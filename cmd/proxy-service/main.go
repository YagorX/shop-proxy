package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/YagorX/shop-proxy/internal/app"
	"github.com/YagorX/shop-proxy/internal/config"
)

func main() {
	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
