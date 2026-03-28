# EquipFlow v1.0.0

Selbstgehostete Equipment-Rental-Software fuer die Veranstaltungstechnik.
Verwaltet Inventar, Projekte, Lager, Rechnungen, Crew und mehr — als Open-Source-Alternative zu Rentman.

## Quick Start

```bash
# 1. Umgebungsvariablen einrichten
cp .env.example .env
# .env bearbeiten und sichere Passwoerter setzen

# 2. Infrastruktur starten (PostgreSQL, Redis, Traefik)
docker compose up -d

# 3. Services bauen und testen
make build
make test
```

## Tech Stack

| Komponente | Technologie |
|------------|-------------|
| Backend | Go 1.23, 19 Microservices, Hexagonale Architektur |
| Frontend | React 18, Vite, TypeScript, SCSS, Framer Motion |
| Datenbank | PostgreSQL 16 (eine DB pro Service) |
| Cache | Redis 7 (Sessions, Pub/Sub) |
| Reverse Proxy | Traefik v3.1 |
| Auth | JWT RS256, Argon2id |

## Projektstruktur

```
services/          19 Go-Microservices
pkg/common/        Geteilte Bibliothek
frontend/src/      React-Anwendung
scripts/           Hilfsskripte
docs/              Architektur, Specs, Planung
```

## Lizenz

TBD
