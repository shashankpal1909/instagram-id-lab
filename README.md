# Instagram ID Generation Lab

This project simulates how Instagram historically generated distributed IDs using PostgreSQL sequences, logical shards, and physical database shards.

## Architecture

The system uses 3 PostgreSQL instances running via Docker Compose to represent physical database shards:
- **PG1** (port 5433): Logical shards 1, 2, 3
- **PG2** (port 5434): Logical shards 4, 5, 6
- **PG3** (port 5435): Logical shards 7, 8

### Routing
Users are routed to a logical shard using the formula:
```go
shardID := int(userID%8) + 1
```

### ID Generation
Each logical shard owns its own PostgreSQL sequence. IDs are 64-bit integers generated as follows:
- **41 bits**: Timestamp (milliseconds since custom epoch to fit in 41 bits without negative sign)
- **13 bits**: Sequence (auto-incrementing PostgreSQL sequence, modulo 8192)
- **10 bits**: Logical Shard ID (modulo 1024)

This ensures highly concurrent inserts without collisions and guarantees time-sortable IDs.

## Requirements
- Docker and Docker Compose
- Go 1.22+

## Setup & Running

1. **Start the databases**
   ```bash
   make up
   ```
   This will bring up 3 PostgreSQL instances and initialize the schemas automatically.

2. **Run the simulation**
   ```bash
   make run
   ```
   This command starts 100 concurrent workers, each performing 10,000 inserts (1,000,000 total). The program validates that no duplicate IDs were generated, prints a per-DB and per-shard breakdown of the generated IDs, and decodes a sample ID.

3. **Reset the databases**
   If you want to clear the databases and start fresh:
   ```bash
   make reset
   ```

4. **Stop the databases**
   ```bash
   make down
   ```
