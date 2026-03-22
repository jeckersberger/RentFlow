# MVP Phasenplan – Event Tech Lagerverwaltungs- & Rechnungssoftware

**Dokumentversion:** 1.0
**Gültig ab:** März 2026
**Zielgruppe:** Development Team, Projektmanagement, Stakeholder

---

## Executive Summary

Dieser Phasenplan beschreibt die Aufteilung der gesamten Softwareentwicklung in 4 fokussierte MVP-Phasen mit jeweils klarem Business-Value. Der Ansatz ermöglicht frühes Feedback, priorisierte Feature-Delivery und handhabbare Sprints.

**Gesamt-Timeline:** ~6 Monate (mit 1 Team à 3–4 Entwickler)
**Go-Live nach Phase 1:** 6–8 Wochen (internes Dogfooding möglich)
**Go-Live nach Phase 3:** 4–5 Monate (Produktiv-Release für Kunden)

---

## Überblick aller Phasen

```
Phase 1 (Wo 1-8)   Phase 2 (Wo 9-16)   Phase 3 (Wo 17-24)   Phase 4 (Wo 25-26)
Fundament+Lager    Projekte+Scanner    Rechnungen+DATEV    Federation+Sub-Rental
│                  │                   │                   │
└─Auth             └─Packlisten        └─Angebote         └─Federation-API
  Equipment          Zeiterfassung       Rechnungen          Sub-Rental
  Lagerplätze        Freelancer          Mahnwesen           Partner-Netzwerk
  Basis-Dashboard    Projekt-Dashboard   DATEV-Export
                                        PDF-Generierung
```

---

# PHASE 1: Fundament + Basis-Lagerverwaltung

## 1.1 Ziel

Nach Phase 1 ist die Grundarchitektur einsatzfähig. **Marco** kann den aktuellen Equipment-Bestand nach Lagerplatz sehen. **Lisa** kann neue Equipment erfassen und in die Lagerplatz-Hierarchie einordnen. Das System läuft in Docker, ist authentifiziert und bietet eine responsive Web-UI für die Kernfunktionen.

**Business-Wert:** Stabiles Fundament, erste echte Datenverwaltung, schnelle Amortisation der Grundentwicklung.

## 1.2 Features

| Feature | Priorität | Beschreibung |
|---------|-----------|-------------|
| **Docker-Setup & Deployment** | Must | Docker Compose (backend, frontend, postgres, nginx). Health-Checks, Restart-Policy. Production-ready Image-Tagging. |
| **User-Authentication (JWT)** | Must | Login/Logout, Session-Management, JWT-Tokens mit Refresh. Basis-Rollenmodell (Admin, Lager, Projektleiter, Buchhaltung, Freelancer). |
| **Rollen & Permissions** | Must | Role-Based Access Control (RBAC). Endpunkt-Schutz basierend auf Rolle. UI-Elemente verstecken je nach Rolle. |
| **Equipment-CRUD** | Must | Equipment anlegen, bearbeiten, löschen. Felder: Name, Typ, Seriennummer, Wert, Status (verfügbar, defekt, verkauft). |
| **Lagerplatz-Hierarchie** | Must | Lagerplätze (max. 3 Ebenen: Halle > Regal > Fach). Equipment zu Lagerplätzen zuordnen. UI mit Tree-View. |
| **Bestand-Übersicht Dashboard** | Must | Tabellenansicht aller Equipment mit Filter (Typ, Status, Lagerplatz). Summen (Anzahl, Gesamtwert). Sortierbar. |
| **Equipment-Detailseite** | Should | Foto hochladen. Wartungshistorie (Text-Notizen). Zuletzt bearbeitet (Wer, Wann). |
| **Datensicherung (Backup)** | Should | Manuelle DB-Backups. Docker Volume Persistence. Restore-Anleitung in Docs. |
| **Basis-Logging** | Could | Application-Logs (Request/Response, Fehler). Strukturiertes Logging (JSON). |
| **Dark Mode** | Could | Optional: Theme Toggle in Settings. Speicherung in localStorage. |

## 1.3 Personas

- **Marco (GF):** ✓ Sieht Bestand, Gesamtwert, Auslastung (wenn vorhanden)
- **Lisa (Lager):** ✓ Erfasst Equipment, ordnet Lagerplätze, trägt Status ein
- **Admin:** ✓ Systemsetup, User-Verwaltung, Backup

## 1.4 Datenbank-Tabellen

```sql
-- Benutzer & Auth
CREATE TABLE users (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  role VARCHAR(50) NOT NULL,  -- admin, lager, projektleiter, buchhaltung, freelancer
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Lagerplätze
CREATE TABLE storage_locations (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  level INT NOT NULL,  -- 1=Halle, 2=Regal, 3=Fach
  parent_id UUID REFERENCES storage_locations(id) ON DELETE SET NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Equipment
CREATE TABLE equipment (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  type VARCHAR(100) NOT NULL,  -- z.B. "Scheinwerfer", "Mischpult"
  serial_number VARCHAR(100) UNIQUE,
  status VARCHAR(50) DEFAULT 'verfügbar',  -- verfügbar, defekt, verkauft, archiviert
  value_eur DECIMAL(10, 2),
  purchase_date DATE,
  last_location_id UUID REFERENCES storage_locations(id) ON DELETE SET NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Equipment-Foto
CREATE TABLE equipment_images (
  id UUID PRIMARY KEY,
  equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  file_path VARCHAR(500),
  created_at TIMESTAMP DEFAULT NOW()
);

-- Wartungshistorie
CREATE TABLE maintenance_log (
  id UUID PRIMARY KEY,
  equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  note TEXT,
  status VARCHAR(50),  -- z.B. "inspiziert", "repariert"
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW()
);
```

## 1.5 API-Endpunkte

### Auth
```
POST   /api/v1/auth/login              (email, password) -> JWT
POST   /api/v1/auth/refresh            () -> new JWT
POST   /api/v1/auth/logout             () -> 200 OK
```

### Users (Admin only)
```
GET    /api/v1/users                   () -> [User]
POST   /api/v1/users                   (userData) -> User
GET    /api/v1/users/:id               () -> User
PUT    /api/v1/users/:id               (userData) -> User
DELETE /api/v1/users/:id               () -> 204
```

### Storage Locations
```
GET    /api/v1/storage-locations       () -> [StorageLocation] (hierarchisch nested)
POST   /api/v1/storage-locations       (name, level, parent_id) -> StorageLocation
GET    /api/v1/storage-locations/:id   () -> StorageLocation
PUT    /api/v1/storage-locations/:id   (name, parent_id) -> StorageLocation
DELETE /api/v1/storage-locations/:id   () -> 204
```

### Equipment
```
GET    /api/v1/equipment               (filters: type, status, location_id, search) -> [Equipment]
POST   /api/v1/equipment               (equipmentData) -> Equipment
GET    /api/v1/equipment/:id           () -> Equipment (mit Fotos + Wartungshistorie)
PUT    /api/v1/equipment/:id           (equipmentData) -> Equipment
DELETE /api/v1/equipment/:id           () -> 204
POST   /api/v1/equipment/:id/move      (location_id) -> Equipment
```

### Equipment Images
```
POST   /api/v1/equipment/:id/images    (file) -> Image
DELETE /api/v1/equipment/:id/images/:imageId () -> 204
```

### Maintenance
```
POST   /api/v1/equipment/:id/maintenance (note, status) -> MaintenanceLog
```

### Dashboard
```
GET    /api/v1/dashboard/summary       () -> { totalEquipment, totalValue, byStatus, byType }
```

## 1.6 Frontend-Seiten & Komponenten

### Pages
- **Login Page** (`/login`)
  - Email, Password Input
  - Error-Message Anzeige
  - "Remember Me" (Optional)

- **Dashboard** (`/dashboard`)
  - Welcome-Card (je nach Rolle)
  - Bestand-Widget (Anzahl Equipment, Gesamtwert)
  - Status-Übersicht (Verfügbar/Defekt/Verkauft)
  - Shortcuts zu häufigen Aktionen

- **Equipment-Übersicht** (`/equipment`)
  - Suchbar, filterbar (Typ, Status, Lagerplatz)
  - Tabelle mit Spalten: Name, Typ, Seriennummer, Lagerplatz, Status, Wert
  - Pagination (50 items/Seite)
  - "Neues Equipment anlegen" Button

- **Equipment-Detail** (`/equipment/:id`)
  - Allgemeine Infos (Name, Typ, Seriennummer, Wert, Status)
  - Lagerplatz (Hierarchie anzeigen, änderbar)
  - Foto-Upload / Galerieansicht
  - Wartungshistorie (Tabelle mit Datum, Note, User)
  - Edit/Delete Buttons
  - Bewegungshistorie (später: Equipment-Historie)

- **Lagerplatz-Verwaltung** (`/storage-locations`)
  - Tree-View der Hierarchie
  - Zum Expandieren klickbar
  - Equipment pro Lagerplatz anzeigen (Collapsible)
  - "Neuer Lagerplatz" Modal
  - Drag & Drop zum Verschieben (Optional, Phase 2)

- **Einstellungen** (`/settings`)
  - Benutzerprofil (Name, Email, Passwort ändern)
  - Sprache (DE/EN) - Placeholder für i18n
  - Theme (Dark/Light) - Optional

- **Admin-Panel** (`/admin`)
  - Benutzer-Liste (Tabelle: Email, Rolle, Status, Aktionen)
  - "Neuer Benutzer" Modal
  - System-Info (DB-Size, Last Backup, Version)

### Komponenten (Wiederverwendbar)
- `Layout` (Header, Sidebar, Footer)
- `Navigation` (Menu je nach Rolle)
- `Table` (Sortierbar, Filterbar, Pagination)
- `Modal` (Generisch für Create/Edit)
- `Button` (Primary, Secondary, Danger)
- `Input` / `Select` / `TextArea`
- `Alert` (Success, Error, Warning, Info)
- `TreeView` (für Lagerplatz-Hierarchie)
- `Card` (für Dashboard)
- `Badge` (für Status, Rolle)

## 1.7 Geschätzter Aufwand

| Bereich | Tage | Bemerkung |
|---------|------|-----------|
| **Backend: Docker + Auth + RBAC** | 8 | Setup, JWT, Role-Middleware |
| **Backend: Equipment CRUD + Tests** | 6 | API-Endpunkte, Validierung, Unit-Tests |
| **Backend: Storage-Locations + Hierarchie** | 5 | Tree-Query, Validierung, Tests |
| **Frontend: Auth + Layout** | 6 | Login, Navigation, Theme-Setup |
| **Frontend: Equipment-Seiten** | 10 | CRUD-Seiten, Tabelle, Detail-Seite |
| **Frontend: Storage-Locations UI** | 6 | Tree-View, Modal, Responsive Design |
| **Frontend: Dashboard** | 4 | Widgets, API-Integration, Responsive |
| **DB + Migrations** | 3 | Schema, Index, Seeds |
| **Testing (Manual + E2E-Basis)** | 4 | Kritische Flows testen |
| **Dokumentation + DevOps** | 4 | README, API-Docs, Docker-Docs |
| **Puffer (15%)** | 8 | Unvorhergesehenes |
| **GESAMT** | **64 Tage** | ~8–10 Wochen (3–4 Entwickler) |

## 1.8 Definition of Done

✓ Alle Must-Features implementiert und getestet
✓ Docker-Setup läuft lokal und remote
✓ API-Tests (mindestens 80% Coverage für kritische Flows)
✓ Manuelles Testing auf Chrome, Firefox, Safari (Desktop)
✓ Responsive Design (Desktop 1920px, Tablet 768px, Mobile 375px)
✓ Auth funktioniert, Rollen-Restrictions greifen
✓ Equipment-CRUD E2E getestet
✓ Lagerplatz-Hierarchie funktioniert
✓ DB-Backups funktionieren
✓ API-Dokumentation vollständig (Postman / OpenAPI)
✓ Deployment auf Test-Umgebung erfolgreich
✓ Stakeholder-Signoff (Marco, Lisa)

## 1.9 Abhängigkeiten

- **Keine** (Phase 1 ist der Anfang)

---

# PHASE 2: Projekte + Scanner + Packlisten

## 2.1 Ziel

Nach Phase 2 können Projekte mit Equipment ausgestattet werden. **Lisa** nutzt die Scanner-Integration zur schnellen Packlisten-Erfassung. **Kevin** sieht seine zugewiesenen Jobs und kann Check-In/Out durchführen. Das Dashboard zeigt Projekt-Status und Auslastung.

**Business-Wert:** Operative Effizienz (Packlisten, Zeiterfassung), Freelancer-Engagement, Datengenauigkeit durch Barcodes.

## 2.2 Features

| Feature | Priorität | Beschreibung |
|---------|-----------|-------------|
| **Projekt-CRUD** | Must | Projekt anlegen (Name, Kunde, Startdatum, Enddatum, Status). Equipment-Liste (Dropdown/Search). Freelancer-Zuweisung (Multi-Select). |
| **Packlisten-Generierung** | Must | Liste aus Projekt-Equipment. Scannable QR-Codes (per Equipment). Druck-Ready Layout. Status-Tracking (Geplant, Verpackt, Vergeben, Zurück). |
| **Scanner-Service (Standalone oder Mobile-Web)** | Must | Barcode/QR-Scan. Equipment-Lookup. Scan-In/Out-Funktionalität. Offline-Buffer (SQLite / LocalStorage). |
| **Scan-Verarbeitung** | Must | Scans mit Packlisten abgleichen. Equipment-Status updaten. Fehlerbehandlung (Unbekannter Scan). |
| **Freelancer-Zeiterfassung** | Must | Check-In (Projekt-Start). Check-Out (Projekt-Ende). Dauer berechnen. API für Scanner-App. |
| **Projekt-Dashboard** | Should | Laufende Projekte, Packlisten-Status, Team-Zusammensetzung, Timeline. |
| **Equipment-Reservierung** | Should | Equipment kann nicht gleichzeitig zwei Projekten zugewiesen sein. Konflikt-Warnung. |
| **Packlisten-Historie** | Could | Vergangene Packlisten anschauen. Scan-Logs. |
| **Mobile-Optimierung Scanner** | Could | PWA für Scanner-App (Home-Screen Icon, Offline). |

## 2.3 Personas

- **Lisa (Lager):** ✓ Erstellt Packlisten, nutzt Scanner, sieht Check-Out-Status
- **Kevin (Freelancer):** ✓ Sieht eigene Projekte, Check-In/Out, verdiente Stunden
- **Marco (GF):** ✓ Sieht Projekt-Status, Auslastung, kritische Pfade

## 2.4 Datenbank-Tabellen

```sql
-- Projekte
CREATE TABLE projects (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  customer_name VARCHAR(255),
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  status VARCHAR(50) DEFAULT 'planung',  -- planung, bereit, aktiv, abgeschlossen
  description TEXT,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Projekt-Equipment (Zuordnung)
CREATE TABLE project_equipment (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  equipment_id UUID NOT NULL REFERENCES equipment(id),
  quantity INT DEFAULT 1,
  status VARCHAR(50) DEFAULT 'geplant',  -- geplant, verpackt, vergeben, zurück
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(project_id, equipment_id)
);

-- Packlisten
CREATE TABLE packing_lists (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  generated_at TIMESTAMP DEFAULT NOW(),
  status VARCHAR(50) DEFAULT 'geplant',  -- geplant, verpackt, vergeben, zurück
  notes TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Packlisten-Positionen (Equipment im Paket)
CREATE TABLE packing_list_items (
  id UUID PRIMARY KEY,
  packing_list_id UUID NOT NULL REFERENCES packing_lists(id) ON DELETE CASCADE,
  equipment_id UUID NOT NULL REFERENCES equipment(id),
  quantity INT DEFAULT 1,
  scanned_quantity INT DEFAULT 0,
  qr_code VARCHAR(500),  -- URL oder Data-URI für QR-Code
  created_at TIMESTAMP DEFAULT NOW()
);

-- Scans
CREATE TABLE scans (
  id UUID PRIMARY KEY,
  packing_list_id UUID REFERENCES packing_lists(id) ON DELETE SET NULL,
  equipment_id UUID NOT NULL REFERENCES equipment(id),
  scan_type VARCHAR(50),  -- in, out, unknown
  scanned_by UUID REFERENCES users(id),
  scanned_at TIMESTAMP DEFAULT NOW(),
  location VARCHAR(100),
  device_id VARCHAR(100)
);

-- Projekt-Freelancer
CREATE TABLE project_freelancers (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  freelancer_id UUID NOT NULL REFERENCES users(id),
  role VARCHAR(100),  -- z.B. "Techniker", "Assistent"
  created_at TIMESTAMP DEFAULT NOW()
);

-- Zeiterfassung
CREATE TABLE timesheets (
  id UUID PRIMARY KEY,
  freelancer_id UUID NOT NULL REFERENCES users(id),
  project_id UUID REFERENCES projects(id),
  check_in_at TIMESTAMP,
  check_out_at TIMESTAMP,
  duration_minutes INT,  -- computed
  status VARCHAR(50) DEFAULT 'aktiv',  -- aktiv, abgeschlossen
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

## 2.5 API-Endpunkte

### Projekte
```
GET    /api/v1/projects                (filters: status, date-range, customer) -> [Project]
POST   /api/v1/projects                (projectData) -> Project
GET    /api/v1/projects/:id            () -> Project (mit Equipment-Liste, Freelancer)
PUT    /api/v1/projects/:id            (projectData) -> Project
DELETE /api/v1/projects/:id            () -> 204
```

### Projekt-Equipment
```
POST   /api/v1/projects/:id/equipment  (equipment_id, quantity) -> ProjectEquipment
DELETE /api/v1/projects/:id/equipment/:equipmentId () -> 204
PATCH  /api/v1/projects/:id/equipment/:equipmentId (quantity, status) -> ProjectEquipment
```

### Packlisten
```
POST   /api/v1/projects/:id/packing-lists (data) -> PackingList (generiert aus Projekt-Equipment)
GET    /api/v1/packing-lists/:id      () -> PackingList (mit Items + QR-Codes)
PATCH  /api/v1/packing-lists/:id      (status, notes) -> PackingList
GET    /api/v1/packing-lists/:id/pdf  () -> PDF-Download
```

### Scanner-Service (Standalone oder in API)
```
POST   /api/v1/scanner/scan            (barcode/qr_code, device_id) -> { equipment, status, message }
GET    /api/v1/scanner/packing-list/:id (offline-json für Scanner-App, QR-cached)
POST   /api/v1/scanner/offline-sync    (scans_array) -> { synced_count, errors }
```

### Freelancer-Projekte
```
GET    /api/v1/freelancer/projects     () -> [Project] (nur eigene Projekte)
GET    /api/v1/freelancer/packing-lists () -> [PackingList] (für aktuelle Projekte)
```

### Zeiterfassung
```
POST   /api/v1/freelancer/check-in     (project_id) -> Timesheet
POST   /api/v1/freelancer/check-out    () -> Timesheet (duration berechnet)
GET    /api/v1/freelancer/timesheets   (date-range, project_id) -> [Timesheet]
GET    /api/v1/users/:id/timesheets    (date-range) -> [Timesheet] (Admin-View)
```

## 2.6 Frontend-Seiten & Komponenten

### Pages
- **Projekte-Übersicht** (`/projects`)
  - Tabelle: Name, Kunde, Datum, Status, Anzahl Equipment, Freelancer
  - Filter: Status, Datum
  - "Neues Projekt" Button
  - Quick-Actions: Packliste, Team anschauen

- **Projekt-Detail** (`/projects/:id`)
  - Projekt-Infos (Name, Kunde, Datum, Status, Beschreibung)
  - Equipment-Liste (Tabelle: Name, Typ, Menge, Status)
  - Equipment hinzufügen (Modal mit Search/Dropdown)
  - Freelancer-Liste (Namen, Rolle, Status)
  - Freelancer hinzufügen (Modal mit User-Search)
  - Actions: Packliste generieren, Status ändern, Projekt archivieren

- **Packlisten** (`/projects/:id/packing-lists`)
  - Liste aller Packlisten des Projekts
  - Generieren-Button
  - Tabelle: Datum, Status, Items-Count, Actions (View, PDF, Print, Delete)

- **Packlisten-Detailansicht** (`/packing-lists/:id`)
  - Projekt-Info (Übersicht)
  - Items-Tabelle: Equipment, Menge, Scanned-Count, Barcode/QR
  - PDF-Download / Print Button
  - Status-Buttons (Geplant > Verpackt > Vergeben > Zurück)
  - Scan-Historie (Mini-Tabelle: Equipment, Zeit, User)

- **Scanner-App** (`/scanner` oder separat hosten)
  - QR/Barcode Input (Autofocus)
  - Liveansicht: "Letzter Scan" mit Equipment-Bild
  - Status: ✓ Erfolgreich / ✗ Fehler / ? Unbekannt
  - Offline-Indicator (Sync-Status)
  - "Sync jetzt" Button
  - Geschichte (Scrollbar mit letzten Scans)
  - Mobile-optimiert (PWA)

- **Freelancer-Dashboard** (`/freelancer/dashboard`)
  - "Aktive Projekte" Widget (Projektnamen, Status)
  - "Check-In/Out" Card (großer Button wenn nicht eingecheckt)
  - "Heute geleistete Stunden" Widget
  - "Meine Packlisten" (Button zum Quick-View)
  - Zeiterfassung-Historie (Tabelle: Projekt, Check-In, Check-Out, Dauer)

- **Zeiterfassung** (`/freelancer/timesheets`)
  - Kalender-View oder Tabelle: Datum, Projekt, Dauer, Status
  - Filter: Monat, Projekt

### Komponenten
- `ProjectForm` (Create/Edit Modal)
- `ProjectEquipmentSelect` (Search + Add Equipment)
- `FreelancerSelect` (Search + Add Freelancer)
- `ScannerInput` (Barcode-Input mit Live-Update)
- `PackingListPrintable` (Druckfreundliches Layout)
- `CheckInOutButton` (Kontextuell, groß und prominent)
- `TimelineVisualization` (Projekt-Timeline mit Equipment-Status)

## 2.7 Geschätzter Aufwand

| Bereich | Tage | Bemerkung |
|---------|------|-----------|
| **Backend: Projekt-CRUD + Validierung** | 5 | Datum-Constraints, Status-Machine |
| **Backend: Projekt-Equipment Zuordnung** | 4 | Konflikterkennung, Reservierung |
| **Backend: Packlisten-Generation** | 5 | QR-Code-Generierung, PDF-Template |
| **Backend: Scanner-Service** | 8 | Barcode-Parsing, Offline-Sync, Error-Handling |
| **Backend: Zeiterfassung** | 4 | Check-In/Out, Duration-Calc, Validation |
| **Backend: Tests** | 6 | Integration-Tests für Scans, Zeiterfassung |
| **Frontend: Projekte + CRUD** | 7 | Seite, Modal, Equipment-Select |
| **Frontend: Packlisten + PDF** | 8 | Generierung, Detail-View, Print-Layout, PDF-Download |
| **Frontend: Scanner-App (Web)** | 8 | Eingabe-Handling, Offline-Storage, Sync, PWA-Setup |
| **Frontend: Freelancer-Seiten** | 6 | Dashboard, Timesheets, Check-In/Out Button |
| **Frontend: Responsive Design** | 4 | Mobile-Testing, Layout-Anpassung |
| **QA + E2E Testing** | 8 | Kritische Flows (Packlisten-Scan, Check-In) |
| **Deployment + Docs** | 4 | Update-Prozess, Scanner-App-Deploy |
| **Puffer (15%)** | 11 | |
| **GESAMT** | **88 Tage** | ~11–14 Wochen (3–4 Entwickler) |

## 2.8 Definition of Done

✓ Projekt-CRUD vollständig funktionierend
✓ Equipment-Reservierung (Konflikt-Prävention)
✓ Packlisten generierbar und druckbar
✓ Barcode/QR-Scans funktionieren lokal + mit Offline-Sync
✓ Check-In/Out funktioniert, Zeiten berechnet
✓ Scanner-App auf Tablet/Mobile getestet
✓ PWA-Features (Offline, Add-to-Home) getestet
✓ Freelancer-Seiten responsive + getestet
✓ Packlisten-PDF-Export funktioniert
✓ API-Tests für kritische Scanner-Flows
✓ E2E-Test: Projekt anlegen → Equipment hinzufügen → Packliste generieren → Scannen
✓ Stakeholder-Signoff (Lisa, Kevin, Marco)

## 2.9 Abhängigkeiten

- **Phase 1** muss vollständig sein
  - Auth, Equipment-CRUD, Lagerplätze, User-Management

---

# PHASE 3: Angebote + Rechnungen + DATEV

## 3.1 Ziel

Nach Phase 3 ist ein vollständiger Rechnungsworkflow implementiert. **Thomas** kann Angebote erstellen, diese in Rechnungen umwandeln, Mahnungen versenden und am Monatsende einen DATEV-Export durchführen. Das System unterstützt Teilrechnungen, Gutschriften und offene Posten-Tracking.

**Business-Wert:** Automatisierte Finanzverwaltung, Compliance (DATEV-Export), Geschäftsdokumentation, Mahnwesen-Effizienz.

## 3.2 Features

| Feature | Priorität | Beschreibung |
|---------|-----------|-------------|
| **Angebote (Quotes)** | Must | Template-basiert. Kunde, Equipment-Liste mit Preisen, Rabatt, MwSt. PDF-Export. Status (Entwurf, Versendet, Angenommen, Abgelehnt). |
| **Rechnungen** | Must | Aus Angebot oder manuell. Rechnungsnummer (Auto-Inkrement), Datum, Fälligkeitsdatum, Zahlungsbedingungen. Equipment-Linien mit Einzelpreis. |
| **Teilrechnungen** | Should | Mehrere Rechnungen pro Projekt. Summen-Tracking. |
| **Gutschriften (Credit Notes)** | Should | Für Retouren/Fehler. Referenz zu Original-Rechnung. |
| **Zahlungs-Tracking** | Should | Offene/Bezahlte Rechnungen. Zahlungsdatum + Betrag erfassen. |
| **Mahnwesen** | Should | Automatische Mahnungen (14, 30, 60 Tage überfällig). E-Mail-Template. Mahngebühr (optional). Status (Erste Mahnung, Zweite Mahnung, etc.). |
| **DATEV-Export** | Must | Monatlicher Export (Posting-Journal). Format: CSV o.ä. gemäß DATEV-Standard. Sachkonto-Mapping konfigurierbar. |
| **PDF-Generierung** | Must | Professionelle Angebots- & Rechnungs-PDFs mit Logo, Kundendaten, Zahlungskonten. |
| **E-Mail-Versand** | Should | Angebote, Rechnungen, Mahnungen direkt versenden. E-Mail-Template anpassbar. Tracking: Versendet, Gelesen. |
| **Belegarchivierung** | Could | Digitale Archivierung (GoBD-konform). Scan + Upload möglich. |

## 3.3 Personas

- **Thomas (Buchhaltung):** ✓ Erstellt Angebote/Rechnungen, versendet Mahnungen, DATEV-Export
- **Marco (GF):** ✓ Sieht offene Posten, Umsatzübersicht
- **Admin:** ✓ DATEV-Mapping konfigurieren, E-Mail-Server konfigurieren

## 3.4 Datenbank-Tabellen

```sql
-- Kunden (vereinfacht, könnten später erweitert werden)
CREATE TABLE customers (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  address VARCHAR(500),
  tax_id VARCHAR(50),  -- Mehrwertsteuernummer o.ä.
  payment_terms INT DEFAULT 14,  -- Tage bis Fälligkeitsdatum
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Angebote
CREATE TABLE quotes (
  id UUID PRIMARY KEY,
  quote_number VARCHAR(50) UNIQUE NOT NULL,  -- z.B. "Q-2026-0001"
  customer_id UUID NOT NULL REFERENCES customers(id),
  project_id UUID REFERENCES projects(id),
  created_by UUID NOT NULL REFERENCES users(id),
  created_date DATE NOT NULL,
  valid_until DATE NOT NULL,
  status VARCHAR(50) DEFAULT 'entwurf',  -- entwurf, versendet, angenommen, abgelehnt
  discount_eur DECIMAL(10, 2) DEFAULT 0,
  tax_rate DECIMAL(5, 2) DEFAULT 19.00,
  notes TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Angebots-Positionen
CREATE TABLE quote_items (
  id UUID PRIMARY KEY,
  quote_id UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
  equipment_id UUID REFERENCES equipment(id),
  description VARCHAR(500),  -- Falls nicht Equipment, dann frei-Text
  quantity INT NOT NULL,
  unit_price_eur DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Rechnungen
CREATE TABLE invoices (
  id UUID PRIMARY KEY,
  invoice_number VARCHAR(50) UNIQUE NOT NULL,  -- z.B. "R-2026-0001"
  quote_id UUID REFERENCES quotes(id),
  customer_id UUID NOT NULL REFERENCES customers(id),
  project_id UUID REFERENCES projects(id),
  created_by UUID NOT NULL REFERENCES users(id),
  invoice_date DATE NOT NULL,
  due_date DATE NOT NULL,
  status VARCHAR(50) DEFAULT 'offen',  -- offen, teilvollbracht, vollständig, überfällig, storniert
  subtotal_eur DECIMAL(12, 2) NOT NULL,
  discount_eur DECIMAL(10, 2) DEFAULT 0,
  tax_amount_eur DECIMAL(10, 2) NOT NULL,
  tax_rate DECIMAL(5, 2) DEFAULT 19.00,
  total_eur DECIMAL(12, 2) NOT NULL,
  paid_amount_eur DECIMAL(12, 2) DEFAULT 0,
  notes TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Rechnungs-Positionen
CREATE TABLE invoice_items (
  id UUID PRIMARY KEY,
  invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  equipment_id UUID REFERENCES equipment(id),
  description VARCHAR(500),
  quantity INT NOT NULL,
  unit_price_eur DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Zahlungen
CREATE TABLE payments (
  id UUID PRIMARY KEY,
  invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  amount_eur DECIMAL(12, 2) NOT NULL,
  payment_date DATE NOT NULL,
  payment_method VARCHAR(50),  -- bank_transfer, cash, credit_card, etc.
  reference VARCHAR(500),  -- Zahlungsreferenz
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW()
);

-- Gutschriften
CREATE TABLE credit_notes (
  id UUID PRIMARY KEY,
  credit_number VARCHAR(50) UNIQUE NOT NULL,  -- z.B. "CN-2026-0001"
  invoice_id UUID NOT NULL REFERENCES invoices(id),
  customer_id UUID NOT NULL REFERENCES customers(id),
  created_by UUID NOT NULL REFERENCES users(id),
  created_date DATE NOT NULL,
  reason VARCHAR(500),  -- z.B. "Retoure", "Fehler"
  amount_eur DECIMAL(12, 2) NOT NULL,
  status VARCHAR(50) DEFAULT 'offen',  -- offen, verrechnet, erstattet
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Mahnungen
CREATE TABLE dunning_letters (
  id UUID PRIMARY KEY,
  invoice_id UUID NOT NULL REFERENCES invoices(id),
  level INT DEFAULT 1,  -- 1, 2, 3 (Stufen)
  dunning_date DATE NOT NULL,
  reminder_email_sent BOOLEAN DEFAULT false,
  dunning_fee_eur DECIMAL(10, 2) DEFAULT 0,
  status VARCHAR(50) DEFAULT 'offen',  -- offen, bezahlt, storniert
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- E-Mail-Versandhistorie
CREATE TABLE email_history (
  id UUID PRIMARY KEY,
  recipient_email VARCHAR(255) NOT NULL,
  subject VARCHAR(500),
  body TEXT,
  document_type VARCHAR(50),  -- quote, invoice, dunning_letter
  document_id UUID,
  sent_at TIMESTAMP DEFAULT NOW(),
  status VARCHAR(50) DEFAULT 'sent',  -- sent, bounced, failed
  error_message TEXT
);

-- DATEV-Konfiguration (Simple Key-Value)
CREATE TABLE datev_config (
  key VARCHAR(100) PRIMARY KEY,
  value TEXT,
  updated_at TIMESTAMP DEFAULT NOW()
);
-- Einträge: sachkonto_equipment, sachkonto_dienstleistung, sachkonto_erlöse, etc.

-- DATEV-Exporte (Audit-Log)
CREATE TABLE datev_exports (
  id UUID PRIMARY KEY,
  export_date DATE NOT NULL,
  file_path VARCHAR(500),
  record_count INT,
  created_by UUID REFERENCES users(id),
  status VARCHAR(50),  -- success, error
  error_message TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);
```

## 3.5 API-Endpunkte

### Kunden
```
GET    /api/v1/customers                () -> [Customer]
POST   /api/v1/customers                (customerData) -> Customer
GET    /api/v1/customers/:id            () -> Customer
PUT    /api/v1/customers/:id            (customerData) -> Customer
```

### Angebote
```
GET    /api/v1/quotes                   (filters: customer, status, date-range) -> [Quote]
POST   /api/v1/quotes                   (quoteData: customer_id, items[], ...) -> Quote
GET    /api/v1/quotes/:id               () -> Quote (mit Items)
PUT    /api/v1/quotes/:id               (quoteData) -> Quote
PATCH  /api/v1/quotes/:id/status        (new_status) -> Quote
GET    /api/v1/quotes/:id/pdf           () -> PDF-Download
POST   /api/v1/quotes/:id/email         (recipient_email) -> { success, message }
POST   /api/v1/quotes/:id/to-invoice    () -> Invoice (konvertiert Angebot zu Rechnung)
DELETE /api/v1/quotes/:id               () -> 204
```

### Rechnungen
```
GET    /api/v1/invoices                 (filters: customer, status, date-range) -> [Invoice]
POST   /api/v1/invoices                 (invoiceData: customer_id, items[], ...) -> Invoice
GET    /api/v1/invoices/:id             () -> Invoice (mit Items, Payments, Dunning)
PUT    /api/v1/invoices/:id             (invoiceData) -> Invoice
GET    /api/v1/invoices/:id/pdf         () -> PDF-Download
POST   /api/v1/invoices/:id/email       (recipient_email) -> { success, message }
PATCH  /api/v1/invoices/:id/status      (new_status) -> Invoice
DELETE /api/v1/invoices/:id             () -> 204 (nur wenn Entwurf)
```

### Zahlungen
```
POST   /api/v1/invoices/:id/payments    (amount_eur, payment_date, method, reference) -> Payment
GET    /api/v1/invoices/:id/payments    () -> [Payment]
```

### Gutschriften
```
POST   /api/v1/invoices/:id/credit-note (reason, amount_eur) -> CreditNote
GET    /api/v1/credit-notes/:id         () -> CreditNote
PATCH  /api/v1/credit-notes/:id/status  (new_status) -> CreditNote
```

### Mahnwesen
```
POST   /api/v1/invoices/:id/dunning     (level) -> DunningLetter
GET    /api/v1/dunning-letters          (filters: status, date-range) -> [DunningLetter]
POST   /api/v1/dunning-letters/:id/email () -> { success, message }
```

### DATEV-Export
```
POST   /api/v1/datev/config             (sachkonto_mapping) -> { success }
GET    /api/v1/datev/exports            (date-range) -> [DatevExport]
POST   /api/v1/datev/export             (start_date, end_date) -> { export_id, download_url, record_count }
GET    /api/v1/datev/exports/:id/download () -> CSV-Download
```

### Dashboard / Übersichten
```
GET    /api/v1/dashboard/finance        () -> { outstanding_amount, revenue_30d, overdue_count, top_customers }
GET    /api/v1/invoices/summary         (date-range) -> { paid, outstanding, overdue }
```

## 3.6 Frontend-Seiten & Komponenten

### Pages
- **Kunden-Verwaltung** (`/customers`)
  - Tabelle: Name, Email, Zahlungsbedingungen, Status
  - Filter, Suchfeld
  - "Neuer Kunde" Button (Modal)
  - "Bearbeiten" / "Löschen" (mit Warnung wenn Rechnungen)

- **Angebote** (`/quotes`)
  - Tabelle: Nummer, Kunde, Datum, Gültig bis, Status, Betrag
  - Filter: Status, Kunde, Datum
  - "Neues Angebot" Button
  - Quick-Actions: PDF, E-Mail, Zu Rechnung konvertieren

- **Angebots-Detail** (`/quotes/:id`)
  - Kunde, Adressen, Datum, Gültig bis
  - Items-Tabelle: Beschreibung, Menge, Einzelpreis, Gesamt
  - Rabatt, MwSt., Summe (Netto, Brutto)
  - Status-Badge
  - Actions: PDF, E-Mail, Edit, Zu Rechnung konvertieren, Löschen
  - Notizen

- **Rechnungen** (`/invoices`)
  - Tabelle: Nummer, Kunde, Datum, Fällig, Status, Betrag, Bezahlt
  - Filter: Status, Kunde, Datum, Überfällig
  - "Neue Rechnung" Button
  - Übersicht: Summe offen, Überfällig, Gesamt Umsatz 30d
  - Quick-Actions: PDF, E-Mail, Zahlung, Gutschrift, Mahnung

- **Rechnungs-Detail** (`/invoices/:id`)
  - Ähnlich wie Angebots-Detail
  - + Zahlungen-Tabelle (Datum, Betrag, Methode)
  - + "Zahlung erfassen" Button
  - + Dunning-Status / Mahnungen-Historie
  - + Gutschriften-Link (falls vorhanden)

- **Gutschriften** (`/credit-notes`)
  - Tabelle: Nummer, Ursprungs-Rechnung, Kunde, Betrag, Status
  - Filter: Status, Datum
  - Links zu Originals

- **Mahnwesen** (`/dunning`)
  - Tabelle: Rechnung, Kunde, Level, Datum, Status
  - Filter: Status, Überfällig-Tage
  - "Mahnung versenden" Button (E-Mail)
  - Gebühren-Kalkulator (Optional)

- **DATEV-Export** (`/datev`)
  - Konfigurations-Form (Sachkonto-Mapping)
  - Tabelle vergangener Exports: Datum, Records, File, Status
  - "Export erstellen" Button (Modal mit Datum-Range)
  - Download-Link

- **Finance Dashboard** (`/finance` oder in Main-Dashboard)
  - KPIs: Ausstehend, Überfällig, Umsatz 30d, Durchschnittliche Zahlungsdauer
  - Charts: Umsatz-Timeline, Zahlungs-Status-Pie
  - Top-Kunden, Offene-Posten-Liste

### Komponenten
- `QuoteForm` / `InvoiceForm` (Create/Edit Modal)
- `ItemsTable` (Dynamic, Add/Remove Items)
- `CustomerSelect` (Search + Select)
- `PricingCalculator` (Automatische Summen-Berechnung)
- `StatusBadge` (Farblich code: offen=rot, bezahlt=grün, überfällig=orange)
- `PaymentInput` (Modal für Zahlungs-Erfassung)
- `PDFPreview` (Render-Vorschau vor Download)
- `EmailTemplateEditor` (WYSIWYG, Variablen einfügen)
- `DatevConfigForm` (Sachkonto-Mapping)

## 3.7 Geschätzter Aufwand

| Bereich | Tage | Bemerkung |
|---------|------|-----------|
| **Backend: Customer + Quote CRUD** | 5 | Quote-Generation, Validierung |
| **Backend: Invoice CRUD + Calculations** | 6 | Summen, Tax-Handling, Status-Machine |
| **Backend: Payment Tracking** | 4 | Payment-Erfassung, Reconciliation |
| **Backend: Credit Notes + Dunning** | 5 | Automatische Mahnungen, Gebühren |
| **Backend: DATEV-Export Logic** | 7 | CSV-Format, Sachkonto-Mapping, Audit-Log |
| **Backend: PDF-Generierung** | 6 | Templates, Fonts, QR-Rechnungscodes (Swiss-QR optional) |
| **Backend: E-Mail-Versand** | 4 | SMTP, Templates, Retry-Logic, Bounce-Handling |
| **Backend: Tests** | 8 | Invoice-Calculation, DATEV-Export, Payment-Reconciliation |
| **Frontend: Customer + Quote Seiten** | 7 | CRUD, Forms, PDF-Preview |
| **Frontend: Invoice Seiten** | 8 | CRUD, Payment-Input, Dunning-UI |
| **Frontend: Gutschriften + Mahnungen** | 6 | Forms, Status-Updates |
| **Frontend: Finance Dashboard** | 8 | KPI-Widgets, Charts (Chart.js/Recharts), Responsive |
| **Frontend: DATEV-Config UI** | 4 | Form, Save/Load, Validation |
| **Frontend: E-Mail-Template Editor** | 5 | WYSIWYG, Variable-Insert, Preview |
| **Frontend: Responsive Design** | 4 | Mobile, Tablet, Print-CSS |
| **QA + E2E Testing** | 10 | Quote-to-Invoice, Payment-Tracking, DATEV-Export |
| **Docs + Deployment** | 4 | API-Docs, DATEV-Format-Docs, Email-Config-Guide |
| **Puffer (15%)** | 18 | |
| **GESAMT** | **124 Tage** | ~16–20 Wochen (3–4 Entwickler) |

## 3.8 Definition of Done

✓ Quote-CRUD vollständig + PDF-Export
✓ Invoice-CRUD + Summations-Logik (Netto, Brutto, Tax)
✓ Payment-Erfassung + Reconciliation
✓ Credit Notes + Dunning-Automatik (basierend auf Fälligkeitsdatum)
✓ E-Mail-Versand (Angebote, Rechnungen, Mahnungen) funktioniert
✓ DATEV-Export (CSV) gemäß Standard korrekt
✓ PDF-Generierung professionell aussehend + korrekte Rechnungsdaten
✓ Finance-Dashboard mit KPIs + Charts
✓ API-Tests für Invoice-Calculation, DATEV-Export
✓ E2E-Test: Angebot → Rechnung → Zahlung → DATEV-Export
✓ Mahnwesen-Test: Überfällige Rechnungen erkennen + Mahnungen versenden
✓ Stakeholder-Signoff (Thomas, Marco)

## 3.9 Abhängigkeiten

- **Phase 1 & 2** müssen vollständig sein
  - Auth, Equipment, Projekte, Freelancer-Data

---

# PHASE 4: Federation + Sub-Rental

## 4.1 Ziel

Nach Phase 4 können Equipment-Partner über eine föderierte API verbunden werden. **Marco** kann bei Partnern (z.B. Stefan) Equipment anfragen. Das System prüft verfügbar Bestände, koordiniert Lieferungen und erstellt automatisch Sub-Rental-Rechnungen. Ein Partner-Netzwerk entsteht.

**Business-Wert:** Erweiterte Geschäftsmöglichkeiten (Equipment-Netzwerk), schnellere Projektakquisition, zusätzliche Einnahmequelle (Sub-Rental).

## 4.2 Features

| Feature | Priorität | Beschreibung |
|---------|-----------|-------------|
| **Federation-API (mTLS)** | Must | Peer-to-Peer Kommunikation zwischen Instanzen. Gegenseitige TLS-Zertifikate. Authentifizierung. Rate-Limiting. |
| **Partner-Verzeichnis** | Must | Partner-Namen, Kontakt, API-Endpoint, Zertifikat-Fingerprint. Status (verbunden, nicht erreichbar). |
| **Equipment-Verfügbarkeit abfragen** | Must | Query-API: GET /federation/equipment?type=Scheinwerfer&start=2026-05-01&end=2026-05-03. Response mit verfügbarem Equipment + Preisen. |
| **Sub-Rental-Anfrage** | Should | Anfrage senden (Equipment, Datum, Menge). Partner antwortet mit Angebot. Auto-Rechnung auf Angebot. |
| **Zustandsdokumentation** | Should | Equipment-Fotos bei Check-Out von Partner. Rückgabe-Inspektionen. Schäden dokumentieren + Ausgleich. |
| **Automatische Rechnungsstellung** | Should | Sub-Rental-Rechnung auto-erstellen. Basis: Partner-Preis + Markup. Fälligkeitsdatum von verbrauchtem Projekt. |
| **Lieferverfolgung** | Could | Wer hat Equipment wann abgeholt? Tracking von Sendungen. |
| **Vertragsvorlagen** | Could | Standard-Nutzungsbedingungen je Partner konfigurierbar. |

## 4.3 Personas

- **Marco (GF):** ✓ Sendet Sub-Rental-Anfragen, sieht verfügbare Equipment bei Partnern
- **Stefan (Partner):** ✓ Nimmt Anfragen entgegen, antwortet mit Verfügbarkeit, prüft Rückgaben
- **Admin:** ✓ Konfiguriert Partner-Verbindungen, Zertifikate, Verträge

## 4.4 Datenbank-Tabellen

```sql
-- Partner
CREATE TABLE partners (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  contact_email VARCHAR(255),
  api_endpoint VARCHAR(500),
  api_cert_fingerprint VARCHAR(128),
  status VARCHAR(50) DEFAULT 'pendente',  -- pending, connected, disconnected, blocked
  markup_percent DECIMAL(5, 2) DEFAULT 15.00,  -- Aufschlag auf Partner-Preis
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Sub-Rental Anfragen
CREATE TABLE sub_rental_requests (
  id UUID PRIMARY KEY,
  partner_id UUID NOT NULL REFERENCES partners(id),
  project_id UUID REFERENCES projects(id),
  request_date DATE NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  status VARCHAR(50) DEFAULT 'pending',  -- pending, quoted, accepted, rejected, completed, cancelled
  notes TEXT,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Sub-Rental Request Items
CREATE TABLE sub_rental_request_items (
  id UUID PRIMARY KEY,
  request_id UUID NOT NULL REFERENCES sub_rental_requests(id) ON DELETE CASCADE,
  equipment_type VARCHAR(100),  -- z.B. "Scheinwerfer 1000W"
  quantity INT NOT NULL,
  requested_at TIMESTAMP DEFAULT NOW()
);

-- Sub-Rental Angebote (vom Partner)
CREATE TABLE sub_rental_quotes (
  id UUID PRIMARY KEY,
  request_id UUID NOT NULL REFERENCES sub_rental_requests(id),
  partner_id UUID NOT NULL REFERENCES partners(id),
  quote_number VARCHAR(50),
  created_at TIMESTAMP DEFAULT NOW(),
  status VARCHAR(50) DEFAULT 'pending',  -- pending, accepted, rejected
  expires_at TIMESTAMP
);

-- Sub-Rental Quote Items
CREATE TABLE sub_rental_quote_items (
  id UUID PRIMARY KEY,
  quote_id UUID NOT NULL REFERENCES sub_rental_quotes(id) ON DELETE CASCADE,
  equipment_id UUID,  -- Partner-seitige Equipment-ID oder Typ-String
  equipment_name VARCHAR(255),
  quantity INT NOT NULL,
  daily_rate_eur DECIMAL(10, 2),
  available BOOLEAN DEFAULT true
);

-- Sub-Rental Bestellungen (akzeptiertes Angebot)
CREATE TABLE sub_rental_orders (
  id UUID PRIMARY KEY,
  quote_id UUID NOT NULL REFERENCES sub_rental_quotes(id),
  invoice_id UUID REFERENCES invoices(id),  -- Auto-generated Invoice
  status VARCHAR(50) DEFAULT 'pending',  -- pending, pickup, delivered, returned, closed
  pickup_date TIMESTAMP,
  delivery_date TIMESTAMP,
  return_date TIMESTAMP,
  pickup_location VARCHAR(500),
  delivery_location VARCHAR(500),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Sub-Rental Zustandsdokumentation
CREATE TABLE sub_rental_conditions (
  id UUID PRIMARY KEY,
  order_id UUID NOT NULL REFERENCES sub_rental_orders(id),
  checkpoint VARCHAR(50),  -- pickup_check, return_check
  equipment_id VARCHAR(255),  -- Partner-Equipment-ID
  condition_status VARCHAR(50),  -- ok, damaged, missing
  notes TEXT,
  photos_path VARCHAR(500),  -- JSON array of image paths
  damage_cost_eur DECIMAL(10, 2),
  documented_by UUID REFERENCES users(id),
  documented_at TIMESTAMP DEFAULT NOW()
);

-- Federation API Calls (Audit Log)
CREATE TABLE federation_audit (
  id UUID PRIMARY KEY,
  partner_id UUID REFERENCES partners(id),
  method VARCHAR(10),  -- GET, POST
  endpoint VARCHAR(500),
  request_body TEXT,
  response_status INT,
  response_body TEXT,
  error_message TEXT,
  called_at TIMESTAMP DEFAULT NOW()
);
```

## 4.5 API-Endpunkte

### Partner-Verwaltung
```
GET    /api/v1/partners                 () -> [Partner]
POST   /api/v1/partners                 (partnerData) -> Partner
GET    /api/v1/partners/:id             () -> Partner
PUT    /api/v1/partners/:id             (partnerData) -> Partner
PATCH  /api/v1/partners/:id/test        () -> { status, message } (Verbindung testen)
```

### Sub-Rental Anfragen
```
POST   /api/v1/sub-rental/request       (partner_id, items[], dates) -> SubRentalRequest
GET    /api/v1/sub-rental/requests      (filters: partner, status, date-range) -> [SubRentalRequest]
GET    /api/v1/sub-rental/requests/:id  () -> SubRentalRequest (mit Quote + Order)
PATCH  /api/v1/sub-rental/requests/:id/status (new_status) -> SubRentalRequest
```

### Sub-Rental Orders
```
POST   /api/v1/sub-rental/orders/:id/accept () -> SubRentalOrder (akzeptiert Quote, erstellt Order + Invoice)
GET    /api/v1/sub-rental/orders/:id    () -> SubRentalOrder
PATCH  /api/v1/sub-rental/orders/:id    (pickup_date, delivery_date) -> SubRentalOrder
```

### Sub-Rental Conditions
```
POST   /api/v1/sub-rental/orders/:id/condition (checkpoint, items[]) -> SubRentalCondition
GET    /api/v1/sub-rental/orders/:id/conditions () -> [SubRentalCondition]
```

### Federation API (extern)
```
GET    /federation/equipment            (type, start_date, end_date) -> AvailableEquipment[] (mTLS)
POST   /federation/sub-rental/request   (requestData) -> { request_id, expires_at } (mTLS)
GET    /federation/sub-rental/quote/:id () -> SubRentalQuote (mTLS)
POST   /federation/sub-rental/quote/:id/respond (accept/reject, items) -> { order_id } (mTLS)
```

## 4.6 Frontend-Seiten & Komponenten

### Pages
- **Partner-Verwaltung** (`/partners`)
  - Tabelle: Name, Status, API-Endpoint, Markupaufschlag, Aktionen
  - "Neuer Partner" Button (Modal)
  - "Verbindung testen" Button
  - Zertifikat-Info anzeigen
  - Status-Indicator (Online/Offline)

- **Sub-Rental Anfragen** (`/sub-rental/requests`)
  - Tabelle: Datum, Partner, Items-Summary, Daten, Status
  - Filter: Status, Partner, Datum
  - "Neue Anfrage" Button (Modal)
  - Quick-Actions: Quote ansehen, Annehmen, Ablehnen

- **Sub-Rental Anfrage-Detail** (`/sub-rental/requests/:id`)
  - Anfrage-Info (Partner, Datum, Items)
  - Quote (vom Partner, wenn vorhanden)
  - Items-Vergleich (Angefordert vs. Angeboten)
  - Actions: Annehmen/Ablehnen, Stornieren

- **Sub-Rental Bestellungen** (`/sub-rental/orders`)
  - Tabelle: Bestellnummer, Partner, Daten, Status, Wert
  - Filter: Status, Partner, Datum
  - Status-Badges (Pending, Pickup, Delivered, Returned, Closed)

- **Sub-Rental Bestellung-Detail** (`/sub-rental/orders/:id`)
  - Bestelldaten, Termine (Pickup, Return)
  - Verknüpfte Invoice
  - Zustandsdokumentation (Pickup + Return Check)
  - "Zustand dokumentieren" Button (Modal mit Fotos, Notizen)
  - Schäden-Kostenkalkulator

### Komponenten
- `PartnerForm` (Create/Edit Modal)
- `SubRentalRequestForm` (Item-Select, Datum-Range)
- `EquipmentAvailabilitySearch` (Partner-Abfrage)
- `QuoteComparison` (Angefordert vs. Angeboten)
- `ConditionCheckModal` (Foto-Upload, Damage-Input)
- `StatusTimeline` (Request → Quote → Order → Pickup → Return → Close)
- `PartnerStatusIndicator` (Online/Offline)

## 4.7 Geschätzter Aufwand

| Bereich | Tage | Bemerkung |
|---------|------|-----------|
| **Backend: mTLS Setup + Cert Management** | 6 | Zertifikat-Validierung, Key-Rotation |
| **Backend: Federation API Endpunkte** | 8 | Equipment-Query, Rate-Limiting, Auth |
| **Backend: Sub-Rental Request CRUD** | 6 | Workflow State-Machine |
| **Backend: Sub-Rental Quote + Acceptance** | 6 | Quote-Handling, Auto-Invoice-Generation |
| **Backend: Zustandsdokumentation** | 5 | Foto-Upload, Damage-Tracking, Cost-Calc |
| **Backend: Partner-Synchronisation** | 4 | Status-Polling, Offline-Handling |
| **Backend: Audit-Logging + Monitoring** | 4 | Federation-Calls tracken |
| **Backend: Tests** | 8 | mTLS-Tests, Sub-Rental-Flows, Edge-Cases |
| **Frontend: Partner-Seiten** | 6 | CRUD, Status-Test |
| **Frontend: Sub-Rental Request Flow** | 8 | Forms, Equipment-Search, Quote-Annahme |
| **Frontend: Sub-Rental Orders** | 6 | Bestellungen, Zustandsdoku, Fotos |
| **Frontend: Partner-Status Dashboard** | 4 | Online-Status, verfügbare Items |
| **Frontend: Responsive Design** | 3 | Mobile-Testing |
| **QA + E2E Testing** | 10 | Full Sub-Rental Flow (Request → Accept → Order → Condition → Invoice) |
| **Docs + Deployment** | 5 | Federation-API-Docs, mTLS-Setup-Guide |
| **Puffer (15%)** | 16 | |
| **GESAMT** | **115 Tage** | ~14–18 Wochen (2–3 Entwickler, parallel zu anderen) |

## 4.8 Definition of Done

✓ mTLS-Kommunikation zwischen zwei Test-Instanzen funktioniert
✓ Equipment-Verfügbarkeit kann abgefragt werden (externe API)
✓ Sub-Rental-Request → Quote → Order Workflow komplett
✓ Auto-Invoice-Generierung beim Order-Akzeptieren
✓ Zustandsdokumentation (Pickup + Return Check) funktioniert
✓ Foto-Upload für Zustandsberichte
✓ Schäden-Kostenverfolgung
✓ Partner-Status-Monitoring (Online/Offline)
✓ Federation Audit-Log vollständig
✓ E2E-Test: Partner hinzufügen → Equipment suchen → Anfrage → Quote → Order → Condition Check → Invoice
✓ API-Dokumentation (mTLS, Endpoints, Fehlerbehandlung)
✓ Stakeholder-Signoff (Marco, Stefan, Admin)

## 4.9 Abhängigkeiten

- **Phase 1, 2, 3** müssen vollständig sein
  - Auth, Equipment, Projekte, Rechnungen, Invoice-Auto-Generation
  - PDF-Export (für Sub-Rental-Dokumente)

---

# Technische Querschnittsthemen

Diese Themen werden **parallel über alle Phasen** umgesetzt:

## i18n (Internationalisierung)

- **Sprachen:** DE (Hauptsprache), EN (optional in Phase 1, aber Infrastruktur vorbereit)
- **Implementation:**
  - Frontend: `react-i18next` (oder `next-i18next` falls Next.js)
  - Backend: Namespace-Ansatz (z.B. `/locales/de.json`, `/locales/en.json`)
  - Datenbankwerte: Enumerations englisch speichern, UI-Labels übersetzen
  - PDF-Templates: mehrsprachig
- **Scope Phase 1:** Deutsche Infrastruktur, Platzhalter für EN
- **Scope Phase 2+:** EN-Übersetzungen hinzufügen
- **Effort:** ~3–5 Tage Pro Phase (je nach Umfang)

## Offline-Modus für Scanner

- **Ziel:** Scanner funktioniert ohne Netzwerk (z.B. in Lagerhallen ohne WLAN)
- **Implementation:**
  - IndexedDB / SQLite für lokale Daten
  - Service Worker für Cache
  - Background Sync API (Sync wenn wieder online)
  - Offline-Indicator in UI
- **Scope Phase 2:** MVP (Basic Offline für Scans)
- **Scope Phase 3+:** Erweiterte Offline-Features
- **Effort:** ~4–6 Tage Phase 2, ~2–3 Tage Phase 3+

## Responsive Design & Mobile-First

- **Breakpoints:**
  - Desktop: 1920px+, 1280px+
  - Tablet: 768px – 1024px
  - Mobile: 375px – 480px (Phones)
- **Scope:** Alle Phasen
- **Critical Paths Mobile:** Scanner, Freelancer-Dashboard, Check-In/Out
- **Testing:** BrowserStack oder Firebase TestLab
- **Effort:** ~2–3 Tage pro Phase (Teil des Feature-Aufwands)

## Testing (Unit, Integration, E2E)

### Unit-Tests
- Backend: Go `testing` Package, Testify für Assertions
- Frontend: Jest + React Testing Library
- Target: 70%+ Coverage für Business-Logic
- **Effort:** ~10–20% des Feature-Aufwands

### Integration-Tests
- API-Tests (Go HTTP-Server)
- Database-Tests (SQLite in-Memory oder Testcontainers)
- Scope Phase 1+
- **Effort:** ~10% des Feature-Aufwands

### E2E-Tests
- Kritische Flows: Auth, Equipment-CRUD, Scanner, Invoice-Erstellung
- Tool: Playwright oder Cypress
- Scope Phase 2+ (Scanner-Flow)
- **Effort:** ~5–10% des Feature-Aufwands

### Manual-Testing (QA)
- Regressions-Checklist pro Phase
- Usability-Testing mit Personas
- Performance-Testing (Datengröße, Load)
- **Effort:** ~1–2 Tage pro Phase

## CI/CD & Deployment

### Continuous Integration
- **Trigger:** Jeder Commit zu `develop`/`main`
- **Pipeline:**
  1. Code-Format Check (gofmt, prettier)
  2. Linting (golangci-lint, ESLint)
  3. Unit-Tests
  4. Build (Go Binary, React Bundle)
  5. Integration-Tests
  6. SAST (Snyk, SonarQube optional)
- **Tool:** GitHub Actions (oder GitLab CI)

### Continuous Deployment
- **Docker Image Build & Push** (auf Registry, z.B. Docker Hub, Quay.io)
- **Auto-Deploy zu Test-Umgebung** (nach CI-Success)
- **Staging-Deploy:** Manual-Trigger
- **Production-Deploy:** Manual mit Genehmigung

### Updates & Rollbacks
- **Docker-Image Versioning:** Semantic Versioning (v1.2.3)
- **Auto-Update Mechanism:** Backend kann prüfen auf neue Versionen, User wird benachrichtigt
- **Rollback:** Docker-Compose mit Image-Version pin
- **Effort:** ~3–4 Tage Setup Phase 1, dann ~0.5 Tage pro Phase für CI/CD-Maintenance

## Performance & Monitoring

### Backend Monitoring
- **Logging:** Structured Logging (JSON) mit Serilog oder zap
- **Metrics:** Prometheus (Request-Rate, Latency, Error-Rate)
- **Tracing:** Jaeger optional
- **Health-Checks:** `/health` Endpoint

### Frontend Monitoring
- **Web Vitals:** Lighthouse, PageSpeed Insights
- **Error-Tracking:** Sentry optional
- **Analytics:** Plausible (privacy-focused)

### Database Performance
- **Indexes:** Strategic indexing auf häufigen Queries
- **Query Profiling:** EXPLAIN ANALYZE
- **Backups:** Täglich, mit Retention-Policy

## Security Best Practices

### Backend
- **Authentication:** JWT mit Refresh-Tokens, HTTPS-only, Secure Cookies
- **Authorization:** RBAC auf allen Endpunkten
- **Input Validation:** All User-Inputs validieren
- **SQL Injection Prevention:** Prepared Statements (ORM oder pgx)
- **CORS:** Nur erlaubte Origins
- **Rate-Limiting:** Per User, Per IP
- **Secrets Management:** Umgebungsvariablen, z.B. `.env` (nicht versioniert)

### Frontend
- **XSS Prevention:** Content Security Policy, Input Sanitization
- **CSRF:** Token-based (SameSite Cookies)
- **HTTPS:** Always
- **Dependency Scanning:** npm audit, Snyk

### Federation (mTLS)
- **Certificate Pinning:** Fingerprint-Validierung
- **Mutual TLS:** Client + Server Certificates
- **Encryption in Transit:** TLS 1.3+

---

# Gesamtbudget & Timeline

## Entwickler-Team Annahmen
- **Team-Größe:** 3–4 Senior/Mid-Level Entwickler
- **Entwickler-Tage:** 5 Tage/Woche, ~7 Stunden/Tag = 35h/Woche
- **Velocity:** ~4–5 Entwickler-Tage pro Woche pro Person (nach Meeting/Admin-Overhead)

## Budgetübersicht

| Phase | Geschätzter Aufwand | Team-Größe | Kalender-Wochen | Go-Live |
|-------|---------------------|-----------|-----------------|---------|
| **Phase 1** | 64 Tage | 3 Dev | 8–10 Wo | Intern (Beta) |
| **Phase 2** | 88 Tage | 3–4 Dev | 11–14 Wo | Intern (Beta) |
| **Phase 3** | 124 Tage | 3–4 Dev | 16–20 Wo | **Produktiv v1.0** |
| **Phase 4** | 115 Tage | 2–3 Dev (parallel zu Phase 3) | 14–18 Wo | v1.1 (optional) |
| **Querschnittsthemen** | ~50 Tage | 1 Dev | Verteilt über alle | Fortlaufend |
| **Gesamt (ohne Phase 4)** | 326 Tage | Avg. 3.2 Dev | ~24 Wochen | v1.0 produktiv |
| **Gesamt (mit Phase 4)** | 441 Tage | Avg. 3.2 Dev | ~28 Wochen | v1.1 |

### Vereinfachte Annahmen
- **Fehlschätzungs-Puffer:** 15% pro Phase (bereits in "Gesamt" enthalten)
- **Querschnittsthemen:** Parallel zu Phasen, ~1 Junior/Mid-Dev mit ~10% Zeit
- **Code-Review + Testing:** Im "Definition of Done" enthalten
- **Meetings + Admin:** ~30% Overhead (berücksichtigt in Kalender-Wochen-Berechnung)

---

# Gantt-Chart Timeline

```
Monat/Wo | 1   2   3   4   5   6   7   8   9  10  11  12  13  14  15  16  17  18  19  20  21  22  23  24  25  26
────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────
Phase 1  │ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ (Wo 1–8, Go-Live: End Wo 8)
         │   [Foundation: Auth, Equipment, Dashboard, Docker]
Phase 2  │                         ━━━━━━━━━━━━━━━━━━━━━━━━━━━━ (Wo 9–16, Go-Live Beta: End Wo 16)
         │                         [Projects, Scanner, Freelancer, Timesheet]
Phase 3  │                                           ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ (Wo 17–24, Go-Live v1.0: End Wo 24)
         │                                           [Quotes, Invoices, DATEV, PDF]
Phase 4  │                                                   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━ (Wo 25–26+, optional v1.1)
         │                                                   [Federation, Sub-Rental]
Querschn.│ ✓ i18n     ✓ Offline      ✓ Testing       ✓ Monitoring (fortlaufend über alle Phasen)
────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────

Legend:
━━━ = Active Development
✓   = Querschnitts-Aktivitäten (parallel)
```

---

# Meilensteine & Signoff-Kriterien

## Meilenstein 1: Phase 1 Complete (Woche 8)
- **Demo:** Marco + Lisa sehen Equipment-Bestand + Lagerplätze
- **Sign-Off:** Marco (GF), Lisa (Lager)
- **Go-Live:** Intern nur

## Meilenstein 2: Phase 2 Complete (Woche 16)
- **Demo:** Lisa scannt Packlisten, Kevin sieht Jobs + Check-In/Out
- **Sign-Off:** Lisa, Kevin, Marco
- **Go-Live:** Intern + Beta-Kunden (optional)

## Meilenstein 3: Phase 3 Complete (Woche 24)
- **Demo:** Thomas erstellt Angebot → Rechnung → DATEV-Export
- **Sign-Off:** Thomas (Buchhaltung), Marco, Admin
- **Go-Live:** **Produktiv v1.0**

## Meilenstein 4: Phase 4 Complete (Woche 26+, optional)
- **Demo:** Marco fragt bei Stefan Sub-Rental-Equipment an
- **Sign-Off:** Marco, Stefan (Partner), Admin
- **Go-Live:** v1.1 (Netzwerk-Effekt)

---

# Risiken & Mitigation

| Risiko | Wahrscheinlichkeit | Impact | Mitigation |
|--------|------------------|--------|-----------|
| **Scope Creep** | Hoch | Hoch | Strikte Phase-Definition, regelmäßige Scope-Review |
| **DATEV-Integration komplexer** | Mittel | Mittel | Frühes Audit mit DATEV-Partner, Mockups vor Implementation |
| **Scanner-Offline-Mode flaky** | Mittel | Mittel | Early PoC Phase 2, umfangreiche Testing |
| **mTLS-Setup fehleranfällig** | Mittel | Hoch | Externe Security-Review, Zertifikat-Management-Tool |
| **Partner-API-Instabilität** | Hoch | Mittel | Fallback-Logik, Polling-Retry, Monitoring |
| **Performance-Probleme bei Datengrowth** | Mittel | Hoch | Early Database-Tuning, Caching-Strategie (Redis optional) |
| **Team-Turnover** | Mittel | Hoch | Dokumentation, Code-Reviews, Knowledge-Sharing |

---

# Lessons Learned & Best Practices

1. **MVP-Fokus:** Strikte Priorisierung (Must/Should/Could) pro Phase. Could-Features verschieben.
2. **Early User-Feedback:** Nach Phase 1 internes Dogfooding mit Lisa + Marco vor Phase 2.
3. **Testing-Early:** Unit + E2E Tests ab Phase 1, nicht am Ende.
4. **Documentation-as-Code:** API-Docs, Architecture-Diagrams im Repo.
5. **CI/CD-First:** Deployment-Pipeline vor Feature-Implementierung.
6. **Security-Reviews:** Externe Reviews vor Phase 4 (Federation).

---

# Anhänge

## A. Tech-Stack Details

| Komponente | Wahl | Begründung |
|-----------|------|-----------|
| Backend | Go + Gin/Chi | Schnell, concurrent, production-ready |
| Frontend | React + TypeScript | Component-based, ecosystem, type-safety |
| Database | PostgreSQL | Relational, ACID, PostGIS optional |
| Auth | JWT + Refresh-Tokens | Stateless, scalable, standard |
| Deployment | Docker + Docker Compose | Portable, reproducible, cloud-agnostic |
| Monitoring | Prometheus + Grafana | Open-source, battle-tested |
| Logging | Structured (JSON) | Machine-readable, searchable |
| Testing | Go: Testify, Frontend: Jest | Standard, good documentation |
| i18n | i18next | React-friendly, large community |

## B. Umgebungen

- **Local:** Docker Compose (3 Container: Frontend, Backend, DB)
- **Test:** Cloud VM (z.B. DigitalOcean, Hetzner) mit CD-Pipeline
- **Staging:** Ähnlich Prod, separate DB
- **Production:** Cloud Kubernetes (optional) oder VPS mit Auto-Updates

## C. Go-Live Checkliste (v1.0)

- [ ] DB-Backups funktionieren und getestet (Recovery-Test)
- [ ] Alle kritischen Flows getestet (Auth, Equipment, Project, Invoice)
- [ ] Performance-Baseline (DB-Größe, Response-Time < 500ms)
- [ ] Security-Audit (OWASP Top 10 Check)
- [ ] Datenschutz (Privacy Policy, GDPR-Compliance)
- [ ] Dokumentation (User-Guide, Admin-Guide, API-Docs)
- [ ] Stakeholder-Schulung (Thomas, Lisa, Kevin, Marco)
- [ ] Support-Plan (Email, Bugtracker, Incident-Response)
- [ ] Monitoring + Alerting aktiv
- [ ] Rollback-Plan dokumentiert

---

**Dokument-Status:** Draft
**Nächste Überprüfung:** Nach Phase 1 Abschluss (Woche 9)
**Verantwortlich:** Tech Lead / Projektmanager

