package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"instagram-id-lab/internal/db"
	"instagram-id-lab/internal/util"
	"instagram-id-lab/internal/worker"
)

const (
	numWorkers = 100
	numInserts = 10000
)

func main() {
	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		cancel()
	}()

	log.Println("Connecting to PostgreSQL instances...")

	// Establish connections to all 3 PostgreSQL instances representing the physical database shards.
	pg1, err := db.Connect(ctx, "postgres://postgres:postgres@localhost:5433/instagram")
	if err != nil {
		log.Fatalf("Failed to connect to PG1: %v", err)
	}
	defer pg1.Close()

	pg2, err := db.Connect(ctx, "postgres://postgres:postgres@localhost:5434/instagram")
	if err != nil {
		log.Fatalf("Failed to connect to PG2: %v", err)
	}
	defer pg2.Close()

	pg3, err := db.Connect(ctx, "postgres://postgres:postgres@localhost:5435/instagram")
	if err != nil {
		log.Fatalf("Failed to connect to PG3: %v", err)
	}
	defer pg3.Close()

	log.Println("Connected to all databases.")

	// Initialize the logical shard router with the active database connections.
	// This maps the 8 logical shards evenly across the 3 physical instances.
	router := db.NewRouter(pg1, pg2, pg3)

	log.Printf("Starting %d workers, each performing %d inserts...", numWorkers, numInserts)

	// Launch a pool of concurrent workers to simulate high insert loads.
	results := make(chan int64, numWorkers*numInserts)
	errs := make(chan error, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker.RunWorker(ctx, router, numInserts, results, errs)
		}()
	}

	// Wait for workers in a separate goroutine to close the results channel
	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// Collect all generated IDs and scan for potential duplicates.
	// We use a map to enforce uniqueness strictly.
	idMap := make(map[int64]struct{})
	duplicateCount := 0
	var sampleID int64
	totalGenerated := 0
	
	// Track per shard counts
	shardCounts := make(map[int]int)

	for id := range results {
		if _, exists := idMap[id]; exists {
			duplicateCount++
			log.Printf("Duplicate ID detected: %d", id)
			// The simulator should strictly avoid duplicates. Any collision is considered a fatal system failure.
			log.Fatalf("Duplicate ID generated: %d! Failing immediately.", id)
		}
		idMap[id] = struct{}{}
		sampleID = id
		totalGenerated++
		
		// Extract shard from ID
		shard := int(id & 1023)
		shardCounts[shard]++
	}

	// Check for any errors during worker execution
	for err := range errs {
		if err != nil {
			log.Fatalf("Worker encountered error: %v", err)
		}
	}

	// Print a comprehensive summary of the test run.
	fmt.Println("-------------------------------------------------")
	fmt.Printf("Total IDs generated: %d\n", totalGenerated)
	fmt.Printf("Duplicate count:     %d\n", duplicateCount)
	fmt.Println("-------------------------------------------------")
	
	fmt.Println("Per DB / Per Shard Summary:")
	fmt.Println("PG1 (Shards 1, 2, 3):")
	for i := 1; i <= 3; i++ {
		fmt.Printf("  Shard %d: %d inserts\n", i, shardCounts[i])
	}
	fmt.Println("PG2 (Shards 4, 5, 6):")
	for i := 4; i <= 6; i++ {
		fmt.Printf("  Shard %d: %d inserts\n", i, shardCounts[i])
	}
	fmt.Println("PG3 (Shards 7, 8):")
	for i := 7; i <= 8; i++ {
		fmt.Printf("  Shard %d: %d inserts\n", i, shardCounts[i])
	}
	fmt.Println("-------------------------------------------------")

	if totalGenerated > 0 {
		fmt.Println("Sample decoded ID:")
		util.Decode(sampleID)
	}
	fmt.Println("-------------------------------------------------")
	fmt.Println("Run completed successfully!")
}
