# Infrastructure, Testing & Performance Specification

**Stand:** 21. März 2026
**Version:** 1.0
**Zielgruppe:** DevOps Engineer, QA Lead, Performance Engineer

---

## 1. Docker Compose Stack (Production)

### 1.1 docker-compose.yml (Complete)

```yaml
version: '3.9'

networks:
  rentflow:
    driver: bridge

volumes:
  kurrentdb_data:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:
  traefik_certs:

x-common-environment: &common-env
  ENVIRONMENT: production
  LOG_LEVEL: info
  TRAEFIK_ENTRYPOINT: https
  TRAEFIK_SCHEME: https

x-postgres-conn: &postgres-conn
  POSTGRES_USER: rentflow_user
  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:?POSTGRES_PASSWORD required}
  POSTGRES_HOST: postgres
  POSTGRES_PORT: 5432

x-kurrentdb-conn: &kurrentdb-conn
  KURRENTDB_HOST: kurrentdb
  KURRENTDB_PORT: 2113
  KURRENTDB_USER: admin
  KURRENTDB_PASSWORD: ${KURRENTDB_PASSWORD:?KURRENTDB_PASSWORD required}

x-redis-conn: &redis-conn
  REDIS_HOST: redis
  REDIS_PORT: 6379
  REDIS_PASSWORD: ${REDIS_PASSWORD:?REDIS_PASSWORD required}

services:
  # ============================================================
  # Infrastructure: Traefik v3 (API Gateway)
  # ============================================================
  traefik:
    image: traefik:v3.0
    container_name: rentflow-traefik
    restart: always
    privileged: true
    networks:
      - rentflow
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Dashboard (admin-only)
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_certs:/letsencrypt
      - ./config/traefik/traefik.yml:/etc/traefik/traefik.yml:ro
      - ./config/traefik/dynamic:/etc/traefik/dynamic:ro
    environment:
      TRAEFIK_API_DASHBOARD: "true"
      TRAEFIK_PROVIDERS_DOCKER: "true"
      TRAEFIK_PROVIDERS_DOCKER_EXPOSEDBYDEFAULT: "false"
      TRAEFIK_PROVIDERS_FILE_DIRECTORY: /etc/traefik/dynamic
      TRAEFIK_ENTRYPOINTS_WEB_ADDRESS: ":80"
      TRAEFIK_ENTRYPOINTS_WEBSECURE_ADDRESS: ":443"
      TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_EMAIL: ${LETSENCRYPT_EMAIL:-admin@rentflow.local}
      TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_STORAGE: /letsencrypt/acme.json
      TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_HTTPCHALLENGE: "true"
      TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_HTTPCHALLENGE_ENTRYPOINT: web
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
    healthcheck:
      test: ["CMD", "traefik", "healthcheck", "--ping"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.dashboard.entrypoints: "websecure"
      traefik.http.routers.dashboard.rule: "Host(`${DOMAIN}`) && PathPrefix(`/admin/dashboard`)"
      traefik.http.routers.dashboard.service: "api@internal"
      traefik.http.routers.dashboard.tls: "true"
      traefik.http.routers.dashboard.tls.certresolver: "letsencrypt"
      traefik.http.middlewares.dashboard-auth.basicauth.users: "admin:${DASHBOARD_HTPASSWD}"
      traefik.http.routers.dashboard.middlewares: "dashboard-auth"

  # ============================================================
  # Data: KurrentDB (Event Store)
  # ============================================================
  kurrentdb:
    image: kurrent/kurrentdb:26.0
    container_name: rentflow-kurrentdb
    restart: always
    networks:
      - rentflow
    ports:
      - "127.0.0.1:2113:2113"
    environment:
      KURRENTDB_DEFAULT_ADMIN_USER_PASSWORD: ${KURRENTDB_PASSWORD}
      KURRENTDB_DEFAULT_ADMIN_USER_ENABLED: "true"
      KURRENTDB_CLUSTER_SIZE: 1
      KURRENTDB_LOG_LEVEL: Info
      KURRENTDB_STATS_PERIOD_SEC: 0  # Disable stats for small deployments
      KURRENTDB_ENABLE_EXTERNAL_TCP: "true"
      KURRENTDB_EXT_TCP_HEARTBEAT_INTERVAL: 750
      KURRENTDB_EXT_TCP_HEARTBEAT_TIMEOUT: 1500
      KURRENTDB_ENABLE_ATOM_PUB_OVER_HTTP: "false"
      KURRENTDB_GOSSIP_SEED: ""
      KURRENTDB_ENABLE_TRUSTED_AUTH: "false"
    volumes:
      - kurrentdb_data:/var/lib/kurrentdb
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1.5G
        reservations:
          cpus: '0.5'
          memory: 512M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:2113/health/live"]
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 30s

  # ============================================================
  # Data: PostgreSQL 16 (Read Models & Projections)
  # ============================================================
  postgres:
    image: postgres:16-alpine
    container_name: rentflow-postgres
    restart: always
    networks:
      - rentflow
    ports:
      - "127.0.0.1:5432:5432"
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-rentflow_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: rentflow
      POSTGRES_INITDB_ARGS: >
        -c shared_buffers=1GB
        -c effective_cache_size=3GB
        -c maintenance_work_mem=256MB
        -c work_mem=16MB
        -c wal_buffers=32MB
        -c default_statistics_target=100
        -c random_page_cost=1.1
        -c effective_io_concurrency=200
        -c max_connections=200
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./config/postgres/postgresql.conf:/etc/postgresql/postgresql.conf:ro
      - ./migrations:/docker-entrypoint-initdb.d:ro
      - ./backups:/backups:rw
    deploy:
      resources:
        limits:
          cpus: '1.5'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-rentflow_user}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 20s

  # ============================================================
  # Data: Redis 7 (Sessions, Cache, Rate Limiting)
  # ============================================================
  redis:
    image: redis:7-alpine
    container_name: rentflow-redis
    restart: always
    networks:
      - rentflow
    ports:
      - "127.0.0.1:6379:6379"
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --maxmemory 512mb
      --maxmemory-policy allkeys-lru
      --timeout 300
      --tcp-backlog 511
      --databases 16
      --loglevel notice
    volumes:
      - redis_data:/data
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  # ============================================================
  # Microservices: Auth Service (Port 8001)
  # ============================================================
  auth-service:
    build:
      context: ./services/auth-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-auth
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: auth-service
      SERVICE_PORT: 8001
      JWT_SECRET: ${JWT_SECRET:?JWT_SECRET required}
      JWT_EXPIRY_HOURS: 24
      JWT_REFRESH_EXPIRY_DAYS: 7
      BCRYPT_COST: 12
      SESSION_TTL_MINUTES: 1440
      MFA_ENABLED: ${MFA_ENABLED:-true}
      TENANT_ISOLATION: "true"
      SCHEMA_NAME: auth_schema
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8001/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.auth.entrypoints: "websecure"
      traefik.http.routers.auth.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/auth`)"
      traefik.http.routers.auth.tls: "true"
      traefik.http.routers.auth.tls.certresolver: "letsencrypt"
      traefik.http.services.auth.loadbalancer.server.port: 8001

  # ============================================================
  # Microservices: Inventory Service (Port 8002)
  # ============================================================
  inventory-service:
    build:
      context: ./services/inventory-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-inventory
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: inventory-service
      SERVICE_PORT: 8002
      SCHEMA_NAME: inventory_schema
      CACHE_TTL_SECONDS: 300
      BATCH_SIZE_DEFAULT: 1000
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8002/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.inventory.entrypoints: "websecure"
      traefik.http.routers.inventory.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/inventory`)"
      traefik.http.routers.inventory.tls: "true"
      traefik.http.routers.inventory.tls.certresolver: "letsencrypt"
      traefik.http.services.inventory.loadbalancer.server.port: 8002

  # ============================================================
  # Microservices: Project Service (Port 8003)
  # ============================================================
  project-service:
    build:
      context: ./services/project-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-project
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: project-service
      SERVICE_PORT: 8003
      SCHEMA_NAME: project_schema
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8003/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.project.entrypoints: "websecure"
      traefik.http.routers.project.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/projects`)"
      traefik.http.routers.project.tls: "true"
      traefik.http.routers.project.tls.certresolver: "letsencrypt"
      traefik.http.services.project.loadbalancer.server.port: 8003

  # ============================================================
  # Microservices: Scanner Service (Port 8004)
  # ============================================================
  scanner-service:
    build:
      context: ./services/scanner-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-scanner
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: scanner-service
      SERVICE_PORT: 8004
      SCHEMA_NAME: scanner_schema
      OFFLINE_SYNC_ENABLED: "true"
      OFFLINE_CACHE_TTL: 3600
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8004/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.scanner.entrypoints: "websecure"
      traefik.http.routers.scanner.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/scanner`)"
      traefik.http.routers.scanner.tls: "true"
      traefik.http.routers.scanner.tls.certresolver: "letsencrypt"
      traefik.http.services.scanner.loadbalancer.server.port: 8004

  # ============================================================
  # Microservices: Warehouse Service (Port 8005)
  # ============================================================
  warehouse-service:
    build:
      context: ./services/warehouse-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-warehouse
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: warehouse-service
      SERVICE_PORT: 8005
      SCHEMA_NAME: warehouse_schema
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8005/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.warehouse.entrypoints: "websecure"
      traefik.http.routers.warehouse.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/warehouse`)"
      traefik.http.routers.warehouse.tls: "true"
      traefik.http.routers.warehouse.tls.certresolver: "letsencrypt"
      traefik.http.services.warehouse.loadbalancer.server.port: 8005

  # ============================================================
  # Microservices: Invoice Service (Port 8006)
  # ============================================================
  invoice-service:
    build:
      context: ./services/invoice-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-invoice
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: invoice-service
      SERVICE_PORT: 8006
      SCHEMA_NAME: invoice_schema
      PDF_GENERATION_TIMEOUT: 30
      DATEV_EXPORT_ENABLED: "true"
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8006/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.invoice.entrypoints: "websecure"
      traefik.http.routers.invoice.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/invoices`)"
      traefik.http.routers.invoice.tls: "true"
      traefik.http.routers.invoice.tls.certresolver: "letsencrypt"
      traefik.http.services.invoice.loadbalancer.server.port: 8006

  # ============================================================
  # Microservices: Document Service (Port 8007)
  # ============================================================
  document-service:
    build:
      context: ./services/document-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-document
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: document-service
      SERVICE_PORT: 8007
      SCHEMA_NAME: document_schema
      STORAGE_PATH: /app/storage/documents
      MAX_FILE_SIZE_MB: 100
    volumes:
      - ./storage/documents:/app/storage/documents:rw
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8007/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.document.entrypoints: "websecure"
      traefik.http.routers.document.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/documents`)"
      traefik.http.routers.document.tls: "true"
      traefik.http.routers.document.tls.certresolver: "letsencrypt"
      traefik.http.services.document.loadbalancer.server.port: 8007

  # ============================================================
  # Microservices: Crew Service (Port 8008)
  # ============================================================
  crew-service:
    build:
      context: ./services/crew-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-crew
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: crew-service
      SERVICE_PORT: 8008
      SCHEMA_NAME: crew_schema
      CALDAV_ENABLED: "true"
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8008/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.crew.entrypoints: "websecure"
      traefik.http.routers.crew.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/crew`)"
      traefik.http.routers.crew.tls: "true"
      traefik.http.routers.crew.tls.certresolver: "letsencrypt"
      traefik.http.services.crew.loadbalancer.server.port: 8008

  # ============================================================
  # Microservices: Federation Service (Port 8009)
  # ============================================================
  federation-service:
    build:
      context: ./services/federation-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-federation
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: federation-service
      SERVICE_PORT: 8009
      SCHEMA_NAME: federation_schema
      MTLS_CERT_PATH: /app/certs/server.crt
      MTLS_KEY_PATH: /app/certs/server.key
      MTLS_CA_PATH: /app/certs/ca.crt
    volumes:
      - ./certs/federation:/app/certs:ro
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8009/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.federation.entrypoints: "websecure"
      traefik.http.routers.federation.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/federation`)"
      traefik.http.routers.federation.tls: "true"
      traefik.http.routers.federation.tls.certresolver: "letsencrypt"
      traefik.http.services.federation.loadbalancer.server.port: 8009

  # ============================================================
  # Microservices: Remaining Services (8010-8017)
  # ============================================================
  maintenance-service:
    build:
      context: ./services/maintenance-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-maintenance
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: maintenance-service
      SERVICE_PORT: 8010
      SCHEMA_NAME: maintenance_schema
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8010/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.maintenance.entrypoints: "websecure"
      traefik.http.routers.maintenance.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/maintenance`)"
      traefik.http.routers.maintenance.tls: "true"
      traefik.http.routers.maintenance.tls.certresolver: "letsencrypt"
      traefik.http.services.maintenance.loadbalancer.server.port: 8010

  transport-service:
    build:
      context: ./services/transport-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-transport
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: transport-service
      SERVICE_PORT: 8011
      SCHEMA_NAME: transport_schema
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8011/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.transport.entrypoints: "websecure"
      traefik.http.routers.transport.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/transport`)"
      traefik.http.routers.transport.tls: "true"
      traefik.http.routers.transport.tls.certresolver: "letsencrypt"
      traefik.http.services.transport.loadbalancer.server.port: 8011

  insurance-service:
    build:
      context: ./services/insurance-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-insurance
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: insurance-service
      SERVICE_PORT: 8012
      SCHEMA_NAME: insurance_schema
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8012/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.insurance.entrypoints: "websecure"
      traefik.http.routers.insurance.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/insurance`)"
      traefik.http.routers.insurance.tls: "true"
      traefik.http.routers.insurance.tls.certresolver: "letsencrypt"
      traefik.http.services.insurance.loadbalancer.server.port: 8012

  workflow-service:
    build:
      context: ./services/workflow-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-workflow
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: workflow-service
      SERVICE_PORT: 8013
      SCHEMA_NAME: workflow_schema
      WORKFLOW_EXECUTION_TIMEOUT: 300
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8013/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.workflow.entrypoints: "websecure"
      traefik.http.routers.workflow.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/workflows`)"
      traefik.http.routers.workflow.tls: "true"
      traefik.http.routers.workflow.tls.certresolver: "letsencrypt"
      traefik.http.services.workflow.loadbalancer.server.port: 8013

  ai-service:
    build:
      context: ./services/ai-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-ai
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: ai-service
      SERVICE_PORT: 8014
      SCHEMA_NAME: ai_schema
      AI_PROVIDERS: ${AI_PROVIDERS:-ollama}
      OLLAMA_ENDPOINT: ${OLLAMA_ENDPOINT:-http://localhost:11434}
      CLAUDE_API_KEY: ${CLAUDE_API_KEY:-}
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.2'
          memory: 256M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8014/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.ai.entrypoints: "websecure"
      traefik.http.routers.ai.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/ai`)"
      traefik.http.routers.ai.tls: "true"
      traefik.http.routers.ai.tls.certresolver: "letsencrypt"
      traefik.http.services.ai.loadbalancer.server.port: 8014

  notification-service:
    build:
      context: ./services/notification-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-notification
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: notification-service
      SERVICE_PORT: 8015
      SCHEMA_NAME: notification_schema
      SMTP_HOST: ${SMTP_HOST:-}
      SMTP_PORT: ${SMTP_PORT:-587}
      SMTP_USER: ${SMTP_USER:-}
      SMTP_PASSWORD: ${SMTP_PASSWORD:-}
      SMTP_FROM: ${SMTP_FROM:-noreply@rentflow.local}
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8015/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.notification.entrypoints: "websecure"
      traefik.http.routers.notification.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/notifications`)"
      traefik.http.routers.notification.tls: "true"
      traefik.http.routers.notification.tls.certresolver: "letsencrypt"
      traefik.http.services.notification.loadbalancer.server.port: 8015

  reporting-service:
    build:
      context: ./services/reporting-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-reporting
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: reporting-service
      SERVICE_PORT: 8016
      SCHEMA_NAME: reporting_schema
      MATERIALIZED_VIEW_REFRESH: 3600
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8016/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.reporting.entrypoints: "websecure"
      traefik.http.routers.reporting.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/reporting`)"
      traefik.http.routers.reporting.tls: "true"
      traefik.http.routers.reporting.tls.certresolver: "letsencrypt"
      traefik.http.services.reporting.loadbalancer.server.port: 8016

  audit-service:
    build:
      context: ./services/audit-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
        BUILD_VERSION: ${APP_VERSION:-dev}
    container_name: rentflow-audit
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      <<: [*common-env, *postgres-conn, *kurrentdb-conn, *redis-conn]
      SERVICE_NAME: audit-service
      SERVICE_PORT: 8017
      SCHEMA_NAME: audit_schema
      GOBD_COMPLIANCE: "true"
      IMMUTABLE_LOG: "true"
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8017/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.audit.entrypoints: "websecure"
      traefik.http.routers.audit.rule: "Host(`${DOMAIN}`) && PathPrefix(`/api/audit`)"
      traefik.http.routers.audit.tls: "true"
      traefik.http.routers.audit.tls.certresolver: "letsencrypt"
      traefik.http.services.audit.loadbalancer.server.port: 8017

  # ============================================================
  # Frontend: React UI
  # ============================================================
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
      args:
        BUILD_VERSION: ${APP_VERSION:-dev}
        VITE_API_BASE_URL: https://${DOMAIN}/api
    container_name: rentflow-frontend
    restart: always
    networks:
      - rentflow
    depends_on:
      - traefik
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
    labels:
      traefik.enable: "true"
      traefik.http.routers.frontend.entrypoints: "websecure"
      traefik.http.routers.frontend.rule: "Host(`${DOMAIN}`) && !PathPrefix(`/api`)"
      traefik.http.routers.frontend.tls: "true"
      traefik.http.routers.frontend.tls.certresolver: "letsencrypt"
      traefik.http.routers.frontend.priority: 1
      traefik.http.services.frontend.loadbalancer.server.port: 80
      traefik.http.middlewares.frontend-gzip.compress: "true"
      traefik.http.routers.frontend.middlewares: "frontend-gzip"

  # ============================================================
  # Observability: Prometheus
  # ============================================================
  prometheus:
    image: prom/prometheus:latest
    container_name: rentflow-prometheus
    restart: always
    networks:
      - rentflow
    ports:
      - "127.0.0.1:9090:9090"
    volumes:
      - ./config/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./config/prometheus/alert-rules.yml:/etc/prometheus/alert-rules.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9090/-/healthy"]
      interval: 10s
      timeout: 3s
      retries: 3

  # ============================================================
  # Observability: Grafana
  # ============================================================
  grafana:
    image: grafana/grafana:latest
    container_name: rentflow-grafana
    restart: always
    networks:
      - rentflow
    ports:
      - "127.0.0.1:3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD:?GRAFANA_PASSWORD required}
      GF_USERS_ALLOW_SIGN_UP: "false"
      GF_LOG_LEVEL: info
      GF_DATE_FORMAT: "DD/MM/YYYY HH:mm:ss"
    volumes:
      - ./config/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./config/grafana/dashboards:/var/lib/grafana/dashboards:ro
      - grafana_data:/var/lib/grafana
    deploy:
      resources:
        limits:
          cpus: '0.3'
          memory: 256M
        reservations:
          cpus: '0.1'
          memory: 128M
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:3000/api/health"]
      interval: 10s
      timeout: 3s
      retries: 3
    labels:
      traefik.enable: "true"
      traefik.http.routers.grafana.entrypoints: "websecure"
      traefik.http.routers.grafana.rule: "Host(`${DOMAIN}`) && PathPrefix(`/admin/grafana`)"
      traefik.http.routers.grafana.tls: "true"
      traefik.http.routers.grafana.tls.certresolver: "letsencrypt"
      traefik.http.services.grafana.loadbalancer.server.port: 3000

  # ============================================================
  # Backup Service
  # ============================================================
  backup-service:
    build:
      context: ./services/backup-service
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.22"
    container_name: rentflow-backup
    restart: always
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
    environment:
      <<: [*postgres-conn, *kurrentdb-conn]
      BACKUP_SCHEDULE: "0 3 * * *"  # 3:00 AM daily
      BACKUP_RETENTION_DAYS: 30
      BACKUP_COMPRESSION: "true"
      BACKUP_ENCRYPTION_KEY: ${BACKUP_ENCRYPTION_KEY:?BACKUP_ENCRYPTION_KEY required}
      S3_ENABLED: ${BACKUP_S3_ENABLED:-false}
      S3_BUCKET: ${BACKUP_S3_BUCKET:-}
      S3_REGION: ${BACKUP_S3_REGION:-eu-central-1}
      S3_ACCESS_KEY: ${BACKUP_S3_ACCESS_KEY:-}
      S3_SECRET_KEY: ${BACKUP_S3_SECRET_KEY:-}
    volumes:
      - ./backups:/app/backups:rw
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8020/health"]
      interval: 10s
      timeout: 3s
      retries: 3
```

---

## 1.2 docker-compose.dev.yml (Development Override)

```yaml
version: '3.9'

services:
  traefik:
    ports:
      - "8080:8080"  # Dashboard accessible
    environment:
      TRAEFIK_CERTIFICATESRESOLVERS_LETSENCRYPT_ACME_EMAIL: "dev@localhost"

  kurrentdb:
    ports:
      - "2113:2113"  # Expose for dev tools
    environment:
      KURRENTDB_LOG_LEVEL: Debug

  postgres:
    ports:
      - "5432:5432"  # Full access for local dev

  redis:
    ports:
      - "6379:6379"  # Full access

  # Hot-reload für alle Services
  auth-service:
    build:
      context: ./services/auth-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/auth-service/cmd:/app/cmd:ro
      - ./services/auth-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8001:8001"
      - "2345:2345"  # Delve debugger
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  inventory-service:
    build:
      context: ./services/inventory-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/inventory-service/cmd:/app/cmd:ro
      - ./services/inventory-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8002:8002"
      - "2346:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  project-service:
    build:
      context: ./services/project-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/project-service/cmd:/app/cmd:ro
      - ./services/project-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8003:8003"
      - "2347:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  scanner-service:
    build:
      context: ./services/scanner-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/scanner-service/cmd:/app/cmd:ro
      - ./services/scanner-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8004:8004"
      - "2348:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  warehouse-service:
    build:
      context: ./services/warehouse-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/warehouse-service/cmd:/app/cmd:ro
      - ./services/warehouse-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8005:8005"
      - "2349:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  invoice-service:
    build:
      context: ./services/invoice-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/invoice-service/cmd:/app/cmd:ro
      - ./services/invoice-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8006:8006"
      - "2350:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  document-service:
    build:
      context: ./services/document-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/document-service/cmd:/app/cmd:ro
      - ./services/document-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
      - ./storage/documents:/app/storage/documents:rw
    ports:
      - "8007:8007"
      - "2351:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  crew-service:
    build:
      context: ./services/crew-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/crew-service/cmd:/app/cmd:ro
      - ./services/crew-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8008:8008"
      - "2352:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  federation-service:
    build:
      context: ./services/federation-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/federation-service/cmd:/app/cmd:ro
      - ./services/federation-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8009:8009"
      - "2353:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  maintenance-service:
    build:
      context: ./services/maintenance-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/maintenance-service/cmd:/app/cmd:ro
      - ./services/maintenance-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8010:8010"
      - "2354:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  transport-service:
    build:
      context: ./services/transport-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/transport-service/cmd:/app/cmd:ro
      - ./services/transport-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8011:8011"
      - "2355:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  insurance-service:
    build:
      context: ./services/insurance-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/insurance-service/cmd:/app/cmd:ro
      - ./services/insurance-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8012:8012"
      - "2356:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  workflow-service:
    build:
      context: ./services/workflow-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/workflow-service/cmd:/app/cmd:ro
      - ./services/workflow-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8013:8013"
      - "2357:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  ai-service:
    build:
      context: ./services/ai-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/ai-service/cmd:/app/cmd:ro
      - ./services/ai-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8014:8014"
      - "2358:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  notification-service:
    build:
      context: ./services/notification-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/notification-service/cmd:/app/cmd:ro
      - ./services/notification-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8015:8015"
      - "2359:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  reporting-service:
    build:
      context: ./services/reporting-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/reporting-service/cmd:/app/cmd:ro
      - ./services/reporting-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8016:8016"
      - "2360:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  audit-service:
    build:
      context: ./services/audit-service
      dockerfile: Dockerfile.dev
    volumes:
      - ./services/audit-service/cmd:/app/cmd:ro
      - ./services/audit-service/internal:/app/internal:ro
      - ./pkg/common:/app/pkg/common:ro
    ports:
      - "8017:8017"
      - "2361:2345"
    command: air -c .air.toml
    environment:
      DEBUG_MODE: "true"
      LOG_LEVEL: debug

  # Hot-reload für Frontend
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.dev
      args:
        VITE_API_BASE_URL: http://localhost:8080
    volumes:
      - ./frontend/src:/app/src:ro
      - ./frontend/index.html:/app/index.html:ro
      - ./frontend/vite.config.ts:/app/vite.config.ts:ro
    ports:
      - "5173:5173"
    command: npm run dev

  # Seed-Daten laden
  init-db:
    build:
      context: ./scripts
      dockerfile: Dockerfile.init
    networks:
      - rentflow
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      <<: *postgres-conn
    volumes:
      - ./scripts/seeds:/app/seeds:ro
    command: /app/init.sh
```

---

## 1.3 .env.example (Complete)

```bash
################################################################################
# ENVIRONMENT CONFIGURATION
################################################################################

# Deployment environment: development, staging, production
ENVIRONMENT=production

# Application version (auto-set during CI/CD)
APP_VERSION=1.0.0
BUILD_COMMIT=unknown

# Domain and SSL
DOMAIN=rentflow.example.com
LETSENCRYPT_EMAIL=admin@rentflow.example.com

# Log level: debug, info, warn, error
LOG_LEVEL=info

################################################################################
# DATABASE CONFIGURATION
################################################################################

# PostgreSQL 16
POSTGRES_USER=rentflow_user
POSTGRES_PASSWORD=change_me_very_secure_password_here
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=rentflow

################################################################################
# KURRENTDB CONFIGURATION
################################################################################

# KurrentDB Event Store
KURRENTDB_HOST=kurrentdb
KURRENTDB_PORT=2113
KURRENTDB_USER=admin
KURRENTDB_PASSWORD=change_me_very_secure_password_here
KURRENTDB_SCHEME=http  # Use https in production with mTLS

################################################################################
# REDIS CONFIGURATION
################################################################################

# Redis 7
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=change_me_very_secure_password_here
REDIS_DB=0

# Cache TTL in seconds
REDIS_CACHE_TTL_SECONDS=300

################################################################################
# SECURITY & JWT
################################################################################

# JWT Secret (min. 32 characters, use: openssl rand -base64 32)
JWT_SECRET=change_me_min_32_characters_base64_encoded
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_DAYS=7

# Session Secret for cookies
SESSION_SECRET=change_me_min_32_characters_base64_encoded

# CORS allowed origins (comma-separated)
ALLOWED_ORIGINS=https://rentflow.example.com,https://app.rentflow.example.com

# Admin user (created on first start)
ADMIN_EMAIL=admin@rentflow.example.com
ADMIN_PASSWORD=change_me_strong_password

# Bcrypt cost factor (10-12 recommended)
BCRYPT_COST=12

# MFA enabled
MFA_ENABLED=true

################################################################################
# TRAEFIK CONFIGURATION
################################################################################

# Dashboard basic auth (user:password hashed with htpasswd)
# Generate: htpasswd -nB admin
DASHBOARD_HTPASSWD=admin:$2y$05$change_me_with_actual_hash

# Entrypoint for Traefik
TRAEFIK_ENTRYPOINT=https

################################################################################
# MONITORING
################################################################################

# Prometheus retention period
PROMETHEUS_RETENTION=30d

# Grafana admin password
GRAFANA_PASSWORD=change_me_strong_password
GRAFANA_LOG_LEVEL=info

################################################################################
# BACKUP CONFIGURATION
################################################################################

# Backup encryption key (min. 32 characters, use: openssl rand -base64 32)
BACKUP_ENCRYPTION_KEY=change_me_min_32_characters_base64_encoded

# Backup schedule (cron format): "0 3 * * *" = daily 3 AM
BACKUP_SCHEDULE=0 3 * * *

# Local backup retention
BACKUP_RETENTION_DAYS=30

# S3 backup destination (optional)
BACKUP_S3_ENABLED=false
BACKUP_S3_BUCKET=rentflow-backups
BACKUP_S3_REGION=eu-central-1
BACKUP_S3_ACCESS_KEY=
BACKUP_S3_SECRET_KEY=

################################################################################
# SERVICE-SPECIFIC CONFIGURATION
################################################################################

# Auth Service
AUTH_SERVICE_PORT=8001
AUTH_SESSION_TTL_MINUTES=1440
AUTH_TENANT_ISOLATION=true

# Inventory Service
INVENTORY_SERVICE_PORT=8002
INVENTORY_CACHE_TTL_SECONDS=300
INVENTORY_BATCH_SIZE_DEFAULT=1000

# Project Service
PROJECT_SERVICE_PORT=8003

# Scanner Service
SCANNER_SERVICE_PORT=8004
SCANNER_OFFLINE_SYNC_ENABLED=true
SCANNER_OFFLINE_CACHE_TTL=3600

# Warehouse Service
WAREHOUSE_SERVICE_PORT=8005

# Invoice Service
INVOICE_SERVICE_PORT=8006
INVOICE_PDF_GENERATION_TIMEOUT=30
INVOICE_DATEV_EXPORT_ENABLED=true

# Document Service
DOCUMENT_SERVICE_PORT=8007
DOCUMENT_STORAGE_PATH=/app/storage/documents
DOCUMENT_MAX_FILE_SIZE_MB=100

# Crew Service
CREW_SERVICE_PORT=8008
CREW_CALDAV_ENABLED=true

# Federation Service
FEDERATION_SERVICE_PORT=8009
FEDERATION_MTLS_CERT_PATH=/app/certs/server.crt
FEDERATION_MTLS_KEY_PATH=/app/certs/server.key
FEDERATION_MTLS_CA_PATH=/app/certs/ca.crt

# Maintenance Service
MAINTENANCE_SERVICE_PORT=8010

# Transport Service
TRANSPORT_SERVICE_PORT=8011

# Insurance Service
INSURANCE_SERVICE_PORT=8012

# Workflow Service
WORKFLOW_SERVICE_PORT=8013
WORKFLOW_EXECUTION_TIMEOUT=300

# AI Service
AI_SERVICE_PORT=8014
AI_PROVIDERS=ollama
OLLAMA_ENDPOINT=http://localhost:11434
CLAUDE_API_KEY=
OPENAI_API_KEY=
MISTRAL_API_KEY=
GEMINI_API_KEY=

# Notification Service
NOTIFICATION_SERVICE_PORT=8015
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=no-reply@example.com
SMTP_PASSWORD=
SMTP_FROM=noreply@rentflow.local
SMTP_TLS_ENABLED=true

# Reporting Service
REPORTING_SERVICE_PORT=8016
REPORTING_MATERIALIZED_VIEW_REFRESH=3600

# Audit Service
AUDIT_SERVICE_PORT=8017
AUDIT_GOBD_COMPLIANCE=true
AUDIT_IMMUTABLE_LOG=true

################################################################################
# FEATURE FLAGS
################################################################################

FEATURE_FEDERATION_ENABLED=true
FEATURE_AI_ENABLED=true
FEATURE_CALDAV_ENABLED=true
FEATURE_DATEV_EXPORT_ENABLED=true
FEATURE_RFID_SCANNER_ENABLED=true
FEATURE_OFFLINE_MODE_ENABLED=true

################################################################################
# OPTIONAL INTEGRATIONS
################################################################################

# DATEV
DATEV_CLIENT_NUMBER=
DATEV_CONSULTANT_NUMBER=
DATEV_SANDBOX=true

# Bank connection (for automated invoice payment detection)
BANK_API_ENABLED=false
BANK_API_ENDPOINT=
BANK_API_KEY=

# SMS notifications
SMS_PROVIDER=twilio  # twilio, messagebird, aws-sns
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_FROM_NUMBER=

# Webhook for external integrations
WEBHOOK_SECRET=change_me_min_32_characters
WEBHOOK_TIMEOUT_SECONDS=10

################################################################################
# DEVELOPMENT SETTINGS (only if ENVIRONMENT=development)
################################################################################

# Debug mode
DEBUG_MODE=false

# Hot reload for services
HOT_RELOAD_ENABLED=true

# Seed data on startup
SEED_DATA=false

# GraphQL introspection
GRAPHQL_INTROSPECTION_ENABLED=false
```

---

## 1.4 Dockerfile pro Service (Go Multi-Stage)

### Base: `services/auth-service/Dockerfile`

```dockerfile
# Stage 1: Builder
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build application
ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
  -a -installsuffix cgo \
  -ldflags="-w -s \
    -X github.com/rentflow/rentflow/internal/version.Version=${BUILD_VERSION} \
    -X github.com/rentflow/rentflow/internal/version.Commit=${BUILD_COMMIT}" \
  -o auth-service ./cmd/main.go

# Stage 2: Runtime (minimal Alpine)
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache curl ca-certificates tzdata

# Create non-root user
RUN addgroup -D -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy binary from builder
COPY --from=builder /build/auth-service /app/auth-service

# Copy migrations if needed
COPY --from=builder /build/migrations /app/migrations

# Change ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=10s --timeout=3s --retries=3 --start-period=10s \
  CMD curl -f http://localhost:8001/health || exit 1

# Expose port (default 8001)
EXPOSE 8001

# Run application
CMD ["./auth-service"]
```

### Development: `services/auth-service/Dockerfile.dev`

```dockerfile
FROM golang:1.22-alpine

WORKDIR /app

# Install dev dependencies
RUN apk add --no-cache git ca-certificates curl build-base air

# Install air (hot-reload tool)
RUN go install github.com/cosmtrek/air@latest

# Copy go files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Install dependencies
RUN go mod tidy

EXPOSE 8001
EXPOSE 2345  # Delve debugger

CMD ["air", "-c", ".air.toml"]
```

---

## 1.5 Resource Limits für 4GB RAM Server (CX21)

```yaml
# Empfohlene Verteilung für 4GB RAM (CX21):

traefik:
  limits:
    cpus: '0.5'
    memory: 256M

kurrentdb:  # Event Store: 1.5-2GB
  limits:
    cpus: '1'
    memory: 1.5G

postgres:   # Read-DB: 1.5-2GB
  limits:
    cpus: '1.5'
    memory: 2G

redis:      # Cache: 512MB
  limits:
    cpus: '0.5'
    memory: 512M

# 17 Go Services: je 256MB
auth-service through audit-service:
  limits:
    cpus: '0.3'
    memory: 256M

# Total: ~6GB with overhead
# Recommendation: CX31 (8GB) for production
```

---

## 2. Traefik v3 Konfiguration

### 2.1 Static Config: `config/traefik/traefik.yml`

```yaml
global:
  checkNewVersion: false
  sendAnonymousUsage: false

# Log configuration
log:
  level: INFO
  format: json

# Access logs
accessLog:
  format: json
  filePath: /var/log/traefik/access.log
  fields:
    defaultMode: keep
    names:
      ClientUsername: drop

# API configuration
api:
  dashboard: true
  debug: false
  insecure: false

# Entrypoints (HTTP und HTTPS)
entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entryPoint:
          scheme: https
          port: 443

  websecure:
    address: ":443"
    http:
      tls:
        certResolver: letsencrypt
        domains:
          - main: ${DOMAIN}
            sans:
              - "*.${DOMAIN}"
      middlewares:
        - securityHeaders@file

# Providers
providers:
  docker:
    endpoint: unix:///var/run/docker.sock
    exposedByDefault: false
    defaultRule: Host(`{{ index .Labels "com.docker.compose.service" }}.{{ env "DOMAIN" }}`)

  file:
    filename: /etc/traefik/dynamic/config.yml
    watch: true

# Certificate Resolvers
certificatesResolvers:
  letsencrypt:
    acme:
      email: ${LETSENCRYPT_EMAIL}
      storage: /letsencrypt/acme.json
      httpChallenge:
        entryPoint: web
      certificatesDuration: 2160h  # 90 days

  selfsigned:
    acme:
      storage: /letsencrypt/self-signed.json
      httpChallenge:
        entryPoint: web
```

### 2.2 Dynamic Config: `config/traefik/dynamic/config.yml`

```yaml
http:
  # Global Middleware
  middlewares:
    # Security headers
    securityHeaders:
      headers:
        accessControlAllowOriginList:
          - https://${DOMAIN}
          - https://*.${DOMAIN}
        accessControlAllowMethods:
          - GET
          - POST
          - PUT
          - DELETE
          - PATCH
          - OPTIONS
        accessControlAllowHeaders:
          - Content-Type
          - Authorization
        accessControlExposeHeaders:
          - X-Total-Count
          - X-Page-Number
        accessControlMaxAge: 86400
        accessControlAllowCredentials: true
        stsSeconds: 31536000
        stsIncludeSubdomains: true
        stsPreload: true
        contentTypeNosniff: true
        browserXssFilter: true
        referrerPolicy: "strict-origin-when-cross-origin"
        permissionsPolicy: "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()"

    # Rate limiting
    rateLimit:
      rateLimit:
        average: 100
        burst: 50
        period: 1s

    # Compression
    gzip:
      compress:
        minResponseBodyBytes: 1024

    # Circuit breaker
    circuitBreaker:
      circuitBreaker:
        expression: "NetworkErrorRatio() > 0.5"
        checkInterval: 100ms
        fallbackDuration: 10s
        maxRequests: 1

    # Retry
    retry:
      retry:
        attempts: 3
        initialInterval: 100ms

    # CORS
    cors:
      headers:
        accessControlAllowOriginList:
          - https://${DOMAIN}
        accessControlAllowMethods:
          - GET
          - POST
          - PUT
          - DELETE
          - OPTIONS
        accessControlAllowHeaders:
          - Content-Type
          - Authorization

  # Routers
  routers:
    # API Router (catch-all for /api/*)
    api-router:
      rule: "Host(`{{ env \"DOMAIN\" }}`) && PathPrefix(`/api`)"
      entryPoints:
        - websecure
      service: api-service
      tls:
        certResolver: letsencrypt
      middlewares:
        - securityHeaders
        - rateLimit
        - circuitBreaker
        - retry
        - gzip

    # Frontend Router
    frontend-router:
      rule: "Host(`{{ env \"DOMAIN\" }}`)"
      entryPoints:
        - websecure
      service: frontend-service
      tls:
        certResolver: letsencrypt
      middlewares:
        - securityHeaders
        - gzip
      priority: 1

    # Dashboard (admin only)
    dashboard:
      rule: "Host(`{{ env \"DOMAIN\" }}`) && PathPrefix(`/admin/dashboard`)"
      entryPoints:
        - websecure
      service: api@internal
      tls:
        certResolver: letsencrypt
      middlewares:
        - dashboard-auth
        - securityHeaders

    # Prometheus (metrics, internal only)
    prometheus:
      rule: "Host(`{{ env \"DOMAIN\" }}`) && PathPrefix(`/metrics`)"
      entryPoints:
        - websecure
      service: prometheus-service
      tls:
        certResolver: letsencrypt
      middlewares:
        - metrics-auth

  # Services
  services:
    api-service:
      weighted:
        services:
          - name: auth-service
            weight: 1
          - name: inventory-service
            weight: 1
          - name: project-service
            weight: 1

    frontend-service:
      loadBalancer:
        servers:
          - url: http://frontend:80

    prometheus-service:
      loadBalancer:
        servers:
          - url: http://prometheus:9090

  # Basic Auth for Dashboard
  middlewares:
    dashboard-auth:
      basicAuth:
        users:
          - "admin:${DASHBOARD_HTPASSWD}"

    metrics-auth:
      basicAuth:
        users:
          - "metrics:${METRICS_HTPASSWD}"
```

---

## 3. CI/CD Pipeline (GitHub Actions)

### 3.1 `.github/workflows/ci.yml`

```yaml
name: CI Pipeline

on:
  push:
    branches:
      - develop
      - main
    paths:
      - 'services/**'
      - 'frontend/**'
      - '.github/workflows/ci.yml'
  pull_request:
    branches:
      - main
    paths:
      - 'services/**'
      - 'frontend/**'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test-go-services:
    name: Test Go Services
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service:
          - auth-service
          - inventory-service
          - project-service
          - scanner-service
          - warehouse-service
          - invoice-service
          - document-service
          - crew-service
          - federation-service
          - maintenance-service
          - transport-service
          - insurance-service
          - workflow-service
          - ai-service
          - notification-service
          - reporting-service
          - audit-service

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: rentflow_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      kurrentdb:
        image: kurrent/kurrentdb:26.0
        env:
          KURRENTDB_DEFAULT_ADMIN_USER_PASSWORD: test_password
        options: >-
          --health-cmd "curl -f http://localhost:2113/health/live"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 2113:2113

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
          cache: true
          cache-dependency-path: 'services/${{ matrix.service }}/go.sum'

      - name: Run unit tests
        run: |
          cd services/${{ matrix.service }}
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out

      - name: Run integration tests
        env:
          POSTGRES_DSN: postgresql://test_user:test_password@localhost:5432/rentflow_test
          KURRENTDB_URL: http://localhost:2113
        run: |
          cd services/${{ matrix.service }}
          go test -v -race -tags=integration -timeout=10m ./...

      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          files: ./services/${{ matrix.service }}/coverage.out
          flags: ${{ matrix.service }}
          fail_ci_if_error: false

  test-frontend:
    name: Test Frontend
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: 'frontend/package-lock.json'

      - name: Install dependencies
        run: |
          cd frontend
          npm ci

      - name: Run lint
        run: |
          cd frontend
          npm run lint

      - name: Run tests
        run: |
          cd frontend
          npm test -- --coverage

      - name: Build frontend
        run: |
          cd frontend
          npm run build

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./frontend/coverage/lcov.info
          flags: frontend
          fail_ci_if_error: false

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [auth-service, inventory-service, project-service]

    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner (Go)
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: 'services/${{ matrix.service }}'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Run gosec (Go security)
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt=json -out=gosec-results.json ./services/${{ matrix.service }}/...'

      - name: Upload Trivy results
        uses: github/codeql-action/upload-sarif@v2
        if: always()
        with:
          sarif_file: 'trivy-results.sarif'
          category: 'Trivy-${{ matrix.service }}'

  build-docker:
    name: Build Docker Images
    needs: [test-go-services, test-frontend, security-scan]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    strategy:
      matrix:
        service:
          - auth-service
          - inventory-service
          - project-service
          - scanner-service
          - warehouse-service
          - invoice-service
          - document-service
          - crew-service
          - federation-service
          - maintenance-service
          - transport-service
          - insurance-service
          - workflow-service
          - ai-service
          - notification-service
          - reporting-service
          - audit-service

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Log in to Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push service image
        uses: docker/build-push-action@v4
        with:
          context: ./services/${{ matrix.service }}
          file: ./services/${{ matrix.service }}/Dockerfile
          push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/develop' }}
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/services/${{ matrix.service }}:latest
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/services/${{ matrix.service }}:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  build-frontend:
    name: Build Frontend Image
    needs: [test-frontend]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Log in to Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push frontend image
        uses: docker/build-push-action@v4
        with:
          context: ./frontend
          file: ./frontend/Dockerfile.prod
          push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/develop' }}
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/frontend:latest
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/frontend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### 3.2 `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write

    outputs:
      version: ${{ steps.get_version.outputs.version }}

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Get version
        id: get_version
        run: echo "version=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Generate changelog
        id: changelog
        run: |
          VERSION=${{ steps.get_version.outputs.version }}
          git log $(git describe --tags --abbrev=0 HEAD^)..HEAD --oneline --format="%h %s" > CHANGELOG.tmp
          cat CHANGELOG.tmp

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          tag_name: ${{ steps.get_version.outputs.version }}
          body_path: CHANGELOG.tmp
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  push-images:
    name: Push Release Images
    needs: release
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    strategy:
      matrix:
        component: [auth-service, inventory-service, project-service, scanner-service, warehouse-service, invoice-service, document-service, crew-service, federation-service, maintenance-service, transport-service, insurance-service, workflow-service, ai-service, notification-service, reporting-service, audit-service, frontend]

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Log in to Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push image
        uses: docker/build-push-action@v4
        with:
          context: ./${{ matrix.component == 'frontend' && 'frontend' || format('services/{0}', matrix.component) }}
          file: ./${{ matrix.component == 'frontend' && 'frontend/Dockerfile.prod' || format('services/{0}/Dockerfile', matrix.component) }}
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/${{ matrix.component == 'frontend' && 'frontend' || format('services/{0}', matrix.component) }}:${{ needs.release.outputs.version }}
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/${{ matrix.component == 'frontend' && 'frontend' || format('services/{0}', matrix.component) }}:latest
```

### 3.3 `.github/workflows/deploy.yml`

```yaml
name: Deploy to Production

on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy (e.g., v1.2.3)'
        required: true

jobs:
  deploy:
    name: Deploy
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://${{ secrets.DEPLOY_DOMAIN }}

    steps:
      - uses: actions/checkout@v4

      - name: SSH and deploy
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.DEPLOY_HOST }}
          username: ${{ secrets.DEPLOY_USER }}
          key: ${{ secrets.DEPLOY_SSH_KEY }}
          script: |
            set -e
            cd /opt/rentflow

            # Backup current state
            docker compose config > docker-compose.backup.yml
            docker exec rentflow-postgres pg_dump -U rentflow_user rentflow > backups/pre-deploy-$(date +%s).sql

            # Pull new images
            export REGISTRY=ghcr.io
            export IMAGE_NAME=${{ github.repository }}
            export APP_VERSION=${{ github.event.inputs.version }}
            docker compose pull

            # Run migrations
            docker compose run --rm migrate

            # Rolling restart
            docker compose up -d

            # Health check
            sleep 30
            for service in auth-service inventory-service project-service scanner-service warehouse-service invoice-service document-service crew-service federation-service maintenance-service transport-service insurance-service workflow-service ai-service notification-service reporting-service audit-service; do
              if ! curl -f http://localhost:8001/health > /dev/null 2>&1; then
                echo "Health check failed for $service, rolling back..."
                docker compose config < docker-compose.backup.yml > docker-compose.yml
                docker compose up -d
                exit 1
              fi
            done

            echo "Deployment successful!"

      - name: Notify deployment
        if: always()
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Deployment of ${{ github.event.inputs.version }} to production'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

---

## 4. Monitoring & Observability

### 4.1 Prometheus Config: `config/prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    environment: production
    cluster: rentflow

alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093

rule_files:
  - /etc/prometheus/alert-rules.yml

scrape_configs:
  - job_name: traefik
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics

  - job_name: 'go-services'
    static_configs:
      - targets:
          - 'auth-service:8001'
          - 'inventory-service:8002'
          - 'project-service:8003'
          - 'scanner-service:8004'
          - 'warehouse-service:8005'
          - 'invoice-service:8006'
          - 'document-service:8007'
          - 'crew-service:8008'
          - 'federation-service:8009'
          - 'maintenance-service:8010'
          - 'transport-service:8011'
          - 'insurance-service:8012'
          - 'workflow-service:8013'
          - 'ai-service:8014'
          - 'notification-service:8015'
          - 'reporting-service:8016'
          - 'audit-service:8017'
    metrics_path: /metrics

  - job_name: 'postgresql'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'kurrentdb'
    static_configs:
      - targets: ['kurrentdb:2113']
    metrics_path: /metrics

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
```

### 4.2 Alert Rules: `config/prometheus/alert-rules.yml`

```yaml
groups:
  - name: service_health
    interval: 30s
    rules:
      - alert: ServiceDown
        expr: up{job="go-services"} == 0
        for: 1m

      - alert: HighErrorRate
        expr: rate(http_request_total{status=~"5.."}[5m]) > 0.05
        for: 5m

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m

  - name: infrastructure
    interval: 30s
    rules:
      - alert: DiskUsageHigh
        expr: node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes < 0.15
        for: 5m

      - alert: MemoryUsageHigh
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) > 0.85
        for: 5m

  - name: database
    interval: 30s
    rules:
      - alert: PostgreSQLDown
        expr: pg_up{job="postgresql"} == 0
        for: 1m

      - alert: PostgreSQLSlowQueries
        expr: rate(pg_slow_queries_total[5m]) > 0.1
        for: 5m

  - name: events
    interval: 30s
    rules:
      - alert: KurrentDBSubscriptionLag
        expr: kurrentdb_subscription_lag > 1000
        for: 5m
```

### 4.3 Grafana Dashboards

**Dashboards:**
1. Service Health Overview — alle 17 Services, Request Rate, Error Rate, p50/p95/p99 Latencies
2. Request Performance — HTTP Duration, Errors, Slowest Endpoints
3. Business Metrics — Active Projects, Scans/Min, Equipment Utilization %, Outstanding Invoices
4. Infrastructure — CPU, RAM, Disk, Network I/O, Container Restarts
5. Database — Query Performance, Connection Pool, Cache Hit Ratio, Slow Queries

### 4.4 Structured Logging (zerolog JSON Format)

Log entry example:
```json
{
  "level": "info",
  "timestamp": "2026-03-21T10:30:45.123Z",
  "service": "auth-service",
  "correlation_id": "abc-123-def",
  "tenant_id": "tenant-001",
  "user_id": "user-123",
  "msg": "User login successful",
  "method": "POST",
  "path": "/api/auth/login",
  "status_code": 200,
  "duration_ms": 45
}
```

---

## 5. Testing Strategy

### 5.1 Unit Tests (Go + testify)

**Naming:** `Test{Function}_{Scenario}_{ExpectedResult}`
**Coverage-Ziel:** >80% Domain Logic, >60% Adapters
**Mocks:** testify/mock oder mockery-generiert
**Fixtures:** testcontainers-go für DB-Tests

Example:
```go
func TestCreateUser_DuplicateEmail_ReturnsError(t *testing.T) {
  repo := new(MockUserRepository)
  repo.On("GetByEmail", mock.Anything, "test@example.com").Return(&User{}, nil)

  svc := NewUserService(repo)
  _, err := svc.CreateUser(context.Background(), "test@example.com", "pw")

  assert.ErrorIs(t, err, ErrUserAlreadyExists)
}
```

### 5.2 Integration Tests (testcontainers-go)

```bash
go test -tags=integration -timeout=10m ./...
```

Startet PostgreSQL + KurrentDB + Redis Container für echte API-Tests.

### 5.3 E2E Tests (Playwright)

**Test-Szenarien pro Persona:**
- Marco: Login → Dashboard → Equipment erstellen → Projekt
- Lisa: Login → Scanner → Packliste → Checkout → Check-in
- Thomas: Rechnungsstellung → DATEV-Export
- Kevin: Mobile Login → Zeiterfassung

```bash
cd frontend && npm run test:e2e
npx playwright show-trace trace.zip
```

### 5.4 Load Tests (k6)

**Normal:** 20 Users, 100 Scans/Min → p95 <200ms
**Peak:** 50 Users, 500 Scans/Min → p95 <200ms
**Stress:** Finding breaking point

```bash
k6 run tests/load/scanner-load.js \
  -e BASE_URL=https://rentflow.local \
  -e AUTH_TOKEN=eyJ... \
  -o json=results.json
```

---

## 6. Performance Optimization

### 6.1 Go Backend

**Connection Pooling:**
- pgxpool: MaxConns=100, MinConns=10
- Connection timeout: 10s
- Idle timeout: 5min

**Caching:**
- Equipment availability: Redis TTL 30s
- User permissions: Redis TTL 5min
- Dashboard aggregations: Redis TTL 60s

**Query Optimization:**
```sql
CREATE INDEX idx_equipment_status ON inventory_schema.equipment(status);
CREATE INDEX idx_project_tenant_date ON project_schema.projects(tenant_id, start_date);
PARTITION events_log BY MONTH;
```

### 6.2 Frontend

**Bundle Optimization:**
- Tree-shaking, code splitting
- Lazy loading (routes + components)
- Image WebP + srcset + lazy loading
- Critical CSS inline

### 6.3 PostgreSQL Tuning (4GB RAM)

```
shared_buffers = 1GB (25% of RAM)
effective_cache_size = 3GB (75% of RAM)
work_mem = 16MB
maintenance_work_mem = 256MB
random_page_cost = 1.1
```

---

## 7. Auto-Update System

**Workflow:**
1. Daily check GitHub Releases API
2. Admin Dashboard banner with changelog
3. One-click update: Backup → Pull Images → Migrate → Restart → Health Check
4. Automatic rollback on failure

---

## 8. Backup System

**Automated Daily (03:00 UTC):**
- PostgreSQL: pg_dump --format=custom, 9x compression
- KurrentDB: Event store backup
- Uploads: rsync
- Encryption: AES-256-GCM
- Retention: 7 daily, 4 weekly, 3 monthly
- Optional: S3 sync

**Restore Test:** Monthly automatic in separate DB

---

**Status:** SPEC COMPLETE — Production-ready Implementation Details

