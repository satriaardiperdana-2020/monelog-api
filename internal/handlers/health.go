package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
)

// HealthService defines the process and dependency health behavior exposed over HTTP.
type HealthService interface {
	Live() bool
	Ready(context.Context) bool
}

// Health translates health service results into stable public responses.
type Health struct {
	service HealthService
}

// NewHealth creates health HTTP handlers.
func NewHealth(service HealthService) *Health {
	return &Health{service: service}
}

// Live handles GET /health/live.
func (h *Health) Live(c *echo.Context) error {
	if !h.service.Live() {
		return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
	}
	return c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

// Ready handles GET /health/ready.
func (h *Health) Ready(c *echo.Context) error {
	if !h.service.Ready(c.Request().Context()) {
		return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
	}
	return c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

type healthResponse struct {
	Status string `json:"status"`
}
