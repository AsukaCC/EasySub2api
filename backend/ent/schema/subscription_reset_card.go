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

// SubscriptionResetCard is one independently expiring weekly-quota reset
// entitlement attached to a concrete user subscription.
type SubscriptionResetCard struct {
	ent.Schema
}

func (SubscriptionResetCard) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "subscription_reset_cards",
			Checks: map[string]string{
				"subscription_reset_cards_status_valid": "status IN ('available', 'consumed', 'expired', 'revoked')",
			},
		},
	}
}

func (SubscriptionResetCard) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.UUIDv7IDMixin{}, mixins.TimeMixin{}}
}

func (SubscriptionResetCard) Fields() []ent.Field {
	return []ent.Field{
		field.String("subscription_id").SchemaType(postgresUUIDSchema),
		field.String("user_id").SchemaType(postgresUUIDSchema),
		field.String("group_id").SchemaType(postgresUUIDSchema),
		field.String("status").MaxLen(16).Default("available"),
		field.Time("issued_at").SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("consumed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("issued_by").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("consumed_by").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("source_type").MaxLen(40).Default("manual"),
		field.String("source_id").MaxLen(128).Default(""),
		field.String("grant_batch_id").SchemaType(postgresUUIDSchema),
		field.Int("grant_index"),
	}
}

func (SubscriptionResetCard) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subscription_id", "status", "expires_at"),
		index.Fields("user_id", "status", "expires_at"),
		index.Fields("expires_at", "status"),
		index.Fields("source_type", "source_id", "subscription_id", "grant_index").Unique(),
		index.Fields("grant_batch_id", "grant_index").Unique(),
	}
}
