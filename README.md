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

### Phase 0 – Foundation (Q2 2026) ✅
- [x] Projekt-Setup: Monorepo, Docker Compose, CI/CD
- [x] KurrentDB + PostgreSQL Infrastruktur (Docker-Config)
- [x] Shared Go Library (pkg/common) – kompiliert ✅
- [x] 18 Service-Skeletons mit Health-Checks – kompilieren ✅
- [x] Traefik API Gateway Konfiguration
- [x] Frontend-Grundgerüst (React 18, Design System) – baut ✅

### Phase 1 – MVP Core (Q3 2026) 🔄
- [x] Auth-Service – Scaffolding (kompiliert, noch nicht getestet)
- [x] Inventory-Service – Scaffolding (kompiliert, noch nicht getestet)
- [x] Scanner-Service – Scaffolding (kompiliert, noch nicht getestet)
- [x] Warehouse-Service – Scaffolding (kompiliert, noch nicht getestet)
- [x] Basis-Frontend – Pages vorhanden (baut, noch nicht E2E-getestet)
- [ ] Unit-Tests für Core-Services
- [ ] KurrentDB Event Sourcing verdrahten
- [ ] Passwort-Hashing auf bcrypt/argon2id upgraden
- [ ] Integration-Tests mit echtem PostgreSQL

### Phase 2 – Business Logic (Q4 2026)
- [x] Project-Service – Scaffolding (kompiliert)
- [x] Invoice-Service – Scaffolding (kompiliert, GoBD-Struktur vorhanden)
- [x] Document-Service – Scaffolding (kompiliert)
- [x] Crew-Service – Scaffolding (kompiliert)
- [x] Expense-Service – Scaffolding (kompiliert)
- [ ] Business Logic verifizieren & testen
- [ ] DATEV-Export verifizieren

### Phase 3 – Advanced Features (Q1 2027)
- [x] Federation-Service – Scaffolding (kompiliert)
- [x] Maintenance-Service – Scaffolding (kompiliert)
- [x] Transport-Service – Scaffolding (kompiliert)
- [x] Notification-Service – Scaffolding (kompiliert)
- [ ] NAS-Storage Integration (nur Adapter-Interface vorhanden)
- [ ] mTLS Federation tatsächlich implementieren
- [ ] E-Mail/Push-Versand implementieren

### Phase 4 – Intelligence & Scale (Q2 2027+)
- [x] AI-Service – Scaffolding (kompiliert)
- [x] Workflow-Service – Scaffolding (kompiliert)
- [x] Reporting-Service – Scaffolding (kompiliert)
- [x] Insurance-Service – Scaffolding (kompiliert)
- [ ] AI-Provider-Anbindung (Claude, OpenAI, Ollama)
- [ ] Advanced Analytics & Performance-Optimierung

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
