package middleware

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

// Register installs the HTTP middleware required by the service foundation.
func Register(e *echo.Echo, logger *slog.Logger) {
	e.Use(echomiddleware.RequestID())
	e.Use(requestLogger(logger))
	e.Use(echomiddleware.RecoverWithConfig(echomiddleware.RecoverConfig{
		DisablePrintStack: true,
		DisableStackAll:   true,
	}))
}

// RegisterCORS permits credentialed browser requests only from configured origins.
// Auth handlers still enforce their own Origin and CSRF checks.
func RegisterCORS(e *echo.Echo, allowedOrigins []string) {
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowCredentials: true,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-CSRF-Token", "If-Match"},
	}))
}

func requestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		HandleError:  true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogRoutePath: true,
		LogRequestID: true,
		LogStatus:    true,
		LogValuesFunc: func(c *echo.Context, values echomiddleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			if values.Status >= 500 {
				level = slog.LevelError
			}
			logger.LogAttrs(
				c.Request().Context(),
				level,
				"http request",
				slog.String("request_id", values.RequestID),
				slog.String("method", values.Method),
				slog.String("route", values.RoutePath),
				slog.Int("status", values.Status),
				slog.Duration("duration", values.Latency),
				slog.String("remote_ip", values.RemoteIP),
			)
			return nil
		},
	})
}
