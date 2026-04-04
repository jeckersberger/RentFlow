# CrateDesk — Tester ↔ Entwickler Kommunikation

## Anleitung
- **Tester**: Bugs hier eintragen (Format unten)
- **Entwickler**: Liest diese Datei, fixt Bugs, markiert als FIXED

## Bug-Format
```
### BUG-XXX: Kurzbeschreibung
- **Endpunkt**: GET/POST/... /api/v1/...
- **Erwartet**: Was sollte passieren
- **Tatsächlich**: Was passiert stattdessen
- **Status**: OPEN / FIXED
```

## Aktive Bugs

### BUG-001: 11/19 API-Routes geben HTML statt JSON (CRITICAL)
- **Status**: FIXED — War stale Traefik-Config nach Redeploy. Alle Routes liefern jetzt JSON.

### BUG-002: DATEV-Export Route kaputt (CRITICAL)
- **Status**: FIXED — Alias-Route /invoices/datev-export hinzugefügt.

### BUG-003: Invoice Finalize gibt 405 (CRITICAL)
- **Status**: FIXED ��� POST /finalize als Alias für PATCH hinzugefügt.

### BUG-004: 18 Config-Endpoints geben 404 (CRITICAL)
- **Status**: FIXED — Config GET gibt jetzt leeren Default {key, value: {}} zurueck statt 404. Setup/status Endpoint hinzugefuegt.

### BUG-005: Rechnungen-Frontend zeigt keine Daten (CRITICAL)
- **Status**: FIXED — Field-Mapping: invoice_number→number, customer_name→client_name, invoice_date→issue_date, total_net→subtotal, total_gross→total; alle /100 für Cents.

### BUG-006: Budget Faktor 100 falsch (HIGH)
- **Status**: FIXED — Budget wird jetzt /100 konvertiert.

### BUG-007: Angebote zeigen "NaN EUR" (HIGH)
- **Status**: FIXED — total_net→sub_total Mapping + /100 Konvertierung.

### BUG-008: Angebots-Items werden ignoriert (HIGH)
- **Status**: BY DESIGN — Items muessen per POST /quotes/{id}/items einzeln hinzugefuegt werden (wie bei Invoices). Frontend muss Items nach Quote-Create posten.

### BUG-009: Projekt-Detail nicht erreichbar (HIGH)
- **Endpunkt**: UI /projects — Klick auf Tabellenzeile
- **Erwartet**: Navigation zu /projects/{id}
- **Tatsaechlich**: Nichts passiert, URL aendert sich nicht
- **Status**: OPEN

### BUG-010: Kundensuche filtert nicht (HIGH)
- **Status**: FIXED — search Parameter im List-Endpoint implementiert (ILIKE auf Name/Email).

### BUG-011: Umlaute werden zu Fragezeichen (HIGH)
- **Status**: NOT A BUG — API ist UTF-8 korrekt (verifiziert direkt + via Cloudflare). Problem war Windows-Terminal-Encoding beim curl-Test. Bereits gespeicherte kaputte Daten muessen manuell korrigiert werden.

### BUG-012: Projekt Status-Update via PUT ignoriert (HIGH)
- **Status**: FIXED — Status-Feld zum UpdateProjectRequest hinzugefuegt, wird jetzt bei PUT verarbeitet.

### BUG-013: Ungueltiger Status wird akzeptiert (HIGH)
- **Status**: FIXED — ValidateStatus() prueft jetzt den Status-Wert bei Update, gibt 400 bei ungueltigem Status.

### BUG-014: Invoice DELETE nicht implementiert (HIGH)
- **Status**: FIXED — DELETE /invoices/{id} implementiert (nur Drafts löschbar).

### BUG-015: Pagination "Seite 1 von 0" (MEDIUM)
- **Status**: FIXED — total wird jetzt aus meta.total gelesen.

### BUG-016: Kundennamen fehlen ueberall (MEDIUM)
- **Endpunkt**: UI /projects, /invoices, /quotes, /calendar
- **Erwartet**: Kundenname in Tabellenspalte
- **Tatsaechlich**: "—" oder leer ueberall
- **Status**: OPEN

### BUG-017: Kalender zeigt Event nicht im Grid (MEDIUM)
- **Endpunkt**: UI /calendar — Konzert Frequency am 5.4.2026
- **Erwartet**: Event im Kalender-Grid auf dem 5. sichtbar
- **Tatsaechlich**: Nur in Seitenliste, fehlt im Grid
- **Status**: OPEN

### BUG-018: Scanner "Noch -1 Tage" (MEDIUM)
- **Endpunkt**: Scanner-App Check-Out Screen
- **Erwartet**: "Noch 1 Tag" (Projekt startet morgen 5.4.)
- **Tatsaechlich**: "Noch -1 Tage" — negative Tage
- **Status**: OPEN

### BUG-019: Scanner Back-Stack zu flach (MEDIUM)
- **Endpunkt**: Scanner-App Navigation
- **Erwartet**: Zurueck-Button geht zum Hauptmenue
- **Tatsaechlich**: Doppeltes Zurueck navigiert zum Login (Session verloren)
- **Status**: OPEN

### BUG-020: Scanner Check-In zeigt alle Projekte (MEDIUM)
- **Endpunkt**: Scanner-App Check-In Screen
- **Erwartet**: Nur Projekte mit heutiger Rueckgabe
- **Tatsaechlich**: Alle Projekte (inkl. Testdaten)
- **Status**: OPEN

### BUG-021: Scanner rohe ISO-Timestamps (MEDIUM)
- **Endpunkt**: Scanner-App Check-In Screen
- **Erwartet**: "05.04.2026" (formatiert)
- **Tatsaechlich**: "2026-04-05T00:00:00Z" (roher ISO-String)
- **Status**: OPEN

### BUG-022: Invoice Item vat_rate nicht gespeichert (MEDIUM)
- **Status**: FIXED — vat_rate Feld zu InvoiceItem hinzugefuegt (Model+DTO+Repo+Migration). Default: Invoice-Level vat_rate.

### BUG-023: Scanner Adressfelder Mismatch (MEDIUM)
- **Endpunkt**: Scanner sendet address_street, API nutzt billing_address_street
- **Erwartet**: Adresse korrekt gespeichert
- **Tatsaechlich**: Adresse leer (Feldnamen stimmen nicht ueberein)
- **Status**: OPEN

### BUG-024: Security Headers fehlen (MEDIUM)
- **Endpunkt**: Alle API-Responses
- **Erwartet**: HSTS, X-Frame-Options, CSP, Referrer-Policy
- **Tatsaechlich**: Keine Security Headers in Responses (Cloudflare leitet sie nicht weiter)
- **Status**: OPEN

### BUG-025: Benutzer-Rollen leer (MEDIUM)
- **Endpunkt**: UI /settings/users
- **Erwartet**: Rolle pro Benutzer (Admin, etc.)
- **Tatsaechlich**: Rollen-Spalte bei allen 3 Usern komplett leer
- **Status**: OPEN

### BUG-026: crew/members und documents/templates Route (MEDIUM)
- **Status**: FIXED — /members und /templates Alias-Routes hinzugefügt.

### BUG-027: 500 Error bei langen Namen (LOW)
- **Endpunkt**: POST /api/v1/projects mit 2000-Zeichen Name
- **Erwartet**: 400 "Name zu lang"
- **Tatsaechlich**: 500 Internal Server Error
- **Status**: OPEN

### BUG-028: 500 Error bei grossen Zahlen (LOW)
- **Endpunkt**: POST /api/v1/projects mit budget=99999999999999
- **Erwartet**: 400 Validation Error
- **Tatsaechlich**: 500 Internal Server Error (int64 overflow)
- **Status**: OPEN

### BUG-029: Negative Budgets/Mengen akzeptiert (LOW)
- **Endpunkt**: POST /api/v1/projects mit budget=-50000, POST /api/v1/invoices mit quantity=-5
- **Erwartet**: 400 "Wert muss positiv sein"
- **Tatsaechlich**: Akzeptiert ohne Fehler
- **Status**: OPEN

### BUG-030: Datumsformat nur YYYY-MM-DD (LOW)
- **Endpunkt**: POST /api/v1/projects mit start_date="10.04.2026"
- **Erwartet**: DD.MM.YYYY akzeptiert (europaeisches Format)
- **Tatsaechlich**: 400 "Ungueltiges Startdatum" — nur YYYY-MM-DD funktioniert
- **Status**: OPEN

### BUG-031: Kalender leere Klammern (LOW)
- **Endpunkt**: UI /calendar
- **Erwartet**: "Firmengala Eventhaus (Eventhaus Muenchen GmbH)"
- **Tatsaechlich**: "Firmengala Eventhaus ()" — Kundenname fehlt in Klammern
- **Status**: OPEN

### BUG-032: Software-Version leer (LOW)
- **Endpunkt**: UI /settings/update
- **Erwartet**: Aktuelle Versionsnummer
- **Tatsaechlich**: "v..." (leer), "Unbekannt"
- **Status**: OPEN

## Server-Info
- **URL**: https://cratedesk.je-soundulight.de
- **Login**: t@t.de / test1234 / tenant: je-soundulight
- **Deployed**: de24c89 (04.04.2026)
- **Alle 19 Services**: Running
