# =============================================================================
# EquipFlow — Makefile
# =============================================================================

.PHONY: dev build test lint docker-up docker-down docker-build migrate help

# Standardziel
.DEFAULT_GOAL := help

# --- Entwicklung ---

dev: ## Frontend + Backend im Dev-Modus starten
	@echo "=== EquipFlow Dev-Modus ==="
	docker compose up -d postgres redis
	@echo "Infrastruktur laeuft. Services einzeln starten oder docker-up nutzen."

# --- Build ---

build: ## Alle Go-Services bauen
	@echo "=== Baue alle Services ==="
	@for dir in services/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "  -> Baue $$(basename $$dir)"; \
			(cd "$$dir" && go build -o bin/$$(basename $$dir) ./cmd/...); \
		fi; \
	done
	@echo "=== Build abgeschlossen ==="

# --- Tests ---

test: ## Alle Tests ausfuehren
	@echo "=== Tests ==="
	@for dir in services/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "  -> Teste $$(basename $$dir)"; \
			(cd "$$dir" && go test ./... -v -cover); \
		fi; \
	done
	@if [ -f "pkg/common/go.mod" ]; then \
		echo "  -> Teste pkg/common"; \
		(cd pkg/common && go test ./... -v -cover); \
	fi
	@echo "=== Tests abgeschlossen ==="

# --- Lint ---

lint: ## Go-Code mit golangci-lint pruefen
	@echo "=== Lint ==="
	@for dir in services/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "  -> Lint $$(basename $$dir)"; \
			(cd "$$dir" && golangci-lint run ./...); \
		fi; \
	done
	@if [ -f "pkg/common/go.mod" ]; then \
		echo "  -> Lint pkg/common"; \
		(cd pkg/common && golangci-lint run ./...); \
	fi
	@echo "=== Lint abgeschlossen ==="

# --- Docker ---

docker-up: ## Alle Container starten (Infrastruktur)
	docker compose up -d
	@echo "=== EquipFlow laeuft ==="
	@echo "  PostgreSQL: localhost:$${POSTGRES_PORT:-5432}"
	@echo "  Redis:      localhost:$${REDIS_PORT:-6379}"
	@echo "  Traefik:    localhost:$${TRAEFIK_DASHBOARD_PORT:-8080}"

docker-down: ## Alle Container stoppen
	docker compose down

docker-build: ## Alle Docker-Images bauen
	docker compose build --parallel

# --- Migration ---

migrate: ## Datenbank-Migrationen ausfuehren
	@echo "=== Migrationen ==="
	@for dir in services/*/; do \
		if [ -d "$$dir/migrations" ]; then \
			echo "  -> Migriere $$(basename $$dir)"; \
			DB_NAME=$$(basename $$dir | tr '-' '_'); \
			migrate -path "$$dir/migrations" \
				-database "postgres://$${POSTGRES_USER:-rentflow}:$${POSTGRES_PASSWORD}@localhost:$${POSTGRES_PORT:-5432}/$$DB_NAME?sslmode=disable" \
				up; \
		fi; \
	done
	@echo "=== Migrationen abgeschlossen ==="

# --- Hilfe ---

help: ## Diese Hilfe anzeigen
	@echo "EquipFlow — Verfuegbare Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
