package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr          string
	DatabaseDSN   string
	AutoMigrate   bool
	JWTSecret     string
	TokenDuration time.Duration
}

func Load() Config {
	ttlHours := envInt("JWT_EXPIRES_HOURS", 2)

	return Config{
		Addr:          env("SERVER_ADDR", ":8080"),
		DatabaseDSN:   env("DATABASE_DSN", "root:lszqaz12038268@tcp(127.0.0.1:3306)/xszs_blog?charset=utf8mb4&parseTime=True&loc=Local"),
		AutoMigrate:   envBool("AUTO_MIGRATE", true),
		JWTSecret:     env("JWT_SECRET", "dev-secret-change-me"),
		TokenDuration: time.Duration(ttlHours) * time.Hour,
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

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
