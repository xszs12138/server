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

type Comment struct {
	ent.Schema
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.Uint64("postId").
			StorageKey("postId"),
		field.Uint64("parentId").
			StorageKey("parentId").
			Optional().
			Nillable(),
		field.String("nickname").
			MaxLen(64).
			NotEmpty(),
		field.String("email").
			MaxLen(128).
			Optional().
			Nillable(),
		field.String("website").
			MaxLen(512).
			Optional().
			Nillable(),
		field.Text("content").
			NotEmpty(),
		field.String("status").
			MaxLen(32).
			Default("pending"),
		field.String("ip").
			MaxLen(64).
			NotEmpty(),
		field.String("userAgent").
			StorageKey("userAgent").
			MaxLen(512).
			Optional().
			Default(""),
		field.Uint64("adminUserId").
			StorageKey("adminUserId").
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

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("comments").
			Unique().
			Required().
			Field("postId"),
		edge.To("replies", Comment.Type).
			From("parent").
			Unique().
			Field("parentId"),
		edge.From("admin", User.Type).
			Ref("adminComments").
			Unique().
			Field("adminUserId"),
	}
}

func (Comment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "comments"},
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("postId", "status", "createdAt").StorageKey("idxCommentsPostStatusCreatedAt"),
		index.Fields("parentId").StorageKey("idxCommentsParentId"),
		index.Fields("status", "createdAt").StorageKey("idxCommentsStatusCreatedAt"),
	}
}
