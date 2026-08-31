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

// SupportTicketRead stores per-reader read progress.
type SupportTicketRead struct{ ent.Schema }

func (SupportTicketRead) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "support_ticket_reads"}}
}

func (SupportTicketRead) Mixin() []ent.Mixin { return []ent.Mixin{mixins.UUIDv7IDMixin{}} }

func (SupportTicketRead) Fields() []ent.Field {
	return []ent.Field{
		field.String("ticket_id").SchemaType(postgresUUIDSchema),
		field.String("reader_id").SchemaType(postgresUUIDSchema),
		field.Time("read_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicketRead) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_id", "reader_id").Unique(),
		index.Fields("reader_id", "read_at"),
	}
}
