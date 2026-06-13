package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a new pgxpool.Pool connected to the given database URL.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
