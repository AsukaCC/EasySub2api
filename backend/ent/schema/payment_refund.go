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

// PaymentRefund records one idempotent refund attempt and its point recovery.
type PaymentRefund struct {
	ent.Schema
}

func (PaymentRefund) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{
		Table: "payment_refunds",
		Checks: map[string]string{
			"payment_refunds_currency_valid":             "currency = 'CNY'",
			"payment_refunds_amounts_nonnegative":        "requested_principal_amount >= 0 AND principal_amount >= 0 AND fee_amount >= 0 AND gateway_amount >= 0 AND base_points >= 0 AND bonus_points >= 0 AND affiliate_rebate_points >= 0 AND bonus_expired_offset >= 0",
			"payment_refunds_gateway_split_valid":        "gateway_amount = principal_amount + fee_amount",
			"payment_refunds_bonus_expired_offset_valid": "bonus_expired_offset <= bonus_points",
			"payment_refunds_targets_nonnegative":        "target_principal_amount >= 0 AND target_fee_amount >= 0 AND target_base_points >= 0 AND target_bonus_points >= 0 AND target_affiliate_points >= 0",
			"payment_refunds_status_valid":               "status IN ('REQUESTED', 'RESERVED', 'SUBMITTING', 'PENDING', 'SUCCEEDED', 'FAILED')",
			"payment_refunds_source_valid":               "source IN ('SELF_SERVICE', 'TICKET', 'ADMIN')",
		},
	}}
}

func (PaymentRefund) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.UUIDv7IDMixin{}}
}

func (PaymentRefund) Fields() []ent.Field {
	money := func(name string) ent.Field {
		return field.Float(name).SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).Default(0)
	}
	points := func(name string) ent.Field {
		return field.Float(name).SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0)
	}
	return []ent.Field{
		field.String("order_id").SchemaType(postgresUUIDSchema),
		field.String("user_id").SchemaType(postgresUUIDSchema),
		field.String("ticket_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("source").MaxLen(24).Default("SELF_SERVICE"),
		field.String("status").MaxLen(30).Default("REQUESTED"),
		field.String("idempotency_key").MaxLen(160),
		field.String("request_fingerprint").MaxLen(160).Default(""),
		field.String("provider_request_id").MaxLen(128),
		field.String("provider_refund_id").MaxLen(128).Optional().Nillable(),
		field.String("currency").MaxLen(3).Default("CNY"),
		money("requested_principal_amount"),
		money("principal_amount"),
		money("fee_amount"),
		money("gateway_amount"),
		points("base_points"),
		points("bonus_points"),
		points("affiliate_rebate_points"),
		points("bonus_expired_offset"),
		money("target_principal_amount"),
		money("target_fee_amount"),
		points("target_base_points"),
		points("target_bonus_points"),
		points("target_affiliate_points"),
		field.String("wallet_hold_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("requested_by").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("reason").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.String("error_code").MaxLen(80).Default(""),
		field.String("error_message").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.Time("submitted_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("settled_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (PaymentRefund) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id", "idempotency_key").Unique(),
		index.Fields("order_id").StorageKey("idx_payment_refunds_order"),
		index.Fields("order_id").
			Unique().
			StorageKey("idx_payment_refunds_one_active_order").
			Annotations(entsql.IndexWhere("status IN ('REQUESTED', 'RESERVED', 'SUBMITTING', 'PENDING')")),
		index.Fields("user_id", "created_at"),
		index.Fields("status"),
	}
}
