package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const AffiliateBindingRewardVersion = 1

var (
	ErrAffiliateInviteCycle          = infraerrors.Conflict("AFFILIATE_INVITE_CYCLE", "invitation relationship would create a cycle")
	ErrAffiliateInviterUnavailable   = infraerrors.BadRequest("AFFILIATE_INVITER_UNAVAILABLE", "inviter is unavailable")
	ErrAffiliateBackfillPreviewStale = infraerrors.Conflict("AFFILIATE_BACKFILL_PREVIEW_STALE", "affiliate reward backfill preview is stale")
	ErrAffiliateBackfillRunning      = infraerrors.Conflict("AFFILIATE_BACKFILL_RUNNING", "an affiliate reward backfill is already running")
)

type AffiliateBindingRewardConfig struct {
	InviterPoints       float64 `json:"inviter_points"`
	InviterValidityDays int     `json:"inviter_validity_days"`
	InviteePoints       float64 `json:"invitee_points"`
	InviteeValidityDays int     `json:"invitee_validity_days"`
}

func DefaultAffiliateBindingRewardConfig() AffiliateBindingRewardConfig {
	return AffiliateBindingRewardConfig{
		InviterValidityDays: AffiliateBindingRewardValidityDefault,
		InviteeValidityDays: AffiliateBindingRewardValidityDefault,
	}
}

type AffiliateBindingReward struct {
	Points    float64    `json:"points"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Applied   bool       `json:"applied"`
}

type AffiliateBindingResult struct {
	Bound         bool                   `json:"bound"`
	AlreadyBound  bool                   `json:"already_bound"`
	InviterReward AffiliateBindingReward `json:"inviter_reward"`
	InviteeReward AffiliateBindingReward `json:"invitee_reward"`
}

type AffiliateRewardBackfillPreview struct {
	Config                 AffiliateBindingRewardConfig `json:"config"`
	EligibleRelations      int                          `json:"eligible_relations"`
	EstimatedInviterGrants int                          `json:"estimated_inviter_grants"`
	EstimatedInviteeGrants int                          `json:"estimated_invitee_grants"`
	EstimatedInviterPoints float64                      `json:"estimated_inviter_points"`
	EstimatedInviteePoints float64                      `json:"estimated_invitee_points"`
	PreviewToken           string                       `json:"preview_token"`
}

type AffiliateRewardBackfillRun struct {
	ID                   string                       `json:"id"`
	Status               string                       `json:"status"`
	Config               AffiliateBindingRewardConfig `json:"config"`
	EligibleRelations    int                          `json:"eligible_relations"`
	ProcessedRelations   int                          `json:"processed_relations"`
	InviterGrants        int                          `json:"inviter_grants"`
	InviteeGrants        int                          `json:"invitee_grants"`
	InviterPointsGranted float64                      `json:"inviter_points_granted"`
	InviteePointsGranted float64                      `json:"invitee_points_granted"`
	ErrorMessage         string                       `json:"error_message,omitempty"`
	CreatedAt            time.Time                    `json:"created_at"`
	StartedAt            *time.Time                   `json:"started_at,omitempty"`
	CompletedAt          *time.Time                   `json:"completed_at,omitempty"`
	UpdatedAt            time.Time                    `json:"updated_at"`
}

type affiliateRewardRepository interface {
	BindInviterWithRewards(context.Context, string, string, AffiliateBindingRewardConfig) (*AffiliateBindingResult, error)
	LegacyBindingRewardScope(context.Context) (int, string, error)
	CreateRewardBackfillRun(context.Context, string, string, AffiliateBindingRewardConfig, int) (*AffiliateRewardBackfillRun, error)
	GetRewardBackfillRun(context.Context, string) (*AffiliateRewardBackfillRun, error)
	GetActiveRewardBackfillRun(context.Context) (*AffiliateRewardBackfillRun, error)
	ProcessRewardBackfillBatch(context.Context, string, int) (int, bool, error)
	FailRewardBackfillRun(context.Context, string, string) error
}

var activeAffiliateBackfills sync.Map

func affiliateBackfillToken(config AffiliateBindingRewardConfig, count int, scopeHash string) string {
	payload := fmt.Sprintf("%.8f:%d:%.8f:%d:%d:%s", config.InviterPoints, config.InviterValidityDays, config.InviteePoints, config.InviteeValidityDays, count, scopeHash)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func (s *AffiliateService) bindingRewardConfig(ctx context.Context) AffiliateBindingRewardConfig {
	if s == nil || s.settingService == nil {
		return DefaultAffiliateBindingRewardConfig()
	}
	return s.settingService.GetAffiliateBindingRewardConfig(ctx)
}

func (s *AffiliateService) BindInviterByCodeWithResult(ctx context.Context, userID, rawCode string) (*AffiliateBindingResult, error) {
	if strings.TrimSpace(rawCode) == "" {
		return &AffiliateBindingResult{}, nil
	}
	inviter, err := s.resolveEligibleInviter(ctx, userID, rawCode)
	if err != nil {
		return nil, err
	}
	repo, ok := s.repo.(affiliateRewardRepository)
	if !ok {
		bound, bindErr := s.repo.BindInviter(ctx, userID, inviter.UserID)
		return &AffiliateBindingResult{Bound: bound, AlreadyBound: !bound}, bindErr
	}
	result, err := repo.BindInviterWithRewards(ctx, userID, inviter.UserID, s.bindingRewardConfig(ctx))
	if err != nil {
		return nil, err
	}
	s.invalidateAffiliateCaches(ctx, userID)
	s.invalidateAffiliateCaches(ctx, inviter.UserID)
	return result, nil
}

func (s *AffiliateService) AdminPreviewRewardBackfill(ctx context.Context) (*AffiliateRewardBackfillPreview, error) {
	repo, ok := s.repo.(affiliateRewardRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate reward repository unavailable")
	}
	count, scopeHash, err := repo.LegacyBindingRewardScope(ctx)
	if err != nil {
		return nil, err
	}
	config := s.bindingRewardConfig(ctx)
	preview := &AffiliateRewardBackfillPreview{
		Config: config, EligibleRelations: count,
		EstimatedInviterPoints: QuantizeUsageBillingAmount(float64(count) * config.InviterPoints),
		EstimatedInviteePoints: QuantizeUsageBillingAmount(float64(count) * config.InviteePoints),
	}
	if config.InviterPoints > 0 {
		preview.EstimatedInviterGrants = count
	}
	if config.InviteePoints > 0 {
		preview.EstimatedInviteeGrants = count
	}
	preview.PreviewToken = affiliateBackfillToken(config, count, scopeHash)
	return preview, nil
}

func (s *AffiliateService) AdminStartRewardBackfill(ctx context.Context, actorID, previewToken string) (*AffiliateRewardBackfillRun, error) {
	if !s.IsEnabled(ctx) {
		return nil, infraerrors.Conflict("FEATURE_DISABLED", "affiliate feature must be enabled before backfill")
	}
	preview, err := s.AdminPreviewRewardBackfill(ctx)
	if err != nil {
		return nil, err
	}
	if preview.PreviewToken != previewToken {
		return nil, ErrAffiliateBackfillPreviewStale
	}
	repo, ok := s.repo.(affiliateRewardRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate reward repository unavailable")
	}
	run, err := repo.CreateRewardBackfillRun(ctx, actorID, previewToken, preview.Config, preview.EligibleRelations)
	if err != nil {
		return nil, err
	}
	s.resumeRewardBackfill(run.ID)
	return run, nil
}

func (s *AffiliateService) AdminGetRewardBackfill(ctx context.Context, runID string) (*AffiliateRewardBackfillRun, error) {
	repo, ok := s.repo.(affiliateRewardRepository)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate reward repository unavailable")
	}
	run, err := repo.GetRewardBackfillRun(ctx, runID)
	if err == nil && (run.Status == "pending" || run.Status == "running") {
		s.resumeRewardBackfill(run.ID)
	}
	return run, err
}

func (s *AffiliateService) resumeRewardBackfill(runID string) {
	if _, loaded := activeAffiliateBackfills.LoadOrStore(runID, struct{}{}); loaded {
		return
	}
	go func() {
		defer activeAffiliateBackfills.Delete(runID)
		repo, ok := s.repo.(affiliateRewardRepository)
		if !ok {
			return
		}
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			processed, done, err := repo.ProcessRewardBackfillBatch(ctx, runID, 100)
			cancel()
			if err != nil {
				failCtx, failCancel := context.WithTimeout(context.Background(), 5*time.Second)
				_ = repo.FailRewardBackfillRun(failCtx, runID, err.Error())
				failCancel()
				return
			}
			if done || processed == 0 {
				return
			}
		}
	}()
}

func (s *AffiliateService) resumePersistedRewardBackfill() {
	repo, ok := s.repo.(affiliateRewardRepository)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	run, err := repo.GetActiveRewardBackfillRun(ctx)
	cancel()
	if err == nil && run != nil {
		s.resumeRewardBackfill(run.ID)
	}
}
