package service

import (
	"context"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type NamespaceService struct {
	namespaceRepo repo.NamespaceRepositoryInterface
	userRepo      repo.IUserRepo
	courseRepo    repo.CourseRepositoryInterface
}

func NewNamespaceService(namespaceRepo repo.NamespaceRepositoryInterface, userRepo repo.IUserRepo, courseRepo repo.CourseRepositoryInterface) *NamespaceService {
	return &NamespaceService{
		namespaceRepo: namespaceRepo,
		userRepo:      userRepo,
		courseRepo:    courseRepo,
	}
}

func (s *NamespaceService) GetNamespaces(ctx context.Context) ([]model.NamespaceResponse, error) {
	namespaces, err := s.namespaceRepo.GetNamespaces(ctx)
	if err != nil {
		return nil, Internal("Failed to get namespaces", err)
	}
	
	// Convert to API response format
	result := make([]model.NamespaceResponse, len(namespaces))
	for i, ns := range namespaces {
		desc := ""
		if ns.Description != nil {
			desc = *ns.Description
		}
		gitlabID := ""
		if ns.GitlabGroupID != nil {
			gitlabID = *ns.GitlabGroupID
		}
		result[i] = model.NamespaceResponse{
			ID:            ns.ID.String(),
			Name:          ns.Name,
			Slug:          ns.Slug,
			Description:   desc,
			GitlabGroupID: gitlabID,
			CoursesCount:  0,
			UsersCount:    0,
		}
	}
	
	return result, nil
}

func (s *NamespaceService) GetNamespace(ctx context.Context, id string) (*model.NamespaceDetailResponse, error) {
	namespace, err := s.namespaceRepo.GetNamespaceByID(ctx, id)
	if err != nil {
		return nil, Internal("Failed to get namespace", err)
	}
	if namespace == nil {
		return nil, NotFound("namespace not found")
	}

	desc := ""
	if namespace.Description != nil {
		desc = *namespace.Description
	}
	gitlabID := ""
	if namespace.GitlabGroupID != nil {
		gitlabID = *namespace.GitlabGroupID
	}

	return &model.NamespaceDetailResponse{
		Namespace: model.NamespaceResponse{
			ID:            namespace.ID.String(),
			Name:          namespace.Name,
			Slug:          namespace.Slug,
			Description:   desc,
			GitlabGroupID: gitlabID,
			CoursesCount:  0,
			UsersCount:    0,
		},
		Users:   []model.NamespaceUser{},
		Courses: []model.NamespaceCourse{},
	}, nil
}

func (s *NamespaceService) GetInstanceSummary(ctx context.Context) (*model.InstanceSummary, error) {
	summary, err := s.namespaceRepo.GetInstanceSummary(ctx)
	if err != nil {
		return nil, Internal("Failed to get instance summary", err)
	}
	return summary, nil
}

func (s *NamespaceService) GetCourseScores(ctx context.Context, courseID string) ([]model.ScoreResponse, error) {
	scores, err := s.namespaceRepo.GetCourseScores(ctx, courseID)
	if err != nil {
		return nil, Internal("Failed to get course scores", err)
	}
	
	// Convert to API response format
	result := make([]model.ScoreResponse, len(scores))
	for i, sc := range scores {
		submitted := ""
		if sc.SubmittedAt != nil {
			submitted = sc.SubmittedAt.Format("2006-01-02")
		}
		result[i] = model.ScoreResponse{
			ID:        sc.ID,
			Student:   sc.Student,
			Score:     sc.Score,
			Submitted: submitted,
		}
	}
	
	return result, nil
}