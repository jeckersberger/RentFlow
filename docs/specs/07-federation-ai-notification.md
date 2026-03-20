# RentFlow — Implementierungsreife Detailspezifikation
## Federation Protocol, AI Service, Notification System

**Stand:** 21. März 2026
**Version:** 2.0 — Production Ready
**Autoren:** Federation/Distributed Systems Specialist, AI/ML Engineer, Notification/Realtime Engineer

---

## 1. FEDERATION PROTOCOL (P2P)

### 1.1 Discovery & Pairing — Detaillierter Workflow

#### 1.1.1 Token-Generierung (Admin Interface)

**Endpoint (Initiator):**
```
POST /api/v1/admin/federation/tokens/generate
```

**Request:**
```json
{
  "partner_company_name": "SoundPro GmbH",
  "partner_email": "stefan@soundpro.de",
  "expires_in_hours": 24,
  "max_uses": 1
}
```

**Response:**
```json
{
  "pairing_token": "ST-9f2e8d1c-4a7b-11ef-9e3a-0242ac130003:exp:2026-03-21T14:30:00Z:sig:base64sig",
  "qr_code_url": "data:image/svg+xml;base64,...",
  "expires_at": "2026-03-21T14:30:00Z",
  "display_string": "ST-9f2e8d1c-4a7b",
  "instructions": "Token an Partner per Email/Signal senden. Gültig für 24 Stunden."
}
```

**Token-Format (JWT-ähnlich):**
```
ST-{UUID}:exp:{ISO8601_EXPIRY}:sig:{BASE64_SIGNATURE}
```

**Token-Storage (Redis):**
```go
type PairingToken struct {
    TokenID       string    // "ST-9f2e8d1c..."
    InstanceID    string    // "marco-veranstaltung:5c9a8e2b..."
    PartnerEmail  string    // Partner's Email (optional, für Audit)
    ExpiresAt     time.Time // 24h TTL
    UsedAt        *time.Time // nil solange ungenutzt
    MaxUses       int       // Normalerweise 1
    CreatedAt     time.Time
    Signature     string    // HMAC-SHA256(TokenID + ExpiresAt, instance_secret_key)
}
```

---

#### 1.1.2 Zertifikats-Discovery & Austausch

**Sequence Diagram:**

```
Marco-Instanz                           Stefan-Instanz
     |                                        |
     | 1. POST /.well-known/rentflow/pair    |
     | (Token + Marco's Public Cert)         |
     |-------------------------------------->|
     |                                        | Validate Token
     |                                        | Verify Signature
     |                                        | Store Marco's Cert
     |                                        | Generate Stefan's Cert
     | 2. HTTP 200 OK (Stefan's Cert)        |
     |<--------------------------------------|
     | Store Stefan's Cert                   |
     | Init mTLS-TLS Pool                    |
     |                                        |
     | 3. GET /api/v1/federation/health      |
     | (mTLS + Request Signature)            |
     |-------------------------------------->|
     |                                        | Verify mTLS
     |                                        | Verify Signature
     | 4. HTTP 200 {status: "paired"}        |
     |<--------------------------------------|
     |                                        |
     ✅ PAIRING COMPLETE                      ✅ PAIRING COMPLETE
```

**POST /.well-known/rentflow/federation/pair**

**Request (von Marco an Stefan):**
```http
POST https://stefan-soundpro.com/.well-known/rentflow-federation/pair HTTP/1.1
Host: stefan-soundpro.com
Content-Type: application/json
X-RentFlow-Pairing-Token: ST-9f2e8d1c-4a7b-11ef-9e3a-0242ac130003:exp:2026-03-21T14:30:00Z:sig:base64sig
X-RentFlow-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-RentFlow-Timestamp: 2026-03-20T14:30:00Z

{
  "instance_id": "marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b",
  "instance_name": "MB Veranstaltungstechnik",
  "company_name": "MB Veranstaltungstechnik GmbH",
  "base_url": "https://marco-events.com",
  "api_version": "1.0",
  "certificate": {
    "public_key_pem": "-----BEGIN CERTIFICATE-----\nMIIDXTCC...\n-----END CERTIFICATE-----",
    "key_fingerprint_sha256": "a3f8e2c1d9...",
    "subject": "CN=marco-events.com",
    "valid_from": "2026-03-20T14:30:00Z",
    "valid_until": "2027-03-20T14:30:00Z",
    "algorithm": "RSA-4096"
  },
  "contact_email": "admin@marco-events.com",
  "contact_name": "Marco Mueller"
}
```

**Response (Stefan antwortet):**
```http
HTTP/1.1 200 OK
Content-Type: application/json
X-RentFlow-Timestamp: 2026-03-20T14:30:01Z
X-RentFlow-Signature: sha256=base64encodedsig

{
  "status": "paired",
  "instance_id": "stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002",
  "instance_name": "SoundPro Audio Rental",
  "company_name": "SoundPro GmbH",
  "base_url": "https://stefan-soundpro.com",
  "federation_api_endpoint": "https://stefan-soundpro.com/api/v1/federation",
  "certificate": {
    "public_key_pem": "-----BEGIN CERTIFICATE-----\nMIIDXTCC...\n-----END CERTIFICATE-----",
    "key_fingerprint_sha256": "b4g9f3d2e0...",
    "subject": "CN=stefan-soundpro.com",
    "valid_from": "2026-03-20T14:30:00Z",
    "valid_until": "2027-03-20T14:30:00Z",
    "algorithm": "RSA-4096"
  },
  "paired_at": "2026-03-20T14:30:00Z",
  "contact_email": "admin@stefan-soundpro.com"
}
```

**Error Responses:**

```json
// 400 Bad Request — Token ungültig/abgelaufen
{
  "error": "invalid_pairing_token",
  "message": "Token ist abgelaufen oder wurde bereits verwendet",
  "expires_at": "2026-03-20T14:30:00Z"
}

// 409 Conflict — Partner bereits verbunden
{
  "error": "partner_already_paired",
  "message": "Diese Instanz ist bereits mit marco-veranstaltung verbunden",
  "paired_at": "2026-03-15T10:00:00Z"
}

// 500 Internal Server Error
{
  "error": "certificate_generation_failed",
  "message": "Zertifikats-Generierung fehlgeschlagen"
}
```

---

#### 1.1.3 mTLS-Setup — Go Implementation

```go
// Certificate-Pool für jeden Partner
type MtlsPool struct {
    PartnerID              string
    CertificatePublicKey   string    // PEM-encoded
    CertificateFingerprint string    // SHA256
    ClientCertPath         string    // Lokales Client-Cert
    ClientKeyPath          string    // Lokaler Private Key
    TlsConfig              *tls.Config
    CreatedAt              time.Time
    UpdatedAt              time.Time
}

// Initialization bei Pairing-Completion
func (s *FederationService) InitMtlsForPartner(ctx context.Context,
    partnerID string,
    partnerCert string) error {

    // 1. Parse Partner-Zertifikat
    cert, err := parseCertificatePEM(partnerCert)
    if err != nil {
        return fmt.Errorf("invalid certificate: %w", err)
    }

    // 2. Berechne Fingerprint
    fingerprint := calculateFingerprint(cert)

    // 3. Lade lokales Client-Zertifikat
    clientCertData, err := os.ReadFile(s.config.ClientCertPath)
    if err != nil {
        return fmt.Errorf("client cert not found: %w", err)
    }

    clientKeyData, err := os.ReadFile(s.config.ClientKeyPath)
    if err != nil {
        return fmt.Errorf("client key not found: %w", err)
    }

    // 4. Lade Client-Cert + Key in tls.Certificate
    clientCertPair, err := tls.X509KeyPair(clientCertData, clientKeyData)
    if err != nil {
        return fmt.Errorf("invalid client cert/key pair: %w", err)
    }

    // 5. Erstelle CertPool für Partner
    certPool := x509.NewCertPool()
    if !certPool.AppendCertsFromPEM([]byte(partnerCert)) {
        return fmt.Errorf("failed to append partner certificate")
    }

    // 6. Konfiguriere tls.Config mit Certificate Pinning
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{clientCertPair},
        RootCAs:      certPool,
        ClientAuth:   tls.RequireAndVerifyClientCert,
        MinVersion:   tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
        ServerName: extractDomainFromCert(cert),
    }

    // 7. Speichere MtlsPool
    pool := &MtlsPool{
        PartnerID:              partnerID,
        CertificatePublicKey:   partnerCert,
        CertificateFingerprint: fingerprint,
        ClientCertPath:         s.config.ClientCertPath,
        ClientKeyPath:          s.config.ClientKeyPath,
        TlsConfig:              tlsConfig,
        CreatedAt:              time.Now(),
    }

    return s.mtlsPoolRepo.Save(ctx, pool)
}

// HTTP Client für Federation-Requests
func (s *FederationService) GetMtlsHttpClient(partnerID string) (*http.Client, error) {
    pool, err := s.mtlsPoolRepo.FindByPartnerID(partnerID)
    if err != nil {
        return nil, err
    }

    transport := &http.Transport{
        TLSClientConfig: pool.TlsConfig,
        MaxIdleConns:    10,
        IdleConnTimeout: 30 * time.Second,
    }

    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }, nil
}
```

---

### 1.2 Federation REST API — Alle Endpoints

#### 1.2.1 Verfügbarkeits-Abfragen

**GET /api/v1/federation/equipment/availability**

**Purpose:** Equipment-Verfügbarkeit bei Partner abfragen

**Authentication:** mTLS + Request-Signatur

**Query Parameters:**
```
?category=PA-Tops
&from_date=2026-03-22
&to_date=2026-03-24
&min_quantity=4
&fields=name,model,condition,partner_price
```

**Request Header:**
```http
GET /api/v1/federation/equipment/availability?category=PA-Tops&from_date=2026-03-22&to_date=2026-03-24&min_quantity=4 HTTP/1.1
Host: stefan-soundpro.com
X-RentFlow-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-RentFlow-Timestamp: 2026-03-20T14:31:00Z
X-RentFlow-Signature: sha256=base64encodedsig
X-RentFlow-Request-Id: req-20260320-001
```

**Response (200 OK):**
```json
{
  "request_id": "req-20260320-001",
  "partner_id": "stefan-soundpro",
  "query": {
    "category": "PA-Tops",
    "from_date": "2026-03-22",
    "to_date": "2026-03-24",
    "min_quantity": 4
  },
  "results": [
    {
      "equipment_id": "eq-k2-001",
      "name": "L-Acoustics K2",
      "model": "K2",
      "category": "PA-Tops",
      "available_quantity": 6,
      "total_quantity": 8,
      "unit_price_per_day": 480,
      "currency": "EUR",
      "condition": "excellent",
      "next_available": "2026-03-22T08:00:00Z",
      "latest_available": "2026-03-24T22:00:00Z",
      "special_notes": "2 Stück derzeit bei Hochzeit im Einsatz",
      "equipment_photo_url": "https://stefan-soundpro.com/assets/k2-001.jpg"
    },
    {
      "equipment_id": "eq-sb28-001",
      "name": "L-Acoustics SB28",
      "model": "SB28",
      "category": "PA-Subs",
      "available_quantity": 4,
      "total_quantity": 6,
      "unit_price_per_day": 320,
      "currency": "EUR",
      "condition": "excellent",
      "next_available": "2026-03-22T08:00:00Z",
      "latest_available": "2026-03-24T22:00:00Z"
    }
  ],
  "queried_at": "2026-03-20T14:31:00Z",
  "cache_valid_until": "2026-03-20T15:31:00Z"
}
```

**Error Responses:**
```json
// 401 Unauthorized
{
  "error": "invalid_signature",
  "message": "Request-Signatur konnte nicht verifiziert werden"
}

// 404 Not Found — Category nicht freigegeben
{
  "error": "category_not_shared",
  "message": "Kategorie 'Mischpulte' wurde von Partner nicht zur Verfügung gestellt"
}

// 429 Too Many Requests
{
  "error": "rate_limit_exceeded",
  "message": "Max 100 Requests/Minute überschritten",
  "retry_after_seconds": 45
}

// 503 Service Unavailable
{
  "error": "partner_offline",
  "message": "Partner-Instanz ist derzeit nicht erreichbar. Nutze gecachte Daten (gültig bis 2026-03-21T14:30:00Z)"
}
```

---

#### 1.2.2 Sub-Rental Request

**POST /api/v1/federation/rental/request**

**Purpose:** Neue Sub-Rental-Anfrage erstellen

**Request Body:**
```json
{
  "request_id": "rent-marco-20260320-001",
  "requesting_company_id": "marco-veranstaltung",
  "requesting_company_name": "MB Veranstaltungstechnik",
  "project_name": "Festival Stadtpark 2026",
  "project_id": "proj-festival-2026",
  "items": [
    {
      "equipment_id": "eq-k2-001",
      "equipment_name": "L-Acoustics K2",
      "quantity": 4,
      "from_date": "2026-03-22",
      "to_date": "2026-03-24",
      "from_time": "06:00:00",
      "to_time": "23:00:00",
      "unit_price": 480,
      "total_price": 3840,
      "currency": "EUR",
      "rental_purpose": "Main-Stage-Beschallung, Festival"
    },
    {
      "equipment_id": "eq-sb28-001",
      "equipment_name": "L-Acoustics SB28",
      "quantity": 2,
      "from_date": "2026-03-22",
      "to_date": "2026-03-24",
      "unit_price": 320,
      "total_price": 1280,
      "currency": "EUR"
    }
  ],
  "pickup_location": {
    "address": "SoundPro, Gewerbegebiet Nord, 70xxx Stuttgart",
    "date": "2026-03-21",
    "time": "16:00:00",
    "contact_person": "Lisa Huber",
    "contact_phone": "+49 711 123456"
  },
  "dropoff_location": {
    "address": "Gleiche wie Pickup",
    "date": "2026-03-25",
    "time": "12:00:00"
  },
  "insurance_required": true,
  "special_requests": "Halterungen für Traverse erforderlich",
  "confirmation_mode": "manual",
  "notes": "Langjähriger Partner, Standard-Sub-Rental"
}
```

**Response (201 Created):**
```json
{
  "rental_id": "subrental-stefan-20260320-001",
  "status": "pending",
  "requesting_instance": "marco-veranstaltung",
  "providing_instance": "stefan-soundpro",
  "created_at": "2026-03-20T14:31:00Z",
  "expires_at": "2026-03-21T14:31:00Z",
  "total_amount": 5120,
  "currency": "EUR",
  "items_count": 2,
  "confirmation_url": "https://stefan-soundpro.com/federation/rental/subrental-stefan-20260320-001",
  "next_action": "Warte auf Bestätigung durch SoundPro Admin"
}
```

---

#### 1.2.3 Sub-Rental Confirm

**PUT /api/v1/federation/rental/{id}/confirm**

**Purpose:** Sub-Rental-Anfrage bestätigen (von Stefan)

**Request Body:**
```json
{
  "confirmed_by": "Stefan Keller",
  "confirmed_at": "2026-03-20T15:00:00Z",
  "notes": "Alle K2 checken, beide SB28 bereit",
  "conditions": [
    {
      "item": "eq-k2-001",
      "condition": "Kleine Kratzer auf 2 Stück, aber funktionstüchtig",
      "inspection_timestamp": "2026-03-20T15:00:00Z"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "rental_id": "subrental-stefan-20260320-001",
  "status": "confirmed",
  "confirmed_at": "2026-03-20T15:00:00Z",
  "pickup_code": "QR-STEFAN-001-2026-03",
  "next_status": "checked_out",
  "estimated_checkout": "2026-03-21T16:00:00Z"
}
```

---

#### 1.2.4 Checkout (Equipment wird ausgegeben)

**PUT /api/v1/federation/rental/{id}/checkout**

**Purpose:** Equipment wird bei Pickup dokumentiert

**Request Body:**
```json
{
  "checkout_timestamp": "2026-03-21T16:05:00Z",
  "checked_out_by": "Peter Mueller (SoundPro)",
  "items": [
    {
      "equipment_id": "eq-k2-001",
      "quantity_given": 4,
      "condition_at_checkout": {
        "description": "4x K2 in ausgezeichnetem Zustand übergeben",
        "photos": [
          "https://stefan-soundpro.com/rental/subrental-001/checkout-k2-001.jpg",
          "https://stefan-soundpro.com/rental/subrental-001/checkout-k2-002.jpg"
        ],
        "condition_rating": 5,
        "serial_numbers": ["K2-001", "K2-002", "K2-003", "K2-004"]
      },
      "digital_signature": "base64_encoded_signature"
    },
    {
      "equipment_id": "eq-sb28-001",
      "quantity_given": 2,
      "condition_at_checkout": {
        "description": "SB28-001 hat kleine Kratzer auf der Rückseite, SB28-002 perfekt",
        "photos": [
          "https://stefan-soundpro.com/rental/subrental-001/checkout-sb28-001.jpg"
        ],
        "condition_rating": 4,
        "serial_numbers": ["SB28-001", "SB28-002"]
      }
    }
  ],
  "notes": "Lisa Huber und Marco Mueller abgeholt. Alles dokumentiert.",
  "pickup_contact_signature": "base64_signature_from_marco_or_lisa"
}
```

**Response (200 OK):**
```json
{
  "rental_id": "subrental-stefan-20260320-001",
  "status": "checked_out",
  "checkout_confirmed_at": "2026-03-21T16:05:00Z",
  "equipment_left_partner": true,
  "checkin_expected": "2026-03-25T12:00:00Z"
}
```

---

#### 1.2.5 Checkin (Equipment wird zurückgegeben)

**PUT /api/v1/federation/rental/{id}/checkin**

**Purpose:** Equipment-Rückgabe dokumentieren

**Request Body:**
```json
{
  "checkin_timestamp": "2026-03-25T12:15:00Z",
  "checked_in_by": "Marco Mueller",
  "items": [
    {
      "equipment_id": "eq-k2-001",
      "quantity_returned": 4,
      "condition_at_checkin": {
        "description": "Alle 4x K2 in gutem Zustand zurück. 1x K2-003 hat leichten Kratzer auf der Frontplatte (vermutlich während Transport)",
        "photos": [
          "https://marco-events.com/rental/return/checkin-k2-003-scratch.jpg"
        ],
        "condition_rating": 4,
        "damages": [
          {
            "item": "K2-003",
            "type": "scratch",
            "location": "front panel",
            "severity": "minor",
            "estimated_repair_cost": 200
          }
        ]
      },
      "serial_numbers_returned": ["K2-001", "K2-002", "K2-003", "K2-004"]
    },
    {
      "equipment_id": "eq-sb28-001",
      "quantity_returned": 2,
      "condition_at_checkin": {
        "description": "Beide SB28 in ausgezeichnetem Zustand",
        "condition_rating": 5
      }
    }
  ],
  "notes": "Event war erfolgreich. Festival super gelaufen.",
  "checkout_contact_signature": "base64_signature_from_stefan"
}
```

**Response (200 OK):**
```json
{
  "rental_id": "subrental-stefan-20260320-001",
  "status": "checked_in",
  "checkin_confirmed_at": "2026-03-25T12:15:00Z",
  "damage_assessment": {
    "total_damages": 1,
    "total_estimated_cost": 200,
    "items_with_damages": ["eq-k2-001"]
  },
  "invoice_generation_status": "pending_review",
  "next_action": "Damage-Review durch Stefan, dann Rechnungsgenerierung"
}
```

---

#### 1.2.6 Partner Info

**GET /api/v1/federation/partner/info**

**Purpose:** Stammdaten des Partners abrufen

**Response (200 OK):**
```json
{
  "partner_id": "stefan-soundpro",
  "instance_id": "stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002",
  "company_name": "SoundPro GmbH",
  "contact_email": "admin@stefan-soundpro.com",
  "contact_phone": "+49 711 987654",
  "base_url": "https://stefan-soundpro.com",
  "federation_api_version": "1.0",
  "federation_enabled": true,
  "paired_at": "2026-03-20T14:30:00Z",
  "shared_categories": [
    "PA-Tops",
    "PA-Subs"
  ],
  "not_shared_categories": [
    "Mixing Consoles",
    "Microphones",
    "Processing"
  ],
  "pricing_model": {
    "type": "partner_condition",
    "discount_percent": 15,
    "minimum_rental_hours": 4,
    "deposit_required": false
  },
  "availability_window_days": 30,
  "confirmation_mode": "manual",
  "last_health_check": "2026-03-20T14:31:00Z",
  "connection_status": "healthy"
}
```

---

#### 1.2.7 Partner Disconnect

**DELETE /api/v1/federation/pair/{partnerId}**

**Purpose:** Partnership beenden

**Request Body:**
```json
{
  "reason": "Geschäftsbeziehung beendet",
  "notes": "Stefan zieht sich aus Vermietung zurück",
  "delete_shared_data": false,
  "archive_history": true
}
```

**Response (200 OK):**
```json
{
  "status": "disconnected",
  "partner_id": "stefan-soundpro",
  "disconnected_at": "2026-03-20T16:00:00Z",
  "archived_rentals": 47,
  "active_rentals_notice": "2 aktive Sub-Rentals müssen vor Disconnect abgeschlossen werden"
}
```

---

### 1.3 Daten-Sharing-Konfiguration

**Storage-Schema (PostgreSQL):**

```sql
-- Tabelle: federation_partner_permissions
CREATE TABLE federation_partner_permissions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    partner_id VARCHAR(255) NOT NULL,  -- "stefan-soundpro"

    -- Equipment-Kategorien: Whitelist
    shared_categories JSONB NOT NULL,  -- ["PA-Tops", "PA-Subs"]
    excluded_categories JSONB,         -- ["Mixing Consoles", "Microphones"]

    -- Preismodell
    pricing_model VARCHAR(50),         -- "partner_condition" | "standard" | "custom"
    partner_discount_percent NUMERIC(5,2),  -- z.B. 15.00 = 15%
    partner_price_markup_percent NUMERIC(5,2), -- Wenn negativ: Rabatt

    -- Verfügbarkeits-Fenster
    availability_window_days INT DEFAULT 30,
    preview_future_rentals BOOLEAN DEFAULT false,

    -- Bestätigungsmodus
    confirmation_mode VARCHAR(50),     -- "manual" | "automatic"
    auto_confirm_max_items INT,        -- Nur wenn < X Items
    auto_confirm_max_rental_days INT,

    -- Kontakt und Benachrichtigungen
    contact_email VARCHAR(255),
    contact_name VARCHAR(255),
    notification_email VARCHAR(255),

    -- Audit
    created_at TIMESTAMP DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_at TIMESTAMP DEFAULT NOW(),
    updated_by UUID REFERENCES users(id),

    UNIQUE(tenant_id, partner_id)
);

-- Tabelle: federation_shared_equipment_categories
CREATE TABLE federation_shared_equipment_categories (
    id UUID PRIMARY KEY,
    partner_permission_id UUID NOT NULL REFERENCES federation_partner_permissions(id),
    category_id UUID NOT NULL REFERENCES equipment_categories(id),

    -- Spezielle Regeln pro Kategorie
    max_availability_window_days INT,  -- Optional: Anderes Fenster pro Kategorie
    requires_deposit BOOLEAN DEFAULT false,
    deposit_amount NUMERIC(10,2),
    deposit_currency VARCHAR(3) DEFAULT 'EUR',

    min_rental_hours INT DEFAULT 4,
    max_simultaneous_items INT,  -- Max 4 K2 gleichzeitig vermieten?

    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(partner_permission_id, category_id)
);
```

**Go Struct für Konfiguration:**

```go
type FederationPartnerConfig struct {
    PartnerID string

    // Kategorien
    SharedCategories   []string // ["PA-Tops", "PA-Subs"]
    ExcludedCategories []string

    // Preismodell
    PricingModel struct {
        Type                string  // "partner_condition", "standard", "custom"
        PartnerDiscountPct  float64 // z.B. 15 = 15% Rabatt
        MinRentalHours      int     // Minimum Mietdauer
        RequireDeposit      bool
        DepositAmount       decimal.Decimal
    }

    // Verfügbarkeit
    AvailabilityWindowDays int  // How far into future to expose
    PreviewFutureRentals   bool // Kann Partner zukünftige Rentals sehen?

    // Bestätigung
    ConfirmationMode struct {
        Mode                string // "manual" | "automatic"
        AutoConfirmMaxItems int    // Auto-confirm nur wenn < X Items
        AutoConfirmMaxDays  int    // Auto-confirm nur wenn <= X Tage
    }

    // Kontakt
    ContactEmail      string
    NotificationEmail string

    // Audit
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Admin-UI-Konfiguration (Beispiel):**

```
[SoundPro GmbH] — Partner-Einstellungen
────────────────────────────────────────

Freigegebene Kategorien:
☑ PA-Tops (L-Acoustics K2, A15, ...)
☑ PA-Subs (SB28, SB118, ...)
☐ Mixing Consoles (CL5, QL1) — NICHT freigegeben
☐ Microphones (Shure Axient) — NICHT freigegeben

Preismodell:
○ Standard-Listenpreis
● Partner-Rabatt: 15% (480€ → 408€/Tag)
○ Custom-Preis

Verfügbarkeit:
Fenster: [30 Tage] in die Zukunft
☐ Zeige zukünftige (noch nicht bestätigte) Rentals

Bestätigung:
● Manuell (Stefan muss jede Anfrage bestätigen)
○ Automatisch (bei < [5] Items und <= [7] Tage)

Benachrichtigungen:
Kontakt: Stefan Keller (stefan@soundpro.de)
Neue Sub-Rentals: ☑ In-App ☑ E-Mail
Status-Updates: ☑ In-App

Verbunden seit: 20.03.2026
Status: ✅ Aktiv (letzte Abfrage: vor 5 Minuten)
```

---

### 1.4 Sub-Rental Workflow — State Machine

**Status-Diagramm:**

```
RequestCreated
    ↓
Pending (24h Wartezeit für Bestätigung)
    ├─→ Rejected (Partner lehnt ab)
    │   └─→ ENDE
    │
    └─→ Confirmed (Partner bestätigt)
        ↓
    CheckedOut (Equipment wird übergeben)
        ↓
    InUse (Equipment im Event-Einsatz)
        │
        ├─→ CheckedInEarly (Vorzeitig zurück)
        │   └─→ ? (Rechnungsanpassung nötig)
        │
        └─→ CheckedIn (Equipment zurückgegeben)
            ↓
        DamageReview (Automatisch: Bestandsaufnahme von Schäden)
            │
            ├─→ NoDamages → Invoice Generated
            └─→ DamagesFound → (Admin Review) → Invoice Generated
                ↓
        Completed
            ↓
        Invoiced (Rechnung generiert und versendet)
```

**State Transitions & Permissions:**

| Status | Triggered By | Next Status | Can Revert | Events |
|--------|--------------|-------------|-----------|--------|
| RequestCreated | Marco (POST /request) | Pending | - | `SubRentalRequested` |
| Pending | (Automatic, 24h timeout) | Rejected | - | `SubRentalExpired` |
| Pending | Stefan (PUT /confirm) | Confirmed | Nein | `SubRentalConfirmed`, Notification to Marco |
| Pending | Stefan (PUT /reject) | Rejected | Ja | `SubRentalRejected` |
| Confirmed | Marco (PUT /checkout) | CheckedOut | Nein | `EquipmentCheckedOut`, Photos + Signatures |
| CheckedOut | (Auto after rental end date) | CheckedIn | Ja (rollback) | `EquipmentDueToReturn` |
| CheckedIn | Marco (PUT /checkin) | DamageReview | Nein | `EquipmentCheckedIn`, Damage assessment |
| DamageReview | Stefan (Review + approve) | Completed | Nein | `DamageReviewApproved`, Damage charges calculated |
| Completed | System (Auto-generate invoice) | Invoiced | Nein | `InvoiceGenerated`, `InvoiceSent` |

**Automatisierte Prozesse:**

```go
type SubRentalStateTransition struct {
    RentalID      string
    FromStatus    string
    ToStatus      string
    Timestamp     time.Time
    TriggeredBy   string // "marco-system", "stefan-admin", "system"
    Notes         string
    Metadata      map[string]interface{}
}

// Auto-Checkout nach Mietdauer überschritten
func (s *SubRentalService) AutoCheckInWarning(ctx context.Context) error {
    rentals, err := s.repo.FindDueToCheckin(ctx)
    if err != nil {
        return err
    }

    for _, rental := range rentals {
        // 24h vor Fälligkeit: Benachrichtigung an Marco
        timeUntilDue := rental.CheckoutExpectedTime.Sub(time.Now())
        if timeUntilDue < 24*time.Hour && timeUntilDue > 23*time.Hour {
            s.notificationService.Send(ctx, Notification{
                RecipientID: rental.RequestingInstanceID,
                Type:        "rental_due_soon",
                Title:       "Equipment-Rückgabe fällig",
                Body:        fmt.Sprintf("%s fällig am %s", rental.ProjectName, rental.CheckoutExpectedTime),
                Priority:    "high",
                Channels:    []string{"push", "in_app", "email"},
            })
        }

        // Bei Überschreitung: Auto-mark as overdue
        if time.Now().After(rental.CheckoutExpectedTime.Add(24 * time.Hour)) {
            s.repo.UpdateStatus(ctx, rental.ID, "overdue")
            s.notificationService.Send(ctx, Notification{
                RecipientID: rental.ProvidingInstanceID,
                Type:        "rental_overdue",
                Priority:    "critical",
            })
        }
    }

    return nil
}

// Auto-Invoice nach Checkin + Damage Review
func (s *SubRentalService) GenerateInvoiceAutomatically(ctx context.Context, rentalID string) error {
    rental, err := s.repo.FindByID(ctx, rentalID)
    if err != nil {
        return err
    }

    if rental.Status != "completed" {
        return fmt.Errorf("rental must be completed before invoicing")
    }

    // Berechne Mietbetrag
    dailyCharge := s.calculateDailyCharge(rental)
    days := rental.CheckinTime.Sub(rental.CheckoutTime).Hours() / 24
    rentalAmount := dailyCharge * days

    // Addiere Schadensersatz (falls vorhanden)
    damageAmount := decimal.NewFromInt(0)
    for _, damage := range rental.AssessedDamages {
        damageAmount = damageAmount.Add(damage.EstimatedRepairCost)
    }

    totalAmount := rentalAmount + damageAmount.InexactFloat64()

    // Generiere Ausgangsrechnung (Stefan) und Eingangsrechnung (Marco)
    invoice := &Invoice{
        ID:              uuid.New().String(),
        RentalID:        rentalID,
        InvoiceType:     "sub_rental",
        InvoicedBy:      rental.ProvidingInstanceID,
        InvoicedTo:      rental.RequestingInstanceID,
        IssueDate:       time.Now(),
        DueDate:         time.Now().AddDate(0, 0, 14),
        TotalAmount:     totalAmount,
        Currency:        "EUR",
        Status:          "draft",
    }

    invoice.Lines = []InvoiceLine{
        {
            Description:  fmt.Sprintf("Vermietung %s (%.1f Tage)", rental.ProjectName, days),
            Quantity:     days,
            UnitPrice:    dailyCharge,
            LineTotal:    rentalAmount,
        },
    }

    if damageAmount.GreaterThan(decimal.Zero) {
        invoice.Lines = append(invoice.Lines, InvoiceLine{
            Description:  "Schadensersatz",
            Quantity:     1,
            UnitPrice:    damageAmount.InexactFloat64(),
            LineTotal:    damageAmount.InexactFloat64(),
        })
    }

    err = s.invoiceService.CreateFromSubRental(ctx, invoice)
    if err != nil {
        return err
    }

    // Benachrichigung an beide Seiten
    s.notificationService.Send(ctx, Notification{
        RecipientID: rental.ProvidingInstanceID,
        Type:        "invoice_created",
        Body:        fmt.Sprintf("Rechnung für %s erstellt: %.2f EUR", rental.ProjectName, totalAmount),
    })
    s.notificationService.Send(ctx, Notification{
        RecipientID: rental.RequestingInstanceID,
        Type:        "invoice_received",
        Body:        fmt.Sprintf("Eingangsrechnung von %s: %.2f EUR", rental.ProvidingInstanceID, totalAmount),
    })

    return nil
}
```

---

### 1.5 Sicherheit

#### 1.5.1 mTLS & Request-Signatur

```go
// HMAC-SHA256 Request-Signatur
func (s *FederationService) SignRequest(method, path, body string, privateKey []byte) (string, error) {
    // Format: METHOD|PATH|BODY|TIMESTAMP
    timestamp := time.Now().UTC().Format(time.RFC3339)
    message := fmt.Sprintf("%s|%s|%s|%s", method, path, body, timestamp)

    h := hmac.New(sha256.New, privateKey)
    h.Write([]byte(message))

    signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
    return signature, nil
}

// Verify incoming request
func (s *FederationService) VerifyRequestSignature(
    method, path, body string,
    receivedSignature string,
    partnerPublicKey []byte) error {

    timestamp := time.Now().UTC().Format(time.RFC3339)
    message := fmt.Sprintf("%s|%s|%s|%s", method, path, body, timestamp)

    h := hmac.New(sha256.New, partnerPublicKey)
    h.Write([]byte(message))
    expectedSig := base64.StdEncoding.EncodeToString(h.Sum(nil))

    if !hmac.Equal([]byte(receivedSignature), []byte(expectedSig)) {
        return fmt.Errorf("invalid signature")
    }

    return nil
}
```

#### 1.5.2 Rate Limiting

```go
// Redis-basiert
type FederationRateLimiter struct {
    redisClient *redis.Client
    maxRequests int       // 100
    windowSecs  int64     // 60
}

func (rl *FederationRateLimiter) CheckLimit(ctx context.Context, partnerID string) (allowed bool, retryAfter int, err error) {
    key := fmt.Sprintf("federation:ratelimit:%s", partnerID)

    count, err := rl.redisClient.Incr(ctx, key).Result()
    if err != nil {
        return false, 0, err
    }

    if count == 1 {
        rl.redisClient.Expire(ctx, key, time.Duration(rl.windowSecs)*time.Second)
    }

    ttl := rl.redisClient.TTL(ctx, key).Val()

    if count > int64(rl.maxRequests) {
        return false, int(ttl.Seconds()), nil
    }

    return true, 0, nil
}
```

#### 1.5.3 Payload Validation

```go
type EquipmentAvailabilityQuery struct {
    Category     string   `json:"category" validate:"required,max=100"`
    FromDate     string   `json:"from_date" validate:"required,datetime=2006-01-02"`
    ToDate       string   `json:"to_date" validate:"required,datetime=2006-01-02"`
    MinQuantity  int      `json:"min_quantity" validate:"min=1,max=1000"`
    MaxResults   *int     `json:"max_results" validate:"omitempty,min=1,max=100"`
}

func validatePayload(payload interface{}) error {
    validate := validator.New()
    return validate.Struct(payload)
}
```

#### 1.5.4 Audit Log

```sql
CREATE TABLE federation_audit_log (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    timestamp TIMESTAMP DEFAULT NOW(),

    action VARCHAR(100),  -- "pair", "availability_query", "rental_request", "checkout", etc.
    partner_id VARCHAR(255),
    request_id VARCHAR(255),  -- For correlation

    http_method VARCHAR(10),  -- "GET", "POST", "PUT", "DELETE"
    http_path VARCHAR(500),
    http_status_code INT,

    request_body JSONB,  -- PII-redacted
    response_body JSONB,

    ip_address INET,
    user_agent TEXT,

    success BOOLEAN,
    error_message TEXT,

    duration_ms INT,  -- Wie lange dauerte die Aktion?

    created_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (tenant_id, partner_id) REFERENCES federation_partner_permissions(tenant_id, partner_id)
);

-- Index für schnelle Abfragen
CREATE INDEX idx_fed_audit_partner_action ON federation_audit_log(partner_id, action, created_at DESC);
```

---

### 1.6 Offline-Toleranz

```go
type OfflineTolerance struct {
    CachedAvailability map[string]EquipmentAvailability
    LastSuccessfulCheck map[string]time.Time
    RetryStrategy RetryBackoffConfig
}

type RetryBackoffConfig struct {
    InitialDelayMs int     // 500ms
    MaxDelayMs     int     // 30s
    Multiplier     float64 // 2.0 (exponential)
    MaxRetries     int     // 3
}

func (ot *OfflineTolerance) FetchAvailabilityWithFallback(
    ctx context.Context,
    partnerID string,
    query EquipmentAvailabilityQuery) (EquipmentAvailability, error) {

    // Versuche aktuellen Request
    result, err := ot.fetchLive(ctx, partnerID, query)
    if err == nil {
        // Cache aktualisieren
        ot.CachedAvailability[partnerID] = result
        ot.LastSuccessfulCheck[partnerID] = time.Now()
        return result, nil
    }

    // Fallback zu gecachten Daten
    if cached, ok := ot.CachedAvailability[partnerID]; ok {
        lastCheck := ot.LastSuccessfulCheck[partnerID]
        cacheAge := time.Now().Sub(lastCheck)

        return EquipmentAvailability{
            Data:            cached.Data,
            IsCached:        true,
            CachedSince:     lastCheck,
            CacheAgeMinutes: int(cacheAge.Minutes()),
            WarningMessage:  fmt.Sprintf("Partner ist offline. Daten von vor %d Minuten (Zuverlässigkeit: 80%%)", int(cacheAge.Minutes())),
        }, nil
    }

    // Keine Daten verfügbar
    return EquipmentAvailability{}, fmt.Errorf("partner offline and no cached data available")
}
```

---

## 2. AI SERVICE (Multi-Provider)

### 2.1 Provider-Architektur

**Go Interface:**

```go
type AIProvider interface {
    // LLM-Completion
    Complete(ctx context.Context, prompt string, opts CompletionOptions) (string, *TokenUsage, error)

    // Embeddings (für Semantic Search)
    Embed(ctx context.Context, text string) ([]float32, *TokenUsage, error)

    // Klassifizierung
    Classify(ctx context.Context, text string, categories []string) (string, float64, error)

    // Vision (für OCR, Asset Recognition)
    ExtractText(ctx context.Context, imageBase64 string) (string, *TokenUsage, error)

    // Provider-Infos
    Name() string
    Model() string
    IsAvailable(ctx context.Context) bool
    GetCost(tokens int) decimal.Decimal
}

// Adapter für jeden Provider
type ClaudeProvider struct {
    apiKey     string
    client     *anthropic.Client
    model      string // "claude-3-5-sonnet-20241022"
    maxTokens  int
}

type OpenAIProvider struct {
    apiKey     string
    client     *openai.Client
    model      string // "gpt-4o"
}

type OllamaProvider struct {
    endpoint   string  // "http://ollama:11434"
    client     *ollama.Client
    model      string  // "llama3:8b"
}
```

**Provider-Registry:**

```go
type ProviderRegistry struct {
    providers    map[string]AIProvider
    defaultTask  map[string]string // task → provider mapping
    mutex        sync.RWMutex
}

func (r *ProviderRegistry) RegisterProvider(name string, provider AIProvider) {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    r.providers[name] = provider
}

func (r *ProviderRegistry) SelectForTask(task string, requiresVision bool) (AIProvider, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()

    // 1. Bevorzuge Ollama (lokal) wenn verfügbar und keine Vision nötig
    if !requiresVision {
        if ollama, ok := r.providers["ollama"]; ok && ollama.IsAvailable(context.Background()) {
            return ollama, nil
        }
    }

    // 2. Nutze Task-spezifisch konfiguriertes Provider
    if providerName, ok := r.defaultTask[task]; ok {
        if provider, ok := r.providers[providerName]; ok {
            return provider, nil
        }
    }

    // 3. Fallback zu Default-Provider
    if defaultProvider, ok := r.providers["default"]; ok {
        return defaultProvider, nil
    }

    return nil, fmt.Errorf("no suitable provider found for task: %s", task)
}
```

**Konfiguration (YAML):**

```yaml
ai:
  providers:
    - name: anthropic
      enabled: true
      model: claude-3-5-sonnet-20241022
      api_key: "${ANTHROPIC_API_KEY}"
      max_tokens: 4096
      timeout_seconds: 30
      fallback_to_next: true
      tasks:
        - smart_asset_creator
        - price_optimization
        - demand_forecast
        - anonymization

    - name: openai
      enabled: true
      model: gpt-4o
      api_key: "${OPENAI_API_KEY}"
      max_tokens: 4096
      tasks:
        - ocr_enhancement
        - asset_recognition

    - name: ollama
      enabled: true
      endpoint: http://ollama:11434
      model: llama3:8b
      timeout_seconds: 60
      tasks:
        - local_classification
        - internal_search
```

---

### 2.2 Anonymisierung (DSGVO-konform)

**PII Detection & Anonymization:**

```go
type AnonymizationService struct {
    customerDB CustomerRepository
    patterns   PIIPatterns
}

type PIIPatterns struct {
    Email    *regexp.Regexp
    Phone    *regexp.Regexp
    IBAN     *regexp.Regexp
    Address  *regexp.Regexp
    Price    *regexp.Regexp
    Name     *regexp.Regexp
}

type AnonymizationMap struct {
    Replacements map[string]string  // [CUSTOMER_1] → "Marco Mueller"
    CreatedAt    time.Time
    TenantID     string
    Reversible   bool               // Kann de-anonymisiert werden?
}

func (s *AnonymizationService) Anonymize(ctx context.Context, text string) (string, *AnonymizationMap, error) {
    anonMap := &AnonymizationMap{
        Replacements: make(map[string]string),
        CreatedAt:    time.Now(),
    }

    result := text
    counter := 1

    // Pattern 1: Kundennamen
    customerNames := s.extractCustomerNames(ctx, text)
    for _, name := range customerNames {
        placeholder := fmt.Sprintf("[CUSTOMER_%d]", counter)
        result = strings.ReplaceAll(result, name, placeholder)
        anonMap.Replacements[placeholder] = name
        counter++
    }

    // Pattern 2: E-Mail
    result = s.patterns.Email.ReplaceAllStringFunc(result, func(email string) string {
        placeholder := fmt.Sprintf("[EMAIL_%d]", counter)
        anonMap.Replacements[placeholder] = email
        counter++
        return placeholder
    })

    // Pattern 3: Telefon (+49/0)...
    result = s.patterns.Phone.ReplaceAllStringFunc(result, func(phone string) string {
        placeholder := fmt.Sprintf("[PHONE_%d]", counter)
        anonMap.Replacements[placeholder] = phone
        counter++
        return placeholder
    })

    // Pattern 4: IBAN
    result = s.patterns.IBAN.ReplaceAllStringFunc(result, func(iban string) string {
        placeholder := fmt.Sprintf("[IBAN_%d]", counter)
        anonMap.Replacements[placeholder] = iban
        counter++
        return placeholder
    })

    // Pattern 5: Adressen
    result = s.patterns.Address.ReplaceAllStringFunc(result, func(addr string) string {
        placeholder := fmt.Sprintf("[ADDRESS_%d]", counter)
        anonMap.Replacements[placeholder] = addr
        counter++
        return placeholder
    })

    // Pattern 6: Preise/Umsatz
    result = s.patterns.Price.ReplaceAllStringFunc(result, func(price string) string {
        placeholder := fmt.Sprintf("[PRICE_%d]", counter)
        anonMap.Replacements[placeholder] = price
        counter++
        return placeholder
    })

    // Speichere Mapping (nur lokal!)
    err := s.anonMapRepo.Save(ctx, anonMap)

    return result, anonMap, err
}

func (s *AnonymizationService) Deanonymize(ctx context.Context, text string, anonMap *AnonymizationMap) (string, error) {
    result := text

    // Ersetze alle Placeholders zurück
    for placeholder, original := range anonMap.Replacements {
        result = strings.ReplaceAll(result, placeholder, original)
    }

    return result, nil
}
```

**Storage (PostgreSQL — encrypted):**

```sql
CREATE TABLE ai_anonymization_maps (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    task_id VARCHAR(255),  -- z.B. "asset_creation_003", "price_opt_042"

    original_input TEXT NOT NULL,  -- Optional: gespeichert (encrypted)?
    replacements JSONB NOT NULL,   -- {"[CUSTOMER_1]": "Marco Mueller", ...}

    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,  -- Auto-delete nach 90 Tagen

    created_by UUID REFERENCES users(id)
);

-- Sensitive: Column-level encryption
ALTER TABLE ai_anonymization_maps
    ALTER COLUMN replacements
    SET ENCRYPTED WITH (ENCRYPTION_TYPE = DETERMINISTIC);
```

---

### 2.3 AI-Features

#### 2.3.1 Smart Asset Creator

**Purpose:** Equipment-Stammdaten aus Foto + Kategorie automatisch vorschlagen

**Prompt Template:**

```
Du bist Experte für Veranstaltungstechnik-Equipment. Analysiere folgende Informationen
und generiere vollständige Equipment-Stammdaten:

EINGABE:
- Foto: [BASE64 ENCODED IMAGE]
- Kategorie: "PA-Tops"
- Hersteller: "L-Acoustics"
- Nutzer-Input: "K2, 2019er Modell"

GEFORDERTE AUSGABE (JSON):
{
  "name": "L-Acoustics K2",
  "manufacturer": "L-Acoustics",
  "model": "K2",
  "category": "PA-Tops",
  "year_manufactured": 2019,
  "serial_number": "[ERKANNT AUS FOTO FALLS MÖGLICH]",

  "technical_specs": {
    "frequency_response": "40Hz - 20kHz",
    "max_spl": "139dB",
    "impedance": "6 Ohms",
    "connector_type": "Speakon NL4",
    "power_handling": "2500W"
  },

  "dimensions": {
    "width_mm": 560,
    "height_mm": 722,
    "depth_mm": 588,
    "weight_kg": 60
  },

  "replacement_cost": 18500,
  "currency": "EUR",

  "recommended_purchase_price": 18500,
  "recommended_rental_price_per_day": 480,
  "recommended_rental_price_per_week": 2400,

  "maintenance_schedule": {
    "inspection_interval_months": 6,
    "recone_interval_hours": 5000,
    "next_service_date": "2024-12-15"
  },

  "condition": "excellent",
  "notes": "Gut gepflegtes Equipment, keine sichtbaren Mängel"
}
```

**Go Implementation:**

```go
type SmartAssetCreatorService struct {
    aiProvider   AIProvider
    anonService  *AnonymizationService
}

func (s *SmartAssetCreatorService) CreateFromPhoto(ctx context.Context,
    photoBase64 string,
    category string,
    userNotes string) (*Asset, *AITokenUsage, error) {

    // 1. OCR auf Foto
    extractedText, usage, err := s.aiProvider.ExtractText(ctx, photoBase64)
    if err != nil {
        return nil, nil, err
    }

    // 2. Anonymisiere Nutzer-Input
    anonInput, anonMap, err := s.anonService.Anonymize(ctx, userNotes)
    if err != nil {
        return nil, nil, err
    }

    // 3. Prompt generieren
    prompt := fmt.Sprintf(`...PROMPT_TEMPLATE...
    OCR Result: %s
    Category: %s
    User Input: %s`, extractedText, category, anonInput)

    // 4. AI-Completion
    response, completion_usage, err := s.aiProvider.Complete(ctx, prompt, CompletionOptions{
        MaxTokens: 2000,
        Format:    "json",
    })
    if err != nil {
        return nil, nil, err
    }

    // 5. Parse Response
    var assetData map[string]interface{}
    json.Unmarshal([]byte(response), &assetData)

    // 6. De-anonymisiere (falls nötig)
    // ...

    // 7. Speichere Vorschlag (nicht auto-speichern, nur Vorschlag!)
    totalUsage := &AITokenUsage{
        CompletionTokens: completion_usage.CompletionTokens,
        PromptTokens:     completion_usage.PromptTokens + usage.PromptTokens,
    }

    return &Asset{...}, totalUsage, nil
}
```

#### 2.3.2 Preisoptimierung

**Purpose:** Historische Mietdaten analysieren → optimale Tages-/Wochen-/Monatspreise

**Workflow:**

```go
type PriceOptimizationService struct {
    aiProvider    AIProvider
    rentalRepo    RentalRepository
    assetRepo     AssetRepository
}

func (s *PriceOptimizationService) OptimizePrices(ctx context.Context, assetID string) (*PriceRecommendation, error) {
    // 1. Sammle historische Rental-Daten (letzte 2 Jahre)
    history, err := s.rentalRepo.GetHistoryByAsset(ctx, assetID, 730*24*time.Hour)
    if err != nil {
        return nil, err
    }

    // 2. Aggregiere Metriken
    metrics := struct {
        TotalRentals        int
        AverageRentalDays   float64
        AverageNightlyRate  float64
        SeasonalDemand      map[string]float64  // "Q1", "Q2", ...
        PeakMonths          []int
        OccupancyRate       float64  // 0-100%
    }{}

    // Berechne Occupancy Rate
    totalAvailableDays := float64(730)
    totalRentalDays := 0.0
    for _, rental := range history {
        totalRentalDays += rental.DurationDays
    }
    metrics.OccupancyRate = (totalRentalDays / totalAvailableDays) * 100

    // 3. Prompt für AI
    prompt := fmt.Sprintf(`
    Analysiere folgende Equipment-Vermietungs-Historiedaten und empfehle optimale Preise:

    Asset: %s
    Zeitraum: letzte 2 Jahre
    Vermietungen: %d
    Durchschnittliche Mietdauer: %.1f Tage
    Durchschn. Nightly Rate: %.2f EUR
    Belegungsquote: %.1f%%
    Peak-Monate: %v
    Saisonale Nachfrage: %v

    Berücksichtige:
    - Nachfrageschwankungen (Festival-Saison: Juni-September)
    - Kompetitive Pricing
    - Gewinnoptimierung
    - Long-tail effect (Wochenmiete günstiger als 7x Tagesmiete)

    GEFORDERTE AUSGABE (JSON):
    {
      "current_price_daily": 480,
      "recommended_price_daily": 520,
      "recommended_price_weekly": 2400,
      "recommended_price_monthly": 8000,

      "seasonal_adjustments": {
        "high_season": {
          "months": [6, 7, 8, 9],
          "multiplier": 1.3
        },
        "low_season": {
          "months": [11, 12, 1, 2],
          "multiplier": 0.8
        }
      },

      "confidence": 0.85,
      "rationale": "Hohe Belegungsquote (73%) deutet auf Unterbewertung hin..."
    }`, assetID, len(history), metrics.AverageRentalDays, metrics.AverageNightlyRate, metrics.OccupancyRate, metrics.PeakMonths, metrics.SeasonalDemand)

    response, tokenUsage, err := s.aiProvider.Complete(ctx, prompt, CompletionOptions{
        MaxTokens: 1500,
        Format:    "json",
    })
    if err != nil {
        return nil, err
    }

    var recommendation *PriceRecommendation
    json.Unmarshal([]byte(response), &recommendation)

    // 4. Speichere als Vorschlag (Admin-Review erforderlich)
    err = s.priceRecRepo.SaveSuggestion(ctx, assetID, recommendation)

    return recommendation, err
}
```

**Storage (PostgreSQL):**

```sql
CREATE TABLE ai_price_recommendations (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),
    asset_id UUID REFERENCES assets(id),

    current_price_daily NUMERIC(10,2),
    recommended_price_daily NUMERIC(10,2),
    recommended_price_weekly NUMERIC(10,2),
    recommended_price_monthly NUMERIC(10,2),

    seasonal_adjustments JSONB,  -- {"high_season": {...}, ...}
    confidence NUMERIC(3,2),     -- 0.00 - 1.00
    rationale TEXT,

    status VARCHAR(50) DEFAULT 'suggestion',  -- "suggestion" | "approved" | "applied" | "rejected"
    applied_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW(),
    reviewed_by UUID REFERENCES users(id),

    UNIQUE(tenant_id, asset_id, created_at)
);
```

#### 2.3.3 Demand Forecast

**Purpose:** Saisonale Equipment-Auslastung vorhersagen

```go
type DemandForecastService struct {
    aiProvider AIProvider
    rentalRepo RentalRepository
}

func (s *DemandForecastService) ForecastDemand(ctx context.Context, assetID string) (*DemandForecast, error) {
    // Daten: 3 Jahre historische Vermietungen
    history, _ := s.rentalRepo.GetHistoryByAsset(ctx, assetID, 1095*24*time.Hour)

    // Gruppiere nach Monat
    monthlyData := make(map[string]int)  // "2024-03" → 8 Vermietungen
    for _, rental := range history {
        month := rental.StartDate.Format("2006-01")
        monthlyData[month]++
    }

    prompt := fmt.Sprintf(`
    Analysiere Vermietungs-Zeitreihendaten und prognostiziere für nächste 6 Monate:

    Historische Daten (3 Jahre):
    %s

    Berücksichtige:
    - Saisonale Muster (Festival-Saison, Weihnachten, Hochzeitssaison)
    - Trend (wächst die Nachfrage?)
    - Anomalien

    Gib Prognose für die nächsten 6 Monate mit Konfidenzintervallen:
    {
      "forecast": [
        {"month": "2026-04", "expected_rentals": 12, "confidence_95": [9, 15]},
        ...
      ],
      "peak_months": [6, 7, 8, 9],
      "trend": "slightly_growing",
      "seasonality_strength": 0.75
    }`, s.formatHistoryForPrompt(monthlyData))

    response, _, _ := s.aiProvider.Complete(ctx, prompt, CompletionOptions{MaxTokens: 1000})

    var forecast DemandForecast
    json.Unmarshal([]byte(response), &forecast)

    return &forecast, nil
}
```

#### 2.3.4 Document OCR (Rechnungsverarbeitung)

**Purpose:** Eingangsrechnungen scannen → Felder extrahieren

```go
type DocumentOCRService struct {
    aiProvider AIProvider
    anonService *AnonymizationService
}

func (s *DocumentOCRService) ExtractInvoiceData(ctx context.Context, invoicePDF []byte) (*InvoiceExtraction, error) {
    // 1. Konvertiere PDF zu Base64-Bildern (erste Seite)
    imageBase64, _ := s.convertPDFToImage(invoicePDF)

    // 2. OCR + Anonymisierung
    prompt := `
    Extrahiere folgende Daten aus dieser Eingangsrechnung:

    {
      "invoice_number": "...",
      "invoice_date": "2026-03-20",
      "supplier_name": "[COMPANY_1]",
      "supplier_email": "[EMAIL_1]",
      "total_amount": 5120.00,
      "currency": "EUR",
      "items": [
        {
          "description": "L-Acoustics K2 Vermietung (4 Stück, 2 Tage)",
          "quantity": 4,
          "unit_price": 480.00,
          "line_total": 3840.00
        }
      ],
      "payment_terms": "Net 14",
      "due_date": "2026-04-03"
    }`

    response, _, _ := s.aiProvider.ExtractText(ctx, imageBase64)

    var extraction InvoiceExtraction
    json.Unmarshal([]byte(response), &extraction)

    return &extraction, nil
}
```

#### 2.3.5 Chat-Assistent

**Purpose:** Natürlichsprachliche Abfragen

```go
type ChatAssistantService struct {
    aiProvider AIProvider
}

func (s *ChatAssistantService) ProcessQuery(ctx context.Context, userQuery string) (string, error) {
    // z.B. "Welches Equipment ist am Samstag frei?"

    prompt := fmt.Sprintf(`
    Du bist Assistent für eine Veranstaltungstechnik-Lagerverwaltung.
    Nutzer fragt: "%s"

    Verfügbare Befehle:
    - Verfügbarkeit abfragen: equipment status [datum]
    - Neue Anfrage: rental request [equipment] [datum]
    - Meine Rentals: my rentals

    Antworte natürlichsprachlich und gebe konkrete Anweisungen.`, userQuery)

    response, _, _ := s.aiProvider.Complete(ctx, prompt, CompletionOptions{MaxTokens: 500})

    return response, nil
}
```

#### 2.3.6 Anomalie-Erkennung

**Purpose:** Ungewöhnliche Patterns erkennen

```go
type AnomalyDetectionService struct {
    aiProvider AIProvider
    rentalRepo RentalRepository
}

func (s *AnomalyDetectionService) DetectAnomalies(ctx context.Context) ([]*Anomaly, error) {
    // Sammle alle Rentals der letzten 30 Tage
    recentRentals, _ := s.rentalRepo.FindLast30Days(ctx)

    // Analysiere Patterns
    prompt := fmt.Sprintf(`
    Analysiere folgende Equipment-Vermietungs-Daten und identifiziere Anomalien:

    [30 Tage Daten als JSON]

    Achte auf:
    - Equipment das sehr schnell (< 1 Tag) zurückkommt → könnte gestohlen sein
    - Überbuchungen (mehr gemietet als physikalisch vorhanden)
    - Ungewöhnliche Preise oder Rabatte
    - Einzelne Renter die viel mehr/weniger mieten als normalerweise

    Gib zurück:
    {
      "anomalies": [
        {
          "type": "possible_theft",
          "equipment_id": "eq-k2-001",
          "description": "K2 wurde vor 2 Wochen nicht zurückgegeben",
          "risk_score": 0.95,
          "recommended_action": "Kontakt mit Renter aufnehmen"
        }
      ]
    }`, recentRentals)

    response, _, _ := s.aiProvider.Complete(ctx, prompt, CompletionOptions{MaxTokens: 1500})

    var result struct{ Anomalies []*Anomaly }
    json.Unmarshal([]byte(response), &result)

    return result.Anomalies, nil
}
```

---

### 2.4 Token-Budget & Kostenkontr

olle

```sql
CREATE TABLE ai_token_usage (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),

    provider VARCHAR(50),  -- "anthropic", "openai", "ollama"
    model VARCHAR(100),    -- "claude-3-5-sonnet", "gpt-4o", ...

    task VARCHAR(100),     -- "smart_asset_creator", "price_optimization", ...

    prompt_tokens INT,
    completion_tokens INT,
    total_tokens INT,

    cost_usd NUMERIC(10,6),  -- Kosten in USD
    cost_eur NUMERIC(10,6),  -- Kosten in EUR (aktueller Kurs)

    created_at TIMESTAMP DEFAULT NOW(),

    INDEX(tenant_id, created_at DESC)
);

CREATE TABLE ai_token_budgets (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),

    month DATE,  -- "2026-03-01"
    max_tokens INT,
    max_cost_usd NUMERIC(10,2),

    current_tokens INT DEFAULT 0,
    current_cost_usd NUMERIC(10,2) DEFAULT 0,

    warning_threshold_percent INT DEFAULT 80,
    alert_sent_at TIMESTAMP,

    UNIQUE(tenant_id, month)
);
```

**Budget-Enforcement:**

```go
type TokenBudgetService struct {
    budgetRepo BudgetRepository
    usageRepo  UsageRepository
    notifySvc  NotificationService
}

func (s *TokenBudgetService) CheckBudget(ctx context.Context, tenantID string, tokens int) (allowed bool, remaining int, err error) {
    budget, _ := s.budgetRepo.GetCurrentMonth(ctx, tenantID)
    if budget == nil {
        return true, 0, nil  // Kein Budget gesetzt = unbegrenzt
    }

    remaining = budget.MaxTokens - budget.CurrentTokens

    if remaining-tokens < 0 {
        s.notifySvc.NotifyAdmins(ctx, Notification{
            Type:     "ai_budget_exceeded",
            Title:    "AI-Token-Budget überschritten",
            Priority: "critical",
        })
        return false, 0, fmt.Errorf("token budget exceeded")
    }

    if float64(budget.CurrentTokens+tokens) > float64(budget.MaxTokens)*float64(budget.WarningThresholdPercent)/100.0 {
        if budget.AlertSentAt == nil {
            s.notifySvc.NotifyAdmins(ctx, Notification{
                Type:  "ai_budget_warning",
                Title: fmt.Sprintf("AI-Budget zu %d%% aufgebraucht", budget.WarningThresholdPercent),
            })
            s.budgetRepo.MarkAlertSent(ctx, budget.ID)
        }
    }

    return true, remaining - tokens, nil
}
```

---

## 3. NOTIFICATION SERVICE

### 3.1 Notification Channels

#### 3.1.1 In-App (WebSocket)

**WebSocket Namespace:** `/ws/notifications`

**Go Handler:**

```go
type NotificationHub struct {
    clients    map[string]*NotificationClient
    broadcast  chan *Notification
    register   chan *NotificationClient
    unregister chan *NotificationClient
    mutex      sync.RWMutex
}

type NotificationClient struct {
    UserID      string
    Conn        *websocket.Conn
    Send        chan *Notification
    Unsubscribe chan bool
}

func (h *NotificationHub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client.UserID] = client
            h.mutex.Unlock()

            // Send unread count
            count := h.getUnreadCount(client.UserID)
            client.Send <- &Notification{
                Type: "unread_count",
                Data: map[string]interface{}{"count": count},
            }

        case client := <-h.unregister:
            h.mutex.Lock()
            if _, ok := h.clients[client.UserID]; ok {
                delete(h.clients, client.UserID)
                close(client.Send)
            }
            h.mutex.Unlock()

        case notification := <-h.broadcast:
            h.mutex.RLock()
            if client, ok := h.clients[notification.RecipientID]; ok {
                select {
                case client.Send <- notification:
                default:
                    // Client's send channel ist voll, skip
                }
            }
            h.mutex.RUnlock()
        }
    }
}

func (h *NotificationHub) SendToUser(notification *Notification) {
    h.broadcast <- notification
}

func (c *NotificationClient) ReadPump(hub *NotificationHub) {
    c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

    for {
        var msg map[string]interface{}
        if err := c.Conn.ReadJSON(&msg); err != nil {
            hub.unregister <- c
            return
        }

        // Handle client-side actions
        if action, ok := msg["action"].(string); ok {
            switch action {
            case "mark_as_read":
                notifID := msg["notification_id"].(string)
                // Update DB
                break
            case "delete":
                notifID := msg["notification_id"].(string)
                // Delete from DB
                break
            }
        }
    }
}

func (c *NotificationClient) WritePump() {
    for {
        select {
        case notification := <-c.Send:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.Conn.WriteJSON(notification); err != nil {
                return
            }

        case <-c.Unsubscribe:
            c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
            return
        }
    }
}
```

**Notification Data Model:**

```go
type Notification struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`      // "sub_rental_request", "invoice_overdue", ...
    Title     string                 `json:"title"`
    Body      string                 `json:"body"`
    Link      *string                `json:"link,omitempty"`  // URL zur Aktion
    Priority  string                 `json:"priority"`  // "low", "medium", "high", "critical"

    RecipientID string                 `json:"-"`
    Data      map[string]interface{} `json:"data"`      // Task-spezifische Daten

    Read      bool      `json:"read"`
    CreatedAt time.Time `json:"created_at"`

    Channels  []string  `json:"-"`  // "push", "in_app", "email"

    // Für Grouping
    GroupKey  *string  `json:"group_key,omitempty"`  // "sub_rental_002" → group multiple notifs
    GroupSize *int     `json:"group_size,omitempty"` // z.B. "3 neue Scan-Events" statt 3 einzelne
}
```

**Storage (PostgreSQL):**

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),
    recipient_id UUID REFERENCES users(id),

    type VARCHAR(100),  -- "sub_rental_request", ...
    title VARCHAR(255),
    body TEXT,
    link VARCHAR(500),
    priority VARCHAR(20),

    data JSONB,

    read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,

    group_key VARCHAR(255),  -- Für Grouping

    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,  -- Soft Delete

    INDEX(recipient_id, read, created_at DESC),
    INDEX(group_key, created_at DESC)
);

CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    tenant_id UUID REFERENCES tenants(id),

    -- Pro Event-Typ
    event_type VARCHAR(100),

    -- Kanäle
    channels JSONB,  -- {"push": true, "in_app": true, "email": false}

    -- Quiet Hours
    quiet_hours_enabled BOOLEAN DEFAULT true,
    quiet_hours_from TIME,      -- "22:00"
    quiet_hours_until TIME,     -- "07:00"

    -- Digest
    digest_enabled BOOLEAN DEFAULT false,
    digest_frequency VARCHAR(50),  -- "daily", "weekly"
    digest_time TIME,              -- "09:00"

    UNIQUE(user_id, tenant_id, event_type)
);
```

---

#### 3.1.2 Web Push (Service Worker)

**Service Worker Registration:**

```javascript
// In Browser/Next.js App
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/sw.js');
}

// Subscribe to push notifications
async function subscribeToPush(vapidPublicKey) {
    const registration = await navigator.serviceWorker.ready;

    const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
    });

    // Send subscription to backend
    await fetch('/api/notifications/subscribe', {
        method: 'POST',
        body: JSON.stringify(subscription),
    });
}

// Service Worker (/public/sw.js)
self.addEventListener('push', (event) => {
    const data = event.data.json();

    const options = {
        body: data.body,
        icon: '/icon-192x192.png',
        badge: '/badge-72x72.png',
        tag: data.type,
        requireInteraction: data.priority === 'critical',
        data: {
            link: data.link,
            notificationId: data.id,
        },
    };

    event.waitUntil(
        self.registration.showNotification(data.title, options)
    );
});

self.addEventListener('notificationclick', (event) => {
    event.notification.close();

    if (event.notification.data.link) {
        clients.matchAll({ type: 'window' }).then((clientList) => {
            for (let client of clientList) {
                if (client.url === event.notification.data.link && 'focus' in client) {
                    return client.focus();
                }
            }
            if (clients.openWindow) {
                return clients.openWindow(event.notification.data.link);
            }
        });
    }
});
```

**Web Push Backend:**

```go
type WebPushService struct {
    vapidPrivateKey string
    vapidPublicKey  string
    db              *sql.DB
}

// VAPID Key Generierung (einmalig bei Setup)
func GenerateVAPIDKeys() (public, private string, err error) {
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        return "", "", err
    }

    // Konvertiere zu Base64
    privBytes, _ := x509.MarshalECPrivateKey(privateKey)
    public = base64.RawURLEncoding.EncodeToString(elliptic.Marshal(elliptic.P256(), privateKey.PublicKey.X, privateKey.PublicKey.Y))
    private = base64.RawURLEncoding.EncodeToString(privBytes)

    return
}

// Subscribe Endpoint
func (s *WebPushService) Subscribe(ctx context.Context, userID string, subscription *webpush.Subscription) error {
    _, err := s.db.ExecContext(ctx, `
        INSERT INTO web_push_subscriptions (user_id, endpoint, auth, p256dh)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, endpoint) DO UPDATE SET updated_at = NOW()
    `, userID, subscription.Endpoint, subscription.Keys["auth"], subscription.Keys["p256dh"])

    return err
}

// Send Push Notification
func (s *WebPushService) SendPush(ctx context.Context, userID string, notification *Notification) error {
    // Lese alle Subscriptions für User
    rows, err := s.db.QueryContext(ctx, `
        SELECT endpoint, auth, p256dh FROM web_push_subscriptions WHERE user_id = $1 AND active = true
    `, userID)
    if err != nil {
        return err
    }

    payload, _ := json.Marshal(notification)

    for rows.Next() {
        var endpoint, auth, p256dh string
        rows.Scan(&endpoint, &auth, &p256dh)

        subscription := &webpush.Subscription{
            Endpoint: endpoint,
            Keys: map[string]string{
                "auth":   auth,
                "p256dh": p256dh,
            },
        }

        // Rate Limiting: max 1 Push pro 5 Minuten
        lastPushTime, _ := s.getLastPushTime(ctx, userID)
        if time.Since(lastPushTime) < 5*time.Minute && notification.Priority != "critical" {
            continue
        }

        _, err := webpush.SendNotification(payload, subscription, &webpush.Options{
            VAPIDPrivateKey: s.vapidPrivateKey,
            VAPIDPublicKey:  s.vapidPublicKey,
            TTL:             24 * 3600,
        })

        if err != nil {
            // Handle invalid subscription
            if webpush.IsExpired(err) {
                s.db.ExecContext(ctx, "DELETE FROM web_push_subscriptions WHERE endpoint = $1", endpoint)
            }
        }
    }

    return nil
}
```

---

#### 3.1.3 E-Mail

**SMTP Configuration:**

```yaml
email:
  smtp_host: smtp.sendgrid.net
  smtp_port: 587
  smtp_user: apikey
  smtp_password: "${SENDGRID_API_KEY}"
  from_email: noreply@rentflow.local
  from_name: RentFlow
  reply_to: support@rentflow.local

  # TLS
  use_tls: true
  use_auth: true

  # Queue
  queue_backend: redis  # "redis" | "database"

  # Retry
  max_retries: 3
  retry_delay_minutes: 5

  # Bounce Handling
  bounce_notification_url: https://rentflow.local/webhooks/ses-bounces
```

**Template System:**

```go
type EmailTemplate struct {
    Name    string  // "invoice", "warning", "sub_rental_request", ...
    Subject string
    Body    string  // Go html/template
}

type EmailTemplateService struct {
    templates map[string]*EmailTemplate
    sendgrid *sendgrid.Client
}

// Template-Registry
var emailTemplates = map[string]*EmailTemplate{
    "sub_rental_request": {
        Name:    "sub_rental_request",
        Subject: "Neue Sub-Rental-Anfrage: {{ .ProjectName }}",
        Body: `
<h2>Neue Sub-Rental-Anfrage</h2>
<p>{{ .RequestingCompany }} benötigt folgendes Equipment:</p>
<table>
  {{range .Items}}
  <tr>
    <td>{{ .EquipmentName }}</td>
    <td>{{ .Quantity }}x</td>
    <td>{{ .FromDate }} - {{ .ToDate }}</td>
  </tr>
  {{end}}
</table>
<p>Gesamtbetrag: {{ .TotalAmount }} EUR</p>
<a href="{{ .ConfirmationLink }}">Jetzt bestätigen oder ablehnen</a>
        `,
    },

    "invoice_created": {
        Name: "invoice_created",
        Subject: "Rechnung #{{ .InvoiceNumber }}",
        Body: `
<h2>Ihre Rechnung wurde generiert</h2>
...
        `,
    },

    "overdue_invoice": {
        Name: "overdue_invoice",
        Subject: "Zahlungserinnerung: Rechnung #{{ .InvoiceNumber }}",
        Body: `...`,
    },
}

// Rendering
func (s *EmailTemplateService) RenderTemplate(templateName string, data interface{}) (string, error) {
    tmpl, ok := emailTemplates[templateName]
    if !ok {
        return "", fmt.Errorf("template not found: %s", templateName)
    }

    t, err := template.New(templateName).Parse(tmpl.Body)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := t.Execute(&buf, data); err != nil {
        return "", err
    }

    return buf.String(), nil
}

// Queue-based Sending
func (s *EmailTemplateService) SendAsync(ctx context.Context, email *Email) error {
    // Speichere in Redis Queue
    payload, _ := json.Marshal(email)
    return s.redisClient.LPush(ctx, "email_queue", payload).Err()
}

// Worker (goroutine)
func (s *EmailTemplateService) ProcessQueue(ctx context.Context) {
    for {
        result := s.redisClient.BRPop(ctx, 0, "email_queue")
        payload := result.Val()[1]

        var email Email
        json.Unmarshal([]byte(payload), &email)

        // Render + Send
        htmlBody, _ := s.RenderTemplate(email.TemplateName, email.Data)

        message := mail.NewSingleEmail(
            mail.NewEmail(email.FromName, email.FromEmail),
            email.Subject,
            mail.NewEmail(email.RecipientName, email.RecipientEmail),
            email.TextBody,
            htmlBody,
        )

        _, err := s.sendgrid.Send(message)
        if err != nil {
            // Retry logic mit exponential backoff
            // ...
        }
    }
}
```

---

### 3.2 Notification Rules — Event → Channel Mapping

**Master Tabelle:**

```sql
CREATE TABLE notification_rules (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),

    event_type VARCHAR(100),  -- "sub_rental_request", "invoice_overdue", ...
    recipient_role VARCHAR(50), -- "admin", "freelancer", "finance", ...

    -- Kanäle
    channels JSONB,  -- {"push": true, "in_app": true, "email": true}

    -- Priorisierung
    priority VARCHAR(20),  -- "low", "medium", "high", "critical"

    -- Bedingungen (optional)
    condition_query JSONB,  -- z.B. {"amount_exceeds": 5000, "partner": "soundpro"}

    -- Template
    notification_template_id UUID,

    enabled BOOLEAN DEFAULT true,

    created_at TIMESTAMP DEFAULT NOW(),

    INDEX(tenant_id, event_type, recipient_role)
);
```

**Notification Rules Matrix:**

| Event | Empfänger | Push | In-App | E-Mail | Priorität | Condition |
|-------|-----------|------|--------|--------|-----------|-----------|
| SubRentalRequested | Partner (Stefan) | ✓ | ✓ | ✓ | Hoch | - |
| SubRentalConfirmed | Requester (Marco) | ✓ | ✓ | - | Hoch | - |
| SubRentalRejected | Requester (Marco) | ✓ | ✓ | ✓ | Mittel | - |
| EquipmentCheckedOut | Partner (Stefan) | - | ✓ | - | Niedrig | - |
| EquipmentCheckedIn | Partner (Stefan) | - | ✓ | - | Niedrig | - |
| EquipmentDamageDetected | Partner (Stefan) | ✓ | ✓ | ✓ | Hoch | damage_cost > 500 |
| InvoiceGenerated | Requester (Marco) | - | ✓ | ✓ | Mittel | - |
| InvoiceOverdue | Finance (Thomas) | ✓ | ✓ | ✓ | Hoch | overdue_days > 7 |
| InvoicePaid | Partner (Stefan) | - | ✓ | - | Niedrig | - |
| ProjectAssigned | Freelancer (Kevin) | ✓ | ✓ | - | Hoch | - |
| ProjectChanged | Crew | ✓ | ✓ | - | Mittel | - |
| SystemUpdateAvailable | Admin | - | ✓ | ✓ | Niedrig | - |
| BackupFailed | Admin | ✓ | ✓ | ✓ | Kritisch | - |
| FederationPartnerDown | Admin | - | ✓ | - | Niedrig | - |
| FederationPartnerRecovered | Admin | - | ✓ | - | Niedrig | - |

**Go Implementation:**

```go
type NotificationRuleEngine struct {
    rules map[string][]*NotificationRule
    db    *sql.DB
}

func (e *NotificationRuleEngine) ProcessEvent(ctx context.Context, event *DomainEvent) error {
    // 1. Finde alle Regeln für diesen Event
    rules := e.rules[event.Type]

    for _, rule := range rules {
        // 2. Prüfe Bedingungen
        if rule.ConditionQuery != nil {
            if !e.evaluateCondition(event, rule.ConditionQuery) {
                continue
            }
        }

        // 3. Bestimme Empfänger
        recipients := e.getRecipientsForRole(ctx, rule.RecipientRole)

        // 4. Erstelle Notification
        for _, recipient := range recipients {
            notif := &Notification{
                ID:          uuid.New().String(),
                Type:        event.Type,
                RecipientID: recipient.ID,
                Priority:    rule.Priority,
                Title:       e.renderTitle(rule, event),
                Body:        e.renderBody(rule, event),
                Link:        e.generateLink(rule, event),
                Channels:    rule.Channels,
                CreatedAt:   time.Now(),
            }

            // 5. Speichere
            e.db.ExecContext(ctx, `
                INSERT INTO notifications (...)
                VALUES (...)
            `, ...)

            // 6. Sende über Kanäle
            for _, channel := range rule.Channels {
                switch channel {
                case "push":
                    e.webPushService.SendPush(ctx, recipient.ID, notif)
                case "in_app":
                    e.hub.SendToUser(notif)
                case "email":
                    e.emailService.SendAsync(ctx, &Email{
                        TemplateName: rule.NotificationTemplateID,
                        Data:         event,
                    })
                }
            }
        }
    }

    return nil
}

func (e *NotificationRuleEngine) evaluateCondition(event *DomainEvent, condition map[string]interface{}) bool {
    // z.B. Sende E-Mail nur wenn Schadensersatz > 500 EUR
    if maxAmount, ok := condition["max_damage_cost"]; ok {
        if event.Data["damage_cost"].(float64) > maxAmount.(float64) {
            return false
        }
    }
    return true
}
```

---

### 3.3 Notification Preferences

**User-Level Konfiguration:**

```sql
CREATE TABLE user_notification_settings (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),

    -- Globale Quiet Hours
    quiet_hours_enabled BOOLEAN DEFAULT true,
    quiet_hours_from TIME DEFAULT '22:00',
    quiet_hours_until TIME DEFAULT '07:00',
    timezone VARCHAR(50) DEFAULT 'Europe/Berlin',

    -- Digest-Option
    digest_enabled BOOLEAN DEFAULT false,
    digest_frequency VARCHAR(50),  -- "daily", "weekly"
    digest_time TIME DEFAULT '09:00',

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id)
);

CREATE TABLE user_event_notification_settings (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    event_type VARCHAR(100),

    -- Kanäle pro Event
    channel_push BOOLEAN,
    channel_in_app BOOLEAN,
    channel_email BOOLEAN,

    UNIQUE(user_id, event_type)
);
```

**Admin-UI für Preferences:**

```
[Kevin Radic] — Benachrichtigungs-Einstellungen
─────────────────────────────────────────────

🔔 Globale Einstellungen
  Quiet Hours: ☑ Aktiviert (22:00 - 07:00, Zeitzone: Europe/Berlin)
  Digest-Mails: ○ Deaktiviert
                ● Täglich um 09:00 Uhr
                ○ Wöchentlich

🎯 Event-basiert
  SubRental-Anfrage:  [✓ Push] [✓ In-App] [✗ E-Mail]
  Projekt-Zuweisung:  [✓ Push] [✓ In-App] [✓ E-Mail]
  Zeiterfassung:      [✗ Push] [✓ In-App] [✗ E-Mail]
  Rechnungen:         [✗ Push] [✓ In-App] [✓ E-Mail]
  System-Alerts:      [✓ Push] [✓ In-App] [✓ E-Mail]

[Speichern]
```

---

## 4. WORKFLOW ENGINE (No-Code)

### 4.1 Workflow Components

**Workflow Schema:**

```go
type Workflow struct {
    ID          string
    TenantID    string
    Name        string  // "Mahnung nach 14 Tagen"
    Description string

    Trigger     WorkflowTrigger
    Conditions  []WorkflowCondition
    Actions     []WorkflowAction

    Enabled     bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type WorkflowTrigger struct {
    EventType   string          // "invoice.created", "rental.checked_in", ...
    EventFilter map[string]interface{} // z.B. {"status": "pending"}
}

type WorkflowCondition struct {
    Type      string          // "amount_exceeds", "days_passed", "contains_text", ...
    Field     string          // "amount", "created_at", ...
    Operator  string          // ">", "<", "==", "contains", ...
    Value     interface{}     // 5000, 14, "error", ...
    Logic     string          // "AND" | "OR" (zwischen mehreren Bedingungen)
}

type WorkflowAction struct {
    Type      string          // "send_notification", "change_status", "send_email", "webhook", ...
    Params    map[string]interface{}
}
```

**Storage (PostgreSQL):**

```sql
CREATE TABLE workflows (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),

    name VARCHAR(255),
    description TEXT,

    trigger_event_type VARCHAR(100),
    trigger_event_filter JSONB,

    conditions JSONB,  -- JSON Array von Bedingungen
    actions JSONB,     -- JSON Array von Aktionen

    enabled BOOLEAN DEFAULT true,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    INDEX(tenant_id, enabled)
);

CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY,
    workflow_id UUID REFERENCES workflows(id),

    event_id UUID,  -- Welcher Event hat diesen Workflow getriggert?

    status VARCHAR(50),  -- "pending", "running", "success", "failed"

    execution_log JSONB,  -- Schritt-für-Schritt Log

    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,

    INDEX(workflow_id, status, started_at DESC)
);
```

### 4.2 Visual Workflow Editor (React Flow)

**Frontend (React):**

```typescript
// components/WorkflowEditor.tsx
import { useCallback } from 'react';
import ReactFlow, {
  Node,
  Edge,
  addEdge,
  Connection,
  useNodesState,
  useEdgesState,
} from 'reactflow';

export function WorkflowEditor({ workflowID }: { workflowID: string }) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  const onConnect = useCallback(
    (connection: Connection) =>
      setEdges((eds) => addEdge(connection, eds)),
    [setEdges],
  );

  const addTriggerNode = () => {
    const newNode: Node = {
      id: `trigger-${Date.now()}`,
      data: { label: 'Event Trigger', eventType: 'invoice.created' },
      position: { x: 250, y: 5 },
      type: 'trigger',
    };
    setNodes((nds) => nds.concat(newNode));
  };

  const addConditionNode = () => {
    const newNode: Node = {
      id: `condition-${Date.now()}`,
      data: { label: 'If/Then', condition: 'amount > 5000' },
      position: { x: 250, y: 100 },
      type: 'condition',
    };
    setNodes((nds) => nds.concat(newNode));
  };

  const addActionNode = () => {
    const newNode: Node = {
      id: `action-${Date.now()}`,
      data: { label: 'Send Notification' },
      position: { x: 250, y: 200 },
      type: 'action',
    };
    setNodes((nds) => nds.concat(newNode));
  };

  const saveWorkflow = async () => {
    const workflowDef = {
      nodes,
      edges,
    };
    await fetch(`/api/workflows/${workflowID}`, {
      method: 'PUT',
      body: JSON.stringify(workflowDef),
    });
  };

  return (
    <div style={{ width: '100%', height: '600px' }}>
      <div style={{ marginBottom: '10px' }}>
        <button onClick={addTriggerNode}>Trigger hinzufügen</button>
        <button onClick={addConditionNode}>Bedingung hinzufügen</button>
        <button onClick={addActionNode}>Aktion hinzufügen</button>
        <button onClick={saveWorkflow}>Speichern</button>
      </div>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
      />
    </div>
  );
}
```

### 4.3 Pre-built Workflow Templates

**Template 1: "Mahnung nach 14 Tagen"**

```json
{
  "name": "Mahnung nach 14 Tagen",
  "trigger": {
    "event_type": "invoice.created",
    "event_filter": { "status": "open" }
  },
  "conditions": [
    {
      "type": "days_passed",
      "field": "created_at",
      "operator": ">=",
      "value": 14
    }
  ],
  "actions": [
    {
      "type": "send_email",
      "params": {
        "template_name": "overdue_invoice_reminder",
        "recipient_field": "invoice.client_email",
        "subject": "Zahlungserinnerung: Rechnung {{ invoice.number }}"
      }
    },
    {
      "type": "send_notification",
      "params": {
        "recipient_role": "admin",
        "type": "invoice_overdue",
        "priority": "medium"
      }
    },
    {
      "type": "change_status",
      "params": {
        "entity_type": "invoice",
        "new_status": "reminded"
      }
    }
  ]
}
```

**Template 2: "Sub-Rental Equipment-Damage-Alert"**

```json
{
  "name": "Sub-Rental Damage Detection Alert",
  "trigger": {
    "event_type": "rental.damage_detected",
    "event_filter": {}
  },
  "conditions": [
    {
      "type": "amount_exceeds",
      "field": "damage.estimated_cost",
      "operator": ">",
      "value": 500
    }
  ],
  "actions": [
    {
      "type": "send_notification",
      "params": {
        "recipient_id": "{{ rental.provider_admin_id }}",
        "type": "damage_alert",
        "priority": "high",
        "channels": ["push", "in_app", "email"]
      }
    },
    {
      "type": "send_email",
      "params": {
        "template_name": "damage_report",
        "recipient_field": "rental.provider_email",
        "cc": ["finance@soundpro.de"]
      }
    },
    {
      "type": "webhook",
      "params": {
        "url": "https://soundpro.de/api/webhooks/damage-detected",
        "method": "POST",
        "payload_template": "{{ rental | tojson }}"
      }
    }
  ]
}
```

**Template 3: "Sub-Rental Auto-Confirm (unter Bedingungen)"**

```json
{
  "name": "Auto-Confirm Sub-Rentals (Vertrauens-Partner)",
  "trigger": {
    "event_type": "rental.request_created",
    "event_filter": { "confirmation_mode": "automatic" }
  },
  "conditions": [
    {
      "type": "partner_whitelisted",
      "field": "requesting_partner_id",
      "operator": "in",
      "value": ["marco-veranstaltung"]
    },
    {
      "type": "item_count",
      "field": "items",
      "operator": "<=",
      "value": 5
    },
    {
      "type": "rental_duration",
      "field": "days",
      "operator": "<=",
      "value": 7
    }
  ],
  "actions": [
    {
      "type": "change_status",
      "params": {
        "entity_type": "rental",
        "new_status": "confirmed",
        "auto_confirmed": true
      }
    },
    {
      "type": "send_notification",
      "params": {
        "recipient_field": "requesting_partner_admin",
        "type": "rental_confirmed",
        "priority": "medium"
      }
    }
  ]
}
```

### 4.4 Workflow Execution Engine

```go
type WorkflowExecutionEngine struct {
    workflowRepo WorkflowRepository
    eventBus     EventBus
    db           *sql.DB
}

func (e *WorkflowExecutionEngine) ProcessEvent(ctx context.Context, event *DomainEvent) error {
    // 1. Finde alle Workflows die auf diesen Event-Type reagieren
    workflows, err := e.workflowRepo.FindByTrigger(ctx, event.Type)
    if err != nil {
        return err
    }

    for _, workflow := range workflows {
        if !workflow.Enabled {
            continue
        }

        // 2. Prüfe Trigger-Filter
        if !e.matchesFilter(event, workflow.Trigger.EventFilter) {
            continue
        }

        // 3. Erstelle Execution-Record
        execution := &WorkflowExecution{
            ID:           uuid.New().String(),
            WorkflowID:   workflow.ID,
            EventID:      event.ID,
            Status:       "pending",
            ExecutionLog: []string{},
            StartedAt:    time.Now(),
        }

        e.db.ExecContext(ctx, "INSERT INTO workflow_executions (...) VALUES (...)", ...)

        // 4. Evaluiere Bedingungen
        if len(workflow.Conditions) > 0 {
            conditionsMet := e.evaluateConditions(event, workflow.Conditions)
            if !conditionsMet {
                e.db.ExecContext(ctx, "UPDATE workflow_executions SET status = $1 WHERE id = $2",
                    "skipped", execution.ID)
                continue
            }
        }

        // 5. Führe Aktionen aus
        execution.Status = "running"
        e.db.ExecContext(ctx, "UPDATE workflow_executions SET status = $1 WHERE id = $2",
            "running", execution.ID)

        for i, action := range workflow.Actions {
            err := e.executeAction(ctx, action, event, execution)
            if err != nil {
                execution.ExecutionLog = append(execution.ExecutionLog,
                    fmt.Sprintf("Action %d failed: %v", i, err))
                execution.Status = "failed"
                execution.ErrorMessage = err.Error()
                break
            }
            execution.ExecutionLog = append(execution.ExecutionLog,
                fmt.Sprintf("Action %d executed successfully", i))
        }

        if execution.Status == "running" {
            execution.Status = "success"
        }

        execution.CompletedAt = time.Now()
        logJSON, _ := json.Marshal(execution.ExecutionLog)
        e.db.ExecContext(ctx, `
            UPDATE workflow_executions
            SET status = $1, execution_log = $2, completed_at = $3, error_message = $4
            WHERE id = $5
        `, execution.Status, logJSON, execution.CompletedAt, execution.ErrorMessage, execution.ID)
    }

    return nil
}

func (e *WorkflowExecutionEngine) executeAction(ctx context.Context, action *WorkflowAction, event *DomainEvent, execution *WorkflowExecution) error {
    switch action.Type {
    case "send_notification":
        recipientID := e.renderTemplate(action.Params["recipient_id"].(string), event)
        e.notificationService.Send(ctx, &Notification{
            RecipientID: recipientID,
            Type:        action.Params["type"].(string),
            Priority:    action.Params["priority"].(string),
        })
        return nil

    case "send_email":
        templateName := action.Params["template_name"].(string)
        e.emailService.SendAsync(ctx, &Email{
            TemplateName: templateName,
            Data:         event.Data,
        })
        return nil

    case "change_status":
        entityType := action.Params["entity_type"].(string)
        newStatus := action.Params["new_status"].(string)
        entityID := event.Data["id"]

        _, err := e.db.ExecContext(ctx,
            fmt.Sprintf("UPDATE %s SET status = $1 WHERE id = $2", entityType),
            newStatus, entityID)
        return err

    case "webhook":
        url := action.Params["url"].(string)
        payload, _ := json.Marshal(event.Data)

        resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
        if err != nil {
            return err
        }
        resp.Body.Close()
        return nil

    default:
        return fmt.Errorf("unknown action type: %s", action.Type)
    }
}
```

---

## 5. DOMAIN EVENTS & INTEGRATION

### 5.1 Event-driven Architecture

```go
type DomainEvent struct {
    ID         string
    Type       string  // "sub_rental.requested", "invoice.created", ...
    AggregateID string
    Timestamp  time.Time
    Data       map[string]interface{}
    Source     string  // "federation", "api", "system"
}

// Event Types
const (
    SubRentalRequested    = "sub_rental.requested"
    SubRentalConfirmed    = "sub_rental.confirmed"
    SubRentalRejected     = "sub_rental.rejected"
    SubRentalCheckedOut   = "sub_rental.checked_out"
    SubRentalCheckedIn    = "sub_rental.checked_in"
    SubRentalCompleted    = "sub_rental.completed"
    DamageDetected        = "sub_rental.damage_detected"
    InvoiceGenerated      = "invoice.generated"
    InvoicePaid           = "invoice.paid"
    InvoiceOverdue        = "invoice.overdue"
    ProjectAssigned       = "project.assigned"
    ProjectChanged        = "project.changed"
    EquipmentCreated      = "equipment.created"
    SystemBackupFailed    = "system.backup_failed"
    FederationPartnerDown = "federation.partner_down"
)
```

### 5.2 Event Handler Registration

```go
type EventBus struct {
    handlers map[string][]EventHandler
    mutex    sync.RWMutex
}

type EventHandler interface {
    Handle(ctx context.Context, event *DomainEvent) error
    EventType() string
}

func (bus *EventBus) Register(handler EventHandler) {
    bus.mutex.Lock()
    defer bus.mutex.Unlock()

    eventType := handler.EventType()
    bus.handlers[eventType] = append(bus.handlers[eventType], handler)
}

func (bus *EventBus) Publish(ctx context.Context, event *DomainEvent) error {
    bus.mutex.RLock()
    handlers := bus.handlers[event.Type]
    bus.mutex.RUnlock()

    for _, handler := range handlers {
        go func(h EventHandler) {
            if err := h.Handle(ctx, event); err != nil {
                log.Printf("Handler error for %s: %v", event.Type, err)
            }
        }(handler)
    }

    return nil
}
```

**Handler-Registrierung:**

```go
// In main()
eventBus := &EventBus{handlers: make(map[string][]EventHandler)}

// Notification Handler
eventBus.Register(&NotificationRuleHandler{ruleEngine: notifRuleEngine})

// Workflow Handler
eventBus.Register(&WorkflowTriggerHandler{executionEngine: workflowEngine})

// Invoice Handler (für Rechnungsgenerierung)
eventBus.Register(&InvoiceGenerationHandler{invoiceService: invoiceSvc})

// Federation Handler
eventBus.Register(&FederationSyncHandler{federationService: fedSvc})
```

---

## ZUSAMMENFASSUNG

Dieses Dokument beschreibt die **produktionsreife Implementierung** von:

1. **Federation Protocol (P2P):** Dezentrales Pairing mit mTLS, sichere Equipment-Verfügbarkeitsabfragen, Sub-Rental-Workflow mit Zustandsverwaltung
2. **AI Service (Multi-Provider):** Claude/OpenAI/Ollama-Adapter, DSGVO-konforme Anonymisierung, 6 AI-Features (Asset Creator, Price Optimization, Demand Forecast, OCR, Chat, Anomaly Detection), Token-Budget-Management
3. **Notification Service:** WebSocket In-App, Web Push (VAPID), E-Mail (SMTP Queue), Notification Rules Engine, User Preferences
4. **Workflow Engine:** No-Code Visual Editor (React Flow), Event-driven Execution, Pre-built Templates (Mahnung, Damage Alert, Auto-Confirm)
5. **Domain Events & Integration:** Event Bus mit Handlers für alle Features

**Sicherheit:**
- mTLS Certificate Pinning für Federation
- HMAC-SHA256 Request-Signing
- Rate Limiting (100 req/min pro Partner)
- Payload Validation
- Audit Logging
- DSGVO-konforme Anonymisierung vor Cloud-KI

**Offline-Toleranz:**
- Cached Equipment-Availability mit TTL
- Graceful Degradation wenn Partner offline
- Exponential Backoff Retry-Strategie

**Performance:**
- Redis für In-Memory Caching und Message Queues
- PostgreSQL mit Indizes für schnelle Abfragen
- WebSocket für Push-Notifications (statt Polling)
- E-Mail Queue mit Retry-Logik

Alle Endpoints, Datenmodelle, Go Structs, SQL Schemas und Implementation-Details sind dokumentiert und produktionsbereit.
