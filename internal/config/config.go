package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Addr          string
	DatabaseDSN   string
	AutoMigrate   bool
	JWTSecret     string
	TokenDuration time.Duration
	SteamAPIKey   string
	SteamID       string
	SteamSyncLang string
	ImageBedAPIURL string
	ImageBedToken  string
	ImageBedAlbumID int
	ImageBedOrder   string
}

type fileConfig struct {
	Server struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"server"`
	Database struct {
		DSN         string `mapstructure:"dsn"`
		AutoMigrate bool   `mapstructure:"autoMigrate"`
	} `mapstructure:"database"`
	JWT struct {
		Secret       string `mapstructure:"secret"`
		ExpiresHours int    `mapstructure:"expiresHours"`
	} `mapstructure:"jwt"`
	Steam struct {
		APIKey   string `mapstructure:"apiKey"`
		SteamID  string `mapstructure:"steamId"`
		SyncLang string `mapstructure:"syncLang"`
	} `mapstructure:"steam"`
	ImageBed struct {
		APIURL   string `mapstructure:"apiUrl"`
		Token    string `mapstructure:"token"`
		AlbumID  int    `mapstructure:"albumId"`
		Order    string `mapstructure:"order"`
	} `mapstructure:"imageBed"`
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf(
			"read config: %w (请复制 config/config.example.yaml 为 config/config.yaml)",
			err,
		)
	}

	var raw fileConfig
	if err := v.Unmarshal(&raw); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	cfg, err := validateAndBuild(raw)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validateAndBuild(raw fileConfig) (Config, error) {
	addr := strings.TrimSpace(raw.Server.Addr)
	if addr == "" {
		addr = ":8080"
	}

	dsn := strings.TrimSpace(raw.Database.DSN)
	if dsn == "" {
		return Config{}, fmt.Errorf("database.dsn 不能为空")
	}

	secret := strings.TrimSpace(raw.JWT.Secret)
	if secret == "" {
		return Config{}, fmt.Errorf("jwt.secret 不能为空")
	}

	expiresHours := raw.JWT.ExpiresHours
	if expiresHours <= 0 {
		expiresHours = 2
	}

	syncLang := strings.TrimSpace(raw.Steam.SyncLang)
	if syncLang == "" {
		syncLang = "schinese"
	}

	imageBedOrder := strings.TrimSpace(raw.ImageBed.Order)
	if imageBedOrder == "" {
		imageBedOrder = "newest"
	}

	return Config{
		Addr:            addr,
		DatabaseDSN:     dsn,
		AutoMigrate:     raw.Database.AutoMigrate,
		JWTSecret:       secret,
		TokenDuration:   time.Duration(expiresHours) * time.Hour,
		SteamAPIKey:     strings.TrimSpace(raw.Steam.APIKey),
		SteamID:         strings.TrimSpace(raw.Steam.SteamID),
		SteamSyncLang:   syncLang,
		ImageBedAPIURL:  strings.TrimSpace(raw.ImageBed.APIURL),
		ImageBedToken:   strings.TrimSpace(raw.ImageBed.Token),
		ImageBedAlbumID: raw.ImageBed.AlbumID,
		ImageBedOrder:   imageBedOrder,
	}, nil
}
