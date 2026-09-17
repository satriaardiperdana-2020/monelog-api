package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

// Categories is the PostgreSQL repository for owner-scoped categories.
type Categories struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewCategories(pool *pgxpool.Pool) *Categories {
	return &Categories{pool: pool, queries: db.New(pool)}
}

func (r *Categories) OwnerIsActive(ctx context.Context, ownerID uuid.UUID) error {
	_, err := r.queries.GetActiveUserByID(ctx, pgUUID(ownerID))
	return err
}

func (r *Categories) Create(ctx context.Context, id, ownerID uuid.UUID, categoryType, name string) (db.Category, error) {
	return r.queries.CreateCategory(ctx, db.CreateCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), Type: categoryType, Name: name})
}

func (r *Categories) Get(ctx context.Context, id, ownerID uuid.UUID, deleted bool) (db.Category, error) {
	params := db.GetActiveCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID)}
	if deleted {
		return r.queries.GetDeletedCategory(ctx, db.GetDeletedCategoryParams(params))
	}
	return r.queries.GetActiveCategory(ctx, params)
}

func (r *Categories) List(ctx context.Context, ownerID uuid.UUID, deleted bool) ([]db.Category, error) {
	if deleted {
		return r.queries.ListDeletedCategories(ctx, pgUUID(ownerID))
	}
	return r.queries.ListActiveCategories(ctx, pgUUID(ownerID))
}

func (r *Categories) Update(ctx context.Context, id, ownerID uuid.UUID, name string, version int32) (db.Category, error) {
	return r.queries.UpdateCategory(ctx, db.UpdateCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), Name: name, ExpectedVersion: version})
}

func (r *Categories) Archive(ctx context.Context, id, ownerID uuid.UUID, version int32) (db.Category, error) {
	return r.queries.SoftDeleteCategory(ctx, db.SoftDeleteCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), ExpectedVersion: version})
}

func (r *Categories) Restore(ctx context.Context, id, ownerID uuid.UUID, version int32) (db.Category, error) {
	return r.queries.RestoreCategory(ctx, db.RestoreCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), ExpectedVersion: version})
}

func (r *Categories) CreateAdmin(ctx context.Context, id, actorID, ownerID uuid.UUID, categoryType, name string) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, id, "create", func(q *db.Queries) (db.Category, error) {
		return q.CreateCategory(ctx, db.CreateCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), Type: categoryType, Name: name})
	})
}

func (r *Categories) UpdateAdmin(ctx context.Context, id, actorID, ownerID uuid.UUID, name string, version int32) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, id, "update", func(q *db.Queries) (db.Category, error) {
		return q.UpdateCategory(ctx, db.UpdateCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), Name: name, ExpectedVersion: version})
	})
}

func (r *Categories) ArchiveAdmin(ctx context.Context, id, actorID, ownerID uuid.UUID, version int32) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, id, "delete", func(q *db.Queries) (db.Category, error) {
		return q.SoftDeleteCategory(ctx, db.SoftDeleteCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), ExpectedVersion: version})
	})
}

func (r *Categories) RestoreAdmin(ctx context.Context, id, actorID, ownerID uuid.UUID, version int32) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, id, "restore", func(q *db.Queries) (db.Category, error) {
		return q.RestoreCategory(ctx, db.RestoreCategoryParams{ID: pgUUID(id), UserID: pgUUID(ownerID), ExpectedVersion: version})
	})
}

func (r *Categories) withAdminWrite(ctx context.Context, actorID, ownerID, resourceID uuid.UUID, action string, write func(*db.Queries) (db.Category, error)) (db.Category, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Category{}, fmt.Errorf("begin category admin write: %w", err)
	}
	defer tx.Rollback(ctx)
	q := r.queries.WithTx(tx)
	item, err := write(q)
	if err != nil {
		return db.Category{}, err
	}
	if _, err := q.InsertAdminAccessEvent(ctx, db.InsertAdminAccessEventParams{
		ID: uuidToPG(uuid.New()), ActorUserID: pgUUID(actorID), TargetUserID: pgUUID(ownerID), ResourceType: "category", ResourceID: pgUUID(resourceID), Action: action, Outcome: "success", RequestID: uuid.NewString(), SafeMetadata: []byte(`{}`),
	}); err != nil {
		return db.Category{}, fmt.Errorf("audit category admin write: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return db.Category{}, fmt.Errorf("commit category admin write: %w", err)
	}
	return item, nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func uuidToPG(value uuid.UUID) pgtype.UUID { return pgUUID(value) }
