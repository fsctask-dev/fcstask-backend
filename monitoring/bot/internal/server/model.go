package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app *fiber.App
}

type AlertmanagerPayload struct {
	Status string        `json:"status"`
	Alerts []AlertDetail `json:"alerts"`
}

type AlertDetail struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}
