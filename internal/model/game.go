package model

import "time"

type Game struct {
	ID                     uint64
	SteamAppID             uint32
	Name                   string
	NameZh                 *string
	Cover                  string
	Genres                 []string
	PlaytimeMinutes        uint32
	Playtime2WeeksMinutes  uint32
	LastPlayedAt           *time.Time
	AchievementUnlocked    *uint32
	AchievementTotal       *uint32
	ProgressPercent        *uint8
	ProgressSource         string
	PlayStatus             string
	IsVisible              bool
	Sort                   int
	SyncedAt               *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type GameMonthlyStat struct {
	YearMonth    string
	TotalMinutes uint32
}
