package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OperationLog struct {
	ent.Schema
}

func (OperationLog) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.Uint64("userId").
			StorageKey("userId"),
		field.String("username").
			MaxLen(64).
			NotEmpty(),
		field.Int("actionValue").
			StorageKey("actionValue").
			Positive(),
		field.String("actionLabel").
			StorageKey("actionLabel").
			MaxLen(64).
			NotEmpty(),
		field.String("ip").
			MaxLen(64).
			NotEmpty(),
		field.String("region").
			MaxLen(128).
			Default("未知区域"),
		field.String("userAgent").
			StorageKey("userAgent").
			MaxLen(512).
			Default(""),
		field.Time("createdAt").
			StorageKey("createdAt").
			Default(time.Now).
			Immutable(),
	}
}

func (OperationLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "operationLogs"},
	}
}

func (OperationLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("createdAt").StorageKey("operationlogCreatedAt"),
		index.Fields("userId").StorageKey("operationlogUserId"),
		index.Fields("actionValue").StorageKey("operationlogActionValue"),
	}
}
