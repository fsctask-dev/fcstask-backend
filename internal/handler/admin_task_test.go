package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/handler"
	"fcstask-backend/internal/service"
)

type MockAdminTaskService struct {
	mock.Mock
}

func (m *MockAdminTaskService) CreateTask(ctx context.Context, userID uuid.UUID, input service.CreateTaskInput) (*model.Task, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockAdminTaskService) GetTask(ctx context.Context, userID, taskID uuid.UUID) (*model.Task, error) {
	args := m.Called(ctx, userID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockAdminTaskService) ListTasks(ctx context.Context, userID, hwID uuid.UUID) ([]model.Task, error) {
	args := m.Called(ctx, userID, hwID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *MockAdminTaskService) UpdateTask(ctx context.Context, userID, taskID uuid.UUID, input service.UpdateTaskInput) (*model.Task, error) {
	args := m.Called(ctx, userID, taskID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockAdminTaskService) PublishTask(ctx context.Context, userID uuid.UUID, input service.PublishTaskInput) (*model.Task, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockAdminTaskService) DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error {
	args := m.Called(ctx, userID, taskID)
	return args.Error(0)
}

func (m *MockAdminTaskService) SetScore(ctx context.Context, userID uuid.UUID, input service.SetTaskScoreInput) (*model.Task, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func TestHandlerCreateTask_Success(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	hwID := uuid.New()
	repoURL := "https://github.com/test/repo"
	taskURL := "https://test.com/task"
	Score := 100

	body := map[string]interface{}{
		"repo_url": repoURL,
		"task_url": taskURL,
		"score":    Score,
	}
	expected := &model.Task{TaskID: uuid.New(), HwID: hwID, RepoURL: &repoURL, TaskURL: &taskURL}

	c, rec := newEchoContextMultiParam(http.MethodPost, "/", body,
		[]string{"courseId", "hwId"},
		[]string{uuid.New().String(), hwID.String()},
	)
	svc.On("CreateTask", mock.Anything, mock.Anything, service.CreateTaskInput{
		HwID:    hwID,
		RepoURL: repoURL,
		TaskURL: taskURL,
		Score:   Score,
	}).Return(expected, nil)

	err := h.CreateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerCreateTask_InvalidHwID(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	c, rec := newEchoContext(http.MethodPost, "/", nil, map[string]string{"hwId": "bad"})

	err := h.CreateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerCreateTask_HomeworkNotFound(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	hwID := uuid.New()
	courseID := uuid.New()
	body := map[string]interface{}{"repo_url": "https://github.com/test/repo"}

	c, rec := newEchoContextMultiParam(http.MethodPost, "/", body,
		[]string{"courseId", "hwId"},
		[]string{courseID.String(), hwID.String()},
	)
	svc.On("CreateTask", mock.Anything, mock.Anything, mock.Anything).Return(nil, service.NotFound("Homework not found"))

	err := h.CreateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerCreateTask_NilURLs(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	hwID := uuid.New()
	courseID := uuid.New()
	expected := &model.Task{TaskID: uuid.New(), HwID: hwID}

	c, rec := newEchoContextMultiParam(http.MethodPost, "/", map[string]interface{}{},
		[]string{"courseId", "hwId"},
		[]string{courseID.String(), hwID.String()},
	)
	svc.On("CreateTask", mock.Anything, mock.Anything, service.CreateTaskInput{HwID: hwID}).Return(expected, nil)

	err := h.CreateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerGetTask_Success(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	expected := &model.Task{TaskID: taskID}

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"taskId": taskID.String()})
	svc.On("GetTask", mock.Anything, mock.Anything, taskID).Return(expected, nil)

	err := h.GetTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerGetTask_InvalidID(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"taskId": "bad"})

	err := h.GetTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerGetTask_NotFound(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"taskId": taskID.String()})
	svc.On("GetTask", mock.Anything, mock.Anything, taskID).Return(nil, service.NotFound("Task not found"))

	err := h.GetTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListTasks_Success(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	hwID := uuid.New()
	courseID := uuid.New()
	expected := []model.Task{
		{TaskID: uuid.New(), HwID: hwID},
		{TaskID: uuid.New(), HwID: hwID},
	}

	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "hwId"},
		[]string{courseID.String(), hwID.String()},
	)
	svc.On("ListTasks", mock.Anything, mock.Anything, hwID).Return(expected, nil)

	err := h.ListTasks(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListTasks_InvalidHwID(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	c, rec := newEchoContext(http.MethodGet, "/", nil, map[string]string{"hwId": "bad"})

	err := h.ListTasks(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerListTasks_HomeworkNotFound(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	hwID := uuid.New()
	courseID := uuid.New()
	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "hwId"},
		[]string{courseID.String(), hwID.String()},
	)
	svc.On("ListTasks", mock.Anything, mock.Anything, hwID).Return(nil, service.NotFound("Homework not found"))

	err := h.ListTasks(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerListTasks_ServiceError(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	hwID := uuid.New()
	courseID := uuid.New()
	c, rec := newEchoContextMultiParam(http.MethodGet, "/", nil,
		[]string{"courseId", "hwId"},
		[]string{courseID.String(), hwID.String()},
	)
	svc.On("ListTasks", mock.Anything, mock.Anything, hwID).Return(nil, service.Internal("Failed to fetch tasks", nil))

	err := h.ListTasks(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerUpdateTask_Success(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	newRepo := "https://github.com/test/new-repo"
	newTask := "https://test.com/new-task"
	newScore := 200
	body := map[string]interface{}{
		"repo_url": newRepo,
		"task_url": newTask,
		"score":    newScore,
	}
	expected := &model.Task{TaskID: taskID, RepoURL: &newRepo, TaskURL: &newTask}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("UpdateTask", mock.Anything, mock.Anything, taskID, service.UpdateTaskInput{
		RepoURL: newRepo,
		TaskURL: newTask,
		Score:   newScore,
	}).Return(expected, nil)

	err := h.UpdateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerUpdateTask_InvalidID(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"taskId": "bad"})

	err := h.UpdateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerUpdateTask_NotFound(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	body := map[string]interface{}{"repo_url": "https://github.com/test/repo"}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("UpdateTask", mock.Anything, mock.Anything, taskID, mock.Anything).Return(nil, service.NotFound("Task not found"))

	err := h.UpdateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerUpdateTask_PartialUpdate(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	newRepo := "https://github.com/test/new-repo"
	body := map[string]interface{}{"repo_url": newRepo}
	expected := &model.Task{TaskID: taskID, RepoURL: &newRepo}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("UpdateTask", mock.Anything, mock.Anything, taskID, service.UpdateTaskInput{RepoURL: newRepo}).Return(expected, nil)

	err := h.UpdateTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerPublishTask_Success(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	body := map[string]interface{}{"is_public": true}
	expected := &model.Task{TaskID: taskID, IsPublic: boolPtr(true)}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("PublishTask", mock.Anything, mock.Anything, service.PublishTaskInput{
		TaskID:   taskID,
		IsPublic: true,
	}).Return(expected, nil)

	err := h.PublishTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerPublishTask_Unpublish(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	body := map[string]interface{}{"is_public": false}
	expected := &model.Task{TaskID: taskID, IsPublic: boolPtr(false)}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("PublishTask", mock.Anything, mock.Anything, service.PublishTaskInput{
		TaskID:   taskID,
		IsPublic: false,
	}).Return(expected, nil)

	err := h.PublishTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerPublishTask_InvalidID(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", map[string]interface{}{"is_public": true}, map[string]string{"taskId": "bad"})

	err := h.PublishTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerPublishTask_NotFound(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	body := map[string]interface{}{"is_public": true}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("PublishTask", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, service.NotFound("Task not found"))

	err := h.PublishTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerDeleteTask_Success(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"taskId": taskID.String()})
	svc.On("DeleteTask", mock.Anything, mock.Anything, taskID).Return(nil)

	err := h.DeleteTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerDeleteTask_InvalidID(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"taskId": "bad"})

	err := h.DeleteTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerDeleteTask_NotFound(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"taskId": taskID.String()})
	svc.On("DeleteTask", mock.Anything, mock.Anything, taskID).Return(service.NotFound("Task not found"))

	err := h.DeleteTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerDeleteTask_DeleteError(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	c, rec := newEchoContext(http.MethodDelete, "/", nil, map[string]string{"taskId": taskID.String()})
	svc.On("DeleteTask", mock.Anything, mock.Anything, taskID).Return(service.Internal("Failed to delete task", nil))

	err := h.DeleteTask(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerSetScore_Success(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	score := 250
	body := map[string]interface{}{"score": score}
	expected := &model.Task{TaskID: taskID}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("SetScore", mock.Anything, mock.Anything, service.SetTaskScoreInput{
		TaskID: taskID,
		Score:  score,
	}).Return(expected, nil)

	err := h.SetScore(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerSetScore_InvalidTaskID(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	c, rec := newEchoContext(http.MethodPatch, "/", nil, map[string]string{"taskId": "bad"})

	err := h.SetScore(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerSetScore_NegativeScore(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	body := map[string]interface{}{"score": -10}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("SetScore", mock.Anything, mock.Anything, service.SetTaskScoreInput{TaskID: taskID, Score: -10}).
		Return(nil, service.BadRequest("score must be positive"))

	err := h.SetScore(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerSetScore_TaskNotFound(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	body := map[string]interface{}{"score": 100}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("SetScore", mock.Anything, mock.Anything, service.SetTaskScoreInput{TaskID: taskID, Score: 100}).
		Return(nil, service.NotFound("Task not found"))

	err := h.SetScore(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandlerSetScore_ServiceError(t *testing.T) {
	svc := new(MockAdminTaskService)
	h := handler.NewAdminTaskHandler(svc)

	taskID := uuid.New()
	body := map[string]interface{}{"score": 100}

	c, rec := newEchoContext(http.MethodPatch, "/", body, map[string]string{"taskId": taskID.String()})
	svc.On("SetScore", mock.Anything, mock.Anything, service.SetTaskScoreInput{TaskID: taskID, Score: 100}).
		Return(nil, service.Internal("Failed to set score", nil))

	err := h.SetScore(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}

func boolPtr(b bool) *bool { return &b }
