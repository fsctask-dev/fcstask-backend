package handler

import (
	"fcstask-backend/internal/db/model"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/service"
)

type PostCourseRequest struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	InviteCode   string `json:"inviteCode,omitempty"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	RepoTemplate string `json:"repoTemplate"`
	Description  string `json:"description"`
}

type JoinCourseRequest struct {
	Code string `json:"code"`
}

type CourseHandler struct {
	courseService *service.CourseService
}

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// GET /courses
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

// GET /course/:courseId
func (h *CourseHandler) GetCourse(ctx echo.Context) error {
	user := ctx.Get(UserContextKey).(*model.User)
	courseID := ctx.Param("courseId")
	course, err := h.courseService.GetCourse(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}
	if course.Type == model.CourseTypePublic {
		return ctx.JSON(http.StatusOK, course)
	}

	courseUUID, _ := uuid.Parse(courseID)
	allowed, err := service.HasScopedPermission(ctx.Request().Context(), h.courseService.RoleRepo, user.ID, courseUUID, service.PermissionHomeworkRead)
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

// PATCH /admin/course/update
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

// GET /course/board
func (h *CourseHandler) GetCourseBoard(ctx echo.Context) error {
	user := ctx.Get(UserContextKey).(*model.User)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}

	courseID := ctx.Param("courseId")

	course, err := h.courseService.GetCourse(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}

	if course.Type == model.CourseTypePrivate {
		allowed, err := service.HasScopedPermission(ctx.Request().Context(), h.courseService.RoleRepo, user.ID, course.ID, service.PermissionHomeworkRead)
		if err != nil {
			return serviceError(ctx, err)
		}
		if !allowed {
			return forbidden(ctx, "You are not a participant of this course")
		}
	}

	board, err := h.courseService.GetCourseBoard(ctx.Request().Context(), courseID)
	if err != nil {
		return serviceError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, board)
}

// POST /courses/:courseId/join
func (h *CourseHandler) JoinCourse(ctx echo.Context) error {
	user := ctx.Get(UserContextKey).(*model.User)
	if user == nil {
		return unauthorized(ctx, "User not found in context")
	}

	var req JoinCourseRequest
	if err := ctx.Bind(&req); err != nil {
		return badRequest(ctx, "invalid JSON payload")
	}

	courseID := ctx.Param("courseId")
	if err := h.courseService.JoinCourse(ctx.Request().Context(), user.ID, courseID, req.Code); err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "successfully joined"})
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
