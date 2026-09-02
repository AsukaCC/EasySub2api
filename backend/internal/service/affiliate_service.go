package service

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
	"github.com/AsukaCC/EasySub2api/internal/pkg/logger"
	appTimezone "github.com/AsukaCC/EasySub2api/internal/pkg/timezone"
)

var (
	ErrAffiliateProfileNotFound       = infraerrors.NotFound("AFFILIATE_PROFILE_NOT_FOUND", "affiliate profile not found")
	ErrAffiliateCodeInvalid           = infraerrors.BadRequest("AFFILIATE_CODE_INVALID", "invalid affiliate code")
	ErrAffiliateCodeTaken             = infraerrors.Conflict("AFFILIATE_CODE_TAKEN", "affiliate code already in use")
	ErrAffiliateAlreadyBound          = infraerrors.Conflict("AFFILIATE_ALREADY_BOUND", "affiliate inviter already bound")
	ErrAffiliateQuotaEmpty            = infraerrors.BadRequest("AFFILIATE_QUOTA_EMPTY", "no affiliate quota available to transfer")
	ErrAffiliateCodeRegenerationLimit = infraerrors.TooManyRequests(
		"AFFILIATE_CODE_REGENERATION_LIMIT",
		"monthly affiliate code regeneration limit reached",
	)
)

const (
	affiliateInviteesLimit                = 100
	affiliateRebateFreezeHours            = 168
	AffiliateCodeRegenerationMonthlyLimit = 3
	// AffiliateCodeMinLength / AffiliateCodeMaxLength bound both system-generated
	// 12-char codes and admin-customized codes (e.g. "VIP2026").
	AffiliateCodeMinLength = 4
	AffiliateCodeMaxLength = 32
)

// affiliateCodeValidChar accepts uppercase letters, digits, underscore and dash.
// All input passes through strings.ToUpper before validation, so lowercase from
// users is normalized — admins may supply mixed case in their UI.
var affiliateCodeValidChar = func() [256]bool {
	var tbl [256]bool
	for c := byte('A'); c <= 'Z'; c++ {
		tbl[c] = true
	}
	for c := byte('0'); c <= '9'; c++ {
		tbl[c] = true
	}
	tbl['_'] = true
	tbl['-'] = true
	return tbl
}()

// isValidAffiliateCodeFormat validates code format for both binding (user input)
// and admin updates. Caller is expected to upper-case the input first.
func isValidAffiliateCodeFormat(code string) bool {
	if len(code) < AffiliateCodeMinLength || len(code) > AffiliateCodeMaxLength {
		return false
	}
	for i := 0; i < len(code); i++ {
		if !affiliateCodeValidChar[code[i]] {
			return false
		}
	}
	return true
}

type AffiliateSummary struct {
	UserID                         string     `json:"user_id"`
	AffCode                        string     `json:"aff_code"`
	AffCodeCustom                  bool       `json:"aff_code_custom"`
	AffRebateRatePercent           *float64   `json:"aff_rebate_rate_percent,omitempty"`
	InviterID                      *string    `json:"inviter_id,omitempty"`
	AffCount                       int        `json:"aff_count"`
	AffQuota                       float64    `json:"aff_quota"`
	AffFrozenQuota                 float64    `json:"aff_frozen_quota"`
	AffHistoryQuota                float64    `json:"aff_history_quota"`
	AffCodeRegenerationPeriodStart *time.Time `json:"-"`
	AffCodeRegenerationCount       int        `json:"-"`
	CreatedAt                      time.Time  `json:"created_at"`
	UpdatedAt                      time.Time  `json:"updated_at"`
}

type AffiliateInvitee struct {
	UserID      string     `json:"user_id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	TotalRebate float64    `json:"total_rebate"`
}

type AffiliateDetail struct {
	UserID          string  `json:"user_id"`
	AffCode         string  `json:"aff_code"`
	InviterID       *string `json:"inviter_id,omitempty"`
	AffCount        int     `json:"aff_count"`
	AffQuota        float64 `json:"aff_quota"`
	AffFrozenQuota  float64 `json:"aff_frozen_quota"`
	AffHistoryQuota float64 `json:"aff_history_quota"`
	// EffectiveRebateRatePercent 是当前用户作为邀请人时实际生效的返利比例：
	// 优先用户自己的专属比例（aff_rebate_rate_percent），否则回退到全局比例。
	// 用于在用户的 /affiliate 页面直观展示「分享后能拿到多少」。
	EffectiveRebateRatePercent       float64            `json:"effective_rebate_rate_percent"`
	RebateRecipient                  string             `json:"rebate_recipient"`
	CodeRegenerationLimit            int                `json:"code_regeneration_limit"`
	CodeRegenerationUsed             int                `json:"code_regeneration_used"`
	CodeRegenerationRemaining        int                `json:"code_regeneration_remaining"`
	CodeRegenerationResetAt          time.Time          `json:"code_regeneration_reset_at"`
	Invitees                         []AffiliateInvitee `json:"invitees"`
	CanBindInviter                   bool               `json:"can_bind_inviter"`
	InviteeBindingRewardPoints       float64            `json:"invitee_binding_reward_points"`
	InviteeBindingRewardValidityDays int                `json:"invitee_binding_reward_validity_days"`
}

type AffiliateCodeRegenerationResult struct {
	AffCode   string    `json:"aff_code"`
	Limit     int       `json:"limit"`
	Used      int       `json:"used"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"`
}

type AffiliateRepository interface {
	EnsureUserAffiliate(ctx context.Context, userID string) (*AffiliateSummary, error)
	GetAffiliateByCode(ctx context.Context, code string) (*AffiliateSummary, error)
	BindInviter(ctx context.Context, userID, inviterID string) (bool, error)
	AccrueQuota(ctx context.Context, inviterID, inviteeUserID string, amount float64, freezeHours int, sourceOrderID *string) (bool, error)
	GetAccruedRebateFromInvitee(ctx context.Context, inviterID, inviteeUserID string) (float64, error)
	ThawFrozenQuota(ctx context.Context, userID string) (float64, error)
	TransferQuotaToBalance(ctx context.Context, userID string) (float64, float64, error)
	ListInvitees(ctx context.Context, inviterID string, limit int) ([]AffiliateInvitee, error)
	RegenerateUserAffCode(ctx context.Context, userID string, periodStart time.Time, limit int) (string, int, error)

	// 管理端：用户级专属配置
	UpdateUserAffCode(ctx context.Context, userID string, newCode string) error
	ResetUserAffCode(ctx context.Context, userID string) (string, error)
	SetUserRebateRate(ctx context.Context, userID string, ratePercent *float64) error
	BatchSetUserRebateRate(ctx context.Context, userIDs []string, ratePercent *float64) error
	ListUsersWithCustomSettings(ctx context.Context, filter AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error)
	ListAffiliateInviteRecords(ctx context.Context, filter AffiliateRecordFilter) ([]AffiliateInviteRecord, int64, error)
	ListAffiliateRebateRecords(ctx context.Context, filter AffiliateRecordFilter) ([]AffiliateRebateRecord, int64, error)
	ListAffiliateTransferRecords(ctx context.Context, filter AffiliateRecordFilter) ([]AffiliateTransferRecord, int64, error)
	GetAffiliateUserOverview(ctx context.Context, userID string) (*AffiliateUserOverview, error)
}

// AffiliateAdminFilter 列表筛选条件
type AffiliateAdminFilter struct {
	Search   string
	Page     int
	PageSize int
}

// AffiliateAdminEntry 专属用户列表条目
type AffiliateAdminEntry struct {
	UserID               string   `json:"user_id"`
	Email                string   `json:"email"`
	Username             string   `json:"username"`
	AffCode              string   `json:"aff_code"`
	AffCodeCustom        bool     `json:"aff_code_custom"`
	AffRebateRatePercent *float64 `json:"aff_rebate_rate_percent,omitempty"`
	AffCount             int      `json:"aff_count"`
}

type AffiliateRecordFilter struct {
	Search   string
	Page     int
	PageSize int
	StartAt  *time.Time
	EndAt    *time.Time
	SortBy   string
	SortDesc bool
}

type AffiliateInviteRecord struct {
	InviterID       string    `json:"inviter_id"`
	InviterEmail    string    `json:"inviter_email"`
	InviterUsername string    `json:"inviter_username"`
	InviteeID       string    `json:"invitee_id"`
	InviteeEmail    string    `json:"invitee_email"`
	InviteeUsername string    `json:"invitee_username"`
	AffCode         string    `json:"aff_code"`
	TotalRebate     float64   `json:"total_rebate"`
	CreatedAt       time.Time `json:"created_at"`
}

type AffiliateRebateRecord struct {
	OrderID                string    `json:"order_id"`
	OutTradeNo             string    `json:"out_trade_no"`
	InviterID              string    `json:"inviter_id"`
	InviterEmail           string    `json:"inviter_email"`
	InviterUsername        string    `json:"inviter_username"`
	InviteeID              string    `json:"invitee_id"`
	InviteeEmail           string    `json:"invitee_email"`
	InviteeUsername        string    `json:"invitee_username"`
	RecipientID            string    `json:"recipient_id"`
	RecipientEmail         string    `json:"recipient_email"`
	RecipientUsername      string    `json:"recipient_username"`
	RebateRecipient        string    `json:"rebate_recipient"`
	OrderAmount            float64   `json:"order_amount"`
	PayAmount              float64   `json:"pay_amount"`
	RebateAmount           float64   `json:"rebate_amount"`
	ReservedReversalPoints float64   `json:"reserved_reversal_points"`
	ReversedPoints         float64   `json:"reversed_points"`
	NetRebatePoints        float64   `json:"net_rebate_points"`
	PaymentType            string    `json:"payment_type"`
	OrderStatus            string    `json:"order_status"`
	CreatedAt              time.Time `json:"created_at"`
}

type AffiliateTransferRecord struct {
	LedgerID            string    `json:"ledger_id"`
	UserID              string    `json:"user_id"`
	UserEmail           string    `json:"user_email"`
	Username            string    `json:"username"`
	Amount              float64   `json:"amount"`
	BalanceAfter        *float64  `json:"balance_after,omitempty"`
	AvailableQuotaAfter *float64  `json:"available_quota_after,omitempty"`
	FrozenQuotaAfter    *float64  `json:"frozen_quota_after,omitempty"`
	HistoryQuotaAfter   *float64  `json:"history_quota_after,omitempty"`
	SnapshotAvailable   bool      `json:"snapshot_available"`
	CurrentBalance      float64   `json:"-"`
	RemainingQuota      float64   `json:"-"`
	FrozenQuota         float64   `json:"-"`
	HistoryQuota        float64   `json:"-"`
	CreatedAt           time.Time `json:"created_at"`
}

type AffiliateUserOverview struct {
	UserID              string  `json:"user_id"`
	Email               string  `json:"email"`
	Username            string  `json:"username"`
	AffCode             string  `json:"aff_code"`
	RebateRatePercent   float64 `json:"rebate_rate_percent"`
	RebateRateCustom    bool    `json:"-"`
	InvitedCount        int     `json:"invited_count"`
	RebatedInviteeCount int     `json:"rebated_invitee_count"`
	AvailableQuota      float64 `json:"available_quota"`
	HistoryQuota        float64 `json:"history_quota"`
}

type AffiliateService struct {
	repo                 AffiliateRepository
	settingService       *SettingService
	authCacheInvalidator APIKeyAuthCacheInvalidator
	billingCacheService  *BillingCacheService
}

func NewAffiliateService(repo AffiliateRepository, settingService *SettingService, authCacheInvalidator APIKeyAuthCacheInvalidator, billingCacheService *BillingCacheService) *AffiliateService {
	service := &AffiliateService{
		repo:                 repo,
		settingService:       settingService,
		authCacheInvalidator: authCacheInvalidator,
		billingCacheService:  billingCacheService,
	}
	go service.resumePersistedRewardBackfill()
	return service
}

// IsEnabled reports whether the affiliate (邀请返利) feature is turned on.
func (s *AffiliateService) IsEnabled(ctx context.Context) bool {
	if s == nil || s.settingService == nil {
		return AffiliateEnabledDefault
	}
	return s.settingService.IsAffiliateEnabled(ctx)
}

func (s *AffiliateService) IsUserAvailable(ctx context.Context) bool {
	return s != nil && s.settingService != nil && s.settingService.IsAffiliateUserAvailable(ctx)
}

func (s *AffiliateService) EnsureUserAffiliate(ctx context.Context, userID string) (*AffiliateSummary, error) {
	if userID == "" {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	return s.repo.EnsureUserAffiliate(ctx, userID)
}

func (s *AffiliateService) GetAffiliateDetail(ctx context.Context, userID string) (*AffiliateDetail, error) {
	if !s.IsEnabled(ctx) {
		return nil, infraerrors.NotFound("FEATURE_DISABLED", "feature is disabled")
	}
	// Lazy thaw: move any matured frozen quota to available before reading.
	if s != nil && s.repo != nil {
		// best-effort: thaw failure is non-fatal
		_, _ = s.repo.ThawFrozenQuota(ctx, userID)
	}

	summary, err := s.EnsureUserAffiliate(ctx, userID)
	if err != nil {
		return nil, err
	}
	invitees, err := s.listInvitees(ctx, userID)
	if err != nil {
		return nil, err
	}
	regenerationUsed, regenerationResetAt := affiliateCodeRegenerationUsage(summary, appTimezone.Now())
	bindingRewards := s.bindingRewardConfig(ctx)
	return &AffiliateDetail{
		UserID:                           summary.UserID,
		AffCode:                          summary.AffCode,
		InviterID:                        summary.InviterID,
		AffCount:                         summary.AffCount,
		AffQuota:                         summary.AffQuota,
		AffFrozenQuota:                   summary.AffFrozenQuota,
		AffHistoryQuota:                  summary.AffHistoryQuota,
		EffectiveRebateRatePercent:       s.resolveRebateRatePercent(ctx, summary),
		RebateRecipient:                  s.rebateRecipient(ctx),
		CodeRegenerationLimit:            AffiliateCodeRegenerationMonthlyLimit,
		CodeRegenerationUsed:             regenerationUsed,
		CodeRegenerationRemaining:        max(AffiliateCodeRegenerationMonthlyLimit-regenerationUsed, 0),
		CodeRegenerationResetAt:          regenerationResetAt,
		Invitees:                         invitees,
		CanBindInviter:                   summary.InviterID == nil,
		InviteeBindingRewardPoints:       bindingRewards.InviteePoints,
		InviteeBindingRewardValidityDays: bindingRewards.InviteeValidityDays,
	}, nil
}

func (s *AffiliateService) RegenerateAffiliateCode(ctx context.Context, userID string) (*AffiliateCodeRegenerationResult, error) {
	if !s.IsEnabled(ctx) {
		return nil, infraerrors.NotFound("FEATURE_DISABLED", "feature is disabled")
	}
	if userID == "" {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}

	now := appTimezone.Now()
	periodStart := appTimezone.StartOfMonth(now)
	resetAt := periodStart.AddDate(0, 1, 0)
	code, used, err := s.repo.RegenerateUserAffCode(
		ctx,
		userID,
		periodStart,
		AffiliateCodeRegenerationMonthlyLimit,
	)
	if err != nil {
		if errors.Is(err, ErrAffiliateCodeRegenerationLimit) {
			return nil, ErrAffiliateCodeRegenerationLimit.WithMetadata(map[string]string{
				"limit":    strconv.Itoa(AffiliateCodeRegenerationMonthlyLimit),
				"reset_at": resetAt.Format(time.RFC3339),
			})
		}
		return nil, err
	}

	return &AffiliateCodeRegenerationResult{
		AffCode:   code,
		Limit:     AffiliateCodeRegenerationMonthlyLimit,
		Used:      used,
		Remaining: max(AffiliateCodeRegenerationMonthlyLimit-used, 0),
		ResetAt:   resetAt,
	}, nil
}

func affiliateCodeRegenerationUsage(summary *AffiliateSummary, now time.Time) (int, time.Time) {
	periodStart := appTimezone.StartOfMonth(now)
	resetAt := periodStart.AddDate(0, 1, 0)
	if summary == nil || summary.AffCodeRegenerationPeriodStart == nil {
		return 0, resetAt
	}
	storedStart := summary.AffCodeRegenerationPeriodStart.In(appTimezone.Location())
	if storedStart.Year() != periodStart.Year() || storedStart.Month() != periodStart.Month() {
		return 0, resetAt
	}
	return min(max(summary.AffCodeRegenerationCount, 0), AffiliateCodeRegenerationMonthlyLimit), resetAt
}

func (s *AffiliateService) BindInviterByCode(ctx context.Context, userID string, rawCode string) error {
	_, err := s.BindInviterByCodeWithResult(ctx, userID, rawCode)
	return err
}

func (s *AffiliateService) resolveEligibleInviter(ctx context.Context, userID string, rawCode string) (*AffiliateSummary, error) {
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	if code == "" {
		return nil, nil
	}
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	// 总开关关闭时，注册阶段静默忽略 aff 参数（不报错，避免阻断注册流程）
	if !s.IsUserAvailable(ctx) {
		return nil, infraerrors.NotFound("FEATURE_DISABLED", "feature is disabled")
	}
	if !isValidAffiliateCodeFormat(code) {
		return nil, ErrAffiliateCodeInvalid
	}

	selfSummary, err := s.repo.EnsureUserAffiliate(ctx, userID)
	if err != nil {
		return nil, err
	}
	if selfSummary.InviterID != nil {
		if existing, lookupErr := s.repo.GetAffiliateByCode(ctx, code); lookupErr == nil && existing.UserID == *selfSummary.InviterID {
			return existing, nil
		}
		return nil, ErrAffiliateAlreadyBound
	}

	inviterSummary, err := s.repo.GetAffiliateByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrAffiliateProfileNotFound) {
			return nil, ErrAffiliateCodeInvalid
		}
		return nil, err
	}
	if inviterSummary == nil || inviterSummary.UserID == "" || inviterSummary.UserID == userID {
		return nil, ErrAffiliateCodeInvalid
	}
	return inviterSummary, nil
}

func (s *AffiliateService) AccrueInviteRebate(ctx context.Context, inviteeUserID string, baseRechargeAmount float64) (float64, error) {
	return s.AccrueInviteRebateForOrder(ctx, inviteeUserID, baseRechargeAmount, nil)
}

func (s *AffiliateService) AccrueInviteRebateForOrder(ctx context.Context, inviteeUserID string, baseRechargeAmount float64, sourceOrderID *string) (float64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	if inviteeUserID == "" || baseRechargeAmount <= 0 || math.IsNaN(baseRechargeAmount) || math.IsInf(baseRechargeAmount, 0) {
		return 0, nil
	}
	// 总开关关闭时，新充值不再产生返利
	if !s.IsEnabled(ctx) {
		return 0, nil
	}

	inviteeSummary, err := s.repo.EnsureUserAffiliate(ctx, inviteeUserID)
	if err != nil {
		return 0, err
	}
	if inviteeSummary.InviterID == nil || *inviteeSummary.InviterID == "" {
		return 0, nil
	}

	// 加载邀请人 profile，专属比例始终属于该邀请码的邀请活动配置。
	inviterSummary, err := s.repo.EnsureUserAffiliate(ctx, *inviteeSummary.InviterID)
	if err != nil {
		return 0, err
	}
	// 有效期检查：超过返利有效期后不再产生返利
	if s.settingService != nil {
		if durationDays := s.settingService.GetAffiliateRebateDurationDays(ctx); durationDays > 0 {
			if time.Now().After(inviteeSummary.CreatedAt.AddDate(0, 0, durationDays)) {
				return 0, nil
			}
		}
	}

	rebateRatePercent := s.resolveRebateRatePercent(ctx, inviterSummary)
	rebate := roundTo(baseRechargeAmount*(rebateRatePercent/100), 8)
	if rebate <= 0 {
		return 0, nil
	}

	recipientID := *inviteeSummary.InviterID
	if s.rebateRecipient(ctx) == AffiliateRebateRecipientInvitee {
		recipientID = inviteeUserID
	}

	// 单人上限检查：按实际接收方精确截断到剩余额度。
	if s.settingService != nil {
		if perInviteeCap := s.settingService.GetAffiliateRebatePerInviteeCap(ctx); perInviteeCap > 0 {
			existing, err := s.repo.GetAccruedRebateFromInvitee(ctx, recipientID, inviteeUserID)
			if err != nil {
				return 0, err
			}
			if existing >= perInviteeCap {
				return 0, nil
			}
			if remaining := perInviteeCap - existing; rebate > remaining {
				rebate = roundTo(remaining, 8)
			}
		}
	}

	applied, err := s.repo.AccrueQuota(ctx, recipientID, inviteeUserID, rebate, affiliateRebateFreezeHours, sourceOrderID)
	if err != nil {
		return 0, err
	}
	if !applied {
		return 0, nil
	}
	return rebate, nil
}

func (s *AffiliateService) rebateRecipient(ctx context.Context) string {
	if s == nil || s.settingService == nil {
		return AffiliateRebateRecipientDefault
	}
	return s.settingService.GetAffiliateRebateRecipient(ctx)
}

// resolveRebateRatePercent returns the inviter's exclusive rate when set,
// otherwise the global setting value (clamped to [Min, Max]).
func (s *AffiliateService) resolveRebateRatePercent(ctx context.Context, inviter *AffiliateSummary) float64 {
	if inviter != nil && inviter.AffRebateRatePercent != nil {
		v := *inviter.AffRebateRatePercent
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return s.globalRebateRatePercent(ctx)
		}
		return clampAffiliateRebateRate(v)
	}
	return s.globalRebateRatePercent(ctx)
}

// globalRebateRatePercent reads the system-wide rebate rate via SettingService,
// returning the documented default when SettingService is unavailable.
func (s *AffiliateService) globalRebateRatePercent(ctx context.Context) float64 {
	if s == nil || s.settingService == nil {
		return AffiliateRebateRateDefault
	}
	return s.settingService.GetAffiliateRebateRatePercent(ctx)
}

func (s *AffiliateService) TransferAffiliateQuota(ctx context.Context, userID string) (float64, float64, error) {
	if !s.IsEnabled(ctx) {
		return 0, 0, infraerrors.NotFound("FEATURE_DISABLED", "feature is disabled")
	}
	if s == nil || s.repo == nil {
		return 0, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}

	transferred, balance, err := s.repo.TransferQuotaToBalance(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	if transferred > 0 {
		s.invalidateAffiliateCaches(ctx, userID)
	}
	return transferred, balance, nil
}

func (s *AffiliateService) listInvitees(ctx context.Context, inviterID string) ([]AffiliateInvitee, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	invitees, err := s.repo.ListInvitees(ctx, inviterID, affiliateInviteesLimit)
	if err != nil {
		return nil, err
	}
	for i := range invitees {
		invitees[i].Email = maskEmail(invitees[i].Email)
	}
	return invitees, nil
}

func roundTo(v float64, scale int) float64 {
	factor := math.Pow10(scale)
	return math.Round(v*factor) / factor
}

func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	at := strings.Index(email, "@")
	if at <= 0 || at >= len(email)-1 {
		return "***"
	}

	local := email[:at]
	domain := email[at+1:]
	dot := strings.LastIndex(domain, ".")

	maskedLocal := maskSegment(local)
	if dot <= 0 || dot >= len(domain)-1 {
		return maskedLocal + "@" + maskSegment(domain)
	}

	domainName := domain[:dot]
	tld := domain[dot:]
	return maskedLocal + "@" + maskSegment(domainName) + tld
}

func maskSegment(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return "***"
	}
	if len(r) == 1 {
		return string(r[0]) + "***"
	}
	return string(r[0]) + "***"
}

func (s *AffiliateService) invalidateAffiliateCaches(ctx context.Context, userID string) {
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	if s.billingCacheService != nil {
		if err := s.billingCacheService.InvalidateUserBalance(ctx, userID); err != nil {
			logger.LegacyPrintf("service.affiliate", "[Affiliate] Failed to invalidate billing cache for user %v: %v", userID, err)
		}
	}
}

// =========================
// Admin: 专属配置管理
// =========================

// validateExclusiveRate ensures a per-user override is finite and within
// [Min, Max]. nil is always valid (means "clear / fall back to global").
func validateExclusiveRate(ratePercent *float64) error {
	if ratePercent == nil {
		return nil
	}
	v := *ratePercent
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return infraerrors.BadRequest("INVALID_RATE", "invalid rebate rate")
	}
	if v < AffiliateRebateRateMin || v > AffiliateRebateRateMax {
		return infraerrors.BadRequest("INVALID_RATE", "rebate rate out of range")
	}
	return nil
}

// AdminUpdateUserAffCode 管理员改写用户的邀请码（专属邀请码）。
func (s *AffiliateService) AdminUpdateUserAffCode(ctx context.Context, userID string, rawCode string) error {
	if s == nil || s.repo == nil {
		return infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	if !isValidAffiliateCodeFormat(code) {
		return ErrAffiliateCodeInvalid
	}
	return s.repo.UpdateUserAffCode(ctx, userID, code)
}

// AdminResetUserAffCode 重置用户邀请码为系统随机码。
func (s *AffiliateService) AdminResetUserAffCode(ctx context.Context, userID string) (string, error) {
	if s == nil || s.repo == nil {
		return "", infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	return s.repo.ResetUserAffCode(ctx, userID)
}

// AdminSetUserRebateRate 设置/清除用户专属返利比例。ratePercent==nil 表示清除。
func (s *AffiliateService) AdminSetUserRebateRate(ctx context.Context, userID string, ratePercent *float64) error {
	if s == nil || s.repo == nil {
		return infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	if err := validateExclusiveRate(ratePercent); err != nil {
		return err
	}
	return s.repo.SetUserRebateRate(ctx, userID, ratePercent)
}

// AdminBatchSetUserRebateRate 批量设置/清除用户专属返利比例。
func (s *AffiliateService) AdminBatchSetUserRebateRate(ctx context.Context, userIDs []string, ratePercent *float64) error {
	if s == nil || s.repo == nil {
		return infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	if err := validateExclusiveRate(ratePercent); err != nil {
		return err
	}
	cleaned := make([]string, 0, len(userIDs))
	for _, uid := range userIDs {
		if uid != "" {
			cleaned = append(cleaned, uid)
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	return s.repo.BatchSetUserRebateRate(ctx, cleaned, ratePercent)
}

// AdminListCustomUsers 列出有专属配置的用户。
func (s *AffiliateService) AdminListCustomUsers(ctx context.Context, filter AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	return s.repo.ListUsersWithCustomSettings(ctx, filter)
}

func (s *AffiliateService) AdminListInviteRecords(ctx context.Context, filter AffiliateRecordFilter) ([]AffiliateInviteRecord, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	return s.repo.ListAffiliateInviteRecords(ctx, normalizeAffiliateRecordFilter(filter))
}

func (s *AffiliateService) AdminListRebateRecords(ctx context.Context, filter AffiliateRecordFilter) ([]AffiliateRebateRecord, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	return s.repo.ListAffiliateRebateRecords(ctx, normalizeAffiliateRecordFilter(filter))
}

func (s *AffiliateService) AdminListTransferRecords(ctx context.Context, filter AffiliateRecordFilter) ([]AffiliateTransferRecord, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	return s.repo.ListAffiliateTransferRecords(ctx, normalizeAffiliateRecordFilter(filter))
}

func (s *AffiliateService) AdminGetUserOverview(ctx context.Context, userID string) (*AffiliateUserOverview, error) {
	if userID == "" {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	overview, err := s.repo.GetAffiliateUserOverview(ctx, userID)
	if err != nil {
		return nil, err
	}
	if overview != nil {
		if !overview.RebateRateCustom {
			overview.RebateRatePercent = s.globalRebateRatePercent(ctx)
		}
		overview.RebateRatePercent = clampAffiliateRebateRate(overview.RebateRatePercent)
	}
	return overview, nil
}

func normalizeAffiliateRecordFilter(filter AffiliateRecordFilter) AffiliateRecordFilter {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	filter.Search = strings.TrimSpace(filter.Search)
	filter.SortBy = strings.TrimSpace(filter.SortBy)
	return filter
}
