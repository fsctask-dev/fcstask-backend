package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type PostCourseRequest struct {
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	Status       string  `json:"status"`
	Type         string  `json:"type"`
	InviteCode   *string `json:"inviteCode,omitempty"`
	StartDate    string  `json:"startDate"`
	EndDate      string  `json:"endDate"`
	RepoTemplate string  `json:"repoTemplate"`
	Description  string  `json:"description"`
}

type JoinCourseRequest struct {
	Code string `json:"code"`
}

type CourseHandler struct {
	courseService ICourseService
}

func NewCourseHandler(courseService ICourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// GET /courses
func (h *CourseHandler) GetCourses(ctx echo.Context) error {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return unauthorized(ctx, "User not found in context")
	}
	courses, err := h.courseService.GetCourses(ctx.Request().Context(), user.ID, ctx.QueryParam("status"))
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, courses)
}

// GET /course/:courseId
func (h *CourseHandler) GetCourse(ctx echo.Context) error {
	courseID := ctx.Param("courseId")
	course, err := h.courseService.GetCourse(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}
	if course.Type == model.CourseTypePublic {
		return ctx.JSON(http.StatusOK, course)
	}

	user, ok := authenticatedUser(ctx)
	if !ok {
		return unauthorized(ctx, "User not found in context")
	}

	allowed, err := h.courseService.CanReadCourse(ctx.Request().Context(), user.ID, course)
	if err != nil {
		return serviceError(ctx, err)
	}
	if !allowed {
		return forbidden(ctx, "You are not a participant of this course")
	}

	return ctx.JSON(http.StatusOK, course)
}

// POST /admin/course/create
func (h *CourseHandler) CreateCourse(ctx echo.Context) error {
	user := mustAuthenticatedUser(ctx)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}

	var req PostCourseRequest
	if !bindRequest(ctx, &req, "invalid JSON payload") {
		return nil
	}
	course, err := h.courseService.CreateCourse(ctx.Request().Context(), user.ID, courseInput(req))
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, course)
}

// PATCH /admin/course/update
func (h *CourseHandler) UpdateCourse(ctx echo.Context) error {
	user := mustAuthenticatedUser(ctx)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}
	var req PostCourseRequest
	if !bindRequest(ctx, &req, "invalid JSON payload") {
		return nil
	}
	course, err := h.courseService.UpdateCourse(ctx.Request().Context(), user.ID, ctx.Param("courseId"), courseInput(req))
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, course)
}

// GET /course/board
func (h *CourseHandler) GetCourseBoard(ctx echo.Context) error {
	user := mustAuthenticatedUser(ctx)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}

	courseID := ctx.Param("courseId")

	course, err := h.courseService.GetCourse(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}

	allowed, err := h.courseService.CanReadCourse(ctx.Request().Context(), user.ID, course)
	if err != nil {
		return serviceError(ctx, err)
	}
	if !allowed {
		return forbidden(ctx, "You are not a participant of this course")
	}

	board, err := h.courseService.GetCourseBoard(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, board)
}

// POST /courses/:courseId/join
func (h *CourseHandler) JoinCourse(ctx echo.Context) error {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return unauthorized(ctx, "User not found in context")
	}

	var req JoinCourseRequest
	if !bindRequest(ctx, &req, "invalid JSON payload") {
		return nil
	}

	courseID := ctx.Param("courseId")
	if err := h.courseService.JoinCourse(ctx.Request().Context(), user.ID, courseID, req.Code); err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "successfully joined"})
}

// GET /api/courses/:courseId/scores
func (h *CourseHandler) GetScores(ctx echo.Context) error {
	user, ok := authenticatedUser(ctx)
	if !ok {
		return unauthorized(ctx, "User not found in context")
	}

	courseID := ctx.Param("courseId")
	entries, err := h.courseService.GetLeaderboard(ctx.Request().Context(), user.ID, courseID)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, entries)
}

func courseInput(req PostCourseRequest) service.CourseInput {
	return service.CourseInput{
		Name:         req.Name,
		Slug:         req.Slug,
		Status:       req.Status,
		Type:         model.CourseType(req.Type),
		InviteCode:   optionalString(req.InviteCode),
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		RepoTemplate: req.RepoTemplate,
		Description:  req.Description,
	}
}

func optionalString(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
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
