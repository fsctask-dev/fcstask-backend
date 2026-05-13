package handler

import (
	"context"
	"github.com/google/uuid"
	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type IAdminHomeworkService interface {
	CreateHomework(ctx context.Context, userID uuid.UUID, input service.CreateHomeworkInput) (*model.Homework, error)
	GetHomework(ctx context.Context, userID uuid.UUID, hwID uuid.UUID) (*model.Homework, error)
	ListHomework(ctx context.Context, userID uuid.UUID, courseID uuid.UUID) ([]model.Homework, error)
	UpdateHomework(ctx context.Context, userID uuid.UUID, hwID uuid.UUID, input service.UpdateHomeworkInput) (*model.Homework, error)
	DeleteHomework(ctx context.Context, userID uuid.UUID, hwID uuid.UUID) error
	PublishHomework(ctx context.Context, userID uuid.UUID, hwID uuid.UUID, isPublic bool) (*model.Homework, error)
	SetDeadline(ctx context.Context, userID uuid.UUID, input service.SetDeadlineInput) (*model.Deadline, error)
	UpdateDeadline(ctx context.Context, userID uuid.UUID, deadlineID uuid.UUID, input service.UpdateDeadlineInput) (*model.Deadline, error)
	DeleteDeadline(ctx context.Context, userID uuid.UUID, deadlineID uuid.UUID) error
}

type IAdminRoleService interface {
	AssignRole(ctx context.Context, userID uuid.UUID, input service.AssignRoleInput) (*model.UserRole, error)
	RevokeRole(ctx context.Context, userID uuid.UUID, input service.RevokeRoleInput) error
	ListUserRoles(ctx context.Context, userID uuid.UUID, courseID uuid.UUID) ([]model.UserRole, error)
	AddPermission(ctx context.Context, input service.AddPermissionInput) (*model.CourseAdminPermission, error)
	RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error
	ListPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error)
}

type IAdminTaskService interface {
	CreateTask(ctx context.Context, userID uuid.UUID, input service.CreateTaskInput) (*model.Task, error)
	GetTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) (*model.Task, error)
	ListTasks(ctx context.Context, userID uuid.UUID, hwID uuid.UUID) ([]model.Task, error)
	UpdateTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, input service.UpdateTaskInput) (*model.Task, error)
	DeleteTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID) error
	SetScore(ctx context.Context, userID uuid.UUID, input service.SetTaskScoreInput) (*model.Task, error)
}