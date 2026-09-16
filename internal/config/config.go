package config

import (
	"encoding/base64"
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
	Auth        AuthConfig
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

// AuthConfig contains security-sensitive authentication configuration.
type AuthConfig struct {
	JWTSigningKey       []byte
	JWTIssuer           string
	JWTAudience         string
	AccessTokenLifetime time.Duration
	AllowedOrigins      []string
}

// Load reads and validates configuration from environment variables.
func Load() (Config, error) {
	lookup, err := loadFileEnvironment(os.LookupEnv, os.ReadFile)
	if err != nil {
		return Config{}, err
	}
	return load(lookup)
}

type lookupEnv func(string) (string, bool)
type readFile func(string) ([]byte, error)

// loadFileEnvironment combines the optional environment-specific file with the
// process environment. Process values always take precedence over file values.
func loadFileEnvironment(process lookupEnv, read readFile) (lookupEnv, error) {
	appEnv := defaultAppEnv
	if value, ok := process("APP_ENV"); ok && strings.TrimSpace(value) != "" {
		appEnv = strings.TrimSpace(value)
	}
	if !safeEnvironmentName(appEnv) {
		return nil, errors.New("APP_ENV contains invalid characters")
	}

	values, err := parseEnvironmentFile(".env."+appEnv, read)
	if err != nil {
		return nil, err
	}
	return func(name string) (string, bool) {
		if value, ok := process(name); ok {
			return value, true
		}
		value, ok := values[name]
		return value, ok
	}, nil
}

func safeEnvironmentName(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func parseEnvironmentFile(name string, read readFile) (map[string]string, error) {
	contents, err := read(name)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}

	values := make(map[string]string)
	for lineNumber, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || !validEnvironmentKey(key) {
			return nil, fmt.Errorf("invalid environment assignment at %s:%d", name, lineNumber+1)
		}
		parsedValue, err := parseEnvironmentValue(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid environment value at %s:%d", name, lineNumber+1)
		}
		values[key] = parsedValue
	}
	return values, nil
}

func validEnvironmentKey(value string) bool {
	if value == "" {
		return false
	}
	for index, character := range value {
		if character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' || character == '_' || index > 0 && character >= '0' && character <= '9' {
			continue
		}
		return false
	}
	return true
}

func parseEnvironmentValue(value string) (string, error) {
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1], nil
	}
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return strconv.Unquote(value)
	}
	if strings.HasPrefix(value, "\"") || strings.HasSuffix(value, "\"") || strings.HasPrefix(value, "'") || strings.HasSuffix(value, "'") {
		return "", errors.New("unbalanced quote")
	}
	return value, nil
}

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

	authConfig, err := loadAuthConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnv:      appEnv,
		DatabaseURL: databaseURL,
		Auth:        authConfig,
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

func loadAuthConfig(lookup lookupEnv) (AuthConfig, error) {
	encodedKey, ok := lookup("AUTH_JWT_HMAC_KEY")
	if !ok || strings.TrimSpace(encodedKey) == "" {
		return AuthConfig{}, errors.New("AUTH_JWT_HMAC_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedKey))
	if err != nil || len(key) < 32 {
		return AuthConfig{}, errors.New("AUTH_JWT_HMAC_KEY must be base64-encoded and at least 32 bytes")
	}
	issuer, err := requiredString(lookup, "AUTH_JWT_ISSUER")
	if err != nil {
		return AuthConfig{}, err
	}
	audience, err := requiredString(lookup, "AUTH_JWT_AUDIENCE")
	if err != nil {
		return AuthConfig{}, err
	}
	accessTokenLifetime, err := durationValue(lookup, "AUTH_ACCESS_TOKEN_TTL", 24*time.Hour)
	if err != nil || accessTokenLifetime > 24*time.Hour {
		return AuthConfig{}, errors.New("AUTH_ACCESS_TOKEN_TTL must be positive and no more than 24 hours")
	}
	origins, err := allowedOrigins(lookup)
	if err != nil {
		return AuthConfig{}, err
	}
	return AuthConfig{JWTSigningKey: key, JWTIssuer: issuer, JWTAudience: audience, AccessTokenLifetime: accessTokenLifetime, AllowedOrigins: origins}, nil
}

func requiredString(lookup lookupEnv, name string) (string, error) {
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return strings.TrimSpace(value), nil
}

func allowedOrigins(lookup lookupEnv) ([]string, error) {
	value, err := requiredString(lookup, "AUTH_ALLOWED_ORIGINS")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	origins := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		origin := strings.TrimSpace(item)
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, errors.New("AUTH_ALLOWED_ORIGINS must contain valid origins")
		}
		if _, ok := seen[origin]; !ok {
			seen[origin] = struct{}{}
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return nil, errors.New("AUTH_ALLOWED_ORIGINS must contain at least one origin")
	}
	return origins, nil
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
