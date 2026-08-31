package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// AccountWeeklyQuotaSnapshot is the dashboard-safe, normalized weekly quota
// projection. It is intentionally built from persisted account state only;
// constructing it never contacts an upstream provider.
type AccountWeeklyQuotaSnapshot struct {
	Known        bool       `json:"known"`
	UsedPercent  float64    `json:"used_percent"`
	RemainingPct float64    `json:"remaining_percent"`
	ResetAt      *time.Time `json:"reset_at,omitempty"`
	ObservedAt   *time.Time `json:"observed_at,omitempty"`
	Source       string     `json:"source,omitempty"`
	Expired      bool       `json:"expired,omitempty"`
}

type AccountQuotaAccount struct {
	ID                  string                     `json:"id"`
	Name                string                     `json:"name"`
	Platform            string                     `json:"platform"`
	Status              string                     `json:"status"`
	Schedulable         bool                       `json:"schedulable"`
	Weekly              AccountWeeklyQuotaSnapshot `json:"weekly"`
	RateLimited         bool                       `json:"rate_limited"`
	RateLimitResetAt    *time.Time                 `json:"rate_limit_reset_at,omitempty"`
	ModelRateLimitCount int                        `json:"model_rate_limit_count,omitempty"`
	State               string                     `json:"state"` // available, rate_limited, used, unknown, unavailable
}

type AccountQuotaAggregate struct {
	AccountCount       int     `json:"account_count"`
	KnownCount         int     `json:"known_count"`
	UnknownCount       int     `json:"unknown_count"`
	RateLimitedCount   int     `json:"rate_limited_count"`
	UnavailableCount   int     `json:"unavailable_count"`
	AvailablePercent   float64 `json:"available_percent"`
	RateLimitedPercent float64 `json:"rate_limited_percent"`
	UsedPercent        float64 `json:"used_percent"`
	UnknownPercent     float64 `json:"unknown_percent"`
	UnavailablePercent float64 `json:"unavailable_percent"`
}

type AccountQuotaGroup struct {
	ID      string                `json:"id"`
	Name    string                `json:"name"`
	Summary AccountQuotaAggregate `json:"summary"`
}

type AccountQuotaPlatform struct {
	Platform string                `json:"platform"`
	Summary  AccountQuotaAggregate `json:"summary"`
	Groups   []AccountQuotaGroup   `json:"groups"`
}

type AccountQuotaDashboard struct {
	GeneratedAt      time.Time              `json:"generated_at"`
	LatestObservedAt *time.Time             `json:"latest_observed_at,omitempty"`
	Platforms        []AccountQuotaPlatform `json:"platforms"`
}

// AccountQuotaDashboardService produces dashboard projections from the
// existing account repository. The repository response is never serialized
// directly, so credentials remain outside the API response.
type AccountQuotaDashboardService struct {
	accounts AccountRepository
}

type accountQuotaProjectionSource interface {
	ListQuotaDashboardAccounts(context.Context) ([]Account, error)
}

func NewAccountQuotaDashboardService(accounts AccountRepository) *AccountQuotaDashboardService {
	return &AccountQuotaDashboardService{accounts: accounts}
}

func (s *AccountQuotaDashboardService) load(ctx context.Context) ([]Account, error) {
	if s == nil || s.accounts == nil {
		return nil, fmt.Errorf("account repository is not configured")
	}
	if source, ok := s.accounts.(accountQuotaProjectionSource); ok {
		return source.ListQuotaDashboardAccounts(ctx)
	}
	// Compatibility fallback for repository implementations used by older
	// binaries/tests. No usage service or provider client is called.
	return s.accounts.ListAllWithFilters(ctx, "", "", "", "", "", "", "")
}

func (s *AccountQuotaDashboardService) GetDashboard(ctx context.Context) (*AccountQuotaDashboard, error) {
	accounts, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	platforms := make(map[string]*AccountQuotaPlatform)
	platformAccounts := make(map[string]map[string]AccountQuotaAccount)
	groupAccounts := make(map[string]map[string]map[string]AccountQuotaAccount)
	groupNames := make(map[string]map[string]string)
	var latest *time.Time

	for _, account := range accounts {
		row := projectAccountQuota(&account, now)
		if row.Weekly.ObservedAt != nil && (latest == nil || row.Weekly.ObservedAt.After(*latest)) {
			observed := *row.Weekly.ObservedAt
			latest = &observed
		}
		platform := platforms[account.Platform]
		if platform == nil {
			platform = &AccountQuotaPlatform{Platform: account.Platform}
			platforms[account.Platform] = platform
			platformAccounts[account.Platform] = make(map[string]AccountQuotaAccount)
		}
		platformAccounts[account.Platform][account.ID] = row

		if len(account.Groups) == 0 {
			addGroupAccount(groupAccounts, groupNames, account.Platform, "__unassigned__", "未分组", account.ID, row)
		} else {
			for _, group := range account.Groups {
				if group == nil {
					continue
				}
				addGroupAccount(groupAccounts, groupNames, account.Platform, group.ID, group.Name, account.ID, row)
			}
		}
	}

	for platformName, platform := range platforms {
		platform.Summary = aggregateRows(platformAccounts[platformName])
		groups := make([]AccountQuotaGroup, 0, len(groupAccounts[platformName]))
		for groupID, rows := range groupAccounts[platformName] {
			name := groupNames[platformName][groupID]
			if name == "" {
				name = groupID
			}
			groups = append(groups, AccountQuotaGroup{ID: groupID, Name: name, Summary: aggregateRows(rows)})
		}
		sort.Slice(groups, func(i, j int) bool {
			return strings.ToLower(groups[i].Name) < strings.ToLower(groups[j].Name)
		})
		platform.Groups = groups
	}

	result := &AccountQuotaDashboard{GeneratedAt: now, LatestObservedAt: latest}
	for _, platform := range platforms {
		result.Platforms = append(result.Platforms, *platform)
	}
	// Stable ordering makes the UI and cache output deterministic.
	for i := 0; i < len(result.Platforms); i++ {
		for j := i + 1; j < len(result.Platforms); j++ {
			if result.Platforms[j].Platform < result.Platforms[i].Platform {
				result.Platforms[i], result.Platforms[j] = result.Platforms[j], result.Platforms[i]
			}
		}
	}
	return result, nil
}

// ListAccounts returns sanitized account rows for an expanded platform/group
// panel. Filtering is performed after the repository's soft-delete filter and
// pagination is applied in memory because the projection depends on JSONB
// compatibility fields from several provider snapshots.
func (s *AccountQuotaDashboardService) ListAccounts(ctx context.Context, platform, groupID string, offset, limit int) ([]AccountQuotaAccount, int, error) {
	accounts, err := s.load(ctx)
	if err != nil {
		return nil, 0, err
	}
	now := time.Now()
	rows := make([]AccountQuotaAccount, 0)
	for _, account := range accounts {
		if platform != "" && account.Platform != platform {
			continue
		}
		matchesGroup := groupID == "" || (groupID == "__unassigned__" && len(account.Groups) == 0)
		if groupID != "" && groupID != "__unassigned__" {
			matchesGroup = false
			for _, group := range account.Groups {
				if group != nil && group.ID == groupID {
					matchesGroup = true
					break
				}
			}
		}
		if matchesGroup {
			rows = append(rows, projectAccountQuota(&account, now))
		}
	}
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if strings.ToLower(rows[j].Name) < strings.ToLower(rows[i].Name) {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}
	total := len(rows)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50
	}
	if offset >= total {
		return []AccountQuotaAccount{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return rows[offset:end], total, nil
}

func addGroupAccount(groups map[string]map[string]map[string]AccountQuotaAccount, names map[string]map[string]string, platform, groupID, groupName, accountID string, row AccountQuotaAccount) {
	if groups[platform] == nil {
		groups[platform] = make(map[string]map[string]AccountQuotaAccount)
	}
	if groups[platform][groupID] == nil {
		groups[platform][groupID] = make(map[string]AccountQuotaAccount)
	}
	groups[platform][groupID][accountID] = row
	if names[platform] == nil {
		names[platform] = make(map[string]string)
	}
	if names[platform][groupID] == "" {
		names[platform][groupID] = groupName
	}
}

func aggregateRows(rows map[string]AccountQuotaAccount) AccountQuotaAggregate {
	var out AccountQuotaAggregate
	out.AccountCount = len(rows)
	if out.AccountCount == 0 {
		return out
	}
	for _, row := range rows {
		if row.State == "unavailable" {
			out.UnavailableCount++
			out.UnavailablePercent += 100
			continue
		}
		if !row.Weekly.Known {
			out.UnknownCount++
			out.UnknownPercent += 100
			continue
		}
		out.KnownCount++
		out.UsedPercent += row.Weekly.UsedPercent
		switch row.State {
		case "rate_limited":
			out.RateLimitedCount++
			out.RateLimitedPercent += row.Weekly.RemainingPct
		default:
			out.AvailablePercent += row.Weekly.RemainingPct
		}
	}
	denom := float64(out.AccountCount)
	out.AvailablePercent = clampPercent(out.AvailablePercent / denom)
	out.RateLimitedPercent = clampPercent(out.RateLimitedPercent / denom)
	out.UsedPercent = clampPercent(out.UsedPercent / denom)
	out.UnknownPercent = clampPercent(out.UnknownPercent / denom)
	out.UnavailablePercent = clampPercent(out.UnavailablePercent / denom)
	return out
}

func projectAccountQuota(account *Account, now time.Time) AccountQuotaAccount {
	row := AccountQuotaAccount{ID: account.ID, Name: account.Name, Platform: account.Platform, Status: account.Status, Schedulable: account.Schedulable}
	row.Weekly = weeklyQuotaSnapshot(account, now)
	row.RateLimited = account.IsRateLimited()
	row.RateLimitResetAt = account.RateLimitResetAt
	row.ModelRateLimitCount = activeModelRateLimitCount(account, now)

	otherUnavailable := account.Status != StatusActive || (!account.Schedulable && !row.RateLimited) || account.IsOverloaded() || activeTempUnschedulable(account, now) || activeAccountExpiry(account, now)
	if otherUnavailable {
		row.State = "unavailable"
	} else if !row.Weekly.Known {
		row.State = "unknown"
	} else if row.RateLimited {
		row.State = "rate_limited"
	} else if row.Weekly.RemainingPct <= 0 {
		row.State = "used"
	} else {
		row.State = "available"
	}
	return row
}

func activeTempUnschedulable(account *Account, now time.Time) bool {
	return account != nil && account.TempUnschedulableUntil != nil && account.TempUnschedulableUntil.After(now)
}

func activeAccountExpiry(account *Account, now time.Time) bool {
	return account != nil && account.AutoPauseOnExpired && account.ExpiresAt != nil && !account.ExpiresAt.After(now)
}

func clampPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func activeModelRateLimitCount(account *Account, now time.Time) int {
	if account == nil || account.Extra == nil {
		return 0
	}
	raw, ok := account.Extra["model_rate_limits"].(map[string]any)
	if !ok {
		return 0
	}
	count := 0
	for _, value := range raw {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if reset, ok := parseQuotaTime(item["rate_limit_reset_at"]); ok && reset.After(now) {
			count++
		}
	}
	return count
}

func weeklyQuotaSnapshot(account *Account, now time.Time) AccountWeeklyQuotaSnapshot {
	if account == nil {
		return AccountWeeklyQuotaSnapshot{}
	}
	candidates := make([]AccountWeeklyQuotaSnapshot, 0, 3)
	if candidate, ok := normalizedWeeklySnapshot(account.Extra, now); ok {
		candidates = append(candidates, candidate)
	}
	if candidate, ok := legacyWeeklySnapshot(account, now); ok {
		candidates = append(candidates, candidate)
	}
	if limit := account.GetQuotaWeeklyLimit(); limit > 0 && !account.IsWeeklyQuotaPeriodExpired() {
		used := account.GetQuotaWeeklyUsed() / limit * 100
		resetAt, _ := parseQuotaTime(account.Extra["quota_weekly_reset_at"])
		if resetAt == nil || resetAt.IsZero() {
			if start, ok := parseQuotaTime(account.Extra["quota_weekly_start"]); ok {
				next := start.Add(7 * 24 * time.Hour)
				resetAt = &next
			}
		}
		observed := account.UpdatedAt
		candidates = append(candidates, AccountWeeklyQuotaSnapshot{Known: true, UsedPercent: clampPercent(used), RemainingPct: clampPercent(100 - used), ResetAt: timePtrIfNonZeroValue(resetAt), ObservedAt: &observed, Source: "local_limit"})
	}
	if len(candidates) == 0 {
		return AccountWeeklyQuotaSnapshot{}
	}
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.RemainingPct < best.RemainingPct {
			best = candidate
		}
	}
	return best
}

func normalizedWeeklySnapshot(extra map[string]any, now time.Time) (AccountWeeklyQuotaSnapshot, bool) {
	if extra == nil {
		return AccountWeeklyQuotaSnapshot{}, false
	}
	raw, ok := extra["passive_weekly_quota"].(map[string]any)
	if !ok {
		return AccountWeeklyQuotaSnapshot{}, false
	}
	used, ok := numberFromMap(raw, "used_percent")
	if !ok {
		return AccountWeeklyQuotaSnapshot{}, false
	}
	resetAt, _ := parseQuotaTime(raw["reset_at"])
	observedAt, _ := parseQuotaTime(raw["observed_at"])
	if quotaSnapshotExpired(resetAt, observedAt, now) {
		return AccountWeeklyQuotaSnapshot{Expired: true}, false
	}
	return makeQuotaSnapshot(used, resetAt, observedAt, "passive_response"), true
}

func legacyWeeklySnapshot(account *Account, now time.Time) (AccountWeeklyQuotaSnapshot, bool) {
	extra := account.Extra
	if extra == nil {
		return AccountWeeklyQuotaSnapshot{}, false
	}
	if util, ok := numberFromMap(extra, "passive_usage_7d_utilization"); ok {
		resetAt, _ := parseQuotaTime(extra["passive_usage_7d_reset"])
		observedAt, _ := parseQuotaTime(extra["passive_usage_sampled_at"])
		if !quotaSnapshotExpired(resetAt, observedAt, now) {
			return makeQuotaSnapshot(util*100, resetAt, observedAt, "anthropic_passive"), true
		}
	}
	if used, ok := numberFromMap(extra, "codex_7d_used_percent"); ok {
		resetAt, _ := parseQuotaTime(extra["codex_7d_reset_at"])
		observedAt := account.UpdatedAt
		if !quotaSnapshotExpired(resetAt, &observedAt, now) {
			return makeQuotaSnapshot(used, resetAt, &observedAt, "openai_passive"), true
		}
	}
	for _, key := range []string{"grok_billing_snapshot", "ollama_cloud_usage_snapshot"} {
		if snapshot, ok := nestedWeeklySnapshot(extra[key], now, key); ok {
			return snapshot, true
		}
	}
	return AccountWeeklyQuotaSnapshot{}, false
}

func nestedWeeklySnapshot(raw any, now time.Time, source string) (AccountWeeklyQuotaSnapshot, bool) {
	root, ok := raw.(map[string]any)
	if !ok {
		return AccountWeeklyQuotaSnapshot{}, false
	}
	if source == "ollama_cloud_usage_snapshot" {
		data, _ := root["data"].(map[string]any)
		week, _ := data["seven_day"].(map[string]any)
		used, ok := numberFromMap(week, "used_percent")
		if !ok {
			return AccountWeeklyQuotaSnapshot{}, false
		}
		resetAt, _ := parseQuotaTime(week["reset_at"])
		observedAt, _ := parseQuotaTime(root["fetched_at"])
		if quotaSnapshotExpired(resetAt, observedAt, now) {
			return AccountWeeklyQuotaSnapshot{}, false
		}
		return makeQuotaSnapshot(used, resetAt, observedAt, "ollama_passive"), true
	}
	for _, key := range []string{"usage_percent", "used_percent"} {
		if used, ok := numberFromMap(root, key); ok {
			resetAt, _ := parseQuotaTime(root["period_end"])
			observedAt, _ := parseQuotaTime(root["updated_at"])
			if quotaSnapshotExpired(resetAt, observedAt, now) {
				return AccountWeeklyQuotaSnapshot{}, false
			}
			return makeQuotaSnapshot(used, resetAt, observedAt, "grok_passive"), true
		}
	}
	return AccountWeeklyQuotaSnapshot{}, false
}

func makeQuotaSnapshot(used float64, resetAt, observedAt *time.Time, source string) AccountWeeklyQuotaSnapshot {
	used = clampPercent(used)
	return AccountWeeklyQuotaSnapshot{Known: true, UsedPercent: used, RemainingPct: clampPercent(100 - used), ResetAt: timePtrIfNonZeroValue(resetAt), ObservedAt: timePtrIfNonZeroValue(observedAt), Source: source}
}

func quotaSnapshotExpired(resetAt, observedAt *time.Time, now time.Time) bool {
	if resetAt != nil && !resetAt.IsZero() {
		return !resetAt.After(now)
	}
	return observedAt != nil && now.Sub(*observedAt) > 7*24*time.Hour
}

func numberFromMap(values map[string]any, key string) (float64, bool) {
	if values == nil {
		return 0, false
	}
	raw, ok := values[key]
	if !ok {
		return 0, false
	}
	value := parseExtraFloat64(raw)
	return value, value >= 0
}

func parseQuotaTime(raw any) (*time.Time, bool) {
	switch value := raw.(type) {
	case *time.Time:
		return value, value != nil
	case time.Time:
		return &value, true
	case float64:
		t := time.Unix(int64(value), 0)
		return &t, true
	case int64:
		t := time.Unix(value, 0)
		return &t, true
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false
		}
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			return &parsed, true
		}
	}
	return nil, false
}

func timePtrIfNonZero(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func timePtrIfNonZeroValue(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	return value
}
