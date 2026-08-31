package schema

import (
	"github.com/AsukaCC/EasySub2api/ent/schema/mixins"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentOrder holds the schema definition for the PaymentOrder entity.
//
// 删除策略：硬删除
// PaymentOrder 使用硬删除而非软删除，原因如下：
//   - 订单通过 status 字段追踪完整生命周期，无需依赖软删除
//   - 订单审计通过 PaymentAuditLog 表记录，删除前可归档
//   - 减少查询复杂度，避免软删除过滤开销
type PaymentOrder struct {
	ent.Schema
}

func (PaymentOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "payment_orders",
			Checks: map[string]string{
				"payment_orders_order_type_valid":       "order_type IN ('balance', 'subscription')",
				"payment_orders_inventory_status_valid": "inventory_status IN ('NONE', 'RESERVED', 'CONSUMED', 'RELEASED')",
			},
		},
	}
}

func (PaymentOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.UUIDv7IDMixin{}}
}

func (PaymentOrder) Fields() []ent.Field {
	return []ent.Field{
		// 用户信息（冗余存储，避免关联查询）
		field.String("user_id").SchemaType(postgresUUIDSchema),
		field.String("user_email").
			MaxLen(255),
		field.String("user_name").
			MaxLen(100),
		field.String("user_notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 金额信息
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("pay_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("fee_rate").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).
			Default(0),
		field.Float("wallet_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("wallet_bonus_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("wallet_recharge_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.String("wallet_hold_id").SchemaType(postgresUUIDSchema).
			Optional().
			Nillable(),
		field.Bool("wallet_only").Default(false),
		field.Float("gateway_base_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		// Recharge orders carry both fiat and points. Keep amount/pay_amount above
		// for API compatibility and use these fields as the unambiguous contract.
		field.Float("principal_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.Float("fee_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.String("currency").
			MaxLen(3).
			Default(""),
		field.Float("base_points").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("bonus_points").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("credited_points").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.JSON("bonus_tier_snapshot", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("bonus_expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("bonus_grant_id").
			SchemaType(postgresUUIDSchema).
			Optional().
			Nillable(),
		field.Float("affiliate_rebate_points").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Time("refund_deadline").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Float("refunded_principal_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.Float("refunded_fee_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.Float("refunded_gateway_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.Float("reversed_base_points").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("reversed_bonus_points").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("reversed_affiliate_points").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.String("recharge_code").
			MaxLen(64),

		// 支付信息
		field.String("out_trade_no").
			MaxLen(64).
			Default(""),
		field.String("payment_type").
			MaxLen(30),
		field.String("payment_trade_no").
			MaxLen(128),
		field.String("pay_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("qr_code").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("qr_code_img").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 订单类型 & 订阅关联
		field.String("order_type").
			MaxLen(20).
			Default("balance"),
		field.String("plan_id").SchemaType(postgresUUIDSchema).
			Optional().
			Nillable(),
		field.String("subscription_group_id").SchemaType(postgresUUIDSchema).
			Optional().
			Nillable(),
		field.Int("subscription_days").
			Optional().
			Nillable(),
		field.String("provider_instance_id").SchemaType(postgresUUIDSchema).
			Optional().
			Nillable().
			MaxLen(64),
		field.String("provider_key").
			Optional().
			Nillable().
			MaxLen(30),
		field.JSON("provider_snapshot", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),

		// 状态
		field.String("status").
			MaxLen(30).
			Default("PENDING"),
		field.String("inventory_status").
			MaxLen(16).
			Default("NONE"),
		field.Time("inventory_reserved_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("inventory_consumed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("inventory_released_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		// 退款信息
		field.Float("refund_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.String("refund_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("refund_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Bool("force_refund").
			Default(false),
		field.Time("refund_requested_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("refund_request_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("refund_requested_by").
			Optional().
			Nillable().
			SchemaType(postgresUUIDSchema),

		// 时间节点
		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("failed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("failed_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 来源信息
		field.String("client_ip").
			MaxLen(50),
		field.String("src_host").
			MaxLen(255),
		field.String("src_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 时间戳
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (PaymentOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_orders").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (PaymentOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("out_trade_no").
			Unique().
			Annotations(entsql.IndexWhere("out_trade_no <> ''")),
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("expires_at"),
		index.Fields("created_at"),
		index.Fields("paid_at"),
		index.Fields("payment_type", "paid_at"),
		index.Fields("order_type"),
		index.Fields("plan_id", "inventory_status"),
	}
}
