#!/bin/bash
set -e

# RentFlow Rollback Script
# Rollback to previous image tag with optional migration rollback
# Usage: ./scripts/rollback.sh [--tag <tag>] [--rollback-migrations] [--dry-run]

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_FILE="${PROJECT_ROOT}/rollback.log"
DRY_RUN=false
ROLLBACK_MIGRATIONS=false
TARGET_TAG=""
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
    --tag)
      TARGET_TAG="$2"
      shift 2
      ;;
    --rollback-migrations)
      ROLLBACK_MIGRATIONS=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --help)
      echo "Usage: $0 [--tag <tag>] [--rollback-migrations] [--dry-run]"
      echo ""
      echo "Options:"
      echo "  --tag <tag>               Specific tag to rollback to (default: previous)"
      echo "  --rollback-migrations     Also rollback database migrations"
      echo "  --dry-run                 Show what would be done without making changes"
      echo "  --help                    Show this help message"
      echo ""
      echo "Examples:"
      echo "  ./scripts/rollback.sh                           # Rollback to previous version"
      echo "  ./scripts/rollback.sh --tag v1.2.3              # Rollback to specific version"
      echo "  ./scripts/rollback.sh --rollback-migrations     # Also rollback database"
      echo "  ./scripts/rollback.sh --tag v1.2.3 --dry-run    # Preview rollback"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
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

# Get current running image tags
get_current_tags() {
  log_info "Retrieving current running image tags..."

  local registry="ghcr.io"
  local owner="rentflow"

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
    local container_name="rentflow-${service}"
    local current_image=$(docker inspect "$container_name" --format='{{.Config.Image}}' 2>/dev/null || echo "unknown")
    echo "$service: $current_image" | tee -a "$LOG_FILE"
  done
}

# Determine target tag to rollback to
determine_rollback_tag() {
  if [ -n "$TARGET_TAG" ]; then
    log_info "Using specified tag: $TARGET_TAG"
    return 0
  fi

  log_info "Attempting to determine previous tag from git..."

  if ! command -v git &> /dev/null; then
    log_error "Git not found and no --tag specified. Please provide target tag with --tag"
    exit 1
  fi

  # Get the previous tag from git
  TARGET_TAG=$(git -C "$PROJECT_ROOT" describe --tags --abbrev=0 2>/dev/null || echo "")

  if [ -z "$TARGET_TAG" ]; then
    log_error "Could not determine previous tag. Please specify with --tag option"
    exit 1
  fi

  log_info "Determined rollback tag: $TARGET_TAG"
}

# Stop services before rollback
stop_services() {
  log_info "Stopping services for safe rollback..."

  execute_cmd "stop all services" \
    "docker-compose -f '$DOCKER_COMPOSE_FILE' stop"

  sleep 2
}

# Rollback Docker images to target tag
rollback_images() {
  log_info "Rolling back Docker images to tag: $TARGET_TAG"

  local registry="ghcr.io"
  local owner="rentflow"

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
    local image="${registry}/${owner}/${service}:${TARGET_TAG}"
    log_info "Pulling image: $image"

    if [ "$DRY_RUN" = false ]; then
      if docker pull "$image" 2>&1 | tee -a "$LOG_FILE"; then
        log_success "Pulled: $image"
      else
        log_warn "Failed to pull $image (service may not exist in this tag)"
      fi
    else
      log_info "[DRY-RUN] Would pull: $image"
    fi
  done

  log_success "Image rollback completed"
}

# Rollback database migrations
rollback_database_migrations() {
  if [ "$ROLLBACK_MIGRATIONS" = false ]; then
    log_info "Skipping database migration rollback (use --rollback-migrations to enable)"
    return 0
  fi

  log_info "Rolling back database migrations..."
  log_warn "This will revert database schema changes!"

  # Source .env file if it exists to get database credentials
  if [ -f "$ENV_FILE" ]; then
    export $(grep -v '^#' "$ENV_FILE" | xargs)
  fi

  # List of services with migrations (in reverse order for rollback)
  local services=(
    "expense-service"
    "audit-service"
    "reporting-service"
    "notification-service"
    "ai-service"
    "workflow-service"
    "insurance-service"
    "transport-service"
    "maintenance-service"
    "federation-service"
    "crew-service"
    "document-service"
    "invoice-service"
    "warehouse-service"
    "scanner-service"
    "project-service"
    "inventory-service"
    "auth-service"
  )

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

    log_info "Rolling back migrations for $service..."

    execute_cmd "rollback $service migrations" \
      "cd '$PROJECT_ROOT' && make migrate-down SERVICE=$service"
  done

  log_success "All migration rollbacks completed"
}

# Start services after rollback
start_services() {
  log_info "Starting services with rolled-back images..."

  execute_cmd "start all services" \
    "docker-compose -f '$DOCKER_COMPOSE_FILE' up -d"

  sleep 3
}

# Verify rollback health
verify_rollback_health() {
  log_info "Verifying rollback health..."

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

  log_success "Rollback health verification passed"
  return 0
}

# Prompt for confirmation
confirm_rollback() {
  if [ "$DRY_RUN" = true ]; then
    log_warn "[DRY-RUN MODE] No changes will be made"
    return 0
  fi

  log_warn "=========================================="
  log_warn "ROLLBACK CONFIRMATION"
  log_warn "=========================================="
  log_warn "Target Tag: ${TARGET_TAG:-previous}"
  log_warn "Rollback Migrations: $ROLLBACK_MIGRATIONS"
  log_warn "This action will:"
  log_warn "  1. Stop all services"
  log_warn "  2. Pull previous images from GHCR"
  if [ "$ROLLBACK_MIGRATIONS" = true ]; then
    log_warn "  3. Rollback database migrations (DATA LOSS POSSIBLE)"
    log_warn "  4. Start services"
  else
    log_warn "  3. Start services"
  fi
  log_warn "=========================================="
  read -p "Are you sure you want to proceed? (yes/no): " -r
  echo

  if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    log_info "Rollback cancelled by user"
    exit 0
  fi

  log_success "Confirmed, proceeding with rollback"
}

# Main rollback flow
main() {
  log_info "=========================================="
  log_info "RentFlow Rollback Started"
  log_info "=========================================="
  log_info "Timestamp: $TIMESTAMP"
  log_info "Project Root: $PROJECT_ROOT"
  log_info "Dry Run: $DRY_RUN"
  log_info "Log File: $LOG_FILE"
  log_info "=========================================="

  check_prerequisites
  determine_rollback_tag
  get_current_tags
  confirm_rollback
  stop_services
  rollback_images
  rollback_database_migrations
  start_services

  if [ "$DRY_RUN" = false ]; then
    verify_rollback_health
  else
    log_info "[DRY-RUN] Skipping health verification"
  fi

  log_info "=========================================="
  log_success "Rollback completed successfully!"
  log_info "=========================================="
  log_info "Rolled back to tag: $TARGET_TAG"
  log_info "Services are running with previous images"
  log_info "Next steps:"
  log_info "1. Verify services are running: docker-compose -f $DOCKER_COMPOSE_FILE ps"
  log_info "2. Check logs: docker-compose -f $DOCKER_COMPOSE_FILE logs -f"
  log_info "3. Monitor health: http://localhost:8080 (Traefik Dashboard)"
  log_info "=========================================="
}

# Run main
main "$@"
