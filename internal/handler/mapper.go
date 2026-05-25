package handler

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"fcstask-backend/internal/api"
	models "fcstask-backend/internal/db/model"
	"fcstask-backend/internal/service"
)

type authResponse struct {
	SessionToken openapi_types.UUID `json:"session_token"`
	User         api.User           `json:"user"`
}

type sessionResponse struct {
	Id        openapi_types.UUID `json:"id"`
	Ip        string             `json:"ip"`
	UserAgent string             `json:"user_agent"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type sessionWithUserResponse struct {
	sessionResponse
	User api.User `json:"user"`
}

type userWithSessionsResponse struct {
	User     api.User          `json:"user"`
	Sessions []sessionResponse `json:"sessions"`
}

type paginatedSessionsResponse struct {
	Items  []sessionWithUserResponse `json:"items"`
	Total  int64                     `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

type paginatedUsersWithSessionsResponse struct {
	Items  []userWithSessionsResponse `json:"items"`
	Total  int64                      `json:"total"`
	Limit  int                        `json:"limit"`
	Offset int                        `json:"offset"`
}

func userToAPI(user *models.User) api.User {
	return api.User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		TgUid:     user.TgUID,
		UserId:    openapi_types.UUID(user.UserID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func sessionToAPI(session models.Session) sessionResponse {
	return sessionResponse{
		Id:        openapi_types.UUID(session.ID),
		Ip:        session.IP,
		UserAgent: session.UserAgent,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}
}

func sessionResultsToAPI(sessions []models.Session) []sessionWithUserResponse {
	items := make([]sessionWithUserResponse, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, sessionWithUserResponse{
			sessionResponse: sessionToAPI(session),
			User:            userToAPI(&session.User),
		})
	}

	return items
}

func userSessionsResultsToAPI(users []models.User) []userWithSessionsResponse {
	items := make([]userWithSessionsResponse, 0, len(users))
	for _, user := range users {
		items = append(items, userWithSessionsResponse{
			User:     userToAPI(&user),
			Sessions: sessionsToAPI(user.Sessions),
		})
	}

	return items
}

func sessionsToAPI(sessions []models.Session) []sessionResponse {
	items := make([]sessionResponse, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, sessionToAPI(session))
	}

	return items
}

func authResultToAPI(result *service.AuthResult) authResponse {
	return authResponse{
		SessionToken: openapi_types.UUID(result.Session.ID),
		User:         userToAPI(result.User),
	}
}
