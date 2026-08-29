package service

import (
	"context"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/domain"
	"github.com/AsukaCC/EasySub2api/internal/model"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

var ErrAccountModelRuleExists = infraerrors.Conflict(
	"ACCOUNT_MODEL_RULE_EXISTS",
	"an account model rule with this name already exists for the platform",
)

// AccountModelRuleRepository defines persistence for reusable account rules.
type AccountModelRuleRepository interface {
	List(ctx context.Context, platform string) ([]*model.AccountModelRule, error)
	GetByID(ctx context.Context, id string) (*model.AccountModelRule, error)
	GetByPlatformAndName(ctx context.Context, platform, name string) (*model.AccountModelRule, error)
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

func (s *AccountModelRuleService) List(ctx context.Context, platform string) ([]*model.AccountModelRule, error) {
	platform = strings.TrimSpace(strings.ToLower(platform))
	if platform != "" {
		if !domain.IsAccountPlatform(platform) {
			return nil, &model.ValidationError{Field: "platform", Message: "unsupported account platform"}
		}
	}
	return s.repo.List(ctx, platform)
}

func (s *AccountModelRuleService) GetByID(ctx context.Context, id string) (*model.AccountModelRule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AccountModelRuleService) Create(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	if existing, err := s.repo.GetByPlatformAndName(ctx, rule.Platform, rule.Name); err != nil {
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
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}
	other, err := s.repo.GetByPlatformAndName(ctx, rule.Platform, rule.Name)
	if err != nil {
		return nil, err
	}
	if other != nil && other.ID != rule.ID {
		return nil, ErrAccountModelRuleExists
	}
	return s.repo.Update(ctx, rule)
}

func (s *AccountModelRuleService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
