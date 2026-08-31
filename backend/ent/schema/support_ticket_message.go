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

// SupportTicketMessage is an immutable timeline entry.
type SupportTicketMessage struct{ ent.Schema }

func (SupportTicketMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{
		Table: "support_ticket_messages",
		Checks: map[string]string{
			"support_ticket_messages_author_role_valid": "author_role IN ('USER', 'ADMIN', 'SYSTEM')",
			"support_ticket_messages_kind_valid":        "kind IN ('COMMENT', 'SYSTEM')",
		},
	}}
}

func (SupportTicketMessage) Mixin() []ent.Mixin { return []ent.Mixin{mixins.UUIDv7IDMixin{}} }

func (SupportTicketMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("ticket_id").SchemaType(postgresUUIDSchema),
		field.String("author_id").SchemaType(postgresUUIDSchema).Optional().Nillable(),
		field.String("author_role").MaxLen(16),
		field.String("kind").MaxLen(16).Default("COMMENT"),
		field.String("body").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.String("event_type").MaxLen(64).Default(""),
		field.String("event_data").SchemaType(map[string]string{dialect.Postgres: "jsonb"}).Default("{}"),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicketMessage) Indexes() []ent.Index {
	return []ent.Index{index.Fields("ticket_id", "created_at")}
}
