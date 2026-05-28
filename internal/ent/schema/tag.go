package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Tag struct {
	ent.Schema
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.String("name").
			MaxLen(64).
			NotEmpty(),
		field.String("slug").
			MaxLen(96).
			NotEmpty().
			Unique(),
		field.String("description").
			MaxLen(255).
			Optional().
			Default(""),
		field.Int("sort").
			Default(100),
		field.Bool("visible").
			Default(true),
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

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("posts", Post.Type).
			Ref("tags"),
	}
}

func (Tag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tags"},
	}
}

func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique().StorageKey("ukTagsSlug"),
		index.Fields("visible", "sort").StorageKey("idxTagsVisibleSort"),
	}
}
