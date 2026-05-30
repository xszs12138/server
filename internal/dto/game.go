package dto

import "time"

type WebGameListItem struct {
	ID                  uint64     `json:"id"`
	SteamAppID          uint32     `json:"steamAppId"`
	Name                string     `json:"name"`
	NameZh              string     `json:"nameZh"`
	Cover               string     `json:"cover"`
	Genres              []string   `json:"genres"`
	PlayStatus          string     `json:"playStatus"`
	PlayStatusLabel     string     `json:"playStatusLabel,omitempty"`
	ProgressPercent     *uint8     `json:"progressPercent"`
	PlaytimeHours       float64    `json:"playtimeHours"`
	Playtime2WeeksHours float64    `json:"playtime2WeeksHours"`
	LastPlayedAt        *time.Time `json:"lastPlayedAt"`
}

type WebGameGenreItem struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type WebGameRecentItem struct {
	ID              uint64  `json:"id"`
	SteamAppID      uint32  `json:"steamAppId"`
	Name            string  `json:"name"`
	NameZh          string  `json:"nameZh"`
	Cover           string  `json:"cover"`
	PlaytimeHours   float64 `json:"playtimeHours"`
	ProgressPercent *uint8  `json:"progressPercent"`
}

type WebGamePlaytimeMonth struct {
	Month string  `json:"month"`
	Hours float64 `json:"hours"`
}

type WebGamePlaytimeStats struct {
	TotalHours float64                `json:"totalHours"`
	Months     []WebGamePlaytimeMonth `json:"months"`
}

type WebGameSidebar struct {
	RecentGames   []WebGameRecentItem  `json:"recentGames"`
	PlaytimeStats WebGamePlaytimeStats `json:"playtimeStats"`
}

type AdminGameListItem struct {
	WebGameListItem
	ProgressSource      string     `json:"progressSource"`
	IsVisible           bool       `json:"isVisible"`
	PlaytimeMinutes     uint32     `json:"playtimeMinutes"`
	AchievementUnlocked *uint32    `json:"achievementUnlocked"`
	AchievementTotal    *uint32    `json:"achievementTotal"`
	SyncedAt            *time.Time `json:"syncedAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type AdminGameUpdateRequest struct {
	NameZh          *string  `json:"nameZh"`
	Genres          []string `json:"genres"`
	PlayStatus      *string  `json:"playStatus"`
	ProgressPercent *uint8   `json:"progressPercent"`
	IsVisible       *bool    `json:"isVisible"`
	Sort            *int     `json:"sort"`
}

type GameSyncResult struct {
	SyncedCount  int        `json:"syncedCount"`
	VisibleCount int        `json:"visibleCount"`
	SyncedAt     time.Time  `json:"syncedAt"`
}
