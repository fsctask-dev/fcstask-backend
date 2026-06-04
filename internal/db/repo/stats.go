package repo

import (
	"context"
	"fcstask-backend/internal/db"
	models "fcstask-backend/internal/db/model"
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

	// Всего курсов
	r.rw.ReadDB().WithContext(ctx).Model(&models.Course{}).Count(&stats.TotalCourses)

	// Публичных
	r.rw.ReadDB().WithContext(ctx).Model(&models.Course{}).
		Where("type = ?", models.CourseTypePublic).Count(&stats.PublicCourses)

	// Приватных
	r.rw.ReadDB().WithContext(ctx).Model(&models.Course{}).
		Where("type = ?", models.CourseTypePrivate).Count(&stats.PrivateCourses)

	// Пользователей
	r.rw.ReadDB().WithContext(ctx).Model(&models.User{}).Count(&stats.TotalUsers)

	return &stats, nil
}
