package app

import (
	"fcstask-backend/internal/controller"
	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/handler"
	"fcstask-backend/internal/metrics"
	authmw "fcstask-backend/internal/middleware"
	"fcstask-backend/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type repositories struct {
	user         repo.IUserRepo
	session      repo.SessionRepositoryInterface
	course       repo.CourseRepositoryInterface
	role         repo.IRoleRepo
	homework     repo.IHomeworkRepo
	task         repo.ITaskRepo
	deadline     repo.IDeadlineRepo
	studentScore repo.IStudentTaskScoreRepo
}

func newRepositories(dbClient *db.Client) repositories {
	return repositories{
		user:         repo.NewUserRepository(dbClient),
		session:      repo.NewSessionRepository(dbClient),
		course:       repo.NewCourseRepository(dbClient),
		role:         repo.NewRoleRepository(dbClient.DB()),
		homework:     repo.NewHomeworkRepository(dbClient.DB()),
		task:         repo.NewTaskRepository(dbClient.DB()),
		deadline:     repo.NewDeadlineRepository(dbClient.DB()),
		studentScore: repo.NewStudentTaskScoreRepository(dbClient.DB()),
	}
}

type services struct {
	user          handler.IUserService
	auth          handler.IAuthService
	session       handler.ISessionService
	course        handler.ICourseService
	adminHomework handler.IAdminHomeworkService
	adminTask     handler.IAdminTaskService
	adminRole     handler.IAdminRoleService
}

func newServices(repositories repositories) services {
	return services{
		user:          service.NewUserService(repositories.user),
		auth:          service.NewAuthService(repositories.user, repositories.session),
		session:       service.NewSessionService(repositories.session),
		course:        service.NewCourseService(repositories.course, repositories.role, repositories.studentScore),
		adminHomework: service.NewAdminHomeworkService(repositories.homework, repositories.deadline, repositories.role),
		adminTask:     service.NewAdminTaskService(repositories.task, repositories.homework, repositories.role),
		adminRole:     service.NewAdminRoleService(repositories.role, repositories.user),
	}
}

type httpHandlers struct {
	apiController *controller.APIController
	adminHomework *handler.AdminHomeworkHandler
	adminTask     *handler.AdminTaskHandler
	adminRole     *handler.AdminRoleHandler
}

func newHTTPHandlers(services services) httpHandlers {
	authHandler := handler.NewAuthHandler(services.auth)
	userHandler := handler.NewUserHandler(services.user)
	sessionHandler := handler.NewSessionHandler(services.session, services.user)
	courseHandler := handler.NewCourseHandler(services.course)

	return httpHandlers{
		apiController: controller.NewAPIController(
			authHandler,
			userHandler,
			sessionHandler,
			courseHandler,
		),
		adminHomework: handler.NewAdminHomeworkHandler(services.adminHomework),
		adminTask:     handler.NewAdminTaskHandler(services.adminTask),
		adminRole:     handler.NewAdminRoleHandler(services.adminRole),
	}
}

func registerHTTPMiddleware(e *echo.Echo, repositories repositories) {
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	e.Use(authmw.Auth(repositories.user, repositories.session, protectedPaths()))
	metrics.EchoPrometheus(e)
}

func registerHTTPRoutes(e *echo.Echo, handlers httpHandlers) {
	controller.RegisterRoutes(
		e,
		handlers.apiController,
		handlers.adminHomework,
		handlers.adminTask,
		handlers.adminRole,
	)
}

func protectedPaths() []string {
	return controller.ProtectedPaths()
}
