package repository

import (
	"context"
	"sort"
	"strings"

	"github.com/AsukaCC/EasySub2api/ent"
	dbaccount "github.com/AsukaCC/EasySub2api/ent/account"
	"github.com/AsukaCC/EasySub2api/ent/accountmodelrule"
	"github.com/AsukaCC/EasySub2api/internal/domain"
	"github.com/AsukaCC/EasySub2api/internal/model"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

type accountModelRuleRepository struct {
	client *ent.Client
}

func NewAccountModelRuleRepository(client *ent.Client) service.AccountModelRuleRepository {
	return &accountModelRuleRepository{client: client}
}

func (r *accountModelRuleRepository) List(ctx context.Context, platform, subscriptionTier string) ([]*model.AccountModelRule, error) {
	query := r.client.AccountModelRule.Query()
	if platform != "" {
		query = query.Where(accountmodelrule.PlatformEQ(platform))
	}
	switch subscriptionTier {
	case service.AccountModelRuleGenericTierFilter:
		query = query.Where(accountmodelrule.SubscriptionTierIsNil())
	case "":
	default:
		query = query.Where(accountmodelrule.Or(
			accountmodelrule.SubscriptionTierEQ(subscriptionTier),
			accountmodelrule.SubscriptionTierIsNil(),
		))
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*model.AccountModelRule, len(rows))
	for i, row := range rows {
		result[i] = accountModelRuleToModel(row)
		result[i].BoundAccountCount, err = r.CountBoundAccounts(ctx, row.ID)
		if err != nil {
			return nil, err
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Platform != right.Platform {
			return left.Platform < right.Platform
		}
		leftExact := left.SubscriptionTier != nil
		rightExact := right.SubscriptionTier != nil
		if leftExact != rightExact {
			return leftExact
		}
		if leftExact && *left.SubscriptionTier != *right.SubscriptionTier {
			return *left.SubscriptionTier < *right.SubscriptionTier
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
	return result, nil
}

func (r *accountModelRuleRepository) GetByID(ctx context.Context, id string) (*model.AccountModelRule, error) {
	row, err := r.client.AccountModelRule.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return accountModelRuleToModel(row), nil
}

func (r *accountModelRuleRepository) GetByScopeAndName(ctx context.Context, platform string, subscriptionTier *string, name string) (*model.AccountModelRule, error) {
	query := r.client.AccountModelRule.Query().Where(
		accountmodelrule.PlatformEQ(platform),
		accountmodelrule.NameEQ(name),
	)
	if subscriptionTier == nil {
		query = query.Where(accountmodelrule.SubscriptionTierIsNil())
	} else {
		query = query.Where(accountmodelrule.SubscriptionTierEQ(*subscriptionTier))
	}
	row, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return accountModelRuleToModel(row), nil
}

func (r *accountModelRuleRepository) SubscriptionTierExists(ctx context.Context, platform, subscriptionTier string) (bool, error) {
	return r.client.Account.Query().Where(
		dbaccount.PlatformEQ(platform),
		dbaccount.SubscriptionTierEQ(subscriptionTier),
	).Exist(ctx)
}

func (r *accountModelRuleRepository) CountBoundAccounts(ctx context.Context, id string) (int, error) {
	return r.client.Account.Query().Where(dbaccount.ModelRuleIDEQ(id)).Count(ctx)
}

func (r *accountModelRuleRepository) Create(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error) {
	builder := r.client.AccountModelRule.Create().
		SetName(rule.Name).
		SetPlatform(rule.Platform).
		SetNillableSubscriptionTier(rule.SubscriptionTier).
		SetModelRoutes(rule.ModelRoutes).
		SetWhitelist([]string{}).
		SetMapping(map[string]string{}).
		SetReasoningEfforts(map[string]string{})
	if rule.Description != nil {
		builder.SetDescription(*rule.Description)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, service.ErrAccountModelRuleExists
		}
		return nil, err
	}
	return accountModelRuleToModel(row), nil
}

func (r *accountModelRuleRepository) Update(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error) {
	builder := r.client.AccountModelRule.UpdateOneID(rule.ID).
		SetName(rule.Name).
		SetPlatform(rule.Platform).
		SetNillableSubscriptionTier(rule.SubscriptionTier).
		SetModelRoutes(rule.ModelRoutes)
	if rule.SubscriptionTier == nil {
		builder.ClearSubscriptionTier()
	}
	if rule.Description != nil {
		builder.SetDescription(*rule.Description)
	} else {
		builder.ClearDescription()
	}
	row, err := builder.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		if ent.IsConstraintError(err) {
			return nil, service.ErrAccountModelRuleExists
		}
		return nil, err
	}
	accountIDs, err := r.client.Account.Query().Where(dbaccount.ModelRuleIDEQ(rule.ID)).Select(dbaccount.FieldID).Strings(ctx)
	if err != nil {
		return nil, err
	}
	for start := 0; start < len(accountIDs); start += postgresParameterBatchSize {
		end := start + postgresParameterBatchSize
		if end > len(accountIDs) {
			end = len(accountIDs)
		}
		if err := enqueueSchedulerOutbox(ctx, r.client, service.SchedulerOutboxEventAccountBulkChanged, nil, nil, map[string]any{"account_ids": accountIDs[start:end]}); err != nil {
			return nil, err
		}
	}
	return accountModelRuleToModel(row), nil
}

func (r *accountModelRuleRepository) Delete(ctx context.Context, id string) error {
	if err := r.client.AccountModelRule.DeleteOneID(id).Exec(ctx); err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func accountModelRuleToModel(row *ent.AccountModelRule) *model.AccountModelRule {
	return &model.AccountModelRule{
		ID:               row.ID,
		Name:             row.Name,
		Description:      row.Description,
		Platform:         row.Platform,
		SubscriptionTier: row.SubscriptionTier,
		ModelRoutes:      append([]domain.AccountModelRoute{}, row.ModelRoutes...),
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
