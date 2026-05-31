package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type AdminTaskHandler struct {
	taskService IAdminTaskService
}

func NewAdminTaskHandler(taskService IAdminTaskService) *AdminTaskHandler {
	return &AdminTaskHandler{taskService: taskService}
}

type CreateTaskRequest struct {
	Title   *string `json:"title"`
	RepoURL *string `json:"repo_url"`
	TaskURL *string `json:"task_url"`
	Score   *int    `json:"score"`
}

type UpdateTaskRequest struct {
	Title   *string `json:"title"`
	RepoURL *string `json:"repo_url"`
	TaskURL *string `json:"task_url"`
	Score   *int    `json:"score"`
}

type SetTaskScoreRequest struct {
	Score int `json:"score"`
}

// POST /admin/courses/:courseId/homework/:hwId/tasks
func (h *AdminTaskHandler) CreateTask(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}

	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	input := service.CreateTaskInput{
		HwID: hwID,
	}
	if req.Title != nil {
		input.Title = req.Title
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
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}
	tasks, err := h.taskService.ListTasks(c.Request().Context(), user.ID, hwID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, tasks)
}

// GET /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) GetTask(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}
	task, err := h.taskService.GetTask(c.Request().Context(), user.ID, taskID)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, task)
}

// PATCH /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) UpdateTask(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	input := service.UpdateTaskInput{}
	if req.Title != nil {
		input.Title = req.Title
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

	task, err := h.taskService.UpdateTask(c.Request().Context(), user.ID, taskID, input)
	if err != nil {
		return serviceError(c, err)
	}

	return c.JSON(http.StatusOK, task)
}

// DELETE /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) DeleteTask(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}
	if err := h.taskService.DeleteTask(c.Request().Context(), user.ID, taskID); err != nil {
		return serviceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /admin/courses/:courseId/homework/:hwId/tasks/:taskId/score
func (h *AdminTaskHandler) SetScore(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}

	var req SetTaskScoreRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
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
