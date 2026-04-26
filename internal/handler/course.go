package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type Course = models.Course
type BoardDeadline = models.BoardDeadline
type BoardTask = models.BoardTask
type BoardGroup = models.BoardGroup
type TaskBoardSummary = models.TaskBoardSummary

type PostCourseRequest struct {
	Name         string            `json:"name"`
	Slug         string            `json:"slug"`
	Status       string            `json:"status"`
	Type         models.CourseType `json:"type"`
	StartDate    string            `json:"startDate"`
	EndDate      string            `json:"endDate"`
	RepoTemplate string            `json:"repoTemplate"`
	Description  string            `json:"description"`
}

type CourseHandler struct {
	courseService *service.CourseService
}

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

func (h *CourseHandler) GetCourses(ctx echo.Context) error {
	courses, err := h.courseService.GetCourses(ctx.Request().Context(), ctx.QueryParam("status"))
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, courses)
}

func (h *CourseHandler) GetCourse(ctx echo.Context) error {
	course, err := h.courseService.GetCourse(ctx.Request().Context(), ctx.Param("courseId"))
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, course)
}

func (h *CourseHandler) CreateCourse(ctx echo.Context) error {
	var req PostCourseRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "invalid JSON payload")
	}

	course, err := h.courseService.CreateCourse(ctx.Request().Context(), courseInput(req))
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) UpdateCourse(ctx echo.Context) error {
	var req PostCourseRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "invalid JSON payload")
	}

	course, err := h.courseService.UpdateCourse(ctx.Request().Context(), ctx.Param("courseId"), courseInput(req))
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, course)
}

func (h *CourseHandler) GetCourseBoard(ctx echo.Context) error {
	board, err := h.courseService.GetCourseBoard(ctx.Request().Context(), ctx.Param("courseId"))
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, board)
}

func courseInput(req PostCourseRequest) service.CourseInput {
	return service.CourseInput{
		Name:         req.Name,
		Slug:         req.Slug,
		Status:       req.Status,
		Type:         req.Type,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		RepoTemplate: req.RepoTemplate,
		Description:  req.Description,
	}
}

func isValidCourseStatus(status string) bool {
	return service.IsValidCourseStatus(status)
}

func isValidDate(date string) bool {
	return service.IsValidDate(date)
}

func isValidDateRange(start, end string) bool {
	return service.IsValidDateRange(start, end)
}
