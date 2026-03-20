# MyRMS — Projektplan & Architektur-Übersicht

**Stand:** 20. März 2026
**Version:** 1.0
**Branch:** `claude/event-tech-inventory-system-H42yD`

---

## 1. Projektvision

MyRMS (My Rental Management System) ist eine selbst-gehostete, Open-Source-Lösung für Verleihunternehmen in der Veranstaltungstechnik. Sie ersetzt Excel, Word und WhatsApp durch ein integriertes System für Lagerverwaltung, Projektplanung, Rechnungsstellung und Partner-Zusammenarbeit.

### Zielgruppe
- Veranstaltungstechnik-Firmen mit 1-30 Mitarbeitern im DACH-Raum
- Unternehmen die Self-Hosting bevorzugen (eigene Daten, keine SaaS-Abhängigkeit)
- Firmen die regelmäßig Equipment untereinander verleihen (Federation)

### Abgrenzung zu Wettbewerbern

| Aspekt | Rentman | Easyjob | MyRMS |
|--------|---------|---------|-------|
| Hosting | Cloud-only | On-Premise (Windows) | Self-hosted Docker (jedes OS) |
| Mobile | Eingeschränkt | Sehr eingeschränkt | Full Web-App (responsive + PWA) |
| Sub-Rental | Nur Rentman↔Rentman | Manuell + Assistent | Federation API (offenes Protokoll) |
| Preis | $19-25/User/Monat | ~4.000€ + Support | Self-hosted, keine Lizenzkosten |
| Buchhaltung | QuickBooks/Xero | Eigenes Modul | DATEV/WISO-Export |
| Scanner | QR/Barcode/RFID | Barcode + App | USB, Android-Handheld, Kamera, RFID |
| Offline | Nein | Ja (Desktop) | PWA mit Offline-Sync |
| KI-Integration | Nein | Nein | Multi-Provider AI (Claude, OpenAI, Ollama, ...) |
| E-Check/DGUV V3 | Nein | Nein | Integriert mit Messgeräte-Import |

---

## 2. Tech-Stack

| Schicht | Technologie | Begründung |
|---------|-------------|------------|
| **Backend** | Go 1.22+ | Performance, Single-Binary, starke Concurrency, einfaches Deployment |
| **Frontend** | TypeScript + React 18 + Sass/SCSS | Großes Ökosystem, TypeScript-First, PWA-fähig |
| **Datenbank** | PostgreSQL 16 | JSON-Support, Row-Level Security, robuste Erweiterbarkeit |
| **Deployment** | Docker / Docker Compose | Reproduzierbar, Self-Hosting, einfache Updates |
| **Reverse Proxy** | Caddy (im Docker-Stack) | Automatisches HTTPS (Let's Encrypt), Zero-Config |
| **i18n** | DE/EN von Beginn an | react-i18next (Frontend), Go embed (Backend) |
| **Scanner** | USB-Barcode, Zebra Android-Handheld, Handy-Kamera, RFID | Alle gängigen Szenarien abgedeckt |
| **PDF-Erzeugung** | Go: go-pdf / chromedp | Rechnungen, Verträge, Prüfprotokolle, Labels |
| **E-Mail** | Go: net/smtp + gomail | Rechnungsversand, Mahnungen, Benachrichtigungen |
| **KI** | Multi-Provider (Claude, OpenAI, Gemini, Mistral, Ollama) | Flexibilität, kein Vendor Lock-in |
| **Echtzeit** | WebSocket (gorilla/websocket) | Live-Updates Dashboard, Scanner-Feedback, Notifications |
| **Kalender-Sync** | CalDAV / ICS-Export | Crew-Kalender kompatibel mit iPhone, Android, Outlook |
| **Backup** | pg_dump + AES-256 Verschlüsselung | Automatisch, Multi-Destination |

---

## 3. System-Architektur

```
┌───────────────────────────────────────────────────────────┐
│                        Clients                             │
│  Desktop-Browser │ Tablet │ Handy (PWA) │ Zebra TC21       │
│  Kunden-Portal   │ Freelancer-App                          │
└──────────────────────────┬────────────────────────────────┘
                           │ HTTPS (Caddy → Let's Encrypt)
┌──────────────────────────▼────────────────────────────────┐
│                    Go Backend (API)                         │
│                                                            │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐ │
│  │ Auth/Rollen  │ │ Inventar &  │ │ Projekte &           │ │
│  │ (JWT+RBAC)   │ │ Lagerorte   │ │ Packlisten           │ │
│  └─────────────┘ └─────────────┘ └──────────────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐ │
│  │ Rechnungen & │ │ Scanner &   │ │ Federation-API       │ │
│  │ Buchhaltung  │ │ RFID        │ │ (mTLS P2P)           │ │
│  └─────────────┘ └─────────────┘ └──────────────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐ │
│  │ Wartung &    │ │ Transport & │ │ Verträge &           │ │
│  │ E-Check      │ │ Logistik    │ │ Versicherung         │ │
│  └─────────────┘ └─────────────┘ └──────────────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐ │
│  │ Crew &       │ │ Workflow-   │ │ AI-Services          │ │
│  │ Personal     │ │ Engine      │ │ (Multi-Provider)     │ │
│  └─────────────┘ └─────────────┘ └──────────────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐ │
│  │ Flightcase   │ │ Notification│ │ Audit-Trail          │ │
│  │ Management   │ │ Center      │ │ (GoBD)               │ │
│  └─────────────┘ └─────────────┘ └──────────────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌──────────────────────┐ │
│  │ Dokument-    │ │ E-Mail      │ │ Backup &             │ │
│  │ Engine       │ │ System      │ │ Recovery             │ │
│  └─────────────┘ └─────────────┘ └──────────────────────┘ │
└──────────────────────────┬────────────────────────────────┘
                           │
          ┌────────────────▼────────────────┐
          │        PostgreSQL 16 DB         │
          │  (Row-Level Security pro Firma) │
          └─────────────────────────────────┘

Firma A ◄──── Federation API (mTLS REST) ────► Firma B
```

---

## 4. Docker Compose Stack

```yaml
services:
  app:
    # Go Backend + eingebettetes React-Frontend (Single Binary)
    image: myrms/myrms:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://myrms:${DB_PASSWORD}@db:5432/myrms?sslmode=disable
      - DOMAIN=${DOMAIN}
      - ADMIN_EMAIL=${ADMIN_EMAIL}
      - SECRET_KEY=${SECRET_KEY}
      - SMTP_HOST=${SMTP_HOST:-}
      - SMTP_PORT=${SMTP_PORT:-587}
      - SMTP_USER=${SMTP_USER:-}
      - SMTP_PASSWORD=${SMTP_PASSWORD:-}
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped
    volumes:
      - uploads:/app/uploads
      - backups:/app/backups

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=myrms
      - POSTGRES_USER=myrms
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myrms"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
    restart: unless-stopped

volumes:
  pgdata:
  uploads:
  backups:
  caddy_data:
```

**Minimal `.env`:**
```env
DOMAIN=myrms.example.com
DB_PASSWORD=sicheres-passwort-hier
ADMIN_EMAIL=marco@example.com
SECRET_KEY=generierter-key
```

---

## 5. Rollen & Berechtigungsmodell

| Rolle | Dashboard | Projekte | Lager/Scan | Preise | Rechnungen | Crew | Federation | Admin |
|-------|-----------|----------|------------|--------|------------|------|------------|-------|
| **Admin** | Voll | Voll | Voll | Ja | Voll | Voll | Voll | Ja |
| **Projektleiter** | Voll | Voll | Lesen | Ja | Freigabe | Planen | Anfragen | Nein |
| **Lager** | Lager-Widget | Packlisten | Voll | Nein | Nein | Nein | In/Out | Nein |
| **Buchhaltung** | Finanz-Widget | Lesen | Nein | Ja | Voll | Nein | Rechnungen | Nein |
| **Freelancer** | Eigene Jobs | Eigene | Scan | Nein | Nein | Eigene | Nein | Nein |
| **Kunden-Portal** | Nein | Eigene | Nein | Eigene | Download | Nein | Nein | Nein |
| **Benutzerdefiniert** | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Konfigurierbar | Nein |

---

## 6. Modulübersicht

### Basis-Module (Phase 1 — MVP)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| AUTH | Authentifizierung & Rollen | JWT, RBAC, 6 Rollen, Einladungslinks | Admin |
| DASH | Dashboard | 30-Sekunden-Überblick, rollenbasiert | Marco |
| INV | Inventar & Assets | Equipment-Verwaltung, Kategorien, Seriennummern | Lisa, Marco |
| PROJ | Projekte & Packlisten | Projekt-Lifecycle, Equipment-Zuordnung, Doppelbuchungs-Check | Marco, Lisa |
| SCAN | Scanner & Check-In/Out | QR, Barcode, RFID, Kamera, Zebra TC21, Offline-fähig | Lisa |
| A1 | Artikelnummern & QR-Labels | Interne Nummern, ZPL-Labels, Metrawatt-Kompatibilität | Lisa, Admin |

### Business-Module (Phase 2)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| L1 | Flexible Lagerorte | Hierarchisch, Emoji-Icons, Bulk-Move | Lisa |
| L2 | Preiskalkulations-Engine | Staffelpreise, Mengenrabatte, Saisonzuschläge, Bundles | Marco |
| DOC | Dokument-Engine | Angebote, Rechnungen, Lieferscheine, Layout-Editor | Thomas, Marco |
| INV-FIN | Rechnungen & Buchhaltung | Angebot→Rechnung, Teilrechnungen, Gutschriften, Mahnwesen | Thomas |
| BANK | Bank-Integration | DATEV/WISO-Export, Kontoauszug-Import, OP-Abgleich | Thomas |
| EMAIL | E-Mail-System | Rechnungs-/Mahnversand, Templates, Anhänge | Thomas |
| FED | Federation | mTLS P2P, Equipment-Sharing, Sub-Rental, Partner-Preise | Stefan/SoundPro |
| N8 | Audit-Trail | GoBD-konform, unveränderliches Log, jede Entität | Thomas, Admin |

### Erweiterte Module (Phase 3)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| N1 | Crew & Personalplanung | Freelancer-DB, Schichtplanung, Zeiterfassung, CalDAV-Sync | Marco, Kevin |
| N5 | Flightcase-Management | 3-Tier-Validierung, RFID-Scan, Typ-basierte Zubehör-Zuordnung | Lisa |
| N2 | E-Check / DGUV V3 | Prüfdoku, Messgeräte-Import (IZYTRON.IQ), Gerätesperre | Lisa, Admin |
| N3 | Verbrauchsmaterial | Mengenbasiert, Mindestbestand, Nachbestellung, Lieferanten | Lisa |
| J1 | Vertragsmanagement | Templates, Platzhalter, Versionierung, digitale Signatur | Marco |
| J2 | Wartung & Predictive Maintenance | Wartungspläne, Checklisten, Erinnerungen, Foto-Doku | Lisa |
| J3 | Transport & Logistik | Fahrzeuge, Fahrer, Tourenplanung, Kapazitätsprüfung | Marco |
| K1 | Workflow-Engine | No-Code WENN→DANN, Event/CRON/Manuelle Trigger | Marco, Admin |
| K2 | Versicherungsmanagement | Policen, Deckungsprüfung, Ablauf-Warnungen | Marco |
| K3 | Schadenmanagement | 8-Schritte-Prozess, KVA, Fotos, Versicherungsmeldung | Lisa, Marco |
| N6 | Kalender & Timeline | Gantt-Projekte, Equipment-Verfügbarkeit, Crew-Kalender | Marco |
| N7 | Benachrichtigungs-Center | In-App Bell, Push, E-Mail, pro-User konfigurierbar | Alle |

### Portal & Erweiterungen (Phase 4)

| ID | Modul | Beschreibung | Persona |
|----|-------|-------------|---------|
| N4 | Kunden-Portal | Self-Service: Angebote, Rechnungen, Verfügbarkeit, Signatur | Kunden |
| L3 | Online-Buchungsportal | Öffentlicher Katalog, Warenkorb, Registrierung | Kunden |
| L4 | Backup & Disaster Recovery | Auto-Backup, Verschlüsselung, Restore-Test | Admin |
| L5 | Nachhaltigkeits-Reporting | CO2-Tracking, Energieverbrauch, ESG-Reports | Marco |
| D2 | Label-Designer | Zebra ZPL, visueller Editor, Artikelnummer+QR | Lisa |

### KI-Module (Phase 5)

| ID | Modul | Beschreibung |
|----|-------|-------------|
| I1-I3 | Multi-KI-Provider | 6 Adapter, Task-Routing, Fallback, Usage-Tracking |
| I6 | KI-Anonymisierung | Cloud-Pflicht-Anonymisierung, 7 Regex-Patterns |
| I9 | Smart Asset Creator | Hersteller→Auto-Fill, Foto-Erkennung, Bulk |
| I10 | KI-Lernsystem | Feedback, Few-Shot, Prompt-Versioning, A/B |
| I11 | KI-Transportkosten | Prognose, TCO, Budget-Forecast, Anomalien |
| I12 | KI-Versicherungsanalyse | Risiko-Bewertung, Deckungslücken, Claim-Assistent |
| I13 | KI-Bestandsoptimierung | Nachfrage-Prognose, Lagermengen, Nachbestellung |
| I14 | KI-Mahnwesen | Zahlungswahrscheinlichkeit, Mahnstufen-Optimierung |
| I15 | KI-Reporting | Anomalie-Erkennung, Trend-Vorhersage, NLP-Summaries |
| I16 | KI-Workflow-Optimierung | Vorschläge, Engpass-Erkennung, Trigger-Optimierung |
| I17 | KI-Packoptimierung | Packsequenz, Gewichtsverteilung, Fehlteile-Vorhersage |

---

## 7. Git-Workflow

```
main (stable releases)
 ├── develop (aktive Entwicklung)
 │    ├── feature/* (pro Modul/Feature)
 │    └── fix/* (Bugfixes)
 └── release/vX.Y.Z (Release-Vorbereitung)
```

### Commit-Konventionen
```
<type>(<scope>): <kurze Beschreibung>

Types: feat, fix, docs, refactor, test, chore, perf
Scopes: auth, inventory, scanner, invoicing, federation, crew, echeck, ...
```

### Versioning (SemVer)
- `MAJOR` = Breaking Changes / große neue Features
- `MINOR` = Neue Features, abwärtskompatibel
- `PATCH` = Bugfixes, kleine Verbesserungen

---

## 8. Verknüpfte Dokumente

| Datei | Inhalt |
|-------|--------|
| [01-DATABASE-SCHEMA.md](./01-DATABASE-SCHEMA.md) | Komplettes PostgreSQL-Schema aller Module |
| [02-BACKEND-SERVICES.md](./02-BACKEND-SERVICES.md) | Go-Packages, Services, API-Endpunkte |
| [03-FRONTEND-COMPONENTS.md](./03-FRONTEND-COMPONENTS.md) | React-Komponenten, Routing, State |
| [04-MODULE-SPECS.md](./04-MODULE-SPECS.md) | Detailspezifikationen aller Module |
| [05-IMPLEMENTATION-PHASES.md](./05-IMPLEMENTATION-PHASES.md) | Phasenplan, Prioritäten, Abhängigkeiten |
