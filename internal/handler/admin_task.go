package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/service"
)

type AdminTaskHandler struct {
	taskService IAdminTaskService
}

func NewAdminTaskHandler(taskService IAdminTaskService) *AdminTaskHandler {
	return &AdminTaskHandler{taskService: taskService}
}

type CreateTaskRequest struct {
	RepoURL *string `json:"repo_url"`
	TaskURL *string `json:"task_url"`
	Score   *int    `json:"score"`
}

type UpdateTaskRequest struct {
	RepoURL *string `json:"repo_url"`
	TaskURL *string `json:"task_url"`
	Score   *int    `json:"score"`
}

type SetTaskScoreRequest struct {
	Score int `json:"score"`
}

// POST /admin/courses/:courseId/homework/:hwId/tasks
func (h *AdminTaskHandler) CreateTask(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	hwID, ok := parseUUIDParam(c, "hwId", "Invalid homework ID")
	if !ok {
		return nil
	}

	var req CreateTaskRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	input := service.CreateTaskInput{
		HwID: hwID,
	}
	if req.RepoURL != nil {
		input.RepoURL = *req.RepoURL
	}
	if req.TaskURL != nil {
		input.TaskURL = *req.TaskURL
	}
	if req.Score != nil {
		input.Score = *req.Score
	}

	task, err := h.taskService.CreateTask(c.Request().Context(), user.ID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusCreated, task)
}

// GET /admin/courses/:courseId/homework/:hwId/tasks
func (h *AdminTaskHandler) ListTasks(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	hwID, ok := parseUUIDParam(c, "hwId", "Invalid homework ID")
	if !ok {
		return nil
	}
	tasks, err := h.taskService.ListTasks(c.Request().Context(), user.ID, hwID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, tasks)
}

// GET /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) GetTask(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	taskID, ok := parseUUIDParam(c, "taskId", "Invalid task ID")
	if !ok {
		return nil
	}
	task, err := h.taskService.GetTask(c.Request().Context(), user.ID, taskID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, task)
}

// PATCH /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) UpdateTask(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	taskID, ok := parseUUIDParam(c, "taskId", "Invalid task ID")
	if !ok {
		return nil
	}

	var req UpdateTaskRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	input := service.UpdateTaskInput{}
	if req.RepoURL != nil {
		input.RepoURL = *req.RepoURL
	}
	if req.TaskURL != nil {
		input.TaskURL = *req.TaskURL
	}
	if req.Score != nil {
		input.Score = *req.Score
	}

	task, err := h.taskService.UpdateTask(c.Request().Context(), user.ID, taskID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, task)
}

// DELETE /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) DeleteTask(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	taskID, ok := parseUUIDParam(c, "taskId", "Invalid task ID")
	if !ok {
		return nil
	}
	if err := h.taskService.DeleteTask(c.Request().Context(), user.ID, taskID); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /admin/courses/:courseId/homework/:hwId/tasks/:taskId/score
func (h *AdminTaskHandler) SetScore(c echo.Context) error {
	user, ok := authenticatedUser(c)
	if !ok {
		return unauthorized(c, "User not found")
	}

	taskID, ok := parseUUIDParam(c, "taskId", "Invalid task ID")
	if !ok {
		return nil
	}

	var req SetTaskScoreRequest
	if !bindRequest(c, &req, "Invalid request body") {
		return nil
	}

	task, err := h.taskService.SetScore(c.Request().Context(), user.ID, service.SetTaskScoreInput{
		TaskID: taskID,
		Score:  req.Score,
	})
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, task)
}
