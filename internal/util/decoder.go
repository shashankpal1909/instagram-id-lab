package util

import (
	"fmt"
	"time"
)

// Decode unpacks a 64-bit ID into its original components and prints them.
// It reverses the bitwise shifts used during ID generation:
//   - The lowest 10 bits represent the shard ID.
//   - The next 13 bits represent the sequence.
//   - The remaining upper bits represent the timestamp.
func Decode(id int64) {

	shard := id & 1023
	sequence := (id >> 10) & 8191
	timestamp := id >> 23
	
	// Add custom epoch back
	epoch := int64(1704067200000) // 2024-01-01T00:00:00Z in milliseconds
	timeObj := time.UnixMilli(timestamp + epoch)

	fmt.Printf("Decoded ID: %d\n", id)
	fmt.Printf("  Timestamp: %d (%s)\n", timestamp, timeObj.Format(time.RFC3339))
	fmt.Printf("  Sequence:  %d\n", sequence)
	fmt.Printf("  Shard ID:  %d\n", shard)
}
