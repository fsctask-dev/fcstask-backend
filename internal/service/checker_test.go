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

func setupCheckerService() (*service.CheckerService, *MockTaskRepo, *MockStudentTaskScoreRepo, *MockLatePolicyRepo) {
	taskRepo := new(MockTaskRepo)
	scoreRepo := new(MockStudentTaskScoreRepo)
	lpRepo := new(MockLatePolicyRepo)
	svc := service.NewCheckerService(taskRepo, scoreRepo, lpRepo)
	return svc, taskRepo, scoreRepo, lpRepo
}

func TestCheckerSubmitGrade_MissingStudentID(t *testing.T) {
	svc, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.Nil,
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "student_id")
}

func TestCheckerSubmitGrade_MissingTaskID(t *testing.T) {
	svc, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.Nil,
		CourseID:    uuid.New(),
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task_id")
}

func TestCheckerSubmitGrade_MissingCourseID(t *testing.T) {
	svc, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.Nil,
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "course_id")
}

func TestCheckerSubmitGrade_NegativeScore(t *testing.T) {
	svc, _, _, _ := setupCheckerService()

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		RawScore:    -1,
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative")
}

func TestCheckerSubmitGrade_TaskNotFound(t *testing.T) {
	svc, taskRepo, _, _ := setupCheckerService()

	taskID := uuid.New()
	taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, errors.New("not found"))

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    uuid.New(),
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	taskRepo.AssertExpectations(t)
}

func TestCheckerSubmitGrade_Success_NoLatePolicy(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	studentID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	lpRepo.On("GetByHwID", mock.Anything, hwID).
		Return(nil, errors.New("not found"))
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.StudentID == studentID &&
			s.TaskID == taskID &&
			s.CourseID == courseID &&
			s.Score == 100
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		RawScore:    100,
		IsPassed:    true,
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 100, result.Score)
	assert.True(t, result.IsPassed)
	taskRepo.AssertExpectations(t)
	scoreRepo.AssertExpectations(t)
}

func TestCheckerSubmitGrade_Success_BeforeSoftDeadline(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	softDeadline := time.Now().Add(24 * time.Hour)
	hardDeadline := time.Now().Add(48 * time.Hour)

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
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
		CourseID:    uuid.New(),
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 100, result.Score)
}

func TestCheckerSubmitGrade_Success_BetweenSoftAndHard(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	softDeadline := time.Now().Add(-24 * time.Hour) // уже прошёл
	hardDeadline := time.Now().Add(24 * time.Hour)  // ещё не прошёл

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
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
		CourseID:    uuid.New(),
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 70, result.Score)
}

func TestCheckerSubmitGrade_Success_AfterHardDeadline(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()
	softDeadline := time.Now().Add(-48 * time.Hour) // прошёл
	hardDeadline := time.Now().Add(-24 * time.Hour) // тоже прошёл

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
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
		CourseID:    uuid.New(),
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
}

func TestCheckerSubmitGrade_UpsertError(t *testing.T) {
	svc, taskRepo, scoreRepo, lpRepo := setupCheckerService()

	taskID := uuid.New()
	hwID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).
		Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	lpRepo.On("GetByHwID", mock.Anything, hwID).
		Return(nil, errors.New("not found"))
	scoreRepo.On("Upsert", mock.Anything, mock.Anything).
		Return(errors.New("db error"))

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    uuid.New(),
		RawScore:    100,
		SubmittedAt: time.Now(),
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save grade")
	scoreRepo.AssertExpectations(t)
}
