# Load environment variables from .env file
include .env
export

# Default target - shows available commands
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make run         - Run the application"
	@echo "  make migrate     - Run database migrations"
	@echo "  make migrate-down - Rollback last migration"
	@echo "  make migrate-force - Force migration version (use with VERSION=x)"
	@echo "  make migrate-goto - Migrate to specific version (use with VERSION=x)"
	@echo "  make migrate-drop - Drop everything in database (careful!)"
	@echo "  make migrate-create - Create new migration (use with NAME=migration_name)"
	@echo "  make db-create   - Create databases"
	@echo "  make db-drop     - Drop databases (careful!)"
	@echo "  make build       - Build the application"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"
	@echo "  make swagger     - Generate Swagger documentation"
	@echo "  make run-swagger - Generate docs and run application"
	@echo "  make install-swagger - Install swagger CLI tool"

# Run the application
.PHONY: run
run:
	@echo "Starting application..."
	go run cmd/api/main.go

# Create databases
.PHONY: db-create
db-create:
	@echo "Creating databases..."
	@echo "Using connection: $(DATABASE_URL)"
	@psql $(DATABASE_URL) -c "SELECT 1" > /dev/null 2>&1 || \
		(psql postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=$(DB_SSLMODE) -c "CREATE DATABASE $(DB_NAME);" && \
		 psql postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=$(DB_SSLMODE) -c "CREATE DATABASE $(DB_NAME)_test;")
	@echo "Databases created"

# Drop databases (be careful!)
.PHONY: db-drop
db-drop:
	@echo "Dropping databases..."
	@read -p "Are you sure? This will delete all data! [y/N] " confirm; \
	if [ "$$confirm" = "y" ]; then \
		psql postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=$(DB_SSLMODE) -c "DROP DATABASE IF EXISTS $(DB_NAME);" && \
		psql postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=$(DB_SSLMODE) -c "DROP DATABASE IF EXISTS $(DB_NAME)_test;" && \
		echo "Databases dropped"; \
	else \
		echo "Operation cancelled"; \
	fi

# Build the application
.PHONY: build
build:
	@echo "Building application..."
	go build -o bin/todo-api cmd/api/main.go
	@echo "Build complete: bin/todo-api"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf docs/
	go clean -cache

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Install development dependencies
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Tools installed"

# ==================== SWAGGER COMMANDS ====================

# Generate swagger documentation
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs
	@echo "Swagger docs generated successfully"

# Generate swagger and run
.PHONY: run-swagger
run-swagger: swagger run

# Install swagger if not present
.PHONY: install-swagger
install-swagger:
	@echo "Installing swagger..."
	@go install github.com/swaggo/swag/cmd/swag@latest

# ==================== MIGRATION COMMANDS ====================

# Run all pending migrations
.PHONY: migrate
migrate:
	@echo "Running migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

# Rollback the last migration
.PHONY: migrate-down
migrate-down:
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

# Rollback all migrations
.PHONY: migrate-down-all
migrate-down-all:
	@echo "Rolling back all migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down

# Force set migration version (use with VERSION=x)
.PHONY: migrate-force
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Please provide VERSION: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(VERSION)..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(VERSION)

# Migrate to a specific version (use with VERSION=x)
.PHONY: migrate-goto
migrate-goto:
	@if [ -z "$(VERSION)" ]; then \
		echo "Please provide VERSION: make migrate-goto VERSION=1"; \
		exit 1; \
	fi
	@echo "Migrating to version $(VERSION)..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" goto $(VERSION)

# Show current migration version
.PHONY: migrate-version
migrate-version:
	@echo "Current migration version:"
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" version

# Drop everything in the database
.PHONY: migrate-drop
migrate-drop:
	@echo "WARNING: This will drop everything in the database!"
	@read -p "Are you sure? [y/N] " confirm; \
	if [ "$$confirm" = "y" ]; then \
		migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" drop -f; \
		echo "Database dropped"; \
	else \
		echo "Operation cancelled"; \
	fi

# Create a new migration (sequential numbering)
.PHONY: migrate-create
migration:
	@if [ -z "$(NAME)" ]; then \
		echo "Please provide a NAME: make migration NAME=your_migration_NAME"; \
		exit 1; \
	fi
	@# Find the highest migration number
	@last_migration=$$(ls -1 migrations/*.up.sql 2>/dev/null | sed 's/.*\/\([0-9]*\)_.*/\1/' | sort -n | tail -1); \
	if [ -z "$$last_migration" ]; then \
		next_number="001"; \
	else \
		next_number=$$(printf "%03d" $$(($$last_migration + 1))); \
	fi; \
	touch migrations/$${next_number}_$(NAME).up.sql; \
	touch migrations/$${next_number}_$(NAME).down.sql; \
	echo "Created migration files:"; \
	echo "  migrations/$${next_number}_$(NAME).up.sql"; \
	echo "  migrations/$${next_number}_$(NAME).down.sql"