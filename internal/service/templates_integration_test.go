//go:build integration

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

func TestTemplateLifecycleAuthorizationAndNoFinancialMutation(t *testing.T) {
	pool := openAuthTestPool(t)
	authService, err := NewAuth(pool, []byte("0123456789abcdef0123456789abcdef"), "monelog-api", "monelog-app", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner, err := authService.Register(ctx, RegisterInput{Email: "template-owner@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := authService.Register(ctx, RegisterInput{Email: "template-other@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := authService.BootstrapInitialAdmin(ctx, RegisterInput{Email: "template-admin@example.com", Password: "correct horse battery staple", Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	categories, err := authService.queries.ListActiveCategories(ctx, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	var expenseCategory int64
	for _, category := range categories {
		if category.Type == CategoryTypeExpense {
			expenseCategory = category.ID
			break
		}
	}
	if expenseCategory == 0 {
		t.Fatal("expense category was not seeded")
	}

	templates := NewTemplates(pool)
	transactions := NewTransactions(pool)
	ownerScope := TemplateScope{Actor: Actor{UserID: owner.ID, Role: RoleUser}, Owner: owner.ID}
	input := TemplateInput{CategoryID: expenseCategory, Type: CategoryTypeExpense, Name: " Lunch ", Amount: "43500.00", Title: " Office lunch "}
	created, err := templates.Create(ctx, ownerScope, input)
	if err != nil || created.Name != "Lunch" || created.Title != "Office lunch" || created.IsDelete || created.Version != 1 {
		t.Fatalf("create=%+v err=%v", created, err)
	}
	for range 2 {
		draft, err := templates.Apply(ctx, ownerScope, created.ID)
		if err != nil || draft.CategoryID != expenseCategory || draft.Amount != "43500.00" {
			t.Fatalf("apply=%+v err=%v", draft, err)
		}
	}
	transactionScope := TransactionScope{Actor: Actor{UserID: owner.ID, Role: RoleUser}, Owner: owner.ID}
	history, err := transactions.List(ctx, transactionScope, TransactionFilter{StartDate: "2026-09-01", EndDate: "2026-09-30"})
	if err != nil || len(history.Items) != 0 {
		t.Fatalf("template mutated transactions: %+v err=%v", history, err)
	}
	summary, err := transactions.GetReportSummary(ctx, transactionScope, ReportFilter{Range: "custom", StartDate: "2026-09-01", EndDate: "2026-09-30"})
	if err != nil || summary.Income != "0.00" || summary.Expense != "0.00" || summary.Difference != "0.00" {
		t.Fatalf("template mutated report: %+v err=%v", summary, err)
	}
	if _, err := templates.Create(ctx, ownerScope, TemplateInput{CategoryID: expenseCategory, Type: CategoryTypeIncome, Name: "Wrong type", Amount: "1.00", Title: "Wrong type"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("type mismatch error=%v", err)
	}
	otherCategories, err := authService.queries.ListActiveCategories(ctx, other.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := templates.Create(ctx, ownerScope, TemplateInput{CategoryID: otherCategories[0].ID, Type: otherCategories[0].Type, Name: "Foreign", Amount: "1.00", Title: "Foreign"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign category error=%v", err)
	}
	if err := templates.SoftDelete(ctx, ownerScope, created.ID, created.Version); err != nil {
		t.Fatal(err)
	}
	if _, err := templates.Get(ctx, ownerScope, created.ID, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("active after delete error=%v", err)
	}
	trash, err := templates.Get(ctx, ownerScope, created.ID, true)
	if err != nil || !trash.IsDelete || trash.Version != 2 {
		t.Fatalf("trash=%+v err=%v", trash, err)
	}
	if _, err := templates.Restore(ctx, ownerScope, created.ID, created.Version); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale restore error=%v", err)
	}
	restored, err := templates.Restore(ctx, ownerScope, created.ID, trash.Version)
	if err != nil || restored.IsDelete || restored.Version != 3 {
		t.Fatalf("restore=%+v err=%v", restored, err)
	}
	categoryService := NewCategories(pool)
	if err := categoryService.Archive(ctx, CategoryScope{Actor: Actor{UserID: owner.ID, Role: RoleUser}, Owner: owner.ID}, expenseCategory, 1); err != nil {
		t.Fatalf("archive template category: %v", err)
	}
	if _, err := templates.Apply(ctx, ownerScope, created.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("apply with deleted category error=%v", err)
	}
	otherScope := TemplateScope{Actor: Actor{UserID: other.ID, Role: RoleUser}, Owner: other.ID}
	if _, err := templates.Get(ctx, otherScope, created.ID, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner template error=%v", err)
	}
	adminScope := TemplateScope{Actor: Actor{UserID: admin.ID, Role: RoleAdmin}, Owner: other.ID, Admin: true, RequestID: "template-admin-request"}
	adminCreated, err := templates.Create(ctx, adminScope, TemplateInput{CategoryID: otherCategories[0].ID, Type: otherCategories[0].Type, Name: "Admin owned", Amount: "1000.00", Title: "Admin template"})
	if err != nil || adminCreated.UserID != other.ID {
		t.Fatalf("admin create=%+v err=%v", adminCreated, err)
	}
	if _, err := templates.Apply(ctx, adminScope, adminCreated.ID); err != nil {
		t.Fatalf("admin apply: %v", err)
	}
	events, err := authService.queries.ListAdminAccessEventsByTarget(ctx, db.ListAdminAccessEventsByTargetParams{TargetUserID: &other.ID, PageSize: 20})
	if err != nil || len(events) < 2 || events[0].ResourceType != "template" {
		t.Fatalf("template audit events=%+v err=%v", events, err)
	}
}
