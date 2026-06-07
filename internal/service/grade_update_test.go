package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

func (m *MockTaskRepo) GetByHomeworkID(ctx context.Context, homeworkID uuid.UUID) ([]model.Task, error) {
	args := m.Called(ctx, homeworkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Task), args.Error(1)
}

func setupGradeUpdateService(hasPermission bool) (*service.GradeUpdateService, *MockTaskRepo, *MockHomeworkRepo, *MockScoreRepo, *MockRoleRepo) {
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepo)
	scoreRepo := new(MockScoreRepo)
	roleRepo := new(MockRoleRepo)
	roleID := uuid.New()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, mock.Anything).Return(true, nil)

	svc := service.NewGradeUpdateService(taskRepo, homeworkRepo, scoreRepo, roleRepo)
	return svc, taskRepo, homeworkRepo, scoreRepo, roleRepo
}

func setupStudentMembershipCheck(roleRepo *MockRoleRepo, studentID, courseID uuid.UUID, isMember bool) {
	roleRepo.On("HasScopedPermission", mock.Anything, studentID, courseID, service.PermissionCourseRead).
		Return(isMember, nil)
}

func TestUpdateGrade_InvalidStudentID(t *testing.T) {
	svc, _, _, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()

	input := service.UpdateGradeInput{
		StudentID: uuid.Nil,
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Score:     intPtr(85),
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "student_id is required")
}

func TestUpdateGrade_InvalidTaskID(t *testing.T) {
	svc, _, _, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()

	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.Nil,
		CourseID:  uuid.New(),
		Score:     intPtr(85),
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task_id is required")
}

func TestUpdateGrade_InvalidCourseID(t *testing.T) {
	svc, _, _, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()

	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.Nil,
		Score:     intPtr(85),
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestUpdateGrade_ScoreIsNil(t *testing.T) {
	svc, _, _, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()

	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Score:     nil,
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "score is required")
}

func TestUpdateGrade_ScoreNegative(t *testing.T) {
	svc, _, _, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()
	negativeScore := -10

	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    uuid.New(),
		CourseID:  uuid.New(),
		Score:     &negativeScore,
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "score must be 0 or higher")
}

func TestUpdateGrade_TaskNotFound(t *testing.T) {
	svc, taskRepo, _, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	taskRepo.On("GetByID", ctx, taskID).Return(nil, errors.New("not found"))

	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    taskID,
		CourseID:  uuid.New(),
		Score:     intPtr(85),
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task not found")
	taskRepo.AssertExpectations(t)
}

func TestUpdateGrade_HomeworkNotFound(t *testing.T) {
	svc, taskRepo, homeworkRepo, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()

	task := &model.Task{TaskID: taskID, HwID: uuid.New()}
	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)

	homeworkRepo.On("GetByID", ctx, task.HwID).Return(nil, errors.New("not found"))

	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    taskID,
		CourseID:  uuid.New(),
		Score:     intPtr(85),
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "homework not found")
	taskRepo.AssertExpectations(t)
	homeworkRepo.AssertExpectations(t)
}

func TestUpdateGrade_TaskWrongCourse(t *testing.T) {
	svc, taskRepo, homeworkRepo, _, _ := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	wrongCourseID := uuid.New()

	task := &model.Task{TaskID: taskID, HwID: uuid.New()}
	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)

	homework := &model.Homework{HwID: task.HwID, CourseID: wrongCourseID}
	homeworkRepo.On("GetByID", ctx, task.HwID).Return(homework, nil)

	input := service.UpdateGradeInput{
		StudentID: uuid.New(),
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     intPtr(85),
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task does not belong to the specified course")
	taskRepo.AssertExpectations(t)
	homeworkRepo.AssertExpectations(t)
}

func TestUpdateGrade_Success(t *testing.T) {
	svc, taskRepo, homeworkRepo, scoreRepo, roleRepo := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 95
	task := &model.Task{TaskID: taskID, HwID: uuid.New()}
	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)
	homework := &model.Homework{HwID: task.HwID, CourseID: courseID}
	homeworkRepo.On("GetByID", ctx, task.HwID).Return(homework, nil)
	scoreRepo.On("Upsert", ctx, mock.Anything).Return(nil)

	input := service.UpdateGradeInput{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     &scoreValue,
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, scoreValue, result.Score)

	taskRepo.AssertExpectations(t)
	homeworkRepo.AssertExpectations(t)
	scoreRepo.AssertExpectations(t)
	roleRepo.AssertExpectations(t)
}

func TestUpdateGrade_SuccessWithDifferentScore(t *testing.T) {
	svc, taskRepo, homeworkRepo, scoreRepo, roleRepo := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 100

	task := &model.Task{TaskID: taskID, HwID: uuid.New()}
	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)

	homework := &model.Homework{HwID: task.HwID, CourseID: courseID}
	homeworkRepo.On("GetByID", ctx, task.HwID).Return(homework, nil)

	setupStudentMembershipCheck(roleRepo, studentID, courseID, true)

	scoreRepo.On("Upsert", ctx, mock.MatchedBy(func(s *model.StudentTaskScore) bool {
		return s.Score == 100
	})).Return(nil)

	input := service.UpdateGradeInput{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     &scoreValue,
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100, result.Score)
}

func TestUpdateGrade_UpsertError(t *testing.T) {
	svc, taskRepo, homeworkRepo, scoreRepo, roleRepo := setupGradeUpdateService(true)
	ctx := context.Background()
	userID := uuid.New()
	studentID := uuid.New()
	taskID := uuid.New()
	courseID := uuid.New()
	scoreValue := 75

	task := &model.Task{TaskID: taskID, HwID: uuid.New()}
	taskRepo.On("GetByID", ctx, taskID).Return(task, nil)

	homework := &model.Homework{HwID: task.HwID, CourseID: courseID}
	homeworkRepo.On("GetByID", ctx, task.HwID).Return(homework, nil)

	setupStudentMembershipCheck(roleRepo, studentID, courseID, true)

	scoreRepo.On("Upsert", ctx, mock.AnythingOfType("*model.StudentTaskScore")).Return(errors.New("db error"))

	input := service.UpdateGradeInput{
		StudentID: studentID,
		TaskID:    taskID,
		CourseID:  courseID,
		Score:     &scoreValue,
		UserID:    userID,
	}

	result, err := svc.UpdateGrade(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to update grade")

	taskRepo.AssertExpectations(t)
	homeworkRepo.AssertExpectations(t)
	scoreRepo.AssertExpectations(t)
}
