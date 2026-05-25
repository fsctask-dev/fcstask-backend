package app

import (
	"context"
	"fcstask-backend/internal/api"
	"fcstask-backend/internal/config"
	"fcstask-backend/internal/controller"
	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/handler"
	"fcstask-backend/internal/metrics"
	authmw "fcstask-backend/internal/middleware"
	"fcstask-backend/internal/server"
	"fcstask-backend/internal/service"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"time"
)

type App struct {
	echo            *echo.Echo
	db              *db.Client
	sessionRepo     repo.SessionRepositoryInterface
	httpServer      server.HTTPServer
	shutdownTimeout time.Duration
	sessionCfg      config.SessionConfig
}

func New(cfg *config.Config) (*App, error) {
	e := echo.New()

	dbClient, err := db.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	userRepo := repo.NewUserRepository(dbClient)
	sessionRepo := repo.NewSessionRepository(dbClient)
	courseRepo := repo.NewCourseRepository(dbClient)
	roleRepo := repo.NewRoleRepository(dbClient.DB())
	homeworkRepo := repo.NewHomeworkRepository(dbClient.DB())
	taskRepo := repo.NewTaskRepository(dbClient.DB())
	deadlineRepo := repo.NewDeadlineRepository(dbClient.DB())
	studentScoreRepo := repo.NewStudentTaskScoreRepository(dbClient.DB())

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, sessionRepo)
	sessionService := service.NewSessionService(sessionRepo)
	courseService := service.NewCourseService(courseRepo, roleRepo, studentScoreRepo)
	adminHomeworkService := service.NewAdminHomeworkService(homeworkRepo, deadlineRepo, roleRepo)
	adminTaskService := service.NewAdminTaskService(taskRepo, homeworkRepo, roleRepo)
	adminRoleService := service.NewAdminRoleService(roleRepo, userRepo)
	namespaceService := service.NewNamespaceService(
		repo.NewNamespaceRepository(dbClient),
		userRepo,
	)

	adminHomeworkHandler := handler.NewAdminHomeworkHandler(adminHomeworkService)
	adminTaskHandler := handler.NewAdminTaskHandler(adminTaskService)
	adminRoleHandler := handler.NewAdminRoleHandler(adminRoleService)
	namespaceHandler := handler.NewNamespaceHandler(namespaceService)

	apiController := controller.NewAPIController(
		handler.NewAuthHandler(authService),
		handler.NewUserHandler(userService),
		handler.NewSessionHandler(sessionService, userService),
		handler.NewCourseHandler(courseService),
		namespaceHandler,
	)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	e.Use(authmw.Auth(userRepo, sessionRepo, []string{
		"/v1/api/me",
		"/api/signout",
		"/api/namespaces",
		"/api/namespaces/:id",
		"/api/namespaces/:id/users",
		"/api/namespaces/:id/courses",
		"/v1/sessions",
		"/v1/users/sessions",
		"/api/signout",
		"/api/courses/:courseId/scores",
		"/api/courses/:courseId/join",
		"/admin/courses/:courseId/homework",
		"/admin/courses/:courseId/homework/:hwId",
		"/admin/courses/:courseId/homework/:hwId/publish",
		"/admin/courses/:courseId/homework/:hwId/deadline",
		"/admin/deadlines/:deadlineId",
		"/admin/courses/:courseId/homework/:hwId/tasks",
		"/admin/courses/:courseId/homework/:hwId/tasks/:taskId",
		"/admin/courses/:courseId/homework/:hwId/tasks/:taskId/score",
		"/admin/courses/:courseId/roles",
		"/admin/courses/:courseId/participants",
		"/admin/courses/:courseId/roles/:roleId/permissions",
		"/admin/courses/:courseId/roles/:roleId/permissions/:permission",
		"/admin/super-admins",
	}))

	api.RegisterHandlers(e, apiController)
	apiController.RegisterCourseRoutes(e)
	apiController.RegisterNamespaceRoutes(e)
	apiController.RegisterAdminRoutes(e, adminHomeworkHandler, adminTaskHandler, adminRoleHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	metrics.EchoPrometheus(e)

	httpServer := server.NewHTTPServer(addr, e)

	return &App{
		echo:            e,
		db:              dbClient,
		sessionRepo:     sessionRepo,
		httpServer:      httpServer,
		shutdownTimeout: cfg.Server.ShutdownTimeout,
		sessionCfg:      cfg.Session,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.httpServer.Start(ctx)
	}()

	go a.runSessionCleanup(ctx)

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

func (a *App) runSessionCleanup(ctx context.Context) {
	ticker := time.NewTicker(a.sessionCfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleted, err := a.sessionRepo.CleanOutdatedSessions(ctx, a.sessionCfg.TTL)
			if err != nil {
				log.Printf("Session cleanup error: %v", err)
			} else if deleted > 0 {
				log.Printf("Session cleanup: removed %d expired sessions", deleted)
			}
		}
	}
}
