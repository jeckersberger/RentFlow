.PHONY: help dev build test test-integration test-e2e lint docker-build docker-up docker-down migrate-up migrate-down proto clean

# Get all services dynamically
SERVICES := $(notdir $(wildcard services/*/))
SERVICE_TARGETS := $(addprefix build-,$(SERVICES))
TEST_TARGETS := $(addprefix test-,$(SERVICES))

# Go paths
GO := ~/.local/go/bin/go
GOBIN := $(shell pwd)/bin

help:
	@echo "RentFlow Microservices - Available Targets"
	@echo ""
	@echo "Development:"
	@echo "  make dev              - Start all services in dev mode"
	@echo "  make build            - Build all services"
	@echo "  make clean            - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run unit tests for all services"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-e2e         - Run end-to-end tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint             - Run golangci-lint"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build     - Build all Docker images"
	@echo "  make docker-up        - Start all services with docker-compose"
	@echo "  make docker-down      - Stop all services"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up       - Run database migrations"
	@echo "  make migrate-down     - Rollback migrations"
	@echo ""
	@echo "Other:"
	@echo "  make proto            - Generate protobuf code (placeholder)"
	@echo "  make help             - Show this help message"
	@echo ""

dev: build
	@echo "Starting all services in dev mode..."
	@for service in $(SERVICES); do \
		echo "Starting $$service..."; \
		(cd services/$$service && $(GO) run ./cmd/server) & \
	done
	@wait

build: $(SERVICE_TARGETS)
	@echo "All services built successfully"

build-%:
	@echo "Building $*..."
	@mkdir -p $(GOBIN)
	@cd services/$* && $(GO) build -o $(GOBIN)/$* ./cmd/server

test: $(TEST_TARGETS)
	@echo "All unit tests completed"

test-%:
	@echo "Running tests for $*..."
	@cd services/$* && $(GO) test -v -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	@for service in $(SERVICES); do \
		echo "Integration tests for $$service..."; \
		cd services/$$service && $(GO) test -v -tags=integration ./... || exit 1; \
	done
	@echo "Integration tests passed"

test-e2e:
	@echo "Running end-to-end tests..."
	@echo "E2E tests require all services running - start with 'make docker-up' first"
	@echo "Placeholder for E2E test suite"

lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed"; exit 1)
	@for service in $(SERVICES); do \
		echo "Linting $$service..."; \
		golangci-lint run services/$$service/... || exit 1; \
	done
	@echo "Lint checks passed"

docker-build:
	@echo "Building Docker images..."
	@if [ -f infra/docker/docker-compose.yml ]; then \
		docker-compose -f infra/docker/docker-compose.yml build; \
	else \
		echo "docker-compose.yml not found at infra/docker/"; \
	fi

docker-up:
	@echo "Starting services with docker-compose..."
	@if [ -f infra/docker/docker-compose.yml ]; then \
		docker-compose -f infra/docker/docker-compose.yml up -d; \
	else \
		echo "docker-compose.yml not found at infra/docker/"; \
	fi

docker-down:
	@echo "Stopping services..."
	@if [ -f infra/docker/docker-compose.yml ]; then \
		docker-compose -f infra/docker/docker-compose.yml down; \
	else \
		echo "docker-compose.yml not found at infra/docker/"; \
	fi

migrate-up:
	@echo "Running database migrations..."
	@echo "Placeholder for database migration tool (e.g., golang-migrate)"
	@echo "Each service has migrations/ directory for its schema"

migrate-down:
	@echo "Rolling back database migrations..."
	@echo "Placeholder for database rollback"

proto:
	@echo "Generating protobuf code..."
	@echo "Placeholder for protobuf generation"
	@echo "When ready, integrate protoc and protoc-gen-go"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(GOBIN)
	@for service in $(SERVICES); do \
		cd services/$$service && $(GO) clean || true; \
	done
	@$(GO) clean -cache
	@echo "Clean complete"

workspace-sync:
	@echo "Syncing Go workspace..."
	@$(GO) work sync
	@echo "Workspace synced"

.DEFAULT_GOAL := help
