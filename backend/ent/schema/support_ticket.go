package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/AsukaCC/EasySub2api/ent/schema/mixins"
)

// SupportTicket represents a user support conversation and its optional refund workflow.
type SupportTicket struct {
	ent.Schema
}

func (SupportTicket) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{
		Table: "support_tickets",
		Checks: map[string]string{
			"support_tickets_category_valid":        "category IN ('ACCOUNT', 'REFUND')",
			"support_tickets_status_valid":          "status IN ('PENDING_ADMIN', 'PENDING_USER', 'IN_PROGRESS', 'RESOLVED', 'CLOSED', 'CANCELLED')",
			"support_tickets_origin_valid":          "origin IN ('USER', 'ADMIN', 'MIGRATED')",
			"support_tickets_refund_decision_valid": "refund_decision IN ('NONE', 'PENDING', 'APPROVED', 'REJECTED')",
			"support_tickets_reopen_count_valid":    "reopen_count >= 0 AND reopen_count <= 1",
			"support_tickets_approved_nonnegative":  "approved_principal_amount IS NULL OR approved_principal_amount >= 0",
		},
	}}
}

func (SupportTicket) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.UUIDv7IDMixin{}}
}

func (SupportTicket) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").SchemaType(postgresUUIDSchema),
		field.String("category").MaxLen(24),
		field.String("status").MaxLen(24).Default("PENDING_ADMIN"),
		field.String("origin").MaxLen(16).Default("USER"),
		field.String("title").MaxLen(120),
		field.String("api_key_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("api_key_name_snapshot").MaxLen(160).Default(""),
		field.String("group_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("group_name_snapshot").MaxLen(160).Default(""),
		field.String("order_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("refund_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("refund_decision").MaxLen(16).Default("NONE"),
		field.Float("approved_principal_amount").SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).Optional().Nillable(),
		field.String("reviewer_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.Time("reviewed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int("reopen_count").Default(0),
		field.Time("resolved_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("closed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("last_user_activity_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("last_admin_activity_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("status", "created_at"),
		index.Fields("category", "created_at"),
		index.Fields("order_id").StorageKey("idx_support_tickets_order"),
		index.Fields("order_id").Unique().StorageKey("idx_support_tickets_one_active_refund").
			Annotations(entsql.IndexWhere("category = 'REFUND' AND order_id IS NOT NULL AND status IN ('PENDING_ADMIN', 'PENDING_USER', 'IN_PROGRESS')")),
	}
}
