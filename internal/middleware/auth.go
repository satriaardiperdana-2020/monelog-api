package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type accessAuthenticator interface {
	AuthenticateAccess(context.Context, string) (service.Actor, error)
}

type actorContextKey struct{}

// Authenticate validates a Bearer token and loads the current actor state.
func Authenticate(authenticator accessAuthenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			raw, ok := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if !ok {
				return unauthorized(c)
			}
			actor, err := authenticator.AuthenticateAccess(c.Request().Context(), raw)
			if err != nil {
				return unauthorized(c)
			}
			request := c.Request().WithContext(context.WithValue(c.Request().Context(), actorContextKey{}, actor))
			c.SetRequest(request)
			c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
			return next(c)
		}
	}
}

// RequireAdmin rejects current actors that are not loaded as administrators.
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		actor, ok := Actor(c.Request().Context())
		if !ok || actor.Role != service.RoleAdmin {
			return c.JSON(http.StatusForbidden, map[string]any{"error": map[string]string{"code": "FORBIDDEN", "message": "Forbidden."}})
		}
		return next(c)
	}
}

// Actor returns the database-backed actor inserted by Authenticate.
func Actor(ctx context.Context) (service.Actor, bool) {
	actor, ok := ctx.Value(actorContextKey{}).(service.Actor)
	return actor, ok && actor.UserID != uuid.Nil
}

func bearerToken(value string) (string, bool) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func unauthorized(c *echo.Context) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	return c.JSON(http.StatusUnauthorized, map[string]any{"error": map[string]string{"code": "AUTHENTICATION_FAILED", "message": "Authentication failed."}})
}
