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

type PostTag struct {
	ent.Schema
}

func (PostTag) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("postId").
			StorageKey("postId"),
		field.Uint64("tagId").
			StorageKey("tagId"),
		field.Time("createdAt").
			StorageKey("createdAt").
			Default(time.Now).
			Immutable(),
	}
}

func (PostTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("post", Post.Type).
			Unique().
			Required().
			Field("postId"),
		edge.To("tag", Tag.Type).
			Unique().
			Required().
			Field("tagId"),
	}
}

func (PostTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "postTags"},
	}
}

func (PostTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("postId", "tagId").Unique().StorageKey("pkPostTags"),
		index.Fields("tagId").StorageKey("idxPostTagsTagId"),
	}
}
