package model

type User struct {
	ID           uint64
	Username     string
	PasswordHash string
	Nickname     string
	Avatar       string
	Role         string
	Status       string
}

func (u User) IsActive() bool {
	return u.Status == "active"
}
