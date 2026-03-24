# RentFlow

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen?style=flat-square)](https://github.com/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware)
[![License](https://img.shields.io/badge/license-TBD-blue?style=flat-square)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.1.0--alpha-red?style=flat-square)](https://github.com/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware/releases)
[![Status](https://img.shields.io/badge/status-pre--alpha-orange?style=flat-square)](https://github.com/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware)

**Self-hosted Lagerverwaltung & Rechnungssoftware fuer Veranstaltungstechnik**

---

## Was ist RentFlow?

RentFlow ist eine **moderne, selbst gehostete Lagerverwaltungs- und Rechnungssoftware** speziell fuer Veranstaltungstechnik-Firmen (1-100 Mitarbeiter) in der DACH-Region.

Statt teure Cloud-Loesungen mit Lizenzgebuehren und Vendor Lock-in nutzen VT-Profis RentFlow auf ihrem eigenen Server -- mit voller Kontrolle ueber ihre Daten, DATEV-Integration und einem innovativen **Federation-API fuer unternehmensuebergreifendes Equipment-Sharing** (Peer-to-Peer, keine zentrale Instanz).

---

## Quick Start

### Systemanforderungen

| Anforderung | Minimum |
|---|---|
| Docker | 24+ |
| Docker Compose | v2 |
| RAM | 4 GB |
| Festplatte | 10 GB |
| OS | Linux (Ubuntu 22.04+ / Debian 12+) |
| Domain | Empfohlen (fuer SSL via Let's Encrypt) |

### Installation in 4 Schritten

```bash
# 1. Repository klonen
git clone https://github.com/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware.git rentflow
cd rentflow

# 2. Environment-Variablen konfigurieren
cp .env.example .env
# .env bearbeiten: Passwoerter, Domain, E-Mail anpassen

# 3. Starten
docker compose up -d

# 4. Browser oeffnen
# http://localhost:3000 -> Setup-Wizard
```

> **Empfohlener Server:** Hetzner CX21 (4 GB RAM, 40 GB SSD) ab 5 EUR/Monat

---

## Standard-Ports

| Service | Port | Beschreibung |
|---|---|---|
| **Frontend** | 3000 | React Web-App (Nginx) |
| **Traefik** | 80 / 443 | Reverse Proxy, SSL Termination |
| **Traefik Dashboard** | 8080 | Admin-Dashboard |
| **PostgreSQL** | 5432 | Datenbank |
| **Redis** | 6379 | Cache & Sessions |
| **KurrentDB** | 2113 | Event Store (Admin-UI) |
| **KurrentDB TCP** | 1113 | Event Store TCP |
| **Prometheus** | 9090 | Metriken |
| auth-service | 8001 | JWT, RBAC, Users |
| inventory-service | 8002 | Equipment, Kategorien |
| project-service | 8003 | Projekte, Packlisten |
| scanner-service | 8004 | QR/Barcode Scanner |
| warehouse-service | 8005 | Lagerplaetze |
| invoice-service | 8006 | Rechnungen, DATEV |
| document-service | 8007 | PDF, Templates, OCR |
| crew-service | 8008 | Personal, Zeiterfassung |
| federation-service | 8009 | P2P Equipment-Sharing |
| maintenance-service | 8010 | Wartung, DGUV V3 |
| transport-service | 8011 | Fahrzeuge, Touren |
| insurance-service | 8012 | Versicherungen |
| workflow-service | 8013 | No-Code Workflows |
| ai-service | 8014 | KI Multi-Provider |
| notification-service | 8015 | E-Mail, Push, In-App |
| reporting-service | 8016 | Reports, KPIs |
| audit-service | 8017 | GoBD Audit-Trail |
| expense-service | 8018 | Ausgabenverwaltung |

---

## Erster Start (Setup-Wizard)

Nach `docker compose up -d` erreichst du unter `http://localhost:3000` den Setup-Wizard:

1. **Firma anlegen** -- Name, Adresse, Logo hochladen
2. **Admin-Account erstellen** -- E-Mail und sicheres Passwort
3. **SMTP konfigurieren** -- E-Mail-Versand fuer Rechnungen und Benachrichtigungen (unter Einstellungen)
4. **AI-Provider** (optional) -- Claude, OpenAI oder Ollama (unter Einstellungen > Integrationen)
5. **Lagerstruktur definieren** -- Standort, Raeume, Regale, Faecher
6. **Benutzer & Rollen** -- Mitarbeiter einladen (Admin, Lager, Buchhaltung, Freelancer)

---

## Entwicklung (Development)

### Voraussetzungen

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose v2

### Lokale Entwicklung starten

```bash
# 1. Infrastruktur starten (KurrentDB, PostgreSQL, Redis)
docker compose -f docker-compose.dev.yml up -d

# 2. Backend-Services starten
make dev

# 3. Frontend starten (Vite mit Hot Reload)
cd frontend
npm install
npm run dev
# Laeuft auf http://localhost:5173
```

### Tests

```bash
# Alle Backend-Tests
make test

# Frontend Tests
cd frontend && npm test

# Integration Tests (benoetigt laufende Infrastruktur)
make test-integration

# E2E Tests
make test-e2e
```

---

## Produktion

Fuer den Produktionsbetrieb den Production-Override verwenden:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

Dies aktiviert:
- Ressource-Limits (Memory, CPU) fuer alle Services
- Nur Ports 80/443 nach aussen (Traefik Routing)
- Security-Haertung (read-only Filesystem, dropped Capabilities)
- Strukturiertes JSON-Logging mit Rotation

---

## Backup

Das mitgelieferte Backup-Script sichert PostgreSQL, KurrentDB und Uploads:

```bash
# Manuelles Backup
./scripts/backup.sh

# Automatisches taegliches Backup (Crontab)
# 0 3 * * * /opt/rentflow/scripts/backup.sh >> /var/log/rentflow-backup.log 2>&1
```

**Was wird gesichert:**
- PostgreSQL (alle 18 Service-Datenbanken)
- KurrentDB Event Store
- Upload-Verzeichnis (Dokumente, Bilder)
- Verschluesselung mit AES-256 (konfigurierbar)
- Optionaler S3-Upload

**Wiederherstellung:**

```bash
./scripts/restore.sh /opt/rentflow/backups/<backup-datei>.enc
```

Konfiguration ueber `.env`:
- `BACKUP_ENCRYPTION_KEY` -- Verschluesselungsschluessel
- `BACKUP_RETENTION_DAYS` -- Aufbewahrungsdauer (Standard: 30 Tage)
- `BACKUP_S3_ENABLED` -- S3-Upload aktivieren

---

## Update

```bash
# Standard-Update
./scripts/update.sh

# Nur bestimmte Services updaten
./scripts/update.sh auth-service frontend
```

Bei Problemen steht ein Rollback-Script bereit:

```bash
./scripts/rollback.sh
```

---

## Tech-Stack

| Komponente | Technologie |
|---|---|
| **Backend** | Go 1.22+ (18 Microservices) |
| **API Gateway** | Traefik v3 |
| **Event Store** | KurrentDB (ehem. EventStoreDB) |
| **Frontend** | React 18 + TypeScript + Sass |
| **Read-DB** | PostgreSQL 16 |
| **Cache** | Redis 7 |
| **Deployment** | Docker Compose |
| **Monitoring** | Prometheus + Grafana + Loki |

---

## Architektur

**Microservice Architecture + Event Sourcing + CQRS**

```
Clients (Browser / PWA / Scanner)
          |
    Traefik v3 (TLS, Routing, Rate Limiting)
          |
    18 Go Microservices
          |
    KurrentDB (Event Store, Source of Truth)
          |
    PostgreSQL 16 (Read-Models, Schema-per-Service)
```

### 18 Microservices

| # | Service | Verantwortung |
|---|---------|---------------|
| 1 | auth-service | JWT, RBAC, Users, Tenants, Sessions |
| 2 | inventory-service | Equipment, Kategorien, Preise, Labels |
| 3 | project-service | Projekte, Packlisten, Reservierungen |
| 4 | scanner-service | QR/Barcode/RFID, Check-In/Out, Offline-Sync |
| 5 | warehouse-service | Lagerplaetze, Warenbewegung, Inventur |
| 6 | invoice-service | Rechnungen, Mahnwesen, DATEV |
| 7 | document-service | PDF, Templates, Vertraege, OCR |
| 8 | crew-service | Personal, Freelancer, Zeiterfassung, CalDAV |
| 9 | federation-service | mTLS P2P, Equipment-Sharing, Sub-Rental |
| 10 | maintenance-service | Wartung, E-Check/DGUV V3 |
| 11 | transport-service | Fahrzeuge, Touren, Routen |
| 12 | insurance-service | Policen, Schaeden, Risiko-Analyse |
| 13 | workflow-service | No-Code Workflows, Event-Trigger |
| 14 | ai-service | Multi-Provider AI, Anonymisierung, Prognosen |
| 15 | notification-service | In-App, Push, E-Mail |
| 16 | reporting-service | CQRS Read-Side, KPIs, Dashboards |
| 17 | audit-service | GoBD Audit-Trail, immutables Log |
| 18 | expense-service | Ausgaben, Belege, Kostenrechnung |

---

## Dokumentation

| Dokument | Inhalt |
|----------|--------|
| [docs/INSTALLATION.md](docs/INSTALLATION.md) | Detaillierte Installationsanleitung |
| [docs/USER_GUIDE.md](docs/USER_GUIDE.md) | Benutzerhandbuch |
| [docs/API.md](docs/API.md) | REST API Dokumentation |
| [docs/FEDERATION.md](docs/FEDERATION.md) | Peer-to-Peer Sharing Setup |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Production-Deployment |
| [docs/DATEV_EXPORT.md](docs/DATEV_EXPORT.md) | DATEV-Integration |
| [docs/ADMIN_GUIDE.md](docs/ADMIN_GUIDE.md) | Administrations- & Wartungshandbuch |

---

## Contributing

1. **Fork** das Repository
2. **Feature-Branch erstellen**: `git checkout -b feature/dein-feature`
3. **Commit**: `git commit -am 'Add some feature'`
4. **Push**: `git push origin feature/dein-feature`
5. **Pull Request erstellen**

**Bug-Reports & Feature-Requests:** [GitHub Issues](https://github.com/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware/issues)

---

## Warum RentFlow statt Rentman/Easyjob?

| Feature | RentFlow | Rentman | Easyjob |
|---------|----------|---------|---------|
| **Self-Hosted** | Ja | Nein (Cloud) | Nein (Cloud) |
| **Lizenzkosten** | Kostenlos | 200-500 EUR/Monat | 150-400 EUR/Monat |
| **Vendor Lock-in** | Nein (offene API) | Ja | Ja |
| **DATEV-Export** | Ja | Nein | Nein |
| **Federation/Sharing** | P2P, offen | Nein | Nein |
| **DSGVO** | Volle Kontrolle | Abhaengig vom Provider | Abhaengig vom Provider |

---

## Lizenz

[LIZENZ IST NOCH FESTZULEGEN -- TBD]

---

## Support & Community

- **Fragen & Diskussionen**: [GitHub Discussions](https://github.com/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware/discussions)
- **Bug-Reports**: [GitHub Issues](https://github.com/jeckersberger/Lagerverwaltung-und-Rechnungsbearbeitungssoftware/issues)
- **Dokumentation**: [docs/](docs/)
