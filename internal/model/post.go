package model

import "time"

type Category struct {
	ID          uint64
	Name        string
	Slug        string
	Description string
	Sort        int
	Visible     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Tag struct {
	ID          uint64
	Name        string
	Slug        string
	Description string
	Sort        int
	Visible     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Post struct {
	ID          uint64
	Title       string
	Slug        string
	Cover       string
	Summary     string
	Content     string
	ContentType string
	Status      string
	IsPinned    bool
	ViewCount   uint64
	CategoryID  *uint64
	AuthorID    uint64
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Category    *Category
	Tags        []Tag
}

type PostNeighbor struct {
	Title       string
	Slug        string
	Cover       string
	PublishedAt *time.Time
}

type ArchivePost struct {
	ID          uint64
	Title       string
	Slug        string
	Cover       string
	Summary     string
	ViewCount   int
	PublishedAt time.Time
}

type PostCountFilter struct {
	CategoryID uint64
	TagID      uint64
}

type PostListFilter struct {
	Page         int
	PageSize     int
	Status       string
	CategoryID   uint64
	CategorySlug string
	TagID        uint64
	TagSlug      string
	Keyword      string
	WebOnly      bool
}

type PostCreateInput struct {
	Title       string
	Slug        string
	Cover       string
	Summary     string
	Content     string
	ContentType string
	Status      string
	IsPinned    bool
	CategoryID  *uint64
	TagIDs      []uint64
	AuthorID    uint64
}

type PostUpdateInput struct {
	Title       string
	Slug        string
	Cover       string
	Summary     string
	Content     string
	ContentType string
	Status      string
	IsPinned    bool
	CategoryID  *uint64
	TagIDs      []uint64
}

type PostStatusUpdate struct {
	ID          uint64
	Status      string
	PublishedAt *time.Time
}
