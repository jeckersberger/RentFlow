# MyRMS — Backend-Services Spezifikation

**Stand:** 20. März 2026
**Version:** 1.0

---

## Überblick

MyRMS besteht aus 17 Go-Microservices. Jeder Service ist ein eigenständiges Go-Binary, das in einem Docker-Container läuft. Alle Services folgen denselben Architekturprinzipien:

- **Hexagonal Architecture** (Ports & Adapters): Business-Logik ist von Infrastruktur getrennt
- **CQRS**: Command-Handler schreiben Events nach KurrentDB, Query-Handler lesen aus PostgreSQL
- **Event Sourcing**: Alle Zustandsänderungen werden als Events persistiert, nie direkte Datenbankupdates
- **Shared Library**: Gemeinsame Funktionen in `pkg/common/` (JWT, Events, Logging, Config, Health)

---

## Gemeinsame Architektur (alle Services)

### Go-Paketstruktur (pro Service)

```
services/<service-name>/
├── cmd/
│   └── server/
│       └── main.go                 # Entrypoint: Config laden, Server starten
├── internal/
│   ├── domain/                     # Reine Business-Logik (kein Framework, kein DB)
│   │   ├── aggregate/              # Aggregate (Zustand + Event-Anwendung)
│   │   ├── command/                # Command-Definitionen (CreateEquipment, etc.)
│   │   ├── event/                  # Domain-Event-Definitionen
│   │   └── valueobject/            # Value Objects (Price, Condition, etc.)
│   ├── application/                # Use Cases / Command Handlers
│   │   ├── command/                # Command-Handler (schreiben Events)
│   │   └── query/                  # Query-Handler (lesen aus PostgreSQL)
│   ├── infrastructure/
│   │   ├── kurrentdb/              # Event Store Adapter
│   │   ├── postgres/               # Read-Model Repositories
│   │   ├── redis/                  # Cache / Session Adapter
│   │   └── http/                   # HTTP-Handlers (Chi-Router)
│   └── projection/                 # KurrentDB-Subscription → PostgreSQL-Projektion
├── migrations/                     # SQL-Migrationen
├── Dockerfile
└── go.mod
```

### Shared Library: `pkg/common/`

```
pkg/common/
├── auth/
│   ├── jwt.go          # JWT-Validation (RS256), Claims-Extraktion
│   ├── rbac.go         # Berechtigungsprüfung: HasPermission(ctx, "equipment.write")
│   └── middleware.go   # HTTP-Middleware: RequireAuth, RequirePermission
├── events/
│   ├── client.go       # KurrentDB-Client-Wrapper (Append, Subscribe, Read)
│   ├── envelope.go     # Standard-Event-Envelope (EventID, TenantID, etc.)
│   └── schemas/        # Go-Structs für alle Domain-Events
├── errors/
│   ├── types.go        # NotFoundError, ValidationError, ConflictError, etc.
│   └── http.go         # HTTP-Fehlerantworten (RFC 7807 Problem Details)
├── logging/
│   └── logger.go       # zerolog-Wrapper mit Correlation-ID und TenantID
├── middleware/
│   ├── tenant.go       # X-Tenant-ID Header extrahieren und validieren
│   ├── requestid.go    # Correlation-ID generieren/extrahieren
│   ├── ratelimit.go    # Redis-basiertes Rate Limiting
│   └── recovery.go     # Panic-Recovery mit Logging
├── health/
│   └── handler.go      # /health/live und /health/ready Endpoints
├── config/
│   └── config.go       # Umgebungsvariablen laden (godotenv + Validierung)
└── database/
    ├── postgres.go     # pgx v5 Connection Pool
    └── migrate.go      # golang-migrate Wrapper
```

### HTTP-Framework und Bibliotheken

| Bibliothek | Verwendung |
|-----------|-----------|
| `go-chi/chi/v5` | HTTP-Router (leichtgewichtig, middleware-freundlich) |
| `jackc/pgx/v5` | PostgreSQL-Treiber (native, performant) |
| `kurrent-io/kurrentdb-client-go` | KurrentDB-Client |
| `redis/go-redis/v9` | Redis-Client |
| `rs/zerolog` | Strukturiertes Logging (JSON) |
| `golang-jwt/jwt/v5` | JWT-Verarbeitung |
| `golang-migrate/migrate/v4` | Datenbankmigrationen |
| `go-playground/validator/v10` | Request-Validierung |
| `google/uuid` | UUID-Generierung |
| `gorilla/websocket` | WebSocket (notification-service) |
| `prometheus/client_golang` | Prometheus-Metriken |
| `swaggo/swag` | OpenAPI/Swagger-Dokumentation |

---

## Service 1: auth-service (Port :8001)

### Verantwortung

JWT-Authentifizierung, RBAC-Autorisierung, Benutzerverwaltung, Mandantenverwaltung, Session-Management.

### REST-API-Endpunkte

```
POST   /api/auth/login                  Login (JWT + Refresh Token)
POST   /api/auth/refresh                JWT erneuern via Refresh Token
POST   /api/auth/logout                 Session beenden
POST   /api/auth/forgot-password        Passwort-Reset anfordern
POST   /api/auth/reset-password         Passwort zurücksetzen
POST   /api/auth/verify-mfa             MFA-Code prüfen
GET    /api/auth/me                     Aktuellen User abrufen

GET    /api/users                       Alle User des Mandanten
POST   /api/users                       Neuen User anlegen (Admin)
GET    /api/users/{id}                  User abrufen
PUT    /api/users/{id}                  User aktualisieren
DELETE /api/users/{id}                  User deaktivieren
POST   /api/users/{id}/roles            Rolle zuweisen
DELETE /api/users/{id}/roles/{role_id}  Rolle entziehen

POST   /api/invitations                 Einladungslink erstellen
GET    /api/invitations/{token}         Einladung prüfen
POST   /api/invitations/{token}/accept  Einladung annehmen

GET    /api/roles                       Alle Rollen
POST   /api/roles                       Neue Rolle erstellen
PUT    /api/roles/{id}                  Rolle bearbeiten
DELETE /api/roles/{id}                  Rolle löschen (wenn nicht Systemrolle)

GET    /api/tenants/current             Mandantendaten abrufen
PUT    /api/tenants/current             Mandantendaten aktualisieren

GET    /health/live
GET    /health/ready
GET    /metrics
```

### gRPC-API (intern, Service-zu-Service)

```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUserPermissions(GetUserPermissionsRequest) returns (GetUserPermissionsResponse);
  rpc GetTenantConfig(GetTenantConfigRequest) returns (GetTenantConfigResponse);
}
```

Alle anderen Services rufen diesen gRPC-Endpunkt auf um JWT-Tokens zu validieren, anstatt selbst die Kryptographie durchzuführen.

### JWT-Aufbau

```json
{
  "sub": "user-uuid",
  "iss": "myrms-auth-service",
  "aud": "myrms",
  "exp": 1712000000,
  "iat": 1711996400,
  "jti": "unique-token-id",
  "tenant_id": "tenant-uuid",
  "email": "marco@example.com",
  "roles": ["admin"],
  "permissions": ["equipment.read", "equipment.write", "invoice.read", ...]
}
```

**Signierung:** RS256 (asymmetrisch). Private Key nur im auth-service, Public Key verteilt.

### Event-Sourcing-Pattern

```go
// Command: LoginUser
type LoginUserCommand struct {
    TenantID  uuid.UUID
    Email     string
    Password  string
    IPAddress string
    UserAgent string
}

// Handler schreibt Event in KurrentDB
func (h *LoginCommandHandler) Handle(ctx context.Context, cmd LoginUserCommand) (*LoginResult, error) {
    user, err := h.userRepo.FindByEmail(ctx, cmd.TenantID, cmd.Email)
    if err != nil { return nil, err }

    if !bcrypt.CheckPassword(user.PasswordHash, cmd.Password) {
        // Event: LoginFailed
        h.eventStore.Append(ctx, "user-"+user.ID.String(), events.LoginFailed{...})
        return nil, ErrInvalidCredentials
    }

    // Event: UserLoggedIn
    h.eventStore.Append(ctx, "user-"+user.ID.String(), events.UserLoggedIn{
        UserID:    user.ID,
        TenantID:  cmd.TenantID,
        IPAddress: cmd.IPAddress,
        UserAgent: cmd.UserAgent,
    })

    // JWT erstellen, Refresh Token speichern
    accessToken, err := h.jwtService.CreateToken(user)
    refreshToken, err := h.tokenRepo.CreateRefreshToken(ctx, user.ID)

    return &LoginResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
```

---

## Service 2: inventory-service (Port :8002)

### Verantwortung

Equipment-Verwaltung, Kategorien, Preisregeln, Bundles, Flightcases, QR-Labels, Verfügbarkeitsabfragen.

### REST-API-Endpunkte

```
GET    /api/equipment                     Equipment-Liste (paginiert, gefiltert)
POST   /api/equipment                     Neues Equipment anlegen
GET    /api/equipment/{id}                Equipment abrufen
PUT    /api/equipment/{id}                Equipment aktualisieren
DELETE /api/equipment/{id}                Equipment als ausgemustert markieren
POST   /api/equipment/{id}/check-out      Equipment ausgeben
POST   /api/equipment/{id}/check-in       Equipment zurücknehmen
GET    /api/equipment/{id}/history        Event-History des Equipments
POST   /api/equipment/{id}/condition      Zustand aktualisieren
GET    /api/equipment/{id}/availability   Verfügbarkeit für Zeitraum prüfen
POST   /api/equipment/bulk-import         Bulk-Import via CSV/Excel
GET    /api/equipment/search              Volltext-Suche
POST   /api/equipment/{id}/label          QR-Label generieren

GET    /api/categories                    Kategorie-Baum
POST   /api/categories                    Neue Kategorie
PUT    /api/categories/{id}               Kategorie bearbeiten
DELETE /api/categories/{id}               Kategorie löschen

GET    /api/price-rules                   Alle Preisregeln
POST   /api/price-rules                   Neue Preisregel
PUT    /api/price-rules/{id}              Preisregel bearbeiten
DELETE /api/price-rules/{id}              Preisregel deaktivieren
POST   /api/price-rules/calculate         Preis für Equipment + Zeitraum berechnen

GET    /api/bundles                       Alle Bundles
POST   /api/bundles                       Neues Bundle
PUT    /api/bundles/{id}                  Bundle bearbeiten

GET    /api/flightcases                   Alle Flightcases
POST   /api/flightcases                   Neuer Flightcase

GET    /health/live
GET    /health/ready
GET    /metrics
```

### gRPC-API (intern)

```protobuf
service InventoryService {
  rpc GetEquipmentInfo(GetEquipmentInfoRequest) returns (EquipmentInfo);
  rpc CheckAvailability(CheckAvailabilityRequest) returns (AvailabilityResponse);
  rpc ReserveEquipment(ReserveEquipmentRequest) returns (ReserveResponse);
  rpc ReleaseReservation(ReleaseReservationRequest) returns (ReleaseResponse);
  rpc GetPriceForPeriod(GetPriceRequest) returns (PriceResponse);
}
```

### Verfügbarkeits-Algorithmus

```go
// Prüft ob Equipment für Zeitraum verfügbar ist
func (s *AvailabilityService) CheckAvailability(
    ctx context.Context,
    tenantID uuid.UUID,
    equipmentID uuid.UUID,
    from, to time.Time,
    quantity int,
) (*AvailabilityResult, error) {

    // 1. Cache prüfen (Redis, 5 Minuten TTL)
    cacheKey := fmt.Sprintf("avail:%s:%s:%s:%s", tenantID, equipmentID, from.Format("20060102"), to.Format("20060102"))
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }

    // 2. Gesamtbestand aus inventory_schema.equipment
    total, err := s.equipmentRepo.GetTotalQuantity(ctx, tenantID, equipmentID)

    // 3. Alle Reservierungen/Check-outs für Zeitraum aus project_schema
    // (via gRPC-Aufruf an project-service)
    reserved, err := s.projectClient.GetReservedQuantity(ctx, equipmentID, from, to)

    // 4. Defekte/Wartung aus maintenance_schema
    // (via gRPC-Aufruf an maintenance-service)
    unavailable, err := s.maintenanceClient.GetUnavailableQuantity(ctx, equipmentID, from, to)

    available := total - reserved - unavailable
    result := &AvailabilityResult{
        TotalQuantity:    total,
        ReservedQuantity: reserved,
        AvailableQuantity: available,
        IsAvailable:       available >= quantity,
    }

    // 5. Ergebnis cachen
    s.cache.Set(ctx, cacheKey, result, 5*time.Minute)

    return result, nil
}
```

---

## Service 3: project-service (Port :8003)

### Verantwortung

Projekt-Lifecycle, Kundenverwaltung, Packlisten, Equipment-Reservierungen, Doppelbuchungs-Prüfung, Kalender-Integration.

### REST-API-Endpunkte

```
GET    /api/projects                          Projekte (paginiert, gefiltert)
POST   /api/projects                          Neues Projekt
GET    /api/projects/{id}                     Projekt abrufen
PUT    /api/projects/{id}                     Projekt aktualisieren
DELETE /api/projects/{id}                     Projekt stornieren
POST   /api/projects/{id}/status              Status ändern
GET    /api/projects/{id}/packing-list        Packliste abrufen
POST   /api/projects/{id}/equipment           Equipment zur Packliste hinzufügen
PUT    /api/projects/{id}/equipment/{eq_id}   Equipment-Position aktualisieren
DELETE /api/projects/{id}/equipment/{eq_id}   Equipment aus Packliste entfernen
GET    /api/projects/{id}/availability        Verfügbarkeitsprüfung aller Equipment
POST   /api/projects/{id}/confirm             Projekt bestätigen (Reservierungen anlegen)
POST   /api/projects/{id}/duplicate           Projekt duplizieren
GET    /api/projects/{id}/timeline            Gantt-Daten
POST   /api/projects/{id}/checklist           Checkliste hinzufügen

GET    /api/customers                         Kundenliste
POST   /api/customers                         Neuer Kunde
GET    /api/customers/{id}                    Kunde abrufen
PUT    /api/customers/{id}                    Kunde aktualisieren
GET    /api/customers/{id}/projects           Projekte des Kunden
GET    /api/customers/{id}/invoices           Rechnungen des Kunden (via invoice-service)

GET    /api/calendar                          Kalender-Daten (Projekte als Ereignisse)
GET    /api/calendar.ics                      ICS-Export für CalDAV-Sync

GET    /health/live
GET    /health/ready
GET    /metrics
```

### Doppelbuchungs-Prüfung (Event-gesteuert)

```go
// Wird aufgerufen wenn ProjectEquipmentAdded-Event eingeht
func (s *ConflictDetectionService) CheckForConflicts(
    ctx context.Context,
    tenantID uuid.UUID,
    equipmentID uuid.UUID,
    from, to time.Time,
    projectID uuid.UUID,
) error {

    // Alle anderen Projekte mit diesem Equipment im Zeitraum
    conflicts, err := s.projectRepo.FindConflictingProjects(
        ctx, tenantID, equipmentID, from, to, projectID)

    for _, conflict := range conflicts {
        // Event: AvailabilityConflictDetected → KurrentDB
        s.eventStore.Append(ctx, "project-"+projectID.String(),
            events.AvailabilityConflictDetected{
                EquipmentID:  equipmentID,
                ConflictingProjectID: conflict.ID,
                From: from,
                To:   to,
            })

        // Benachrichtigung: workflow-service und notification-service reagieren auf dieses Event
    }

    return nil
}
```

---

## Service 4: scanner-service (Port :8004)

### Verantwortung

QR/Barcode/RFID-Scanning, Check-In/Out-Workflow, Offline-Sync, Scan-Hardware-Integration.

### REST-API-Endpunkte

```
POST   /api/scan                     Einzelnen Scan verarbeiten
POST   /api/scan/batch               Batch-Scan (mehrere Scans auf einmal)
POST   /api/check-in                 Equipment einchecken
POST   /api/check-out                Equipment auschecken
GET    /api/scan/history             Scan-Historie
GET    /api/scan/active-checkouts    Alle aktiven Check-outs

POST   /api/sync/offline-queue       Offline-Scans synchronisieren (von Geräten)
GET    /api/sync/status/{device_id}  Sync-Status abfragen

GET    /api/equipment/{barcode}/info Artikel-Info über Barcode (für Scanner-Display)

WebSocket: /ws/scanner/{device_id}   Echtzeit-Feedback (Scan-Bestätigung, Farb-Signal)

GET    /health/live
GET    /health/ready
GET    /metrics
```

### Scan-Verarbeitungs-Pipeline

```go
func (s *ScanProcessor) ProcessScan(ctx context.Context, req ScanRequest) (*ScanResult, error) {

    // 1. Barcode/QR-Code → Equipment-ID auflösen
    equipmentID, err := s.resolveCode(ctx, req.TenantID, req.BarcodeValue)
    if err != nil {
        return &ScanResult{Status: "unknown", Color: "red"}, nil
    }

    // 2. Equipment-Infos laden (gRPC → inventory-service)
    equipment, err := s.inventoryClient.GetEquipmentInfo(ctx, equipmentID)

    // 3. Kontext bestimmen: welcher Typ?
    switch req.ScanType {
    case "checkout":
        return s.handleCheckout(ctx, req, equipment)
    case "checkin":
        return s.handleCheckin(ctx, req, equipment)
    case "inventory":
        return s.handleInventoryScan(ctx, req, equipment)
    case "locate":
        return s.handleLocate(ctx, req, equipment)
    }

    // 4. Event schreiben: ScanCompleted → KurrentDB
    s.eventStore.Append(ctx, "scan-events", events.ScanCompleted{...})

    // 5. Echtzeit-Feedback via WebSocket an das Scanner-Gerät
    s.wsHub.Send(req.DeviceID, ScanFeedback{
        Color:   "green",
        Sound:   "beep_ok",
        Message: equipment.Name,
    })

    return result, nil
}
```

### Offline-Modus (PWA + ServiceWorker)

Der scanner-service akzeptiert Batches von Offline-Scans die während des Offline-Betriebs auf dem Gerät gepuffert wurden:

1. Gerät geht offline → Scans werden in IndexedDB gespeichert
2. Gerät geht online → ServiceWorker triggert `/api/sync/offline-queue`
3. scanner-service verarbeitet jeden Scan in Reihenfolge
4. Konflikte werden gemeldet (z.B. Equipment bereits von anderem User eingecheckt)

---

## Service 5: warehouse-service (Port :8005)

### Verantwortung

Lagerorte-Verwaltung, Warenbewegungen, Inventur, Lager-Optimierung, Bestands-Reporting.

### REST-API-Endpunkte

```
GET    /api/warehouse/locations               Lagerorte (Baum-Struktur)
POST   /api/warehouse/locations               Neuen Lagerort anlegen
PUT    /api/warehouse/locations/{id}          Lagerort bearbeiten
DELETE /api/warehouse/locations/{id}          Lagerort löschen

GET    /api/warehouse/equipment/{id}/location Wo ist dieses Equipment?
PUT    /api/warehouse/equipment/{id}/location Lagerort zuweisen

POST   /api/warehouse/movements               Warenbewegung buchen
GET    /api/warehouse/movements               Bewegungsprotokoll

GET    /api/warehouse/inventory-counts        Inventuren
POST   /api/warehouse/inventory-counts        Neue Inventur starten
GET    /api/warehouse/inventory-counts/{id}   Inventur abrufen
POST   /api/warehouse/inventory-counts/{id}/items/{eq_id} Gezählte Menge eintragen
POST   /api/warehouse/inventory-counts/{id}/complete      Inventur abschließen

GET    /api/warehouse/stats                   Lager-Statistiken
GET    /api/warehouse/optimization/suggestions Einlagerungsvorschläge (KI)

GET    /health/live
GET    /health/ready
GET    /metrics
```

---

## Service 6: invoice-service (Port :8006)

### Verantwortung

Angebots- und Rechnungserstellung, Mahnwesen, DATEV-Export, Bank-Import, Zahlungsverfolgung.

### REST-API-Endpunkte

```
GET    /api/quotes                    Angebote
POST   /api/quotes                    Neues Angebot
GET    /api/quotes/{id}               Angebot abrufen
PUT    /api/quotes/{id}               Angebot bearbeiten
POST   /api/quotes/{id}/send          Angebot versenden
POST   /api/quotes/{id}/convert       Angebot → Rechnung konvertieren
POST   /api/quotes/{id}/duplicate     Angebot duplizieren

GET    /api/invoices                  Rechnungen
POST   /api/invoices                  Neue Rechnung
GET    /api/invoices/{id}             Rechnung abrufen
PUT    /api/invoices/{id}             Rechnung bearbeiten (nur Draft)
POST   /api/invoices/{id}/send        Rechnung versenden
POST   /api/invoices/{id}/cancel      Rechnung stornieren
POST   /api/invoices/{id}/credit-note Gutschrift erstellen
POST   /api/invoices/{id}/partial     Teilrechnung erstellen
GET    /api/invoices/{id}/pdf         PDF herunterladen
POST   /api/invoices/{id}/payment     Zahlung buchen

GET    /api/invoices/open-positions   Offene Posten (Mahnliste)
POST   /api/invoices/dunning/run      Mahnlauf durchführen

GET    /api/bank/transactions         Bank-Transaktionen
POST   /api/bank/import               Kontoauszug importieren (MT940/CAMT.053)
POST   /api/bank/match                Automatischer OP-Abgleich
PUT    /api/bank/transactions/{id}/match  Manuelle Zuordnung

GET    /api/datev/export              DATEV-Export generieren
GET    /api/datev/export/invoices     Ausgangsrechnungen (DATEV-Format)
GET    /api/datev/export/costs        Eingangsrechnungen (DATEV-Format)

GET    /health/live
GET    /health/ready
GET    /metrics
```

### DATEV-Export-Logik

Der DATEV-Export erzeugt CSV-Dateien im DATEV-Buchungsstapel-Format:

```go
type DATEVExportLine struct {
    Umsatz        string    // Betrag (netto)
    Soll_Haben    string    // "S" für Soll, "H" für Haben
    WKZ           string    // Währungskennzeichen "EUR"
    Kurs          string    // Wechselkurs (leer bei EUR)
    BasisUmsatz   string    // Basiswährungsbetrag
    BasisWKZ      string
    Konto         string    // Debitorenkonto (10000-69999)
    Gegenkonto    string    // z.B. "8400" (Erlöse 19%)
    BU_Schluessel string    // Buchungsschlüssel
    Belegdatum    string    // TTMM
    Belegfeld1    string    // Rechnungsnummer
    Belegfeld2    string
    Skonto        string
    Buchungstext  string    // Kundenname + Rechnungsnummer
}
```

---

## Service 7: document-service (Port :8007)

### Verantwortung

PDF-Generierung aus Templates, Template-Verwaltung, Dokumenten-Archiv, Label-Druck (ZPL), OCR.

### REST-API-Endpunkte

```
GET    /api/documents/templates         Templates
POST   /api/documents/templates         Template anlegen
PUT    /api/documents/templates/{id}    Template bearbeiten
POST   /api/documents/templates/preview Template-Vorschau rendern

POST   /api/documents/generate          Dokument aus Template generieren
GET    /api/documents/{id}              Dokument abrufen
GET    /api/documents/{id}/download     Dokument herunterladen

POST   /api/labels/generate             QR/Barcode-Label als ZPL/PDF
POST   /api/labels/{template_id}/print  Label drucken (an Zebra-Drucker)
GET    /api/labels/templates            Label-Vorlagen

POST   /api/documents/ocr               Dokument hochladen und OCR durchführen

GET    /health/live
GET    /health/ready
GET    /metrics
```

### PDF-Generierung mit chromedp

```go
func (g *PDFGenerator) GeneratePDF(ctx context.Context, templateHTML string, data interface{}) ([]byte, error) {
    // 1. Go-Template mit Daten rendern
    tmpl, err := template.New("doc").Funcs(sprig.FuncMap()).Parse(templateHTML)
    var buf bytes.Buffer
    tmpl.Execute(&buf, data)

    // 2. chromedp: HTML → PDF (unterstützt komplexes CSS, Tabellen, Seitenumbrüche)
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    var pdfBuf []byte
    chromedp.Run(ctx,
        chromedp.Navigate("about:blank"),
        chromedp.ActionFunc(func(ctx context.Context) error {
            frameTree, _ := page.GetFrameTree().Do(ctx)
            page.SetDocumentContent(frameTree.Frame.ID, buf.String()).Do(ctx)
            return nil
        }),
        chromedp.ActionFunc(func(ctx context.Context) error {
            pdfBuf, _, _ = page.PrintToPDF().
                WithPrintBackground(true).
                WithMarginTop(0.5).
                WithMarginBottom(0.5).
                Do(ctx)
            return nil
        }),
    )

    return pdfBuf, nil
}
```

---

## Service 8: crew-service (Port :8008)

### Verantwortung

Personal- und Freelancer-Verwaltung, Qualifikationen, Verfügbarkeitsplanung, Zeiterfassung, CalDAV-Sync.

### REST-API-Endpunkte

```
GET    /api/crew                      Crew-Mitglieder
POST   /api/crew                      Neues Crew-Mitglied
GET    /api/crew/{id}                 Crew-Mitglied abrufen
PUT    /api/crew/{id}                 Crew-Mitglied aktualisieren
GET    /api/crew/{id}/availability    Verfügbarkeit prüfen
GET    /api/crew/{id}/assignments     Projektzuweisungen
GET    /api/crew/{id}/time-entries    Zeiteinträge
GET    /api/crew/{id}/calendar.ics    CalDAV-Feed (ICS)

POST   /api/crew/assignments          Crew-Mitglied zu Projekt zuweisen
PUT    /api/crew/assignments/{id}     Zuweisung aktualisieren
DELETE /api/crew/assignments/{id}     Zuweisung entfernen

POST   /api/crew/time-entries         Zeiteintrag erstellen (Check-In)
PUT    /api/crew/time-entries/{id}    Zeiteintrag aktualisieren (Check-Out)
GET    /api/crew/time-entries         Alle Zeiteinträge (für Lohnabrechnung)

GET    /api/crew/schedule             Wochenplan-Übersicht

GET    /health/live
GET    /health/ready
GET    /metrics
```

---

## Service 9: federation-service (Port :8009)

### Verantwortung

mTLS-gesicherte P2P-Kommunikation mit Partner-Instanzen, Equipment-Sharing, Sub-Rental-Workflow.

### REST-API-Endpunkte (intern, für eigene UI)

```
GET    /api/federation/partners                  Partner-Liste
POST   /api/federation/partners/invite           Partner einladen (Token generieren)
POST   /api/federation/partners/accept/{token}   Einladung annehmen
DELETE /api/federation/partners/{id}             Partnerschaft beenden

GET    /api/federation/partners/{id}/availability Equipment-Verfügbarkeit beim Partner
POST   /api/federation/sub-rentals               Sub-Rental anfragen
GET    /api/federation/sub-rentals               Sub-Rental-Anfragen
PUT    /api/federation/sub-rentals/{id}/confirm  Sub-Rental bestätigen
PUT    /api/federation/sub-rentals/{id}/complete Sub-Rental abschließen
```

### Federation-API (extern, für Partner-Instanzen, mTLS)

```
GET    /federation/v1/catalog                    Equipment-Katalog (freigegeben)
POST   /federation/v1/availability               Verfügbarkeit abfragen
POST   /federation/v1/sub-rentals                Sub-Rental-Anfrage senden
PUT    /federation/v1/sub-rentals/{id}/respond   Antwort auf Anfrage (bestätigt/abgelehnt)
POST   /federation/v1/handover                   Übergabe dokumentieren
POST   /federation/v1/return                     Rückgabe dokumentieren
```

### mTLS-Implementierung

```go
func (s *FederationServer) CreateMTLSClient(partner *FederationPartner) (*http.Client, error) {
    // Eigenes Zertifikat für diesen Partner laden
    cert, err := tls.X509KeyPair(
        []byte(partner.OurCertPEM),
        []byte(partner.OurCertKey),
    )

    // Partner-Zertifikat als CA hinzufügen
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM([]byte(partner.CertificatePEM))

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS13,
    }

    return &http.Client{
        Transport: &http.Transport{TLSClientConfig: tlsConfig},
        Timeout:   30 * time.Second,
    }, nil
}
```

---

## Service 10: maintenance-service (Port :8010)

### Verantwortung

Wartungsplanung, Wartungsaufgaben, E-Check/DGUV V3, Checklisten, Messgeräte-Import (IZYTRON.IQ).

### REST-API-Endpunkte

```
GET    /api/maintenance/tasks               Wartungsaufgaben (gefiltert)
POST   /api/maintenance/tasks               Neue Aufgabe
GET    /api/maintenance/tasks/{id}          Aufgabe abrufen
PUT    /api/maintenance/tasks/{id}          Aufgabe aktualisieren
POST   /api/maintenance/tasks/{id}/complete Aufgabe abschließen
GET    /api/maintenance/tasks/overdue       Überfällige Aufgaben
GET    /api/maintenance/tasks/upcoming      Anstehende Aufgaben (30 Tage)

GET    /api/maintenance/plans               Wartungspläne
POST   /api/maintenance/plans               Wartungsplan anlegen
PUT    /api/maintenance/plans/{id}          Plan bearbeiten
DELETE /api/maintenance/plans/{id}          Plan deaktivieren

GET    /api/echeck/records                  E-Check-Prüfprotokolle
POST   /api/echeck/records                  E-Check manuell eintragen
GET    /api/echeck/records/{id}             Prüfprotokoll abrufen
POST   /api/echeck/import/izytron           IZYTRON.IQ-Daten importieren (XML)
GET    /api/echeck/due                      Fällige E-Checks

GET    /health/live
GET    /health/ready
GET    /metrics
```

---

## Service 11: transport-service (Port :8011)

### Verantwortung

Fahrzeugverwaltung, Tourenplanung, Routenoptimierung, Transportkosten-Tracking.

### REST-API-Endpunkte

```
GET    /api/vehicles                  Fahrzeuge
POST   /api/vehicles                  Neues Fahrzeug
GET    /api/vehicles/{id}             Fahrzeug abrufen
PUT    /api/vehicles/{id}             Fahrzeug aktualisieren
GET    /api/vehicles/availability     Fahrzeug-Verfügbarkeit (Zeitraum)

GET    /api/tours                     Touren
POST   /api/tours                     Neue Tour planen
GET    /api/tours/{id}                Tour abrufen
PUT    /api/tours/{id}                Tour aktualisieren
POST   /api/tours/{id}/start          Tour starten
POST   /api/tours/{id}/complete       Tour abschließen
POST   /api/tours/{id}/optimize       Route optimieren (KI via ai-service)
GET    /api/tours/{id}/costs          Transportkosten berechnen

GET    /health/live
GET    /health/ready
GET    /metrics
```

---

## Service 12: insurance-service (Port :8012)

### Verantwortung

Versicherungspolizzen-Verwaltung, Schadensmeldungen, KI-gestützte Risikobewertung.

### REST-API-Endpunkte

```
GET    /api/insurance/policies           Versicherungen
POST   /api/insurance/policies           Neue Police anlegen
GET    /api/insurance/policies/{id}      Police abrufen
PUT    /api/insurance/policies/{id}      Police aktualisieren
GET    /api/insurance/policies/expiring  Ablaufende Policen

GET    /api/damage-reports               Schadensmeldungen
POST   /api/damage-reports               Schaden melden
GET    /api/damage-reports/{id}          Schadensmeldung abrufen
PUT    /api/damage-reports/{id}          Schadensmeldung aktualisieren
POST   /api/damage-reports/{id}/submit   An Versicherung melden

GET    /api/insurance/coverage-check/{equipment_id}  Deckungsprüfung
POST   /api/insurance/risk-analysis      KI-Risikoanalyse (via ai-service)

GET    /health/live
GET    /health/ready
GET    /metrics
```

---

## Service 13: workflow-service (Port :8013)

### Verantwortung

No-Code-Workflow-Engine, Event-getriggerte Automatisierungen, CRON-Workflows, Genehmigungsprozesse.

### REST-API-Endpunkte

```
GET    /api/workflows                  Workflows
POST   /api/workflows                  Neuen Workflow erstellen
GET    /api/workflows/{id}             Workflow abrufen
PUT    /api/workflows/{id}             Workflow bearbeiten
DELETE /api/workflows/{id}             Workflow löschen
POST   /api/workflows/{id}/activate    Workflow aktivieren
POST   /api/workflows/{id}/deactivate  Workflow deaktivieren
POST   /api/workflows/{id}/test        Workflow manuell testen
GET    /api/workflows/{id}/executions  Ausführungshistorie

GET    /api/workflows/templates        Workflow-Vorlagen
POST   /api/workflows/from-template    Aus Vorlage erstellen

GET    /health/live
GET    /health/ready
GET    /metrics
```

### Workflow-Execution-Engine

```go
// KurrentDB Catch-up Subscription auf $all
func (e *WorkflowEngine) StartSubscription(ctx context.Context) {
    sub, _ := e.eventStore.SubscribeToAll(ctx, kurrentdb.SubscribeToAllOptions{
        From: kurrentdb.Start{},
    })

    for event := range sub.Channel() {
        // Alle aktiven Workflows prüfen
        workflows, _ := e.workflowRepo.FindByTrigger(ctx, event.EventType)

        for _, workflow := range workflows {
            // Bedingungen prüfen
            if e.evaluateConditions(workflow.Conditions, event.Data) {
                // Asynchrone Ausführung
                go e.executeWorkflow(ctx, workflow, event)
            }
        }
    }
}

// Verfügbare Aktionen
var WorkflowActions = map[string]ActionHandler{
    "send_notification": &SendNotificationAction{},
    "send_email":        &SendEmailAction{},
    "create_task":       &CreateMaintenanceTaskAction{},
    "update_status":     &UpdateStatusAction{},
    "create_invoice":    &CreateInvoiceAction{},
    "webhook":           &WebhookAction{},
    "approve_required":  &ApprovalRequiredAction{},
}
```

---

## Service 14: ai-service (Port :8014)

### Verantwortung

Multi-Provider-KI-Schnittstelle, Preisoptimierung, Nachfrageprognose, Asset-Erkennung aus Fotos, Anonymisierung.

### REST-API-Endpunkte

```
GET    /api/ai/providers                   Konfigurierte KI-Provider
POST   /api/ai/providers                   Provider konfigurieren
PUT    /api/ai/providers/{id}              Provider-Einstellungen
POST   /api/ai/providers/{id}/test         Verbindung testen

POST   /api/ai/price-optimization          Preisoptimierungsvorschläge
POST   /api/ai/demand-forecast             Nachfrageprognose
POST   /api/ai/asset-recognition           Equipment aus Foto erkennen
POST   /api/ai/maintenance-prediction      Wartungsvorhersage
POST   /api/ai/transport-optimization      Routen-/Kostenoptimierung
POST   /api/ai/insurance-risk              Versicherungsrisiko bewerten
POST   /api/ai/pack-optimization           Packoptimierung für Flightcases
POST   /api/ai/dunning-prediction          Zahlungswahrscheinlichkeit

POST   /api/ai/anonymize                   Text anonymisieren (vor Cloud-API-Aufruf)

GET    /api/ai/usage                       Token-/Kosten-Tracking
GET    /api/ai/feedback                    Feedback-Übersicht

GET    /health/live
GET    /health/ready
GET    /metrics
```

### Provider-Abstraktion

```go
type AIProvider interface {
    Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error)
    Name() string
}

// Implementierungen für:
type AnthropicProvider struct { client *anthropic.Client }
type OpenAIProvider struct { client *openai.Client }
type GoogleProvider struct { client *google.Client }
type MistralProvider struct { client *mistral.Client }
type OllamaProvider struct { baseURL string }  // Self-hosted

// Router: Task-Typ → optimaler Provider
func (r *ProviderRouter) RouteTask(task AITaskType) AIProvider {
    switch task {
    case PriceOptimization:
        return r.getProvider("anthropic")  // claude-3-5-sonnet
    case AssetRecognition:
        return r.getProvider("openai")     // gpt-4o (Vision)
    case DemandForecast:
        return r.getProvider("google")     // gemini-1.5-pro
    case LocalTask:
        return r.getProvider("ollama")     // llama3 (self-hosted, keine Daten-Übertragung)
    }
}
```

---

## Service 15: notification-service (Port :8015)

### Verantwortung

In-App-Benachrichtigungen, E-Mail-Versand, Push-Notifications, User-Präferenzen.

### REST-API-Endpunkte

```
GET    /api/notifications              Eigene Benachrichtigungen
POST   /api/notifications/{id}/read    Als gelesen markieren
POST   /api/notifications/read-all     Alle als gelesen
DELETE /api/notifications/{id}         Löschen

GET    /api/notifications/preferences  Eigene Präferenzen
PUT    /api/notifications/preferences  Präferenzen aktualisieren

WebSocket: /ws/notifications           Echtzeit-Benachrichtigungen

GET    /health/live
GET    /health/ready
GET    /metrics
```

### Event-Subscription für Benachrichtigungen

```go
// notification-service subscribt auf $all und transformiert Domain-Events in Notifications
var NotificationRules = map[string]NotificationRule{
    "invoice.overdue": {
        Title:     "Rechnung überfällig: {{.InvoiceNumber}}",
        Body:      "{{.CustomerName}}: {{.Amount}} EUR seit {{.DaysSinceDue}} Tagen überfällig",
        Channels:  []string{"in_app", "email"},
        Roles:     []string{"admin", "accountant"},
    },
    "equipment.checked_out": {
        Title:     "Equipment ausgegeben: {{.EquipmentName}}",
        Body:      "{{.ProjectName}} — ausgegeben von {{.UserName}}",
        Channels:  []string{"in_app"},
        Roles:     []string{"admin", "project_manager"},
    },
    "maintenance.overdue": {
        Title:     "Wartung überfällig: {{.EquipmentName}}",
        Body:      "{{.TaskTitle}} war fällig am {{.DueDate}}",
        Channels:  []string{"in_app", "email"},
        Roles:     []string{"admin", "warehouse"},
    },
}
```

---

## Service 16: reporting-service (Port :8016)

### Verantwortung

CQRS Read-Side für Dashboards, KPI-Berechnungen, Business-Intelligence, Export.

### REST-API-Endpunkte

```
GET    /api/dashboard                  Dashboard-Daten (rollenbasiert)
GET    /api/dashboard/config           Dashboard-Konfiguration
PUT    /api/dashboard/config           Dashboard anpassen (Widgets)

GET    /api/reports/revenue            Umsatz-Report (Zeitraum, Kunde, Kategorie)
GET    /api/reports/utilization        Auslastungs-Report (Equipment)
GET    /api/reports/equipment-value    Inventarwert-Report
GET    /api/reports/crew-hours         Personaleinsatz-Report
GET    /api/reports/projects           Projektlisten-Report
GET    /api/reports/open-positions     Offene-Posten-Report
GET    /api/reports/co2                CO2-/Nachhaltigkeits-Report (ESG)

POST   /api/reports/export             Report als CSV/XLSX/PDF exportieren

GET    /api/kpis                       Live-KPIs
GET    /api/kpis/history               KPI-Verlauf (für Charts)

GET    /health/live
GET    /health/ready
GET    /metrics
```

### CQRS Read-Model-Aufbau

reporting-service subscribt auf `$all` in KurrentDB und pflegt seine eigenen Materialized Views:

```go
// Subscription-Handler
func (p *ReportingProjection) HandleEvent(event kurrentdb.Event) {
    switch event.EventType {
    case "invoice.paid":
        p.updateRevenueKPI(event)
        p.updateCustomerRevenue(event)
    case "equipment.checked_out":
        p.updateEquipmentUtilization(event)
    case "project.completed":
        p.updateProjectStats(event)
        p.snapshotKPIs(event.TenantID, time.Now())
    }
}
```

---

## Service 17: audit-service (Port :8017)

### Verantwortung

GoBD-konformer, unveränderlicher Audit-Trail aller Events aller Services, Checksummen-Kette.

### REST-API-Endpunkte

```
GET    /api/audit/log                  Audit-Log (gefiltert, paginiert)
GET    /api/audit/log/{entity_type}/{id} Alle Events einer Entität
GET    /api/audit/verify               Checksummen-Kette verifizieren
GET    /api/audit/export               Audit-Log exportieren (GoBD-konformes Format)

GET    /health/live
GET    /health/ready
GET    /metrics
```

### Unveränderlichkeits-Garantie

```go
// audit-service subscribt auf $all und schreibt JEDEN Event in die audit_log-Tabelle
// Der Datenbankuser des audit-service hat NUR INSERT-Rechte auf audit_log
// UPDATE und DELETE sind auf Datenbankebene verboten

-- PostgreSQL: Nur INSERT erlaubt
REVOKE UPDATE, DELETE ON audit_schema.audit_log FROM audit_service_user;
GRANT INSERT, SELECT ON audit_schema.audit_log TO audit_service_user;

// SHA-256-Kette: Jeder Eintrag enthält Hash des vorherigen
func (r *AuditLogRepo) Append(ctx context.Context, event AuditEntry) error {
    lastChecksum, _ := r.getLastChecksum(ctx, event.TenantID)
    payload, _ := json.Marshal(event.Payload)

    checksum := sha256.Sum256([]byte(
        event.EventID.String() + string(payload) + lastChecksum,
    ))

    _, err := r.db.Exec(ctx, `
        INSERT INTO audit_schema.audit_log
        (log_uuid, tenant_id, event_id, event_type, aggregate_type, aggregate_id,
         user_id, source_service, payload, checksum, prev_checksum, sequence_number, recorded_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
                (SELECT COALESCE(MAX(sequence_number), 0) + 1 FROM audit_schema.audit_log WHERE tenant_id = $2),
                NOW())`,
        event.LogUUID, event.TenantID, ..., hex.EncodeToString(checksum[:]), lastChecksum)

    return err
}
```

---

## Traefik v3 — Routing-Konfiguration

### Docker-Labels pro Service

```yaml
# Beispiel: inventory-service
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.inventory.rule=Host(`${DOMAIN}`) && PathPrefix(`/api/equipment`, `/api/categories`, `/api/price-rules`, `/api/bundles`, `/api/flightcases`, `/api/labels`)"
  - "traefik.http.routers.inventory.entrypoints=websecure"
  - "traefik.http.routers.inventory.tls.certresolver=letsencrypt"
  - "traefik.http.routers.inventory.middlewares=auth-headers,rate-limit,compress"
  - "traefik.http.services.inventory.loadbalancer.server.port=8002"
  - "traefik.http.services.inventory.loadbalancer.healthcheck.path=/health/ready"
  - "traefik.http.services.inventory.loadbalancer.healthcheck.interval=10s"
```

### Middleware-Definitionen

```yaml
# traefik-middlewares.yml (oder via Docker-Labels auf Traefik-Container)

# Rate Limiting (via Redis)
http:
  middlewares:
    rate-limit:
      rateLimit:
        burst: 50
        average: 20        # 20 Requests/Sekunde

    auth-headers:
      headers:
        customRequestHeaders:
          X-Request-ID: ""       # Wird von Traefik generiert wenn leer
        customResponseHeaders:
          X-Content-Type-Options: "nosniff"
          X-Frame-Options: "DENY"
          Strict-Transport-Security: "max-age=31536000; includeSubDomains"
          Content-Security-Policy: "default-src 'self'"

    compress:
      compress:
        minResponseBodyBytes: 1024
```

### Routing-Tabelle (alle Services)

| Service | Path-Prefix(e) |
|---------|----------------|
| auth-service | `/api/auth`, `/api/users`, `/api/roles`, `/api/invitations`, `/api/tenants` |
| inventory-service | `/api/equipment`, `/api/categories`, `/api/price-rules`, `/api/bundles`, `/api/flightcases`, `/api/labels` |
| project-service | `/api/projects`, `/api/customers`, `/api/calendar` |
| scanner-service | `/api/scan`, `/api/check-in`, `/api/check-out`, `/api/sync` |
| warehouse-service | `/api/warehouse` |
| invoice-service | `/api/quotes`, `/api/invoices`, `/api/bank`, `/api/datev` |
| document-service | `/api/documents`, `/api/templates`, `/api/labels` |
| crew-service | `/api/crew` |
| federation-service | `/api/federation`, `/federation/v1` |
| maintenance-service | `/api/maintenance`, `/api/echeck` |
| transport-service | `/api/vehicles`, `/api/tours` |
| insurance-service | `/api/insurance`, `/api/damage-reports` |
| workflow-service | `/api/workflows` |
| ai-service | `/api/ai` |
| notification-service | `/api/notifications`, `/ws/notifications` |
| reporting-service | `/api/reports`, `/api/dashboard`, `/api/kpis` |
| audit-service | `/api/audit` |

---

## Service-zu-Service-Kommunikation

### Synchrone Kommunikation (gRPC)

Direkte Abhängigkeiten werden über gRPC gelöst, wenn ein sofortiger Rückgabewert benötigt wird:

```
inventory-service ──gRPC──► auth-service        (Token validieren)
project-service   ──gRPC──► inventory-service   (Verfügbarkeit prüfen)
scanner-service   ──gRPC──► inventory-service   (Equipment-Info per Barcode)
scanner-service   ──gRPC──► project-service     (Projekt-Kontext prüfen)
invoice-service   ──gRPC──► project-service     (Packliste für Rechnung)
document-service  ──gRPC──► invoice-service     (Rechnungsdaten für PDF)
document-service  ──gRPC──► project-service     (Projektdaten für Lieferschein)
transport-service ──gRPC──► ai-service          (Routenoptimierung)
```

### Asynchrone Kommunikation (KurrentDB Events)

Lose Kopplung über Events für alle anderen Abhängigkeiten:

```
inventory-service appended  →  warehouse-service, reporting-service, audit-service
project-service appended    →  inventory-service (Reservierung), notification-service, audit-service
scanner-service appended    →  inventory-service (Status), warehouse-service, audit-service
invoice-service appended    →  notification-service (Mahnung fällig), workflow-service, audit-service
maintenance-service appended → inventory-service (Status auf "Maintenance"), notification-service
```

---

## Authentifizierung und Autorisierung

### JWT-Flow

```
Client → POST /api/auth/login
       ← { access_token (15min), refresh_token (30 Tage) }

Client → GET /api/equipment  (Header: Authorization: Bearer <access_token>)
Traefik → Weiterleitung an inventory-service (Header unverändert)
inventory-service → Middleware: JWT validieren (Public Key)
                 → Tenant-ID aus Token extrahieren
                 → Permissions prüfen: HasPermission("equipment.read")
                 ← 200 OK + Daten
```

### Berechtigungs-Modell (RBAC)

Permissions sind fein granular und werden beim Login in den JWT gepackt:

```
equipment.read, equipment.write, equipment.delete
project.read, project.write, project.confirm
invoice.read, invoice.write, invoice.send, invoice.approve
warehouse.read, warehouse.write
crew.read, crew.write, crew.assign
scanner.use
federation.read, federation.manage
maintenance.read, maintenance.write
admin.users, admin.roles, admin.system
reports.read
audit.read
```

### Multi-Mandanten-Isolation

Jeder Request enthält die `tenant_id` aus dem JWT. Alle Datenbankabfragen filtern automatisch nach `tenant_id`:

```go
// PostgreSQL Row Level Security
ALTER TABLE inventory_schema.equipment ENABLE ROW LEVEL SECURITY;

CREATE POLICY equipment_tenant_isolation ON inventory_schema.equipment
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- In jedem Request: SET LOCAL app.current_tenant_id = 'tenant-uuid-aus-jwt'
```

---

## Observability

### Prometheus-Metriken (alle Services)

```go
// Standard-Metriken pro Service:
var (
    httpRequestsTotal = prometheus.NewCounterVec(...)     // request_total{method, path, status}
    httpDuration      = prometheus.NewHistogramVec(...)   // request_duration_seconds
    dbQueryDuration   = prometheus.NewHistogramVec(...)   // db_query_duration_seconds
    eventStoreAppends = prometheus.NewCounterVec(...)     // eventstore_appends_total{stream}
    activeWebSockets  = prometheus.NewGauge(...)          // active_websocket_connections
)
```

### Structured Logging (zerolog)

```json
{
  "level": "info",
  "service": "inventory-service",
  "tenant_id": "uuid",
  "user_id": "uuid",
  "correlation_id": "uuid",
  "method": "POST",
  "path": "/api/equipment",
  "status": 201,
  "duration_ms": 45,
  "time": "2026-03-20T10:30:00Z"
}
```

### Health-Check-Endpoints

Alle Services implementieren:
- `GET /health/live` → HTTP 200 wenn Prozess läuft (für Liveness Probe)
- `GET /health/ready` → HTTP 200 wenn DB + KurrentDB erreichbar (für Readiness Probe)
- `GET /metrics` → Prometheus-Metriken
