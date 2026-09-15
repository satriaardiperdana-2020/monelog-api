package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAppEnv            = "development"
	defaultHTTPAddr          = "127.0.0.1:8080"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 10 * time.Second
	defaultWriteTimeout      = 15 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultReadyTimeout      = 2 * time.Second
)

// Config contains all runtime configuration required by the API process.
type Config struct {
	AppEnv      string
	HTTP        HTTPConfig
	DatabaseURL string
}

// HTTPConfig contains HTTP server and lifecycle timeouts.
type HTTPConfig struct {
	Addr              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	ReadyTimeout      time.Duration
}

// Load reads and validates configuration from environment variables.
func Load() (Config, error) {
	return load(os.LookupEnv)
}

type lookupEnv func(string) (string, bool)

func load(lookup lookupEnv) (Config, error) {
	appEnv, err := stringValue(lookup, "APP_ENV", defaultAppEnv)
	if err != nil {
		return Config{}, err
	}

	addr, err := stringValue(lookup, "HTTP_ADDR", defaultHTTPAddr)
	if err != nil {
		return Config{}, err
	}
	if err := validateAddress(addr); err != nil {
		return Config{}, err
	}

	readHeaderTimeout, err := durationValue(lookup, "HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return Config{}, err
	}
	readTimeout, err := durationValue(lookup, "HTTP_READ_TIMEOUT", defaultReadTimeout)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := durationValue(lookup, "HTTP_WRITE_TIMEOUT", defaultWriteTimeout)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := durationValue(lookup, "HTTP_IDLE_TIMEOUT", defaultIdleTimeout)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationValue(lookup, "SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}
	readyTimeout, err := durationValue(lookup, "HEALTH_READY_TIMEOUT", defaultReadyTimeout)
	if err != nil {
		return Config{}, err
	}

	databaseURL, ok := lookup("DATABASE_URL")
	if !ok || strings.TrimSpace(databaseURL) == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if err := validateDatabaseURL(databaseURL); err != nil {
		return Config{}, err
	}

	return Config{
		AppEnv:      appEnv,
		DatabaseURL: databaseURL,
		HTTP: HTTPConfig{
			Addr:              addr,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
			ShutdownTimeout:   shutdownTimeout,
			ReadyTimeout:      readyTimeout,
		},
	}, nil
}

func stringValue(lookup lookupEnv, name, fallback string) (string, error) {
	value, ok := lookup(name)
	if !ok {
		return fallback, nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s must not be empty", name)
	}
	return value, nil
}

func durationValue(lookup lookupEnv, name string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(name)
	if !ok {
		return fallback, nil
	}
	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", name)
	}
	return duration, nil
}

func validateAddress(addr string) error {
	host, portText, err := net.SplitHostPort(addr)
	if err != nil || strings.TrimSpace(host) == "" {
		return errors.New("HTTP_ADDR must use a valid host:port")
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("HTTP_ADDR must use a port between 1 and 65535")
	}
	return nil
}

func validateDatabaseURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") || parsed.Host == "" {
		return errors.New("DATABASE_URL must be a valid PostgreSQL URL")
	}
	return nil
}
