// Package schema 定义 Ent ORM 的数据库 schema。
package schema

import (
	"github.com/AsukaCC/EasySub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AccountModelRule is a reusable, platform-scoped model restriction template.
// Applying a rule copies its mapping into an account; there is deliberately no
// edge to accounts so rule edits/deletes never mutate existing credentials.
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
		field.JSON("mapping", map[string]string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("whitelist", []string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (AccountModelRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform"),
		index.Fields("platform", "name").Unique(),
	}
}
