package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// UUIDv7IDMixin provides a string-backed UUIDv7 primary key stored as native
// PostgreSQL uuid. UUIDv7 keeps newly-created rows roughly time-ordered while
// avoiding database sequences.
type UUIDv7IDMixin struct {
	mixin.Schema
}

func (UUIDv7IDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			DefaultFunc(newUUIDv7).
			SchemaType(map[string]string{
				dialect.Postgres: "uuid",
			}),
	}
}

func newUUIDv7() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return id.String()
}
