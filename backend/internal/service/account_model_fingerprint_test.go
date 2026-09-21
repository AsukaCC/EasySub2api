//go:build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	apperrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type fingerprintTestRepo struct {
	AccountRepository
	mu        sync.Mutex
	snapshots map[string]*ModelFingerprintSnapshot
	done      chan *ModelFingerprintSnapshot
	mapping   map[string]any
}

func (r *fingerprintTestRepo) GetByID(_ context.Context, id string) (*Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &Account{ID: id, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive,
		Credentials: map[string]any{"api_key": "test-only", "base_url": "https://example.com", "model_mapping": r.mapping},
		Extra:       map[string]any{"openai_responses_mode": "force_chat_completions", ModelFingerprintExtraKey: r.snapshots[id]}}, nil
}
func (r *fingerprintTestRepo) ClaimModelFingerprint(_ context.Context, id string, snapshot *ModelFingerprintSnapshot) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if previous := r.snapshots[id]; previous != nil && previous.Status == "running" && time.Now().Before(previous.ExpiresAt) {
		return false, nil
	}
	copy := *snapshot
	r.snapshots[id] = &copy
	return true, nil
}
func (r *fingerprintTestRepo) SaveModelFingerprint(_ context.Context, id string, snapshot *ModelFingerprintSnapshot) error {
	r.mu.Lock()
	copy := *snapshot
	r.snapshots[id] = &copy
	r.mu.Unlock()
	if snapshot.Status != "running" {
		r.done <- &copy
	}
	return nil
}

type fingerprintTestTransport struct {
	HTTPUpstream
	mu       sync.Mutex
	requests map[string][][]byte
	sessions map[string][]string
	entered  chan string
	release  chan struct{}
	invalid  bool
	fail     bool
}

func (u *fingerprintTestTransport) DoWithTLS(req *http.Request, _ string, accountID string, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	u.mu.Lock()
	u.requests[accountID] = append(u.requests[accountID], body)
	u.sessions[accountID] = append(u.sessions[accountID], req.Header.Get("Session_ID"))
	first := len(u.requests[accountID]) == 1
	u.mu.Unlock()
	if u.fail {
		return nil, fmt.Errorf("upstream unavailable")
	}
	if first && u.entered != nil {
		u.entered <- accountID
		select {
		case <-u.release:
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
	text := strings.Repeat("17,42,138,251,", 60)
	if u.invalid {
		text = "I cannot provide numbers"
	}
	delta, _ := json.Marshal(map[string]any{"usage": map[string]int{"prompt_tokens": 40, "completion_tokens": 80}, "choices": []any{map[string]any{"delta": map[string]string{"content": text}}}})
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(fmt.Sprintf("data: %s\n\ndata: [DONE]\n\n", delta)))}, nil
}

func newFingerprintTestService() (*AccountTestService, *fingerprintTestRepo, *fingerprintTestTransport) {
	repo := &fingerprintTestRepo{snapshots: make(map[string]*ModelFingerprintSnapshot), done: make(chan *ModelFingerprintSnapshot, 8)}
	upstream := &fingerprintTestTransport{requests: make(map[string][][]byte), sessions: make(map[string][]string)}
	return &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: &config.Config{}, modelFingerprintUsage: &fingerprintUsageStub{}}, repo, upstream
}

type fingerprintUsageStub struct {
	mu   sync.Mutex
	logs []UsageLog
	err  error
}

func (r *fingerprintUsageStub) Create(ctx context.Context, log *UsageLog) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	r.logs = append(r.logs, *log)
	return r.err == nil, r.err
}

func waitFingerprint(t *testing.T, done <-chan *ModelFingerprintSnapshot) *ModelFingerprintSnapshot {
	t.Helper()
	select {
	case result := <-done:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("test worker did not finish")
		return nil
	}
}

func TestModelFingerprintRetention(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name     string
		status   string
		age      time.Duration
		finished bool
		visible  bool
	}{
		{"completed before boundary", "completed", 2*time.Hour - time.Nanosecond, true, true},
		{"completed at boundary", "completed", 2 * time.Hour, true, false},
		{"completed after boundary", "completed", 3 * time.Hour, true, false},
		{"failure at boundary", "failed", 2 * time.Hour, true, false},
		{"active lease", "running", -time.Minute, false, true},
		{"recent interrupted worker", "running", time.Minute, false, true},
		{"expired interrupted worker", "running", 2 * time.Hour, false, false},
		{"legacy result", "completed", time.Hour, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			finished := now.Add(-tc.age)
			snapshot := &ModelFingerprintSnapshot{ID: "test", Status: tc.status, StartedAt: finished.Add(-4 * time.Minute), ExpiresAt: finished}
			if tc.finished {
				snapshot.FinishedAt = &finished
				// A worker lease that is still recent cannot extend an old result.
				snapshot.ExpiresAt = now.Add(time.Minute)
			}
			got, err := ParseModelFingerprintSnapshot(snapshot, now)
			require.NoError(t, err)
			require.Equal(t, tc.visible, got != nil)
			if tc.visible && tc.status == "running" && tc.age > 0 {
				require.Equal(t, "interrupted", got.Error)
				require.Equal(t, finished, *got.FinishedAt)
				require.Equal(t, "running", snapshot.Status)
			}
		})
	}
	svc, repo, _ := newFingerprintTestService()
	old := time.Now().Add(-3 * time.Hour)
	repo.snapshots["expired"] = &ModelFingerprintSnapshot{Status: "completed", FinishedAt: &old}
	got, err := svc.GetModelFingerprint(context.Background(), "expired")
	require.NoError(t, err)
	require.Nil(t, got)
}

type fingerprintExpiryRepo struct {
	AccountRepository
	cleanupErr error
	cleaned    bool
	paused     bool
}

func (r *fingerprintExpiryRepo) DeleteExpiredModelFingerprints(ctx context.Context, now time.Time) (int64, error) {
	r.cleaned = ctx.Err() == nil && time.Since(now) < time.Second
	return 0, r.cleanupErr
}

func (r *fingerprintExpiryRepo) AutoPauseExpiredAccounts(context.Context, time.Time) (int64, error) {
	r.paused = true
	return 0, nil
}

func TestModelFingerprintCleanupRunsWithAccountExpiry(t *testing.T) {
	for _, cleanupErr := range []error{nil, fmt.Errorf("cleanup unavailable")} {
		repo := &fingerprintExpiryRepo{cleanupErr: cleanupErr}
		NewAccountExpiryService(repo, time.Minute).runOnce()
		require.True(t, repo.cleaned)
		require.True(t, repo.paused, "cleanup failure must not block account expiry")
	}
}

func TestModelFingerprintSingleConversationAndAccountIsolation(t *testing.T) {
	svc, repo, upstream := newFingerprintTestService()
	upstream.entered = make(chan string, 2)
	upstream.release = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	first, err := svc.StartModelFingerprint(ctx, "one", "gpt-6-astra", "admin")
	require.NoError(t, err)
	cancel() // Closing the browser request must not cancel the worker.
	select {
	case <-upstream.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("worker not started")
	}
	_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin")
	require.Equal(t, 409, apperrors.Code(err))
	second, err := svc.StartModelFingerprint(context.Background(), "two", "gpt-6-astra", "admin")
	require.NoError(t, err)
	select {
	case <-upstream.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("other account cannot run concurrently")
	}
	upstream.mu.Lock()
	require.Len(t, upstream.requests["one"], 1)
	require.Len(t, upstream.requests["two"], 1)
	upstream.mu.Unlock()
	close(upstream.release)
	for range 2 {
		result := waitFingerprint(t, repo.done)
		require.Equal(t, "completed", result.Status)
		require.Equal(t, 3, result.Valid)
		require.NotNil(t, result.Result)
	}
	upstream.mu.Lock()
	defer upstream.mu.Unlock()
	for _, item := range []struct{ id, session string }{{"one", first.ID}, {"two", second.ID}} {
		require.Len(t, upstream.requests[item.id], 3)
		for i, body := range upstream.requests[item.id] {
			require.Len(t, gjson.GetBytes(body, "messages").Array(), 2*i+1)
			require.Equal(t, item.session, upstream.sessions[item.id][i])
			require.Equal(t, int64(modelFingerprintMaxTokens), gjson.GetBytes(body, "max_completion_tokens").Int())
			require.Equal(t, "low", gjson.GetBytes(body, "reasoning_effort").String())
			if i > 0 {
				require.Equal(t, "assistant", gjson.GetBytes(body, "messages.1.role").String())
				require.Contains(t, gjson.GetBytes(body, "messages.1.content").String(), "17,42")
			}
		}
	}
}

func TestModelFingerprintInvalidSamplesAndMedia(t *testing.T) {
	svc, repo, upstream := newFingerprintTestService()
	upstream.invalid = true
	_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-image-2", "admin")
	require.Equal(t, 400, apperrors.Code(err))
	_, err = svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin")
	require.NoError(t, err)
	result := waitFingerprint(t, repo.done)
	require.Equal(t, "insufficient_samples", result.Error)
	require.Nil(t, result.Result)
	require.Equal(t, 3, result.Completed)
	stale := &ModelFingerprintSnapshot{Status: "running", ExpiresAt: time.Now().Add(-time.Minute)}
	repo.mu.Lock()
	repo.snapshots["stale"] = stale
	repo.mu.Unlock()
	status, err := svc.GetModelFingerprint(context.Background(), "stale")
	require.NoError(t, err)
	require.Equal(t, "interrupted", status.Error)
}

func TestModelFingerprintPayloadProtocols(t *testing.T) {
	conversation := &modelFingerprintConversation{id: "session", turns: []modelFingerprintTurn{{role: "user", text: "first"}, {role: "assistant", text: "17,42"}}}
	ctx := context.WithValue(context.Background(), modelFingerprintContextKey{}, &modelFingerprintProbe{conversation: conversation, prompt: "next"})
	for _, tc := range []struct{ protocol, key, role string }{{"anthropic", "messages", "assistant"}, {"chat", "messages", "assistant"}, {"responses", "input", "assistant"}, {"gemini", "contents", "model"}} {
		payload := map[string]any{"model": "gpt-6-astra"}
		applyModelFingerprintPayload(ctx, payload, tc.protocol, false)
		turns := payload[tc.key].([]map[string]any)
		require.Len(t, turns, 3)
		require.Equal(t, tc.role, turns[1]["role"])
	}
	codex := map[string]any{}
	applyModelFingerprintPayload(ctx, codex, "responses", true)
	require.NotContains(t, codex, "max_output_tokens")
	unchanged := map[string]any{"input": "hi"}
	applyModelFingerprintPayload(context.Background(), unchanged, "responses", false)
	require.Equal(t, map[string]any{"input": "hi"}, unchanged)
}

func TestModelFingerprintDirectModelAndUsage(t *testing.T) {
	for _, mapping := range []map[string]any{
		{"other-model": "other-model"},
		{"gpt-6-astra": "gpt-image-2"},
	} {
		svc, repo, upstream := newFingerprintTestService()
		repo.mapping = mapping
		job, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin")
		require.NoError(t, err)
		require.Equal(t, "completed", waitFingerprint(t, repo.done).Status)
		writer := svc.modelFingerprintUsage.(*fingerprintUsageStub)
		writer.mu.Lock()
		require.Len(t, writer.logs, 3)
		for i, log := range writer.logs {
			require.Empty(t, log.APIKeyID)
			require.Equal(t, "admin", log.UserID)
			require.Equal(t, "one", log.AccountID)
			require.Equal(t, RequestTypeTest, log.RequestType)
			require.Equal(t, "low", *log.ReasoningEffort)
			require.Equal(t, "low", *log.RequestedReasoningEffort)
			require.Equal(t, "gpt-6-astra", log.Model)
			require.Equal(t, job.ID, *log.SessionID)
			require.Equal(t, fmt.Sprintf("fingerprint:%s:%d", job.ID, i+1), log.RequestID)
			require.Zero(t, log.ActualCost)
			require.Equal(t, 40, log.InputTokens)
			require.Equal(t, 80, log.OutputTokens)
			require.NotNil(t, log.DurationMs)
		}
		writer.mu.Unlock()
		upstream.mu.Lock()
		for _, body := range upstream.requests["one"] {
			require.Equal(t, "gpt-6-astra", gjson.GetBytes(body, "model").String())
		}
		upstream.mu.Unlock()
		account, err := repo.GetByID(context.Background(), "one")
		require.NoError(t, err)
		require.Equal(t, mapping, account.Credentials["model_mapping"])
	}
}

func TestModelFingerprintFailureUsage(t *testing.T) {
	for _, writeFailure := range []bool{false, true} {
		svc, repo, upstream := newFingerprintTestService()
		writer := svc.modelFingerprintUsage.(*fingerprintUsageStub)
		if writeFailure {
			writer.err = fmt.Errorf("storage unavailable")
		} else {
			upstream.fail = true
		}
		_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin")
		require.NoError(t, err)
		result := waitFingerprint(t, repo.done)
		require.Equal(t, "failed", result.Status)
		require.Equal(t, 1, result.Completed)
		writer.mu.Lock()
		require.Len(t, writer.logs, 1)
		require.Equal(t, RequestTypeTest, writer.logs[0].RequestType)
		writer.mu.Unlock()
		upstream.mu.Lock()
		require.Len(t, upstream.requests["one"], 1, "failures must not retry billable model requests")
		upstream.mu.Unlock()
	}
}

func TestModelFingerprintNativeUsageCounters(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		events               []string
		input, output, cache int
	}{
		{"anthropic", []string{`{"message":{"model":"claude-sonnet-4-6","usage":{"input_tokens":12,"cache_read_input_tokens":8}}}`, `{"usage":{"output_tokens":20}}`}, 12, 20, 8},
		{"responses", []string{`{"response":{"usage":{"input_tokens":30,"output_tokens":14,"input_tokens_details":{"cached_tokens":10}}}}`}, 20, 14, 10},
		{"chat", []string{`{"usage":{"prompt_tokens":25,"completion_tokens":9,"prompt_tokens_details":{"cached_tokens":5}}}`}, 20, 9, 5},
		{"gemini", []string{`{"response":{"usageMetadata":{"promptTokenCount":24,"candidatesTokenCount":10,"thoughtsTokenCount":5,"cachedContentTokenCount":4}}}`}, 20, 15, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := &modelFingerprintProbe{}
			ctx := context.WithValue(context.Background(), modelFingerprintContextKey{}, probe)
			for _, event := range tc.events {
				var data map[string]any
				require.NoError(t, json.Unmarshal([]byte(event), &data))
				collectModelFingerprintUsage(ctx, data)
			}
			require.Equal(t, tc.input, probe.usage.InputTokens)
			require.Equal(t, tc.output, probe.usage.OutputTokens)
			require.Equal(t, tc.cache, probe.usage.CacheReadTokens)
			require.Zero(t, probe.usage.ActualCost)
		})
	}
}

func TestModelFingerprintLowEffort(t *testing.T) {
	for _, tc := range []struct {
		name, model, protocol, path string
		want                        any
	}{
		{"chat", "gpt-6-astra", "chat", "reasoning_effort", "low"},
		{"responses", "gpt-6-astra", "responses", "reasoning.effort", "low"},
		{"grok chat", "grok-3-mini", "chat", "reasoning_effort", "low"},
		{"grok responses", "grok-4.3", "responses", "reasoning.effort", "low"},
		{"claude", "claude-sonnet-4-6", "anthropic", "output_config.effort", "low"},
		{"gemini level", "gemini-3.1-pro-preview", "gemini", "generationConfig.thinkingConfig.thinkingLevel", "low"},
		{"gemini budget", "gemini-2.5-pro", "gemini", "generationConfig.thinkingConfig.thinkingBudget", float64(1024)},
		{"antigravity claude", "claude-sonnet-4-6", "gemini", "generationConfig.thinkingConfig.thinkingBudget", float64(1024)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := &modelFingerprintProbe{model: tc.model, conversation: &modelFingerprintConversation{}, prompt: "numbers"}
			ctx := context.WithValue(context.Background(), modelFingerprintContextKey{}, probe)
			payload := map[string]any{}
			if tc.protocol != "gemini" {
				payload["model"] = tc.model
			}
			switch tc.protocol {
			case "chat":
				payload["reasoning_effort"] = "high"
			case "responses":
				payload["reasoning"] = map[string]any{"effort": "xhigh"}
			case "anthropic":
				payload["thinking"] = map[string]any{"type": "enabled", "budget_tokens": 8192}
				payload["output_config"] = map[string]any{"effort": "max"}
			case "gemini":
				payload["generationConfig"] = map[string]any{"thinkingConfig": map[string]any{"thinkingBudget": -1, "thinkingLevel": "high"}}
			}
			applyModelFingerprintPayload(ctx, payload, tc.protocol, tc.protocol == "responses")
			body, err := json.Marshal(payload)
			require.NoError(t, err)
			require.Equal(t, tc.want, gjson.GetBytes(body, tc.path).Value())
			require.Equal(t, "low", *probe.usage.ReasoningEffort)
			require.Equal(t, "low", *probe.usage.RequestedReasoningEffort)
			if tc.protocol == "anthropic" {
				require.NotContains(t, payload, "thinking")
			}
		})
	}
	for _, tc := range []struct{ model, protocol, field string }{
		{"gpt-4.1", "chat", "reasoning_effort"},
		{"gpt-4.1", "responses", "reasoning"},
		{"claude-haiku-4-5", "anthropic", "output_config"},
		{"gemini-2.0-flash", "gemini", "generationConfig.thinkingConfig"},
	} {
		probe := &modelFingerprintProbe{model: tc.model, conversation: &modelFingerprintConversation{}}
		ctx := context.WithValue(context.Background(), modelFingerprintContextKey{}, probe)
		payload := map[string]any{"model": tc.model}
		applyModelFingerprintPayload(ctx, payload, tc.protocol, false)
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		require.False(t, gjson.GetBytes(body, tc.field).Exists(), "unsupported model %s", tc.model)
		require.Nil(t, probe.usage.ReasoningEffort)
	}
}
