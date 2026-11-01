package service

import (
	"context"
	"time"
)

type modelFingerprintAccountLoad interface {
	GetAccountsLoadBatchFresh(context.Context, []AccountWithConcurrency) (map[string]*AccountLoadInfo, error)
}

// Use persisted quota windows and live request counts without querying upstream.
func (s *AccountTestService) fingerprintSkipReason(ctx context.Context, account *Account) string {
	if account == nil {
		return "account_unavailable"
	}
	now := time.Now()
	for _, window := range []string{"5h", "7d"} {
		if usage := buildCodexUsageProgressFromExtra(account.Extra, window, now); usage != nil && usage.Utilization >= 100 {
			return "quota_exhausted"
		}
	}
	utilization, hasUtilization := account.Extra["session_window_utilization"]
	if (parseExtraFloat64(utilization) >= 1 || (!hasUtilization && account.SessionWindowStatus == "rejected")) &&
		(account.SessionWindowEnd == nil || now.Before(*account.SessionWindowEnd)) {
		return "quota_exhausted"
	}
	if usage := buildPassiveUsageWindow(account.Extra, "passive_usage_7d_utilization", "passive_usage_7d_reset"); usage != nil &&
		usage.Utilization >= 100 && (usage.ResetsAt == nil || now.Before(*usage.ResetsAt)) {
		return "quota_exhausted"
	}
	if !account.IsSchedulable() {
		return "account_unavailable"
	}
	if s.modelFingerprintLoad == nil {
		return "account_activity_unknown"
	}
	loadCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	loads, err := s.modelFingerprintLoad.GetAccountsLoadBatchFresh(loadCtx, []AccountWithConcurrency{{ID: account.ID, MaxConcurrency: account.Concurrency}})
	if err != nil || loads[account.ID] == nil {
		return "account_activity_unknown"
	}
	load := loads[account.ID]
	if load.CurrentConcurrency > 0 || load.WaitingCount > 0 {
		return "account_busy"
	}
	return ""
}
