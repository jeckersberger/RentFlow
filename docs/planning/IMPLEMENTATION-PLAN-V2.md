# RentFlow — Implementation Plan V2

> Letzte Aktualisierung: 2026-03-23
> Status: Aktiver Entwicklungsplan

---

## Übersicht

| Phase | Fokus | Aufwand | Status |
|-------|-------|---------|--------|
| **Phase 1** | Foundation & Usability | ~8 Tasks | Offen |
| **Phase 2** | Scanner & Mobile (PWA) | ~5 Tasks | Offen |
| **Phase 3** | Kommunikation & AI-Automation | ~7 Tasks | Offen |
| **Phase 4** | Federation & RFID | ~6 Tasks | Offen |
| **Phase 5** | Polish & Production | ~6 Tasks | Offen |

---

## Phase 1: Foundation & Usability

**Ziel**: Ein nicht-technischer User kann sich einloggen, Einstellungen verwalten, Mitarbeiter anlegen, und die Kern-Workflows (Equipment/Projekte/Rechnungen) ohne Verwirrung nutzen.

### 1.1 Settings — SMTP-Konfiguration
**Was**: Neuer Tab "E-Mail / SMTP" in Settings. Formular: SMTP Host, Port, Username, Passwort (maskiert), Absender-Email, Absender-Name, TLS Toggle, "Test senden" Button.
**Backend**: `PUT /api/v1/tenants/{id}/smtp-config` und `POST /api/v1/tenants/{id}/smtp-test` in auth-service.
**Warum**: Ohne SMTP-Config kann das System keine E-Mails versenden (Rechnungen, Mahnungen, Passwort-Reset).
**Dateien**:
- `frontend/src/pages/Settings/SettingsPage.tsx` — Tab + Formular + Mutation
- `frontend/src/services/api.ts` — `tenantApi.updateSmtpConfig()`, `tenantApi.testSmtp()`
- `services/auth-service/internal/adapters/http/handlers.go` — neue Handler
- `services/auth-service/internal/adapters/http/router.go` — neue Routen
**Komplexität**: M

### 1.2 Settings — AI Provider / API-Key Konfiguration
**Was**: Tab "Integrationen" in Settings. Pro AI-Provider (Anthropic, OpenAI, Google, Mistral, Ollama): API-Key (verschlüsselt), Modell-Auswahl, Enable/Disable Toggle, Default-Provider.
**Backend**: Existiert bereits (`POST /api/v1/ai/providers`, `GET /api/v1/ai/providers`).
**Warum**: AI-Features sind ein Kern-Differentiator. Ohne Config-UI kann niemand AI aktivieren.
**Dateien**:
- `frontend/src/pages/Settings/SettingsPage.tsx` — Tab "Integrationen"
- `frontend/src/services/api.ts` — `aiApi.providers()`, `aiApi.registerProvider()` verdrahten
**Komplexität**: M

### 1.3 Settings — Benachrichtigungs-Präferenzen
**Was**: Hardcoded Checkboxen im Notifications-Tab durch echtes Formular ersetzen, das die notification-service Preference API nutzt (`GET /preferences`, `PUT /preferences`). Pro Event-Typ: E-Mail/In-App/Push Toggles, Ruhezeiten, Digest-Modus.
**Warum**: User müssen kontrollieren können, welche Benachrichtigungen sie bekommen.
**Dateien**:
- `frontend/src/pages/Settings/SettingsPage.tsx` — Notifications-Tab umschreiben
- `frontend/src/services/api.ts` — `notificationApi.getPreferences()`, `notificationApi.updatePreference()`
**Komplexität**: S

### 1.4 Settings — Backup-Konfiguration
**Was**: Tab "Backup" in Settings. Letztes Backup, Zeitplan (täglich/wöchentlich), Ziel (lokal/S3), manueller "Jetzt sichern" Button.
**Warum**: Self-hosted User brauchen Backup-Management.
**Dateien**:
- `frontend/src/pages/Settings/SettingsPage.tsx` — Tab "Backup"
- Backend: Backup-Endpoint in auth-service oder reporting-service
**Komplexität**: M

### 1.5 User Management — Vollständiges CRUD
**Was**: Backend hat bereits: `GET /api/v1/users`, `PUT /api/v1/users/{id}/roles`, `DELETE /api/v1/users/{id}`. Frontend zeigt nur read-only Liste. Hinzufügen:
- "Neuer Benutzer" Button → Modal mit E-Mail, Name, Rollen-Auswahl
- Klick auf User → Edit-Drawer mit Rollen-Dropdown, Aktiv/Deaktiviert Toggle
- Deaktivierungs-Bestätigungsdialog
**Warum**: Geschäftsführer muss Mitarbeiter einladen und Rollen zuweisen können ohne Kommandozeile.
**Dateien**:
- `frontend/src/pages/Settings/SettingsPage.tsx` — Users-Tab erweitern
- `frontend/src/services/api.ts` — `userApi.create()`, `userApi.updateRole()`, `userApi.deactivate()`
- `services/auth-service/internal/adapters/http/handlers.go` — CreateUser/InviteUser sicherstellen
**Komplexität**: M

### 1.6 Deutsche Übersetzung vervollständigen
**Was**: i18n-System existiert mit DE/EN für common, auth, equipment, projects, invoices. Aber die meisten Seiten haben hardcoded deutsche Strings statt `t('key')`. Fehlende Namespaces erstellen: scanner, federation, settings, warehouse, transport, maintenance, crew, reports, ai. Status-Labels wie "available" → "Verfügbar" durchgängig.
**Warum**: Konsistenz und Vorbereitung für EN-Markt.
**Dateien**:
- `frontend/src/i18n/locales/de/*.json` — neue Dateien
- `frontend/src/i18n/locales/en/*.json` — neue Dateien
- Alle Seiten-Komponenten mit hardcoded Strings
**Komplexität**: L

### 1.7 UI-Polish und Konsistenz
**Was**: (a) Inline-Styles durch SCSS-Module ersetzen. (b) Gemeinsame Patterns extrahieren (Status-Badges, DataTables). (c) Alle Status-Werte konsistent mit deutschen Labels und Farb-Badges. (d) Loading-Skeletons statt "wird geladen..." Text.
**Warum**: Nicht-technische User bewerten Software nach visueller Konsistenz.
**Komplexität**: M

### 1.8 Mock-Daten entfernen
**Was**: `api.ts` hat massiven Block hardcoded Mock-Daten und `MOCK_MODE` Conditionals. Alle API-Calls auf echtes Backend umstellen. Mock-Daten in separates `__mocks__/` Verzeichnis verschieben.
**Warum**: Mock-Mode maskiert echte Bugs und schafft Verwirrung.
**Komplexität**: M

---

## Phase 2: Scanner & Mobile

**Ziel**: Lisa kann mit Handy oder Zebra TC21 Barcodes scannen, Equipment ein-/auschecken, und das System funktioniert offline.

### 2.1 Scanner Session UI-Flow
**Was**: Scanner-Backend hat volle Session-Verwaltung. Frontend braucht:
- Session-Start: Projekt wählen, Modus (Check-Out/Check-In/Inventur), "Session starten"
- Während Session: Kontext-Bar oben (Projekt, Modus, Scan-Zähler)
- Pro Scan: sofortiges Feedback (grüner Haken / rotes X mit Sound)
- Session beenden: Zusammenfassung, PDF-Export
**Warum**: Lisa braucht einen klaren Workflow: Projekt wählen → scannen → fertig.
**Abhängigkeiten**: Task 1.8
**Komplexität**: L

### 2.2 PWA Service Worker mit Background Sync
**Was**: VitePWA ist konfiguriert. IndexedDB Offline-Queue existiert. Aber Background Sync API wird nicht genutzt. Hinzufügen:
- Workbox Background Sync Queue für Scan-Submissions
- Offline-Scans queuen automatisch und werden bei Reconnect gesendet
- Offline-Indikator im App-Shell
- Equipment- und Projektliste für Offline-Barcode-Auflösung cachen
**Warum**: Lager haben oft schlechtes WiFi. Scans dürfen nie verloren gehen.
**Abhängigkeiten**: Task 2.1
**Komplexität**: L

### 2.3 Manuelle Eingabe und Barcode-Formate
**Was**: Manuelles Barcode-Eingabefeld (für unleserliche Labels). Vibrations-Feedback bei erfolgreichem Scan (`navigator.vibrate`). Audio-Beep bei Scan (Web Audio API).
**Warum**: Nicht alle Barcodes sind kameratauglich. Audio/Vibration ist auf der lauten Lagerhalle essentiell.
**Komplexität**: S

### 2.4 Scanner Ergebnis-Detail und Quick Actions
**Was**: Nach Scan → Equipment-Detailkarte: Name, Foto, Status, aktuelles Projekt, letzte Wartung, Zustand. Quick-Action Buttons: "Check-Out an Projekt", "Check-In", "Defekt melden", "Standort ändern".
**Warum**: Lisa muss sofort nach dem Scan handeln können, nicht zu einer anderen Seite navigieren.
**Komplexität**: M

### 2.5 Multi-Device Support
**Was**: Testen und fixen auf: iPhone Safari (PWA), Android Chrome, Zebra TC21 (Hardware-Scanner sendet Keyboard-Events). Für Zebra: Keyboard-Speed-Input erkennen und auto-submit.
**Warum**: Kevin nutzt iPhone, Lisa nutzt Zebra. Beides muss funktionieren.
**Komplexität**: M

---

## Phase 3: Kommunikation & AI-Automation

**Ziel**: E-Mails gehen automatisch raus, Benachrichtigungen sind konfigurierbar, AI liefert echten Mehrwert.

### 3.1 E-Mail Template System
**Was**: Go Template Engine mit HTML-E-Mail-Templates: Rechnung, Mahnung (Stufe 1/2/3), Projektbestätigung, Passwort-Reset, Einladung, Wartungserinnerung, Rückgabe überfällig. Templates in `services/notification-service/templates/`.
**Warum**: Automatisierte E-Mails sind Kernfunktion.
**Abhängigkeiten**: Task 1.1 (SMTP Config)
**Komplexität**: L

### 3.2 Notification Event Triggers
**Was**: Business-Events lösen Benachrichtigungen aus:
- Rechnung überfällig → E-Mail + In-App
- Equipment Check-Out → In-App an Projektleiter
- Wartung fällig → E-Mail + In-App an Lagerpersonal
- Neues Projekt → In-App an alle
- Scanner-Session abgeschlossen → In-App Zusammenfassung
**Komplexität**: M

### 3.3 AI Preis-Optimierung Widget
**Was**: Widget auf Equipment-Detailseite. Zeigt: aktueller Preis, empfohlener Preis, Konfidenz, Begründung, Akzeptieren/Ablehnen. Backend-Endpoint existiert.
**Warum**: Marquee AI-Feature. Sofortiger Nutzen für nicht-technische User.
**Abhängigkeiten**: Task 1.2
**Komplexität**: M

### 3.4 AI Dashboard Insights
**Was**: Dashboard-Sektion mit KI-Empfehlungen: "5 Geräte sollten vor der Festival-Saison gewartet werden", Token-Verbrauch, akzeptierte/abgelehnte Vorschläge.
**Warum**: Marco will 30-Sekunden-Überblick. AI Insights erhöhen den wahrgenommenen Wert.
**Komplexität**: M

### 3.5 AI Smart Asset Creator
**Was**: "KI Equipment anlegen" Button. User lädt Foto hoch oder gibt Beschreibung ein. AI gibt Vorschläge: Name, Kategorie, Hersteller, Modell, Gewicht, Leistung. User prüft und bestätigt.
**Warum**: Spart Lisa 5 Minuten pro Equipment. Reduziert Fehler.
**Komplexität**: M

### 3.6 Predictive Maintenance
**Was**: "KI Wartungsvorhersage" auf Maintenance-Seite. Pro Equipment: vorhergesagtes nächstes Wartungsdatum, Konfidenz, Begründung. Batch-Action: "Wartungspläne für alle überfälligen erstellen".
**Warum**: Verhindert Geräteausfälle bei Events.
**Komplexität**: M

### 3.7 Scheduled Jobs Infrastruktur
**Was**: Cron-Scheduler im workflow-service mit `robfig/cron/v3`:
- Täglich: Überfällige Rechnungen prüfen, Mahnungen senden
- Wöchentlich: Wartungserinnerungen
- Monatlich: AI Demand Forecast Refresh
**Warum**: Automation braucht Background-Jobs.
**Abhängigkeiten**: Tasks 3.1, 3.2
**Komplexität**: L

---

## Phase 4: Federation & RFID

**Ziel**: Zwei RentFlow-Instanzen können sich verbinden, Equipment des Partners durchsuchen, Sub-Rental-Anfragen stellen. RFID-Hardware funktioniert für Bulk-Scanning.

### 4.1 Federation Connection Wizard
**Was**: Frontend-Wizard:
1. Partner-URL eingeben
2. Auto-Discovery via `GET /.well-known/rentflow-federation`
3. Zertifikat-Austausch (mTLS)
4. Verbindung bestätigen
**Warum**: Federation ist Alleinstellungsmerkmal. Verbindung muss einfach sein.
**Komplexität**: L

### 4.2 Partner Equipment Katalog
**Was**: Nach Verbindung: Equipment-Liste des Partners mit Filtern (Kategorie, Verfügbarkeit, Datumsbereich). Verfügbarkeitskalender pro Item. "Anfrage senden" Button.
**Warum**: So findet Marco Equipment von SoundPro für das Festival.
**Komplexität**: M

### 4.3 Sub-Rental Anfrage- und Genehmigungsflow
**Was**: Vollständiger Request → Approve → Reject → Complete Flow im UI. Backend-Endpoints existieren. Übergabedokument-Generierung bei Genehmigung. Rechnung bei Abschluss.
**Komplexität**: L

### 4.4 RFID Hardware Integration — Backend
**Was**: RFID-Scan-Typen definieren. `POST /api/v1/scanner/rfid/bulk` Endpoint. RFID-Tags auf Equipment mappen via `equipment.rfid_tag` Feld. Support: Zebra FX9600 (Netzwerk-API), CF-H906 (Keyboard Wedge).
**Warum**: RFID ermöglicht Bulk-Scanning (100 Items in 2 Sekunden).
**Komplexität**: L

### 4.5 RFID Tag Management UI
**Was**: RFID-Sektion auf Equipment-Detailseite. Aktueller Tag anzeigen. "Tag zuweisen" per Scan. Bulk-Tag-Zuweisung. Tag-Status (aktiv/deaktiviert).
**Komplexität**: M

### 4.6 Cross-Tenant RFID Reading
**Was**: RFID-Tags von Federation-Partnern erkennen. Unbekannte Tags über Federation-Partner auflösen. "Partner-Equipment" anzeigen.
**Warum**: Verhindert Verwirrung wenn Partner-Equipment im eigenen Lager ist.
**Abhängigkeiten**: Tasks 4.1, 4.4
**Komplexität**: L

---

## Phase 5: Polish & Production

**Ziel**: System ist produktionsreif mit Event Sourcing, Reporting, Security, Deployment-Dokumentation.

### 5.1 KurrentDB Event Sourcing
**Was**: Services schreiben Domain-Events nach KurrentDB. Projektionen updaten PostgreSQL Read-Models. Start mit inventory-service als Pilot.
**Komplexität**: XL

### 5.2 Reporting & Analytics Dashboard
**Was**: Echte Daten in ReportsPage: Umsatz/Monat, Equipment-Auslastung, Top 10 nach Umsatz, Überfällige Rechnungen. CSV/PDF Export.
**Komplexität**: L

### 5.3 Dokumenten-Management
**Was**: PDF-Generierung für Rechnungen. Vertrags-Templates. QR-Label-Druck. Dokument-Upload/Storage.
**Komplexität**: L

### 5.4 Security Audit
**Was**: Auth-Middleware auf allen Endpoints. RBAC Permission Checks. Rate Limiting. CORS. Input Validation. SQL Injection Review.
**Komplexität**: L

### 5.5 Performance Optimierung
**Was**: Redis Caching. Datenbank-Index-Review. Frontend Code Splitting. Bild-Optimierung.
**Komplexität**: M

### 5.6 Deployment Dokumentation
**Was**: README, docker-compose.prod.yml, .env.example, Backup-Script, Update-Guide, Monitoring-Setup.
**Komplexität**: M

---

## Prioritäts-Matrix

| Prio | Task | Phase | Komplexität | User Impact |
|------|------|-------|-------------|-------------|
| P0 | 1.5 User Management UI | 1 | M | Kann keine Mitarbeiter anlegen |
| P0 | 1.1 SMTP Konfiguration | 1 | M | Kann keine E-Mails senden |
| P0 | 1.8 Mock-Daten entfernen | 1 | M | Echte Daten erforderlich |
| P1 | 1.2 AI Provider Config | 1 | M | Kann AI nicht nutzen |
| P1 | 2.1 Scanner Session UI | 2 | L | Scanner ohne Flow unbrauchbar |
| P1 | 1.3 Notification Prefs | 1 | S | Benachrichtigungen nicht konfigurierbar |
| P1 | 3.1 Email Templates | 3 | L | Keine automatisierten E-Mails |
| P2 | 2.2 PWA Background Sync | 2 | L | Offline-Scans gehen verloren |
| P2 | 3.3 AI Preis-Widget | 3 | M | Kern-Differentiator |
| P2 | 1.6 i18n vervollständigen | 1 | L | Gemischte Sprachen im UI |
| P2 | 1.7 UI Polish | 1 | M | Professionelles Erscheinungsbild |
| P3 | 4.1 Federation Wizard | 4 | L | Federation nicht nutzbar |
| P3 | 4.4 RFID Backend | 4 | L | RFID nicht nutzbar |
| P3 | 5.1 Event Sourcing | 5 | XL | Architektur-Schulden |
| P4 | 5.2 Reporting Dashboard | 5 | L | Nice-to-have |
| P4 | 5.6 Deployment Docs | 5 | M | Für Go-Live nötig |

---

## Kritische Dateien

| Datei | Warum kritisch |
|-------|---------------|
| `frontend/src/pages/Settings/SettingsPage.tsx` | Höchste Priorität — muss erweitert werden mit SMTP, Integrationen, Backup Tabs |
| `frontend/src/services/api.ts` | Zentrales API-Gateway mit Mock-Daten-Blöcken — muss auf echte Endpoints umgestellt werden |
| `services/auth-service/internal/adapters/http/router.go` | User Management, SMTP Config, Invitation Endpoints |
| `frontend/src/pages/Scanner/ScannerPage.tsx` | Session-Management, RFID-Modus, Background Sync |
| `services/notification-service/internal/adapters/http/handlers.go` | E-Mail-Templates, SMTP Integration, Event-Trigger |
