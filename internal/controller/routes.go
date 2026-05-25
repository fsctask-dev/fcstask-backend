package controller

import (
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"

	"fcstask-backend/internal/api"
	"fcstask-backend/internal/handler"
)

type routeMethod string

const (
	routeMethodGet    routeMethod = http.MethodGet
	routeMethodPost   routeMethod = http.MethodPost
	routeMethodPut    routeMethod = http.MethodPut
	routeMethodPatch  routeMethod = http.MethodPatch
	routeMethodDelete routeMethod = http.MethodDelete
)

type routeOperation string

const (
	routeSignIn               routeOperation = "sign_in"
	routeSignOut              routeOperation = "sign_out"
	routeSignUp               routeOperation = "sign_up"
	routeGetMe                routeOperation = "get_me"
	routePostV1Echo           routeOperation = "post_v1_echo"
	routeGetSessions          routeOperation = "get_sessions"
	routeCreateUser           routeOperation = "create_user"
	routeGetUserByEmail       routeOperation = "get_user_by_email"
	routeGetUsersWithSessions routeOperation = "get_users_with_sessions"
	routeGetUserByUsername    routeOperation = "get_user_by_username"
	routeGetUserByID          routeOperation = "get_user_by_id"
	routeGetCourses           routeOperation = "get_courses"
	routeCreateCourse         routeOperation = "create_course"
	routeGetCourse            routeOperation = "get_course"
	routeUpdateCourse         routeOperation = "update_course"
	routeGetCourseBoard       routeOperation = "get_course_board"
	routeGetCourseScores      routeOperation = "get_course_scores"
	routeJoinCourse           routeOperation = "join_course"
	routeCreateHomework       routeOperation = "create_homework"
	routeGetHomework          routeOperation = "get_homework"
	routeListHomework         routeOperation = "list_homework"
	routeUpdateHomework       routeOperation = "update_homework"
	routeDeleteHomework       routeOperation = "delete_homework"
	routePublishHomework      routeOperation = "publish_homework"
	routeSetDeadline          routeOperation = "set_deadline"
	routeUpdateDeadline       routeOperation = "update_deadline"
	routeDeleteDeadline       routeOperation = "delete_deadline"
	routeCreateTask           routeOperation = "create_task"
	routeListTasks            routeOperation = "list_tasks"
	routeGetTask              routeOperation = "get_task"
	routeUpdateTask           routeOperation = "update_task"
	routeDeleteTask           routeOperation = "delete_task"
	routeSetScore             routeOperation = "set_score"
	routeAssignCourseAdmin    routeOperation = "assign_course_admin"
	routeRevokeCourseAdmin    routeOperation = "revoke_course_admin"
	routeRemoveParticipant    routeOperation = "remove_course_participant"
	routeListUserRoles        routeOperation = "list_user_roles"
	routeAddPermission        routeOperation = "add_permission"
	routeRemovePermission     routeOperation = "remove_permission"
	routeListPermissions      routeOperation = "list_permissions"
	routeCreateSuperAdmin     routeOperation = "create_super_admin"
)

type routeSpec struct {
	method    routeMethod
	path      string
	protected bool
	operation routeOperation
}

var routeCatalog = []routeSpec{
	{method: routeMethodPost, path: "/api/signin", operation: routeSignIn},
	{method: routeMethodPost, path: "/api/signout", protected: true, operation: routeSignOut},
	{method: routeMethodPost, path: "/api/signup", operation: routeSignUp},
	{method: routeMethodGet, path: "/v1/api/me", protected: true, operation: routeGetMe},
	{method: routeMethodPost, path: "/v1/echo", operation: routePostV1Echo},
	{method: routeMethodGet, path: "/v1/sessions", protected: true, operation: routeGetSessions},
	{method: routeMethodPost, path: "/v1/users", operation: routeCreateUser},
	{method: routeMethodGet, path: "/v1/users/email/:email", operation: routeGetUserByEmail},
	{method: routeMethodGet, path: "/v1/users/sessions", protected: true, operation: routeGetUsersWithSessions},
	{method: routeMethodGet, path: "/v1/users/username/:username", operation: routeGetUserByUsername},
	{method: routeMethodGet, path: "/v1/users/:id", operation: routeGetUserByID},
	{method: routeMethodGet, path: "/api/courses", operation: routeGetCourses},
	{method: routeMethodPost, path: "/api/courses", operation: routeCreateCourse},
	{method: routeMethodGet, path: "/api/courses/:courseId", operation: routeGetCourse},
	{method: routeMethodPut, path: "/api/courses/:courseId", operation: routeUpdateCourse},
	{method: routeMethodGet, path: "/api/courses/:courseId/board", operation: routeGetCourseBoard},
	{method: routeMethodGet, path: "/api/courses/:courseId/scores", protected: true, operation: routeGetCourseScores},
	{method: routeMethodPost, path: "/api/courses/:courseId/join", protected: true, operation: routeJoinCourse},
	{method: routeMethodPost, path: "/admin/courses/:courseId/homework", protected: true, operation: routeCreateHomework},
	{method: routeMethodGet, path: "/admin/courses/:courseId/homework/:hwId", protected: true, operation: routeGetHomework},
	{method: routeMethodGet, path: "/admin/courses/:courseId/homework", protected: true, operation: routeListHomework},
	{method: routeMethodPatch, path: "/admin/courses/:courseId/homework/:hwId", protected: true, operation: routeUpdateHomework},
	{method: routeMethodDelete, path: "/admin/courses/:courseId/homework/:hwId", protected: true, operation: routeDeleteHomework},
	{method: routeMethodPatch, path: "/admin/courses/:courseId/homework/:hwId/publish", protected: true, operation: routePublishHomework},
	{method: routeMethodPut, path: "/admin/courses/:courseId/homework/:hwId/deadline", protected: true, operation: routeSetDeadline},
	{method: routeMethodPatch, path: "/admin/deadlines/:deadlineId", protected: true, operation: routeUpdateDeadline},
	{method: routeMethodDelete, path: "/admin/deadlines/:deadlineId", protected: true, operation: routeDeleteDeadline},
	{method: routeMethodPost, path: "/admin/courses/:courseId/homework/:hwId/tasks", protected: true, operation: routeCreateTask},
	{method: routeMethodGet, path: "/admin/courses/:courseId/homework/:hwId/tasks", protected: true, operation: routeListTasks},
	{method: routeMethodGet, path: "/admin/courses/:courseId/homework/:hwId/tasks/:taskId", protected: true, operation: routeGetTask},
	{method: routeMethodPatch, path: "/admin/courses/:courseId/homework/:hwId/tasks/:taskId", protected: true, operation: routeUpdateTask},
	{method: routeMethodDelete, path: "/admin/courses/:courseId/homework/:hwId/tasks/:taskId", protected: true, operation: routeDeleteTask},
	{method: routeMethodPatch, path: "/admin/courses/:courseId/homework/:hwId/tasks/:taskId/score", protected: true, operation: routeSetScore},
	{method: routeMethodPost, path: "/admin/courses/:courseId/roles", protected: true, operation: routeAssignCourseAdmin},
	{method: routeMethodDelete, path: "/admin/courses/:courseId/roles", protected: true, operation: routeRevokeCourseAdmin},
	{method: routeMethodDelete, path: "/admin/courses/:courseId/participants", protected: true, operation: routeRemoveParticipant},
	{method: routeMethodGet, path: "/admin/courses/:courseId/roles", protected: true, operation: routeListUserRoles},
	{method: routeMethodPost, path: "/admin/courses/:courseId/roles/:roleId/permissions", protected: true, operation: routeAddPermission},
	{method: routeMethodDelete, path: "/admin/courses/:courseId/roles/:roleId/permissions/:permission", protected: true, operation: routeRemovePermission},
	{method: routeMethodGet, path: "/admin/courses/:courseId/roles/:roleId/permissions", protected: true, operation: routeListPermissions},
	{method: routeMethodPost, path: "/admin/super-admins", protected: true, operation: routeCreateSuperAdmin},
}

func RegisterRoutes(
	e *echo.Echo,
	apiController *APIController,
	adminHomework *handler.AdminHomeworkHandler,
	adminTask *handler.AdminTaskHandler,
	adminRole *handler.AdminRoleHandler,
) {
	wrapper := api.ServerInterfaceWrapper{Handler: apiController}

	for _, route := range routeCatalog {
		registerRoute(e, route, resolveRouteHandler(route.operation, wrapper, apiController, adminHomework, adminTask, adminRole))
	}
}

func ProtectedPaths() []string {
	protected := make(map[string]struct{})
	for _, route := range routeCatalog {
		if route.protected {
			protected[route.path] = struct{}{}
		}
	}

	paths := make([]string, 0, len(protected))
	for path := range protected {
		paths = append(paths, path)
	}

	sort.Strings(paths)
	return paths
}

func resolveRouteHandler(
	operation routeOperation,
	wrapper api.ServerInterfaceWrapper,
	apiController *APIController,
	adminHomework *handler.AdminHomeworkHandler,
	adminTask *handler.AdminTaskHandler,
	adminRole *handler.AdminRoleHandler,
) echo.HandlerFunc {
	switch operation {
	case routeSignIn:
		return wrapper.SignIn
	case routeSignOut:
		return wrapper.SignOut
	case routeSignUp:
		return wrapper.SignUp
	case routeGetMe:
		return wrapper.GetMe
	case routePostV1Echo:
		return wrapper.PostV1Echo
	case routeGetSessions:
		return wrapper.GetSessions
	case routeCreateUser:
		return wrapper.CreateUser
	case routeGetUserByEmail:
		return wrapper.GetUserByEmail
	case routeGetUsersWithSessions:
		return wrapper.GetUsersWithSessions
	case routeGetUserByUsername:
		return wrapper.GetUserByUsername
	case routeGetUserByID:
		return wrapper.GetUserByID
	case routeGetCourses:
		return apiController.courseHandler.GetCourses
	case routeCreateCourse:
		return apiController.courseHandler.CreateCourse
	case routeGetCourse:
		return apiController.courseHandler.GetCourse
	case routeUpdateCourse:
		return apiController.courseHandler.UpdateCourse
	case routeGetCourseBoard:
		return apiController.courseHandler.GetCourseBoard
	case routeGetCourseScores:
		return apiController.courseHandler.GetScores
	case routeJoinCourse:
		return apiController.courseHandler.JoinCourse
	case routeCreateHomework:
		return adminHomework.CreateHomework
	case routeGetHomework:
		return adminHomework.GetHomework
	case routeListHomework:
		return adminHomework.ListHomework
	case routeUpdateHomework:
		return adminHomework.UpdateHomework
	case routeDeleteHomework:
		return adminHomework.DeleteHomework
	case routePublishHomework:
		return adminHomework.PublishHomework
	case routeSetDeadline:
		return adminHomework.SetDeadline
	case routeUpdateDeadline:
		return adminHomework.UpdateDeadline
	case routeDeleteDeadline:
		return adminHomework.DeleteDeadline
	case routeCreateTask:
		return adminTask.CreateTask
	case routeListTasks:
		return adminTask.ListTasks
	case routeGetTask:
		return adminTask.GetTask
	case routeUpdateTask:
		return adminTask.UpdateTask
	case routeDeleteTask:
		return adminTask.DeleteTask
	case routeSetScore:
		return adminTask.SetScore
	case routeAssignCourseAdmin:
		return adminRole.AssignCourseAdmin
	case routeRevokeCourseAdmin:
		return adminRole.RevokeCourseAdmin
	case routeRemoveParticipant:
		return adminRole.RemoveCourseParticipant
	case routeListUserRoles:
		return adminRole.ListUserRoles
	case routeAddPermission:
		return adminRole.AddPermission
	case routeRemovePermission:
		return adminRole.RemovePermission
	case routeListPermissions:
		return adminRole.ListPermissions
	case routeCreateSuperAdmin:
		return adminRole.CreateSuperAdmin
	default:
		panic("unsupported route operation: " + string(operation))
	}
}

func registerRoute(e *echo.Echo, route routeSpec, handler echo.HandlerFunc) {
	switch route.method {
	case routeMethodGet:
		e.GET(route.path, handler)
	case routeMethodPost:
		e.POST(route.path, handler)
	case routeMethodPut:
		e.PUT(route.path, handler)
	case routeMethodPatch:
		e.PATCH(route.path, handler)
	case routeMethodDelete:
		e.DELETE(route.path, handler)
	default:
		panic("unsupported route method: " + string(route.method))
	}
}
