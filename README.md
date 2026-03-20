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
- Docker & Docker Compose
- Mindestens 2 GB RAM und 10 GB Speicher

### Installation in 5 Schritten

#### 1. Repository klonen
```bash
git clone https://github.com/rentflow/rentflow.git
cd rentflow
```

#### 2. Environment-Variablen setzen
```bash
cp .env.example .env
# Passe folgende Werte in .env an:
# - DATABASE_PASSWORD (sichere Passphrase)
# - APP_URL (deine Domain oder IP)
```

#### 3. Docker Compose starten
```bash
docker compose up -d
```

#### 4. Datenbank initialisieren
```bash
docker compose exec api go run cmd/migrate/main.go
```

#### 5. Im Browser öffnen
Navigiere zu `https://localhost` (oder deine konfigurierte Domain).

**Standardanmeldedaten:**
- Benutzername: `admin@rentflow.local`
- Passwort: `[wird beim Setup gesetzt]`

> ⚠️ Ändere das Admin-Passwort nach der ersten Anmeldung!

---

## 🏗️ Tech-Stack

| Komponente | Technologie | Hinweis |
|-----------|-----------|--------|
| **Backend** | Go (REST API) | Modular, fast, gering RAM-Verbrauch |
| **Frontend** | React + TypeScript | Modern, responsive, offline-ready |
| **Datenbank** | PostgreSQL 14+ | ACID-Compliance, JSON-Support |
| **Deployment** | Docker Compose | Single-File-Setup, Versionskontrolle |
| **Reverse Proxy** | Caddy | Auto-HTTPS, Zero-Config |
| **Scanner** | Barcode/QR (USB/Mobile) | Open-Source Decoder |

---

## 📐 Systemarchitektur

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                          │
│  ┌──────────────────┐  ┌──────────────────┐             │
│  │ Web UI (React)   │  │ Mobile Scanner   │             │
│  │ (localhost:3000) │  │ (Offline-Mode)   │             │
│  └────────┬─────────┘  └─────────┬────────┘             │
└───────────┼──────────────────────┼──────────────────────┘
            │                      │
     ┌──────┴──────────────────────┴──────┐
     │  Caddy Reverse Proxy (Port 443)    │
     │  - HTTPS Termination               │
     │  - Auto-Renewal (Let's Encrypt)    │
     └──────────────┬─────────────────────┘
                    │
        ┌───────────┴──────────┐
        │  REST API (Go)       │
        │  - /api/v1/...       │
        │  - WebSocket (WS)    │
        │  (localhost:8080)    │
        └───────────┬──────────┘
                    │
        ┌───────────┴──────────┐
        │   PostgreSQL 14+     │
        │   - equipment        │
        │   - projects         │
        │   - invoices         │
        │   (localhost:5432)   │
        └──────────────────────┘

Federation API
┌──────────────────────────────────┐
│ P2P Equipment-Sharing Protocol    │
│ (optional, externe Firmen)       │
└──────────────────────────────────┘
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
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Docker & Docker Compose (optional)

### Lokale Entwicklung starten

#### 1. Dependencies installieren
```bash
# Backend
cd backend
go mod download

# Frontend
cd ../frontend
npm install
```

#### 2. PostgreSQL lokal starten
```bash
docker run -d \
  --name rentflow-db \
  -e POSTGRES_PASSWORD=dev \
  -e POSTGRES_DB=rentflow \
  -p 5432:5432 \
  postgres:14
```

#### 3. Migrations durchführen
```bash
cd backend
go run cmd/migrate/main.go
```

#### 4. Backend starten
```bash
cd backend
go run cmd/api/main.go
# Läuft auf http://localhost:8080
```

#### 5. Frontend starten (in neuem Terminal)
```bash
cd frontend
npm start
# Läuft auf http://localhost:3000
```

### Testing
```bash
# Backend Tests
cd backend
go test ./...

# Frontend Tests
cd frontend
npm test
```

---

## 🗺️ Roadmap

### Phase 1 – Grundlagen (Q2 2024) ✅
- [x] Equipment-Verwaltung
- [x] QR-Scanner (USB)
- [x] Einfache Rechnungserstellung
- [x] PostgreSQL Integration

### Phase 2 – Professionalisierung (Q3 2024) 🔄
- [ ] DATEV-Export
- [ ] Rollen & Berechtigungen
- [ ] Mobile Scanner-App
- [ ] Teilrechnungen & Gutschriften
- [ ] Offline-Modus

### Phase 3 – Federation (Q4 2024)
- [ ] Federation API (Peer-to-Peer)
- [ ] Equipment-Sharing zwischen Firmen
- [ ] Automatische Synchronisation
- [ ] Vertragsmanagement

### Phase 4 – Enterprise (2025+)
- [ ] Multi-Tenant Option (optional)
- [ ] Advanced Reporting & Analytics
- [ ] Mobile App (iOS/Android)
- [ ] Integrationen (Shopify, Eventbrite, etc.)
- [ ] Automatisiertes Mahnwesen

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
- PostgreSQL Team
- Open-Source Projekten (Caddy, Docker, etc.)
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
