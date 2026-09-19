package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type fakeAuthenticator struct {
	actor service.Actor
	err   error
}

func (f fakeAuthenticator) AuthenticateAccess(context.Context, string) (service.Actor, error) {
	return f.actor, f.err
}

func TestAuthenticateLoadsActorAndNoStore(t *testing.T) {
	e := echo.New()
	actor := service.Actor{UserID: 1, Role: service.RoleAdmin}
	e.GET("/protected", func(c *echo.Context) error {
		got, ok := Actor(c.Request().Context())
		if !ok || got != actor {
			t.Fatal("actor was not added to the request context")
		}
		return c.NoContent(http.StatusNoContent)
	}, Authenticate(fakeAuthenticator{actor: actor}))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set(echo.HeaderAuthorization, "Bearer token")
	e.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || response.Header().Get(echo.HeaderCacheControl) != "no-store" {
		t.Fatalf("status=%d headers=%v", response.Code, response.Header())
	}
}

func TestAuthenticateRejectsMalformedBearer(t *testing.T) {
	e := echo.New()
	e.GET("/protected", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) }, Authenticate(fakeAuthenticator{}))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set(echo.HeaderAuthorization, "Basic token")
	e.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || response.Header().Get(echo.HeaderCacheControl) != "no-store" {
		t.Fatalf("status=%d headers=%v", response.Code, response.Header())
	}
}

func TestAuthenticateReturnsUnauthorizedForMissingOrInvalidTokens(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		authenticator fakeAuthenticator
	}{
		{name: "missing token"},
		{name: "expired or invalid token", authorization: "Bearer token", authenticator: fakeAuthenticator{err: service.ErrAuthentication}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			e.GET("/protected", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) }, Authenticate(test.authenticator))
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if test.authorization != "" {
				request.Header.Set(echo.HeaderAuthorization, test.authorization)
			}
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized || response.Header().Get(echo.HeaderCacheControl) != "no-store" || response.Body.String() != "{\"error\":{\"code\":\"AUTHENTICATION_FAILED\",\"message\":\"Authentication failed.\"}}\n" {
				t.Fatalf("status=%d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
			}
		})
	}
}

func TestRequireAdminRejectsUser(t *testing.T) {
	e := echo.New()
	e.GET("/admin", RequireAdmin(func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) }))
	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want 403", response.Code)
	}
}
