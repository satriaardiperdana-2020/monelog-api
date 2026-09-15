package handlers

import "github.com/labstack/echo/v5"

// RegisterRoutes registers the public routes implemented by ISSUE-001.
func RegisterRoutes(e *echo.Echo, health *Health) {
	e.GET("/health/live", health.Live)
	e.GET("/health/ready", health.Ready)
}
