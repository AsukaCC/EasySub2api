package repository

import (
	"context"

	"github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/accountmodelrule"
	"github.com/AsukaCC/EasySub2api/internal/model"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

type accountModelRuleRepository struct {
	client *ent.Client
}

func NewAccountModelRuleRepository(client *ent.Client) service.AccountModelRuleRepository {
	return &accountModelRuleRepository{client: client}
}

func (r *accountModelRuleRepository) List(ctx context.Context, platform string) ([]*model.AccountModelRule, error) {
	query := r.client.AccountModelRule.Query().
		Order(ent.Asc(accountmodelrule.FieldPlatform), ent.Asc(accountmodelrule.FieldName))
	if platform != "" {
		query = query.Where(accountmodelrule.PlatformEQ(platform))
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*model.AccountModelRule, len(rows))
	for i, row := range rows {
		result[i] = accountModelRuleToModel(row)
	}
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

func (r *accountModelRuleRepository) GetByPlatformAndName(ctx context.Context, platform, name string) (*model.AccountModelRule, error) {
	row, err := r.client.AccountModelRule.Query().
		Where(
			accountmodelrule.PlatformEQ(platform),
			accountmodelrule.NameEQ(name),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return accountModelRuleToModel(row), nil
}

func (r *accountModelRuleRepository) Create(ctx context.Context, rule *model.AccountModelRule) (*model.AccountModelRule, error) {
	builder := r.client.AccountModelRule.Create().
		SetName(rule.Name).
		SetPlatform(rule.Platform).
		SetWhitelist(rule.Whitelist).
		SetMapping(rule.Mapping).
		SetReasoningEfforts(rule.ReasoningEfforts)
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
		SetWhitelist(rule.Whitelist).
		SetMapping(rule.Mapping).
		SetReasoningEfforts(rule.ReasoningEfforts)
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
	mapping := make(map[string]string, len(row.Mapping))
	for from, to := range row.Mapping {
		mapping[from] = to
	}
	reasoningEfforts := make(map[string]string, len(row.ReasoningEfforts))
	for modelName, effort := range row.ReasoningEfforts {
		reasoningEfforts[modelName] = effort
	}
	return &model.AccountModelRule{
		ID:               row.ID,
		Name:             row.Name,
		Description:      row.Description,
		Platform:         row.Platform,
		Whitelist:        append([]string(nil), row.Whitelist...),
		Mapping:          mapping,
		ReasoningEfforts: reasoningEfforts,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
