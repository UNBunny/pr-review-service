.PHONY: help run build test test-all test-api test-coverage lint clean docker-up docker-down loadtest migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application with docker-compose
	docker-compose up --build

build: ## Build the Go application binary
	go build -o bin/server ./cmd/app

test: ## Run integration tests
	powershell -Command "$$env:TEST_DATABASE_URL='postgres://postgres:postgres@127.0.0.1:5433/pr_review?sslmode=disable'; go test -v ./tests/..."

lint: ## Run golangci-lint
	golangci-lint run ./...

clean: ## Clean up build artifacts and docker volumes
	rm -rf bin/
	docker-compose down -v

docker-up: ## Start docker-compose services
	docker-compose up -d

docker-down: ## Stop docker-compose services
	docker-compose down

loadtest: ## Run load test (requires service to be running)
	go run ./cmd/loadtest/main.go

migrate: ## Run migrations manually (requires DATABASE_URL env var)
	go run ./cmd/app/main.go

deps: ## Download Go dependencies
	go mod download
	go mod tidy

dev: ## Run in development mode (local)
	powershell -Command "$$env:DATABASE_URL='postgres://postgres:postgres@localhost:5433/pr_review?sslmode=disable'; $$env:SERVER_PORT='8080'; $$env:MIGRATIONS_PATH='./migrations'; go run ./cmd/app/main.go"

test-all: ## Run all automated tests (PowerShell script)
	powershell -ExecutionPolicy Bypass -File ./test-all.ps1

test-api: ## Run API tests (PowerShell script)
	powershell -ExecutionPolicy Bypass -File ./test-api.ps1

test-coverage: ## Run tests with coverage report
	powershell -Command "$$env:TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/pr_review?sslmode=disable'; go test -v -race -coverprofile=coverage.out -covermode=atomic ./tests/..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests via HTTP API
	docker-compose up -d
	@echo "Waiting for services to start..."
	@timeout /t 10 /nobreak >nul
	powershell -ExecutionPolicy Bypass -File ./run-integration-tests.ps1

test-integration-go: ## Run Go integration tests (requires local DB)
	docker-compose up -d db
	@echo "Waiting for database..."
	@timeout /t 5 /nobreak >nul
	powershell -Command "$$env:TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5433/pr_review?sslmode=disable'; go test -v -count=1 ./tests/..."

verify: ## Verify project (build + lint + test)
	@echo "🔍 Verifying project..."
	@$(MAKE) build
	@$(MAKE) lint
	@$(MAKE) test-integration
	@echo "✅ Verification complete!"
