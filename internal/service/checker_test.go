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

func setupCheckerService() (*service.CheckerService, *MockTaskRepo, *MockHomeworkRepo, *MockScoreRepo, *MockDeadlineRepo, *MockCourseLateRepo, *MockRoleRepo) {
	taskRepo := new(MockTaskRepo)
	hwRepo := new(MockHomeworkRepo)
	scoreRepo := new(MockScoreRepo)
	deadlineRepo := new(MockDeadlineRepo)
	lateRepo := new(MockCourseLateRepo)
	roleRepo := new(MockRoleRepo)

	roleID := uuid.New()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, service.PermissionTaskSubmit).Return(true, nil)
	deadlineRepo.On("GetByHomeworkID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	svc := service.NewCheckerService(taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, roleRepo)
	return svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, roleRepo
}

func intPtr(v int) *int         { return &v }
func f64Ptr(v float64) *float64 { return &v }

func TestSubmitGrade_Success_Passed(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 85
	submittedAt := time.Now().Add(-24 * time.Hour)

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.StudentID == studentID && s.TaskID == taskID && s.Score == scoreValue && s.IsPassed == true
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, scoreValue, result.Score)
	assert.True(t, result.IsPassed)

	taskRepo.AssertExpectations(t)
	hwRepo.AssertExpectations(t)
	scoreRepo.AssertExpectations(t)
}

func TestSubmitGrade_Success_Fail(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	submittedAt := time.Now().Add(-24 * time.Hour)

	task := &model.Task{TaskID: taskID, HwID: hwID}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0 && s.IsPassed == false
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "fail",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Score)
	assert.False(t, result.IsPassed)

	taskRepo.AssertExpectations(t)
	hwRepo.AssertExpectations(t)
	scoreRepo.AssertExpectations(t)
}

func TestSubmitGrade_MissingStudentID(t *testing.T) {
	svc, _, _, _, _, _, _ := setupCheckerService()
	ctx := context.Background()

	input := service.SubmitGradeInput{
		StudentID:   uuid.Nil,
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now(),
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "You don't have permission to access this resource")
}

func TestSubmitGrade_MissingTaskID(t *testing.T) {
	svc, _, _, _, _, _, _ := setupCheckerService()
	ctx := context.Background()

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.Nil,
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now(),
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task_id is required")
}

func TestSubmitGrade_MissingCourseID(t *testing.T) {
	svc, _, _, _, _, _, _ := setupCheckerService()
	ctx := context.Background()

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.Nil,
		Status:      "passed",
		SubmittedAt: time.Now(),
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestSubmitGrade_InvalidStatus(t *testing.T) {
	svc, _, _, _, _, _, _ := setupCheckerService()
	ctx := context.Background()

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "invalid",
		SubmittedAt: time.Now(),
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "status must be 'passed' or 'fail'")
}

func TestSubmitGrade_MissingSubmittedAt(t *testing.T) {
	svc, _, _, _, _, _, _ := setupCheckerService()
	ctx := context.Background()

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Time{},
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "submitted_at is required")
}

func TestSubmitGrade_FutureSubmittedAt(t *testing.T) {
	svc, _, _, _, _, _, _ := setupCheckerService()
	ctx := context.Background()

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(24 * time.Hour),
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "submitted_at cannot be in the future")
}

func TestSubmitGrade_TaskNotFound(t *testing.T) {
	svc, taskRepo, _, _, _, _, _ := setupCheckerService()
	ctx := context.Background()

	taskRepo.On("GetByID", ctx, mock.Anything).Return(nil, assert.AnError)

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(-24 * time.Hour),
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Task not found")
	taskRepo.AssertExpectations(t)
}

func TestSubmitGrade_HomeworkNotFound(t *testing.T) {
	svc, taskRepo, hwRepo, _, _, _, _ := setupCheckerService()
	ctx := context.Background()
	taskID := uuid.New()
	hwID := uuid.New()

	task := &model.Task{TaskID: taskID, HwID: hwID}
	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(nil, assert.AnError)

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(-24 * time.Hour),
	}

	result, err := svc.SubmitGrade(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Homework not found")
	taskRepo.AssertExpectations(t)
	hwRepo.AssertExpectations(t)
}

func TestSubmitGrade_TaskCourseMismatch(t *testing.T) {
	svc, taskRepo, hwRepo, _, _, _, _ := setupCheckerService()
	ctx := context.Background()
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	otherCourseID := uuid.New()

	task := &model.Task{TaskID: taskID, HwID: hwID}
	hw := &model.Homework{HwID: hwID, CourseID: otherCourseID}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)

	input := service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now().Add(-24 * time.Hour),
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task does not belong to the specified course")
}

func TestSubmitGrade_UpsertError(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	submittedAt := time.Now().Add(-24 * time.Hour)

	task := &model.Task{TaskID: taskID, HwID: hwID}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(assert.AnError)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "fail",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to save grade")
}

func TestSubmitGrade_LatePolicy_NoPenalty_OnTime(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	softDeadline := time.Now().Add(-48 * time.Hour)
	submittedAt := softDeadline.Add(-1 * time.Hour) // before soft deadline

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: softDeadline, HardDeadline: softDeadline.Add(72 * time.Hour)}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeLinear, SoftPenalty: 0.2, HardDeadlineScore: 0.5}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(policy, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == scoreValue
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, scoreValue, result.Score)
}

func TestSubmitGrade_LatePolicy_LinearPenalty(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	softDeadline := time.Now().Add(-48 * time.Hour)
	hardDeadline := softDeadline.Add(48 * time.Hour)
	submittedAt := softDeadline.Add(24 * time.Hour)

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: softDeadline, HardDeadline: hardDeadline}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeLinear, SoftPenalty: 0.2, HardDeadlineScore: 0.5}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(policy, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 65
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, 65, result.Score)
}

func TestSubmitGrade_LatePolicy_StepPenalty(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	softDeadline := time.Now().Add(-72 * time.Hour)
	hardDeadline := softDeadline.Add(120 * time.Hour)
	submittedAt := softDeadline.Add(48 * time.Hour) // 2 days late
	stepPercent := 0.1

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: softDeadline, HardDeadline: hardDeadline}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeStep, StepPercent: &stepPercent}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(policy, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 80
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, 80, result.Score)
}

func TestSubmitGrade_LatePolicy_StepPenalty_MinZero(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	softDeadline := time.Now().Add(-720 * time.Hour)
	hardDeadline := softDeadline.Add(720 * time.Hour)
	submittedAt := softDeadline.Add(720 * time.Hour)
	stepPercent := 0.1

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: softDeadline, HardDeadline: hardDeadline}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeStep, StepPercent: &stepPercent}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(policy, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
}

func TestSubmitGrade_LatePolicy_CoefficientPenalty(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	softDeadline := time.Now().Add(-48 * time.Hour)
	hardDeadline := softDeadline.Add(48 * time.Hour)
	submittedAt := softDeadline.Add(24 * time.Hour)
	coefficient := 0.75

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: softDeadline, HardDeadline: hardDeadline}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeCoefficient, Coefficient: &coefficient}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(policy, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 75
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, 75, result.Score)
}

func TestSubmitGrade_LatePolicy_AfterHardDeadline_ZeroScore(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	softDeadline := time.Now().Add(-96 * time.Hour)
	hardDeadline := softDeadline.Add(48 * time.Hour)
	submittedAt := hardDeadline.Add(1 * time.Hour) // after hard deadline

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: softDeadline, HardDeadline: hardDeadline}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeLinear}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(policy, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
}

func TestSubmitGrade_LatePolicy_TaskScoreNil_ReturnsZero(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	submittedAt := time.Now().Add(-24 * time.Hour)

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: nil}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
}

func TestSubmitGrade_LatePolicy_DeadlineRepoError_Fallback(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, _, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	submittedAt := time.Now()

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(nil, assert.AnError)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == scoreValue
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, scoreValue, result.Score)
}

func TestSubmitGrade_LatePolicy_LatePolicyRepoError_Fallback(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, deadlineRepo, lateRepo, _ := setupCheckerService()
	ctx := context.Background()

	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	scoreValue := 100
	softDeadline := time.Now().Add(-48 * time.Hour)
	hardDeadline := softDeadline.Add(48 * time.Hour)
	submittedAt := softDeadline.Add(24 * time.Hour)

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreValue}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: softDeadline, HardDeadline: hardDeadline}

	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	deadlineRepo.On("GetByHomeworkID", ctx, hwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", ctx, courseID).Return(nil, assert.AnError)
	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == scoreValue // fallback to base score
	})).Return(nil)

	input := service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	}

	result, err := svc.SubmitGrade(ctx, courseID, input)
	assert.NoError(t, err)
	assert.Equal(t, scoreValue, result.Score)
}
