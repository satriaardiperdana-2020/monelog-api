package handlers

import (
	"github.com/labstack/echo/v5"

	api "github.com/satriaardiperdana-2020/monelog-api/internal/api"
)

// RegisterRoutes registers the OpenAPI-defined HTTP routes.
func RegisterRoutes(e *echo.Echo, health *Health, auth *Auth, authenticate echo.MiddlewareFunc) {
	RegisterSwagger(e)
	if auth == nil || authenticate == nil {
		e.GET("/health/live", health.Live)
		e.GET("/health/ready", health.Ready)
		return
	}
	api.RegisterHandlersWithOptions(e, newOpenAPIServer(health, auth), api.RegisterHandlersOptions{
		OperationMiddlewares: map[string][]echo.MiddlewareFunc{
			"deleteMe": {authenticate},
			"getMe":    {authenticate},
			"updateMe": {authenticate},
		},
	})
}
