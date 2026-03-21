#!/bin/bash
set -e

# RentFlow Deployment Script
# Self-hosted deployment with image pull, migration, and rolling restarts
# Usage: ./scripts/deploy.sh [--dry-run]

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_FILE="${PROJECT_ROOT}/deployment.log"
DRY_RUN=false
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
DOCKER_COMPOSE_FILE="${PROJECT_ROOT}/infra/docker/docker-compose.yml"
ENV_FILE="${PROJECT_ROOT}/.env"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--dry-run]"
      exit 1
      ;;
  esac
done

# Logging function
log() {
  local level=$1
  shift
  local message="$@"
  echo -e "[${TIMESTAMP}] [${level}] ${message}" | tee -a "$LOG_FILE"
}

log_info() {
  log "${BLUE}INFO${NC}" "$@"
}

log_success() {
  log "${GREEN}SUCCESS${NC}" "$@"
}

log_warn() {
  log "${YELLOW}WARN${NC}" "$@"
}

log_error() {
  log "${RED}ERROR${NC}" "$@"
}

# Execute command with optional dry-run
execute_cmd() {
  local description=$1
  local cmd=$2

  if [ "$DRY_RUN" = true ]; then
    log_info "[DRY-RUN] Would execute: $description"
    log_info "[DRY-RUN] Command: $cmd"
  else
    log_info "Executing: $description"
    if eval "$cmd"; then
      log_success "Completed: $description"
    else
      log_error "Failed: $description"
      exit 1
    fi
  fi
}

# Check prerequisites
check_prerequisites() {
  log_info "Checking prerequisites..."

  if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed"
    exit 1
  fi

  if ! command -v docker-compose &> /dev/null; then
    log_error "Docker Compose is not installed"
    exit 1
  fi

  if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
    log_error "docker-compose.yml not found at $DOCKER_COMPOSE_FILE"
    exit 1
  fi

  if [ ! -f "$ENV_FILE" ]; then
    log_warn ".env file not found, using defaults from docker-compose.yml"
  fi

  log_success "Prerequisites check passed"
}

# Pull latest images from GHCR
pull_images() {
  log_info "Pulling latest Docker images from GHCR..."

  local registry="ghcr.io"
  local owner="rentflow"

  # List of services to update
  local services=(
    "auth-service"
    "inventory-service"
    "project-service"
    "scanner-service"
    "warehouse-service"
    "invoice-service"
    "document-service"
    "crew-service"
    "federation-service"
    "maintenance-service"
    "transport-service"
    "insurance-service"
    "workflow-service"
    "ai-service"
    "notification-service"
    "reporting-service"
    "audit-service"
    "expense-service"
  )

  for service in "${services[@]}"; do
    local image="${registry}/${owner}/${service}:latest"
    log_info "Pulling image: $image"

    if [ "$DRY_RUN" = false ]; then
      if ! docker pull "$image" 2>&1 | tee -a "$LOG_FILE"; then
        log_warn "Failed to pull $image (service may not exist yet)"
      fi
    else
      log_info "[DRY-RUN] Would pull: $image"
    fi
  done

  log_success "Image pull completed"
}

# Run database migrations
run_migrations() {
  log_info "Running database migrations..."

  # List of services with migrations
  local services=(
    "auth-service"
    "inventory-service"
    "project-service"
    "scanner-service"
    "warehouse-service"
    "invoice-service"
    "document-service"
    "crew-service"
    "federation-service"
    "maintenance-service"
    "transport-service"
    "insurance-service"
    "workflow-service"
    "ai-service"
    "notification-service"
    "reporting-service"
    "audit-service"
    "expense-service"
  )

  # Source .env file if it exists to get database credentials
  if [ -f "$ENV_FILE" ]; then
    export $(grep -v '^#' "$ENV_FILE" | xargs)
  fi

  for service in "${services[@]}"; do
    local migrations_dir="${PROJECT_ROOT}/services/${service}/migrations"

    if [ ! -d "$migrations_dir" ]; then
      log_warn "Migrations directory not found for $service, skipping"
      continue
    fi

    if [ ! "$(ls -A "$migrations_dir")" ]; then
      log_warn "No migration files found for $service, skipping"
      continue
    fi

    log_info "Running migrations for $service..."

    execute_cmd "migrate $service" \
      "cd '$PROJECT_ROOT' && make migrate-up SERVICE=$service"
  done

  log_success "All migrations completed"
}

# Perform rolling restart of services
rolling_restart() {
  log_info "Performing rolling restart of services..."

  # Services to restart in order (critical services first)
  local services=(
    "postgres"
    "redis"
    "kurrentdb"
    "traefik"
    "auth-service"
    "inventory-service"
    "project-service"
    "scanner-service"
    "warehouse-service"
    "invoice-service"
    "document-service"
    "crew-service"
    "federation-service"
    "maintenance-service"
    "transport-service"
    "insurance-service"
    "workflow-service"
    "ai-service"
    "notification-service"
    "reporting-service"
    "audit-service"
    "expense-service"
  )

  for service in "${services[@]}"; do
    log_info "Restarting service: $service"

    execute_cmd "restart $service" \
      "docker-compose -f '$DOCKER_COMPOSE_FILE' restart $service"

    # Wait a bit for service to stabilize
    sleep 3
  done

  log_success "Rolling restart completed"
}

# Verify deployment health
verify_health() {
  log_info "Verifying deployment health..."

  local max_attempts=30
  local attempt=0
  local health_check_passed=false

  while [ $attempt -lt $max_attempts ]; do
    log_info "Health check attempt $((attempt + 1))/$max_attempts"

    # Check if postgres is healthy
    if docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T postgres pg_isready -U "${POSTGRES_USER:-postgres}" &> /dev/null; then
      log_success "PostgreSQL is healthy"
      health_check_passed=true
      break
    fi

    attempt=$((attempt + 1))
    sleep 2
  done

  if [ "$health_check_passed" = false ]; then
    log_error "Health check failed after $max_attempts attempts"
    return 1
  fi

  log_success "Deployment health verification passed"
  return 0
}

# Main deployment flow
main() {
  log_info "=========================================="
  log_info "RentFlow Deployment Started"
  log_info "=========================================="
  log_info "Timestamp: $TIMESTAMP"
  log_info "Project Root: $PROJECT_ROOT"
  log_info "Dry Run: $DRY_RUN"
  log_info "Log File: $LOG_FILE"
  log_info "=========================================="

  check_prerequisites
  pull_images
  run_migrations
  rolling_restart

  if [ "$DRY_RUN" = false ]; then
    verify_health
  else
    log_info "[DRY-RUN] Skipping health verification"
  fi

  log_info "=========================================="
  log_success "Deployment completed successfully!"
  log_info "=========================================="
  log_info "Next steps:"
  log_info "1. Verify services are running: docker-compose -f $DOCKER_COMPOSE_FILE ps"
  log_info "2. Check logs: docker-compose -f $DOCKER_COMPOSE_FILE logs -f"
  log_info "3. Monitor health: http://localhost:8080 (Traefik Dashboard)"
}

# Run main
main "$@"
