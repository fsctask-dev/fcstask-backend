package repo

import (
	"context"
	"fcstask-backend/internal/db"
	models "fcstask-backend/internal/db/model"
	"fmt"
)

type StatsRepositoryInterface interface {
	GetStats(ctx context.Context) (*models.PlatformStats, error)
}

type StatsRepository struct {
	rw db.ReadWriter
}

func NewStatsRepository(rw db.ReadWriter) *StatsRepository {
	return &StatsRepository{rw: rw}
}

func (r *StatsRepository) GetStats(ctx context.Context) (*models.PlatformStats, error) {
	var stats models.PlatformStats

	if err := r.rw.ReadDB().WithContext(ctx).Model(&models.Course{}).Count(&stats.TotalCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count total courses: %w", err)
	}

	if err := r.rw.ReadDB().WithContext(ctx).Model(&models.Course{}).
		Where("type = ?", models.CourseTypePublic).Count(&stats.PublicCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count public courses: %w", err)
	}

	if err := r.rw.ReadDB().WithContext(ctx).Model(&models.Course{}).
		Where("type = ?", models.CourseTypePrivate).Count(&stats.PrivateCourses).Error; err != nil {
		return nil, fmt.Errorf("failed to count private courses: %w", err)
	}

	if err := r.rw.ReadDB().WithContext(ctx).Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	return &stats, nil
}
