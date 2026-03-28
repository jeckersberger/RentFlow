# Claude Code Configuration — EquipFlow

## Projekt

EquipFlow ist eine selbstgehostete Equipment-Rental-Software fuer die Veranstaltungstechnik-Branche.
Entwickelt als Ersatz fuer Rentman, vollstaendig Open Source, optimiert fuer kleine bis mittlere Verleiher.

## Architektur

- **19 Go-Microservices** unter `services/` (Hexagonale Architektur pro Service)
- **React 18 + Vite Frontend** unter `frontend/`
- **Shared Library** unter `pkg/common/`
- **PostgreSQL 16** mit separater Datenbank pro Service
- **Redis 7** fuer Sessions, Cache, Pub/Sub
- **Traefik v3.1** als Reverse Proxy (Production)

### Services

auth, inventory, project, scanner, warehouse, invoice, document, crew,
federation, maintenance, transport, insurance, workflow, ai, notification,
reporting, audit, expense, customer

## Verbindliche Regeln

### Go-Backend

- Package-Name fuer HTTP-Handler: `httphandler` (NICHT `http`)
- Hexagonale Architektur: `domain/` -> `ports/` -> `adapters/` -> `cmd/`
- Geldbetraege IMMER als `int64` (Cent), NIE float
- Schema-per-Service: Jeder Service hat seine eigene PostgreSQL-Datenbank
- Error Handling: Eigene Domain-Error-Typen, kein nacktes `error`
- Context-Propagation: Immer `context.Context` als ersten Parameter

### Frontend

- Plain `.scss` Dateien, KEINE `.module.scss`
- Framer Motion fuer Animationen
- React Router v6 mit Lazy Loading
- Zustand fuer State Management
- Alle API-Calls ueber zentrale `apiClient`-Instanz

### Qualitaet

- Tests MUESSEN bestehen bevor zum naechsten Service gewechselt wird
- Jeder Service braucht mindestens: Unit Tests (Domain), Integration Tests (Adapter), Handler Tests
- Coverage-Ziel: >80% fuer Domain-Logic
- `golangci-lint` muss fehlerfrei durchlaufen

## Build & Test

```bash
# Infrastruktur starten
make docker-up

# Alle Services bauen
make build

# Alle Tests
make test

# Lint
make lint

# Migrationen
make migrate
```

## Referenz-Dokumentation

- Architektur: `docs/architecture/`
- Service-Specs: `docs/specs/`
- Planungsdokumente: `docs/planning/`
- KI-Hirn: `C:\Users\jecke\Documents\Claude\Obsidian Hirn\KI-Hirn\Projekte\RentFlow TODOs.md`

## Datei-Organisation

- `/services/{name}/` — Microservice mit eigenem go.mod
- `/pkg/common/` — Geteilte Bibliothek (Logger, Middleware, DB-Helpers)
- `/frontend/src/` — React-Anwendung
- `/scripts/` — Hilfsskripte (DB-Init, Deployment)
- `/docs/` — Dokumentation, ADRs, Specs
- `/config/` — Konfigurationsdateien

## Sicherheit

- JWT RS256 mit Rotation
- Argon2id fuer Passwort-Hashing
- Brute-Force-Protection via Redis Rate Limiting
- Input-Validierung an allen System-Grenzen
- KEINE Secrets in Source-Files oder Commits
