package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type DictItem struct {
	ent.Schema
}

func (DictItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.String("dictType").
			StorageKey("dictType").
			MaxLen(64).
			NotEmpty(),
		field.Int("value").
			Positive(),
		field.String("code").
			MaxLen(64).
			Optional().
			Nillable(),
		field.String("label").
			MaxLen(64).
			NotEmpty(),
		field.Bool("enabled").
			Default(true),
		field.Int("sort").
			Default(100),
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

func (DictItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dictItems"},
	}
}

func (DictItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("dictType", "value").Unique().StorageKey("ukDictItemsTypeValue"),
		index.Fields("dictType", "code").Unique().StorageKey("ukDictItemsTypeCode"),
		index.Fields("dictType", "enabled", "sort").StorageKey("idxDictItemsTypeEnabledSort"),
	}
}
