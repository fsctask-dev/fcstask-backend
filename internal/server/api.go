package server

import (
	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/repo"
)

type APIServer struct {
	db          *db.Client
	userRepo    repo.UserRepositoryInterface
	sessionRepo repo.SessionRepositoryInterface
}

func NewAPIServer(db *db.Client) *APIServer {
	userRepo := repo.NewUserRepository(db)
	sessionRepo := repo.NewSessionRepository(db)

	return &APIServer{
		db:          db,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *APIServer) UserRepo() repo.UserRepositoryInterface {
	return s.userRepo
}

func (s *APIServer) SessionRepo() repo.SessionRepositoryInterface {
	return s.sessionRepo
}
