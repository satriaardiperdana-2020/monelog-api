package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/term"

	"github.com/satriaardiperdana-2020/monelog-api/internal/config"
	"github.com/satriaardiperdana-2020/monelog-api/internal/repository/postgresql"
	"github.com/satriaardiperdana-2020/monelog-api/internal/service"
)

func main() {
	email := flag.String("email", "", "initial admin email")
	timezone := flag.String("timezone", "Asia/Jakarta", "IANA timezone")
	flag.Parse()
	if *email == "" {
		fmt.Fprintln(os.Stderr, "-email is required")
		os.Exit(2)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr, "password must be read from an interactive terminal")
		os.Exit(2)
	}
	fmt.Fprint(os.Stderr, "Initial admin password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read password")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration")
		os.Exit(1)
	}
	if err := bootstrap(context.Background(), cfg, *email, *timezone, string(password)); err != nil {
		fmt.Fprintln(os.Stderr, "admin bootstrap failed")
		os.Exit(1)
	}
	slog.Info("initial admin created", "email", *email)
}

func bootstrap(ctx context.Context, cfg config.Config, email, timezone, password string) error {
	pool, err := postgresql.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	authService, err := service.NewAuth(pool, cfg.Auth.JWTSigningKey, cfg.Auth.JWTIssuer, cfg.Auth.JWTAudience, cfg.Auth.AccessTokenLifetime)
	if err != nil {
		return err
	}
	_, err = authService.BootstrapInitialAdmin(ctx, service.RegisterInput{Email: email, Password: password, Timezone: timezone})
	if errors.Is(err, service.ErrConflict) {
		return errors.New("an active initial admin already exists")
	}
	return err
}
