//go:build integration

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

func TestTransactionHistoryDailySummaryAndLifecycle(t *testing.T) {
	pool := openAuthTestPool(t)
	authService, err := NewAuth(pool, []byte("0123456789abcdef0123456789abcdef"), "monelog-api", "monelog-app", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner, err := authService.Register(ctx, RegisterInput{Email: "transactions@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
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

	transactions := NewTransactions(pool)
	transactions.now = func() time.Time { return time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC) }
	scope := TransactionScope{Actor: Actor{UserID: owner.ID, Role: RoleUser}, Owner: owner.ID}
	requestID := int64(1)
	input := TransactionInput{TransactionDate: "2026-09-16", Type: CategoryTypeExpense, CategoryID: expenseCategory, Amount: "43500.00", Title: " Belanja ", ClientRequestID: requestID}
	created, err := transactions.Create(ctx, scope, input)
	if err != nil || !created.Created || created.Transaction.Amount != "43500.00" || created.Transaction.Title != "Belanja" {
		t.Fatalf("create=%+v err=%v", created, err)
	}
	replay, err := transactions.Create(ctx, scope, input)
	if err != nil || replay.Created || replay.Transaction.ID != created.Transaction.ID {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	conflict := input
	conflict.Amount = "1.00"
	if _, err := transactions.Create(ctx, scope, conflict); !errors.Is(err, ErrConflict) {
		t.Fatalf("different replay error=%v", err)
	}

	history, err := transactions.List(ctx, scope, TransactionFilter{StartDate: "2026-09-01", EndDate: "2026-09-17"})
	if err != nil || len(history.Items) != 1 || history.Items[0].ID != created.Transaction.ID {
		t.Fatalf("history=%+v err=%v", history, err)
	}
	daily, err := transactions.ListDailySummaries(ctx, scope, "2026-09-01", "2026-09-17", 30, "")
	if err != nil || len(daily.Items) != 1 || daily.Items[0].Expense != "43500.00" || daily.Items[0].Difference != "-43500.00" {
		t.Fatalf("daily=%+v err=%v", daily, err)
	}

	if err := transactions.SoftDeleteTransaction(ctx, scope, created.Transaction.ID, created.Transaction.Version); err != nil {
		t.Fatal(err)
	}
	history, err = transactions.List(ctx, scope, TransactionFilter{StartDate: "2026-09-01", EndDate: "2026-09-17"})
	if err != nil || len(history.Items) != 0 {
		t.Fatalf("active after delete=%+v err=%v", history, err)
	}
	trash, err := transactions.List(ctx, scope, TransactionFilter{StartDate: "2026-09-01", EndDate: "2026-09-17", Deleted: true})
	if err != nil || len(trash.Items) != 1 || !trash.Items[0].IsDelete {
		t.Fatalf("trash=%+v err=%v", trash, err)
	}
	daily, err = transactions.ListDailySummaries(ctx, scope, "2026-09-01", "2026-09-17", 30, "")
	if err != nil || len(daily.Items) != 0 {
		t.Fatalf("daily after delete=%+v err=%v", daily, err)
	}
	restored, err := transactions.RestoreTransaction(ctx, scope, created.Transaction.ID, trash.Items[0].Version)
	if err != nil || restored.IsDelete || restored.Version != 3 {
		t.Fatalf("restore=%+v err=%v", restored, err)
	}
}

func TestAdminTransactionScopeAuditAndStableCursor(t *testing.T) {
	pool := openAuthTestPool(t)
	authService, err := NewAuth(pool, []byte("0123456789abcdef0123456789abcdef"), "monelog-api", "monelog-app", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner, err := authService.Register(ctx, RegisterInput{Email: "transaction-owner@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := authService.Register(ctx, RegisterInput{Email: "transaction-other@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := authService.BootstrapInitialAdmin(ctx, RegisterInput{Email: "transaction-admin@example.com", Password: "correct horse battery staple", Timezone: "UTC"})
	if err != nil {
		t.Fatal(err)
	}
	categories, err := authService.queries.ListActiveCategories(ctx, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	var categoryID int64
	for _, category := range categories {
		if category.Type == CategoryTypeExpense {
			categoryID = category.ID
			break
		}
	}
	transactions := NewTransactions(pool)
	transactions.now = func() time.Time { return time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC) }
	adminScope := TransactionScope{Actor: Actor{UserID: admin.ID, Role: RoleAdmin}, Owner: owner.ID, Admin: true, RequestID: "admin-request"}
	createdIDs := make(map[int64]struct{})
	for i, amount := range []string{"1.00", "2.00", "3.00"} {
		created, err := transactions.Create(ctx, adminScope, TransactionInput{TransactionDate: "2026-09-16", Type: CategoryTypeExpense, CategoryID: categoryID, Amount: amount, Title: "Admin entry " + amount, ClientRequestID: int64(i + 1)})
		if err != nil {
			t.Fatalf("admin create %d: %v", i, err)
		}
		if created.Transaction.UserID != owner.ID || created.Transaction.CreatedBy != admin.ID {
			t.Fatalf("owner/actor=%+v", created.Transaction)
		}
		createdIDs[created.Transaction.ID] = struct{}{}
	}
	first, err := transactions.List(ctx, adminScope, TransactionFilter{StartDate: "2026-09-01", EndDate: "2026-09-17", Limit: 2})
	if err != nil || len(first.Items) != 2 || first.NextCursor == nil {
		t.Fatalf("first page=%+v err=%v", first, err)
	}
	second, err := transactions.List(ctx, adminScope, TransactionFilter{StartDate: "2026-09-01", EndDate: "2026-09-17", Limit: 2, Cursor: *first.NextCursor})
	if err != nil || len(second.Items) != 1 || second.NextCursor != nil {
		t.Fatalf("second page=%+v err=%v", second, err)
	}
	seen := map[int64]struct{}{}
	for _, item := range append(first.Items, second.Items...) {
		if _, exists := seen[item.ID]; exists {
			t.Fatalf("duplicate paginated ID %d", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	if len(seen) != len(createdIDs) {
		t.Fatalf("paginated IDs=%d created=%d", len(seen), len(createdIDs))
	}
	otherScope := TransactionScope{Actor: Actor{UserID: other.ID, Role: RoleUser}, Owner: other.ID}
	for id := range createdIDs {
		if _, err := transactions.Get(ctx, otherScope, id, false); !errors.Is(err, ErrNotFound) {
			t.Fatalf("cross-owner get error=%v", err)
		}
		break
	}
	events, err := authService.queries.ListAdminAccessEventsByTarget(ctx, db.ListAdminAccessEventsByTargetParams{TargetUserID: &owner.ID, PageSize: 20})
	if err != nil || len(events) < 5 {
		t.Fatalf("admin audit events=%d err=%v", len(events), err)
	}
}
