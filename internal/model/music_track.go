package model

import "time"

type MusicTrack struct {
	ID              uint64
	Name            string
	Artist          string
	AudioURL        string
	CoverURL        string
	Lrc             string
	DurationSeconds int
	Sort            int
	Visible         bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type MusicTrackListFilter struct {
	Page     int
	PageSize int
	Keyword  string
	Visible  *bool
}

type MusicTrackSaveInput struct {
	Name            string
	Artist          string
	AudioURL        string
	CoverURL        string
	Lrc             string
	DurationSeconds int
	Sort            int
	Visible         bool
}
