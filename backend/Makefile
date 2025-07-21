# Project configuration
CMD_DIR := cmd
BIN_DIR := bin
LOG_DIR := logs
DB_PATH := internal/db/db.db
MIGRATIONS_DIR := internal/db/migrations
SCHEMA_PATH := internal/db/schema.sql

.PHONY: setup-env build-all build-% db-generate db-reset db-init db-migrate-create db-migrate-up db-migrate-down clean

# Setup environment (prints export statements for eval)
setup-env:
	@echo "export ENV_PATH=$(CURDIR)/.env"
	@echo "export BIN_DIR=$(CURDIR)/$(BIN_DIR)"
	@echo "export LOG_DIR=$(CURDIR)/$(LOG_DIR)"

# Build specific binary (e.g., make build-create-mcp-client-instance)
build-%:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$* ./$(CMD_DIR)/$*/main.go

# Build all core binaries
build-all:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/create-mcp-client-instance ./$(CMD_DIR)/createMCPClientInstance

# Generate code from SQLC
db-generate:
	sqlc generate

# Reset database: delete, migrate up, generate code, init data
db-reset:
	rm -f $(DB_PATH)
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up
	sqlc generate
	# go run $(CMD_DIR)/initDB/main.go

# Init data (no schema changes)
db-init:
	go run $(CMD_DIR)/initDB/main.go

# Create a new migration: make db-migrate-create name=create_users_table
db-migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: you must provide a name, e.g. make db-migrate-create name=create_users_table"; \
		exit 1; \
	else \
		goose -dir $(MIGRATIONS_DIR) create $(name) sql; \
	fi

# Migrate up
db-migrate-up:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

# Migrate down
db-migrate-down:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) down

# Remove built binaries
clean:
	rm -f $(BIN_DIR)/*
