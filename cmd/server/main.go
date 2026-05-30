package main

import (
	"context"
	"log"

	"blog-server/internal/config"
	"blog-server/internal/database"
	"blog-server/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	ctx := context.Background()

	client, err := database.Open(ctx, cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer client.Close()

	engine := router.New(cfg, client)

	if err := engine.Run(cfg.Addr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
