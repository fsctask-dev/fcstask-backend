package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

func setupCheckerService() (*service.CheckerService, *MockTaskRepo, *MockHomeworkRepo, *MockScoreRepo, *MockHwDeadlineRepo, *MockCourseLateRepo) {
	taskRepo := new(MockTaskRepo)
	hwRepo := new(MockHomeworkRepo)
	scoreRepo := new(MockScoreRepo)
	hwDlRepo := new(MockHwDeadlineRepo)
	lateRepo := new(MockCourseLateRepo)
	svc := service.NewCheckerService(taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo)
	return svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo
}

func intPtr(v int) *int         { return &v }
func f64Ptr(v float64) *float64 { return &v }

func TestSubmitGrade_MissingStudentID(t *testing.T) {
	svc, _, _, _, _, _ := setupCheckerService()
	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		TaskID: uuid.New(), CourseID: uuid.New(), Status: "passed", SubmittedAt: time.Now(),
	})
	assert.ErrorContains(t, err, "student_id is required")
}

func TestSubmitGrade_MissingTaskID(t *testing.T) {
	svc, _, _, _, _, _ := setupCheckerService()
	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID: uuid.New(), CourseID: uuid.New(), Status: "passed", SubmittedAt: time.Now(),
	})
	assert.ErrorContains(t, err, "task_id is required")
}

func TestSubmitGrade_InvalidStatus(t *testing.T) {
	svc, _, _, _, _, _ := setupCheckerService()
	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID: uuid.New(), TaskID: uuid.New(), CourseID: uuid.New(),
		Status: "unknown", SubmittedAt: time.Now(),
	})
	assert.ErrorContains(t, err, "status must be")
}

func TestSubmitGrade_TaskNotFound(t *testing.T) {
	svc, taskRepo, _, _, _, _ := setupCheckerService()
	ctx := context.Background()
	taskID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(nil, assert.AnError)
	_, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: uuid.New(), TaskID: taskID, CourseID: uuid.New(),
		Status: "passed", SubmittedAt: time.Now(),
	})
	assert.ErrorContains(t, err, "Task not found")
}

func TestSubmitGrade_CourseMismatch(t *testing.T) {
	svc, taskRepo, hwRepo, _, _, _ := setupCheckerService()
	ctx := context.Background()
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil) // different course

	_, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: uuid.New(), TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: time.Now(),
	})
	assert.ErrorContains(t, err, "does not belong to the specified course")
}

func TestSubmitGrade_PassedNoLatePolicy(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(nil, assert.AnError) // no deadlines
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: time.Now(),
	})
	assert.NoError(t, err)
	assert.Equal(t, score, result.Score)
	assert.True(t, result.IsPassed)

	_ = lateRepo
}

func TestSubmitGrade_FailStatus(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, _ := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(nil, assert.AnError)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "fail", SubmittedAt: time.Now(),
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
	assert.False(t, result.IsPassed)
}

func TestSubmitGrade_LinearPolicy_BeforeSoft(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100
	now := time.Now()
	soft := now.Add(2 * time.Hour)
	hard := now.Add(4 * time.Hour)

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(&model.HomeworkDeadline{SoftDeadline: soft, HardDeadline: hard}, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(&model.CourseLatePolicy{
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.5,
	}, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: now,
	})
	assert.NoError(t, err)
	assert.Equal(t, 100, result.Score)
}

func TestSubmitGrade_LinearPolicy_AfterHard(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100
	now := time.Now()
	soft := now.Add(-4 * time.Hour)
	hard := now.Add(-2 * time.Hour)

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(&model.HomeworkDeadline{SoftDeadline: soft, HardDeadline: hard}, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(&model.CourseLatePolicy{
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.5,
	}, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: now,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
}

func TestSubmitGrade_LinearPolicy_BetweenSoftAndHard(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100
	now := time.Now()
	soft := now.Add(-2 * time.Hour)
	hard := now.Add(2 * time.Hour)

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(&model.HomeworkDeadline{SoftDeadline: soft, HardDeadline: hard}, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(&model.CourseLatePolicy{
		PolicyType: model.PolicyTypeLinear, HardDeadlineScore: 0.0,
	}, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: now,
	})
	assert.NoError(t, err)
	assert.Equal(t, 50, result.Score)
}

func TestSubmitGrade_StepPolicy(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100
	now := time.Now()
	soft := now.Add(-23 * time.Hour)
	hard := now.Add(48 * time.Hour)

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(&model.HomeworkDeadline{SoftDeadline: soft, HardDeadline: hard}, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(&model.CourseLatePolicy{
		PolicyType:  model.PolicyTypeStep,
		StepPercent: f64Ptr(0.1),
	}, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: now,
	})
	assert.NoError(t, err)
	assert.Equal(t, 90, result.Score)
}

func TestSubmitGrade_CoefficientPolicy(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100
	now := time.Now()
	soft := now.Add(-1 * time.Hour)
	hard := now.Add(48 * time.Hour)

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(&model.HomeworkDeadline{SoftDeadline: soft, HardDeadline: hard}, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(&model.CourseLatePolicy{
		PolicyType:  model.PolicyTypeCoefficient,
		Coefficient: f64Ptr(0.8),
	}, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: now,
	})
	assert.NoError(t, err)
	assert.Equal(t, 80, result.Score)
}

func TestSubmitGrade_LinearPolicy_WithSoftPenalty(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, hwDlRepo, lateRepo := setupCheckerService()
	ctx := context.Background()
	taskID, hwID, courseID, studentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	score := 100
	now := time.Now()
	soft := now.Add(-2 * time.Hour)
	hard := now.Add(2 * time.Hour)

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &score}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	hwDlRepo.On("GetByHwID", ctx, hwID).Return(&model.HomeworkDeadline{SoftDeadline: soft, HardDeadline: hard}, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(&model.CourseLatePolicy{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.1,
		HardDeadlineScore: 0.3,
	}, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	result, err := svc.SubmitGrade(ctx, service.SubmitGradeInput{
		StudentID: studentID, TaskID: taskID, CourseID: courseID,
		Status: "passed", SubmittedAt: now,
	})
	assert.NoError(t, err)
	assert.Equal(t, 60, result.Score)
}
