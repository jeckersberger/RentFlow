# MyRMS — Kern-Module Spezifikation

**Stand:** 20. März 2026
**Version:** 1.0

Dieses Dokument beschreibt die Detailspezifikation der drei Kernmodule: **Inventory** (Artikelverwaltung), **Project** (Projekte und Buchungen) und **Invoice** (Rechnungsstellung und Buchhaltung).

---

## Modul 1: Inventory — Artikelverwaltung

### Überblick

Das Inventory-Modul ist das Herzstück von MyRMS. Es verwaltet das gesamte Equipment-Portfolio: von der 5€-Kabelverbindung bis zur 50.000€-PA-Anlage. Zielgruppe: Lisa (tägliche Lagerarbeit) und Marco (strategischer Überblick).

### 1.1 Equipment-Stammdaten

#### Pflichtfelder beim Anlegen

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `name` | String | Vollständige Bezeichnung (z.B. "L-Acoustics K2 Lautsprecher-Modul") |
| `category_id` | UUID | Pflichtauswahl aus Kategoriepfad |
| `quantity_total` | Integer | Gesamtbestand (mind. 1) |

#### Optionale, aber wichtige Felder

| Feld | Typ | Beschreibung |
|------|-----|-------------|
| `internal_number` | String | Interne Artikelnummer (z.B. "LT-0042", "PA-K2-001") |
| `serial_number` | String | Herstellerseriennummer (wichtig für Wartung + Versicherung) |
| `manufacturer` | String | Herstellername |
| `model` | String | Modellbezeichnung |
| `purchase_price` | Decimal | Einkaufspreis (nur für Admin/Buchhaltung sichtbar) |
| `replacement_value` | Decimal | Wiederbeschaffungswert (für Versicherung) |
| `weight_kg` | Decimal | Gewicht in kg (für Transportplanung) |
| `power_consumption_w` | Integer | Leistungsaufnahme (für Stromanschluss-Kalkulation) |
| `barcode` / `qr_code` | String | Scanner-Codes (auto-generiert oder manuell) |
| `images` | Array | Produktfotos (min. 1 empfohlen) |
| `technical_specs` | JSON | Hersteller-Spezifikationen (frei strukturierbar) |
| `internal_notes` | Text | Interne Hinweise (nur Staff) |

#### Auto-Generierung der Artikelnummer

Beim Erstellen eines neuen Equipments wird automatisch eine interne Nummer generiert, wenn keine angegeben:

```
Format: {Kategorie-Prefix}-{4-stellige-Nummer}
Beispiele: LT-0042 (Licht), TO-0103 (Ton), VI-0005 (Video), CA-0018 (Kabel)
```

Der Benutzer kann die generierte Nummer überschreiben. Die Nummer muss innerhalb des Mandanten eindeutig sein.

### 1.2 Kategorie-System

Kategorien sind hierarchisch (bis 3 Ebenen tief) und mandanten-spezifisch:

```
Licht
├── Moving Heads
│   ├── Beam
│   ├── Wash
│   └── Spot/Beam/Wash
├── Conventional
│   ├── PAR
│   └── Fresnel
└── LED-Panels

Ton
├── PA-Systeme
│   ├── Line Arrays
│   └── Point Source
├── Subwoofer
├── Monitoring
└── Mischpulte

Video
├── Projektoren
├── LED-Wände
└── Screens

Rigging & Struktur
├── Traversensysteme
└── Hoist / Motor

Kabel & Adapter
├── Audiokabel
├── DMX-Kabel
├── Starkstrom
└── Adapter

Verbrauchsmaterial
├── Lampen
├── Batterien
└── Kabelbinder
```

Jede Kategorie hat:
- Name (mehrsprachig via i18n-Key)
- Icon (Emoji oder Lucide-Icon-Name)
- Farbe (für visuelle Unterscheidung in der UI)
- Kategorie-Prefix für Artikelnummern
- Übergeordnete Kategorie (optional)

### 1.3 Mengenverwaltung und Zustand

#### Mengenfelder (denormalisiert für Performance)

```
quantity_total      = Gesamtbestand
quantity_available  = Verfügbar (im Lager, funktionsfähig)
quantity_rented     = Verliehen (ausgecheckt)
quantity_defect     = Defekt (gesperrt)
quantity_maintenance = In Wartung

Invariante: quantity_total = quantity_available + quantity_rented + quantity_defect + quantity_maintenance
```

Diese Felder werden durch Event-Projektion aktuell gehalten, nie direkt geschrieben.

#### Zustände eines Artikels

| Zustand | Code | Beschreibung |
|---------|------|-------------|
| Verfügbar | `available` | Im Lager, kann ausgegeben werden |
| Verliehen | `rented` | Aktuell beim Kunden oder auf Baustelle |
| Reserviert | `reserved` | Für Projekt reserviert (noch im Lager) |
| Defekt | `defect` | Gesperrt, nicht ausgebbar |
| In Wartung | `maintenance` | In Reparatur/Prüfung |
| Ausgemustert | `retired` | Aus dem Bestand entfernt, nicht mehr verfügbar |

#### Zustandsänderungs-Regeln

```
available → rented        : Durch Check-Out-Scan
rented → available        : Durch Check-In-Scan
available → defect        : Durch Defekt-Meldung (Scan oder manuell)
defect → maintenance      : Durch Wartungsauftrag
maintenance → available   : Durch Abschluss der Wartung
maintenance → defect      : Wenn Reparatur fehlschlägt
available → retired       : Nur durch Admin, unwiderruflich (kein direktes zurück)
```

### 1.4 Preiskalkulations-Engine

Die Preiskalkulation in MyRMS unterstützt komplexe Preisstrukturen, die in der Veranstaltungstechnik üblich sind.

#### Preisregel-Typen

**Tagessatz (Standard)**
```
base_price = 100€/Tag
Berechnung: Tage × 100€
```

**Wochenpauschale**
```
base_price = 500€/Woche
Berechnung: Beim Überschreiten von 7 Tagen greift Wochensatz
```

**Staffelpreise (Zeitrabatt)**
```
1-3 Tage:   100€/Tag
4-7 Tage:   80€/Tag   (tier_2_from=4, tier_2_price=80)
8+ Tage:    60€/Tag   (tier_3_from=8, tier_3_price=60)
```

**Mengenstaffel**
```
1-4 Stück:  100€/Stück/Tag
5+ Stück:   85€/Stück/Tag  (qty_discount_threshold=5, qty_discount_pct=15)
```

**Saisonzuschlag**
```
Normaler Preis: 100€/Tag
Festival-Saison (Jun-Aug): +20% → 120€/Tag
(season_surcharge_pct=20, season_start="06-01", season_end="08-31")
```

**Bundle-Preis**
```
Ton-Paket A (CL5 + 4x Wedge + Kabelset):
  Einzelpreise: 300 + 4×50 + 100 = 600€
  Bundle-Preis: 480€ (20% günstiger als Einzelteile)
```

#### Preis-Berechnungslogik

```go
func (e *PriceEngine) Calculate(req PriceRequest) *PriceResult {
    // 1. Passende Preisregel finden (Priorität: Equipment > Kategorie > Standard)
    rule := e.findBestRule(req.EquipmentID, req.CategoryID, req.Date)
    if rule == nil {
        return &PriceResult{Price: 0, Error: "Keine Preisregel gefunden"}
    }

    days := req.RentalDays
    qty := req.Quantity

    // 2. Tagessatz bestimmen (mit Staffelpreis)
    dailyRate := rule.BasePrice
    if rule.Tier3From != nil && days >= *rule.Tier3From {
        dailyRate = *rule.Tier3Price
    } else if rule.Tier2From != nil && days >= *rule.Tier2From {
        dailyRate = *rule.Tier2Price
    }

    // 3. Mengenstaffel
    if rule.QtyDiscountThreshold != nil && qty >= *rule.QtyDiscountThreshold {
        dailyRate = dailyRate * (1 - *rule.QtyDiscountPct/100)
    }

    // 4. Saisonzuschlag
    if rule.IsInSeason(req.Date) {
        dailyRate = dailyRate * (1 + rule.SeasonSurchargePct/100)
    }

    unitPrice := dailyRate * float64(days)
    total := unitPrice * float64(qty)

    return &PriceResult{
        UnitPrice:    unitPrice,
        DailyRate:    dailyRate,
        RentalDays:   days,
        Quantity:     qty,
        Total:        total,
        PriceRuleID:  rule.ID,
        Calculation:  fmt.Sprintf("%.2f€/Tag × %d Tage × %d Stück", dailyRate, days, qty),
    }
}
```

### 1.5 QR-Labels und Barcode-Management

#### Label-Generierung

Jedes Equipment-Label enthält:
- QR-Code (Equipment-UUID, für Scanner-App)
- Barcode CODE-128 (Interne Artikelnummer, für USB-Scanner)
- Klartext: Interne Nummer + Name
- Logo (optional, mandanten-spezifisch)

#### ZPL-Template für Zebra-Drucker

```zpl
^XA
^MMT
^PW812
^LL0203
^LS0

^FO50,20^GFA,800,800,...  // Logo
^FO50,80^BY3^BCN,80,Y,N,N^FD{{.InternalNumber}}^FS  // Barcode
^FO50,180^BQN,2,5^FDQA,{{.ID}}^FS  // QR-Code
^FO300,80^A0N,30,30^FD{{.InternalNumber}}^FS
^FO300,120^A0N,25,25^FD{{.Name | trunc 20}}^FS
^FO300,155^A0N,20,20^FD{{.Manufacturer}}^FS

^PQ1,0,1,Y
^XZ
```

#### Label-Druck-Workflow

1. Lisa scannt zu beschriftende Artikel
2. Oder: Bulk-Auswahl in der Tabelle
3. Klick auf "Labels drucken"
4. Dialog: Anzahl Kopien, Drucker auswählen (wenn mehrere konfiguriert)
5. ZPL wird an Zebra-Drucker via USB/WLAN gesendet
6. Alternativ: PDF-Download für Office-Drucker

### 1.6 Flightcase-Management (Modul N5)

Flightcases sind die Behälter für Equipment. MyRMS verwaltet sie als eigene Entitäten:

#### 3-Tier-Validierung bei Check-Out

```
Tier 1: Harter Constraint (BLOCKIERT Check-Out)
  → Pflicht-Equipment fehlt (z.B. Netzkabel zum PA-Modul)
  → Zeige Fehlermeldung: "Netzkabel C13/C14 ist Pflichtinhalt, fehlt in diesem Flightcase"

Tier 2: Weicher Constraint (WARNUNG, aber weiter möglich)
  → Empfohlenes Equipment fehlt (z.B. Reservelampe)
  → Zeige Warnung: "Reservelampe Par64 empfohlen, aber nicht im Case"

Tier 3: Information (NUR HINWEIS)
  → Optionales Zubehör verfügbar aber nicht eingeplant
  → Zeige Info: "Es gibt optionale Adapter für dieses Set"
```

Jeder Flightcase hat eine `case_type`-Konfiguration, die definiert welches Equipment wie validiert wird.

---

## Modul 2: Project — Projekte, Buchungen, Verfügbarkeit

### Überblick

Das Project-Modul steuert den kompletten Projekt-Lebenszyklus: von der ersten Kundenanfrage bis zur abgeschlossenen Veranstaltung. Es enthält die kritische Verfügbarkeitslogik, die Doppelbuchungen verhindert.

### 2.1 Projekt-Lifecycle

```
Statusübergänge:

  inquiry ────────────────────────────────► cancelled
     │
     ▼
  offer_sent ──────────────────────────────► cancelled
     │
     ▼ (Kunde bestätigt)
  confirmed ──────────────────────────────► cancelled
     │
     ▼ (Lager bereitet vor)
  in_preparation
     │
     ▼ (Equipment geht raus)
  active
     │
     ▼ (Alles zurück, alles abgehakt)
  completed
     │
     ▼ (Rechnung geschrieben)
  invoiced (Endzustand)
```

**Statusübergangs-Regeln:**
- Nur bestätigte Projekte reservieren Equipment (blockieren Verfügbarkeit)
- `active` wird automatisch gesetzt wenn erstes Equipment per Scanner ausgegeben wird
- `completed` wird automatisch gesetzt wenn letztes Equipment zurückgegangen UND alle Packlisten-Positionen abgehakt sind
- `invoiced` wird von invoice-service gesetzt wenn Rechnung erstellt wurde
- Jeder Statuswechsel erzeugt ein KurrentDB-Event: `ProjectStatusChanged`

### 2.2 Packlisten-Verwaltung

#### Packlisten-Erstellung

Die Packliste entsteht aus dem Angebot:

1. Marco erstellt Angebot mit Equipment-Positionen (inventory-service liefert Preise)
2. Kunde bestätigt → Projekt wird auf `confirmed` gesetzt
3. Packliste wird automatisch aus Angebotspositionen generiert
4. Lisa sieht Packliste auf Zebra-Handheld, sortiert nach Lagerplatz

#### Lagerplatz-Sortierung für optimale Kommissionierroute

```go
// Packlisten-Positionen werden nach Lagerplatz sortiert,
// damit Lisa nicht quer durchs Lager läuft

func sortByOptimalPickRoute(items []PackingListItem, warehouseLayout WarehouseLayout) []PackingListItem {
    // 1. Lagerplätze zu Regalnummern auflösen
    for i := range items {
        items[i].location = warehouse.GetLocation(items[i].locationID)
    }

    // 2. Sortierung: Zone → Regal → Fach (alphabetisch/numerisch)
    sort.Slice(items, func(i, j int) bool {
        locI := items[i].location
        locJ := items[j].location
        if locI.Zone != locJ.Zone { return locI.Zone < locJ.Zone }
        if locI.Rack != locJ.Rack { return locI.Rack < locJ.Rack }
        return locI.Bay < locJ.Bay
    })

    return items
}
```

#### Packlisten-Abschnitte

Für große Projekte werden Packlisten in Abschnitte unterteilt:

```
Festival Setup - Packliste
├── Ton (32 Positionen)
│   ├── K2 Tops × 4  [Regal A1]  ✓ Ausgegeben
│   ├── SB28 Subs × 2  [Regal A2]  ⬜ Ausstehend
│   └── CL5 Pult × 1  [Regal B3]  ⬜ Ausstehend
├── Licht (18 Positionen)
│   ├── Claypaky B-Eye × 8  [Regal C1]  ✓ Ausgegeben
│   └── ...
└── Kabel & Adapter (47 Positionen)
    └── ...
```

### 2.3 Verfügbarkeitsprüfung und Doppelbuchungs-Schutz

#### Echtzeit-Verfügbarkeit

```go
// Verfügbarkeit wird bei jedem Hinzufügen zur Packliste geprüft
// Ergebnis wird im Frontend direkt angezeigt

type AvailabilityCheck struct {
    EquipmentID    uuid.UUID
    ProjectID      uuid.UUID    // Ausschluss: eigenes Projekt nicht als Konflikt werten
    From           time.Time
    To             time.Time
    RequestedQty   int
}

type AvailabilityResult struct {
    IsAvailable      bool
    AvailableQty     int
    TotalQty         int
    ConflictingProjects []ConflictingProject
    AlternativeSuggestions []AlternativeEquipment  // KI: ähnliche Artikel
}
```

#### Visuelles Feedback im Frontend

```
Verfügbar:    ✓ Grün — "4 von 6 verfügbar"
Knapp:        ⚠ Gelb — "2 von 6 verfügbar (benötigt: 4)"
Nicht verfügbar: ✗ Rot — "0 verfügbar — belegt durch: Festival Stadtpark 15.-17.03."
```

#### Konflikterkennung bei Doppelbuchung

Wenn zwei bestätigte Projekte dasselbe Equipment im selben Zeitraum benötigen:

1. System erkennt Konflikt beim Bestätigen des zweiten Projekts
2. Event `AvailabilityConflictDetected` → KurrentDB
3. workflow-service triggert Benachrichtigung an Projektleiter
4. Konflikt erscheint im Dashboard mit rotem Alert
5. Marco entscheidet: Equipment umplanen, Sub-Rental anfragen, oder Priorität setzen

#### Sub-Rental-Vorschlag bei Knappheit

Wenn Equipment nicht verfügbar und Federation-Partner konfiguriert:

```
Dashboard-Alert: "CL5 nicht verfügbar (15.-17.03)"
Automatischer Vorschlag: "SoundPro hat 2× CL5 verfügbar (Verfügbarkeitsprüfung via Federation)"
Aktion: "Sub-Rental anfragen" → öffnet Federation-Modul
```

### 2.4 Kalender und Timeline-Ansicht (Modul N6)

#### Gantt-Ansicht (Projektleiter-Perspektive)

- Horizontal: Zeitachse (Tage/Wochen)
- Vertikal: Projekte
- Farbkodierung nach Status
- Klick auf Balken: Schnellübersicht
- Drag & Drop: Projekt-Termine verschieben (mit Re-Validierung)

#### Equipment-Verfügbarkeits-Kalender

- Zeigt welches Equipment wann ausgebucht ist
- Hilfreich für Angebots-Erstellung: "Wann sind die K2 das nächste Mal frei?"
- Filter: Nach Kategorie, nach Projekt, nach Zeitraum

#### CalDAV-Integration

```
Projekte werden als CalDAV-Events exportiert:
  - Aufbau-Zeitraum: DTSTART/DTEND
  - Veranstaltung: als Subevent
  - Abbau: als Subevent

CalDAV-Feed-URL pro User:
  https://myrms.example.com/api/calendar/{user_id}/calendar.ics?token={caldav_token}

Kompatibel mit:
  - iPhone Kalender (iCloud)
  - Android Kalender (Google Calendar)
  - Outlook (via CalDAV-Plugin)
  - Thunderbird
```

### 2.5 Kundenverwaltung

#### Kundenprofil

Jeder Kunde (Firma oder Privatperson) hat:
- Stammdaten: Name, Adresse, Kontakt
- Zahlungskonditionen: Standard-Zahlungsziel (z.B. 14 Tage), Kreditlimit
- Projekt-Historie: Alle vergangenen und aktuellen Projekte
- Rechnungs-Historie: Alle Rechnungen, Mahnungen, Zahlungen
- Tags: Freie Kategorisierung (z.B. "VIP-Kunde", "Festival", "Corporate")
- Notizen: Interne Anmerkungen

#### Portal-Zugang für Kunden (Modul N4)

Kunden können optional einen Portal-Zugang erhalten:
- Nur eigene Projekte und Rechnungen sichtbar
- Angebote online bestätigen (mit digitaler Unterschrift)
- Rechnungen herunterladen
- Verfügbarkeit anfragen (öffentlicher Katalog)
- Kommunikation: Nachrichten an Projektleiter

---

## Modul 3: Invoice — Rechnungen, Angebote, Buchhaltung

### Überblick

Das Invoice-Modul deckt den gesamten kaufmännischen Prozess ab: Angebot schreiben, Rechnung erstellen, Mahnungen versenden, DATEV-Export und Bank-Abgleich. Primäre Zielgruppe: Thomas (Buchhaltung).

### 3.1 Angebots-Erstellung

#### Angebots-Workflow

```
1. Neues Angebot erstellen
   ├── Verknüpfung mit bestehendem Projekt (empfohlen)
   └── Oder: Standalone-Angebot

2. Positionen aus Packliste übernehmen (ein Klick)
   → Preise werden automatisch berechnet (inventory-service Preisregel)
   → Mengen aus Packliste übernommen
   → Manuell ergänzen: Serviceleistungen, Transportkosten, etc.

3. Anpassungen
   ├── Mengen anpassen
   ├── Preise überschreiben (mit Begründung)
   ├── Rabatt gewähren (gesamt oder pro Position)
   └── Sonderkonditionen eintragen

4. Angebot versenden
   ├── PDF per E-Mail (via SMTP, E-Mail-Template)
   └── Link zum Kunden-Portal (digitale Bestätigung)

5. Angebot wird angenommen
   └── Ein-Klick-Konvertierung zu Rechnung
```

#### Angebots-Template

Das Angebots-PDF enthält:
- Firmenbriefkopf (Logo, Adresse, Steuernummer, IBAN)
- Kundendaten und Anschrift
- Angebotsnummer und -datum
- Gültigkeitsdatum
- Projektbeschreibung
- Positionen (Tabelle): Nr., Beschreibung, Menge, Einheit, Einzelpreis, Rabatt, Gesamtpreis
- Zwischensumme, MwSt (19%), Gesamtbetrag
- Zahlungskonditionen
- AGB / Hinweise
- Unterschriften-Bereich (bei Papierversion)

### 3.2 Rechnungserstellung

#### Aus Angebot konvertieren (1-Klick)

```go
func (s *InvoiceService) ConvertQuoteToInvoice(ctx context.Context, quoteID uuid.UUID) (*Invoice, error) {
    quote, err := s.quoteRepo.GetByID(ctx, quoteID)
    if err != nil { return nil, err }

    if quote.Status != "accepted" {
        return nil, ErrQuoteNotAccepted
    }

    // Neue Rechnungsnummer generieren
    invoiceNumber, err := s.generateInvoiceNumber(ctx, quote.TenantID)

    // Rechnung erstellen
    invoice := &Invoice{
        TenantID:      quote.TenantID,
        CustomerID:    quote.CustomerID,
        ProjectID:     quote.ProjectID,
        QuoteID:       &quote.ID,
        InvoiceNumber: invoiceNumber,
        InvoiceType:   "invoice",
        Status:        "draft",
        Subject:       quote.Subject,
        IntroText:     quote.IntroText,
        OutroText:     quote.OutroText,
        TaxRate:       quote.TaxRate,
        InvoiceDate:   time.Now().Truncate(24 * time.Hour),
        DueDate:       time.Now().Add(time.Duration(quote.PaymentTerms) * 24 * time.Hour).Truncate(24 * time.Hour),
        PaymentTerms:  quote.PaymentTerms,
        Items:         convertItems(quote.Items),  // Positionen übernehmen
    }

    // Summen berechnen
    invoice.RecalculateTotals()

    // In DB speichern
    err = s.invoiceRepo.Create(ctx, invoice)

    // Event: InvoiceCreated → KurrentDB
    s.eventStore.Append(ctx, "invoice-"+invoice.ID.String(), events.InvoiceCreated{...})

    return invoice, nil
}
```

#### Direktrechnung (ohne Angebot)

Wenn kein vorheriges Angebot existiert:
1. Projekt auswählen → Packliste als Basis
2. Tatsächlich gelieferte Mengen anpassen (wichtig: Scanner-Daten fließen automatisch ein)
3. Zusatzpositionen hinzufügen
4. Rechnung erstellen

**Integration Scanner → Rechnung:**
Der scanner-service sendet bei jedem Check-In/Check-Out Events. invoice-service kann diese Events auswerten und der Rechnung "tatsächlich gelieferte Menge" vorschlagen:

```
Angebot sagte: 8× Moving Head
Tatsächlich ausgegeben (laut Scanner): 10× Moving Head (Kunde hat 2 kurzfristig dazugenommen)
invoice-service zeigt: "Hinweis: 2 zusätzliche Moving Heads ausgegeben (ohne Angebot)"
Thomas klickt: "Differenz übernehmen" → 10 in der Rechnung
```

### 3.3 Teilrechnungen

Häufig bei Großprojekten: Zahlung in Raten (z.B. 30/50/20 Split).

#### Teilrechnungs-Setup

```
Projekt: Festival Stadtpark 2026
Gesamtwert: 15.000€

Teilrechnungsplan:
  1. Anzahlung (bei Bestätigung):  30% = 4.500€  → Rechnung RE-2026-042
  2. Nach Aufbau:                  50% = 7.500€  → Rechnung RE-2026-043
  3. Nach Abbau (Schlussrechnung): 20% = 3.000€  → Rechnung RE-2026-044

System berechnet automatisch:
  - Beträge pro Teilrechnung
  - Offenen Restbetrag
  - Bereits bezahlte Teile
  - Schlusskorrekturen (wenn sich Gesamtsumme ändert)
```

#### Teilrechnungs-Berechnung bei Auftragsänderung

```
Ursprünglicher Gesamtwert: 15.000€
Bereits gesendet: RE-2026-042 (Anzahlung) 30% = 4.500€

Auftragsänderung: +1.500€ → Neuer Gesamtwert: 16.500€

Berechnung RE-2026-043 (50% = 8.250€):
  Gesamtwert neu:          16.500€
  Soll-Stand nach 2. TR:   50% + 30% = 80% = 13.200€
  Bereits bezahlt:         4.500€ (Anzahlung)
  Zu berechnen:            13.200€ − 4.500€ = 8.700€

RE-2026-044 (Schlussrechnung):
  Restbetrag: 16.500€ − 4.500€ − 8.700€ = 3.300€
```

Das System berechnet dies automatisch. Thomas muss nur prüfen und bestätigen.

### 3.4 Mahnwesen

#### Automatischer Mahnlauf

Konfigurierbar pro Mandant:

```yaml
dunning_config:
  payment_reminder:       # Zahlungserinnerung (kein Mahngebühr)
    days_after_due: 3
    subject: "Zahlungserinnerung {{invoice_number}}"
    template: "reminder_email"

  dunning_level_1:        # Erste Mahnung
    days_after_due: 14
    fee: 5.00             # Mahngebühr in EUR
    subject: "1. Mahnung — Rechnung {{invoice_number}}"
    template: "dunning_1_email"

  dunning_level_2:        # Zweite Mahnung
    days_after_due: 28
    fee: 15.00
    subject: "2. Mahnung — Letzte Aufforderung"
    template: "dunning_2_email"

  dunning_level_3:        # Letzte Mahnung vor Inkasso
    days_after_due: 45
    fee: 25.00
    subject: "3. und letzte Mahnung"
    template: "dunning_3_email"
```

#### Mahnlauf-Trigger

- **Automatisch**: workflow-service CRON täglich um 08:00 Uhr
- **Manuell**: Thomas klickt "Mahnlauf jetzt ausführen" im Dashboard
- **Ausnahmen**: Kunden können von automatischen Mahnungen ausgenommen werden

#### KI-gestützte Mahnstufenwahl (Modul I14)

```go
// ai-service: Bevor Mahnung versendet wird, analysiert KI den Kunden
func (s *DunningService) GetRecommendedLevel(ctx context.Context, invoice Invoice) DunningRecommendation {
    history := s.getCustomerPaymentHistory(ctx, invoice.CustomerID)
    score := s.aiClient.PredictPaymentProbability(ctx, history, invoice)

    // Empfehlung basierend auf Score:
    if score.PaymentProbability > 0.85 {
        return DunningRecommendation{Level: "reminder", Note: "Guter Zahler, erstmal freundlich erinnern"}
    } else if score.PaymentProbability > 0.50 {
        return DunningRecommendation{Level: "dunning_1", Note: "Formale Mahnung"}
    } else {
        return DunningRecommendation{Level: "dunning_2", Note: "Risiko-Kunde, direkter vorgehen"}
    }
}
```

### 3.5 DATEV-Export

#### Unterstützte Exportformate

| Format | Verwendung |
|--------|-----------|
| DATEV Buchungsstapel CSV | Ausgangs- und Eingangsrechnungen |
| DATEV Debitoren/Kreditoren | Kundenstammdaten |
| DATEV SELF | Selbst buchen (vereinfacht) |
| WISO/Lexware CSV | Alternative Buchhaltungssoftware |

#### DATEV-Mapping

```go
// Kontenrahmen SKR03 (Standard für kleine Unternehmen)
var DATEVAccountMapping = map[string]string{
    "revenue_19pct":         "8400",  // Erlöse 19% MwSt
    "revenue_7pct":          "8300",  // Erlöse 7% MwSt
    "revenue_0pct":          "8125",  // Steuerfreie Erlöse
    "tax_liability_19pct":   "1776",  // Umsatzsteuer 19%
    "accounts_receivable":   "1400",  // Forderungen Debitoren (10000-69999 range)
    "accounts_payable":      "1600",  // Verbindlichkeiten Kreditoren
    "bank":                  "1200",  // Bank
}

// Debitorennummer: 10000 + laufende Kundennummer
func (e *DATEVExporter) GetDebtorAccount(customerNumber int) string {
    return fmt.Sprintf("%d", 10000 + customerNumber)
}
```

#### Export-Prozess

1. Thomas wählt Zeitraum (Monat/Quartal) und Format
2. System aggregiert alle relevanten Buchungen
3. DATEV-CSV wird generiert und als Download bereitgestellt
4. Optional: Direktversand per E-Mail an Steuerberater
5. Export-Protokoll: Was wurde exportiert, Zeitstempel, User

### 3.6 Bank-Import und OP-Abgleich

#### Unterstützte Importformate

- **MT940** (SWIFT-Standard, alle deutschen Banken)
- **CAMT.053** (ISO 20022, neuerer Standard)
- **CSV** (bankspezifische Exporte, konfigurierbar)

#### Automatischer OP-Abgleich

```go
func (s *BankService) AutoMatchTransactions(ctx context.Context, tenantID uuid.UUID) MatchResult {
    unmatched, _ := s.transactionRepo.GetUnmatched(ctx, tenantID)
    openInvoices, _ := s.invoiceRepo.GetOpenInvoices(ctx, tenantID)

    for _, tx := range unmatched {
        // Strategie 1: Exakter Betragsabgleich + Rechnungsnummer im Verwendungszweck
        for _, inv := range openInvoices {
            if tx.Amount == inv.Outstanding && strings.Contains(tx.Reference, inv.InvoiceNumber) {
                s.matchTransaction(ctx, tx.ID, inv.ID, "exact_match")
                continue
            }
        }

        // Strategie 2: Betrag passt (±1 Cent Toleranz für Überweisungsgebühren)
        for _, inv := range openInvoices {
            if math.Abs(float64(tx.Amount - inv.Outstanding)) < 0.02 {
                s.matchTransaction(ctx, tx.ID, inv.ID, "amount_match_tolerance")
            }
        }

        // Strategie 3: Kundennamen-Match
        // ...
    }
}
```

#### Dashboard: Offene Posten

Thomas sieht täglich auf einen Blick:

```
OFFENE POSTEN — Stand: 20.03.2026

  Gesamtaußenstände: 47.832,50€

  Überfällig (rot):
    RE-2026-032  Stadtwerke GmbH   3.420,00€  seit 32 Tagen   [2. Mahnung]
    RE-2026-038  Event AG          1.800,00€  seit 21 Tagen   [1. Mahnung]

  Fällig in < 7 Tagen (gelb):
    RE-2026-041  BMW AG           12.600,00€  fällig 22.03.

  Nicht fällig (grau):
    RE-2026-044  SoundEvents       5.800,00€  fällig 05.04.
    ...

  Aktionen:
    [Mahnlauf ausführen]  [DATEV exportieren]  [Offene Posten exportieren]
```

### 3.7 Gutschriften und Stornierungen

#### Gutschrift erstellen

Eine Gutschrift ist eine Rechnung mit negativem Betrag und Bezug zur Original-Rechnung:

```go
func (s *InvoiceService) CreateCreditNote(ctx context.Context, invoiceID uuid.UUID, items []CreditNoteItem) (*Invoice, error) {
    original, err := s.invoiceRepo.GetByID(ctx, invoiceID)

    creditNote := &Invoice{
        TenantID:        original.TenantID,
        CustomerID:      original.CustomerID,
        ProjectID:       original.ProjectID,
        ParentInvoiceID: &original.ID,
        InvoiceType:     "credit_note",
        Subject:         fmt.Sprintf("Gutschrift zu %s", original.InvoiceNumber),
        IntroText:       fmt.Sprintf("Hiermit erteilen wir Ihnen eine Gutschrift zu Rechnung %s.", original.InvoiceNumber),
        Items:           items,
        TaxRate:         original.TaxRate,
    }

    creditNote.RecalculateTotals()
    // Negativer Gesamtbetrag

    // Event: CreditNoteCreated → KurrentDB
    // audit-service protokolliert die Verbindung original → gutschrift

    return creditNote, nil
}
```

#### Rechnungsstorno

Eine stornierte Rechnung wird durch eine vollständige Gutschrift ausgeglichen:
- Original-Rechnung bleibt erhalten (GoBD-Konformität!)
- Gegenbuchung als Gutschrift über denselben Betrag
- Beide Dokumente sind miteinander verknüpft
- Status der Original-Rechnung: `cancelled` (nur als Anzeige, Dokument unverändert)

**Wichtig:** Direkte Löschung von Rechnungen ist aus GoBD-Gründen verboten. Das System bietet keine Lösch-Funktion für bestätigte Rechnungen an — nur Storno.
