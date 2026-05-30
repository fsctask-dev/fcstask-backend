package service

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type NamespaceWithCounts struct {
	model.Namespace
	CoursesCount int64 `json:"coursesCount"`
	UsersCount   int64 `json:"usersCount"`
}

type NamespaceUserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
}

type NamespaceCourseInfo struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
	URL    string    `json:"url"`
}

type NamespaceService struct {
	repo repo.INamespaceRepo
}

func NewNamespaceService(r repo.INamespaceRepo) *NamespaceService {
	return &NamespaceService{repo: r}
}

func (s *NamespaceService) ListNamespaces(ctx context.Context) ([]NamespaceWithCounts, error) {
	namespaces, err := s.repo.ListNamespaces(ctx)
	if err != nil {
		return nil, Internal("Failed to list namespaces", err)
	}

	result := make([]NamespaceWithCounts, len(namespaces))
	for i, ns := range namespaces {
		courses, err := s.repo.CountCourses(ctx, ns.ID)
		if err != nil {
			return nil, Internal("Failed to count courses", err)
		}
		users, err := s.repo.CountUsers(ctx, ns.ID)
		if err != nil {
			return nil, Internal("Failed to count users", err)
		}
		result[i] = NamespaceWithCounts{
			Namespace:    ns,
			CoursesCount: courses,
			UsersCount:   users,
		}
	}
	return result, nil
}

func (s *NamespaceService) GetNamespace(ctx context.Context, id uuid.UUID) (*NamespaceWithCounts, error) {
	ns, err := s.repo.GetNamespaceByID(ctx, id)
	if err != nil {
		return nil, Internal("Failed to get namespace", err)
	}
	if ns == nil {
		return nil, NotFound("namespace not found")
	}

	courses, err := s.repo.CountCourses(ctx, id)
	if err != nil {
		return nil, Internal("Failed to count courses", err)
	}
	users, err := s.repo.CountUsers(ctx, id)
	if err != nil {
		return nil, Internal("Failed to count users", err)
	}

	return &NamespaceWithCounts{
		Namespace:    *ns,
		CoursesCount: courses,
		UsersCount:   users,
	}, nil
}

func (s *NamespaceService) GetNamespaceUsers(ctx context.Context, id uuid.UUID) ([]NamespaceUserInfo, error) {
	if _, err := s.GetNamespace(ctx, id); err != nil {
		return nil, err
	}

	rows, err := s.repo.GetNamespaceUsers(ctx, id)
	if err != nil {
		return nil, Internal("Failed to get namespace users", err)
	}

	result := make([]NamespaceUserInfo, len(rows))
	for i, r := range rows {
		result[i] = NamespaceUserInfo{
			ID:       r.UserID,
			Username: r.Username,
			Role:     r.Role,
		}
	}
	return result, nil
}

func (s *NamespaceService) GetNamespaceCourses(ctx context.Context, id uuid.UUID) ([]NamespaceCourseInfo, error) {
	if _, err := s.GetNamespace(ctx, id); err != nil {
		return nil, err
	}

	courses, err := s.repo.GetNamespaceCourses(ctx, id)
	if err != nil {
		return nil, Internal("Failed to get namespace courses", err)
	}

	result := make([]NamespaceCourseInfo, len(courses))
	for i, c := range courses {
		result[i] = NamespaceCourseInfo{
			ID:     c.ID,
			Name:   c.Name,
			Status: c.Status,
			URL:    c.URL,
		}
	}
	return result, nil
}
