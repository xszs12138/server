package config

import (
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Addr          string
	JWTSecret     string
	TokenDuration time.Duration
	Admin         AdminConfig
}

type AdminConfig struct {
	ID           uint64
	Username     string
	PasswordHash string
	Nickname     string
	Avatar       string
	Role         string
	Status       string
}

func Load() Config {
	ttlHours := envInt("JWT_EXPIRES_HOURS", 2)
	passwordHash := os.Getenv("ADMIN_PASSWORD_HASH")
	if passwordHash == "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(env("ADMIN_PASSWORD", "password")), bcrypt.DefaultCost)
		passwordHash = string(hash)
	}

	return Config{
		Addr:          env("SERVER_ADDR", ":8080"),
		JWTSecret:     env("JWT_SECRET", "dev-secret-change-me"),
		TokenDuration: time.Duration(ttlHours) * time.Hour,
		Admin: AdminConfig{
			ID:           uint64(envInt("ADMIN_ID", 1)),
			Username:     env("ADMIN_USERNAME", "admin"),
			PasswordHash: passwordHash,
			Nickname:     env("ADMIN_NICKNAME", "管理员"),
			Avatar:       env("ADMIN_AVATAR", ""),
			Role:         env("ADMIN_ROLE", "admin"),
			Status:       env("ADMIN_STATUS", "active"),
		},
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
