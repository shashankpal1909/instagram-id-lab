.PHONY: up down run reset

up:
	docker compose up -d
	@echo "Waiting for databases to be ready..."
	@sleep 5

down:
	docker compose down -v

reset: down up

run:
	go run ./cmd
