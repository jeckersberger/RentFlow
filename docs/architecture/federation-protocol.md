# CrateDesk Federation Protocol

## 1. Protokoll-Übersicht

### Architektur: Dezentrales Peer-to-Peer System

CrateDesk Federation ist ein **dezentrales, serverlos Protokoll** für die Vernetzung unabhängiger Lagerverwaltungs-Instanzen. Jede VT-Firma behält ihre komplette Kontrolle:

- **Eigener Server, eigene Daten:** Alle Equipment- und Geschäftsdaten bleiben lokal
- **Keine zentrale Plattform:** Keine Abhängigkeit von Drittanbietern
- **Direkte Peer-to-Peer Kommunikation:** Instanz A spricht direkt mit Instanz B
- **Autonome Kontrollmöglichkeit:** Jeder Admin entscheidet, was geteilt wird und unter welchen Bedingungen
- **Graceful Degradation:** Partner-Ausfälle beeinflussen nicht die eigene Instanz

### Prinzipien

1. **Datensparsamkeit:** Nur das absolut notwendige wird übertragen
2. **Autonomie:** Kein Partner kann eine andere Instanz kontrollieren
3. **Transparenz:** Alle Federation-Aktivitäten sind auditierbar
4. **Einfachheit:** Admin-Workflows sind nicht-technisch gestaltet
5. **Sicherheit:** mTLS + Request-Signing, kein zentraler Token-Server

---

## 2. Pairing-Prozess

### Schritt-für-Schritt: Zwei Instanzen vernetzen sich

#### Phase 1: Token-Generierung (Admin Interface)

**Bei Stefan (SoundPro):**

1. Admin öffnet "Einstellungen > Federation > Neuen Partner hinzufügen"
2. System generiert **Pairing Token** (zeitlich limitiert, 24h gültig)
   ```
   PAIRING_TOKEN=ST-9f2e8d1c-4a7b-11ef-9e3a-0242ac130003:exp:2026-03-21T14:30:00Z:sig:base64sig
   ```
3. Admin sieht QR-Code + Textfeld mit Token
4. Token wird **nicht persistent gespeichert** – nur temporär in Redis (TTL 24h)

#### Phase 2: Token-Austausch (Manuell)

1. Stefan kopiert Token (oder scannt QR-Code mit Handy)
2. Stefan sendet Token an Marco (per Email, Signal, etc.)
3. Marco öffnet seine Instanz: "Einstellungen > Federation > Partner verbinden"
4. Marco fügt Stefans Token ein
5. System validiert Token + startet Zertifikats-Austausch

#### Phase 3: Zertifikats-Discovery & Austausch

**Marco sendet HTTP POST an Stefans Discovery-Endpoint:**

```http
POST https://stefan-soundpro.com/.well-known/cratedesk-federation/pair
Content-Type: application/json
X-CrateDesk-Pairing-Token: ST-9f2e8d1c-4a7b-11ef-9e3a-0242ac130003
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b

{
  "instance_id": "marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b",
  "instance_name": "Marco Veranstaltungstechnik",
  "base_url": "https://marco-events.com",
  "certificate_public_key": "-----BEGIN CERTIFICATE-----\nMIIDXTCC...\n-----END CERTIFICATE-----",
  "api_version": "1.0"
}
```

**Stefan antwortet mit eigenen Zertifikat + bestätigt Pairing:**

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "paired",
  "instance_id": "stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002",
  "instance_name": "SoundPro Audio Rental",
  "certificate_public_key": "-----BEGIN CERTIFICATE-----\nMIIDXTCC...\n-----END CERTIFICATE-----",
  "paired_at": "2026-03-20T14:30:00Z",
  "federation_url": "https://stefan-soundpro.com/api/v1/federation"
}
```

#### Phase 4: mTLS-Setup & Verifizierung

Nach erfolgreicher Pairing werden beide Zertifikate in der lokalen Datenbank gespeichert mit **Certificate Pinning:**

**In Stefan's DB (Partner Marco):**
```sql
INSERT INTO federation_partners (
  partner_id,
  instance_id,
  base_url,
  certificate_fingerprint,  -- SHA256
  certificate_public_key,
  status,
  paired_at
) VALUES (
  'marco-veranstaltung',
  '5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b',
  'https://marco-events.com',
  'sha256:a3f8e2c1d9...',
  '-----BEGIN CERTIFICATE-----...',
  'active',
  '2026-03-20T14:30:00Z'
);
```

**Test-Handshake:** Marco sendet signed HTTP GET an Stefan:

```http
GET /api/v1/federation/health HTTP/1.1
Host: stefan-soundpro.com
X-CrateDesk-Instance-Id: marco-veranstaltung
X-CrateDesk-Timestamp: 2026-03-20T14:31:00Z
X-CrateDesk-Signature: sha256=base64encodedsig
X-CrateDesk-Client-Cert-Fingerprint: sha256:a3f8e2c1d9...
```

**Stefan verifiziert:**
1. mTLS-Handshake erfolgt (Zertifikat matcht pinned fingerprint)
2. Request-Signatur validiert
3. Antwortet mit `{"status": "healthy", "federation_enabled": true}`

✅ **Pairing abgeschlossen!**

---

## 3. mTLS-Setup

### Automatische Zertifikatsgenerierung

**Bei Installation jede Instanz generiert selbst:**

1. **Self-Signed Root CA** (pro Instanz, einmalig):
   ```
   Subject: CN=CrateDesk CA (stefan-soundpro)
   Issuer: (self-signed)
   Validity: 10 years
   Key Size: 4096-bit RSA
   Purpose: CA for federation certificates
   ```

2. **Server-Zertifikat** (automatisch generiert):
   ```
   Subject: CN=stefan-soundpro.com
   Issuer: CrateDesk CA (stefan-soundpro)
   Validity: 1 year (auto-renewal 30 days vor Ablauf)
   Key Size: 4096-bit RSA
   SANs: stefan-soundpro.com, *.stefan-soundpro.com
   Extended Key Usage: TLS Web Server Authentication, TLS Web Client Authentication
   ```

3. **Speicherung:**
   ```
   /opt/cratedesk/certs/
   ├── ca.crt              (Root CA certificate)
   ├── ca.key              (Root CA private key - protected)
   ├── server.crt          (Server certificate)
   ├── server.key          (Server private key - protected)
   └── federation/
       ├── marco_ca.crt    (Marco's Root CA - pinned)
       └── ...
   ```

### Certificate Pinning

**Stefans System speichert Marcos Root CA Zertifikat:**

```go
type PinnedCertificate struct {
    PartnerID              string    // marco-veranstaltung
    InstanceID             string    // 5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
    RootCAFingerprint      string    // SHA256(marcos-ca.crt)
    CertificateChain       string    // PEM-encoded
    PinnedAt               time.Time
    ValidUntil             time.Time
}
```

**mTLS-Handshake Validierung:**

```go
func ValidateMTLSPeer(ctx context.Context, cert *x509.Certificate, partnerID string) error {
    // 1. Zertifikat nicht abgelaufen?
    if cert.NotAfter.Before(time.Now()) {
        return errors.New("certificate expired")
    }

    // 2. Von gepinntem Root CA signiert?
    pinnedCA := GetPinnedCACertificate(partnerID)
    err := cert.VerifySignatureFrom(pinnedCA)
    if err != nil {
        return fmt.Errorf("certificate not signed by pinned CA: %w", err)
    }

    // 3. Fingerprint korrekt?
    fingerprint := fmt.Sprintf("sha256:%s", SHA256Hex(cert.Raw))
    if fingerprint != pinnedCA.Fingerprint {
        return errors.New("fingerprint mismatch")
    }

    return nil
}
```

### Certificate Rotation & Auto-Renewal

- Instanzen überprüfen täglich, ob Zertifikate in 30 Tagen ablaufen
- Neues Zertifikat wird generiert (mit privatem Key)
- **Neue Zertifikate werden NICHT automatisch zu Partnern gepusht** – Partner erhalten beim nächsten Request die neuen Zertifikate via TLS
- Admin wird notifiziert wenn Zertifikat bald abläuft
- Alte Zertifikate können 30 Tage nach Ablauf manuell gelöscht werden

---

## 4. API-Endpunkte (Federation)

Alle Federation-Endpunkte liegen unter `/api/v1/federation/` und erfordern:
- mTLS mit gepinntem Zertifikat
- Signed Request (`X-CrateDesk-Signature` Header)
- Gültigen `X-CrateDesk-Instance-Id` Header

### 4.1 Discovery: `/.well-known/cratedesk-federation`

**Öffentlicher Endpunkt – KEINE mTLS erforderlich.**

```http
GET /.well-known/cratedesk-federation HTTP/1.1
Host: stefan-soundpro.com
```

**Response:**
```json
{
  "federation_enabled": true,
  "instance_id": "stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002",
  "instance_name": "SoundPro Audio Rental",
  "protocol_version": "1.0",
  "api_endpoint": "https://stefan-soundpro.com/api/v1/federation",
  "capabilities": [
    "equipment_catalog",
    "availability_query",
    "sub_rental",
    "handover",
    "invoicing"
  ],
  "root_ca_certificate": "-----BEGIN CERTIFICATE-----\nMIIBkTCB...\n-----END CERTIFICATE-----",
  "root_ca_fingerprint": "sha256:a3f8e2c1d9b4f6e8a2c5d7f9a1b3e5c7...",
  "supported_api_versions": ["1.0"],
  "max_lookahead_days": 30,
  "contact_email": "federation@stefan-soundpro.com"
}
```

### 4.2 Equipment-Katalog

#### 4.2.1 List Freigegebene Kategorien

```http
GET /api/v1/federation/equipment/categories HTTP/1.1
Host: stefan-soundpro.com
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-CrateDesk-Timestamp: 2026-03-20T14:31:00Z
X-CrateDesk-Signature: sha256=base64sig
```

**Response:**
```json
{
  "categories": [
    {
      "id": "cat-audio-001",
      "name": "Audio - Mischpulte",
      "equipment_count": 12,
      "can_query_availability": true,
      "pricing_model": "hourly_rate"
    },
    {
      "id": "cat-audio-002",
      "name": "Audio - Mikrofone",
      "equipment_count": 45,
      "can_query_availability": true,
      "pricing_model": "hourly_rate"
    },
    {
      "id": "cat-lights-001",
      "name": "Beleuchtung - LED",
      "equipment_count": 8,
      "can_query_availability": false,
      "reason": "Configured as unavailable for this partner"
    }
  ]
}
```

#### 4.2.2 List Equipment einer Kategorie

```http
GET /api/v1/federation/equipment/categories/cat-audio-001/items HTTP/1.1
Host: stefan-soundpro.com
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-CrateDesk-Timestamp: 2026-03-20T14:31:00Z
X-CrateDesk-Signature: sha256=base64sig
```

**Response:**
```json
{
  "category_id": "cat-audio-001",
  "category_name": "Audio - Mischpulte",
  "equipment": [
    {
      "id": "eq-mix-001",
      "name": "Soundcraft Si Impact 32",
      "description": "Digital mixing console, 32 channels",
      "quantity_available": 2,
      "base_rental_price": {
        "amount": 150.00,
        "currency": "EUR",
        "per_unit": "day"
      },
      "partner_discount_percent": -15,
      "final_price": {
        "amount": 127.50,
        "currency": "EUR",
        "per_unit": "day"
      },
      "conditions": {
        "min_rental_days": 1,
        "deposit_required": true,
        "delivery_available": true
      }
    },
    {
      "id": "eq-mix-002",
      "name": "Yamaha CL5",
      "description": "Digital mixing console, 32 channels",
      "quantity_available": 1,
      "base_rental_price": {
        "amount": 200.00,
        "currency": "EUR",
        "per_unit": "day"
      },
      "partner_discount_percent": -15,
      "final_price": {
        "amount": 170.00,
        "currency": "EUR",
        "per_unit": "day"
      },
      "conditions": {
        "min_rental_days": 1,
        "deposit_required": true,
        "delivery_available": true
      }
    }
  ]
}
```

**Wichtig:** Preise sind NIEMALS hart-kodiert. Sie beziehen sich auf Partner-spezifische Konfiguration (siehe Kapitel 6).

### 4.3 Verfügbarkeit

#### 4.3.1 Verfügbarkeitsabfrage

```http
POST /api/v1/federation/availability/query HTTP/1.1
Host: stefan-soundpro.com
Content-Type: application/json
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-CrateDesk-Timestamp: 2026-03-20T14:31:00Z
X-CrateDesk-Signature: sha256=base64sig

{
  "queries": [
    {
      "equipment_id": "eq-mix-001",
      "from_date": "2026-04-01T09:00:00Z",
      "to_date": "2026-04-03T17:00:00Z",
      "quantity_required": 1
    },
    {
      "equipment_id": "eq-mix-002",
      "from_date": "2026-04-01T09:00:00Z",
      "to_date": "2026-04-03T17:00:00Z",
      "quantity_required": 1
    }
  ]
}
```

**Response:**
```json
{
  "query_id": "aq-5f8d9c2e-3b4a-11ef-a5c7-0242ac130005",
  "results": [
    {
      "equipment_id": "eq-mix-001",
      "requested_quantity": 1,
      "available_quantity": 2,
      "is_available": true,
      "availability_periods": [
        {
          "from": "2026-04-01T09:00:00Z",
          "to": "2026-04-03T17:00:00Z",
          "quantity_available": 2,
          "status": "available"
        }
      ]
    },
    {
      "equipment_id": "eq-mix-002",
      "requested_quantity": 1,
      "available_quantity": 0,
      "is_available": false,
      "availability_periods": [
        {
          "from": "2026-04-01T09:00:00Z",
          "to": "2026-04-02T14:00:00Z",
          "quantity_available": 0,
          "status": "reserved"
        },
        {
          "from": "2026-04-02T14:00:00Z",
          "to": "2026-04-03T17:00:00Z",
          "quantity_available": 1,
          "status": "available"
        }
      ]
    }
  ],
  "timestamp": "2026-03-20T14:31:00Z",
  "expires_at": "2026-03-20T14:46:00Z"
}
```

**Begrenzungen:**
- Max. 30 Tage Lookahead (konfigurierbar pro Partner, Min 7 Tage)
- Max. 100 Equipment pro Query
- Query-Result gültig 15 Minuten (dann muss neu abgefragt werden)

### 4.4 Sub-Rental: Anfrage, Status, Bestätigung

#### 4.4.1 Sub-Rental-Anfrage senden

```http
POST /api/v1/federation/sub-rentals HTTP/1.1
Host: stefan-soundpro.com
Content-Type: application/json
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-CrateDesk-Timestamp: 2026-03-20T14:35:00Z
X-CrateDesk-Signature: sha256=base64sig

{
  "sub_rental_id": "sr-marco-001-2026-04-01",
  "requesting_partner_id": "marco-veranstaltung",
  "requested_equipment": [
    {
      "equipment_id": "eq-mix-001",
      "quantity": 1,
      "rental_from": "2026-04-01T09:00:00Z",
      "rental_to": "2026-04-03T17:00:00Z"
    }
  ],
  "customer": {
    "name": "Marco's Veranstaltung GmbH",
    "contact_email": "rentals@marco-events.com",
    "phone": "+49 30 1234567"
  },
  "event_details": {
    "event_name": "Großkonzert 2026",
    "event_date": "2026-04-02",
    "event_location": "Berlin"
  },
  "delivery": {
    "type": "pickup",
    "location": "Stefan's Studio, Berlin",
    "scheduled_time": "2026-04-01T08:00:00Z"
  },
  "special_conditions": "Please ensure console is pre-programmed for rock concert setup.",
  "auto_confirm_requested": false
}
```

**Response:**
```json
{
  "status": "pending",
  "sub_rental_id": "sr-marco-001-2026-04-01",
  "server_sub_rental_id": "sr-stefan-2026-03-20-001",
  "created_at": "2026-03-20T14:35:00Z",
  "requires_manual_confirmation": true,
  "confirmation_deadline": "2026-03-20T18:35:00Z",
  "estimated_total_price": {
    "amount": 382.50,
    "currency": "EUR",
    "breakdown": {
      "rental_days": 2.5,
      "daily_rate": 127.50,
      "delivery_fee": 50.00,
      "insurance": 25.00
    }
  }
}
```

#### 4.4.2 Sub-Rental Status abfragen

```http
GET /api/v1/federation/sub-rentals/sr-stefan-2026-03-20-001 HTTP/1.1
Host: stefan-soundpro.com
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-CrateDesk-Timestamp: 2026-03-20T14:40:00Z
X-CrateDesk-Signature: sha256=base64sig
```

**Response:**
```json
{
  "sub_rental_id": "sr-stefan-2026-03-20-001",
  "partner_sub_rental_id": "sr-marco-001-2026-04-01",
  "status": "pending",
  "status_history": [
    {
      "status": "pending",
      "timestamp": "2026-03-20T14:35:00Z",
      "reason": "Awaiting manual confirmation"
    }
  ],
  "equipment": [
    {
      "equipment_id": "eq-mix-001",
      "name": "Soundcraft Si Impact 32",
      "quantity": 1,
      "rental_from": "2026-04-01T09:00:00Z",
      "rental_to": "2026-04-03T17:00:00Z"
    }
  ],
  "total_price": {
    "amount": 382.50,
    "currency": "EUR"
  },
  "next_action": "awaiting_confirmation"
}
```

#### 4.4.3 Sub-Rental Bestätigung/Ablehnung

```http
POST /api/v1/federation/sub-rentals/sr-stefan-2026-03-20-001/confirm HTTP/1.1
Host: marco-events.com
Content-Type: application/json
X-CrateDesk-Instance-Id: stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002
X-CrateDesk-Timestamp: 2026-03-20T15:00:00Z
X-CrateDesk-Signature: sha256=base64sig

{
  "confirmed": true,
  "confirmed_by": "admin@stefan-soundpro.de",
  "delivery_note": "Console will be pre-programmed. Contact: Stefan +49 30 9876543"
}
```

**Response:**
```json
{
  "sub_rental_id": "sr-stefan-2026-03-20-001",
  "status": "confirmed",
  "confirmed_at": "2026-03-20T15:00:00Z",
  "confirmed_by_instance": "stefan-soundpro",
  "final_price": {
    "amount": 382.50,
    "currency": "EUR"
  },
  "handover_checklist_url": "https://stefan-soundpro.com/handover/sr-stefan-2026-03-20-001"
}
```

**Oder Ablehnung:**
```http
POST /api/v1/federation/sub-rentals/sr-stefan-2026-03-20-001/reject HTTP/1.1

{
  "rejected": true,
  "rejected_by": "admin@marco-events.com",
  "reason": "Equipment already booked for same period",
  "suggested_alternative_dates": [
    {
      "from": "2026-04-05T09:00:00Z",
      "to": "2026-04-07T17:00:00Z"
    }
  ]
}
```

### 4.5 Handover: Zustandsdokumentation

#### 4.5.1 Handover starten (Pickup)

```http
POST /api/v1/federation/sub-rentals/sr-stefan-2026-03-20-001/handover/start HTTP/1.1
Host: stefan-soundpro.com
Content-Type: application/json
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-CrateDesk-Timestamp: 2026-04-01T07:45:00Z
X-CrateDesk-Signature: sha256=base64sig

{
  "handover_type": "pickup",
  "location": "Stefan's Studio, Berlin",
  "handover_at": "2026-04-01T08:00:00Z",
  "performed_by": {
    "name": "Marco Techniker",
    "role": "technician",
    "contact": "marco@marco-events.com"
  }
}
```

**Response:**
```json
{
  "handover_id": "hnd-sr-stefan-2026-03-20-001-pickup",
  "sub_rental_id": "sr-stefan-2026-03-20-001",
  "handover_type": "pickup",
  "status": "in_progress",
  "created_at": "2026-04-01T07:45:00Z",
  "inspection_checklist": [
    {
      "item_id": "chk-001",
      "item_name": "Console - Power Supply",
      "required": true
    },
    {
      "item_id": "chk-002",
      "item_name": "Console - Network Cable",
      "required": true
    },
    {
      "item_id": "chk-003",
      "item_name": "Console - XLR Cables (5x)",
      "required": true
    }
  ],
  "document_upload_url": "https://stefan-soundpro.com/api/v1/federation/handover/hnd-sr-stefan-2026-03-20-001-pickup/documents"
}
```

#### 4.5.2 Handover-Dokumentation hochladen

```http
POST /api/v1/federation/handover/hnd-sr-stefan-2026-03-20-001-pickup/documents HTTP/1.1
Host: stefan-soundpro.com
Content-Type: multipart/form-data
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
X-CrateDesk-Timestamp: 2026-04-01T08:15:00Z
X-CrateDesk-Signature: sha256=base64sig

--boundary
Content-Disposition: form-data; name="documents"
Content-Type: application/json

{
  "checklist_completion": [
    {
      "item_id": "chk-001",
      "status": "ok",
      "notes": "Power supply working correctly"
    },
    {
      "item_id": "chk-002",
      "status": "ok",
      "notes": ""
    },
    {
      "item_id": "chk-003",
      "status": "ok",
      "notes": "5x XLR cables present and tested"
    }
  ],
  "general_condition": "Excellent",
  "notes": "Console fully functional, pre-programmed as requested"
}
--boundary
Content-Disposition: form-data; name="photo_pickup_overall"
Content-Type: image/jpeg

[BINARY JPG DATA]
--boundary
Content-Disposition: form-data; name="photo_display"
Content-Type: image/jpeg

[BINARY JPG DATA]
--boundary
Content-Disposition: form-data; name="digital_signature"
Content-Type: application/json

{
  "timestamp": "2026-04-01T08:20:00Z",
  "signer_name": "Marco Techniker",
  "signature_data": "base64encodedsignature"
}
--boundary--
```

**Response:**
```json
{
  "handover_id": "hnd-sr-stefan-2026-03-20-001-pickup",
  "status": "completed",
  "completed_at": "2026-04-01T08:20:00Z",
  "documents_uploaded": 3,
  "signature_verified": true,
  "return_handover_url": "https://marco-events.com/handover/hnd-sr-stefan-2026-03-20-001-return"
}
```

#### 4.5.3 Return-Handover

```http
POST /api/v1/federation/sub-rentals/sr-stefan-2026-03-20-001/handover/return HTTP/1.1
Host: marco-events.com
Content-Type: multipart/form-data
X-CrateDesk-Instance-Id: stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002
X-CrateDesk-Timestamp: 2026-04-03T18:00:00Z
X-CrateDesk-Signature: sha256=base64sig

[... similar document upload structure ...]
```

**Response von Marco an Stefan:**
```json
{
  "handover_id": "hnd-sr-stefan-2026-03-20-001-return",
  "sub_rental_id": "sr-stefan-2026-03-20-001",
  "status": "completed",
  "return_condition": "Good",
  "damage_report": [],
  "returned_at": "2026-04-03T18:00:00Z",
  "invoice_ready": true,
  "invoice_url": "https://marco-events.com/api/v1/federation/invoices/inv-2026-03-20-001"
}
```

### 4.6 Invoicing: Automatische Rechnungsdaten

#### 4.6.1 Invoice abrufen

```http
GET /api/v1/federation/invoices/inv-2026-03-20-001 HTTP/1.1
Host: marco-events.com
X-CrateDesk-Instance-Id: stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002
X-CrateDesk-Timestamp: 2026-03-20T18:30:00Z
X-CrateDesk-Signature: sha256=base64sig
```

**Response:**
```json
{
  "invoice_id": "inv-2026-03-20-001",
  "sub_rental_id": "sr-stefan-2026-03-20-001",
  "invoice_number": "2026-001",
  "issuer": {
    "company": "Marco Veranstaltungstechnik",
    "instance_id": "marco-veranstaltung",
    "tax_id": "DE123456789"
  },
  "recipient": {
    "company": "SoundPro Audio Rental",
    "instance_id": "stefan-soundpro",
    "tax_id": "DE987654321"
  },
  "invoice_date": "2026-04-03",
  "due_date": "2026-04-10",
  "line_items": [
    {
      "description": "Soundcraft Si Impact 32 - Rental",
      "quantity": 1,
      "unit": "day",
      "unit_price": 127.50,
      "total_days": 2.5,
      "amount": 318.75,
      "tax_rate": 0.19,
      "tax_amount": 60.56,
      "line_total": 379.31
    },
    {
      "description": "Delivery Service",
      "quantity": 1,
      "unit": "service",
      "unit_price": 50.00,
      "amount": 50.00,
      "tax_rate": 0.19,
      "tax_amount": 9.50,
      "line_total": 59.50
    },
    {
      "description": "Equipment Insurance",
      "quantity": 1,
      "unit": "service",
      "unit_price": 25.00,
      "amount": 25.00,
      "tax_rate": 0.19,
      "tax_amount": 4.75,
      "line_total": 29.75
    }
  ],
  "subtotal": 393.75,
  "tax_total": 74.81,
  "total": 468.56,
  "currency": "EUR",
  "payment_terms": "Net 7 days",
  "status": "issued",
  "invoice_pdf_url": "https://marco-events.com/api/v1/federation/invoices/inv-2026-03-20-001/pdf"
}
```

#### 4.6.2 Invoice als PDF abrufen

```http
GET /api/v1/federation/invoices/inv-2026-03-20-001/pdf HTTP/1.1
Host: marco-events.com
X-CrateDesk-Instance-Id: stefan-soundpro:3b2d8f1a-7c4e-11ef-8b9d-0242ac110002
X-CrateDesk-Timestamp: 2026-03-20T18:35:00Z
X-CrateDesk-Signature: sha256=base64sig
```

**Response:** PDF binary mit Content-Type `application/pdf`

---

## 5. Datenmodell: Was wird geteilt, was nicht?

### Grundprinzip: Datensparsamkeit

Jede Instanz exponiert nur das **absolut notwendige Minimum**. Partner sehen NIEMALS:
- Interne Preiskalkulationen (nur finales Preis-Modell pro Partner)
- Kundendetails (nur aggregierte Verfügbarkeit)
- Lagerverwaltungs-Details (Seriennummern, Reparaturhistorie, etc.)
- Interne Buchungssystem-Struktur
- Compliance/Audit-Daten

### Was WIRD geteilt?

| Entity | Partner sieht | Stefan behält privat |
|--------|---------------|--------------------|
| **Kategorie** | Name, ID, Equipment-Count | Interne Struktur, alle SKUs |
| **Equipment** | Name, Beschreibung, Zustand, Partner-Preis | Einkaufspreis, Seriennummer, Reparaturhistorie |
| **Verfügbarkeit** | Verfügbar/Nicht-Verfügbar (Pro Partner konfigurierbar) | Warum nicht verfügbar, Intern-Details |
| **Sub-Rental** | Anfrage-Status, bestätigte Daten, Preisberechnung | Interne Kosten, Profitability-Metriken |
| **Handover-Docs** | Checklist-Status, Schäden (wenn vorhanden), Fotos | Interne Qualitätsmessungen |
| **Invoice** | Detaillierte Abrechnung (MUSS genau sein) | Interne Rechnungskategorien, Gewinn |

### Datenmodell: Federation-relevante Entities

```
federation_partners:
  - partner_id: marco-veranstaltung
  - instance_id: 5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
  - base_url: https://marco-events.com
  - certificate_fingerprint: sha256:a3f8e2c1d9...
  - certificate_public_key: -----BEGIN CERTIFICATE-----
  - paired_at: 2026-03-20T14:30:00Z
  - status: active

partner_configurations:
  - partner_id: marco-veranstaltung
  - shared_categories: [cat-audio-001, cat-audio-002]
  - pricing_model: partner_discount
  - discount_percent: -15
  - min_lookahead_days: 7
  - max_lookahead_days: 30
  - auto_confirm_rentals: false
  - auto_confirm_threshold_eur: 500.00
  - allowed_delivery_types: [pickup, delivery]
  - api_rate_limit_requests_per_hour: 600

federation_requests_audit:
  - id: uuid
  - partner_id: marco-veranstaltung
  - request_method: GET
  - request_path: /api/v1/federation/availability/query
  - response_status: 200
  - timestamp: 2026-03-20T14:31:00Z
  - signature_valid: true
  - mtls_certificate_fingerprint: sha256:a3f8e2c1d9...
  - request_size_bytes: 412
  - response_size_bytes: 1847
```

---

## 6. Konfigurations-Modell

### Was kann der Admin pro Partner einstellen?

**Admin Interface: Settings > Federation > Partner "Marco Veranstaltung"**

#### 6.1 Sharing & Visibility

```
✓ Kategorien freigeben
  ☑ Audio - Mischpulte
  ☑ Audio - Mikrofone
  ☐ Beleuchtung - LED
  ☐ Video - Kameras

ⓘ "LED-Beleuchtung ist zu wertvoll, möchte ich nicht teilen"
```

#### 6.2 Pricing Model

```
Preismodell für Marco:
  ◎ Base Price (keine Anpassung)
  ◎ Fixed Discount: -15%
    └─ Alle Mietpreise werden um 15% reduziert

  ◎ Custom Per-Category:
    ├─ Audio: -20%
    ├─ Video: -10%
    └─ Sonstiges: -5%

  ◎ Markup (falls Marco über uns re-rented): +30%

Beispiel Berechnung:
  Base Rate: €150/Tag
  Marco sieht: €150 × (1 - 0.15) = €127.50/Tag
```

#### 6.3 Booking & Auto-Confirmation

```
Anfrage-Management:

  ◎ Manuelle Bestätigung (Default)
    ├─ Ich muss jede Anfrage manuel bestätigen
    └─ Frist: 4 Stunden

  ◎ Auto-Confirm unter Schwelle:
    ├─ Auto-Bestätigung wenn < €500
    └─ Größere Anfragen brauchen manuelle Genehmigung

  ◎ Vollautomatisch
    └─ Alle Anfragen werden sofort bestätigt

Benachrichtigungen:
  ☑ Email bei neuer Anfrage
  ☑ Email bei Bestätigung durch Partner
  ☐ SMS Alerts
```

#### 6.4 Verfügbarkeit & Lookback-Fenster

```
Verfügbarkeit-Konfiguration:

  Maximales Voraus-Fenster:
    [30] Tage (Min: 7, Max: 90)
    ⓘ "Maximal 30 Tage in die Zukunft sichtbar"

  Verfügbarkeit-Caching:
    [15] Minuten
    ⓘ "Verfügbarkeitsabfragen werden 15 Min. gecacht"

  Private Fenster (Marco kann nicht abfragen):
    Von: [___________] Bis: [___________]
    ⓘ "Zeitfenster für eigene Buchungen (z.B. Weihnachtsmessen)"
```

#### 6.5 Lieferung & Logistics

```
Lieferoptionen für Marco:

  ☑ Pickup (Marco holt ab)
    └─ Ort: Stefan's Studio, Berlin

  ☑ Delivery Service (Stefan liefert)
    ├─ Gebühr: €50
    ├─ Lieferradius: 50 km
    └─ Mindestbestellung: €300

  ☐ Dropoff
```

#### 6.6 Rate Limiting & Quotas

```
API Limits pro Stunde:
  - Availability Queries: [600] (10/Min)
  - Sub-Rental Requests: [100]
  - Status Checks: [Unlimited]

Request Timeout:
  [30] Sekunden

Rollover (zu viele Requests):
  ◎ Block + Error 429
  ◎ Warn Email
```

#### 6.7 Data Retention

```
Historische Daten pro Partner:

  Sub-Rentals:     [6] Monate (dann archiviert)
  Invoices:        [10] Jahre (rechtlich erforderlich)
  Handover Docs:   [3] Jahre
  Audit Logs:      [1] Jahr

  Automatisches Löschen nach Ablauf:
  ☑ Aktiviert
```

### Konfiguration Speicherung

```sql
CREATE TABLE partner_config (
    partner_id VARCHAR(255) PRIMARY KEY,
    shared_category_ids JSON,           -- ["cat-001", "cat-002"]
    pricing_model VARCHAR(50),          -- 'base', 'discount', 'custom', 'markup'
    pricing_rules JSON,                 -- flexible: {discount_percent: -15} or per-category
    auto_confirm_enabled BOOLEAN,
    auto_confirm_threshold_eur DECIMAL(10,2),
    max_lookahead_days INT,
    private_date_windows JSON,          -- [{from: "2026-12-20", to: "2026-12-31"}]
    allowed_delivery_types JSON,        -- ["pickup", "delivery"]
    delivery_fee_eur DECIMAL(10,2),
    api_rate_limit_per_hour INT,
    data_retention_days INT,
    updated_at TIMESTAMP,
    updated_by VARCHAR(255)
);
```

---

## 7. Sicherheit

### 7.1 mTLS (Mutual TLS)

Alle Federation-Endpunkte (außer `/.well-known/...`) erzwingen:

1. **Server-Authentifizierung:** Client validiert Stefans Zertifikat gegen gepinntes Root-CA
2. **Client-Authentifizierung:** Server validiert Marcos Zertifikat gegen gepinntes Root-CA
3. **Fingerprint Validation:** Zusätzlich zur Cert-Chain-Validierung

```go
// Server-seitige Validierung (Stefan)
func EnforcePartnerMTLS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cert := r.TLS.PeerCertificates[0] // Client-Zertifikat (Marco)

        // 1. Zertifikat abgelaufen?
        if time.Now().After(cert.NotAfter) {
            http.Error(w, "Certificate expired", 401)
            return
        }

        // 2. Fingerprint in pinned certs?
        fingerprint := fmt.Sprintf("sha256:%s", SHA256Hex(cert.Raw))
        partner := FindPartnerByFingerprint(fingerprint)
        if partner == nil {
            http.Error(w, "Unknown partner certificate", 401)
            return
        }

        // 3. Von gepinntem Root-CA signiert?
        pinnedCA := GetPartnerRootCA(partner.ID)
        if err := cert.VerifySignatureFrom(pinnedCA); err != nil {
            http.Error(w, "Certificate not signed by trusted CA", 401)
            return
        }

        // OK - Request weitergeben
        ctx := context.WithValue(r.Context(), "partner_id", partner.ID)
        ctx = context.WithValue(ctx, "cert_fingerprint", fingerprint)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 7.2 Request-Signing (HMAC-SHA256)

Zusätzlich zu mTLS: Alle Requests müssen signiert sein.

**Signature Header Format:**
```
X-CrateDesk-Signature: sha256=base64encodedsignature
X-CrateDesk-Timestamp: 2026-03-20T14:31:00Z
X-CrateDesk-Instance-Id: marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
```

**Signatur-Berechnung (Marco unterschreibt für Stefan):**

```
1. Canonical Request aufbauen:
   METHOD HTTP_METHOD
   PATH /api/v1/federation/...
   TIMESTAMP 2026-03-20T14:31:00Z
   INSTANCE marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b
   BODY_HASH sha256(request_body)

2. Mit Marcos geheimen Schlüssel signieren (wird bei Pairing ausgetauscht):
   signature = HMAC-SHA256(canonical_request, marcos_secret_key)

3. Base64-kodiert als Header einfügen
```

**Stefan validiert:**
```go
func ValidateRequestSignature(r *http.Request, partner *Partner) error {
    body, _ := ioutil.ReadAll(r.Body)
    r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

    timestamp := r.Header.Get("X-CrateDesk-Timestamp")
    instanceID := r.Header.Get("X-CrateDesk-Instance-Id")
    signature := r.Header.Get("X-CrateDesk-Signature")

    // Signatur aus Request extrahieren
    sig := strings.TrimPrefix(signature, "sha256=")
    decodedSig, _ := base64.StdEncoding.DecodeString(sig)

    // Canonical Request nachbauen
    bodyHash := SHA256(body)
    canonical := fmt.Sprintf(
        "GET /api/v1/federation/...\n%s\n%s\n%s",
        timestamp, instanceID, bodyHash,
    )

    // Mit Marcos geheimem Schlüssel validieren
    expectedSig := HMAC_SHA256(canonical, partner.SecretKey)
    if !bytes.Equal(decodedSig, expectedSig) {
        return errors.New("invalid signature")
    }

    // Timestamp nicht älter als 5 Min?
    requestTime, _ := time.Parse(time.RFC3339, timestamp)
    if time.Since(requestTime) > 5*time.Minute {
        return errors.New("timestamp too old (replay attack?)")
    }

    return nil
}
```

### 7.3 Certificate Pinning + Rotation Handling

**Pinning Strategy:**
- Root CA Zertifikat wird gepinnt (nicht einzelne Server-Certs)
- Erlaubt automatische Rotation von Server-Zerts ohne Partner-Update
- Aber: Wenn Root CA wechselt (Notfall), muss manuell neu gepairt werden

**Was passiert wenn Marcos Root CA abläuft?**

1. 30 Tage vorher: System warnt Admin
2. Stefan kann mit altem Cert noch kommunizieren
3. Marco generiert neues Root CA + neue Server-Certs
4. Marco muss "Zertifikat erneuern" im UI klicken
5. System sendet neues Root CA zu Stefan
6. Stefan updatest Pinning
7. Beide Instanzen führen kurz-Handshake durch, um zu verifizieren

```sql
-- Zertifikat-Rotation Prozess
INSERT INTO federation_cert_rotation_requests (
    partner_id,
    request_type,
    old_fingerprint,
    new_fingerprint,
    status,
    created_at,
    expires_at
) VALUES (
    'marco-veranstaltung',
    'root_ca_renewal',
    'sha256:old_fingerprint',
    'sha256:new_fingerprint',
    'pending_confirmation',
    NOW(),
    NOW() + INTERVAL '24 hours'
);
```

### 7.4 Rate-Limiting

**Pro Partner, pro Stunde:**

```go
type RateLimiter struct {
    partnerID string
    limits map[string]int // endpoint -> requests/hour
}

func (rl *RateLimiter) Check(endpoint string) error {
    limit := rl.limits[endpoint] // z.B. 600 für /availability
    current := redis.Get(fmt.Sprintf("rate:%s:%s:%s", rl.partnerID, endpoint, currentHour))

    if current >= limit {
        return fmt.Errorf("rate limit exceeded: %d/%d", current, limit)
    }

    redis.Incr(...)
    return nil
}
```

**Response bei Überschreitung:**
```http
HTTP/1.1 429 Too Many Requests
Retry-After: 1200
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 2026-03-20T15:31:00Z
```

### 7.5 Audit-Logging

**Alle Federation-Requests werden protokolliert:**

```sql
CREATE TABLE federation_audit_log (
    id UUID PRIMARY KEY,
    partner_id VARCHAR(255),
    instance_id VARCHAR(255),
    cert_fingerprint VARCHAR(255),
    request_method VARCHAR(10),
    request_path VARCHAR(512),
    request_timestamp TIMESTAMP,
    response_status INT,
    response_time_ms INT,
    request_size_bytes INT,
    response_size_bytes INT,
    signature_valid BOOLEAN,
    signature_error VARCHAR(255),
    rate_limit_exceeded BOOLEAN,
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Audit-Log Beispiel
INSERT INTO federation_audit_log VALUES (
    '7b3a2f1e-5c9d-4e8a-1b2c-9d7f8e6a5c3d',
    'marco-veranstaltung',
    '5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b',
    'sha256:a3f8e2c1d9...',
    'POST',
    '/api/v1/federation/sub-rentals',
    '2026-03-20T14:35:00Z',
    200,
    147,
    412,
    1847,
    true,
    NULL,
    false,
    '192.0.2.42',
    'CrateDesk-Federation/1.0',
    NOW()
);
```

**Audit-Log Aufbewahrung:** 1 Jahr (konfigurierbar pro Partner)

### 7.6 Keine dauerhafte Speicherung von Partner-Daten

**Was wird NICHT persistent gespeichert:**

- Detaillierte API Response Bodies
- Persönliche Daten von Kunden (Name, Email, Phone) – nur Hash für Deduplication
- Verfügbarkeitsabfrage-Details nach 30 Tagen
- Handover-Fotos nach 3 Jahren (automatisch gelöscht)

**Was wird persistent gespeichert:**

- Sub-Rental Metadata (für Geschäftsprozesse)
- Invoice Daten (Rechtspflicht)
- Audit Logs (1 Jahr)
- Partner-Konfiguration (solange aktiv)

---

## 8. Fehlerbehandlung & Offline-Resilienz

### 8.1 Timeout & Retry-Strategie

**Timeout pro Request:**
```go
timeout := 30 * time.Second  // Konfigurierbar pro Partner
client := &http.Client{
    Timeout: timeout,
}
```

**Retry-Logik:**
```go
type RetryConfig struct {
    MaxRetries      int           // Default: 3
    BackoffBase     time.Duration // 1 Sekunde
    BackoffMultiplier float64      // 2x
    MaxBackoff      time.Duration // 32 Sekunden
}

func (r *RetryConfig) Retry(fn func() error) error {
    var lastErr error

    for attempt := 0; attempt < r.MaxRetries; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }

        lastErr = err

        if attempt < r.MaxRetries-1 {
            backoff := r.BackoffBase * time.Duration(math.Pow(r.BackoffMultiplier, float64(attempt)))
            if backoff > r.MaxBackoff {
                backoff = r.MaxBackoff
            }
            time.Sleep(backoff)
        }
    }

    return lastErr
}

// Verwendung:
errs := retryConfig.Retry(func() error {
    return marcoClient.QueryAvailability(ctx, query)
})
```

**Retry-Bedingungen:**
- ✅ Retry: Connection Timeout, Network Error, 5xx Server Errors
- ✅ Retry: 429 Too Many Requests
- ❌ No Retry: 400 Bad Request, 401 Unauthorized, 403 Forbidden
- ❌ No Retry: 404 Not Found

### 8.2 Graceful Degradation: Partner ist offline

**Szenario: Marcos Server ist 2 Stunden down**

**Stefan kann weiterhin:**
- ✅ Sub-Rentals aus lokalem Cache bearbeiten
- ✅ Lokale Equipment-Kataloge durchsuchen
- ✅ Lokale Rechnungen generieren
- ✅ Mit anderen Partnern kommunizieren

**Stefan kann NICHT:**
- ❌ Marcos Verfügbarkeit abfragen (erhält Fehler mit Timestamp des letzten erfolgreichen Checks)
- ❌ Neue Sub-Rental-Anfragen an Marco senden (Anfrage bleibt in `pending_send` lokal)
- ❌ Status-Updates von Marco-Anfragen abrufen (zeigt "Last updated 2h ago")

**UI zeigt dem Admin:**
```
⚠️  Marco Veranstaltungstechnik ist offline seit 2h
    Verfügbarkeitsabfragen funktionieren nicht.
    Letzte Verbindung: 2026-03-20 12:35 UTC
    Automatischer Wiederverbindungs-Versuch in 5 Minuten.
```

### 8.3 Queue für offline Requests

**Wenn Partner offline ist, werden Requests lokal gepuffert:**

```sql
CREATE TABLE federation_pending_requests (
    id UUID PRIMARY KEY,
    partner_id VARCHAR(255),
    request_method VARCHAR(10),
    request_path VARCHAR(512),
    request_body JSONB,
    request_headers JSONB,
    status VARCHAR(50),  -- pending, sent, failed, expired
    created_at TIMESTAMP,
    last_retry_at TIMESTAMP,
    retry_count INT,
    expires_at TIMESTAMP  -- Nach 24h ablaufen
);
```

**Background Worker versucht alle 5 Minuten, gepufferte Requests zu versenden:**

```go
func (w *FederationWorker) ProcessPendingRequests(ctx context.Context) {
    pending := db.GetPendingRequests(ctx, limit=100)

    for _, req := range pending {
        partner := db.GetPartner(ctx, req.PartnerID)

        if partner.LastHealthCheck.Add(2*time.Minute) > time.Now() {
            // Partner ist möglicherweise noch offline, skip
            continue
        }

        err := sendRequest(ctx, partner, req)
        if err != nil {
            req.RetryCount++
            if req.RetryCount >= 288 { // Nach 24h (retries alle 5 Min)
                req.Status = "expired"
                NotifyAdmin("Federation request expired", req)
            }
            db.UpdateRequest(ctx, req)
        } else {
            req.Status = "sent"
            db.UpdateRequest(ctx, req)
        }
    }
}
```

### 8.4 Health-Check Monitoring

**Stefan checkt Marcos Status alle 5 Minuten:**

```go
func (h *HealthChecker) CheckPartnerHealth(ctx context.Context, partner *Partner) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    resp, err := h.client.Get(ctx, partner.BaseURL + "/.well-known/cratedesk-federation")

    if err != nil {
        partner.LastHealthStatus = "offline"
        partner.LastHealthCheck = time.Now()
        partner.OfflineSince = time.Now()  // First failure
        db.UpdatePartner(ctx, partner)

        if time.Since(partner.OfflineSince) > 30*time.Minute {
            AlertAdmin("Partner offline for 30+ minutes")
        }
        return
    }

    partner.LastHealthStatus = "online"
    partner.LastHealthCheck = time.Now()
    partner.OfflineSince = nil
    db.UpdatePartner(ctx, partner)
}
```

### 8.5 Fehler-Responses

**Allgemeines Error Format:**

```json
{
  "error": {
    "code": "FEDERATION_ERROR_CODE",
    "message": "Menschenlesbarer Fehler",
    "timestamp": "2026-03-20T14:35:00Z",
    "request_id": "req-5f8d9c2e-3b4a-11ef-a5c7-0242ac130005"
  }
}
```

**Spezifische Error Codes:**

| Code | HTTP | Beschreibung |
|------|------|-------------|
| INVALID_CERTIFICATE | 401 | mTLS Zertifikat ungültig/abgelaufen |
| INVALID_SIGNATURE | 401 | Request-Signatur ungültig |
| RATE_LIMIT_EXCEEDED | 429 | Rate-Limit überschritten |
| EQUIPMENT_NOT_FOUND | 404 | Equipment nicht vorhanden/nicht geteilt |
| AVAILABILITY_NOT_FOUND | 404 | Verfügbarkeit kann nicht abgefragt werden |
| PARTNER_OFFLINE | 503 | Partner ist offline, Retry später |
| INVALID_DATE_RANGE | 400 | Datum-Bereich ungültig (zu weit in Zukunft) |
| AUTO_CONFIRM_FAILED | 422 | Anfrage konnte nicht auto-bestätigt werden |
| HANDOVER_INCOMPLETE | 409 | Handover noch nicht abgeschlossen |

---

## 9. Versionierung & Capability Negotiation

### 9.1 API-Versionen

**Aktuell:** `1.0`

**Zukünftige Versionen:**
- `1.1`: Neue Features (backward-compatible)
- `2.0`: Breaking Changes (neue Endpunkte)

**Version in Discovery:**
```json
{
  "api_version": "1.0",
  "supported_api_versions": ["1.0", "1.1"],
  "capabilities": [
    "equipment_catalog",
    "availability_query",
    "sub_rental",
    "handover",
    "invoicing"
  ]
}
```

### 9.2 Capability-Negotiation

**Beim Pairing werden Capabilities ausgehandelt:**

```json
{
  "instance_id": "marco-veranstaltung:5c9a8e2b-3f1d-4c9e-8a2b-7d6f4e1a3c9b",
  "base_url": "https://marco-events.com",
  "api_version": "1.0",
  "capabilities": [
    "equipment_catalog",
    "availability_query",
    "sub_rental",
    "handover",
    "invoicing"
  ],
  "feature_flags": {
    "handover_photo_compression": true,
    "invoice_pdf_generation": true,
    "automatic_certificate_rotation": true
  }
}
```

### 9.3 Handling incompatibler Versionen

**Wenn Stefan API 2.0 hat, Marco aber nur 1.0:**

```
Stefan macht Request mit: X-CrateDesk-API-Version: 2.0
Marco antwortet:
HTTP/1.1 505 HTTP Version Not Supported
Content-Type: application/json

{
  "error": {
    "code": "API_VERSION_NOT_SUPPORTED",
    "message": "API version 2.0 not supported",
    "supported_versions": ["1.0"]
  }
}
```

**Fallback-Strategie:**
1. Stefan versucht mit API 2.0
2. Falls 505, downgrade auf 1.0
3. Log als Warnung
4. Admin erhält Hinweis: "Partner unterstützt noch API 1.0"

---

## 10. Sequenz-Diagramme

### 10.1 Pairing-Prozess

```
Stefan Admin              Stefan System            Marco System           Marco Admin
    |                         |                        |                       |
    | 1. Klick "Partner hinzufügen"                   |                       |
    |--[Admin UI]---------->  |                        |                       |
    |                         |                        |                       |
    |      2. Generiere Pairing Token (TTL 24h)       |                       |
    |  <---------[PAIRING_TOKEN]--                    |                       |
    |                         |                        |                       |
    | 3. QR-Code/Token anzeigen                       |                       |
    | <--------[UI]--------    |                        |                       |
    |                         |                        |                       |
    | 4. Copy Token                                   |                        |
    | 5. Send per Mail ---------------------------------------->   Marco (Email)
    |                         |                        |
    |                         |      6. Paste Token in Admin UI
    |                         |   <--------[Admin UI]--
    |                         |
    |                         |      7. GET /.well-known/...
    |                         |  <------[mTLS Handshake]----[No mTLS yet, plaintext]
    |                         |
    |  8. Validates Token, Extracts Marcos Cert
    |  <---------[PAIRING_TOKEN validation]---
    |                         |
    |  9. POST /pair mit Marcos Zertifikat + Token
    |  <------[HTTP POST + marcos-root-ca.crt]
    |                         |
    |     10. Validiert Token, speichert Marcos Cert
    |  --------[Token validation + cert pinning]---->
    |                         |
    |  11. HTTP 200 OK + Stefans Zertifikat
    |  ------[stefan-root-ca.crt]----------------->
    |                         |
    |     12. Speichert Stefans Cert, pinned
    |  <---------[cert pinning]---
    |                         |
    |  13. Test-Handshake: GET /health mit mTLS
    |  ------[mTLS + Signature]----[mTLS Validation]-->
    |                         |
    |  14. HTTP 200 OK
    |  --------[health check response]------------>
    |                         |
    | 15. Admin: "Pairing erfolgreich!"              |
    | <--------[UI Success Message]                   |
    |                         |
    |                         |                       |  16. Admin: "Pairing erfolgreich!"
    |                         |                       |  <-------[UI Success Message]
    |                         |                       |
```

### 10.2 Equipment-Verfügbarkeit abfragen

```
Marco Admin/UI       Marco System        Stefan System        Stefan DB
    |                    |                   |                   |
    | 1. Suche "Audio-Mischpulte" bei Stefan
    |---[UI Search]------>  |
    |                       |
    |                       | 2. mTLS + Signed GET /equipment/categories
    |                       |----[HTTPS + Certificate + Signature]---->
    |                       |
    |                       |             3. Validiere mTLS + Signature
    |                       |             <---------[Cert + Sig Validation]
    |                       |
    |                       |             4. SELECT categories WHERE partner_id='marco'
    |                       |             <-----[Query]------
    |                       |
    |                       |             5. [cat-audio-001, cat-audio-002]
    |                       |             --------[Result]----->
    |                       |
    |                       | 6. JSON mit Kategorien + Equipment-Count
    |                       |  <-----[JSON Response]----
    |                       |
    | 7. Zeige Kategorien an
    | <------[Categories List UI]------
    |
    | 8. Klick "Soundcraft Si Impact - Verfügbarkeit"
    |-----[Equipment Selection]----->
    |
    |                       | 9. POST /availability/query
    |                       |  Anfrage: 2026-04-01 bis 2026-04-03
    |                       |----[HTTPS + mTLS + Signature]---->
    |                       |
    |                       |             10. Validiere Request
    |                       |             Berechne Verfügbarkeit aus lokalen Daten
    |                       |             <---------[Query]------
    |                       |
    |                       |             11. Verfügbarkeitskalender (2.5 Tage)
    |                       |             --------[Result]----->
    |                       |
    |                       | 12. JSON mit Verfügbarkeit
    |                       |  <-----[JSON Response]----
    |                       |
    | 13. Zeige Kalender: "2. verfügbar - Preis €127.50/Tag"
    | <------[Availability UI]------
    |
```

### 10.3 Sub-Rental Workflow (Komplett)

```
Marco Admin         Marco System        Stefan System       Stefan Admin
    |                   |                   |                   |
    | 1. Klick "Angebot erstellen"
    |---[UI Button]------>  |
    |                       |
    |                       | 2. POST /sub-rentals
    |                       |----[mTLS + Sig + Sub-Rental JSON]---->
    |                       |
    |                       |             3. Validiere Request
    |                       |             Prüfe Verfügbarkeit
    |                       |             <---------[Local Check]
    |                       |
    |                       |             4. INSERT sub-rental (status=pending)
    |                       |             <---------[DB]-------
    |                       |
    |                       | 5. HTTP 202 Accepted
    |                       | <----[Response + ID]----
    |                       |
    | 6. "Anfrage eingereicht, warte auf Bestätigung"
    | <----[UI Message]-----
    |                       |
    |                       | (Optional: Send Email to Stefan)
    |                       |
    |                       |                   | 7. Email: "Neue Sub-Rental-Anfrage"
    |                       |                   | <-----[Email]---
    |                       |
    |                       |                   | 8. Öffnet Admin UI
    |                       |                   |-----[Dashboard]---->
    |                       |
    |                       |                   | 9. Sieht Anfrage "Pending"
    |                       |                   | <-----[UI List]---
    |                       |
    |                       |                   | 10. Klick "Details anschauen"
    |                       |                   |------[UI]-------->
    |                       |
    | Marco warte auf Feedback...             | 11. Liest Anfrage, prüft Verfügbarkeit
    |                                         | Klick "Bestätigen"
    |                                         |-------[UI Button]----->
    |                       |
    |                       | 12. POST /sub-rentals/{id}/confirm
    |                       | <----[mTLS + Sig + Confirm JSON]----
    |                       |
    |                       |             13. UPDATE sub-rental (status=confirmed)
    |                       |             <---------[DB]-------
    |                       |
    |                       | 14. HTTP 200 OK
    |                       | <----[Confirm Response]----
    |                       |
    | 15. "Sub-Rental bestätigt!"
    | <----[UI Message]-----
    |
    | 16. Marco plant Pickup am 2026-04-01 08:00 UTC
    |
    |                       | 17. POST /handover/start
    |                       |----[Pickup Intent]---->
    |                       |
    |                       | 18. HTTP 200 + Checklist
    |                       | <----[Checklist Items]----
    |                       |
    | 19. Techniker bestätigt Geräte bei Pickup
    | Fotos machen
    |-----[Photo + Checklist]----->
    |
    |                       | 20. POST /handover/{id}/documents
    |                       |----[Photos + Signatures + Checklist]---->
    |                       |
    |                       | 21. HTTP 200 Handover Complete
    |                       | <----[Response]----
    |                       |
    | 22. "Equipment übernommen"
    | <----[UI Success]-----
    |
    | 23. Marco nutzt Equipment 2026-04-01 bis 04-03
    |
    | 24. Sub-Rental endet, Return-Handover
    |-----[Return Documentation]----->
    |                       |
    |                       | 25. POST /handover/return
    |                       |----[Return + Condition Docs]---->
    |                       |
    |                       | 26. HTTP 200 Return Complete
    |                       | <----[Response + Invoice URL]----
    |                       |
    | 27. "Equipment zurückgegeben"
    | <----[UI Success]-----
    |
    | 28. Marco fragt: GET /invoices/{id}
    |-----[Invoice Request]----->
    |                       |
    |                       | 29. GET /invoices/inv-2026-03-20-001
    |                       | <----[mTLS + Sig]----
    |                       |
    |                       | 30. HTTP 200 + Invoice JSON
    |                       | <----[Invoice Data]----
    |                       |
    | 31. "Rechnung: €468.56 fällig am 2026-04-10"
    | <----[Invoice UI]-----
    |
```

### 10.4 Handover-Prozess (Detailed)

```
Marco Techniker      Marco System        Stefan System       Stefan Techniker
    |                    |                   |                   |
    | 1. Komme zu Stefans Studio für Pickup
    |                    |                   |
    |                    | 2. POST /handover/start (type=pickup)
    |                    |----[Start Intent]---->
    |                    |
    |                    |             3. CREATE handover (id=hnd-...)
    |                    |             GET checklist items
    |                    |             <---------[DB]-------
    |                    |
    |                    | 4. HTTP 200 + Checklist
    |                    | <----[Checklist Template]----
    |                    |
    | 5. Mobile App zeigt Checklist:
    |    ☐ Soundcraft Si Impact
    |    ☐ Power Supply
    |    ☐ XLR Cables (5x)
    | <----[UI Checklist]-----
    |
    | 6. Equipment prüfen
    |    ☑ Soundcraft Si Impact - OK
    |    ☑ Power Supply - OK
    |    ☑ XLR Cables - OK
    |
    | 7. Gesamtzustand Foto machen
    |-----[Photo Capture]----->
    |
    |                    | 8. POST /handover/{id}/documents
    |                    |----[Checklist Status + Photos + Metadata]---->
    |                    |
    |                    |             9. VALIDATE Photos (not empty)
    |                    |             STORE Photos in S3 (encrypted)
    |                    |             STORE Checklist completion
    |                    |             <---------[S3 + DB]-------
    |                    |
    |                    | 10. HTTP 200 OK
    |                    | <----[Confirmation + Signature Request]----
    |                    |
    | 11. "Bitte signieren mit Handy-Stift"
    | <----[Signature Pad]-----
    |
    | 12. Unterschreibe auf Handy-Display
    |-----[Signature Data]----->
    |
    |                    | 13. POST /handover/{id}/sign
    |                    |----[Signature]---->
    |                    |
    |                    |             14. VERIFY Signature Format
    |                    |             SAVE Signature (encrypted)
    |                    |             UPDATE handover status=completed
    |                    |             <---------[DB]-------
    |                    |
    |                    | 15. HTTP 200 Handover Complete
    |                    | <----[Success + Summary]----
    |                    |
    | 16. "Handover abgeschlossen!"
    | <----[UI Success]-----
    |
    |                    |             (Stefan erhält Notification)
    |                    |             17. Admin Alert: "Handover received from Marco"
    |                    |             <---------[Notification]
    |                    |
    |                    |                       | 18. Admin: "Handover überprüfen"
    |                    |                       |-----[Admin UI]---->
    |                    |
    |                    |                       | 19. Sieht Fotos + Checklist
    |                    |                       | <-----[View Docs]---
    |                    |                       |
    |                    |                       | 20. Alles OK?
    |                    |                       | "Akzeptiert" Klick
    |                    |                       |-----[Accept Button]---->
    |                    |
    |                    | 21. POST /handover/{id}/acknowledge
    |                    | <----[Accept]----
    |                    |
    |                    | 22. HTTP 200 Acknowledged
    |                    | <----[Response]----
    |                    |
    | 23. "Equipment offiziell übernommen"
    | <----[Final UI]-----
    |
```

---

## 11. Vergleich mit Rentman

### 11.1 Rentman's proprietäres Modell (Probleme)

| Aspekt | Rentman | CrateDesk Federation |
|--------|---------|-------------------|
| **Zentraler Server** | ✅ Ja (rentman.io) | ❌ Nein – P2P |
| **Datenkontrolle** | ⚠️ Rentman kontrolliert alles | ✅ Jeder Admin kontrolliert seine Daten |
| **Single Point of Failure** | ✅ Ja (Cloud-Ausfallrisiko) | ❌ Nein – dezentral |
| **Vendor Lock-In** | ✅ Ja (nur via Rentman) | ❌ Nein – offenes Protokoll |
| **Abhängigkeit** | ✅ Von Rentman-Infrastruktur | ❌ Nur von eigener Infrastruktur |
| **Kosten** | ⚠️ SaaS-Gebühren + Integration | ✅ Kostenlos (nur API-Calls) |
| **Transparenz** | ❌ Closed Source | ✅ Offenes Protokoll (dokumentiert) |
| **Datenschutz** | ⚠️ Drittanbieter sieht alles | ✅ Keine Zentralisierung |
| **Offline-Fähigkeit** | ❌ Funktioniert nicht ohne Rentman | ✅ Alle Funktionen lokal verfügbar |
| **Customization** | ❌ Rentman bestimmt Features | ✅ Jeder kann erweitern |
| **Switching Costs** | ✅ Sehr hoch (Export schwierig) | ❌ Niedrig (Standard REST API) |

### 11.2 Warum CrateDesk Federation besser ist

#### 1. **Wahre Dezentralisierung**
- Rentman = Zentraler Hub (alle Daten fließen durch Rentman)
- CrateDesk = Direkt Peer-to-Peer (Stefan ↔ Marco, kein Drittel-Server)

#### 2. **Datensouveränität**
```
Rentman Model:
  VT-Firma A [Server]
       ↓
    [Rentman Cloud]  ← alle Daten sichtbar für Rentman
       ↑
  VT-Firma B [Server]

CrateDesk Model:
  VT-Firma A [Server]
       ↕ (direkt, nur notwendige Daten)
  VT-Firma B [Server]

  → Rentman sieht nichts
  → Nur A und B sehen die Daten
```

#### 3. **Offline-Resilienz**
- **Rentman:** Wenn rentman.io down ist → keine Vernetzung möglich
- **CrateDesk:** Wenn Marcos Server down ist → Stefan arbeitet normal weiter, Marco ist nur nicht erreichbar

#### 4. **Kosten**
- **Rentman:** SaaS-Gebühren + API-Kosten + Setup-Gebühren
- **CrateDesk:** Kostenlos – nur API-Calls zwischen Servern

#### 5. **Transparenz & Security**
- **Rentman:** Proprietär, Black Box (wie werden Daten verarbeitet? Wer hat Zugriff?)
- **CrateDesk:** Dokumentiertes Protokoll, Open Source, Jeder kann überprüfen (mTLS, Signatures, kein Backdoor)

#### 6. **Customization**
- **Rentman:** Features kommen von Rentman, nicht konfigurierbar
- **CrateDesk:** Jede Firma konfiguriert selbst (Preise, Kategorien, Auto-Confirm, etc.)

#### 7. **Switching Costs**
- **Rentman:** Hohe Abhängigkeit, Daten fest in Rentman
- **CrateDesk:** Standard REST JSON API – jederzeit aussteigen, Daten bleiben bei dir

#### 8. **Compliance & Datenschutz**
```
Rentman-Problem:
  "Equipment verfügbar in Partner B"
  → Diese Daten fließen durch Rentman-Server
  → Rentman muss als Datenverarbeiter-Vertrag (AVV) mit dir haben
  → Rentman könnte theoretisch deine Daten für ML-Training nutzen
  → GDPR-Risiko: Daten bei US-Firma (Rentman gegründet in USA/EU?)

CrateDesk-Lösung:
  "Equipment verfügbar in Partner B"
  → Daten fließen DIREKT von Marcos Server zu Stefans Server
  → Keine Zentralisierung
  → Keine Third-Party-Sichtbarkeit
  → GDPR-konform: nur Partner sehen die Daten
```

#### 9. **Skalierbarkeit**
- **Rentman:** Alle Abfragen gehen durch Rentman → Bottleneck
- **CrateDesk:** Jeder Partner handled seine eigenen Abfragen → N×N-Skalierung

#### 10. **Ecosystem**
```
Rentman:
  Rentman API ← nur proprietäre Features

CrateDesk:
  ✓ Offenes Protokoll
  ✓ Jeder Entwickler kann Tools bauen
  ✓ Andere Software kann integrieren
  ✓ Standard HTTPS + mTLS (Branchenstandard)
  ✓ JSON ist universell
```

### 11.3 Zitat: Stefans Vision

> "Ich will nicht dass Rentman alles über meine Kunden und meine Preisgestaltung weiß. Mein Server, meine Daten. Ich möchte mit Marco direkt reden können, ohne dass Rentman dazwischenschaut. Und wenn Rentman morgen von einer Private-Equity-Firma gekauft wird und die Features auf einmal kosten, bin ich nicht in einer Falle – ich kann einfach weiterhin mit meinen Partnern communizieren, weil das Protokoll offen ist."

---

## Zusammenfassung

**CrateDesk Federation** ist ein **dezentrales, sicheres und kontrollierbares Netzwerk** für unabhängige VT-Firmen:

✅ **Keine zentrale Plattform** – P2P direkt zwischen Servern
✅ **Volle Kontrolle** – Jeder Admin bestimmt, was geteilt wird
✅ **Sicher** – mTLS + Request-Signing + Certificate Pinning
✅ **Zukunftssicher** – Offenes Protokoll, Standard-Technologie
✅ **Resilient** – Funktioniert auch wenn Partner offline sind
✅ **Konform** – GDPR-ready, keine Zentralisierung von Daten
✅ **Kostenfrei** – Keine Gebühren an Drittanbieter

**Differenzierungspunkt vs. Rentman:** Nicht bessere Features, sondern **bessere Kontrollstruktur** – das ist was Stefan braucht.

