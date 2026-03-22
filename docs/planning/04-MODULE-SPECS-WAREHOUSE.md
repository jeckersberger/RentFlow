# MyRMS — Lager, Scanner, Federation und Audit-Module Spezifikation

**Stand:** 20. März 2026
**Version:** 1.0

Dieses Dokument beschreibt die Detailspezifikation der Module: **Warehouse** (Lager), **Scanner** (QR/Barcode/RFID), **Federation** (Mehrstandort) und **Audit** (GoBD-Compliance).

---

## Modul 13: Warehouse — Lager und Lagerverwaltung (Module L1, L3-L5)

### Überblick

Das Warehouse-Modul verwaltet die physische Struktur des Lagers: Lagerhallen, Zonen, Regale und einzelne Lagerplätze. Es trackt wo sich jedes Equipment-Stück befindet und ermöglicht eine digitale Inventur.

Zielgruppe: Lisa (tägliche Lagerarbeit), Marco (Lager-Überblick), Admin (Lagerstruktur einrichten).

### 13.1 Lagerstruktur (Hierarchisch)

```
Mandant: MB Veranstaltungstechnik
└── Lager: Hauptlager (Gewerbegebiet Musterstadt)
    ├── Zone: Halle A (Licht & Rigging)
    │   ├── Regal A1 (Moving Heads)
    │   │   ├── Fach A1-01 (Claypaky Sharpy × 4)
    │   │   ├── Fach A1-02 (Martin MAC Quantum × 6)
    │   │   └── Fach A1-03 (GLP Impression X4 × 8)
    │   ├── Regal A2 (Conventional + PAR)
    │   └── Regal A3 (Traversensystem)
    │
    ├── Zone: Halle B (Ton)
    │   ├── Regal B1 (PA-Module: L-Acoustics)
    │   ├── Regal B2 (PA-Module: d&b)
    │   ├── Regal B3 (Mischpulte)
    │   └── Boden B-Groß (Subwoofer, Fluggestelle)
    │
    ├── Zone: Halle C (Kabel, Adapter, Kleinteile)
    │   ├── Regal C1 (Audiokabel 10m)
    │   ├── Regal C2 (Audiokabel 15m+)
    │   ├── Regal C3 (DMX-Kabel)
    │   └── Regal C4 (Adapter, DI-Boxen)
    │
    └── Zone: Büro
        └── Büroregal (Laptops, Tablets, Zubehör)

Optional: Außenlager, Fahrzeuge als Lagerorte
```

#### Lagerort-Einrichtung (Admin)

```typescript
// Lagerorte einrichten: Drag & Drop Tree-Editor
function WarehouseStructureEditor() {
  const { data: warehouses } = useWarehouses();

  return (
    <div className={styles.editor}>
      <h2>Lagerstruktur</h2>
      <p className={styles.hint}>
        Erstellen Sie Ihre Lagerstruktur. Jeder Lagerplatz bekommt einen QR-Code
        den Sie ausdrucken und ans Regal kleben können.
      </p>

      {warehouses?.map(warehouse => (
        <WarehouseTreeNode
          key={warehouse.id}
          warehouse={warehouse}
          onAddZone={(id) => addZone(id)}
          onAddLocation={(parentId) => addLocation(parentId)}
        />
      ))}

      <Button onClick={() => addWarehouse()}>
        <Plus size={16} /> Neues Lager hinzufügen
      </Button>
    </div>
  );
}
```

#### Lagerplatz-QR-Codes

Jeder Lagerort hat einen eindeutigen QR-Code. Lisa kann diesen scannen um:
- Zu sehen was an diesem Lagerort sein sollte
- Equipment diesem Lagerort zuzuweisen ("Einlagern hier")
- Die aktuelle Belegung zu prüfen

```
QR-Code-Inhalt: loc:{location-uuid}
Beispiel: loc:a7f3c891-4d2e-...

Label für Regal A1-01:
┌─────────────────────┐
│ █████████ █ █ █████ │
│ █  QR  █ REGAL A1  │
│ █ CODE █ Fach 01   │
│ █████████          │
│ Moving Heads Zone  │
│ a7f3c891           │
└─────────────────────┘
```

### 13.2 Equipment-Lagerplatz-Tracking

#### Einlagern / Umlagern

```go
// Equipment einem Lagerplatz zuweisen (nach Rückgabe oder Umlagerung)
func (s *WarehouseService) AssignLocation(ctx context.Context, cmd AssignLocationCommand) error {
    // Validierung: Lagerort existiert und ist aktiv
    location, err := s.locationRepo.GetByID(ctx, cmd.LocationID)
    if err != nil { return err }
    if !location.IsActive { return ErrLocationInactive }

    // Alten Lagerort entfernen
    s.locationRepo.RemoveEquipment(ctx, cmd.EquipmentID, cmd.TenantID)

    // Neuen Lagerort setzen
    s.locationRepo.AssignEquipment(ctx, cmd.EquipmentID, cmd.LocationID, cmd.Quantity)

    // Lagerbewegung protokollieren
    s.movementRepo.Create(ctx, StockMovement{
        EquipmentID:    cmd.EquipmentID,
        MovementType:   "transfer",
        Quantity:       cmd.Quantity,
        ToLocationID:   &cmd.LocationID,
        UserID:         cmd.UserID,
    })

    // Event: EquipmentLocated → KurrentDB
    s.eventStore.Append(ctx, "warehouse-events", events.EquipmentLocated{
        EquipmentID: cmd.EquipmentID,
        LocationID:  cmd.LocationID,
        LocationName: location.FullPath(),  // "Halle B / Regal B1 / Fach 03"
    })

    return nil
}
```

#### Lagerplatz-Suche

```
Kevin fragt Lisa: "Wo sind die DMX-Splitter?"
Lisa öffnet App, tippt "DMX-Splitter":
→ App zeigt: "4× Elation 4Port DMX Splitter — Regal C3, Fach 02"
→ Karte des Lagers mit markiertem Lagerplatz (optional)
```

### 13.3 Warenbewegungen

Jede Bewegung von Equipment wird automatisch protokolliert:

| Bewegungstyp | Auslöser |
|-------------|---------|
| `inbound` | Neue Einlagerung (Neukauf, Rückgabe) |
| `outbound` | Ausgabe für Projekt |
| `transfer` | Umlagern innerhalb des Lagers |
| `adjustment` | Korrektur bei Inventur |
| `sub_rental_in` | Eingang von Partner-Firma |
| `sub_rental_out` | Ausgang an Partner-Firma |
| `maintenance_in` | Ins Lager nach Wartung |
| `maintenance_out` | Aus dem Lager zur Wartung |

### 13.4 Inventur-System

#### Inventur-Typen

| Typ | Beschreibung | Wann |
|-----|-------------|------|
| **Vollständige Inventur** | Alle Artikel im Lager | Jährlich (Jahresabschluss) |
| **Partielle Inventur** | Bestimmte Zone oder Kategorie | Wöchentlich/Monatlich |
| **Stichproben-Inventur** | Zufällige Auswahl | Laufend |
| **Projektbezogene Inventur** | Nach großem Projekt | Nach Festival-Saison |

#### Inventur-Workflow

```
Schritt 1: Inventur starten
  → Marco/Admin erstellt neue Inventur (Typ, Scope, Datum)
  → System generiert Inventurliste mit Soll-Beständen

Schritt 2: Zählen (Lisa mit Zebra-Handheld)
  → Lisa geht Zone für Zone durch
  → Für jeden Lagerplatz: Artikel scannen → Anzahl eingeben
  → Offline-fähig: Scans werden lokal gespeichert, bei Sync übertragen

Schritt 3: Differenzen anzeigen
  → System zeigt: Was stimmt? Was fehlt? Was ist mehr da?
  ┌────────────────────────────────────────────────┐
  │ INVENTUR ERGEBNISSE — Regal A1                 │
  │                                                │
  │ Claypaky Sharpy        Soll: 8  Ist: 7  -1 🔴 │
  │ Martin MAC Quantum     Soll: 6  Ist: 6   0 ✓  │
  │ GLP Impression X4      Soll: 8  Ist: 9  +1 🟡 │
  └────────────────────────────────────────────────┘

Schritt 4: Differenzen klären
  → Für jede Differenz: Ursache recherchieren
  → Mögliche Ursachen: auf Baustelle vergessen, Diebstahl, Fehler beim Zählen
  → Erklärung dokumentieren

Schritt 5: Inventur abschließen
  → Bestand wird angepasst (Inventurdifferenz gebucht)
  → Inventur-Protokoll als PDF exportieren (für Steuerberater)
  → GoBD-konform: Differenzen werden als Buchungsbeleg archiviert
```

#### Inventur-Mobile-UI

```typescript
// Mobil-optimierte Inventur-Zählung (Zebra TC21)
function InventoryCountingScreen({ countId }: { countId: string }) {
  const { data: count } = useInventoryCount(countId);
  const updateItem = useUpdateInventoryCountItem();
  const [activeLocation, setActiveLocation] = useState<string | null>(null);

  const handleScan = async (barcode: string) => {
    // Barcode → Location oder Equipment identifizieren
    const identified = await identifyBarcode(barcode);

    if (identified.type === 'location') {
      setActiveLocation(identified.id);
      // Zeige alle Items die an diesem Lagerplatz sein sollten
    } else if (identified.type === 'equipment') {
      // Item zählen: Feld "counted_qty" +1 oder Modal für Mengen-Eingabe
      const item = count.items.find(i => i.equipmentId === identified.id);
      updateItem.mutate({ itemId: item.id, countedQty: (item.countedQty ?? 0) + 1 });
    }
  };

  return (
    <div className={styles.counting}>
      {/* Aktueller Lagerplatz */}
      <div className={styles.currentLocation}>
        <MapPin size={20} />
        {activeLocation ? (
          <span>{getLocationName(activeLocation)}</span>
        ) : (
          <span className={styles.hint}>Lagerort-QR scannen zum Starten</span>
        )}
      </div>

      {/* Fortschritt */}
      <ProgressBar
        value={count.items.filter(i => i.countedQty !== null).length}
        max={count.items.length}
        label="Positionen gezählt"
      />

      {/* Items an aktivem Lagerplatz */}
      {activeLocation && (
        <ItemsList
          items={count.items.filter(i => i.locationId === activeLocation)}
          onQuantityChange={(itemId, qty) =>
            updateItem.mutate({ itemId, countedQty: qty })
          }
        />
      )}

      <BarcodeInputField onScan={handleScan} autoFocus />
    </div>
  );
}
```

### 13.5 Lager-Optimierung (KI-gestützt)

```go
// KI-Vorschläge für optimale Einlagerung
func (s *WarehouseOptimizationService) GetStorageSuggestions(ctx context.Context, equipmentID uuid.UUID) []StorageSuggestion {
    equipment, _ := s.inventoryClient.GetEquipmentInfo(ctx, equipmentID)
    usageHistory, _ := s.movementRepo.GetRecentUsage(ctx, equipmentID, 90)

    // Analyse: Mit welchem anderen Equipment wird dieses oft zusammen verwendet?
    frequentCombinations := s.analyzeFrequentCombinations(usageHistory)

    suggestions := []StorageSuggestion{}

    // Vorschlag 1: Nah bei häufig kombinierten Artikeln
    for _, combo := range frequentCombinations {
        comboLocation, _ := s.locationRepo.GetEquipmentLocation(ctx, combo.EquipmentID)
        suggestions = append(suggestions, StorageSuggestion{
            LocationID: comboLocation.NearbyFreeLocation(),
            Reason: fmt.Sprintf("Oft zusammen verwendet mit %s (in %s)", combo.Name, comboLocation.Name),
            Score: combo.CoUsageCount,
        })
    }

    // Vorschlag 2: Häufig verwendetes Equipment nahe Ausgang
    if len(usageHistory) > 10 {
        suggestions = append(suggestions, StorageSuggestion{
            LocationID: s.getLocationNearExit(ctx),
            Reason: "Häufig verwendetes Equipment — nahe am Ausgang einlagern",
            Score: len(usageHistory),
        })
    }

    sort.Slice(suggestions, func(i, j int) bool { return suggestions[i].Score > suggestions[j].Score })
    return suggestions
}
```

---

## Modul 14: Scanner — QR/Barcode/RFID-Scanning

### Überblick

Das Scanner-Modul ist die physische Schnittstelle zwischen Lager und Software. Es unterstützt alle in der VT-Branche üblichen Scan-Methoden und ist speziell für den professionellen Einsatz mit Zebra-Handhelds optimiert.

### 14.1 Unterstützte Hardware

| Hardware | Verwendung | Integration |
|---------|-----------|-------------|
| Zebra TC21/TC26 Android | Lisa (Lager, Tagesgeschäft) | DataWedge → Android-App |
| USB-Barcode-Scanner | Am PC / Büro | USB-HID → Web-Browser direkt |
| iPhone/Android-Kamera | Kevin (Baustelle), Freelancer | @zxing/browser Kamera-API |
| RFID-Reader (Zebra) | Große Lager, Bulk-Scanning | Zebra RFID SDK |
| Webcam Desktop | Lisa am PC | @zxing/browser |

### 14.2 Zebra DataWedge Integration

Zebra-Geräte verwenden DataWedge als Middleware zwischen Scanner-Hardware und Apps:

```javascript
// Android: DataWedge sendet Scans als Broadcast Intent
// MyRMS App (Android-WebView) lauscht auf diesen Intent

// In der PWA (Android Wrapper oder WebView):
document.addEventListener('DWBarcodeScan', (event) => {
    const barcode = event.detail.barcode;
    const format = event.detail.format; // CODE_128, QR_CODE, etc.

    // Verarbeite den Scan
    processScan(barcode, format);
});

// Zebra DataWedge-Profil konfigurieren (einmalig):
{
  "Profile Name": "MyRMS Scanner",
  "Enabled": true,
  "Data Capture Plus": {
    "Barcode Input": { "Enabled": true },
    "Intent Output": {
      "Enabled": true,
      "Intent Action": "com.myrms.BARCODE_SCAN",
      "Intent Category": "android.intent.category.DEFAULT"
    }
  }
}
```

### 14.3 Scan-Kontexte

Der Scanner weiß immer in welchem Kontext er operiert:

```typescript
// Scanner hat immer einen aktiven Kontext
type ScanContext = {
  mode: 'checkout' | 'checkin' | 'inventory' | 'locate' | 'damage' | 'restock';
  projectId?: string;    // Für welches Projekt wird ausgegeben/eingenommen?
  locationId?: string;   // In welchen Lagerort wird eingelagert?
  userId: string;
};
```

#### Check-Out-Kontext

```
Kontext: Lisa checkt Equipment für Festival Stadtpark aus

1. Lisa öffnet Packliste für Festival Stadtpark
2. Klickt: "Ausgabe starten"
3. Scanner-Modus: "checkout" mit projectId = "festival-stadtpark-id"

Für jeden Scan:
  → Equipment-Info laden
  → Prüfen: Gehört dieses Equipment zur Packliste des Projekts?
  → Prüfen: Equipment in gutem Zustand?
  → Lagerplatz löschen (Equipment ist jetzt "ausgegeben")
  → Packlisten-Position abhaken

Feedback:
  GRÜN:  "✓ Ausgegeben: Claypaky Sharpy #003 — für Festival Stadtpark"
  GELB:  "⚠ Nicht in Packliste — trotzdem ausgeben?" [Ja] [Nein]
  ROT:   "✗ Equipment defekt! Ausgabe nicht möglich."
  ROT:   "✗ Für ANDEREN Job! (Festival München)"
```

#### Check-In-Kontext

```
Kontext: Rückgabe nach Festival

1. Lisa scannt erstes Teil beim Entladen des LKW
2. System erkennt automatisch: "Kommt von Festival Stadtpark"
3. Scanner-Modus: "checkin" mit auto-erkanntem projectId

Für jeden Scan:
  → Prüfen: Kommt von welchem Projekt?
  → Zeigt: "Zustand bei Ausgabe: OK (vor 3 Tagen)"
  → Lisa bewertet Zustand: OK / Leicht beschädigt / Defekt
  → Bei Defekt: Foto + Notiz in 3 Taps
  → Lagerplatz vorschlagen: "Einlagern in A1-01?" [Ja] [Anderer Ort]
```

### 14.4 RFID-Unterstützung

```go
// RFID-Leser können mehrere Tags gleichzeitig lesen (Bulk-Scan)
// Besonders sinnvoll für: Vollständigkeitsprüfung von Flightcases

type RFIDScanResult struct {
    Tags      []RFIDTag
    Timestamp time.Time
    ReaderID  string
}

// Use Case: Flightcase-Vollständigkeitsprüfung
func (s *ScannerService) CheckFlightcaseCompleteness(ctx context.Context, caseID uuid.UUID) CompletenessResult {
    // Alle RFID-Tags im Flightcase lesen (Bulk-Scan in < 1 Sekunde)
    tags, _ := s.rfidReader.ReadAllInRange(ctx)

    // Equipment-IDs aus Tags auflösen
    equipment := s.resolveRFIDTags(ctx, tags)

    // Mit Soll-Inhalt des Flightcases vergleichen
    expectedContents, _ := s.inventoryClient.GetFlightcaseContents(ctx, caseID)

    missing := s.findMissing(expectedContents, equipment)
    extra := s.findExtra(expectedContents, equipment)

    return CompletenessResult{
        TotalExpected: len(expectedContents),
        Found:         len(equipment),
        Missing:       missing,
        Extra:         extra,
        IsComplete:    len(missing) == 0,
    }
}
```

### 14.5 Offline-Modus

Der Offline-Modus ist kritisch für Lisa im Lager und Kevin auf der Baustelle:

#### Offline-Datenstrategie

```typescript
// Was wird lokal im Browser/App gecacht für Offline-Zugriff?
const offlineCachedData = {
  // Pflicht (immer gecacht):
  equipment_basics: {
    ttl: '24h',
    items: 'alle Equipment-Basisinfos (barcode → id, name, status, location)',
    size: '~2MB für 1000 Artikel',
  },
  packing_lists_active: {
    ttl: '4h',
    items: 'alle Packlisten aktiver Projekte (aktuell + nächste 7 Tage)',
    size: '~500KB',
  },
  locations_tree: {
    ttl: '24h',
    items: 'Lagerorte-Struktur für Einlagerungsvorschläge',
    size: '~100KB',
  },

  // Optional (wenn Platz):
  equipment_photos: {
    ttl: '7 Tage',
    items: 'Thumbnails aller Artikel (für Wiedererkennung)',
    size: '~20MB für 1000 Artikel',
  },
};
```

#### Offline-Scan-Verarbeitung

```typescript
// Service Worker: Scans während Offline queuen
class OfflineScanQueue {
  private db: IDBDatabase;

  async addScan(scan: ScanData): Promise<void> {
    const entry: OfflineEntry = {
      ...scan,
      localId: crypto.randomUUID(),
      localTimestamp: new Date().toISOString(),
      synced: false,
    };

    await this.db.add('offline_scans', entry);

    // Versuche sofort zu synchronisieren
    if (navigator.onLine) {
      await this.syncPending();
    }
  }

  async syncPending(): Promise<SyncResult> {
    const pending = await this.db.getAll('offline_scans', IDBKeyRange.only(false));

    const result: SyncResult = { synced: 0, failed: 0, conflicts: [] };

    for (const entry of pending) {
      try {
        const response = await fetch('/api/sync/offline-queue', {
          method: 'POST',
          body: JSON.stringify(entry),
        });

        if (response.ok) {
          await this.db.put('offline_scans', { ...entry, synced: true });
          result.synced++;
        } else if (response.status === 409) {
          // Konflikt: Equipment wurde von anderem User verändert
          const conflict = await response.json();
          result.conflicts.push({ localScan: entry, serverState: conflict });
        }
      } catch {
        result.failed++;
      }
    }

    return result;
  }
}
```

---

## Modul 15: Federation — Mehrstandort und Netzwerk-Austausch

### Überblick

Das Federation-Modul ermöglicht MyRMS-Instanzen verschiedener Firmen miteinander zu verbinden. Jede Firma behält vollständige Datensouveränität — es gibt keinen zentralen Server.

### 15.1 Federation-Architektur

```
Firma A (MB Veranstaltungstechnik)    Firma B (SoundPro GmbH)
┌─────────────────────────────┐        ┌──────────────────────────────┐
│  federation-service         │◄─mTLS─►│  federation-service          │
│  Port: 8009                 │        │  Port: 8009                  │
│                             │        │                              │
│  GIBT FREI:                 │        │  GIBT FREI:                  │
│  - Licht-Equipment          │        │  - PA-Systeme (K2, A15)      │
│  - Traversensystem          │        │  - Mischpulte (auf Anfrage)  │
│                             │        │                              │
│  SIEHT NICHT:               │        │  SIEHT NICHT:                │
│  - Stefans Kundenliste      │        │  - Marcos Finanzdaten        │
│  - Umsatzdaten              │        │  - Ecuipment-Gesamtwert      │
└─────────────────────────────┘        └──────────────────────────────┘
```

### 15.2 Partnerschaft einrichten

#### Schritt-für-Schritt-Einrichtung

```
1. EINLADUNG ERSTELLEN (Firma A):
   → Admin → Federation → "Partner einladen"
   → Firma-Name eingeben: "SoundPro GmbH"
   → Token generieren (gültig 72 Stunden)
   → Marco schickt Stefan den Token (per E-Mail, WhatsApp, persönlich)

   Generierter Token: MYRMS-FED-2026-a4f8c2e1-...

2. EINLADUNG ANNEHMEN (Firma B):
   → Stefan in seiner MyRMS-Instanz: Admin → Federation → "Einladung eingeben"
   → Token eingeben
   → Firmenname von Firma A wird angezeigt: "MB Veranstaltungstechnik"
   → Stefan klickt: "Partnerschaft annehmen"
   → Beide Systeme tauschen automatisch mTLS-Zertifikate aus

3. KONFIGURATION (Firma B):
   → Stefan konfiguriert was Marco sehen darf:
   ┌─────────────────────────────────────────────────────┐
   │ FREIGABE FÜR: MB Veranstaltungstechnik               │
   │                                                      │
   │ Freigegebene Kategorien:                             │
   │ ☑ PA-Systeme → L-Acoustics K2        (Verfügbarkeit + Preis)
   │ ☑ PA-Systeme → L-Acoustics A15       (Verfügbarkeit + Preis)
   │ ☑ PA-Systeme → Subwoofer SB28        (Verfügbarkeit + Preis)
   │ ☐ Mischpulte                          (nicht freigegeben)
   │ ☐ Mikrofone                           (nicht freigegeben)
   │                                                      │
   │ Preismodell: Partner-Kondition (-15% von Listenpreis)│
   │ Bestätigungsmodus: ○ Automatisch ● Manuell           │
   │ Sichtbarkeit: Verfügbarkeit bis 30 Tage im Voraus   │
   └─────────────────────────────────────────────────────┘

4. FERTIG:
   → Marco sieht in "Sub-Rental": SoundPro GmbH (aktiv)
   → Marco kann Verfügbarkeit prüfen ohne anzurufen
```

### 15.3 Sub-Rental-Workflow

#### Verfügbarkeitsabfrage

```go
// Marcos System fragt Stefans System via mTLS ab
func (s *FederationService) QueryPartnerAvailability(ctx context.Context, partnerID uuid.UUID, query AvailabilityQuery) (*AvailabilityResponse, error) {
    partner, _ := s.partnerRepo.GetByID(ctx, partnerID)

    // mTLS-Client erstellen
    client, _ := s.createMTLSClient(partner)

    // Federation-API aufrufen (Stefans Server)
    req := FederationAvailabilityRequest{
        CategorySlugs: query.Categories,
        DateFrom:      query.From,
        DateTo:        query.To,
        Quantity:      query.Quantity,
    }

    resp, _ := client.Post(partner.URL+"/federation/v1/availability", req)

    // Antwort: Liste verfügbarer Equipment mit Partner-Preisen
    // KEINE sensiblen Daten (keine Kundenliste, kein Gesamtumsatz)
    return resp, nil
}
```

#### Sub-Rental-Anfrage

```
MARCO (Firma A) — erstellt Anfrage:
  Equipment: 4× L-Acoustics K2 Tops
  Zeitraum: 22.-24. März 2026
  Projekt: Festival Stadtpark (intern, Stefan sieht nur den Zeitraum)
  Anmerkung: "Brauche für Freilichtbühne, 3.000 Pers."

SYSTEM sendet via Federation-API an Stefans Instanz:
  {
    "requester_name": "MB Veranstaltungstechnik",
    "equipment_external_id": "lacoust-k2-tops",  // Kein UUID aus unserem System!
    "equipment_name": "L-Acoustics K2 Tops",
    "quantity": 4,
    "date_from": "2026-03-22T07:00:00Z",
    "date_to": "2026-03-24T22:00:00Z",
    "note": "Brauche für Freilichtbühne, 3.000 Pers."
  }

STEFAN (Firma B) — sieht in seiner Instanz:
  Eingehende Anfrage: MB Veranstaltungstechnik
  Anfrage: 4× L-Acoustics K2 Tops, 22.-24.03.2026
  [Bestätigen] [Ablehnen] [Alternative vorschlagen]

STEFAN bestätigt → Marcos System zeigt:
  ✓ Bestätigt: 4× K2 Tops @ 480€/Tag × 3 Tage = 5.760€
  [Auf Rechnung anwenden] [Abholplan erstellen]
```

### 15.4 Übergabe-Dokumentation

```go
// Bei physischer Übergabe: Zustandsdokumentation in BEIDEN Systemen

// Firma A schickt Fahrer zu Firma B:
func (s *FederationService) DocumentHandover(ctx context.Context, requestID uuid.UUID, handover HandoverData) error {
    // 1. Fotos und Zustandsnotizen aufnehmen (bei Firma B)
    // 2. Handover-Dokument erstellen
    // 3. An eigenes System senden (Rückgabebeweis)
    // 4. An Partner senden via Federation-API (zu Informationszwecken)

    // Digitale Unterschrift von Stefans Mitarbeiter
    // Zeitgestempel Fotos für jede K2-Box (Kratzer, Beschädigungen vorher dokumentieren)

    // QR-Codes der abgeholten Geräte scannen
    // → In Marcos System: Status = "sub_rental_in" mit Fotos
    // → In Stefans System: Status = "sub_rental_out" mit denselben Fotos

    return nil
}
```

### 15.5 Automatische Abrechnung nach Sub-Rental

```
Nach Rückgabe der K2 (Montag 08:00):

1. Lisa bringt K2 zurück zu Stefan, scannt Rückgabe
2. Stefans System registriert: 4× K2, zurück, Zustand OK, Zeitraum abgeschlossen
3. AUTOMATISCH: Stefans invoice-service erstellt Ausgangsrechnung an Marcos Firma
4. Marcos invoice-service erhält die Rechnung via Federation-API als Eingangsrechnung
5. Thomas sieht in seiner Eingangsrechnungsliste:
   "SoundPro GmbH — 4× K2 Tops, 22.-24.03. — 5.760,00€"
   → Zum direkten Verbuchen bereit
```

---

## Modul 16: Audit — GoBD-konformer Audit-Trail (Modul N8)

### Überblick

Das Audit-Modul protokolliert unveränderlich jede relevante Aktion im System. Es erfüllt die Anforderungen der GoBD (Grundsätze ordnungsgemäßer Buchführung und Dokumentation bei IT-gestützten Buchführungssystemen) für deutsche Unternehmen.

### 16.1 GoBD-Anforderungen

Die GoBD schreibt für EDV-gestützte Buchführungssysteme vor:

| Anforderung | MyRMS-Implementierung |
|------------|----------------------|
| **Unveränderlichkeit** | Kein UPDATE/DELETE auf audit_log, nur INSERT |
| **Vollständigkeit** | Jedes Domain-Event wird protokolliert ($all Subscription) |
| **Nachvollziehbarkeit** | Jeder Eintrag hat User-ID, Zeitstempel, Event-Typ |
| **Ordnung** | Sequenznummer, Checksummen-Kette |
| **Zeitgerechte Buchung** | Events werden sofort beim Auftreten protokolliert |
| **Aufbewahrung** | 10 Jahre (konfigurierbar) |
| **Exportierbarkeit** | Export für Betriebsprüfung |

### 16.2 Checksummen-Kette

```
Unveränderliche Kettensignatur:

Eintrag 1:   checksum = SHA256(event_id_1 || payload_1 || "")
Eintrag 2:   checksum = SHA256(event_id_2 || payload_2 || checksum_1)
Eintrag 3:   checksum = SHA256(event_id_3 || payload_3 || checksum_2)
...
Eintrag N:   checksum = SHA256(event_id_N || payload_N || checksum_N-1)

Wenn Eintrag 2 verändert würde:
  → checksum_2 würde sich ändern
  → checksum_3 (abhängig von checksum_2) wäre ungültig
  → Alle folgenden Checksummen wären ungültig
  → Manipulation sofort erkennbar
```

```go
func (r *AuditRepo) VerifyChain(ctx context.Context, tenantID uuid.UUID, from, to int64) VerificationResult {
    entries, _ := r.db.Query(ctx, `
        SELECT sequence_number, event_id, payload, checksum, prev_checksum
        FROM audit_schema.audit_log
        WHERE tenant_id = $1 AND sequence_number BETWEEN $2 AND $3
        ORDER BY sequence_number ASC
    `, tenantID, from, to)

    prevChecksum := ""
    for _, entry := range entries {
        // Erwarteten Checksum berechnen
        payload, _ := json.Marshal(entry.Payload)
        expected := sha256.Sum256([]byte(entry.EventID.String() + string(payload) + prevChecksum))
        expectedHex := hex.EncodeToString(expected[:])

        if entry.Checksum != expectedHex {
            return VerificationResult{
                IsValid:         false,
                ErrorAtSequence: entry.SequenceNumber,
                Message: fmt.Sprintf("Checksum-Fehler bei Sequenz %d", entry.SequenceNumber),
            }
        }

        prevChecksum = entry.Checksum
    }

    return VerificationResult{IsValid: true, VerifiedEntries: len(entries)}
}
```

### 16.3 Audit-Log Abfragen

```
Audit-Log-Suche (nur für Admin und Buchhaltung)

Filter-Optionen:
  Zeitraum:         Von — Bis (Datum/Uhrzeit)
  Event-Typ:        invoice.created, equipment.checked_out, user.login, ...
  Aggregate-Typ:    invoice, equipment, project, user, ...
  User:             Wer hat die Aktion ausgeführt?
  Service:          Welcher Service hat das Event erzeugt?

Beispiel-Abfrage:
  "Alle Änderungen an Rechnung RE-2026-042"
  → Zeigt Timeline aller Events für diese Rechnung:
    2026-03-15 10:23 | invoice.created       | thomas@mbva.de  | invoice-service
    2026-03-15 10:24 | invoice.item_added    | thomas@mbva.de  | invoice-service
    2026-03-15 14:30 | invoice.sent          | thomas@mbva.de  | invoice-service
    2026-03-28 09:15 | invoice.payment_received | system       | invoice-service
    2026-03-28 09:15 | invoice.paid          | system          | invoice-service
```

### 16.4 DSGVO-Konformität

```go
// DSGVO Recht auf Vergessenwerden
// Problem: Audit-Log ist unveränderlich (GoBD), aber DSGVO fordert Löschbarkeit

// Lösung: Pseudonymisierung statt Löschung
func (s *AuditService) PseudonymizeUser(ctx context.Context, userID uuid.UUID, reason string) error {
    // Ersetze alle user_email-Einträge mit anonymem Wert
    // user_id bleibt (für Rückverfolgbarkeit intern)
    // aber Email/Name werden unkenntlich gemacht

    s.db.Exec(ctx, `
        UPDATE audit_schema.audit_log
        SET user_email = 'deleted-user@anon.invalid'
        WHERE user_id = $1
    `, userID)

    // ABER: Die eigentlichen Event-Payloads bleiben unverändert
    // (GoBD: Buchungsbelege dürfen nicht geändert werden)
    // → Kompromiss: Personenbezogene Daten in Metadaten entfernen,
    //   buchungsrelevante Daten bleiben erhalten

    // Protokolliere die Pseudonymisierung selbst als Audit-Event
    s.appendAuditEntry(ctx, AuditEntry{
        EventType:  "gdpr.user_pseudonymized",
        AggregateType: "user",
        AggregateID: userID,
        Payload: map[string]interface{}{
            "reason": reason,
            "gdpr_request_date": time.Now(),
        },
    })

    return nil
}
```

### 16.5 Betriebsprüfungs-Export

```
Export für Betriebsprüfung (Finanzamt):

Format: GoBD-konformes ZIP-Paket mit:
  /audit-log/
    audit_log_2024_Q1.csv    → Alle Buchungsevents Q1 2024
    audit_log_2024_Q2.csv    → Alle Buchungsevents Q2 2024
    ...
    chain_verification.txt   → Protokoll der Checksummen-Prüfung
  /invoices/
    RE-2024-001.pdf          → Alle Rechnungs-PDFs
    RE-2024-002.pdf
    ...
    GU-2024-001.pdf          → Gutschriften
  /documents/
    lieferschein_001.pdf     → Alle Lieferscheine
    ...
  /metadata/
    export_info.txt          → Exportzeitraum, Exporteur, Checksumme des Exports

Zeitstempel und Checksumme des gesamten Exports:
  → Thomas kann beweisen dass der Export unverändert ist
```
