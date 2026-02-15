package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fcstask/internal/app"
	"fcstask/internal/config"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	app, err := app.New(cfg)
	if err != nil {
		log.Fatal("Failed to create app:", err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
