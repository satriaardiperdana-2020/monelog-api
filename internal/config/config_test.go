package config

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoadFileEnvironmentUsesEnvironmentSpecificFileAndProcessOverrides(t *testing.T) {
	process := map[string]string{
		"APP_ENV":      "development",
		"DATABASE_URL": "postgres://process:password@localhost:5432/monelog",
	}
	lookup, err := loadFileEnvironment(mapLookup(process), func(name string) ([]byte, error) {
		if name != ".env.development" {
			t.Fatalf("environment file = %q, want .env.development", name)
		}
		return []byte("DATABASE_URL=postgres://file:password@localhost:5432/monelog\nAUTH_JWT_ISSUER='monelog api'\nexport AUTH_JWT_AUDIENCE=monelog-app\n"), nil
	})
	if err != nil {
		t.Fatalf("loadFileEnvironment() error = %v", err)
	}
	if got, _ := lookup("DATABASE_URL"); got != process["DATABASE_URL"] {
		t.Fatalf("DATABASE_URL = %q, want process value", got)
	}
	if got, _ := lookup("AUTH_JWT_ISSUER"); got != "monelog api" {
		t.Fatalf("AUTH_JWT_ISSUER = %q", got)
	}
	if got, _ := lookup("AUTH_JWT_AUDIENCE"); got != "monelog-app" {
		t.Fatalf("AUTH_JWT_AUDIENCE = %q", got)
	}
}

func TestLoadFileEnvironmentRejectsMalformedOrUnsafeFiles(t *testing.T) {
	tests := []struct {
		name    string
		process map[string]string
		read    func(string) ([]byte, error)
	}{
		{name: "unsafe app environment", process: map[string]string{"APP_ENV": "../production"}, read: func(string) ([]byte, error) { return nil, nil }},
		{name: "invalid assignment", process: map[string]string{}, read: func(string) ([]byte, error) { return []byte("NOT VALID"), nil }},
		{name: "missing file", process: map[string]string{}, read: func(string) ([]byte, error) { return nil, os.ErrNotExist }},
		{name: "read failure", process: map[string]string{}, read: func(string) ([]byte, error) { return nil, errors.New("unavailable") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadFileEnvironment(mapLookup(test.process), test.read)
			if test.name == "missing file" {
				if err != nil {
					t.Fatalf("loadFileEnvironment() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("loadFileEnvironment() error = nil, want error")
			}
		})
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL":         "postgres://user:password@localhost:5432/monelog?sslmode=disable",
		"AUTH_JWT_HMAC_KEY":    "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"AUTH_JWT_ISSUER":      "monelog-api",
		"AUTH_JWT_AUDIENCE":    "monelog-app",
		"AUTH_ALLOWED_ORIGINS": "https://app.example.com",
	}

	got, err := load(mapLookup(values))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if got.AppEnv != defaultAppEnv {
		t.Errorf("AppEnv = %q, want %q", got.AppEnv, defaultAppEnv)
	}
	if got.HTTP.Addr != defaultHTTPAddr {
		t.Errorf("HTTP.Addr = %q, want %q", got.HTTP.Addr, defaultHTTPAddr)
	}
	if got.HTTP.ReadHeaderTimeout != defaultReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %s, want %s", got.HTTP.ReadHeaderTimeout, defaultReadHeaderTimeout)
	}
	if got.HTTP.ReadyTimeout != defaultReadyTimeout {
		t.Errorf("ReadyTimeout = %s, want %s", got.HTTP.ReadyTimeout, defaultReadyTimeout)
	}
	if got.Auth.AccessTokenLifetime != 24*time.Hour {
		t.Errorf("AccessTokenLifetime = %s, want 24 hours", got.Auth.AccessTokenLifetime)
	}
}

func TestLoadAcceptsOverrides(t *testing.T) {
	values := map[string]string{
		"APP_ENV":                  "test",
		"HTTP_ADDR":                "localhost:9090",
		"HTTP_READ_HEADER_TIMEOUT": "1s",
		"HTTP_READ_TIMEOUT":        "2s",
		"HTTP_WRITE_TIMEOUT":       "3s",
		"HTTP_IDLE_TIMEOUT":        "4s",
		"SHUTDOWN_TIMEOUT":         "5s",
		"HEALTH_READY_TIMEOUT":     "750ms",
		"DATABASE_URL":             "postgresql://user:password@localhost:5432/monelog",
		"AUTH_JWT_HMAC_KEY":        "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"AUTH_JWT_ISSUER":          "monelog-api",
		"AUTH_JWT_AUDIENCE":        "monelog-app",
		"AUTH_ACCESS_TOKEN_TTL":    "24h",
		"AUTH_ALLOWED_ORIGINS":     "https://app.example.com,http://localhost:5173",
	}

	got, err := load(mapLookup(values))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if got.AppEnv != "test" || got.HTTP.Addr != "localhost:9090" {
		t.Fatalf("unexpected string configuration: %+v", got)
	}
	if got.HTTP.ReadyTimeout != 750*time.Millisecond || got.HTTP.ShutdownTimeout != 5*time.Second || got.Auth.AccessTokenLifetime != 24*time.Hour {
		t.Fatalf("unexpected timeout configuration: %+v", got.HTTP)
	}
}

func TestLoadRejectsInvalidConfigurationWithoutLeakingDatabaseURL(t *testing.T) {
	const sensitiveURL = "mysql://private-user:private-password@db.internal/monelog"

	tests := []struct {
		name   string
		values map[string]string
	}{
		{name: "missing database URL", values: authValues(map[string]string{})},
		{name: "invalid database URL", values: authValues(map[string]string{"DATABASE_URL": sensitiveURL})},
		{name: "missing auth key", values: map[string]string{"DATABASE_URL": validDatabaseURL}},
		{name: "invalid auth key", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "AUTH_JWT_HMAC_KEY": "bad"})},
		{name: "invalid origin", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "AUTH_ALLOWED_ORIGINS": "https://app.example.com/path"})},
		{name: "access token lifetime too long", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "AUTH_ACCESS_TOKEN_TTL": "25h"})},
		{name: "empty app environment", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "APP_ENV": " "})},
		{name: "invalid address", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "HTTP_ADDR": "localhost"})},
		{name: "invalid port", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "HTTP_ADDR": "localhost:70000"})},
		{name: "invalid duration", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "HTTP_READ_TIMEOUT": "soon"})},
		{name: "zero duration", values: authValues(map[string]string{"DATABASE_URL": validDatabaseURL, "SHUTDOWN_TIMEOUT": "0s"})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := load(mapLookup(test.values))
			if err == nil {
				t.Fatal("load() error = nil, want validation error")
			}
			if strings.Contains(err.Error(), sensitiveURL) || strings.Contains(err.Error(), "private-password") {
				t.Fatalf("error leaked DATABASE_URL: %v", err)
			}
		})
	}
}

func authValues(values map[string]string) map[string]string {
	if _, ok := values["AUTH_JWT_HMAC_KEY"]; !ok {
		values["AUTH_JWT_HMAC_KEY"] = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
	}
	if _, ok := values["AUTH_JWT_ISSUER"]; !ok {
		values["AUTH_JWT_ISSUER"] = "monelog-api"
	}
	if _, ok := values["AUTH_JWT_AUDIENCE"]; !ok {
		values["AUTH_JWT_AUDIENCE"] = "monelog-app"
	}
	if _, ok := values["AUTH_ACCESS_TOKEN_TTL"]; !ok {
		values["AUTH_ACCESS_TOKEN_TTL"] = "24h"
	}
	if _, ok := values["AUTH_ALLOWED_ORIGINS"]; !ok {
		values["AUTH_ALLOWED_ORIGINS"] = "https://app.example.com"
	}
	return values
}

const validDatabaseURL = "postgres://user:password@localhost:5432/monelog"

func mapLookup(values map[string]string) lookupEnv {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
