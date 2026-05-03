package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fcstask-backend/internal/db/model"
)

type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) AssignRole(ctx context.Context, userRole *model.UserRole) error {
	args := m.Called(ctx, userRole)
	return args.Error(0)
}

func (m *MockRoleRepo) RevokeRole(ctx context.Context, userID, courseID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, courseID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepo) GetByCourseID(ctx context.Context, courseID uuid.UUID) ([]model.UserRole, error) {
	args := m.Called(ctx, courseID)
	return args.Get(0).([]model.UserRole), args.Error(1)
}

func (m *MockRoleRepo) AddPermission(ctx context.Context, perm *model.CourseAdminPermission) error {
	args := m.Called(ctx, perm)
	return args.Error(0)
}

func (m *MockRoleRepo) RemovePermission(ctx context.Context, roleID uuid.UUID, permission string) error {
	args := m.Called(ctx, roleID, permission)
	return args.Error(0)
}

func (m *MockRoleRepo) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]model.CourseAdminPermission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]model.CourseAdminPermission), args.Error(1)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func setupTest(t *testing.T) (*echo.Echo, *MockRoleRepo, *MockUserRepo, *AdminRoleHandler) {
	e := echo.New()
	mockRoleRepo := new(MockRoleRepo)
	mockUserRepo := new(MockUserRepo)
	handler := NewAdminRoleHandler(mockRoleRepo, mockUserRepo)
	return e, mockRoleRepo, mockUserRepo, handler
}

func TestAdminAssignRoleHandler(t *testing.T) {
	tests := []struct {
		name           string
		courseID       string
		reqBody        AssignRoleRequest
		setupMocks     func(*MockRoleRepo, *MockUserRepo, AssignRoleRequest, uuid.UUID)
		expectedStatus int
		expectedBody   bool
	}{
		{
			name:     "Success",
			courseID: uuid.New().String(),
			reqBody: AssignRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			setupMocks: func(mr *MockRoleRepo, mu *MockUserRepo, req AssignRoleRequest, courseID uuid.UUID) {
				mu.On("GetByID", mock.Anything, req.UserID).Return(&model.User{ID: req.UserID}, nil)
				mr.On("AssignRole", mock.Anything, mock.MatchedBy(func(ur *model.UserRole) bool {
					return ur.UserID == req.UserID && ur.CourseID == courseID && ur.RoleID == req.RoleID
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   true,
		},
		{
			name:     "Invalid Course ID",
			courseID: "invalid-uuid",
			reqBody: AssignRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			setupMocks:     func(mr *MockRoleRepo, mu *MockUserRepo, req AssignRoleRequest, courseID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   false,
		},
		{
			name:     "Missing User ID",
			courseID: uuid.New().String(),
			reqBody: AssignRoleRequest{
				UserID: uuid.Nil,
				RoleID: uuid.New(),
			},
			setupMocks:     func(mr *MockRoleRepo, mu *MockUserRepo, req AssignRoleRequest, courseID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   false,
		},
		{
			name:     "Missing Role ID",
			courseID: uuid.New().String(),
			reqBody: AssignRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.Nil,
			},
			setupMocks:     func(mr *MockRoleRepo, mu *MockUserRepo, req AssignRoleRequest, courseID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   false,
		},
		{
			name:     "User Not Found",
			courseID: uuid.New().String(),
			reqBody: AssignRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			setupMocks: func(mr *MockRoleRepo, mu *MockUserRepo, req AssignRoleRequest, courseID uuid.UUID) {
				mu.On("GetByID", mock.Anything, req.UserID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   false,
		},
		{
			name:     "Assign Role Fails",
			courseID: uuid.New().String(),
			reqBody: AssignRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			setupMocks: func(mr *MockRoleRepo, mu *MockUserRepo, req AssignRoleRequest, courseID uuid.UUID) {
				mu.On("GetByID", mock.Anything, req.UserID).Return(&model.User{ID: req.UserID}, nil)
				mr.On("AssignRole", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockRoleRepo, mockUserRepo, handler := setupTest(t)

			courseUUID, _ := uuid.Parse(tt.courseID)
			tt.setupMocks(mockRoleRepo, mockUserRepo, tt.reqBody, courseUUID)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/admin/courses/"+tt.courseID+"/roles", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("courseId")
			c.SetParamValues(tt.courseID)

			err := handler.AdminAssignRoleHandler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody {
				var response model.UserRole
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.reqBody.UserID, response.UserID)
				assert.Equal(t, tt.reqBody.RoleID, response.RoleID)
				if courseUUID != uuid.Nil {
					assert.Equal(t, courseUUID, response.CourseID)
				}
			}

			mockRoleRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAdminRevokeRoleHandler(t *testing.T) {
	tests := []struct {
		name           string
		courseID       string
		reqBody        RevokeRoleRequest
		setupMocks     func(*MockRoleRepo, RevokeRoleRequest, uuid.UUID)
		expectedStatus int
	}{
		{
			name:     "Success",
			courseID: uuid.New().String(),
			reqBody: RevokeRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			setupMocks: func(mr *MockRoleRepo, req RevokeRoleRequest, courseID uuid.UUID) {
				mr.On("RevokeRole", mock.Anything, req.UserID, courseID, req.RoleID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:     "Invalid Course ID",
			courseID: "invalid-uuid",
			reqBody: RevokeRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			setupMocks:     func(mr *MockRoleRepo, req RevokeRoleRequest, courseID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Missing User ID",
			courseID: uuid.New().String(),
			reqBody: RevokeRoleRequest{
				UserID: uuid.Nil,
				RoleID: uuid.New(),
			},
			setupMocks:     func(mr *MockRoleRepo, req RevokeRoleRequest, courseID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Missing Role ID",
			courseID: uuid.New().String(),
			reqBody: RevokeRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.Nil,
			},
			setupMocks:     func(mr *MockRoleRepo, req RevokeRoleRequest, courseID uuid.UUID) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Revoke Role Fails",
			courseID: uuid.New().String(),
			reqBody: RevokeRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			setupMocks: func(mr *MockRoleRepo, req RevokeRoleRequest, courseID uuid.UUID) {
				mr.On("RevokeRole", mock.Anything, req.UserID, courseID, req.RoleID).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockRoleRepo, _, handler := setupTest(t)

			courseUUID, _ := uuid.Parse(tt.courseID)
			tt.setupMocks(mockRoleRepo, tt.reqBody, courseUUID)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodDelete, "/admin/courses/"+tt.courseID+"/roles", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("courseId")
			c.SetParamValues(tt.courseID)

			err := handler.AdminRevokeRoleHandler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockRoleRepo.AssertExpectations(t)
		})
	}
}

func TestAdminListUserRolesHandler(t *testing.T) {
	tests := []struct {
		name           string
		courseID       string
		setupMocks     func(*MockRoleRepo, uuid.UUID)
		expectedStatus int
		expectedLen    int
	}{
		{
			name:     "Success",
			courseID: uuid.New().String(),
			setupMocks: func(mr *MockRoleRepo, courseID uuid.UUID) {
				mr.On("GetByCourseID", mock.Anything, courseID).Return([]model.UserRole{
					{UserID: uuid.New(), CourseID: courseID, RoleID: uuid.New()},
					{UserID: uuid.New(), CourseID: courseID, RoleID: uuid.New()},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:     "Invalid Course ID",
			courseID: "invalid-uuid",
			setupMocks: func(mr *MockRoleRepo, courseID uuid.UUID) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
		},
		{
			name:     "Empty List",
			courseID: uuid.New().String(),
			setupMocks: func(mr *MockRoleRepo, courseID uuid.UUID) {
				mr.On("GetByCourseID", mock.Anything, courseID).Return([]model.UserRole{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name:     "Database Error",
			courseID: uuid.New().String(),
			setupMocks: func(mr *MockRoleRepo, courseID uuid.UUID) {
				mr.On("GetByCourseID", mock.Anything, courseID).Return([]model.UserRole{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockRoleRepo, _, handler := setupTest(t)

			courseUUID, _ := uuid.Parse(tt.courseID)
			tt.setupMocks(mockRoleRepo, courseUUID)

			req := httptest.NewRequest(http.MethodGet, "/admin/courses/"+tt.courseID+"/roles", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("courseId")
			c.SetParamValues(tt.courseID)

			err := handler.AdminListUserRolesHandler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var roles []model.UserRole
				err := json.Unmarshal(rec.Body.Bytes(), &roles)
				assert.NoError(t, err)
				assert.Len(t, roles, tt.expectedLen)
			}

			mockRoleRepo.AssertExpectations(t)
		})
	}
}

func TestAdminAddPermissionHandler(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		reqBody        AddPermissionRequest
		setupMocks     func(*MockRoleRepo, uuid.UUID, string)
		expectedStatus int
	}{
		{
			name:   "Success",
			roleID: uuid.New().String(),
			reqBody: AddPermissionRequest{
				Permission: "edit_course",
			},
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {
				mr.On("AddPermission", mock.Anything, mock.MatchedBy(func(p *model.CourseAdminPermission) bool {
					return p.RoleID == roleID && p.Permission == permission
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid Role ID",
			roleID: "invalid-uuid",
			reqBody: AddPermissionRequest{
				Permission: "edit_course",
			},
			setupMocks:     func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Empty Permission",
			roleID: uuid.New().String(),
			reqBody: AddPermissionRequest{
				Permission: "",
			},
			setupMocks:     func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Add Permission Fails",
			roleID: uuid.New().String(),
			reqBody: AddPermissionRequest{
				Permission: "edit_course",
			},
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {
				mr.On("AddPermission", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockRoleRepo, _, handler := setupTest(t)

			roleUUID, _ := uuid.Parse(tt.roleID)
			tt.setupMocks(mockRoleRepo, roleUUID, tt.reqBody.Permission)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/admin/roles/"+tt.roleID+"/permissions", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("roleId")
			c.SetParamValues(tt.roleID)

			err := handler.AdminAddPermissionHandler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusCreated {
				var perm model.CourseAdminPermission
				err := json.Unmarshal(rec.Body.Bytes(), &perm)
				assert.NoError(t, err)
				assert.Equal(t, tt.reqBody.Permission, perm.Permission)
				if roleUUID != uuid.Nil {
					assert.Equal(t, roleUUID, perm.RoleID)
				}
			}

			mockRoleRepo.AssertExpectations(t)
		})
	}
}

func TestAdminRemovePermissionHandler(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		permission     string
		setupMocks     func(*MockRoleRepo, uuid.UUID, string)
		expectedStatus int
	}{
		{
			name:       "Success",
			roleID:     uuid.New().String(),
			permission: "edit_course",
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {
				mr.On("RemovePermission", mock.Anything, roleID, permission).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "Invalid Role ID",
			roleID:     "invalid-uuid",
			permission: "edit_course",
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty Permission",
			roleID:     uuid.New().String(),
			permission: "",
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Remove Permission Fails",
			roleID:     uuid.New().String(),
			permission: "edit_course",
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID, permission string) {
				mr.On("RemovePermission", mock.Anything, roleID, permission).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockRoleRepo, _, handler := setupTest(t)

			roleUUID, _ := uuid.Parse(tt.roleID)
			tt.setupMocks(mockRoleRepo, roleUUID, tt.permission)

			req := httptest.NewRequest(http.MethodDelete, "/admin/roles/"+tt.roleID+"/permissions/"+tt.permission, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("roleId", "permission")
			c.SetParamValues(tt.roleID, tt.permission)

			err := handler.AdminRemovePermissionHandler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			mockRoleRepo.AssertExpectations(t)
		})
	}
}

func TestAdminListPermissionsHandler(t *testing.T) {
	tests := []struct {
		name           string
		roleID         string
		setupMocks     func(*MockRoleRepo, uuid.UUID)
		expectedStatus int
		expectedLen    int
	}{
		{
			name:   "Success",
			roleID: uuid.New().String(),
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID) {
				mr.On("GetPermissions", mock.Anything, roleID).Return([]model.CourseAdminPermission{
					{RoleID: roleID, Permission: "edit_course"},
					{RoleID: roleID, Permission: "delete_course"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    2,
		},
		{
			name:   "Invalid Role ID",
			roleID: "invalid-uuid",
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedLen:    0,
		},
		{
			name:   "Empty List",
			roleID: uuid.New().String(),
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID) {
				mr.On("GetPermissions", mock.Anything, roleID).Return([]model.CourseAdminPermission{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedLen:    0,
		},
		{
			name:   "Database Error",
			roleID: uuid.New().String(),
			setupMocks: func(mr *MockRoleRepo, roleID uuid.UUID) {
				mr.On("GetPermissions", mock.Anything, roleID).Return([]model.CourseAdminPermission{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockRoleRepo, _, handler := setupTest(t)

			roleUUID, _ := uuid.Parse(tt.roleID)
			tt.setupMocks(mockRoleRepo, roleUUID)

			req := httptest.NewRequest(http.MethodGet, "/admin/roles/"+tt.roleID+"/permissions", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("roleId")
			c.SetParamValues(tt.roleID)

			err := handler.AdminListPermissionsHandler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedStatus == http.StatusOK {
				var perms []model.CourseAdminPermission
				err := json.Unmarshal(rec.Body.Bytes(), &perms)
				assert.NoError(t, err)
				assert.Len(t, perms, tt.expectedLen)
			}

			mockRoleRepo.AssertExpectations(t)
		})
	}
}
