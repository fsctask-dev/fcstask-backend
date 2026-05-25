package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"fcstask-backend/internal/db/model"
)

func newHelperContext(method, target string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec), rec
}

func TestAuthenticatedUser_Success(t *testing.T) {
	ctx, _ := newHelperContext(http.MethodGet, "/", "")
	expected := &model.User{ID: uuid.New()}
	ctx.Set(UserContextKey, expected)

	user, ok := authenticatedUser(ctx)

	assert.Same(t, expected, user)
	assert.True(t, ok)
}

func TestAuthenticatedUser_MissingUser(t *testing.T) {
	ctx, _ := newHelperContext(http.MethodGet, "/", "")

	user, ok := authenticatedUser(ctx)

	assert.Nil(t, user)
	assert.False(t, ok)
}

func TestMustAuthenticatedUser_PanicsWithoutUser(t *testing.T) {
	ctx, _ := newHelperContext(http.MethodGet, "/", "")

	assert.Panics(t, func() {
		_ = mustAuthenticatedUser(ctx)
	})
}

func TestParseUUIDParam_InvalidValue(t *testing.T) {
	ctx, rec := newHelperContext(http.MethodGet, "/", "")
	ctx.SetParamNames("courseId")
	ctx.SetParamValues("bad")

	value, ok := parseUUIDParam(ctx, "courseId", "Invalid course ID")

	assert.Equal(t, uuid.Nil, value)
	assert.False(t, ok)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBindRequest_InvalidPayload(t *testing.T) {
	ctx, rec := newHelperContext(http.MethodPost, "/", "{")
	ctx.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	var payload struct {
		Name string `json:"name"`
	}

	ok := bindRequest(ctx, &payload, "Invalid request body")

	assert.False(t, ok)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
