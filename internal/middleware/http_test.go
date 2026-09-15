package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestRegisterAddsRequestIDAndSafeLogging(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	e := echo.New()
	Register(e, logger)
	e.GET("/items/:id", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	const secret = "must-not-appear"
	request := httptest.NewRequest(http.MethodGet, "/items/123?token="+secret, strings.NewReader(secret))
	request.Header.Set("Authorization", "Bearer "+secret)
	request.Header.Set("Cookie", "session="+secret)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	if response.Header().Get(echo.HeaderXRequestID) == "" {
		t.Fatal("X-Request-ID response header is empty")
	}
	logLine := output.String()
	for _, want := range []string{"request_id", "GET", "/items/:id", `"status":200`} {
		if !strings.Contains(logLine, want) {
			t.Errorf("log %q does not contain %q", logLine, want)
		}
	}
	if strings.Contains(logLine, secret) || strings.Contains(logLine, "Authorization") || strings.Contains(logLine, "Cookie") {
		t.Fatalf("request log contains sensitive request data: %s", logLine)
	}
}

func TestRegisterRecoversFromPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	e := echo.New()
	Register(e, logger)
	e.GET("/panic", func(*echo.Context) error {
		panic("sensitive panic value")
	})

	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "sensitive panic value") {
		t.Fatalf("response leaked panic value: %s", response.Body.String())
	}
}
