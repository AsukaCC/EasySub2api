package schema

import (
	"time"

	"github.com/AsukaCC/EasySub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PendingSubscription stores a granted subscription that is waiting for the
// current subscription on the same platform to end.
type PendingSubscription struct {
	ent.Schema
}

func (PendingSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "pending_subscriptions",
			Checks: map[string]string{
				"pending_subscriptions_status_valid":      "status IN ('PENDING', 'ACTIVATED', 'CANCELLED')",
				"pending_subscriptions_validity_positive": "validity_days > 0",
			},
		},
	}
}

func (PendingSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.UUIDv7IDMixin{}}
}

func (PendingSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").SchemaType(postgresUUIDSchema),
		field.String("group_id").SchemaType(postgresUUIDSchema),
		field.String("platform").MaxLen(50),
		field.Int("validity_days"),
		field.String("source_type").MaxLen(32),
		field.String("source_id").MaxLen(128).Default(""),
		field.String("blocked_by_subscription_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.Time("expected_activation_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("status").MaxLen(16).Default("PENDING"),
		field.String("activated_subscription_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("activation_mode").MaxLen(16).Default(""),
		field.JSON("forfeited_subscription_ids", []string{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("activated_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("cancelled_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("last_error").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.String("assigned_by").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("notes").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (PendingSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "platform").Unique().Annotations(entsql.IndexWhere("status = 'PENDING'")),
		index.Fields("source_type", "source_id").Unique().Annotations(entsql.IndexWhere("source_id <> ''")),
		index.Fields("status", "expected_activation_at"),
		index.Fields("group_id"),
	}
}
