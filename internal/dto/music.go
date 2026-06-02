package dto

import "time"

type MusicTrackRequest struct {
	Name            string `json:"name" binding:"required"`
	Artist          string `json:"artist"`
	AudioURL        string `json:"audioUrl" binding:"required"`
	CoverURL        string `json:"coverUrl"`
	Lrc             string `json:"lrc"`
	DurationSeconds int    `json:"durationSeconds"`
	Sort            int    `json:"sort"`
	Visible         bool   `json:"visible"`
}

type AdminMusicTrackListItem struct {
	ID              uint64    `json:"id"`
	Name            string    `json:"name"`
	Artist          string    `json:"artist"`
	AudioURL        string    `json:"audioUrl"`
	CoverURL        string    `json:"coverUrl"`
	DurationSeconds int       `json:"durationSeconds"`
	Sort            int       `json:"sort"`
	Visible         bool      `json:"visible"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AdminMusicTrackDetail struct {
	AdminMusicTrackListItem
	Lrc string `json:"lrc"`
}

type WebMusicTrackItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	URL    string `json:"url"`
	Cover  string `json:"cover,omitempty"`
	Lrc    string `json:"lrc,omitempty"`
}

type WebMusicPlaylistResponse struct {
	Items []WebMusicTrackItem `json:"items"`
}
