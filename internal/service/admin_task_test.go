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

type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) Create(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockTaskRepo) GetByHwID(ctx context.Context, hwID uuid.UUID) ([]model.Task, error) {
	args := m.Called(ctx, hwID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *MockTaskRepo) Update(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepo) SetScore(ctx context.Context, id uuid.UUID, score int) error {
	args := m.Called(ctx, id, score)
	return args.Error(0)
}

type MockHomeworkRepoForTask struct {
	mock.Mock
}

func (m *MockHomeworkRepoForTask) Create(ctx context.Context, hw *model.Homework) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *MockHomeworkRepoForTask) GetByID(ctx context.Context, id uuid.UUID) (*model.Homework, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockHomeworkRepoForTask) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Homework, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Homework), args.Error(1)
}

func (m *MockHomeworkRepoForTask) Update(ctx context.Context, hw *model.Homework) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *MockHomeworkRepoForTask) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTaskService() (*service.AdminTaskService, *MockTaskRepo, *MockHomeworkRepoForTask) {
	taskRepo := new(MockTaskRepo)
	hwRepo := new(MockHomeworkRepoForTask)
	roleRepo := new(MockRoleRepo)
	roleID := uuid.New()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, mock.Anything).Return(true, nil)
	svc := service.NewAdminTaskService(taskRepo, hwRepo, roleRepo)
	return svc, taskRepo, hwRepo
}

func TestCreateTask_Success(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()
	userID := uuid.New()

	hwID := uuid.New()
	title := "Test Task"
	repoURL := "https://github.com/test/repo"
	taskURL := "https://test.com/task"
	Score := 100

	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Create", ctx, mock.MatchedBy(func(task *model.Task) bool {
		return task.HwID == hwID &&
			task.Title == title &&
			*task.RepoURL == repoURL &&
			*task.TaskURL == taskURL &&
			*task.Score == Score
	})).Return(nil)

	result, err := svc.CreateTask(ctx, userID, service.CreateTaskInput{
		HwID:    hwID,
		Title:   &title,
		RepoURL: repoURL,
		TaskURL: taskURL,
		Score:   Score,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, hwID, result.HwID)
	assert.Equal(t, title, result.Title)
	assert.Equal(t, repoURL, *result.RepoURL)
	assert.Equal(t, taskURL, *result.TaskURL)
	assert.Equal(t, Score, *result.Score)
	hwRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
}

func TestCreateTask_EmptyHwID(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.CreateTask(ctx, uuid.New(), service.CreateTaskInput{
		HwID:    uuid.Nil,
		RepoURL: "https://github.com/test/repo",
		TaskURL: "https://test.com/task",
		Score:   100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "homework_id is required")
}

func TestCreateTask_HomeworkNotFound(t *testing.T) {
	svc, _, hwRepo := setupTaskService()
	ctx := context.Background()

	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(nil, assert.AnError)

	result, err := svc.CreateTask(ctx, uuid.New(), service.CreateTaskInput{
		HwID:    hwID,
		RepoURL: "https://github.com/test/repo",
		TaskURL: "https://test.com/task",
		Score:   100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Homework not found")
	hwRepo.AssertExpectations(t)
}

func TestCreateTask_RepoError(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Create", ctx, mock.AnythingOfType("*model.Task")).Return(assert.AnError)

	result, err := svc.CreateTask(ctx, uuid.New(), service.CreateTaskInput{
		HwID:    hwID,
		RepoURL: "https://github.com/test/repo",
		Score:   100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to create task")
	hwRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
}

func TestCreateTask_NilURLs(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Create", ctx, mock.MatchedBy(func(task *model.Task) bool {
		return task.HwID == hwID && task.RepoURL == nil && task.TaskURL == nil
	})).Return(nil)

	result, err := svc.CreateTask(ctx, uuid.New(), service.CreateTaskInput{
		HwID:  hwID,
		Score: 100,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.RepoURL)
	assert.Nil(t, result.TaskURL)
	hwRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
}

func TestGetTask_Success(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()
	userID := uuid.New()

	taskID := uuid.New()
	hwID := uuid.New()
	expected := &model.Task{TaskID: taskID, HwID: hwID, Title: "Test"}
	taskRepo.On("GetByID", ctx, taskID).Return(expected, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)

	result, err := svc.GetTask(ctx, userID, taskID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	taskRepo.AssertExpectations(t)
}

func TestGetTask_EmptyID(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.GetTask(ctx, uuid.New(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task_id is required")
}

func TestGetTask_NotFound(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(nil, assert.AnError)

	result, err := svc.GetTask(ctx, uuid.New(), taskID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Task not found")
	taskRepo.AssertExpectations(t)
}

func TestListTasks_Success(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	hwID := uuid.New()
	expectedTasks := []model.Task{
		{TaskID: uuid.New(), HwID: hwID, Title: "Task 1"},
		{TaskID: uuid.New(), HwID: hwID, Title: "Task 2"},
	}

	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("GetByHwID", ctx, hwID).Return(expectedTasks, nil)

	result, err := svc.ListTasks(ctx, uuid.New(), hwID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedTasks, result)
	hwRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
}

func TestListTasks_EmptyHwID(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.ListTasks(ctx, uuid.New(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "homework_id is required")
}

func TestListTasks_HomeworkNotFound(t *testing.T) {
	svc, _, hwRepo := setupTaskService()
	ctx := context.Background()

	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(nil, assert.AnError)

	result, err := svc.ListTasks(ctx, uuid.New(), hwID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Homework not found")
	hwRepo.AssertExpectations(t)
}

func TestListTasks_RepoError(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("GetByHwID", ctx, hwID).Return(nil, assert.AnError)

	result, err := svc.ListTasks(ctx, uuid.New(), hwID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to fetch tasks")
	hwRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
}

func TestUpdateTask_Success(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()
	userID := uuid.New()

	taskID := uuid.New()
	hwID := uuid.New()
	oldTitle := "Old Task"
	newTitle := "Updated Task"
	oldRepoURL := "https://github.com/test/old-repo"
	oldTaskURL := "https://test.com/old-task"
	newRepoURL := "https://github.com/test/updated-repo"
	newTaskURL := "https://test.com/updated-task"
	oldTaskScore := 100
	newTaskScore := 200

	existingTask := &model.Task{
		TaskID:  taskID,
		HwID:    hwID,
		Title:   oldTitle,
		RepoURL: &oldRepoURL,
		TaskURL: &oldTaskURL,
		Score:   &oldTaskScore,
	}

	taskRepo.On("GetByID", ctx, taskID).Return(existingTask, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Update", ctx, mock.MatchedBy(func(task *model.Task) bool {
		return task.Title == newTitle &&
			*task.RepoURL == newRepoURL &&
			*task.TaskURL == newTaskURL &&
			*task.Score == newTaskScore
	})).Return(nil)

	result, err := svc.UpdateTask(ctx, userID, taskID, service.UpdateTaskInput{
		Title:   &newTitle,
		RepoURL: newRepoURL,
		TaskURL: newTaskURL,
		Score:   newTaskScore,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newTitle, result.Title)
	assert.Equal(t, newRepoURL, *result.RepoURL)
	assert.Equal(t, newTaskURL, *result.TaskURL)
	assert.Equal(t, newTaskScore, *result.Score)
	taskRepo.AssertExpectations(t)
}

func TestUpdateTask_PartialUpdate(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()
	userID := uuid.New()

	taskID := uuid.New()
	hwID := uuid.New()
	oldRepoURL := "https://github.com/test/old-repo"
	oldTaskURL := "https://test.com/old-task"
	newRepoURL := "https://github.com/test/updated-repo"

	existingTask := &model.Task{
		TaskID:  taskID,
		HwID:    hwID,
		Title:   "Old",
		RepoURL: &oldRepoURL,
		TaskURL: &oldTaskURL,
	}

	taskRepo.On("GetByID", ctx, taskID).Return(existingTask, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Update", ctx, mock.MatchedBy(func(task *model.Task) bool {
		return *task.RepoURL == newRepoURL && *task.TaskURL == oldTaskURL
	})).Return(nil)

	result, err := svc.UpdateTask(ctx, userID, taskID, service.UpdateTaskInput{
		RepoURL: newRepoURL,
	})

	assert.NoError(t, err)
	assert.Equal(t, newRepoURL, *result.RepoURL)
	assert.Equal(t, oldTaskURL, *result.TaskURL)
	taskRepo.AssertExpectations(t)
}

func TestUpdateTask_NotFound(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(nil, assert.AnError)

	result, err := svc.UpdateTask(ctx, uuid.New(), taskID, service.UpdateTaskInput{
		RepoURL: "https://github.com/test/repo",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Task not found")
	taskRepo.AssertExpectations(t)
}

func TestUpdateTask_UpdateError(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()
	userID := uuid.New()

	taskID := uuid.New()
	hwID := uuid.New()
	existingTask := &model.Task{TaskID: taskID, HwID: hwID, Title: "Test"}

	taskRepo.On("GetByID", ctx, taskID).Return(existingTask, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Update", ctx, mock.AnythingOfType("*model.Task")).Return(assert.AnError)

	result, err := svc.UpdateTask(ctx, userID, taskID, service.UpdateTaskInput{
		RepoURL: "https://github.com/test/repo",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to update task")
	taskRepo.AssertExpectations(t)
}

func TestDeleteTask_Success(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	hwID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Title: "Test"}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Delete", ctx, taskID).Return(nil)

	err := svc.DeleteTask(ctx, uuid.New(), taskID)

	assert.NoError(t, err)
	taskRepo.AssertExpectations(t)
}

func TestDeleteTask_EmptyID(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	err := svc.DeleteTask(ctx, uuid.New(), uuid.Nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task_id is required")
}

func TestDeleteTask_NotFound(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(nil, assert.AnError)

	err := svc.DeleteTask(ctx, uuid.New(), taskID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Task not found")
	taskRepo.AssertExpectations(t)
}

func TestDeleteTask_DeleteError(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	hwID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Title: "Test"}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("Delete", ctx, taskID).Return(assert.AnError)

	err := svc.DeleteTask(ctx, uuid.New(), taskID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to delete task")
	taskRepo.AssertExpectations(t)
}

func TestSetScore_Success(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	hwID := uuid.New()
	score := 250

	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Title: "Test"}, nil).Twice()
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("SetScore", ctx, taskID, score).Return(nil)

	result, err := svc.SetScore(ctx, uuid.New(), service.SetTaskScoreInput{
		TaskID: taskID,
		Score:  score,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, taskID, result.TaskID)
	taskRepo.AssertExpectations(t)
}

func TestSetScore_EmptyTaskID(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.SetScore(ctx, uuid.New(), service.SetTaskScoreInput{
		TaskID: uuid.Nil,
		Score:  100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task_id is required")
}

func TestSetScore_NegativeScore(t *testing.T) {
	svc, _, _ := setupTaskService()
	ctx := context.Background()

	result, err := svc.SetScore(ctx, uuid.New(), service.SetTaskScoreInput{
		TaskID: uuid.New(),
		Score:  -10,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "score must be positive")
}

func TestSetScore_TaskNotFound(t *testing.T) {
	svc, taskRepo, _ := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(nil, assert.AnError)

	result, err := svc.SetScore(ctx, uuid.New(), service.SetTaskScoreInput{
		TaskID: taskID,
		Score:  100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Task not found")
	taskRepo.AssertExpectations(t)
}

func TestSetScore_SetScoreError(t *testing.T) {
	svc, taskRepo, hwRepo := setupTaskService()
	ctx := context.Background()

	taskID := uuid.New()
	hwID := uuid.New()
	taskRepo.On("GetByID", ctx, taskID).Return(&model.Task{TaskID: taskID, HwID: hwID, Title: "Test"}, nil)
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	taskRepo.On("SetScore", ctx, taskID, 100).Return(assert.AnError)

	result, err := svc.SetScore(ctx, uuid.New(), service.SetTaskScoreInput{
		TaskID: taskID,
		Score:  100,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to set score")
	taskRepo.AssertExpectations(t)
}
