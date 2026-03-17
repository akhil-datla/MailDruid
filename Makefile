APP_NAME := maildruid
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: help build run test lint clean docker-build docker-up docker-down migrate frontend deps fmt

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

frontend: ## Build the frontend
	cd web && npm install --legacy-peer-deps && npm run build
	rm -rf internal/server/frontend
	cp -r web/dist internal/server/frontend

build: frontend ## Build the binary (includes frontend)
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/maildruid

build-go: ## Build Go binary only (skip frontend)
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/maildruid

run: build ## Build and run the server
	./bin/$(APP_NAME) serve

dev: ## Run frontend dev server (with API proxy to :8080)
	cd web && npm run dev

test: ## Run tests
	go test -race -cover ./...

lint: ## Run linter
	golangci-lint run ./...

clean: ## Remove build artifacts
	rm -rf bin/ web/dist internal/server/frontend

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-up: ## Start with docker-compose
	docker compose up -d

docker-down: ## Stop docker-compose services
	docker compose down

migrate: ## Run database migrations
	go run ./cmd/maildruid migrate

fmt: ## Format code
	go fmt ./...

deps: ## Download all dependencies
	go mod download && go mod tidy
	cd web && npm install --legacy-peer-deps
