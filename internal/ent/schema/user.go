package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.String("username").
			MaxLen(64).
			NotEmpty().
			Unique(),
		field.String("passwordHash").
			StorageKey("passwordHash").
			MaxLen(255).
			NotEmpty(),
		field.String("nickname").
			MaxLen(64).
			NotEmpty(),
		field.String("avatar").
			MaxLen(512).
			Optional().
			Default(""),
		field.String("email").
			MaxLen(128).
			Optional().
			Nillable().
			Unique(),
		field.String("status").
			MaxLen(32).
			Default("active"),
		field.Time("lastLoginAt").
			StorageKey("lastLoginAt").
			Optional().
			Nillable(),
		field.Time("createdAt").
			StorageKey("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			StorageKey("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deletedAt").
			StorageKey("deletedAt").
			Optional().
			Nillable(),
	}
}
