package dto

import "time"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"required"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
}

type OperationMeta struct {
	IP        string
	Region    string
	UserAgent string
}

type LoginResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpiresIn   int64     `json:"expiresIn"`
	User        AdminUser `json:"user"`
}

type AdminUser struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
}

type DictItemRequest struct {
	DictType string `json:"dictType" binding:"required"`
	Value    int    `json:"value" binding:"required"`
	Label    string `json:"label" binding:"required"`
	Enabled  bool   `json:"enabled"`
	Sort     int    `json:"sort"`
}

type DictItem struct {
	ID        uint64    `json:"id"`
	DictType  string    `json:"dictType"`
	Value     int       `json:"value"`
	Label     string    `json:"label"`
	Enabled   bool      `json:"enabled"`
	Sort      int       `json:"sort"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type OperationLog struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"userId"`
	Username    string    `json:"username"`
	ActionValue int       `json:"actionValue"`
	ActionLabel string    `json:"actionLabel"`
	IP          string    `json:"ip"`
	Region      string    `json:"region"`
	UserAgent   string    `json:"userAgent"`
	CreatedAt   time.Time `json:"createdAt"`
}
