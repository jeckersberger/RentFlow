# CrateDesk — Verifiziertes Code Review

**Datum:** 7. April 2026
**Reviewer:** Claude Code (Opus 4.6)
**Umfang:** ~160.000 Zeilen Code
- 396 Go-Dateien (102.336 Zeilen Production + 8.872 Zeilen Tests)
- 211 TypeScript/TSX-Dateien (~51.572 Zeilen Frontend)
- 19 Microservices + React 18 Frontend

**Methodik:** Verifikation gegen tatsächlichen Source-Code via parallele Analyse-Agents (Backend, Frontend, Security/Testing, Infrastruktur) + manueller Deep-Dive in kritische Bereiche.

---

## Gesamtbewertung: **Gut** ⭐⭐⭐⭐ (4/5)

Solides Projekt mit professioneller Architektur. Hauptrisiken: fehlende Test-Coverage, kein Monitoring, hardcoded Production-Credentials in Tests, und 10+ offene Bugs.

---

## 1. Backend-Architektur (Go) — 9/10

### Stärken

- **19/19 Services vollständig** gebaut mit konsistenter hexagonaler Architektur
- Jeder Service hat: `cmd/server/main.go`, `internal/domain/`, `internal/application/`, `internal/adapters/httphandler/`, `internal/infrastructure/postgres/`, `migrations/`
- **chi v5 Router** (nicht Gin) mit korrektem Middleware-Ordering
- **Context-Propagation: 100%** — alle DB/Service-Calls akzeptieren `context.Context`
- **Parametrisierte SQL-Queries überall** — kein SQL-Injection-Risiko gefunden
- **Transaktionen** mit `database.WithTx()` und Row-Level-Locking (`FOR UPDATE`)
- **Connection Pooling**: MaxConns=20, MinConns=2, 30min Lifetime
- **Go 1.24** konsistent in allen 19 Services
- **WebSocket Origin-Validation** sauber implementiert in `notification-service`

### Shared `pkg/common/` — 9 Middleware

| Modul | Status | Zweck |
|-------|--------|-------|
| `auth.go` | ✅ | JWT RS256 Validation |
| `authz.go` | ✅ | `RequireRole(roles...)` RBAC (69 Verwendungen in 19 Files) |
| `bodylimit.go` | ✅ | 1MB Body-Limit |
| `cors.go` | ⚠️ | CORS + Security Headers (siehe Issue #3) |
| `ratelimit.go` | ✅ | Redis-basiert per IP/Endpoint |
| `recovery.go` | ✅ | Panic Recovery |
| `requestid.go` | ✅ | Request-ID Propagation |
| `security_headers.go` | ✅ | HSTS, X-Frame-Options, CSP |
| `tenant.go` | ✅ | Tenant-Context aus JWT |

### Probleme

| # | Severity | Problem | Datei |
|---|----------|---------|-------|
| 1 | Medium | **CORS-Wildcard-Reflektion**: Bei `*` wird request-origin statt `*` als Header gesetzt | `pkg/common/middleware/cors.go:52-53` |
| 2 | Low | **Error ignoriert** bei Invoice-Status-Update nach Email-Versand | `services/invoice/.../invoice_handler.go:432` |
| 3 | Medium | **Invoice-Erstellung nicht transaktional**: Invoice + Items in separaten TXs | `services/invoice/.../invoice_service.go:171-215` |
| 4 | Low | `parseUUID()` in 10+ Services dupliziert statt in `pkg/common` | `*/httphandler/helpers.go` |

---

## 2. Frontend (React/TypeScript) — 7.5/10

### Stärken

- **151 Komponenten** (43 shared + 108 Pages)
- **Zustand** für State (6 Stores mit Persist-Middleware)
- **Axios** mit zentralen Interceptors: Token-Injection, Refresh-Token-Queue, Envelope-Unwrapping
- **React Router v6** mit Lazy Loading für 60+ Routes
- **TypeScript strict: ON** (aber `noImplicitAny: false`)
- **ErrorBoundary** vorhanden (class component, mit Stack-Traces)
- **i18n**: Deutsch/Englisch komplett (15 Namespaces), FR/NL teilweise
- **Vite** mit manuellen Chunks (vendor, query, ui)
- **Radix UI** + Lucide Icons + Framer Motion
- **Code-Splitting** via React.lazy für 60+ Routes

### Probleme

| # | Severity | Problem |
|---|----------|---------|
| 1 | Medium | **13 Seiten >800 Zeilen** — SetupWizard (1.093), ExpensesPage (1.001), EquipmentDetail (989) |
| 2 | Medium | **58× `eslint-disable @typescript-eslint/no-explicit-any`** — Type-Safety-Lücken |
| 3 | Medium | **api.ts hat 2.059 Zeilen** — alle 19 API-Module in einer Datei |
| 4 | Medium | **Keine Virtualisierung** für lange Listen (kein react-window) |
| 5 | Low | **Kein React.memo** verwendet |
| 6 | Low | **SCSS-Konvention gemischt**: 9× `.module.scss` + 10× plain `.scss` |
| 7 | Low | **Kein Prettier-Config** |
| 8 | Low | **3 TODOs** offen: NotificationCenter:70, CrewDetail:71, MyAssignments:27 |

---

## 3. Security — 7/10

### Stärken

- **JWT RS256** mit Key-Rotation via PEM-Dateien
- **Argon2id** für Passwort-Hashing (time=1, memory=64KB, threads=4, keyLen=32)
- **Rate Limiting** (Redis, per IP+Endpoint, Cloudflare CF-Connecting-IP trusted)
- **RequireRole** RBAC (69× in 19 Files)
- **Body-Size-Limit** 1MB auf allen Services
- **Security Headers**: HSTS, X-Frame-Options, CSP, Referrer-Policy
- **Alle Ports 127.0.0.1** gebunden (BSI CERT-Bund Fix)
- **Parametrisierte SQL** überall — keine Injection-Lücken
- **`.env` in .gitignore**, keine Secrets in Git-History

### 🔴 KRITISCHE Findings

#### K1: Hardcoded Production-Credentials in 12 Dateien
**Severity:** Hoch
**Dateien:**
- `frontend/e2e/login.spec.ts` (Zeilen 5-6, 20-21)
- `frontend/e2e/smoke.spec.ts` (Zeilen 23, 33, 45)
- `frontend/src/pages/Legal/DatenschutzPage.tsx`
- `frontend/src/pages/Legal/ImpressumPage.tsx`
- `docs/api/openapi.yaml`
- `scripts/test_auth.py`, `test_customer.py`, `test_inventory.py`, `test_invoice.py`, `test_project.py`, `test_scanner.py`, `test_warehouse.py`, `test_e2e.py`

**Problem:** Echte Production-Admin-Credentials (`j.eckersberger@je-soundulight.de` + Passwort) sind im Source-Code committed. Wer Zugriff auf das Repo hat, hat Production-Admin-Zugriff.

**Bonus-Problem:** `login.spec.ts:4` ruft `http://46.224.105.10/login` auf — **HTTP statt HTTPS**, umgeht TLS vollständig.

**Sofortmaßnahmen:**
1. Credentials in Tests durch `process.env.TEST_EMAIL / TEST_PASSWORD` ersetzen
2. Production-Passwort **rotieren** (ist bereits in Git-History exposed)
3. `DatenschutzPage.tsx` / `ImpressumPage.tsx`: Email-Adresse prüfen, ob wirklich öffentlich nötig
4. E2E-Tests auf HTTPS umstellen

---

#### K2: Setup-Wizard Trust-On-First-Use (TOFU) Vulnerability
**Datei:** `services/auth/internal/application/setup_service.go:87-101`
**Severity:** Mittel-Hoch

```go
if state == nil {
    tokenHash := hashSetupToken(req.Token)
    newState := &domain.SetupState{ TokenHash: tokenHash, ... }
    s.setupRepo.CreateState(ctx, newState)
    return true, nil  // accepts ANY token on first call
}
```

**Problem:** Der **erste Aufrufer** von `/setup/init` bestimmt den Token. Wenn ein frisches System vor manueller Setup-Initialisierung erreichbar ist, kann ein Angreifer eigene Admin-Credentials etablieren.

**Mitigation aktuell:** Keine. Kein Out-of-Band-Token, kein Time-Window-Limit, keine IP-Whitelist.

**Empfehlung:** Token muss beim ersten Server-Start generiert und **nur im Server-Log** ausgegeben werden. `/setup/init` akzeptiert nur den vorgenerierten Hash.

---

#### K3: CORS Wildcard-Reflektion
**Datei:** `pkg/common/middleware/cors.go:52-53`
**Severity:** Mittel

```go
if allowedOrigins["*"] || allowedOrigins[origin] {
    w.Header().Set("Access-Control-Allow-Origin", origin)  // reflects ANY origin
}
```

**Problem:** Mit `CORS_ORIGINS=*` wird die Request-Origin reflektiert — funktional gleich wie `*`, aber semantisch verwirrend. `DefaultCORSConfig()` hat `*` als Default. Line 60-62 setzt Credentials nur wenn NICHT wildcard, was den Schaden begrenzt — aber die Logik sollte trotzdem klar sein.

**Fix:**
```go
if allowedOrigins["*"] {
    w.Header().Set("Access-Control-Allow-Origin", "*")
} else if allowedOrigins[origin] {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```

---

### ⚠️ Weitere Security-Issues

| # | Severity | Problem | Quelle |
|---|----------|---------|--------|
| S1 | Medium | Docker-Container laufen als **root** — kein `USER` in Dockerfiles | `services/*/Dockerfile` |
| S2 | Medium | **Input-Validierung unvollständig** — 500er bei langen Namen / großen Zahlen / negativen Werten | BUG-027/028/029 |
| S3 | Low | Keine **Length-Limits** auf Text-Feldern | `user_handler.go:110-114` |
| S4 | Low | Kein **Secrets-Management** (Environment-Variablen statt Vault/Docker Secrets) | `docker-compose.yml` |

---

## 4. Testing — 4/10 (Schwächster Bereich)

### Zahlen

| Metrik | Wert |
|--------|------|
| Go-Test-Dateien | 25 |
| Go-Test-Zeilen | 8.872 |
| Go-Production-Zeilen | 102.336 |
| **Test/Production-Ratio** | **8,7%** |
| Handler-Tests | **0** |
| Repository-Tests | **0** |
| Integration-Tests | **0** |
| Frontend Unit-Tests | **0** |
| E2E-Tests (Playwright) | 2 Dateien (6 Tests) |

### Gefundene Tests

**Go:**
- 19× `domain/models_test.go` (Entity-Validierung)
- 3× Application-Tests: `auth_service_test.go`, `pricing_service_test.go`, `datev_service_test.go`
- 1× Claude-Provider-Test (ai-service)
- 1× Skill-Match-Test (crew-service)
- 1× Smart-Asset-Test (ai-service)

**E2E:** `frontend/e2e/smoke.spec.ts`, `frontend/e2e/login.spec.ts`
**Python Scripts:** 8× `scripts/test_*.py` (manuelle API-Tests, keine CI-Integration)

### Kritische Lücken

- **CI lässt Test-Failures durch** (`.github/workflows/ci.yml:41`: `|| true`)
- **Frontend-Job hat NULL Tests** (nur `npm run build`)
- **Kein Coverage-Reporting**
- **HTTP-Layer komplett ungetestet** (77 Handler, 0 Tests)

---

## 5. Infrastruktur & DevOps — 7.5/10

### Stärken

- **docker-compose.yml** mit 20 Containern
- **Multi-Stage Dockerfiles** (golang:1.24-alpine → alpine:3.20)
- **Traefik v3.1** mit Docker-Provider und PathPrefix-Routing
- **CI/CD**: `.github/workflows/ci.yml` (Push/PR auf develop/main)
- **Backup**: `pg_dumpall` + NAS-Replikation, 7 Tage Retention
- **go.work** für Workspace-Modus
- **57+ SQL-Migrationen** über alle Services
- **Health Checks**: `/health` (DB+Redis) + `/livez` (K8s-ready)
- **Alle Ports auf 127.0.0.1** (BSI-Fix)

### Probleme

| # | Severity | Problem |
|---|----------|---------|
| I1 | **High** | **Kein Monitoring** — `infra/prometheus/` leer, keine Dashboards, keine Alerts, keine `/metrics` Endpoints |
| I2 | Medium | **CI akzeptiert fehlschlagende Tests** (`\|\| true`) |
| I3 | Medium | **Keine Resource-Limits** in docker-compose (kein `deploy.resources.limits`) |
| I4 | Medium | **Docker-Container als root** (kein `USER` in Dockerfiles) |
| I5 | Low | **README veraltet** — Go 1.23 statt 1.24 |
| I6 | Low | **Deploy-Script baut sequentiell** (nötig wegen 4GB RAM auf CX33) |

---

## 6. Bekannte Bugs (TESTER_BUGS.md)

**31 Bugs dokumentiert, davon ~10 OPEN:**

| Bug | Severity | Status | Root Cause |
|-----|----------|--------|------------|
| BUG-009 | HIGH | OPEN | Projekt-Detail nicht klickbar — **Code-Check: onRowClick ist in `ProjectList.tsx:204` implementiert → evtl. bereits gefixt** |
| BUG-016 | MEDIUM | OPEN | Kundennamen fehlen überall — **Root Cause: `project_repo.go:20-27` speichert nur `customer_id`, keine JOIN zum customer-service (Microservice-Split)** |
| BUG-017 | MEDIUM | OPEN | Kalender zeigt Events nicht im Grid |
| BUG-018 | MEDIUM | OPEN | Scanner "Noch -1 Tage" |
| BUG-019 | MEDIUM | OPEN | Scanner Back-Stack zu flach |
| BUG-020 | MEDIUM | OPEN | Scanner Check-In zeigt alle Projekte |
| BUG-021 | MEDIUM | OPEN | Scanner rohe ISO-Timestamps |
| BUG-023 | MEDIUM | OPEN | Scanner Adressfelder Mismatch (`address_street` vs `billing_address_street`) |
| BUG-024 | MEDIUM | OPEN | Security Headers via Cloudflare nicht durchgereicht |
| BUG-025 | MEDIUM | OPEN | Benutzer-Rollen in UI leer |
| BUG-027 | LOW | OPEN | 500 bei 2000-Zeichen Name (Input-Validierung fehlt) |
| BUG-028 | LOW | OPEN | 500 bei großen Zahlen (int64 overflow) |
| BUG-029 | LOW | OPEN | Negative Budgets akzeptiert |
| BUG-030 | LOW | OPEN | Datumsformat nur YYYY-MM-DD (kein DD.MM.YYYY) |
| BUG-031 | LOW | OPEN | Kalender leere Klammern |

---

## 7. Priorisierte Empfehlungen

### 🔴 P0 — Sofort (Security)

1. **Production-Passwort rotieren** (in Git-History exposed)
2. **E2E-Credentials** via Env-Vars (`TEST_EMAIL`/`TEST_PASSWORD`)
3. **Setup-Wizard absichern** (out-of-band Token statt TOFU)
4. **CORS-Logik klarstellen** (`*` Branch)
5. **CI-Fix**: `|| true` entfernen, Tests müssen grün sein

### 🟡 P1 — Vor Go-Live

6. **Input-Validierung** mit `go-playground/validator` auf allen POST/PUT Endpoints
7. **Docker non-root User** in allen Dockerfiles
8. **Invoice-Erstellung transaktional** machen
9. **Handler-Tests** schreiben (HTTP-Layer komplett ungetestet)
10. **Offene Bugs fixen**: BUG-009, BUG-016, BUG-017, BUG-025
11. **Resource-Limits** in docker-compose setzen

### 🟢 P2 — Erste Wochen nach Go-Live

12. **Monitoring aufbauen**: Prometheus + Grafana + Alertmanager
13. **Frontend-Tests** einführen (Vitest + Testing Library)
14. **api.ts aufteilen** in 19 Module-Dateien
15. **Große Seiten refactoren** (13 Files >800 LOC)
16. **List-Virtualisierung** für Equipment/Invoice-Listen
17. **OpenAPI/Swagger** Dokumentation generieren
18. **SCSS-Konvention** standardisieren
19. **README aktualisieren** (Go 1.24)

---

## 8. Score-Zusammenfassung

| Bereich | Score | Status |
|---------|-------|--------|
| **Backend-Architektur** | 9.0/10 | Hexagonal, konsistent, sauber |
| **Code-Qualität (Go)** | 8.5/10 | Context 100%, parametrisierte SQL, sauberes DI |
| **Frontend** | 7.5/10 | Gut strukturiert, aber große Dateien + `any`-Lücken |
| **Security** | 7.0/10 | Starke Primitives, aber hardcoded Creds + TOFU |
| **Testing** | 4.0/10 | 8,7% Ratio, keine Handler/Frontend-Tests |
| **Infrastruktur** | 7.5/10 | Solides Docker-Setup, aber kein Monitoring |
| **Dokumentation** | 6.0/10 | README veraltet, keine API-Docs |

**Gesamt: 7.1 / 10 ⭐⭐⭐⭐**

---

## 9. Fazit

Die **Architektur und Code-Qualität sind stark** — hexagonale Struktur, saubere Context-Propagation, parametrisierte Queries, moderne Auth (Argon2id + RS256). Das Projekt zeigt professionelle Software-Engineering-Praktiken.

Die Hauptrisiken liegen in:

1. **Security-Hygiene**: Hardcoded Credentials und TOFU-Setup machen das System vor Go-Live angreifbar
2. **Test-Abdeckung**: 8,7% Ratio und CI-Test-Suppression sind nicht produktionsreif
3. **Observability**: Ohne Monitoring fliegt man blind

**Für Go-Live zwingend:** P0-Items (Credential-Rotation, Setup-Hardening, CI-Fix, Input-Validierung).

**Nice-to-have für V1:** Monitoring, Handler-Tests, API-Docs.

---

**Erstellt:** 7. April 2026
**Methode:** Automatisches verifiziertes Review (4 parallele Agents + manueller Deep-Dive)
**Reviewer:** Claude Code (Opus 4.6, 1M Context)
