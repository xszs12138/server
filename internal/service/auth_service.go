package service

import (
	"context"
	"errors"
	"time"

	"blog-server/internal/dao"
	"blog-server/internal/dto"
	"blog-server/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")

type AuthService struct {
	users         dao.UserDAO
	jwtSecret     []byte
	tokenDuration time.Duration
}

func NewAuthService(users dao.UserDAO, jwtSecret string, tokenDuration time.Duration) *AuthService {
	return &AuthService{
		users:         users,
		jwtSecret:     []byte(jwtSecret),
		tokenDuration: tokenDuration,
	}
}

func (svc *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := svc.users.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive() {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := svc.issueToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: token,
		ExpiresIn:   int64(svc.tokenDuration.Seconds()),
		User: dto.AdminUser{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Role:     user.Role,
		},
	}, nil
}

func (svc *AuthService) issueToken(user *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"iat":      now.Unix(),
		"exp":      now.Add(svc.tokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(svc.jwtSecret)
}
