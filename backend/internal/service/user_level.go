package service

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	gocache "github.com/patrickmn/go-cache"
)

const (
	// Kept as a source-compatibility marker for old settings rows. Runtime
	// calculation never reads settings.user_level_settings anymore.
	SettingKeyUserLevelSettings = "user_level_settings"
	userLevelProfileCacheTTL    = time.Minute
)

// UserLevelSettings is retained only for source compatibility with older
// clients. Fixed L1/L2/L3 thresholds are no longer used by the service.
type UserLevelSettings struct {
	L2MinSpend  float64 `json:"l2_min_spend"`
	L3MinSpend  float64 `json:"l3_min_spend"`
	WindowHours int     `json:"window_hours"`
}

func DefaultUserLevelSettings() UserLevelSettings {
	return UserLevelSettings{WindowHours: 168}
}

func ValidateUserLevelSettings(settings UserLevelSettings) (UserLevelSettings, error) {
	if math.IsNaN(settings.L2MinSpend) || math.IsInf(settings.L2MinSpend, 0) || settings.L2MinSpend < 0 {
		return settings, infraerrors.BadRequest("USER_LEVEL_L2_INVALID", "L2 minimum spend must be nonnegative")
	}
	if math.IsNaN(settings.L3MinSpend) || math.IsInf(settings.L3MinSpend, 0) || settings.L3MinSpend < 0 {
		return settings, infraerrors.BadRequest("USER_LEVEL_L3_INVALID", "L3 minimum spend must be nonnegative")
	}
	settings.L2MinSpend = QuantizeUsageBillingAmount(settings.L2MinSpend)
	settings.L3MinSpend = QuantizeUsageBillingAmount(settings.L3MinSpend)
	settings.WindowHours = 168
	return settings, nil
}

type UserLevelProfile struct {
	UserID string `json:"user_id"`
	// Level is a compatibility ordinal. It is zero when the user has no
	// assigned rule; tier UUIDs in Rules/CurrentTierIDs are authoritative.
	Level               int                    `json:"level"`
	Configured          bool                   `json:"configured"`
	Usage7d             float64                `json:"usage_7d"`
	WindowFrom          time.Time              `json:"window_from"`
	CalculatedAt        time.Time              `json:"calculated_at"`
	Rules               []UserLevelRuleProfile `json:"rules"`
	CurrentTierIDs      []string               `json:"current_tier_ids"`
	UserLevelMultiplier *float64               `json:"user_level_multiplier,omitempty"`
}

// UserLevelDashboard is the user-safe summary exposed by the personal
// dashboard. Legacy fields remain in the JSON shape for old clients, but no
// fixed L1/L2/L3 thresholds are returned or used for calculation.
type UserLevelDashboard struct {
	UserID              string                 `json:"user_id"`
	Level               int                    `json:"level"`
	Configured          bool                   `json:"configured"`
	Usage7d             float64                `json:"usage_7d"`
	WindowHours         int                    `json:"window_hours"`
	WindowFrom          time.Time              `json:"window_from"`
	CalculatedAt        time.Time              `json:"calculated_at"`
	Rules               []UserLevelRuleProfile `json:"rules"`
	CurrentTierIDs      []string               `json:"current_tier_ids"`
	UserLevelMultiplier *float64               `json:"user_level_multiplier,omitempty"`
	GroupRuleMultiplier *float64               `json:"group_rule_multiplier,omitempty"`
	EffectiveSource     string                 `json:"effective_source,omitempty"`
	LevelMultiplier     *float64               `json:"level_multiplier,omitempty"`
	EffectiveMultiplier *float64               `json:"effective_multiplier,omitempty"`
	MultiplierGroup     string                 `json:"multiplier_group,omitempty"`
	NextLevelMultiplier *float64               `json:"next_level_multiplier,omitempty"`
	NextMultiplierGroup string                 `json:"next_multiplier_group,omitempty"`
	// Deprecated fixed-setting fields. They are always zero.
	L2MinSpend float64 `json:"l2_min_spend"`
	L3MinSpend float64 `json:"l3_min_spend"`
}

type DynamicRateUsageKey struct {
	RuleID   string
	QuotaKey string
}

// UserLevelRepository is the minimal rolling-spend/quota contract used by the
// gateway. Rule CRUD lives in UserLevelRulesRepository so existing focused
// implementations do not need to grow this interface.
type UserLevelRepository interface {
	GetRollingSpend(ctx context.Context, userID string, since, until time.Time) (float64, error)
	GetRollingSpendBatch(ctx context.Context, userIDs []string, since, until time.Time) (map[string]float64, error)
	GetDynamicRateUsage(ctx context.Context, userID, groupID string, keys []DynamicRateUsageKey) (map[DynamicRateUsageKey]float64, error)
	GetSharedDynamicRateUsage(ctx context.Context, groupID string, keys []DynamicRateUsageKey) (map[DynamicRateUsageKey]float64, error)
}

type DynamicRateCandidate struct {
	RuleID              string  `json:"rule_id"`
	RuleName            string  `json:"rule_name"`
	StartAt             string  `json:"start_at"`
	EndAt               string  `json:"end_at"`
	QuotaKey            string  `json:"quota_key"`
	Multiplier          float64 `json:"-"`
	DiscountCoefficient float64 `json:"discount_coefficient"`
	SharedQuotaAmount   float64 `json:"shared_quota_amount"`
	SharedUsedAmount    float64 `json:"shared_used_amount"`
	PersonalQuotaAmount float64 `json:"personal_quota_amount"`
	PersonalUsedAmount  float64 `json:"personal_used_amount"`
}

const (
	DynamicRateStatusLegacy     = "legacy"
	DynamicRateStatusNotStarted = "not_started"
	DynamicRateStatusActive     = "active"
	DynamicRateStatusExpired    = "expired"
	DynamicRateStatusInvalid    = "invalid"
)

type DynamicRateUsageSummary struct {
	RuleID   string `json:"rule_id"`
	RuleName string `json:"rule_name"`
	StartAt  string `json:"start_at"`
	EndAt    string `json:"end_at"`
	Status   string `json:"status"`
	DiscountCoefficient float64 `json:"discount_coefficient"`
	// Shared fields are retained for old response consumers. Live selection is
	// per-user and never reads the group-wide counter.
	SharedQuotaAmount     float64  `json:"shared_quota_amount"`
	SharedUsedAmount      float64  `json:"shared_used_amount"`
	SharedRemainingAmount *float64 `json:"shared_remaining_amount"`
	PersonalQuotaAmount   float64  `json:"personal_quota_amount"`
	UsageScope            string   `json:"usage_scope"`
}

// UserRatePlan is the one pricing snapshot shared by scheduling and billing.
// User-level and group-side candidates are kept separately so diagnostics can
// explain why the final user-side multiplier was selected.
type UserRatePlan struct {
	GroupID   string  `json:"group_id"`
	UserLevel int     `json:"user_level"`
	Usage7d   float64 `json:"usage_7d"`
	// BaseMultiplier is retained for old consumers and now represents the
	// final pre-peak user-side multiplier.
	BaseMultiplier          float64                `json:"base_multiplier"`
	RateMultiplier          float64                `json:"rate_multiplier"`
	PeakMultiplier          float64                `json:"peak_multiplier"`
	EffectiveMultiplier     float64                `json:"effective_multiplier"`
	Source                  string                 `json:"source"`
	DynamicCandidates       []DynamicRateCandidate `json:"dynamic_candidates"`
	SelectedDynamicRuleID   string                 `json:"selected_dynamic_rule_id,omitempty"`
	UserLevelMultiplier     *float64               `json:"user_level_multiplier,omitempty"`
	UserRateMultiplier      *float64               `json:"user_rate_multiplier,omitempty"`
	GroupRuleMultiplier     *float64               `json:"group_rule_multiplier,omitempty"`
	EffectiveBaseMultiplier float64                `json:"effective_base_multiplier"`
	EffectiveSource         string                 `json:"effective_source"`
	// NonDynamicMultiplier is the baseline a dynamic rule must beat. It keeps
	// a more expensive dynamic rule from consuming quota when a user-level or
	// static group candidate is already cheaper.
	NonDynamicMultiplier float64 `json:"non_dynamic_multiplier"`
}

type RankedUserGroup struct {
	Group        *Group
	Subscription *UserSubscription
	Plan         UserRatePlan
}

type UserLevelService struct {
	repo UserLevelRepository
	// settingRepo is retained only to preserve the constructor used by Wire and
	// external integrations. It is deliberately never read.
	settingRepo   SettingRepository
	groupRepo     GroupRepository
	userRateRepo  UserGroupRateRepository
	subRepo       UserSubscriptionRepository
	rulesRepo     UserLevelRulesRepository
	profileCache  *gocache.Cache
	cacheRevision atomic.Uint64
}

func NewUserLevelService(repo UserLevelRepository, settingRepo SettingRepository, groupRepo GroupRepository, userRateRepo UserGroupRateRepository, subRepo UserSubscriptionRepository) *UserLevelService {
	s := &UserLevelService{
		repo: repo, settingRepo: settingRepo, groupRepo: groupRepo, userRateRepo: userRateRepo, subRepo: subRepo,
		profileCache: gocache.New(userLevelProfileCacheTTL, time.Minute),
	}
	if typed, ok := repo.(UserLevelRulesRepository); ok {
		s.rulesRepo = typed
	}
	return s
}

// GetSettings is a compatibility read. The old settings row is intentionally
// ignored so it cannot affect new rule calculations.
func (s *UserLevelService) GetSettings(context.Context) (UserLevelSettings, error) {
	return DefaultUserLevelSettings(), nil
}

// UpdateSettings is retained for old source consumers, but fixed thresholds
// are no longer mutable through the service.
func (s *UserLevelService) UpdateSettings(context.Context, UserLevelSettings) (UserLevelSettings, error) {
	return DefaultUserLevelSettings(), errors.New("fixed user level settings have been removed; use level rules")
}

func (s *UserLevelService) invalidateProfiles() {
	if s == nil {
		return
	}
	s.cacheRevision.Add(1)
	if s.profileCache != nil {
		s.profileCache.Flush()
	}
}

func (s *UserLevelService) profileCacheKey(userID string, at time.Time) string {
	return strings.TrimSpace(userID) + ":" + timeFormatInt64(at.UTC().Truncate(time.Minute).Unix()) + ":" + timeFormatUint64(s.cacheRevision.Load())
}

func (s *UserLevelService) ResolveProfile(ctx context.Context, userID string, at time.Time) (UserLevelProfile, error) {
	if at.IsZero() {
		at = time.Now()
	}
	profile := UserLevelProfile{
		UserID: userID, WindowFrom: at.Add(-7 * 24 * time.Hour), CalculatedAt: at,
		Rules: []UserLevelRuleProfile{}, CurrentTierIDs: []string{},
	}
	if s == nil || strings.TrimSpace(userID) == "" || s.rulesRepo == nil {
		return profile, nil
	}
	key := s.profileCacheKey(userID, at)
	if s.profileCache != nil {
		if cached, ok := s.profileCache.Get(key); ok {
			if value, valid := cached.(UserLevelProfile); valid {
				return cloneUserLevelProfile(value), nil
			}
		}
	}
	rulesByUser, err := s.rulesRepo.GetAssignedLevelRulesBatch(ctx, []string{userID})
	if err != nil {
		return profile, err
	}
	rules := rulesByUser[userID]
	spends, err := s.loadRuleSpends(ctx, []string{userID}, rules, at)
	if err != nil {
		return profile, err
	}
	profile = buildUserLevelProfile(userID, rules, spends[userID], at)
	if s.profileCache != nil {
		s.profileCache.Set(key, cloneUserLevelProfile(profile), userLevelProfileCacheTTL)
	}
	return profile, nil
}

func (s *UserLevelService) GetProfiles(ctx context.Context, userIDs []string, at time.Time) (map[string]UserLevelProfile, error) {
	if at.IsZero() {
		at = time.Now()
	}
	out := make(map[string]UserLevelProfile, len(userIDs))
	unique := uniqueStrings(userIDs)
	for _, id := range unique {
		out[id] = UserLevelProfile{UserID: id, WindowFrom: at.Add(-7 * 24 * time.Hour), CalculatedAt: at, Rules: []UserLevelRuleProfile{}, CurrentTierIDs: []string{}}
	}
	if len(unique) == 0 || s == nil || s.rulesRepo == nil {
		return out, nil
	}
	rulesByUser, err := s.rulesRepo.GetAssignedLevelRulesBatch(ctx, unique)
	if err != nil {
		return nil, err
	}
	spends, err := s.loadRuleSpends(ctx, unique, flattenAssignedRules(rulesByUser), at)
	if err != nil {
		return nil, err
	}
	for _, id := range unique {
		profile := buildUserLevelProfile(id, rulesByUser[id], spends[id], at)
		out[id] = profile
		if s.profileCache != nil {
			s.profileCache.Set(s.profileCacheKey(id, at), cloneUserLevelProfile(profile), userLevelProfileCacheTTL)
		}
	}
	return out, nil
}

// loadRuleSpends performs at most one aggregate query per supported window.
// The returned map is user -> window days -> spend.
func (s *UserLevelService) loadRuleSpends(ctx context.Context, userIDs []string, rules []UserLevelRule, at time.Time) (map[string]map[int]float64, error) {
	out := make(map[string]map[int]float64, len(userIDs))
	for _, id := range userIDs {
		out[id] = make(map[int]float64)
	}
	if s == nil || s.repo == nil || len(userIDs) == 0 {
		return out, nil
	}
	// Dynamic group rules use the rolling seven-day spend threshold even when a
	// user has no assigned level rule. Keep that baseline independent from the
	// set of configured user-level windows.
	windows := map[int]struct{}{7: struct{}{}}
	for _, rule := range rules {
		if rule.WindowDays == 7 || rule.WindowDays == 14 || rule.WindowDays == 30 {
			windows[rule.WindowDays] = struct{}{}
		}
	}
	for days := range windows {
		spendByUser, err := s.repo.GetRollingSpendBatch(ctx, userIDs, at.AddDate(0, 0, -days), at)
		if err != nil {
			return nil, err
		}
		for _, id := range userIDs {
			out[id][days] = QuantizeUsageBillingAmount(spendByUser[id])
		}
	}
	return out, nil
}

func buildUserLevelProfile(userID string, rules []UserLevelRule, spends map[int]float64, at time.Time) UserLevelProfile {
	profile := UserLevelProfile{
		UserID: userID, CalculatedAt: at, Usage7d: QuantizeUsageBillingAmount(spends[7]),
		WindowFrom: at.Add(-7 * 24 * time.Hour), Rules: make([]UserLevelRuleProfile, 0, len(rules)), CurrentTierIDs: []string{},
	}
	if len(rules) == 0 {
		return profile
	}
	var lowest *float64
	for _, rule := range rules {
		if rule.WindowDays != 7 && rule.WindowDays != 14 && rule.WindowDays != 30 {
			continue
		}
		tiers := append([]UserLevelTier(nil), rule.Tiers...)
		sort.SliceStable(tiers, func(i, j int) bool { return tiers[i].SortOrder < tiers[j].SortOrder })
		spend := QuantizeUsageBillingAmount(spends[rule.WindowDays])
		current := chooseUserLevelTier(tiers, spend)
		entry := UserLevelRuleProfile{
			RuleID: rule.ID, RuleName: rule.Name, WindowDays: rule.WindowDays, Enabled: rule.Enabled,
			Spend: spend, WindowFrom: at.AddDate(0, 0, -rule.WindowDays), CalculatedAt: at,
			Tiers: tiers,
		}
		if current != nil {
			if rule.Enabled {
				profile.Configured = true
				entry.CurrentTierID = current.ID
				entry.CurrentTierName = current.Name
				entry.CurrentTierOrder = current.SortOrder
				entry.MinSpend = current.MinSpend
				entry.DefaultMultiplier = cloneFloatPtr(current.DefaultMultiplier)
				profile.CurrentTierIDs = append(profile.CurrentTierIDs, current.ID)
				if current.DefaultMultiplier != nil && (lowest == nil || *current.DefaultMultiplier < *lowest) {
					value := *current.DefaultMultiplier
					lowest = &value
				}
				if current.SortOrder+1 > profile.Level {
					profile.Level = current.SortOrder + 1
				}
			}
		}
		if rule.WindowDays == 7 {
			profile.Usage7d = spend
			profile.WindowFrom = entry.WindowFrom
		}
		profile.Rules = append(profile.Rules, entry)
	}
	if len(profile.Rules) == 0 {
		profile.Configured = false
	}
	profile.UserLevelMultiplier = lowest
	profile.CurrentTierIDs = uniqueStrings(profile.CurrentTierIDs)
	return profile
}

func chooseUserLevelTier(tiers []UserLevelTier, spend float64) *UserLevelTier {
	var current *UserLevelTier
	for i := range tiers {
		if tiers[i].SortOrder == 0 || spend >= tiers[i].MinSpend {
			if current == nil || tiers[i].MinSpend >= current.MinSpend {
				candidate := tiers[i]
				current = &candidate
			}
		}
	}
	return current
}

func cloneFloatPtr(value *float64) *float64 {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func cloneUserLevelProfile(value UserLevelProfile) UserLevelProfile {
	value.CurrentTierIDs = append([]string(nil), value.CurrentTierIDs...)
	value.Rules = append([]UserLevelRuleProfile(nil), value.Rules...)
	for i := range value.Rules {
		value.Rules[i].DefaultMultiplier = cloneFloatPtr(value.Rules[i].DefaultMultiplier)
		value.Rules[i].Tiers = append([]UserLevelTier(nil), value.Rules[i].Tiers...)
		for j := range value.Rules[i].Tiers {
			value.Rules[i].Tiers[j].DefaultMultiplier = cloneFloatPtr(value.Rules[i].Tiers[j].DefaultMultiplier)
		}
	}
	value.UserLevelMultiplier = cloneFloatPtr(value.UserLevelMultiplier)
	return value
}

func flattenAssignedRules(byUser map[string][]UserLevelRule) []UserLevelRule {
	seen := make(map[string]struct{})
	out := make([]UserLevelRule, 0)
	for _, rules := range byUser {
		for _, rule := range rules {
			key := rule.ID + ":" + formatUnsignedDecimal(uint64(rule.WindowDays))
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, rule)
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// ResolveDashboard resolves the user's configured rules and the lowest final
// rate across authorized groups.
func (s *UserLevelService) ResolveDashboard(ctx context.Context, userID string, groupIDs []string, at time.Time) (UserLevelDashboard, error) {
	if at.IsZero() {
		at = time.Now()
	}
	profile, err := s.ResolveProfile(ctx, userID, at)
	if err != nil {
		return UserLevelDashboard{}, err
	}
	out := UserLevelDashboard{
		UserID: profile.UserID, Level: profile.Level, Configured: profile.Configured,
		Usage7d: profile.Usage7d, WindowHours: 168, WindowFrom: profile.WindowFrom, CalculatedAt: profile.CalculatedAt,
		Rules: cloneUserLevelProfile(profile).Rules, CurrentTierIDs: append([]string(nil), profile.CurrentTierIDs...),
		UserLevelMultiplier: cloneFloatPtr(profile.UserLevelMultiplier),
	}
	if len(groupIDs) == 0 {
		return out, nil
	}
	ranked, err := s.RankGroups(ctx, userID, groupIDs, at, "")
	if err != nil {
		return out, err
	}
	for i := range ranked {
		candidate := &ranked[i]
		if candidate.Plan.GroupRuleMultiplier != nil && (out.GroupRuleMultiplier == nil || *candidate.Plan.GroupRuleMultiplier < *out.GroupRuleMultiplier) {
			value := *candidate.Plan.GroupRuleMultiplier
			out.GroupRuleMultiplier = &value
		}
		if out.LevelMultiplier == nil || candidate.Plan.EffectiveBaseMultiplier < *out.LevelMultiplier {
			value := candidate.Plan.EffectiveBaseMultiplier
			out.LevelMultiplier = &value
		}
		if out.EffectiveMultiplier == nil || candidate.Plan.EffectiveMultiplier < *out.EffectiveMultiplier {
			value := candidate.Plan.EffectiveMultiplier
			out.EffectiveMultiplier = &value
			out.EffectiveSource = candidate.Plan.EffectiveSource
			if candidate.Group != nil {
				out.MultiplierGroup = candidate.Group.Name
			}
		}
	}
	return out, nil
}

// RecordSpend invalidates the rule profile cache. The usage log is the source
// of truth, so incrementing an old fixed-level snapshot would be incorrect when
// multiple rule windows are assigned to the same user.
func (s *UserLevelService) RecordSpend(userID string, amount float64) {
	if s == nil || amount <= 0 || strings.TrimSpace(userID) == "" {
		return
	}
	s.invalidateProfiles()
}

func dynamicRuleApplies(rule GroupDynamicRateRule, profile UserLevelProfile, at time.Time) (string, bool) {
	if !rule.Enabled || profile.Usage7d < rule.ActivationSpend {
		return "", false
	}
	start, end, quotaKey, ok := parseDynamicRateWindow(rule)
	if !ok || at.Before(start) || !at.Before(end) {
		return "", false
	}
	return quotaKey, true
}

func (s *UserLevelService) resolveGroupPlan(ctx context.Context, userID string, group *Group, profile UserLevelProfile, at time.Time) (UserRatePlan, error) {
	if group == nil {
		return UserRatePlan{}, ErrGroupNotFound
	}
	if !finiteNonnegative(group.RateMultiplier) {
		return UserRatePlan{}, errors.New("group rate multiplier is invalid")
	}
	groupMultiplier := group.RateMultiplier
	userMultiplier := 1.0
	if s.userRateRepo != nil {
		userRate, err := s.userRateRepo.GetByUserAndGroup(ctx, userID, group.ID)
		if err != nil {
			return UserRatePlan{}, err
		}
		if userRate != nil {
			if !finiteNonnegative(*userRate) {
				return UserRatePlan{}, errors.New("user rate multiplier is invalid")
			}
			userMultiplier = *userRate
		}
	}
	selectedBase := groupMultiplier * userMultiplier
	if math.IsNaN(selectedBase) || math.IsInf(selectedBase, 0) || selectedBase < 0 {
		return UserRatePlan{}, errors.New("effective rate multiplier is invalid")
	}
	// Dynamic rules are discounts on top of the static group*user multiplier.
	// A rule is eligible only while its absolute window is active, its rolling
	// spend threshold is met, and its independent per-user quota has remaining
	// capacity. Overlapping eligible rules resolve to the lowest coefficient.
	candidates := make([]DynamicRateCandidate, 0)
	keys := make([]DynamicRateUsageKey, 0)
	for _, rule := range group.DynamicRateRules {
		quotaKey, ok := dynamicRuleApplies(rule, profile, at)
		if !ok {
			continue
		}
		coefficient := rule.DiscountCoefficient
		if coefficient == 0 {
			coefficient = rule.Multiplier
		}
		if !finitePositive(coefficient) || coefficient < 0.01 || coefficient > 1 {
			continue
		}
		keys = append(keys, DynamicRateUsageKey{RuleID: rule.ID, QuotaKey: quotaKey})
		candidates = append(candidates, DynamicRateCandidate{
			RuleID: rule.ID, RuleName: rule.Name, StartAt: rule.StartAt, EndAt: rule.EndAt,
			QuotaKey: quotaKey, Multiplier: coefficient, DiscountCoefficient: coefficient,
			PersonalQuotaAmount: dynamicRatePersonalQuotaAmount(rule),
		})
	}
	if len(keys) > 0 && s.repo != nil {
		used, err := s.repo.GetDynamicRateUsage(ctx, userID, group.ID, keys)
		if err != nil {
			return UserRatePlan{}, err
		}
		filtered := candidates[:0]
		for _, candidate := range candidates {
			key := DynamicRateUsageKey{RuleID: candidate.RuleID, QuotaKey: candidate.QuotaKey}
			candidate.PersonalUsedAmount = used[key]
			if candidate.PersonalQuotaAmount > 0 && candidate.PersonalUsedAmount >= candidate.PersonalQuotaAmount {
				continue
			}
			filtered = append(filtered, candidate)
		}
		candidates = filtered
	}
	selectedRuleID := ""
	effectiveMultiplier := selectedBase
	if len(candidates) > 0 {
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].DiscountCoefficient != candidates[j].DiscountCoefficient {
				return candidates[i].DiscountCoefficient < candidates[j].DiscountCoefficient
			}
			return candidates[i].RuleID < candidates[j].RuleID
		})
		selectedRuleID = candidates[0].RuleID
		effectiveMultiplier = selectedBase * candidates[0].DiscountCoefficient
	}
	return UserRatePlan{
		GroupID: group.ID, UserLevel: profile.Level, Usage7d: profile.Usage7d,
		BaseMultiplier: selectedBase, RateMultiplier: selectedBase, PeakMultiplier: 1, EffectiveMultiplier: effectiveMultiplier,
		Source: "group_times_user", DynamicCandidates: candidates, SelectedDynamicRuleID: selectedRuleID,
		UserLevelMultiplier: nil, UserRateMultiplier: rateCandidatePtr(rateCandidate{value: userMultiplier}), GroupRuleMultiplier: rateCandidatePtr(rateCandidate{value: groupMultiplier}),
		EffectiveBaseMultiplier: selectedBase, EffectiveSource: "group_times_user", NonDynamicMultiplier: selectedBase,
	}, nil
}

type rateCandidate struct {
	value  float64
	source string
}

func lowestRateCandidate(candidates []rateCandidate) rateCandidate {
	if len(candidates) == 0 {
		return rateCandidate{value: 1, source: "default"}
	}
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.value < best.value {
			best = candidate
		}
	}
	return best
}

func rateCandidatePtr(candidate rateCandidate) *float64 {
	value := candidate.value
	return &value
}

func userLevelFloatPtr(value float64) *float64 {
	return &value
}

func finitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func sanitizePeakMultiplier(value float64) float64 {
	// A zero peak multiplier is a valid free window. Only malformed values
	// should fall back to the neutral multiplier.
	if !finiteNonnegative(value) {
		return 1
	}
	return value
}

func finiteNonnegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func dynamicRateRuleStatus(rule GroupDynamicRateRule, at time.Time) string {
	if isLegacyDynamicRateRule(rule) || len(rule.Levels) > 0 {
		return DynamicRateStatusLegacy
	}
	start, end, _, ok := parseDynamicRateWindow(rule)
	if !ok {
		return DynamicRateStatusInvalid
	}
	if at.Before(start) {
		return DynamicRateStatusNotStarted
	}
	if at.Before(end) {
		return DynamicRateStatusActive
	}
	return DynamicRateStatusExpired
}

func (s *UserLevelService) GetDynamicRateUsageSummary(ctx context.Context, groupID string, at time.Time) ([]DynamicRateUsageSummary, error) {
	if s == nil || s.groupRepo == nil {
		return nil, errors.New("user level group repository is unavailable")
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, ErrGroupNotFound
	}
	if at.IsZero() {
		at = time.Now()
	}
	group, err := s.groupRepo.GetByIDLite(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}
	result := make([]DynamicRateUsageSummary, 0, len(group.DynamicRateRules))
	for _, rule := range group.DynamicRateRules {
		start, end, _, validWindow := parseDynamicRateWindow(rule)
		summary := DynamicRateUsageSummary{
			RuleID: rule.ID, RuleName: rule.Name, Status: dynamicRateRuleStatus(rule, at),
			DiscountCoefficient: func() float64 { if rule.DiscountCoefficient > 0 { return rule.DiscountCoefficient }; if rule.Multiplier > 0 { return rule.Multiplier }; return 1 }(),
			PersonalQuotaAmount: QuantizeUsageBillingAmount(dynamicRatePersonalQuotaAmount(rule)), UsageScope: "per_user",
		}
		if validWindow {
			summary.StartAt = start.Format(time.RFC3339Nano)
			summary.EndAt = end.Format(time.RFC3339Nano)
		}
		result = append(result, summary)
	}
	return result, nil
}

// ResolvePlan resolves one group using the same profile and candidate logic as
// RankGroups. Gateway paths that did not precompute a selection must call this
// instead of falling back to the historical user-overrides-group behavior.
func (s *UserLevelService) ResolvePlan(ctx context.Context, userID string, group *Group, at time.Time) (UserRatePlan, error) {
	profile, err := s.ResolveProfile(ctx, userID, at)
	if err != nil {
		return UserRatePlan{}, err
	}
	return s.resolveGroupPlan(ctx, userID, group, profile, at)
}

// resolveGroupBaseMultiplier is kept for old dashboard callers. It now uses
// the current tier UUID candidates and the independent user-level minimum.
func (s *UserLevelService) resolveGroupBaseMultiplier(ctx context.Context, userID string, group *Group, level int) (float64, error) {
	_ = level
	plan, err := s.ResolvePlan(ctx, userID, group, time.Now())
	return plan.EffectiveBaseMultiplier, err
}

func (s *UserLevelService) RankGroups(ctx context.Context, userID string, groupIDs []string, at time.Time, platform string) ([]RankedUserGroup, error) {
	if s == nil || s.groupRepo == nil || strings.TrimSpace(userID) == "" {
		return nil, nil
	}
	profile, err := s.ResolveProfile(ctx, userID, at)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(groupIDs))
	ranked := make([]RankedUserGroup, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		groupID = strings.TrimSpace(groupID)
		if groupID == "" {
			continue
		}
		if _, exists := seen[groupID]; exists {
			continue
		}
		seen[groupID] = struct{}{}
		group, loadErr := s.groupRepo.GetByIDLite(ctx, groupID)
		if loadErr != nil || group == nil || !group.IsActive() {
			continue
		}
		if platform != "" && group.Platform != PlatformComposite && !strings.EqualFold(group.Platform, platform) {
			continue
		}
		var subscription *UserSubscription
		if group.IsSubscriptionType() {
			if s.subRepo == nil {
				continue
			}
			subscription, loadErr = s.subRepo.GetActiveByUserIDAndGroupID(ctx, userID, group.ID)
			if loadErr != nil || subscription == nil || !subscription.IsActive() {
				continue
			}
			daily, weekly, monthly := subscription.CheckAllLimits(group, 0)
			if !daily || !weekly || !monthly {
				continue
			}
		}
		plan, planErr := s.resolveGroupPlan(ctx, userID, group, profile, at)
		if planErr != nil {
			continue
		}
		ranked = append(ranked, RankedUserGroup{Group: group, Subscription: subscription, Plan: plan})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Plan.EffectiveMultiplier != ranked[j].Plan.EffectiveMultiplier {
			return ranked[i].Plan.EffectiveMultiplier < ranked[j].Plan.EffectiveMultiplier
		}
		return ranked[i].Group.ID < ranked[j].Group.ID
	})
	return ranked, nil
}

// The decimal helpers keep profile cache keys independent of a settings JSON
// representation and avoid a mutable configuration object in the key.
func formatUnsignedDecimal(value uint64) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}

func formatSignedDecimal(value int64) string {
	if value >= 0 {
		return formatUnsignedDecimal(uint64(value))
	}
	return "-" + formatUnsignedDecimal(uint64(-(value + 1))) + "1"
}

func timeFormatInt64(value int64) string   { return formatSignedDecimal(value) }
func timeFormatUint64(value uint64) string { return formatUnsignedDecimal(value) }
