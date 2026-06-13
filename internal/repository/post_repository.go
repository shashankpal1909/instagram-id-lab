package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GenerateID generates a 64-bit ID using Instagram's distributed ID generation strategy:
//
//   - 41 bits: Timestamp in milliseconds (using a custom epoch to prevent overflow)
//   - 13 bits: Logical Shard ID
//   - 10 bits: Auto-incrementing sequence from PostgreSQL
//
// This structure guarantees that IDs are time-sortable and globally unique across shards.
func GenerateID(ctx context.Context, db *pgxpool.Pool, sequence string, shardID int) (int64, error) {
	// 1. Fetch sequence from PostgreSQL
	query := fmt.Sprintf("SELECT nextval('%s')", sequence)
	var seq int64
	err := db.QueryRow(ctx, query).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("failed to get sequence: %w", err)
	}

	// 2. Build ID
	// Use a custom epoch (e.g., Jan 1, 2024) to keep the timestamp small enough
	// to fit in 41 bits without setting the sign bit of the int64
	epoch := int64(1704067200000) // 2024-01-01T00:00:00Z in milliseconds
	timestamp := time.Now().UnixMilli() - epoch

	// Modulo sequence to prevent overflow into timestamp bits (2^13 = 8192)
	// Modulo shardID to prevent overflow into sequence bits (2^10 = 1024)
	seqComponent := seq % 8192
	shardComponent := shardID % 1024

	// Combine the components into a single 64-bit integer.
	// The timestamp takes the highest 41 bits.
	// The sequence takes the next 13 bits.
	// The shard ID takes the lowest 10 bits.
	id := (timestamp << 23) | (seqComponent << 10) | int64(shardComponent)

	return id, nil
}

// InsertPost inserts a row into the posts table on the appropriate physical database
func InsertPost(ctx context.Context, db *pgxpool.Pool, id int64, userID int64, shardID int, content string) error {
	query := `
		INSERT INTO posts (id, user_id, shard_id, content)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(ctx, query, id, userID, shardID, content)
	if err != nil {
		return fmt.Errorf("failed to insert post: %w", err)
	}
	return nil
}
