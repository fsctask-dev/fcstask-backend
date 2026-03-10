package server

import (
	"fcstask/internal/db"
	"fcstask/internal/db/repo"
)

type APIServer struct {
	db       *db.Client
	userRepo repo.UserRepositoryInterface
}

func NewAPIServer(db *db.Client) *APIServer {
	userRepo := repo.NewUserRepository(db.DB())

	return &APIServer{
		db:       db,
		userRepo: userRepo,
	}
}
