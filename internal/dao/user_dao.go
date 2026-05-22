package dao

import (
	"context"
	"errors"

	"blog-server/internal/config"
	"blog-server/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserDAO interface {
	FindByUsername(ctx context.Context, username string) (*model.User, error)
}

type MemoryUserDAO struct {
	users map[string]model.User
}

func NewMemoryUserDAO(admin config.AdminConfig) *MemoryUserDAO {
	user := model.User{
		ID:           admin.ID,
		Username:     admin.Username,
		PasswordHash: admin.PasswordHash,
		Nickname:     admin.Nickname,
		Avatar:       admin.Avatar,
		Role:         admin.Role,
		Status:       admin.Status,
	}

	return &MemoryUserDAO{
		users: map[string]model.User{
			user.Username: user,
		},
	}
}

func (dao *MemoryUserDAO) FindByUsername(_ context.Context, username string) (*model.User, error) {
	user, ok := dao.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &user, nil
}
