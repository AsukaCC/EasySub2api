package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	openAIWSResponseAccountCachePrefix = "openai:response:"
	openAIWSStateStoreCleanupInterval  = time.Minute
	openAIWSStateStoreCleanupMaxPerMap = 512
	openAIWSStateStoreMaxEntriesPerMap = 65536
	openAIWSStateStoreRedisTimeout     = 3 * time.Second
)

type openAIWSAccountBinding struct {
	accountID string
	expiresAt time.Time
}

type openAIWSConnBinding struct {
	connID    string
	expiresAt time.Time
}

type openAIWSTurnStateBinding struct {
	turnState string
	// accountID 铸造该 turn-state 的上游账号；空串表示旧式无溯源绑定。
	// failover 换号后新账号不得回放旧账号铸造的 blob（跨账号矛盾信号）。
	accountID string
	expiresAt time.Time
}

type openAIWSSessionConnBinding struct {
	connID    string
	expiresAt time.Time
}

// OpenAIWSStateStore 管理 WSv2 的粘连状态。
// - response_id -> account_id 用于续链路由
// - response_id -> conn_id 用于连接内上下文复用
//
// response_id -> account_id 优先走 GatewayCache（Redis），同时维护本地热缓存。
// response_id -> conn_id 仅在本进程内有效。
type OpenAIWSStateStore interface {
	BindResponseAccount(ctx context.Context, groupID string, responseID string, accountID string, ttl time.Duration) error
	GetResponseAccount(ctx context.Context, groupID string, responseID string) (string, error)
	DeleteResponseAccount(ctx context.Context, groupID string, responseID string) error

	BindResponseConn(responseID, connID string, ttl time.Duration)
	GetResponseConn(responseID string) (string, bool)
	DeleteResponseConn(responseID string)

	BindSessionTurnState(groupID string, sessionHash, turnState string, ttl time.Duration)
	GetSessionTurnState(groupID string, sessionHash string) (string, bool)
	DeleteSessionTurnState(groupID string, sessionHash string)
	// BindSessionTurnStateForAccount / GetSessionTurnStateForAccount 是带账号溯源的变体：
	// 绑定记录铸造账号；读取时仅当绑定属于同一账号（或为旧式无溯源绑定）才回放，
	// 避免 failover 换号后把旧账号的 turn-state 带进新账号的握手。
	BindSessionTurnStateForAccount(groupID string, sessionHash, accountID, turnState string, ttl time.Duration)
	GetSessionTurnStateForAccount(groupID string, sessionHash, accountID string) (string, bool)

	BindSessionConn(groupID string, sessionHash, connID string, ttl time.Duration)
	GetSessionConn(groupID string, sessionHash string) (string, bool)
	DeleteSessionConn(groupID string, sessionHash string)

	// invalid_encrypted_content lineage：按会话记录已被上游拒绝的
	// encrypted_content 摘要，后续 turn 进场时仅剥离命中项，避免同一失效
	// 密文随客户端历史反复触发"整包被拒→剥离→重试/重连"。仅进程内有效。
	MarkSessionInvalidEncryptedContent(groupID string, sessionHash string, digests []string, ttl time.Duration)
	GetSessionInvalidEncryptedContentDigests(groupID string, sessionHash string) map[string]struct{}
	// HasAnySessionInvalidEncryptedContent 是热路径快速探测：全局无记录时
	// 调用方可跳过会话哈希计算与摘要匹配。
	HasAnySessionInvalidEncryptedContent() bool
}

// openAIWSInvalidEncryptedBinding 记录一个会话中已被上游判定失效的
// encrypted_content 摘要（invalid_encrypted_content lineage）。
type openAIWSInvalidEncryptedBinding struct {
	digests   map[string]struct{}
	expiresAt time.Time
}

// openAIWSInvalidEncryptedDigestsPerSession 是单会话摘要集合的存储自保护上限。
// 超出后新增摘要被忽略，仅退化为"该项下次仍触发一次上游拒绝后的常规 recovery"，
// 不影响正确性。
const openAIWSInvalidEncryptedDigestsPerSession = 512

type defaultOpenAIWSStateStore struct {
	cache GatewayCache

	responseToAccountMu  sync.RWMutex
	responseToAccount    map[string]openAIWSAccountBinding
	responseToConnMu     sync.RWMutex
	responseToConn       map[string]openAIWSConnBinding
	sessionToTurnStateMu sync.RWMutex
	sessionToTurnState   map[string]openAIWSTurnStateBinding
	sessionToConnMu      sync.RWMutex
	sessionToConn        map[string]openAIWSSessionConnBinding

	sessionInvalidEncryptedMu sync.RWMutex
	sessionInvalidEncrypted   map[string]openAIWSInvalidEncryptedBinding

	lastCleanupUnixNano atomic.Int64
}

// NewOpenAIWSStateStore 创建默认 WS 状态存储。
func NewOpenAIWSStateStore(cache GatewayCache) OpenAIWSStateStore {
	store := &defaultOpenAIWSStateStore{
		cache:                   cache,
		responseToAccount:       make(map[string]openAIWSAccountBinding, 256),
		responseToConn:          make(map[string]openAIWSConnBinding, 256),
		sessionToTurnState:      make(map[string]openAIWSTurnStateBinding, 256),
		sessionToConn:           make(map[string]openAIWSSessionConnBinding, 256),
		sessionInvalidEncrypted: make(map[string]openAIWSInvalidEncryptedBinding),
	}
	store.lastCleanupUnixNano.Store(time.Now().UnixNano())
	return store
}

func (s *defaultOpenAIWSStateStore) BindResponseAccount(ctx context.Context, groupID string, responseID string, accountID string, ttl time.Duration) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" || accountID == "" {
		return nil
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	expiresAt := time.Now().Add(ttl)
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	s.responseToAccountMu.Lock()
	ensureBindingCapacity(s.responseToAccount, mapKey, openAIWSStateStoreMaxEntriesPerMap)
	s.responseToAccount[mapKey] = openAIWSAccountBinding{accountID: accountID, expiresAt: expiresAt}
	s.responseToAccountMu.Unlock()

	if s.cache == nil {
		return nil
	}
	cacheKey := openAIWSResponseAccountCacheKey(id)
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	return s.cache.SetSessionAccountID(cacheCtx, groupID, cacheKey, accountID, ttl)
}

func (s *defaultOpenAIWSStateStore) GetResponseAccount(ctx context.Context, groupID string, responseID string) (string, error) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return "", nil
	}
	s.maybeCleanup()

	now := time.Now()
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	s.responseToAccountMu.RLock()
	if binding, ok := s.responseToAccount[mapKey]; ok {
		if now.Before(binding.expiresAt) {
			accountID := binding.accountID
			s.responseToAccountMu.RUnlock()
			return accountID, nil
		}
	}
	s.responseToAccountMu.RUnlock()

	if s.cache == nil {
		return "", nil
	}

	cacheKey := openAIWSResponseAccountCacheKey(id)
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	accountID, err := s.cache.GetSessionAccountID(cacheCtx, groupID, cacheKey)
	if err != nil || accountID == "" {
		// 缓存读取失败不阻断主流程，按未命中降级。
		return "", nil
	}
	return accountID, nil
}

func (s *defaultOpenAIWSStateStore) DeleteResponseAccount(ctx context.Context, groupID string, responseID string) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return nil
	}
	s.responseToAccountMu.Lock()
	delete(s.responseToAccount, openAIWSResponseAccountMapKey(groupID, id))
	s.responseToAccountMu.Unlock()

	if s.cache == nil {
		return nil
	}
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	return s.cache.DeleteSessionAccountID(cacheCtx, groupID, openAIWSResponseAccountCacheKey(id))
}

func (s *defaultOpenAIWSStateStore) BindResponseConn(responseID, connID string, ttl time.Duration) {
	id := normalizeOpenAIWSResponseID(responseID)
	conn := strings.TrimSpace(connID)
	if id == "" || conn == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.responseToConnMu.Lock()
	ensureBindingCapacity(s.responseToConn, id, openAIWSStateStoreMaxEntriesPerMap)
	s.responseToConn[id] = openAIWSConnBinding{
		connID:    conn,
		expiresAt: time.Now().Add(ttl),
	}
	s.responseToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetResponseConn(responseID string) (string, bool) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.responseToConnMu.RLock()
	binding, ok := s.responseToConn[id]
	s.responseToConnMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.connID) == "" {
		return "", false
	}
	return binding.connID, true
}

func (s *defaultOpenAIWSStateStore) DeleteResponseConn(responseID string) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return
	}
	s.responseToConnMu.Lock()
	delete(s.responseToConn, id)
	s.responseToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindSessionTurnState(groupID string, sessionHash, turnState string, ttl time.Duration) {
	s.BindSessionTurnStateForAccount(groupID, sessionHash, "", turnState, ttl)
}

func (s *defaultOpenAIWSStateStore) BindSessionTurnStateForAccount(groupID string, sessionHash, accountID, turnState string, ttl time.Duration) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	state := strings.TrimSpace(turnState)
	if key == "" || state == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.sessionToTurnStateMu.Lock()
	ensureBindingCapacity(s.sessionToTurnState, key, openAIWSStateStoreMaxEntriesPerMap)
	s.sessionToTurnState[key] = openAIWSTurnStateBinding{
		turnState: state,
		accountID: strings.TrimSpace(accountID),
		expiresAt: time.Now().Add(ttl),
	}
	s.sessionToTurnStateMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetSessionTurnState(groupID string, sessionHash string) (string, bool) {
	binding, ok := s.loadSessionTurnStateBinding(groupID, sessionHash)
	if !ok {
		return "", false
	}
	return binding.turnState, true
}

// GetSessionTurnStateForAccount 仅回放同账号铸造（或旧式无溯源）的 turn-state。
// 账号不匹配时视为不存在：调用方按「无 turn-state」出站，由上游为新账号重新铸造。
func (s *defaultOpenAIWSStateStore) GetSessionTurnStateForAccount(groupID string, sessionHash, accountID string) (string, bool) {
	binding, ok := s.loadSessionTurnStateBinding(groupID, sessionHash)
	if !ok {
		return "", false
	}
	accountID = strings.TrimSpace(accountID)
	if binding.accountID != "" && accountID != "" && binding.accountID != accountID {
		return "", false
	}
	return binding.turnState, true
}

func (s *defaultOpenAIWSStateStore) loadSessionTurnStateBinding(groupID string, sessionHash string) (openAIWSTurnStateBinding, bool) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return openAIWSTurnStateBinding{}, false
	}
	s.maybeCleanup()

	now := time.Now()
	s.sessionToTurnStateMu.RLock()
	binding, ok := s.sessionToTurnState[key]
	s.sessionToTurnStateMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.turnState) == "" {
		return openAIWSTurnStateBinding{}, false
	}
	return binding, true
}

func (s *defaultOpenAIWSStateStore) DeleteSessionTurnState(groupID string, sessionHash string) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return
	}
	s.sessionToTurnStateMu.Lock()
	delete(s.sessionToTurnState, key)
	s.sessionToTurnStateMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindSessionConn(groupID string, sessionHash, connID string, ttl time.Duration) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	conn := strings.TrimSpace(connID)
	if key == "" || conn == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.sessionToConnMu.Lock()
	ensureBindingCapacity(s.sessionToConn, key, openAIWSStateStoreMaxEntriesPerMap)
	s.sessionToConn[key] = openAIWSSessionConnBinding{
		connID:    conn,
		expiresAt: time.Now().Add(ttl),
	}
	s.sessionToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetSessionConn(groupID string, sessionHash string) (string, bool) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.sessionToConnMu.RLock()
	binding, ok := s.sessionToConn[key]
	s.sessionToConnMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.connID) == "" {
		return "", false
	}
	return binding.connID, true
}

func (s *defaultOpenAIWSStateStore) DeleteSessionConn(groupID string, sessionHash string) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return
	}
	s.sessionToConnMu.Lock()
	delete(s.sessionToConn, key)
	s.sessionToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) MarkSessionInvalidEncryptedContent(groupID string, sessionHash string, digests []string, ttl time.Duration) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" || len(digests) == 0 {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	now := time.Now()
	overflowed := 0
	s.sessionInvalidEncryptedMu.Lock()
	ensureBindingCapacity(s.sessionInvalidEncrypted, key, openAIWSStateStoreMaxEntriesPerMap)
	binding, ok := s.sessionInvalidEncrypted[key]
	if !ok || now.After(binding.expiresAt) || binding.digests == nil {
		binding = openAIWSInvalidEncryptedBinding{digests: make(map[string]struct{}, len(digests))}
	}
	for _, digest := range digests {
		digest = strings.TrimSpace(digest)
		if digest == "" {
			continue
		}
		if len(binding.digests) >= openAIWSInvalidEncryptedDigestsPerSession {
			if _, exists := binding.digests[digest]; !exists {
				overflowed++
			}
			continue
		}
		binding.digests[digest] = struct{}{}
	}
	binding.expiresAt = now.Add(ttl)
	s.sessionInvalidEncrypted[key] = binding
	s.sessionInvalidEncryptedMu.Unlock()

	// HasAny/Get 在每个 OpenAI 请求热路径上抢同一把锁，日志 IO 必须在锁外。
	if overflowed > 0 {
		logOpenAIWSModeInfo(
			"invalid_encrypted_lineage_capacity_overflow dropped_digests=%d capacity=%d",
			overflowed,
			openAIWSInvalidEncryptedDigestsPerSession,
		)
	}
}

func (s *defaultOpenAIWSStateStore) GetSessionInvalidEncryptedContentDigests(groupID string, sessionHash string) map[string]struct{} {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return nil
	}
	s.maybeCleanup()

	now := time.Now()
	s.sessionInvalidEncryptedMu.RLock()
	defer s.sessionInvalidEncryptedMu.RUnlock()
	binding, ok := s.sessionInvalidEncrypted[key]
	if !ok || now.After(binding.expiresAt) || len(binding.digests) == 0 {
		return nil
	}
	digests := make(map[string]struct{}, len(binding.digests))
	for digest := range binding.digests {
		digests[digest] = struct{}{}
	}
	return digests
}

func (s *defaultOpenAIWSStateStore) HasAnySessionInvalidEncryptedContent() bool {
	s.sessionInvalidEncryptedMu.RLock()
	defer s.sessionInvalidEncryptedMu.RUnlock()
	return len(s.sessionInvalidEncrypted) > 0
}

func (s *defaultOpenAIWSStateStore) maybeCleanup() {
	if s == nil {
		return
	}
	now := time.Now()
	last := time.Unix(0, s.lastCleanupUnixNano.Load())
	if now.Sub(last) < openAIWSStateStoreCleanupInterval {
		return
	}
	if !s.lastCleanupUnixNano.CompareAndSwap(last.UnixNano(), now.UnixNano()) {
		return
	}

	// 增量限额清理，避免高规模下一次性全量扫描导致长时间阻塞。
	s.responseToAccountMu.Lock()
	cleanupExpiredAccountBindings(s.responseToAccount, now, openAIWSStateStoreCleanupMaxPerMap)
	s.responseToAccountMu.Unlock()

	s.responseToConnMu.Lock()
	cleanupExpiredConnBindings(s.responseToConn, now, openAIWSStateStoreCleanupMaxPerMap)
	s.responseToConnMu.Unlock()

	s.sessionToTurnStateMu.Lock()
	cleanupExpiredTurnStateBindings(s.sessionToTurnState, now, openAIWSStateStoreCleanupMaxPerMap)
	s.sessionToTurnStateMu.Unlock()

	s.sessionToConnMu.Lock()
	cleanupExpiredSessionConnBindings(s.sessionToConn, now, openAIWSStateStoreCleanupMaxPerMap)
	s.sessionToConnMu.Unlock()

	s.sessionInvalidEncryptedMu.Lock()
	cleanupExpiredInvalidEncryptedBindings(s.sessionInvalidEncrypted, now, openAIWSStateStoreCleanupMaxPerMap)
	s.sessionInvalidEncryptedMu.Unlock()
}

func cleanupExpiredInvalidEncryptedBindings(bindings map[string]openAIWSInvalidEncryptedBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredAccountBindings(bindings map[string]openAIWSAccountBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredConnBindings(bindings map[string]openAIWSConnBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredTurnStateBindings(bindings map[string]openAIWSTurnStateBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredSessionConnBindings(bindings map[string]openAIWSSessionConnBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func ensureBindingCapacity[T any](bindings map[string]T, incomingKey string, maxEntries int) {
	if len(bindings) < maxEntries || maxEntries <= 0 {
		return
	}
	if _, exists := bindings[incomingKey]; exists {
		return
	}
	// 固定上限保护：淘汰任意一项，优先保证内存有界。
	for key := range bindings {
		delete(bindings, key)
		return
	}
}

func normalizeOpenAIWSResponseID(responseID string) string {
	return strings.TrimSpace(responseID)
}

func openAIWSResponseAccountCacheKey(responseID string) string {
	sum := sha256.Sum256([]byte(responseID))
	return openAIWSResponseAccountCachePrefix + hex.EncodeToString(sum[:])
}

// openAIWSResponseAccountMapKey 本地热缓存按分组隔离的 key，与 Redis 层保持一致，避免跨组命中。
func openAIWSResponseAccountMapKey(groupID string, responseID string) string {
	return fmt.Sprintf("%v:%s", groupID, responseID)
}

func normalizeOpenAIWSTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return time.Hour
	}
	return ttl
}

func openAIWSSessionTurnStateKey(groupID string, sessionHash string) string {
	hash := strings.TrimSpace(sessionHash)
	if hash == "" {
		return ""
	}
	return fmt.Sprintf("%v:%s", groupID, hash)
}

func withOpenAIWSStateStoreRedisTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, openAIWSStateStoreRedisTimeout)
}
