package service

import (
	"context"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type SessionService struct {
	sessionRepo repo.SessionRepositoryInterface
}

func NewSessionService(sessionRepo repo.SessionRepositoryInterface) *SessionService {
	return &SessionService{sessionRepo: sessionRepo}
}

func (s *SessionService) GetSessions(ctx context.Context, limit, offset int) ([]models.Session, int64, error) {
	limit, offset, err := ParsePagination(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.sessionRepo.CountSessions(ctx)
	if err != nil {
		return nil, 0, Internal("Failed to count sessions", err)
	}

	sessions, err := s.sessionRepo.GetSessionsWithUser(ctx, limit, offset)
	if err != nil {
		return nil, 0, Internal("Failed to get sessions", err)
	}

	return sessions, total, nil
}
