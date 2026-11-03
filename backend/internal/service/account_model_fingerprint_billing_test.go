//go:build unit

package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fingerprintBillingStore struct {
	UserRepository
	mu                sync.Mutex
	balance           float64
	commands          map[string]UsageBillingCommand
	exhaustAfterFirst bool
}

func (r *fingerprintBillingStore) GetByID(context.Context, string) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return &User{ID: "admin", Status: StatusActive, Balance: r.balance}, nil
}

func (r *fingerprintBillingStore) Apply(ctx context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.commands[cmd.RequestID]; exists {
		return &UsageBillingApplyResult{}, nil
	}
	r.commands[cmd.RequestID] = *cmd
	r.balance -= cmd.BalanceCost
	if r.exhaustAfterFirst {
		r.balance = 0
	}
	balance := r.balance
	return &UsageBillingApplyResult{Applied: true, NewBalance: &balance}, nil
}

type fingerprintBillingLogs struct {
	UsageLogRepository
	mu   sync.Mutex
	logs map[string]UsageLog
}

func (r *fingerprintBillingLogs) Create(ctx context.Context, log *UsageLog) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.logs[log.RequestID]
	r.logs[log.RequestID] = *log
	return !exists, nil
}

func attachFingerprintBilling(svc *AccountTestService) (*fingerprintBillingStore, *fingerprintBillingLogs) {
	store := &fingerprintBillingStore{balance: 100, commands: make(map[string]UsageBillingCommand)}
	logs := &fingerprintBillingLogs{logs: make(map[string]UsageLog)}
	gateway := newOpenAIRecordUsageServiceWithBillingRepoForTest(logs, store, store, nil, nil)
	gateway.billingCacheService.userRepo = store
	svc.modelFingerprintUsage = &modelFingerprintBilling{gateway: gateway, keyQuota: &openAIRecordUsageAPIKeyQuotaStub{}}
	svc.modelFingerprintGroups.(*fingerprintGroupRepo).group.RateMultiplier = 2
	svc.modelFingerprintKeys.(*fingerprintKeyRepo).keys[0].Quota = 100
	svc.modelFingerprintKeys.(*fingerprintKeyRepo).keys[0].Group = &Group{ID: "other-primary-group", RateMultiplier: 99}
	return store, logs
}

func TestModelFingerprintBillingEveryRequest(t *testing.T) {
	for _, source := range []string{"manual", "scheduled", "native"} {
		t.Run(source, func(t *testing.T) {
			svc, repo, native := newFingerprintTestService(t)
			store, logs := attachFingerprintBilling(svc)
			svc.httpUpstream = &fingerprintAPITransport{}
			option := ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat", Source: source}
			if source == "native" {
				svc.accountRepo = &nativeFingerprintRepo{repo}
				native.responses = true
				svc.httpUpstream = native
				option.Protocol, option.Source = "auto", "manual"
			}
			job, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", option)
			require.NoError(t, err)
			require.Equal(t, "completed", waitFingerprint(t, repo.done).Status)
			store.mu.Lock()
			defer store.mu.Unlock()
			logs.mu.Lock()
			defer logs.mu.Unlock()
			require.Len(t, store.commands, 3)
			require.Len(t, logs.logs, 3)
			total := 0.0
			for i := 1; i <= 3; i++ {
				id := fmt.Sprintf("fingerprint:%s:%d", job.ID, i)
				cmd, ok := store.commands[id]
				require.True(t, ok)
				require.Positive(t, cmd.BalanceCost)
				require.Equal(t, cmd.BalanceCost, cmd.APIKeyQuotaCost)
				require.Equal(t, "local-key", cmd.APIKeyID)
				require.Equal(t, 40, cmd.InputTokens)
				require.Equal(t, 80, cmd.OutputTokens)
				log := logs.logs[id]
				require.Equal(t, RequestTypeTest, log.RequestType)
				require.Equal(t, 2.0, log.RateMultiplier)
				require.Equal(t, cmd.BalanceCost, log.ActualCost)
				total += cmd.BalanceCost
			}
			require.InDelta(t, 100-total, store.balance, 1e-8)
		})
	}
}

func TestModelFingerprintBillingStopsWhenBalanceExhausted(t *testing.T) {
	svc, repo, _ := newFingerprintTestService(t)
	store, _ := attachFingerprintBilling(svc)
	store.exhaustAfterFirst = true
	transport := &fingerprintAPITransport{}
	svc.httpUpstream = transport
	_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat"})
	require.NoError(t, err)
	require.Equal(t, "failed", waitFingerprint(t, repo.done).Status)
	store.mu.Lock()
	require.Len(t, store.commands, 1)
	store.mu.Unlock()
	transport.mu.Lock()
	defer transport.mu.Unlock()
	require.Len(t, transport.payloads, 1)
}

func TestModelFingerprintBillingCancelledPartialUsageAndIdempotency(t *testing.T) {
	svc, _, _ := newFingerprintTestService(t)
	store, logs := attachFingerprintBilling(svc)
	account, err := svc.accountRepo.GetByID(context.Background(), "one")
	require.NoError(t, err)
	key, err := svc.fingerprintKey(context.Background(), account, "admin", "local-key")
	require.NoError(t, err)
	prepared, err := svc.modelFingerprintUsage.Prepare(context.Background(), account, key)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	probe := &modelFingerprintProbe{billing: prepared, startedAt: time.Now(), cancel: cancel, expected: 2}
	probe.event(TestEvent{Type: "content", Text: "1,2,"})
	require.True(t, probe.enough)
	require.NoError(t, ctx.Err(), "sampling must leave the stream open for final usage")
	probe.usage = UsageLog{InputTokens: 10, OutputTokens: 20, CacheReadTokens: 30, CacheCreationTokens: 40}
	cancel()
	job := &ModelFingerprintSnapshot{ID: "partial", Model: "gpt-6-astra", APIKeyID: key.ID, GroupID: key.GroupID}
	for range 2 {
		require.NoError(t, svc.recordModelFingerprintUsage("admin", "one", job, 0, probe))
	}
	require.Len(t, store.commands, 1)
	cmd := store.commands["fingerprint:partial:1"]
	require.Equal(t, 10, cmd.InputTokens)
	require.Equal(t, 30, cmd.CacheReadTokens)
	require.Equal(t, 40, cmd.CacheCreationTokens)
	require.Positive(t, cmd.BalanceCost)
	require.InDelta(t, 100-cmd.BalanceCost, store.balance, 1e-8)
	require.Equal(t, cmd.BalanceCost, logs.logs[cmd.RequestID].ActualCost)
}

func TestModelFingerprintBillingErrorResponseUsage(t *testing.T) {
	svc, repo, _ := newFingerprintTestService(t)
	store, _ := attachFingerprintBilling(svc)
	svc.httpUpstream = &fingerprintAPITransport{status: 429}
	_, err := svc.StartModelFingerprint(context.Background(), "one", "gpt-6-astra", "admin", ModelFingerprintOptions{APIKeyID: "local-key", Protocol: "chat"})
	require.NoError(t, err)
	require.Equal(t, "failed", waitFingerprint(t, repo.done).Status)
	store.mu.Lock()
	defer store.mu.Unlock()
	require.Len(t, store.commands, 1)
	for _, cmd := range store.commands {
		require.Positive(t, cmd.BalanceCost)
	}
}

type fingerprintBillingSubscriptionRepo struct {
	UserSubscriptionRepository
	subscription *UserSubscription
}

func (r *fingerprintBillingSubscriptionRepo) GetActiveByUserIDAndGroupID(context.Context, string, string) (*UserSubscription, error) {
	return r.subscription, nil
}

func TestModelFingerprintBillingUsesSubscription(t *testing.T) {
	svc, _, _ := newFingerprintTestService(t)
	store, logs := attachFingerprintBilling(svc)
	billing := svc.modelFingerprintUsage.(*modelFingerprintBilling)
	subRepo := &fingerprintBillingSubscriptionRepo{subscription: &UserSubscription{
		ID: "subscription", UserID: "admin", GroupID: "fingerprint-group",
		Status: SubscriptionStatusActive, ExpiresAt: time.Now().Add(time.Hour),
	}}
	billing.gateway.userSubRepo = subRepo
	billing.gateway.billingCacheService.subRepo = subRepo
	svc.modelFingerprintGroups.(*fingerprintGroupRepo).group.SubscriptionType = SubscriptionTypeSubscription
	account, err := svc.accountRepo.GetByID(context.Background(), "one")
	require.NoError(t, err)
	key, err := svc.fingerprintKey(context.Background(), account, "admin", "local-key")
	require.NoError(t, err)
	prepared, err := billing.Prepare(context.Background(), account, key)
	require.NoError(t, err)
	log := &UsageLog{RequestID: "fingerprint:subscription:1", Model: "gpt-6-astra", InputTokens: 40, OutputTokens: 80, CreatedAt: time.Now()}
	require.NoError(t, billing.Record(context.Background(), log, prepared))
	cmd := store.commands[log.RequestID]
	require.Zero(t, cmd.BalanceCost)
	require.Positive(t, cmd.SubscriptionCost)
	require.Equal(t, "subscription", *cmd.SubscriptionID)
	require.Equal(t, cmd.SubscriptionCost, cmd.APIKeyQuotaCost)
	require.Equal(t, BillingTypeSubscription, logs.logs[log.RequestID].BillingType)
	require.Equal(t, 100.0, store.balance)
}
