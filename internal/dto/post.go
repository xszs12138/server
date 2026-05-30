package dto

import "time"

type WebPostCategory struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type WebPostTag struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type WebPostListItem struct {
	ID          uint64           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Cover       string           `json:"cover"`
	Summary     string           `json:"summary"`
	IsPinned    bool             `json:"isPinned"`
	ViewCount   uint64           `json:"viewCount"`
	PublishedAt *time.Time       `json:"publishedAt"`
	Category    *WebPostCategory `json:"category,omitempty"`
	Tags        []WebPostTag     `json:"tags"`
}

type WebPostNeighbor struct {
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Cover       string     `json:"cover"`
	PublishedAt *time.Time `json:"publishedAt"`
}

type WebPostDetail struct {
	ID          uint64           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Cover       string           `json:"cover"`
	Summary     string           `json:"summary"`
	Content     string           `json:"content"`
	ContentType string           `json:"contentType"`
	ViewCount   uint64           `json:"viewCount"`
	PublishedAt *time.Time       `json:"publishedAt"`
	Category    *WebPostCategory `json:"category,omitempty"`
	Tags        []WebPostTag     `json:"tags"`
	PrevPost    *WebPostNeighbor `json:"prevPost,omitempty"`
	NextPost    *WebPostNeighbor `json:"nextPost,omitempty"`
}

type AdminPostCategory struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type AdminPostTag struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type AdminPostListItem struct {
	ID          uint64             `json:"id"`
	Title       string             `json:"title"`
	Slug        string             `json:"slug"`
	Cover       string             `json:"cover"`
	Summary     string             `json:"summary"`
	Status      string             `json:"status"`
	IsPinned    bool               `json:"isPinned"`
	ViewCount   uint64             `json:"viewCount"`
	Category    *AdminPostCategory `json:"category,omitempty"`
	Tags        []AdminPostTag     `json:"tags"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	PublishedAt *time.Time         `json:"publishedAt"`
}

type AdminPostDetail struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Cover       string     `json:"cover"`
	Summary     string     `json:"summary"`
	Content     string     `json:"content"`
	ContentType string     `json:"contentType"`
	Status      string     `json:"status"`
	IsPinned    bool       `json:"isPinned"`
	ViewCount   uint64     `json:"viewCount"`
	CategoryID  *uint64    `json:"categoryId"`
	TagIDs      []uint64   `json:"tagIds"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	PublishedAt *time.Time `json:"publishedAt"`
}

type PostRequest struct {
	Title       string   `json:"title" binding:"required"`
	Slug        string   `json:"slug" binding:"required"`
	Cover       string   `json:"cover"`
	Summary     string   `json:"summary"`
	Content     string   `json:"content" binding:"required"`
	ContentType string   `json:"contentType"`
	Status      string   `json:"status"`
	IsPinned    bool     `json:"isPinned"`
	CategoryID  *uint64  `json:"categoryId"`
	TagIDs      []uint64 `json:"tagIds"`
}

type PostIDResponse struct {
	ID uint64 `json:"id"`
}

type PostStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type PostStatusResponse struct {
	ID          uint64     `json:"id"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"publishedAt"`
}
