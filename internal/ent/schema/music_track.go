package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MusicTrack struct {
	ent.Schema
}

func (MusicTrack) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.String("name").
			MaxLen(160).
			NotEmpty(),
		field.String("artist").
			MaxLen(128).
			Default(""),
		field.String("audioUrl").
			StorageKey("audioUrl").
			MaxLen(512).
			NotEmpty(),
		field.String("coverUrl").
			StorageKey("coverUrl").
			MaxLen(512).
			Optional().
			Default(""),
		field.Text("lrc").
			Optional(),
		field.Int("durationSeconds").
			StorageKey("durationSeconds").
			Default(0),
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

func (MusicTrack) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "musicTracks"},
	}
}

func (MusicTrack) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("visible", "sort").StorageKey("idxMusicTracksVisibleSort"),
	}
}
