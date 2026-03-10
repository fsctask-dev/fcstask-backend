package server

import (
	"fcstask/internal/db"
	"fcstask/internal/db/repo"
)

type APIServer struct {
	db          *db.Client
	userRepo    repo.UserRepositoryInterface
	sessionRepo repo.SessionRepositoryInterface
}

func NewAPIServer(db *db.Client) *APIServer {
	userRepo := repo.NewUserRepository(db.DB())
	sessionRepo := repo.NewSessionRepository(db.DB())

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
