package dto

import "time"

type WebCommentReply struct {
	ID        uint64    `json:"id"`
	Nickname  string    `json:"nickname"`
	Website   string    `json:"website,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type WebCommentItem struct {
	ID        uint64            `json:"id"`
	Nickname  string            `json:"nickname"`
	Website   string            `json:"website,omitempty"`
	Content   string            `json:"content"`
	CreatedAt time.Time         `json:"createdAt"`
	Replies   []WebCommentReply `json:"replies"`
}

type WebCommentCreateRequest struct {
	Nickname string  `json:"nickname" binding:"required,max=64"`
	Email    string  `json:"email" binding:"omitempty,max=128,email"`
	Website  string  `json:"website" binding:"omitempty,max=512,url"`
	Content  string  `json:"content" binding:"required,max=2000"`
	ParentID *uint64 `json:"parentId"`
}

type WebCommentCreateResponse struct {
	ID     uint64 `json:"id"`
	Status string `json:"status"`
}

type AdminCommentPost struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

type AdminCommentListItem struct {
	ID        uint64            `json:"id"`
	Post      AdminCommentPost  `json:"post"`
	ParentID  *uint64           `json:"parentId"`
	Nickname  string            `json:"nickname"`
	Email     string            `json:"email,omitempty"`
	Website   string            `json:"website,omitempty"`
	Content   string            `json:"content"`
	Status    string            `json:"status"`
	IP        string            `json:"ip"`
	CreatedAt time.Time         `json:"createdAt"`
}

type AdminCommentReplyRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type CommentStatusResponse struct {
	ID     uint64 `json:"id"`
	Status string `json:"status"`
}
