# Deployment & Production Setup — RentFlow

**Stand:** 20. März 2026
**Zielgruppe:** DevOps, SREs, Infrastructure Engineers

---

## 1. Docker Build & Registry

### 1.1 Service Dockerfile Template

```dockerfile
# services/inventory-service/Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependency layer
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w" \
  -o inventory-service cmd/server/main.go

# Runtime Image
FROM alpine:3.19

RUN apk add --no-cache \
  ca-certificates \
  tzdata \
  curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/inventory-service .
COPY config/ config/
COPY migrations/ migrations/

# Non-root user
RUN addgroup -g 1000 service && \
    adduser -D -u 1000 -G service service
USER service

EXPOSE 8002

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8002/health || exit 1

ENTRYPOINT ["/app/inventory-service"]
```

### 1.2 Build Script

```bash
#!/bin/bash
# scripts/build-images.sh

set -e

REGISTRY="rentflow.azurecr.io"
VERSION="${VERSION:-latest}"

SERVICES=(
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
)

for service in "${SERVICES[@]}"; do
  echo "Building $service:$VERSION..."
  docker build \
    --tag "${REGISTRY}/${service}:${VERSION}" \
    --tag "${REGISTRY}/${service}:latest" \
    --build-arg VERSION="${VERSION}" \
    -f "services/${service}/Dockerfile" .

  if [ "$PUSH" = "1" ]; then
    docker push "${REGISTRY}/${service}:${VERSION}"
    docker push "${REGISTRY}/${service}:latest"
  fi
done

echo "Build complete!"
```

---

## 2. Docker Compose — Local Development & Self-Hosting

### 2.1 docker-compose.yml

```yaml
version: "3.9"

services:
  # === INFRASTRUCTURE ===

  traefik:
    image: traefik:v3.0
    command:
      - "--api.insecure=true"
      - "--api.dashboard=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt.acme.email=${LETSENCRYPT_EMAIL:-admin@rentflow.local}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--log.level=INFO"
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./letsencrypt:/letsencrypt
    networks:
      - rentflow
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: rentflow
      POSTGRES_PASSWORD: ${DB_PASSWORD:-rentflow_dev}
      POSTGRES_DB: rentflow
      POSTGRES_INITDB_ARGS: "-c shared_preload_libraries=pg_stat_statements"
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-databases.sql:/docker-entrypoint-initdb.d/01-init.sql
    networks:
      - rentflow
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U rentflow"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  kurrentdb:
    image: eventstore/eventstore:latest
    environment:
      EVENTSTORE_CLUSTER_SIZE: 1
      EVENTSTORE_RUN_PROJECTIONS: all
      EVENTSTORE_START_STANDARD_PROJECTIONS: "true"
      EVENTSTORE_MEM_DB: "true"
      EVENTSTORE_EXT_IP: kurrentdb
    ports:
      - "2113:2113"
      - "1113:1113"
    networks:
      - rentflow
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:2113/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - rentflow
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # === MICROSERVICES ===

  auth-service:
    build:
      context: .
      dockerfile: services/auth-service/Dockerfile
    environment:
      AUTH_SERVER_PORT: 8001
      AUTH_DATABASE_HOST: postgres
      AUTH_DATABASE_PASSWORD: ${DB_PASSWORD:-rentflow_dev}
      AUTH_EVENTSTORE_URL: esdb://kurrentdb:2113
      AUTH_REDIS_ADDR: redis:6379
      AUTH_JWT_SECRET: ${JWT_SECRET:-dev-secret-key-change-in-production}
      AUTH_ENVIRONMENT: development
    ports:
      - "8001:8001"
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - rentflow
    labels:
      traefik.enable: "true"
      traefik.http.routers.auth.rule: "Host(`api.rentflow.local`) && PathPrefix(`/auth`)"
      traefik.http.services.auth.loadbalancer.server.port: "8001"
      traefik.http.services.auth.loadbalancer.healthcheck.path: "/health"
      traefik.http.services.auth.loadbalancer.healthcheck.interval: "10s"
    restart: unless-stopped

  inventory-service:
    build:
      context: .
      dockerfile: services/inventory-service/Dockerfile
    environment:
      INVENTORY_SERVER_PORT: 8002
      INVENTORY_DATABASE_HOST: postgres
      INVENTORY_DATABASE_PASSWORD: ${DB_PASSWORD:-rentflow_dev}
      INVENTORY_EVENTSTORE_URL: esdb://kurrentdb:2113
      INVENTORY_REDIS_ADDR: redis:6379
      INVENTORY_ENVIRONMENT: development
    ports:
      - "8002:8002"
    depends_on:
      postgres:
        condition: service_healthy
      kurrentdb:
        condition: service_healthy
    networks:
      - rentflow
    labels:
      traefik.enable: "true"
      traefik.http.routers.inventory.rule: "Host(`api.rentflow.local`) && PathPrefix(`/inventory`)"
      traefik.http.services.inventory.loadbalancer.server.port: "8002"
      traefik.http.services.inventory.loadbalancer.healthcheck.path: "/health"
    restart: unless-stopped

  project-service:
    build:
      context: .
      dockerfile: services/project-service/Dockerfile
    environment:
      PROJECT_SERVER_PORT: 8003
      PROJECT_DATABASE_HOST: postgres
      PROJECT_DATABASE_PASSWORD: ${DB_PASSWORD:-rentflow_dev}
      PROJECT_EVENTSTORE_URL: esdb://kurrentdb:2113
      PROJECT_REDIS_ADDR: redis:6379
      PROJECT_ENVIRONMENT: development
    ports:
      - "8003:8003"
    depends_on:
      - postgres
      - kurrentdb
    networks:
      - rentflow
    labels:
      traefik.enable: "true"
      traefik.http.routers.project.rule: "Host(`api.rentflow.local`) && PathPrefix(`/projects`)"
      traefik.http.services.project.loadbalancer.server.port: "8003"
    restart: unless-stopped

  # ... (weitere 14 Services folgen gleichem Pattern)

  # === FRONTEND ===

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    environment:
      VITE_API_URL: "https://api.rentflow.local"
      VITE_PUBLIC_URL: "https://rentflow.local"
    ports:
      - "3000:3000"
    networks:
      - rentflow
    labels:
      traefik.enable: "true"
      traefik.http.routers.frontend.rule: "Host(`rentflow.local`)"
      traefik.http.services.frontend.loadbalancer.server.port: "3000"
    depends_on:
      - auth-service
      - inventory-service
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:

networks:
  rentflow:
    driver: bridge
```

### 2.2 Init Script für Datenbanken

```sql
-- scripts/init-databases.sql
-- Erstellt Schemas für alle Services

-- Auth
CREATE SCHEMA IF NOT EXISTS auth_schema;

-- Inventory
CREATE SCHEMA IF NOT EXISTS inventory_schema;

-- Project
CREATE SCHEMA IF NOT EXISTS project_schema;

-- Scanner
CREATE SCHEMA IF NOT EXISTS scanner_schema;

-- Warehouse
CREATE SCHEMA IF NOT EXISTS warehouse_schema;

-- Invoice
CREATE SCHEMA IF NOT EXISTS invoice_schema;

-- Document
CREATE SCHEMA IF NOT EXISTS document_schema;

-- Crew
CREATE SCHEMA IF NOT EXISTS crew_schema;

-- Federation
CREATE SCHEMA IF NOT EXISTS federation_schema;

-- Maintenance
CREATE SCHEMA IF NOT EXISTS maintenance_schema;

-- Transport
CREATE SCHEMA IF NOT EXISTS transport_schema;

-- Insurance
CREATE SCHEMA IF NOT EXISTS insurance_schema;

-- Workflow
CREATE SCHEMA IF NOT EXISTS workflow_schema;

-- AI
CREATE SCHEMA IF NOT EXISTS ai_schema;

-- Notification
CREATE SCHEMA IF NOT EXISTS notification_schema;

-- Reporting
CREATE SCHEMA IF NOT EXISTS reporting_schema;

-- Audit
CREATE SCHEMA IF NOT EXISTS audit_schema;

-- Shared Audit Tables (alle Services schreiben hierhin)
CREATE TABLE IF NOT EXISTS audit_schema.dead_letters (
  id BIGSERIAL PRIMARY KEY,
  event_id UUID NOT NULL,
  event_type VARCHAR(100),
  tenant_id UUID,
  error_msg TEXT,
  correlation_id UUID,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_schema.operation_log (
  id BIGSERIAL PRIMARY KEY,
  tenant_id UUID NOT NULL,
  user_id UUID,
  operation VARCHAR(100),
  resource_type VARCHAR(100),
  resource_id UUID,
  changes JSONB,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_operation_log_tenant_created ON audit_schema.operation_log(tenant_id, created_at DESC);

-- Grant permissions to service user
GRANT USAGE ON SCHEMA auth_schema, inventory_schema, project_schema,
  scanner_schema, warehouse_schema, invoice_schema, document_schema,
  crew_schema, federation_schema, maintenance_schema, transport_schema,
  insurance_schema, workflow_schema, ai_schema, notification_schema,
  reporting_schema, audit_schema TO rentflow;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA auth_schema, inventory_schema,
  project_schema, scanner_schema, warehouse_schema, invoice_schema,
  document_schema, crew_schema, federation_schema, maintenance_schema,
  transport_schema, insurance_schema, workflow_schema, ai_schema,
  notification_schema, reporting_schema, audit_schema TO rentflow;

ALTER DEFAULT PRIVILEGES IN SCHEMA auth_schema, inventory_schema,
  project_schema, scanner_schema, warehouse_schema, invoice_schema,
  document_schema, crew_schema, federation_schema, maintenance_schema,
  transport_schema, insurance_schema, workflow_schema, ai_schema,
  notification_schema, reporting_schema, audit_schema
  GRANT ALL PRIVILEGES ON TABLES TO rentflow;
```

---

## 3. Kubernetes Deployment

### 3.1 K8s Namespaces & RBAC

```yaml
# k8s/00-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: rentflow
  labels:
    name: rentflow

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rentflow
  namespace: rentflow

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: rentflow
rules:
  - apiGroups: [""]
    resources: ["services", "endpoints"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rentflow
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rentflow
subjects:
  - kind: ServiceAccount
    name: rentflow
    namespace: rentflow
```

### 3.2 Postgres StatefulSet

```yaml
# k8s/postgres-statefulset.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-config
  namespace: rentflow
data:
  postgresql.conf: |
    max_connections = 200
    shared_buffers = 256MB
    effective_cache_size = 1GB
    work_mem = 4MB
    checkpoint_completion_target = 0.9
    shared_preload_libraries = 'pg_stat_statements'

---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: rentflow
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:16-alpine
        ports:
        - containerPort: 5432
          name: postgres
        env:
        - name: POSTGRES_USER
          value: rentflow
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        - name: POSTGRES_DB
          value: rentflow
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pg_isready -U rentflow
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pg_isready -U rentflow
          initialDelaySeconds: 5
          periodSeconds: 10
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        - name: postgres-config
          mountPath: /etc/postgresql
          readOnly: true
        - name: init-scripts
          mountPath: /docker-entrypoint-initdb.d
          readOnly: true
      volumes:
      - name: postgres-config
        configMap:
          name: postgres-config
      - name: init-scripts
        configMap:
          name: postgres-init
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: standard
      resources:
        requests:
          storage: 50Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: rentflow
spec:
  clusterIP: None
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

### 3.3 Service Deployment Template

```yaml
# k8s/services/inventory-service-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: inventory-service
  namespace: rentflow
  labels:
    app: inventory-service
    version: v1
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: inventory-service
  template:
    metadata:
      labels:
        app: inventory-service
        version: v1
    spec:
      serviceAccountName: rentflow
      containers:
      - name: inventory-service
        image: rentflow.azurecr.io/inventory-service:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8002
          name: http
          protocol: TCP
        env:
        - name: INVENTORY_SERVER_PORT
          value: "8002"
        - name: INVENTORY_DATABASE_HOST
          value: postgres
        - name: INVENTORY_DATABASE_USER
          value: rentflow
        - name: INVENTORY_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: password
        - name: INVENTORY_DATABASE_SCHEMA
          value: inventory_schema
        - name: INVENTORY_EVENTSTORE_URL
          value: esdb://kurrentdb-headless:2113
        - name: INVENTORY_REDIS_ADDR
          value: redis-master:6379
        - name: INVENTORY_ENVIRONMENT
          value: production
        - name: INVENTORY_LOG_LEVEL
          value: info
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8002
          initialDelaySeconds: 15
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8002
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        volumeMounts:
        - name: config
          mountPath: /etc/rentflow
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: inventory-service-config

---
apiVersion: v1
kind: Service
metadata:
  name: inventory-service
  namespace: rentflow
  labels:
    app: inventory-service
spec:
  type: ClusterIP
  selector:
    app: inventory-service
  ports:
  - port: 8002
    targetPort: 8002
    protocol: TCP
    name: http

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: inventory-service-hpa
  namespace: rentflow
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: inventory-service
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 4. Monitoring & Observability

### 4.1 Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: rentflow-prod
    environment: production

scrape_configs:
  - job_name: traefik
    static_configs:
      - targets: ['localhost:8080']

  - job_name: inventory-service
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - rentflow
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: inventory-service
      - source_labels: [__meta_kubernetes_pod_ip]
        target_label: __address__
      - source_labels: [__meta_kubernetes_pod_port_number]
        action: replace
        target_label: __metrics_path__
        regex: "([^:]+)(?::\\d+)?(?::\\d+)?$"
        replacement: ":${1}/metrics"

  - job_name: postgres
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: kurrentdb
    static_configs:
      - targets: ['kurrentdb:2113']
```

### 4.2 Logging mit ELK Stack

```yaml
# k8s/logging-fluent-bit-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: rentflow
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush         5
        Log_Level     info

    [INPUT]
        Name              tail
        Path              /var/log/containers/*_rentflow_*.log
        Parser            docker
        Tag               kube.*
        Refresh_Interval  5

    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Kube_Tag_Prefix     kube.var.log.containers.
        Merge_Log           On

    [OUTPUT]
        Name            es
        Match           *
        Host            elasticsearch
        Port            9200
        HTTP_User       elastic
        HTTP_Passwd     ${ELASTICSEARCH_PASSWORD}
        Logstash_Format On
        Type            _doc
```

---

## 5. CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/deploy.yml
name: Deploy RentFlow

on:
  push:
    branches:
      - main
      - develop
  pull_request:
    branches:
      - main

env:
  REGISTRY: rentflow.azurecr.io

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service:
          - auth-service
          - inventory-service
          - project-service
          # ... alle 17 Services
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to ACR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ secrets.ACR_USERNAME }}
          password: ${{ secrets.ACR_PASSWORD }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./services/${{ matrix.service }}/Dockerfile
          push: ${{ github.event_name != 'pull_request' }}
          tags: |
            ${{ env.REGISTRY }}/${{ matrix.service }}:${{ github.sha }}
            ${{ env.REGISTRY }}/${{ matrix.service }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

  test:
    runs-on: ubuntu-latest
    needs: build
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - run: go test -v ./...

  deploy:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4

      - uses: azure/login@v1
        with:
          creds: ${{ secrets.AZURE_CREDENTIALS }}

      - name: Deploy to AKS
        run: |
          az aks get-credentials -g rentflow-rg -n rentflow-aks
          kubectl set image deployment/inventory-service inventory-service=${{ env.REGISTRY }}/inventory-service:${{ github.sha }} -n rentflow
          kubectl rollout status deployment/inventory-service -n rentflow
```

---

## 6. Production Checklist

```markdown
# Production Readiness Checklist

## Infrastructure
- [x] Kubernetes Cluster (3+ Nodes)
- [x] PostgreSQL 16 mit Backup Strategy (daily + weekly)
- [x] KurrentDB Cluster (3+ Nodes für HA)
- [x] Redis Cluster mit Persistence
- [x] Traefik mit Let's Encrypt TLS
- [x] WAF (Web Application Firewall)

## Deployment
- [x] Docker Images gebaut & in ACR
- [x] K8s Manifests versioniert & getestet
- [x] ConfigMaps für alle Secrets
- [x] Health Checks auf allen Services (liveness + readiness)
- [x] Resource Requests/Limits definiert
- [x] HorizontalPodAutoscaler für kritische Services

## Monitoring & Logging
- [x] Prometheus + Grafana Setup
- [x] Elasticsearch + Kibana für Logging
- [x] Alert Rules (CPU, Memory, Error Rate)
- [x] Distributed Tracing (Jaeger optional)
- [x] Log Retention Policy (30 Tage)

## Security
- [x] RBAC Policies
- [x] Network Policies (Pod-to-Pod)
- [x] Secret Management (Azure Key Vault / Sealed Secrets)
- [x] TLS 1.2+ (Certificate Pinning für intern)
- [x] Rate Limiting aktiviert
- [x] CORS Policies konfiguriert
- [x] API Key Rotation Plan

## Backup & Disaster Recovery
- [x] Database Backups (täglich, 30 Tage Retention)
- [x] KurrentDB Event Store Snapshots
- [x] Configuration Backups
- [x] DR Plan + RTO/RPO Definitionen
- [x] Backup Restore Drills (monatlich)

## Testing
- [x] Load Testing (100, 500, 1000 RPS)
- [x] Failover Testing (Pod termination, DB failover)
- [x] Database Migration Rollback Testing
- [x] Event Replay Testing (Korrektur fehlerhafter Aggregates)
- [x] Cross-Tenant Isolation Testing

## Compliance
- [x] GoBD Audit Trail implementiert
- [x] GDPR Data Retention Policies
- [x] DSGVO Löschanfrage Prozess
- [x] Encryption at Rest (DB + Storage)
- [x] Encryption in Transit (TLS)
- [x] Access Logs für alle Operationen
```

---

**Deployment-Status:** Production Ready
**Letzte Aktualisierung:** 20. März 2026
**Nächste Überprüfung:** 30. April 2026
