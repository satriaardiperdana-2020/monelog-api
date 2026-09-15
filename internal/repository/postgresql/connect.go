package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a lazy PostgreSQL connection pool. Availability is reported
// by the readiness endpoint rather than making a temporary outage fatal.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("invalid PostgreSQL configuration")
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, errors.New("could not create PostgreSQL pool")
	}
	return pool, nil
}
