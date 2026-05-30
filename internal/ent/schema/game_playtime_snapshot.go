package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type GamePlaytimeSnapshot struct {
	ent.Schema
}

func (GamePlaytimeSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.Uint32("steamAppId").
			StorageKey("steamAppId"),
		field.Uint32("playtimeMinutes").
			StorageKey("playtimeMinutes"),
		field.Time("snapshotAt").
			StorageKey("snapshotAt").
			Default(time.Now),
	}
}

func (GamePlaytimeSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gamePlaytimeSnapshots"},
	}
}

func (GamePlaytimeSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("steamAppId", "snapshotAt").StorageKey("idxGamePlaytimeSnapshotsAppIdTime"),
	}
}
