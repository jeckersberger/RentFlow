# Architektur-Übersicht

## Tech-Stack

| Komponente | Technologie |
|-----------|-------------|
| Frontend | TypeScript + React + Sass/SCSS |
| Backend | Go (REST API) |
| Datenbank | PostgreSQL |
| Deployment | Docker / Docker Compose |
| i18n | Mehrsprachig (DE/EN) von Beginn an |
| Scanner | USB-Barcode, Android-Handheld, Handy-Kamera |

## System-Architektur

```
┌─────────────────────────────────────────────────┐
│                    Clients                       │
│  Browser (PC/Laptop) │ Tablet │ Handy │ Scanner  │
└──────────────────────┬──────────────────────────┘
                       │ HTTPS
┌──────────────────────▼──────────────────────────┐
│              Go Backend (API)                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │ Lager    │ │Rechnungen│ │ Federation-API   │ │
│  │ Modul    │ │ Modul    │ │ (Firma↔Firma)    │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │ Auth/    │ │ Scanner  │ │ PDF/Export       │ │
│  │ Rollen   │ │ Service  │ │ Service          │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
└──────────────────────┬──────────────────────────┘
                       │
         ┌─────────────▼─────────────┐
         │      PostgreSQL DB        │
         └───────────────────────────┘

Firma A ◄──── Federation API (mTLS) ────► Firma B
```

## Multi-Firma Federation

- Jede Firma hostet eigene Instanz (eigener Server, eigene DB)
- Peer-to-Peer Verbindung über sichere REST-API mit mTLS (mutual TLS)
- Firmen können sich gegenseitig "verknüpfen" (Einladung/Bestätigung)
- Mögliche Interaktionen: Equipment-Verfügbarkeit abfragen, Ausleihen, Sub-Rental
- Jede Firma kontrolliert selbst welche Daten geteilt werden

## Versionierung & Update-System

### Semantic Versioning (SemVer)
- Format: `MAJOR.MINOR.PATCH` (z.B. `1.2.3`)
- MAJOR = Breaking Changes / große neue Features
- MINOR = Neue Features, abwärtskompatibel
- PATCH = Bugfixes, kleine Verbesserungen

### Git-Workflow
```
main (stable releases)
 ├── develop (aktive Entwicklung)
 │    ├── feature/*
 │    └── fix/*
 └── release/vX.Y.Z (Release-Vorbereitung)
```

### Commit-Konventionen (Conventional Commits)
```
<type>(<scope>): <kurze Beschreibung>
```
Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`

### Auto-Update via UI
1. Backend prüft GitHub auf neue Releases
2. Admin sieht Update-Banner mit Changelog
3. Klick auf "Jetzt updaten"
4. Automatisch: Docker-Image pull → DB-Backup → Migration → Health-Check
5. Bei Fehler: automatischer Rollback

### DB-Migrationen
- Automatisch beim App-Start
- Versionierte Migrations-Dateien im Repo
- Rollback-Skripte für jede Migration
- Automatisches Backup vor jeder Migration

## Entscheidungen

| Entscheidung | Begründung |
|-------------|-----------|
| Go statt Node.js | Performance, einfaches Deployment (single binary), starke Concurrency |
| PostgreSQL statt MySQL | Bessere JSON-Unterstützung, bessere Erweiterbarkeit, robuster |
| React statt Vue | Größeres Ökosystem, mehr Libraries, TypeScript-First |
| Docker | Einfaches Self-Hosting, reproduzierbare Umgebung, einfache Updates |
| mTLS für Federation | Gegenseitige Authentifizierung, keine zentrale Authority nötig |
| REST statt gRPC (Federation) | Einfacher zu debuggen, Firewall-freundlicher, breiter unterstützt |
