#!/bin/bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVICES_DIR="$PROJECT_ROOT/services"

# Function to print usage
usage() {
    cat << EOF
${BLUE}RentFlow Service Template Generator${NC}

${YELLOW}Usage:${NC}
    $0 <service-name> [options]

${YELLOW}Arguments:${NC}
    service-name        Name of the service (lowercase, hyphens only)
                        Example: user-service, payment-service

${YELLOW}Options:${NC}
    -h, --help          Show this help message
    -f, --force         Overwrite existing service (use with caution)

${YELLOW}Examples:${NC}
    $0 notification-service
    $0 payment-service --force

EOF
}

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $*"
}

print_success() {
    echo -e "${GREEN}✓${NC} $*"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $*"
}

print_error() {
    echo -e "${RED}✗${NC} $*"
}

# Function to validate service name format
validate_service_name() {
    local name=$1
    if [[ ! $name =~ ^[a-z][a-z0-9]*(-[a-z0-9]+)*$ ]]; then
        print_error "Invalid service name: $name"
        echo "Service name must:"
        echo "  - Start with a lowercase letter"
        echo "  - Contain only lowercase letters, numbers, and hyphens"
        echo "  - Not end with a hyphen"
        return 1
    fi
    return 0
}

# Function to convert service name to Go package format
to_go_package() {
    echo "$1" | tr '-' '_'
}

# Main script logic
main() {
    # Parse arguments
    SERVICE_NAME=""
    FORCE_FLAG=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -f|--force)
                FORCE_FLAG=true
                shift
                ;;
            -*)
                print_error "Unknown option: $1"
                usage
                exit 1
                ;;
            *)
                if [ -z "$SERVICE_NAME" ]; then
                    SERVICE_NAME=$1
                else
                    print_error "Too many arguments"
                    usage
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate arguments
    if [ -z "$SERVICE_NAME" ]; then
        print_error "Service name is required"
        usage
        exit 1
    fi

    # Validate service name format
    if ! validate_service_name "$SERVICE_NAME"; then
        exit 1
    fi

    SERVICE_DIR="$SERVICES_DIR/$SERVICE_NAME"
    GO_PACKAGE=$(to_go_package "$SERVICE_NAME")

    # Check if service already exists
    if [ -d "$SERVICE_DIR" ]; then
        if [ "$FORCE_FLAG" = false ]; then
            print_error "Service directory already exists: $SERVICE_DIR"
            echo "Use --force to overwrite"
            exit 1
        else
            print_warning "Overwriting existing service: $SERVICE_NAME"
            rm -rf "$SERVICE_DIR"
        fi
    fi

    print_info "Creating service: $SERVICE_NAME"
    echo ""

    # Create directory structure
    print_info "Creating directory structure..."
    mkdir -p "$SERVICE_DIR"/{cmd/server,internal/{domain,application,adapters/{http,grpc}},internal/infrastructure/repositories,migrations}
    print_success "Directory structure created"

    # Create main.go
    print_info "Generating main.go..."
    create_main_go "$SERVICE_DIR" "$SERVICE_NAME" "$GO_PACKAGE"
    print_success "main.go created"

    # Create domain files
    print_info "Generating domain files..."
    create_domain_models "$SERVICE_DIR" "$GO_PACKAGE"
    create_domain_events "$SERVICE_DIR" "$GO_PACKAGE"
    create_domain_repository "$SERVICE_DIR" "$GO_PACKAGE"
    print_success "Domain files created"

    # Create application layer
    print_info "Generating application layer..."
    create_application_service "$SERVICE_DIR" "$GO_PACKAGE"
    print_success "Application service created"

    # Create adapters
    print_info "Generating HTTP adapter..."
    create_http_router "$SERVICE_DIR" "$GO_PACKAGE"
    create_http_handlers "$SERVICE_DIR" "$GO_PACKAGE"
    print_success "HTTP adapter created"

    # Create infrastructure
    print_info "Generating infrastructure layer..."
    create_postgres_repository "$SERVICE_DIR" "$GO_PACKAGE"
    print_success "Infrastructure layer created"

    # Create migration files
    print_info "Generating migration files..."
    create_migrations "$SERVICE_DIR"
    print_success "Migration files created"

    # Create go.mod
    print_info "Generating go.mod..."
    create_go_mod "$SERVICE_DIR" "$SERVICE_NAME"
    print_success "go.mod created"

    # Create Dockerfile
    print_info "Generating Dockerfile..."
    create_dockerfile "$SERVICE_DIR" "$SERVICE_NAME"
    print_success "Dockerfile created"

    # Update go.work
    print_info "Updating go.work..."
    update_go_work "$SERVICE_NAME"
    print_success "go.work updated"

    # Print summary
    echo ""
    print_success "Service generated successfully!"
    echo ""
    echo "${BLUE}Next steps:${NC}"
    echo "1. Navigate to the service directory:"
    echo "   cd $SERVICE_DIR"
    echo ""
    echo "2. Install dependencies:"
    echo "   go mod download"
    echo ""
    echo "3. Verify the code compiles:"
    echo "   go build ./cmd/server"
    echo ""
    echo "4. Update domain models in:"
    echo "   $SERVICE_DIR/internal/domain/"
    echo ""
    echo "5. Implement business logic in:"
    echo "   $SERVICE_DIR/internal/application/"
    echo ""
    echo "6. Add HTTP handlers in:"
    echo "   $SERVICE_DIR/internal/adapters/http/"
    echo ""
    echo "7. Implement database layer in:"
    echo "   $SERVICE_DIR/internal/infrastructure/repositories/"
    echo ""
    print_info "Files created:"
    find "$SERVICE_DIR" -type f | sort | sed 's/^/   /'
}

# Generate main.go
create_main_go() {
    local service_dir=$1
    local service_name=$2
    local go_package=$3

    cat > "$service_dir/cmd/server/main.go" << 'MAIN_EOF'
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeckersberger/rentflow/shared/logger"
	"github.com/jeckersberger/rentflow/shared/config"
)

func main() {
	// Initialize logger
	log := logger.New()
	defer log.Sync()

	log.Info("Starting service", "version", "1.0.0")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready"}`)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Starting HTTP server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", "error", err)
		os.Exit(1)
	}

	log.Info("Server stopped")
}
MAIN_EOF
}

# Generate domain models
create_domain_models() {
    local service_dir=$1
    local go_package=$2

    cat > "$service_dir/internal/domain/models.go" << 'MODELS_EOF'
package domain

import "time"

// Entity represents a base entity with common fields
type Entity struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewEntity creates a new entity with generated ID and timestamps
func NewEntity(id string) Entity {
	now := time.Now().UTC()
	return Entity{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
MODELS_EOF
}

# Generate domain events
create_domain_events() {
    local service_dir=$1
    local go_package=$2

    cat > "$service_dir/internal/domain/events.go" << 'EVENTS_EOF'
package domain

import "time"

// DomainEvent represents a domain event interface
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
	Aggregate() string
}

// BaseDomainEvent provides common event fields
type BaseDomainEvent struct {
	eventID    string
	eventType  string
	occurredAt time.Time
	aggregate  string
}

// EventID returns the event ID
func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

// EventType returns the event type
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// OccurredAt returns when the event occurred
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// Aggregate returns the aggregate ID
func (e BaseDomainEvent) Aggregate() string {
	return e.aggregate
}
EVENTS_EOF
}

# Generate domain repository interface
create_domain_repository() {
    local service_dir=$1
    local go_package=$2

    cat > "$service_dir/internal/domain/repository.go" << 'REPOSITORY_EOF'
package domain

import "context"

// Repository defines the interface for data persistence
type Repository interface {
	// Ping checks the repository connection
	Ping(ctx context.Context) error
	// Close closes the repository connection
	Close() error
}
REPOSITORY_EOF
}

# Generate application service
create_application_service() {
    local service_dir=$1
    local go_package=$2

    cat > "$service_dir/internal/application/service.go" << 'SERVICE_EOF'
package application

import (
	"context"

	"github.com/jeckersberger/rentflow/shared/logger"
)

// Service defines the application service
type Service struct {
	log logger.Logger
}

// NewService creates a new application service
func NewService(log logger.Logger) *Service {
	return &Service{
		log: log,
	}
}

// Health returns the health status
func (s *Service) Health(ctx context.Context) map[string]string {
	return map[string]string{
		"status": "healthy",
	}
}
SERVICE_EOF
}

# Generate HTTP router
create_http_router() {
    local service_dir=$1
    local go_package=$2

    cat > "$service_dir/internal/adapters/http/router.go" << 'ROUTER_EOF'
package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/shared/logger"
)

// Router sets up HTTP routes
type Router struct {
	mux    *http.ServeMux
	log    logger.Logger
	handle *Handler
}

// NewRouter creates a new HTTP router
func NewRouter(log logger.Logger, handler *Handler) *Router {
	mux := http.NewServeMux()
	router := &Router{
		mux:    mux,
		log:    log,
		handle: handler,
	}
	router.setupRoutes()
	return router
}

// setupRoutes configures all routes
func (r *Router) setupRoutes() {
	// Add your routes here
	// Example: r.mux.HandleFunc("/api/v1/endpoint", r.handle.Endpoint)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
ROUTER_EOF
}

# Generate HTTP handlers
create_http_handlers() {
    local service_dir=$1
    local go_package=$2

    cat > "$service_dir/internal/adapters/http/handlers.go" << 'HANDLERS_EOF'
package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/shared/logger"
)

// Handler handles HTTP requests
type Handler struct {
	log logger.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(log logger.Logger) *Handler {
	return &Handler{
		log: log,
	}
}

// WriteJSON writes a JSON response
func (h *Handler) WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response
func (h *Handler) WriteError(w http.ResponseWriter, status int, message string) {
	h.WriteJSON(w, status, map[string]string{"error": message})
}
HANDLERS_EOF
}

# Generate PostgreSQL repository
create_postgres_repository() {
    local service_dir=$1
    local go_package=$2

    cat > "$service_dir/internal/infrastructure/repositories/postgres.go" << 'POSTGRES_EOF'
package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/shared/logger"
	"github.com/jeckersberger/rentflow/services/[SERVICE_NAME]/internal/domain"
	_ "github.com/lib/pq"
)

// PostgresRepository implements the repository interface using PostgreSQL
type PostgresRepository struct {
	db  *sql.DB
	log logger.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(dsn string, log logger.Logger) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{
		db:  db,
		log: log,
	}, nil
}

// Ping checks the database connection
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// Compile-time proof of interface implementation
var _ domain.Repository = (*PostgresRepository)(nil)
POSTGRES_EOF

    # Replace placeholder
    sed -i "s/\[SERVICE_NAME\]/$1/g" "$service_dir/internal/infrastructure/repositories/postgres.go"
}

# Generate migration files
create_migrations() {
    local service_dir=$1

    cat > "$service_dir/migrations/001_initial.up.sql" << 'MIGRATION_UP_EOF'
-- Initial migration for service
-- Add your schema here

CREATE TABLE IF NOT EXISTS schema_version (
    version INT PRIMARY KEY,
    description TEXT NOT NULL,
    installed_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_version (version, description) VALUES (1, 'Initial schema');
MIGRATION_UP_EOF

    cat > "$service_dir/migrations/001_initial.down.sql" << 'MIGRATION_DOWN_EOF'
-- Rollback initial migration
DROP TABLE IF EXISTS schema_version;
MIGRATION_DOWN_EOF
}

# Generate go.mod
create_go_mod() {
    local service_dir=$1
    local service_name=$2

    cat > "$service_dir/go.mod" << GO_MOD_EOF
module github.com/jeckersberger/rentflow/services/$service_name

go 1.24

require (
	github.com/jeckersberger/rentflow/shared v0.1.0
	github.com/lib/pq v1.10.9
)
GO_MOD_EOF
}

# Generate Dockerfile
create_dockerfile() {
    local service_dir=$1
    local service_name=$2

    cat > "$service_dir/Dockerfile" << DOCKERFILE_EOF
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \\
    -ldflags="-w -s" \\
    -o server \\
    ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \\
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./server"]
DOCKERFILE_EOF
}

# Update go.work
update_go_work() {
    local service_name=$1
    local go_work_file="$PROJECT_ROOT/go.work"

    if [ ! -f "$go_work_file" ]; then
        print_warning "go.work not found at $go_work_file"
        return
    fi

    local service_path="./services/$service_name"

    # Check if service is already in go.work
    if grep -q "$service_path" "$go_work_file"; then
        print_info "Service already in go.work"
        return
    fi

    # Add service to go.work
    {
        echo ""
        echo "use $service_path"
    } >> "$go_work_file"
}

# Run main function
main "$@"
