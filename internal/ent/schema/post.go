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

type Post struct {
	ent.Schema
}

func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.String("title").
			MaxLen(160).
			NotEmpty(),
		field.String("slug").
			MaxLen(180).
			NotEmpty().
			Unique(),
		field.String("cover").
			MaxLen(512).
			Optional().
			Default(""),
		field.String("summary").
			MaxLen(500).
			Optional().
			Default(""),
		field.Text("content").
			NotEmpty(),
		field.String("contentType").
			StorageKey("contentType").
			MaxLen(32).
			Default("markdown"),
		field.String("status").
			MaxLen(32).
			Default("draft"),
		field.Bool("isPinned").
			StorageKey("isPinned").
			Default(false),
		field.Uint64("viewCount").
			StorageKey("viewCount").
			Default(0),
		field.Uint64("categoryId").
			StorageKey("categoryId").
			Optional().
			Nillable(),
		field.Uint64("authorId").
			StorageKey("authorId"),
		field.Time("publishedAt").
			StorageKey("publishedAt").
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

func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("posts").
			Unique().
			Required().
			Field("authorId"),
		edge.From("category", Category.Type).
			Ref("posts").
			Unique().
			Field("categoryId"),
		edge.To("tags", Tag.Type).
			Through("postTags", PostTag.Type),
		edge.To("comments", Comment.Type),
	}
}

func (Post) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "posts"},
	}
}

func (Post) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique().StorageKey("ukPostsSlug"),
		index.Fields("status", "publishedAt").StorageKey("idxPostsStatusPublishedAt"),
		index.Fields("categoryId", "status").StorageKey("idxPostsCategoryStatus"),
		index.Fields("authorId").StorageKey("idxPostsAuthorId"),
		index.Fields("isPinned", "publishedAt").StorageKey("idxPostsIsPinnedPublishedAt"),
	}
}
