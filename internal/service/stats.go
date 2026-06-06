package service

import (
	"context"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"

	"github.com/google/uuid"
)

type StatsService struct {
	statsRepo repo.StatsRepositoryInterface
	roleRepo  repo.IRoleRepo
}

func NewStatsService(statsRepo repo.StatsRepositoryInterface, roleRepo repo.IRoleRepo) *StatsService {
	return &StatsService{statsRepo: statsRepo, roleRepo: roleRepo}
}

func (s *StatsService) GetStats(ctx context.Context, userID uuid.UUID) (*models.PlatformStats, error) {
	if err := RequireScopedPermission(ctx, s.roleRepo, userID, uuid.Nil, PermissionStatsRead); err != nil {
		return nil, err
	}

	stats, err := s.statsRepo.GetStats(ctx)
	if err != nil {
		return nil, Internal("Failed to get platform stats", err)
	}
	return stats, nil
}
