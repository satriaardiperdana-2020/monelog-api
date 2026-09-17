//go:build integration

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

func TestCategoryLifecycleAndOwnership(t *testing.T) {
	pool := openAuthTestPool(t)
	authService, err := NewAuth(pool, []byte("0123456789abcdef0123456789abcdef"), "monelog-api", "monelog-app", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner, err := authService.Register(ctx, RegisterInput{Email: "owner-categories@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := authService.Register(ctx, RegisterInput{Email: "other-categories@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := authService.BootstrapInitialAdmin(ctx, RegisterInput{Email: "admin-categories@example.com", Password: "correct horse battery staple", Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	categories := NewCategories(pool)
	ownerScope := CategoryScope{Actor: Actor{UserID: owner.ID, Role: RoleUser}, Owner: owner.ID}
	created, err := categories.Create(ctx, ownerScope, CategoryTypeExpense, "  Subscriptions  ")
	if err != nil || created.Name != "Subscriptions" || created.IsDelete || created.Version != 1 {
		t.Fatalf("create=%+v err=%v", created, err)
	}
	if _, err := categories.Create(ctx, ownerScope, CategoryTypeExpense, "subscriptions"); !errors.Is(err, ErrConflict) {
		t.Fatalf("case-insensitive duplicate error=%v", err)
	}
	if err := categories.Archive(ctx, ownerScope, created.ID, created.Version); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if _, err := categories.Get(ctx, ownerScope, created.ID, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("active read after archive error=%v", err)
	}
	archived, err := categories.Get(ctx, ownerScope, created.ID, true)
	if err != nil || !archived.IsDelete || archived.Version != 2 {
		t.Fatalf("trash category=%+v err=%v", archived, err)
	}
	if _, err := categories.Create(ctx, ownerScope, CategoryTypeExpense, "SUBSCRIPTIONS"); !errors.Is(err, ErrConflict) {
		t.Fatalf("archived name was reused: %v", err)
	}
	if _, err := categories.Restore(ctx, ownerScope, created.ID, 1); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale restore error=%v", err)
	}
	restored, err := categories.Restore(ctx, ownerScope, created.ID, archived.Version)
	if err != nil || restored.IsDelete || restored.Version != 3 {
		t.Fatalf("restore=%+v err=%v", restored, err)
	}
	otherScope := CategoryScope{Actor: Actor{UserID: other.ID, Role: RoleUser}, Owner: other.ID}
	if _, err := categories.Get(ctx, otherScope, created.ID, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner get error=%v", err)
	}
	adminScope := CategoryScope{Actor: Actor{UserID: admin.ID, Role: RoleAdmin}, Owner: other.ID, Admin: true}
	if _, err := categories.Create(ctx, adminScope, CategoryTypeIncome, "Admin owned"); err != nil {
		t.Fatalf("admin create: %v", err)
	}
	events, err := authService.queries.ListAdminAccessEventsByTarget(ctx, db.ListAdminAccessEventsByTargetParams{TargetUserID: toPGUUID(other.ID), PageSize: 10})
	if err != nil || len(events) != 1 || events[0].ActorUserID.Bytes != toPGUUID(admin.ID).Bytes || events[0].Action != "create" {
		t.Fatalf("admin audit events=%+v err=%v", events, err)
	}
}
