package main

import (
	"log"

	"blog-server/internal/config"
	"blog-server/internal/router"
)

func main() {
	cfg := config.Load()
	engine := router.New(cfg)

	if err := engine.Run(cfg.Addr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
