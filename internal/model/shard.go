package model

import "github.com/jackc/pgx/v5/pgxpool"

// Shard represents a logical shard mapping to a specific PostgreSQL connection pool and sequence.
type Shard struct {
	ID       int
	Name     string
	Sequence string
	DB       *pgxpool.Pool
}
