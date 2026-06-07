package handler

import (
	"context"
	"fcstask-backend/internal/db/model"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type StatsServiceInterface interface {
	GetStats(ctx context.Context, userID uuid.UUID) (*model.PlatformStats, error)
}

type StatsHandler struct {
	statsService StatsServiceInterface
}

func NewStatsHandler(statsService StatsServiceInterface) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetStats(c echo.Context) error {
	user, ok := c.Get(UserContextKey).(*model.User)
	if !ok || user == nil {
		return unauthorized(c, "User not found")
	}

	stats, err := h.statsService.GetStats(c.Request().Context(), user.ID)
	if err != nil {
		return serviceError(c, err)
	}
	return c.JSON(http.StatusOK, stats)
}
