package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

func (r *Categories) OwnerIsActive(ctx context.Context, ownerID int64) error {
	_, err := r.queries.GetActiveUserByID(ctx, ownerID)
	return err
}

func (r *Categories) Create(ctx context.Context, ownerID int64, categoryType, name string) (db.Category, error) {
	return r.queries.CreateCategory(ctx, db.CreateCategoryParams{UserID: ownerID, Type: categoryType, Name: name})
}

func (r *Categories) Get(ctx context.Context, id, ownerID int64, deleted bool) (db.Category, error) {
	params := db.GetActiveCategoryParams{ID: id, UserID: ownerID}
	if deleted {
		return r.queries.GetDeletedCategory(ctx, db.GetDeletedCategoryParams(params))
	}
	return r.queries.GetActiveCategory(ctx, params)
}

func (r *Categories) List(ctx context.Context, ownerID int64, deleted bool) ([]db.Category, error) {
	if deleted {
		return r.queries.ListDeletedCategories(ctx, ownerID)
	}
	return r.queries.ListActiveCategories(ctx, ownerID)
}

func (r *Categories) Update(ctx context.Context, id, ownerID int64, name string, version int32) (db.Category, error) {
	return r.queries.UpdateCategory(ctx, db.UpdateCategoryParams{ID: id, UserID: ownerID, Name: name, ExpectedVersion: version})
}

func (r *Categories) Archive(ctx context.Context, id, ownerID int64, version int32) (db.Category, error) {
	return r.queries.SoftDeleteCategory(ctx, db.SoftDeleteCategoryParams{ID: id, UserID: ownerID, ExpectedVersion: version})
}

func (r *Categories) Restore(ctx context.Context, id, ownerID int64, version int32) (db.Category, error) {
	return r.queries.RestoreCategory(ctx, db.RestoreCategoryParams{ID: id, UserID: ownerID, ExpectedVersion: version})
}

func (r *Categories) CreateAdmin(ctx context.Context, actorID, ownerID int64, categoryType, name string) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, "create", func(q *db.Queries) (db.Category, error) {
		return q.CreateCategory(ctx, db.CreateCategoryParams{UserID: ownerID, Type: categoryType, Name: name})
	})
}

func (r *Categories) UpdateAdmin(ctx context.Context, id, actorID, ownerID int64, name string, version int32) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, "update", func(q *db.Queries) (db.Category, error) {
		return q.UpdateCategory(ctx, db.UpdateCategoryParams{ID: id, UserID: ownerID, Name: name, ExpectedVersion: version})
	})
}

func (r *Categories) ArchiveAdmin(ctx context.Context, id, actorID, ownerID int64, version int32) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, "delete", func(q *db.Queries) (db.Category, error) {
		return q.SoftDeleteCategory(ctx, db.SoftDeleteCategoryParams{ID: id, UserID: ownerID, ExpectedVersion: version})
	})
}

func (r *Categories) RestoreAdmin(ctx context.Context, id, actorID, ownerID int64, version int32) (db.Category, error) {
	return r.withAdminWrite(ctx, actorID, ownerID, "restore", func(q *db.Queries) (db.Category, error) {
		return q.RestoreCategory(ctx, db.RestoreCategoryParams{ID: id, UserID: ownerID, ExpectedVersion: version})
	})
}

func (r *Categories) withAdminWrite(ctx context.Context, actorID, ownerID int64, action string, write func(*db.Queries) (db.Category, error)) (db.Category, error) {
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
		ActorUserID: actorID, TargetUserID: &ownerID, ResourceType: "category", ResourceID: &item.ID, Action: action, Outcome: "success", RequestID: uuid.NewString(), SafeMetadata: []byte(`{}`),
	}); err != nil {
		return db.Category{}, fmt.Errorf("audit category admin write: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return db.Category{}, fmt.Errorf("commit category admin write: %w", err)
	}
	return item, nil
}
