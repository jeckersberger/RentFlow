# CrateDesk — Vollstaendiger Masterplan

> Stand: 29.03.2026 — Basiert auf ALLEN Spec-Dokumenten
> 19 Services, 150+ Endpoints, 40+ DB-Tabellen, 200+ UI-Seiten

## Legende
- [x] Existiert und funktioniert
- [~] Existiert aber unvollstaendig/buggy
- [ ] Fehlt komplett

---

## TIER 1: GELD VERDIENEN (Wochen 1-4)
*Ohne das kann Janis keine Auftraege abwickeln*

### 1.1 Document Engine + PDF-Generierung
- [ ] HTML-Template-System (Go html/template mit Platzhaltern)
- [ ] 8 Default-Templates: Rechnung, Angebot, Gutschrift, Mahnung 1-3, Lieferschein, Packliste, Vertrag, E-Check-Protokoll
- [ ] PDF-Generierung via chromedp (HTML→PDF)
- [ ] GET /api/v1/invoices/{id}/pdf
- [ ] GET /api/v1/documents/templates — Template-Liste
- [ ] PUT /api/v1/documents/templates/{id} — Template bearbeiten
- [ ] Template-Felder: Logo-URL, Primaerfarbe, Schriftart, Kopf-/Fusszeile, Impressum
- [ ] Unveraenderliches Archiv (SHA-256 Checksummen, GoBD)
- [ ] Frontend: Template-Editor mit Live-Preview
- [ ] Frontend: PDF-Download + Preview auf Rechnungs-/Angebotsseite

### 1.2 E-Mail-Versand (SMTP)
- [ ] SMTP-Client (gomail) im notification-service
- [ ] HTML E-Mail-Templates: Rechnung, Mahnung, Angebot, Einladung, Passwort-Reset
- [ ] POST /api/v1/invoices/{id}/send — PDF als Anhang
- [ ] POST /api/v1/offers/{id}/send — Angebot versenden
- [ ] Frontend: "Per E-Mail senden" Button
- [ ] Frontend: E-Mail-Vorschau vor Versand
- [ ] Frontend: SMTP-Einstellungen in Settings

### 1.3 Angebots-System (Quotes)
- [ ] DB: quotes + quote_items Tabellen
- [ ] POST /api/v1/offers — Angebot erstellen
- [ ] GET /api/v1/offers — Angebotsliste
- [ ] GET /api/v1/offers/{id}/pdf — Angebots-PDF
- [ ] POST /api/v1/offers/{id}/send — Per E-Mail
- [ ] POST /api/v1/offers/{id}/convert-to-invoice — In Rechnung umwandeln
- [ ] Status: draft/sent/accepted/rejected/expired
- [ ] Frontend: Angebots-Editor (Positionen, Rabatte, Intro-/Outro-Text)
- [ ] Frontend: 1-Klick Angebot→Rechnung Konvertierung

### 1.4 Mahnwesen (3-stufig)
- [ ] DB: dunning_entries Tabelle
- [ ] POST /api/v1/dunning/send-reminder
- [ ] POST /api/v1/invoices/dunning-run — Automatischer Mahnlauf
- [ ] Mahnstufen: Zahlungserinnerung (7d), Mahnung 1 (14d), Mahnung 2 (28d)
- [ ] Konfigurierbare Fristen + Mahngebuehren
- [ ] Frontend: Faellige-Rechnungen-Uebersicht mit Mahnstatus

### 1.5 Preiskalkulations-Engine
- [ ] DB: price_rules Tabelle
- [ ] 6 Preistypen: Tagesatz, Wochenpauschale, Staffelpreise, Mengenstaffel, Saisonzuschlag, Bundle
- [ ] POST /api/v1/equipment/{id}/price-calculation
- [ ] Automatische Preisberechnung: Zeitraum × Tagespreis × Rabatt
- [ ] Frontend: Preis-Kalkulator in Angebots-/Rechnungs-Editor

---

## TIER 2: BUCHHALTUNG (Wochen 5-6)
*Steuerberater braucht Daten, Bank muss abgeglichen werden*

### 2.1 Bank-Import + OP-Abgleich
- [ ] DB: bank_transactions Tabelle
- [ ] POST /api/v1/banking/import-transactions — MT940/CAMT.053/CSV Upload
- [ ] POST /api/v1/banking/auto-match — Automatischer OP-Abgleich
- [ ] Matching-Algorithmus: Exakter Betrag + Rechnungsnummer → Betrag ±1ct → Kundenname
- [ ] Frontend: Bank-Import (Datei-Upload, Format waehlen)
- [ ] Frontend: Matching-UI (Transaktion↔Rechnung zuordnen)
- [ ] Frontend: Offene-Posten-Dashboard

### 2.2 DATEV/WISO-Export
- [ ] POST /api/v1/export/datev — DATEV CSV (SKR03/SKR04)
- [ ] WISO-Export Format
- [ ] Debitorennummern-Mapping (10000 + Kundennummer)
- [ ] Frontend: Export-Dialog (Zeitraum, Format waehlen)

### 2.3 Teilrechnungen + Gutschriften
- [ ] DB: partial_invoices Tabelle
- [ ] POST /api/v1/invoices/{id}/partial-invoice
- [ ] POST /api/v1/invoices/{id}/credit-note — Gutschrift mit Verweis auf Original
- [ ] Restbetrag-Tracking bei Teilrechnungen
- [x] GoBD: Keine Loeschung, nur Gegenbuchung (Hash-Kette existiert)

---

## TIER 3: PROJEKT-WORKFLOW (Wochen 7-8)
*Der taegliche Arbeitsablauf: Anfrage→Angebot→Projekt→Lieferschein→Rechnung*

### 3.1 Projekt-Lifecycle
- [~] Projekt CRUD existiert, aber Status-Maschine fehlt
- [ ] Status-Uebergaenge: inquiry→offer_sent→confirmed→in_preparation→active→completed→invoiced→archived
- [ ] Automatischer Status-Wechsel bei Check-Out/Check-In
- [ ] Frontend: Kanban-Board fuer Projekte
- [ ] Frontend: Projekt-Detail mit Tabs (Stammdaten, Equipment, Packliste, Finanzen, Timeline)

### 3.2 Equipment-Zuweisung + Verfuegbarkeit
- [~] project_equipment Tabelle existiert
- [ ] Verfuegbarkeitscheck-Algorithmus: Bestand - Reserviert - Wartung
- [ ] Doppelbuchungs-Erkennung ueber alle Projekte
- [ ] Sub-Rental-Vorschlag bei Knappheit (Federation)
- [ ] Frontend: Equipment-Auswahl mit Live-Verfuegbarkeits-Ampel (gruen/gelb/rot)
- [ ] Frontend: Verfuegbarkeits-Timeline pro Equipment

### 3.3 Packlisten
- [~] packlists + packlist_items existieren
- [ ] Packlisten-Generierung sortiert nach Lagerort (optimale Kommissionierroute)
- [ ] Status pro Item: PLANNED→PACKED→LOADED→ON_SITE→RETURNED
- [ ] PDF-Export als Packliste
- [ ] Frontend: Packlisten-Editor (abhaken, Drag&Drop)

### 3.4 Kalender + Timeline
- [ ] GET /api/v1/projects/calendar — Kalender-API
- [ ] GET /api/v1/projects/{id}/calendar.ics — CalDAV-Export
- [ ] Frontend: Kalender-Monatsansicht (Projekte als Balken)
- [ ] Frontend: Equipment-Auslastungs-Timeline

---

## TIER 4: LAGER + SCANNER (Wochen 9-10)
*Physisches Equipment verwalten*

### 4.1 Scanner (CameraX + Offline)
- [~] Scanner-App existiert, aber nur Text-Input
- [ ] CameraX Barcode/QR-Scanning
- [ ] Scan-Kontexte: checkout, checkin, inventory, locate, damage, restock
- [ ] Offline-Queue (Room DB / IndexedDB)
- [ ] Background-Sync bei Reconnect
- [ ] Check-Out: Equipment→Projekt mit Packlisten-Validierung
- [ ] Check-In: Zustandsbewertung (OK/beschaedigt/defekt) + Foto bei Defekt
- [ ] RFID-Flightcase-Validator (Bulk-Scan)
- [ ] Digitale Unterschrift bei Uebergabe

### 4.2 Lagerverwaltung (hierarchisch)
- [~] warehouse_service existiert mit Basis-CRUD
- [ ] Hierarchie: Lager→Zone→Regal→Fach (max 5 Ebenen)
- [ ] Lagerbewegungsprotokoll (inbound/outbound/transfer/adjustment)
- [ ] Inventur-Workflow (Voll/Teil/Stichprobe)
- [ ] KI-Lagerplatz-Optimierung (haeufig zusammen→nahe beieinander)
- [ ] Frontend: Lagerstruktur-Baum-Editor
- [ ] Frontend: Inventur-Zaehl-UI (Mobile, Offline)
- [ ] Frontend: Soll-Ist-Vergleich Inventur

---

## TIER 5: CREW + PERSONAL (Wochen 11-12)
*Mitarbeiter und Freelancer verwalten*

### 5.1 Crew-Management
- [~] crew-service existiert mit Basis-CRUD
- [ ] Employee vs. Freelancer Unterscheidung
- [ ] Qualifikations-Matrix (IPAF, Rigger, Tonmeister)
- [ ] Verfuegbarkeits-Matrix (Wochenansicht)
- [ ] Konflikt-Erkennung (Doppelbuchung)
- [ ] CalDAV-Sync fuer Kalender-Apps
- [ ] Frontend: Crew-Uebersicht mit Qualifikations-Badges
- [ ] Frontend: Verfuegbarkeits-Kalender

### 5.2 Zeiterfassung
- [ ] POST /api/v1/timesheets/check-in + check-out
- [ ] Stunden, Pausen, Nachtzuschlag, Ueberstunden
- [ ] GPS bei Check-In
- [ ] Stundenexport CSV fuer Lohnabrechnung
- [ ] Frontend: Check-In/Out Widget (grosser Button, mobile-optimiert)

### 5.3 Freelancer-Portal
- [ ] Freelancer sieht nur eigene Einsaetze
- [ ] Einladungs-Workflow (Link-Generator)
- [ ] Rechnungsgenerierung aus Timesheets

---

## TIER 6: TRANSPORT + WARTUNG (Wochen 13-14)
*Logistik und Geraetepflege*

### 6.1 Transport
- [~] transport-service existiert mit Basis-CRUD
- [ ] Fahrzeug-Kapazitaetspruefung (Gewicht + Volumen)
- [ ] Tour-Planung (Haltestellen, Reihenfolge)
- [ ] KI-Routenoptimierung
- [ ] Transportkosten-Tracking (Kraftstoff, Maut, Fahrzeit)
- [ ] CO2-Berechnung pro Tour
- [ ] Frontend: Tour-Planungs-Dialog mit Kapazitaets-Indikator

### 6.2 Wartung + E-Check
- [~] maintenance-service existiert mit Basis-CRUD
- [ ] Wartungsplan-Typen: zeitbasiert, einsatzbasiert, zustandsbasiert, E-Check
- [ ] DGUV V3 Workflow (Fristenverwaltung, Messwert-Validierung)
- [ ] IZYTRON.IQ XML-Import (Pruefergebnis-Import)
- [ ] Pruefplaketten-Druck (ZPL)
- [ ] Frontend: Wartungs-Dashboard (ueberfaellig/diese Woche/naechste 30 Tage)
- [ ] Frontend: E-Check-Import-Dialog

### 6.3 Versicherung
- [~] insurance-service existiert mit Basis-CRUD
- [ ] 8-Schritte-Schadensfall-Prozess
- [ ] Deckungspruefung automatisch
- [ ] Fotodokumentation unveraenderlich (SHA-256)
- [ ] Schadensformular-Export (Versicherungs-Vorlage)
- [ ] Frontend: Schadensfall-Meldeformular mit Kamera

---

## TIER 7: ADVANCED (Wochen 15-18)
*Automatisierung, KI, Partnernetzwerk*

### 7.1 KI Multi-Provider
- [~] ai-service existiert mit Basis-CRUD
- [ ] Provider-Adapter: Anthropic Claude, OpenAI GPT-4o, Google Gemini, Mistral, Ollama
- [ ] Smart Asset Creator (Foto→Equipment-Daten)
- [ ] Preis-Optimierung (historische Daten + Saisonalitaet)
- [ ] Nachfrage-Prognose
- [ ] KI-Mahnstufenwahl (Zahlungswahrscheinlichkeit)
- [ ] Anonymisierung vor Cloud-API (DSGVO)
- [ ] Few-Shot-Learning (Feedback-Loop)
- [ ] Frontend: KI-Chat-Interface
- [ ] Frontend: Preis-Widget auf Equipment-Detail

### 7.2 Workflow-Engine (No-Code)
- [~] workflow-service existiert mit Basis-CRUD
- [ ] Trigger: Event-basiert, CRON, Manuell, Webhook
- [ ] Aktionen: E-Mail, Notification, Task erstellen, Status aendern, Warten, Bedingung, Genehmigung
- [ ] 7 vordefinierte Templates
- [ ] Frontend: Visueller Editor (React Flow, Drag&Drop)

### 7.3 Federation (P2P Equipment-Sharing)
- [~] federation-service existiert mit Basis-CRUD
- [ ] mTLS-Zertifikat-Austausch
- [ ] Partner-Equipment-Verfuegbarkeit abfragen
- [ ] Sub-Rental-Workflow: Anfrage→Bestaetigung→Uebergabe→Rueckgabe→Rechnung
- [ ] Uebergabe-Dokumentation mit Fotos + Unterschrift
- [ ] Automatische Rechnungserstellung nach Rueckgabe
- [ ] Frontend: Partner-Pairing-Dialog
- [ ] Frontend: Sub-Rental-Anfrage + Tracking

### 7.4 Notification-Center
- [~] notification-service existiert mit Basis-CRUD
- [ ] Kanaele: In-App (WebSocket), Web Push (VAPID), E-Mail
- [ ] Nutzer-Praeferenzen pro Event-Typ
- [ ] Ruhezeiten (22-07 Uhr)
- [ ] Daily Digest E-Mail
- [ ] Frontend: Glocken-Icon mit Unread-Badge + Popover

### 7.5 Audit-Trail (GoBD)
- [~] audit-service existiert mit Basis-CRUD
- [ ] Unveraenderliches Log mit SHA-256-Kette
- [ ] Checksummen-Verifikation
- [ ] Betriebspruefungs-Export (ZIP)
- [ ] DSGVO-Pseudonymisierung
- [ ] Frontend: Audit-Log-Suche + Verifikation

---

## TIER 8: PORTALE + EXTRAS (Wochen 19-22)
*Nice-to-have, aber differenzierend*

### 8.1 Kunden-Portal
- [ ] Self-Service: Angebote ansehen, digital unterschreiben
- [ ] Rechnungen herunterladen
- [ ] Verfuegbarkeit pruefen
- [ ] Oeffentliche Equipment-Info-Seite (via QR-Code)

### 8.2 Ausgaben-Verwaltung (Receipt OCR)
- [~] expense-service existiert mit Basis-CRUD
- [ ] Mobile PWA fuer Kassenbon-Fotos
- [ ] OCR via KI (Claude Vision / OpenAI Vision)
- [ ] Automatische Kategorisierung (SKR04)
- [ ] Genehmigungs-Workflow
- [ ] Wiederkehrende Ausgaben
- [ ] Budget-Tracking
- [ ] NAS-Speicher-Integration

### 8.3 Reporting + Dashboards
- [~] reporting-service existiert mit Basis-CRUD
- [ ] Rollenspezifische KPI-Dashboards (GF/Lager/Buchhaltung)
- [ ] Umsatz-Report (Zeitreihe, Breakdown nach Kategorie/Kunde)
- [ ] Auslastungs-Report (Equipment-Nutzung %)
- [ ] Crew-Stunden-Report
- [ ] CO2/ESG-Report
- [ ] Diagramme (Recharts)
- [ ] CSV/PDF/Excel-Export

### 8.4 Frontend UX
- [ ] Dark/Light Mode Toggle
- [ ] Command Palette (Cmd+K)
- [ ] i18n DE/EN (react-i18next)
- [ ] PWA Offline komplett (Service Worker)
- [ ] WebSocket Echtzeit-Updates
- [ ] Responsive Design (Mobile-First)
- [ ] Loading Skeletons ueberall
- [ ] Toast-Benachrichtigungen ueberall

---

## WAS BEREITS EXISTIERT (Ist-Stand 29.03.2026)

### Backend (19 Go-Services, alle deployed)
- [x] auth-service: Login, Register, JWT RS256, RBAC, Brute-Force-Schutz
- [x] inventory-service: Equipment CRUD, Categories, Types, Flightcases, History, Search
- [x] project-service: Projects CRUD, Project Equipment, Packlists, Reservations
- [x] scanner-service: Scan, Checkout, Checkin, BulkSync, AdhocBooking, Devices
- [x] warehouse-service: Warehouses, Zones, Racks, Locations, Movements, Inventory Checks
- [x] invoice-service: Invoices CRUD, Items, Finalize (GoBD), Payments, Kleinunternehmer
- [x] document-service: Templates, Documents, Attachments (nur DB, keine PDF-Generierung)
- [x] crew-service: Members, Assignments, Qualifications
- [x] federation-service: Partners, Shared Listings, Requests
- [x] maintenance-service: Schedules, Tasks, Logs
- [x] transport-service: Vehicles, Transport Orders, Items
- [x] insurance-service: Policies, Insured Equipment, Claims
- [x] workflow-service: Definitions, Instances, Actions
- [x] ai-service: Predictions, Suggestions, Training Data
- [x] notification-service: Notifications, Preferences, Unread Count
- [x] reporting-service: Report Definitions, Snapshots, Dashboard Widgets
- [x] audit-service: Audit Logs, Audit Policies
- [x] expense-service: Categories, Expenses, Receipts, Approval
- [x] customer-service: Customers CRUD, Contacts, Contact Notes

### Frontend
- [x] Login, Dashboard (echte API-Daten), Sidebar-Navigation
- [x] Listen: Equipment, Kunden, Projekte, Rechnungen, Crew, Lager, Dokumente, Ausgaben, Wartung, Transport, Reports
- [x] Detail-Seiten: Equipment, Kunden, Projekte, Rechnungen
- [x] Create/Edit Forms: Equipment, Kunden, Projekte, Rechnungen
- [x] Account/Profil-Seite
- [x] Dark Theme, Framer Motion Animationen

### Infrastruktur
- [x] Hetzner CX33, Docker, 23 Container, Cloudflare Tunnel
- [x] PostgreSQL 16, Redis 7, Traefik
- [x] CORS fix, Seed-Script
- [x] Scanner-App (Kotlin, APK signiert, Login funktioniert)

---

## Zeitschaetzung

| Tier | Wochen | Features |
|------|--------|----------|
| 1: Geld verdienen | 4 | PDF, E-Mail, Angebote, Mahnwesen, Preise |
| 2: Buchhaltung | 2 | Bank-Import, DATEV, Teilrechnungen |
| 3: Projekt-Workflow | 2 | Status-Maschine, Verfuegbarkeit, Kalender |
| 4: Lager + Scanner | 2 | CameraX, Offline, Inventur |
| 5: Crew + Personal | 2 | Zeiterfassung, CalDAV, Freelancer |
| 6: Transport + Wartung | 2 | Kapazitaet, E-Check, Versicherung |
| 7: Advanced | 4 | KI, Workflows, Federation, Notifications |
| 8: Portale + Extras | 4 | Kunden-Portal, OCR, Reports, UX |
| **Gesamt** | **~22 Wochen** | **~150 Features** |
