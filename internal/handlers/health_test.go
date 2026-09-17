package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type fakeHealthService struct {
	live       bool
	ready      bool
	readyCalls int
}

func (f *fakeHealthService) Live() bool {
	return f.live
}

func (f *fakeHealthService) Ready(context.Context) bool {
	f.readyCalls++
	return f.ready
}

func TestHealthEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		service    *fakeHealthService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "liveness is independent of readiness",
			path:       "/health/live",
			service:    &fakeHealthService{live: true, ready: false},
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok"}`,
		},
		{
			name:       "database is ready",
			path:       "/health/ready",
			service:    &fakeHealthService{live: true, ready: true},
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok"}`,
		},
		{
			name:       "database is unavailable",
			path:       "/health/ready",
			service:    &fakeHealthService{live: true, ready: false},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"status":"unavailable"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			RegisterRoutes(e, NewHealth(test.service), nil, nil, nil)

			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if strings.TrimSpace(response.Body.String()) != test.wantBody {
				t.Errorf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
			if test.path == "/health/live" && test.service.readyCalls != 0 {
				t.Errorf("liveness called readiness %d times", test.service.readyCalls)
			}
		})
	}
}

func TestUnknownDomainRouteIsNotImplemented(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, NewHealth(&fakeHealthService{live: true, ready: true}), nil, nil, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
