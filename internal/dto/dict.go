package dto

import "time"

type DictTypeItem struct {
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	WebPublic   bool       `json:"webPublic"`
}

type DictTypeUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type DictItemRequest struct {
	DictType string `json:"dictType"`
	Value    int    `json:"value"`
	Code     string `json:"code"`
	Label    string `json:"label"`
	Enabled  bool   `json:"enabled"`
	Sort     int    `json:"sort"`
}

type DictItem struct {
	ID        uint64    `json:"id"`
	DictType  string    `json:"dictType"`
	Value     int       `json:"value"`
	Code      string    `json:"code"`
	Label     string    `json:"label"`
	Enabled   bool      `json:"enabled"`
	Sort      int       `json:"sort"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WebDictItem struct {
	Value int    `json:"value"`
	Code  string `json:"code"`
	Label string `json:"label"`
	Sort  int    `json:"sort"`
}
