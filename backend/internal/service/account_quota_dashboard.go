package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/usagestats"
)

// AccountWeeklyQuotaSnapshot is a dashboard-safe projection of one account's
// passively observed upstream seven-day quota and its matching local account
// cost. Reading this projection never contacts an upstream provider.
type AccountWeeklyQuotaSnapshot struct {
	Known                 bool       `json:"known"`
	UsedPercent           float64    `json:"used_percent"`
	RemainingPct          float64    `json:"remaining_percent"`
	UsedPoints            float64    `json:"used_points"`
	EstimatedTotalPoints  float64    `json:"estimated_total_points"`
	EstimatedRemainPoints float64    `json:"estimated_remaining_points"`
	WindowStart           *time.Time `json:"window_start,omitempty"`
	WindowEnd             *time.Time `json:"window_end,omitempty"`
	ResetAt               *time.Time `json:"reset_at,omitempty"`
	ObservedAt            *time.Time `json:"observed_at,omitempty"`
	Source                string     `json:"source,omitempty"`
	Expired               bool       `json:"expired,omitempty"`
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
	State               string                     `json:"state"` // available, rate_limited, temporarily_unavailable, unknown
}

// AccountQuotaAggregate is point-weighted. Unknown accounts are represented
// by count only because their capacity cannot be inferred without fabricating
// a numeric denominator.
type AccountQuotaAggregate struct {
	EnabledAccountCount  int     `json:"enabled_account_count"`
	KnownAccountCount    int     `json:"known_account_count"`
	UnknownAccountCount  int     `json:"unknown_account_count"`
	RateLimitedCount     int     `json:"rate_limited_count"`
	EstimatedTotalPoints float64 `json:"estimated_total_points"`
	UsedPoints           float64 `json:"used_points"`
	AvailablePoints      float64 `json:"available_points"`
	RateLimitedPoints    float64 `json:"rate_limited_points"`
	UsedPercent          float64 `json:"used_percent"`
	AvailablePercent     float64 `json:"available_percent"`
	RateLimitedPercent   float64 `json:"rate_limited_percent"`
	CoverageComplete     bool    `json:"coverage_complete"`
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

type accountQuotaProjectionSource interface {
	ListQuotaDashboardAccounts(context.Context) ([]Account, error)
}

type accountQuotaWindowReader interface {
	GetAccountWindowStatsByWindows(context.Context, []usagestats.AccountStatsWindow) (map[string]*usagestats.AccountStats, error)
}

// AccountQuotaDashboardService estimates weekly point capacity from persisted
// upstream percentages and local usage logs. It intentionally has no provider
// client dependency.
type AccountQuotaDashboardService struct {
	accounts AccountRepository
	groups   GroupRepository
	usage    UsageLogRepository
}

// NewAccountQuotaDashboardService remains as a compatibility constructor for
// older unit-test fixtures. Production wiring uses the provider below.
func NewAccountQuotaDashboardService(accounts AccountRepository) *AccountQuotaDashboardService {
	return &AccountQuotaDashboardService{accounts: accounts}
}

func ProvideAccountQuotaDashboardService(accounts AccountRepository, groups GroupRepository, usage UsageLogRepository) *AccountQuotaDashboardService {
	return &AccountQuotaDashboardService{accounts: accounts, groups: groups, usage: usage}
}

func (s *AccountQuotaDashboardService) loadAccounts(ctx context.Context) ([]Account, error) {
	if s == nil || s.accounts == nil {
		return nil, fmt.Errorf("account repository is not configured")
	}
	if source, ok := s.accounts.(accountQuotaProjectionSource); ok {
		return source.ListQuotaDashboardAccounts(ctx)
	}
	return s.accounts.ListAllWithFilters(ctx, "", "", "", "", "", "", "")
}

func (s *AccountQuotaDashboardService) loadActiveGroups(ctx context.Context, accounts []Account) ([]Group, error) {
	if s.groups != nil {
		return s.groups.ListActive(ctx)
	}
	// Compatibility fallback for old fixtures. Production always uses the
	// group repository so active empty groups are retained.
	seen := make(map[string]Group)
	for i := range accounts {
		for _, group := range accounts[i].Groups {
			if group != nil && group.Status == StatusActive {
				seen[group.ID] = *group
			}
		}
	}
	groups := make([]Group, 0, len(seen))
	for _, group := range seen {
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].SortOrder == groups[j].SortOrder {
			return groups[i].ID < groups[j].ID
		}
		return groups[i].SortOrder < groups[j].SortOrder
	})
	return groups, nil
}

func (s *AccountQuotaDashboardService) GetDashboard(ctx context.Context) (*AccountQuotaDashboard, error) {
	accounts, err := s.loadAccounts(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := s.loadActiveGroups(ctx, accounts)
	if err != nil {
		return nil, err
	}

	activeGroups := make(map[string]Group, len(groups))
	for i := range groups {
		activeGroups[groups[i].ID] = groups[i]
	}
	enabledAccounts := filterQuotaAccounts(accounts, activeGroups, "")
	rows, latest, err := s.projectAccounts(ctx, enabledAccounts, time.Now())
	if err != nil {
		return nil, err
	}

	rowsByID := make(map[string]AccountQuotaAccount, len(rows))
	accountsByID := make(map[string]Account, len(enabledAccounts))
	for i := range rows {
		rowsByID[rows[i].ID] = rows[i]
	}
	for i := range enabledAccounts {
		accountsByID[enabledAccounts[i].ID] = enabledAccounts[i]
	}

	platformGroups := make(map[string][]AccountQuotaGroup)
	platformRows := make(map[string]map[string]AccountQuotaAccount)
	for i := range groups {
		group := groups[i]
		groupRowMap := make(map[string]AccountQuotaAccount)
		for accountID, account := range accountsByID {
			if accountBelongsToGroup(&account, group.ID) {
				groupRowMap[accountID] = rowsByID[accountID]
				if platformRows[group.Platform] == nil {
					platformRows[group.Platform] = make(map[string]AccountQuotaAccount)
				}
				platformRows[group.Platform][accountID] = rowsByID[accountID]
			}
		}
		platformGroups[group.Platform] = append(platformGroups[group.Platform], AccountQuotaGroup{
			ID:      group.ID,
			Name:    group.Name,
			Summary: aggregateQuotaRows(groupRowMap),
		})
		// Keep a platform card for active empty groups.
		if platformRows[group.Platform] == nil {
			platformRows[group.Platform] = make(map[string]AccountQuotaAccount)
		}
	}

	result := &AccountQuotaDashboard{GeneratedAt: time.Now(), LatestObservedAt: latest}
	platformNames := make([]string, 0, len(platformGroups))
	for platform := range platformGroups {
		platformNames = append(platformNames, platform)
	}
	sort.Strings(platformNames)
	for _, platform := range platformNames {
		result.Platforms = append(result.Platforms, AccountQuotaPlatform{
			Platform: platform,
			Summary:  aggregateQuotaRows(platformRows[platform]),
			Groups:   platformGroups[platform],
		})
	}
	return result, nil
}

func (s *AccountQuotaDashboardService) ListAccounts(ctx context.Context, platform, groupID string, offset, limit int) ([]AccountQuotaAccount, int, error) {
	accounts, err := s.loadAccounts(ctx)
	if err != nil {
		return nil, 0, err
	}
	groups, err := s.loadActiveGroups(ctx, accounts)
	if err != nil {
		return nil, 0, err
	}
	activeGroups := make(map[string]Group, len(groups))
	for i := range groups {
		activeGroups[groups[i].ID] = groups[i]
	}
	group, exists := activeGroups[groupID]
	if groupID == "" || !exists || (platform != "" && group.Platform != platform) {
		return []AccountQuotaAccount{}, 0, nil
	}

	filtered := filterQuotaAccounts(accounts, activeGroups, groupID)
	rows, _, err := s.projectAccounts(ctx, filtered, time.Now())
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name)
	})
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

func filterQuotaAccounts(accounts []Account, activeGroups map[string]Group, groupID string) []Account {
	result := make([]Account, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		if account.Status != StatusActive || !account.Schedulable {
			continue
		}
		matched := false
		for _, group := range account.Groups {
			if group == nil {
				continue
			}
			if _, active := activeGroups[group.ID]; !active {
				continue
			}
			if groupID == "" || group.ID == groupID {
				matched = true
				break
			}
		}
		if matched {
			result = append(result, account)
		}
	}
	return result
}

func accountBelongsToGroup(account *Account, groupID string) bool {
	if account == nil {
		return false
	}
	for _, group := range account.Groups {
		if group != nil && group.ID == groupID {
			return true
		}
	}
	return false
}

func (s *AccountQuotaDashboardService) projectAccounts(ctx context.Context, accounts []Account, now time.Time) ([]AccountQuotaAccount, *time.Time, error) {
	snapshots := make(map[string]AccountWeeklyQuotaSnapshot, len(accounts))
	windows := make([]usagestats.AccountStatsWindow, 0, len(accounts))
	var latest *time.Time
	for i := range accounts {
		snapshot := passiveWeeklyQuotaSnapshot(&accounts[i], now)
		if snapshot.ObservedAt != nil && (latest == nil || snapshot.ObservedAt.After(*latest)) {
			observed := *snapshot.ObservedAt
			latest = &observed
		}
		if window, ok := quotaStatsWindow(accounts[i].ID, &snapshot); ok {
			snapshot.WindowStart = &window.StartTime
			snapshot.WindowEnd = &window.EndTime
			windows = append(windows, window)
		}
		snapshots[accounts[i].ID] = snapshot
	}

	stats := make(map[string]*usagestats.AccountStats)
	if len(windows) > 0 {
		reader, ok := s.usage.(accountQuotaWindowReader)
		if !ok {
			return nil, nil, fmt.Errorf("usage repository does not support independent quota windows")
		}
		var err error
		stats, err = reader.GetAccountWindowStatsByWindows(ctx, windows)
		if err != nil {
			return nil, nil, err
		}
	}

	rows := make([]AccountQuotaAccount, 0, len(accounts))
	for i := range accounts {
		snapshot := snapshots[accounts[i].ID]
		applyPointEstimate(&snapshot, stats[accounts[i].ID])
		rows = append(rows, projectAccountQuota(&accounts[i], snapshot, now))
	}
	return rows, latest, nil
}

func quotaStatsWindow(accountID string, snapshot *AccountWeeklyQuotaSnapshot) (usagestats.AccountStatsWindow, bool) {
	if snapshot == nil || !snapshot.Known || snapshot.UsedPercent <= 0 || snapshot.ObservedAt == nil || snapshot.ObservedAt.IsZero() {
		return usagestats.AccountStatsWindow{}, false
	}
	end := *snapshot.ObservedAt
	start := end.Add(-7 * 24 * time.Hour)
	if snapshot.ResetAt != nil && !snapshot.ResetAt.IsZero() {
		start = snapshot.ResetAt.Add(-7 * 24 * time.Hour)
	}
	if !start.Before(end) {
		return usagestats.AccountStatsWindow{}, false
	}
	return usagestats.AccountStatsWindow{AccountID: accountID, StartTime: start, EndTime: end}, true
}

func applyPointEstimate(snapshot *AccountWeeklyQuotaSnapshot, stats *usagestats.AccountStats) {
	if snapshot == nil || !snapshot.Known || snapshot.UsedPercent <= 0 || stats == nil || stats.Cost <= 0 {
		if snapshot != nil {
			snapshot.Known = false
		}
		return
	}
	usedPercent := clampPercent(snapshot.UsedPercent)
	total := stats.Cost / (usedPercent / 100)
	if math.IsNaN(total) || math.IsInf(total, 0) || total <= 0 {
		snapshot.Known = false
		return
	}
	snapshot.UsedPoints = stats.Cost
	snapshot.EstimatedTotalPoints = total
	snapshot.EstimatedRemainPoints = math.Max(0, total-stats.Cost)
	snapshot.RemainingPct = clampPercent(100 - usedPercent)
}

func aggregateQuotaRows(rows map[string]AccountQuotaAccount) AccountQuotaAggregate {
	out := AccountQuotaAggregate{EnabledAccountCount: len(rows)}
	for _, row := range rows {
		if row.RateLimited {
			out.RateLimitedCount++
		}
		if !row.Weekly.Known {
			out.UnknownAccountCount++
			continue
		}
		out.KnownAccountCount++
		out.EstimatedTotalPoints += row.Weekly.EstimatedTotalPoints
		out.UsedPoints += row.Weekly.UsedPoints
		if row.RateLimited {
			out.RateLimitedPoints += row.Weekly.EstimatedRemainPoints
		} else {
			out.AvailablePoints += row.Weekly.EstimatedRemainPoints
		}
	}
	out.CoverageComplete = out.EnabledAccountCount > 0 && out.UnknownAccountCount == 0
	if out.EstimatedTotalPoints > 0 {
		out.UsedPercent = clampPercent(out.UsedPoints / out.EstimatedTotalPoints * 100)
		out.AvailablePercent = clampPercent(out.AvailablePoints / out.EstimatedTotalPoints * 100)
		out.RateLimitedPercent = clampPercent(out.RateLimitedPoints / out.EstimatedTotalPoints * 100)
	}
	return out
}

func projectAccountQuota(account *Account, snapshot AccountWeeklyQuotaSnapshot, now time.Time) AccountQuotaAccount {
	row := AccountQuotaAccount{
		ID:          account.ID,
		Name:        account.Name,
		Platform:    account.Platform,
		Status:      account.Status,
		Schedulable: account.Schedulable,
		Weekly:      snapshot,
	}
	row.RateLimited = account.IsRateLimited()
	row.RateLimitResetAt = account.RateLimitResetAt
	row.ModelRateLimitCount = activeModelRateLimitCount(account, now)
	if !row.Weekly.Known {
		row.State = "unknown"
	} else if row.RateLimited {
		row.State = "rate_limited"
	} else if account.IsOverloaded() || activeTempUnschedulable(account, now) || activeAccountExpiry(account, now) {
		row.State = "temporarily_unavailable"
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

// passiveWeeklyQuotaSnapshot intentionally excludes configured local weekly
// limits. The dashboard estimate is valid only when the denominator is a real
// passively observed upstream seven-day percentage.
func passiveWeeklyQuotaSnapshot(account *Account, now time.Time) AccountWeeklyQuotaSnapshot {
	if account == nil {
		return AccountWeeklyQuotaSnapshot{}
	}
	if candidate, ok := normalizedWeeklySnapshot(account.Extra, now); ok {
		return candidate
	}
	if candidate, ok := legacyWeeklySnapshot(account, now); ok {
		return candidate
	}
	return AccountWeeklyQuotaSnapshot{}
}

func normalizedWeeklySnapshot(extra map[string]any, now time.Time) (AccountWeeklyQuotaSnapshot, bool) {
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
	source, _ := raw["source"].(string)
	if source == "" {
		source = "passive_response"
	}
	return makeQuotaSnapshot(used, resetAt, observedAt, source), true
}

func legacyWeeklySnapshot(account *Account, now time.Time) (AccountWeeklyQuotaSnapshot, bool) {
	extra := account.Extra
	if extra == nil {
		return AccountWeeklyQuotaSnapshot{}, false
	}
	if util, ok := numberFromMap(extra, "passive_usage_7d_utilization"); ok {
		resetAt, _ := parseQuotaTime(extra["passive_usage_7d_reset"])
		observedAt, _ := parseQuotaTime(extra["passive_usage_sampled_at"])
		if observedAt != nil && !quotaSnapshotExpired(resetAt, observedAt, now) {
			return makeQuotaSnapshot(util*100, resetAt, observedAt, "anthropic_passive"), true
		}
	}
	if used, ok := numberFromMap(extra, "codex_7d_used_percent"); ok {
		resetAt, _ := parseQuotaTime(extra["codex_7d_reset_at"])
		observedAt, _ := parseQuotaTime(extra["codex_usage_updated_at"])
		if observedAt == nil && !account.UpdatedAt.IsZero() {
			observed := account.UpdatedAt
			observedAt = &observed
		}
		if observedAt != nil && !quotaSnapshotExpired(resetAt, observedAt, now) {
			return makeQuotaSnapshot(used, resetAt, observedAt, "openai_passive"), true
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
		if observedAt == nil || quotaSnapshotExpired(resetAt, observedAt, now) {
			return AccountWeeklyQuotaSnapshot{}, false
		}
		return makeQuotaSnapshot(used, resetAt, observedAt, "ollama_passive"), true
	}
	for _, key := range []string{"usage_percent", "used_percent"} {
		if used, ok := numberFromMap(root, key); ok {
			resetAt, _ := parseQuotaTime(root["period_end"])
			observedAt, _ := parseQuotaTime(root["updated_at"])
			if observedAt == nil || quotaSnapshotExpired(resetAt, observedAt, now) {
				return AccountWeeklyQuotaSnapshot{}, false
			}
			return makeQuotaSnapshot(used, resetAt, observedAt, "grok_passive"), true
		}
	}
	return AccountWeeklyQuotaSnapshot{}, false
}

func makeQuotaSnapshot(used float64, resetAt, observedAt *time.Time, source string) AccountWeeklyQuotaSnapshot {
	used = clampPercent(used)
	return AccountWeeklyQuotaSnapshot{
		Known:        true,
		UsedPercent:  used,
		RemainingPct: clampPercent(100 - used),
		ResetAt:      timePtrIfNonZeroValue(resetAt),
		ObservedAt:   timePtrIfNonZeroValue(observedAt),
		Source:       source,
	}
}

func quotaSnapshotExpired(resetAt, observedAt *time.Time, now time.Time) bool {
	if resetAt != nil && !resetAt.IsZero() {
		return !resetAt.After(now)
	}
	return observedAt == nil || now.Sub(*observedAt) > 7*24*time.Hour
}

func numberFromMap(values map[string]any, key string) (float64, bool) {
	if values == nil {
		return 0, false
	}
	raw, ok := values[key]
	if !ok || raw == nil {
		return 0, false
	}
	var value float64
	switch typed := raw.(type) {
	case float64:
		value = typed
	case float32:
		value = float64(typed)
	case int:
		value = float64(typed)
	case int64:
		value = float64(typed)
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		value = parsed
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, false
		}
		value = parsed
	default:
		return 0, false
	}
	return value, !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
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

func timePtrIfNonZeroValue(value *time.Time) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	return value
}

// NormalizePassiveWeeklyQuotaUpdate adds the canonical passive snapshot when
// a platform-specific UpdateExtra call contains a reliable seven-day quota.
// It does not clear the last good snapshot when the current response carries
// no quota data.
func NormalizePassiveWeeklyQuotaUpdate(updates map[string]any, observedAt time.Time) map[string]any {
	if len(updates) == 0 {
		return updates
	}
	result := make(map[string]any, len(updates)+1)
	for key, value := range updates {
		result[key] = value
	}
	if raw, ok := result["passive_weekly_quota"].(map[string]any); ok {
		if _, present := raw["observed_at"]; !present {
			copyRaw := make(map[string]any, len(raw)+1)
			for key, value := range raw {
				copyRaw[key] = value
			}
			copyRaw["observed_at"] = observedAt
			result["passive_weekly_quota"] = copyRaw
		}
		return result
	}

	type candidate struct {
		used     float64
		reset    *time.Time
		observed *time.Time
		source   string
	}
	var selected *candidate
	if util, ok := numberFromMap(result, "passive_usage_7d_utilization"); ok {
		reset, _ := parseQuotaTime(result["passive_usage_7d_reset"])
		observed, _ := parseQuotaTime(result["passive_usage_sampled_at"])
		selected = &candidate{used: util * 100, reset: reset, observed: observed, source: "anthropic_passive"}
	} else if used, ok := numberFromMap(result, "codex_7d_used_percent"); ok {
		reset, _ := parseQuotaTime(result["codex_7d_reset_at"])
		observed, _ := parseQuotaTime(result["codex_usage_updated_at"])
		selected = &candidate{used: used, reset: reset, observed: observed, source: "openai_passive"}
	} else if snapshot, ok := nestedWeeklySnapshot(result["grok_billing_snapshot"], observedAt, "grok_billing_snapshot"); ok {
		selected = &candidate{used: snapshot.UsedPercent, reset: snapshot.ResetAt, observed: snapshot.ObservedAt, source: snapshot.Source}
	} else if snapshot, ok := nestedWeeklySnapshot(result["ollama_cloud_usage_snapshot"], observedAt, "ollama_cloud_usage_snapshot"); ok {
		selected = &candidate{used: snapshot.UsedPercent, reset: snapshot.ResetAt, observed: snapshot.ObservedAt, source: snapshot.Source}
	}
	if selected == nil {
		return result
	}
	if selected.observed == nil || selected.observed.IsZero() {
		selected.observed = &observedAt
	}
	canonical := map[string]any{
		"used_percent": clampPercent(selected.used),
		"observed_at":  selected.observed,
		"source":       selected.source,
	}
	if selected.reset != nil && !selected.reset.IsZero() {
		canonical["reset_at"] = selected.reset
	}
	result["passive_weekly_quota"] = canonical
	return result
}
