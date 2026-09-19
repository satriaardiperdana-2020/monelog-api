package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	appauth "github.com/satriaardiperdana-2020/monelog-api/internal/auth"
	"github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

const (
	refreshCookieName = "__Host-monelog-refresh"
	csrfCookieName    = "__Host-monelog-csrf"
	refreshLifetime   = 30 * 24 * time.Hour
)

type authService interface {
	Register(context.Context, service.RegisterInput) (service.User, error)
	Login(context.Context, service.LoginInput) (service.Session, error)
	Refresh(context.Context, string) (service.Session, error)
	Logout(context.Context, string) error
	Me(context.Context, service.Actor) (service.User, error)
	UpdateProfile(context.Context, service.Actor, string, int32) (service.User, error)
	DeleteMe(context.Context, service.Actor, int32) error
}

// Auth exposes authentication and self-profile HTTP routes.
type Auth struct {
	service        authService
	allowedOrigins map[string]struct{}
	limiter        *middleware.LoginRateLimiter
}

func NewAuth(authService authService, allowedOrigins []string, limiter *middleware.LoginRateLimiter) *Auth {
	origins := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origins[origin] = struct{}{}
	}
	return &Auth{service: authService, allowedOrigins: origins, limiter: limiter}
}

func (h *Auth) Register(c *echo.Context) error {
	var request registerRequest
	if err := decodeJSON(c, &request); err != nil {
		return validationError(c)
	}
	user, err := h.service.Register(c.Request().Context(), service.RegisterInput{Email: request.Email, Password: request.Password, Timezone: request.Timezone})
	if err != nil {
		return h.writeServiceError(c, err)
	}
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	return c.JSON(http.StatusCreated, map[string]any{"data": userResponse(user)})
}

func (h *Auth) Login(c *echo.Context) error {
	var request loginRequest
	if err := decodeJSON(c, &request); err != nil || (request.ClientType != "web" && request.ClientType != "native") {
		return validationError(c)
	}
	if !h.limiter.Allow(strings.ToLower(strings.TrimSpace(request.Email)), c.Request().RemoteAddr) {
		c.Response().Header().Set("Retry-After", strconv.Itoa(int((15 * time.Minute).Seconds())))
		return c.JSON(http.StatusTooManyRequests, errorBody("RATE_LIMITED", "Too many login attempts."))
	}
	if request.ClientType == "web" && !h.allowedOrigin(c.Request()) {
		return c.JSON(http.StatusForbidden, errorBody("FORBIDDEN", "Forbidden."))
	}
	session, err := h.service.Login(c.Request().Context(), service.LoginInput{Email: request.Email, Password: request.Password})
	if err != nil {
		return authenticationError(c)
	}
	return h.writeSession(c, request.ClientType, session)
}

func (h *Auth) Refresh(c *echo.Context) error {
	var request refreshRequest
	if err := decodeJSON(c, &request); err != nil || (request.ClientType != "web" && request.ClientType != "native") {
		return validationError(c)
	}
	if request.ClientType == "web" && !h.allowedOrigin(c.Request()) {
		return forbiddenError(c)
	}
	refreshToken, ok := h.transportToken(c, request.ClientType, request.RefreshToken, true)
	if !ok {
		if request.ClientType == "web" {
			return forbiddenError(c)
		}
		return authenticationError(c)
	}
	session, err := h.service.Refresh(c.Request().Context(), refreshToken)
	if err != nil {
		return authenticationError(c)
	}
	return h.writeSession(c, request.ClientType, session)
}

func (h *Auth) Logout(c *echo.Context) error {
	var request refreshRequest
	if err := decodeJSON(c, &request); err != nil || (request.ClientType != "web" && request.ClientType != "native") {
		return validationError(c)
	}
	if request.ClientType == "web" && !h.allowedOrigin(c.Request()) {
		return forbiddenError(c)
	}
	refreshToken, ok := h.transportToken(c, request.ClientType, request.RefreshToken, true)
	if !ok && request.ClientType == "web" {
		return forbiddenError(c)
	}
	if ok {
		if err := h.service.Logout(c.Request().Context(), refreshToken); err != nil {
			return h.writeServiceError(c, err)
		}
	}
	if request.ClientType == "web" {
		h.clearBrowserCookies(c)
	}
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	return c.NoContent(http.StatusNoContent)
}

func (h *Auth) Me(c *echo.Context) error {
	actor, ok := middleware.Actor(c.Request().Context())
	if !ok {
		return authenticationError(c)
	}
	user, err := h.service.Me(c.Request().Context(), actor)
	if err != nil {
		return authenticationError(c)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": userResponse(user)})
}

func (h *Auth) UpdateProfile(c *echo.Context) error {
	actor, ok := middleware.Actor(c.Request().Context())
	if !ok {
		return authenticationError(c)
	}
	var request profileRequest
	if err := decodeJSON(c, &request); err != nil {
		return validationError(c)
	}
	user, err := h.service.UpdateProfile(c.Request().Context(), actor, request.Timezone, request.Version)
	if err != nil {
		return h.writeServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": userResponse(user)})
}

func (h *Auth) DeleteMe(c *echo.Context) error {
	actor, ok := middleware.Actor(c.Request().Context())
	if !ok {
		return authenticationError(c)
	}
	version, err := ifMatchVersion(c.Request().Header.Get("If-Match"))
	if err != nil {
		return validationError(c)
	}
	if err := h.service.DeleteMe(c.Request().Context(), actor, version); err != nil {
		return h.writeServiceError(c, err)
	}
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	return c.NoContent(http.StatusNoContent)
}

func (h *Auth) transportToken(c *echo.Context, clientType, nativeToken string, requireCSRF bool) (string, bool) {
	if clientType == "native" {
		return nativeToken, nativeToken != ""
	}
	if !h.allowedOrigin(c.Request()) {
		return "", false
	}
	refreshCookie, err := c.Cookie(refreshCookieName)
	if err != nil || refreshCookie.Value == "" {
		return "", false
	}
	if requireCSRF {
		csrfCookie, err := c.Cookie(csrfCookieName)
		if err != nil || !appauth.EqualSecret(csrfCookie.Value, c.Request().Header.Get("X-CSRF-Token")) {
			return "", false
		}
	}
	return refreshCookie.Value, true
}

func (h *Auth) allowedOrigin(request *http.Request) bool {
	_, ok := h.allowedOrigins[request.Header.Get("Origin")]
	return ok
}

func (h *Auth) writeSession(c *echo.Context, clientType string, session service.Session) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	if clientType == "web" {
		csrf, err := appauth.NewCSRFToken()
		if err != nil {
			return internalError(c)
		}
		h.setBrowserCookies(c, session.RefreshToken, csrf)
		return c.JSON(http.StatusOK, map[string]any{"access_token": session.AccessToken, "token_type": "Bearer", "expires_in": int64(time.Until(session.ExpiresAt).Seconds())})
	}
	return c.JSON(http.StatusOK, map[string]any{"access_token": session.AccessToken, "refresh_token": session.RefreshToken, "token_type": "Bearer", "expires_in": int64(time.Until(session.ExpiresAt).Seconds())})
}

func (h *Auth) setBrowserCookies(c *echo.Context, refreshToken, csrf string) {
	expires := time.Now().Add(refreshLifetime)
	c.SetCookie(&http.Cookie{Name: refreshCookieName, Value: refreshToken, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: int(refreshLifetime.Seconds())})
	c.SetCookie(&http.Cookie{Name: csrfCookieName, Value: csrf, Path: "/", HttpOnly: false, Secure: true, SameSite: http.SameSiteStrictMode, Expires: expires, MaxAge: int(refreshLifetime.Seconds())})
}

func (h *Auth) clearBrowserCookies(c *echo.Context) {
	for _, name := range []string{refreshCookieName, csrfCookieName} {
		c.SetCookie(&http.Cookie{Name: name, Value: "", Path: "/", HttpOnly: name == refreshCookieName, Secure: true, SameSite: http.SameSiteStrictMode, Expires: time.Unix(1, 0), MaxAge: -1})
	}
}

func (h *Auth) writeServiceError(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrAuthentication):
		return authenticationError(c)
	case errors.Is(err, service.ErrValidation):
		return validationError(c)
	case errors.Is(err, service.ErrConflict):
		return c.JSON(http.StatusConflict, errorBody("CONFLICT", "Conflict."))
	case errors.Is(err, service.ErrNotFound):
		return c.JSON(http.StatusNotFound, errorBody("NOT_FOUND", "Not found."))
	default:
		return internalError(c)
	}
}

func decodeJSON(c *echo.Context, destination any) error {
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values")
	}
	return nil
}

func ifMatchVersion(value string) (int32, error) {
	value = strings.Trim(value, "\"")
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed <= 0 {
		return 0, errors.New("invalid version")
	}
	return int32(parsed), nil
}

func authenticationError(c *echo.Context) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	return c.JSON(http.StatusUnauthorized, errorBody("AUTHENTICATION_FAILED", "Authentication failed."))
}

func validationError(c *echo.Context) error {
	return c.JSON(http.StatusUnprocessableEntity, errorBody("VALIDATION_ERROR", "Invalid request."))
}

func forbiddenError(c *echo.Context) error {
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
	return c.JSON(http.StatusForbidden, errorBody("FORBIDDEN", "Forbidden."))
}

func internalError(c *echo.Context) error {
	return c.JSON(http.StatusInternalServerError, errorBody("INTERNAL_ERROR", "Internal server error."))
}

func errorBody(code, message string) map[string]any {
	return map[string]any{"error": map[string]string{"code": code, "message": message}}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

type loginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	ClientType string `json:"client_type"`
}

type refreshRequest struct {
	ClientType   string `json:"client_type"`
	RefreshToken string `json:"refresh_token"`
}

type profileRequest struct {
	Timezone string `json:"timezone"`
	Version  int32  `json:"version"`
}

type userPayload struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Timezone string `json:"timezone"`
	Currency string `json:"currency"`
	IsDelete bool   `json:"isDelete"`
	Version  int32  `json:"version"`
}

func userResponse(user service.User) userPayload {
	return userPayload{ID: user.ID, Email: user.Email, Role: user.Role, Timezone: user.Timezone, Currency: user.Currency, IsDelete: user.IsDelete, Version: user.Version}
}
