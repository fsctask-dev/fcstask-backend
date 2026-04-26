package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	models "fcstask-backend/internal/db/model"
)

type courseServiceRepo struct {
	courses map[string]models.Course
	boards  map[string]models.TaskBoardSummary
}

func (r *courseServiceRepo) GetCourses(ctx context.Context) ([]models.Course, error) {
	courses := make([]models.Course, 0, len(r.courses))
	for _, course := range r.courses {
		courses = append(courses, course)
	}
	return courses, nil
}

func (r *courseServiceRepo) GetCourseByID(ctx context.Context, courseID string) (*models.Course, error) {
	course, ok := r.courses[courseID]
	if !ok {
		return nil, nil
	}
	return &course, nil
}

func (r *courseServiceRepo) CreateCourse(ctx context.Context, course models.Course) (*models.Course, error) {
	r.courses[course.ID] = course
	return &course, nil
}

func (r *courseServiceRepo) UpdateCourse(ctx context.Context, courseID string, course models.Course) (*models.Course, error) {
	r.courses[courseID] = course
	return &course, nil
}

func (r *courseServiceRepo) GetCourseBoard(ctx context.Context, courseID string) (*models.TaskBoardSummary, bool, error) {
	board, ok := r.boards[courseID]
	if !ok {
		return nil, false, nil
	}
	return &board, true, nil
}

func newCourseServiceRepo() *courseServiceRepo {
	return &courseServiceRepo{
		courses: map[string]models.Course{
			"go": {
				ID:           "go",
				Name:         "Go",
				Status:       "created",
				StartDate:    "2026-01-01",
				EndDate:      "2026-02-01",
				RepoTemplate: "git@test/go.git",
				Description:  "Go course",
				URL:          "/course/go",
			},
		},
		boards: map[string]models.TaskBoardSummary{},
	}
}

func TestCourseService_CreateCourseSuccess(t *testing.T) {
	repo := newCourseServiceRepo()
	svc := NewCourseService(repo)

	course, err := svc.CreateCourse(context.Background(), CourseInput{
		Name:         "Rust",
		Slug:         "rust",
		Status:       "created",
		StartDate:    "2026-03-01",
		EndDate:      "2026-04-01",
		RepoTemplate: "git@test/rust.git",
		Description:  "Rust course",
	})

	assert.NoError(t, err)
	assert.Equal(t, "rust", course.ID)
	assert.Equal(t, "/course/rust", course.URL)
	assert.Contains(t, repo.courses, "rust")
}

func TestCourseService_CreateCourseConflict(t *testing.T) {
	svc := NewCourseService(newCourseServiceRepo())

	course, err := svc.CreateCourse(context.Background(), CourseInput{
		Name:         "Go",
		Slug:         "go",
		Status:       "created",
		StartDate:    "2026-01-01",
		EndDate:      "2026-02-01",
		RepoTemplate: "git@test/go.git",
		Description:  "Go course",
	})

	assert.Nil(t, course)
	var serviceErr *Error
	assert.True(t, errors.As(err, &serviceErr))
	assert.Equal(t, "conflict", serviceErr.Code)
}

func TestCourseService_UpdateCoursePartial(t *testing.T) {
	repo := newCourseServiceRepo()
	svc := NewCourseService(repo)

	course, err := svc.UpdateCourse(context.Background(), "go", CourseInput{Name: "Go Advanced"})

	assert.NoError(t, err)
	assert.Equal(t, "Go Advanced", course.Name)
	assert.Equal(t, "created", course.Status)
	assert.Equal(t, "2026-01-01", course.StartDate)
}

func TestCourseService_GetCourseBoardEmpty(t *testing.T) {
	svc := NewCourseService(newCourseServiceRepo())

	board, err := svc.GetCourseBoard(context.Background(), "go")

	assert.NoError(t, err)
	assert.Equal(t, "Go", board.CourseName)
	assert.Equal(t, "created", board.CourseStatus)
	assert.Empty(t, board.Groups)
}
