package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

func setupGradeUpdateService(allowPermission bool) (*service.GradeUpdateService, *MockScoreRepo, *MockRoleRepo) {
	taskRepo := new(MockTaskRepo)
	scoreRepo := new(MockScoreRepo)
	roleRepo := new(MockRoleRepo)
	roleID := uuid.New()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, mock.Anything).Return(allowPermission, nil)
	svc := service.NewGradeUpdateService(taskRepo, scoreRepo, roleRepo)
	return svc, scoreRepo, roleRepo
}

func TestUpdateGrade_MissingPermission(t *testing.T) {
	svc, _, roleRepo := setupGradeUpdateService(false)
	userID := uuid.New()
	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Score:     intPtr(85),
	}
	_, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.Error(t, err)
	roleRepo.AssertExpectations(t)
}

func TestUpdateGrade_InvalidStudentID(t *testing.T) {
	svc, _, _ := setupGradeUpdateService(true)
	userID := uuid.New()
	input := service.UpdateGradeInput{
		StudentID: uuid.Nil,
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Score:     intPtr(85),
	}
	_, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.ErrorContains(t, err, "student_id is required")
}

func TestUpdateGrade_InvalidTaskID(t *testing.T) {
	svc, _, _ := setupGradeUpdateService(true)
	userID := uuid.New()
	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.Nil,
		CourseID:  uuid.New(),
		Score:     intPtr(85),
	}
	_, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.ErrorContains(t, err, "task_id is required")
}

func TestUpdateGrade_InvalidCourseID(t *testing.T) {
	svc, _, _ := setupGradeUpdateService(true)
	userID := uuid.New()
	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.Nil,
		Score:     intPtr(85),
	}
	_, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.ErrorContains(t, err, "course_id is required")
}

func TestUpdateGrade_InvalidScore_Nil(t *testing.T) {
	svc, _, _ := setupGradeUpdateService(true)
	userID := uuid.New()
	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Score:     nil,
	}
	_, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.ErrorContains(t, err, "score is required")
}

func TestUpdateGrade_InvalidScore_Negative(t *testing.T) {
	svc, _, _ := setupGradeUpdateService(true)
	userID := uuid.New()
	negativeScore := -10
	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Score:     &negativeScore,
	}
	_, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.ErrorContains(t, err, "score must be 0 or higher")
}

func TestUpdateGrade_Success(t *testing.T) {
	svc, scoreRepo, roleRepo := setupGradeUpdateService(true)
	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 95

	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.StudentID == studentID &&
			s.TaskID == taskID &&
			s.CourseID == courseID &&
			s.Score == scoreValue
	})).Return(nil)

	input := service.UpdateGradeInput{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     &scoreValue,
	}
	result, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.NoError(t, err)
	assert.Equal(t, scoreValue, result.Score)

	scoreRepo.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

func TestUpdateGrade_UpsertError(t *testing.T) {
	svc, scoreRepo, _ := setupGradeUpdateService(true)
	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 75

	scoreRepo.On("Upsert", mock.Anything, mock.Anything).Return(assert.AnError)

	input := service.UpdateGradeInput{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     &scoreValue,
	}
	_, err := svc.UpdateGrade(context.Background(), userID, input)
	assert.ErrorContains(t, err, "Failed to update grade")
	scoreRepo.AssertExpectations(t)
}
