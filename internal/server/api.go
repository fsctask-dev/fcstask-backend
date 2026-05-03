package server

import (
	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/mailer"
	"fcstask-backend/internal/oauth"
	"fcstask-backend/internal/server/handler"
)

type APIServer struct {
	db                   *db.Client
	userRepo             repo.IUserRepo
	sessionRepo          repo.SessionRepositoryInterface
	identityRepo         repo.IOAuthIdentityRepo
	registrationRepo     repo.IRegistrationSessionRepo
	emailRegRepo         repo.IEmailRegistrationRepo
	passwordResetRepo    repo.IPasswordResetRepo
	courseRepo           repo.CourseRepositoryInterface
	courseHandler        *handler.CourseHandler
	oauthHandler         *handler.OAuthHandler
	signUpHandler        *handler.SignUpHandler
	passwordResetHandler *handler.PasswordResetHandler
}

func NewAPIServer(db *db.Client, oauthCfg config.OAuthConfig, mailerCfg config.MailerConfig) *APIServer {
	userRepo := repo.NewUserRepository(db.DB())
	sessionRepo := repo.NewSessionRepository(db.DB())
	courseRepo := repo.NewCourseRepository(db.DB())
	identityRepo := repo.NewOAuthIdentityRepository(db.DB())
	registrationRepo := repo.NewRegistrationSessionRepository(db.DB())
	emailRegRepo := repo.NewEmailRegistrationRepository(db.DB())
	passwordResetRepo := repo.NewPasswordResetRepository(db.DB())

	registry := oauth.NewRegistry(
		oauth.NewGitLabProvider(oauthCfg.GitLab),
		oauth.NewGoogleProvider(oauthCfg.Google),
		oauth.NewTelegramProvider(oauthCfg.Telegram),
	)

	var m mailer.Mailer
	if mailerCfg.Enabled {
		m = mailer.NewSMTPMailer(mailerCfg)
	} else {
		m = mailer.NewLogMailer()
	}

	return &APIServer{
		db:                   db,
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		identityRepo:         identityRepo,
		registrationRepo:     registrationRepo,
		emailRegRepo:         emailRegRepo,
		passwordResetRepo:    passwordResetRepo,
		courseRepo:           courseRepo,
		courseHandler:        handler.NewCourseHandler(courseRepo),
		oauthHandler:         handler.NewOAuthHandler(userRepo, sessionRepo, identityRepo, registrationRepo, registry),
		signUpHandler:        handler.NewSignUpHandler(userRepo, sessionRepo, emailRegRepo, m, mailerCfg),
		passwordResetHandler: handler.NewPasswordResetHandler(userRepo, sessionRepo, passwordResetRepo, m, mailerCfg),
	}
}

func (s *APIServer) UserRepo() repo.IUserRepo {
	return s.userRepo
}

func (s *APIServer) SessionRepo() repo.SessionRepositoryInterface {
	return s.sessionRepo
}

func (s *APIServer) EmailRegRepo() repo.IEmailRegistrationRepo {
	return s.emailRegRepo
}

func (s *APIServer) PasswordResetRepo() repo.IPasswordResetRepo {
	return s.passwordResetRepo
}
