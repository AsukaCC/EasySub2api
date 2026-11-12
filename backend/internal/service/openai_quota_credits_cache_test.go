package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheCodexCreditsSnapshotPerAccount(t *testing.T) {
	repo := &stubQuotaAccountRepo{}
	svc := &OpenAIQuotaService{accountRepo: repo}
	balance := "123.4567890123456789"
	usage := &OpenAIQuotaUsage{FetchedAt: 123, Credits: &OpenAICredits{HasCredits: true, Balance: &balance}}
	require.NoError(t, svc.CacheCreditsSnapshot(context.Background(), "one", usage))
	zero := "0"
	require.NoError(t, svc.CacheCreditsSnapshot(context.Background(), "two", &OpenAIQuotaUsage{
		FetchedAt: 456, Credits: &OpenAICredits{Balance: &zero},
	}))
	body, err := json.Marshal(repo.extraUpdates["one"][openaiQuotaCreditsKey])
	require.NoError(t, err)
	require.JSONEq(t, `{"credits":{"has_credits":true,"unlimited":false,"balance":"123.4567890123456789"},"fetched_at":123}`, string(body))
	body, err = json.Marshal(repo.extraUpdates["two"][openaiQuotaCreditsKey])
	require.NoError(t, err)
	require.JSONEq(t, `{"credits":{"has_credits":false,"unlimited":false,"balance":"0"},"fetched_at":456}`, string(body))
	require.Error(t, svc.CacheCreditsSnapshot(context.Background(), "one", nil))
	repo.extraUpdateErr = errors.New("database unavailable")
	require.Error(t, svc.CacheCreditsSnapshot(context.Background(), "one", &OpenAIQuotaUsage{FetchedAt: 789}))
	require.Equal(t, int64(123), repo.extraUpdates["one"][openaiQuotaCreditsKey].(openAICreditsSnapshot).FetchedAt)
}
