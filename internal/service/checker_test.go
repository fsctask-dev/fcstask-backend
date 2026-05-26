package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type MockStudentTaskScoreRepo struct {
	mock.Mock
}

func (m *MockStudentTaskScoreRepo) Upsert(ctx context.Context, score *model.StudentTaskScore) error {
	return m.Called(ctx, score).Error(0)
}

func (m *MockStudentTaskScoreRepo) GetByStudentAndCourse(ctx context.Context, studentID, courseID uuid.UUID) ([]model.StudentTaskScore, error) {
	args := m.Called(ctx, studentID, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StudentTaskScore), args.Error(1)
}

func setupCheckerService() (*service.CheckerService, *MockTaskRepo, *MockStudentTaskScoreRepo, *MockLatePolicyRepo, *MockHomeworkRepo) {
	taskRepo := new(MockTaskRepo)
	scoreRepo := new(MockStudentTaskScoreRepo)
	lpRepo := new(MockLatePolicyRepo)
	hwRepo := new(MockHomeworkRepo)
	svc := service.NewCheckerService(taskRepo, hwRepo, scoreRepo, lpRepo)
	return svc, taskRepo, scoreRepo, lpRepo, hwRepo
}

func TestCheckerSubmitGrade_MissingStudentID(t *testing.T) {
	svc, _, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.Nil,
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "student_id")
}

func TestCheckerSubmitGrade_MissingTaskID(t *testing.T) {
	svc, _, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.Nil,
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task_id")
}

func TestCheckerSubmitGrade_MissingCourseID(t *testing.T) {
	svc, _, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.Nil,
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "course_id")
}

func TestCheckerSubmitGrade_InvalidStatus(t *testing.T) {
	svc, _, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "unknown",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestCheckerSubmitGrade_TaskNotFound(t *testing.T) {
	svc, taskRepo, _, _, _ := setupCheckerService()

	taskID := uuid.New()
	taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, errors.New("not found"))

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Task not found")
	taskRepo.AssertExpectations(t)
}

func TestCheckerSubmitGrade_HomeworkNotFound(t *testing.T) {
	svc, taskRepo, _, _, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(nil, errors.New("not found"))

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Homework not found")
	taskRepo.AssertExpectations(t)
	hwRepo.AssertExpectations(t)
}

func TestCheckerSubmitGrade_TaskNotBelongToCourse(t *testing.T) {
	svc, taskRepo, _, _, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	otherCourseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(&model.Homework{HwID: hwID, CourseID: otherCourseID}, nil)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong")
	taskRepo.AssertExpectations(t)
	hwRepo.AssertExpectations(t)
}

func TestCheckerSubmitGrade_Fail_ScoreZero(t *testing.T) {
	svc, taskRepo, scoreRepo, _, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0 && !s.IsPassed
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "fail",
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
	assert.False(t, result.IsPassed)
	scoreRepo.AssertExpectations(t)
}

func TestCheckerSubmitGrade_Passed_NoLatePolicy(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	maxScore := 100

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &maxScore}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	lpRepo.On("GetByHwID", mock.Anything, hwID).
		Return(nil, errors.New("not found"))
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 100 && s.IsPassed
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 100, result.Score)
	assert.True(t, result.IsPassed)
	scoreRepo.AssertExpectations(t)
}

func TestCheckerSubmitGrade_Passed_BeforeSoftDeadline(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	maxScore := 100
	softDeadline := time.Now().Add(24 * time.Hour)
	hardDeadline := time.Now().Add(48 * time.Hour)

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &maxScore}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	lpRepo.On("GetByHwID", mock.Anything, hwID).Return(&model.LatePolicy{
		HwID:         hwID,
		SoftDeadline: softDeadline,
		HardDeadline: hardDeadline,
		SoftPenalty:  0.7,
		HardPenalty:  0.0,
	}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 100
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 100, result.Score)
}

func TestCheckerSubmitGrade_Passed_BetweenSoftAndHard(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	maxScore := 100
	softDeadline := time.Now().Add(-24 * time.Hour)
	hardDeadline := time.Now().Add(24 * time.Hour)

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &maxScore}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	lpRepo.On("GetByHwID", mock.Anything, hwID).Return(&model.LatePolicy{
		HwID:         hwID,
		SoftDeadline: softDeadline,
		HardDeadline: hardDeadline,
		SoftPenalty:  0.7,
		HardPenalty:  0.0,
	}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 70
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 70, result.Score)
}

func TestCheckerSubmitGrade_Passed_AfterHardDeadline(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	maxScore := 100
	softDeadline := time.Now().Add(-48 * time.Hour)
	hardDeadline := time.Now().Add(-24 * time.Hour)

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &maxScore}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	lpRepo.On("GetByHwID", mock.Anything, hwID).Return(&model.LatePolicy{
		HwID:         hwID,
		SoftDeadline: softDeadline,
		HardDeadline: hardDeadline,
		SoftPenalty:  0.7,
		HardPenalty:  0.0,
	}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
}

func TestCheckerSubmitGrade_UpsertError(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo, hwRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	maxScore := 100

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &maxScore}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).
		Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	lpRepo.On("GetByHwID", mock.Anything, hwID).
		Return(nil, errors.New("not found"))
	scoreRepo.On("Upsert", mock.Anything, mock.Anything).
		Return(errors.New("db error"))

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save grade")
	scoreRepo.AssertExpectations(t)
}
