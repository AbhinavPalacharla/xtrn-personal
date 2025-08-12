# Project configuration
CMD_DIR := cmd
BIN_DIR := bin
SCRIPTS_DIR := scripts
LOG_DIR := logs
DB_PATH := internal/db/db.db
MIGRATIONS_DIR := internal/db/migrations

.PHONY: setup-env

setup-env:
	@echo "export ENV_PATH=$(CURDIR)/.env"
	@echo "export BIN_DIR=$(CURDIR)/$(BIN_DIR)"
	@echo "export LOG_DIR=$(CURDIR)/$(LOG_DIR)"

build-start-mcp-instance:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/start-mcp-instance ./$(CMD_DIR)/start-mcp-instance/main.go

# Remove built binaries
clean:
	rm -f $(BIN_DIR)/*

######################## SCRIPTS ########################

delete-mcp-instances:
	go run $(SCRIPTS_DIR)/delete_mcp_instances/main.go

init-mcp-instances:
	go run $(SCRIPTS_DIR)/init_mcp_server_instances/main.go

init-oauth-providers:
	go run $(SCRIPTS_DIR)/init_oauth_providers/main.go

init-mcp-images:
	go run $(SCRIPTS_DIR)/init_mcp_server_images/main.go

######################## DATABASE ########################

db-generate:
	sqlc generate

# Reset database: delete, migrate up, generate code, init data
db-reset:
	rm -f $(DB_PATH)
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up
	sqlc generate

# Create a new migration: make db-migrate-create name=create_users_table
db-migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: you must provide a name, e.g. make db-migrate-create name=<migration_name>"; \
		exit 1; \
	else \
		goose -dir $(MIGRATIONS_DIR) create $(name) sql; \
	fi

db-migrate-up:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

db-migrate-down:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) down
