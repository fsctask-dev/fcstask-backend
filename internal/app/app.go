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
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type App struct {
	echo             *echo.Echo
	db               *db.Client
	sessionRepo      repo.SessionRepositoryInterface
	httpServer       server.HTTPServer
	metrics          *metrics.Metrics
	metricsServer    *metrics.Server
	shutdownTimeout  time.Duration
	sessionCfg       config.SessionConfig
	observabilityCfg config.ObservabilityConfig
}

func New(cfg *config.Config) (*App, error) {
	e := echo.New()

	dbClient, err := db.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	m := metrics.New()

	userRepo := repo.NewUserRepository(dbClient)
	sessionRepo := repo.NewSessionRepository(dbClient)
	courseRepo := repo.NewCourseRepository(dbClient)
	roleRepo := repo.NewRoleRepository(dbClient.DB())
	homeworkRepo := repo.NewHomeworkRepository(dbClient.DB())
	taskRepo := repo.NewTaskRepository(dbClient.DB())
	deadlineRepo := repo.NewDeadlineRepository(dbClient.DB())
	studentScoreRepo := repo.NewStudentTaskScoreRepository(dbClient.DB())
	courseLateRepo := repo.NewCourseLatePolicy(dbClient.DB())

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, sessionRepo).WithMetrics(m.Auth, m.Session)
	sessionService := service.NewSessionService(sessionRepo)
	courseService := service.NewCourseService(courseRepo, roleRepo, studentScoreRepo).WithMetrics(m.Course)
	adminHomeworkService := service.NewAdminHomeworkService(homeworkRepo, deadlineRepo, roleRepo).WithMetrics(m.Admin)
	adminTaskService := service.NewAdminTaskService(taskRepo, homeworkRepo, roleRepo).WithMetrics(m.Admin)
	adminRoleService := service.NewAdminRoleService(roleRepo, userRepo).WithMetrics(m.Admin)
	checkerService := service.NewCheckerService(taskRepo, homeworkRepo, studentScoreRepo, deadlineRepo, courseLateRepo, roleRepo).WithMetrics(m.Checker)
	courseLateService := service.NewCourseLatePolicy(courseLateRepo, roleRepo).WithMetrics(m.LatePolicy)
	gradeUpdateService := service.NewGradeUpdateService(taskRepo, studentScoreRepo, roleRepo).WithMetrics(m.Admin)

	adminHomeworkHandler := handler.NewAdminHomeworkHandler(adminHomeworkService)
	adminTaskHandler := handler.NewAdminTaskHandler(adminTaskService)
	adminRoleHandler := handler.NewAdminRoleHandler(adminRoleService)
	courseLateHandler := handler.NewCourseLateHandler(courseLateService)
	gradeUpdateHandler := handler.NewGradeUpdateHandler(gradeUpdateService)

	apiController := controller.NewAPIController(
		handler.NewAuthHandler(authService),
		handler.NewUserHandler(userService),
		handler.NewSessionHandler(sessionService, userService),
		handler.NewCourseHandler(courseService),
		adminHomeworkHandler,
		handler.NewCheckerHandler(checkerService),
		handler.NewGradeUpdateHandler(gradeUpdateService),
	)

	e.Use(metrics.EchoMiddleware(m.HTTP))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	e.Use(authmw.Auth(userRepo, sessionRepo, []string{
		"/v1/api/me",
		"/v1/sessions",
		"/v1/users/sessions",
		"/api/signout",
		"/api/courses",
		"/api/courses/:courseId/scores",
		"/api/courses/:courseId/join",
		"/api/courses/:courseId/invite",
		"/api/grades",
		"/admin/courses/:courseId/homework",
		"/admin/courses/:courseId/homework/:hwId",
		"/admin/courses/:courseId/homework/:hwId/publish",
		"/admin/courses/:courseId/homework/:hwId/deadline",
		"/admin/deadlines/:deadlineId",
		"/admin/courses/:courseId/homework/:hwId/tasks",
		"/admin/courses/:courseId/homework/:hwId/tasks/:taskId",
		"/admin/courses/:courseId/homework/:hwId/tasks/:taskId/publish",
		"/admin/courses/:courseId/homework/:hwId/tasks/:taskId/score",
		"/admin/courses/:courseId/homework/:hwId/tasks/:taskId/update_grade",
		"/admin/courses/:courseId/late-policy",
		"/admin/courses/:courseId/roles",
		"/admin/courses/:courseId/participants",
		"/admin/courses/:courseId/roles/:roleId/permissions",
		"/admin/courses/:courseId/roles/:roleId/permissions/:permission",
		"/admin/super-admins",
		"/admin/homework/:hwId/deadline",
	}))

	api.RegisterHandlers(e, apiController)
	apiController.RegisterCourseRoutes(e)
	apiController.RegisterHomeworkRoutes(e)
	apiController.RegisterAdminRoutes(e, adminHomeworkHandler, adminTaskHandler, adminRoleHandler, courseLateHandler, gradeUpdateHandler)
	apiController.RegisterCheckerRoutes(e)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := server.NewHTTPServer(addr, e)

	metricsServer := metrics.NewServer(cfg.Observability.MetricsAddr, m.Registry)

	return &App{
		echo:             e,
		db:               dbClient,
		sessionRepo:      sessionRepo,
		httpServer:       httpServer,
		metrics:          m,
		metricsServer:    metricsServer,
		shutdownTimeout:  cfg.Server.ShutdownTimeout,
		sessionCfg:       cfg.Session,
		observabilityCfg: cfg.Observability,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- a.httpServer.Start(ctx)
	}()

	go func() {
		errCh <- a.metricsServer.Start(ctx)
	}()

	go a.runSessionCleanup(ctx)
	go a.db.RunStatsCollector(ctx, a.metrics.DB, a.observabilityCfg.DBStatsInterval)

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
				a.metrics.Session.IncCleanupError()
				log.Printf("Session cleanup error: %v", err)
			} else if deleted > 0 {
				a.metrics.Session.AddCleanupDeleted(deleted)
				log.Printf("Session cleanup: removed %d expired sessions", deleted)
			}
		}
	}
}
