package service

import (
	"context"

	"github.com/google/uuid"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type NamespaceDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Description   *string `json:"description,omitempty"`
	GitlabGroupID *string `json:"gitlabGroupId,omitempty"`
	UsersCount    int     `json:"usersCount"`
	CoursesCount  int     `json:"coursesCount"`
}

type NamespaceUserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type NamespaceCourseDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

type NamespaceService struct {
	nsRepo   repo.NamespaceRepositoryInterface
	userRepo repo.IUserRepo
}

func NewNamespaceService(
	nsRepo repo.NamespaceRepositoryInterface,
	userRepo repo.IUserRepo,
) *NamespaceService {
	return &NamespaceService{
		nsRepo:   nsRepo,
		userRepo: userRepo,
	}
}

func (s *NamespaceService) GetNamespaces(ctx context.Context, userID uuid.UUID) ([]NamespaceDTO, error) {
	namespaces, err := s.nsRepo.GetAll(ctx)
	if err != nil {
		return nil, Internal("Failed to fetch namespaces", err)
	}

	nsDTOs := make([]NamespaceDTO, 0, len(namespaces))
	for _, ns := range namespaces {
		users, err := s.nsRepo.GetUsers(ctx, ns.ID)
		if err != nil {
			return nil, Internal("Failed to fetch namespace users", err)
		}

		nsDTOs = append(nsDTOs, NamespaceDTO{
			ID:            ns.ID.String(),
			Name:          ns.Name,
			Slug:          ns.Slug,
			Description:   ns.Description,
			GitlabGroupID: ns.GitlabGroupID,
			UsersCount:    len(users),
			CoursesCount:  0,
		})
	}

	return nsDTOs, nil
}

func (s *NamespaceService) GetNamespace(ctx context.Context, nsID uuid.UUID) (*models.Namespace, error) {
	ns, err := s.nsRepo.GetByID(ctx, nsID)
	if err != nil {
		return nil, Internal("Failed to fetch namespace", err)
	}
	if ns == nil {
		return nil, NotFound("namespace not found")
	}
	return ns, nil
}

func (s *NamespaceService) GetNamespaceUsers(ctx context.Context, nsID uuid.UUID) ([]NamespaceUserDTO, error) {
	users, err := s.nsRepo.GetUsers(ctx, nsID)
	if err != nil {
		return nil, Internal("Failed to fetch namespace users", err)
	}

	result := make([]NamespaceUserDTO, 0, len(users))
	for _, nu := range users {
		user, err := s.userRepo.GetUserByID(ctx, nu.UserID)
		if err != nil || user == nil {
			continue
		}
		result = append(result, NamespaceUserDTO{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     nu.Role,
		})
	}
	return result, nil
}

func (s *NamespaceService) GetNamespaceCourses(ctx context.Context, nsID uuid.UUID) ([]NamespaceCourseDTO, error) {
	return []NamespaceCourseDTO{}, nil
}
