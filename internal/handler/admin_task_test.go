package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/handler"
)

type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) Create(ctx context.Context, task *model.Task) error {
	return m.Called(ctx, task).Error(0)
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
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *MockTaskRepo) Update(ctx context.Context, task *model.Task) error {
	return m.Called(ctx, task).Error(0)
}

func (m *MockTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockTaskRepo) SetScore(ctx context.Context, id uuid.UUID, score int) error {
	return m.Called(ctx, id, score).Error(0)
}

type MockHomeworkRepoForTask struct {
	mock.Mock
}

func (m *MockHomeworkRepoForTask) Create(ctx context.Context, hw *model.Homework) error {
	return m.Called(ctx, hw).Error(0)
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
	return args.Get(0).([]model.Homework), args.Error(1)
}

func (m *MockHomeworkRepoForTask) Update(ctx context.Context, hw *model.Homework) error {
	return m.Called(ctx, hw).Error(0)
}

func (m *MockHomeworkRepoForTask) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func TestAdminCreateTask_Success(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	repoURL := "https://github.com/test/repo"
	taskURL := "https://test.com/task"

	body := map[string]interface{}{
		"repo_url": repoURL,
		"task_url": taskURL,
	}
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())

	homework := &model.Homework{HwID: hwID}
	homeworkRepo.On("GetByID", mock.Anything, hwID).Return(homework, nil)
	taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)

	err := h.AdminCreateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var result model.Task
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Equal(t, hwID, result.HwID)
	assert.Equal(t, &repoURL, result.RepoURL)
	assert.Equal(t, &taskURL, result.TaskURL)
}

func TestAdminCreateTask_InvalidHomeworkID(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/homework/invalid-uuid/tasks", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), "invalid-uuid")

	err := h.AdminCreateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	taskRepo.AssertNotCalled(t, "Create")
}

func TestAdminCreateTask_HomeworkNotFound(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())
	homeworkRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)
	err := h.AdminCreateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	taskRepo.AssertNotCalled(t, "Create")
}

func TestAdminCreateTask_InvalidRequestBody(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodPost, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks", "invalid body")
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())

	homework := &model.Homework{HwID: hwID}
	homeworkRepo.On("GetByID", mock.Anything, hwID).Return(homework, nil)

	err := h.AdminCreateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminListTasks_Success(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())

	homework := &model.Homework{HwID: hwID}
	expectedTasks := []model.Task{
		{TaskID: uuid.New(), HwID: hwID},
		{TaskID: uuid.New(), HwID: hwID},
	}
	homeworkRepo.On("GetByID", mock.Anything, hwID).Return(homework, nil)
	taskRepo.On("GetByHwID", mock.Anything, hwID).Return(expectedTasks, nil)

	err := h.AdminListTasksHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result []model.Task
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Len(t, result, 2)
}

func TestAdminListTasks_InvalidHomeworkID(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+courseID.String()+"/homework/invalid-uuid/tasks", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), "invalid-uuid")

	err := h.AdminListTasksHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	taskRepo.AssertNotCalled(t, "GetByHwID")
}

func TestAdminListTasks_HomeworkNotFound(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId")
	c.SetParamValues(courseID.String(), hwID.String())

	homeworkRepo.On("GetByID", mock.Anything, hwID).Return(nil, assert.AnError)

	err := h.AdminListTasksHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	taskRepo.AssertNotCalled(t, "GetByHwID")
}

func TestAdminGetTask_Success(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	expectedTask := &model.Task{TaskID: taskID, HwID: hwID}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(expectedTask, nil)

	err := h.AdminGetTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result model.Task
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Equal(t, taskID, result.TaskID)
}

func TestAdminGetTask_InvalidTaskID(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/invalid-uuid", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), "invalid-uuid")

	err := h.AdminGetTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminGetTask_NotFound(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	req, rec := newRequest(http.MethodGet, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)

	err := h.AdminGetTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminUpdateTask_Success(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	newRepoURL := "https://github.com/test/updated-repo"
	newTaskURL := "https://test.com/updated-task"

	body := map[string]interface{}{
		"repo_url": newRepoURL,
		"task_url": newTaskURL,
	}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String(), body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	oldRepoURL := "https://github.com/test/old-repo"
	oldTaskURL := "https://test.com/old-task"
	existingTask := &model.Task{
		TaskID:  taskID,
		HwID:    hwID,
		RepoURL: &oldRepoURL,
		TaskURL: &oldTaskURL,
	}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil)
	taskRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)

	err := h.AdminUpdateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result model.Task
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Equal(t, &newRepoURL, result.RepoURL)
	assert.Equal(t, &newTaskURL, result.TaskURL)
}

func TestAdminUpdateTask_InvalidTaskID(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/invalid-uuid", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), "invalid-uuid")

	err := h.AdminUpdateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminUpdateTask_TaskNotFound(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)

	err := h.AdminUpdateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminUpdateTask_UpdateRepoURLOnly(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	newRepoURL := "https://github.com/test/updated-repo"

	body := map[string]interface{}{
		"repo_url": newRepoURL,
	}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String(), body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	oldRepoURL := "https://github.com/test/old-repo"
	oldTaskURL := "https://test.com/old-task"
	existingTask := &model.Task{
		TaskID:  taskID,
		HwID:    hwID,
		RepoURL: &oldRepoURL,
		TaskURL: &oldTaskURL,
	}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil)
	taskRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)

	err := h.AdminUpdateTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result model.Task
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Equal(t, &newRepoURL, result.RepoURL)
	assert.Equal(t, &oldTaskURL, result.TaskURL)
}

func TestAdminDeleteTask_Success(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	existingTask := &model.Task{TaskID: taskID, HwID: hwID}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil)
	taskRepo.On("Delete", mock.Anything, taskID).Return(nil)

	err := h.AdminDeleteTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestAdminDeleteTask_InvalidTaskID(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/invalid-uuid", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), "invalid-uuid")

	err := h.AdminDeleteTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminDeleteTask_TaskNotFound(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	req, rec := newRequest(http.MethodDelete, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String(), nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)

	err := h.AdminDeleteTaskHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminSetTaskScore_Success(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	score := 95

	body := map[string]interface{}{
		"score": score,
	}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String()+"/score", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	existingTask := &model.Task{TaskID: taskID, HwID: hwID}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil)
	taskRepo.On("SetScore", mock.Anything, taskID, score).Return(nil)

	err := h.AdminSetTaskScoreHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Equal(t, taskID.String(), result["task_id"])
	assert.Equal(t, float64(score), result["score"])
}

func TestAdminSetTaskScore_InvalidTaskID(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/invalid-uuid/score", nil)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), "invalid-uuid")

	err := h.AdminSetTaskScoreHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminSetTaskScore_TaskNotFound(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	body := map[string]interface{}{"score": 100}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String()+"/score", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	taskRepo.On("GetByID", mock.Anything, taskID).Return(nil, assert.AnError)

	err := h.AdminSetTaskScoreHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminSetTaskScore_InvalidRequestBody(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String()+"/score", "invalid")
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	existingTask := &model.Task{TaskID: taskID, HwID: hwID}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil)

	err := h.AdminSetTaskScoreHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminSetTaskScore_NegativeScore(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	body := map[string]interface{}{"score": -10}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String()+"/score", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	existingTask := &model.Task{TaskID: taskID, HwID: hwID}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil)

	err := h.AdminSetTaskScoreHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAdminSetTaskScore_ZeroScore(t *testing.T) {
	e := setupEcho()
	taskRepo := new(MockTaskRepo)
	homeworkRepo := new(MockHomeworkRepoForTask)
	h := handler.NewAdminTaskHandler(taskRepo, homeworkRepo)

	courseID := uuid.New()
	hwID := uuid.New()
	taskID := uuid.New()
	score := 0

	body := map[string]interface{}{
		"score": score,
	}
	req, rec := newRequest(http.MethodPatch, "/admin/courses/"+courseID.String()+"/homework/"+hwID.String()+"/tasks/"+taskID.String()+"/score", body)
	c := e.NewContext(req, rec)
	c.SetParamNames("courseId", "hwId", "taskId")
	c.SetParamValues(courseID.String(), hwID.String(), taskID.String())

	existingTask := &model.Task{TaskID: taskID, HwID: hwID}
	taskRepo.On("GetByID", mock.Anything, taskID).Return(existingTask, nil)
	taskRepo.On("SetScore", mock.Anything, taskID, score).Return(nil)

	err := h.AdminSetTaskScoreHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	assert.Equal(t, taskID.String(), result["task_id"])
	assert.Equal(t, float64(score), result["score"])
}
