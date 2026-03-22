# RentFlow — Implementierungsreife Detailspezifikation
## Expense Tracking Service (Ausgabenverwaltung) & NAS Storage Architecture

**Stand:** 21. März 2026
**Version:** 1.0 — Production Ready
**Autoren:** Finanz-/Compliance Engineer, Backend-Architekt, Storage Systems Specialist

---

## INHALTSVERZEICHNIS

1. [Übersicht & Anforderungen](#1-übersicht--anforderungen)
2. [Section 1: Expense Tracking Service](#section-1-expense-tracking-service)
3. [Section 2: NAS Storage Architecture](#section-2-nas-storage-architecture)
4. [Cross-Service Integration](#cross-service-integration)
5. [Sicherheit & Compliance](#5-sicherheit--compliance)
6. [Deployment & Betrieb](#6-deployment--betrieb)

---

## 1. ÜBERSICHT & ANFORDERUNGEN

### 1.1 Geschäftsanforderungen

**Nutzer-Request (Original):**
> *"Ich brauche noch die Möglichkeit auch Ausgaben in der Software festzuhalten, also über das Handy Fotos von Kassenzetteln zu machen und diese dann auszulesen oder von Rechnungen und diese dann in die Ausgaben automatisch aufzunehmen, damit die dann zum Schluss auch mit zur Steuer exportiert werden."*

**Scope:**
- Mobile PWA für Receipt-Fotografie (Kassenbons, Rechnungen)
- Automatische OCR-Verarbeitung via ai-service (Claude/OpenAI/Gemini/Mistral/Ollama)
- Intelligente Datenextraktion (Vendor, Datum, Betrag, USt., Positionen, Zahlmittel)
- Automatische Kategorisierung (SKR03/SKR04)
- Genehmigungsworkflow mit konfigurierbaren Schwellwerten
- Steuer-Export (DATEV, WISO, CSV) für GoBD-konforme Archivierung
- NAS-basierte Speicherung aller Dateien (Bilder, PDFs, Dokumente)
- Rechtssichere Archivierung nach deutschem Steuerrecht (AO §147, GoBD, DSGVO)

### 1.2 Nicht im Scope

- Eigenständige Geschäftskonto-Integration (wird via invoice-service verwaltet)
- Sepa-Payment-Zahlungen (Pflicht: manuell oder über Fremd-Integration)
- Kategorisierung nach Branchenstandard-ESCO-Codes

### 1.3 Architektur-Entscheidungen

| Entscheidung | Wert | Begründung |
|-------------|------|-----------|
| **Service Nr.** | 18 (expense-service) | Nächste verfügbare Nummer |
| **Port** | :8018 | Fortlaufend nach reporter-service (:8017 = audit-service) |
| **Event Store** | KurrentDB | Immutable Audit Trail für Steuerkompliance |
| **Read-DB** | PostgreSQL (expense_schema) | Schnelle Abfragen, Reporting-Integration |
| **Speicher Backend** | NAS (Synology SMB/NFS) | Lokal gehostet, vollständige Datenkontrolle |
| **OCR Provider** | ai-service (Multi-Provider) | Kostenoptimierung, Fallback-Strategie |
| **Genehmigung** | Konfig-basiert | Auto-Approve < Schwelle, Manual > Schwelle |
| **Retention** | 10 Jahre (AO §147) | In KurrentDB immutable, Löschung nur mit Audit-Trail |
| **Export-Formate** | DATEV, WISO CSV, PDF-Bericht | Häufigste Accounting-Software in DE |

---

# SECTION 1: EXPENSE TRACKING SERVICE

## 2. FUNKTIONALE ANFORDERUNGEN

### 2.1 Kernfunktionalitäten

#### 2.1.1 Receipt Photo Capture (Mobile PWA)

**Anwendungsfall:** Nutzer fotografiert Kassenbon/Rechnung mit Smartphone

**Workflow:**
```
Nutzer öffnet PWA
  ↓
Camera Permission Request
  ↓
Live Preview (optional Crop)
  ↓
Capture Photo (JPEG, PNG bis 10 MB)
  ↓
Metadata erfassen (Datum, Vendor optional)
  ↓
Upload (Queue wenn offline)
  ↓
OCR triggern
  ↓
Ergebnis Review (UI)
```

**Frontend-Komponenten (React/TypeScript):**
```typescript
// src/components/ExpenseCapture/ReceiptCamera.tsx
interface ReceiptCameraProps {
  onCapture: (photo: Blob, metadata: PhotoMetadata) => void;
  maxFileSize?: number; // default: 10MB
  supportedFormats?: string[]; // default: ['image/jpeg', 'image/png', 'application/pdf']
}

interface PhotoMetadata {
  timestamp: ISO8601;
  gpsCoordinates?: { lat: number; lng: number };
  deviceType: 'mobile' | 'tablet' | 'desktop';
  offlineMode: boolean;
}

// src/components/ExpenseCapture/ManualCorrection.tsx
interface ManualCorrectionProps {
  ocrResult: OCRResult;
  onSave: (corrected: ExpenseData) => void;
  categories: ExpenseCategory[];
}

// src/components/ExpenseList/ExpenseList.tsx
interface ExpenseListProps {
  filters: ExpenseFilter;
  onApprove?: (id: string) => void;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
}
```

**Offline-Queueing (Service Worker):**
```typescript
// src/workers/expenseQueue.worker.ts
type QueuedExpense = {
  id: string; // temp UUID
  photo: Blob;
  metadata: PhotoMetadata;
  queuedAt: ISO8601;
  retryCount: number;
  lastError?: string;
};

// Bei Online-Rückkehr: sync()
// Exponential Backoff: 1s → 2s → 4s → 8s (max 4 Retries)
```

#### 2.1.2 OCR Processing & Data Extraction

**Trigger-Punkt:** Nach Receipt-Upload (async Job)

**OCR-Pipeline:**

```
Receipt Photo
  ↓
[AI-Service Multi-Provider Call]
  ├─ Provider 1: Claude (gpt-4-vision)
  │  └─ Fallback wenn Rate-Limit
  ├─ Provider 2: OpenAI (vision)
  │  └─ Fallback wenn Claude ausfällt
  ├─ Provider 3: Google Gemini
  │  └─ Fallback für Lokalisierung
  └─ Provider 4: Ollama (local, offline)
      └─ Backup für kritische Szenarien
  ↓
[Structured Output Extraction]
  ├─ Vendor Name (fuzzy match gegen Vendor-DB)
  ├─ Transaction Date (ISO8601)
  ├─ Total Amount (EUR, mit Dezimaltrennern)
  ├─ Tax Amount (USt. in %)
  ├─ Line Items (optional)
  ├─ Payment Method (Cash/Card/Check/Transfer)
  └─ Currency (default EUR, ggf. Conversion)
  ↓
[Confidence Scoring & Validation]
  ├─ Vendor Confidence: 0.0 - 1.0
  ├─ Amount Confidence: 0.0 - 1.0
  └─ Flag für Manual Review wenn < 0.75
  ↓
Extracted Data → Manual Correction UI
```

**ai-service Integration (HTTP):**

```bash
POST /api/v1/process-document
Content-Type: application/json

{
  "source": "expense-service:ocr-receipt",
  "document": {
    "type": "receipt",
    "format": "image/jpeg",
    "data_base64": "...",
    "filename": "kassenbon-2026-03-21.jpg"
  },
  "extraction_schema": {
    "vendor_name": "string",
    "transaction_date": "ISO8601",
    "total_amount": "decimal(10,2)",
    "currency": "string (EUR|USD|GBP|CHF)",
    "tax_amount": "decimal(10,2)",
    "tax_rate": "decimal(4,2)",
    "line_items": [
      {
        "description": "string",
        "amount": "decimal(10,2)",
        "quantity": "integer"
      }
    ],
    "payment_method": "enum(cash|card|check|transfer)",
    "iban_bic": "object (optional)"
  },
  "processing_instructions": {
    "language": "de",
    "confidence_threshold": 0.75,
    "anonymize_personal_data": true
  },
  "ai_provider_priority": ["claude", "openai", "gemini", "ollama"]
}

HTTP/1.1 200 OK
Content-Type: application/json

{
  "request_id": "uuid",
  "status": "completed",
  "extracted_data": {
    "vendor_name": {
      "value": "REWE Markt GmbH",
      "confidence": 0.98,
      "vendor_id": "vendor-12345" // optional: Falls bekannt
    },
    "transaction_date": {
      "value": "2026-03-20T14:30:00Z",
      "confidence": 0.95
    },
    "total_amount": {
      "value": 47.99,
      "confidence": 0.99
    },
    "currency": {
      "value": "EUR",
      "confidence": 1.0
    },
    "tax_amount": {
      "value": 7.66,
      "confidence": 0.92
    },
    "tax_rate": {
      "value": 19.0,
      "confidence": 0.90
    },
    "line_items": [
      {
        "description": "Kaffeebohnen (1kg)",
        "amount": 12.99,
        "quantity": 1,
        "confidence": 0.87
      },
      {
        "description": "Notizblöcke (5er Pack)",
        "amount": 4.99,
        "quantity": 1,
        "confidence": 0.85
      }
    ],
    "payment_method": {
      "value": "cash",
      "confidence": 0.80
    }
  },
  "ai_provider_used": "claude",
  "processing_duration_ms": 2345,
  "flags_for_review": [
    "line_items_confidence_below_threshold",
    "payment_method_unclear"
  ]
}
```

**Expense-Service → OCR-Completion Event:**

```go
type OCRCompletedEvent struct {
  ReceiptID string
  ExtractedData OCRExtraction
  ConfidenceScores map[string]float64 // vendor_name, amount, date, etc.
  RequiresManualReview bool
  ProcessedAt time.Time
  AIProvider string // "claude", "openai", "gemini", "ollama"
}
```

#### 2.1.3 Automatische Kategorisierung

**Standard-Kategorien (SKR03/SKR04 kompatibel):**

```sql
-- PostgreSQL: expense_schema.categories
CREATE TABLE IF NOT EXISTS categories (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  accounting_code VARCHAR(10) NOT NULL, -- "4302", "6360", etc.
  name_de VARCHAR(255) NOT NULL,
  name_en VARCHAR(255),
  description TEXT,
  gkt_code VARCHAR(20), -- Gemeinsamer Klassifikationskatalog (optional)
  skr04_mapping VARCHAR(20), -- SKR04 Kontonummer
  tax_category ENUM('full_deductible', 'partial_deductible', 'non_deductible'),
  tax_rate_default DECIMAL(4,2), -- 0, 7, 19 (%)
  requires_approval BOOLEAN DEFAULT false,
  approval_threshold_cents BIGINT, -- NULL = keine Schwelle
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  UNIQUE(tenant_id, accounting_code),
  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id)
);

-- Beispiel-Daten:
INSERT INTO categories VALUES
  (uuid_generate_v4(), tenant1, '4320', 'Büromaterial', 'Office Supplies', '...', NULL, '4320', 'full_deductible', 19.00, false, NULL, true, now(), now()),
  (uuid_generate_v4(), tenant1, '6360', 'Reinigungskosten', 'Cleaning Services', '...', NULL, '6360', 'full_deductible', 19.00, false, 50000, true, now(), now()),
  (uuid_generate_v4(), tenant1, '4370', 'Fassaden/Außenanlagen', '...', '...', NULL, '4370', 'full_deductible', 19.00, false, NULL, true, now(), now());
```

**AI-basierte Kategorisierung (im expense-service):**

```go
type CategorizationEngine struct {
  categoryCache map[string]Category // Lookup
  vendorCache   map[string]VendorProfile // Historical
  ml interface{} // Optional: ML-Model für Vorhersagen
}

func (e *CategorizationEngine) Categorize(ctx context.Context, expense OCRExtraction) (*Category, float64, error) {
  // 1. Vendor-History prüfen (häufigste Kategorie)
  if profile, ok := e.vendorCache[expense.VendorName]; ok {
    return profile.DefaultCategory, 0.95, nil
  }

  // 2. Keyword-Matching gegen Category-Namen
  matches := e.keywordMatch(expense.VendorName, expense.LineItems)
  if len(matches) > 0 {
    return matches[0].Category, matches[0].Confidence, nil
  }

  // 3. ML-Modell (optional, trainiert auf historischen Daten)
  // if e.ml != nil {
  //   return e.ml.Predict(expense)
  // }

  // 4. Default: Manual Review erforderlich
  return nil, 0.0, ErrNoCategoryMatch
}
```

**Kategorisierungs-Event:**

```go
type ExpenseCategorizedEvent struct {
  ExpenseID string
  Category Category
  Confidence float64 // 0.0 - 1.0
  Method string // "vendor_history", "keyword_match", "ml_model", "manual"
  CategorizedAt time.Time
}
```

#### 2.1.4 Manual Correction UI

**Workflow:**

```
OCR Result präsentiert
  ↓
Nutzer korrigiert Felder (falls nötig)
  ├─ Vendor Name
  ├─ Datum
  ├─ Betrag
  ├─ USt. (%/Betrag)
  ├─ Zahlmittel
  ├─ Kategorisierung
  └─ Beschreibung/Notizen
  ↓
Speichern → Event: ExpenseCorrectedByUser
  ↓
Genehmigungslogik prüfen
```

**React Component (TypeScript):**

```typescript
interface ManualCorrectionUIProps {
  expense: Expense;
  ocrExtraction: OCRExtraction;
  categories: Category[];
  onSave: (corrected: Expense) => Promise<void>;
  readOnly?: boolean;
}

interface EditableCorrectionFields {
  vendorName: string;
  transactionDate: ISO8601;
  amount: number;
  taxRate: number;
  taxAmount: number;
  paymentMethod: PaymentMethod;
  category: Category;
  description?: string;
  lineItems?: LineItem[];
}

// Inline Validation:
// - Amount > 0
// - Date <= today
// - Category != null (oder setze auf Genehmigung)
// - Tax Rate in [0, 7, 19] (de-spezifisch)
```

#### 2.1.5 Recurring Expenses

**Usecase:** Monatliche Versicherung, Miete, Leasing, SaaS-Abos

**Data Model:**

```go
type RecurringExpense struct {
  ID string
  TenantID string

  Name string // "Haftpflichtversicherung", "Büro-Internet"
  Category Category
  Amount decimal.Decimal
  Currency string

  // Recurrence
  Frequency enum // MONTHLY, QUARTERLY, ANNUALLY, CUSTOM
  StartDate time.Time
  EndDate *time.Time // nil = unbegrenzt

  // Auto-Generation
  AutoGenerateExpense bool // true = alle X Tage/Monate automatisch erzeugen
  NextGenerationDate time.Time

  // Genehmigung
  ApprovalStatus enum // PENDING, APPROVED, REJECTED

  // Audit
  CreatedAt time.Time
  UpdatedAt time.Time
  CreatedBy string
}

// KurrentDB Event:
type RecurringExpenseCreatedEvent struct {
  RecurringExpenseID string
  TenantID string
  Name string
  Amount decimal.Decimal
  Frequency string
  Category Category
  StartDate time.Time
  CreatedAt time.Time
}

type RecurringExpenseGeneratedEvent struct {
  RecurringExpenseID string
  GeneratedExpenseID string
  Amount decimal.Decimal
  DueDate time.Time
  GeneratedAt time.Time
}
```

**Scheduled Job (via workflow-service oder intern):**

```go
// Täglich um 02:00 Uhr (KurrentDB Subscriber)
func (s *ExpenseService) ProcessRecurringExpenses(ctx context.Context) error {
  expenses, _ := s.repo.GetRecurringExpensesMaturityTomorrow(ctx)

  for _, rec := range expenses {
    expense := &Expense{
      ID: uuid.New().String(),
      TenantID: rec.TenantID,
      Category: rec.Category,
      Amount: rec.Amount,
      Currency: rec.Currency,
      VendorName: rec.Name,
      TransactionDate: time.Now(),
      PaymentMethod: "unknown",
      Status: PENDING, // oder AUTO_APPROVED wenn konfiguriert
      IsRecurring: true,
      RecurringExpenseID: rec.ID,
    }

    s.eventBus.Publish(ctx, &ExpenseCreatedEvent{...})
    rec.NextGenerationDate = rec.nextOccurrence()
    s.eventBus.Publish(ctx, &RecurringExpenseGeneratedEvent{...})
  }
  return nil
}
```

#### 2.1.6 Approval Workflow

**Genehmigungskonfiguration:**

```sql
-- PostgreSQL: expense_schema.approval_settings
CREATE TABLE IF NOT EXISTS approval_settings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL UNIQUE,
  auto_approve_under_cents BIGINT, -- NULL = alle genehmigungspflichtig
  auto_reject_over_cents BIGINT, -- NULL = unbegrenzt
  requires_document_approval BOOLEAN DEFAULT true, -- Photo/PDF erforderlich?
  approval_user_ids UUID[] DEFAULT '{}', -- Genehmiger-Liste (fallback: Tenants)
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id)
);
```

**Genehmigungslogik:**

```
Expense created (Status: PENDING)
  ↓
Check approval_settings
  ├─ IF amount < auto_approve_under_cents
  │  └─ Status: AUTO_APPROVED
  │     Event: ExpenseAutoApprovedEvent
  │
  ├─ IF amount > auto_reject_over_cents
  │  └─ Status: AUTO_REJECTED (for review)
  │     Event: ExpenseAutoRejectedEvent
  │     Notification: Approver
  │
  └─ ELSE
     └─ Status: PENDING_APPROVAL
        Notification: Approver
        Queue: Approval Dashboard
```

**Manual Approval Endpoint:**

```bash
POST /api/v1/expenses/{expenseId}/approve
Content-Type: application/json

{
  "approved": true,
  "approver_notes": "OK, für Büromaterial korrekt kategorisiert",
  "reject_reason": null  // wenn approved=false
}

# KurrentDB Event:
type ExpenseApprovedEvent struct {
  ExpenseID string
  ApprovedBy string
  ApprovedAt time.Time
  Notes string
  NewStatus enum // APPROVED, REJECTED
}
```

**Approval Queue (Read-Model):**

```sql
-- PostgreSQL: expense_schema.approval_queue_view
CREATE OR REPLACE VIEW approval_queue_view AS
SELECT
  e.id,
  e.tenant_id,
  e.vendor_name,
  e.amount,
  e.currency,
  e.category_id,
  c.name_de AS category_name,
  e.transaction_date,
  e.ocr_confidence,
  e.status,
  e.created_at,
  e.receipt_file_id,
  CASE
    WHEN e.amount < 2000 THEN 'low'
    WHEN e.amount < 10000 THEN 'medium'
    ELSE 'high'
  END AS risk_level
FROM expenses e
LEFT JOIN categories c ON e.category_id = c.id
WHERE e.status = 'PENDING_APPROVAL'
  AND e.tenant_id = current_tenant_id()
ORDER BY risk_level DESC, e.created_at ASC;
```

#### 2.1.7 Tax Export Integration

**Export-Formate:**

**A. DATEV Format (Goldstandard für deutsche Steuerkanzleien)**

```
DATEV Ausgabenbelegkopien-Format (EXTF)
Version: 700

"Konto";"Gegenkonto";"Belegdatum";"Belegdatum bis";"Beleg Rechnungsnummer";"Beleg Beschreibung";"Betrag (EUR)";"Umsatzsteuer";"Steuersatz";"Kostenstelle";"Belegnummer";"Belegart";"Belege RoE";"Belege Scan-Hash";"Belege Pfad"

"4320";"1000";"20260320";"20260320";"REC-001";"REWE Markt - Büromaterial";"47,99";"7,66";"19";"KST001";"kassenbon-20260320.jpg";"Kassenbon";"0";"sha256:abc123...";"nas://rentflow/receipts/2026/03/kassenbon-20260320.jpg"
```

**B. WISO Steuerung CSV Format:**

```csv
Betrag,Kategorie,Datum,Beschreibung,Belegnummer,USt.-Betrag,USt.-Satz,Zahlart,Notizen
"47,99","4320","20.03.2026","REWE - Büromaterial","REC-001","7,66","19%","Kasse","Kassenbon fotografiert"
```

**C. Interne PDF-Reportierung (für Vorsteuerabzug):**

```
RentFlow — Ausgabenbericht Q1 2026
═════════════════════════════════════════

Zeitraum: 01.01.2026 — 31.03.2026
Gesamtausgaben: EUR 12.456,78
Vorsteuer (19%): EUR 2.367,79
Vorsteuer (7%): EUR 234,56
Sonstige: EUR 0,00

Kategorieverteilung:
├─ Büromaterial (4320): EUR 847,99 (6,8%)
├─ Reinigung (6360): EUR 2.345,67 (18,8%)
├─ Energieverbrauch (4670): EUR 3.456,78 (27,7%)
├─ Telefon/Internet (6150): EUR 1.234,56 (9,9%)
└─ Sonstige: EUR 4.571,78 (36,8%)

[Tabellarische Detailliste mit alle Einzelpositionen]
```

**Export REST API:**

```bash
POST /api/v1/expenses/export
Content-Type: application/json

{
  "format": "datev" | "wiso_csv" | "pdf_report" | "all",
  "period": {
    "start_date": "2026-01-01",
    "end_date": "2026-03-31"
  },
  "filters": {
    "categories": ["4320", "6360"], // optional: nur bestimmte Kategorien
    "min_amount": null,
    "max_amount": null,
    "approval_status": "APPROVED" // nur genehmigte Ausgaben
  },
  "include_documents": true, // Belege inline (base64) oder als ZIP?
  "include_tax_report": true,
  "recipient_email": "steuerkanzlei@example.de" // optional: per Mail
}

# Response:
HTTP/1.1 200 OK
Content-Type: application/zip

[Binary ZIP mit Dateien:
  - export_DATEV_Q1_2026.txt
  - export_WISO_Q1_2026.csv
  - report_Q1_2026.pdf
  - receipts/ (optional)
]

# Oder für einzelnes Format:
POST /api/v1/expenses/export?format=pdf_report
→ application/pdf
```

**Vor-Export-Validierung:**

```go
type ExportValidator struct {}

func (v *ExportValidator) ValidateForTaxExport(expenses []Expense) error {
  errs := []string{}

  for _, exp := range expenses {
    // 1. Status-Check
    if exp.Status != APPROVED {
      errs = append(errs, fmt.Sprintf("Expense %s: not approved", exp.ID))
    }

    // 2. Category-Check
    if exp.Category == nil {
      errs = append(errs, fmt.Sprintf("Expense %s: no category", exp.ID))
    }

    // 3. Document-Check (GoBD)
    if exp.ReceiptFileID == "" {
      errs = append(errs, fmt.Sprintf("Expense %s: missing receipt", exp.ID))
    }

    // 4. Tax-Check
    if exp.TaxAmount > 0 && exp.TaxRate == 0 {
      errs = append(errs, fmt.Sprintf("Expense %s: tax rate missing", exp.ID))
    }

    // 5. Vendor-Check
    if exp.VendorName == "" {
      errs = append(errs, fmt.Sprintf("Expense %s: no vendor", exp.ID))
    }
  }

  if len(errs) > 0 {
    return fmt.Errorf("export validation failed: %v", errs)
  }
  return nil
}
```

#### 2.1.8 Multi-Currency Support

**Währungs-Konvertierung (via Fixer.io, ECB, oder HubSpot-integriert):**

```sql
CREATE TABLE IF NOT EXISTS currency_rates (
  id UUID PRIMARY KEY,
  base_currency VARCHAR(3) DEFAULT 'EUR',
  target_currency VARCHAR(3),
  rate DECIMAL(10,6),
  rate_date DATE NOT NULL,
  source VARCHAR(50), -- 'ecb', 'fixer', 'openexchangerates'
  created_at TIMESTAMP NOT NULL,
  UNIQUE(base_currency, target_currency, rate_date)
);

-- Täglich aktualisiert via Cronjob oder Event-Handler
```

**Conversion bei Receipt-Upload:**

```go
type CurrencyConverter struct {
  rateCache map[string]ExchangeRate
}

func (c *CurrencyConverter) ConvertToEUR(amount decimal.Decimal, fromCurrency string, date time.Time) (decimal.Decimal, error) {
  if fromCurrency == "EUR" {
    return amount, nil
  }

  rate, err := c.getRate(fromCurrency, "EUR", date)
  if err != nil {
    return decimal.Zero, err
  }

  return amount.Mul(rate.Rate), nil
}

// Event: ExpenseCurrencyConvertedEvent
type ExpenseCurrencyConvertedEvent struct {
  ExpenseID string
  OriginalAmount decimal.Decimal
  OriginalCurrency string
  ConvertedAmount decimal.Decimal
  ConvertedCurrency string // "EUR"
  ExchangeRate decimal.Decimal
  ConversionDate time.Time
  ConvertedAt time.Time
}
```

#### 2.1.9 Budget Tracking

**Budgets pro Kategorie/Projekt/Monat:**

```sql
CREATE TABLE IF NOT EXISTS budgets (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  period_type ENUM('monthly', 'quarterly', 'annually'),
  period_start_date DATE NOT NULL,
  period_end_date DATE NOT NULL,

  scope_type ENUM('category', 'project', 'global'),
  category_id UUID, -- NULL wenn project oder global
  project_id UUID,  -- NULL wenn category oder global

  budget_amount DECIMAL(12,2) NOT NULL,
  currency VARCHAR(3) DEFAULT 'EUR',

  alert_threshold_percent INTEGER DEFAULT 80, -- 80% = Warnung

  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id),
  FOREIGN KEY (category_id) REFERENCES categories(id),
  FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- Budget Status View
CREATE OR REPLACE VIEW budget_status_view AS
SELECT
  b.id,
  b.tenant_id,
  b.period_start_date,
  b.period_end_date,
  b.scope_type,
  COALESCE(c.name_de, p.name, 'Global') AS scope_name,
  b.budget_amount,
  COALESCE(SUM(e.amount), 0) AS spent_amount,
  ROUND(100.0 * COALESCE(SUM(e.amount), 0) / b.budget_amount, 2) AS spent_percent,
  CASE
    WHEN COALESCE(SUM(e.amount), 0) > b.budget_amount THEN 'EXCEEDED'
    WHEN COALESCE(SUM(e.amount), 0) > (b.budget_amount * b.alert_threshold_percent / 100) THEN 'WARNING'
    ELSE 'OK'
  END AS status
FROM budgets b
LEFT JOIN categories c ON b.category_id = c.id
LEFT JOIN projects p ON b.project_id = p.id
LEFT JOIN expenses e ON (
  (b.category_id = e.category_id OR b.category_id IS NULL)
  AND (b.project_id = e.project_id OR b.project_id IS NULL)
  AND e.transaction_date BETWEEN b.period_start_date AND b.period_end_date
  AND e.status = 'APPROVED'
  AND e.tenant_id = b.tenant_id
)
GROUP BY b.id, c.id, p.id;
```

**Alerts via Notification-Service:**

```go
type BudgetAlertEvent struct {
  BudgetID string
  ScopeType string // "category", "project", "global"
  ScopeName string
  BudgetAmount decimal.Decimal
  SpentAmount decimal.Decimal
  SpentPercent float64
  Status enum // "WARNING", "EXCEEDED"
  TriggeredAt time.Time
}

// notification-service subscribet → User erhält Push/E-Mail
```

---

## 3. TECHNISCHE DESIGN (expense-service)

### 3.1 Service-Architektur (Hexagonal)

```
expense-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   ├── expense.go (Expense Aggregate)
│   │   ├── receipt.go
│   │   ├── category.go
│   │   ├── recurring.go
│   │   ├── approval.go
│   │   ├── events.go (KurrentDB Events)
│   │   └── errors.go
│   ├── application/
│   │   ├── services/
│   │   │   ├── expense_service.go
│   │   │   ├── ocr_service.go
│   │   │   ├── categorization_service.go
│   │   │   ├── approval_service.go
│   │   │   ├── export_service.go
│   │   │   └── budget_service.go
│   │   └── dto/
│   │       ├── expense_dto.go
│   │       ├── ocr_dto.go
│   │       └── export_dto.go
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── event_store.go (KurrentDB)
│   │   │   ├── expense_repository.go
│   │   │   ├── category_repository.go
│   │   │   └── postgres_migrations.go
│   │   ├── external/
│   │   │   ├── ai_client.go (→ ai-service)
│   │   │   ├── storage_client.go (→ storage-adapter)
│   │   │   ├── invoice_client.go (→ invoice-service)
│   │   │   └── project_client.go (→ project-service)
│   │   ├── http/
│   │   │   ├── handlers.go
│   │   │   └── routes.go
│   │   └── messagebus/
│   │       └── kurrentdb_subscriber.go
│   └── config/
│       └── config.go
├── migrations/
│   └── postgres/
│       ├── 001_expense_schema.sql
│       ├── 002_approval_settings.sql
│       ├── 003_budgets.sql
│       └── 004_recurring.sql
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── go.mod
└── docker-compose.override.yml
```

### 3.2 KurrentDB Event Definitions

**Event Store Streams:**

```
aggregate-streams:
  expense-{uuid}         → alle Events einer Ausgabe
  receipt-{uuid}         → alle Upload-Events einer Quittung
  recurring-{uuid}       → Änderungen bei Dauerausgabe

category-streams (auto):
  $ce-expense            → alle Expense-Events
  $ce-receipt            → alle Receipt-Events
  $ce-recurring          → alle Recurring-Events
```

**Event Definitions (Go):**

```go
// domain/events.go

// ===== Expense Lifecycle =====

type ExpenseCreatedEvent struct {
  EventID string `json:"event_id"`
  AggregateID string `json:"aggregate_id"` // expense-{uuid}
  AggregateType string `json:"aggregate_type"` // "expense"
  EventType string `json:"event_type"` // "expense.created"
  Version int `json:"version"`
  Timestamp time.Time `json:"timestamp"`
  TenantID string `json:"tenant_id"`

  UserID string `json:"user_id"`
  VendorName string `json:"vendor_name"`
  Amount decimal.Decimal `json:"amount"`
  Currency string `json:"currency"`
  TransactionDate time.Time `json:"transaction_date"`
  Description string `json:"description"`
  PaymentMethod string `json:"payment_method"` // cash, card, check, transfer

  ReceiptFileID *string `json:"receipt_file_id"` // optional
  IsRecurring bool `json:"is_recurring"`
  RecurringExpenseID *string `json:"recurring_expense_id"`

  InitialStatus string `json:"initial_status"` // PENDING, PENDING_APPROVAL, AUTO_APPROVED
}

type ReceiptUploadedEvent struct {
  EventID string
  AggregateID string // receipt-{uuid}
  AggregateType string // "receipt"
  EventType string // "receipt.uploaded"
  Timestamp time.Time
  TenantID string

  UserID string
  ExpenseID *string // optional: kann mehreren Ausgaben zugeordnet werden
  FileName string
  FileSize int64
  MimeType string
  StorageReference string // NAS-Pfad: /mnt/nas/rentflow/receipts/2026/03/...

  UploadMethod string // "mobile_camera", "file_upload", "email", "api"
  OfflineMode bool
}

type OCRCompletedEvent struct {
  EventID string
  AggregateID string // expense-{uuid} oder receipt-{uuid}
  Timestamp time.Time
  TenantID string

  ReceiptFileID string
  ExtractedVendorName string
  ExtractedAmount decimal.Decimal
  ExtractedCurrency string
  ExtractedDate time.Time
  ExtractedTaxAmount decimal.Decimal
  ExtractedTaxRate decimal.Decimal
  ExtractedLineItems []OCRLineItem
  ExtractedPaymentMethod string

  ConfidenceScores map[string]float64 // vendor_name, amount, date, etc.
  RequiresManualReview bool
  AIProvider string // "claude", "openai", "gemini", "ollama"
  ProcessingDurationMs int64
}

type ExpenseCorrectedByUserEvent struct {
  EventID string
  AggregateID string // expense-{uuid}
  Timestamp time.Time
  TenantID string

  UserID string
  FieldsChanged []string // ["vendor_name", "amount", "category_id"]

  // Korrigierte Werte
  VendorName *string
  Amount *decimal.Decimal
  Currency *string
  TransactionDate *time.Time
  TaxAmount *decimal.Decimal
  TaxRate *decimal.Decimal
  PaymentMethod *string
  CategoryID *string
  Description *string

  Notes string // Warum wurde korrigiert?
}

type ExpenseCategorizedEvent struct {
  EventID string
  AggregateID string // expense-{uuid}
  Timestamp time.Time
  TenantID string

  CategoryID string
  CategoryName string
  Confidence float64
  Method string // "vendor_history", "keyword_match", "ml_model", "manual"

  OldCategoryID *string // falls vorher kategorisiert
}

type ExpenseApprovedEvent struct {
  EventID string
  AggregateID string // expense-{uuid}
  Timestamp time.Time
  TenantID string

  ApprovedBy string
  Status string // "APPROVED" oder "REJECTED"
  RejectReason *string
  ApproverNotes string
}

type ExpenseAutoApprovedEvent struct {
  EventID string
  AggregateID string
  Timestamp time.Time
  TenantID string

  Reason string // "under_threshold"
}

type ExpenseDeletedEvent struct {
  EventID string
  AggregateID string // expense-{uuid}
  Timestamp time.Time
  TenantID string

  DeletedBy string
  Reason string
  // Soft-Delete für GoBD!
  // Hard-Delete nur nach 10-Jahres-Retention
}

type ExpenseTaxExportedEvent struct {
  EventID string
  AggregateID string
  Timestamp time.Time
  TenantID string

  ExportFormat string // "datev", "wiso_csv", "pdf_report"
  ExportID string
  ExportedAt time.Time
  ExportedBy string

  IncludedExpenseCount int
  IncludedAmount decimal.Decimal
}

// ===== Recurring Expenses =====

type RecurringExpenseCreatedEvent struct {
  EventID string
  AggregateID string // recurring-{uuid}
  AggregateType string // "recurring-expense"
  EventType string // "recurring-expense.created"
  Timestamp time.Time
  TenantID string

  UserID string
  Name string
  Category Category
  Amount decimal.Decimal
  Currency string
  Frequency string // MONTHLY, QUARTERLY, ANNUALLY
  StartDate time.Time
  EndDate *time.Time
  AutoGenerate bool
  NextGenerationDate time.Time
}

type RecurringExpenseGeneratedEvent struct {
  EventID string
  AggregateID string // recurring-{uuid}
  Timestamp time.Time
  TenantID string

  GeneratedExpenseID string
  GeneratedAmount decimal.Decimal
  GeneratedDate time.Time
}

// ===== Currency Conversion =====

type ExpenseCurrencyConvertedEvent struct {
  EventID string
  AggregateID string
  Timestamp time.Time
  TenantID string

  OriginalAmount decimal.Decimal
  OriginalCurrency string
  ConvertedAmount decimal.Decimal
  ConvertedCurrency string // "EUR"
  ExchangeRate decimal.Decimal
  ConversionDate time.Time
}

// ===== Budget Alerts =====

type BudgetAlertTriggeredEvent struct {
  EventID string
  AggregateID string // budget-{uuid}
  Timestamp time.Time
  TenantID string

  AlertType string // "WARNING", "EXCEEDED"
  BudgetID string
  ScopeType string // category, project, global
  ScopeName string
  BudgetAmount decimal.Decimal
  SpentAmount decimal.Decimal
  SpentPercent float64
}
```

### 3.3 PostgreSQL Schema (Read-Model)

```sql
-- expense_schema.sql

-- ===== Expenses (Haupttabelle) =====

CREATE SCHEMA IF NOT EXISTS expense_schema;

CREATE TABLE IF NOT EXISTS expense_schema.expenses (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,

  -- Kernfelder
  vendor_name VARCHAR(255) NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  currency VARCHAR(3) DEFAULT 'EUR',
  transaction_date DATE NOT NULL,
  tax_amount DECIMAL(12,2),
  tax_rate DECIMAL(4,2), -- 0, 7, 19 (%),
  description TEXT,
  payment_method VARCHAR(50), -- cash, card, check, transfer, unknown

  -- Kategorisierung
  category_id UUID,

  -- Genehmigung
  approval_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    -- PENDING, PENDING_APPROVAL, AUTO_APPROVED, APPROVED, REJECTED
  approved_by UUID,
  approved_at TIMESTAMP,
  approval_notes TEXT,
  rejection_reason TEXT,

  -- Quittung/Beleg
  receipt_file_id UUID,
  receipt_uploaded_at TIMESTAMP,

  -- OCR
  ocr_vendor_name VARCHAR(255),
  ocr_amount DECIMAL(12,2),
  ocr_confidence DECIMAL(4,2), -- 0.0 - 1.0
  ocr_completed_at TIMESTAMP,
  ocr_provider VARCHAR(50), -- claude, openai, gemini, ollama
  requires_manual_review BOOLEAN DEFAULT false,

  -- Dauerauftrag
  is_recurring BOOLEAN DEFAULT false,
  recurring_expense_id UUID,

  -- Steuer-Export
  exported_at TIMESTAMP,
  export_format VARCHAR(50), -- datev, wiso_csv, pdf_report
  export_id UUID,

  -- Audit
  created_by UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP, -- Soft-Delete für GoBD

  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id),
  FOREIGN KEY (category_id) REFERENCES expense_schema.categories(id),
  FOREIGN KEY (approved_by) REFERENCES auth.users(id),
  FOREIGN KEY (receipt_file_id) REFERENCES document_schema.files(id),
  FOREIGN KEY (recurring_expense_id) REFERENCES expense_schema.recurring_expenses(id)
);

CREATE INDEX idx_expenses_tenant_status ON expense_schema.expenses(tenant_id, approval_status);
CREATE INDEX idx_expenses_tenant_date ON expense_schema.expenses(tenant_id, transaction_date);
CREATE INDEX idx_expenses_category ON expense_schema.expenses(category_id);
CREATE INDEX idx_expenses_receipt ON expense_schema.expenses(receipt_file_id);

-- ===== Categories =====

CREATE TABLE IF NOT EXISTS expense_schema.categories (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  accounting_code VARCHAR(10) NOT NULL,
  name_de VARCHAR(255) NOT NULL,
  name_en VARCHAR(255),
  description TEXT,
  skr04_mapping VARCHAR(20), -- SKR04 Kontonummer
  tax_category VARCHAR(50) NOT NULL,
    -- 'full_deductible' (19%), 'partial_deductible' (7%), 'non_deductible' (0%)
  tax_rate_default DECIMAL(4,2),
  requires_approval BOOLEAN DEFAULT false,
  approval_threshold_cents BIGINT, -- NULL = keine Schwelle
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

  UNIQUE(tenant_id, accounting_code),
  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id)
);

-- ===== Approval Settings =====

CREATE TABLE IF NOT EXISTS expense_schema.approval_settings (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL UNIQUE,
  auto_approve_under_cents BIGINT, -- NULL = alle genehmigungspflichtig
  auto_reject_over_cents BIGINT, -- NULL = unbegrenzt
  requires_document_approval BOOLEAN DEFAULT true,
  approval_user_ids UUID[] DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id)
);

-- ===== Recurring Expenses =====

CREATE TABLE IF NOT EXISTS expense_schema.recurring_expenses (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  category_id UUID NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  currency VARCHAR(3) DEFAULT 'EUR',
  frequency VARCHAR(50) NOT NULL, -- MONTHLY, QUARTERLY, ANNUALLY
  start_date DATE NOT NULL,
  end_date DATE,
  auto_generate BOOLEAN DEFAULT false,
  next_generation_date DATE,
  approval_status VARCHAR(50) DEFAULT 'PENDING',
  created_by UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id),
  FOREIGN KEY (category_id) REFERENCES expense_schema.categories(id),
  FOREIGN KEY (created_by) REFERENCES auth.users(id)
);

-- ===== Budgets =====

CREATE TABLE IF NOT EXISTS expense_schema.budgets (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  period_type VARCHAR(50) NOT NULL, -- monthly, quarterly, annually
  period_start_date DATE NOT NULL,
  period_end_date DATE NOT NULL,
  scope_type VARCHAR(50) NOT NULL, -- category, project, global
  category_id UUID,
  project_id UUID,
  budget_amount DECIMAL(12,2) NOT NULL,
  currency VARCHAR(3) DEFAULT 'EUR',
  alert_threshold_percent INTEGER DEFAULT 80,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

  FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id),
  FOREIGN KEY (category_id) REFERENCES expense_schema.categories(id)
);

-- ===== Currency Rates (für Konvertierung) =====

CREATE TABLE IF NOT EXISTS expense_schema.currency_rates (
  id UUID PRIMARY KEY,
  base_currency VARCHAR(3) DEFAULT 'EUR',
  target_currency VARCHAR(3) NOT NULL,
  rate DECIMAL(10,6) NOT NULL,
  rate_date DATE NOT NULL,
  source VARCHAR(50), -- ecb, fixer, openexchangerates
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),

  UNIQUE(base_currency, target_currency, rate_date)
);

-- ===== Approval Queue View =====

CREATE OR REPLACE VIEW expense_schema.approval_queue AS
SELECT
  e.id,
  e.tenant_id,
  e.vendor_name,
  e.amount,
  e.currency,
  c.name_de AS category_name,
  e.transaction_date,
  e.ocr_confidence,
  e.approval_status,
  e.created_at,
  CASE
    WHEN e.amount < 2000 THEN 'low'
    WHEN e.amount < 10000 THEN 'medium'
    ELSE 'high'
  END AS risk_level
FROM expense_schema.expenses e
LEFT JOIN expense_schema.categories c ON e.category_id = c.id
WHERE e.approval_status = 'PENDING_APPROVAL'
  AND e.deleted_at IS NULL
ORDER BY risk_level DESC, e.created_at ASC;

-- ===== Budget Status View =====

CREATE OR REPLACE VIEW expense_schema.budget_status AS
SELECT
  b.id,
  b.tenant_id,
  b.period_start_date,
  b.period_end_date,
  COALESCE(c.name_de, 'Global') AS scope_name,
  b.budget_amount,
  COALESCE(SUM(e.amount), 0) AS spent_amount,
  ROUND(100.0 * COALESCE(SUM(e.amount), 0) / b.budget_amount, 2) AS spent_percent,
  CASE
    WHEN COALESCE(SUM(e.amount), 0) > b.budget_amount THEN 'EXCEEDED'
    WHEN COALESCE(SUM(e.amount), 0) > (b.budget_amount * b.alert_threshold_percent / 100) THEN 'WARNING'
    ELSE 'OK'
  END AS status
FROM expense_schema.budgets b
LEFT JOIN expense_schema.categories c ON b.category_id = c.id
LEFT JOIN expense_schema.expenses e ON (
  (b.category_id = e.category_id OR b.category_id IS NULL)
  AND e.transaction_date BETWEEN b.period_start_date AND b.period_end_date
  AND e.approval_status = 'APPROVED'
  AND e.tenant_id = b.tenant_id
  AND e.deleted_at IS NULL
)
GROUP BY b.id, c.id;
```

### 3.4 REST API Endpoints

| Method | Endpoint | Beschreibung | Auth |
|--------|----------|-------------|------|
| **POST** | `/api/v1/expenses` | Neue Ausgabe erstellen (manual) | TENANT |
| **POST** | `/api/v1/expenses/from-ocr` | Ausgabe aus OCR-Ergebnis | TENANT |
| **GET** | `/api/v1/expenses` | Ausgaben-Liste (mit Filter/Paging) | TENANT |
| **GET** | `/api/v1/expenses/{id}` | Einzelne Ausgabe | TENANT |
| **PATCH** | `/api/v1/expenses/{id}` | Korrigieren (manuell) | TENANT |
| **DELETE** | `/api/v1/expenses/{id}` | Soft-Delete | TENANT |
| **POST** | `/api/v1/expenses/{id}/approve` | Genehmigung | EXPENSE_APPROVER |
| **POST** | `/api/v1/expenses/{id}/reject` | Ablehnung | EXPENSE_APPROVER |
| **GET** | `/api/v1/expenses/approval-queue` | Pending-Liste | EXPENSE_APPROVER |
| **POST** | `/api/v1/receipts/upload` | Foto/PDF hochladen | TENANT |
| **POST** | `/api/v1/receipts/{id}/process-ocr` | OCR manuell triggern | TENANT |
| **GET** | `/api/v1/receipts/{id}` | Receipt-Metadaten | TENANT |
| **GET** | `/api/v1/categories` | Kategorien-Liste | TENANT |
| **POST** | `/api/v1/categories` | Neue Kategorie | ADMIN |
| **PATCH** | `/api/v1/categories/{id}` | Kategorie anpassen | ADMIN |
| **GET** | `/api/v1/recurring` | Daueraufträge | TENANT |
| **POST** | `/api/v1/recurring` | Neuer Dauerauftrag | TENANT |
| **PATCH** | `/api/v1/recurring/{id}` | Dauerauftrag ändern | TENANT |
| **DELETE** | `/api/v1/recurring/{id}` | Dauerauftrag löschen | TENANT |
| **GET** | `/api/v1/budgets` | Budget-Übersicht | TENANT |
| **POST** | `/api/v1/budgets` | Neues Budget | ADMIN |
| **PATCH** | `/api/v1/budgets/{id}` | Budget anpassen | ADMIN |
| **GET** | `/api/v1/expenses/export` | Export DATEV/WISO | FINANCE_ADMIN |
| **POST** | `/api/v1/expenses/export` | Export triggern | FINANCE_ADMIN |
| **GET** | `/api/v1/approval-settings` | Approval-Config | ADMIN |
| **PATCH** | `/api/v1/approval-settings` | Approval-Config ändern | ADMIN |

### 3.5 Service Interaction Sequence Diagram

```
User (Mobile PWA)
  │
  ├─> [1] POST /api/v1/receipts/upload
  │   ├─> expense-service
  │   ├─> storage-adapter (NAS Mount)
  │   └─> KurrentDB: ReceiptUploadedEvent
  │
  ├─> [2] Trigger OCR (async)
  │   ├─> expense-service → ai-service
  │   │   ├─ Claude API (gpt-4-vision)
  │   │   ├─ OpenAI (vision)
  │   │   ├─ Gemini
  │   │   └─ Ollama (fallback)
  │   │
  │   └─> KurrentDB: OCRCompletedEvent
  │
  ├─> [3] Manual Correction (optional)
  │   ├─> expense-service
  │   └─> KurrentDB: ExpenseCorrectedByUserEvent
  │
  ├─> [4] Save Expense
  │   ├─> expense-service
  │   ├─> Categorization Engine (vendor history / keyword match)
  │   ├─> KurrentDB: ExpenseCreatedEvent, ExpenseCategorizedEvent
  │   │
  │   └─> Approval Logic:
  │       ├─ amount < threshold
  │       │  └─ KurrentDB: ExpenseAutoApprovedEvent
  │       └─ amount > threshold
  │          └─ notification-service → Approver
  │
  └─> [5] After Approval:
      ├─> KurrentDB: ExpenseApprovedEvent
      └─> Trigger Budget Check
         └─ budget-service (ggf. BudgetAlertTriggeredEvent)

Finance Manager
  │
  └─> [6] Export Tax Data
      ├─> expense-service: GET /api/v1/expenses/export
      ├─> Validation (status, category, receipt, etc.)
      ├─> Generate DATEV/WISO/PDF
      ├─> storage-adapter: Fetch Receipt-PDFs
      ├─ KurrentDB: ExpenseTaxExportedEvent
      └─> Download ZIP (DATEV + Belege + Report)
```

---

# SECTION 2: NAS STORAGE ARCHITECTURE

## 4. STORAGE-ADAPTER (Shared Library)

### 4.1 Architektur-Übersicht

**Ziel:** Abstraktion der Dateispeicherung = Plugin-Architektur für lokale/NAS/Cloud-Speicherung

**Location:** `pkg/common/storage/`

```
pkg/common/storage/
├── adapter.go          # StorageAdapter Interface
├── local_adapter.go    # LocalStorageAdapter (Fallback)
├── nas_adapter.go      # NASStorageAdapter (Primary)
├── s3_adapter.go       # S3StorageAdapter (Future)
├── metadata.go         # FileMetadata, FileReference
└── errors.go           # Custom Errors
```

### 4.2 StorageAdapter Interface

```go
// pkg/common/storage/adapter.go

package storage

import (
  "context"
  "io"
  "time"
)

type StorageAdapter interface {
  // Store speichert eine Datei und gibt eine Referenz zurück
  Store(ctx context.Context, file io.Reader, metadata FileMetadata) (*FileReference, error)

  // Retrieve lädt eine Datei
  Retrieve(ctx context.Context, fileRef *FileReference) (io.ReadCloser, error)

  // Delete löscht eine Datei (mit Restrictions für GoBD)
  Delete(ctx context.Context, fileRef *FileReference, reason string) error

  // List listet Dateien einer Entity
  List(ctx context.Context, entityType string, entityID string) ([]FileReference, error)

  // GetSignedURL generiert eine zeitlich begrenzte URL
  GetSignedURL(ctx context.Context, fileRef *FileReference, expireDuration time.Duration) (string, error)

  // HealthCheck prüft Verfügbarkeit
  HealthCheck(ctx context.Context) error

  // Capabilities gibt Infos zu diesem Adapter
  Capabilities() AdapterCapabilities
}

type FileMetadata struct {
  TenantID string           // Multi-Tenancy
  EntityType string         // "expense", "equipment", "invoice", etc.
  EntityID string           // UUID der Entität
  FileName string           // Original-Dateiname
  MimeType string           // application/pdf, image/jpeg, etc.
  FileSize int64            // Bytes
  Checksum string           // SHA-256
  IsPrivate bool            // true = nur authenticated access
  RetentionYears int        // 10 = 10 Jahre GoBD
  UploadedBy string         // User UUID
  UploadedAt time.Time
  Description string        // Metadaten-Info
  Tags []string             // z.B. ["receipt", "tax-relevant", "2026-Q1"]
}

type FileReference struct {
  ID string                 // UUID
  StorageKey string         // Backend-spezifischer Key (/mnt/nas/rentflow/receipts/...)
  TenantID string
  EntityType string
  EntityID string
  FileName string
  MimeType string
  FileSize int64
  Checksum string
  CreatedAt time.Time
  ExpiresAt *time.Time       // nil = never expires (normal), set = temporary
  AccessLog []AccessLogEntry // Wer hat wann zugegriffen?
}

type AccessLogEntry struct {
  AccessedBy string   // User UUID
  AccessedAt time.Time
  Method string       // "read", "download", "preview"
  IPAddress string
}

type AdapterCapabilities struct {
  SupportsSignedURLs bool
  SupportsEncryption bool
  SupportsVersioning bool
  MaxFileSize int64 // Bytes
  AllowedMimeTypes []string
  IsReadOnly bool
}
```

### 4.3 NAS Storage Adapter Implementation

```go
// pkg/common/storage/nas_adapter.go

package storage

import (
  "context"
  "crypto/sha256"
  "fmt"
  "io"
  "os"
  "path/filepath"
  "time"
)

type NASStorageAdapter struct {
  config NASConfig
  logger Logger
}

type NASConfig struct {
  MountPath string        // "/mnt/nas/rentflow"
  NASType string          // "smb", "nfs"
  SMBUsername string      // für SMB
  SMBPassword string      // für SMB (encrypted!)
  SMBDomain string        // für SMB
  SMBAddress string       // \\nas.local\rentflow
  NFSAddress string       // nas.local:/export/rentflow (für NFS)
  MaxFileSize int64       // 100 * 1024 * 1024 (100 MB)
  AllowedMimeTypes []string
  VerifySSL bool
  HealthCheckInterval time.Duration
}

func (n *NASStorageAdapter) Store(ctx context.Context, file io.Reader, metadata FileMetadata) (*FileReference, error) {
  // 1. Validierung
  if err := n.validateMetadata(metadata); err != nil {
    return nil, fmt.Errorf("invalid metadata: %w", err)
  }

  // 2. Zielverzeichnis bestimmen
  destDir := n.buildPath(metadata)

  // 3. Verzeichnis erstellen (falls nötig)
  if err := os.MkdirAll(destDir, 0755); err != nil {
    return nil, fmt.Errorf("failed to create directory: %w", err)
  }

  // 4. Dateiname generieren (UUID + Originalname)
  fileID := uuid.New().String()
  destFile := filepath.Join(destDir, fmt.Sprintf("%s_%s", fileID, metadata.FileName))

  // 5. Datei speichern + SHA-256 berechnen
  f, err := os.Create(destFile)
  if err != nil {
    return nil, fmt.Errorf("failed to create file: %w", err)
  }
  defer f.Close()

  hash := sha256.New()
  tee := io.TeeReader(file, hash)

  written, err := io.Copy(f, tee)
  if err != nil {
    os.Remove(destFile) // Cleanup
    return nil, fmt.Errorf("failed to write file: %w", err)
  }

  if written != metadata.FileSize {
    os.Remove(destFile)
    return nil, fmt.Errorf("size mismatch: wrote %d, expected %d", written, metadata.FileSize)
  }

  checksum := fmt.Sprintf("sha256:%x", hash.Sum(nil))

  // 6. FileReference zurückgeben
  ref := &FileReference{
    ID: fileID,
    StorageKey: destFile, // absoluter Pfad für NAS
    TenantID: metadata.TenantID,
    EntityType: metadata.EntityType,
    EntityID: metadata.EntityID,
    FileName: metadata.FileName,
    MimeType: metadata.MimeType,
    FileSize: written,
    Checksum: checksum,
    CreatedAt: time.Now(),
    AccessLog: []AccessLogEntry{
      {
        AccessedBy: metadata.UploadedBy,
        AccessedAt: time.Now(),
        Method: "write",
      },
    },
  }

  return ref, nil
}

func (n *NASStorageAdapter) Retrieve(ctx context.Context, fileRef *FileReference) (io.ReadCloser, error) {
  // 1. Existence Check
  if _, err := os.Stat(fileRef.StorageKey); err != nil {
    return nil, fmt.Errorf("file not found: %w", err)
  }

  // 2. Open + Log
  f, err := os.Open(fileRef.StorageKey)
  if err != nil {
    return nil, err
  }

  // Zugriff loggen (async)
  go func() {
    // In KurrentDB event
    fileRef.AccessLog = append(fileRef.AccessLog, AccessLogEntry{
      AccessedAt: time.Now(),
      Method: "read",
    })
  }()

  return f, nil
}

func (n *NASStorageAdapter) Delete(ctx context.Context, fileRef *FileReference, reason string) error {
  // GoBD Compliance: Löschung nur möglich, wenn Retention abgelaufen
  if fileRef.CreatedAt.AddDate(10, 0, 0).After(time.Now()) { // 10 Jahre Retention
    return fmt.Errorf("file protected by retention policy (GoBD §147). Delete allowed after %s",
      fileRef.CreatedAt.AddDate(10, 0, 0).Format("2006-01-02"))
  }

  if err := os.Remove(fileRef.StorageKey); err != nil {
    return fmt.Errorf("failed to delete file: %w", err)
  }

  // Löschung in Audit-Log
  // → audit-service.LogAction(DeleteFile, reason)

  return nil
}

func (n *NASStorageAdapter) List(ctx context.Context, entityType string, entityID string) ([]FileReference, error) {
  // Liste alle Dateien einer Entity
  // SELECT * FROM file_references WHERE entity_type = ? AND entity_id = ?
  // (aus PostgreSQL, nicht vom Filesystem!)

  // Das ist Read-Model. Filesystem ist nur für Speicher,
  // Metadaten-Abfragen gehen über DB.

  return nil, fmt.Errorf("use PostgreSQL for queries, not filesystem")
}

func (n *NASStorageAdapter) GetSignedURL(ctx context.Context, fileRef *FileReference, expireDuration time.Duration) (string, error) {
  // NAS-Dateien werden NICHT direkt exposed
  // Stattdessen: Service generiert einen API-Token
  // und speichert in Redis: token -> fileRef Mapping (mit TTL)

  token := generateSecureToken(32)
  // redis.Set("file-token:"+token, fileRef.ID, expireDuration)

  return fmt.Sprintf("/api/v1/files/download/%s", token), nil
}

func (n *NASStorageAdapter) HealthCheck(ctx context.Context) error {
  // Prüfe, ob NAS erreichbar ist
  testFile := filepath.Join(n.config.MountPath, ".healthcheck")

  data := []byte(time.Now().Format(time.RFC3339))
  if err := os.WriteFile(testFile, data, 0644); err != nil {
    return fmt.Errorf("NAS health check failed: %w", err)
  }

  if err := os.Remove(testFile); err != nil {
    return fmt.Errorf("NAS cleanup failed: %w", err)
  }

  return nil
}

func (n *NASStorageAdapter) Capabilities() AdapterCapabilities {
  return AdapterCapabilities{
    SupportsSignedURLs: true,    // Via Token in Redis
    SupportsEncryption: false,   // NAS RAID + TLS ist genug
    SupportsVersioning: false,
    MaxFileSize: n.config.MaxFileSize,
    AllowedMimeTypes: n.config.AllowedMimeTypes,
    IsReadOnly: false,
  }
}

// Helper: Pfad-Struktur aufbauen
func (n *NASStorageAdapter) buildPath(metadata FileMetadata) string {
  year := time.Now().Format("2006")
  month := time.Now().Format("01")

  switch metadata.EntityType {
  case "expense":
    return filepath.Join(n.config.MountPath, "tenants", metadata.TenantID, "receipts", year, month)
  case "equipment":
    return filepath.Join(n.config.MountPath, "tenants", metadata.TenantID, "equipment", metadata.EntityID)
  case "document":
    return filepath.Join(n.config.MountPath, "tenants", metadata.TenantID, "documents", year, month)
  case "invoice":
    return filepath.Join(n.config.MountPath, "tenants", metadata.TenantID, "invoices", year, month)
  case "maintenance":
    return filepath.Join(n.config.MountPath, "tenants", metadata.TenantID, "maintenance", metadata.EntityID)
  default:
    return filepath.Join(n.config.MountPath, "tenants", metadata.TenantID, "other")
  }
}
```

### 4.4 Local Storage Adapter (Fallback)

```go
// pkg/common/storage/local_adapter.go
// (Vereinfacht dargestellt)

type LocalStorageAdapter struct {
  basePath string
  logger Logger
}

func (l *LocalStorageAdapter) Store(ctx context.Context, file io.Reader, metadata FileMetadata) (*FileReference, error) {
  // Ähnlich wie NASStorageAdapter, aber speichert lokal
  // basePath = "./storage" oder "/var/lib/rentflow/storage"
  // Für Development und Fallback (falls NAS ausfällt)

  // ...
}
```

### 4.5 File Organization on NAS (Directory Structure)

```
/mnt/nas/rentflow/
│
├── tenants/
│   ├── {tenant-uuid-1}/
│   │   ├── receipts/           # Ausgaben-Belege (Kassenbons, Rechnungen)
│   │   │   └── {year}/
│   │   │       ├── 01/
│   │   │       │   ├── abc123-kassenbon-2026-01-15.jpg
│   │   │       │   ├── def456-rechnung-2026-01-20.pdf
│   │   │       │   └── ghi789-kvrechnung-2026-01-22.pdf
│   │   │       └── 02/, 03/, ...
│   │   │
│   │   ├── equipment/          # Ausrüstungs-Fotos
│   │   │   ├── {equipment-uuid-1}/
│   │   │   │   ├── jkl012-foto-vorne.jpg
│   │   │   │   ├── mno345-foto-hinten.jpg
│   │   │   │   └── pqr678-qr-code.png
│   │   │   └── {equipment-uuid-2}/, ...
│   │   │
│   │   ├── documents/          # Generierte PDFs, Verträge
│   │   │   └── {year}/
│   │   │       ├── 01/
│   │   │       │   ├── xyz789-mietvertrag-2026-01.pdf
│   │   │       │   ├── stu901-packliste-project-abc.pdf
│   │   │       │   └── vwx234-rechnung-proj-xyz.pdf
│   │   │       └── 02/, 03/, ...
│   │   │
│   │   ├── invoices/           # Eingangsr echnung (von Lieferanten)
│   │   │   └── {year}/
│   │   │       ├── 01/
│   │   │       │   ├── yyy111-lieferer-abc-inv-2345.pdf
│   │   │       │   └── yyy222-lieferer-xyz-inv-9876.pdf
│   │   │       └── 02/, 03/, ...
│   │   │
│   │   ├── projects/           # Projekt-spezifische Dateien
│   │   │   ├── {project-uuid-1}/
│   │   │   │   ├── zzz333-lastenheft.pdf
│   │   │   │   ├── zzz444-equipment-list.xlsx
│   │   │   │   └── zzz555-crew-schedule.pdf
│   │   │   └── {project-uuid-2}/, ...
│   │   │
│   │   └── maintenance/        # E-Check Reports, DGUV V3 Zertifikate
│   │       ├── {equipment-uuid-1}/
│   │       │   ├── check-2025-06-15.pdf
│   │       │   ├── check-2024-06-10.pdf
│   │       │   └── prüfprotokoll-2024.pdf
│   │       └── {equipment-uuid-2}/, ...
│   │
│   └── {tenant-uuid-2}/
│       ├── receipts/, equipment/, documents/, ...
│
└── .metadata/                  # Systemdateien
    ├── healthcheck.txt
    └── backup-manifest.json
```

### 4.6 NAS Access Control & Security

**Sicherheitsmodell:**

```
Browser/Client
    ↓
    [HTTPS/TLS]
    ↓
RentFlow Service (mit authentication)
    │
    ├─ User Authorization Check (RBAC)
    │  └─ User darf Datei sehen?
    │
    ├─ File Permission Check
    │  └─ Datei ist private oder public?
    │
    ├─ Tenant Isolation Check
    │  └─ Datei gehört zum User's Tenant?
    │
    └─ Retrieve from NAS
       ├─ storage-adapter.Retrieve()
       │  └─ Read /mnt/nas/rentflow/...
       │
       └─ Stream zurück (mit Virus-Scan optional)
```

**Keine direkten SMB-Shares nach außen!** SMB-Mount ist nur für RentFlow-Service sichtbar.

### 4.7 Encryption & Backup

**Encryption at Rest:**

```
Option 1 (Synology Native):
  ├─ Synology NAS mit ECC RAM + AES-256 Encryption
  ├─ RAID-6 (redundancy)
  └─ Automatic snapshots

Option 2 (File-Level):
  ├─ LUKS Encrypted Volume (/dev/mapper/nas-rentflow)
  ├─ Mounted at /mnt/nas/rentflow
  └─ Keys in Vault (HashiCorp Vault oder Docker Secrets)

Option 3 (ZFS):
  ├─ ZFS with encryption at rest
  ├─ ZFS snapshots for versioning
  └─ Replication to offsite NAS
```

**Backup Strategy:**

```yaml
Backup Plan:
  Primary: NAS RAID-6 (onsite)
  Secondary:
    - Daily snapshots (keep last 7 days)
    - Weekly snapshots (keep last 4 weeks)
    - Monthly backups (offsite, S3-compatible or external drive)

  Retention:
    - Current Year: Full weekly backups
    - Archive: Monthly backups for 10 years (legal requirement)
    - Disaster Recovery: Test restore quarterly
```

**Implementation (via docker-compose):**

```yaml
# docker-compose.yml (excerpt)

nas_sync:
  image: lsyncd
  volumes:
    - /mnt/nas/rentflow:/source:ro
    - /mnt/backup-external/rentflow:/backup
  environment:
    SYNC_INTERVAL: 86400 # Daily
  command: |
    lsyncd -nodaemon /etc/lsyncd/sync.conf
```

---

## 5. SICHERHEIT & COMPLIANCE

### 5.1 GoBD (Grundsätze ordnungsmäßiger Buchführung)

**Anforderungen:**

| Anforderung | Implementierung in RentFlow |
|-------------|---------------------------|
| **Belegablage** | NAS mit Zugriffskontrolle, keine Modifizierbarkeit |
| **Datenintegrität** | SHA-256 Checksums in KurrentDB, Verifizierung bei Download |
| **Vollständigkeit** | Alle Transaktionen im Event Store (audit trail) |
| **Nachvollziehbarkeit** | audit-service loggt Zugriffe & Änderungen |
| **Unveränderbarkeit** | KurrentDB Events sind immutable (append-only) |
| **Verfügbarkeit** | NAS Redundancy + Backups |

**GoBD-konforme Archivierung:**

```go
// expense-service/internal/domain/gobdcompliance.go

type GoBDCompliance struct {
  receiptID string
  uploadedAt time.Time
  fileHash string // SHA-256
  tenantID string
  retentionYears int // 10 (gemäß AO §147)
}

// Bei jeder Datei:
// ✓ Speichern im KurrentDB Event mit Hash
// ✓ Nicht modifizierbar nach Speicherung
// ✓ Zugriff nur über Service (nicht direkt SMB-Share)
// ✓ Löschung erst nach 10 Jahren möglich
```

### 5.2 DSGVO (Datenschutz)

**Anforderungen:**

| Anforderung | Implementierung |
|-------------|-----------------|
| **Datenminimierung** | OCR: Anonymisierung von Personalien |
| **Speicherbegrenzung** | Retention Policy: 10 Jahre (tax), dann Löschung |
| **Zugriffskontrolle** | RBAC pro Tenant, keine Cross-Tenant Zugriffe |
| **Recht auf Löschung** | Soft-Delete in DB, Hard-Delete nach Retention |
| **Audit** | audit-service trackt alle Zugriffe |
| **Verschlüsselung** | TLS in Transit, AES-256 at Rest (Synology) |

**Anonymisierung in OCR:**

```json
// ai-service request:
{
  "anonymize_personal_data": true,
  "mask_patterns": [
    "IBAN",
    "SSN",
    "email",
    "phone"
  ]
}

// Response (sensitiv e Felder gemask t):
{
  "extracted_data": {
    "vendor_name": "REWE Markt GmbH",
    "transaction_date": "2026-03-20",
    "amount": 47.99,
    // KEINE persönlichen Daten in line_items!
    "payment_method": "cash"
  }
}
```

### 5.3 §14 UStG (Vorsteuerabzug)

**Anforderungen für Vorsteuer-Geltendmachung:**

```
Folgende Infos MÜSSEN auf dem Beleg sein:
  ✓ Ausstellungsdatum (oder Empfangsdatum)
  ✓ Rechnungs-/Belegnummer (eindeutig)
  ✓ Leistungsbeschreibung
  ✓ Betrag (brutto + netto)
  ✓ USt.-Satz (explizit)
  ✓ USt.-Betrag (explizit)
  ✓ Lieferant-Name & Adresse
  ✓ Empfänger-Name (oder implizit = Tenant)

In RentFlow:
  ├─ OCR extrahiert alle Felder
  ├─ Manual Correction falls OCR ungenau
  ├─ Validation vor Export (keine fehlenden Felder)
  └─ Export-Format garantiert korrekte Struktur (DATEV)
```

**Tax Validation vor Export:**

```go
func (s *ExportService) ValidateForVorsteuerabzug(expense Expense) error {
  checks := []struct {
    name string
    check func() bool
  }{
    {"date", expense.TransactionDate != nil},
    {"vendor", expense.VendorName != ""},
    {"amount", expense.Amount > 0},
    {"tax_rate", expense.TaxRate > 0 && (expense.TaxRate == 7 || expense.TaxRate == 19)},
    {"tax_amount", expense.TaxAmount > 0},
    {"receipt", expense.ReceiptFileID != ""},
    {"status", expense.Status == "APPROVED"},
  }

  for _, c := range checks {
    if !c.check() {
      return fmt.Errorf("validation failed: missing %s", c.name)
    }
  }

  return nil
}
```

---

## 6. DEPLOYMENT & BETRIEB

### 6.1 Docker Compose Integration

```yaml
# docker-compose.yml (excerpt)

services:
  expense-service:
    image: rentflow/expense-service:latest
    ports:
      - "8018:8018"
    environment:
      # KurrentDB
      KURRENTDB_URL: "esdb://kurrentdb:2113"
      KURRENTDB_USER: "admin"
      KURRENTDB_PASSWORD: "${KURRENTDB_PASSWORD}"

      # PostgreSQL
      DATABASE_URL: "postgres://rentflow:${DB_PASSWORD}@postgres:5432/rentflow?sslmode=disable"
      DATABASE_SCHEMA: "expense_schema"

      # NAS Storage
      STORAGE_BACKEND: "nas"
      NAS_MOUNT_PATH: "/mnt/nas/rentflow"
      NAS_TYPE: "smb"
      # SMB-Credentials (in .env oder Docker Secrets)
      NAS_SMB_USERNAME: "${NAS_USER}"
      NAS_SMB_PASSWORD: "${NAS_PASSWORD}"
      NAS_SMB_DOMAIN: "WORKGROUP"
      NAS_SMB_ADDRESS: "nas.local"

      # AI Service Integration
      AI_SERVICE_URL: "http://ai-service:8014"
      AI_PROVIDERS: "claude,openai,gemini,ollama"
      CLAUDE_API_KEY: "${CLAUDE_API_KEY}"
      OPENAI_API_KEY: "${OPENAI_API_KEY}"

      # Logging
      LOG_LEVEL: "info"
      LOG_FORMAT: "json"

      # CORS
      CORS_ALLOWED_ORIGINS: "https://app.rentflow.local"

    depends_on:
      - kurrentdb
      - postgres
      - ai-service
      - nas

    volumes:
      - /mnt/nas/rentflow:/mnt/nas/rentflow:rw

    networks:
      - rentflow

    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8018/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  nas:
    image: synology-dsm:7.2  # Oder Docker SMB Server
    ports:
      - "445:445"      # SMB
      - "139:139"      # NetBIOS
    environment:
      SMB_USER: "${NAS_USER}"
      SMB_PASSWORD: "${NAS_PASSWORD}"
    volumes:
      - nas-storage:/var/lib/nas
    networks:
      - rentflow

  # NAS Mount Sidecar (für lokale Tests)
  nas-mount:
    image: alpine:latest
    command: sh -c "apk add cifs-utils && mount -t cifs //nas/rentflow /mnt/nas/rentflow -o username=$NAS_USER,password=$NAS_PASSWORD,uid=1000,gid=1000"
    volumes:
      - nas-rentflow:/mnt/nas/rentflow
    depends_on:
      - nas
    networks:
      - rentflow

volumes:
  nas-storage:
  nas-rentflow:

networks:
  rentflow:
    driver: bridge
```

### 6.2 Konfigurationsmanagement

```yaml
# config/expense-service.yaml

server:
  port: 8018
  tlsEnabled: true
  tlsCertFile: /etc/rentflow/certs/expense-service.crt
  tlsKeyFile: /etc/rentflow/certs/expense-service.key

kurrentdb:
  url: ${KURRENTDB_URL}
  credentials:
    username: ${KURRENTDB_USER}
    password: ${KURRENTDB_PASSWORD}
  poolSize: 10
  reconnectDelay: 5s

database:
  url: ${DATABASE_URL}
  maxConnections: 20
  sslMode: require

storage:
  backend: ${STORAGE_BACKEND} # "local", "nas", "s3"

  nas:
    mountPath: ${NAS_MOUNT_PATH}
    type: ${NAS_TYPE} # "smb", "nfs"
    maxFileSize: 104857600 # 100 MB
    allowedMimeTypes:
      - image/jpeg
      - image/png
      - application/pdf
    encryption: aes-256
    backupStrategy:
      enabled: true
      interval: 86400 # daily
      retention: 2592000 # 30 days

  local:
    basePath: ./storage
    maxFileSize: 104857600

ai:
  providers:
    - name: claude
      apiKey: ${CLAUDE_API_KEY}
      model: gpt-4-vision
      priority: 1
    - name: openai
      apiKey: ${OPENAI_API_KEY}
      model: gpt-4-vision
      priority: 2
    - name: gemini
      apiKey: ${GOOGLE_API_KEY}
      model: gemini-1.5-vision
      priority: 3
    - name: ollama
      endpoint: http://localhost:11434
      model: llava
      priority: 4 # Fallback

  defaults:
    confidenceThreshold: 0.75
    anonymizePersonalData: true
    language: de
    processingTimeout: 30s

approval:
  autoApproveUnder: 5000 # EUR (cents), NULL = all manual
  autoRejectOver: 100000 # EUR (cents), NULL = no auto-rejection
  requiresDocumentApproval: true

retention:
  expensesYears: 10 # GoBD AO §147
  archiveAfterMonths: 36
  hardDeleteAfterYears: 10

budget:
  alertingEnabled: true
  alertThreshold: 80 # percent
  checkInterval: 3600 # seconds

logging:
  level: ${LOG_LEVEL}
  format: ${LOG_FORMAT} # "json" or "text"
  outputs:
    - stdout
    - file:///var/log/rentflow/expense-service.log
```

### 6.3 Monitoring & Observability

```go
// internal/infrastructure/monitoring/metrics.go

import "github.com/prometheus/client_golang/prometheus"

var (
  expenseCreatedCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
      Name: "rentflow_expenses_created_total",
      Help: "Total number of expenses created",
    },
    []string{"tenant_id", "status"},
  )

  ocrProcessingDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Name: "rentflow_ocr_processing_duration_seconds",
      Help: "OCR processing duration",
      Buckets: prometheus.ExponentialBuckets(0.1, 2, 8),
    },
    []string{"provider", "status"},
  )

  nasStorageUsage = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Name: "rentflow_nas_storage_bytes",
      Help: "NAS storage usage",
    },
    []string{"tenant_id", "type"},
  )

  approvalQueueLength = prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Name: "rentflow_approval_queue_length",
      Help: "Number of expenses awaiting approval",
    },
    []string{"tenant_id"},
  )
)

// Initialization
func init() {
  prometheus.MustRegister(expenseCreatedCounter)
  prometheus.MustRegister(ocrProcessingDuration)
  prometheus.MustRegister(nasStorageUsage)
  prometheus.MustRegister(approvalQueueLength)
}
```

**Grafana Dashboards:**

```json
{
  "dashboard": {
    "title": "RentFlow Expense Tracking",
    "panels": [
      {
        "title": "Expenses Created (Last 24h)",
        "targets": [
          { "expr": "increase(rentflow_expenses_created_total[24h])" }
        ]
      },
      {
        "title": "OCR Success Rate",
        "targets": [
          { "expr": "increase(rentflow_ocr_processing_duration_seconds_bucket{status='success'}[24h]) / increase(rentflow_ocr_processing_duration_seconds_count[24h])" }
        ]
      },
      {
        "title": "NAS Storage Usage",
        "targets": [
          { "expr": "rentflow_nas_storage_bytes" }
        ]
      },
      {
        "title": "Approval Queue",
        "targets": [
          { "expr": "rentflow_approval_queue_length" }
        ]
      }
    ]
  }
}
```

### 6.4 CI/CD Pipeline

```yaml
# .github/workflows/expense-service.yml

name: Expense Service CI/CD

on:
  push:
    branches: [main, develop]
    paths:
      - 'services/expense-service/**'
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      kurrentdb:
        image: eventstore/eventstore:latest
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - run: go build -o expense-service ./services/expense-service/cmd
      - run: docker build -t rentflow/expense-service:${{ github.sha }} .
      - run: docker push ghcr.io/rentflow/expense-service:${{ github.sha }}

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: aquasecurity/trivy-action@master
        with:
          image-ref: ghcr.io/rentflow/expense-service:${{ github.sha }}
          format: 'sarif'
```

---

## 7. CROSS-SERVICE INTEGRATION

### 7.1 Integration mit invoice-service

**Usecase:** Eingangsrechnung wird gescannt → Automatisch als Expense kategorisiert

**Event Flow:**

```
invoice-service: InvoiceUploadedEvent
  ↓
[KurrentDB Subscription]
  ↓
expense-service: ProcessInvoiceAsExpense()
  ├─ Trigger OCR (invoice_type="incoming_invoice")
  ├─ Extract: vendor, amount, date, tax
  ├─ Link to invoice-service via invoice_id
  └─ Create Expense with status: AUTO_CATEGORIZED
```

**REST Integration:**

```bash
# expense-service calls invoice-service
GET /api/v1/invoices/{invoice-id}
→ { id, vendor, amount, tax_rate, invoice_number, date }

# Then creates Expense
POST /api/v1/expenses
{
  "vendor_name": "...",
  "amount": ...,
  "linked_invoice_id": invoice-id,
  ...
}
```

### 7.2 Integration mit project-service

**Usecase:** Ausgaben einem Projekt zuordnen (für Projektkalkulation)

```go
type Expense struct {
  // ...
  ProjectID *uuid.UUID // Optional: Expense kann Projekt zugeordnet sein
}

// Query: Alle Ausgaben eines Projekts
GET /api/v1/projects/{project-id}/expenses
→ [Expense, ...]

// Projekt-Kosten werden in Reporting miteinbezogen
reporting-service: ProjectCostBreakdown()
  ├─ Fixed Costs (Equipment, Crew)
  ├─ Variable Costs (Expenses aus expense-service)
  └─ Margin Calculation
```

### 7.3 Integration mit reporting-service

**KPI Dashboards:**

```
reporting-service:
  ├─ Total Expenses (month/quarter/year)
  ├─ Expenses by Category (pie chart)
  ├─ Tax Deductibility Analysis (Vorsteuerabzug potential)
  ├─ Budget vs. Actual (per Category)
  ├─ Expense Trend (time series)
  ├─ Approval Efficiency (avg time to approval)
  └─ OCR Accuracy (confidence trend)
```

### 7.4 Integration mit audit-service

**Audit Trail:**

Alle Änderungen an Expenses werden geloggt:

```
audit-service.LogAction(
  action: "ExpenseCreated",
  entityType: "Expense",
  entityID: expense.ID,
  user: user.ID,
  timestamp: now,
  changes: {...}
)

audit-service.LogAction(
  action: "ExpenseApproved",
  entityType: "Expense",
  entityID: expense.ID,
  user: approver.ID,
  changes: {status: "APPROVED"}
)
```

---

## 8. TESTING-STRATEGIE

### 8.1 Unit Tests

```go
// internal/application/services/categorization_service_test.go

func TestCategorizationEngine_VendorHistory(t *testing.T) {
  engine := &CategorizationEngine{
    vendorCache: map[string]VendorProfile{
      "REWE Markt GmbH": {
        DefaultCategory: categoryBüromaterial,
        HistoricalCount: 15,
      },
    },
  }

  expense := OCRExtraction{
    VendorName: "REWE Markt GmbH",
  }

  cat, conf, _ := engine.Categorize(context.Background(), expense)

  assert.Equal(t, categoryBüromaterial.ID, cat.ID)
  assert.Greater(t, conf, 0.9)
}

func TestOCRValidation_MissingFields(t *testing.T) {
  expense := Expense{
    VendorName: "",  // Missing!
    Amount: 47.99,
    TaxRate: 19.0,
  }

  err := ValidateExpenseForApproval(expense)
  assert.Error(t, err)
  assert.Contains(t, err.Error(), "vendor_name")
}
```

### 8.2 Integration Tests

```go
// tests/integration/expense_ocr_test.go

func TestExpenseOCREndToEnd(t *testing.T) {
  // Setup
  ctx := context.Background()
  service := setupExpenseService(t)

  // 1. Upload Receipt Photo
  receiptFile, _ := os.Open("testdata/kassenbon.jpg")
  receiptRef, _ := service.StorageAdapter.Store(ctx, receiptFile, FileMetadata{
    EntityType: "expense",
    FileName: "kassenbon.jpg",
    MimeType: "image/jpeg",
  })

  // 2. Trigger OCR
  ocrResult, _ := service.ProcessOCR(ctx, receiptRef)

  // Assertions
  assert.Equal(t, "REWE Markt GmbH", ocrResult.ExtractedVendorName)
  assert.Equal(t, decimal.NewFromFloat(47.99), ocrResult.ExtractedAmount)
  assert.Greater(t, ocrResult.ConfidenceScores["amount"], 0.9)

  // 3. Create Expense from OCR
  expense, _ := service.CreateExpenseFromOCR(ctx, ocrResult)

  assert.Equal(t, "PENDING_APPROVAL", expense.ApprovalStatus)
  assert.NotNil(t, expense.CategoryID) // Auto-categorized
}
```

### 8.3 End-to-End Tests (PWA)

```javascript
// e2e/tests/receipt-capture.e2e.js (Cypress)

describe('Receipt Photo Capture & OCR', () => {
  it('should capture receipt, process OCR, and approve', () => {
    cy.visit('/app/expenses');
    cy.contains('New Expense').click();

    // Simulate camera capture
    cy.get('[data-testid="camera-button"]').click();
    cy.uploadFile('testdata/kassenbon.jpg', 'image-input');

    // Wait for OCR
    cy.contains('Processing...', { timeout: 5000 }).should('not.exist');

    // Verify OCR Result
    cy.get('[data-testid="vendor-name"]').should('have.value', 'REWE Markt GmbH');
    cy.get('[data-testid="amount"]').should('have.value', '47.99');

    // Save
    cy.contains('Save').click();
    cy.contains('Expense created successfully');
  });
});
```

---

## 9. ROLLOUT-PLAN

### Phase 1: MVP (Woche 1-2)

- [x] Expense Service Scaffolding (Go)
- [x] PostgreSQL Schema (expenses, categories)
- [x] KurrentDB Events (ExpenseCreatedEvent, ExpenseApprovedEvent)
- [x] REST API (CRUD + approve endpoint)
- [x] Manual Expense Creation

### Phase 2: OCR (Woche 3-4)

- [x] Storage Adapter (NAS + Local)
- [x] Receipt Upload (multipart/form-data)
- [x] ai-service Integration (Claude, OpenAI)
- [x] Manual Correction UI
- [x] Mobile PWA (Camera + Upload)

### Phase 3: Automation (Woche 5-6)

- [x] Categorization Engine
- [x] Approval Workflow (configurable)
- [x] Recurring Expenses
- [x] Budget Tracking

### Phase 4: Tax Export (Woche 7-8)

- [x] DATEV Format
- [x] WISO CSV
- [x] PDF Reporting
- [x] Vorsteuer-Validation

### Phase 5: Polish & Compliance (Woche 9-10)

- [x] GoBD-Compliance Testing
- [x] DSGVO Anonymization
- [x] Audit Logging
- [x] Performance Optimization

---

## 10. ERFOLGSKRITERIEN

| Kriterium | Zielwert |
|----------|----------|
| OCR Accuracy (Betrag) | > 98% |
| OCR Processing Time | < 2 Sekunden |
| Approval Queue Resolution | < 24 Stunden |
| Storage Redundancy | RAID-6 + Backups |
| API Availability | > 99.9% |
| Compliance Score (GoBD) | 100% |
| User Adoption (PWA) | > 70% mit Mobile |

---

## ANHANG: Code-Snippets

### A1: Go Service Initialization

```go
// cmd/main.go

func main() {
  ctx := context.Background()
  config := loadConfig()

  // Database
  db, _ := sql.Open("postgres", config.Database.URL)

  // Event Store
  client, _ := esdb.Connect(config.KurrentDB.URL, esdb.WithCredentials(
    config.KurrentDB.Username,
    config.KurrentDB.Password,
  ))
  defer client.Close()

  // Storage Adapter
  var storageAdapter storage.StorageAdapter
  switch config.Storage.Backend {
  case "nas":
    storageAdapter = storage.NewNASStorageAdapter(config.Storage.NAS)
  case "local":
    storageAdapter = storage.NewLocalStorageAdapter(config.Storage.Local)
  }

  // Services
  expenseRepo := expense.NewRepository(db)
  aiClient := ai.NewClient(config.AI)
  expenseService := application.NewExpenseService(expenseRepo, client, storageAdapter, aiClient)

  // HTTP Server
  router := gin.Default()
  handlers := http.NewExpenseHandlers(expenseService)
  handlers.RegisterRoutes(router)

  router.Run(":" + config.Server.Port)
}
```

---

**Dokument ende**

**Dokumentlänge:** ~800 Zeilen
**Formatierung:** Markdown mit Tabellen, Code-Blöcken, Diagrams
**Sprache:** Deutsch (Technical Terms in English)
**Status:** Production-Ready für Implementation
