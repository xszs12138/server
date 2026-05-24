package dao

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/ent"
	entuser "blog-server/internal/ent/user"
	"blog-server/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

type UserDAO interface {
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, user model.User) (*model.User, error)
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateLastLoginAt(ctx context.Context, id uint64) error
}

type EntUserDAO struct {
	client *ent.Client
}

func NewEntUserDAO(client *ent.Client) *EntUserDAO {
	return &EntUserDAO{client: client}
}

func (dao *EntUserDAO) Count(ctx context.Context) (int, error) {
	return dao.client.User.Query().
		Where(entuser.DeletedAtIsNil()).
		Count(ctx)
}

func (dao *EntUserDAO) Create(ctx context.Context, user model.User) (*model.User, error) {
	create := dao.client.User.Create().
		SetUsername(user.Username).
		SetPasswordHash(user.PasswordHash).
		SetNickname(user.Nickname).
		SetAvatar(user.Avatar).
		SetStatus(user.Status)
	if user.Email != "" {
		create.SetEmail(user.Email)
	}

	created, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return toModel(created), nil
}

func (dao *EntUserDAO) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	user, err := dao.client.User.Query().
		Where(
			entuser.IDEQ(id),
			entuser.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return toModel(user), nil
}

func (dao *EntUserDAO) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := dao.client.User.Query().
		Where(
			entuser.UsernameEQ(username),
			entuser.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return toModel(user), nil
}

func (dao *EntUserDAO) UpdateLastLoginAt(ctx context.Context, id uint64) error {
	return dao.client.User.UpdateOneID(id).
		SetLastLoginAt(time.Now()).
		Exec(ctx)
}

func toModel(user *ent.User) *model.User {
	return &model.User{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Email:        valueOrEmpty(user.Email),
		Status:       user.Status,
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
