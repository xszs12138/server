package dto

import "time"

type TaxonomyRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Visible     bool   `json:"visible"`
}

type AdminCategoryListItem struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	Visible     bool      `json:"visible"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AdminCategoryDetail struct {
	AdminCategoryListItem
}

type AdminTagListItem struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	Visible     bool      `json:"visible"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AdminTagDetail struct {
	AdminTagListItem
}

type TaxonomyIDResponse struct {
	ID uint64 `json:"id"`
}

type WebCategoryItem struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	PostCount   int    `json:"postCount"`
}

type WebTagItem struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int    `json:"postCount"`
}
