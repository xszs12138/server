package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
	Avatar   string `json:"avatar,omitempty"`
	Role     string `json:"role"`
}
