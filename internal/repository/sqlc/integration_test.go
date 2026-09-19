//go:build integration

package sqlc_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

func TestSchemaAndScopedLifecycle(t *testing.T) {
	pool := openMigratedTestPool(t)
	queries := db.New(pool)
	ctx := context.Background()

	owner, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        " Owner@Example.com ",
		PasswordHash: "owner-password-hash",
		Timezone:     "Asia/Jakarta",
		Currency:     "IDR",
	})
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if owner.Email != "owner@example.com" || owner.Role != "user" || owner.IsDelete || owner.Version != 1 {
		t.Fatalf("unexpected owner defaults: %+v", owner)
	}
	ownerID := owner.ID

	actor, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        "actor@example.com",
		PasswordHash: "actor-password-hash",
		Timezone:     "Asia/Jakarta",
		Currency:     "IDR",
	})
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	if actor.Role != "user" {
		t.Fatalf("registration query bypassed role default: %q", actor.Role)
	}
	actorID := actor.ID

	ownerCategory, err := queries.CreateCategory(ctx, db.CreateCategoryParams{
		UserID: ownerID, Type: "expense", Name: " Food ",
	})
	if err != nil {
		t.Fatalf("create owner category: %v", err)
	}
	if ownerCategory.Name != "Food" || ownerCategory.IsDelete || ownerCategory.Version != 1 {
		t.Fatalf("unexpected category defaults: %+v", ownerCategory)
	}
	ownerCategoryID := ownerCategory.ID

	template, err := queries.CreateTransactionTemplate(ctx, db.CreateTransactionTemplateParams{
		UserID: ownerID, CategoryID: ownerCategoryID, Type: "expense", Name: " Lunch template ", Amount: testNumeric(43500, 0), Title: " Template lunch ",
	})
	if err != nil {
		t.Fatalf("create transaction template: %v", err)
	}
	if template.Name != "Lunch template" || template.Title != "Template lunch" || template.IsDelete || template.Version != 1 {
		t.Fatalf("unexpected template defaults: %+v", template)
	}
	if active, err := queries.ListActiveTransactionTemplates(ctx, ownerID); err != nil || len(active) != 1 || active[0].ID != template.ID {
		t.Fatalf("list active templates=%+v err=%v", active, err)
	}

	actorCategory, err := queries.CreateCategory(ctx, db.CreateCategoryParams{
		UserID: actorID, Type: "expense", Name: "Food",
	})
	if err != nil {
		t.Fatalf("create actor category: %v", err)
	}
	actorCategoryID := actorCategory.ID

	if _, err := queries.CreateTransactionTemplate(ctx, db.CreateTransactionTemplateParams{
		UserID: ownerID, CategoryID: actorCategoryID, Type: "expense", Name: "Foreign template", Amount: testNumeric(1, 0), Title: "Foreign template",
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("cross-owner category should not create a template, got %v", err)
	}
	if _, err := queries.CreateTransactionTemplate(ctx, db.CreateTransactionTemplateParams{
		UserID: ownerID, CategoryID: ownerCategoryID, Type: "income", Name: "Wrong type template", Amount: testNumeric(1, 0), Title: "Wrong type template",
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("incompatible category should not create a template, got %v", err)
	}
	deletedTemplate, err := queries.SoftDeleteTransactionTemplate(ctx, db.SoftDeleteTransactionTemplateParams{ID: template.ID, UserID: ownerID, ExpectedVersion: 1})
	if err != nil || !deletedTemplate.IsDelete || deletedTemplate.Version != 2 {
		t.Fatalf("soft delete template=%+v err=%v", deletedTemplate, err)
	}
	if active, err := queries.ListActiveTransactionTemplates(ctx, ownerID); err != nil || len(active) != 0 {
		t.Fatalf("deleted template remained active=%+v err=%v", active, err)
	}
	if trashTemplates, err := queries.ListDeletedTransactionTemplates(ctx, ownerID); err != nil || len(trashTemplates) != 1 || trashTemplates[0].ID != template.ID {
		t.Fatalf("list deleted templates=%+v err=%v", trashTemplates, err)
	}
	if _, err := queries.RestoreTransactionTemplate(ctx, db.RestoreTransactionTemplateParams{ID: template.ID, UserID: ownerID, ExpectedVersion: 1}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("stale template restore should affect no row, got %v", err)
	}
	restoredTemplate, err := queries.RestoreTransactionTemplate(ctx, db.RestoreTransactionTemplateParams{ID: template.ID, UserID: ownerID, ExpectedVersion: 2})
	if err != nil || restoredTemplate.IsDelete || restoredTemplate.Version != 3 {
		t.Fatalf("restore template=%+v err=%v", restoredTemplate, err)
	}

	requestID := int64(6)
	transaction, err := queries.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:          ownerID,
		CategoryID:      ownerCategoryID,
		Type:            "expense",
		Amount:          testNumeric(43500, 0),
		TransactionDate: testDate(2026, time.September, 16),
		Title:           " Lunch ",
		ClientRequestID: requestID,
		RequestHash:     "request-hash",
		ActorUserID:     actorID,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if transaction.Title != "Lunch" || transaction.IsDelete || transaction.Version != 1 {
		t.Fatalf("unexpected transaction defaults: %+v", transaction)
	}
	transactionID := transaction.ID
	if transaction.UserID != ownerID || transaction.CreatedBy != actorID || transaction.UpdatedBy != actorID {
		t.Fatalf("owner/actor attribution was not preserved: %+v", transaction)
	}

	_, err = queries.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:          ownerID,
		CategoryID:      actorCategoryID,
		Type:            "expense",
		Amount:          testNumeric(1, 0),
		TransactionDate: testDate(2026, time.September, 16),
		Title:           "Wrong owner",
		ClientRequestID: 8,
		RequestHash:     "wrong-owner-hash",
		ActorUserID:     actorID,
	})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("cross-owner category should not create a transaction, got %v", err)
	}

	_, err = queries.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:          ownerID,
		CategoryID:      ownerCategoryID,
		Type:            "expense",
		Amount:          testNumeric(10, 0),
		TransactionDate: testDate(2026, time.September, 16),
		Title:           "Duplicate request",
		ClientRequestID: requestID,
		RequestHash:     "different-hash",
		ActorUserID:     actorID,
	})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("duplicate idempotency key should return no inserted row, got %v", err)
	}

	_, err = queries.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:          ownerID,
		CategoryID:      ownerCategoryID,
		Type:            "expense",
		Amount:          testNumeric(0, 0),
		TransactionDate: testDate(2026, time.September, 16),
		Title:           "Invalid amount",
		ClientRequestID: 11,
		RequestHash:     "invalid-amount-hash",
		ActorUserID:     actorID,
	})
	assertPostgresCode(t, err, "23514")

	deleted, err := queries.SoftDeleteTransaction(ctx, db.SoftDeleteTransactionParams{
		ID: transactionID, UserID: ownerID, ActorUserID: actorID, ExpectedVersion: 1,
	})
	if err != nil {
		t.Fatalf("soft delete transaction: %v", err)
	}
	if !deleted.IsDelete || deleted.Version != 2 {
		t.Fatalf("unexpected deleted transaction: %+v", deleted)
	}
	if _, err := queries.GetActiveTransaction(ctx, db.GetActiveTransactionParams{ID: transactionID, UserID: ownerID}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("deleted transaction remained active: %v", err)
	}
	trash, err := queries.ListDeletedTransactions(ctx, db.ListDeletedTransactionsParams{UserID: ownerID, StartDate: testDate(2026, time.January, 1), EndDate: testDate(2026, time.December, 31), PageSize: 10})
	if err != nil || len(trash) != 1 {
		t.Fatalf("list transaction trash: len=%d err=%v", len(trash), err)
	}
	if _, err := queries.RestoreTransaction(ctx, db.RestoreTransactionParams{
		ID: transactionID, UserID: ownerID, ActorUserID: actorID, ExpectedVersion: 1,
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("stale restore should affect no row: %v", err)
	}
	restored, err := queries.RestoreTransaction(ctx, db.RestoreTransactionParams{
		ID: transactionID, UserID: ownerID, ActorUserID: actorID, ExpectedVersion: 2,
	})
	if err != nil {
		t.Fatalf("restore transaction: %v", err)
	}
	if restored.IsDelete || restored.Version != 3 {
		t.Fatalf("unexpected restored transaction: %+v", restored)
	}

	if _, err := queries.CreateCategory(ctx, db.CreateCategoryParams{
		UserID: ownerID, Type: "expense", Name: "food",
	}); err == nil {
		t.Fatal("case-insensitive duplicate category name was accepted")
	} else {
		assertPostgresCode(t, err, "23505")
	}

	event, err := queries.InsertAdminAccessEvent(ctx, db.InsertAdminAccessEventParams{
		ActorUserID:  actorID,
		TargetUserID: &ownerID,
		ResourceType: "transaction",
		ResourceID:   &transactionID,
		Action:       "restore",
		Outcome:      "success",
		RequestID:    "request-14",
		SafeMetadata: []byte(`{"fields":["isDelete"]}`),
	})
	if err != nil {
		t.Fatalf("insert audit event: %v", err)
	}
	if event.ActorUserID != actorID || event.TargetUserID == nil || *event.TargetUserID != ownerID {
		t.Fatalf("audit actor/target attribution was not preserved: %+v", event)
	}

	familyID := int64(16)
	if _, err := queries.CreateRefreshSession(ctx, db.CreateRefreshSessionParams{
		UserID:    ownerID,
		FamilyID:  &familyID,
		TokenHash: "refresh-token-hash",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
	}); err != nil {
		t.Fatalf("create refresh session: %v", err)
	}
}

func openMigratedTestPool(t *testing.T) *pgxpool.Pool {
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
		t.Fatalf("create schema suffix: %v", err)
	}
	schema := "monelog_test_" + hex.EncodeToString(random)
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminPool.Exec(context.Background(), "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("open schema-scoped pool: %v", err)
	}
	t.Cleanup(pool.Close)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve integration test path")
	}
	migrationDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "db", "migrations")
	migrationPaths, err := filepath.Glob(filepath.Join(migrationDir, "*.up.sql"))
	if err != nil {
		t.Fatalf("find migrations: %v", err)
	}
	if len(migrationPaths) == 0 {
		t.Fatal("no up migrations found")
	}
	sort.Strings(migrationPaths)
	for _, migrationPath := range migrationPaths {
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", filepath.Base(migrationPath), err)
		}
		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatalf("apply migration %s: %v", filepath.Base(migrationPath), err)
		}
	}

	return pool
}

func testNumeric(value int64, exponent int32) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(value), Exp: exponent, Valid: true}
}

func testDate(year int, month time.Month, day int) pgtype.Date {
	return pgtype.Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC), Valid: true}
}

func assertPostgresCode(t *testing.T, err error, code string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != code {
		t.Fatalf("expected PostgreSQL error %s, got %v", code, err)
	}
}
