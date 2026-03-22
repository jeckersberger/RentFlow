# MyRMS — KI-, Reporting-, Workflow- und Notification-Module Spezifikation

**Stand:** 20. März 2026
**Version:** 1.0

Dieses Dokument beschreibt die Detailspezifikation der KI-Module (**AI-Modul**, **Reporting**, **Workflow-Engine** und **Notification**).

---

## Modul 9: AI — Künstliche Intelligenz (Module I1-I17)

### Überblick

Das AI-Modul ist ein zentraler Orchestrator für alle KI-Funktionen in MyRMS. Es unterstützt mehrere KI-Provider gleichzeitig, routet Anfragen an den optimalen Provider und anonymisiert sensible Daten bevor sie an Cloud-APIs gesendet werden.

**Wichtiger Grundsatz:** KI-Funktionen sind immer optional und erfordern eine explizite Aktivierung. Das System funktioniert vollständig ohne KI. Alle KI-Vorschläge sind Empfehlungen, keine automatischen Aktionen (außer wenn explizit als Automatisierung konfiguriert).

### 9.1 Multi-Provider-Architektur (Module I1-I3)

#### Unterstützte Provider

| Provider | Modelle | Eignung |
|---------|---------|---------|
| Anthropic (Claude) | claude-3-5-sonnet, claude-3-haiku | Reasoning, Text, Analyse |
| OpenAI | gpt-4o, gpt-4o-mini | Vision, Code, Chat |
| Google | gemini-1.5-pro, gemini-1.5-flash | Lange Kontexte, Multimodal |
| Mistral | mistral-large, mistral-small | Europäischer Anbieter, DSGVO-freundlicher |
| Ollama (Self-hosted) | llama3, mistral, phi3, etc. | Komplett lokal, keine Datenweitergabe |

#### Provider-Konfiguration (Admin-Panel)

```yaml
ai_providers:
  - provider: anthropic
    api_key: sk-ant-...   # Verschlüsselt gespeichert
    model: claude-3-5-sonnet-20241022
    is_default: true
    tasks:                # Für welche Tasks dieser Provider verwendet wird
      - price_optimization
      - demand_forecast
      - text_generation
      - anonymization

  - provider: openai
    api_key: sk-...
    model: gpt-4o
    tasks:
      - asset_recognition  # Vision-Fähigkeit für Equipment-Fotos
      - ocr_enhancement

  - provider: ollama
    endpoint: http://ollama:11434  # Self-hosted im Docker-Stack
    model: llama3:8b
    tasks:
      - local_classification  # Klassifizierung ohne Datenweitergabe
      - internal_search
```

#### Task-Routing-Logik

```go
func (r *ProviderRouter) SelectProvider(taskType AITaskType, requiresVision bool) AIProvider {
    // 1. Wenn kein Provider konfiguriert: Fehler
    if len(r.providers) == 0 {
        return nil, ErrNoProviderConfigured
    }

    // 2. Wenn Benutzerdaten enthalten sind: Anonymisieren oder nur lokalen Provider verwenden
    if r.containsPersonalData(taskType) && !r.hasLocalProvider() {
        // Anonymisierung erzwingen
        return r.getProvider("anthropic"), nil  // Nach Anonymisierung OK
    }

    // 3. Task-spezifisches Routing
    switch taskType {
    case AssetRecognition:
        return r.getProviderForTask("asset_recognition")  // GPT-4o (Vision)
    case DemandForecast:
        return r.getProviderForTask("demand_forecast")
    default:
        return r.getDefaultProvider()
    }
}
```

#### Usage-Tracking und Kostenkontrolle

```sql
-- Tägliches Kosten-Monitoring (reporting_schema.kpi_snapshots)
SELECT
    DATE(created_at) as date,
    provider,
    COUNT(*) as requests,
    SUM(total_tokens) as tokens,
    SUM(cost_usd) as cost_usd,
    SUM(cost_usd) * 0.93 as cost_eur  -- USD → EUR
FROM ai_schema.ai_requests
WHERE tenant_id = $1
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at), provider
ORDER BY date DESC, cost_usd DESC;
```

Budget-Warnung: Wenn monatliche KI-Kosten > konfigurierbarer Schwellenwert → Benachrichtigung an Admin.

### 9.2 KI-Anonymisierung (Modul I6)

Bevor sensible Kundendaten an Cloud-KI-APIs gesendet werden, anonymisiert der ai-service:

```go
// Vor jedem API-Aufruf mit potenziell sensiblen Daten
func (s *AnonymizationService) Anonymize(ctx context.Context, text string) (string, AnonymizationMap) {
    replacements := AnonymizationMap{}

    // Pattern 1: Kundennamen
    customerNames := s.extractCustomerNames(ctx, text)
    for i, name := range customerNames {
        placeholder := fmt.Sprintf("[KUNDE_%d]", i+1)
        text = strings.ReplaceAll(text, name, placeholder)
        replacements[placeholder] = name
    }

    // Pattern 2: E-Mail-Adressen
    emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
    text = emailRegex.ReplaceAllStringFunc(text, func(email string) string {
        placeholder := fmt.Sprintf("[EMAIL_%d]", len(replacements))
        replacements[placeholder] = email
        return placeholder
    })

    // Pattern 3: Telefonnummern
    phoneRegex := regexp.MustCompile(`(\+49|0)\d{5,15}`)
    // ...

    // Pattern 4: IBAN
    ibanRegex := regexp.MustCompile(`[A-Z]{2}\d{2}[A-Z0-9]{4}\d{7}([A-Z0-9]?){0,16}`)
    // ...

    // Pattern 5: Straßenadressen
    addressRegex := regexp.MustCompile(`[A-Za-zäöüÄÖÜß\-]+ \d+[a-z]?`)
    // ...

    // Pattern 6: Firmennamen (aus Kundendatenbank)
    // ...

    // Pattern 7: Preise / Umsatzdaten
    priceRegex := regexp.MustCompile(`\d+[.,]\d{2}\s*€`)
    // ...

    return text, replacements
}

// Nach dem KI-Aufruf: Anonymisierung rückgängig machen für Anzeige
func (s *AnonymizationService) Deanonymize(text string, mapping AnonymizationMap) string {
    for placeholder, original := range mapping {
        text = strings.ReplaceAll(text, placeholder, original)
    }
    return text
}
```

**DSGVO-Konformität:** Bei Ollama (Self-hosted) ist keine Anonymisierung nötig, da Daten das System nicht verlassen.

### 9.3 Preisoptimierung (Modul I1/I2)

#### Preisoptimierungs-Algorithmus

```
Eingangsdaten für KI:
  - Equipment-Typ und -Kategorie (anonymisiert)
  - Historische Vermietungsdaten (Zeiträume, Auslastung)
  - Saisonale Muster (Festival-Saison: Juni-August)
  - Wettbewerbs-Preise (optional, manuell eingepflegt)
  - Aktuelle Auslastungsrate

KI-Ausgabe:
  {
    "current_price": 100.00,
    "suggested_price": 115.00,
    "confidence": 0.78,
    "reasoning": "Saisonale Hochphase (Aug). Ähnliches Equipment wurde 15% über Basispreis gebucht.",
    "price_range": { "min": 95.00, "max": 130.00 },
    "factors": [
      { "factor": "Saisonalität", "impact": "+12%", "description": "Festival-Saison" },
      { "factor": "Auslastung", "impact": "+5%", "description": "Auslastung 82% (Schwellenwert 75%)" },
      { "factor": "Wettbewerb", "impact": "-2%", "description": "SoundPro bietet ähnliches günstiger an" }
    ]
  }
```

#### Integration in UI

```typescript
// Preisoptimierungs-Widget in der Equipment-Detailseite
function PriceOptimizationWidget({ equipmentId }: { equipmentId: string }) {
  const { data: suggestion, isLoading } = usePriceOptimization(equipmentId);

  return (
    <Card>
      <CardHeader>KI-Preisempfehlung</CardHeader>
      <CardBody>
        {isLoading ? (
          <Skeleton height={100} />
        ) : suggestion ? (
          <>
            <div className={styles.priceComparison}>
              <span>Aktuell: {formatCurrency(suggestion.currentPrice)}</span>
              <ArrowRight size={16} />
              <span className={styles.suggested}>
                Empfehlung: {formatCurrency(suggestion.suggestedPrice)}
                <ConfidenceBadge value={suggestion.confidence} />
              </span>
            </div>
            <p className={styles.reasoning}>{suggestion.reasoning}</p>
            <details>
              <summary>Einflussfaktoren</summary>
              <FactorsList factors={suggestion.factors} />
            </details>
            <div className={styles.actions}>
              <Button variant="outline" size="sm"
                onClick={() => applyPrice(suggestion.suggestedPrice)}>
                Preis übernehmen
              </Button>
              <Button variant="ghost" size="sm"
                onClick={() => dismissSuggestion()}>
                Ablehnen
              </Button>
            </div>
          </>
        ) : (
          <EmptyState message="Nicht genug Daten für Preisempfehlung" />
        )}
      </CardBody>
    </Card>
  );
}
```

### 9.4 Nachfrageprognose (Modul I13)

```
Für welchen Zeitraum wird Equipment wie stark nachgefragt?

Eingangsdaten:
  - Historische Buchungsdaten (2+ Jahre)
  - Saisonale Muster
  - Regionale Veranstaltungskalender (manuell eingepflegt oder via API)
  - Festival-Daten (Open-Air-Saison: Apr-Sep in DACH)

Ausgabe pro Equipment-Typ:
  "Nachfrage-Prognose: L-Acoustics K2 Arrays"
  April 2026:  Erwartete Auslastung 68% (+8% ggü. Vorjahr)
  Mai 2026:    Erwartete Auslastung 85% (Frühlings-Festivals)
  Juni 2026:   Erwartete Auslastung 95% (ACHTUNG: Engpass wahrscheinlich!)
  ...

Empfehlungen:
  - "Für die Festival-Saison 2026: Sub-Rental von 4× K2 bei SoundPro anfragen"
  - "Preis für Juni-August um 15-20% erhöhen (hohe Nachfrage)"
  - "Wartung der K2 bis Ende März planen (vor der Saison)"
```

### 9.5 Smart Asset Creator — Equipment-Erkennung aus Fotos (Modul I9)

```
Workflow:
1. Lisa fotografiert neues Equipment mit dem Handy
2. Foto wird an ai-service (OpenAI GPT-4o Vision) gesendet
3. KI identifiziert:
   - Gerätekategorie (Lautsprecher, Scheinwerfer, Mischpult, ...)
   - Hersteller und Modell (wenn sichtbar)
   - Zustand (neu, gebraucht, Kratzer sichtbar)
   - Technische Merkmale (Anzahl Kanäle, Watts, etc.)

4. Ergebnis wird als Vorschlag angezeigt:
   Erkannt: L-Acoustics ARCS Wide (82% Konfidenz)
   Kategorie: PA-Systeme → Line Arrays → L-Acoustics
   Gewicht: ca. 18 kg (aus Datenbank)
   Leistungsaufnahme: 900W (aus Datenbank)

5. Marco/Lisa prüft und bestätigt → Equipment-Stammdaten werden angelegt
6. Fehlende Felder (Seriennummer, Kaufpreis) manuell ergänzen
```

#### Hersteller-Datenbank-Integration

```go
// Nach KI-Erkennung: Stammdaten aus Herstellerdatenbank laden
func (s *AssetCreatorService) EnrichWithManufacturerData(model string) *ManufacturerData {
    // Interne Datenbank: häufige VT-Equipment-Typen mit Stammdaten
    // (community-gepflegt oder API: Open Library VT-Equipment)
    data, found := s.manufacturerDB.Lookup(model)
    if !found {
        // Fallback: Allgemeine Web-Suche (via KI)
        data = s.aiClient.SearchManufacturerData(model)
    }
    return data
}
```

### 9.6 Wartungsvorhersage (Predictive Maintenance)

```
Basierend auf:
  - Nutzungshäufigkeit (Anzahl Einsätze, Betriebsstunden)
  - Letzte Wartungsprotokolle (welche Mängel wurden gefunden?)
  - Equipment-Typ (Moving Heads: Lampen haben ~2000h Lebensdauer)
  - Fehlerberichte (Scan-basierte Zustandsmeldungen)

KI-Ausgabe:
  "Vorhergesagte Lampenwechsel in 3 Wochen (15.04.):
   - 4× Claypaky Sharpy (Lampe ist > 80% der Lebensdauer)
   - 2× Martin MAC Quantum (Lampenwechsel überfällig)
   Wartungskosten: ca. 400€ (4× 80€ + 2× 40€)
   Empfehlung: Wartungstermin für KW 14 planen (vor Festival-Saison)"
```

### 9.7 KI-Lernsystem und Feedback (Modul I10)

```go
// Jede KI-Empfehlung kann bewertet werden (Few-Shot Learning)
type AIFeedback struct {
    RequestID   uuid.UUID
    TaskType    AITaskType
    UserAction  string       // "accepted", "rejected", "modified"
    UserComment string
    ModifiedTo  *float64     // Bei Preisänderung: was wurde tatsächlich gewählt
}

// Feedback wird gesammelt und in zukünftige Prompts eingebettet
func (s *LearningService) GetFewShotExamples(ctx context.Context, taskType AITaskType) []Example {
    // Die letzten 10 akzeptierten Empfehlungen als Beispiele
    feedback, _ := s.feedbackRepo.GetAccepted(ctx, taskType, 10)
    return convertToExamples(feedback)
}
```

**A/B-Testing:** Für dasselbe Equipment können zwei verschiedene KI-Vorschläge parallel getestet werden. Nach 4 Wochen wird ausgewertet welcher Vorschlag zu mehr Buchungen/besserem Ergebnis führte.

---

## Modul 10: Reporting — Dashboards, KPIs, Business Intelligence

### Überblick

Das Reporting-Modul ist die CQRS Read-Side für Geschäftsdaten. Es subscribt auf alle KurrentDB-Events und pflegt eigene Materialized Views für schnelle Dashboard-Abfragen.

### 10.1 Dashboard-KPIs

#### KPIs nach Rolle

**Geschäftsführer (Marco) — 30-Sekunden-Überblick:**
```
HEUTE / DIESE WOCHE:
  Aktive Projekte: 4    (2× im Aufbau, 1× aktiv, 1× im Abbau)
  Equipment draußen: 89 Artikel (von 312 gesamt = 28,5% Auslastung)
  Offene Rechnungen: 47.832,50€  (davon 5.220€ überfällig)
  Geplante Projekte (7 Tage): 3 Projekte, ~28.000€ Umsatz

WARNUNGEN:
  🔴 2 Doppelbuchungen — Equipment überschneidet sich
  🟡 E-Check: 3 Geräte überfällig
  🟡 Wartung: 5 Geräte überfällig
```

**Lagerverwaltung (Lisa) — Lager-Übersicht:**
```
IM LAGER VERFÜGBAR: 223 Artikel
HEUTE RAUSGEHEND: 45 Artikel (Projekt: Festival Stadtpark)
HEUTE KOMMEND: 12 Artikel (Rückgabe: Corporate Event)
DEFEKT: 8 Artikel (zu reparieren)
WARTUNG: 3 Artikel (in Reparatur)
```

**Buchhaltung (Thomas) — Finanz-Überblick:**
```
MONAT MÄRZ 2026:
  Umsatz (fakturiert): 34.500€
  Umsatz (offen): 18.200€
  Eingangsrechnungen: 6.800€
  Ergebnis: ~27.700€ (vor Fixkosten)

OFFENE POSTEN:
  Gesamt: 47.832,50€
  Davon überfällig: 5.220€
  Mahnungen diese Woche: 3
```

### 10.2 Reports

#### Umsatz-Report

```
Umsatzanalyse April 2025 — März 2026

Nach Monat:
  Apr: 42.100€  Mai: 38.500€  Jun: 52.000€  Jul: 67.500€  Aug: 71.200€
  Sep: 55.300€  Okt: 48.800€  Nov: 39.200€  Dez: 44.100€  Jan: 31.500€
  Feb: 35.800€  Mär: 34.500€

Nach Kategorie:
  Ton:    38%  (210.000€)
  Licht:  31%  (171.000€)
  Video:  18%  (99.000€)
  Rigging: 8%  (44.000€)
  Sonstiges: 5% (28.000€)

Nach Kunde (Top 5):
  1. BMW AG                    45.200€  (8,2%)
  2. Eventmanagement XY GmbH   38.900€  (7,1%)
  3. Stadt Musterstadt         32.100€  (5,8%)
  4. ...

Top-Equipment:
  1. d&b audiotechnik J-Series (48 Einsätze, 89.200€ Umsatz)
  2. L-Acoustics K2 (36 Einsätze, 67.400€)
  3. ...
```

#### Equipment-Auslastungs-Report

```
AUSLASTUNGS-ANALYSE — März 2026

Gesamtauslastung: 34,2% (von 312 Artikeln: 0 Stk. = 0 Tage vermietet)

Auslastung nach Kategorie:
  PA-Systeme:   62,1%  █████████████████░░░░░  (sehr hoch)
  Moving Heads: 38,4%  ████████████░░░░░░░░░░░
  Mischpulte:   71,3%  ██████████████████████░  (kritisch: fast immer ausgebucht)
  Kabelsets:    22,1%  ███████░░░░░░░░░░░░░░░░  (Überbestand?)

Unterausgelastetes Equipment (< 10% im letzten Quartal):
  - 12× PAR64 Conventional    2,3%  → Verkauf oder Abgang prüfen?
  - 4× Nebuliser Tiny          3,1%  → Preis senken?
```

#### Crew-Stunden-Report (für Lohnabrechnung)

```
PERSONALKOSTEN — März 2026

Kevin Radic (Freelancer):
  Projekt Festival Stadtpark:    2 Tage × 300€ = 600€
  Projekt Corporate Event BMW:   1 Tag × 280€  = 280€
  Projekt Stadtfest:             1 Tag × 300€  = 300€
  GESAMT:                                        1.180€

Max Müller (Festangestellt):
  Gesamtstunden: 168h (Mo-Fr)
  → Stundenlohn laut Arbeitsvertrag
  Überstunden: 12h (werden als Freizeit ausgeglichen)
```

### 10.3 Nachhaltigkeits-Reporting / ESG (Modul L5)

```
CO2-TRACKING — 2025

Gesamtemissionen: 12,4 t CO₂e

Aufschlüsselung:
  Transport (Fahrten):      7,2 t CO₂e  (58%)
    - 28.000 km × 18L/100km × 2,63kg CO₂/L = 7,1t
  Energieverbrauch Lager:   2,8 t CO₂e  (23%)
    - 14.500 kWh × 0,192 kg CO₂/kWh = 2,8t
  Entsorgung Equipment:     1,4 t CO₂e  (11%)
  Sonstiges:                0,9 t CO₂e  (8%)

ZIELE FÜR 2026:
  - CO₂-Reduktion 10%: Elektro-Transporter (2× Sprinter E) → -2,4t
  - Ökostrom für Lager: -2,8t (falls Anbieterwechsel)
  - Reparieren statt Ersetzen: Ziel 90% Reparaturquote

ESG-Bericht (für Kunden-Anforderungen):
  - Exportierbar als PDF (für Kundenanforderungen)
```

### 10.4 Anomalie-Erkennung (Modul I15)

```go
// reporting-service nutzt ai-service für Anomalie-Erkennung
func (s *ReportingService) DetectAnomalies(ctx context.Context) []Anomaly {
    anomalies := []Anomaly{}

    // Anomalie 1: Umsatzeinbruch
    lastMonthRevenue := s.getMonthlyRevenue(ctx, -1)
    avgRevenue := s.getAverageMonthlyRevenue(ctx, 6)
    if lastMonthRevenue < avgRevenue*0.7 {
        anomalies = append(anomalies, Anomaly{
            Type:     "revenue_drop",
            Severity: "warning",
            Message:  fmt.Sprintf("Umsatz letzten Monat 30%% unter Durchschnitt (%.0f€ vs %.0f€)", lastMonthRevenue, avgRevenue),
        })
    }

    // Anomalie 2: Ungewöhnlich viele Schäden
    recentDamages := s.countDamageReports(ctx, 30)
    avgDamages := s.getAverageDamageReports(ctx, 6)
    if float64(recentDamages) > avgDamages*2 {
        anomalies = append(anomalies, Anomaly{
            Type:     "damage_spike",
            Severity: "high",
            Message:  "Schadenshäufigkeit 2× über Durchschnitt — Handling-Prozesse prüfen",
        })
    }

    // Weitere Anomalien: Zahlungsausfälle, Inventardifferenzen, E-Check-Verstöße...

    return anomalies
}
```

---

## Modul 11: Workflow — Automatisierungen und Genehmigungsprozesse (Modul K1)

### Überblick

Die Workflow-Engine ermöglicht No-Code-Automatisierungen nach dem WENN-DANN-Prinzip. Zielgruppe: Marco (Geschäftsführer) und Admin.

### 11.1 Workflow-Konzept

```
WORKFLOW: "Rechnung bei Projektabschluss"

TRIGGER: Event "project.completed" empfangen

BEDINGUNGEN:
  - Projektstatus = "completed"
  - Projekt hat gültiges Angebot

AKTIONEN:
  1. [Warte] 1 Stunde (für eventuelle Nachkorrekturen)
  2. [Erstelle Aufgabe] "Rechnung erstellen" → Thomas zuweisen
  3. [Benachrichtige] Thomas: "Projekt abgeschlossen — Rechnung erstellen"
  4. [Benachrichtige] Marco: "Abschluss bestätigt: {projekt_name}"

OPTIONALE AKTION: (wenn "Automatisch erstellen" aktiviert)
  2b. [Erstelle Rechnung automatisch] aus Angebot → Status: "draft"
  3b. [Benachrichtige] Thomas: "Entwurf erstellt — bitte prüfen"
```

### 11.2 Verfügbare Trigger

| Trigger-Typ | Beschreibung | Beispiel-Events |
|------------|-------------|----------------|
| **KurrentDB Event** | Domain-Event empfangen | `project.completed`, `equipment.checked_in`, `invoice.overdue` |
| **Zeitplan (CRON)** | Regelmäßig ausführen | Täglich 08:00, Montags, Monatlich |
| **Manuell** | User klickt auf "Ausführen" | Ad-hoc Workflows |
| **Webhook** | Externer HTTP-Aufruf | Integration von außen |

### 11.3 Verfügbare Aktionen

#### Kommunikation

| Aktion | Parameter |
|--------|---------|
| `send_email` | Empfänger, Betreff, Template, Variablen |
| `send_notification` | User/Rolle, Titel, Text, Link |
| `send_sms` | Telefonnummer, Text (via SMPP/SMS-Provider) |
| `send_webhook` | URL, Method, Headers, Body-Template |

#### Daten-Aktionen

| Aktion | Parameter |
|--------|---------|
| `create_task` | Titel, Beschreibung, Fälligkeitsdatum, Zuweisung |
| `create_invoice` | aus Projekt, Status (draft/send) |
| `update_status` | Entitätstyp, Status |
| `create_maintenance_task` | Equipment-ID, Aufgabe, Priorität |

#### Fluss-Kontrolle

| Aktion | Parameter |
|--------|---------|
| `wait` | Dauer (Minuten, Stunden, Tage) |
| `wait_for_approval` | Genehmiger (User/Rolle), Frist |
| `condition` | Wenn-Dann-Verzweigung im Workflow |

### 11.4 Vordefinierte Workflow-Vorlagen

Das System liefert fertige Workflow-Vorlagen mit, die der Admin aktivieren kann:

```
Vorlage 1: "Automatische Rechnung bei Projektabschluss"
  Trigger: project.completed
  Aktion: Rechnungs-Entwurf erstellen + Thomas benachrichtigen

Vorlage 2: "Mahnlauf täglich"
  Trigger: CRON 08:00 täglich
  Aktion: Offene Posten prüfen, fällige Mahnungen versenden

Vorlage 3: "E-Check-Erinnerung"
  Trigger: CRON 07:00 täglich
  Bedingung: E-Check fällig in <= 30 Tagen
  Aktion: Wartungsaufgabe erstellen + Admin benachrichtigen

Vorlage 4: "Doppelbuchungs-Alert"
  Trigger: availability.conflict_detected
  Aktion: Projektleiter sofort benachrichtigen + Dashboard-Alert

Vorlage 5: "Willkommen neuer Freelancer"
  Trigger: invitation.accepted (Rolle = Freelancer)
  Aktion: E-Mail mit Onboarding-Informationen senden

Vorlage 6: "Versicherungspolizze läuft ab"
  Trigger: CRON monatlich
  Bedingung: Polizze läuft in <= 60 Tagen ab
  Aktion: Admin per E-Mail erinnern

Vorlage 7: "Equipment-Rückgabe überfällig"
  Trigger: CRON täglich
  Bedingung: Check-Out vor > 7 Tagen, kein Check-In
  Aktion: Projektleiter benachrichtigen
```

### 11.5 Genehmigungsprozesse

```go
// Workflow mit Genehmigungsschritt
type ApprovalStep struct {
    Title        string
    Description  string
    Approvers    []Approver  // User-IDs oder Rollen
    Deadline     time.Duration
    EscalateTo   []Approver  // Falls keine Antwort nach Frist
    OnApprove    []Action
    OnReject     []Action
}

// Beispiel: Großauftrag > 10.000€ braucht Geschäftsführer-Freigabe
Workflow: "Rechnungsfreigabe"
  TRIGGER: invoice.created (wenn invoice.total > 10000)
  AKTION: wait_for_approval(
    approver: role="admin",
    deadline: 2 Tage,
    on_approve: invoice.send(),
    on_reject: create_task("Rechnung anpassen", assignee=creator)
  )
```

### 11.6 Workflow-Editor (Frontend)

```typescript
// Visueller No-Code Editor im Browser
// Drag & Drop mit React Flow

function WorkflowEditor({ workflowId }: { workflowId?: string }) {
  const [nodes, setNodes] = useState<WorkflowNode[]>([]);
  const [edges, setEdges] = useState<WorkflowEdge[]>([]);

  // Jeder Node repräsentiert einen Trigger, eine Bedingung oder eine Aktion
  const nodeTypes = {
    trigger: TriggerNode,      // Auslöser (Event, Zeitplan)
    condition: ConditionNode,  // WENN-Bedingung
    action: ActionNode,        // Dann-Aktion
    wait: WaitNode,            // Warte-Schritt
    approval: ApprovalNode,    // Genehmigungsschritt
  };

  return (
    <div className={styles.editor}>
      <NodeToolbar>
        {/* Drag-and-Drop aus Sidebar */}
        <DraggableNode type="trigger" icon={<Zap />} label="Trigger" />
        <DraggableNode type="condition" icon={<GitBranch />} label="Bedingung" />
        <DraggableNode type="action" icon={<Play />} label="Aktion" />
        <DraggableNode type="wait" icon={<Clock />} label="Warten" />
      </NodeToolbar>

      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onConnect={onConnect}
        fitView
      >
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>

      <div className={styles.actions}>
        <Button variant="outline" onClick={() => testWorkflow()}>
          Testen
        </Button>
        <Button onClick={() => saveAndActivate()}>
          Speichern & Aktivieren
        </Button>
      </div>
    </div>
  );
}
```

---

## Modul 12: Notification — Benachrichtigungs-Center (Modul N7)

### Überblick

Das Notification-Modul verteilt Benachrichtigungen über alle Kanäle (In-App, E-Mail, Push, SMS) und lässt jeden User seine Präferenzen selbst konfigurieren.

### 12.1 Benachrichtigungs-Typen

| Typ | Standard-Kanal | Beschreibung |
|-----|---------------|-------------|
| `project.conflict` | In-App + E-Mail | Doppelbuchung erkannt |
| `project.confirmed` | In-App | Projekt von Kunde bestätigt |
| `project.status_changed` | In-App | Statuswechsel |
| `invoice.overdue` | E-Mail + In-App | Rechnung überfällig |
| `invoice.paid` | In-App | Zahlung eingegangen |
| `equipment.defect_reported` | In-App | Defekt gemeldet |
| `equipment.checkin_complete` | In-App | Alle Artikel zurück |
| `maintenance.overdue` | E-Mail + In-App | Wartung überfällig |
| `echeck.overdue` | E-Mail + In-App | E-Check überfällig |
| `crew.assignment_new` | Push + E-Mail | Neue Job-Zuweisung |
| `crew.assignment_changed` | Push | Job-Änderung |
| `federation.request_received` | In-App + E-Mail | Sub-Rental-Anfrage |
| `insurance.expiring` | E-Mail | Police läuft ab |
| `system.update_available` | In-App | Software-Update verfügbar |
| `ai.suggestion_ready` | In-App | KI-Empfehlung |

### 12.2 Kanal-Konfiguration

#### E-Mail (SMTP)

```go
// E-Mail-Versand via SMTP (konfigurierbar in .env)
type EmailConfig struct {
    Host     string  // smtp.example.com
    Port     int     // 587 (STARTTLS)
    Username string
    Password string
    From     string  // noreply@myrms.example.com
    FromName string  // MB Veranstaltungstechnik
}

func (s *EmailService) Send(ctx context.Context, to, subject, templateName string, data interface{}) error {
    // Template rendern
    html, _ := s.renderEmailTemplate(templateName, data)
    text, _ := s.renderTextTemplate(templateName, data)

    msg := gomail.NewMessage()
    msg.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From))
    msg.SetHeader("To", to)
    msg.SetHeader("Subject", subject)
    msg.SetBody("text/plain", text)
    msg.AddAlternative("text/html", html)

    // Retry bei Fehler (3× mit exponentialem Backoff)
    return retry.Do(func() error {
        return s.dialer.DialAndSend(msg)
    }, retry.Attempts(3), retry.Delay(5*time.Second))
}
```

#### Push-Notifications (Web Push)

```go
// Web Push API (für PWA auf Desktop und Mobile)
type PushService struct {
    vapidPublicKey  string
    vapidPrivateKey string
}

func (s *PushService) Send(ctx context.Context, userID uuid.UUID, notification PushPayload) {
    // Alle Push-Tokens des Users laden
    tokens, _ := s.tokenRepo.GetByUserID(ctx, userID)

    for _, token := range tokens {
        payload, _ := json.Marshal(notification)
        sub := webpush.Subscription{
            Endpoint: token.Endpoint,
            Keys:     webpush.Keys{P256dh: token.P256dh, Auth: token.Auth},
        }
        webpush.SendNotification(payload, &sub, &webpush.Options{
            VAPIDPublicKey:  s.vapidPublicKey,
            VAPIDPrivateKey: s.vapidPrivateKey,
        })
    }
}
```

### 12.3 User-Präferenzen

Jeder User kann pro Benachrichtigungs-Typ konfigurieren:
- In-App: Ja / Nein (Standard: Ja)
- E-Mail: Ja / Nein / Täglich-Digest (Standard: Ja für wichtige)
- Push: Ja / Nein (Standard: Nein, muss explizit aktiviert werden)
- Ruhezeiten: "Keine Benachrichtigungen zwischen 22:00 und 07:00"

```typescript
// Präferenzen-UI
function NotificationPreferences() {
  const { data: prefs } = useNotificationPreferences();
  const updatePrefs = useUpdateNotificationPreferences();

  return (
    <div>
      <h2>Benachrichtigungs-Einstellungen</h2>
      {notificationTypes.map(type => (
        <div key={type.id} className={styles.row}>
          <div className={styles.info}>
            <span className={styles.name}>{type.name}</span>
            <span className={styles.description}>{type.description}</span>
          </div>
          <div className={styles.channels}>
            <ChannelToggle
              label="In-App"
              checked={prefs?.[type.id]?.inApp ?? true}
              onChange={(v) => updatePrefs.mutate({ type: type.id, inApp: v })}
            />
            <ChannelToggle
              label="E-Mail"
              checked={prefs?.[type.id]?.email ?? false}
              onChange={(v) => updatePrefs.mutate({ type: type.id, email: v })}
            />
            <ChannelToggle
              label="Push"
              checked={prefs?.[type.id]?.push ?? false}
              onChange={(v) => updatePrefs.mutate({ type: type.id, push: v })}
            />
          </div>
        </div>
      ))}

      <Separator />
      <FormField label="Ruhezeiten">
        <div className={styles.quietHours}>
          <TimeInput label="Von" name="quietFrom" />
          <span>bis</span>
          <TimeInput label="Bis" name="quietUntil" />
        </div>
      </FormField>
    </div>
  );
}
```

### 12.4 Notification-Center (Frontend)

```typescript
// In-App Benachrichtigungs-Center (Glocken-Icon in TopBar)
function NotificationCenter() {
  const { notifications, unreadCount, markAsRead, markAllAsRead } = useNotifications();
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className={styles.trigger}>
          <Bell size={20} />
          {unreadCount > 0 && (
            <span className={styles.badge}>{unreadCount > 99 ? '99+' : unreadCount}</span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className={styles.content}>
        <div className={styles.header}>
          <h3>Benachrichtigungen</h3>
          {unreadCount > 0 && (
            <Button variant="ghost" size="sm" onClick={markAllAsRead}>
              Alle als gelesen
            </Button>
          )}
        </div>

        <ScrollArea className={styles.list}>
          {notifications.length === 0 ? (
            <EmptyState message="Keine Benachrichtigungen" />
          ) : (
            notifications.map(n => (
              <NotificationItem
                key={n.id}
                notification={n}
                onClick={() => {
                  markAsRead(n.id);
                  if (n.referenceType) navigate(getLink(n));
                  setOpen(false);
                }}
              />
            ))
          )}
        </ScrollArea>

        <div className={styles.footer}>
          <Link to="/settings/notifications">Einstellungen</Link>
        </div>
      </PopoverContent>
    </Popover>
  );
}
```

### 12.5 Tages-Zusammenfassung (Daily Digest E-Mail)

Für User die E-Mail-Benachrichtigungen als "Daily Digest" konfiguriert haben, werden alle Notifications des Tages in einer E-Mail zusammengefasst:

```
SUBJECT: MyRMS Tagesübersicht — 20. März 2026

Guten Morgen, Marco!

HEUTE ZU BEACHTEN:
  🔴 2 Doppelbuchungen — Sofort handeln!
  🟡 Rechnung RE-2026-032 seit 32 Tagen überfällig (3.420€)
  🟡 E-Check: 3 Geräte in 5 Tagen fällig

NEUE NACHRICHTEN:
  ✓ Festival Stadtpark — Projekt als abgeschlossen markiert
  ✓ Zahlung eingegangen: Corporate BMW 14.875€

ANSTEHEND DIESE WOCHE:
  [Mo] Festival Stadtpark — Aufbau ab 07:00
  [Mi] Corporate Webinar — Equipment ab 08:00
  ...

[Zum Dashboard] [Doppelbuchungen lösen] [Einstellungen]
```
