package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type GameMonthlyStat struct {
	ent.Schema
}

func (GameMonthlyStat) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.String("yearMonth").
			StorageKey("yearMonth").
			MaxLen(7).
			NotEmpty(),
		field.Uint32("totalMinutes").
			StorageKey("totalMinutes").
			Default(0),
		field.Time("updatedAt").
			StorageKey("updatedAt").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (GameMonthlyStat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gameMonthlyStats"},
	}
}

func (GameMonthlyStat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("yearMonth").Unique().StorageKey("ukGameMonthlyStatsYearMonth"),
	}
}
