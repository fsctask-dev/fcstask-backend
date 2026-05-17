package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db"
	models "fcstask-backend/internal/db/model"
)

type NamespaceRepositoryInterface interface {
	GetNamespaces(ctx context.Context) ([]models.Namespace, error)
	GetNamespaceByID(ctx context.Context, id string) (*models.Namespace, error)
	GetInstanceSummary(ctx context.Context) (*models.InstanceSummary, error)
	GetCourseScores(ctx context.Context, courseID string) ([]models.Score, error)
}

type NamespaceRepository struct {
	rw db.ReadWriter
}

func NewNamespaceRepository(rw db.ReadWriter) NamespaceRepositoryInterface {
	return &NamespaceRepository{rw: rw}
}

func (r *NamespaceRepository) GetNamespaces(ctx context.Context) ([]models.Namespace, error) {
	var namespaces []models.Namespace
	if err := r.rw.ReadDB().WithContext(ctx).Find(&namespaces).Error; err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}
	return namespaces, nil
}

func (r *NamespaceRepository) GetNamespaceByID(ctx context.Context, id string) (*models.Namespace, error) {
	var ns models.Namespace
	query := r.rw.ReadDB().WithContext(ctx)
	
	if parsedID, err := uuid.Parse(id); err == nil {
		query = query.Where("id = ?", parsedID)
	} else {
		query = query.Where("slug = ?", id)
	}
	
	if err := query.First(&ns).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}
	
	return &ns, nil
}

func (r *NamespaceRepository) GetInstanceSummary(ctx context.Context) (*models.InstanceSummary, error) {
	var totalCourses, totalUsers, totalNamespaces int64

	r.rw.ReadDB().WithContext(ctx).Model(&models.Course{}).Count(&totalCourses)
	r.rw.ReadDB().WithContext(ctx).Model(&models.User{}).Count(&totalUsers)
	r.rw.ReadDB().WithContext(ctx).Model(&models.Namespace{}).Count(&totalNamespaces)

	return &models.InstanceSummary{
		TotalCourses:    int(totalCourses),
		TotalUsers:      int(totalUsers),
		TotalNamespaces: int(totalNamespaces),
		HealthStatus:    "ok",
	}, nil
}

func (r *NamespaceRepository) GetCourseScores(ctx context.Context, courseID string) ([]models.Score, error) {
	var scores []models.Score

	query := r.rw.ReadDB().WithContext(ctx)
	
	if parsedID, err := uuid.Parse(courseID); err == nil {
		query = query.Where("course_id = ?", parsedID)
	} else {
		// Try to find course by slug and get its ID
		var course models.Course
		if err := r.rw.ReadDB().WithContext(ctx).Where("slug = ?", courseID).First(&course).Error; err == nil {
			query = query.Where("course_id = ?", course.ID)
		} else {
			return []models.Score{}, nil
		}
	}

	if err := query.Order("score DESC").Find(&scores).Error; err != nil {
		return nil, fmt.Errorf("failed to get scores: %w", err)
	}

	return scores, nil
}