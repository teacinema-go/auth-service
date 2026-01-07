-include .env
export

MIGRATIONS_DIR=internal/database/migrations

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build:
	go build -o bin/auth-service cmd/auth-service/main.go

.PHONY: run
run:
	go run cmd/auth-service/main.go

.PHONY: migrate-create
migrate-create: ## Usage: make migrate-create NAME=init
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required"; \
		exit 1; \
	fi
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

.PHONY: migrate-up
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_NAME) sslmode=disable" up
	@echo "Migrations applied successfully"

.PHONY: migrate-down
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_NAME) sslmode=disable" down
	@echo "Migration rolled back"

.PHONY: migrate-status
migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "host=$(POSTGRES_HOST) port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_NAME) sslmode=disable" status

.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate
	@echo "sqlc code generated successfully"