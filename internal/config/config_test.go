package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	values := map[string]string{
		"DATABASE_URL": "postgres://user:password@localhost:5432/monelog?sslmode=disable",
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
	}

	got, err := load(mapLookup(values))
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if got.AppEnv != "test" || got.HTTP.Addr != "localhost:9090" {
		t.Fatalf("unexpected string configuration: %+v", got)
	}
	if got.HTTP.ReadyTimeout != 750*time.Millisecond || got.HTTP.ShutdownTimeout != 5*time.Second {
		t.Fatalf("unexpected timeout configuration: %+v", got.HTTP)
	}
}

func TestLoadRejectsInvalidConfigurationWithoutLeakingDatabaseURL(t *testing.T) {
	const sensitiveURL = "mysql://private-user:private-password@db.internal/monelog"

	tests := []struct {
		name   string
		values map[string]string
	}{
		{name: "missing database URL", values: map[string]string{}},
		{name: "invalid database URL", values: map[string]string{"DATABASE_URL": sensitiveURL}},
		{name: "empty app environment", values: map[string]string{"DATABASE_URL": validDatabaseURL, "APP_ENV": " "}},
		{name: "invalid address", values: map[string]string{"DATABASE_URL": validDatabaseURL, "HTTP_ADDR": "localhost"}},
		{name: "invalid port", values: map[string]string{"DATABASE_URL": validDatabaseURL, "HTTP_ADDR": "localhost:70000"}},
		{name: "invalid duration", values: map[string]string{"DATABASE_URL": validDatabaseURL, "HTTP_READ_TIMEOUT": "soon"}},
		{name: "zero duration", values: map[string]string{"DATABASE_URL": validDatabaseURL, "SHUTDOWN_TIMEOUT": "0s"}},
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

const validDatabaseURL = "postgres://user:password@localhost:5432/monelog"

func mapLookup(values map[string]string) lookupEnv {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
