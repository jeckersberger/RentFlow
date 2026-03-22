# Detailspezifikation: Project, Invoice & Crew Domain (Module P1-P8, F1-F8, H1-H7)

**Status:** Implementierungsreif | **Version:** 1.0 | **Datum:** 2026-03-21

---

## 1. PROJECT DOMAIN MODEL

### 1.1 Projekt-Lifecycle State Machine

```
Draft ──[confirm]──> Quoted ──[accept]──> Confirmed
                              ↓
                           [reject]
                              ↓
                           Rejected

Confirmed ──[prepare]──> InPreparation ──[dispatch]──> LoadingOut
                                                            ↓
                                                      [loaded]
                                                            ↓
                                                        OnSite
                                                            ↓
                                                      [complete_onsite]
                                                            ↓
                                                       LoadingIn
                                                            ↓
                                                      [equipment_returned]
                                                            ↓
                                                       Returned
                                                            ↓
                                                      [invoice_sent]
                                                            ↓
                                                       Invoiced
                                                            ↓
                                                      [completed]
                                                            ↓
                                                      Completed
                                                            ↓
                                                      [archive]
                                                            ↓
                                                      Archived
```

### 1.2 State Transition Rules (detailliert)

| Übergang | Wer darf? | Vorbedingungen | Was passiert? | Events |
|----------|-----------|----------------|---------------|--------|
| **Draft → Quoted** | Marco (PM) | Projekt erstellt, Angebot vorbereitet | Angebot-Version 1 generiert, Entwurfsstatus | `ProjectQuotedV1`, `OfferCreated` |
| **Quoted → Confirmed** | Marco (PM) | Kunde hat Angebot akzeptiert | Angebot als "Accepted" markiert, Projekt → Confirmed, Equipment-Reservierungen aktiv | `ProjectConfirmed`, `OfferAccepted`, `EquipmentReserved` |
| **Quoted → Rejected** | Marco (PM) | Kunde lehnt ab oder Time-out (z.B. 14 Tage) | Projekt → Rejected, Equipment-Reservierungen freigegeben | `ProjectRejected`, `OfferExpired` |
| **Confirmed → InPreparation** | Marco (PM), Lisa (Lager) | Projektstart rückt näher (<14 Tage) | Packlisten generieren, Equipment-Check starten, Crew-Zuweisungen bestätigen | `ProjectInPreparation`, `PacklistGeneratedV1` |
| **InPreparation → LoadingOut** | Lisa (Lager) | Packlisten vollständig, Equipment verfügbar/eingecheckt | Equipment scannen, auf LKW laden, Gewicht/Volumen erfassen | `LoadingStarted`, `EquipmentLoadedEvent` |
| **LoadingOut → OnSite** | Kevin (Techniker) beim Laden, Marco (Dispatcher) | LKW verlässt Halle | GPS-Tracking aktiv, Live-Status Update | `ProjectDispatched`, `LKWDeparted` |
| **OnSite → LoadingIn** | Kevin (Techniker) vor Ort | Setup komplett, Show zu Ende, Equipment packbar | Abbau-Event triggern, Equipment-Rückgabe-Scan vorbereiten | `OnSiteLoadingStarted`, `EventComplete` |
| **LoadingIn → Returned** | Lisa (Lager) | Alle Equipment eingelagert & eingecheckt | Projekt-Status abgeschlossen für Lagerseite, Zustandsberichte verfügbar | `EquipmentReturned`, `ProjectReadyForInvoicing` |
| **Returned → Invoiced** | Thomas (Buchhaltung) | Rechnung erstellt, geprüft, freigegeben | Rechnung-Event triggert Buchhaltungs-Export, Projekt → Invoiced | `InvoiceSent`, `ProjectInvoiced` |
| **Invoiced → Completed** | Marco (PM) oder System (automatisch nach Zahlungseingang) | Rechnung bezahlt ODER 90 Tage später | Zahlungseingang bestätigt, Projekt archivierbar | `ProjectCompleted`, `PaymentReceived` |
| **Completed → Archived** | Marco (PM) oder Auto-Archiv (jährlich) | 1 Jahr in Completed-Status | Projekt → Read-Only, Audit-Trail, keine Änderungen mehr | `ProjectArchived` |

### 1.3 Projekt-Datenmodell

```go
type Project struct {
    ID                      string            // UUID, z.B. "PRJ-2026-0042"
    Number                  string            // Fortlaufende Nummer für Reports
    Status                  ProjectStatus     // enum: Draft, Quoted, Confirmed, ...
    StatusHistory           []StatusChange    // Zeitstempel + Benutzer

    // Basis-Infos
    Name                    string            // z.B. "Festival Stadtpark"
    Description             string            // längere Beschreibung
    CustomerID              string            // FK zu Customer
    CustomerName            string            // denormalisiert für schnelle Reports

    // Zeitplan
    PlannedStartDate        time.Time         // Aufbau-Beginn
    PlannedEndDate          time.Time         // Abbau-Ende
    ActualStartDate         *time.Time        // On-Site Start (nullable)
    ActualEndDate           *time.Time        // On-Site End (nullable)

    // Ort
    LocationName            string            // "Rathaus Marktplatz"
    LocationAddress         string
    LocationCity            string
    LocationZIP             string
    LocationCountry         string            // ISO 3166-1 alpha-2
    LocationGPS             struct {
        Latitude            float64
        Longitude           float64
    }
    LocationContactPerson   string
    LocationContactPhone    string

    // Equipment
    EquipmentItems          []ProjectEquipment
    ReservationStartDate    time.Time         // Wann ist Equipment reserviert
    ReservationEndDate      time.Time
    EquipmentConflicts      []ConflictReport  // Doppelbuchungen erkannt

    // Packlisten
    Packlists              []Packlist         // Multiple Versionen
    CurrentPacklistVersion  int               // Version 1, 2, 3...

    // Crew
    CrewAssignments        []CrewAssignment   // Wer arbeitet wann?

    // Angebot & Rechnung
    OfferID                string             // FK zu Offer
    OfferVersion           int                // welche Version akzeptiert
    InvoiceID              *string            // FK zu Invoice (nullable bis Returned)
    PartialInvoices        []string           // IDs aller Teilrechnungen

    // Kosten & Kalkulation
    EstimatedEquipmentCost float64            // Netto €
    EstimatedCrewCost      float64            // Netto €
    EstimatedTransport     float64            // Pauschal oder berechnet
    EstimatedSubRentalCost float64            // Von anderen Partnern
    OfferedPrice           float64            // Gesamtangebot Brutto
    ExpectedMargin         float64            // % für Marco

    ActualEquipmentCost    float64            // Nach Projekt (mit Extras)
    ActualCrewCost         float64            // Nach Zeiterfassung
    ActualTransport        float64            // GPS-Tracking
    ActualSubRentalCost    float64
    ActualTotal            float64
    RealisedMargin         float64

    // Meta
    CreatedBy              string             // User ID
    CreatedAt              time.Time
    UpdatedBy              string
    UpdatedAt              time.Time
    Notes                  string             // Interne Notizen (nicht für Kunde)
    Template               *string            // Falls aus Template erstellt
}

type ProjectStatus string
const (
    Draft           ProjectStatus = "DRAFT"
    Quoted          ProjectStatus = "QUOTED"
    Confirmed       ProjectStatus = "CONFIRMED"
    InPreparation   ProjectStatus = "IN_PREPARATION"
    LoadingOut      ProjectStatus = "LOADING_OUT"
    OnSite          ProjectStatus = "ON_SITE"
    LoadingIn       ProjectStatus = "LOADING_IN"
    Returned        ProjectStatus = "RETURNED"
    Invoiced        ProjectStatus = "INVOICED"
    Completed       ProjectStatus = "COMPLETED"
    Rejected        ProjectStatus = "REJECTED"
    Archived        ProjectStatus = "ARCHIVED"
)

type StatusChange struct {
    FromStatus      ProjectStatus
    ToStatus        ProjectStatus
    ChangedBy       string         // User ID
    ChangedAt       time.Time
    Reason          string         // warum?
}
```

### 1.4 Equipment-Zuordnung & Konflikt-Erkennung

```go
type ProjectEquipment struct {
    ID                  string           // UUID
    ProjectID           string           // FK
    EquipmentID         string           // FK zu Equipment-Master
    EquipmentName       string           // denormalisiert
    Category            string           // "PA-System", "Licht", "Video"
    Quantity            int              // Menge
    UnitPrice           float64          // Netto €/Stück/Tag
    PlannedDays         float64          // 2.5 Tage?
    CalculatedCost      float64          // Qty × UnitPrice × PlannedDays (Netto)

    // Status
    Status              EquipmentStatus  // PLANNED, RESERVED, PACKED, LOADED, ON_SITE, RETURNED, STORED
    StatusHistory       []struct {
        Status          EquipmentStatus
        Timestamp       time.Time
        Location        string           // "Lager Regal A3", "On-Site", "SoundPro"
    }

    // Reservierung
    ReservationMode     ReservationMode  // FIXED_ASSIGNMENT, FLEXIBLE_POOL
    SourceLocation      string           // "Eigenes Lager" oder "SoundPro" (Sub-Rental)
    IsSubRental         bool
    SubRentalPartnerID  *string          // FK zu Partner wenn isSubRental=true

    // Konflikt-Info
    ConflictDetected    bool
    ConflictMessage     string           // z.B. "K2 ist für Projekt XY 28.-30.3. gebucht"
    AlternativeOptions  []struct {
        EquipmentID     string
        Name            string
        InStock         bool
        AvailableQty    int
        PriceDelta      float64          // wie viel teurer/günstiger
        Reason          string           // "E.Serie 3000€ Mietpreis statt K2 2500€"
    }

    // Zustand
    ConditionOnDispatch string           // "OK", "DEFECT_MARKED", "NEEDS_CHECK"
    ConditionOnReturn   string           // "OK", "DEFECT", "DAMAGED", "MISSING"
    DefectReport        *DefectReport    // wenn ConditionOnReturn != "OK"
}

type EquipmentStatus string
const (
    PLANNED       EquipmentStatus = "PLANNED"        // In Angebot
    RESERVED      EquipmentStatus = "RESERVED"       // Confirmed, Equipment-Pool gesperrt
    PACKED        EquipmentStatus = "PACKED"         // In Packliste, im Lager vorbereitet
    LOADED        EquipmentStatus = "LOADED"         // Auf dem LKW
    ON_SITE       EquipmentStatus = "ON_SITE"        // Auf der Baustelle
    RETURNED      EquipmentStatus = "RETURNED"       // Zurück im Lager eingecheckt
    STORED        EquipmentStatus = "STORED"         // Eingelagert
)

type ReservationMode string
const (
    FIXED_ASSIGNMENT ReservationMode = "FIXED"       // Dieser spezielle K2 Serial# XYZ
    FLEXIBLE_POOL    ReservationMode = "FLEXIBLE"    // Irgendein K2, Hauptsache da
)

// Konflikt-Erkennung: Service-Logik
func (svc *ProjectService) CheckEquipmentConflicts(
    ctx context.Context,
    projectID string,
    equipmentList []ProjectEquipment,
) ([]ConflictReport, error) {

    // für jedes Equipment in der Liste:
    // 1. Abfrage: Welche anderen Projekte in diesem Zeitraum nutzen das Equipment?
    // 2. Ist Qty verfügbar? (Total - Reserved - On-Site)
    // 3. Wenn Konflikt: Alternativen suchen (ähnliche Equipment in gleicher Kategorie)
    // 4. Alternatives mit Preis-Delta berechnen
    // 5. Return ConflictReport

    conflicts := make([]ConflictReport, 0)

    for _, item := range equipmentList {
        bookings, err := svc.repo.GetConcurrentBookings(
            ctx,
            item.EquipmentID,
            item.PlannedStartDate,
            item.PlannedEndDate,
        )
        // ... Logik

        if conflictFound {
            alts := svc.findAlternatives(ctx, item)
            conflicts = append(conflicts, ConflictReport{
                EquipmentID: item.EquipmentID,
                ConflictingProjects: bookings,
                Alternatives: alts,
            })
        }
    }

    return conflicts, nil
}
```

### 1.5 Packlisten: Generierung, Bearbeitung, Mehrbenutzer-Sync

```go
type Packlist struct {
    ID                  string
    ProjectID           string           // FK
    Version             int              // V1, V2, V3... wenn Änderungen
    CreatedBy           string           // User ID
    CreatedAt           time.Time
    UpdatedBy           string
    UpdatedAt           time.Time

    Items               []PacklistItem

    // Meta
    Status              PacklistStatus   // DRAFT, FINALIZED, IN_PROGRESS, COMPLETED
    SortingMode         string           // "WAREHOUSE_LOCATION" (optimiert für Packvorgänge)
    EstimatedWeight     float64          // kg
    EstimatedVolume     float64          // m³ (für LKW-Planung)
    MultiUserLocking    map[string]UserLock // wer bearbeitet gerade?
}

type PacklistItem struct {
    ID                  string
    EquipmentID         string
    EquipmentName       string
    Quantity            int
    Category            string           // "Ton", "Licht", "Video"

    // Lagerplatz-Info (optimiert für Picking)
    WarehouseLocation   struct {
        Hall            string           // "A", "B"
        Rack            string           // "R12"
        Shelf           string           // "3"
        Bin             string           // "05"
        FullPath        string           // "A-R12-3-05" für Sortierung
    }
    CaseLabel           string           // was steht auf der Kiste?
    CasePhoto           string           // URL zu Foto

    // Status pro Item
    Status              PacklistItemStatus // PLANNED, PACKED, LOADED, ON_SITE, RETURNED, STORED
    PackedBy            *string            // User ID
    PackedAt            *time.Time
    LoadedBy            *string
    LoadedAt            *time.Time
    ReturnedBy          *string
    ReturnedAt          *time.Time

    // Zustand
    ConditionOnDispatch string           // "OK", "DEFECT_MARKED"
    ConditionOnReturn   string           // "OK", "DEFECT", "DAMAGED", "MISSING"

    // Notizen
    Notes               string           // "Kabel Funk, Pin 7 wackelt — aber OK"
}

type PacklistStatus string
const (
    DRAFT       PacklistStatus = "DRAFT"
    FINALIZED   PacklistStatus = "FINALIZED"    // Ready for dispatch
    IN_PROGRESS PacklistStatus = "IN_PROGRESS"  // Aktives Packen
    COMPLETED   PacklistStatus = "COMPLETED"    // Alles zurück
)

type PacklistItemStatus string
const (
    PLANNED     PacklistItemStatus = "PLANNED"
    PACKED      PacklistItemStatus = "PACKED"
    LOADED      PacklistItemStatus = "LOADED"
    ON_SITE     PacklistItemStatus = "ON_SITE"
    RETURNED    PacklistItemStatus = "RETURNED"
    STORED      PacklistItemStatus = "STORED"
)

// WebSocket-Event für Mehrbenutzer-Sync
type PacklistUpdateEvent struct {
    PacklistID      string
    ItemID          string
    Action          string              // "pack", "load", "unpack", "update_status"
    NewStatus       PacklistItemStatus
    UpdatedBy       string
    Timestamp       time.Time
    ConflictsWith   *string             // wenn gleichzeitig andere User bearbeitet
}

// Packlisten-Generierung
func (svc *ProjectService) GeneratePacklist(
    ctx context.Context,
    projectID string,
) (*Packlist, error) {

    project, err := svc.repo.GetProject(ctx, projectID)
    // 1. Alle ProjectEquipment laden
    // 2. Nach WarehouseLocation sortieren (Warehouse-Service)
    // 3. Packlist-Items mit optimaler Reihenfolge erstellen
    // 4. CasePhotos, Labels abrufen
    // 5. Packlist speichern als V1
    // 6. Event: PacklistGenerated → Lager-UI aktualisiert

    return packlist, nil
}

// Packlisten-Bearbeitung mit Optimistisches Locking
func (svc *ProjectService) UpdatePacklistItem(
    ctx context.Context,
    packlistID string,
    itemID string,
    newStatus PacklistItemStatus,
    updatedBy string,
) error {

    // 1. Optimistic Lock: Check version
    item, version := svc.repo.GetPacklistItem(ctx, packlistID, itemID)

    // 2. Prüfen: Darf dieser User diese Aktion?
    //    - PACKED: nur Lager-Staff
    //    - LOADED: nur LKW-Driver oder Techniker
    //    - ON_SITE: nur Techniker
    //    - RETURNED: nur Lager-Staff

    // 3. Zustandsübergang validieren
    //    z.B. PLANNED → PACKED ok, aber LOADED → PLANNED nicht

    // 4. Speichern mit Version-Check
    updated := svc.repo.UpdatePacklistItem(ctx, packlistID, itemID,
        version, newStatus, updatedBy)

    // 5. WebSocket-Event senden an alle Clients
    svc.wsHub.Broadcast(ctx, &PacklistUpdateEvent{
        PacklistID: packlistID,
        ItemID: itemID,
        NewStatus: newStatus,
        UpdatedBy: updatedBy,
        Timestamp: time.Now(),
    })

    return nil
}
```

### 1.6 Projekt-Templates

```go
type ProjectTemplate struct {
    ID                  string
    Name                string              // "Standard Firmenfeier 50-100 Personen"
    Description         string
    CreatedBy           string
    CreatedAt           time.Time

    // Template-Inhalt
    EquipmentTemplate   []EquipmentTemplate // Vordefinierte Equipment-Zusammenstellung
    CrewTemplate        []CrewTemplate      // Typische Crew-Zusammensetzung
    LocationTemplate    struct {
        DefaultCity     string
        SetupType       string              // "Theater", "Gala", "Festival", "Corporate"
    }
    PricingTemplate     struct {
        StandardMargin  float64             // %
        DiscountRules   []DiscountRule
    }

    // Verwendung
    UsageCount          int
    LastUsedAt          *time.Time
}

// Templates erstellen aus bestehendem Projekt:
func (svc *ProjectService) CreateTemplateFromProject(
    ctx context.Context,
    projectID string,
    templateName string,
) (*ProjectTemplate, error) {
    // Projekt laden, Equipment/Crew extrahieren, als Template speichern
}

// Template auf neues Projekt anwenden:
func (svc *ProjectService) ApplyTemplate(
    ctx context.Context,
    projectID string,
    templateID string,
) error {
    // Equipment, Crew, Pricing aus Template in neues Projekt kopieren
}
```

### 1.7 Kalkulation: Equipment + Crew + Transport + Sub-Rental vs. Angebot

```go
type ProjectCostCalculation struct {
    // Eingaben
    EquipmentItems      []ProjectEquipment
    CrewAssignments     []CrewAssignment
    TransportMode       TransportMode    // INCLUDED, CALCULATED, EXTERNAL
    SubRentalItems      []ProjectEquipment

    // Berechnungen (alle Netto)
    EquipmentCost struct {
        OwnEquipment    float64          // Eigene Geräte × Tage × Tagesmiete
        DepreciationAdd float64          // Falls Neuware mitgenommen (für Kalkulation)
        Total           float64
    }

    CrewCost struct {
        BaseSalary      float64          // Techniker × Tage × Tagessatz
        NightShift      float64          // 22-6 Uhr Zuschlag +25%
        Overtime        float64          // >10h/Tag +50%
        SpecialSkills   float64          // Rigging/Safety +15%
        Total           float64
    }

    TransportCost struct {
        CalculatedKm    float64          // 0.65€/km wenn berechnet
        LKWRental       float64          // wenn gemieteter LKW
        Fuel            float64
        Logistics       float64          // Packen, Fahren, Ausladen
        Total           float64
    }

    SubRentalCost struct {
        PartnerCharges  []struct {
            Partner     string
            Items       []string
            Cost        float64
        }
        Total           float64
    }

    // Gesamten
    TotalCostNetto      float64

    // vs. Angebot
    OfferedPrice        float64          // Brutto aus Offer
    OfferedPriceNetto   float64          // ohne MwSt
    MarginAbsolute      float64          // OfferedPriceNetto - TotalCostNetto
    MarginPercent       float64          // % auf Kosten
    BreakEven           bool             // wenn MarginPercent < 0
}

type TransportMode string
const (
    INCLUDED   TransportMode = "INCLUDED"    // in Netto-Preis enthalten
    CALCULATED TransportMode = "CALCULATED"  // dynamisch nach km/Zeit
    EXTERNAL   TransportMode = "EXTERNAL"    // Kunde arrangiert selbst
)

// Service: Laufend Kosten neu berechnen wenn sich Equipment/Crew ändert
func (svc *ProjectService) RecalculateProjectMargin(
    ctx context.Context,
    projectID string,
) (*ProjectCostCalculation, error) {

    project, _ := svc.repo.GetProject(ctx, projectID)
    calc := &ProjectCostCalculation{}

    // 1. Equipment durchlaufen
    for _, eq := range project.EquipmentItems {
        if eq.IsSubRental {
            calc.SubRentalCost.Total += eq.CalculatedCost
        } else {
            calc.EquipmentCost.Total += eq.CalculatedCost
        }
    }

    // 2. Crew durchlaufen (Zeiterfassungs-Records)
    crew, _ := svc.crewService.GetProjectCrew(ctx, projectID)
    for _, assignment := range crew {
        records, _ := svc.crewService.GetTimeEntries(ctx, assignment.FreelancerID, projectID)
        // Stunden berechnen, Nachtzuschläge, Überstunden
        calc.CrewCost.Total += calculateCrewTotal(records)
    }

    // 3. Transport
    if project.ReservationMode == CALCULATED {
        calc.TransportCost.CalculatedKm = calculateDistance(project.LocationGPS) * 0.65
    }

    // 4. Marge gegen Angebot
    calc.OfferedPrice = project.OfferedPrice
    calc.OfferedPriceNetto = project.OfferedPrice / 1.19
    calc.TotalCostNetto = calc.EquipmentCost.Total +
                          calc.CrewCost.Total +
                          calc.TransportCost.Total +
                          calc.SubRentalCost.Total
    calc.MarginAbsolute = calc.OfferedPriceNetto - calc.TotalCostNetto
    calc.MarginPercent = (calc.MarginAbsolute / calc.TotalCostNetto) * 100

    return calc, nil
}
```

---

## 2. ANGEBOTS-SYSTEM (Offer Module O1-O5)

### 2.1 Angebots-Datenmodell

```go
type Offer struct {
    ID                  string            // UUID "OFF-2026-0001"
    ProjectID           *string           // FK zu Project (nullable, Angebot kann vor Projekt sein)
    Version             int               // V1, V2, V3...
    Status              OfferStatus       // DRAFT, SENT, ACCEPTED, REJECTED, EXPIRED
    StatusHistory       []struct {
        Status          OfferStatus
        Timestamp       time.Time
        UpdatedBy       string
    }

    // Kunde
    CustomerID          string            // FK
    CustomerName        string
    CustomerEmail       string
    CustomerAddress     string
    CustomerCity        string
    CustomerZIP         string
    CustomerCountry     string            // ISO 3166-1

    // Daten
    CreatedBy           string            // User (Marco)
    CreatedAt           time.Time
    ValidUntil          time.Time         // z.B. 14 Tage
    SentAt              *time.Time
    SentTo              *string           // Email-Adresse an die gesendet
    AcceptedAt          *time.Time
    AcceptedBy          *string           // Kunde-Kontakt oder User-ID?
    RejectedAt          *time.Time
    RejectionReason     *string

    // Inhalt
    Positions           []OfferPosition

    // Preisberechnung
    SubtotalNetto       float64           // Summe aller Positionen (Netto)
    TaxAmount           float64           // MwSt
    TaxRate             float64           // 19%, 7%, 0%
    TotalBrutto         float64           // zum Zahlen

    // Meta
    Notes               string            // Interne Notizen
    TermsOfService      string            // Text (aus Config)
    PaymentTerms        string            // z.B. "30 Tage Zahlungsziel"
    DeliveryTerms       string

    // PDF-Generierung
    PDFGenerated        *time.Time
    PDFPath             string            // S3/Blob-Storage
}

type OfferStatus string
const (
    DRAFT       OfferStatus = "DRAFT"
    SENT        OfferStatus = "SENT"
    ACCEPTED    OfferStatus = "ACCEPTED"
    REJECTED    OfferStatus = "REJECTED"
    EXPIRED     OfferStatus = "EXPIRED"
)

type OfferPosition struct {
    ID                  string
    PositionNumber      int              // 1, 2, 3... (für PDF)
    Type                PositionType      // EQUIPMENT, CREW, FLAT_RATE, DISCOUNT, OPTIONAL

    // Artikel-Daten
    Description         string            // z.B. "L-Acoustics K2 (4 Stück, 3 Tage, FOH)"
    Quantity            float64           // 4
    Unit                string            // "Stück", "Tag", "Pauschal"
    UnitPrice           float64           // Netto €

    // Berechnung
    Amount              float64           // Qty × UnitPrice (Netto)
    TaxAmount           float64           // Amount × TaxRate%
    TotalBrutto         float64           // Amount + TaxAmount

    // Besonderheiten
    Optional            bool              // Kunde kann ja/nein wählen
    DeliveryDate        *string           // z.B. "28.03.2026"
    Notes               string

    // Referenz
    EquipmentID         *string           // wenn Type=EQUIPMENT
    CrewCategoryID      *string           // wenn Type=CREW

    // Visuals
    EquipmentPhoto      *string           // URL (optional für fancy Angebote)
}

type PositionType string
const (
    EQUIPMENT   PositionType = "EQUIPMENT"    // Gerät
    CREW        PositionType = "CREW"         // Personal
    FLAT_RATE   PositionType = "FLAT_RATE"    // Pauschalpreis (Transport, Setup)
    DISCOUNT    PositionType = "DISCOUNT"     // Rabatt (-10%, -200€)
    OPTIONAL    PositionType = "OPTIONAL"     // Optional (Kunde wählt)
)
```

### 2.2 Angebots-Workflow

```
Workflow-Schritte:

1. Angebot erstellen (DRAFT)
   - Marco klickt "Neues Angebot"
   - Kunde + Projekt auswählen
   - Positionen hinzufügen
   - Preis berechnen
   - Zwischenspeichern möglich

2. Angebot editieren (DRAFT)
   - Positionen ändern/löschen
   - Preise anpassen
   - Versionierung automatisch (V1 → V2 wenn gesendet, dann DRAFT ist V2)

3. Angebot prüfen & versenden (DRAFT → SENT)
   - Marco klickt "Versenden"
   - System generiert PDF (Go-Template + chromedp)
   - PDF an Kunde per E-Mail
   - Status → SENT, SentAt = now(), SentTo = email
   - ValidUntil berechnet: jetzt + 14 Tage

4. Kunde akzeptiert (SENT → ACCEPTED)
   - Option A: Kunde klickt Link in Mail "Ich akzeptiere" → Web-Form
   - Option B: Marco markiert per UI "Akzeptiert am TT.MM"
   - Status → ACCEPTED, AcceptedAt = now()
   - Event: OfferAccepted → Trigger für Projekt-Konvertierung

5. Kunde lehnt ab (SENT → REJECTED)
   - Kunde antwortet per Mail
   - Marco markiert "Abgelehnt"
   - RejectionReason optional notieren
   - Status → REJECTED, RejectedAt = now()

6. Automatisch verfallen (SENT → EXPIRED)
   - Cron-Job: jeden Tag um 6 Uhr
   - Wenn now() > offer.ValidUntil und status = SENT
   - Status → EXPIRED
   - Event: OfferExpired → Notifikation an Marco
```

### 2.3 Positionen detailliert

```go
// Equipment-Position
// "L-Acoustics K2, 4x, 3 Tage (Fr-So), FOH-Setup"
// Preis: 2500€/Stück/Miettag × 3 Tage = 7500€/Stück × 4 = 30.000€ Netto

func (svc *OfferService) AddEquipmentPosition(
    ctx context.Context,
    offerID string,
    equipmentID string,
    quantity int,
    daysOfRental float64,
    notes string,
) (*OfferPosition, error) {

    eq, _ := svc.equipmentService.GetEquipment(ctx, equipmentID)

    pos := &OfferPosition{
        Type: EQUIPMENT,
        Description: fmt.Sprintf("%s (%dx, %.1f Tage)", eq.Name, quantity, daysOfRental),
        Quantity: float64(quantity),
        Unit: "Stück",
        UnitPrice: eq.DailyRentalPrice * daysOfRental,  // z.B. 2500 * 3 = 7500
        Amount: float64(quantity) * eq.DailyRentalPrice * daysOfRental,
        EquipmentID: &equipmentID,
    }

    return pos, nil
}

// Crew-Position
// "Tontechniker, 2x, 3 Tage (Fr-So)"
// Preis: 300€/Tag × 3 Tage = 900€/Stück × 2 = 1800€ Netto

func (svc *OfferService) AddCrewPosition(
    ctx context.Context,
    offerID string,
    crewCategory string,  // "Tontechniker", "Licht", "Rigging"
    quantity int,
    days float64,
    dailyRate float64,
) (*OfferPosition, error) {

    pos := &OfferPosition{
        Type: CREW,
        Description: fmt.Sprintf("%s, %dx, %.1f Tage", crewCategory, quantity, days),
        Quantity: float64(quantity),
        Unit: "Stück",
        UnitPrice: dailyRate * days,
        Amount: float64(quantity) * dailyRate * days,
    }

    return pos, nil
}

// Pauschal-Position
// "Transportfahrt + Aufbau Pauschal" = 2000€

func (svc *OfferService) AddFlatRatePosition(
    ctx context.Context,
    offerID string,
    description string,
    amountNetto float64,
) (*OfferPosition, error) {

    pos := &OfferPosition{
        Type: FLAT_RATE,
        Description: description,
        Quantity: 1,
        Unit: "Pauschal",
        UnitPrice: amountNetto,
        Amount: amountNetto,
    }

    return pos, nil
}

// Rabatt-Position
// "Treue-Rabatt 10%" oder "Rabatt €200 auf PA-System"

func (svc *OfferService) AddDiscountPosition(
    ctx context.Context,
    offerID string,
    reason string,
    discountType DiscountType,  // PERCENT, ABSOLUTE
    discountValue float64,       // 10 (für %), 200 (für €)
) (*OfferPosition, error) {

    offer, _ := svc.repo.GetOffer(ctx, offerID)

    var amount float64
    if discountType == PERCENT {
        // Berechne auf Subtotal (exkl. andere Rabatte)
        subtotal := 0.0
        for _, pos := range offer.Positions {
            if pos.Type != DISCOUNT {
                subtotal += pos.Amount
            }
        }
        amount = subtotal * (discountValue / 100.0)
    } else {
        amount = discountValue
    }

    pos := &OfferPosition{
        Type: DISCOUNT,
        Description: fmt.Sprintf("%s: -%s", reason, formatEuro(amount)),
        Quantity: 1,
        Unit: "Pauschal",
        UnitPrice: -amount,
        Amount: -amount,
    }

    return pos, nil
}

// Optionale Position
// Kunde kann ja/nein wählen, z.B. "Video-Streaming +1500€?"

func (svc *OfferService) AddOptionalPosition(
    ctx context.Context,
    offerID string,
    description string,
    amountNetto float64,
) (*OfferPosition, error) {

    pos := &OfferPosition{
        Type: OPTIONAL,
        Description: description,
        Quantity: 1,
        Unit: "Optional",
        UnitPrice: amountNetto,
        Amount: amountNetto,
        Optional: true,
    }

    return pos, nil
}
```

### 2.4 Preisberechnung & MwSt

```go
type TaxCalculation struct {
    SubtotalNetto       float64
    TaxRate             float64          // 0.19, 0.07, 0.0
    TaxAmount           float64
    TotalBrutto         float64
}

func (svc *OfferService) CalculateTax(
    ctx context.Context,
    offerID string,
) (*TaxCalculation, error) {

    offer, _ := svc.repo.GetOffer(ctx, offerID)

    // 1. Subtotal berechnen (alle Positionen außer Rabatt)
    subtotal := 0.0
    for _, pos := range offer.Positions {
        subtotal += pos.Amount
    }

    // 2. Rabatte abziehen
    discount := 0.0
    for _, pos := range offer.Positions {
        if pos.Type == DISCOUNT {
            discount += pos.Amount  // negative
        }
    }

    nettoAmount := subtotal + discount

    // 3. MwSt-Satz bestimmen
    taxRate := svc.determineTaxRate(ctx, offer.CustomerCountry, offer.CustomerTaxID)

    // 4. MwSt berechnen
    taxAmount := nettoAmount * taxRate
    brutto := nettoAmount + taxAmount

    return &TaxCalculation{
        SubtotalNetto: subtotal,
        TaxRate: taxRate,
        TaxAmount: taxAmount,
        TotalBrutto: brutto,
    }, nil
}

// MwSt-Satz-Erkennung
func (svc *OfferService) determineTaxRate(
    ctx context.Context,
    customerCountry string,      // "DE", "AT", "NL"
    customerTaxID *string,       // USt-IdNr wenn vorhanden
) float64 {

    config, _ := svc.configService.GetConfig(ctx)

    // 1. Kleinunternehmer?
    if config.VATExemptionSmallBusiness {
        return 0.0  // 0% für Marco
    }

    // 2. Reverse Charge (innergemeinschaftlich)?
    //    Wenn Kunde in EU + gültige USt-IdNr vorhanden → 0%
    if svc.isEUCountry(customerCountry) && customerTaxID != nil {
        if svc.validateVATID(customerCountry, *customerTaxID) {
            return 0.0
        }
    }

    // 3. Ausland (nicht EU)? → 0%
    if customerCountry != "DE" && !svc.isEUCountry(customerCountry) {
        return 0.0
    }

    // 4. Inland → Standard 19% oder ermäßigt 7%
    //    (Generell 19% für VT-Dienstleistungen)
    return 0.19
}
```

### 2.5 Angebots-PDF Generierung

```go
// Template in Go (nicht in HTML!)
const offerPDFTemplate = `
{{- $offer := . -}}
┌─────────────────────────────────────────────┐
│         MB VERANSTALTUNGSTECHNIK GmbH       │
│         [Logo eingebettet als Base64]        │
└─────────────────────────────────────────────┘

ANGEBOT NR. {{ $offer.ID }}
Gültig bis: {{ $offer.ValidUntil | formatDate }}

EMPFÄNGER:
{{ $offer.CustomerName }}
{{ $offer.CustomerAddress }}
{{ $offer.CustomerZIP }} {{ $offer.CustomerCity }}

────────────────────────────────────────────────────

LEISTUNGSÜBERSICHT:

┌────┬──────────────────────┬──────┬────────┬─────────┐
│ Pos│ Beschreibung         │ Qty  │ Preis  │ Betrag  │
├────┼──────────────────────┼──────┼────────┼─────────┤
{{ range $i, $pos := $offer.Positions }}
│ {{ $pos.PositionNumber | printf "%2d" }} │ {{ $pos.Description | truncate 20 }} │ {{ $pos.Quantity | printf "%.1f" }} {{ $pos.Unit | printf "%-3s" }} │ {{ $pos.UnitPrice | formatEuro }} │ {{ $pos.Amount | formatEuro }} │
{{ end }}
├────┴──────────────────────┴──────┴────────┼─────────┤
│                         Netto: │ {{ $offer.SubtotalNetto | formatEuro }} │
│                MwSt ({{ $offer.TaxRate | multiply 100 }}%): │ {{ $offer.TaxAmount | formatEuro }} │
│                        Brutto: │ {{ $offer.TotalBrutto | formatEuro }} │
└─────────────────────────────────────────────────────┘

ZAHLUNGSBEDINGUNGEN:
{{ $offer.PaymentTerms }}

LIEFERBEDINGUNGEN:
{{ $offer.DeliveryTerms }}

ALLGEMEINE GESCHÄFTSBEDINGUNGEN:
{{ $offer.TermsOfService }}

────────────────────────────────────────────────────
Gültig: {{ $offer.CreatedAt | formatDate }} - {{ $offer.ValidUntil | formatDate }}

Mit freundlichen Grüßen,
{{ $offer.CreatedBy }}
MB Veranstaltungstechnik GmbH
`

func (svc *OfferService) GeneratePDF(
    ctx context.Context,
    offerID string,
) ([]byte, error) {

    offer, _ := svc.repo.GetOffer(ctx, offerID)

    // 1. Template rendern
    tmpl, _ := template.New("offer").Parse(offerPDFTemplate)
    var buf bytes.Buffer
    tmpl.Execute(&buf, offer)
    html := buf.String()

    // 2. HTML → PDF mit chromedp
    var pdfBytes []byte
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    chromedp.Run(ctx,
        chromedp.SetContent(html),
        chromedp.WaitVisible("body"),
        chromedp.PrintToPDF(
            "/tmp/offer.pdf",
            &protocol.PrintToPDFParams{
                DisplayHeaderFooter: true,
                HeaderTemplate: "<span>MB Veranstaltungstechnik</span>",
                FooterTemplate: "<span>Seite <span class=\"pageNumber\"></span></span>",
                Landscape: false,
                PaperWidth: 8.27,  // A4
                PaperHeight: 11.69,
                MarginTop: 0.5,
                MarginBottom: 0.5,
                MarginLeft: 0.5,
                MarginRight: 0.5,
            },
            &pdfBytes,
        ),
    )

    // 3. PDF in Storage speichern (S3 oder Blob)
    pdfPath := fmt.Sprintf("offers/%s/offer-v%d.pdf", offerID, offer.Version)
    svc.storage.Upload(ctx, pdfPath, pdfBytes)

    // 4. Offer aktualisieren mit PDF-Path
    svc.repo.UpdateOfferPDFPath(ctx, offerID, pdfPath)

    return pdfBytes, nil
}
```

### 2.6 Versionierung & Angebot → Projekt-Konvertierung

```go
// Versionierung
type OfferVersion struct {
    ID              string
    OfferID         string
    Version         int              // 1, 2, 3...
    CreatedAt       time.Time
    CreatedBy       string
    Changes         string           // JSON: was hat sich geändert
    SentAt          *time.Time
    AcceptedAt      *time.Time
}

// Neue Version erstellen (wenn DRAFT→Änderung)
func (svc *OfferService) UpdateOffer(
    ctx context.Context,
    offerID string,
    updates OfferUpdate,
) (*Offer, error) {

    offer, _ := svc.repo.GetOffer(ctx, offerID)

    // Wenn Status = DRAFT: in-place Update
    // Wenn Status = SENT/ACCEPTED: neue Version erstellen

    if offer.Status == SENT || offer.Status == ACCEPTED {
        // Neue Version
        newVersion := offer.Version + 1
        newOfferID := fmt.Sprintf("%s_v%d", offerID, newVersion)

        offer.Version = newVersion
        offer.ID = newOfferID
        offer.Status = DRAFT
        // ... apply updates
        svc.repo.CreateOfferVersion(ctx, offer)
    } else {
        // In-place Update
        // ... apply updates
        svc.repo.UpdateOffer(ctx, offer)
    }

    return offer, nil
}

// Angebot → Projekt-Konvertierung (große Funktion!)
func (svc *ProjectService) ConvertOfferToProject(
    ctx context.Context,
    offerID string,
    projectName string,
) (*Project, error) {

    offer, _ := svc.offerService.GetOffer(ctx, offerID)

    if offer.Status != ACCEPTED {
        return nil, errors.New("Nur akzeptierte Angebote können konvertiert werden")
    }

    // 1. Neues Projekt erstellen
    project := &Project{
        ID: generateProjectID(),
        Name: projectName,
        CustomerID: offer.CustomerID,
        Status: CONFIRMED,
        OfferID: &offerID,
        OfferVersion: offer.Version,
        OfferedPrice: offer.TotalBrutto,
    }

    // 2. Equipment-Positionen aus Angebot in Projekt-Equipment übernehmen
    for _, pos := range offer.Positions {
        if pos.Type == EQUIPMENT && pos.EquipmentID != nil {
            // Equipment-Details laden
            eq, _ := svc.equipmentService.GetEquipment(ctx, *pos.EquipmentID)

            projEq := &ProjectEquipment{
                ID: generateUUID(),
                EquipmentID: *pos.EquipmentID,
                EquipmentName: eq.Name,
                Quantity: int(pos.Quantity),
                UnitPrice: pos.UnitPrice,
                PlannedDays: pos.UnitPrice / eq.DailyRentalPrice, // rückwärts berechnet
                Status: PLANNED,
                ReservationMode: FLEXIBLE_POOL,
            }

            project.EquipmentItems = append(project.EquipmentItems, projEq)
        }
    }

    // 3. Crew-Positionen in Projekt-Crew-Anforderungen übernehmen
    for _, pos := range offer.Positions {
        if pos.Type == CREW {
            // Crew-Assignment als Anforderung erstellen
            assignment := &CrewAssignment{
                ID: generateUUID(),
                ProjectID: project.ID,
                Role: pos.Description,  // z.B. "Tontechniker"
                Quantity: int(pos.Quantity),
                DailyRate: pos.UnitPrice / pos.Quantity,  // rückwärts
                Status: REQUESTED,     // noch nicht bestätigt
            }
            project.CrewAssignments = append(project.CrewAssignments, assignment)
        }
    }

    // 4. Equipment-Reservierungen aktivieren
    for _, eq := range project.EquipmentItems {
        svc.equipmentService.ReserveEquipment(ctx, eq.EquipmentID, eq.Quantity,
            project.PlannedStartDate, project.PlannedEndDate)
    }

    // 5. Projekt speichern
    svc.repo.CreateProject(ctx, project)

    // 6. Event: ProjectCreatedFromOffer
    svc.pubsub.Publish(ctx, &ProjectCreatedFromOfferEvent{
        ProjectID: project.ID,
        OfferID: offerID,
        Timestamp: time.Now(),
    })

    return project, nil
}
```

---

## 3. RECHNUNGS-SYSTEM (Invoice Module F1-F8)

### 3.1 Rechnungs-Datenmodell (§14 UStG konform)

```go
type Invoice struct {
    // Pflichtfelder gemäß §14 UStG
    ID                  string            // UUID "RE-2026-0001"
    Number              string            // Fortlaufend ohne Lücken (GoBD)
    Type                InvoiceType       // REGULAR, PARTIAL, ADVANCE, CREDIT_NOTE, REVERSAL, PROFORMA
    Status              InvoiceStatus     // DRAFT, FINALIZED, SENT, PARTIAL_PAID, PAID, OVERDUE, CANCELLED

    // Ausstellung
    IssuedDate          time.Time         // Rechnungsdatum §14(1) UStG
    IssuedBy            string            // User ID

    // Leistungszeitraum
    PerformanceDateFrom time.Time         // Beginn Leistung
    PerformanceDateTo   time.Time         // Ende Leistung

    // Kunde
    CustomerID          string
    CustomerName        string
    CustomerAddress     string
    CustomerCity        string
    CustomerZIP         string
    CustomerCountry     string            // ISO 3166
    CustomerTaxID       *string           // USt-IdNr für Reverse Charge
    CustomerContactEmail string

    // Verkäufer (Marcos Firma)
    SellerName          string            // "MB Veranstaltungstechnik GmbH"
    SellerAddress       string
    SellerCity          string
    SellerZIP           string
    SellerCountry       string
    SellerTaxID         string            // USt-IdNr
    SellerBankAccount   string            // IBAN
    SellerBIC           string

    // Positionen
    Positions           []InvoicePosition

    // Preisberechnung
    SubtotalNetto       float64
    TaxAmount           float64
    TaxRate             float64           // 0.19, 0.07, 0.0
    TotalBrutto         float64           // Zahlbetrag

    // Zahlungsbedingungen
    PaymentTermDays     int               // z.B. 30
    DueDate             time.Time         // IssuedDate + PaymentTermDays
    DiscountDays        *int              // z.B. 10 (Skonto)
    DiscountPercent     *float64          // z.B. 2% bei Zahlung in 10 Tagen

    // Referenzen
    ProjectID           *string           // FK zu Project
    OfferID             *string           // FK zu Offer
    PreviousInvoiceID   *string           // Bei Storno/Reversal

    // Teilrechnungen
    IsPartialInvoice    bool
    PartialInvoiceConfig *PartialInvoiceConfig  // für Nachberechnungen
    AllPreviousPartials  []string               // IDs aller bisherigen Teilrechnungen

    // Zahlungen
    PaidAmount          float64           // Cumulative
    RemainingAmount     float64           // TotalBrutto - PaidAmount
    PaymentRecords      []PaymentRecord

    // Mahnwesen
    DunningLevel        int               // 0 (aktuell), 1 (1. Mahnung), 2 (2. Mahnung)
    DunningPausedUntil  *time.Time        // Falls pausiert (Reklamation)

    // Metadata
    Notes               string
    InternalNotes       string            // nicht für Kunde
    Attachments         []string          // URL zu PDF, Fotos, Belege

    // Meta
    CreatedAt           time.Time
    UpdatedAt           time.Time
    SentAt              *time.Time        // per E-Mail versendet
    CancelledAt         *time.Time
    CancellationReason  *string

    // GoBD Compliance
    CreatedBy           string            // wer hat die Rechnung erstellt
    ArchivedInDATEV     bool              // zur Kontrolle
    DATEVExportDate     *time.Time
    DATEVReference      *string           // Belegnummer in Buchhaltung
}

type InvoiceType string
const (
    REGULAR         InvoiceType = "REGULAR"         // Schlussrechnung
    PARTIAL         InvoiceType = "PARTIAL"         // Teilrechnung
    ADVANCE         InvoiceType = "ADVANCE"         // Abschlagsrechnung
    CREDIT_NOTE     InvoiceType = "CREDIT_NOTE"     // Gutschrift
    REVERSAL        InvoiceType = "REVERSAL"        // Storno-Rechnung
    PROFORMA        InvoiceType = "PROFORMA"        // Proforma (keine MwSt-Wirkung)
)

type InvoiceStatus string
const (
    DRAFT           InvoiceStatus = "DRAFT"
    FINALIZED       InvoiceStatus = "FINALIZED"
    SENT            InvoiceStatus = "SENT"
    PARTIAL_PAID    InvoiceStatus = "PARTIAL_PAID"
    PAID            InvoiceStatus = "PAID"
    OVERDUE         InvoiceStatus = "OVERDUE"
    CANCELLED       InvoiceStatus = "CANCELLED"
)

type InvoicePosition struct {
    ID                  string
    PositionNumber      int              // 1, 2, 3...
    Type                PositionType     // EQUIPMENT, CREW, SERVICE, DISCOUNT

    Description         string           // "L-Acoustics K2, 4x, 3 Tage"
    Quantity            float64
    Unit                string
    UnitPrice           float64          // Netto
    Amount              float64          // Qty × UnitPrice (Netto)

    // Referenz
    EquipmentID         *string
    CrewAssignmentID    *string
    ProjectEquipmentID  *string

    // MwSt
    TaxRate             float64          // kann pro Position unterschiedlich sein!
    TaxAmount           float64

    // Deliverable
    DeliveredQuantity   int              // was wurde wirklich geliefert? (vs. Quantity)
    Remarks             string           // z.B. "Kunde wollte doch 2 extra"
}
```

### 3.2 Rechnungsarten und Nummern-Schema

```go
type PartialInvoiceConfig struct {
    // Splitting-Konfiguration
    Splits              []struct {
        Name            string           // "Anzahlung", "Nach Aufbau", "Nach Abbau"
        Percent         float64          // 30%, 50%, 20%
        DaysAfterStart  int              // Wann fällig: sofort, nach X Tagen
        Description     string           // Beschreibung für Kunde
    }

    // Automatische Nachberechnung bei Änderungen
    AutoAdjustEnabled    bool            // Summen anpassen bei Auftrags-Änderungen?
    TotalProjectAmount   float64         // Original Gesamtsumme (für Prozent-Berechnung)
    LastRecalcAt        *time.Time       // Wann wurde zuletzt neu gerechnet?

    // Fortlaufende Referenz
    AllPartialInvoices  []struct {
        InvoiceID       string
        SplitName       string
        Percent         float64
        Amount          float64
        Status          string           // SENT, PAID, OVERDUE
        IssuedDate      time.Time
        DueDate         time.Time
        PaidAt          *time.Time
    }
}

// Nummern-Schema Generator
type InvoiceNumberGenerator struct {
    Prefix              string           // "RE", "GS", "AR", "RG" (Gutschrift, Abschlag, Storno)
    Year                int              // 2026
    Sequence            int              // 0001, 0002...
}

func (gen *InvoiceNumberGenerator) Generate(invoiceType InvoiceType) string {
    prefix := ""
    switch invoiceType {
    case REGULAR:
        prefix = "RE"  // Rechnung
    case PARTIAL:
        prefix = "RE"  // auch Rechnungen
    case ADVANCE:
        prefix = "AR"  // Abschlagsrechnung
    case CREDIT_NOTE:
        prefix = "GS"  // Gutschrift
    case REVERSAL:
        prefix = "RG"  // Rechnungsgutschrift (Storno)
    case PROFORMA:
        prefix = "PRO" // Proforma
    }

    return fmt.Sprintf("%s-%d-%04d", prefix, gen.Year, gen.Sequence)
}

// Sequenz aus DB laden und incrementieren (ACID)
func (svc *InvoiceService) AllocateInvoiceNumber(
    ctx context.Context,
    invoiceType InvoiceType,
) (string, error) {

    tx, _ := svc.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // SELECT * FROM invoice_sequence WHERE year = 2026 FOR UPDATE
    seq, _ := svc.repo.GetInvoiceSequence(ctx, 2026)

    number := seq.AllocateNumber(invoiceType)  // seq.Sequence++, return formatted

    // UPDATE invoice_sequence SET sequence = ... WHERE year = 2026
    svc.repo.UpdateInvoiceSequence(ctx, seq)

    tx.Commit()
    return number, nil
}
```

### 3.3 Teilrechnungs-Logik (komplexester Part!)

```go
// Szenario:
// Festival-Projekt: 10.000€ Brutto angeboten
// Splitting: 30% Anzahlung, 50% nach Aufbau, 20% nach Abbau
//
// 1. Anzahlung (30% × 10.000€ = 3.000€)
//    → Rechnung RE-2026-0001, fällig sofort
//    → Kunde bezahlt
//
// 2. Projekt geht live, Aufbau fertig
//    → Kunde hat nachbestellt: +2.000€ (total jetzt 12.000€!)
//    → 2. Teilrechnung (50% × 12.000€ = 6.000€)
//    → ABER: Kunde hat schon 3.000€ gezahlt (30% der alten 10.000€)
//    → Neue Rechnung: 6.000€ − 3.000€ = 3.000€ nur für die Differenz?
//    → Oder: 6.000€ und Hinweis "davon 3.000€ bereits gezahlt"?
//
// Lösung: Pro Teilrechnung Betrag ABSOLUT berechnen,
//         automatisch "bereits gezahlte Anzahlungen" berücksichtigen

func (svc *InvoiceService) CreatePartialInvoice(
    ctx context.Context,
    projectID string,
    splitName string,  // "Anzahlung", "Nach Aufbau", etc.
) (*Invoice, error) {

    project, _ := svc.projectService.GetProject(ctx, projectID)

    // 1. Konfiguration laden
    config := project.PartialInvoiceConfig
    split := config.FindSplit(splitName)

    // 2. Aktuelle Projekt-Summe berechnen (mit Nachbestellungen!)
    currentTotal := svc.calculateProjectTotal(ctx, projectID)  // 12.000€

    // 3. Betrag für diese Teilrechnung
    amountForThisInvoice := currentTotal * (split.Percent / 100.0)  // 50% × 12.000€ = 6.000€

    // 4. Bereits bezahlte Anzahlungen berücksichtigen
    previousPayments := 0.0
    for _, prev := range config.AllPartialInvoices {
        if prev.PaidAt != nil && prev.IssuedDate.Before(time.Now()) {
            previousPayments += prev.Amount
        }
    }

    // 5. Netto-Betrag für THIS Invoice
    //    (Ansatz A: Kunde zahlt für diese Phase, unabhängig von Vorherigem)
    invoiceAmount := amountForThisInvoice  // 6.000€

    // 6. Positionen aus Projekt übernehmen (oder neuberechnen?)
    //    → Pro Teilrechnung mit aktuellen Positionen!
    invoice := &Invoice{
        ID: generateUUID(),
        Number: svc.AllocateInvoiceNumber(ctx, PARTIAL),
        Type: PARTIAL,
        IssuedDate: time.Now(),
        ProjectID: &projectID,
        CustomerID: project.CustomerID,
        TotalBrutto: invoiceAmount,  // Diese Phase
        IsPartialInvoice: true,
        PartialInvoiceConfig: config,
        AllPreviousPartials: config.AllPartialInvoices,
    }

    // Positionen aus aktuellem Projekt-Stand kopieren
    for _, eq := range project.EquipmentItems {
        pos := &InvoicePosition{
            // ... aus ProjectEquipment
        }
        invoice.Positions = append(invoice.Positions, pos)
    }

    // 7. Speichern
    svc.repo.CreateInvoice(ctx, invoice)

    // 8. Config aktualisieren (diese Teilrechnung hinzufügen)
    config.AllPartialInvoices = append(config.AllPartialInvoices, struct{}{
        InvoiceID: invoice.ID,
        SplitName: splitName,
        Percent: split.Percent,
        Amount: invoiceAmount,
        Status: "SENT",
        IssuedDate: invoice.IssuedDate,
        DueDate: invoice.DueDate,
    })
    svc.repo.UpdatePartialInvoiceConfig(ctx, projectID, config)

    return invoice, nil
}

// Auto-Nachberechnung wenn Projekt-Summe sich ändert
func (svc *InvoiceService) RecalculatePartialInvoices(
    ctx context.Context,
    projectID string,
) error {

    project, _ := svc.projectService.GetProject(ctx, projectID)
    config := project.PartialInvoiceConfig

    oldTotal := config.TotalProjectAmount  // Original 10.000€
    newTotal := svc.calculateProjectTotal(ctx, projectID)  // Jetzt 12.000€

    if oldTotal == newTotal {
        return nil  // kein Unterschied
    }

    // Differenz berechnen
    diff := newTotal - oldTotal  // +2.000€

    // Für jede noch nicht gezahlte Teilrechnung: neu berechnen?
    // ODER: Eine Zusatz-Rechnung für die Differenz?
    //
    // Best Practice: Zusatz-Rechnung, damit Audit-Trail klar ist
    //
    // Aber: Wenn Splitting prozentual ist (30/50/20), dann muss alles
    // auf Basis neuer Summe neu berechnet werden!
    //
    // Implementierung: Generiere neue Rechnung für "Nachforderung Zusatz-Equipment"

    additionalInvoice := &Invoice{
        ID: generateUUID(),
        Number: svc.AllocateInvoiceNumber(ctx, PARTIAL),
        Type: PARTIAL,
        IssuedDate: time.Now(),
        ProjectID: &projectID,
        TotalBrutto: diff,  // nur die Differenz
        Description: "Nachforderung für zusätzliches Equipment",
    }

    svc.repo.CreateInvoice(ctx, additionalInvoice)
    svc.repo.UpdatePartialInvoiceConfig(ctx, projectID,
        config.WithNewTotal(newTotal).WithAdditionalInvoice(additionalInvoice.ID),
    )

    return nil
}
```

### 3.4 MwSt-Handling

```go
func (svc *InvoiceService) DetermineTaxRate(
    ctx context.Context,
    invoice *Invoice,
) (float64, error) {

    config, _ := svc.configService.GetConfig(ctx)

    // 1. Kleinunternehmer (§19 UStG)?
    if config.SmallBusinessExemption {
        return 0.0  // Keine MwSt
    }

    // 2. Reverse Charge (Innergemeinschaftlich)?
    //    Bedingung: Kunde EU + gültige USt-IdNr
    if invoice.CustomerCountry != "DE" {
        if svc.isEUCountry(invoice.CustomerCountry) {
            if invoice.CustomerTaxID != nil &&
               svc.validateVATID(*invoice.CustomerTaxID, invoice.CustomerCountry) {
                // Reverse Charge: Marco erhebt keine MwSt
                // → Invoice mit 0% MwSt, Notiz "Reverse Charge"
                return 0.0
            }
        }
    }

    // 3. Export (nicht EU)?
    if !svc.isEUCountry(invoice.CustomerCountry) && invoice.CustomerCountry != "DE" {
        return 0.0  // Ausfuhr steuerfrei
    }

    // 4. Inland oder EU mit ungültiger USt-IdNr → Standard 19%
    return 0.19
}

// Angebote und Rechnungen mit unterschiedlichen MwSt-Sätzen
// Beispiel: Equipment 19%, aber Dienstleistung 7%? (nicht typisch für VT)
// → Implementation trotzdem unterstützen

type InvoicePosition struct {
    // ... wie oben
    TaxRate float64  // kann unterschiedlich sein!
}

func (svc *InvoiceService) CalculateInvoiceTotals(invoice *Invoice) {
    subtotalNetto := 0.0
    totalTaxAmount := 0.0

    for _, pos := range invoice.Positions {
        subtotalNetto += pos.Amount

        // MwSt pro Position
        taxForPos := pos.Amount * pos.TaxRate
        totalTaxAmount += taxForPos
        pos.TaxAmount = taxForPos
    }

    invoice.SubtotalNetto = subtotalNetto
    invoice.TaxAmount = totalTaxAmount
    invoice.TotalBrutto = subtotalNetto + totalTaxAmount

    // Für Rechnungskopf: Durchschnittlicher Steuersatz
    if subtotalNetto > 0 {
        invoice.TaxRate = totalTaxAmount / subtotalNetto
    }
}
```

### 3.5 Rechnung aus Projekt generieren

```go
func (svc *InvoiceService) CreateInvoiceFromProject(
    ctx context.Context,
    projectID string,
    invoiceType InvoiceType,  // REGULAR oder PARTIAL
) (*Invoice, error) {

    project, _ := svc.projectService.GetProject(ctx, projectID)

    // Prüfungen
    if project.Status != RETURNED {
        return nil, errors.New("Projekt muss im Status 'Returned' sein")
    }

    if project.InvoiceID != nil && invoiceType == REGULAR {
        return nil, errors.New("Rechnung existiert bereits für dieses Projekt")
    }

    // 1. Neue Rechnungsnummer
    number, _ := svc.AllocateInvoiceNumber(ctx, invoiceType)

    // 2. Rechnung-Header
    invoice := &Invoice{
        ID: generateUUID(),
        Number: number,
        Type: invoiceType,
        IssuedDate: time.Now(),
        IssuedBy: getUserID(ctx),
        PerformanceDateFrom: project.ActualStartDate,
        PerformanceDateTo: project.ActualEndDate,
        ProjectID: &projectID,
        OfferID: project.OfferID,
        CustomerID: project.CustomerID,
        PaymentTermDays: 30,  // default
        DueDate: time.Now().AddDate(0, 0, 30),
        Status: DRAFT,
    }

    // 3. Positionen aus Projekt übernehmen
    //    → Was wurde WIRKLICH geliefert? (nicht nur geplant)
    for i, eq := range project.EquipmentItems {
        pos := &InvoicePosition{
            ID: generateUUID(),
            PositionNumber: i + 1,
            Type: EQUIPMENT,
            Description: fmt.Sprintf("%s, %dx, %.1f Tage",
                eq.EquipmentName, eq.Quantity, eq.PlannedDays),
            Quantity: float64(eq.Quantity),
            Unit: "Stück",
            UnitPrice: eq.UnitPrice,
            Amount: float64(eq.Quantity) * eq.UnitPrice,
            EquipmentID: &eq.EquipmentID,
        }
        invoice.Positions = append(invoice.Positions, pos)
    }

    // 4. Crew-Kosten aus Zeiterfassung
    crewRecords, _ := svc.crewService.GetProjectTimeEntries(ctx, projectID)
    crewIndex := len(invoice.Positions)
    for _, record := range crewRecords {
        hours := record.EndTime.Sub(record.StartTime).Hours()

        // Berechne Tagesrate oder Stundensatz
        dayRate := 300.0  // default
        if record.FreelancerID != "" {
            freelancer, _ := svc.crewService.GetFreelancer(ctx, record.FreelancerID)
            dayRate = freelancer.DailyRate
        }

        // Nachtzuschlag?
        nightHours := 0.0
        dayHours := 0.0
        // ... komplexe Logik für Nacht-Berechnung

        amount := (dayHours * (dayRate / 8.0)) + (nightHours * (dayRate / 8.0) * 1.25)

        pos := &InvoicePosition{
            PositionNumber: crewIndex + 1,
            Type: CREW,
            Description: fmt.Sprintf("Tontechniker %s, %.1f Stunden",
                record.FreelancerName, hours),
            Quantity: hours,
            Unit: "Stunde",
            UnitPrice: dayRate / 8.0,
            Amount: amount,
        }
        invoice.Positions = append(invoice.Positions, pos)
        crewIndex++
    }

    // 5. Differenzen zu Angebot prüfen (Marco Review)
    offer, _ := svc.offerService.GetOffer(ctx, *project.OfferID)
    if svc.hasSignificantDifferences(invoice, offer) {
        invoice.InternalNotes = "WARNUNG: Abweichung vom Angebot. Bitte prüfen!"
    }

    // 6. MwSt berechnen
    taxRate, _ := svc.DetermineTaxRate(ctx, invoice)
    invoice.TaxRate = taxRate
    svc.CalculateInvoiceTotals(invoice)

    // 7. Speichern
    svc.repo.CreateInvoice(ctx, invoice)

    // 8. Projekt aktualisieren
    svc.projectService.UpdateProjectInvoice(ctx, projectID, invoice.ID)

    return invoice, nil
}

// Kontrolle: Unterschiede erkennen
func (svc *InvoiceService) hasSignificantDifferences(
    invoice *Invoice,
    offer *Offer,
) bool {

    // Vergleiche Positionen
    offerNetto := 0.0
    invoiceNetto := 0.0

    for _, pos := range offer.Positions {
        offerNetto += pos.Amount
    }

    for _, pos := range invoice.Positions {
        invoiceNetto += pos.Amount
    }

    // Wenn Differenz > 5%, als significant markieren
    diff := (invoiceNetto - offerNetto) / offerNetto
    return diff > 0.05 || diff < -0.05
}
```

### 3.6 Gutschrift & Storno

```go
// Gutschrift (positiv, für Kunde)
// Szenario: 2 Moving Heads funktionieren nicht, Kunde bekommt Rabatt über 600€
func (svc *InvoiceService) CreateCreditNote(
    ctx context.Context,
    originalInvoiceID string,
    amount float64,
    reason string,
) (*Invoice, error) {

    original, _ := svc.repo.GetInvoice(ctx, originalInvoiceID)

    creditNote := &Invoice{
        ID: generateUUID(),
        Number: svc.AllocateInvoiceNumber(ctx, CREDIT_NOTE),
        Type: CREDIT_NOTE,
        IssuedDate: time.Now(),
        PreviousInvoiceID: &originalInvoiceID,
        CustomerID: original.CustomerID,
        TotalBrutto: -amount,  // NEGATIV!
        Status: FINALIZED,
        InternalNotes: fmt.Sprintf("Gutschrift: %s. Bezug auf Rechnung %s",
            reason, original.Number),
    }

    // Position für die Gutschrift
    pos := &InvoicePosition{
        PositionNumber: 1,
        Type: SERVICE,
        Description: reason,
        Quantity: 1,
        Unit: "Pauschal",
        UnitPrice: -amount,
        Amount: -amount,
    }
    creditNote.Positions = append(creditNote.Positions, pos)

    svc.repo.CreateInvoice(ctx, creditNote)

    // Kunde zahlt weniger
    original.RemainingAmount -= amount
    svc.repo.UpdateInvoice(ctx, original)

    return creditNote, nil
}

// Storno-Rechnung (wenn ganze Original-Rechnung ungültig)
// Szenario: Rechnung RE-2026-0001 wurde mit falschem Preis versendet
//           → RE-2026-0001-ST (Storno) zum Vorzeichen-Umkehren
func (svc *InvoiceService) CreateReversalInvoice(
    ctx context.Context,
    originalInvoiceID string,
    reason string,
) (*Invoice, error) {

    original, _ := svc.repo.GetInvoice(ctx, originalInvoiceID)

    // Storno = exakte Kopie mit negativen Beträgen
    reversal := &Invoice{
        ID: generateUUID(),
        Number: fmt.Sprintf("%s-RG", original.Number),
        Type: REVERSAL,
        IssuedDate: time.Now(),
        PreviousInvoiceID: &originalInvoiceID,
        CustomerID: original.CustomerID,
        TotalBrutto: -original.TotalBrutto,  // NEGATIV
        Status: FINALIZED,
        InternalNotes: fmt.Sprintf("Storno-Rechnung: %s. Bezug auf Rechnung %s",
            reason, original.Number),
    }

    // Positionen kopieren mit negativem Vorzeichen
    for i, pos := range original.Positions {
        storno := &InvoicePosition{
            PositionNumber: i + 1,
            Type: pos.Type,
            Description: pos.Description,
            Quantity: pos.Quantity,
            Unit: pos.Unit,
            UnitPrice: -pos.UnitPrice,
            Amount: -pos.Amount,
        }
        reversal.Positions = append(reversal.Positions, storno)
    }

    svc.repo.CreateInvoice(ctx, reversal)

    // Original als CANCELLED markieren
    original.Status = CANCELLED
    original.CancelledAt = timePtr(time.Now())
    original.CancellationReason = &reason
    svc.repo.UpdateInvoice(ctx, original)

    return reversal, nil
}
```

---

## 4. MAHNWESEN (Dunning Module D1-D4)

### 4.1 Mahnungs-Datenmodell

```go
type DunningConfiguration struct {
    ID                  string
    Enabled             bool
    AutomaticSend       bool            // oder manuell?
    Levels              []DunningLevel
    Timezone            string          // "Europe/Berlin"
    ReceiversEmail      string          // Thomas' Mail für Benachrichtigungen
}

type DunningLevel struct {
    Level               int              // 1, 2, 3...
    Name                string           // "Zahlungserinnerung", "1. Mahnung", "2. Mahnung"
    DaysAfterDueDate    int              // z.B. 7, 21, 35
    ReminderTemplate    string           // HTML-Template für PDF
    FeeAmount           float64          // Mahngebühr (wenn überhaupt)
    FeeAppliedAfterLevel int             // ab welcher Level?
    NewPaymentTermDays  int              // neues Zahlungsziel (z.B. 10)
    AutomaticSend       bool             // diese Level automatisch versenden?
}

type DunningRecord struct {
    ID                  string
    InvoiceID           string
    Level               int
    ScheduledDate       time.Time        // wann sollte Mahnung versendet werden
    SentDate            *time.Time       // wirklich versendet?
    SentTo              *string          // E-Mail-Adresse
    PDFPath             *string          // gespeicherte PDF
    Status              DunningStatus    // SCHEDULED, SENT, PAUSED
    PausedUntil         *time.Time       // wenn pausiert (Reklamation)
    PausedReason        *string
    CreatedBy           string           // wer hat Mahnung erstellt
    CreatedAt           time.Time
}

type DunningStatus string
const (
    SCHEDULED DunningStatus = "SCHEDULED"
    SENT      DunningStatus = "SENT"
    PAUSED    DunningStatus = "PAUSED"
    CANCELLED DunningStatus = "CANCELLED"
)

// Open Items mit Mahnungs-Status
type OpenItem struct {
    InvoiceID           string
    InvoiceNumber       string
    CustomerName        string
    Amount              float64
    IssuedDate          time.Time
    DueDate             time.Time
    DaysOverdue         int
    PaidAmount          float64
    RemainingAmount     float64
    CurrentDunningLevel int
    NextDunningDate     *time.Time
    Status              OpenItemStatus   // CURRENT, OVERDUE, PAUSED, DISPUTED
}

type OpenItemStatus string
const (
    CURRENT   OpenItemStatus = "CURRENT"    // noch nicht überfällig
    OVERDUE   OpenItemStatus = "OVERDUE"    // überfällig
    PAUSED    OpenItemStatus = "PAUSED"     // Mahnung pausiert
    DISPUTED  OpenItemStatus = "DISPUTED"   // Kunde reklamiert
)
```

### 4.2 Mahnungs-Workflow & Automatisierung

```go
// Täglich um 06:00 Uhr: Dunning-Check
// (Cron-Job oder Cloud Function)
func (svc *DunningService) ProcessDailyDunning(ctx context.Context) error {

    config, _ := svc.repo.GetDunningConfig(ctx)
    if !config.Enabled || !config.AutomaticSend {
        return nil
    }

    // 1. Alle offenen Rechnungen laden
    openInvoices, _ := svc.invoiceService.GetOpenInvoices(ctx)

    for _, invoice := range openInvoices {
        // 2. Prüfe: Ist Mahnung-Pause aktiv?
        if invoice.DunningPausedUntil != nil && time.Now().Before(*invoice.DunningPausedUntil) {
            continue  // pausiert, skip
        }

        // 3. Berechne: Wie viele Tage überfällig?
        daysOverdue := int(time.Since(invoice.DueDate).Hours() / 24)

        // 4. Bestimme nächste Mahnung-Level
        nextLevel := svc.determineNextLevel(invoice.DunningLevel, daysOverdue, config.Levels)

        if nextLevel == nil {
            continue  // keine weitere Mahnung nötig
        }

        // 5. Mahnung erstellen & versenden
        dunning, _ := svc.CreateAndSendDunningNotice(ctx, invoice.ID, nextLevel)

        // 6. Invoice aktualisieren
        invoice.DunningLevel = nextLevel.Level
        svc.invoiceService.UpdateInvoice(ctx, invoice)

        // 7. Event für Notification
        svc.pubsub.Publish(ctx, &DunningIssuedEvent{
            InvoiceID: invoice.ID,
            DunningLevel: nextLevel.Level,
            Timestamp: time.Now(),
        })
    }

    return nil
}

func (svc *DunningService) determineNextLevel(
    currentLevel int,
    daysOverdue int,
    levels []DunningLevel,
) *DunningLevel {

    for _, level := range levels {
        if level.Level > currentLevel && daysOverdue >= level.DaysAfterDueDate {
            return &level
        }
    }
    return nil
}

// Mahnung erstellen und versenden
func (svc *DunningService) CreateAndSendDunningNotice(
    ctx context.Context,
    invoiceID string,
    level *DunningLevel,
) (*DunningRecord, error) {

    invoice, _ := svc.invoiceService.GetInvoice(ctx, invoiceID)

    // 1. Mahnungs-PDF generieren
    pdfBytes, _ := svc.generateDunningPDF(ctx, invoice, level)

    // 2. PDF speichern
    pdfPath := fmt.Sprintf("dunning/%s/level-%d.pdf",
        invoiceID, level.Level)
    svc.storage.Upload(ctx, pdfPath, pdfBytes)

    // 3. DunningRecord erstellen
    record := &DunningRecord{
        ID: generateUUID(),
        InvoiceID: invoiceID,
        Level: level.Level,
        ScheduledDate: time.Now(),
        SentDate: timePtr(time.Now()),
        SentTo: &invoice.CustomerContactEmail,
        PDFPath: &pdfPath,
        Status: SENT,
        CreatedBy: "SYSTEM",
        CreatedAt: time.Now(),
    }
    svc.repo.CreateDunningRecord(ctx, record)

    // 4. E-Mail versenden
    svc.mailService.Send(ctx, &MailRequest{
        To: invoice.CustomerContactEmail,
        Subject: fmt.Sprintf("Mahnung: Rechnung %s", invoice.Number),
        Body: fmt.Sprintf(`
            Sehr geehrte Damen und Herren,

            anbei erhalten Sie unsere %s zu Rechnung %s vom %s.

            Offener Betrag: %.2f €
            Neues Zahlungsziel: %s

            Bitte begleichen Sie die Zahlung unverzüglich.

            Mit freundlichen Grüßen,
            MB Veranstaltungstechnik GmbH
        `,
            level.Name, invoice.Number,
            invoice.IssuedDate.Format("02.01.2006"),
            invoice.RemainingAmount,
            time.Now().AddDate(0, 0, level.NewPaymentTermDays).Format("02.01.2006"),
        ),
        Attachments: []string{pdfPath},
    })

    return record, nil
}

// Mahnung pausieren (z.B. Kunde reklamiert)
func (svc *DunningService) PauseDunning(
    ctx context.Context,
    invoiceID string,
    pausedUntil time.Time,
    reason string,
) error {

    invoice, _ := svc.invoiceService.GetInvoice(ctx, invoiceID)
    invoice.DunningPausedUntil = &pausedUntil
    svc.invoiceService.UpdateInvoice(ctx, invoice)

    // Event
    svc.pubsub.Publish(ctx, &DunningPausedEvent{
        InvoiceID: invoiceID,
        Reason: reason,
        ResumeAt: pausedUntil,
    })

    return nil
}

// Offene-Posten-Dashboard (für Thomas)
func (svc *DunningService) GetOpenItemsDashboard(
    ctx context.Context,
) (*OpenItemsDashboard, error) {

    openInvoices, _ := svc.invoiceService.GetOpenInvoices(ctx)

    var items []OpenItem
    var totalOverdue, totalCurrent, totalPaused float64

    for _, inv := range openInvoices {
        daysOverdue := int(time.Since(inv.DueDate).Hours() / 24)
        status := CURRENT

        if inv.DunningPausedUntil != nil && time.Now().Before(*inv.DunningPausedUntil) {
            status = PAUSED
        } else if daysOverdue > 0 {
            status = OVERDUE
        }

        item := OpenItem{
            InvoiceID: inv.ID,
            InvoiceNumber: inv.Number,
            CustomerName: inv.CustomerName,
            Amount: inv.TotalBrutto,
            DaysOverdue: daysOverdue,
            RemainingAmount: inv.RemainingAmount,
            CurrentDunningLevel: inv.DunningLevel,
            Status: status,
        }
        items = append(items, item)

        // Summation
        if status == OVERDUE {
            totalOverdue += item.RemainingAmount
        } else if status == CURRENT {
            totalCurrent += item.RemainingAmount
        } else {
            totalPaused += item.RemainingAmount
        }
    }

    // Sortierung: überfällig oben
    sort.Slice(items, func(i, j int) bool {
        if items[i].Status == OVERDUE && items[j].Status != OVERDUE {
            return true
        }
        return items[i].DaysOverdue > items[j].DaysOverdue
    })

    return &OpenItemsDashboard{
        Items: items,
        Summary: struct{}{
            TotalOverdue: totalOverdue,
            TotalCurrent: totalCurrent,
            TotalPaused: totalPaused,
            CountOverdue: countByStatus(items, OVERDUE),
            CountCurrent: countByStatus(items, CURRENT),
        },
    }, nil
}
```

---

## 5. DATEV-EXPORT (Accounting Module A1-A3)

### 5.1 Export-Format & Konten-Mapping

```go
// DATEV Buchungsstapel (ASCII CSV)
// Format: Feldtrenner = Semikolon, Dezimal = Komma

type DATEVExportRow struct {
    BelegDatum          string  // "28.03.2026" (Rechnungsdatum)
    BelegFeld1          string  // Rechnungsnummer
    BelegFeld2          string  // optional weitere Info
    Buchungstext        string  // "Verkauf Ton-Equipment 19%" oder "Eingangsrechnung"
    Umsatz              string  // "30000,00" (immer Netto!)
    SollHabenKennzeichen string // "S" = Soll, "H" = Haben
    Sollkonto           string  // z.B. "1400" (Forderungen)
    Gegenkonto          string  // z.B. "8400" (Erlöse 19%)
    BUSchluessel        string  // BU-Schlüssel (optional)
}

// Kontenrahmen SKR03 (Standard für Mittelstand)
type ChartsOfAccounts struct {
    // Aktiva
    BankkontoEN         string  // "1100" (EN = Einzelkonten)
    BargeldenEN         string  // "1110"
    ForderungenEN       string  // "1400"
    ForderungenVA       string  // "1410" (für Vorgangskonten)

    // Passiva
    DarlehensEN         string  // "2700"

    // Erträge (Betriebsergebnis)
    ErloeseMitMWSt19    string  // "8400" (Umsatzerlöse 19%)
    ErloeseMitMWSt7     string  // "8300" (Umsatzerlöse 7%)
    ErloeseMitMWSt0     string  // "8200" (Umsatzerlöse 0% / Reverse Charge)
    NebenerloeseEN      string  // "8500"

    // Aufwendungen
    WareneinkaufEN      string  // "4000"
    PersonalkostenEN    string  // "4100" (oder je nach Konfiguration)
    TransportEN         string  // "4930"

    // MwSt
    VorsteuervEN        string  // "1576" (Vorsteuer 19%)
    VorsteuervEN7       string  // "1406" (Vorsteuer 7%)
    UmsatzsteuerEN      string  // "1700" (zu zahlende Umsatzsteuer)
}

// Konfigurierbar
type DATEVExportConfig struct {
    SKR                 string  // "SKR03" oder "SKR04"
    Accounts            ChartsOfAccounts
    ExportSeparateDebtors bool // Kundenkonten separat (1401, 1402, etc.)
    DisableMWST         bool   // "Kleinunternehmer" → keine MwSt-Trennung
}
```

### 5.2 Export-Logik (monatlich)

```go
func (svc *AccountingService) ExportDATEV(
    ctx context.Context,
    fromDate time.Time,
    toDate time.Time,
) ([]byte, error) {

    config, _ := svc.repo.GetDATEVConfig(ctx)

    // 1. Alle Rechnungen im Zeitraum laden (Status = SENT oder PAID)
    invoices, _ := svc.invoiceService.GetInvoicesByDateRange(ctx, fromDate, toDate)

    // 2. CSV-Zeilen generieren
    var rows []DATEVExportRow

    for _, inv := range invoices {
        // Pro Rechnung: eine Zeile pro MwSt-Satz-Gruppe

        // Beispiel Rechnung RE-2026-0042: 30.000€ Netto (19% MwSt)
        // → 2 Zeilen in DATEV:
        //   Zeile 1: 30.000€ Soll auf 1400 (Forderungen), Haben auf 8400 (Erlöse)
        //   Zeile 2: 5.700€ Soll auf 1700 (Umsatzsteuer), Haben auf 8400 (Erlöse)
        //
        // ODER kompakter:
        //   Zeile 1: 30.000€ Haben auf 8400 (Erlöse 19%), Soll auf 1400
        //
        // Konvention: von Buchhalterperspektive
        //   - Forderung (1400) = Soll-Konto (Aufzählung zu Kunden)
        //   - Erlöse (8400) = Haben-Konto (Einnahmen)

        nettoAmount := inv.SubtotalNetto

        // Gruppiere Positionen nach MwSt-Satz
        positionsByRate := groupBy(inv.Positions, func(p InvoicePosition) float64 {
            return p.TaxRate
        })

        for rate, positions := range positionsByRate {
            var totalForRate float64
            for _, pos := range positions {
                totalForRate += pos.Amount
            }

            // Erlös-Konto bestimmen (je nach MwSt-Satz)
            revenueAccount := ""
            if rate == 0.19 {
                revenueAccount = config.Accounts.ErloeseMitMWSt19  // "8400"
            } else if rate == 0.07 {
                revenueAccount = config.Accounts.ErloeseMitMWSt7   // "8300"
            } else {
                revenueAccount = config.Accounts.ErloeseMitMWSt0   // "8200"
            }

            // Zeile 1: Forderung
            rows = append(rows, DATEVExportRow{
                BelegDatum: inv.IssuedDate.Format("02.01.2006"),
                BelegFeld1: inv.Number,
                Buchungstext: fmt.Sprintf("Umsatz Equipment & Dienstleistungen (MwSt %d%%)", int(rate*100)),
                Umsatz: formatDATEVAmount(totalForRate),
                SollHabenKennzeichen: "S",  // Soll
                Sollkonto: config.Accounts.ForderungenEN,  // "1400"
                Gegenkonto: revenueAccount,
            })

            // Zeile 2: MwSt
            if rate > 0 {
                taxAmount := totalForRate * rate

                vstAccount := ""
                if rate == 0.19 {
                    vstAccount = config.Accounts.UmsatzsteuerEN  // "1700" (zu zahlen!)
                } else if rate == 0.07 {
                    vstAccount = config.Accounts.UmsatzsteuerEN
                }

                rows = append(rows, DATEVExportRow{
                    BelegDatum: inv.IssuedDate.Format("02.01.2006"),
                    BelegFeld1: inv.Number + "-USt",
                    Buchungstext: fmt.Sprintf("Umsatzsteuer %d%%", int(rate*100)),
                    Umsatz: formatDATEVAmount(taxAmount),
                    SollHabenKennzeichen: "S",
                    Sollkonto: vstAccount,  // "1700"
                    Gegenkonto: revenueAccount,
                })
            }
        }
    }

    // 3. Eingangsrechnungen (Ausgaben)
    purchaseInvoices, _ := svc.purchaseInvoiceService.GetByDateRange(ctx, fromDate, toDate)

    for _, pinv := range purchaseInvoices {
        // Beispiel: SoundPro Sub-Rental 4.800€ Netto (19% MwSt)
        // → Zeile 1: 4.800€ Soll auf 4930 (Transport/Fremdleistungen), Haben auf 1100 (Bank)
        // → Zeile 2: 912€ Soll auf 1576 (Vorsteuer 19%), Haben auf 1100

        nettoAmount := pinv.SubtotalNetto

        // Ausgaben-Konto bestimmen (nach Typ)
        expenseAccount := ""
        if strings.Contains(pinv.Description, "Transport") {
            expenseAccount = config.Accounts.TransportEN  // "4930"
        } else {
            expenseAccount = config.Accounts.WareneinkaufEN  // "4000"
        }

        // Zeile 1: Ausgabe
        rows = append(rows, DATEVExportRow{
            BelegDatum: pinv.InvoiceDate.Format("02.01.2006"),
            BelegFeld1: pinv.InvoiceNumber,
            Buchungstext: fmt.Sprintf("Eingangsrechnung %s", pinv.VendorName),
            Umsatz: formatDATEVAmount(nettoAmount),
            SollHabenKennzeichen: "S",
            Sollkonto: expenseAccount,
            Gegenkonto: config.Accounts.BankkontoEN,  // "1100"
        })

        // Zeile 2: Vorsteuer
        if pinv.VATRate > 0 {
            vstAmount := nettoAmount * pinv.VATRate

            vstAccount := ""
            if pinv.VATRate == 0.19 {
                vstAccount = config.Accounts.VorsteuervEN  // "1576"
            } else if pinv.VATRate == 0.07 {
                vstAccount = config.Accounts.VorsteuervEN7  // "1406"
            }

            rows = append(rows, DATEVExportRow{
                BelegDatum: pinv.InvoiceDate.Format("02.01.2006"),
                BelegFeld1: pinv.InvoiceNumber + "-VSt",
                Buchungstext: fmt.Sprintf("Vorsteuer %d%%", int(pinv.VATRate*100)),
                Umsatz: formatDATEVAmount(vstAmount),
                SollHabenKennzeichen: "S",
                Sollkonto: vstAccount,
                Gegenkonto: config.Accounts.BankkontoEN,
            })
        }
    }

    // 4. CSV generieren
    csv := generateDATEVCSV(rows)

    // 5. Validierung
    errors := validateDATEVExport(rows)
    if len(errors) > 0 {
        return nil, fmt.Errorf("DATEV Export Fehler: %v", errors)
    }

    // 6. Archivierung markieren
    for _, inv := range invoices {
        inv.ArchivedInDATEV = true
        inv.DATEVExportDate = timePtr(time.Now())
        svc.invoiceService.UpdateInvoice(ctx, inv)
    }

    return csv, nil
}

func formatDATEVAmount(amount float64) string {
    // 30000.50 → "30000,50" (mit Komma!)
    return strings.Replace(fmt.Sprintf("%.2f", amount), ".", ",", 1)
}

func generateDATEVCSV(rows []DATEVExportRow) []byte {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)
    writer.Comma = ';'  // Semikolon als Trennzeichen

    // Header
    writer.Write([]string{
        "Belegdatum",
        "Belegfeld 1",
        "Belegfeld 2",
        "Buchungstext",
        "Umsatz",
        "Soll/Haben-Kennzeichen",
        "Sollkonto",
        "Gegenkonto",
        "BU-Schlüssel",
    })

    // Daten
    for _, row := range rows {
        writer.Write([]string{
            row.BelegDatum,
            row.BelegFeld1,
            row.BelegFeld2,
            row.Buchungstext,
            row.Umsatz,
            row.SollHabenKennzeichen,
            row.Sollkonto,
            row.Gegenkonto,
            row.BUSchluessel,
        })
    }

    writer.Flush()
    return buf.Bytes()
}

func validateDATEVExport(rows []DATEVExportRow) []string {
    var errors []string

    for i, row := range rows {
        // Prüfe: Sollkonto und Gegenkonto sind verschiedene Konten
        if row.Sollkonto == row.Gegenkonto {
            errors = append(errors, fmt.Sprintf("Row %d: Sollkonto = Gegenkonto", i+1))
        }

        // Prüfe: Umsatz ist nicht 0
        if row.Umsatz == "0,00" {
            errors = append(errors, fmt.Sprintf("Row %d: Umsatz = 0", i+1))
        }

        // Prüfe: Soll/Haben ist "S" oder "H"
        if row.SollHabenKennzeichen != "S" && row.SollHabenKennzeichen != "H" {
            errors = append(errors, fmt.Sprintf("Row %d: Ungültige Soll/Haben-Kennzeichen", i+1))
        }
    }

    return errors
}
```

---

## 6. CREW DOMAIN MODEL (HR Module H1-H7)

### 6.1 Mitarbeiter vs. Freelancer Datenmodell

```go
type CrewMember struct {
    ID                  string
    Type                CrewType        // EMPLOYEE oder FREELANCER
    FirstName           string
    LastName            string
    Email               string
    Phone               string

    // Persönlich
    BirthDate           time.Time       // optional, für Lohnsteuer
    Address             string
    City                string
    ZIP                 string
    Country             string

    // Qualifikationen & Rollen
    Qualifications      []Qualification
    Languages           []string        // ["Deutsch", "Englisch", "Französisch"]
    Notes               string          // "Kann auch Rigging", "Nur Wochenenden"

    // Verfügbarkeit
    CalendarID          string          // für CalDAV
    CalendarURL         string          // "webcal://..." oder "iCal" Feed
    Unavailable         []UnavailablePeriod // Urlaub, Krankheit

    // Nur Freelancer
    SelfEmployed        *SelfEmployedInfo
    TaxID               *string         // USt-IdNr
    BankAccount         *string         // IBAN
    HourlyRate          *float64        // €/Stunde (alternativ DailyRate)
    DailyRate           *float64        // €/Tag
    NightShiftBonus     *float64        // % Zuschlag 22-6 Uhr
    OvertimeBonus       *float64        // % Zuschlag >10h/Tag
    InvoicingFrequency  *string         // "WEEKLY", "MONTHLY"
    PaymentTerms        *string         // "Net 7", "Net 30"

    // Nur Mitarbeiter
    EmploymentContract  *EmploymentInfo
    Salary              *float64        // €/Monat
    TaxClass            *string         // "I", "III", "IV", etc.
    HealthInsurance     *string         // "Krankenkasse", optional
    SocialSecurityID    *string         // SVNR

    // Meta
    Active              bool
    CreatedAt           time.Time
    UpdatedAt           time.Time
    CreatedBy           string
}

type CrewType string
const (
    EMPLOYEE    CrewType = "EMPLOYEE"
    FREELANCER  CrewType = "FREELANCER"
)

type Qualification struct {
    ID                  string
    Name                string              // "Tontechniker", "Lichttechniker", "Rigging"
    Certification       *string             // "Fachkraft für Veranstaltungstechnik"
    CertificationDate   *time.Time          // wann erworben?
    ExpiryDate          *time.Time          // Auffrischung nötig?
    Level               string              // "Junior", "Senior", "Master"
    Notes               string
}

type SelfEmployedInfo struct {
    CompanyName         string              // z.B. "Kevin Radic - Freelance Veranstaltungstechnik"
    USTProbeCertificate bool
    DATABankNumber      *string             // optional Behörden-ID
}

type EmploymentInfo struct {
    StartDate           time.Time
    EndDate             *time.Time          // NULL = noch aktiv
    Position            string
    Department          string
    ContractType        string              // "Vollzeit", "Teilzeit", "Mini-Job"
    HolidayDays         int                 // Urlaubstage pro Jahr
    HolidayUsed         int                 // bisher genommen
}

type UnavailablePeriod struct {
    StartDate           time.Time
    EndDate             time.Time
    Type                string              // "VACATION", "SICK", "TRAINING", "OTHER"
    Description         string
    ApprovedBy          *string
    ApprovedAt          *time.Time
}
```

### 6.2 Qualifikationen & Skill-Matching

```go
type ProjectCrewRequirement struct {
    ID                  string
    ProjectID           string
    Role                string              // "Tontechniker", "Lichttechniker", "Generalmanager"
    Quantity            int                 // wie viele?
    RequiredQualifications []string         // ["Ton", "PA-Systeme", "Monitoring"]
    PreferredQualifications []string        // ["Deutschkenntnisse", "PEAK Experience"]
    DailyRate           float64
    From                time.Time
    To                  time.Time
    DaysRequired        float64             // 3.5 Tage?
    Status              CrewAssignmentStatus  // REQUESTED, OFFERED, CONFIRMED, REJECTED, COMPLETED

    // Zugeordnete Crew
    Assignments         []CrewAssignment
}

type CrewAssignment struct {
    ID                  string
    ProjectID           string
    CrewMemberID        string
    RequirementID       string              // FK zu ProjectCrewRequirement

    // Bestätigung
    Status              CrewAssignmentStatus
    OfferedAt           *time.Time
    ConfirmedAt         *time.Time
    ConfirmedBy         string              // Crew-Member selbst oder Marco?
    RejectedAt          *time.Time
    RejectionReason     *string
    RejectedBy          *string

    // Zeiterfassung
    TimeEntries         []TimeEntry
    TotalHours          float64             // berechnet
    TotalAmount         float64             // Netto

    // Notifications
    NotificationSentAt  *time.Time
    ReminderSentAt      *time.Time          // 24h vorher

    // Meta
    Notes               string
    CreatedAt           time.Time
}

type CrewAssignmentStatus string
const (
    REQUESTED   CrewAssignmentStatus = "REQUESTED"    // Marco anfragen?
    OFFERED     CrewAssignmentStatus = "OFFERED"      // Angeboten, warte auf Antwort
    CONFIRMED   CrewAssignmentStatus = "CONFIRMED"    // Freelancer hat zugesagt
    REJECTED    CrewAssignmentStatus = "REJECTED"     // Freelancer lehnt ab
    COMPLETED   CrewAssignmentStatus = "COMPLETED"    // Projekt fertig, bezahlt
    CANCELLED   CrewAssignmentStatus = "CANCELLED"    // Marco hat abgebrochen
)

// Skill-Matching Service
func (svc *CrewService) FindAvailableCrew(
    ctx context.Context,
    requirement *ProjectCrewRequirement,
) ([]CrewMatch, error) {

    // 1. Alle Crew mit erforderlichen Qualifikationen
    candidates, _ := svc.repo.GetCrewByQualifications(ctx,
        requirement.RequiredQualifications)

    // 2. Verfügbarkeit prüfen
    var available []CrewMatch
    for _, crew := range candidates {
        avail := svc.checkAvailability(crew.ID,
            requirement.From, requirement.To)

        if avail {
            // Skill-Score berechnen
            score := svc.calculateSkillScore(crew, requirement)

            available = append(available, CrewMatch{
                CrewMember: crew,
                SkillScore: score,
                MatchReason: fmt.Sprintf("%.1f%% Match", score*100),
            })
        }
    }

    // 3. Nach Skill-Score sortieren (beste oben)
    sort.Slice(available, func(i, j int) bool {
        return available[i].SkillScore > available[j].SkillScore
    })

    return available, nil
}

func (svc *CrewService) calculateSkillScore(
    crew *CrewMember,
    requirement *ProjectCrewRequirement,
) float64 {

    score := 0.0
    matched := 0.0

    // Required Skills: 70% Weight
    for _, reqSkill := range requirement.RequiredQualifications {
        for _, crewQual := range crew.Qualifications {
            if strings.EqualFold(crewQual.Name, reqSkill) {
                matched += 0.7 / float64(len(requirement.RequiredQualifications))
                break
            }
        }
    }

    // Preferred Skills: 30% Weight
    for _, prefSkill := range requirement.PreferredQualifications {
        for _, crewQual := range crew.Qualifications {
            if strings.EqualFold(crewQual.Name, prefSkill) {
                matched += 0.3 / float64(len(requirement.PreferredQualifications))
                break
            }
        }
    }

    return matched
}
```

### 6.3 Zeiterfassung & Abrechnung

```go
type TimeEntry struct {
    ID                  string
    CrewMemberID        string
    ProjectID           string
    AssignmentID        string

    // Zeiten
    StartTime           time.Time        // "28.03.2026 07:00"
    EndTime             time.Time        // "28.03.2026 19:30"
    Duration            float64          // berechnet in Stunden
    BreakMinutes        int              // Pausen
    ActualWorkHours     float64          // nach Abzug Pausen

    // Ort & Kontext
    LocationName        string           // "Stadtpark Marktbühne"
    LocationGPS         *struct{
        Latitude        float64
        Longitude       float64
    }
    CheckInMethod       string           // "MANUAL", "GPS", "QR_CODE"

    // Besonderheiten
    NightHours          float64          // Stunden zw. 22-6 Uhr (mit Zuschlag)
    OvertimeHours       float64          // Stunden >10h/Tag (mit Zuschlag)
    SpecialTasks        []string         // ["Rigging", "Safety-Briefing"]
    SpecialBonus        float64          // % wenn nötig

    // Abrechnung
    BasalHourlyRate     float64          // €/h Grundsatz
    NightHourlyRate     float64          // €/h mit 25% Zuschlag
    OvertimeHourlyRate  float64          // €/h mit 50% Zuschlag
    CalculatedAmount    float64          // Netto € (in Rechnung)

    // Dokumentation
    Notes               string           // "Kranführer war nicht da, musste Setup machen"
    EquipmentDamage     *DefectReport    // Falls Gerät kaputt gemacht
    ApprovedBy          *string          // Marco approval
    ApprovedAt          *time.Time

    // Meta
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

// Zeiterfassung berechnen (komplexe Logik)
func (svc *CrewService) CalculateTimeEntryAmount(
    ctx context.Context,
    entry *TimeEntry,
) error {

    crew, _ := svc.repo.GetCrewMember(ctx, entry.CrewMemberID)

    // 1. Basis: Tagesrate oder Stundensatz?
    basalRate := 0.0
    if crew.DailyRate != nil && entry.ActualWorkHours <= 8 {
        basalRate = *crew.DailyRate / 8.0  // umrechnen auf Stundensatz
    } else if crew.HourlyRate != nil {
        basalRate = *crew.HourlyRate
    }

    // 2. Nachtstunden berechnen (22-6 Uhr)
    nightHours := calculateNightHours(entry.StartTime, entry.EndTime)
    dayHours := entry.ActualWorkHours - nightHours

    // 3. Überstunden (>10h/Tag)
    overtimeHours := 0.0
    if entry.ActualWorkHours > 10 {
        overtimeHours = entry.ActualWorkHours - 10
    }

    // 4. Beträge berechnen
    dayAmount := dayHours * basalRate
    nightAmount := nightHours * basalRate * 1.25  // 25% Zuschlag
    overtimeAmount := overtimeHours * basalRate * 1.50  // 50% Zuschlag

    entry.CalculatedAmount = dayAmount + nightAmount + overtimeAmount

    return nil
}

// Mobile Check-In/Check-Out
func (svc *CrewService) CheckIn(
    ctx context.Context,
    crewMemberID string,
    projectID string,
) error {

    // Prüfe: Hat dieser Crew Member einen Assignment für diesen Tag & Projekt?
    assignment, _ := svc.repo.GetTodaysAssignment(ctx, crewMemberID, projectID)

    if assignment == nil {
        return errors.New("Kein Job zugeordnet")
    }

    // Erstelle TimeEntry
    entry := &TimeEntry{
        ID: generateUUID(),
        CrewMemberID: crewMemberID,
        ProjectID: projectID,
        AssignmentID: assignment.ID,
        StartTime: time.Now(),
    }

    svc.repo.CreateTimeEntry(ctx, entry)

    // Event für Live-Notification
    svc.pubsub.Publish(ctx, &CrewCheckedInEvent{
        CrewID: crewMemberID,
        ProjectID: projectID,
        Timestamp: time.Now(),
    })

    return nil
}

func (svc *CrewService) CheckOut(
    ctx context.Context,
    crewMemberID string,
    projectID string,
) error {

    // Finde aktive TimeEntry
    entry, _ := svc.repo.GetActiveTimeEntry(ctx, crewMemberID, projectID)

    if entry == nil {
        return errors.New("Kein aktiver Check-In")
    }

    entry.EndTime = time.Now()
    entry.Duration = entry.EndTime.Sub(entry.StartTime).Hours()
    entry.ActualWorkHours = entry.Duration - float64(entry.BreakMinutes)/60.0

    // Berechne Betrag
    svc.CalculateTimeEntryAmount(ctx, entry)

    // Speichern
    svc.repo.UpdateTimeEntry(ctx, entry)

    // Event
    svc.pubsub.Publish(ctx, &CrewCheckedOutEvent{
        CrewID: crewMemberID,
        ProjectID: projectID,
        Hours: entry.ActualWorkHours,
        Timestamp: time.Now(),
    })

    return nil
}

// Monatsabschluss: Freelancer-Rechnung generieren
func (svc *CrewService) GenerateFreelancerInvoice(
    ctx context.Context,
    crewMemberID string,
    fromDate time.Time,
    toDate time.Time,
) (*Invoice, error) {

    crew, _ := svc.repo.GetCrewMember(ctx, crewMemberID)

    if crew.Type != FREELANCER {
        return nil, errors.New("Nur Freelancer können Rechnungen stellen")
    }

    // Alle TimeEntries laden
    entries, _ := svc.repo.GetTimeEntries(ctx, crewMemberID, fromDate, toDate)

    // Gruppiere nach Projekt
    byProject := make(map[string][]*TimeEntry)
    for _, entry := range entries {
        if entry.ApprovedBy == nil {
            return nil, fmt.Errorf("TimeEntry %s nicht genehmigt", entry.ID)
        }
        byProject[entry.ProjectID] = append(byProject[entry.ProjectID], entry)
    }

    // Freelancer-Rechnung erstellen
    invoice := &Invoice{
        ID: generateUUID(),
        Number: fmt.Sprintf("FL-%d-%04d", toDate.Year(), /* sequence */ 1),
        Type: REGULAR,
        IssuedDate: time.Now(),
        PerformanceDateFrom: fromDate,
        PerformanceDateTo: toDate,
    }

    // Positionen pro Projekt
    posNum := 1
    totalAmount := 0.0

    for projectID, timeEntries := range byProject {
        project, _ := svc.projectService.GetProject(ctx, projectID)

        projectHours := 0.0
        projectAmount := 0.0

        for _, entry := range timeEntries {
            projectHours += entry.ActualWorkHours
            projectAmount += entry.CalculatedAmount
        }

        pos := &InvoicePosition{
            PositionNumber: posNum,
            Type: CREW,
            Description: fmt.Sprintf("Veranstaltungstechnik für %s, %.1f Stunden",
                project.Name, projectHours),
            Quantity: projectHours,
            Unit: "Stunde",
            Amount: projectAmount,
        }
        invoice.Positions = append(invoice.Positions, pos)
        posNum++
        totalAmount += projectAmount
    }

    invoice.TotalBrutto = totalAmount
    svc.invoiceService.CreateInvoice(ctx, invoice)

    return invoice, nil
}
```

### 6.4 CalDAV Integration & Push-Notifications

```go
// Exportiere Crew-Verfügbarkeit als iCal
func (svc *CrewService) ExportCalendarAsICAL(
    ctx context.Context,
    crewMemberID string,
) (string, error) {

    crew, _ := svc.repo.GetCrewMember(ctx, crewMemberID)

    cal := ics.NewCalendar()
    cal.SetName(fmt.Sprintf("%s %s - RentFlow Einsätze",
        crew.FirstName, crew.LastName))
    cal.SetDescription("Automatisch generiert aus RentFlow")

    // Alle Crew-Zuweisungen laden (nächste 3 Monate)
    assignments, _ := svc.repo.GetCrewAssignments(ctx, crewMemberID,
        time.Now(), time.Now().AddDate(0, 3, 0))

    for _, assignment := range assignments {
        project, _ := svc.projectService.GetProject(ctx, assignment.ProjectID)

        event := ics.NewEvent(assignment.ID)
        event.SetStartAt(project.PlannedStartDate)
        event.SetEndAt(project.PlannedEndDate)
        event.SetSummary(fmt.Sprintf("Job: %s", project.Name))
        event.SetDescription(fmt.Sprintf(
            "Location: %s\nContact: %s\nStatus: %s",
            project.LocationName,
            project.LocationContactPerson,
            assignment.Status,
        ))
        event.SetLocation(project.LocationAddress)
        event.AddProperty("DESCRIPTION",
            fmt.Sprintf("iCal für %s", crew.FirstName))

        cal.AddEvent(event)
    }

    return cal.String(), nil
}

// Push-Notification bei neuer Zuweisung
func (svc *CrewService) NotifyCrewAssignment(
    ctx context.Context,
    assignmentID string,
) error {

    assignment, _ := svc.repo.GetCrewAssignment(ctx, assignmentID)
    crew, _ := svc.repo.GetCrewMember(ctx, assignment.CrewMemberID)
    project, _ := svc.projectService.GetProject(ctx, assignment.ProjectID)

    // Push an Handy
    svc.pushService.Send(ctx, &PushNotification{
        UserID: crew.ID,
        Title: "Neuer Job zugewiesen",
        Body: fmt.Sprintf("%s am %s in %s",
            project.Name,
            project.PlannedStartDate.Format("02. Jan 2006"),
            project.LocationCity,
        ),
        Data: map[string]string{
            "projectID": assignment.ProjectID,
            "action": "VIEW_JOB",
        },
    })

    // E-Mail als Backup
    svc.mailService.Send(ctx, &MailRequest{
        To: crew.Email,
        Subject: fmt.Sprintf("Neue Jobzuweisung: %s", project.Name),
        Body: fmt.Sprintf(`
            Hallo %s,

            dir wurde ein neuer Job zugewiesen:

            Projekt: %s
            Datum: %s - %s
            Ort: %s
            Bezahlung: %.2f €/Tag

            Bitte bestätige deine Zusage in der RentFlow-App.

            Grüße,
            Marco
        `, crew.FirstName, project.Name,
            project.PlannedStartDate.Format("02.01.2006"),
            project.PlannedEndDate.Format("02.01.2006"),
            project.LocationCity,
            assignment.DailyRate,
        ),
    })

    return nil
}

// Reminder 24h vorher
func (svc *CrewService) SendJobReminders(ctx context.Context) error {

    // Cron-Job: täglich um 10:00 Uhr
    // Finde alle Zuweisungen mit StartDate = morgen (innerhalb 24-48h)
    tomorrow := time.Now().AddDate(0, 0, 1)
    dayAfter := time.Now().AddDate(0, 0, 2)

    assignments, _ := svc.repo.GetAssignmentsInDateRange(ctx, tomorrow, dayAfter)

    for _, assignment := range assignments {
        crew, _ := svc.repo.GetCrewMember(ctx, assignment.CrewMemberID)
        project, _ := svc.projectService.GetProject(ctx, assignment.ProjectID)

        svc.pushService.Send(ctx, &PushNotification{
            UserID: crew.ID,
            Title: "Erinnerung: Job morgen!",
            Body: fmt.Sprintf("%s - %s um %s",
                project.Name,
                project.LocationCity,
                project.PlannedStartDate.Format("15:04"),
            ),
            Data: map[string]string{
                "projectID": assignment.ProjectID,
                "action": "VIEW_PACKLIST",
            },
        })

        assignment.ReminderSentAt = timePtr(time.Now())
        svc.repo.UpdateCrewAssignment(ctx, assignment)
    }

    return nil
}
```

---

## 7. CROSS-DOMAIN WORKFLOWS

### 7.1 Projekt abgeschlossen → Rechnung erstellen

```
Trigger: Projekt-Status ändert sich zu "Returned"
         (Equipment ist eingelagert, Crew hat ausgecheckt)

Workflow:
1. Event: ProjectStatusChanged{ProjectID, Status="RETURNED"}
   → Invoice-Service empfängt Event

2. Invoice-Service: CreateInvoiceFromProject(projectID)
   a) Alle ProjectEquipment laden
   b) Alle TimeEntries für diesen Project laden
   c) Positionen zusammenbauen
   d) MwSt berechnen
   e) Rechnung erstellen (Status = DRAFT)
   f) Notification an Thomas: "Neue Rechnung RE-2026-XXXX zur Prüfung"

3. Thomas prüft:
   - Alle Mengen korrekt?
   - Alle Zeiteinträge genehmigt?
   - Unterschiede zum Angebot? (Auto-Warnung wenn >5%)

4. Thomas gibt frei → "Finalize"
   a) Invoice-Status → FINALIZED
   b) PDF generieren
   c) Event: InvoiceFinalized

5. Marco oder System sendet Rechnung per E-Mail
   a) Event: InvoiceSent
   b) Invoice-Status → SENT
   c) DueDate berechnen
   d) Dunning-Prozess startet

6. Projekt-Status → Invoiced (nur noch für Archivierung)
```

### 7.2 Rechnung überfällig → Mahnung automatisch

```
Trigger: Täglich um 06:00 Uhr (Dunning-Service)

Workflow:
1. Alle Rechnungen mit Status SENT, PARTIAL_PAID, OVERDUE laden
2. Für jede Rechnung:
   a) Prüfe: DueDate + DaysAfterDueDate < heute?
   b) Prüfe: DunningPausedUntil > heute? (Reklamation)
   c) Berechne: nextDunningLevel aus Config
   d) Erstelle DunningRecord
   e) Generiere PDF
   f) Versende E-Mail
   g) Update Invoice.DunningLevel
   h) Event: DunningIssuedLevel1

3. Nach 2. Mahnung (Level 3):
   a) Kundenname rot in Thomas' Dashboard
   b) Evtl. Inkasso-Hinweis in Mahnung
```

### 7.3 Equipment defekt auf Job → Gutschrift

```
Trigger: Kevin markiert Equipment als defekt bei Check-In
         (oder Rückgabe)

Workflow:
1. Equipment-Event: EquipmentDamageReported{ProjectID, EquipmentID, DefectReport}
   → Project-Service empfängt

2. Project-Service:
   a) Markiere ProjectEquipment.ConditionOnReturn = "DEFECT"
   b) DefectReport speichern (Foto, Beschreibung, Kosten)
   c) Event: ProjectEquipmentDamaged

3. Invoice-Service empfängt Event:
   a) Ist Rechnung schon erstellt? Nein → warte bis Returned
      Ja → generiere Gutschrift (Credit Note)
   b) Betrag: Tagesmiete des Geräts × Tage(minus?)
   c) Notification an Thomas: "Gutschrift nötig für Projekt XY"

4. Marco oder System erstellt Gutschrift:
   a) CreateCreditNote(originalInvoiceID, -amount, "Defektes Equipment XY")
   b) Gutschrift-PDF
   c) Kunde-Notification (optional)
```

---

## 8. DATABASE SCHEMA (SQL-Snippets)

```sql
-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    number VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL,
    offer_id UUID,
    invoice_id UUID,
    planned_start_date TIMESTAMP NOT NULL,
    planned_end_date TIMESTAMP NOT NULL,
    actual_start_date TIMESTAMP,
    actual_end_date TIMESTAMP,
    offered_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Project Equipment
CREATE TABLE project_equipment (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id),
    equipment_id UUID NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'PLANNED',
    condition_on_dispatch VARCHAR(50),
    condition_on_return VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Invoices
CREATE TABLE invoices (
    id UUID PRIMARY KEY,
    number VARCHAR(20) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,  -- REGULAR, PARTIAL, ADVANCE, CREDIT_NOTE
    status VARCHAR(50) NOT NULL,
    issued_date DATE NOT NULL,
    performance_date_from DATE NOT NULL,
    performance_date_to DATE NOT NULL,
    due_date DATE NOT NULL,
    customer_id UUID NOT NULL,
    project_id UUID,
    subtotal_netto DECIMAL(10,2),
    tax_amount DECIMAL(10,2),
    tax_rate DECIMAL(5,4),  -- 0.19, 0.07, 0.0
    total_brutto DECIMAL(10,2),
    paid_amount DECIMAL(10,2) DEFAULT 0,
    remaining_amount DECIMAL(10,2),
    dunning_level INT DEFAULT 0,
    dunning_paused_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    created_by UUID NOT NULL
);

-- Invoice Positions
CREATE TABLE invoice_positions (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    position_number INT,
    type VARCHAR(50),  -- EQUIPMENT, CREW, SERVICE, DISCOUNT
    description TEXT,
    quantity DECIMAL(10,2),
    unit VARCHAR(50),
    unit_price DECIMAL(10,2),
    amount DECIMAL(10,2),
    tax_rate DECIMAL(5,4),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Time Entries
CREATE TABLE time_entries (
    id UUID PRIMARY KEY,
    crew_member_id UUID NOT NULL,
    project_id UUID NOT NULL,
    assignment_id UUID,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    actual_work_hours DECIMAL(5,2),
    break_minutes INT DEFAULT 0,
    night_hours DECIMAL(5,2) DEFAULT 0,
    overtime_hours DECIMAL(5,2) DEFAULT 0,
    calculated_amount DECIMAL(10,2),
    approved_by UUID,
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Crew Assignments
CREATE TABLE crew_assignments (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    crew_member_id UUID NOT NULL,
    status VARCHAR(50),  -- REQUESTED, OFFERED, CONFIRMED, COMPLETED
    offered_at TIMESTAMP,
    confirmed_at TIMESTAMP,
    confirmed_by UUID,
    daily_rate DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Offers
CREATE TABLE offers (
    id UUID PRIMARY KEY,
    number VARCHAR(20) UNIQUE NOT NULL,
    project_id UUID,
    version INT DEFAULT 1,
    status VARCHAR(50),  -- DRAFT, SENT, ACCEPTED, REJECTED, EXPIRED
    customer_id UUID NOT NULL,
    customer_email VARCHAR(255),
    valid_until DATE,
    subtotal_netto DECIMAL(10,2),
    tax_amount DECIMAL(10,2),
    total_brutto DECIMAL(10,2),
    pdf_path VARCHAR(1024),
    sent_at TIMESTAMP,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    created_by UUID NOT NULL
);

-- Dunning Records
CREATE TABLE dunning_records (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    level INT NOT NULL,
    scheduled_date DATE NOT NULL,
    sent_date DATE,
    sent_to VARCHAR(255),
    status VARCHAR(50),  -- SCHEDULED, SENT, PAUSED
    paused_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Packlists
CREATE TABLE packlists (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    version INT DEFAULT 1,
    status VARCHAR(50),  -- DRAFT, FINALIZED, IN_PROGRESS, COMPLETED
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Packlist Items
CREATE TABLE packlist_items (
    id UUID PRIMARY KEY,
    packlist_id UUID NOT NULL REFERENCES packlists(id),
    equipment_id UUID NOT NULL,
    quantity INT NOT NULL,
    warehouse_location VARCHAR(100),
    status VARCHAR(50),  -- PLANNED, PACKED, LOADED, ON_SITE, RETURNED
    packed_by UUID,
    packed_at TIMESTAMP,
    returned_by UUID,
    returned_at TIMESTAMP,
    condition_on_dispatch VARCHAR(50),
    condition_on_return VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes für Performance
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_customer_id ON projects(customer_id);
CREATE INDEX idx_invoices_number ON invoices(number);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_time_entries_project_id ON time_entries(project_id);
CREATE INDEX idx_time_entries_crew_id ON time_entries(crew_member_id);
CREATE INDEX idx_packlist_items_equipment_id ON packlist_items(equipment_id);
```

---

**Implementierungsnoten:**
- Alle Datum/Zeit-Felder: UTC speichern, lokal darstellen (basierend auf Kundenzone)
- Beträge: DECIMAL, niemals FLOAT (für Geldrechnung!)
- GoBD-Compliance: CreatedBy, CreatedAt, UpdatedAt für alle Finanzdokumente immutable halten
- Event-Sourcing: Alle State Changes als Events persistent speichern
- Optimistic Locking: Version-Felder für concurrency bei Mehrbenutzern (Packlisten)
- WebSocket: für Packlisten-Echtzeit-Sync (separate Implementierung)

**Stand:** 2026-03-21 | **Review-Status:** Ready for Development