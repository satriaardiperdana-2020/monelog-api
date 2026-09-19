//go:build integration

package service

import (
	"context"
	"testing"
	"time"
)

func TestReportSummaryAndBreakdown(t *testing.T) {
	pool := openAuthTestPool(t)
	authService, err := NewAuth(pool, []byte("0123456789abcdef0123456789abcdef"), "monelog-api", "monelog-app", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	owner, err := authService.Register(ctx, RegisterInput{Email: "reports@example.com", Password: "correct horse battery staple", Timezone: "Asia/Jakarta"})
	if err != nil {
		t.Fatal(err)
	}
	categories, err := authService.queries.ListActiveCategories(ctx, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	var incomeCategory, expenseCategory int64
	for _, category := range categories {
		if category.Type == CategoryTypeIncome {
			incomeCategory = category.ID
		} else if category.Type == CategoryTypeExpense {
			expenseCategory = category.ID
		}
	}

	transactions := NewTransactions(pool)
	transactions.now = func() time.Time { return time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC) }
	scope := TransactionScope{Actor: Actor{UserID: owner.ID, Role: RoleUser}, Owner: owner.ID}

	// Create transactions
	for _, input := range []TransactionInput{
		{TransactionDate: "2026-09-10", Type: CategoryTypeIncome, CategoryID: incomeCategory, Amount: "1000.00", Title: "Income 1", ClientRequestID: 1},
		{TransactionDate: "2026-09-11", Type: CategoryTypeExpense, CategoryID: expenseCategory, Amount: "500.00", Title: "Expense 1", ClientRequestID: 2},
		{TransactionDate: "2026-09-12", Type: CategoryTypeExpense, CategoryID: expenseCategory, Amount: "200.00", Title: "Expense 2", ClientRequestID: 3},
	} {
		if _, err := transactions.Create(ctx, scope, input); err != nil {
			t.Fatalf("create transaction: %v", err)
		}
	}

	// Test Summary
	filter := ReportFilter{Range: "custom", StartDate: "2026-09-01", EndDate: "2026-09-17"}
	summary, err := transactions.GetReportSummary(ctx, scope, filter)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Income != "1000.00" || summary.Expense != "700.00" || summary.Difference != "300.00" {
		t.Errorf("summary mismatch: %+v", summary)
	}

	// Test Breakdown
	breakdown, err := transactions.GetReportBreakdown(ctx, scope, ReportFilter{Range: "custom", StartDate: "2026-09-01", EndDate: "2026-09-17", GroupBy: "week"})
	if err != nil {
		t.Fatalf("breakdown: %v", err)
	}
	if len(breakdown.Periods) == 0 {
		t.Error("breakdown empty")
	}
}
