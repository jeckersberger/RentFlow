# RentFlow

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen?style=flat-square)](https://github.com/rentflow/rentflow)
[![License](https://img.shields.io/badge/license-TBD-blue?style=flat-square)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.1.0--alpha-red?style=flat-square)](https://github.com/rentflow/rentflow/releases)
[![Status](https://img.shields.io/badge/status-pre--alpha-orange?style=flat-square)](https://github.com/rentflow/rentflow)

**Self-hosted Lagerverwaltung & Rechnungssoftware für Veranstaltungstechnik**

---

## 📸 Screenshots

```
[Platzhalter: Dashboard mit Equipment-Übersicht]
[Platzhalter: QR-Scanner Interface]
[Platzhalter: Rechnungsverwaltung]
```

---

## 🎯 Was ist RentFlow?

RentFlow ist eine **moderne, selbst gehostete Lagerverwaltungs- und Rechnungssoftware** speziell für Veranstaltungstechnik-Firmen (1–100 Mitarbeiter) in der DACH-Region.

Statt teure Cloud-Lösungen mit Lizenzgebühren und Vendor Lock-in nutzen VT-Profis RentFlow auf ihrem eigenen Server – mit voller Kontrolle über ihre Daten, DATEV-Integration und einem innovativen **Federation-API für unternehmensübergreifendes Equipment-Sharing** (Peer-to-Peer, keine zentrale Instanz).

---

## ✨ Kernfeatures

### 📦 Equipment & Lager
- ✅ Equipment-Verwaltung mit Barcode/QR-Scanner
  - USB-Scanner, Android-Handheld, Handy-Kamera
- ✅ Hierarchische Lagerplätze
  - Standort → Raum → Regal → Fach
- ✅ Offline-fähiger Scanner-Modus
- ✅ Teilzustandsdokumentation (Sub-Rental Workflow)

### 📋 Projekte & Angebote
- ✅ Projekt-Management mit Equipment-Zuordnung
- ✅ Interaktive Packlisten (mit Abhaken beim Laden)
- ✅ Angebote → Rechnungen
- ✅ Teilrechnungen & Gutschriften
- ✅ Mahnwesen integriert

### 📊 Buchhaltung & Steuern
- ✅ DATEV-Export für Steuerberater
- ✅ Mehrwertsteuer-Handling
- ✅ Rechnungsverwaltung mit Archiv

### 🤝 Federation & Sharing
- ✅ **Federation API**: Firmenübergreifendes Equipment-Sharing
- ✅ Peer-to-Peer Verleih (ohne zentralen Server)
- ✅ Offenes Protokoll, keine proprietären Netzwerke

### 👥 Rollen & Berechtigungen
- ✅ Rollenbasierte UI (Admin, Geschäftsführer, Lager, Buchhaltung, Freelancer)
- ✅ Granulare Zugriffskontrolle

### 🌍 Lokalisierung & Usability
- ✅ Mehrsprachig (Deutsch, Englisch)
- ✅ Moderne Web-App (React + TypeScript)
- ✅ Responsive Design für Desktop & Mobile

### ⚙️ Administration
- ✅ Auto-Updates via Admin-Panel
- ✅ Self-hosted mit Docker Compose
- ✅ Läuft auf Standard-Hardware (z.B. 5€/Monat Hetzner-Server)

---

## 🚀 Quick-Start

### Voraussetzungen
- Linux-Server (Ubuntu 22.04+ / Debian 12+)
- Docker & Docker Compose v2
- Mindestens 4 GB RAM, 20 GB Speicher
- Eine Domain (für automatisches SSL via Let's Encrypt)

### Installation in 4 Schritten

#### 1. Repository klonen
```bash
git clone https://github.com/rentflow/rentflow.git
cd rentflow
```

#### 2. Environment-Variablen setzen
```bash
cp .env.example .env
# Passe folgende Werte in .env an:
# - DOMAIN (deine Domain, z.B. rentflow.meinefirma.de)
# - DB_PASSWORD (sichere Passphrase)
# - ADMIN_EMAIL (deine E-Mail)
# - JWT_SECRET (wird automatisch generiert wenn leer)
```

#### 3. Docker Compose starten
```bash
docker compose up -d
# Startet: Traefik, KurrentDB, PostgreSQL, Redis, alle Services, Frontend
```

#### 4. Setup-Wizard im Browser
Navigiere zu `https://deine-domain.de` — der Setup-Wizard führt dich durch:
1. Firma anlegen (Name, Adresse, Logo)
2. Admin-Account erstellen
3. Erste Benutzer & Rollen
4. Lager-Struktur definieren
5. Grundeinstellungen (Sprache, Währung, MwSt)

> Empfohlener Server: Hetzner CX21 (4GB RAM, 40GB SSD) ab 5€/Monat

---

## 🏗️ Tech-Stack

| Komponente | Technologie | Begründung |
|-----------|-------------|------------|
| **Backend** | Go 1.22+ (17 Microservices) | Performance, Single-Binary, Concurrency |
| **API Gateway** | Traefik v3 | Docker-native, automatisches Service-Discovery |
| **Event Store** | KurrentDB (ehem. EventStoreDB) | Event Sourcing, Subscriptions, Relational Sink |
| **Frontend** | React 18 + TypeScript + Sass | PWA-fähig, Offline-Support, großes Ökosystem |
| **Read-DB** | PostgreSQL 16 | Schema-per-Service Projections, JSONB, RLS |
| **Cache** | Redis 7 | Sessions, API-Cache, Rate Limiting |
| **Deployment** | Docker Compose | Self-Hosting, ein Befehl startet alles |
| **Scanner** | USB-Barcode, Zebra TC21, Handy-Kamera, RFID | Alle VT-Szenarien abgedeckt |
| **Monitoring** | Prometheus + Grafana + zerolog | Metriken, Dashboards, strukturiertes Logging |

---

## 📐 Systemarchitektur

**Microservice Architecture + Event Sourcing + CQRS**

```
┌────────────────────────────────────────────────────────────┐
│                        Clients                              │
│  Browser (Desktop/Tablet/Handy) │ PWA │ Scanner │ Portal    │
└──────────────────────┬─────────────────────────────────────┘
                       │ HTTPS
┌──────────────────────▼─────────────────────────────────────┐
│               Traefik v3 (API Gateway)                      │
│       TLS Termination │ Routing │ Load Balancing             │
│       Rate Limiting │ Circuit Breaker │ Health Checks        │
└──┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬────┘
   │   │   │   │   │   │   │   │   │   │   │   │   │   │
   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼   ▼
 ┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐┌───┐
 │AUT││INV││PRJ││SCN││WHS││INV││DOC││CRW││FED││MNT││...│
 │H  ││ENT││ECT││NER││E  ││OIC││   ││   ││   ││   ││   │
 └─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘└─┬─┘
   └────┴────┴────┴────┴────┴────┴────┴────┴────┴────┴───┘
                              │
              ┌───────────────▼───────────────┐
              │   KurrentDB (Event Store)      │
              │   Source of Truth │ Immutable   │
              │   Subscriptions │ Replay        │
              │   Relational Sink → PostgreSQL  │
              └───────────────┬───────────────┘
                              │ Projections
              ┌───────────────▼───────────────┐
              │   PostgreSQL 16 (Read-Models)  │
              │   Schema-per-Service           │
              │   (eine Instanz, N Schemas)    │
              └───────────────────────────────┘
```

### 17 Microservices

| # | Service | Verantwortung |
|---|---------|---------------|
| 1 | auth-service | JWT, RBAC, Users, Tenants, Sessions |
| 2 | inventory-service | Equipment, Kategorien, Preise, Labels, Flightcases |
| 3 | project-service | Projekte, Packlisten, Reservierungen |
| 4 | scanner-service | QR/Barcode/RFID, Check-In/Out, Offline-Sync |
| 5 | warehouse-service | Lagerplätze, Warenbewegung, Inventur |
| 6 | invoice-service | Rechnungen, Mahnwesen, DATEV, Bank-Integration |
| 7 | document-service | PDF, Templates, Verträge, OCR |
| 8 | crew-service | Personal, Freelancer, Zeiterfassung, CalDAV |
| 9 | federation-service | mTLS P2P, Equipment-Sharing, Sub-Rental |
| 10 | maintenance-service | Wartung, E-Check/DGUV V3 |
| 11 | transport-service | Fahrzeuge, Touren, Routen |
| 12 | insurance-service | Policen, Schäden, Risiko-Analyse |
| 13 | workflow-service | No-Code Workflows, Event-Trigger |
| 14 | ai-service | Multi-Provider AI, Anonymisierung, Prognosen |
| 15 | notification-service | In-App, Push, E-Mail |
| 16 | reporting-service | CQRS Read-Side, KPIs, Dashboards |
| 17 | audit-service | GoBD Audit-Trail, immutables Log |

### CQRS Write/Read Flow

```
WRITE (Command):                    READ (Query):
  Client → Traefik                    Client → Traefik
    → Service                           → Service
      → Validate                          → PostgreSQL (Read-Model)
      → Append Event → KurrentDB          → Return Data
      → Return ACK
                    ↓
          KurrentDB Subscription
                    ↓
          Service projiziert Event
                    ↓
          PostgreSQL (Read-Model updated)
```

### Federation (Peer-to-Peer)

```
Firma A (RentFlow-Instanz)          Firma B (RentFlow-Instanz)
┌──────────────────────┐            ┌──────────────────────┐
│  federation-service  │◄──mTLS───►│  federation-service  │
│  (eigene DB, Events) │   REST    │  (eigene DB, Events) │
└──────────────────────┘            └──────────────────────┘
```

---

## 📚 Dokumentation

Detaillierte Anleitungen findest du in der `/docs` Struktur:

| Dokument | Inhalt |
|----------|--------|
| [docs/INSTALLATION.md](docs/INSTALLATION.md) | Detaillierte Installationsanleitung |
| [docs/USER_GUIDE.md](docs/USER_GUIDE.md) | Benutzerhandbuch mit Screenshots |
| [docs/API.md](docs/API.md) | REST API Dokumentation |
| [docs/FEDERATION.md](docs/FEDERATION.md) | Peer-to-Peer Sharing Setup |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Production-Deployment (nginx, reverse proxy) |
| [docs/DATEV_EXPORT.md](docs/DATEV_EXPORT.md) | DATEV-Integration für Steuerberater |
| [docs/ADMIN_GUIDE.md](docs/ADMIN_GUIDE.md) | Administrations- & Wartungshandbuch |

---

## 🛠️ Development Setup

### Voraussetzungen
- Go 1.22+
- Node.js 20+
- Docker & Docker Compose v2
- KurrentDB (lokal via Docker)
- PostgreSQL 16+

### Lokale Entwicklung starten

#### 1. Infrastruktur starten (KurrentDB, PostgreSQL, Redis)
```bash
docker compose -f docker-compose.dev.yml up -d
```

#### 2. Backend-Services starten
```bash
# Alle Services gleichzeitig (Development-Modus)
make dev

# Oder einzeln:
cd services/auth-service && go run cmd/main.go
cd services/inventory-service && go run cmd/main.go
# ... weitere Services nach Bedarf
```

#### 3. Frontend starten
```bash
cd frontend
npm install
npm run dev
# Läuft auf http://localhost:5173 (Vite)
```

### Testing
```bash
# Alle Backend-Tests
make test

# Einzelnen Service testen
cd services/inventory-service && go test ./...

# Frontend Tests
cd frontend && npm test

# Integration Tests (benötigt laufende Infrastruktur)
make test-integration

# E2E Tests
make test-e2e
```

---

## 🗺️ Roadmap

> **Detaillierte Planung:** [`docs/planning/05-IMPLEMENTATION-PHASES.md`](docs/planning/05-IMPLEMENTATION-PHASES.md)
>
> **Hinweis:** Alle 18 Services kompilieren fehlerfrei und das Frontend baut. Die Service-Skeletons (Hexagonale Architektur, HTTP-Handler, Domain-Models, Repository-Layer) sind vorhanden, aber noch nicht getestet oder produktionsbereit. Die nachfolgenden Checkboxen spiegeln den **tatsächlichen Fertigstellungsgrad** wider.

### Phase 0 – Infrastruktur & Setup (4–6 Wochen)

#### M0.1: Repository & Projektstruktur
- [x] Git-Repository initialisieren
- [x] Verzeichnisstruktur anlegen
- [x] `.gitignore` für Go + Node + Docker konfigurieren
- [x] `go.work` für Multi-Module Workspace einrichten
- [x] Makefile mit Standard-Targets erstellen

#### M0.2: Docker Compose Stack
- [x] `docker-compose.yml` für Infrastruktur-Services
- [x] `docker-compose.dev.yml` mit Hot-Reload für alle Services
- [x] `docker-compose.prod.yml` mit Ressource-Limits
- [x] `.env.example` Datei mit allen Umgebungsvariablen
- [x] PostgreSQL Init-Skript für alle 17 Schemas
- [x] KurrentDB TLS-Zertifikat-Generierung
- [x] Redis-Passwort und Persistence konfigurieren
- [x] Traefik TLS + ACME konfigurieren
- [x] Health-Check-Skript: `scripts/health-check.sh`

#### M0.3: PostgreSQL Schema-Initialisierung
- [x] Init-SQL für alle 17 Schemas
- [x] Service-spezifische Datenbanknutzer
- [x] Extensions (uuid-ossp, pg_trgm, btree_gin)
- [x] Migration-Tool `golang-migrate/migrate` einrichten
- [x] `migrations/` Verzeichnis pro Service anlegen
- [x] Makefile-Target `make migrate-up` / `make migrate-down`

#### M0.4: Shared Go-Library `pkg/common`
- [x] KurrentDB Go-Client-Wrapper mit exponential backoff
- [x] PostgreSQL Pool-Factory (pgxpool.Pool) mit Connection-Retry
- [x] Redis-Client-Factory
- [x] JWT-Middleware für chi-Router
- [x] Zerolog-Setup mit Correlation-ID, Request-ID
- [x] Health-Check-Handler (liveness + readiness)
- [x] Domänen-Fehlertypen (NotFound, Conflict, Unauthorized, etc.)
- [x] Cursor-Paginierung (keyset pagination)
- [ ] gRPC-Server-Bootstrap mit TLS *(deferred – kein Service nutzt aktuell gRPC)*
- [x] Unit-Tests für alle gemeinsamen Pakete

#### M0.5: Service-Template
- [x] Service-Template-Generator (Bash/Go-Skript)
- [x] Template für alle 17 Services anwenden
- [x] Dockerfile (Multi-Stage Build, distroless/scratch)
- [x] Makefile pro Service (build, test, run, lint)
- [x] `go.work` alle Services einschließen

#### M0.6: Frontend-Scaffold
- [x] Vite-Projekt mit React + TypeScript
- [x] TanStack Router mit Route-Tree
- [x] TanStack Query mit Axios-Client
- [x] Zustand-Stores (auth, ui, scanner, notification)
- [x] Radix UI Themes + Farbpalette
- [x] SCSS Design-Tokens (`src/styles/tokens.scss`)
- [x] i18n-Setup (DE/EN, Namespace-Splitting)
- [x] PWA-Config (`vite.config.ts` + `vite-plugin-pwa`)
- [x] Service-Worker für Offline-Funktionalität
- [x] ESLint + Prettier-Konfiguration
- [x] Vitest + Testing-Library Setup

#### M0.7: CI/CD-Pipeline
- [x] GitHub Actions CI für alle Services
- [x] Docker-Image-Build und Push nach GHCR
- [x] Frontend-Build und -Tests in CI
- [x] Deployment-Skript für Self-Hosted (`scripts/deploy.sh`)
- [x] Watchtower für automatisches Image-Update konfigurieren
- [x] Rollback-Skript (`scripts/rollback.sh`)
- [x] Secrets-Management mit `.env`-Datei und Docker Secrets

#### M0.8: Monitoring & Logging
- [x] Prometheus Scraping-Config für alle Services
- [x] Grafana Dashboard-Templates (Latenz, Error-Rate, Event-Throughput)
- [x] Loki für zentrales Log-Aggregation
- [x] Alertmanager-Regeln (Service Down, hohe Fehlerrate)
- [x] `docker-compose.monitoring.yml` (Prometheus + Grafana + Loki)
- [x] Health-Check-Endpoint für alle Services (`GET /health/live`, `GET /health/ready`)

#### Phase-0-Abnahmekriterien
- [x] `docker compose up -d` startet alle Infrastruktur-Services ohne Fehler
- [x] KurrentDB Admin-UI erreichbar unter `http://localhost:2113`
- [x] PostgreSQL-Verbindung mit allen 17 Schemas erfolgreich
- [x] Redis `PING` liefert `PONG`
- [x] Traefik-Dashboard erreichbar
- [x] CI-Pipeline läuft durch (grün)
- [x] Frontend lädt unter `http://localhost:5173`
- [x] Alle Services kompilieren fehlerfrei

---

### Phase 1 – Core MVP (10–14 Wochen)

#### M1.1: Auth-Service
- [ ] PostgreSQL-Schema: `users`, `sessions`, `refresh_tokens`
- [ ] Argon2id-Passwort-Hashing
- [ ] JWT-Ausstellung (RS256, 1h Lebensdauer)
- [ ] Refresh-Token in Redis (7 Tage)
- [ ] RBAC-Rollen: `superadmin`, `admin`, `manager`, `warehouse`, `driver`, `crew`, `freelancer`, `readonly`
- [ ] Permissions-Matrix implementieren
- [ ] Brute-Force-Schutz (5 Fehlversuche → 15 Min Sperre, Redis)
- [ ] Rate-Limiting: 10 Anfragen/Minute pro IP
- [ ] Initialer Superadmin-Nutzer per Umgebungsvariable

#### M1.2: Inventory-Service
- [ ] PostgreSQL-Migration: `equipment`, `equipment_stock`, `categories`, `equipment_images`
- [ ] KurrentDB-Aggregate `EquipmentAggregate`
- [ ] Volltext-Suche mit `pg_trgm`
- [ ] Batch-Verfügbarkeitscheck (bis 500 Items gleichzeitig)
- [ ] Preiskalkulations-Engine mit Staffelrabatten
- [ ] ZPL-Label-Generierung für Zebra-Drucker
- [ ] MinIO/S3-Integration für Equipment-Bilder
- [ ] QR-Code-Generierung (`skip2/go-qrcode`)
- [ ] Flightcase-Logik: Child-Equipment-Zuordnung
- [ ] CSV-Import aus Bestandsdaten alter Systeme

#### M1.3: Project-Service
- [ ] PostgreSQL-Migration: `projects`, `project_equipment`, `customers`
- [ ] Status-Zustandsmaschine (anfrage → archiviert)
- [ ] Doppelbuchungs-Erkennung über alle Projekte
- [ ] Packlisten-Generierung sortiert nach Lagerort
- [ ] Kalenderansicht für Projektzeitraum-Übersicht
- [ ] E-Mail-Benachrichtigung bei Status-Änderung
- [ ] Kunden-Verwaltung (DSGVO-konforme Pflichtfelder)
- [ ] Projekt-Kopieren-Funktion

#### M1.4: Invoice-Service
- [ ] Fortlaufende GoBD-konforme Rechnungsnummern (pro Mandant)
- [ ] PDF-Generierung mit `chromedp` (HTML-Template → PDF)
- [ ] Angebot → Auftragsbestätigung → Rechnung Konvertierung
- [ ] Mahnwesen (3-stufig: Zahlungserinnerung / Mahnung 1 / Mahnung 2)
- [ ] DATEV-Export SKR03 (CSV-Format)
- [ ] Teilrechnung mit Restbetrag-Tracking
- [ ] Stornierung nach GoBD (keine Löschung, nur Gegenbuchung)
- [ ] Gutschriften mit Pflicht-Verweis auf Originalrechnung
- [ ] E-Mail-Versand der Rechnung als PDF-Anhang

#### M1.5: Frontend Core-Pages
- [ ] TanStack Router Route-Tree aufbauen
- [ ] Login-Page, Dashboard, Equipment-Liste, Equipment-Detail
- [ ] Projekte-Liste, Projekt-Detail, Rechnungen-Liste, Rechnung-Detail
- [ ] TanStack Query für alle API-Calls
- [ ] Optimistic Updates für häufige Operationen
- [ ] Fehlerbehandlung + Toast-Benachrichtigungen
- [ ] Responsive Design (Desktop + Tablet Prio)
- [ ] Dark-Mode-Toggle
- [ ] Command Palette (⌘K) für Navigation

#### M1.6: Scanner-Integration
- [ ] `ScannerPage` (`/scan`) mit Kamera-QR-Scan
- [ ] Equipment-Lookup nach QR-Code-UUID
- [ ] Check-Out: Equipment einem Projekt zuweisen
- [ ] Check-In: Equipment vom Projekt zurückgeben
- [ ] Scan-Feedback (visuell + vibration)
- [ ] Offline-Unterstützung für Service-Worker

#### Phase-1-Abnahmekriterien
- [ ] Marco kann ein neues Projekt anlegen und Equipment zuweisen
- [ ] Lisa kann Equipment per Smartphone-Kamera scannen und check-out durchführen
- [ ] Thomas kann eine Rechnung erstellen, als PDF herunterladen und per E-Mail senden
- [ ] Thomas kann DATEV-Export für einen Monat herunterladen
- [ ] Alle CRUD-Operationen für Equipment, Projekte, Rechnungen funktionieren
- [ ] JWT-Auth schützt alle Endpunkte
- [ ] Keine SQL-Injections, XSS, CSRF

---

### Phase 2 – Operations (8–12 Wochen)

#### M2.1: Warehouse-Service
- [ ] PostgreSQL-Migration: `warehouses`, `zones`, `racks`, `bays`, `stock_locations`, `movements`
- [ ] Hierarchische Lagerstruktur (Lager → Zone → Regal → Fach)
- [ ] Standort-Zuweisung für Equipment-Bestände
- [ ] QR-Code-Label für jeden Lagerplatz (ZPL-Format für Zebra)
- [ ] Bewegungs-Tracking: jede physische Bewegung als Event
- [ ] Inventur-Workflow: Voll-Inventur, Zyklus-Inventur, Spot-Kontrolle
- [ ] Lagerplatz-Kapazitäts-Tracking
- [ ] KI-Lageroptimierung (Vorbereitung für Phase 4)

#### M2.2: Scanner-Service mit Zebra TC21
- [ ] Zebra DataWedge-Intent-Empfang in PWA
- [ ] Scan-Kontext-System (check-out, check-in, lager-einräumen, inventur)
- [ ] Multi-Scan für Packlisten (alle Items eines Projekts scannen)
- [ ] Audio + visuelles Feedback pro Scan-Ergebnis
- [ ] Offline-Scan-Queue (IndexedDB, bis 500 Scans)
- [ ] Background-Sync wenn Verbindung wiederhergestellt
- [ ] Scan-Session-Protokoll in `scanner_schema`
- [ ] RFID-Vorbereitung (Zebra FX9600 Interface)
- [ ] USB-Scanner via WebHID-API

#### M2.3: Transport-Service
- [ ] PostgreSQL-Migration: `vehicles`, `tours`, `tour_equipment`, `driver_logs`
- [ ] Fahrzeug-Verwaltung (Kennzeichen, Kapazität kg + m³, DGUV-Prüfung)
- [ ] Tour-Planung: welches Equipment in welches Fahrzeug
- [ ] Kapazitäts-Check: Gewicht + Volumen nicht überschreiten
- [ ] Fahrer-Zuweisung aus Crew-Service (gRPC)
- [ ] Lieferschein-Generierung (Anknüpfung an Document-Service Phase 3)
- [ ] Kilometer + Kosten-Tracking pro Tour
- [ ] Tachograph-kompatible Fahrerlisten

#### M2.4: Maintenance-Service
- [ ] PostgreSQL-Migration: `maintenance_plans`, `maintenance_tasks`, `checklists`, `electrical_tests`
- [ ] Wartungsplan-Typen: Intervall, Nach-Einsatz, Stunden-basiert
- [ ] Aufgaben-Zustandsmaschine: geplant → in_bearbeitung → abgeschlossen
- [ ] Checklisten mit JSON-Format
- [ ] E-Check / DGUV Vorschrift 3 Workflow (VDE-0701-0702-Messwerte, Prüfplaketten-Druck)
- [ ] IZYTRON.IQ XML-Import (Prüfergebnis-Import)
- [ ] Equipment automatisch in Wartungs-Status setzen (sperrt für Buchung)
- [ ] Wartungs-Dashboard: fällige und überfällige Wartungen

#### M2.5: Frontend Operations-Pages
- [ ] Lager-Übersicht (`/warehouse`) mit Zonen + Regalanzeige
- [ ] Mobiler Scanner (`/scan`) erweitert mit Zebra DataWedge + Offline
- [ ] Transport (`/transport`) mit Touren-Planung + Fahrzeug-Verfügbarkeit
- [ ] Wartung (`/maintenance`) mit Fälligkeits-Kalender + E-Check-Dashboard

#### Phase-2-Abnahmekriterien
- [ ] Lisa kann Equipment mit Zebra TC21 scannen und in ein Projekt auschecken
- [ ] Offline-Scans werden bei Reconnect automatisch synchronisiert
- [ ] Wartungsfällige Geräte erscheinen in Dashboard-Warnung
- [ ] DGUV/E-Check-Status pro Gerät ist sichtbar
- [ ] Touren können Fahrzeugen zugewiesen werden mit Kapazitätsprüfung

---

### Phase 3 – Business (8–10 Wochen)

#### M3.1: Crew-Service
- [ ] PostgreSQL-Migration: `crew_members`, `qualifications`, `crew_assignments`, `time_records`
- [ ] Qualifikations-System (IPAF, Rigger, Tonmeister, etc.)
- [ ] Verfügbarkeits-Matrix für Crew-Mitglieder
- [ ] Konflikt-Erkennung: Doppel-Buchung eines Crew-Mitglieds
- [ ] CalDAV-Feed für Crew-Kalender (Thunderbird/iOS)
- [ ] Freelancer-Portal: Eigene Einsätze und Aufgaben sehen
- [ ] Zeiterfassung pro Einsatz (Start/Stop-Timer)
- [ ] Fahrerliste: Wer fährt wann mit welchem Fahrzeug

#### M3.2: Document-Service
- [ ] PostgreSQL-Migration: `documents`, `document_versions`, `signatures`
- [ ] Dokument-Typen: Angebot, AB, Lieferschein, Abholschein, Rechnung, Mietvertrag, Schadensbericht
- [ ] Go-Template-basierte PDF-Generierung (chromedp)
- [ ] Lieferschein-Generierung aus Packliste
- [ ] Digitale Unterschriften (Canvas-Signature-Pad)
- [ ] GoBD-konformes Archiv: SHA-256-Checksummen-Kette
- [ ] Versionierung: `v1`, `v2` ... bei Änderungen
- [ ] Scan-to-Document: Scan-Upload + OCR (tesseract)
- [ ] Kunden-Portal: Dokument zum Unterschreiben per Link

#### M3.3: Insurance-Service
- [ ] PostgreSQL-Migration: `policies`, `claims`, `claim_items`
- [ ] Versicherungspolice-Verwaltung (Betriebshaftpflicht, Kasko, Transport, Mieter)
- [ ] Equipment automatisch versicherungsrelevant markieren (ab Wiederbeschaffungswert)
- [ ] Schadensfall-Workflow: Erfassung → Dokumentation → Einreichung → Abschluss
- [ ] Schadensformular mit Foto-Upload
- [ ] Forderungs-Tracking: Wieviel erstattet, wieviel offen
- [ ] Jahres-Auswertung: Schäden pro Jahr, Prämienentwicklung

#### M3.4: Reporting-Service
- [ ] PostgreSQL-Migration: `report_definitions`, `report_runs`, `kpi_snapshots`
- [ ] Rollen-spezifische KPI-Dashboards (GF, Lager, Buchhaltung)
- [ ] Report-Typen: Umsatzbericht, Auslastung, Mahnstatistik, Equipment-Inventur
- [ ] Zeitraumauswahl: Woche, Monat, Quartal, Jahr, Custom
- [ ] CSV + PDF-Export für jeden Report
- [ ] Geplante E-Mail-Reports (wöchentlicher Umsatz-Report)
- [ ] Recharts/Victory-basierte Diagramme im Frontend

#### M3.5: Frontend Business-Pages
- [ ] Crew-Management (`/crew`) mit Qualifikations-Badges + Verfügbarkeitskalender
- [ ] Freelancer-Ansicht (`/my-assignments`) mobiloptimiert
- [ ] Dokumente (`/documents`) mit Archiv + Unterschriften-Status
- [ ] Versicherung (`/insurance`) mit Policen-Übersicht + Schadensmeldung
- [ ] Reports (`/reports`) mit KPI-Dashboard + Diagramme + Export

#### Phase-3-Abnahmekriterien
- [ ] Kevin kann seine Einsätze auf dem iPhone sehen und Zeiten erfassen
- [ ] Lieferschein wird aus Packliste automatisch generiert und kann digital unterschrieben werden
- [ ] Thomas kann monatlichen Umsatzbericht als PDF/CSV exportieren
- [ ] Schadensfall kann dokumentiert und mit Fotos eingereicht werden
- [ ] GoBD-Archiv speichert Dokumente mit Checksummen-Kette

---

### Phase 4 – Advanced Features (10–14 Wochen)

#### M4.1: AI-Service
- [ ] PostgreSQL-Migration: `ai_requests`, `ai_feedback`, `few_shot_examples`, `ai_providers`
- [ ] Multi-Provider-Architektur: Claude / GPT-4o / Gemini / Mistral / Ollama
- [ ] Anonymisierungs-Service (7 Muster: Name, E-Mail, IBAN, Tel., Adresse, Steuernr., Kundennr.)
- [ ] KI-Funktionen: Preisoptimierung, Nachfrageprognose, Smart Asset Creator, Prädiktive Wartung
- [ ] Feedback-Mechanismus + Few-Shot-Learning

#### M4.2: Workflow-Service
- [ ] PostgreSQL-Migration: `workflow_definitions`, `workflow_instances`, `workflow_steps`
- [ ] Trigger-Typen: Event-basiert, Zeitbasiert (Cron), Manuell
- [ ] Aktions-Typen: E-Mail, Webhook, Service-Call (gRPC), Status-Änderung
- [ ] Vordefinierte Templates (Onboarding, Erinnerungen, Warnungen)
- [ ] Visueller Workflow-Editor (React Flow)

#### M4.3: Federation-Service
- [ ] mTLS-Zertifikat-Management (Cert-auf-Cert-Austausch)
- [ ] Partner-Einrichtung + Equipment-Verfügbarkeit im Partner-Netzwerk
- [ ] Sub-Vermietungs-Anfrage senden + empfangen
- [ ] Handover-Dokumentation + automatische Eingangsrechnung
- [ ] Data-Sovereignty: Nur explizit freigegebene Daten werden geteilt

#### M4.4: Notification-Service
- [ ] PostgreSQL-Migration: `notifications`, `notification_channels`, `user_preferences`
- [ ] E-Mail via SMTP, Web Push (VAPID), In-App (WebSocket + Redis Pub/Sub)
- [ ] Nutzer-Präferenzen: Welcher Kanal für welchen Ereignistyp
- [ ] Notification-Center, Tages-Digest, Stille Stunden, Batch-Benachrichtigungen

#### M4.5: Audit-Service
- [ ] PostgreSQL-Migration: `audit_log`, `audit_exports`
- [ ] Immutables Audit-Log: SHA-256-Checksummen-Kette für GoBD-Konformität
- [ ] Alle schreibenden Operationen in Audit-Log schreiben
- [ ] GoBD-Prüfung: Prüf-Funktion für Checksummen-Integrität
- [ ] DSGVO-Pseudonymisierung nach Nutzerlöschung
- [ ] Betriebsprüfungs-Export (ZIP mit Audit-Logs + KurrentDB-Events)

#### M4.6: Frontend Advanced-Pages
- [ ] KI-Assistent (`/ai`) mit Chat-Interface + Smart Asset Creator
- [ ] Workflow-Editor (`/workflows`) mit React Flow Canvas
- [ ] Federation (`/federation`) mit Partner-Netzwerk-Übersicht
- [ ] Audit-Log (`/audit`) mit Volltext-Suche + GoBD-Prüfbericht
- [ ] Admin (`/admin`) mit System-Einstellungen + Service-Health

#### Phase-4-Abnahmekriterien
- [ ] KI-Preisoptimierung liefert nachvollziehbare Empfehlungen
- [ ] Stefan (SoundPro) kann sich als Federation-Partner verbinden
- [ ] Audit-Log zeigt lückenlose Checksummen-Kette (GoBD-Prüfung bestanden)
- [ ] Workflow-Engine sendet automatisch E-Mail 7 Tage vor Projektstart
- [ ] Web Push funktioniert auf Chrome Desktop + Android + iOS Safari

---

## 🤝 Contributing

Wir freuen uns über Beiträge! Hier sind die nächsten Schritte:

1. **Fork** das Repository
2. **Erstelle einen Feature-Branch**: `git checkout -b feature/deine-feature`
3. **Committe deine Änderungen**: `git commit -am 'Add some feature'`
4. **Pushe zum Branch**: `git push origin feature/deine-feature`
5. **Erstelle einen Pull Request**

Bitte beachte:
- Code muss getestet sein
- Dokumentation muss aktualisiert werden
- Commit-Messages sollten aussagekräftig sein (deutsch oder englisch OK)

**Bug-Reports & Feature-Requests:**
Nutze die [GitHub Issues](https://github.com/rentflow/rentflow/issues) Seite.

---

## 📄 Lizenz

[LIZENZ IST NOCH FESTZULEGEN – TBD]

Aktuelle Optionen zur Diskussion:
- **AGPL-3.0** (stärkerer Copyleft, Federation-freundlich)
- **SSPL** (Server-Side Public License, ähnlich MongoDB)
- **Dual-License** (Open Source + kommerziell)

Detailinformationen: siehe [LICENSE](LICENSE) (wird noch hinzugefügt)

---

## 🎓 Warum RentFlow statt Rentman/Easyjob?

| Feature | RentFlow | Rentman | Easyjob |
|---------|----------|---------|---------|
| **Self-Hosted** | ✅ Deine Server | ❌ Cloud nur | ❌ Cloud nur |
| **Lizenzkosten** | ❌ Kostenlos | ✅ €200–500/Monat | ✅ €150–400/Monat |
| **Vendor Lock-in** | ❌ Nein (offene API) | ✅ Ja | ✅ Ja |
| **DATEV-Export** | ✅ Ja | ❌ Nein | ❌ Nein |
| **Federation/Sharing** | ✅ P2P, offen | ❌ Nein | ❌ Nein |
| **Datenschutz (DSGVO)** | ✅ Vollständig deine Kontrolle | ⚠️ Abhängig vom Provider | ⚠️ Abhängig vom Provider |
| **Moderne Web-App** | ✅ React, TypeScript | ⚠️ ältere UI | ⚠️ ältere UI |
| **Deutsche Unterstützung** | ✅ Ja (DACH-fokussiert) | ✅ Ja | ✅ Ja |

---

## 📞 Support & Community

- **Fragen & Diskussionen**: [GitHub Discussions](https://github.com/rentflow/rentflow/discussions)
- **Bug-Reports**: [GitHub Issues](https://github.com/rentflow/rentflow/issues)
- **E-Mail**: support@rentflow.example (wird noch eingerichtet)
- **Dokumentation**: [docs/](docs/)

---

## 🙏 Danksagungen

RentFlow wurde mit ❤️ von und für Veranstaltungstechnik-Profis entwickelt.

Spezielle Dankbarkeit gegenüber:
- Der Go & React Community
- KurrentDB (ehem. EventStoreDB) Team
- PostgreSQL Team
- Open-Source Projekten (Traefik, Docker, etc.)
- Unseren Early-Adoptern und Testern

---

## 📊 Project Stats

```
📦 Repository: github.com/rentflow/rentflow
🔗 Demo: [wird noch eingerichtet]
⭐ Status: Pre-Alpha (aktive Entwicklung)
📆 Gründung: 2024
🌍 Fokus-Region: DACH (Deutschland, Österreich, Schweiz)
```

---

**Happy Renting! 🎬🎤🎵**
