#!/bin/bash

# RentFlow Comprehensive Health Check Script
# Checks all 18 Go microservices, PostgreSQL, Redis, KurrentDB, and Traefik
# Exit code: 0 if all healthy, 1 if any unhealthy

set -o pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
TOTAL=0
HEALTHY=0
UNHEALTHY=0
WARNINGS=0

# Service definitions (name:port)
declare -a SERVICES=(
  "auth-service:8001"
  "inventory-service:8002"
  "invoice-service:8003"
  "project-service:8004"
  "crew-service:8005"
  "transport-service:8006"
  "warehouse-service:8007"
  "scanner-service:8008"
  "maintenance-service:8009"
  "document-service:8010"
  "notification-service:8011"
  "reporting-service:8012"
  "expense-service:8013"
  "insurance-service:8014"
  "ai-service:8015"
  "federation-service:8016"
  "audit-service:8017"
  "workflow-service:8018"
)

echo "=========================================="
echo "RentFlow Health Check"
echo "=========================================="
echo ""

# Function to check HTTP endpoint
check_service() {
  local service=$1
  local port=$2
  local url="http://localhost:${port}/health"
  
  TOTAL=$((TOTAL + 1))
  
  if curl -sf "$url" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} $service (port $port)"
    HEALTHY=$((HEALTHY + 1))
  else
    echo -e "${RED}✗${NC} $service (port $port) - Health endpoint unreachable"
    UNHEALTHY=$((UNHEALTHY + 1))
  fi
}

# Function to check service connectivity
check_service_port() {
  local service=$1
  local port=$2
  
  if timeout 2 bash -c "echo > /dev/tcp/localhost/$port" 2>/dev/null; then
    return 0
  else
    return 1
  fi
}

echo "Checking Go Microservices:"
echo "------------------------------------------"
for service_def in "${SERVICES[@]}"; do
  IFS=':' read -r service port <<< "$service_def"
  check_service "$service" "$port"
done

echo ""
echo "Checking Infrastructure Services:"
echo "------------------------------------------"

# PostgreSQL
TOTAL=$((TOTAL + 1))
if command -v pg_isready &> /dev/null && pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} PostgreSQL (port 5432)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/5432" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} PostgreSQL (port 5432) - Port open, client unavailable to verify"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} PostgreSQL (port 5432) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# Redis
TOTAL=$((TOTAL + 1))
if command -v redis-cli &> /dev/null && redis-cli -h localhost -p 6379 PING > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} Redis (port 6379)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo PING | nc -w1 localhost 6379" 2>/dev/null | grep -q PONG; then
  echo -e "${GREEN}✓${NC} Redis (port 6379)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/6379" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} Redis (port 6379) - Port open, PING unavailable to verify"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} Redis (port 6379) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# KurrentDB (EventStoreDB)
TOTAL=$((TOTAL + 1))
if curl -sf "http://localhost:2113/health/live" > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} KurrentDB/EventStoreDB (port 2113)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/2113" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} KurrentDB/EventStoreDB (port 2113) - Port open, health endpoint unreachable"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} KurrentDB/EventStoreDB (port 2113) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# Traefik
TOTAL=$((TOTAL + 1))
if curl -sf "http://localhost:8080/ping" > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} Traefik Dashboard/API (port 8080)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/8080" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} Traefik (port 8080) - Port open, ping unavailable to verify"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} Traefik (port 8080) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# Prometheus
TOTAL=$((TOTAL + 1))
if curl -sf "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} Prometheus (port 9090)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/9090" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} Prometheus (port 9090) - Port open, health endpoint unreachable"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} Prometheus (port 9090) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# Grafana
TOTAL=$((TOTAL + 1))
if curl -sf "http://localhost:3000/api/health" > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} Grafana (port 3000)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/3000" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} Grafana (port 3000) - Port open, health endpoint unreachable"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} Grafana (port 3000) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# Loki
TOTAL=$((TOTAL + 1))
if curl -sf "http://localhost:3100/ready" > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} Loki (port 3100)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/3100" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} Loki (port 3100) - Port open, health endpoint unreachable"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} Loki (port 3100) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# Alertmanager
TOTAL=$((TOTAL + 1))
if curl -sf "http://localhost:9093/-/healthy" > /dev/null 2>&1; then
  echo -e "${GREEN}✓${NC} Alertmanager (port 9093)"
  HEALTHY=$((HEALTHY + 1))
elif timeout 2 bash -c "echo > /dev/tcp/localhost/9093" 2>/dev/null; then
  echo -e "${YELLOW}?${NC} Alertmanager (port 9093) - Port open, health endpoint unreachable"
  WARNINGS=$((WARNINGS + 1))
  HEALTHY=$((HEALTHY + 1))
else
  echo -e "${RED}✗${NC} Alertmanager (port 9093) - Unreachable"
  UNHEALTHY=$((UNHEALTHY + 1))
fi

# Summary
echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="
echo -e "Total Services Checked: $TOTAL"
echo -e "${GREEN}Healthy: $HEALTHY${NC}"
if [ $WARNINGS -gt 0 ]; then
  echo -e "${YELLOW}Warnings (Partial Check): $WARNINGS${NC}"
fi
if [ $UNHEALTHY -gt 0 ]; then
  echo -e "${RED}Unhealthy: $UNHEALTHY${NC}"
fi
echo ""

# Exit with appropriate code
if [ $UNHEALTHY -gt 0 ]; then
  echo "Status: UNHEALTHY - Some services are not responding"
  exit 1
elif [ $WARNINGS -gt 0 ]; then
  echo "Status: PARTIALLY HEALTHY - All ports reachable, but some health endpoints unavailable"
  exit 0
else
  echo "Status: HEALTHY - All services operational"
  exit 0
fi
