package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type MockHomeworkRepo struct {
	mock.Mock
}

func (m *MockHomeworkRepo) Create(ctx context.Context, hw *model.Homework) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *MockHomeworkRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Homework, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Homework), args.Error(1)
}

func (m *MockHomeworkRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Homework, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Homework), args.Error(1)
}

func (m *MockHomeworkRepo) Update(ctx context.Context, hw *model.Homework) error {
	args := m.Called(ctx, hw)
	return args.Error(0)
}

func (m *MockHomeworkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockDeadlineRepo struct {
	mock.Mock
}

func (m *MockDeadlineRepo) Create(ctx context.Context, deadline *model.Deadline) error {
	args := m.Called(ctx, deadline)
	return args.Error(0)
}

func (m *MockDeadlineRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Deadline, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func (m *MockDeadlineRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.Deadline, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Deadline), args.Error(1)
}

func (m *MockDeadlineRepo) Update(ctx context.Context, deadline *model.Deadline) error {
	args := m.Called(ctx, deadline)
	return args.Error(0)
}

func (m *MockDeadlineRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeadlineRepo) GetByHomeworkID(ctx context.Context, homeworkID uuid.UUID) (*model.Deadline, error) {
	args := m.Called(ctx, homeworkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func setupService() (*service.AdminHomeworkService, *MockHomeworkRepo, *MockDeadlineRepo, *MockRoleRepo) {
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	roleRepo := new(MockRoleRepo)
	roleID := uuid.New()
	roleRepo.On("GetRoleIDByUserAndCourse", mock.Anything, mock.Anything, mock.Anything).Return(roleID, nil)
	roleRepo.On("HasPermission", mock.Anything, roleID, mock.Anything).Return(true, nil)
	svc := service.NewAdminHomeworkService(hwRepo, dlRepo, roleRepo)
	return svc, hwRepo, dlRepo, roleRepo
}

func TestCreateHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CreateHomeworkInput{
		CourseID:  courseID,
		Title:     "Week 1",
		StartDate: "2025-01-01",
		EndDate:   "2025-06-01",
	}

	hwRepo.On("Create", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.CourseID == courseID &&
			hw.Title == "Week 1" &&
			hw.StartDate != nil &&
			hw.EndDate != nil
	})).Return(nil)

	result, err := svc.CreateHomework(ctx, userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, courseID, result.CourseID)
	assert.Equal(t, "Week 1", result.Title)
	hwRepo.AssertExpectations(t)
}

func TestCreateHomework_WithDescriptionAndPosition(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CreateHomeworkInput{
		CourseID:    courseID,
		Title:       "Week 2",
		Description: "Arrays and sorting",
		Position:    3,
		StartDate:   "2025-02-01",
		EndDate:     "2025-02-28",
	}

	hwRepo.On("Create", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.CourseID == courseID &&
			hw.Title == "Week 2" &&
			hw.Description != nil && *hw.Description == "Arrays and sorting" &&
			hw.Position == 3
	})).Return(nil)

	result, err := svc.CreateHomework(ctx, userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.Position)
	assert.NotNil(t, result.Description)
	assert.Equal(t, "Arrays and sorting", *result.Description)
	hwRepo.AssertExpectations(t)
}

func TestCreateHomework_EmptyCourseID(t *testing.T) {
	svc, _, _, roleRepo := setupService()
	ctx := context.Background()

	input := service.CreateHomeworkInput{
		CourseID:  uuid.Nil,
		Title:     "Test",
		StartDate: "2025-01-01",
		EndDate:   "2025-06-01",
	}

	result, err := svc.CreateHomework(ctx, uuid.New(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
	roleRepo.AssertNotCalled(t, "GetRoleIDByUserAndCourse")
}

func TestCreateHomework_EmptyTitle(t *testing.T) {
	svc, _, _, roleRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CreateHomeworkInput{
		CourseID:  courseID,
		Title:     "",
		StartDate: "2025-01-01",
		EndDate:   "2025-06-01",
	}

	result, err := svc.CreateHomework(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "title is required")
	roleRepo.AssertNotCalled(t, "GetRoleIDByUserAndCourse")
}

func TestCreateHomework_InvalidStartDateFormat(t *testing.T) {
	svc, _, _, roleRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CreateHomeworkInput{
		CourseID:  courseID,
		Title:     "Test",
		StartDate: "not-a-date",
		EndDate:   "2025-06-01",
	}

	result, err := svc.CreateHomework(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "start date must be in format")
	roleRepo.AssertNotCalled(t, "GetRoleIDByUserAndCourse")
}

func TestCreateHomework_InvalidEndDateFormat(t *testing.T) {
	svc, _, _, roleRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CreateHomeworkInput{
		CourseID:  courseID,
		Title:     "Test",
		StartDate: "2025-01-01",
		EndDate:   "not-a-date",
	}

	result, err := svc.CreateHomework(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "end date must be in format")
	roleRepo.AssertNotCalled(t, "GetRoleIDByUserAndCourse")
}

func TestCreateHomework_EndDateBeforeStartDate(t *testing.T) {
	svc, _, _, roleRepo := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CreateHomeworkInput{
		CourseID:  courseID,
		Title:     "Test",
		StartDate: "2025-12-01",
		EndDate:   "2025-01-01",
	}

	result, err := svc.CreateHomework(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "end date must be after start_date")
	roleRepo.AssertNotCalled(t, "GetRoleIDByUserAndCourse")
}

func TestCreateHomework_RepoError(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.CreateHomeworkInput{
		CourseID:  courseID,
		Title:     "Test",
		StartDate: "2025-01-01",
		EndDate:   "2025-06-01",
	}

	hwRepo.On("Create", ctx, mock.AnythingOfType("*model.Homework")).Return(errors.New("db error"))

	result, err := svc.CreateHomework(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to create homework")
	hwRepo.AssertExpectations(t)
}

func TestGetHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	expected := &model.Homework{HwID: hwID, CourseID: courseID}
	hwRepo.On("GetByID", ctx, hwID).Return(expected, nil)

	result, err := svc.GetHomework(ctx, userID, hwID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	hwRepo.AssertExpectations(t)
}

func TestGetHomework_NilID(t *testing.T) {
	svc, _, _, _ := setupService()
	ctx := context.Background()

	result, err := svc.GetHomework(ctx, uuid.New(), uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "homework ID is required")
}

func TestGetHomework_NotFound(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	hwID := uuid.New()

	hwRepo.On("GetByID", ctx, hwID).Return(nil, errors.New("not found"))

	result, err := svc.GetHomework(ctx, uuid.New(), hwID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Homework not found")
	hwRepo.AssertExpectations(t)
}

func TestListHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	expected := []model.Homework{
		{HwID: uuid.New(), CourseID: courseID, Title: "HW1"},
		{HwID: uuid.New(), CourseID: courseID, Title: "HW2"},
	}
	hwRepo.On("GetByCourseID", ctx, courseID).Return(expected, nil)

	result, err := svc.ListHomework(ctx, userID, courseID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Len(t, result, 2)
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	existing := &model.Homework{
		HwID:     hwID,
		CourseID: courseID,
		Title:    "Original Title",
	}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.Title == "Updated Title"
	})).Return(nil)

	newTitle := "Updated Title"
	input := service.UpdateHomeworkInput{
		Title: &newTitle,
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Title", result.Title)
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_UpdateEndDate(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	existing := &model.Homework{
		HwID:     hwID,
		CourseID: courseID,
	}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.Anything).Return(nil)

	input := service.UpdateHomeworkInput{
		EndDate: "2026-01-01",
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.EndDate)
	assert.Equal(t, "2026-01-01", result.EndDate.Format("2006-01-02"))
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_UpdatePositionToZero(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	existing := &model.Homework{
		HwID:     hwID,
		CourseID: courseID,
		Position: 5,
	}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.Position == 0
	})).Return(nil)

	pos := 0
	input := service.UpdateHomeworkInput{
		Position: &pos,
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Position)
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_DescriptionNotChangedIfNil(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	desc := "Original description"

	existing := &model.Homework{
		HwID:        hwID,
		CourseID:    courseID,
		Description: &desc,
	}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.Description != nil && *hw.Description == "Original description"
	})).Return(nil)

	updatedTitle := "Updated Title"
	input := service.UpdateHomeworkInput{
		Title: &updatedTitle,
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Description)
	assert.Equal(t, "Original description", *result.Description)
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_ClearDescription(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()
	desc := "Old description"

	existing := &model.Homework{
		HwID:        hwID,
		CourseID:    courseID,
		Description: &desc,
	}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.Description == nil
	})).Return(nil)

	emptyDesc := ""
	input := service.UpdateHomeworkInput{
		Description: &emptyDesc,
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.Description)
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_EmptyTitleError(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	existing := &model.Homework{HwID: hwID, CourseID: courseID}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)

	emptyTitle := ""
	input := service.UpdateHomeworkInput{
		Title: &emptyTitle,
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "title cannot be empty")
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_EndBeforeStart(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	existing := &model.Homework{
		HwID:      hwID,
		CourseID:  courseID,
		StartDate: &start,
	}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)

	input := service.UpdateHomeworkInput{
		EndDate: "2025-01-01",
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "end date must be after start_date")
	hwRepo.AssertExpectations(t)
}

func TestDeleteHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	existing := &model.Homework{HwID: hwID, CourseID: courseID}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Delete", ctx, hwID).Return(nil)

	err := svc.DeleteHomework(ctx, userID, hwID)

	assert.NoError(t, err)
	hwRepo.AssertExpectations(t)
}

func TestDeleteHomework_NotFound(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	hwID := uuid.New()

	hwRepo.On("GetByID", ctx, hwID).Return(nil, errors.New("not found"))

	err := svc.DeleteHomework(ctx, uuid.New(), hwID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Homework not found")
	hwRepo.AssertExpectations(t)
}

func TestPublishHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	existing := &model.Homework{HwID: hwID, CourseID: courseID}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.IsPublic != nil && *hw.IsPublic == true
	})).Return(nil)

	result, err := svc.PublishHomework(ctx, userID, hwID, true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.IsPublic)
	assert.True(t, *result.IsPublic)
	hwRepo.AssertExpectations(t)
}

func TestSetDeadline_Success(t *testing.T) {
	svc, hwRepo, dlRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	assignedBy := uuid.New()

	existing := &model.Homework{HwID: hwID, CourseID: courseID}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	dlRepo.On("Create", ctx, mock.MatchedBy(func(dl *model.Deadline) bool {
		return dl.Title == "Final Deadline" &&
			dl.CourseID == courseID &&
			dl.HomeworkID == hwID
	})).Return(nil)

	softDeadline := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	hardDeadline := time.Date(2025, 6, 8, 0, 0, 0, 0, time.UTC)

	input := service.SetDeadlineInput{
		CourseID:     courseID,
		HomeworkID:   hwID,
		Title:        "Final Deadline",
		Description:  "Submit your work",
		SoftDeadline: softDeadline,
		HardDeadline: hardDeadline,
		AssignedBy:   &assignedBy,
	}

	result, err := svc.SetDeadline(ctx, userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Final Deadline", result.Title)
	assert.Equal(t, softDeadline, result.SoftDeadline)
	assert.Equal(t, hardDeadline, result.HardDeadline)
	hwRepo.AssertExpectations(t)
	dlRepo.AssertExpectations(t)
}

func TestSetDeadline_SoftDeadlineDefaultsToHardDeadline(t *testing.T) {
	svc, hwRepo, dlRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()

	existing := &model.Homework{HwID: hwID, CourseID: courseID}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	dlRepo.On("Create", ctx, mock.MatchedBy(func(dl *model.Deadline) bool {
		return dl.SoftDeadline.Equal(dl.HardDeadline)
	})).Return(nil)

	hardDeadline := time.Date(2025, 6, 8, 0, 0, 0, 0, time.UTC)

	input := service.SetDeadlineInput{
		CourseID:     courseID,
		HomeworkID:   hwID,
		Title:        "Deadline",
		HardDeadline: hardDeadline,
	}

	result, err := svc.SetDeadline(ctx, userID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, hardDeadline, result.SoftDeadline)
	assert.Equal(t, hardDeadline, result.HardDeadline)
}

func TestSetDeadline_EmptyCourseID(t *testing.T) {
	svc, _, _, _ := setupService()
	ctx := context.Background()

	input := service.SetDeadlineInput{
		CourseID:   uuid.Nil,
		HomeworkID: uuid.New(),
		Title:      "Test",
	}

	result, err := svc.SetDeadline(ctx, uuid.New(), input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "course_id is required")
}

func TestSetDeadline_EmptyTitle(t *testing.T) {
	svc, _, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.SetDeadlineInput{
		CourseID:   courseID,
		HomeworkID: uuid.New(),
		Title:      "",
	}

	result, err := svc.SetDeadline(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "title is required")
}

func TestSetDeadline_MissingHardDeadline(t *testing.T) {
	svc, _, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()

	input := service.SetDeadlineInput{
		CourseID:   courseID,
		HomeworkID: uuid.New(),
		Title:      "Deadline",
	}

	result, err := svc.SetDeadline(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "hard_deadline is required")
}

func TestSetDeadline_HomeworkNotFound(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(nil, errors.New("not found"))

	input := service.SetDeadlineInput{
		CourseID:     courseID,
		HomeworkID:   hwID,
		Title:        "Deadline",
		HardDeadline: time.Now(),
	}

	result, err := svc.SetDeadline(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Homework not found")
	hwRepo.AssertExpectations(t)
}

func TestSetDeadline_HomeworkWrongCourse(t *testing.T) {
	svc, hwRepo, _, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	courseID := uuid.New()
	wrongCourseID := uuid.New()
	hwID := uuid.New()

	existing := &model.Homework{HwID: hwID, CourseID: wrongCourseID}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)

	input := service.SetDeadlineInput{
		CourseID:     courseID,
		HomeworkID:   hwID,
		Title:        "Deadline",
		HardDeadline: time.Now(),
	}

	result, err := svc.SetDeadline(ctx, userID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "homework does not belong to this course")
	hwRepo.AssertExpectations(t)
}

func TestGetDeadlineByHomeworkID_DeadlineNotFound(t *testing.T) {
	svc, hwRepo, dlRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	courseID := uuid.New()

	hw := &model.Homework{HwID: hwID, CourseID: courseID}
	hwRepo.On("GetByID", ctx, hwID).Return(hw, nil)
	dlRepo.On("GetByHomeworkID", ctx, hwID).Return(nil, errors.New("not found"))

	result, err := svc.GetDeadlineByHomeworkID(ctx, userID, hwID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Deadline not found")
}

func TestUpdateDeadline_Success(t *testing.T) {
	svc, _, dlRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	dlID := uuid.New()
	courseID := uuid.New()

	existing := &model.Deadline{
		ID:       dlID,
		CourseID: courseID,
		Title:    "Old Title",
	}
	dlRepo.On("GetByID", ctx, dlID).Return(existing, nil)
	dlRepo.On("Update", ctx, mock.MatchedBy(func(dl *model.Deadline) bool {
		return dl.Title == "Updated Title"
	})).Return(nil)

	input := service.UpdateDeadlineInput{
		Title: "Updated Title",
	}

	result, err := svc.UpdateDeadline(ctx, userID, dlID, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Title", result.Title)
	dlRepo.AssertExpectations(t)
}

func TestUpdateDeadline_NotFound(t *testing.T) {
	svc, _, dlRepo, _ := setupService()
	ctx := context.Background()
	dlID := uuid.New()

	dlRepo.On("GetByID", ctx, dlID).Return(nil, errors.New("not found"))

	input := service.UpdateDeadlineInput{
		Title: "New Title",
	}

	result, err := svc.UpdateDeadline(ctx, uuid.New(), dlID, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Deadline not found")
	dlRepo.AssertExpectations(t)
}

func TestDeleteDeadline_Success(t *testing.T) {
	svc, _, dlRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	dlID := uuid.New()
	courseID := uuid.New()

	existing := &model.Deadline{ID: dlID, CourseID: courseID}
	dlRepo.On("GetByID", ctx, dlID).Return(existing, nil)
	dlRepo.On("Delete", ctx, dlID).Return(nil)

	err := svc.DeleteDeadline(ctx, userID, dlID)

	assert.NoError(t, err)
	dlRepo.AssertExpectations(t)
}

func TestDeleteDeadline_NotFound(t *testing.T) {
	svc, _, dlRepo, _ := setupService()
	ctx := context.Background()
	dlID := uuid.New()

	dlRepo.On("GetByID", ctx, dlID).Return(nil, errors.New("not found"))

	err := svc.DeleteDeadline(ctx, uuid.New(), dlID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Deadline not found")
	dlRepo.AssertExpectations(t)
}
