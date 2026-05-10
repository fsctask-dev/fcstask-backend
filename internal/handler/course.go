package handler

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/google/uuid"
	"fcstask-backend/internal/db/repo"
	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type Course = model.Course
type BoardDeadline = model.BoardDeadline
type BoardTask = model.BoardTask
type BoardGroup = model.BoardGroup
type TaskBoardSummary = model.TaskBoardSummary

type PostCourseRequest struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Status       string `json:"status"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	RepoTemplate string `json:"repoTemplate"`
	Description  string `json:"description"`
}

type CourseHandler struct {
	courseService *service.CourseService
	roleRepo      repo.IRoleRepo
}

func NewCourseHandler(courseService *service.CourseService, roleRepo repo.IRoleRepo) *CourseHandler {
	return &CourseHandler{courseService: courseService, roleRepo: roleRepo}
}

func (h *CourseHandler) GetCourses(ctx echo.Context) error {
	user := ctx.Get(UserContextKey).(*model.User)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}
	courses, err := h.courseService.GetCourses(ctx.Request().Context(), user.ID, ctx.QueryParam("status"))
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, courses)
}

func (h *CourseHandler) GetCourse(ctx echo.Context) error {
	user := ctx.Get(UserContextKey).(*model.User)
	courseID := ctx.Param("courseId")
	course, err := h.courseService.GetCourse(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}
	courseUUID, _ := uuid.Parse(courseID)
	isParticipant, err := service.IsCourseParticipant(ctx.Request().Context(), h.roleRepo, user.ID, courseUUID)
	if err != nil {
		return serviceError(ctx, err)
	}
	if !isParticipant {
		return forbidden(ctx, "You are not a participant of this course")
	}

	return ctx.JSON(http.StatusOK, course)
}

func (h *CourseHandler) CreateCourse(ctx echo.Context) error {
	user := ctx.Get(UserContextKey).(*model.User)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}
	var req PostCourseRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "invalid JSON payload")
	}
	course, err := h.courseService.CreateCourse(ctx.Request().Context(), user.ID, courseInput(req))
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) UpdateCourse(ctx echo.Context) error {
	user := ctx.Get(UserContextKey).(*model.User)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}
	var req PostCourseRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "invalid JSON payload")
	}
	course, err := h.courseService.UpdateCourse(ctx.Request().Context(), user.ID, ctx.Param("courseId"), courseInput(req))
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