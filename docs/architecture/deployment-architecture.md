# Deployment-Architektur: Veranstaltungstechnik Lagerverwaltung

## Übersicht

Diese Dokumentation beschreibt die Production-ready Deployment-Architektur für die Lagerverwaltungs- und Rechnungssoftware. Das System ist optimiert für kleine Veranstaltungstechnik-Firmen (3-15 Mitarbeiter) mit minimalen DevOps-Kenntnissen.

**Kernziele:**
- One-Click Deployment: `docker compose up -d`
- Automatisierte SSL, Backups, Updates, Monitoring
- Läuft auf Hetzner CX21 (4GB RAM, 40GB SSD)
- Keine externen Services erforderlich (Redis, RabbitMQ)

---

## 1. Docker Compose Stack

### 1.1 Production docker-compose.yml

```yaml
version: '3.9'

services:
  # PostgreSQL Datenbankserver
  db:
    image: postgres:16-alpine
    container_name: vt-db
    restart: always
    environment:
      POSTGRES_USER: ${DB_USER:-vtadmin}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme_required}
      POSTGRES_DB: ${DB_NAME:-vt_inventory}
      POSTGRES_INITDB_ARGS: "-c shared_buffers=256MB -c max_connections=100 -c effective_cache_size=1GB"
    volumes:
      # Persistente Datenspeicherung
      - db_data:/var/lib/postgresql/data
      # Backup-Verzeichnis (für pg_dump)
      - ./backups:/backups:rw
      # PostgreSQL Config für Tuning
      - ./postgres.conf:/etc/postgresql/postgresql.conf:ro
    ports:
      # Nur lokal zugänglich (Sicherheit)
      - "127.0.0.1:5432:5432"
    networks:
      - vt_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-vtadmin} -d ${DB_NAME:-vt_inventory}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 20s
    # Resource-Limits für kleine Server
    deploy:
      resources:
        limits:
          cpus: '1.5'
          memory: 1.5G
        reservations:
          cpus: '0.5'
          memory: 512M

  # Go Backend + React Frontend (Single Container)
  app:
    build:
      context: .
      dockerfile: Dockerfile.prod
      args:
        # Build-Arguments für Versioning
        BUILD_VERSION: ${APP_VERSION:-dev}
        BUILD_COMMIT: ${BUILD_COMMIT:-unknown}
    container_name: vt-app
    restart: always
    environment:
      # Datenbankverbindung
      DATABASE_URL: postgresql://${DB_USER:-vtadmin}:${DB_PASSWORD:-changeme_required}@db:5432/${DB_NAME:-vt_inventory}?sslmode=disable

      # App-Konfiguration
      APP_ENV: ${APP_ENV:-production}
      APP_PORT: 3000
      APP_HOST: 0.0.0.0

      # JWT/Security
      JWT_SECRET: ${JWT_SECRET:-changeme_required}
      SESSION_SECRET: ${SESSION_SECRET:-changeme_required}

      # Admin User (nur bei First-Start)
      ADMIN_EMAIL: ${ADMIN_EMAIL:-admin@example.com}
      ADMIN_PASSWORD: ${ADMIN_PASSWORD:-changeme_required}

      # Optional: Backup-Konfiguration
      BACKUP_S3_ENABLED: ${BACKUP_S3_ENABLED:-false}
      BACKUP_S3_BUCKET: ${BACKUP_S3_BUCKET:-}
      BACKUP_S3_REGION: ${BACKUP_S3_REGION:-eu-central-1}
      BACKUP_S3_ACCESS_KEY: ${BACKUP_S3_ACCESS_KEY:-}
      BACKUP_S3_SECRET_KEY: ${BACKUP_S3_SECRET_KEY:-}

      # CORS & Security
      ALLOWED_ORIGINS: ${ALLOWED_ORIGINS:-https://your-domain.com}
      ENABLE_CORS: "true"

      # Logging
      LOG_LEVEL: ${LOG_LEVEL:-info}
      LOG_FORMAT: json

    depends_on:
      db:
        condition: service_healthy

    volumes:
      # Backups (shared mit DB)
      - ./backups:/app/backups:rw
      # Logs
      - ./logs:/app/logs:rw

    ports:
      # Nur intern zugänglich (über Caddy)
      - "127.0.0.1:3000:3000"

    networks:
      - vt_network

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s

    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G

  # Caddy Reverse Proxy + Auto SSL
  caddy:
    image: caddy:2-alpine
    container_name: vt-caddy
    restart: always

    ports:
      - "80:80"
      - "443:443"
      - "443:443/udp"

    environment:
      ACME_AGREE: "true"

    volumes:
      # Caddyfile (Reverse-Proxy Konfiguration)
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      # SSL Zertifikate (persistent)
      - caddy_data:/data
      # Let's Encrypt Cache
      - caddy_config:/config

    networks:
      - vt_network

    depends_on:
      - app

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:2019/status"]
      interval: 30s
      timeout: 5s
      retries: 3

    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M

volumes:
  db_data:
    driver: local
  caddy_data:
    driver: local
  caddy_config:
    driver: local

networks:
  vt_network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.28.0.0/16
```

### 1.2 .env.example

```bash
# ============================================================================
# ESSENTIAL CONFIGURATION - REQUIRED
# ============================================================================

# Domain für SSL/HTTPS (z.B. vt.your-company.de)
DOMAIN=your-domain.com

# Database Credentials (CHANGE THESE!)
DB_USER=vtadmin
DB_PASSWORD=very_secure_password_change_me
DB_NAME=vt_inventory

# Admin User für First-Start
ADMIN_EMAIL=admin@your-company.de
ADMIN_PASSWORD=admin_initial_password_change_me

# JWT & Session Security (CHANGE THESE!)
JWT_SECRET=your_jwt_secret_key_min_32_chars_change_me
SESSION_SECRET=your_session_secret_key_min_32_chars_change_me

# ============================================================================
# OPTIONAL CONFIGURATION - HAS DEFAULTS
# ============================================================================

# App Environment (development, staging, production)
APP_ENV=production

# App Version (wird auto-gesetzt von CI/CD)
APP_VERSION=latest

# Allowed CORS Origins (Comma-separated)
ALLOWED_ORIGINS=https://your-domain.com

# Logging
LOG_LEVEL=info

# ============================================================================
# OPTIONAL: S3 BACKUPS
# ============================================================================

# Enable S3 Backups
BACKUP_S3_ENABLED=false

# AWS S3 Credentials
BACKUP_S3_BUCKET=my-backup-bucket
BACKUP_S3_REGION=eu-central-1
BACKUP_S3_ACCESS_KEY=
BACKUP_S3_SECRET_KEY=

# ============================================================================
# CI/CD BUILD VARIABLES (wird auto-gesetzt)
# ============================================================================
BUILD_COMMIT=unknown
```

---

## 2. Container-Architektur

### 2.1 Design: Single Container vs. Multi-Container

**Entscheidung: Single Container für App, Separate Container für DB & Proxy**

```
┌─────────────────────────────────────────────┐
│          Docker Host (CX21)                 │
├─────────────────────────────────────────────┤
│                                             │
│  ┌───────────────────────────────────────┐  │
│  │ Caddy (Reverse Proxy + SSL)           │  │
│  │ Port 80 (HTTP → HTTPS redirect)       │  │
│  │ Port 443 (HTTPS)                      │  │
│  └───────────────────────────────────────┘  │
│           ↓ (proxy)                         │
│  ┌───────────────────────────────────────┐  │
│  │ App Container (vt-app)                │  │
│  │ Go Backend + React Frontend (SPA)     │  │
│  │ Port 3000 (lokal, nicht exposed)      │  │
│  │ - /api/... → Go Router                │  │
│  │ - /... → React Static Files           │  │
│  └───────────────────────────────────────┘  │
│           ↓ (TCP 5432)                      │
│  ┌───────────────────────────────────────┐  │
│  │ PostgreSQL (vt-db)                    │  │
│  │ Port 5432 (lokal, nicht exposed)      │  │
│  └───────────────────────────────────────┘  │
│           ↓ (Volume)                        │
│  ┌───────────────────────────────────────┐  │
│  │ Shared Backups Volume (/backups)      │  │
│  │ App + DB können hier schreiben        │  │
│  └───────────────────────────────────────┘  │
│                                             │
└─────────────────────────────────────────────┘
```

**Warum Single Container für App?**
- ✅ Einfacher zu deployen (nur ein Binary + React dist)
- ✅ Kleinere Ressourcennutzung
- ✅ Kein Inter-Container IPC erforderlich
- ✅ Go Server kann Static Files direkt serve

**Container Responsibilities:**

| Container | Aufgaben | Restart Policy |
|-----------|----------|-----------------|
| `vt-app` | Go Backend + React SPA | always |
| `vt-db` | PostgreSQL 16 | always |
| `vt-caddy` | SSL/TLS + HTTP Redirect | always |

---

## 3. Build-Pipeline

### 3.1 Multi-Stage Dockerfile (Dockerfile.prod)

```dockerfile
# ============================================================================
# STAGE 1: Build Go Backend
# ============================================================================
FROM golang:1.22-alpine AS go_builder

WORKDIR /app

# Dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy Go modules
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build Binary
# Flags:
# -w -s: Strip symbols (kleineres Binary)
# -X: Set build variables (für Versioning)
ARG BUILD_VERSION=dev
ARG BUILD_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.Commit=${BUILD_COMMIT}" \
    -o /tmp/app ./cmd/server/main.go

# ============================================================================
# STAGE 2: Build React Frontend
# ============================================================================
FROM node:22-alpine AS react_builder

WORKDIR /frontend

# Copy package files
COPY frontend/package.json frontend/package-lock.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source
COPY frontend/ ./

# Build production bundle
RUN npm run build

# ============================================================================
# STAGE 3: Runtime Container
# ============================================================================
FROM alpine:3.19

# Install Runtime Dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    postgresql-client \
    curl \
    bash

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy Go binary from builder
COPY --from=go_builder /tmp/app /app/app

# Copy React build artifacts
COPY --from=react_builder /frontend/dist /app/public

# Create necessary directories
RUN mkdir -p /app/backups /app/logs && \
    chown -R appuser:appuser /app

# Set user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=30s \
    CMD curl -f http://localhost:3000/api/health || exit 1

# Expose port
EXPOSE 3000

# Run app
CMD ["/app/app"]
```

### 3.2 Build-Optimierungen

**Image-Größe-Reduzierung:**
```dockerfile
# ✅ Multi-Stage Build: ~50MB statt 500MB+
# ✅ Alpine base: ~5MB statt 100MB+ (Debian)
# ✅ Go Binary stripping: -ldflags="-w -s"
# ✅ Nur Prod Dependencies: npm ci --only=production
```

**Sicherheit im Build:**
```dockerfile
# ✅ Non-root User (appuser:1000)
# ✅ Read-only Filesystems wo möglich
# ✅ No secrets in Image Layers
# ✅ Distroless/Alpine nur essentials
```

---

## 4. Reverse-Proxy: Caddy-Konfiguration

### 4.1 Caddyfile (Automatisches SSL + HTTP Redirect)

```
# ============================================================================
# Caddyfile - Reverse Proxy & SSL Configuration
# ============================================================================

# Global Options
{
    email admin@{$DOMAIN}
    auto_https on
    log {
        output stdout
        format json
        level info
    }
    # Metrics für Monitoring
    admin localhost:2019
}

# HTTP → HTTPS Redirect
http://{$DOMAIN} {
    redir https://{$DOMAIN}{uri} permanent
}

# HTTPS Main Endpoint
https://{$DOMAIN} {
    # SSL Zertifikat (automatisch via Let's Encrypt)
    encode gzip

    # Health Check Endpoint (nicht proxied)
    @health path /api/health
    handle @health {
        uri strip_prefix /api
        reverse_proxy localhost:3000 {
            health_uri /health
            health_timeout 5s
            health_interval 10s
        }
    }

    # API Endpoints
    @api path /api/*
    handle @api {
        # Rate Limiting: 100 requests/minute pro IP
        rate_limit {
            zone api 100r/m
        }

        # Compression
        encode gzip

        # API Reverse Proxy
        reverse_proxy localhost:3000 {
            # Timeouts
            timeout 30s

            # Headers
            header_up X-Real-IP {remote_host}
            header_up X-Forwarded-For {remote_host}
            header_up X-Forwarded-Proto https
            header_up X-Forwarded-Host {host}

            # WebSocket Support (für zukünftige Real-Time)
            websocket
        }
    }

    # Static Files + SPA Fallback
    @static {
        path /assets/* /img/* /fonts/* /favicon.ico /robots.txt
    }
    handle @static {
        file_server {
            root /app/public
        }
        header Cache-Control max-age=31536000
    }

    # React SPA: Alle anderen Routes → index.html
    @spa {
        not path /api/*
        not path /assets/*
    }
    handle @spa {
        reverse_proxy localhost:3000 {
            header_up X-Forwarded-Proto https
        }
    }

    # Security Headers
    header {
        # Content Security Policy
        Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:"

        # Prevent Clickjacking
        X-Frame-Options "SAMEORIGIN"

        # Prevent MIME Type Sniffing
        X-Content-Type-Options "nosniff"

        # Enable XSS Protection
        X-XSS-Protection "1; mode=block"

        # Referrer Policy
        Referrer-Policy "strict-origin-when-cross-origin"
    }

    # Access Logs
    log {
        output stdout
        format json
        level info
    }
}

# Admin API (lokal only)
localhost:2019 {
    respond /status "OK" 200
}
```

### 4.2 SSL Zertifikate

**Automatische Verwaltung via Caddy + Let's Encrypt:**
- Automatische Erneuerung 30 Tage vor Ablauf
- Zertifikate in `caddy_data` Volume persistent
- HTTP-01 Challenge für DNS-freie Verifizierung
- Fehlerbehandlung mit automatischem Retry

---

## 5. Datenbank-Konfiguration

### 5.1 PostgreSQL Setup

**postgres.conf (Tuning für CX21):**
```ini
# Connection Management
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# Query Planner
random_page_cost = 1.1
effective_io_concurrency = 200

# Logging
log_statement = 'all'
log_duration = on
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

### 5.2 Connection Pooling (PgBouncer optional)

**Für Apps mit vielen Connections:**
```yaml
  pgbouncer:
    image: pgbouncer:latest
    # ...pooling config...
```

**Go Backend Connection Pool (Standard):**
```go
db, _ := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)        // Max 25 connections
db.SetMaxIdleConns(5)         // Idle pool size
db.SetConnMaxLifetime(5 * time.Minute)
```

### 5.3 Migrations-Strategie

**Zwei Optionen:**

**Option A: golang-migrate (Empfohlen)**
```go
// migrations/ Verzeichnis
// 000001_initial_schema.up.sql
// 000001_initial_schema.down.sql
// 000002_add_users_table.up.sql
// ...

import "github.com/golang-migrate/migrate/v4"

func runMigrations(dbURL string) error {
    m, _ := migrate.New("file://migrations", dbURL)
    return m.Up()
}
```

**Aufruf im App-Start:**
```go
// main.go
if err := runMigrations(os.Getenv("DATABASE_URL")); err != nil {
    log.Fatal("Migration failed:", err)
}
```

**Option B: SQL Scripts im Container**
```bash
# docker-entrypoint.sh
#!/bin/sh
psql "$DATABASE_URL" < /app/migrations/init.sql
exec /app/app
```

---

## 6. Backup-System

### 6.1 Automatische Backups (täglich)

**backup.sh (läuft im App Container):**
```bash
#!/bin/bash

# Configuration
BACKUP_DIR="/app/backups"
DB_USER="${DB_USER}"
DB_NAME="${DB_NAME}"
DB_HOST="db"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Ensure backup directory
mkdir -p "${BACKUP_DIR}"

# Create backup
BACKUP_FILE="${BACKUP_DIR}/backup_${TIMESTAMP}.sql.gz"

pg_dump \
    -h "${DB_HOST}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --verbose \
    --format=plain \
    --compress=9 \
    --if-exists \
    --create \
    > "${BACKUP_FILE}" 2>&1

if [ $? -eq 0 ]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') - Backup created: ${BACKUP_FILE} ($(du -h ${BACKUP_FILE} | cut -f1))" >> "${BACKUP_DIR}/backup.log"

    # S3 Upload (optional)
    if [ "${BACKUP_S3_ENABLED}" = "true" ]; then
        aws s3 cp "${BACKUP_FILE}" "s3://${BACKUP_S3_BUCKET}/backups/"
    fi

    # Cleanup old backups (retention policy)
    find "${BACKUP_DIR}" -name "backup_*.sql.gz" -mtime "+${RETENTION_DAYS}" -delete
else
    echo "$(date '+%Y-%m-%d %H:%M:%S') - BACKUP FAILED!" >> "${BACKUP_DIR}/backup.log"
    exit 1
fi
```

### 6.2 Backup Scheduler (Go Cronjob im App)

```go
// internal/backup/scheduler.go
package backup

import (
    "time"
    "github.com/robfig/cron/v3"
)

func StartBackupScheduler() {
    c := cron.New()

    // Daily at 2 AM
    c.AddFunc("0 2 * * *", func() {
        if err := CreateBackup(); err != nil {
            log.Errorf("Backup failed: %v", err)
            alertAdmin("Backup failed", err.Error())
        }
    })

    c.Start()
}

// Retention Cleanup
func CleanupOldBackups(retentionDays int) {
    cutoff := time.Now().AddDate(0, 0, -retentionDays)
    files, _ := ioutil.ReadDir("/app/backups")
    for _, f := range files {
        if f.ModTime().Before(cutoff) && filepath.Ext(f.Name()) == ".gz" {
            os.Remove(f.Name())
        }
    }
}
```

### 6.3 Restore-Prozedur

```bash
#!/bin/bash
# restore.sh
BACKUP_FILE="$1"
DB_USER="${DB_USER}"
DB_NAME="${DB_NAME}"
DB_HOST="db"

# Restore
gunzip < "${BACKUP_FILE}" | psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}"

if [ $? -eq 0 ]; then
    echo "Restore erfolgreich"
else
    echo "Restore fehlgeschlagen"
    exit 1
fi
```

### 6.4 Backup Retention Policy

| Szenario | Retention | Konfiguration |
|----------|-----------|-----------------|
| Default (kleine Firmen) | 7 Tage | `BACKUP_RETENTION_DAYS=7` |
| Medium | 4 Wochen | `BACKUP_RETENTION_DAYS=28` |
| Large | 3 Monate | `BACKUP_RETENTION_DAYS=90` |
| S3 Archive | Unbegrenzt | Lifecycle Policies |

---

## 7. Auto-Update-Mechanismus

### 7.1 Update Flow

```
┌─────────────────────────────────────────────────────┐
│ 1. Admin drückt "Nach Updates prüfen"               │
│    GET /api/admin/updates/check                     │
└─────────────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────┐
│ 2. App prüft GitHub Releases API                    │
│    HEAD https://api.github.com/repos/...            │
│    Vergleicht: CURRENT_VERSION vs LATEST_VERSION    │
└─────────────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────┐
│ 3. Changelog anzeigen (optional)                    │
│    GET /api/admin/updates/changelog                 │
│    Extrakt aus Release Notes                        │
└─────────────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────┐
│ 4. Admin bestätigt: "Jetzt updaten"                 │
│    POST /api/admin/updates/apply                    │
└─────────────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────┐
│ 5. Update-Sequence:                                 │
│    a) DB Backup erstellen                           │
│    b) Docker Image pull (new version)               │
│    c) Container stop                                │
│    d) Container start (new version)                 │
│    e) DB Migrations laufen                          │
│    f) Health-Check                                  │
└─────────────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────────────┐
│ 6a. Erfolg → Update abgeschlossen                   │
│    Status: OK                                       │
└─────────────────────────────────────────────────────┘

            ODER

┌─────────────────────────────────────────────────────┐
│ 6b. Fehler → Automatisches Rollback                 │
│    - Container stop                                 │
│    - Alte Version starten                           │
│    - DB Restore aus Backup                          │
│    - Admin benachrichtigen                          │
└─────────────────────────────────────────────────────┘
```

### 7.2 Go Backend Implementation

```go
// internal/updates/manager.go
package updates

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "time"
)

type Release struct {
    Version string `json:"tag_name"`
    URL     string `json:"html_url"`
    Notes   string `json:"body"`
}

type UpdateManager struct {
    RepoOwner string
    RepoName  string
    ImageName string
}

// Check für neue Versionen
func (um *UpdateManager) CheckForUpdates() (*Release, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
        um.RepoOwner, um.RepoName)

    resp, _ := http.Get(url)
    var release Release
    json.NewDecoder(resp.Body).Decode(&release)

    currentVersion := os.Getenv("APP_VERSION")
    if release.Version != currentVersion {
        return &release, nil
    }
    return nil, nil
}

// Update durchführen
func (um *UpdateManager) ApplyUpdate(newVersion string) error {
    // 1. Backup erstellen
    log.Info("Creating database backup...")
    if err := CreateBackup(); err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }

    // 2. Docker Image pull
    log.Info("Pulling new Docker image...")
    cmd := exec.Command("docker", "compose", "pull")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("docker pull failed: %w", err)
    }

    // 3. Services neustarten
    log.Info("Restarting services...")
    cmd = exec.Command("docker", "compose", "up", "-d")
    if err := cmd.Run(); err != nil {
        return um.RollbackUpdate()
    }

    // 4. Health Check
    log.Info("Running health checks...")
    for i := 0; i < 30; i++ {
        if um.healthCheck() {
            log.Info("Update successful!")
            return nil
        }
        time.Sleep(1 * time.Second)
    }

    // 5. Rollback bei Fehler
    return um.RollbackUpdate()
}

// Rollback bei Fehler
func (um *UpdateManager) RollbackUpdate() error {
    log.Error("Health check failed, rolling back...")

    // Alte Version starten
    cmd := exec.Command("docker", "compose", "down")
    cmd.Run()

    // Aus Backup restore
    if err := RestoreFromLatestBackup(); err != nil {
        log.Errorf("Rollback failed: %v", err)
        // Alert Admin!
        alertAdmin("CRITICAL: Auto-update rollback failed", err.Error())
        return err
    }

    // Alte Container starten
    cmd = exec.Command("docker", "compose", "up", "-d")
    return cmd.Run()
}

func (um *UpdateManager) healthCheck() bool {
    resp, err := http.Get("http://localhost:3000/api/health")
    return err == nil && resp.StatusCode == 200
}
```

### 7.3 API Endpoints

```go
// routes.go
router.GET("/api/admin/updates/check", handlers.CheckUpdates)
router.GET("/api/admin/updates/changelog", handlers.GetChangelog)
router.POST("/api/admin/updates/apply", authMiddleware, handlers.ApplyUpdate)
router.GET("/api/admin/updates/status", handlers.GetUpdateStatus)
```

---

## 8. Monitoring & Observability

### 8.1 Health Check Endpoint

```go
// handlers/health.go
type HealthResponse struct {
    Status      string    `json:"status"`
    Timestamp   time.Time `json:"timestamp"`
    Version     string    `json:"version"`
    Uptime      int64     `json:"uptime_seconds"`
    Database    string    `json:"database"`
    Memory      MemInfo   `json:"memory"`
    Disk        DiskInfo  `json:"disk"`
    ActiveUsers int       `json:"active_users"`
}

func HealthCheck(c *gin.Context) {
    var dbStatus string
    err := db.QueryRow("SELECT 1").Scan()
    if err != nil {
        dbStatus = "unhealthy"
    } else {
        dbStatus = "healthy"
    }

    c.JSON(200, HealthResponse{
        Status:    "ok",
        Timestamp: time.Now(),
        Version:   os.Getenv("APP_VERSION"),
        Uptime:    time.Since(startTime).Milliseconds() / 1000,
        Database:  dbStatus,
        Memory:    getMemoryInfo(),
        Disk:      getDiskInfo(),
    })
}
```

### 8.2 Metrics Collection (Prometheus Format)

```go
// metrics.go
import "github.com/prometheus/client_golang/prometheus"

var (
    // Counter: Request Total
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    // Gauge: Current DB Connections
    dbConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_current",
            Help: "Current database connections",
        },
    )

    // Gauge: Memory Usage
    memoryUsage = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "memory_usage_bytes",
            Help: "Memory usage in bytes",
        },
    )

    // Gauge: Disk Usage
    diskUsage = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "disk_usage_percent",
            Help: "Disk usage percentage",
        },
    )
)

// Collector (läuft kontinuierlich)
func CollectMetrics() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        // Memory
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        memoryUsage.Set(float64(m.Alloc))

        // DB Connections
        dbConnections.Set(float64(db.Stats().OpenConnections))

        // Disk
        diskUsage.Set(float64(getDiskUsagePercent()))
    }
}
```

### 8.3 Monitoring Admin-Panel

**Dashboard zeigt:**

```typescript
// Frontend: /admin/monitoring
interface MonitoringData {
  cpu: {
    percent: number;
    cores: number;
  };
  memory: {
    used_mb: number;
    total_mb: number;
    percent: number;
  };
  disk: {
    used_gb: number;
    available_gb: number;
    percent: number;
  };
  database: {
    size_mb: number;
    connections: number;
    last_backup: Date;
    backup_size_mb: number;
  };
  app: {
    uptime_hours: number;
    version: string;
    http_requests_per_minute: number;
    errors_last_hour: number;
  };
  users: {
    total: number;
    active_today: number;
    active_now: number;
  };
}
```

**Alerts (optional):**
- Disk > 80%: ⚠️ Warning
- Memory > 85%: ⚠️ Warning
- Database > 90% of size limit: 🔴 Critical
- Health Check failed (3x): 🔴 Critical (Auto-Restart)

---

## 9. Logging

### 9.1 Strukturiertes JSON Logging

```go
// internal/logger/logger.go
import "go.uber.org/zap"

var log *zap.SugaredLogger

func init() {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{"stdout", "/app/logs/app.log"}

    base, _ := config.Build()
    log = base.Sugar()
}

// Beispiele:
log.Infow("User login", "user_id", 123, "ip", "192.168.1.1")
log.Errorw("Payment failed", "invoice_id", 456, "error", err)
log.Debugw("Query executed", "duration_ms", 15, "rows", 42)
```

**Log Output (JSON):**
```json
{
  "level": "info",
  "ts": 1695034800.123,
  "logger": "app",
  "msg": "User login",
  "user_id": 123,
  "ip": "192.168.1.1"
}
```

### 9.2 Log Rotation

**docker-compose.yml:**
```yaml
app:
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
```

**Oder via Logrotate (Host):**
```
/app/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    notifempty
    create 0640 appuser appuser
    sharedscripts
}
```

### 9.3 Log Viewer im Admin-Panel

```go
// handlers/admin_logs.go
func GetLogs(c *gin.Context) {
    lines := c.DefaultQuery("lines", "100")
    filter := c.DefaultQuery("filter", "")

    // Read last N lines from /app/logs/app.log
    logs := readLogFile("/app/logs/app.log", lines, filter)

    c.JSON(200, gin.H{"logs": logs})
}
```

---

## 10. Sicherheit

### 10.1 Non-Root Container Execution

```dockerfile
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

USER appuser

# Read-Only Filesystems wo möglich
# Volume mit rw nur für /app/logs und /app/backups
```

### 10.2 Network Isolation

```yaml
networks:
  vt_network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.28.0.0/16

services:
  app:
    networks:
      - vt_network
    # Nur interne Kommunikation, keine externen Ports außer via Caddy
```

### 10.3 Secrets Management

**Option 1: Docker Secrets (Swarm)**
```bash
echo "password123" | docker secret create db_password -
```

**Option 2: .env.local (Empfohlen für Single Server)**
```bash
# .env.local (NICHT ins Git!)
DB_PASSWORD=very_secure_password
JWT_SECRET=...
```

**Best Practice:**
1. Generiere Secrets beim First-Start
2. Schreibe in `.env.local`
3. Lockdown Permissions: `chmod 600 .env.local`
4. Backup Secrets verschlüsselt

### 10.4 Rate Limiting

**In Caddy (bereits konfiguriert):**
```
rate_limit {
    zone api 100r/m      # 100 requests/minute
    zone auth 5r/m       # 5 attempts/minute für Login
}
```

**Im Go Backend (für API):**
```go
import "github.com/ulule/limiter/v3"

// 10 requests per second per IP
limiter := limiter.New(limiter.Store, rate.Limit{Period: 1*time.Second, Limit: 10})
```

### 10.5 HTTPS/TLS Erzwingen

```yaml
# Caddyfile
@insecure {
    protocol http
}
handle @insecure {
    redir https://{host}{uri} permanent
}
```

### 10.6 SQL Injection Prevention

```go
// ✅ Prepared Statements (Go Standard)
err := db.QueryRow("SELECT * FROM users WHERE id = $1", userID).Scan(&user)

// ❌ Niemals String Concatenation!
// query := "SELECT * FROM users WHERE id = " + userID  // DANGER!
```

### 10.7 CORS Configuration

```go
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{os.Getenv("ALLOWED_ORIGINS")},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Authorization", "Content-Type"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

---

## 11. Development Setup

### 11.1 docker-compose.dev.yml (Hot Reload)

```yaml
version: '3.9'

services:
  # PostgreSQL (normal)
  db:
    image: postgres:16-alpine
    container_name: vt-db-dev
    environment:
      POSTGRES_USER: vtadmin
      POSTGRES_PASSWORD: dev_password
      POSTGRES_DB: vt_inventory_dev
    volumes:
      - db_data_dev:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    networks:
      - vt_network_dev

  # Go Backend (mit Hot Reload)
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    container_name: vt-app-dev
    restart: on-failure
    environment:
      DATABASE_URL: postgresql://vtadmin:dev_password@db:5432/vt_inventory_dev?sslmode=disable
      APP_ENV: development
      LOG_LEVEL: debug
    volumes:
      # Source Code (für hot reload via Air)
      - ./cmd:/app/cmd
      - ./internal:/app/internal
      - ./migrations:/app/migrations
      # Frontend (für Vite hot reload)
      - ./frontend/src:/app/frontend/src
      - ./frontend/public:/app/frontend/public
    ports:
      - "3000:3000"
      - "5173:5173"  # Vite dev server
    depends_on:
      - db
    networks:
      - vt_network_dev
    command: sh -c "air -c .air.toml"

  # Optional: pgAdmin für DB Management
  pgadmin:
    image: dpage/pgadmin4:latest
    container_name: vt-pgadmin-dev
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@dev.local
      PGADMIN_DEFAULT_PASSWORD: admin
    ports:
      - "5050:80"
    depends_on:
      - db
    networks:
      - vt_network_dev

volumes:
  db_data_dev:

networks:
  vt_network_dev:
    driver: bridge
```

### 11.2 Dockerfile.dev (mit Air für Go Hot Reload)

```dockerfile
FROM golang:1.22-alpine

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Air für Hot Reload
RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 3000

CMD ["air", "-c", ".air.toml"]
```

### 11.3 .air.toml (Go Hot Reload Config)

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/server/main.go"
  bin = "./tmp/main"
  full_bin = ""
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor", "frontend/dist"]
  include_dir = []
  exclude_file = []
  delay = 1000
  stop_on_error = true
  log = "build-errors.log"
  poll = false
  poll_interval = 0

[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  yellow = "yellow"
```

### 11.4 Frontend Hot Reload (Vite)

**vite.config.ts:**
```typescript
export default defineConfig({
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      }
    }
  }
})
```

**Start:**
```bash
docker-compose -f docker-compose.dev.yml up

# Backend: http://localhost:3000/api
# Frontend (Vite): http://localhost:5173
```

---

## 12. CI/CD Pipeline (GitHub Actions)

### 12.1 GitHub Actions Workflow

**.github/workflows/deploy.yml:**
```yaml
name: Build & Deploy

on:
  push:
    branches: [main, develop]
    tags: ['v*']
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: vt_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Run Tests
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/vt_test?sslmode=disable
        run: |
          go test ./... -v -race -coverprofile=coverage.out

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Log in to Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract Metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha

      - name: Build and Push Image
        uses: docker/build-push-action@v4
        with:
          context: .
          file: ./Dockerfile.prod
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          build-args: |
            BUILD_VERSION=${{ github.ref_name }}
            BUILD_COMMIT=${{ github.sha }}

      - name: Create Release
        if: startsWith(github.ref, 'refs/tags/v')
        uses: softprops/action-gh-release@v1
        with:
          generate_release_notes: true
```

---

## 13. Erste Installation (5 Schritte)

### Schritt 1: Server vorbereiten

```bash
# Hetzner CX21 (Ubuntu 22.04 LTS) oder beliebig

# Docker & Docker Compose installieren
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Firewall
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

### Schritt 2: Repository klonen & Verzeichnis vorbereiten

```bash
mkdir /opt/vt-inventory
cd /opt/vt-inventory

# Docker Compose Files kopieren
wget https://raw.githubusercontent.com/your-repo/main/docker-compose.yml
wget https://raw.githubusercontent.com/your-repo/main/.env.example

# Konfigurieren
cp .env.example .env
nano .env

# Wichtigste Änderungen:
# DOMAIN=vt.your-company.de
# DB_PASSWORD=<generiert>
# JWT_SECRET=<generiert>
```

### Schritt 3: Verzeichnisse & Permissions setzen

```bash
mkdir -p backups logs
chmod 755 backups logs
touch backups/.gitkeep

# Docker-Benutzer für Logs
sudo chown -R 1000:1000 logs
```

### Schritt 4: Container starten

```bash
docker-compose pull
docker-compose up -d

# Logs überprüfen
docker-compose logs -f app

# Health Check
curl http://localhost:3000/api/health
```

### Schritt 5: Admin-Account & SSL Testen

```bash
# HTTPS Test (wartet auf Caddy cert)
sleep 30
curl https://vt.your-company.de/

# Admin-Account setzen (erste Anmeldung im UI)
# Browser: https://vt.your-company.de
# E-Mail: admin@your-company.de (aus .env)
# Passwort: (aus .env)
```

**Fertig! ✅**

---

## 14. Server-Sizing Empfehlungen

### Für verschiedene Firmengrößen:

| Firmensize | Nutzer | Hetzner | RAM | SSD | vCPU | Kosten | Notes |
|-----------|--------|---------|-----|-----|------|--------|-------|
| **Klein (MVP)** | 3-5 | CX11 | 2GB | 25GB | 1 | 4€/Mo | ⚠️ Nur Test/Dev |
| **Small (Standard)** | 5-10 | CX21 | 4GB | 40GB | 2 | 5€/Mo | ✅ Empfohlen (diese Doku) |
| **Medium** | 10-20 | CX31 | 8GB | 80GB | 4 | 11€/Mo | ✅ Gutes Headroom |
| **Large** | 20-50 | CX41 | 16GB | 160GB | 8 | 24€/Mo | ✅ Sehr komfortabel |

**Capacity Planning (CX21):**

```
Speichernutzung:
- OS + Docker: 2 GB
- App Code: 50 MB
- PostgreSQL (100k Artikel): ~200 MB
- Logs (7 Tage): ~500 MB
- Backups (7x): ~1.4 GB
━━━━━━━━━━━━━━━━━━━━
  Verfügbar: ~40 GB ✅

RAM Allocation:
- PostgreSQL: 1.5 GB
- Go App: 1.5 GB
- OS + Buffer: 0.5 GB
- Caddy: 256 MB
━━━━━━━━━━━━━━━━━━━━━
  Total: ~3.75 GB ✅ (4 GB verfügbar)

CPU Load:
- Baseline: 5-10%
- Spike (Rechnung generieren): ~40-60%
- Max Concurrent: 5-10 Benutzer ✅
```

**Vertical Scaling Trigger:**

```
CX21 → CX31 wenn:
  - RAM > 85% sustained
  - CPU > 70% sustained
  - 15+ concurrent users
  - Disk > 80%

CX31 → CX41 wenn:
  - 30+ concurrent users
  - Mehrere Lagerorte / Filialen
  - Externe API Integrationen
```

---

## 15. Troubleshooting

### Problem: Container starten nicht

```bash
# Logs überprüfen
docker-compose logs app
docker-compose logs db
docker-compose logs caddy

# Container Status
docker-compose ps

# Neu starten
docker-compose down
docker-compose up -d
```

### Problem: SSL Zertifikat ungültig

```bash
# Caddy Logs
docker-compose logs caddy

# Manuelle Cleanup (bei Problemen)
docker-compose down
rm -rf caddy_data caddy_config
docker-compose up -d caddy

# Let's Encrypt Rate Limit? → Warten Sie 1 Stunde oder nutzen Staging API
```

### Problem: Datenbank voll

```bash
# DB Größe prüfen
docker-compose exec db psql -U vtadmin -d vt_inventory -c "SELECT pg_size_pretty(pg_database_size('vt_inventory'));"

# Alte Daten löschen (optional, manuell)
# Oder: Upgrade auf größere SSD
```

### Problem: Backup schlägt fehl

```bash
# Test Backup
docker-compose exec app sh -c '/app/backup.sh'

# Backup Status
ls -lh backups/

# Restore testen
docker-compose exec db psql -U vtadmin -d vt_inventory < backups/backup_latest.sql
```

---

## 16. Checkliste Deployment

- [ ] Server: Ubuntu 22.04 LTS, 4GB RAM, 40GB SSD
- [ ] Docker + Docker Compose installiert
- [ ] Domain registriert & DNS konfiguriert (A-Record zu Server IP)
- [ ] `.env` Datei erstellt mit Secrets
- [ ] `docker-compose.yml` auf Server
- [ ] `docker-compose up -d` erfolgreich
- [ ] Health Check: `curl https://domain.com/api/health` → 200 OK
- [ ] HTTPS funktioniert (kein Browser Warning)
- [ ] Admin-Account erstellt
- [ ] Erstes Backup erfolgreich
- [ ] Monitoring Dashboard zugänglich
- [ ] Auto-Update-Check funktioniert
- [ ] Logs werden geschrieben
- [ ] Rate Limiting getestet

---

## 17. Zusätzliche Resources

**Dokumentation:**
- [Docker Compose Docs](https://docs.docker.com/compose/)
- [Caddy Reverse Proxy](https://caddyserver.com/)
- [PostgreSQL Performance](https://www.postgresql.org/docs/16/runtime-config.html)
- [Let's Encrypt](https://letsencrypt.org/)

**Tools:**
- **Monitoring:** Prometheus + Grafana (optional, für größere Setups)
- **Log Aggregation:** ELK Stack (optional, für multi-server)
- **Database Backup:** pg_dump (included)
- **Container Registry:** GitHub Container Registry (free)

---

## 18. Zusammenfassung

Diese Architektur bietet:

✅ **Einfachheit:** `docker-compose up -d` → Läuft

✅ **Sicherheit:** HTTPS, Non-root User, Network Isolation, SQL Injection Prevention

✅ **Zuverlässigkeit:** Automatische Backups, Health Checks, Auto-Rollback Updates

✅ **Überwachung:** Real-time Monitoring, Logging, Error Tracking

✅ **Skalierbarkeit:** Einfaches Vertical Scaling (Hardware Upgrade)

✅ **Developer Experience:** Hot Reload, Lokal entwickeln wie Production

✅ **Cost-Efficient:** 5€/Monat auf Hetzner CX21 für 10+ User

---

**Version:** 1.0
**Zuletzt aktualisiert:** März 2026
**Maintainer:** DevOps Team
