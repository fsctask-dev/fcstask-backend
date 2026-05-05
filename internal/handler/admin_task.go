package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

type AdminTaskHandler struct {
	taskRepo     repo.ITaskRepo
	homeworkRepo repo.IHomeworkRepo
}

func NewAdminTaskHandler(taskRepo repo.ITaskRepo, homeworkRepo repo.IHomeworkRepo) *AdminTaskHandler {
	return &AdminTaskHandler{
		taskRepo:     taskRepo,
		homeworkRepo: homeworkRepo,
	}
}

type CreateTaskRequest struct {
	RepoURL *string `json:"repo_url"`
	TaskURL *string `json:"task_url"`
}

type UpdateTaskRequest struct {
	RepoURL *string `json:"repo_url"`
	TaskURL *string `json:"task_url"`
}

// POST /admin/courses/:courseId/homework/:hwId/tasks
func (h *AdminTaskHandler) AdminCreateTaskHandler(c echo.Context) error {
	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}
	if _, err := h.homeworkRepo.GetByID(c.Request().Context(), hwID); err != nil {
		return notFound(c, "Homework not found")
	}

	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	task := &model.Task{
		HwID:    hwID,
		RepoURL: req.RepoURL,
		TaskURL: req.TaskURL,
	}

	if err := h.taskRepo.Create(c.Request().Context(), task); err != nil {
		return internalError(c, "Failed to create task")
	}
	return c.JSON(http.StatusCreated, task)
}

// GET /admin/courses/:courseId/homework/:hwId/tasks
func (h *AdminTaskHandler) AdminListTasksHandler(c echo.Context) error {
	hwID, err := uuid.Parse(c.Param("hwId"))
	if err != nil {
		return badRequest(c, "Invalid homework ID")
	}
	if _, err := h.homeworkRepo.GetByID(c.Request().Context(), hwID); err != nil {
		return notFound(c, "Homework not found")
	}

	tasks, err := h.taskRepo.GetByHwID(c.Request().Context(), hwID)
	if err != nil {
		return internalError(c, "Failed to fetch tasks")
	}
	return c.JSON(http.StatusOK, tasks)
}

// GET /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) AdminGetTaskHandler(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}

	task, err := h.taskRepo.GetByID(c.Request().Context(), taskID)
	if err != nil {
		return notFound(c, "Task not found")
	}

	return c.JSON(http.StatusOK, task)
}

// PATCH /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) AdminUpdateTaskHandler(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}

	task, err := h.taskRepo.GetByID(c.Request().Context(), taskID)
	if err != nil {
		return notFound(c, "Task not found")
	}

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	if req.RepoURL != nil {
		task.RepoURL = req.RepoURL
	}
	if req.TaskURL != nil {
		task.TaskURL = req.TaskURL
	}

	if err := h.taskRepo.Update(c.Request().Context(), task); err != nil {
		return internalError(c, "Failed to update task")
	}

	return c.JSON(http.StatusOK, task)
}

// DELETE /admin/courses/:courseId/homework/:hwId/tasks/:taskId
func (h *AdminTaskHandler) AdminDeleteTaskHandler(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}

	if _, err := h.taskRepo.GetByID(c.Request().Context(), taskID); err != nil {
		return notFound(c, "Task not found")
	}

	if err := h.taskRepo.Delete(c.Request().Context(), taskID); err != nil {
		return internalError(c, "Failed to delete task")
	}

	return c.NoContent(http.StatusNoContent)
}

// PATCH /admin/courses/:courseId/homework/:hwId/tasks/:taskId/score
func (h *AdminTaskHandler) AdminSetTaskScoreHandler(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		return badRequest(c, "Invalid task ID")
	}

	if _, err := h.taskRepo.GetByID(c.Request().Context(), taskID); err != nil {
		return notFound(c, "Task not found")
	}

	var req struct {
		Score int `json:"score"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "Invalid request body")
	}
	if req.Score < 0 {
		return badRequest(c, "Score must be non-negative")
	}

	if err := h.taskRepo.SetScore(c.Request().Context(), taskID, req.Score); err != nil {
		return internalError(c, "Failed to set score")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"task_id": taskID,
		"score":   req.Score,
	})
}
