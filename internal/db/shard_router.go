package db

import (
	"fmt"

	"instagram-id-lab/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Router maintains the mapping from logical shards to physical databases
type Router struct {
	shards map[int]*model.Shard
}

// NewRouter creates a Router with 8 logical shards distributed across 3 physical pools
func NewRouter(pg1, pg2, pg3 *pgxpool.Pool) *Router {
	r := &Router{
		shards: make(map[int]*model.Shard),
	}

	// PG1: shards 1, 2, 3
	for i := 1; i <= 3; i++ {
		r.shards[i] = &model.Shard{
			ID:       i,
			Name:     fmt.Sprintf("shard_%d", i),
			Sequence: fmt.Sprintf("shard_%d_seq", i),
			DB:       pg1,
		}
	}

	// PG2: shards 4, 5, 6
	for i := 4; i <= 6; i++ {
		r.shards[i] = &model.Shard{
			ID:       i,
			Name:     fmt.Sprintf("shard_%d", i),
			Sequence: fmt.Sprintf("shard_%d_seq", i),
			DB:       pg2,
		}
	}

	// PG3: shards 7, 8
	for i := 7; i <= 8; i++ {
		r.shards[i] = &model.Shard{
			ID:       i,
			Name:     fmt.Sprintf("shard_%d", i),
			Sequence: fmt.Sprintf("shard_%d_seq", i),
			DB:       pg3,
		}
	}

	return r
}

// ShardForUser determines which logical shard a given userID belongs to.
// It uses a modulo operation to evenly distribute users across the 8 logical shards.
func (r *Router) ShardForUser(userID int64) *model.Shard {
	shardID := int(userID%8) + 1
	return r.shards[shardID]
}
