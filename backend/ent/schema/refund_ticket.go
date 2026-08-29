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

// RefundTicket represents a manually reviewed, out-of-window refund request.
type RefundTicket struct {
	ent.Schema
}

func (RefundTicket) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{
		Table: "refund_tickets",
		Checks: map[string]string{
			"refund_tickets_status_valid":                   "status IN ('PENDING', 'APPROVED', 'PROCESSING', 'COMPLETED', 'REJECTED', 'CANCELLED', 'FAILED')",
			"refund_tickets_approved_principal_nonnegative": "approved_principal_amount IS NULL OR approved_principal_amount >= 0",
			"refund_tickets_affiliate_action_valid":         "affiliate_action = 'MANUAL'",
		},
	}}
}

func (RefundTicket) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.UUIDv7IDMixin{}}
}

func (RefundTicket) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_id").SchemaType(postgresUUIDSchema),
		field.String("user_id").SchemaType(postgresUUIDSchema),
		field.String("refund_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("status").MaxLen(24).Default("PENDING"),
		field.String("comment").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.Float("approved_principal_amount").SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).Optional().Nillable(),
		field.String("reviewer_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("review_note").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.String("affiliate_action").MaxLen(24).Default("MANUAL"),
		field.Time("reviewed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (RefundTicket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id").StorageKey("idx_refund_tickets_order"),
		index.Fields("order_id").
			Unique().
			StorageKey("idx_refund_tickets_one_active_order").
			Annotations(entsql.IndexWhere("status IN ('PENDING', 'APPROVED', 'PROCESSING')")),
		index.Fields("user_id", "created_at"),
		index.Fields("status", "created_at"),
	}
}
