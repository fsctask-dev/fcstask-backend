//go:build integration

package repo

import (
	"context"
	"testing"
	"time"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func integrationCourseConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:            "localhost",
		Port:            6432,
		Username:        "postgres",
		Password:        "postgres",
		Database:        "fcstask",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Minute,
	}
}

func migrateCourseAccessModels(t *testing.T, client *db.Client) {
	t.Helper()
	tx := client.WriteDB().Session(&gorm.Session{})
	tx.Config.IgnoreRelationshipsWhenMigrating = true
	if err := tx.AutoMigrate(&model.Course{}, &model.UserRole{}, &model.CourseAdminPermission{}); err != nil {
		t.Fatalf("AutoMigrate course access models: %v", err)
	}
}

func createCourseFixture(t *testing.T, client *db.Client, course model.Course) {
	t.Helper()
	if err := client.WriteDB().Create(&course).Error; err != nil {
		t.Fatalf("Create course: %v", err)
	}
	t.Cleanup(func() {
		_ = client.WriteDB().Where("id = ?", course.ID).Delete(&model.Course{}).Error
	})
}

func createUserRoleFixture(t *testing.T, client *db.Client, role model.UserRole) {
	t.Helper()
	if err := client.WriteDB().Create(&role).Error; err != nil {
		t.Fatalf("Create user role: %v", err)
	}
	t.Cleanup(func() {
		_ = client.WriteDB().Where("user_id = ? AND course_id = ? AND role_id = ?", role.UserID, role.CourseID, role.RoleID).
			Delete(&model.UserRole{}).Error
	})
}

func createPermissionFixture(t *testing.T, client *db.Client, perm model.CourseAdminPermission) {
	t.Helper()
	if err := client.WriteDB().Create(&perm).Error; err != nil {
		t.Fatalf("Create permission: %v", err)
	}
	t.Cleanup(func() {
		_ = client.WriteDB().Where("role_id = ? AND permission = ?", perm.RoleID, perm.Permission).
			Delete(&model.CourseAdminPermission{}).Error
	})
}

func TestCourseRepository_GetPublicCourses_HidesHidden(t *testing.T) {
	ctx := context.Background()
	client, err := db.New(integrationCourseConfig())
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	migrateCourseAccessModels(t, client)

	repo := NewCourseRepository(client)
	visibleID := uuid.New()
	hiddenID := uuid.New()

	createCourseFixture(t, client, model.Course{
		ID:     visibleID,
		Name:   "Visible Public",
		Slug:   "visible-public-" + visibleID.String(),
		Status: "created",
		Type:   model.CourseTypePublic,
		URL:    "/course/visible-public",
	})
	createCourseFixture(t, client, model.Course{
		ID:     hiddenID,
		Name:   "Hidden Public",
		Slug:   "hidden-public-" + hiddenID.String(),
		Status: "hidden",
		Type:   model.CourseTypePublic,
		URL:    "/course/hidden-public",
	})

	courses, err := repo.GetPublicCourses(ctx)
	if err != nil {
		t.Fatalf("GetPublicCourses: %v", err)
	}

	var foundVisible, foundHidden bool
	for _, course := range courses {
		if course.ID == visibleID {
			foundVisible = true
		}
		if course.ID == hiddenID {
			foundHidden = true
		}
	}

	if !foundVisible {
		t.Fatalf("expected visible public course in result")
	}
	if foundHidden {
		t.Fatalf("did not expect hidden public course in result")
	}
}

func TestCourseRepository_GetCoursesByUserID_HiddenRequiresPermission(t *testing.T) {
	ctx := context.Background()
	client, err := db.New(integrationCourseConfig())
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	migrateCourseAccessModels(t, client)

	repo := NewCourseRepository(client)
	userID := uuid.New()
	visibleID := uuid.New()
	hiddenID := uuid.New()
	visibleRoleID := uuid.New()
	hiddenRoleID := uuid.New()

	createCourseFixture(t, client, model.Course{
		ID:     visibleID,
		Name:   "Visible Course",
		Slug:   "visible-course-" + visibleID.String(),
		Status: "created",
		Type:   model.CourseTypePrivate,
		URL:    "/course/visible-course",
	})
	createCourseFixture(t, client, model.Course{
		ID:     hiddenID,
		Name:   "Hidden Course",
		Slug:   "hidden-course-" + hiddenID.String(),
		Status: "hidden",
		Type:   model.CourseTypePrivate,
		URL:    "/course/hidden-course",
	})

	createUserRoleFixture(t, client, model.UserRole{UserID: userID, CourseID: visibleID, RoleID: visibleRoleID})
	createUserRoleFixture(t, client, model.UserRole{UserID: userID, CourseID: hiddenID, RoleID: hiddenRoleID})
	createPermissionFixture(t, client, model.CourseAdminPermission{RoleID: visibleRoleID, Permission: "course.read"})

	courses, err := repo.GetCoursesByUserID(ctx, userID, "")
	if err != nil {
		t.Fatalf("GetCoursesByUserID: %v", err)
	}

	var foundVisible, foundHidden bool
	for _, course := range courses {
		if course.ID == visibleID {
			foundVisible = true
		}
		if course.ID == hiddenID {
			foundHidden = true
		}
	}

	if !foundVisible {
		t.Fatalf("expected visible course in result")
	}
	if foundHidden {
		t.Fatalf("did not expect hidden course without permission")
	}
}

func TestCourseRepository_GetCoursesByUserID_HiddenVisibleForSuperAdmin(t *testing.T) {
	ctx := context.Background()
	client, err := db.New(integrationCourseConfig())
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	migrateCourseAccessModels(t, client)

	repo := NewCourseRepository(client)
	userID := uuid.New()
	hiddenID := uuid.New()
	globalRoleID := uuid.New()

	createCourseFixture(t, client, model.Course{
		ID:     hiddenID,
		Name:   "Hidden Course",
		Slug:   "hidden-super-" + hiddenID.String(),
		Status: "hidden",
		Type:   model.CourseTypePrivate,
		URL:    "/course/hidden-super",
	})
	createUserRoleFixture(t, client, model.UserRole{UserID: userID, CourseID: uuid.Nil, RoleID: globalRoleID})
	createPermissionFixture(t, client, model.CourseAdminPermission{RoleID: globalRoleID, Permission: "is_super_admin"})

	courses, err := repo.GetCoursesByUserID(ctx, userID, "")
	if err != nil {
		t.Fatalf("GetCoursesByUserID: %v", err)
	}

	var foundHidden bool
	for _, course := range courses {
		if course.ID == hiddenID {
			foundHidden = true
		}
	}

	if !foundHidden {
		t.Fatalf("expected hidden course in result for super admin")
	}
}
