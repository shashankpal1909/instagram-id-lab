package worker

import (
	"context"
	"fmt"
	"math/rand"

	"instagram-id-lab/internal/db"
	"instagram-id-lab/internal/repository"
)

// RunWorker simulates a client making multiple concurrent insert requests
func RunWorker(ctx context.Context, router *db.Router, numInserts int, results chan<- int64, errs chan<- error) {
	for i := 0; i < numInserts; i++ {
		// 1. Generate random user id
		userID := rand.Int63n(1_000_000)

		// 2. Resolve logical shard
		shard := router.ShardForUser(userID)

		// 3. Generate ID
		id, err := repository.GenerateID(ctx, shard.DB, shard.Sequence, shard.ID)
		if err != nil {
			errs <- fmt.Errorf("failed to generate ID for user %d: %w", userID, err)
			return
		}

		content := fmt.Sprintf("Hello from user %d on shard %d!", userID, shard.ID)

		// 4. Insert row
		err = repository.InsertPost(ctx, shard.DB, id, userID, shard.ID, content)
		if err != nil {
			errs <- fmt.Errorf("failed to insert post for user %d: %w", userID, err)
			return
		}

		// 5. Send generated ID into a channel
		select {
		case <-ctx.Done():
			return
		case results <- id:
		}
	}
}
