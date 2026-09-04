package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/domain"
	"github.com/AsukaCC/EasySub2api/internal/model"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const AccountModelRuleGenericTierFilter = "__all__"

var ErrAccountModelRuleExists = infraerrors.Conflict(
	"ACCOUNT_MODEL_RULE_EXISTS",
	"an account model rule with this name already exists for the platform and subscription tier",
)

type AccountModelRuleRepository interface {
	List(ctx context.Context, platform, subscriptionTier string) ([]*model.AccountModelRule, error)
	GetByID(ctx context.Context, id string) (*model.AccountModelRule, error)
	GetByScopeAndName(ctx context.Context, platform string, subscriptionTier *string, name string) (*model.AccountModelRule, error)
	SubscriptionTierExists(ctx context.Context, platform, subscriptionTier string) (bool, error)
	CountBoundAccounts(ctx context.Context, id string) (int, error)
	Create(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error)
	Update(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error)
	Delete(ctx context.Context, id string) error
}

type AccountModelRuleService struct {
	repo AccountModelRuleRepository
}

func NewAccountModelRuleService(repo AccountModelRuleRepository) *AccountModelRuleService {
	return &AccountModelRuleService{repo: repo}
}

func (s *AccountModelRuleService) List(ctx context.Context, platform, subscriptionTier string) ([]*model.AccountModelRule, error) {
	platform = strings.TrimSpace(strings.ToLower(platform))
	if platform != "" && !domain.IsAccountPlatform(platform) {
		return nil, &model.ValidationError{Field: "platform", Message: "unsupported account platform"}
	}
	subscriptionTier = strings.TrimSpace(subscriptionTier)
	if subscriptionTier != "" && subscriptionTier != AccountModelRuleGenericTierFilter {
		subscriptionTier = model.NormalizeSubscriptionTier(subscriptionTier)
		if subscriptionTier == "" {
			return nil, &model.ValidationError{Field: "subscription_tier", Message: "invalid subscription tier"}
		}
	}
	return s.repo.List(ctx, platform, subscriptionTier)
}

func (s *AccountModelRuleService) GetByID(ctx context.Context, id string) (*model.AccountModelRule, error) {
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil || rule == nil {
		return rule, err
	}
	rule.BoundAccountCount, err = s.repo.CountBoundAccounts(ctx, id)
	return rule, err
}

func (s *AccountModelRuleService) Create(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	if err := s.validateSpecificTier(ctx, rule); err != nil {
		return nil, err
	}
	if existing, err := s.repo.GetByScopeAndName(ctx, rule.Platform, rule.SubscriptionTier, rule.Name); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, ErrAccountModelRuleExists
	}
	return s.repo.Create(ctx, rule)
}

func (s *AccountModelRuleService) Update(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetByID(ctx, rule.ID)
	if err != nil || existing == nil {
		return existing, err
	}
	boundCount, err := s.repo.CountBoundAccounts(ctx, rule.ID)
	if err != nil {
		return nil, err
	}
	if boundCount > 0 && (existing.Platform != rule.Platform || !sameOptionalTier(existing.SubscriptionTier, rule.SubscriptionTier)) {
		return nil, infraerrors.Conflict(
			"ACCOUNT_MODEL_RULE_SCOPE_BOUND",
			"platform and subscription tier cannot be changed while accounts are bound",
		).WithMetadata(map[string]string{"bound_account_count": strconv.Itoa(boundCount)})
	}
	if err := s.validateSpecificTier(ctx, rule); err != nil {
		return nil, err
	}
	other, err := s.repo.GetByScopeAndName(ctx, rule.Platform, rule.SubscriptionTier, rule.Name)
	if err != nil {
		return nil, err
	}
	if other != nil && other.ID != rule.ID {
		return nil, ErrAccountModelRuleExists
	}
	updated, err := s.repo.Update(ctx, rule)
	if updated != nil {
		updated.BoundAccountCount = boundCount
	}
	return updated, err
}

func (s *AccountModelRuleService) Delete(ctx context.Context, id string) error {
	boundCount, err := s.repo.CountBoundAccounts(ctx, id)
	if err != nil {
		return err
	}
	if boundCount > 0 {
		return infraerrors.Conflict(
			"ACCOUNT_MODEL_RULE_BOUND",
			"account model rule is still bound to accounts",
		).WithMetadata(map[string]string{"bound_account_count": strconv.Itoa(boundCount)})
	}
	return s.repo.Delete(ctx, id)
}

func (s *AccountModelRuleService) ValidateBinding(ctx context.Context, id, platform, subscriptionTier string) (*model.AccountModelRule, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, infraerrors.New(http.StatusBadRequest, "ACCOUNT_MODEL_RULE_NOT_FOUND", "account model rule not found")
	}
	if rule.Platform != strings.ToLower(strings.TrimSpace(platform)) {
		return nil, infraerrors.BadRequest("ACCOUNT_MODEL_RULE_PLATFORM_MISMATCH", "account model rule platform does not match account platform")
	}
	if rule.SubscriptionTier != nil && *rule.SubscriptionTier != model.NormalizeSubscriptionTier(subscriptionTier) {
		return nil, infraerrors.BadRequest("ACCOUNT_MODEL_RULE_TIER_MISMATCH", "account model rule subscription tier does not match account subscription tier")
	}
	return rule, nil
}

func (s *AccountModelRuleService) validateSpecificTier(ctx context.Context, rule *model.AccountModelRule) error {
	if rule.SubscriptionTier == nil {
		return nil
	}
	exists, err := s.repo.SubscriptionTierExists(ctx, rule.Platform, *rule.SubscriptionTier)
	if err != nil {
		return err
	}
	if !exists {
		return infraerrors.BadRequest("ACCOUNT_MODEL_RULE_TIER_NOT_FOUND", "subscription tier is not present on any current account for this platform")
	}
	return nil
}

func sameOptionalTier(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
