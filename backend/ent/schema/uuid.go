package schema

import "entgo.io/ent/dialect"

var postgresUUIDSchema = map[string]string{
	dialect.Postgres: "uuid",
}
