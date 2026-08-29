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

// AnnouncementRead holds the schema definition for the AnnouncementRead entity.
//
// 记录用户对公告的已读状态（首次已读时间）。
type AnnouncementRead struct {
	ent.Schema
}

func (AnnouncementRead) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "announcement_reads"},
	}
}

func (AnnouncementRead) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.UUIDv7IDMixin{}}
}

func (AnnouncementRead) Fields() []ent.Field {
	return []ent.Field{
		field.String("announcement_id").SchemaType(postgresUUIDSchema),
		field.String("user_id").SchemaType(postgresUUIDSchema),
		field.Time("read_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("用户首次已读时间"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (AnnouncementRead) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("announcement", Announcement.Type).
			Ref("reads").
			Field("announcement_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("announcement_reads").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (AnnouncementRead) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("announcement_id"),
		index.Fields("user_id"),
		index.Fields("read_at"),
		index.Fields("announcement_id", "user_id").Unique(),
	}
}
