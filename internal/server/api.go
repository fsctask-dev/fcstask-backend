package server

import (
	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/server/handler"
)

type APIServer struct {
	db            *db.Client
	userRepo      repo.IUserRepo
	sessionRepo   repo.SessionRepositoryInterface
	courseRepo    repo.CourseRepositoryInterface
	courseHandler *handler.CourseHandler
}

func NewAPIServer(db *db.Client) *APIServer {
	userRepo := repo.NewUserRepository(db)
	sessionRepo := repo.NewSessionRepository(db)
	courseRepo := repo.NewCourseRepository(db.DB())

	return &APIServer{
		db:            db,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		courseRepo:    courseRepo,
		courseHandler: handler.NewCourseHandler(courseRepo),
	}
}

func (s *APIServer) UserRepo() repo.IUserRepo {
	return s.userRepo
}

func (s *APIServer) SessionRepo() repo.SessionRepositoryInterface {
	return s.sessionRepo
}
