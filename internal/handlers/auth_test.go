package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

type fakeAuthService struct {
	registerInput service.RegisterInput
	loginInput    service.LoginInput
	loginErr      error
	refreshErr    error
	logoutToken   string
}

func (f *fakeAuthService) Register(_ context.Context, input service.RegisterInput) (service.User, error) {
	f.registerInput = input
	return testUser(), nil
}

func (f *fakeAuthService) Login(_ context.Context, input service.LoginInput) (service.Session, error) {
	f.loginInput = input
	if f.loginErr != nil {
		return service.Session{}, f.loginErr
	}
	return testSession(), nil
}

func (f *fakeAuthService) Refresh(context.Context, string) (service.Session, error) {
	if f.refreshErr != nil {
		return service.Session{}, f.refreshErr
	}
	return testSession(), nil
}

func (f *fakeAuthService) Logout(_ context.Context, token string) error {
	f.logoutToken = token
	return nil
}
func (f *fakeAuthService) Me(context.Context, service.Actor) (service.User, error) {
	return testUser(), nil
}
func (f *fakeAuthService) UpdateProfile(context.Context, service.Actor, string, int32) (service.User, error) {
	return testUser(), nil
}
func (f *fakeAuthService) DeleteMe(context.Context, service.Actor, int32) error { return nil }

func TestRegistrationRejectsRoleInjection(t *testing.T) {
	fake := &fakeAuthService{}
	handler := NewAuth(fake, []string{"https://app.example.com"}, middleware.NewLoginRateLimiter())
	e := echo.New()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"person@example.com","password":"correct horse battery staple","timezone":"Asia/Jakarta","role":"admin"}`))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	response := httptest.NewRecorder()
	e.NewContext(request, response)
	if err := handler.Register(e.NewContext(request, response)); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusUnprocessableEntity || fake.registerInput.Email != "" {
		t.Fatalf("role injection status=%d input=%+v", response.Code, fake.registerInput)
	}
}

func TestBrowserAndNativeTokenTransport(t *testing.T) {
	fake := &fakeAuthService{}
	handler := NewAuth(fake, []string{"https://app.example.com"}, middleware.NewLoginRateLimiter())
	e := echo.New()

	browserRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"person@example.com","password":"correct horse battery staple","client_type":"web"}`))
	browserRequest.Header.Set("Origin", "https://app.example.com")
	browserResponse := httptest.NewRecorder()
	if err := handler.Login(e.NewContext(browserRequest, browserResponse)); err != nil {
		t.Fatal(err)
	}
	if browserResponse.Code != http.StatusOK || !strings.Contains(browserResponse.Header().Get("Set-Cookie"), refreshCookieName) || !strings.Contains(browserResponse.Header().Get("Set-Cookie"), "HttpOnly") || !strings.Contains(browserResponse.Header().Get("Set-Cookie"), "Secure") {
		t.Fatalf("browser response did not set secure refresh cookie: status=%d headers=%v", browserResponse.Code, browserResponse.Header())
	}
	if strings.Contains(browserResponse.Body.String(), "refresh-token") || browserResponse.Header().Get(echo.HeaderCacheControl) != "no-store" {
		t.Fatalf("browser response leaked a refresh token or was cacheable: %s", browserResponse.Body.String())
	}

	nativeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"native@example.com","password":"correct horse battery staple","client_type":"native"}`))
	nativeResponse := httptest.NewRecorder()
	if err := handler.Login(e.NewContext(nativeRequest, nativeResponse)); err != nil {
		t.Fatal(err)
	}
	if nativeResponse.Code != http.StatusOK || !strings.Contains(nativeResponse.Body.String(), "refresh_token") || nativeResponse.Header().Get("Set-Cookie") != "" {
		t.Fatalf("native response used the wrong transport: status=%d body=%s headers=%v", nativeResponse.Code, nativeResponse.Body.String(), nativeResponse.Header())
	}
}

func TestBrowserRefreshRequiresOriginAndCSRF(t *testing.T) {
	handler := NewAuth(&fakeAuthService{}, []string{"https://app.example.com"}, middleware.NewLoginRateLimiter())
	e := echo.New()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(`{"client_type":"web"}`))
	request.Header.Set("Origin", "https://app.example.com")
	request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "refresh-token"})
	request.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-token"})
	response := httptest.NewRecorder()
	if err := handler.Refresh(e.NewContext(request, response)); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusForbidden {
		t.Fatalf("refresh without CSRF status=%d, want 403", response.Code)
	}
}

func TestLoginFailureIsGeneric(t *testing.T) {
	handler := NewAuth(&fakeAuthService{loginErr: service.ErrAuthentication}, []string{"https://app.example.com"}, middleware.NewLoginRateLimiter())
	e := echo.New()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"person@example.com","password":"correct horse battery staple","client_type":"native"}`))
	response := httptest.NewRecorder()
	if err := handler.Login(e.NewContext(request, response)); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusUnauthorized || strings.Contains(response.Body.String(), "person@example.com") {
		t.Fatalf("non-generic login failure: %d %s", response.Code, response.Body.String())
	}
}

func testUser() service.User {
	return service.User{ID: 1, Email: "person@example.com", Role: service.RoleUser, Timezone: "Asia/Jakarta", Currency: "IDR", Version: 1}
}

func testSession() service.Session {
	return service.Session{AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresAt: time.Now().Add(24 * time.Hour)}
}

func TestUserResponseDoesNotContainPasswordFields(t *testing.T) {
	payload, err := json.Marshal(userResponse(testUser()))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"password", "token_hash", "refresh"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("user payload leaked %q: %s", forbidden, payload)
		}
	}
}
