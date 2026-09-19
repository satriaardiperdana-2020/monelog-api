package service

import (
	"testing"
	"time"
)

func TestTransactionCanonicalHashUsesNormalizedValues(t *testing.T) {
	owner, category := int64(1), int64(2)
	one := validatedTransaction{date: time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC), kind: CategoryTypeExpense, categoryID: category, amount: Money{cents: 4350000}, title: "Belanja"}
	two := one
	if canonicalTransactionHash(owner, one) != canonicalTransactionHash(owner, two) {
		t.Fatal("equivalent canonical transactions produced different hashes")
	}
	two.title = "Belanja lain"
	if canonicalTransactionHash(owner, one) == canonicalTransactionHash(owner, two) {
		t.Fatal("different transactions produced the same hash")
	}
}

func TestTransactionCursorBindsScopeModeAndFilters(t *testing.T) {
	actor, owner := int64(1), int64(2)
	scope := TransactionScope{Actor: Actor{UserID: actor, Role: RoleAdmin}, Owner: owner, Admin: true}
	filter := TransactionFilter{StartDate: "2026-09-01", EndDate: "2026-09-17", Type: CategoryTypeExpense}
	encoded := encodeCursor(cursorFor(scope, filter, "transactions", time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC), time.Now(), 3))
	cursor, err := decodeTransactionCursor(encoded)
	if err != nil || !cursorMatches(cursor, scope, filter, "transactions") {
		t.Fatalf("cursor round trip failed: %v", err)
	}
	scope.Admin = false
	if cursorMatches(cursor, scope, filter, "transactions") {
		t.Fatal("admin cursor was accepted in personal mode")
	}
}

func TestDateRangeMaximumIsInclusive366Days(t *testing.T) {
	if _, _, err := validateDateRange("2025-09-17", "2026-09-17"); err != nil {
		t.Fatalf("366-day inclusive range rejected: %v", err)
	}
	if _, _, err := validateDateRange("2025-09-16", "2026-09-17"); err == nil {
		t.Fatal("range over 366 inclusive days was accepted")
	}
}
