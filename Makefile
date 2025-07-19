# Todo API Backend Makefile
# This Makefile provides common development tasks for the Todo API Backend project

# Variables
BINARY_NAME=server
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/server/main.go
DOCKER_IMAGE=todo-api-backend
DOCKER_TAG=latest
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Default target
.DEFAULT_GOAL := help

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

##@ Help
.PHONY: help
help: ## Display this help message
	@echo "$(BLUE)Todo API Backend - Available Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(GREEN)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: build
build: ## Build the application binary
	@echo "$(BLUE)Building application...$(NC)"
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)Build completed: $(BINARY_PATH)$(NC)"

.PHONY: run
run: ## Run the application locally
	@echo "$(BLUE)Starting application...$(NC)"
	$(GOCMD) run $(MAIN_PATH)

.PHONY: run-build
run-build: build ## Build and run the application
	@echo "$(BLUE)Running built application...$(NC)"
	./$(BINARY_PATH)

.PHONY: clean
clean: ## Clean build artifacts and temporary files
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf bin/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "$(GREEN)Clean completed$(NC)"

.PHONY: deps
deps: ## Download and install dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

.PHONY: deps-update
deps-update: ## Update all dependencies to latest versions
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated to latest versions$(NC)"

##@ Testing
.PHONY: test
test: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run unit tests (short mode)
	@echo "$(BLUE)Running unit tests (short mode)...$(NC)"
	$(GOTEST) -short -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_FILE)$(NC)"

.PHONY: test-coverage-html
test-coverage-html: test-coverage ## Generate HTML coverage report
	@echo "$(BLUE)Generating HTML coverage report...$(NC)"
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)HTML coverage report generated: $(COVERAGE_HTML)$(NC)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	$(GOTEST) -v ./tests/integration/...

.PHONY: test-all
test-all: test test-integration ## Run all tests (unit + integration)

.PHONY: benchmark
benchmark: ## Run benchmark tests
	@echo "$(BLUE)Running benchmark tests...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

##@ Code Quality
.PHONY: fmt
fmt: ## Format Go code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	@echo "$(GREEN)Code formatting completed$(NC)"

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "$(BLUE)Checking code formatting...$(NC)"
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "$(RED)Code is not formatted. Run 'make fmt' to fix.$(NC)"; \
		$(GOFMT) -l .; \
		exit 1; \
	else \
		echo "$(GREEN)Code is properly formatted$(NC)"; \
	fi

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
		echo "$(GREEN)Linting completed$(NC)"; \
	else \
		echo "$(YELLOW)golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)Vet completed$(NC)"

.PHONY: check
check: fmt-check vet lint ## Run all code quality checks

##@ Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

.PHONY: docker-run
docker-run: ## Run application in Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-run-detached
docker-run-detached: ## Run application in Docker container (detached)
	@echo "$(BLUE)Running Docker container in background...$(NC)"
	docker run -d --name todo-api -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "$(GREEN)Container started: todo-api$(NC)"

.PHONY: docker-stop
docker-stop: ## Stop running Docker container
	@echo "$(BLUE)Stopping Docker container...$(NC)"
	docker stop todo-api || true
	docker rm todo-api || true
	@echo "$(GREEN)Container stopped$(NC)"

.PHONY: docker-clean
docker-clean: ## Remove Docker image and containers
	@echo "$(BLUE)Cleaning Docker artifacts...$(NC)"
	docker stop todo-api || true
	docker rm todo-api || true
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	@echo "$(GREEN)Docker cleanup completed$(NC)"

##@ Docker Compose
.PHONY: compose-up
compose-up: ## Start services with docker-compose
	@echo "$(BLUE)Starting services with docker-compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Services started$(NC)"

.PHONY: compose-up-build
compose-up-build: ## Start services with docker-compose (rebuild images)
	@echo "$(BLUE)Starting services with docker-compose (rebuild)...$(NC)"
	docker-compose up -d --build
	@echo "$(GREEN)Services started with rebuild$(NC)"

.PHONY: compose-down
compose-down: ## Stop and remove docker-compose services
	@echo "$(BLUE)Stopping docker-compose services...$(NC)"
	docker-compose down
	@echo "$(GREEN)Services stopped$(NC)"

.PHONY: compose-down-volumes
compose-down-volumes: ## Stop services and remove volumes
	@echo "$(BLUE)Stopping services and removing volumes...$(NC)"
	docker-compose down -v
	@echo "$(GREEN)Services stopped and volumes removed$(NC)"

.PHONY: compose-logs
compose-logs: ## View docker-compose logs
	@echo "$(BLUE)Viewing docker-compose logs...$(NC)"
	docker-compose logs -f

.PHONY: compose-logs-api
compose-logs-api: ## View API service logs
	@echo "$(BLUE)Viewing API service logs...$(NC)"
	docker-compose logs -f api

.PHONY: compose-logs-db
compose-logs-db: ## View database service logs
	@echo "$(BLUE)Viewing database service logs...$(NC)"
	docker-compose logs -f postgres

.PHONY: compose-restart
compose-restart: compose-down compose-up ## Restart docker-compose services

##@ Database
.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(BLUE)Running database migrations...$(NC)"
	@if docker-compose ps postgres | grep -q "Up"; then \
		docker-compose exec postgres psql -U user -d todoapi -f /docker-entrypoint-initdb.d/init-db.sql; \
	else \
		echo "$(YELLOW)PostgreSQL container not running. Starting services...$(NC)"; \
		make compose-up; \
		sleep 10; \
		docker-compose exec postgres psql -U user -d todoapi -f /docker-entrypoint-initdb.d/init-db.sql; \
	fi
	@echo "$(GREEN)Database migrations completed$(NC)"

.PHONY: db-reset
db-reset: ## Reset database (drop and recreate)
	@echo "$(BLUE)Resetting database...$(NC)"
	@if docker-compose ps postgres | grep -q "Up"; then \
		docker-compose exec postgres psql -U user -c "DROP DATABASE IF EXISTS todoapi;"; \
		docker-compose exec postgres psql -U user -c "CREATE DATABASE todoapi;"; \
		docker-compose exec postgres psql -U user -d todoapi -f /docker-entrypoint-initdb.d/init-db.sql; \
	else \
		echo "$(RED)PostgreSQL container not running. Start services first with 'make compose-up'$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Database reset completed$(NC)"

.PHONY: db-shell
db-shell: ## Connect to database shell
	@echo "$(BLUE)Connecting to database shell...$(NC)"
	@if docker-compose ps postgres | grep -q "Up"; then \
		docker-compose exec postgres psql -U user -d todoapi; \
	else \
		echo "$(RED)PostgreSQL container not running. Start services first with 'make compose-up'$(NC)"; \
		exit 1; \
	fi

.PHONY: db-backup
db-backup: ## Create database backup
	@echo "$(BLUE)Creating database backup...$(NC)"
	@mkdir -p backups
	@if docker-compose ps postgres | grep -q "Up"; then \
		docker-compose exec postgres pg_dump -U user todoapi > backups/backup_$$(date +%Y%m%d_%H%M%S).sql; \
		echo "$(GREEN)Database backup created in backups/ directory$(NC)"; \
	else \
		echo "$(RED)PostgreSQL container not running. Start services first with 'make compose-up'$(NC)"; \
		exit 1; \
	fi

##@ Documentation
.PHONY: docs-generate
docs-generate: ## Generate API documentation
	@echo "$(BLUE)Generating API documentation...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs/; \
		echo "$(GREEN)API documentation generated$(NC)"; \
	else \
		echo "$(YELLOW)swag not found. Install it with: go install github.com/swaggo/swag/cmd/swag@latest$(NC)"; \
	fi

.PHONY: docs-serve
docs-serve: ## Serve documentation locally
	@echo "$(BLUE)Starting application to serve documentation...$(NC)"
	@echo "$(GREEN)Documentation will be available at: http://localhost:8080/swagger/index.html$(NC)"
	$(GOCMD) run $(MAIN_PATH)

##@ Utilities
.PHONY: env-example
env-example: ## Create .env file from .env.example
	@if [ ! -f .env ]; then \
		echo "$(BLUE)Creating .env file from .env.example...$(NC)"; \
		cp .env.example .env; \
		echo "$(GREEN).env file created. Please review and update the values.$(NC)"; \
	else \
		echo "$(YELLOW).env file already exists$(NC)"; \
	fi

.PHONY: health-check
health-check: ## Check application health
	@echo "$(BLUE)Checking application health...$(NC)"
	@if curl -s http://localhost:8080/health > /dev/null; then \
		echo "$(GREEN)Application is healthy$(NC)"; \
	else \
		echo "$(RED)Application is not responding$(NC)"; \
		exit 1; \
	fi

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)Development tools installed$(NC)"

##@ CI/CD
.PHONY: ci-test
ci-test: ## Run tests for CI/CD pipeline
	@echo "$(BLUE)Running CI tests...$(NC)"
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

.PHONY: ci-build
ci-build: ## Build for CI/CD pipeline
	@echo "$(BLUE)Running CI build...$(NC)"
	$(GOBUILD) -race -o $(BINARY_PATH) $(MAIN_PATH)

.PHONY: ci-check
ci-check: fmt-check vet lint ci-test ## Run all CI checks

##@ Release
.PHONY: release-build
release-build: ## Build optimized release binary
	@echo "$(BLUE)Building release binary...$(NC)"
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) \
		-ldflags="-w -s" \
		-o $(BINARY_PATH)-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) \
		-ldflags="-w -s" \
		-o $(BINARY_PATH)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) \
		-ldflags="-w -s" \
		-o $(BINARY_PATH)-windows-amd64.exe $(MAIN_PATH)
	@echo "$(GREEN)Release binaries built$(NC)"

.PHONY: version
version: ## Show version information
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "Go version: $$(go version)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $$(date)"

##@ Quick Start
.PHONY: setup
setup: env-example deps install-tools ## Initial project setup
	@echo "$(GREEN)Project setup completed!$(NC)"
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Review and update .env file"
	@echo "  2. Run 'make compose-up' to start services"
	@echo "  3. Run 'make health-check' to verify setup"

.PHONY: dev
dev: setup compose-up ## Complete development environment setup
	@echo "$(GREEN)Development environment ready!$(NC)"
	@echo "$(BLUE)API available at: http://localhost:8080$(NC)"
	@echo "$(BLUE)Documentation at: http://localhost:8080/swagger/index.html$(NC)"

.PHONY: quick-test
quick-test: compose-up ## Quick test of the entire system
	@echo "$(BLUE)Running quick system test...$(NC)"
	@sleep 5
	@make health-check
	@echo "$(GREEN)Quick test completed successfully!$(NC)"