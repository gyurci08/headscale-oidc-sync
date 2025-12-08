BINARY_NAME  ?= app
CMD_PATH     ?= ./
OUTPUT_DIR   ?= bin

.DEFAULT_GOAL := help

# ==============================================================================
#  Development Cycle Commands
# ==============================================================================

.PHONY: all
all: clean build run ## Cleans previous builds, rebuilds, runs the application.

.PHONY: build
build: ## Compiles the Go application into a static Linux binary.
	@echo "==> Compiling application..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(OUTPUT_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build successful! Binary at: $(OUTPUT_DIR)/$(BINARY_NAME)"

.PHONY: test
test: ## Runs all tests in the project with the race detector.
	@echo "==> Running tests..."
	go test -v -race ./...

.PHONY: run
run: build ## Starts the application (host binary).
	@echo "==> Starting application (Binary)..."
	$(OUTPUT_DIR)/$(BINARY_NAME)

# ==============================================================================
#  Docker / Compose
# ==============================================================================

.PHONY: docker-build
docker-build: ## Builds the application's Docker image.
	@echo "==> Building Docker image..."
	docker compose build app

.PHONY: docker-up
docker-up: ## Starts the application and related services using Docker Compose.
	@echo "==> Starting application (Docker Compose)..."
	docker compose up --build -d

.PHONY: docker-down
docker-down: ## Stops services started by Docker Compose.
	@echo "==> Stopping application..."
	docker compose down

.PHONY: docker-logs
docker-logs: ## Follows the application container logs.
	@echo "==> Following logs..."
	docker compose logs -f app

# ==============================================================================
#  Maintenance
# ==============================================================================

.PHONY: clean
clean: ## Removes files generated during the build.
	@echo "==> Cleaning..."
	@rm -rf $(OUTPUT_DIR)

.PHONY: lint
lint: ## Runs golangci-lint to check code quality.
	@echo "==> Checking code quality (lint)..."
	@golangci-lint run

.PHONY: deps
deps: ## Downloads and tidies Go module dependencies.
	@echo "==> Managing dependencies..."
	go mod tidy
	go mod download

# ==============================================================================
#  Help
# ==============================================================================

.PHONY: help
help: ## Displays this help.
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)