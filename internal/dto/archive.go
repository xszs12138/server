package dto

import "time"

type WebArchivePost struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	PublishedAt time.Time `json:"publishedAt"`
}

type WebArchiveMonth struct {
	Month int              `json:"month"`
	Posts []WebArchivePost `json:"posts"`
}

type WebArchiveYear struct {
	Year   int               `json:"year"`
	Months []WebArchiveMonth `json:"months"`
}
