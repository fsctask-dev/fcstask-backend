package handler

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

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
	GetDeadlineByHomeworkID(ctx context.Context, userID, hwID uuid.UUID) (*model.Deadline, error)
}

type IAdminRoleService interface {
	CreateSuperAdmin(ctx context.Context, userID uuid.UUID, input service.CreateSuperAdminInput) (*model.UserRole, error)
	AssignCourseAdmin(ctx context.Context, userID uuid.UUID, input service.AssignCourseAdminInput) (*model.UserRole, error)
	GrantCourseCreatePermission(ctx context.Context, userID uuid.UUID, targetUserID uuid.UUID) (*model.UserRole, error)
    RevokeCourseCreatePermission(ctx context.Context, userID uuid.UUID, targetUserID uuid.UUID) error
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
	PublishTask(ctx context.Context, userID uuid.UUID, input service.PublishTaskInput) (*model.Task, error)
	DeleteTask(ctx context.Context, userID, taskID uuid.UUID) error
	SetScore(ctx context.Context, userID uuid.UUID, input service.SetTaskScoreInput) (*model.Task, error)
}
