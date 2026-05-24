package model

import "time"

type User struct {
	ID           uint64
	Username     string
	PasswordHash string
	Nickname     string
	Avatar       string
	Email        string
	Status       string
}

func (u User) IsActive() bool {
	return u.Status == "active"
}

type DictItem struct {
	ID        uint64
	DictType  string
	Value     int
	Label     string
	Enabled   bool
	Sort      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OperationLog struct {
	ID          uint64
	UserID      uint64
	Username    string
	ActionValue int
	ActionLabel string
	IP          string
	Region      string
	UserAgent   string
	CreatedAt   time.Time
}
