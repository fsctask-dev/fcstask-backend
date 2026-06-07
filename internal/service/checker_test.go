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

func intPtr(v int) *int         { return &v }
func f64Ptr(v float64) *float64 { return &v }

func setupChecker(dlErr bool) (
	*service.CheckerService,
	*MockTaskRepo,
	*MockHomeworkRepo,
	*MockScoreRepo,
	*MockDeadlineRepo,
	*MockCourseLateRepo,
) {
	taskRepo := new(MockTaskRepo)
	hwRepo := new(MockHomeworkRepo)
	scoreRepo := new(MockScoreRepo)
	dlRepo := new(MockDeadlineRepo)
	lateRepo := new(MockCourseLateRepo)
	roleRepo := newPermissiveRoleRepo()

	if dlErr {
		dlRepo.On("GetByHomeworkID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	}

	svc := service.NewCheckerService(taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo, roleRepo)
	return svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo
}

func TestSubmitGrade_MissingStudentID(t *testing.T) {
	svc, _, _, _, _, _ := setupChecker(false)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.Nil,
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "You don't have permission")
}

func TestSubmitGrade_MissingTaskID(t *testing.T) {
	svc, _, _, _, _, _ := setupChecker(false)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.Nil,
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "task_id is required")
}

func TestSubmitGrade_MissingCourseID(t *testing.T) {
	svc, _, _, _, _, _ := setupChecker(false)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.Nil,
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "course_id is required")
}

func TestSubmitGrade_InvalidStatus(t *testing.T) {
	svc, _, _, _, _, _ := setupChecker(false)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "unknown",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "status must be 'passed' or 'fail'")
}

func TestSubmitGrade_MissingSubmittedAt(t *testing.T) {
	svc, _, _, _, _, _ := setupChecker(false)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Status:    "passed",
	})
	assert.ErrorContains(t, err, "submitted_at is required")
}

func TestSubmitGrade_FutureSubmittedAt(t *testing.T) {
	svc, _, _, _, _, _ := setupChecker(false)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(24 * time.Hour),
	})
	assert.ErrorContains(t, err, "submitted_at cannot be in the future")
}

func TestSubmitGrade_TaskNotFound(t *testing.T) {
	svc, taskRepo, _, _, _, _ := setupChecker(false)
	taskRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      uuid.New(),
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "Task not found")
	taskRepo.AssertExpectations(t)
}

func TestSubmitGrade_HomeworkNotFound(t *testing.T) {
	svc, taskRepo, hwRepo, _, _, _ := setupChecker(false)
	taskID := uuid.New()
	hwID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    uuid.New(),
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "Homework not found")
}

func TestSubmitGrade_TaskCourseMismatch(t *testing.T) {
	svc, taskRepo, hwRepo, _, _, _ := setupChecker(false)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "task does not belong to the specified course")
}

func TestSubmitGrade_UpsertError(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _ := setupChecker(true)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.Anything).Return(assert.AnError)

	_, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "fail",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.ErrorContains(t, err, "Failed to save grade")
}

func TestSubmitGrade_Passed_BaseScore(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _ := setupChecker(true)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	scoreVal := 85

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.StudentID == studentID && s.Score == scoreVal && s.IsPassed == true
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, scoreVal, result.Score)
	assert.True(t, result.IsPassed)
}

func TestSubmitGrade_Fail_ZeroScore(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _ := setupChecker(true)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0 && s.IsPassed == false
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "fail",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
	assert.False(t, result.IsPassed)
}

func TestSubmitGrade_TaskScoreNil_ReturnsZero(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, _, _ := setupChecker(true)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: nil}, nil)
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 0
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
}

func setupLatePolicyMocks(
	taskRepo *MockTaskRepo,
	hwRepo *MockHomeworkRepo,
	dlRepo *MockDeadlineRepo,
	lateRepo *MockCourseLateRepo,
	scoreRepo *MockScoreRepo,
	task *model.Task,
	hw *model.Homework,
	deadline *model.Deadline,
	policy *model.CourseLatePolicy,
	expectedScore int,
	expectedPassed bool,
	studentID uuid.UUID,
) {
	taskRepo.On("GetByID", mock.Anything, task.TaskID).Return(task, nil).Once()
	taskRepo.On("GetByID", mock.Anything, task.TaskID).Return(task, nil).Once()
	hwRepo.On("GetByID", mock.Anything, hw.HwID).Return(hw, nil).Once()
	hwRepo.On("GetByID", mock.Anything, hw.HwID).Return(hw, nil).Once()
	dlRepo.On("GetByHomeworkID", mock.Anything, task.HwID).Return(deadline, nil)
	lateRepo.On("GetByCourseID", mock.Anything, hw.CourseID).Return(policy, nil)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.StudentID == studentID &&
			s.TaskID == task.TaskID &&
			s.CourseID == hw.CourseID &&
			s.Score == expectedScore &&
			s.IsPassed == expectedPassed
	})).Return(nil)
}

func TestSubmitGrade_OnTime_Nopenalty(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo := setupChecker(false)
	scoreVal := 100
	soft := time.Now().Add(-48 * time.Hour)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: soft, HardDeadline: soft.Add(72 * time.Hour)}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeLinear, SoftPenalty: 0.2, HardDeadlineScore: 0.5}

	setupLatePolicyMocks(taskRepo, hwRepo, dlRepo, lateRepo, scoreRepo,
		task, hw, deadline, policy, scoreVal, true, studentID)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: soft.Add(-time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, scoreVal, result.Score)
}

func TestSubmitGrade_LinearPenalty_MidWindow(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo := setupChecker(false)
	scoreVal := 100
	soft := time.Now().Add(-48 * time.Hour)
	hard := soft.Add(48 * time.Hour)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: soft, HardDeadline: hard}
	policy := &model.CourseLatePolicy{
		PolicyType: model.PolicyTypeLinear, SoftPenalty: 0.2, HardDeadlineScore: 0.5,
	}

	setupLatePolicyMocks(taskRepo, hwRepo, dlRepo, lateRepo, scoreRepo,
		task, hw, deadline, policy, 65, true, studentID)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: soft.Add(24 * time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, 65, result.Score)
}

func TestSubmitGrade_StepPenalty_ThreeDaysLate(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo := setupChecker(false)
	scoreVal := 100
	soft := time.Now().Add(-72 * time.Hour)
	hard := soft.Add(120 * time.Hour)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	stepPercent := 0.1

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: soft, HardDeadline: hard}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeStep, StepPercent: &stepPercent}

	setupLatePolicyMocks(taskRepo, hwRepo, dlRepo, lateRepo, scoreRepo,
		task, hw, deadline, policy, 70, true, studentID)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: soft.Add(48 * time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, 70, result.Score)
}

func TestSubmitGrade_AfterHardDeadline(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo := setupChecker(false)
	scoreVal := 100
	soft := time.Now().Add(-96 * time.Hour)
	hard := soft.Add(48 * time.Hour) // hard deadline is 48 hours after soft
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: soft, HardDeadline: hard}
	policy := &model.CourseLatePolicy{
		PolicyType:        model.PolicyTypeLinear,
		SoftPenalty:       0.2,
		HardDeadlineScore: 0.5,
	}

	setupLatePolicyMocks(taskRepo, hwRepo, dlRepo, lateRepo, scoreRepo,
		task, hw, deadline, policy, 0, true, studentID)

	// Submit after hard deadline
	submittedAt := hard.Add(1 * time.Hour)
	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
	assert.True(t, result.IsPassed)
}

func TestSubmitGrade_StepPenalty_ClampedToZero(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo := setupChecker(false)
	scoreVal := 100
	soft := time.Now().Add(-240 * time.Hour)
	hard := soft.Add(480 * time.Hour)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	stepPercent := 0.15

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: soft, HardDeadline: hard}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeStep, StepPercent: &stepPercent}
	setupLatePolicyMocks(taskRepo, hwRepo, dlRepo, lateRepo, scoreRepo,
		task, hw, deadline, policy, 0, true, studentID)
	submittedAt := soft.Add(168 * time.Hour)
	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: submittedAt,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score)
	assert.True(t, result.IsPassed)
}

func TestSubmitGrade_CoefficientPenalty(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo := setupChecker(false)
	scoreVal := 100
	soft := time.Now().Add(-48 * time.Hour)
	hard := soft.Add(48 * time.Hour)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	studentID := uuid.New()
	coeff := 0.75

	task := &model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}
	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	deadline := &model.Deadline{SoftDeadline: soft, HardDeadline: hard}
	policy := &model.CourseLatePolicy{PolicyType: model.PolicyTypeCoefficient, Coefficient: &coeff}

	setupLatePolicyMocks(taskRepo, hwRepo, dlRepo, lateRepo, scoreRepo,
		task, hw, deadline, policy, 75, true, studentID)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   studentID,
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: soft.Add(24 * time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, 75, result.Score)
}

func TestSubmitGrade_DeadlineRepoError_FallbackToBaseScore(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, _ := setupChecker(false)
	scoreVal := 100
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}, nil).Once()
	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}, nil).Once()
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil).Once()
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil).Once()
	dlRepo.On("GetByHomeworkID", mock.Anything, hwID).Return(nil, assert.AnError)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == scoreVal
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: time.Now().Add(-time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, scoreVal, result.Score)
}

func TestSubmitGrade_LatePolicyRepoError_FallbackToBaseScore(t *testing.T) {
	svc, taskRepo, hwRepo, scoreRepo, dlRepo, lateRepo := setupChecker(false)
	scoreVal := 100
	soft := time.Now().Add(-48 * time.Hour)
	hard := soft.Add(48 * time.Hour)
	taskID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}, nil).Once()
	taskRepo.On("GetByID", mock.Anything, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Score: &scoreVal}, nil).Once()
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil).Once()
	hwRepo.On("GetByID", mock.Anything, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil).Once()
	dlRepo.On("GetByHomeworkID", mock.Anything, hwID).Return(
		&model.Deadline{SoftDeadline: soft, HardDeadline: hard}, nil,
	)
	lateRepo.On("GetByCourseID", mock.Anything, courseID).Return(nil, assert.AnError)
	scoreRepo.On("Upsert", mock.Anything, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == scoreVal
	})).Return(nil)

	result, err := svc.SubmitGrade(context.Background(), service.SubmitGradeInput{
		StudentID:   uuid.New(),
		TaskID:      taskID,
		CourseID:    courseID,
		Status:      "passed",
		SubmittedAt: soft.Add(24 * time.Hour),
	})
	assert.NoError(t, err)
	assert.Equal(t, scoreVal, result.Score)
}
