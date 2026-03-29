# CrateDesk Roadmap Q2 2026

> Erstellt: 29.03.2026 — Basiert auf vollstaendiger Spec-Analyse

## Prinzipien
- Jede Woche muss ein deploybares Ergebnis liefern
- Business-Value first: Was Janis JETZT braucht um Geld zu verdienen
- Kein Overengineering: Event Sourcing/KurrentDB kommt spaeter, CRUD reicht erstmal
- PDF + E-Mail + Bank = Kern-Business, alles andere ist nice-to-have

---

## Sprint 1: Document Engine + PDF (Woche 1)
**Ziel: Rechnungen/Angebote als PDF drucken und per Mail senden**

### Backend (document-service + invoice-service)
- [ ] HTML-Template-System mit Go html/template
- [ ] 8 Default-Templates: RE, AN, LS, AB, MA, PL, MV, Etiketten
- [ ] Platzhalter-System: {{.Kunde.Name}}, {{range .Positionen}}, etc.
- [ ] PDF-Generierung via chromedp (HTML → PDF)
- [ ] GET /api/v1/invoices/{id}/pdf
- [ ] GET /api/v1/documents/templates (Liste aller Templates)
- [ ] PUT /api/v1/documents/templates/{id} (Template bearbeiten)
- [ ] Template-Felder: Logo-URL, Primaerfarbe, Schriftart, Kopfzeile, Fusszeile, Impressum

### Frontend
- [ ] Template-Editor Seite (/settings/templates)
- [ ] Live-Preview (HTML rendered im iframe)
- [ ] Anpassbar: Logo hochladen, Farben, Texte, Spalten ein/aus
- [ ] PDF-Download Button auf Rechnungs-Detail
- [ ] PDF-Preview Modal

### Infrastruktur
- [ ] chromedp oder wkhtmltopdf im document-service Container
- [ ] /uploads Volume fuer Logos

---

## Sprint 2: E-Mail + Mahnwesen (Woche 2)
**Ziel: Rechnungen/Mahnungen per E-Mail versenden**

### Backend (notification-service)
- [ ] SMTP-Client (gomail)
- [ ] E-Mail-Templates (HTML) fuer: Rechnung, Mahnung 1/2/3, Angebot, Einladung
- [ ] POST /api/v1/invoices/{id}/send (PDF als Anhang)
- [ ] Mahnwesen: 3 Stufen (Erinnerung nach 7d, Mahnung nach 14d, letzte Mahnung nach 28d)
- [ ] Mahnstatus in invoice-service tracken
- [ ] POST /api/v1/invoices/{id}/remind

### Frontend
- [ ] "Per E-Mail senden" Button auf Rechnung
- [ ] Mahnungs-Uebersicht (faellige Rechnungen)
- [ ] SMTP-Einstellungen in Settings
- [ ] E-Mail-Vorschau vor dem Senden

---

## Sprint 3: Preiskalkulation + Angebot→Rechnung Flow (Woche 3)
**Ziel: Kompletter Verkaufsprozess**

### Backend (inventory-service + invoice-service)
- [ ] Preiskalkulations-Engine: Tagespreis, Wochenpreis, Staffelrabatte
- [ ] Mengenrabatt-Regeln pro Kategorie
- [ ] Saisonzuschlaege (konfigurierbar)
- [ ] Angebot erstellen (invoice_type: 'quote')
- [ ] Angebot → Auftragsbestaetigung → Rechnung Konvertierung
- [ ] POST /api/v1/invoices/{id}/convert (quote→invoice)
- [ ] Teilrechnungen + Restbetrag-Tracking

### Frontend
- [ ] Angebots-Erstellung mit Equipment-Auswahl
- [ ] Preis-Kalkulator (Zeitraum × Tagespreis × Rabatt)
- [ ] "In Rechnung umwandeln" Button
- [ ] Teilrechnung erstellen

---

## Sprint 4: Bank-Import + DATEV (Woche 4)
**Ziel: Zahlungsabgleich und Steuer-Export**

### Backend (invoice-service)
- [ ] CSV-Import fuer Kontoauszuege (Sparkasse, VR, Deutsche Bank Formate)
- [ ] MT940-Parser (SWIFT-Format)
- [ ] Automatischer OP-Abgleich (Rechnungsnummer in Verwendungszweck matchen)
- [ ] DATEV-Export (SKR03 CSV)
- [ ] WISO-Export
- [ ] GET /api/v1/bank/import (Upload + Parse)
- [ ] POST /api/v1/bank/match (Vorschlaege bestaetigen)
- [ ] GET /api/v1/invoices/datev-export?from=&to=

### Frontend
- [ ] Bank-Import Seite (CSV/MT940 Upload)
- [ ] Matching-UI: Transaktion ↔ Rechnung zuordnen
- [ ] DATEV-Export Button in Rechnungsliste
- [ ] Offene-Posten-Uebersicht

---

## Sprint 5: Customizable Document Templates (Woche 5)
**Ziel: Jeder Dokumenttyp visuell anpassbar**

### Frontend (Template-Editor erweitern)
- [ ] Drag & Drop Sektionen (Kopf, Positionen, Fuss, Summen)
- [ ] Spalten-Konfigurator (welche Spalten in Positionstabelle)
- [ ] Bedingte Bloecke (z.B. Kleinunternehmer-Hinweis nur wenn aktiv)
- [ ] Template pro Dokumenttyp: RE, AN, LS, AB, MA, PL, MV
- [ ] Template-Vorschau mit echten Daten
- [ ] Logo + Briefkopf Editor
- [ ] Schriftart-Auswahl (5-6 Optionen)
- [ ] Farb-Picker fuer Akzentfarbe

---

## Sprint 6: Verfuegbarkeit + Kalender (Woche 6)
**Ziel: Doppelbuchungen verhindern, Projektuebersicht**

### Backend
- [ ] Verfuegbarkeitscheck-Algorithmus (Bestand - Reserviert - Wartung)
- [ ] Konflikt-Erkennung bei Equipment-Zuweisung
- [ ] Kalender-API: GET /api/v1/projects/calendar?from=&to=
- [ ] Equipment-Auslastungs-API

### Frontend
- [ ] Kalender-Ansicht fuer Projekte (Monats/Wochen-View)
- [ ] Equipment-Verfuegbarkeits-Timeline
- [ ] Warnung bei Doppelbuchung
- [ ] Auslastungs-Dashboard

---

## Sprint 7: Scanner + Offline (Woche 7)
**Ziel: Equipment scannen auf dem Handy**

### Scanner-App (Kotlin)
- [ ] CameraX Barcode/QR-Scanning
- [ ] Offline-Queue (Room DB)
- [ ] Background-Sync bei Reconnect
- [ ] Foto bei Check-In (Schadensdoku)

### Backend (scanner-service)
- [ ] GET /api/v1/scan/resolve/{identifier} (Barcode→Equipment)
- [ ] PUT /api/v1/scan/equipment/{id}/rfid (RFID zuordnen)
- [ ] Scanner-Sessions mit Protokoll
- [ ] Digitale Unterschrift bei Uebergabe

---

## Sprint 8: Crew + Zeiterfassung (Woche 8)
**Ziel: Mitarbeiter planen und Zeiten erfassen**

- [ ] Crew-Zuordnung zu Projekten
- [ ] Zeiterfassung Start/Stop
- [ ] Qualifikationen verwalten
- [ ] Freelancer-Ansicht (eigene Einsaetze)
- [ ] CalDAV-Feed fuer Kalender-Apps

---

## Spaeter (nach Q2)
- Event Sourcing / KurrentDB
- KI Multi-Provider
- Federation (P2P Equipment-Sharing)
- Workflow-Engine (No-Code)
- Kunden-Portal
- PWA Offline komplett
- WebSocket Echtzeit
- Notification-Center
- DGUV V3 / E-Check Integration
- NAS-Speicher-Anbindung
- i18n (EN)
- Gantt-Timeline
- Report-Diagramme
