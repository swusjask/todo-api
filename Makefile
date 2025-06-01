# Default target - shows available commands
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make run         - Run the application"
	@echo "  make test        - Run all tests"
	@echo "  make test-verbose - Run tests with detailed output"
	@echo "  make migrate     - Run database migrations"
	@echo "  make migrate-down - Rollback last migration"
	@echo "  make db-create   - Create databases"
	@echo "  make db-drop     - Drop databases (careful!)"
	@echo "  make build       - Build the application"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"

# Run the application
.PHONY: run
run:
	@echo "Starting application..."
	go run cmd/api/main.go

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Run tests with verbose output
.PHONY: test-verbose
test-verbose:
	@echo "Running tests with detailed output..."
	go test -v -cover -race ./...

# Run only integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration

# Create databases
.PHONY: db-create
db-create:
	@echo "Creating databases..."
	psql -U postgres -c "CREATE DATABASE todos;" || true
	psql -U postgres -c "CREATE DATABASE todos_test;" || true
	@echo "Databases created"

# Drop databases (be careful!)
.PHONY: db-drop
db-drop:
	@echo "Dropping databases..."
	@read -p "Are you sure? This will delete all data! [y/N] " confirm; \
	if [ "$$confirm" = "y" ]; then \
		psql -U postgres -c "DROP DATABASE IF EXISTS todos;" && \
		psql -U postgres -c "DROP DATABASE IF EXISTS todos_test;" && \
		echo "Databases dropped"; \
	else \
		echo "Operation cancelled"; \
	fi

# Run migrations
.PHONY: migrate
migrate:
	@echo "Running migrations..."
	go run cmd/api/main.go -migrate-only

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
	@echo "Tools installed"

# Create a new migration
# Usage: make migration name=create_users_table
.PHONY: migration
migration:
	@if [ -z "$(name)" ]; then \
		echo "Please provide a name: make migration name=your_migration_name"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	touch migrations/$${timestamp}_$(name).up.sql; \
	touch migrations/$${timestamp}_$(name).down.sql; \
	echo "Created migration files:"; \
	echo "  migrations/$${timestamp}_$(name).up.sql"; \
	echo "  migrations/$${timestamp}_$(name).down.sql"