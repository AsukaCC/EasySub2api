// Package schema 定义 Ent ORM 的数据库 schema。
package schema

import (
	"github.com/AsukaCC/EasySub2api/ent/schema/mixins"
	"github.com/AsukaCC/EasySub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AccountModelRule is a reusable platform/tier-scoped model routing rule.
type AccountModelRule struct {
	ent.Schema
}

func (AccountModelRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "account_model_rules"},
	}
}

func (AccountModelRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.UUIDv7IDMixin{},
		mixins.TimeMixin{},
	}
}

func (AccountModelRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("platform").
			MaxLen(50).
			NotEmpty(),
		field.String("subscription_tier").
			MaxLen(100).
			Optional().
			Nillable(),
		field.JSON("model_routes", []domain.AccountModelRoute{}).
			Default([]domain.AccountModelRoute{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("mapping", map[string]string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("reasoning_efforts", map[string]string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("whitelist", []string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (AccountModelRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform"),
		index.Fields("platform", "subscription_tier"),
	}
}
