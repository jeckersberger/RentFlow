.PHONY: help dev build test test-common test-integration test-e2e lint docker-build docker-up docker-down migration-audit migrate-up migrate-all migrate-down proto clean workspace-sync verify

# Get all services dynamically
SERVICES := $(notdir $(wildcard services/*/))
SERVICE_TARGETS := $(addprefix build-,$(SERVICES))
TEST_TARGETS := $(addprefix test-,$(SERVICES))

# Tooling. Override from the environment when needed (for example GO=/opt/go/bin/go).
GO ?= go
NPM ?= npm
COMPOSE ?= docker compose
GOBIN := $(shell pwd)/bin

# Database configuration (from .env or defaults)
POSTGRES_USER ?= rentflow_user
POSTGRES_PASSWORD ?= change_me_very_secure_password_here_32chars_min
POSTGRES_HOST ?= postgres
POSTGRES_PORT ?= 5432
POSTGRES_DB ?= rentflow

help:
	@echo "CrateDesk / EquipFlow - Available Targets"
	@echo ""
	@echo "Development:"
	@echo "  make dev              - Start all services in dev mode"
	@echo "  make build            - Build all Go services"
	@echo "  make clean            - Clean build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run pkg/common and all service unit tests"
	@echo "  make test-integration - Run integration-tagged Go tests"
	@echo "  make test-e2e         - Run Playwright smoke tests (requires backend and E2E credentials)"
	@echo "  make verify           - Run unit tests, builds and frontend checks"
	@echo "  make migration-audit  - Validate migration filenames/version uniqueness"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint             - Run golangci-lint"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build     - Build compose images"
	@echo "  make docker-up        - Start compose stack"
	@echo "  make docker-down      - Stop compose stack"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up SERVICE=<service> - Run pending versioned migrations for one service"
	@echo "  make migrate-all                  - Run pending versioned migrations for all services"
	@echo "  make migrate-down                 - Deliberately unsupported until rollback semantics are defined"
	@echo ""
	@echo "E2E environment:"
	@echo "  E2E_LOGIN_USER=<user> E2E_LOGIN_PASS=<pass> [E2E_BASE_URL=http://localhost:3000] make test-e2e"
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
	@cd services/$* && GOWORK=off $(GO) build -o $(GOBIN)/$* ./cmd/server

test: test-common $(TEST_TARGETS)
	@echo "All unit tests completed"

test-common:
	@echo "Running tests for pkg/common..."
	@cd pkg/common && GOWORK=off $(GO) test -v -race ./...

test-%:
	@echo "Running tests for $*..."
	@cd services/$* && GOWORK=off $(GO) test -v -race ./...

test-integration:
	@echo "Running integration tests..."
	@cd pkg/common && GOWORK=off $(GO) test -v -tags=integration ./...
	@for service in $(SERVICES); do \
		echo "Integration tests for $$service..."; \
		(cd services/$$service && GOWORK=off $(GO) test -v -tags=integration ./...) || exit 1; \
	done
	@echo "Integration tests passed"

test-e2e:
	@test -n "$(E2E_LOGIN_USER)" || (echo "E2E_LOGIN_USER is required"; exit 1)
	@test -n "$(E2E_LOGIN_PASS)" || (echo "E2E_LOGIN_PASS is required"; exit 1)
	@echo "Running Playwright smoke tests against $${E2E_BASE_URL:-http://localhost:3000}..."
	@cd frontend && $(NPM) ci && npx playwright install --with-deps chromium && $(NPM) run test:e2e

verify: test build
	@echo "Running frontend verification..."
	@cd frontend && $(NPM) ci && $(NPM) run lint && $(NPM) run type-check && $(NPM) run build && $(NPM) run test:run
	@echo "Verification completed successfully"

migration-audit:
	@bash scripts/audit-migrations.sh services

lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not installed"; exit 1)
	@echo "Linting pkg/common..."
	@cd pkg/common && GOWORK=off golangci-lint run ./...
	@for service in $(SERVICES); do \
		echo "Linting $$service..."; \
		(cd services/$$service && GOWORK=off golangci-lint run ./...) || exit 1; \
	done
	@echo "Lint checks passed"

docker-build:
	@echo "Building Docker images..."
	@$(COMPOSE) build

docker-up:
	@echo "Starting services with docker compose..."
	@$(COMPOSE) up -d

docker-down:
	@echo "Stopping services..."
	@$(COMPOSE) down

migrate-up: migration-audit
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE variable not set"; \
		echo "Usage: make migrate-up SERVICE=auth-service"; \
		exit 1; \
	fi
	@if [ ! -d "services/$(SERVICE)/migrations" ]; then \
		echo "Error: services/$(SERVICE)/migrations directory not found"; \
		exit 1; \
	fi
	@echo "Running versioned migrations for $(SERVICE)..."
	@$(COMPOSE) run --rm migrations "$(SERVICE)"

migrate-all: migration-audit
	@echo "Running versioned migrations for all services..."
	@$(COMPOSE) run --rm migrations

migrate-down:
	@echo "ERROR: automatic down migrations are intentionally disabled." >&2
	@echo "Rollback requires an explicit, reviewed data migration/restore plan." >&2
	@exit 2

proto:
	@echo "Generating protobuf code..."
	@echo "Placeholder for protobuf generation"
	@echo "When ready, integrate protoc and protoc-gen-go"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(GOBIN)
	@for service in $(SERVICES); do \
		(cd services/$$service && $(GO) clean) || true; \
	done
	@$(GO) clean -cache
	@echo "Clean complete"

workspace-sync:
	@echo "Syncing Go workspace..."
	@$(GO) work sync
	@echo "Workspace synced"

.DEFAULT_GOAL := help
