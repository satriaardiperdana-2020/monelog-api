package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/satriaardiperdana-2020/monelog-api/internal/repository/sqlc"
)

// Templates provides owner-scoped access to reusable transaction templates.
type Templates struct {
	queries *db.Queries
}

func NewTemplates(pool *pgxpool.Pool) *Templates {
	return &Templates{queries: db.New(pool)}
}

func (r *Templates) OwnerIsActive(ctx context.Context, ownerID int64) error {
	_, err := r.queries.GetActiveUserByID(ctx, ownerID)
	return err
}

func (r *Templates) List(ctx context.Context, ownerID int64, deleted bool) ([]db.TransactionTemplate, error) {
	if deleted {
		return r.queries.ListDeletedTransactionTemplates(ctx, ownerID)
	}
	return r.queries.ListActiveTransactionTemplates(ctx, ownerID)
}

func (r *Templates) Get(ctx context.Context, id, ownerID int64, deleted bool) (db.TransactionTemplate, error) {
	if deleted {
		return r.queries.GetDeletedTransactionTemplate(ctx, db.GetDeletedTransactionTemplateParams{ID: id, UserID: ownerID})
	}
	return r.queries.GetActiveTransactionTemplate(ctx, db.GetActiveTransactionTemplateParams{ID: id, UserID: ownerID})
}

func (r *Templates) Queries() *db.Queries { return r.queries }
