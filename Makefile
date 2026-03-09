.PHONY: help setup build run test clean docker-up docker-down migrate

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o bin/api-server ./cmd/api-server
	go build -o bin/processor-worker ./cmd/processor-worker

run: ## Run the API server
	go run cmd/api-server/main.go

test: ## Run tests
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

migrate-up: ## Run database migrations up
	go run cmd/migrator/main.go up

migrate-down: ## Run database migrations down
	go run cmd/migrator/main.go down

dev: docker-up ## Start development environment
	@echo "Waiting for database..."
	@sleep 5
	@make migrate-up
	@make run

.DEFAULT_GOAL := help
