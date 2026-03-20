# MyRMS — Erweiterte Module Spezifikation

**Stand:** 20. März 2026
**Version:** 1.0

Dieses Dokument beschreibt die Detailspezifikation der erweiterten Module: **Crew**, **Transport**, **Maintenance**, **Insurance** und **Document**.

---

## Modul 4: Crew — Personalverwaltung und Einsatzplanung (Modul N1)

### Überblick

Das Crew-Modul verwaltet Festangestellte und Freelancer, plant deren Einsätze und erfasst Arbeitszeiten. Zielgruppe: Marco (Einsatzplanung), Kevin (eigene Jobs und Zeiterfassung).

### 4.1 Crew-Stammdaten

#### Personenprofil

| Feld | Beschreibung |
|------|-------------|
| Grunddaten | Name, E-Mail, Telefon, Mobil, Adresse |
| Anstellungsform | Festangestellt / Freelancer |
| Qualifikationen | Fachkraft VT, Rigger, Ton-Techniker, Licht-Techniker, SFX, Streaming, etc. |
| Zertifikate | Führerscheinklassen, IPAF, Rigging-Zertifikat — jeweils mit Ablaufdatum |
| Tages-/Stundensatz | Standardhonorare (überschreibbar pro Projekt) |
| Steuer-ID | Für Freelancer (für Steuermeldungen) |
| IBAN | Für Überweisungen |
| Notfallkontakt | Name + Telefon |
| CalDAV-URL | Für Kalender-Synchronisation (optional) |

#### Qualifikationen und Zertifikate

```go
type Qualification struct {
    ID          string    // z.B. "fachkraft_va", "rigging_cat_3", "ton_foh"
    Name        string    // "Fachkraft für Veranstaltungstechnik"
    Category    string    // "license", "qualification", "certification"
    ValidUntil  *time.Time // Ablaufdatum (nil = unbefristet)
    CertFile    string    // URL zum Zertifikat-Scan
}
```

**Vordefinierte Qualifikationen:**
- Fachkraft für Veranstaltungstechnik (§ 39 DVG V)
- Veranstaltungsmeister
- Rigging (CAT 1-3, IGVW SQ P2)
- IPAF (Hubarbeitsbühnen)
- Führerschein (B, BE, C, CE, Fahrerkarte)
- BGV C1 (Elektroprüfung, DGUV V3)
- Erste Hilfe (mit Datum)
- Strahlenschutz (für Pyrotechnik)
- Tauchschein (für Unterwasser-Events)

### 4.2 Verfügbarkeitsplanung

#### Verfügbarkeits-Matrix (Wochenansicht)

Marco sieht in einer Übersicht:
- Alle Crew-Mitglieder als Zeilen
- Tage als Spalten
- Farbkodierung: Grün = verfügbar, Rot = belegt, Grau = abwesend, Gestreift = tentativ

```
Woche 13/2026 — 25.-31.03.2026

              Mo  Di  Mi  Do  Fr  Sa  So
Kevin Radic   ██  ██  ██  ██  ██  ██  --   (Alle Tage verfügbar bis Sa)
Lisa Huber    ██  ██  ██  ██  ██  --  --   (Urlaub Sa-So)
Max Mueller   ██  FE  FE  FE  ██  ██  --   (Di-Do: Festival Stadtpark)
...

Legende: ██ verfügbar  FE belegt (Festival)  -- frei gewählt / Urlaub
```

#### Verfügbarkeitseingabe durch Crew-Mitglieder

Kevin kann in seinem (mobilen) Account seine Verfügbarkeit angeben:
- Einzelne Tage: Verfügbar / Nicht verfügbar / Auf Anfrage
- Zeiträume: "14.-21.04. Urlaub"
- CalDAV-Synchronisation: Wenn Kevin im iPhone-Kalender einen Termin einträgt und der CalDAV-Feed synchronisiert ist, erscheint er automatisch als "belegt"

#### Konflikterkennung bei Zuweisung

```go
func (s *CrewService) AssignToCrew(ctx context.Context, cmd AssignCrewCommand) error {
    // 1. Prüfe: Ist Kevin bereits für denselben Zeitraum zugewiesen?
    conflicts, err := s.assignmentRepo.FindConflicts(
        ctx, cmd.CrewMemberID, cmd.DateFrom, cmd.DateTo)

    if len(conflicts) > 0 {
        // Event: CrewConflictDetected
        return &ConflictError{
            Message: fmt.Sprintf("%s bereits belegt: %s",
                crew.Name, conflicts[0].ProjectName),
        }
    }

    // 2. Prüfe: Hat Kevin die nötigen Qualifikationen?
    if !s.hasRequiredQualifications(cmd.CrewMemberID, cmd.RequiredQualifications) {
        return &QualificationError{Missing: missingQuals}
    }

    // 3. Zuweisung erstellen
    // ...
}
```

### 4.3 Zeiterfassung

#### Check-in / Check-out (für Kevin)

Kevin nutzt die mobile App (ein Knopf-Prinzip):
- Ankunft auf der Baustelle: "Check-in" → Zeitstempel wird gespeichert, GPS optional
- Abreise: "Check-out" → Dauer wird automatisch berechnet

```typescript
// Frontend: Zeiterfassung (sehr simpel für Kevin)
function TimeTracker({ assignment }: { assignment: Assignment }) {
  const { activeEntry, checkIn, checkOut } = useTimeTracking();

  return (
    <div className={styles.tracker}>
      {activeEntry ? (
        <>
          <div className={styles.running}>
            <Clock size={24} />
            <span>Läuft seit {formatDuration(activeEntry.startTime)}</span>
          </div>
          <Button
            size="xl"           // Großer Button für Baustelle
            variant="danger"
            onClick={() => checkOut(activeEntry.id)}
          >
            Check-out
          </Button>
        </>
      ) : (
        <Button
          size="xl"
          variant="success"
          onClick={() => checkIn(assignment.id)}
        >
          Check-in — {assignment.projectName}
        </Button>
      )}
    </div>
  );
}
```

#### Zeiterfassungs-Export

Marco kann die Stunden für alle Crew-Mitglieder exportieren:
- CSV-Export: Datum, Crew-Mitglied, Projekt, Stunden, Stundensatz, Betrag
- Filterbar: Nach Zeitraum, nach Projekt, nach Person
- Basis für Rechnungen von Freelancern (Kevin kann seinen Report selbst exportieren)
- Integration: Zeiteinträge werden in Projektkosten eingerechnet

### 4.4 Freelancer-Einladungs-Workflow

```
1. Marco erstellt Einladungslink für Kevin
   → auth-service generiert Token, versendet E-Mail

2. Kevin klickt Link, setzt Passwort
   → Account mit eingeschränkter Rolle "Freelancer" erstellt

3. Kevin sieht im System:
   ✓ Eigene zugewiesene Jobs (Datum, Ort, Uhrzeit, Anweisungen)
   ✓ Packlisten für seine Jobs
   ✓ Equipment-Basisinfos
   ✗ Keine Preise
   ✗ Keine Kundendaten
   ✗ Keine Rechnungen
   ✗ Keine Daten anderer Personen
```

---

## Modul 5: Transport — Fahrzeuge, Touren, Routenplanung (Modul J3)

### Überblick

Das Transport-Modul verwaltet Fahrzeuge und plant Touren (Lieferung/Abholung von Equipment). KI-Integration für Routenoptimierung und Kostenprognose.

### 5.1 Fahrzeugverwaltung

#### Fahrzeugdaten

| Feld | Beschreibung |
|------|-------------|
| Kennzeichen | Eindeutige Identifikation |
| Typ | Sprinter, LKW 7,5t, Transporter, PKW |
| Nutzlast (kg) | Maximale Zuladung |
| Ladevolumen (m³) | Verfügbarer Laderaum |
| Kraftstoffverbrauch | L/100km (für Kostenkalkulation) |
| Kraftstofftyp | Diesel, Elektro, Hybrid |
| HU (Hauptuntersuchung) | Fälligkeitsdatum |
| Nächster Service | Datum oder km-Stand |
| Status | Verfügbar / Im Einsatz / Wartung |

#### Kapazitätsprüfung

```go
// Beim Planen einer Tour: Passt das Equipment ins Fahrzeug?
func (s *TransportService) CheckVehicleCapacity(tourID uuid.UUID) CapacityResult {
    tour := s.tourRepo.Get(tourID)
    vehicle := s.vehicleRepo.Get(tour.VehicleID)

    // Gesamtgewicht und Volumen aller Flightcases/Equipment berechnen
    totalWeight := 0.0
    totalVolume := 0.0
    for _, item := range tour.Items {
        equipment := s.inventoryClient.GetEquipmentInfo(item.EquipmentID)
        totalWeight += equipment.WeightKg * float64(item.Quantity)
        if equipment.Dimensions != nil {
            totalVolume += equipment.Dimensions.Volume() * float64(item.Quantity)
        }
    }

    return CapacityResult{
        Vehicle:        vehicle,
        TotalWeightKg:  totalWeight,
        TotalVolumeM3:  totalVolume,
        WeightOK:       totalWeight <= vehicle.PayloadKg,
        VolumeOK:       totalVolume <= vehicle.VolumeM3,
        WeightPercent:  totalWeight / vehicle.PayloadKg * 100,
        VolumePercent:  totalVolume / vehicle.VolumeM3 * 100,
    }
}
```

### 5.2 Tourenplanung

#### Touren-Typen

| Typ | Beschreibung |
|-----|-------------|
| `delivery` | Lieferung Equipment zur Location |
| `pickup` | Abholung Equipment von Location |
| `delivery_and_pickup` | Hin- und Rückfahrt in einer Tour |
| `sub_rental_delivery` | Equipment von Partner-Firma abholen |

#### Tourplan-Erstellung

Marco erstellt eine Tour manuell oder lässt das System vorschlagen:

```
Tour für Festival Stadtpark (Fr 22.03.)

Fahrzeug: Sprinter MB-VA 123 (3,5t, Nutzlast 1.100kg)
Fahrer:   Max Mueller
Abfahrt:  07:00 Uhr ab Lager

Route:
  1. [Lager] Ladezeit 30 min
  2. [Location: Stadtpark Hauptbühne] Anlieferung, Abladezeit 45 min
  3. [Rückfahrt Lager] optional wenn noch Aufträge

Kapazität: 820 kg / 1.100 kg (74%) | 12 m³ / 15 m³ (80%)

Geplante Fahrzeit: 45 min (laut Maps-API)
Transportkosten: ~25€ Kraftstoff + 15€ Fahrzeit = 40€
```

#### KI-Routenoptimierung (Modul I16 / ai-service)

Bei mehreren Stops in einer Tour:

```go
// Beispiel: An einem Tag 3 Jobs anliefern
stops := []Stop{
    {Location: "Hauptbahnhof Halle A", Items: items1},
    {Location: "Messe Nord Halle 7", Items: items2},
    {Location: "Stadtpark Freilichtbühne", Items: items3},
}

// KI optimiert Reihenfolge basierend auf:
// - Geografischer Nähe (Travelling-Salesman-Heuristik)
// - Zeitfenstern der Locations (wenn bekannt)
// - Fahrzeuggewicht nach jeder Entladung
// - Verkehrsprognose (via Maps-API)

optimizedRoute, err := aiService.OptimizeRoute(ctx, depot, stops, vehicle, departureTime)
// → optimizedRoute: [Hauptbahnhof → Stadtpark → Messe Nord] (6% kürzer als naiver Ansatz)
```

### 5.3 Transportkosten-Tracking

Transportkosten werden pro Tour erfasst und können auf Projekte umgelegt werden:

| Kostenart | Erfassung |
|----------|----------|
| Kraftstoff | Manuell oder Berechnung (km × Verbrauch × Kraftstoffpreis) |
| Fahrzeit | Automatisch aus Check-In/Check-Out des Fahrers |
| Mautgebühren | Manuell |
| Parkgebühren | Manuell |
| Fahrzeugmiete | Bei Fremdfahrzeugen |

**Integration in Rechnungsstellung:**

Wenn Transportkosten auf Projekte umgelegt werden:
- Automatisch als Zusatzposition in der Rechnung (z.B. "Transport und Logistik: 85,00€")
- Konfigurierbar: Pauschale oder tatsächliche Kosten
- Konfigurierbar: Welche Kostenpositionen werden dem Kunden berechnet

---

## Modul 6: Maintenance — Wartungsplanung und DGUV V3 (Module J2, N2)

### Überblick

Das Maintenance-Modul verwaltet Wartungsplanung, -aufgaben und gesetzlich vorgeschriebene Prüfungen (E-Check / DGUV Vorschrift 3 und DGUV Vorschrift 4). Primäre Zielgruppe: Lisa und Admin.

### 6.1 Wartungsplanung

#### Wartungsplan-Typen

| Typ | Trigger | Beispiel |
|-----|---------|---------|
| `scheduled` | Zeitbasiert (täglich/monatlich/jährlich) | Jährliche Durchsicht Moving Heads |
| `after_use` | Nach N Einsätzen | Nach 50 Einsätzen: Lampenwechsel |
| `condition_based` | Wenn bestimmter Zustand erreicht | Wenn "Betriebsstunden > 500" |
| `echeck` | Gesetzlich (DGUV V3) | Alle 6 Monate BGV A3 Prüfung |

#### Wartungsaufgabe — Lebenszyklus

```
open → in_progress → completed
  └── cancelled (wenn nicht mehr relevant)

Bei "completed":
  → nächste Wartungsaufgabe wird automatisch erstellt
  → equipment.last_maintenance_at aktualisiert
  → equipment.next_maintenance_at gesetzt
  → Event: MaintenanceCompleted → KurrentDB
```

#### Wartungs-Dashboard

```
WARTUNG ÜBERSICHT

ÜBERFÄLLIG (5):
  └── Yamaha CL5 #001        Durchsicht           fällig 28.02. (20 Tage)  🔴
  └── Claypaky B-Eye #003    Lampenwechsel        fällig 01.03. (19 Tage)  🔴
  └── ...

DIESE WOCHE (3):
  └── Crown XTi #002         Lüfter reinigen      fällig 23.03.            🟡
  └── ...

NÄCHSTE 30 TAGE (12):
  └── ...

  [Neue Aufgabe]  [Wartungsplan erstellen]  [Export als PDF]
```

#### Wartungs-Checklisten

Für jeden Wartungsplan kann eine Checkliste definiert werden:

```json
{
  "equipment_type": "moving_head",
  "plan_name": "Jährliche Durchsicht Moving Head",
  "checklist": [
    { "id": "1", "task": "Lampe prüfen/tauschen", "required": true },
    { "id": "2", "task": "Optik reinigen (Linsen, Reflektoren)", "required": true },
    { "id": "3", "task": "Lüfter reinigen und auf Funktion prüfen", "required": true },
    { "id": "4", "task": "Kabelverbindungen prüfen (auf Beschädigungen)", "required": true },
    { "id": "5", "task": "Pan/Tilt-Kalibrierung prüfen", "required": true },
    { "id": "6", "task": "Firmware-Update prüfen", "required": false },
    { "id": "7", "task": "Außengehäuse auf Beschädigungen prüfen", "required": true },
    { "id": "8", "task": "Probelauf: alle Funktionen testen", "required": true }
  ]
}
```

### 6.2 E-Check / DGUV Vorschrift 3 (Elektrische Prüfungen)

#### Rechtlicher Hintergrund

Die DGUV Vorschrift 3 ("Elektrische Anlagen und Betriebsmittel") schreibt vor:
- Regelmäßige Prüfung elektrischer Betriebsmittel durch Elektrofachkräfte
- Prüfintervalle: 6 Monate (Baustellen, ortsveränderlich) bis 4 Jahre (ortsfest, Büros)
- Für Veranstaltungstechnik: typisch 6-12 Monate für ortsveränderliches Equipment
- Prüfprotokoll muss aufbewahrt werden (GoBD-relevant)

#### E-Check-Workflow

```
1. Fälligkeits-Erkennung
   → System markiert Equipment als "E-Check fällig"
   → Warnung: "E-Check fällig in 30 Tagen"
   → Sperre: "E-Check seit 14 Tagen überfällig — Ausgabe gesperrt!"

2. Prüfung durchführen
   → Elektriker prüft mit Prüfgerät (z.B. Metrawatt Secutest)
   → Messwerte werden aufgezeichnet:
      - Schutzleiterwiderstand (< 0,3 Ω)
      - Isolationswiderstand (> 1 MΩ)
      - Ableitstrom (< 3,5 mA)
      - Differenzstrom
   → Ergebnis: Bestanden / Bedingt bestanden / Nicht bestanden

3. Dokumentation im System
   → Prüfdatum, Prüfer, Messgerät, Messwerte, Ergebnis
   → Nächstes Prüfdatum wird berechnet
   → PDF-Prüfprotokoll generiert (für Aufbewahrungspflicht)
   → QR-Label mit neuem Prüfdatum kann gedruckt werden

4. Bei Nicht-Bestanden
   → Equipment wird automatisch gesperrt (Status: "defect")
   → Wartungsaufgabe "Reparatur elektrisch" erstellt
   → Ausgabe über Scanner blockiert mit Fehlermeldung
```

#### IZYTRON.IQ Import

IZYTRON.IQ ist eine verbreitete Software für elektrische Prüfungen in der VT-Branche:

```go
// XML-Import von IZYTRON.IQ Prüfergebnissen
func (s *ECheckService) ImportFromIZYTRON(ctx context.Context, xmlData []byte) (*ImportResult, error) {
    var report IzytronReport
    xml.Unmarshal(xmlData, &report)

    result := &ImportResult{}

    for _, item := range report.Items {
        // Gerät via Seriennummer oder interner Nummer zuordnen
        equipment, err := s.inventoryClient.FindBySerial(ctx, item.SerialNumber)
        if err != nil {
            result.NotFound = append(result.NotFound, item.SerialNumber)
            continue
        }

        // E-Check-Protokoll anlegen
        echeck := &ECheckRecord{
            EquipmentID:           equipment.ID,
            CheckDate:             item.TestDate,
            NextCheckDate:         item.NextTestDate,
            Result:                mapIzytronResult(item.Result),
            PerformedBy:           item.TesterName,
            MeasuringDevice:       item.TestDeviceName,
            MeasuringDeviceSerial: item.TestDeviceSerial,
            Measurements: map[string]float64{
                "protection_conductor_resistance": item.Rpe,
                "insulation_resistance":           item.Riso,
                "leakage_current":                 item.Ileak,
            },
            IzytronImportID: item.ID,
        }

        s.echeckRepo.Create(ctx, echeck)

        // Equipment-Status aktualisieren
        if echeck.Result == "failed" {
            s.inventoryClient.SetStatus(ctx, equipment.ID, "defect")
        } else {
            s.inventoryClient.UpdateECheckDates(ctx, equipment.ID, echeck.CheckDate, echeck.NextCheckDate)
        }

        result.Imported++
    }

    return result, nil
}
```

#### E-Check-Fristenübersicht

```
E-CHECK ÜBERSICHT — Alle 142 prüfpflichtigen Geräte

STATUS:
  ✓ Geprüft, gültig:          115 Geräte
  ⚠ Fällig in < 30 Tagen:     12 Geräte
  🔴 Überfällig:               11 Geräte
  🚫 Gesperrt (nicht bestanden): 4 Geräte

NÄCHSTE FÄLLIGKEITEN:
  Secutest 701 #004       fällig 25.03.2026   (5 Tage)
  Multicore 32/4 #002     fällig 28.03.2026   (8 Tage)
  ...

  [Export für Prüfer]  [IZYTRON Import]  [Alle fälligen anzeigen]
```

---

## Modul 7: Insurance — Versicherungen und Schadensfälle (Module K2, K3)

### Überblick

Das Insurance-Modul verwaltet Versicherungspolizzen, prüft Deckung bei Anfragen und begleitet den Schadensfall-Prozess.

### 7.1 Versicherungsverwaltung

#### Polizzen-Typen in der VT-Branche

| Typ | Beschreibung |
|-----|-------------|
| Allgemeine Betriebshaftpflicht | Schäden an Dritten |
| Equipment-Versicherung (All Risk) | Schäden am eigenen Equipment |
| Transportversicherung | Schäden beim Transport |
| Veranstalterhaftpflicht | Für spezifische Events |
| Elektronikversicherung | Elektronische Geräte (Mischpulte, etc.) |

#### Deckungsprüfung

```go
// Wird aufgerufen wenn ein Schaden gemeldet wird
func (s *InsuranceService) CheckCoverage(ctx context.Context, equipmentID uuid.UUID, incidentDate time.Time) CoverageResult {
    // Alle aktiven Polizzen für den Mandanten
    policies, _ := s.policyRepo.GetActivePolicies(ctx, s.tenantID)

    covered := false
    var coveringPolicy *InsurancePolicy
    details := []string{}

    for _, policy := range policies {
        // Prüfe ob Polizze zu diesem Zeitpunkt aktiv war
        if !policy.IsActiveAt(incidentDate) { continue }

        // Prüfe ob Equipment durch diese Polizze gedeckt ist
        // (Je nach Polizzen-Typ: nach Wert, nach Kategorie, All-Risk)
        if policy.CoversEquipment(equipmentID) {
            covered = true
            coveringPolicy = &policy
            details = append(details, fmt.Sprintf("Gedeckt durch: %s (Polizze %s)", policy.InsurerName, policy.PolicyNumber))
            break
        }
    }

    return CoverageResult{
        IsCovered:      covered,
        Policy:         coveringPolicy,
        Details:        details,
        Recommendation: s.getRecommendation(covered, equipmentID),
    }
}
```

#### Ablauf-Warnungen

```
VERSICHERUNGS-ÜBERSICHT

LÄUFT AB IN < 60 TAGEN:
  ⚠ All-Risk-Equipment-Police    VHV Versicherungen   fällig 15.04.2026   (26 Tage)
  ⚠ Betriebshaftpflicht          Allianz             fällig 01.05.2026   (42 Tage)

AKTIVE POLICEN:
  ✓ All-Risk-Equipment-Police     Deckungssumme: 2.500.000€
  ✓ Betriebshaftpflicht           Deckungssumme: 5.000.000€
  ✓ Transportversicherung         Deckungssumme: 500.000€

  [Neue Police anlegen]  [Ablauf-Erinnerungen konfigurieren]
```

### 7.2 Schadensfallmanagement (8-Schritte-Prozess)

```
Schritt 1: MELDUNG
  Lisa meldet via Scanner-App: "Claypaky Sharpy defekt — Gehäuse gebrochen"
  → Fotos werden direkt hochgeladen
  → Schadensfall wird mit Datum, Equipment, Projekt verknüpft

Schritt 2: ERSTE EINSCHÄTZUNG
  Marco schaut den Schadensfall an:
  → Kostenschätzung: Reparatur ~500€ oder Totalschaden?
  → Versicherungsdeckung prüfen (automatisch via insurance-service)
  → Entscheidung: Versicherung informieren? Selbst zahlen?

Schritt 3: FOTODOKUMENTATION
  → Minimum 3 Fotos aus verschiedenen Winkeln
  → Fotos werden unveränderlich gespeichert (Zeitstempel, Checksum)
  → "Vor dem Schaden" Fotos aus historischen Scans (wenn vorhanden)

Schritt 4: REPARATURANGEBOT (KVA)
  → Lieferantenauftrag: Equipment zur Reparatur einschicken
  → Kostenvoranschlag (KVA) hochladen
  → Entscheidung: Reparatur / Totalabschreibung

Schritt 5: VERSICHERUNGSMELDUNG (wenn über Selbstbehalt)
  → Schadensformular der Versicherung ausfüllen (System generiert Vorlage)
  → Alle Dokumente als ZIP-Paket exportieren (Fotos, KVA, Prüfprotokolle)
  → Fallnummer der Versicherung eintragen

Schritt 6: REPARATUR / ERSATZ
  → Reparatur: Equipment bleibt auf Status "maintenance"
  → Totalschaden: Equipment wird als "retired" markiert, Ersatzbeschaffung geplant

Schritt 7: ABSCHLUSS
  → Versicherungsleistung eingegangen (Betrag eintragen)
  → Eigenanteil / Selbstbehalt verbucht
  → Schadensfall geschlossen

Schritt 8: ANALYSE (optional, mit KI)
  → KI analysiert: Häufen sich Schäden an bestimmtem Equipment-Typ?
  → Empfehlung: Häufigere Wartung / bessere Flightcases / anderes Handling
```

---

## Modul 8: Document — Lieferscheine, Verträge, Templates (Module DOC, J1)

### Überblick

Das Document-Modul generiert alle PDF-Dokumente in MyRMS: Angebote, Rechnungen, Lieferscheine, Packlisten, Verträge und QR-Labels. Es bietet einen visuellen Template-Editor und archiviert alle generierten Dokumente GoBD-konform.

### 8.1 Dokument-Typen

| Dokument-Typ | Auslöser | Empfänger |
|-------------|---------|----------|
| Angebot (PDF) | Manuell / Projekt | Kunde |
| Rechnung (PDF) | Manuell / Projekt abgeschlossen | Kunde |
| Gutschrift (PDF) | Manuell | Kunde |
| Mahnung (PDF) | Automatisch / Manuell | Kunde |
| Lieferschein (PDF) | Bei Check-Out | Fahrer / Kunde |
| Rückgabeprotokoll (PDF) | Bei Check-In | Archiv / Kunde |
| Packliste (PDF) | Manuell / Lisa | Lager / Fahrer |
| Übergabeprotokoll (PDF) | Sub-Rental Übergabe | Partner-Firma |
| Vertrag (PDF) | Manuell | Kunde |
| E-Check-Protokoll (PDF) | Nach E-Check | Archiv / Versicherung |
| Equipment-Label (ZPL/PDF) | Manuell / Bulk | Drucker |

### 8.2 Template-Engine

Templates werden in HTML mit Go-Template-Syntax erstellt. Im Frontend gibt es einen WYSIWYG-ähnlichen Editor mit Vorschau.

#### Template-Variablen (Platzhalter)

```handlebars
{{/* Allgemeine Variablen */}}
{{ .Tenant.Name }}              → MB Veranstaltungstechnik GmbH
{{ .Tenant.Address }}           → Musterstraße 42, 12345 Musterstadt
{{ .Tenant.VAT_ID }}            → DE123456789
{{ .Tenant.IBAN }}              → DE89 3704 0044 0532 0130 00
{{ .Tenant.Logo }}              → URL zum Logo (als Base64 eingebettet)
{{ .CurrentDate }}              → 20. März 2026
{{ .CurrentDateISO }}           → 2026-03-20

{{/* Angebots-/Rechnungsdaten */}}
{{ .Invoice.Number }}           → RE-2026-042
{{ .Invoice.Date }}             → 20.03.2026
{{ .Invoice.DueDate }}          → 03.04.2026
{{ .Invoice.Subject }}          → Festival Stadtpark 2026 — Ton und Licht
{{ .Invoice.Subtotal | currency }} → 12.500,00 €
{{ .Invoice.TaxAmount | currency }} → 2.375,00 €
{{ .Invoice.Total | currency }} → 14.875,00 €

{{/* Kundendaten */}}
{{ .Customer.Name }}            → SoundEvents GmbH
{{ .Customer.Address }}         → ...

{{/* Positionen (Schleife) */}}
{{ range .Items }}
  {{ .Position }}  {{ .Description }}  {{ .Quantity }} × {{ .UnitPrice | currency }} = {{ .Total | currency }}
{{ end }}
```

#### Standard-Templates

Das System wird mit folgenden Standard-Templates ausgeliefert:
- `standard_quote.html` — Professionelles Angebots-Layout
- `standard_invoice.html` — GoBD-konformes Rechnungslayout
- `credit_note.html` — Gutschrift
- `dunning_1.html` bis `dunning_3.html` — Mahnungen (steigender Ton)
- `delivery_note.html` — Lieferschein mit Unterschriftsfeld
- `return_protocol.html` — Rückgabeprotokoll mit Zustandsbewertung
- `packing_list.html` — Packliste sortiert nach Lagerplatz
- `contract_rental.html` — Mietvertrag mit AGB
- `echeck_protocol.html` — E-Check/DGUV V3 Prüfprotokoll

### 8.3 Lieferschein-Workflow

#### Lieferschein-Erstellung

```go
// Lieferschein wird automatisch erzeugt wenn:
// 1. Equipment für Projekt-Transport-Tour zusammengestellt wird, ODER
// 2. Lisa bestätigt dass Kommissionierung abgeschlossen ist

func (s *DocumentService) GenerateDeliveryNote(ctx context.Context, projectID uuid.UUID, tourID *uuid.UUID) (*Document, error) {
    project := s.projectClient.GetProject(ctx, projectID)
    items := s.packingListClient.GetPackedItems(ctx, projectID)  // Nur tatsächlich gescannte Items

    data := DeliveryNoteData{
        Tenant:         s.tenantClient.GetTenant(ctx, project.TenantID),
        Customer:       project.Customer,
        Project:        project,
        Items:          items,
        DeliveryDate:   time.Now(),
        DeliveryNumber: s.generateDeliveryNumber(ctx, project.TenantID),
        SignatureField:  true,   // Unterschriftsfeld für Empfänger
    }

    // HTML rendern
    html, _ := s.renderTemplate(ctx, "delivery_note", data)

    // PDF generieren via chromedp
    pdf, _ := s.pdfGenerator.Generate(ctx, html)

    // Speichern + Checksum für GoBD
    doc := s.archiveDocument(ctx, pdf, "delivery_note", projectID)

    return doc, nil
}
```

#### Digitale Unterschrift auf dem Tablet

Bei Übergabe kann der Empfänger direkt auf dem Tablet unterschreiben:

```typescript
// Signature-Pad Komponente (für Lieferschein-Unterzeichnung)
function SignaturePad({ onSigned }: { onSigned: (signature: Blob) => void }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const signaturePad = useSignaturePad(canvasRef);

  return (
    <div className={styles.signaturePad}>
      <p>Bitte hier unterschreiben:</p>
      <canvas
        ref={canvasRef}
        className={styles.canvas}
        width={600}
        height={200}
      />
      <div className={styles.actions}>
        <Button variant="ghost" onClick={() => signaturePad.clear()}>
          Löschen
        </Button>
        <Button
          onClick={() => signaturePad.toBlob(onSigned)}
          disabled={signaturePad.isEmpty()}
        >
          Bestätigen
        </Button>
      </div>
    </div>
  );
}
```

Die Unterschrift wird als PNG in das PDF eingebettet und das fertige PDF wird im Archiv gespeichert.

### 8.4 Vertragsmanagement (Modul J1)

#### Vertragsvorlagen

Vertragsvorlagen enthalten Platzhalter für:
- Projektdaten (Ort, Datum, Beschreibung)
- Mietgegenstände (aus Packliste)
- Preise und Zahlungskonditionen
- AGB (Mandanten-spezifisch)
- Haftungsklauseln
- Schadensselbstbehalt des Kunden
- Unterschriftsfelder

#### Versionierung

Verträge werden versioniert gespeichert. Wenn ein Vertrag nach Unterzeichnung geändert wird:
- Alte Version bleibt im Archiv (Unveränderlichkeit)
- Neue Version wird als "Version 2" gespeichert
- Beide Parteien müssen neue Version bestätigen

#### Digitale Signatur

Unterstützte Signatur-Methoden:
1. **Papier-Ausdruck**: PDF drucken, unterschreiben, einscannen
2. **Tablet-Unterschrift**: Direkt auf Touchscreen (wie oben beschrieben)
3. **E-Mail-Bestätigung**: Kunde bestätigt via E-Mail-Link (rechtlich eingeschränkt)
4. **Qualifizierte E-Signatur** (Zukunft): Integration mit DocuSign/Adobe Sign

### 8.5 Dokumenten-Archiv (GoBD-konform)

Alle generierten Dokumente werden mit folgenden Metadaten archiviert:

```go
type ArchivedDocument struct {
    ID            uuid.UUID
    TenantID      uuid.UUID
    ReferenceID   uuid.UUID       // Invoice-ID, Project-ID, etc.
    ReferenceType string          // "invoice", "project", "delivery_note"
    Filename      string          // RE-2026-042.pdf
    FilePath      string          // /data/documents/2026/03/RE-2026-042.pdf
    FileSizeBytes int
    MimeType      string          // application/pdf
    Checksum      string          // SHA-256 (für Unveränderlichkeitsprüfung)
    GeneratedAt   time.Time       // Unveränderlicher Zeitstempel
    GeneratedBy   uuid.UUID       // User-ID
}
```

**GoBD-Anforderungen:**
- Dokumente können nach Erstellung nicht mehr gelöscht oder geändert werden
- Checksumme wird beim Archivieren berechnet und kann jederzeit verifiziert werden
- Aufbewahrungsfrist: 10 Jahre (konfigurierbar nach nationalem Recht)
- Export: Alle Dokumente als ZIP für Betriebsprüfung

### 8.6 Scan-to-Document (OCR) — Modul I17

Eingehende Dokumente (Lieferantenrechnungen, Sub-Rental-Verträge) können eingescannt und automatisch verarbeitet werden:

```
Workflow:
1. PDF/Bild hochladen
2. OCR-Engine (Tesseract oder Cloud-API) extrahiert Text
3. KI (ai-service) interpretiert Felder:
   - Rechnungsnummer
   - Datum
   - Lieferantenname
   - Positionen und Beträge
4. Erkannte Daten werden als Eingangsrechnung vorgeschlagen
5. Thomas prüft und bestätigt (oder korrigiert) die erkannten Daten
6. Eingangsrechnung wird in invoice_schema gespeichert
```

Erkennungsrate: ~85-95% für gut lesbare Dokumente. Alle unbekannten Felder werden zur manuellen Überprüfung markiert.
