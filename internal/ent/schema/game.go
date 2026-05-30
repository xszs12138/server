package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Game struct {
	ent.Schema
}

func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Immutable().
			Unique(),
		field.Uint32("steamAppId").
			StorageKey("steamAppId").
			Unique(),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("nameZh").
			StorageKey("nameZh").
			MaxLen(255).
			Optional().
			Nillable(),
		field.String("cover").
			MaxLen(512).
			Optional().
			Default(""),
		field.JSON("genres", []string{}).
			Default([]string{}),
		field.Uint32("playtimeMinutes").
			StorageKey("playtimeMinutes").
			Default(0),
		field.Uint32("playtime2WeeksMinutes").
			StorageKey("playtime2WeeksMinutes").
			Default(0),
		field.Time("lastPlayedAt").
			StorageKey("lastPlayedAt").
			Optional().
			Nillable(),
		field.Uint32("achievementUnlocked").
			StorageKey("achievementUnlocked").
			Optional().
			Nillable(),
		field.Uint32("achievementTotal").
			StorageKey("achievementTotal").
			Optional().
			Nillable(),
		field.Uint8("progressPercent").
			StorageKey("progressPercent").
			Optional().
			Nillable(),
		field.String("progressSource").
			StorageKey("progressSource").
			MaxLen(32).
			Default("none"),
		field.String("playStatus").
			StorageKey("playStatus").
			MaxLen(32).
			Default("backlog"),
		field.Bool("isVisible").
			StorageKey("isVisible").
			Default(false),
		field.Int("sort").
			Default(100),
		field.Time("syncedAt").
			StorageKey("syncedAt").
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
	}
}

func (Game) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "games"},
	}
}

func (Game) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("steamAppId").Unique().StorageKey("ukGamesSteamAppId"),
		index.Fields("isVisible", "sort").StorageKey("idxGamesVisibleSort"),
		index.Fields("lastPlayedAt").StorageKey("idxGamesLastPlayedAt"),
	}
}
