package service

import (
	"context"
	"github.com/google/uuid"
	"fcstask-backend/internal/db/repo"
)

func IsCourseAdmin(ctx context.Context, roleRepo repo.IRoleRepo, userID, courseID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || courseID == uuid.Nil {
		return false, nil
	}

	roles, err := roleRepo.GetUserCourseRoles(ctx, userID, courseID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		perms, err := roleRepo.GetPermissions(ctx, role.RoleID)
		if err != nil {
			return false, err
		}
		for _, p := range perms {
			if p.Permission == "admin" {
				return true, nil
			}
		}
	}

	return false, nil
}

func IsCourseParticipant(ctx context.Context, roleRepo repo.IRoleRepo, userID, courseID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || courseID == uuid.Nil {
		return false, nil
	}

	roles, err := roleRepo.GetUserCourseRoles(ctx, userID, courseID)
	if err != nil {
		return false, err
	}

	return len(roles) > 0, nil
}