package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	gocache "github.com/patrickmn/go-cache"
)

const (
	SettingKeyUserLevelSettings = "user_level_settings"
	userLevelWindow             = 7 * 24 * time.Hour
	userLevelProfileCacheTTL    = time.Minute
	userLevelSettingsCacheTTL   = 30 * time.Second
)

type UserLevelSettings struct {
	L2MinSpend  float64 `json:"l2_min_spend"`
	L3MinSpend  float64 `json:"l3_min_spend"`
	WindowHours int     `json:"window_hours"`
}

func DefaultUserLevelSettings() UserLevelSettings {
	return UserLevelSettings{L2MinSpend: 50, L3MinSpend: 200, WindowHours: 168}
}

func ValidateUserLevelSettings(settings UserLevelSettings) (UserLevelSettings, error) {
	if math.IsNaN(settings.L2MinSpend) || math.IsInf(settings.L2MinSpend, 0) || settings.L2MinSpend < 0 {
		return settings, infraerrors.BadRequest("USER_LEVEL_L2_INVALID", "L2 minimum spend must be nonnegative")
	}
	if math.IsNaN(settings.L3MinSpend) || math.IsInf(settings.L3MinSpend, 0) || settings.L3MinSpend <= settings.L2MinSpend {
		return settings, infraerrors.BadRequest("USER_LEVEL_L3_INVALID", "L3 minimum spend must be greater than L2")
	}
	settings.L2MinSpend = QuantizeUsageBillingAmount(settings.L2MinSpend)
	settings.L3MinSpend = QuantizeUsageBillingAmount(settings.L3MinSpend)
	settings.WindowHours = int(userLevelWindow / time.Hour)
	return settings, nil
}

type UserLevelProfile struct {
	UserID       string    `json:"user_id"`
	Level        int       `json:"level"`
	Usage7d      float64   `json:"usage_7d"`
	WindowFrom   time.Time `json:"window_from"`
	CalculatedAt time.Time `json:"calculated_at"`
}

// UserLevelDashboard is the user-safe summary exposed by the personal
// dashboard. LevelMultiplier is the lowest level/base multiplier among the
// user's currently available groups; EffectiveMultiplier also includes any
// active dynamic and peak multipliers.
type UserLevelDashboard struct {
	UserID              string    `json:"user_id"`
	Level               int       `json:"level"`
	Usage7d             float64   `json:"usage_7d"`
	WindowHours         int       `json:"window_hours"`
	WindowFrom          time.Time `json:"window_from"`
	CalculatedAt        time.Time `json:"calculated_at"`
	L2MinSpend          float64   `json:"l2_min_spend"`
	L3MinSpend          float64   `json:"l3_min_spend"`
	LevelMultiplier     *float64  `json:"level_multiplier,omitempty"`
	EffectiveMultiplier *float64  `json:"effective_multiplier,omitempty"`
	MultiplierGroup     string    `json:"multiplier_group,omitempty"`
	NextLevelMultiplier *float64  `json:"next_level_multiplier,omitempty"`
	NextMultiplierGroup string    `json:"next_multiplier_group,omitempty"`
}

type DynamicRateUsageKey struct {
	RuleID   string
	QuotaKey string
}

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
	Multiplier          float64 `json:"multiplier"`
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
	RuleID                string   `json:"rule_id"`
	RuleName              string   `json:"rule_name"`
	StartAt               string   `json:"start_at"`
	EndAt                 string   `json:"end_at"`
	Status                string   `json:"status"`
	SharedQuotaAmount     float64  `json:"shared_quota_amount"`
	SharedUsedAmount      float64  `json:"shared_used_amount"`
	SharedRemainingAmount *float64 `json:"shared_remaining_amount"`
}

type UserRatePlan struct {
	GroupID             string                 `json:"group_id"`
	UserLevel           int                    `json:"user_level"`
	Usage7d             float64                `json:"usage_7d"`
	BaseMultiplier      float64                `json:"base_multiplier"`
	PeakMultiplier      float64                `json:"peak_multiplier"`
	EffectiveMultiplier float64                `json:"effective_multiplier"`
	Source              string                 `json:"source"`
	DynamicCandidates   []DynamicRateCandidate `json:"dynamic_candidates"`
}

type RankedUserGroup struct {
	Group        *Group
	Subscription *UserSubscription
	Plan         UserRatePlan
}

type cachedUserLevelSettings struct {
	value     UserLevelSettings
	expiresAt time.Time
}

type UserLevelService struct {
	repo          UserLevelRepository
	settingRepo   SettingRepository
	groupRepo     GroupRepository
	userRateRepo  UserGroupRateRepository
	subRepo       UserSubscriptionRepository
	profileCache  *gocache.Cache
	settingsMu    sync.Mutex
	settingsCache cachedUserLevelSettings
}

func NewUserLevelService(repo UserLevelRepository, settingRepo SettingRepository, groupRepo GroupRepository, userRateRepo UserGroupRateRepository, subRepo UserSubscriptionRepository) *UserLevelService {
	return &UserLevelService{
		repo: repo, settingRepo: settingRepo, groupRepo: groupRepo, userRateRepo: userRateRepo, subRepo: subRepo,
		profileCache: gocache.New(userLevelProfileCacheTTL, time.Minute),
	}
}

func (s *UserLevelService) GetSettings(ctx context.Context) (UserLevelSettings, error) {
	defaults := DefaultUserLevelSettings()
	if s == nil || s.settingRepo == nil {
		return defaults, nil
	}
	now := time.Now()
	s.settingsMu.Lock()
	if now.Before(s.settingsCache.expiresAt) {
		value := s.settingsCache.value
		s.settingsMu.Unlock()
		return value, nil
	}
	s.settingsMu.Unlock()

	raw, err := s.settingRepo.GetValue(ctx, SettingKeyUserLevelSettings)
	if err != nil {
		if !errors.Is(err, ErrSettingNotFound) {
			return defaults, err
		}
	} else if strings.TrimSpace(raw) != "" {
		if unmarshalErr := json.Unmarshal([]byte(raw), &defaults); unmarshalErr != nil {
			return DefaultUserLevelSettings(), fmt.Errorf("parse user level settings: %w", unmarshalErr)
		}
	}
	settings, err := ValidateUserLevelSettings(defaults)
	if err != nil {
		return DefaultUserLevelSettings(), err
	}
	s.settingsMu.Lock()
	s.settingsCache = cachedUserLevelSettings{value: settings, expiresAt: now.Add(userLevelSettingsCacheTTL)}
	s.settingsMu.Unlock()
	return settings, nil
}

func (s *UserLevelService) UpdateSettings(ctx context.Context, settings UserLevelSettings) (UserLevelSettings, error) {
	settings, err := ValidateUserLevelSettings(settings)
	if err != nil {
		return settings, err
	}
	if s == nil || s.settingRepo == nil {
		return settings, errors.New("user level setting repository is unavailable")
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return settings, err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyUserLevelSettings, string(raw)); err != nil {
		return settings, err
	}
	s.settingsMu.Lock()
	s.settingsCache = cachedUserLevelSettings{value: settings, expiresAt: time.Now().Add(userLevelSettingsCacheTTL)}
	s.settingsMu.Unlock()
	if s.profileCache != nil {
		s.profileCache.Flush()
	}
	return settings, nil
}

func userLevelForSpend(spend float64, settings UserLevelSettings) int {
	if spend >= settings.L3MinSpend {
		return 3
	}
	if spend >= settings.L2MinSpend {
		return 2
	}
	return 1
}

func (s *UserLevelService) ResolveProfile(ctx context.Context, userID string, at time.Time) (UserLevelProfile, error) {
	if at.IsZero() {
		at = time.Now()
	}
	profile := UserLevelProfile{UserID: userID, Level: 1, WindowFrom: at.Add(-userLevelWindow), CalculatedAt: at}
	if s == nil || s.repo == nil || strings.TrimSpace(userID) == "" {
		return profile, nil
	}
	if cached, ok := s.profileCache.Get(userID); ok {
		if value, valid := cached.(UserLevelProfile); valid {
			return value, nil
		}
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return profile, err
	}
	spend, err := s.repo.GetRollingSpend(ctx, userID, profile.WindowFrom, at)
	if err != nil {
		return profile, err
	}
	profile.Usage7d = QuantizeUsageBillingAmount(spend)
	profile.Level = userLevelForSpend(profile.Usage7d, settings)
	s.profileCache.Set(userID, profile, userLevelProfileCacheTTL)
	return profile, nil
}

func (s *UserLevelService) GetProfiles(ctx context.Context, userIDs []string, at time.Time) (map[string]UserLevelProfile, error) {
	if at.IsZero() {
		at = time.Now()
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	spendByUser, err := s.repo.GetRollingSpendBatch(ctx, userIDs, at.Add(-userLevelWindow), at)
	if err != nil {
		return nil, err
	}
	out := make(map[string]UserLevelProfile, len(userIDs))
	for _, userID := range userIDs {
		spend := QuantizeUsageBillingAmount(spendByUser[userID])
		profile := UserLevelProfile{UserID: userID, Level: userLevelForSpend(spend, settings), Usage7d: spend, WindowFrom: at.Add(-userLevelWindow), CalculatedAt: at}
		out[userID] = profile
		if s.profileCache != nil {
			s.profileCache.Set(userID, profile, userLevelProfileCacheTTL)
		}
	}
	return out, nil
}

// ResolveDashboard resolves the current user's level and the lowest rate
// visible across the groups the caller has already authorized for the user.
// Group ranking remains centralized in RankGroups so dynamic quota and
// subscription checks use exactly the same rules as request scheduling.
func (s *UserLevelService) ResolveDashboard(ctx context.Context, userID string, groupIDs []string, at time.Time) (UserLevelDashboard, error) {
	if at.IsZero() {
		at = time.Now()
	}
	profile, err := s.ResolveProfile(ctx, userID, at)
	if err != nil {
		return UserLevelDashboard{}, err
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return UserLevelDashboard{}, err
	}
	out := UserLevelDashboard{
		UserID:       profile.UserID,
		Level:        profile.Level,
		Usage7d:      profile.Usage7d,
		WindowHours:  settings.WindowHours,
		WindowFrom:   profile.WindowFrom,
		CalculatedAt: profile.CalculatedAt,
		L2MinSpend:   settings.L2MinSpend,
		L3MinSpend:   settings.L3MinSpend,
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
		if out.LevelMultiplier == nil || candidate.Plan.BaseMultiplier < *out.LevelMultiplier {
			value := candidate.Plan.BaseMultiplier
			out.LevelMultiplier = &value
		}
		if out.EffectiveMultiplier == nil || candidate.Plan.EffectiveMultiplier < *out.EffectiveMultiplier {
			value := candidate.Plan.EffectiveMultiplier
			out.EffectiveMultiplier = &value
			if candidate.Group != nil {
				out.MultiplierGroup = candidate.Group.Name
			}
		}
		if profile.Level < 3 && candidate.Group != nil {
			nextMultiplier, resolveErr := s.resolveGroupBaseMultiplier(ctx, userID, candidate.Group, profile.Level+1)
			if resolveErr != nil {
				return out, resolveErr
			}
			if out.NextLevelMultiplier == nil || nextMultiplier < *out.NextLevelMultiplier {
				value := nextMultiplier
				out.NextLevelMultiplier = &value
				out.NextMultiplierGroup = candidate.Group.Name
			}
		}
	}
	return out, nil
}

func (s *UserLevelService) RecordSpend(userID string, amount float64) {
	if s == nil || s.profileCache == nil || amount <= 0 {
		return
	}
	cached, ok := s.profileCache.Get(userID)
	if !ok {
		return
	}
	profile, ok := cached.(UserLevelProfile)
	if !ok {
		return
	}
	settings, err := s.GetSettings(context.Background())
	if err != nil {
		s.profileCache.Delete(userID)
		return
	}
	profile.Usage7d = QuantizeUsageBillingAmount(profile.Usage7d + amount)
	profile.Level = userLevelForSpend(profile.Usage7d, settings)
	profile.CalculatedAt = time.Now()
	s.profileCache.Set(userID, profile, userLevelProfileCacheTTL)
}

func dynamicRuleApplies(rule GroupDynamicRateRule, profile UserLevelProfile, at time.Time) (string, bool) {
	if !rule.Enabled || profile.Usage7d < rule.ActivationSpend {
		return "", false
	}
	if len(rule.Levels) > 0 {
		matched := false
		for _, level := range rule.Levels {
			if level == profile.Level {
				matched = true
				break
			}
		}
		if !matched {
			return "", false
		}
	}
	start, end, quotaKey, ok := parseDynamicRateWindow(rule)
	if !ok {
		return "", false
	}
	if at.Before(start) || !at.Before(end) {
		return "", false
	}
	return quotaKey, true
}

func (s *UserLevelService) resolveGroupPlan(ctx context.Context, userID string, group *Group, profile UserLevelProfile, at time.Time) (UserRatePlan, error) {
	base, source, err := s.resolveGroupBaseRate(ctx, userID, group, profile.Level)
	if err != nil {
		return UserRatePlan{}, err
	}

	candidates := make([]DynamicRateCandidate, 0)
	keys := make([]DynamicRateUsageKey, 0)
	for _, rule := range group.DynamicRateRules {
		quotaKey, ok := dynamicRuleApplies(rule, profile, at)
		if !ok {
			continue
		}
		start, end, _, validWindow := parseDynamicRateWindow(rule)
		if !validWindow {
			continue
		}
		candidate := DynamicRateCandidate{
			RuleID:              rule.ID,
			RuleName:            rule.Name,
			StartAt:             start.Format(time.RFC3339Nano),
			EndAt:               end.Format(time.RFC3339Nano),
			QuotaKey:            quotaKey,
			Multiplier:          rule.Multiplier,
			SharedQuotaAmount:   rule.SharedQuotaAmount,
			PersonalQuotaAmount: rule.PersonalQuotaAmount,
		}
		candidates = append(candidates, candidate)
		if rule.SharedQuotaAmount > 0 || rule.PersonalQuotaAmount > 0 {
			keys = append(keys, DynamicRateUsageKey{RuleID: rule.ID, QuotaKey: quotaKey})
		}
	}
	if len(keys) > 0 && s.repo != nil {
		personalUsage, err := s.repo.GetDynamicRateUsage(ctx, userID, group.ID, keys)
		if err != nil {
			return UserRatePlan{}, err
		}
		sharedUsage, err := s.repo.GetSharedDynamicRateUsage(ctx, group.ID, keys)
		if err != nil {
			return UserRatePlan{}, err
		}
		for i := range candidates {
			key := DynamicRateUsageKey{RuleID: candidates[i].RuleID, QuotaKey: candidates[i].QuotaKey}
			candidates[i].PersonalUsedAmount = personalUsage[key]
			candidates[i].SharedUsedAmount = sharedUsage[key]
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Multiplier != candidates[j].Multiplier {
			return candidates[i].Multiplier < candidates[j].Multiplier
		}
		return candidates[i].RuleID < candidates[j].RuleID
	})
	usable := candidates[:0]
	for _, candidate := range candidates {
		sharedAvailable := candidate.SharedQuotaAmount == 0 || candidate.SharedUsedAmount < candidate.SharedQuotaAmount
		personalAvailable := candidate.PersonalQuotaAmount == 0 || candidate.PersonalUsedAmount < candidate.PersonalQuotaAmount
		if sharedAvailable && personalAvailable {
			usable = append(usable, candidate)
		}
	}
	candidates = usable
	selected := base
	if len(candidates) > 0 && candidates[0].Multiplier < selected {
		selected = candidates[0].Multiplier
		source = "dynamic"
	}
	peak := group.PeakMultiplierAt(at)
	return UserRatePlan{
		GroupID: group.ID, UserLevel: profile.Level, Usage7d: profile.Usage7d,
		BaseMultiplier: base, PeakMultiplier: peak, EffectiveMultiplier: selected * peak,
		Source: source, DynamicCandidates: candidates,
	}, nil
}

func dynamicRateRuleStatus(rule GroupDynamicRateRule, at time.Time) string {
	if isLegacyDynamicRateRule(rule) {
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

	keys := make([]DynamicRateUsageKey, 0, len(group.DynamicRateRules))
	for _, rule := range group.DynamicRateRules {
		if _, _, quotaKey, ok := parseDynamicRateWindow(rule); ok {
			keys = append(keys, DynamicRateUsageKey{RuleID: rule.ID, QuotaKey: quotaKey})
		}
	}
	sharedUsage := make(map[DynamicRateUsageKey]float64, len(keys))
	if len(keys) > 0 && s.repo != nil {
		sharedUsage, err = s.repo.GetSharedDynamicRateUsage(ctx, group.ID, keys)
		if err != nil {
			return nil, err
		}
	}

	result := make([]DynamicRateUsageSummary, 0, len(group.DynamicRateRules))
	for _, rule := range group.DynamicRateRules {
		start, end, quotaKey, validWindow := parseDynamicRateWindow(rule)
		summary := DynamicRateUsageSummary{
			RuleID:            rule.ID,
			RuleName:          rule.Name,
			Status:            dynamicRateRuleStatus(rule, at),
			SharedQuotaAmount: QuantizeUsageBillingAmount(rule.SharedQuotaAmount),
		}
		if validWindow {
			summary.StartAt = start.Format(time.RFC3339Nano)
			summary.EndAt = end.Format(time.RFC3339Nano)
			summary.SharedUsedAmount = QuantizeUsageBillingAmount(sharedUsage[DynamicRateUsageKey{RuleID: rule.ID, QuotaKey: quotaKey}])
		}
		if rule.SharedQuotaAmount > 0 {
			remaining := rule.SharedQuotaAmount - summary.SharedUsedAmount
			if remaining < 0 {
				remaining = 0
			}
			remaining = QuantizeUsageBillingAmount(remaining)
			summary.SharedRemainingAmount = &remaining
		}
		result = append(result, summary)
	}
	return result, nil
}

func (s *UserLevelService) resolveGroupBaseMultiplier(ctx context.Context, userID string, group *Group, level int) (float64, error) {
	multiplier, _, err := s.resolveGroupBaseRate(ctx, userID, group, level)
	return multiplier, err
}

// resolveGroupBaseRate is shared by live scheduling and the dashboard's next
// level preview. User-specific group rates continue to override level rates,
// so the preview always matches the rate that would actually be selected after
// the user advances.
func (s *UserLevelService) resolveGroupBaseRate(ctx context.Context, userID string, group *Group, level int) (float64, string, error) {
	base := group.RateMultiplier
	source := "group"
	if levelRate, ok := group.LevelRateMultipliers[strconv.Itoa(level)]; ok {
		base = levelRate
		source = "level"
	}
	if s.userRateRepo != nil {
		userRate, err := s.userRateRepo.GetByUserAndGroup(ctx, userID, group.ID)
		if err != nil {
			return 0, "", err
		}
		if userRate != nil {
			base = *userRate
			source = "user"
		}
	}
	return base, source, nil
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
		if groupID = strings.TrimSpace(groupID); groupID == "" {
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
