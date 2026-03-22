# Marktanalyse: Rental-Software für Veranstaltungstechnik

## Rentman (Cloud-SaaS, Niederlande, international)

### Überblick
- Cloud-basierte Operations-Management-Plattform für Event-Technik, AV-Rental und Media-Produktion
- Modulare Preisgestaltung: $19-25/User/Monat + $39 Plattformgebühr
- 92% Nutzerzufriedenheit (232+ Reviews)

### Kernmodule
- **Equipment (Lagerverwaltung)** — Echtzeit-Verfügbarkeit, Engpass-Erkennung, Reparatur-Tracking, Equipment-Kombinationen (Flightcases)
- **Crew Management** — Freelancer-Verwaltung, Drag&Drop-Planung, Verfügbarkeit, Zeiterfassung
- **Projektplanung** — Equipment + Crew + Transport + Finanzen pro Projekt
- **Angebote & Rechnungen** (Add-on) — Digitale Signatur, Rabattstufen, Zahlungs-Tracking
- **CRM** — Kunden/Lieferanten/Venues mit Tags, Zahlungsbedingungen, Warnungen
- **Transport & Logistik** — Fahrzeugverwaltung, Transportplanung
- **Reporting** — Equipment-Profitabilität, Umsatztrends, Auslastung

### Equipment-Tracking
- **QR-Codes** (empfohlen), **Barcodes**, **RFID** nativ unterstützt
- Items = Equipment-Typen (z.B. "Shure SM58") mit einem QR-Code
- Seriennummern = individuelle Einheiten mit eigenem QR-Code
- Teure Geräte: per Seriennummer. Bulk-Items (Kabel, Stative): per Menge
- "Scan Return" ohne Projekt-Zuordnung — System erkennt automatisch wohin

### Multi-Warehouse
- Warehouses (volle Funktionalität) und Storage Locations (einfaches Tracking)
- Interne Sub-Rentals zwischen eigenen Standorten
- Transfer-Projekte für permanente Umzüge
- Add-on: $6/Power-User/Monat pro zusätzlichem Warehouse

### Sub-Rental / Cross-Company
- **Digitale Sub-Rental-Anfragen** zwischen Rentman-Nutzern (proprietäres Netzwerk)
- Manuelle Anfragen für Nicht-Rentman-Lieferanten
- Engpass-Workflow: System erkennt Mangel → schlägt eigenes Inventar, internes Sub-Rental oder externes Sub-Rental vor

### Stärken
- Equipment + Crew kombiniert (Alleinstellungsmerkmal)
- RFID nativ (wenige Konkurrenten bieten das)
- Digitale Sub-Rental-Anfragen (ähnlich unserer Federation-Idee)
- Echtzeit-Packlisten (mehrere Mitarbeiter gleichzeitig)
- Free Basic User für Freelancer/Crew
- Verfügbarkeits-Timeline mit Doppelbuchungs-Erkennung
- REST API für eigene Integrationen

### Schwächen
- **Mobile App eingeschränkt** — keine Angebote, keine Projektänderungen
- **Steile Lernkurve** — überladene Navigation
- **Wenige Buchhaltungs-Integrationen** — nur QuickBooks, Xero, Exact Online (kein DATEV!)
- **Cloud-only** — kein Offline-Modus
- **Teuer & intransparent** — Modul-Aufpreise, pro-User-Kosten
- **Seriennummer-Tracking** hat Lücken
- **Dashboard** kaum anpassbar

---

## Easyjob (On-Premise, Deutschland, DACH-Marktführer)

### Überblick
- Full-ERP-System von Protonic Software GmbH, speziell für Veranstaltungstechnik
- Marktführer in DACH mit 2000+ Firmen
- Perpetual License + jährliche Wartungsgebühr (~4.000€ + ~1.100€/Jahr)
- Updates alle 6 Wochen

### Editionen
| Edition | Zielgruppe |
|---------|-----------|
| S | 1-Person-Betriebe |
| M | ~5 Mitarbeiter |
| L | ~15 Mitarbeiter, ~2 Standorte |
| XL | Großunternehmen, multi-site |

### Kernmodule
- **Projektplanung/-kalkulation/-management** — Volles Projekt-Lifecycle
- **Lagerverwaltung** — Hierarchisch (Standort → Raum → Regal → Fach)
- **Kontaktverwaltung** — CRM-Funktionen
- **Angebote & Rechnungen** — Integriert
- **Disposition** — Verfügbarkeitsprüfung mit Überbuchungs-Verhinderung
- **Sub-Rental-Management** — "Zumietassistent" mit automatischer Engpass-Erkennung
- **Personalverwaltung** — CrewBrain/Personio-Integration
- **Fahrzeug- & Transportmanagement**
- **Profitabilitätsplaner**
- **Raumplaner**

### Equipment-Tracking
- 1D und 2D Barcodes
- Dedizierte Scanner-App (iOS/Android/Windows) mit Kamera
- **Venue Equipment Tracker**: digitale Übergabe mit Scan, Foto, Unterschrift, Schadensdokumentation
- **Barcode Info Modul**: öffentliche WebApp mit technischen Daten pro gescanntem Barcode

### Multi-Warehouse
- Ab Edition L verfügbar
- Material-Transfers zwischen Standorten
- Verfügbarkeit pro Standort bei Buchung sichtbar
- Intelligentes Konflikt-Tool
- Hierarchische Lagerplätze mit eigenen Barcodes
- Routenoptimierte Picklisten

### Sub-Rental
- **"Zumietassistent"** erkennt Engpässe automatisch und schlägt Partner vor
- Partner-Verfügbarkeitsprüfung
- Sub-Rental-Jobs über mehrere Projekte
- Kombinierte Kauf-/Sub-Rental-Bestellungen
- Check-in/Check-out für gemietetes Equipment

### Architektur
- **Client-Server, On-Premise** (kann via IaaS gehostet werden, aber nicht cloud-native)
- Microsoft SQL Server (2016/2017/2019/2025) erforderlich
- Windows-only Desktop-Client für volle Funktionalität
- WebApi für JSON-basierte Integrationen
- Server-Komponenten: Servermanager, Mobile Device Service, Web Access Service, Print Server

### Stärken
- Volles ERP: Projekt → Kalkulation → Disposition → Lager → Rechnung
- Zumietassistent (automatische Engpass-Erkennung)
- Venue Equipment Tracker (digitale Übergabe)
- Barcode Info WebApp (öffentlich)
- Hierarchisches Lagersystem mit routenoptimierten Picklisten
- All-in-One Lizenz
- Bewährte Workflows (2000+ DACH-Firmen)

### Schwächen
- **Windows-only Desktop-Client** für volle Funktionalität
- **Nicht cloud-native**, erfordert MS SQL Server
- **Teuer** für kleine Firmen
- **Überladene UI** — zu viele Features auf einmal sichtbar
- **Mobile/WebApp** nur eingeschränkt (read-heavy)
- **Support** teuer und separat zu bezahlen
- **Updates** können Kompatibilitätsprobleme verursachen

---

## Vergleich & Unsere Positionierung

| Aspekt | Rentman | Easyjob | Unsere Software |
|--------|---------|---------|-----------------|
| Hosting | Cloud-only | On-Premise (Windows) | Self-hosted Docker (jedes OS) |
| Mobile | Eingeschränkt | Sehr eingeschränkt | Full Web-App (responsive) |
| Sub-Rental | Nur Rentman↔Rentman | Manuell + Assistent | Federation API (offen) |
| Preis | $19-25/User/Monat | ~4.000€ + Support | Self-hosted, keine Lizenzkosten |
| Buchhaltung | QuickBooks/Xero | Eigenes Modul | DATEV-Export + offene Schnittstellen |
| Offline | Nein | Ja (Desktop) | PWA mit Offline-Sync (geplant) |
| Scanner | QR/Barcode/RFID | Barcode + App | Alle Typen (USB, Android, Kamera) |
| Lernkurve | Steil | Steil | Rollenbasierte UI |

## Feature-Inspiration

### Von Rentman übernehmen
1. Equipment-Kombinationen (Flightcase mit Inhalt tracken)
2. Scan Return ohne Projektzuordnung
3. Echtzeit-Packlisten (mehrere Nutzer gleichzeitig)
4. Verfügbarkeits-Timeline mit Doppelbuchungs-Warnung
5. Free Freelancer-Zugang (eingeschränkte Rolle)

### Von Easyjob übernehmen
6. Zumietassistent (automatische Engpass-Erkennung + Partner-Vorschlag)
7. Venue Equipment Tracker (digitale Übergabe: Scan + Foto + Unterschrift)
8. Barcode Info Page (öffentliche Geräte-Info per QR-Code)
9. Hierarchisches Lager (Standort → Raum → Regal → Fach)
10. Routenoptimierte Picklisten

### Unsere Eigenentwicklungen
11. Federation API (offenes Protokoll für Cross-Company Equipment-Sharing)
12. DATEV-Export (essentiell für DACH-Markt, fehlt bei Rentman)
13. Rollenbasierte UI (Persona sieht nur relevante Features)
14. Auto-Update via UI mit Rollback-Sicherheit
15. Self-hosted mit Docker Compose (`docker compose up -d` → läuft)
