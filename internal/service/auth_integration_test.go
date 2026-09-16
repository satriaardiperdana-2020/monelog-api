//go:build integration

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appauth "github.com/satriaardiperdana-2020/monelog-api/internal/auth"
)

func TestAuthenticationLifecycleAndRefreshReplay(t *testing.T) {
	pool := openAuthTestPool(t)
	authService, err := NewAuth(pool, []byte("0123456789abcdef0123456789abcdef"), "monelog-api", "monelog-app", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	user, err := authService.Register(ctx, RegisterInput{Email: "person@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Role != RoleUser || user.IsDelete {
		t.Fatalf("registration role/state = %+v", user)
	}
	stored, err := authService.queries.GetActiveUserByID(ctx, toPGUUID(user.ID))
	if err != nil || stored.PasswordHash == "correct horse battery staple" || !appauth.VerifyPassword(stored.PasswordHash, "correct horse battery staple") {
		t.Fatalf("stored password was not an Argon2id hash: user=%+v err=%v", stored, err)
	}
	categories, err := authService.queries.ListActiveCategories(ctx, stored.ID)
	if err != nil || len(categories) != len(defaultCategories) {
		t.Fatalf("default categories=%d err=%v, want %d", len(categories), err, len(defaultCategories))
	}

	login, err := authService.Login(ctx, LoginInput{Email: "person@example.com", Password: "correct horse battery staple"})
	if err != nil || login.RefreshToken == "" || login.AccessToken == "" {
		t.Fatalf("login=%+v err=%v", login, err)
	}
	if _, err := authService.AuthenticateAccess(ctx, login.AccessToken); err != nil {
		t.Fatalf("authenticate access token: %v", err)
	}

	type result struct {
		session Session
		err     error
	}
	results := make(chan result, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			session, err := authService.Refresh(ctx, login.RefreshToken)
			results <- result{session: session, err: err}
		}()
	}
	wait.Wait()
	close(results)
	var rotated Session
	successes := 0
	for result := range results {
		if result.err == nil {
			successes++
			rotated = result.session
			continue
		}
		if !errors.Is(result.err, ErrAuthentication) {
			t.Fatalf("concurrent refresh error = %v", result.err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent refresh successes=%d, want 1", successes)
	}
	if _, err := authService.Refresh(ctx, rotated.RefreshToken); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("replayed family successor refresh error=%v, want authentication failure", err)
	}

	newLogin, err := authService.Login(ctx, LoginInput{Email: "person@example.com", Password: "correct horse battery staple"})
	if err != nil {
		t.Fatal(err)
	}
	if err := authService.Logout(ctx, newLogin.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := authService.Refresh(ctx, newLogin.RefreshToken); !errors.Is(err, ErrAuthentication) {
		t.Fatalf("logged out refresh error=%v, want authentication failure", err)
	}
	if _, err := authService.UpdateProfile(ctx, Actor{UserID: user.ID, Role: RoleUser}, "Not/AZone", user.Version); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid timezone error=%v, want validation failure", err)
	}
	admin, err := authService.BootstrapInitialAdmin(ctx, RegisterInput{Email: "admin@example.com", Password: "correct horse battery staple", Timezone: "UTC"})
	if err != nil || admin.Role != RoleAdmin {
		t.Fatalf("bootstrap admin=%+v err=%v", admin, err)
	}
	if _, err := authService.BootstrapInitialAdmin(ctx, RegisterInput{Email: "other-admin@example.com", Password: "correct horse battery staple", Timezone: "UTC"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("second bootstrap error=%v, want conflict", err)
	}
}

func openAuthTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(adminPool.Close)
	if err := adminPool.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	schema := "monelog_auth_" + hex.EncodeToString(random)
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminPool.Exec(context.Background(), "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Errorf("drop schema: %v", err)
		}
	})
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations", "*.up.sql"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("find migrations: %v", err)
	}
	sort.Strings(paths)
	for _, path := range paths {
		migration, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatalf("apply %s: %v", filepath.Base(path), err)
		}
	}
	return pool
}
