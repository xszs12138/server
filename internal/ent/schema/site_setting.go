package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SiteSetting struct {
	ent.Schema
}

func (SiteSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.String("settingKey").
			StorageKey("key").
			MaxLen(96).
			NotEmpty(),
		field.JSON("value", json.RawMessage{}).
			Default(json.RawMessage("{}")),
		field.String("description").
			Optional().
			Nillable().
			MaxLen(255),
		field.Time("createdAt").
			StorageKey("createdAt").
			Default(time.Now).
			Immutable(),
		field.Time("updatedAt").
			StorageKey("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SiteSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "siteSettings"},
	}
}

func (SiteSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("settingKey").Unique().StorageKey("ukSiteSettingsKey"),
	}
}
