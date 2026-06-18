package service

import (
	"context"

	"github.com/google/uuid"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/db/repo"
)

func requireHomeworkInCourse(ctx context.Context, homeworkRepo repo.IHomeworkRepo, hwID, courseID uuid.UUID) (*model.Homework, error) {
	hw, err := homeworkRepo.GetByID(ctx, hwID)
	if err != nil {
		return nil, NotFound("Homework not found")
	}
	if hw.CourseID != courseID {
		return nil, NotFound("Homework not found")
	}
	return hw, nil
}

func requireTaskInHomework(ctx context.Context, taskRepo repo.ITaskRepo, taskID, hwID uuid.UUID) (*model.Task, error) {
	task, err := taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, NotFound("Task not found")
	}
	if task.HwID != hwID {
		return nil, NotFound("Task not found")
	}
	return task, nil
}
