package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type MockHomeworkRepo struct{ mock.Mock }

func (m *MockHomeworkRepo) Create(ctx context.Context, hw *model.Homework) error {
	return m.Called(ctx, hw).Error(0)
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
	return m.Called(ctx, hw).Error(0)
}
func (m *MockHomeworkRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type MockDeadlineRepo struct{ mock.Mock }

func (m *MockDeadlineRepo) Create(ctx context.Context, d *model.Deadline) error {
	return m.Called(ctx, d).Error(0)
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

func (m *MockDeadlineRepo) Update(ctx context.Context, d *model.Deadline) error {
	return m.Called(ctx, d).Error(0)
}

func (m *MockDeadlineRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockDeadlineRepo) GetByHomeworkID(ctx context.Context, hwID uuid.UUID) (*model.Deadline, error) {
	args := m.Called(ctx, hwID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deadline), args.Error(1)
}

func setupHomeworkService() (*service.AdminHomeworkService, *MockHomeworkRepo, *MockDeadlineRepo, *MockHwDeadlineRepo) {
	hwRepo := new(MockHomeworkRepo)
	dlRepo := new(MockDeadlineRepo)
	hwDlRepo := new(MockHwDeadlineRepo)
	roleRepo := newPermissiveRoleRepo()
	svc := service.NewAdminHomeworkService(hwRepo, dlRepo, roleRepo, hwDlRepo)
	return svc, hwRepo, dlRepo, hwDlRepo
}

func softHard() (time.Time, time.Time) {
	soft := time.Now().Add(24 * time.Hour)
	hard := time.Now().Add(48 * time.Hour)
	return soft, hard
}

func TestCreateHomework_Success(t *testing.T) {
	svc, hwRepo, _, hwDlRepo := setupHomeworkService()
	ctx := context.Background()
	courseID := uuid.New()
	soft, hard := softHard()

	hwRepo.On("Create", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.CourseID == courseID &&
			hw.Title == "Week 1" &&
			hw.StartDate != nil &&
			hw.EndDate != nil
	})).Return(nil)

	hwDlRepo.On("Create", ctx, mock.MatchedBy(func(d *model.HomeworkDeadline) bool {
		return d.SoftDeadline.Equal(soft) && d.HardDeadline.Equal(hard)
	})).Return(nil)

	result, err := svc.CreateHomework(ctx, uuid.New(), service.CreateHomeworkInput{
		CourseID:     courseID,
		Title:        "Week 1",
		StartDate:    "2025-01-01",
		EndDate:      "2025-06-01",
		SoftDeadline: soft,
		HardDeadline: hard,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, courseID, result.CourseID)
	assert.Equal(t, "Week 1", result.Title)
	assert.NotNil(t, result.StartDate)
	assert.NotNil(t, result.EndDate)

	hwRepo.AssertExpectations(t)
	hwDlRepo.AssertExpectations(t)
}

func TestCreateHomework_EmptyTitle(t *testing.T) {
	svc, _, _ := setupService()
	ctx := context.Background()

	input := service.CreateHomeworkInput{
		CourseID:  uuid.New(),
		Title:     "",
		StartDate: "2025-01-01",
		EndDate:   "2025-06-01",
	}

	result, err := svc.CreateHomework(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "title is required")
}

func TestCreateHomework_WithDescriptionAndPosition(t *testing.T) {
	svc, hwRepo, _ := setupService()
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
	svc, _, _, _ := setupHomeworkService()
	soft, hard := softHard()
	_, err := svc.CreateHomework(context.Background(), uuid.New(), service.CreateHomeworkInput{
		CourseID: uuid.Nil, StartDate: "2025-01-01", EndDate: "2025-06-01",
		SoftDeadline: soft, HardDeadline: hard,
	})
	assert.ErrorContains(t, err, "course_id is required")
}

func TestCreateHomework_InvalidStartDate(t *testing.T) {
	svc, _, _, _ := setupHomeworkService()
	soft, hard := softHard()

	_, err := svc.CreateHomework(context.Background(), uuid.New(), service.CreateHomeworkInput{
		CourseID:     uuid.New(),
		Title:        "Week 1",
		StartDate:    "not-a-date",
		EndDate:      "2025-06-01",
		SoftDeadline: soft,
		HardDeadline: hard,
	})
	assert.ErrorContains(t, err, "start date must be in format")
}

func TestCreateHomework_EndBeforeStart(t *testing.T) {
	svc, _, _, _ := setupHomeworkService()
	soft, hard := softHard()

	_, err := svc.CreateHomework(context.Background(), uuid.New(), service.CreateHomeworkInput{
		CourseID:     uuid.New(),
		Title:        "Week 1",
		StartDate:    "2025-12-01",
		EndDate:      "2025-01-01",
		SoftDeadline: soft,
		HardDeadline: hard,
	})
	assert.ErrorContains(t, err, "end date must be after start_date")
}

func TestCreateHomework_MissingSoftDeadline(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwRepo.On("Create", ctx, mock.Anything).Return(nil)
	_, err := svc.CreateHomework(ctx, uuid.New(), service.CreateHomeworkInput{
		CourseID: uuid.New(), StartDate: "2025-01-01", EndDate: "2025-06-01",
		HardDeadline: time.Now().Add(48 * time.Hour),
	})
	assert.ErrorContains(t, err, "soft_deadline is required")
}

func TestCreateHomework_HardBeforeSoft(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwRepo.On("Create", ctx, mock.Anything).Return(nil)
	now := time.Now()
	_, err := svc.CreateHomework(ctx, uuid.New(), service.CreateHomeworkInput{
		CourseID: uuid.New(), StartDate: "2025-01-01", EndDate: "2025-06-01",
		SoftDeadline: now.Add(48 * time.Hour),
		HardDeadline: now.Add(24 * time.Hour),
	})
	assert.ErrorContains(t, err, "hard_deadline must be after soft_deadline")
}

func TestCreateHomework_RepoError(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	soft, hard := softHard()

	input := service.CreateHomeworkInput{
		CourseID:     uuid.New(),
		Title:        "Week 1",
		StartDate:    "2025-01-01",
		EndDate:      "2025-06-01",
		SoftDeadline: soft,
		HardDeadline: hard,
	}

	hwRepo.On("Create", ctx, mock.AnythingOfType("*model.Homework")).Return(assert.AnError)

	result, err := svc.CreateHomework(ctx, uuid.New(), input)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Failed to create homework")
	hwRepo.AssertExpectations(t)
}

func TestGetHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwID := uuid.New()
	expected := &model.Homework{HwID: hwID, CourseID: uuid.New()}
	hwRepo.On("GetByID", ctx, hwID).Return(expected, nil)
	result, err := svc.GetHomework(ctx, uuid.New(), hwID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetHomework_NilID(t *testing.T) {
	svc, _, _, _ := setupHomeworkService()
	_, err := svc.GetHomework(context.Background(), uuid.New(), uuid.Nil)
	assert.ErrorContains(t, err, "homework ID is required")
}

func TestGetHomework_NotFound(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(nil, assert.AnError)
	_, err := svc.GetHomework(ctx, uuid.New(), hwID)
	assert.ErrorContains(t, err, "Homework not found")
}

func TestListHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	courseID := uuid.New()
	expected := []model.Homework{{HwID: uuid.New()}}
	hwRepo.On("GetByCourseID", ctx, courseID).Return(expected, nil)
	result, err := svc.ListHomework(ctx, uuid.New(), courseID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestListHomework_NilCourseID(t *testing.T) {
	svc, _, _, _ := setupHomeworkService()
	_, err := svc.ListHomework(context.Background(), uuid.New(), uuid.Nil)
	assert.ErrorContains(t, err, "course ID is required")
}

func TestUpdateHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwID := uuid.New()
	existing := &model.Homework{HwID: hwID, CourseID: uuid.New()}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.EndDate != nil && hw.EndDate.Format("2006-01-02") == "2026-01-01"
	})).Return(nil)
	result, err := svc.UpdateHomework(ctx, uuid.New(), hwID, service.UpdateHomeworkInput{EndDate: "2026-01-01"})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestUpdateHomework_WithPositionZero(t *testing.T) {
	svc, hwRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()

	existing := &model.Homework{HwID: hwID, CourseID: uuid.New()}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.Position == 0 && hw.Title == "First HW"
	})).Return(nil)

	pos := 0
	firstHW := "First HW"
	input := service.UpdateHomeworkInput{
		Title:    &firstHW,
		Position: &pos,
	}

	result, err := svc.UpdateHomework(ctx, userID, hwID, input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.Position)
	hwRepo.AssertExpectations(t)
}

func TestUpdateHomework_DescriptionNotChangedIfNil(t *testing.T) {
	svc, hwRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	desc := "Original description"

	existing := &model.Homework{HwID: hwID, CourseID: uuid.New(), Description: &desc}
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
	svc, hwRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	desc := "Old description"
	emptyDesc := ""

	existing := &model.Homework{HwID: hwID, CourseID: uuid.New(), Description: &desc}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.Description == nil
	})).Return(nil)

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
	svc, hwRepo, _ := setupService()
	ctx := context.Background()
	userID := uuid.New()
	hwID := uuid.New()
	emptyTitle := ""

	existing := &model.Homework{HwID: hwID, CourseID: uuid.New()}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)

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
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwID := uuid.New()
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	existing := &model.Homework{HwID: hwID, CourseID: uuid.New(), StartDate: &start}
	hwRepo.On("GetByID", ctx, hwID).Return(existing, nil)
	_, err := svc.UpdateHomework(ctx, uuid.New(), hwID, service.UpdateHomeworkInput{EndDate: "2025-01-01"})
	assert.ErrorContains(t, err, "end date must be after start_date")
}

func TestDeleteHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	hwRepo.On("Delete", ctx, hwID).Return(nil)
	assert.NoError(t, svc.DeleteHomework(ctx, uuid.New(), hwID))
}

func TestDeleteHomework_NotFound(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(nil, assert.AnError)
	assert.ErrorContains(t, svc.DeleteHomework(ctx, uuid.New(), hwID), "Homework not found")
}

func TestPublishHomework_Success(t *testing.T) {
	svc, hwRepo, _, _ := setupHomeworkService()
	ctx := context.Background()
	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: uuid.New()}, nil)
	hwRepo.On("Update", ctx, mock.MatchedBy(func(hw *model.Homework) bool {
		return hw.IsPublic != nil && *hw.IsPublic
	})).Return(nil)
	result, err := svc.PublishHomework(ctx, uuid.New(), hwID, true)
	assert.NoError(t, err)
	assert.True(t, *result.IsPublic)
}

func TestSetDeadline_Success(t *testing.T) {
	svc, hwRepo, dlRepo, _ := setupHomeworkService()
	ctx := context.Background()
	courseID := uuid.New()
	hwID := uuid.New()
	hwRepo.On("GetByID", ctx, hwID).Return(&model.Homework{HwID: hwID, CourseID: courseID}, nil)
	dlRepo.On("Create", ctx, mock.MatchedBy(func(dl *model.Deadline) bool {
		return dl.Title == "Deadline 1" && dl.CourseID == courseID
	})).Return(nil)
	result, err := svc.SetDeadline(ctx, uuid.New(), service.SetDeadlineInput{
		CourseID: courseID, HomeworkID: hwID,
		Title: "Deadline 1", DueDate: "2025-12-31T23:59:59Z",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Deadline 1", result.Title)
}

func TestSetDeadline_MissingTitle(t *testing.T) {
	svc, _, _, _ := setupHomeworkService()
	_, err := svc.SetDeadline(context.Background(), uuid.New(), service.SetDeadlineInput{
		CourseID: uuid.New(), HomeworkID: uuid.New(), DueDate: "2025-12-31T23:59:59Z",
	})
	assert.ErrorContains(t, err, "title is required")
}

func TestSetDeadline_InvalidDueDate(t *testing.T) {
	svc, _, _, _ := setupHomeworkService()
	_, err := svc.SetDeadline(context.Background(), uuid.New(), service.SetDeadlineInput{
		CourseID: uuid.New(), HomeworkID: uuid.New(), Title: "D", DueDate: "bad",
	})
	assert.ErrorContains(t, err, "due date must be in RFC3339 format")
}

func TestUpdateDeadline_Success(t *testing.T) {
	svc, _, dlRepo, _ := setupHomeworkService()
	ctx := context.Background()
	dlID := uuid.New()
	dlRepo.On("GetByID", ctx, dlID).Return(&model.Deadline{ID: dlID, CourseID: uuid.New(), Title: "Old"}, nil)
	dlRepo.On("Update", ctx, mock.MatchedBy(func(dl *model.Deadline) bool {
		return dl.Title == "Updated"
	})).Return(nil)
	result, err := svc.UpdateDeadline(ctx, uuid.New(), dlID, service.UpdateDeadlineInput{Title: "Updated"})
	assert.NoError(t, err)
	assert.Equal(t, "Updated", result.Title)
}

func TestUpdateDeadline_NotFound(t *testing.T) {
	svc, _, dlRepo, _ := setupHomeworkService()
	ctx := context.Background()
	dlID := uuid.New()
	dlRepo.On("GetByID", ctx, dlID).Return(nil, assert.AnError)
	_, err := svc.UpdateDeadline(ctx, uuid.New(), dlID, service.UpdateDeadlineInput{Title: "X"})
	assert.ErrorContains(t, err, "Deadline not found")
}

func TestDeleteDeadline_Success(t *testing.T) {
	svc, _, dlRepo, _ := setupHomeworkService()
	ctx := context.Background()
	dlID := uuid.New()
	dlRepo.On("GetByID", ctx, dlID).Return(&model.Deadline{ID: dlID, CourseID: uuid.New()}, nil)
	dlRepo.On("Delete", ctx, dlID).Return(nil)
	assert.NoError(t, svc.DeleteDeadline(ctx, uuid.New(), dlID))
}

func TestDeleteDeadline_NotFound(t *testing.T) {
	svc, _, dlRepo, _ := setupHomeworkService()
	ctx := context.Background()
	dlID := uuid.New()
	dlRepo.On("GetByID", ctx, dlID).Return(nil, assert.AnError)
	assert.ErrorContains(t, svc.DeleteDeadline(ctx, uuid.New(), dlID), "Deadline not found")
}
