//go:build unit

package admin

import (
	"encoding/json"
	"testing"
)

func TestGroupLimitPointsInputAlias(t *testing.T) {
	var req CreateGroupRequest
	if err := json.Unmarshal([]byte(`{"name":"points","daily_limit_points":12.5}`), &req); err != nil {
		t.Fatalf("decode group request: %v", err)
	}
	resolved, err := resolveLimitFieldAlias(req.DailyLimitPoints, req.DailyLimitUSD, "daily_limit_points", "daily_limit_usd")
	if err != nil {
		t.Fatalf("resolve points alias: %v", err)
	}
	if resolved.value == nil || *resolved.value != 12.5 {
		t.Fatalf("resolved daily limit = %v, want 12.5", resolved.value)
	}
}

func TestGroupLimitPointsInputRejectsConflictingLegacyValue(t *testing.T) {
	var req UpdateGroupRequest
	if err := json.Unmarshal([]byte(`{"daily_limit_points":12.5,"daily_limit_usd":10}`), &req); err != nil {
		t.Fatalf("decode group request: %v", err)
	}
	if _, err := resolveLimitFieldAlias(req.DailyLimitPoints, req.DailyLimitUSD, "daily_limit_points", "daily_limit_usd"); err == nil {
		t.Fatal("expected conflicting points and legacy limits to be rejected")
	}
}

func TestUserPlatformQuotaPointsInputAlias(t *testing.T) {
	var req UpdateUserPlatformQuotasRequest
	if err := json.Unmarshal([]byte(`{"quotas":[{"platform":"anthropic","daily_limit_points":8.25}]}`), &req); err != nil {
		t.Fatalf("decode user platform quota request: %v", err)
	}
	if len(req.Quotas) != 1 {
		t.Fatalf("quota count = %d, want 1", len(req.Quotas))
	}
	resolved, err := resolveLimitFieldAlias(req.Quotas[0].DailyLimitPoints, req.Quotas[0].DailyLimitUSD, "daily_limit_points", "daily_limit_usd")
	if err != nil {
		t.Fatalf("resolve points alias: %v", err)
	}
	if resolved.value == nil || *resolved.value != 8.25 {
		t.Fatalf("resolved daily limit = %v, want 8.25", resolved.value)
	}
}

func TestLimitPointsInputAllowsMatchingLegacyValue(t *testing.T) {
	var req CreateGroupRequest
	if err := json.Unmarshal([]byte(`{"name":"points","monthly_limit_points":100,"monthly_limit_usd":100}`), &req); err != nil {
		t.Fatalf("decode group request: %v", err)
	}
	resolved, err := resolveLimitFieldAlias(req.MonthlyLimitPoints, req.MonthlyLimitUSD, "monthly_limit_points", "monthly_limit_usd")
	if err != nil {
		t.Fatalf("matching aliases should be accepted: %v", err)
	}
	if !resolved.set || resolved.value == nil || *resolved.value != 100 {
		t.Fatalf("resolved monthly limit = %+v, want 100", resolved)
	}
}
