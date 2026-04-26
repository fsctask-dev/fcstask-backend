package server

import (
	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/oauth"
	"fcstask-backend/internal/server/handler"
)

type APIServer struct {
	db               *db.Client
	userRepo         repo.IUserRepo
	sessionRepo      repo.SessionRepositoryInterface
	identityRepo     repo.IOAuthIdentityRepo
	registrationRepo repo.IRegistrationSessionRepo
	courseRepo       repo.CourseRepositoryInterface
	courseHandler    *handler.CourseHandler
	oauthHandler     *handler.OAuthHandler
}

func NewAPIServer(db *db.Client, oauthCfg config.OAuthConfig) *APIServer {
	userRepo := repo.NewUserRepository(db.DB())
	sessionRepo := repo.NewSessionRepository(db.DB())
	courseRepo := repo.NewCourseRepository(db.DB())
	identityRepo := repo.NewOAuthIdentityRepository(db.DB())
	registrationRepo := repo.NewRegistrationSessionRepository(db.DB())

	registry := oauth.NewRegistry(
		oauth.NewGitLabProvider(oauthCfg.GitLab),
		oauth.NewGoogleProvider(oauthCfg.Google),
		oauth.NewTelegramProvider(oauthCfg.Telegram),
	)

	return &APIServer{
		db:               db,
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		identityRepo:     identityRepo,
		registrationRepo: registrationRepo,
		courseRepo:       courseRepo,
		courseHandler:    handler.NewCourseHandler(courseRepo),
		oauthHandler:     handler.NewOAuthHandler(userRepo, sessionRepo, identityRepo, registrationRepo, registry),
	}
}

func (s *APIServer) UserRepo() repo.IUserRepo {
	return s.userRepo
}

func (s *APIServer) SessionRepo() repo.SessionRepositoryInterface {
	return s.sessionRepo
}
