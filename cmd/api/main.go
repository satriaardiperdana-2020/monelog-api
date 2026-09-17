package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"

	"github.com/satriaardiperdana-2020/monelog-api/internal/config"
	"github.com/satriaardiperdana-2020/monelog-api/internal/handlers"
	appmiddleware "github.com/satriaardiperdana-2020/monelog-api/internal/middleware"
	"github.com/satriaardiperdana-2020/monelog-api/internal/repository/postgresql"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	pool, err := postgresql.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	healthService := service.NewHealth(pool, cfg.HTTP.ReadyTimeout)
	healthHandler := handlers.NewHealth(healthService)
	authService, err := service.NewAuth(pool, cfg.Auth.JWTSigningKey, cfg.Auth.JWTIssuer, cfg.Auth.JWTAudience, cfg.Auth.AccessTokenLifetime)
	if err != nil {
		return errors.New("initialize authentication")
	}
	authHandler := handlers.NewAuth(authService, cfg.Auth.AllowedOrigins, appmiddleware.NewLoginRateLimiter())
	categoryService := service.NewCategories(pool)
	categoryHandler := handlers.NewCategories(categoryService)

	e := echo.New()
	appmiddleware.Register(e, logger)
	handlers.RegisterRoutes(e, healthHandler, authHandler, categoryHandler, appmiddleware.Authenticate(authService))

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           e,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	listener, err := net.Listen("tcp", cfg.HTTP.Addr)
	if err != nil {
		return fmt.Errorf("HTTP listener failed: %w", err)
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API listening", "address", cfg.HTTP.Addr, "environment", cfg.AppEnv)
		serverErrors <- server.Serve(listener)
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("HTTP server failed: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return errors.New("HTTP server shutdown timed out")
		}
		if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server failed during shutdown: %w", err)
		}
		logger.Info("API stopped gracefully")
		return nil
	}
}
