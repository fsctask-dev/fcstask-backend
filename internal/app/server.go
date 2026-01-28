package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"

	"fcstask/internal/api"
	"fcstask/internal/config"
	"fcstask/internal/db"
	"fcstask/internal/db/model"
	"fcstask/internal/server"
)

type App struct {
	echo            *echo.Echo
	db              *db.Client
	apiServer       *server.APIServer
	httpServer      server.HTTPServer
	shutdownTimeout time.Duration
}

func New(cfg *config.Config) (*App, error) {
	e := echo.New()

	dbClient, err := db.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	if err := dbClient.AutoMigrate(&model.User{}); err != nil {
		log.Printf("Warning: failed to run migrations: %v", err)
	}

	apiServer := server.NewAPIServer(dbClient)

	api.RegisterHandlers(e, apiServer)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	httpServer := server.NewHTTPServer(addr, e)

	return &App{
		echo:            e,
		db:              dbClient,
		apiServer:       apiServer,
		httpServer:      httpServer,
		shutdownTimeout: cfg.Server.ShutdownTimeout,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.httpServer.Start(ctx)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			a.shutdownTimeout,
		)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
		}

		if err := a.db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}

		return nil
	}
}
