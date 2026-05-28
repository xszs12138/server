package model

import "time"

const (
	CommentStatusPending  = "pending"
	CommentStatusApproved = "approved"
	CommentStatusRejected = "rejected"
	CommentStatusSpam     = "spam"
)

type Comment struct {
	ID          uint64
	PostID      uint64
	ParentID    *uint64
	Nickname    string
	Email       string
	Website     string
	Content     string
	Status      string
	IP          string
	UserAgent   string
	AdminUserID *uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PostTitle   string
	PostSlug    string
	Replies     []Comment
}

type CommentCreateInput struct {
	PostID      uint64
	ParentID    *uint64
	Nickname    string
	Email       string
	Website     string
	Content     string
	Status      string
	IP          string
	UserAgent   string
	AdminUserID *uint64
}

type CommentListFilter struct {
	Page     int
	PageSize int
	PostID   uint64
	Status   string
	Keyword  string
	WebOnly  bool
}
