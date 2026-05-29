package handler

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type IAuthService interface {
	SignUp(ctx context.Context, input service.SignUpInput) (*service.AuthResult, error)
	SignIn(ctx context.Context, input service.SignInInput) (*service.AuthResult, error)
	GetMe(ctx context.Context, user *model.User) (string, string, error)
	SignOut(ctx context.Context, session *model.Session) error
}

type IUserService interface {
	CreateUser(ctx context.Context, input service.CreateUserInput) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUsersWithSessions(ctx context.Context, limit, offset int) ([]model.User, int64, error)
}

type ISessionService interface {
	GetSessions(ctx context.Context, limit, offset int) ([]model.Session, int64, error)
}

type ICourseService interface {
	GetCourses(ctx context.Context, userID uuid.UUID, status string) ([]model.Course, error)
	GetCourse(ctx context.Context, courseID string) (*model.Course, error)
	CanReadCourse(ctx context.Context, userID uuid.UUID, course *model.Course) (bool, error)
	CreateCourse(ctx context.Context, userID uuid.UUID, input service.CourseInput) (*model.Course, error)
	UpdateCourse(ctx context.Context, userID uuid.UUID, courseID string, input service.CourseInput) (*model.Course, error)
	GetCourseBoard(ctx context.Context, courseID string) (*model.TaskBoardSummary, error)
	JoinCourse(ctx context.Context, userID uuid.UUID, courseID string, code string) error
	GetLeaderboard(ctx context.Context, userID uuid.UUID, courseID string) ([]model.LeaderboardEntry, error)
}

type IAdminHomeworkService interface {
	CreateHomework(ctx context.Context, userID uuid.UUID, input service.CreateHomeworkInput) (*model.Homework, error)
	GetHomework(ctx context.Context, userID, hwID uuid.UUID) (*model.Homework, error)
	ListHomework(ctx context.Context, userID, courseID uuid.UUID) ([]model.Homework, error)
	UpdateHomework(ctx context.Context, userID, hwID uuid.UUID, input service.UpdateHomeworkInput) (*model.Homework, error)
	DeleteHomework(ctx context.Context, userID, hwID uuid.UUID) error
	PublishHomework(ctx context.Context, userID, hwID uuid.UUID, isPublic bool) (*model.Homework, error)
	SetDeadline(ctx context.Context, userID uuid.UUID, input service.SetDeadlineInput) (*model.Deadline, error)
	UpdateDeadline(ctx context.Context, userID, deadlineID uuid.UUID, input service.UpdateDeadlineInput) (*model.Deadline, error)
	DeleteDeadline(ctx context.Context, userID, deadlineID uuid.UUID) error
}

type IAdminRoleService interface {
	CreateSuperAdmin(ctx context.Context, userID uuid.UUID, input service.CreateSuperAdminInput) (*model.UserRole, error)
	AssignCourseAdmin(ctx context.Context, userID uuid.UUID, input service.AssignCourseAdminInput) (*model.UserRole, error)
	RevokeCourseAdmin(ctx context.Context, userID uuid.UUID, input service.RevokeCourseAdminInput) error
	RemoveCourseParticipant(ctx context.Context, userID uuid.UUID, input service.RemoveCourseParticipantInput) error
	ListUserRoles(ctx context.Context, userID, courseID uuid.UUID) ([]model.UserRole, error)
	AddPermission(ctx context.Context, userID uuid.UUID, input service.AddPermissionInput) (*model.CourseAdminPermission, error)
	RemovePermission(ctx context.Context, userID, courseID, roleID uuid.UUID, permission string) error
	ListPermissions(ctx context.Context, userID, courseID, roleID uuid.UUID) ([]model.CourseAdminPermission, error)
}

type IAdminTaskService interface {
	CreateTask(ctx context.Context, userID uuid.UUID, input service.CreateTaskInput) (*model.Task, error)
	GetTask(ctx context.Context, userID, taskID uuid.UUID) (*model.Task, error)
	ListTasks(ctx context.Context, userID, hwID uuid.UUID) ([]model.Task, error)
	UpdateTask(ctx context.Context, userID, taskID uuid.UUID, input service.UpdateTaskInput) (*model.Task, error)
	DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error
	SetScore(ctx context.Context, userID uuid.UUID, input service.SetTaskScoreInput) (*model.Task, error)
}

var (
	_ IAuthService          = (*service.AuthService)(nil)
	_ IUserService          = (*service.UserService)(nil)
	_ ISessionService       = (*service.SessionService)(nil)
	_ ICourseService        = (*service.CourseService)(nil)
	_ IAdminHomeworkService = (*service.AdminHomeworkService)(nil)
	_ IAdminRoleService     = (*service.AdminRoleService)(nil)
	_ IAdminTaskService     = (*service.AdminTaskService)(nil)
)
