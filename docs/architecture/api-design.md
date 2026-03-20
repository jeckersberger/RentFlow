# REST API Design - Veranstaltungstechnik Lagerverwaltung & Abrechnung

**Version:** 1.0
**API Basis-URL:** `https://api.example.com/api/v1`
**Stand:** 2026-03-20

---

## Inhaltsverzeichnis

1. [Allgemeine Konventionen](#allgemeine-konventionen)
2. [Authentication & Autorisierung](#authentication--autorisierung)
3. [Error Handling](#error-handling)
4. [Rate Limiting & Versionierung](#rate-limiting--versionierung)
5. [File Uploads](#file-uploads)
6. [Real-Time Updates](#real-time-updates)
7. [Auth Module](#auth-module)
8. [Equipment & Lager Module](#equipment--lager-module)
9. [Projekte Module](#projekte-module)
10. [Angebote & Rechnungen Module](#angebote--rechnungen-module)
11. [Sub-Rental Module](#sub-rental-module)
12. [Federation Module](#federation-module)
13. [Admin Module](#admin-module)
14. [Scanner Module](#scanner-module)
15. [Dashboard Module](#dashboard-module)
16. [Zeiterfassung Module](#zeiterfassung-module)

---

## Allgemeine Konventionen

### API-Versionierung
- Alle Endpunkte verwenden `/api/v1/` als Basis-Prefix
- Versionierung erfolgt im URL-Path (nicht im Header)
- Zukünftige Versionen: `/api/v2/`, `/api/v3/` etc.

### HTTP Status Codes
- `200 OK` - Erfolgreich, mit Response Body
- `201 Created` - Ressource erstellt
- `204 No Content` - Erfolgreich, kein Response Body
- `400 Bad Request` - Ungültige Request-Daten
- `401 Unauthorized` - Authentifizierung erforderlich
- `403 Forbidden` - Authentifiziert, aber Zugriff nicht erlaubt
- `404 Not Found` - Ressource nicht vorhanden
- `409 Conflict` - Ressource-Konflikt (z.B. doppelte Seriennummer)
- `422 Unprocessable Entity` - Validierungsfehler
- `429 Too Many Requests` - Rate Limit überschritten
- `500 Internal Server Error` - Server-Fehler
- `503 Service Unavailable` - Wartung/Unavailable

### Pagination Standard
```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

**Query-Parameter:**
- `page` (default: 1)
- `per_page` (default: 20, max: 100)
- `sort_by` (Feldname)
- `sort_order` (asc|desc, default: asc)

### Standard-Filterung
- `search` - Volltextsuche in relevanten Feldern
- `filter[field]` - Exakte Filterung nach Feld (mehrere möglich)
- Beispiel: `GET /equipment?filter[status]=aktiv&filter[type_id]=5&sort_by=name&sort_order=asc`

### Timestamps
- ISO 8601 Format: `2026-03-20T14:30:00Z`
- Alle Timestamps in UTC
- Felder: `created_at`, `updated_at`, `deleted_at` (soft-delete) auf allen Ressourcen

---

## Authentication & Autorisierung

### Rollen und Berechtigungen
```
Admin: Voller Zugriff auf alle Funktionen
Projektleiter: Projekte, Angebote, Rechnungsfreigabe, Sub-Rental, Equipment-Verfügbarkeit
Lager: Equipment CRUD, Lagerplätze, Packlisten, Defekte, Scans. KEINE Preise/Finanzen
Buchhaltung: Rechnungen, Mahnungen, DATEV-Export, offene Posten. KEIN Lager
Freelancer: Nur eigene Jobs, Packlisten, Zeiterfassung. KEINE Preise/Kunden/andere Jobs
Custom: Individuell konfigurierbar
```

### JWT Token Struktur
```json
{
  "sub": "user_uuid",
  "email": "user@company.de",
  "company_id": "company_uuid",
  "roles": ["Projektleiter", "Custom"],
  "permissions": ["project:read", "project:write", "offer:read"],
  "iat": 1711000000,
  "exp": 1711003600,
  "type": "access"
}
```

### Token-Typen
- **Access Token**: 1 Stunde Gültigkeit
- **Refresh Token**: 30 Tage Gültigkeit (in HttpOnly Cookie gespeichert)

### Authorization Header
```
Authorization: Bearer <access_token>
```

---

## Error Handling

### Standard-Error-Response Format
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Equipment mit ID 'xyz' nicht gefunden",
    "status": 404,
    "timestamp": "2026-03-20T14:30:00Z",
    "request_id": "req_abc123xyz",
    "details": {
      "field": "equipment_id",
      "value": "xyz"
    }
  }
}
```

### Validierungsfehler (422)
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validierungsfehler in Request",
    "status": 422,
    "timestamp": "2026-03-20T14:30:00Z",
    "request_id": "req_abc123xyz",
    "details": {
      "errors": [
        {
          "field": "serial_number",
          "message": "Seriennummer bereits vergeben"
        },
        {
          "field": "name",
          "message": "Name ist erforderlich"
        }
      ]
    }
  }
}
```

### Rate Limit Response (429)
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Zu viele Anfragen. Bitte warten Sie.",
    "status": 429,
    "timestamp": "2026-03-20T14:30:00Z",
    "details": {
      "retry_after": 60
    }
  }
}
```

---

## Rate Limiting & Versionierung

### Rate Limiting Strategy
- **Standard**: 100 Requests pro Minute pro User/API-Key
- **Scanner Endpoints**: 1000 Requests pro Minute (für Scanning in Echtzeit)
- **File Upload**: 50 Uploads pro Minute
- **Response Headers**:
  ```
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 87
  X-RateLimit-Reset: 1711003600
  ```

### API Version Management
- Aktuelle Version: `v1`
- Deprecation Timeline: 12 Monate vor Entfernung
- Deprecation Header: `Deprecation: true`, `Sunset: Sun, 31 Dec 2027 23:59:59 GMT`

---

## File Uploads

### Upload-Strategie
- **Zielformate**: JPEG, PNG, WebP
- **Max Größe**: 10 MB pro Datei
- **Storage**: S3-kompatibel (Minio) oder lokales Filesystem
- **URLs**: Signierte URLs mit Expiration (24h)

### Upload-Endpunkt
```
POST /api/v1/files/upload
Content-Type: multipart/form-data

Request:
{
  "file": <binary>,
  "context": "equipment_defect|equipment_photo|company_logo",
  "metadata": {
    "equipment_id": "uuid",
    "project_id": "uuid"
  }
}

Response 201:
{
  "id": "file_uuid",
  "url": "https://api.example.com/api/v1/files/file_uuid/download",
  "signed_url": "https://storage.example.com/...",
  "signed_url_expires_at": "2026-03-21T14:30:00Z",
  "filename": "original_filename.jpg",
  "size": 2048576,
  "mime_type": "image/jpeg",
  "context": "equipment_defect",
  "created_at": "2026-03-20T14:30:00Z"
}
```

### Download mit Authentifizierung
```
GET /api/v1/files/{file_id}/download
GET /api/v1/files/{file_id}/thumbnail?size=small|medium|large

Response:
- Content-Type: image/jpeg|image/png|image/webp
- Content-Disposition: attachment|inline
```

---

## Real-Time Updates

### WebSocket Connection (für Scanner & Packlisten)
```
wss://api.example.com/api/v1/ws

Connection:
Authorization: Bearer <access_token>
```

### WebSocket Event-Typen

#### Packlisten-Updates
```json
{
  "type": "packing_list.scan_event",
  "project_id": "uuid",
  "data": {
    "scan_event_id": "uuid",
    "barcode": "EQ-12345",
    "equipment_id": "uuid",
    "event": "checkin|checkout",
    "timestamp": "2026-03-20T14:30:00Z",
    "scanned_by": "user_uuid",
    "status": "success|error"
  }
}
```

#### Sub-Rental-Updates
```json
{
  "type": "sub_rental.status_changed",
  "sub_rental_id": "uuid",
  "data": {
    "old_status": "Angefragt",
    "new_status": "Bestätigt",
    "timestamp": "2026-03-20T14:30:00Z"
  }
}
```

#### Project Status
```json
{
  "type": "project.status_changed",
  "project_id": "uuid",
  "data": {
    "old_status": "Geplant",
    "new_status": "Aufbau",
    "timestamp": "2026-03-20T14:30:00Z"
  }
}
```

### Server-Sent Events (Alternative zu WebSocket)
```
GET /api/v1/events/stream?filter=project_id:uuid

Content-Type: text/event-stream

data: {"type": "packing_list.scan_event", "data": {...}}
```

---

# Module API Endpoints

---

## Auth Module

### 1. Login
```
POST /api/v1/auth/login
No Authorization required

Request:
{
  "email": "user@company.de",
  "password": "secure_password"
}

Response 200:
{
  "user": {
    "id": "user_uuid",
    "email": "user@company.de",
    "first_name": "Max",
    "last_name": "Mustermann",
    "company_id": "company_uuid",
    "roles": ["Projektleiter"],
    "avatar_url": "https://...",
    "created_at": "2025-12-01T10:00:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGc...",
    "access_token_expires_in": 3600,
    "refresh_token": "rtk_...",
    "refresh_token_expires_in": 2592000
  }
}

Response 401:
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Email oder Passwort ungültig"
  }
}
```

### 2. Refresh Token
```
POST /api/v1/auth/refresh
No Authorization required (Cookie-based refresh token)

Response 200:
{
  "tokens": {
    "access_token": "eyJhbGc...",
    "access_token_expires_in": 3600
  }
}

Response 401:
{
  "error": {
    "code": "REFRESH_TOKEN_EXPIRED",
    "message": "Refresh Token abgelaufen. Bitte neu anmelden."
  }
}
```

### 3. Logout
```
POST /api/v1/auth/logout
Authorization: Bearer <access_token>

Response 204: No Content
```

### 4. Passwort Reset - Request
```
POST /api/v1/auth/password-reset/request
No Authorization required

Request:
{
  "email": "user@company.de"
}

Response 200:
{
  "message": "Passwort-Reset-Link wurde gesendet"
}

Response 404:
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "Benutzer mit dieser E-Mail nicht gefunden"
  }
}
```

### 5. Passwort Reset - Confirm
```
POST /api/v1/auth/password-reset/confirm
No Authorization required

Request:
{
  "token": "prt_xyz123abc",
  "password": "new_secure_password",
  "password_confirm": "new_secure_password"
}

Response 200:
{
  "message": "Passwort erfolgreich geändert"
}

Response 400:
{
  "error": {
    "code": "INVALID_RESET_TOKEN",
    "message": "Reset-Token ungültig oder abgelaufen"
  }
}
```

### 6. Benutzer Profil
```
GET /api/v1/auth/me
Authorization: Bearer <access_token>

Response 200:
{
  "id": "user_uuid",
  "email": "user@company.de",
  "first_name": "Max",
  "last_name": "Mustermann",
  "company_id": "company_uuid",
  "roles": ["Projektleiter"],
  "permissions": ["project:read", "project:write", "offer:read"],
  "avatar_url": "https://...",
  "phone": "+49123456789",
  "settings": {
    "language": "de",
    "timezone": "Europe/Berlin",
    "email_notifications": true
  },
  "created_at": "2025-12-01T10:00:00Z",
  "updated_at": "2026-03-20T14:30:00Z"
}
```

### 7. Profil aktualisieren
```
PATCH /api/v1/auth/me
Authorization: Bearer <access_token>

Request:
{
  "first_name": "Maxim",
  "last_name": "Musterstein",
  "phone": "+49987654321",
  "settings": {
    "language": "de",
    "timezone": "Europe/Berlin"
  }
}

Response 200:
{
  "id": "user_uuid",
  "email": "user@company.de",
  "first_name": "Maxim",
  "last_name": "Musterstein",
  ...
}
```

### 8. Passwort ändern
```
POST /api/v1/auth/change-password
Authorization: Bearer <access_token>

Request:
{
  "old_password": "current_password",
  "new_password": "new_secure_password",
  "new_password_confirm": "new_secure_password"
}

Response 200:
{
  "message": "Passwort erfolgreich geändert"
}

Response 400:
{
  "error": {
    "code": "INVALID_OLD_PASSWORD",
    "message": "Aktuelles Passwort ist nicht korrekt"
  }
}
```

### 9. Benutzer einladen (Admin only)
```
POST /api/v1/auth/invitations
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "email": "newuser@company.de",
  "first_name": "Neue",
  "last_name": "Person",
  "roles": ["Lager"],
  "message": "Willkommen im Team!"
}

Response 201:
{
  "id": "invitation_uuid",
  "email": "newuser@company.de",
  "token": "inv_xyz123abc",
  "invite_url": "https://app.example.com/auth/accept-invitation?token=inv_xyz123abc",
  "expires_at": "2026-03-27T14:30:00Z",
  "created_at": "2026-03-20T14:30:00Z"
}
```

### 10. Einladung akzeptieren
```
POST /api/v1/auth/invitations/{token}/accept
No Authorization required

Request:
{
  "password": "secure_password",
  "password_confirm": "secure_password"
}

Response 200:
{
  "user": { ... },
  "tokens": { ... }
}

Response 400:
{
  "error": {
    "code": "INVALID_INVITATION_TOKEN",
    "message": "Einladungslink ungültig oder abgelaufen"
  }
}
```

### 11. Einladungen auflisten (Admin only)
```
GET /api/v1/auth/invitations
Authorization: Bearer <access_token>
Required Role: Admin

Query Parameters:
- status=pending|accepted|expired (default: pending)
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "invitation_uuid",
      "email": "user@company.de",
      "status": "pending",
      "expires_at": "2026-03-27T14:30:00Z",
      "created_at": "2026-03-20T14:30:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 12. Einladung stornieren (Admin only)
```
DELETE /api/v1/auth/invitations/{invitation_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 204: No Content
```

---

## Equipment & Lager Module

### Equipment Type Management

#### 1. Equipment-Typen auflisten
```
GET /api/v1/equipment-types
Authorization: Bearer <access_token>

Query Parameters:
- search=name_search
- sort_by=name|created_at
- sort_order=asc|desc
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "type_uuid",
      "name": "Beamer",
      "description": "Projektoren verschiedener Leuchtstärke",
      "category": "Projektion",
      "rental_unit": "Stück",
      "default_daily_rate": 150.00,
      "default_setup_fee": 50.00,
      "is_consumable": false,
      "requires_serial_number": true,
      "created_at": "2025-12-01T10:00:00Z",
      "updated_at": "2026-03-20T14:30:00Z"
    }
  ],
  "pagination": { ... }
}
```

#### 2. Equipment-Typ erstellen (Admin only)
```
POST /api/v1/equipment-types
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "name": "Beamer",
  "description": "Projektoren verschiedener Leuchtstärke",
  "category": "Projektion",
  "rental_unit": "Stück",
  "default_daily_rate": 150.00,
  "default_setup_fee": 50.00,
  "is_consumable": false,
  "requires_serial_number": true
}

Response 201:
{
  "id": "type_uuid",
  "name": "Beamer",
  ...
}
```

#### 3. Equipment-Typ aktualisieren (Admin only)
```
PATCH /api/v1/equipment-types/{type_id}
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "default_daily_rate": 160.00,
  "default_setup_fee": 55.00
}

Response 200:
{
  "id": "type_uuid",
  "name": "Beamer",
  ...
}
```

#### 4. Equipment-Typ löschen (Admin only)
```
DELETE /api/v1/equipment-types/{type_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 204: No Content
```

### Individual Equipment Management

#### 5. Equipment-Liste
```
GET /api/v1/equipment
Authorization: Bearer <access_token>

Query Parameters:
- filter[type_id]=uuid
- filter[status]=aktiv|reserviert|defekt|archiviert
- filter[storage_location_id]=uuid
- search=serial_number|name
- sort_by=name|serial_number|created_at
- sort_order=asc|desc
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "equipment_uuid",
      "type_id": "type_uuid",
      "type_name": "Beamer",
      "serial_number": "BM-001-2024",
      "inventory_number": "INV-12345",
      "name": "Beamer Sony VPL-FHZ100",
      "description": "Native 4K-Projektor",
      "status": "aktiv",
      "condition": "gut",
      "storage_location_id": "location_uuid",
      "storage_location_path": "Lager / Regal A / Fach 3",
      "purchase_date": "2023-06-15",
      "purchase_price": 15000.00,
      "last_maintenance": "2026-02-10",
      "next_maintenance_due": "2026-08-10",
      "defect_notes": null,
      "photos": [
        {
          "id": "file_uuid",
          "url": "https://...",
          "context": "equipment_photo"
        }
      ],
      "created_at": "2025-12-01T10:00:00Z",
      "updated_at": "2026-03-20T14:30:00Z"
    }
  ],
  "pagination": { ... }
}
```

#### 6. Equipment erstellen
```
POST /api/v1/equipment
Authorization: Bearer <access_token>
Required Role: Admin, Lager

Request:
{
  "type_id": "type_uuid",
  "serial_number": "BM-001-2024",
  "inventory_number": "INV-12345",
  "name": "Beamer Sony VPL-FHZ100",
  "description": "Native 4K-Projektor",
  "storage_location_id": "location_uuid",
  "purchase_date": "2023-06-15",
  "purchase_price": 15000.00,
  "next_maintenance_due": "2026-08-10"
}

Response 201:
{
  "id": "equipment_uuid",
  "type_id": "type_uuid",
  ...
}

Response 409:
{
  "error": {
    "code": "DUPLICATE_SERIAL_NUMBER",
    "message": "Seriennummer bereits vergeben",
    "details": {
      "existing_equipment_id": "other_uuid"
    }
  }
}
```

#### 7. Equipment abrufen
```
GET /api/v1/equipment/{equipment_id}
Authorization: Bearer <access_token>

Response 200:
{
  "id": "equipment_uuid",
  "type_id": "type_uuid",
  "serial_number": "BM-001-2024",
  ...
}
```

#### 8. Equipment aktualisieren
```
PATCH /api/v1/equipment/{equipment_id}
Authorization: Bearer <access_token>
Required Role: Admin, Lager

Request:
{
  "name": "Beamer Sony VPL-FHZ100 (neu kalibriert)",
  "storage_location_id": "new_location_uuid",
  "condition": "ausgezeichnet",
  "next_maintenance_due": "2026-09-10"
}

Response 200:
{
  "id": "equipment_uuid",
  ...
}
```

#### 9. Equipment löschen/archivieren
```
DELETE /api/v1/equipment/{equipment_id}
Authorization: Bearer <access_token>
Required Role: Admin

Request (optional):
{
  "soft_delete": true,
  "archive_reason": "Verkauft"
}

Response 204: No Content
```

### Storage Locations

#### 10. Lagerplätze auflisten
```
GET /api/v1/storage-locations
Authorization: Bearer <access_token>

Query Parameters:
- search=name
- include_empty=true|false
- hierarchy=true|false (nested tree)

Response 200:
{
  "data": [
    {
      "id": "location_uuid",
      "name": "Regal A",
      "parent_id": "parent_location_uuid",
      "path": "Lager / Regal A",
      "description": "Hauptlager, Regalsystem",
      "capacity": 50,
      "current_equipment_count": 23,
      "sub_locations": [
        {
          "id": "sub_location_uuid",
          "name": "Fach 1",
          ...
        }
      ],
      "created_at": "2025-12-01T10:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

#### 11. Lagerplatz erstellen
```
POST /api/v1/storage-locations
Authorization: Bearer <access_token>
Required Role: Admin, Lager

Request:
{
  "name": "Regal A",
  "parent_id": "parent_uuid",
  "description": "Hauptlager, Regalsystem",
  "capacity": 50
}

Response 201:
{
  "id": "location_uuid",
  "name": "Regal A",
  ...
}
```

#### 12. Lagerplatz aktualisieren
```
PATCH /api/v1/storage-locations/{location_id}
Authorization: Bearer <access_token>
Required Role: Admin, Lager

Request:
{
  "name": "Regal A (reorganisiert)",
  "capacity": 60
}

Response 200:
{
  "id": "location_uuid",
  ...
}
```

#### 13. Lagerplatz löschen
```
DELETE /api/v1/storage-locations/{location_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 204: No Content
```

### Equipment Combinations (Flightcases)

#### 14. Equipment-Kombinationen auflisten
```
GET /api/v1/equipment-combinations
Authorization: Bearer <access_token>

Query Parameters:
- filter[type]=flightcase|bundle|set
- search=name
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "combination_uuid",
      "name": "Beamer Flightcase Set",
      "description": "Komplettes Beamer-Setup mit Zubehör",
      "type": "flightcase",
      "items": [
        {
          "equipment_id": "uuid",
          "equipment_name": "Beamer Sony VPL-FHZ100",
          "quantity": 1,
          "order": 1
        },
        {
          "equipment_id": "uuid",
          "equipment_name": "Lens Standard",
          "quantity": 1,
          "order": 2
        }
      ],
      "default_rental_price": 500.00,
      "is_available": true,
      "created_at": "2025-12-01T10:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

#### 15. Equipment-Kombination erstellen
```
POST /api/v1/equipment-combinations
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "name": "Beamer Flightcase Set",
  "description": "Komplettes Beamer-Setup mit Zubehör",
  "type": "flightcase",
  "items": [
    {
      "equipment_id": "uuid",
      "quantity": 1,
      "order": 1
    },
    {
      "equipment_id": "uuid",
      "quantity": 1,
      "order": 2
    }
  ],
  "default_rental_price": 500.00
}

Response 201:
{
  "id": "combination_uuid",
  ...
}
```

#### 16. Equipment-Kombination aktualisieren
```
PATCH /api/v1/equipment-combinations/{combination_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "default_rental_price": 550.00,
  "items": [...]
}

Response 200:
{
  "id": "combination_uuid",
  ...
}
```

#### 17. Equipment-Kombination löschen
```
DELETE /api/v1/equipment-combinations/{combination_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 204: No Content
```

### Defect Management

#### 18. Defekte melden
```
POST /api/v1/equipment/{equipment_id}/defects
Authorization: Bearer <access_token>
Required Role: Lager, Projektleiter, Admin

Request:
{
  "title": "Display zerkratzt",
  "description": "Kleine Kratzer auf dem Linsenscutz",
  "severity": "low|medium|high|critical",
  "reported_by": "user_uuid",
  "location": "Lager",
  "action_required": "Reinigung|Reparatur|Austausch|Dokumentation",
  "photos": ["file_uuid_1", "file_uuid_2"],
  "tags": ["Display", "Oberflächenschaden"]
}

Response 201:
{
  "id": "defect_uuid",
  "equipment_id": "equipment_uuid",
  "title": "Display zerkratzt",
  "description": "Kleine Kratzer auf dem Linsenscutz",
  "severity": "low",
  "status": "offen",
  "reported_by_name": "Max Mustermann",
  "reported_at": "2026-03-20T14:30:00Z",
  "location": "Lager",
  "action_required": "Reinigung",
  "photos": [...],
  "tags": ["Display", "Oberflächenschaden"],
  "created_at": "2026-03-20T14:30:00Z"
}
```

#### 19. Defekte auflisten
```
GET /api/v1/equipment/{equipment_id}/defects
Authorization: Bearer <access_token>

Query Parameters:
- filter[status]=offen|in_progress|resolved|closed
- filter[severity]=low|medium|high|critical
- sort_by=created_at|severity
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "defect_uuid",
      "equipment_id": "equipment_uuid",
      "title": "Display zerkratzt",
      ...
    }
  ],
  "pagination": { ... }
}
```

#### 20. Defekt aktualisieren
```
PATCH /api/v1/equipment/defects/{defect_id}
Authorization: Bearer <access_token>
Required Role: Lager, Admin

Request:
{
  "status": "resolved",
  "resolution_notes": "Objektiv gereinigt und kalibriert",
  "resolved_by": "user_uuid",
  "resolved_at": "2026-03-20T15:00:00Z"
}

Response 200:
{
  "id": "defect_uuid",
  ...
}
```

#### 21. Defekt löschen
```
DELETE /api/v1/equipment/defects/{defect_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 204: No Content
```

### Equipment Availability

#### 22. Verfügbarkeit abfragen
```
GET /api/v1/equipment/{equipment_id}/availability
Authorization: Bearer <access_token>

Query Parameters:
- from_date=2026-03-25
- to_date=2026-03-30
- exclude_project_id=uuid (optional)

Response 200:
{
  "equipment_id": "equipment_uuid",
  "equipment_name": "Beamer Sony VPL-FHZ100",
  "is_available": true,
  "available_from": "2026-03-20T14:30:00Z",
  "available_until": "2026-04-05T18:00:00Z",
  "requested_period": {
    "from": "2026-03-25T08:00:00Z",
    "to": "2026-03-30T22:00:00Z",
    "is_available_in_period": true
  },
  "current_assignment": {
    "project_id": "uuid",
    "project_name": "Konferenz TechWorld 2026",
    "start_date": "2026-03-20T08:00:00Z",
    "end_date": "2026-03-23T22:00:00Z"
  },
  "bookings": [
    {
      "project_id": "uuid",
      "project_name": "Festival Summer 2026",
      "start_date": "2026-07-15T08:00:00Z",
      "end_date": "2026-07-18T22:00:00Z"
    }
  ]
}
```

#### 23. Verfügbarkeit für mehrere Geräte
```
POST /api/v1/equipment/availability/check-multiple
Authorization: Bearer <access_token>

Request:
{
  "equipment_ids": ["uuid1", "uuid2", "uuid3"],
  "from_date": "2026-03-25",
  "to_date": "2026-03-30"
}

Response 200:
{
  "requested_period": {
    "from": "2026-03-25T08:00:00Z",
    "to": "2026-03-30T22:00:00Z"
  },
  "results": [
    {
      "equipment_id": "uuid1",
      "equipment_name": "Beamer 1",
      "is_available": true
    },
    {
      "equipment_id": "uuid2",
      "equipment_name": "Beamer 2",
      "is_available": false
    }
  ],
  "summary": {
    "total_checked": 3,
    "available": 2,
    "unavailable": 1
  }
}
```

---

## Projekte Module

### 1. Projekte auflisten
```
GET /api/v1/projects
Authorization: Bearer <access_token>

Query Parameters:
- filter[status]=Anfrage|Geplant|Aufbau|Live|Abbau|Abgeschlossen|Abgerechnet
- filter[client_id]=uuid
- filter[project_manager_id]=uuid
- search=project_name|client_name
- sort_by=name|created_at|start_date|end_date
- sort_order=asc|desc
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "project_uuid",
      "name": "Konferenz TechWorld 2026",
      "client_id": "client_uuid",
      "client_name": "TechWorld GmbH",
      "project_manager_id": "user_uuid",
      "project_manager_name": "Max Mustermann",
      "description": "Technische Konferenz für IT-Professionals",
      "status": "Geplant",
      "start_date": "2026-03-25T08:00:00Z",
      "end_date": "2026-03-27T22:00:00Z",
      "location": "Berlin Convention Center",
      "budget": 25000.00,
      "current_cost": 18500.00,
      "equipment_count": 12,
      "freelancer_count": 3,
      "notes": "VIP-Bereich mit zusätzlicher Beleuchtung",
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-03-20T14:30:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 2. Projekt erstellen
```
POST /api/v1/projects
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "name": "Konferenz TechWorld 2026",
  "client_id": "client_uuid",
  "project_manager_id": "user_uuid",
  "description": "Technische Konferenz für IT-Professionals",
  "status": "Anfrage",
  "start_date": "2026-03-25T08:00:00Z",
  "end_date": "2026-03-27T22:00:00Z",
  "location": "Berlin Convention Center",
  "budget": 25000.00,
  "notes": "VIP-Bereich mit zusätzlicher Beleuchtung"
}

Response 201:
{
  "id": "project_uuid",
  "name": "Konferenz TechWorld 2026",
  ...
}
```

### 3. Projekt abrufen
```
GET /api/v1/projects/{project_id}
Authorization: Bearer <access_token>

Response 200:
{
  "id": "project_uuid",
  "name": "Konferenz TechWorld 2026",
  "equipment": [
    {
      "assignment_id": "assignment_uuid",
      "equipment_id": "equipment_uuid",
      "equipment_name": "Beamer Sony VPL-FHZ100",
      "quantity": 2,
      "status": "zugewiesen",
      "notes": "Hauptbeamer für Auditorium 1 und 2"
    }
  ],
  "freelancers": [
    {
      "assignment_id": "assignment_uuid",
      "freelancer_id": "user_uuid",
      "name": "Anna Freelancer",
      "role": "Elektrikerin",
      "daily_rate": 350.00,
      "start_date": "2026-03-25T08:00:00Z",
      "end_date": "2026-03-27T22:00:00Z"
    }
  ],
  "invoices": [
    {
      "id": "invoice_uuid",
      "invoice_number": "REC-2026-001",
      "status": "Entwurf",
      "total": 5000.00
    }
  ],
  ...
}
```

### 4. Projekt aktualisieren
```
PATCH /api/v1/projects/{project_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "name": "Konferenz TechWorld 2026 (erweitert)",
  "status": "Geplant",
  "budget": 28000.00,
  "notes": "Zusätzlicher Workshop-Bereich"
}

Response 200:
{
  "id": "project_uuid",
  ...
}
```

### 5. Projekt löschen
```
DELETE /api/v1/projects/{project_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 204: No Content
```

### Project Equipment Assignment

### 6. Equipment zu Projekt hinzufügen
```
POST /api/v1/projects/{project_id}/equipment
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "equipment_id": "equipment_uuid",
  "quantity": 2,
  "assignment_type": "Rental|Loan",
  "notes": "Hauptbeamer für Auditorium 1 und 2",
  "from_combination": false,
  "combination_id": null
}

Response 201:
{
  "assignment_id": "assignment_uuid",
  "equipment_id": "equipment_uuid",
  "equipment_name": "Beamer Sony VPL-FHZ100",
  "quantity": 2,
  "assignment_type": "Rental",
  "status": "zugewiesen",
  "notes": "Hauptbeamer für Auditorium 1 und 2"
}

Response 409:
{
  "error": {
    "code": "EQUIPMENT_NOT_AVAILABLE",
    "message": "Equipment ist im gewünschten Zeitraum nicht verfügbar",
    "details": {
      "equipment_id": "uuid",
      "conflict_with_project": "project_name"
    }
  }
}
```

### 7. Equipment aus Projekt entfernen
```
DELETE /api/v1/projects/{project_id}/equipment/{assignment_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Response 204: No Content
```

### 8. Projekt-Equipment-Assignment aktualisieren
```
PATCH /api/v1/projects/{project_id}/equipment/{assignment_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "quantity": 3,
  "notes": "Zusätzlicher Beamer hinzugefügt"
}

Response 200:
{
  "assignment_id": "assignment_uuid",
  ...
}
```

### Project Freelancer Assignment

### 9. Freelancer zu Projekt hinzufügen
```
POST /api/v1/projects/{project_id}/freelancers
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "freelancer_id": "user_uuid",
  "role": "Elektrikerin",
  "daily_rate": 350.00,
  "start_date": "2026-03-25T08:00:00Z",
  "end_date": "2026-03-27T22:00:00Z",
  "notes": "Verantwortlich für gesamte Elektrik"
}

Response 201:
{
  "assignment_id": "assignment_uuid",
  "freelancer_id": "user_uuid",
  "name": "Anna Freelancer",
  "role": "Elektrikerin",
  "daily_rate": 350.00,
  "start_date": "2026-03-25T08:00:00Z",
  "end_date": "2026-03-27T22:00:00Z"
}
```

### 10. Freelancer aus Projekt entfernen
```
DELETE /api/v1/projects/{project_id}/freelancers/{assignment_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Response 204: No Content
```

### 11. Freelancer-Assignment aktualisieren
```
PATCH /api/v1/projects/{project_id}/freelancers/{assignment_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "daily_rate": 400.00,
  "end_date": "2026-03-28T10:00:00Z"
}

Response 200:
{
  "assignment_id": "assignment_uuid",
  ...
}
```

### Packing Lists

### 12. Packliste generieren
```
POST /api/v1/projects/{project_id}/packing-list/generate
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Lager

Request:
{
  "optimize_by_location": true,
  "include_defect_items": false,
  "format": "optimized|alphabetical|by_type"
}

Response 200:
{
  "packing_list_id": "packing_list_uuid",
  "project_id": "project_uuid",
  "project_name": "Konferenz TechWorld 2026",
  "generated_at": "2026-03-20T14:30:00Z",
  "total_items": 24,
  "grouped_by": "location",
  "items": [
    {
      "group_id": "location_uuid",
      "group_name": "Lager / Regal A / Fach 3",
      "items": [
        {
          "packing_item_id": "item_uuid",
          "equipment_id": "equipment_uuid",
          "equipment_name": "Beamer Sony VPL-FHZ100",
          "serial_number": "BM-001-2024",
          "quantity": 2,
          "barcode": "EQ-BM-001-2024",
          "status": "pending",
          "scanned": false,
          "scanned_at": null
        }
      ]
    }
  ]
}
```

### 13. Packliste abrufen
```
GET /api/v1/packing-lists/{packing_list_id}
Authorization: Bearer <access_token>

Response 200:
{
  "packing_list_id": "packing_list_uuid",
  "project_id": "project_uuid",
  "project_name": "Konferenz TechWorld 2026",
  "generated_at": "2026-03-20T14:30:00Z",
  "completion_status": "85%",
  "items": [...]
}
```

### 14. Packliste drucken/exportieren
```
GET /api/v1/packing-lists/{packing_list_id}/export
Authorization: Bearer <access_token>

Query Parameters:
- format=pdf|csv|excel
- print_barcodes=true|false
- group_by=location|type|equipment_name

Response:
Content-Type: application/pdf|text/csv|application/vnd.ms-excel
Content-Disposition: attachment; filename="packlist.pdf"
```

### 15. Packlisten auflisten für Projekt
```
GET /api/v1/projects/{project_id}/packing-lists
Authorization: Bearer <access_token>

Query Parameters:
- status=active|completed|archived
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "packing_list_id": "packing_list_uuid",
      "project_id": "project_uuid",
      "generated_at": "2026-03-20T14:30:00Z",
      "total_items": 24,
      "scanned_items": 20,
      "completion_percentage": 83.33,
      "status": "active"
    }
  ],
  "pagination": { ... }
}
```

---

## Angebote & Rechnungen Module

### Offers / Quotations

### 1. Angebote auflisten
```
GET /api/v1/offers
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Query Parameters:
- filter[status]=Entwurf|Gesendet|Akzeptiert|Abgelehnt|Storniert
- filter[project_id]=uuid
- filter[client_id]=uuid
- search=offer_number|client_name
- sort_by=created_at|offer_number|total
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "offer_uuid",
      "offer_number": "ANG-2026-001",
      "project_id": "project_uuid",
      "project_name": "Konferenz TechWorld 2026",
      "client_id": "client_uuid",
      "client_name": "TechWorld GmbH",
      "client_contact": "Max Mustermann",
      "client_email": "max@techworld.de",
      "status": "Gesendet",
      "valid_until": "2026-04-20T23:59:59Z",
      "subtotal": 18500.00,
      "tax_amount": 3515.00,
      "total": 22015.00,
      "items_count": 12,
      "created_at": "2026-03-15T10:00:00Z",
      "created_by": "user_name",
      "sent_at": "2026-03-16T09:00:00Z",
      "updated_at": "2026-03-20T14:30:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 2. Angebot erstellen
```
POST /api/v1/offers
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "project_id": "project_uuid",
  "client_id": "client_uuid",
  "client_contact": "Max Mustermann",
  "client_email": "max@techworld.de",
  "valid_days": 31,
  "notes": "Angebot ist 31 Tage gültig",
  "items": [
    {
      "description": "Beamer Setup (2x Sony VPL-FHZ100)",
      "equipment_type_id": "type_uuid",
      "equipment_id": "equipment_uuid",
      "quantity": 2,
      "unit": "Stück",
      "unit_price": 500.00,
      "discount_percent": 5.0,
      "notes": "10% Mengenrabatt verhandelt"
    },
    {
      "description": "Elektrik-Service",
      "unit": "Stunden",
      "quantity": 16,
      "unit_price": 75.00,
      "discount_percent": 0.0
    }
  ]
}

Response 201:
{
  "id": "offer_uuid",
  "offer_number": "ANG-2026-001",
  "project_id": "project_uuid",
  "status": "Entwurf",
  "items": [...],
  "subtotal": 18500.00,
  "tax_amount": 3515.00,
  "total": 22015.00,
  ...
}
```

### 3. Angebot abrufen
```
GET /api/v1/offers/{offer_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Response 200:
{
  "id": "offer_uuid",
  "offer_number": "ANG-2026-001",
  "items": [
    {
      "id": "item_uuid",
      "description": "Beamer Setup (2x Sony VPL-FHZ100)",
      "equipment_type_id": "type_uuid",
      "equipment_id": "equipment_uuid",
      "quantity": 2,
      "unit": "Stück",
      "unit_price": 500.00,
      "discount_percent": 5.0,
      "line_total": 950.00,
      "notes": "10% Mengenrabatt"
    }
  ],
  "subtotal": 18500.00,
  "tax_rate": 19.0,
  "tax_amount": 3515.00,
  "total": 22015.00,
  ...
}
```

### 4. Angebot aktualisieren
```
PATCH /api/v1/offers/{offer_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "status": "Entwurf",
  "valid_days": 45,
  "notes": "Aktualisiertes Angebot mit zusätzlichen Services",
  "items": [...]
}

Response 200:
{
  "id": "offer_uuid",
  ...
}
```

### 5. Angebot versenden
```
POST /api/v1/offers/{offer_id}/send
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "recipient_email": "max@techworld.de",
  "message": "Anbei finden Sie unser aktualisiertes Angebot.",
  "send_copy_to": ["boss@company.de"]
}

Response 200:
{
  "id": "offer_uuid",
  "status": "Gesendet",
  "sent_at": "2026-03-20T14:30:00Z",
  "sent_to": "max@techworld.de"
}
```

### 6. Angebot als PDF exportieren
```
GET /api/v1/offers/{offer_id}/pdf
Authorization: Bearer <access_token>

Response:
Content-Type: application/pdf
Content-Disposition: attachment; filename="ANG-2026-001.pdf"
```

### 7. Angebot in Rechnung konvertieren
```
POST /api/v1/offers/{offer_id}/convert-to-invoice
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "invoice_date": "2026-03-20",
  "due_date": "2026-04-20",
  "notes": "Rechnung erstellt aus Angebot ANG-2026-001"
}

Response 201:
{
  "id": "invoice_uuid",
  "invoice_number": "REC-2026-001",
  "offer_id": "offer_uuid",
  "status": "Entwurf",
  "items": [...],
  ...
}
```

### 8. Angebot stornieren
```
POST /api/v1/offers/{offer_id}/cancel
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "reason": "Kunde hat Auftrag storniert"
}

Response 200:
{
  "id": "offer_uuid",
  "status": "Storniert",
  "cancelled_at": "2026-03-20T14:30:00Z",
  "cancellation_reason": "Kunde hat Auftrag storniert"
}
```

### Invoices

### 9. Rechnungen auflisten
```
GET /api/v1/invoices
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Query Parameters:
- filter[status]=Entwurf|Versendet|Bezahlt|Teilbezahlt|Überfällig|Storniert
- filter[project_id]=uuid
- filter[client_id]=uuid
- filter[payment_status]=Ausstehend|Bezahlt|Teilbezahlt|Überfällig
- search=invoice_number|client_name
- sort_by=created_at|invoice_number|due_date|total
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "invoice_uuid",
      "invoice_number": "REC-2026-001",
      "project_id": "project_uuid",
      "project_name": "Konferenz TechWorld 2026",
      "client_id": "client_uuid",
      "client_name": "TechWorld GmbH",
      "client_contact": "Max Mustermann",
      "status": "Versendet",
      "payment_status": "Teilbezahlt",
      "invoice_date": "2026-03-20",
      "due_date": "2026-04-20",
      "subtotal": 18500.00,
      "tax_amount": 3515.00,
      "total": 22015.00,
      "paid_amount": 11007.50,
      "remaining_amount": 11007.50,
      "items_count": 12,
      "sent_at": "2026-03-20T14:30:00Z",
      "created_at": "2026-03-20T10:00:00Z",
      "updated_at": "2026-03-20T14:30:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 10. Rechnung erstellen
```
POST /api/v1/invoices
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "project_id": "project_uuid",
  "client_id": "client_uuid",
  "client_contact": "Max Mustermann",
  "client_email": "max@techworld.de",
  "invoice_date": "2026-03-20",
  "due_date": "2026-04-20",
  "offer_id": "offer_uuid",
  "notes": "Rechnung für Konferenz TechWorld 2026",
  "items": [
    {
      "description": "Beamer Setup (2x Sony VPL-FHZ100)",
      "equipment_type_id": "type_uuid",
      "quantity": 2,
      "unit": "Stück",
      "unit_price": 500.00,
      "discount_percent": 5.0
    }
  ]
}

Response 201:
{
  "id": "invoice_uuid",
  "invoice_number": "REC-2026-001",
  "status": "Entwurf",
  "items": [...],
  "subtotal": 18500.00,
  "tax_amount": 3515.00,
  "total": 22015.00,
  ...
}
```

### 11. Rechnung abrufen
```
GET /api/v1/invoices/{invoice_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Response 200:
{
  "id": "invoice_uuid",
  "invoice_number": "REC-2026-001",
  "items": [...],
  "subtotal": 18500.00,
  "tax_rate": 19.0,
  "tax_amount": 3515.00,
  "total": 22015.00,
  "paid_amount": 0.00,
  "remaining_amount": 22015.00,
  "payment_history": [
    {
      "id": "payment_uuid",
      "amount": 11007.50,
      "payment_date": "2026-03-25",
      "method": "Überweisung",
      "reference": "REC-2026-001-1"
    }
  ],
  ...
}
```

### 12. Rechnung aktualisieren
```
PATCH /api/v1/invoices/{invoice_id}
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "status": "Entwurf",
  "due_date": "2026-04-25",
  "notes": "Zahlungsziel verlängert"
}

Response 200:
{
  "id": "invoice_uuid",
  ...
}
```

### 13. Rechnung versenden
```
POST /api/v1/invoices/{invoice_id}/send
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "recipient_email": "max@techworld.de",
  "message": "Anbei finden Sie unsere Rechnung.",
  "send_copy_to": ["accounting@company.de"]
}

Response 200:
{
  "id": "invoice_uuid",
  "status": "Versendet",
  "sent_at": "2026-03-20T14:30:00Z"
}
```

### 14. Rechnung als PDF exportieren
```
GET /api/v1/invoices/{invoice_id}/pdf
Authorization: Bearer <access_token>

Response:
Content-Type: application/pdf
Content-Disposition: attachment; filename="REC-2026-001.pdf"
```

### 15. Zahlungseingänge registrieren
```
POST /api/v1/invoices/{invoice_id}/payments
Authorization: Bearer <access_token>
Required Role: Buchhaltung, Admin

Request:
{
  "amount": 11007.50,
  "payment_date": "2026-03-25",
  "method": "Überweisung",
  "reference": "REC-2026-001-1",
  "notes": "Teilzahlung erhalten"
}

Response 201:
{
  "id": "payment_uuid",
  "invoice_id": "invoice_uuid",
  "amount": 11007.50,
  "payment_date": "2026-03-25",
  "method": "Überweisung",
  "reference": "REC-2026-001-1",
  "remaining_amount": 11007.50
}
```

### 16. Zahlungseingang stornieren
```
DELETE /api/v1/invoices/{invoice_id}/payments/{payment_id}
Authorization: Bearer <access_token>
Required Role: Admin, Buchhaltung

Response 204: No Content
```

### 17. Teilrechnung erstellen
```
POST /api/v1/invoices/{invoice_id}/partial-invoice
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "allocation_percent": 50.0,
  "description": "Erste Teilrechnung - 50% des Gesamtauftrags"
}

Response 201:
{
  "id": "partial_invoice_uuid",
  "invoice_number": "REC-2026-001-1",
  "parent_invoice_id": "invoice_uuid",
  "status": "Entwurf",
  "items": [...],
  "subtotal": 9250.00,
  "tax_amount": 1757.50,
  "total": 11007.50,
  ...
}
```

### 18. Gutschrift erstellen
```
POST /api/v1/invoices/{invoice_id}/credit-note
Authorization: Bearer <access_token>
Required Role: Admin, Buchhaltung

Request:
{
  "amount": 500.00,
  "reason": "Reklamation Beamer 2",
  "items": [
    {
      "description": "Rückgutschrift Beamer Setup",
      "quantity": 0.5,
      "unit_price": 500.00
    }
  ]
}

Response 201:
{
  "id": "credit_note_uuid",
  "invoice_id": "invoice_uuid",
  "credit_note_number": "GUT-2026-001",
  "total": 500.00,
  "reason": "Reklamation Beamer 2",
  "created_at": "2026-03-20T14:30:00Z"
}
```

### Dunning / Reminders

### 19. Mahnungen auflisten
```
GET /api/v1/dunning/reminders
Authorization: Bearer <access_token>
Required Role: Buchhaltung, Admin

Query Parameters:
- filter[status]=Pending|Sent|Paid|Overdue
- filter[level]=1|2|3
- sort_by=due_date|created_at
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "dunning_uuid",
      "invoice_id": "invoice_uuid",
      "invoice_number": "REC-2026-001",
      "client_name": "TechWorld GmbH",
      "level": 1,
      "status": "Sent",
      "amount_due": 11007.50,
      "original_due_date": "2026-04-20",
      "days_overdue": 5,
      "sent_at": "2026-03-20T14:30:00Z",
      "next_reminder_date": "2026-04-10",
      "created_at": "2026-03-20T10:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 20. Mahnung versenden
```
POST /api/v1/dunning/send-reminder
Authorization: Bearer <access_token>
Required Role: Buchhaltung, Admin

Request:
{
  "invoice_id": "invoice_uuid",
  "level": 1,
  "recipient_email": "max@techworld.de",
  "message": "Freundliche Erinnerung: Bitte begleichen Sie Ihre ausstehende Rechnung."
}

Response 200:
{
  "id": "dunning_uuid",
  "status": "Sent",
  "sent_at": "2026-03-20T14:30:00Z",
  "next_reminder_level": 2
}
```

### 21. Mahnwesen-Konfiguration auslesen (Admin only)
```
GET /api/v1/settings/dunning-config
Authorization: Bearer <access_token>
Required Role: Admin

Response 200:
{
  "levels": [
    {
      "level": 1,
      "days_after_due_date": 5,
      "template": "Freundliche Erinnerung",
      "enabled": true
    },
    {
      "level": 2,
      "days_after_due_date": 14,
      "template": "Zweite Mahnung",
      "enabled": true
    },
    {
      "level": 3,
      "days_after_due_date": 28,
      "template": "Letzte Mahnung vor rechtlichen Schritten",
      "enabled": true
    }
  ],
  "auto_send": true,
  "auto_send_interval_days": 5
}
```

### 22. DATEV-Export
```
GET /api/v1/export/datev
Authorization: Bearer <access_token>
Required Role: Buchhaltung, Admin

Query Parameters:
- from_date=2026-01-01
- to_date=2026-03-31
- format=csv|xml

Response:
Content-Type: text/csv|application/xml
Content-Disposition: attachment; filename="DATEV_2026_Q1.csv"
[DATEV-compliant export data]
```

---

## Sub-Rental Module

### 1. Sub-Rental Anfragen auflisten
```
GET /api/v1/sub-rentals
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Query Parameters:
- filter[status]=Angefragt|Bestätigt|Ausgegeben|Zurückgegeben|Storniert
- filter[partner_id]=uuid
- filter[direction]=Outgoing|Incoming
- search=project_name|partner_name
- sort_by=created_at|due_date
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "sub_rental_uuid",
      "request_number": "SR-2026-001",
      "direction": "Outgoing",
      "project_id": "project_uuid",
      "project_name": "Konferenz TechWorld 2026",
      "partner_id": "partner_uuid",
      "partner_name": "Licht-Partner GmbH",
      "status": "Bestätigt",
      "total_items": 5,
      "request_date": "2026-03-15",
      "pickup_date": "2026-03-24",
      "return_date": "2026-03-28",
      "daily_rate": 250.00,
      "total_cost": 1000.00,
      "created_at": "2026-03-15T10:00:00Z",
      "updated_at": "2026-03-20T14:30:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 2. Sub-Rental Anfrage erstellen
```
POST /api/v1/sub-rentals
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "project_id": "project_uuid",
  "partner_id": "partner_uuid",
  "direction": "Outgoing",
  "items": [
    {
      "equipment_id": "equipment_uuid",
      "equipment_name": "Scheinwerfer LED 200W",
      "quantity": 5,
      "daily_rate": 50.00
    }
  ],
  "pickup_date": "2026-03-24",
  "return_date": "2026-03-28",
  "notes": "Für Workshop-Beleuchtung"
}

Response 201:
{
  "id": "sub_rental_uuid",
  "request_number": "SR-2026-001",
  "direction": "Outgoing",
  "status": "Angefragt",
  "items": [...],
  "total_cost": 1000.00,
  ...
}
```

### 3. Sub-Rental Anfrage abrufen
```
GET /api/v1/sub-rentals/{sub_rental_id}
Authorization: Bearer <access_token>

Response 200:
{
  "id": "sub_rental_uuid",
  "request_number": "SR-2026-001",
  "direction": "Outgoing",
  "project_id": "project_uuid",
  "partner_id": "partner_uuid",
  "status": "Bestätigt",
  "items": [
    {
      "id": "item_uuid",
      "equipment_id": "equipment_uuid",
      "equipment_name": "Scheinwerfer LED 200W",
      "quantity": 5,
      "daily_rate": 50.00,
      "total": 1000.00
    }
  ],
  "pickup_date": "2026-03-24",
  "return_date": "2026-03-28",
  "total_cost": 1000.00,
  "condition_photos_pickup": [...],
  "condition_photos_return": [...],
  "notes": "Für Workshop-Beleuchtung",
  ...
}
```

### 4. Sub-Rental Status aktualisieren
```
PATCH /api/v1/sub-rentals/{sub_rental_id}/status
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "status": "Bestätigt",
  "notes": "Partner hat Verfügbarkeit bestätigt"
}

Response 200:
{
  "id": "sub_rental_uuid",
  "status": "Bestätigt",
  "updated_at": "2026-03-20T14:30:00Z"
}
```

### 5. Sub-Rental Rückgabe dokumentieren
```
POST /api/v1/sub-rentals/{sub_rental_id}/document-return
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Lager

Request:
{
  "return_date": "2026-03-28",
  "condition": "good|acceptable|damaged",
  "return_notes": "Ein Scheinwerfer mit Kratzer auf dem Gehäuse",
  "condition_photos": ["file_uuid_1", "file_uuid_2"],
  "missing_items": [
    {
      "equipment_id": "equipment_uuid",
      "quantity": 1,
      "notes": "Stromkabel fehlt"
    }
  ]
}

Response 200:
{
  "id": "sub_rental_uuid",
  "status": "Zurückgegeben",
  "return_date": "2026-03-28",
  "condition": "damaged",
  "condition_photos": [...],
  "missing_items": [...]
}
```

### 6. Sub-Rental Ausgabe dokumentieren
```
POST /api/v1/sub-rentals/{sub_rental_id}/document-handover
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Lager

Request:
{
  "pickup_date": "2026-03-24",
  "pickup_location": "Unser Lager, Berlin",
  "condition": "new|good|acceptable",
  "condition_notes": "Alle Geräte in neuwertigem Zustand",
  "condition_photos": ["file_uuid_1", "file_uuid_2"],
  "handed_over_by": "user_uuid",
  "received_by": "Partner-Mitarbeiter Name"
}

Response 200:
{
  "id": "sub_rental_uuid",
  "status": "Ausgegeben",
  "handover_date": "2026-03-24",
  "condition": "new",
  "condition_photos": [...],
  "handed_over_by": "Max Mustermann"
}
```

### 7. Sub-Rental Rechnung generieren
```
POST /api/v1/sub-rentals/{sub_rental_id}/generate-invoice
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "invoice_date": "2026-03-28",
  "due_date": "2026-04-28"
}

Response 201:
{
  "id": "invoice_uuid",
  "invoice_number": "REC-SR-2026-001",
  "sub_rental_id": "sub_rental_uuid",
  "status": "Entwurf",
  "total": 1000.00
}
```

### 8. Sub-Rental stornieren
```
POST /api/v1/sub-rentals/{sub_rental_id}/cancel
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "reason": "Projekt verschoben"
}

Response 200:
{
  "id": "sub_rental_uuid",
  "status": "Storniert",
  "cancelled_at": "2026-03-20T14:30:00Z"
}
```

---

## Federation Module

### Partner Management

### 1. Partner auflisten
```
GET /api/v1/federation/partners
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Query Parameters:
- search=partner_name|location
- filter[status]=active|inactive
- filter[contract_status]=valid|expired|pending
- sort_by=name|created_at
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "partner_uuid",
      "name": "Licht-Partner GmbH",
      "location": "München",
      "contact_name": "Thomas Schmidt",
      "contact_email": "thomas@licht-partner.de",
      "contact_phone": "+49890123456",
      "status": "active",
      "contract_status": "valid",
      "contract_expires_at": "2027-12-31",
      "api_token_created_at": "2025-12-01T10:00:00Z",
      "rental_conditions": {
        "daily_rate_markup": 1.2,
        "minimum_rental_period_days": 1,
        "payment_terms_days": 30,
        "cancellation_buffer_hours": 24
      },
      "equipment_count": 45,
      "created_at": "2025-12-01T10:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 2. Partner erstellen (Admin only)
```
POST /api/v1/federation/partners
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "name": "Licht-Partner GmbH",
  "location": "München",
  "contact_name": "Thomas Schmidt",
  "contact_email": "thomas@licht-partner.de",
  "contact_phone": "+49890123456",
  "company_registration": "HRB 123456",
  "rental_conditions": {
    "daily_rate_markup": 1.2,
    "minimum_rental_period_days": 1,
    "payment_terms_days": 30,
    "cancellation_buffer_hours": 24
  }
}

Response 201:
{
  "id": "partner_uuid",
  "name": "Licht-Partner GmbH",
  ...
}
```

### 3. Partner abrufen
```
GET /api/v1/federation/partners/{partner_id}
Authorization: Bearer <access_token>

Response 200:
{
  "id": "partner_uuid",
  "name": "Licht-Partner GmbH",
  "location": "München",
  ...
}
```

### 4. Partner aktualisieren (Admin only)
```
PATCH /api/v1/federation/partners/{partner_id}
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "contact_email": "neuer@licht-partner.de",
  "rental_conditions": {
    "daily_rate_markup": 1.25
  }
}

Response 200:
{
  "id": "partner_uuid",
  ...
}
```

### 5. API-Token für Partner generieren (Admin only)
```
POST /api/v1/federation/partners/{partner_id}/generate-token
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "valid_days": 365,
  "name": "API Token 2026"
}

Response 201:
{
  "id": "api_token_uuid",
  "token": "fed_xyz123abc_secret",
  "token_display": "fed_xyz123abc_***",
  "created_at": "2026-03-20T14:30:00Z",
  "expires_at": "2027-03-20T14:30:00Z",
  "name": "API Token 2026"
}
```

### Partner Equipment Availability

### 6. Partner Equipment verfügbar abfragen
```
GET /api/v1/federation/partners/{partner_id}/equipment/availability
Authorization: Bearer <access_token>

Query Parameters:
- equipment_type_ids=uuid1,uuid2,uuid3
- from_date=2026-03-25
- to_date=2026-03-30

Response 200:
{
  "partner_id": "partner_uuid",
  "partner_name": "Licht-Partner GmbH",
  "query_period": {
    "from": "2026-03-25T08:00:00Z",
    "to": "2026-03-30T22:00:00Z"
  },
  "available_equipment": [
    {
      "type_id": "type_uuid",
      "type_name": "Scheinwerfer LED 200W",
      "available_quantity": 8,
      "daily_rate": 50.00,
      "description": "LED Scheinwerfer, dimmbar"
    }
  ],
  "response_time_ms": 250
}
```

### 7. Partner Equipment katalog
```
GET /api/v1/federation/partners/{partner_id}/equipment-catalog
Authorization: Bearer <access_token>

Query Parameters:
- search=equipment_name
- filter[category]=Lichttechnik|Projektoren|Sound
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "type_id": "type_uuid",
      "name": "Scheinwerfer LED 200W",
      "category": "Lichttechnik",
      "description": "LED Scheinwerfer, dimmbar, 200W",
      "daily_rate": 50.00,
      "minimum_rental_days": 1,
      "available": 12
    }
  ],
  "pagination": { ... }
}
```

### Partner Sub-Rental Requests

### 8. Sub-Rental Anfrage an Partner senden
```
POST /api/v1/federation/partners/{partner_id}/sub-rental-requests
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Request:
{
  "project_id": "project_uuid",
  "items": [
    {
      "equipment_type_id": "type_uuid",
      "equipment_type_name": "Scheinwerfer LED 200W",
      "quantity": 5,
      "daily_rate": 50.00
    }
  ],
  "pickup_date": "2026-03-24",
  "return_date": "2026-03-28",
  "notes": "Für Workshop-Beleuchtung, Abholung in Berlin"
}

Response 201:
{
  "id": "partner_request_uuid",
  "partner_id": "partner_uuid",
  "project_id": "project_uuid",
  "status": "Requested",
  "items": [...],
  "total_estimated_cost": 1000.00,
  "sent_at": "2026-03-20T14:30:00Z"
}
```

### 9. Partner Sub-Rental Anfragen auflisten
```
GET /api/v1/federation/partner-requests
Authorization: Bearer <access_token>

Query Parameters:
- filter[status]=Requested|Confirmed|Rejected|Cancelled
- filter[partner_id]=uuid
- sort_by=created_at|due_date
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "partner_request_uuid",
      "partner_id": "partner_uuid",
      "partner_name": "Licht-Partner GmbH",
      "project_id": "project_uuid",
      "status": "Confirmed",
      "confirmation_date": "2026-03-20T12:30:00Z",
      "items": [...],
      "total_cost": 1000.00
    }
  ],
  "pagination": { ... }
}
```

### 10. Partner Anfrage aktualisieren
```
PATCH /api/v1/federation/partner-requests/{request_id}
Authorization: Bearer <access_token>

Request:
{
  "status": "Confirmed",
  "confirmation_notes": "Alle Scheinwerfer verfügbar und bereit"
}

Response 200:
{
  "id": "partner_request_uuid",
  "status": "Confirmed",
  "confirmation_date": "2026-03-20T14:30:00Z"
}
```

---

## Admin Module

### User Management

### 1. Benutzer auflisten (Admin only)
```
GET /api/v1/admin/users
Authorization: Bearer <access_token>
Required Role: Admin

Query Parameters:
- search=email|name
- filter[status]=active|inactive|pending
- filter[role]=Admin|Projektleiter|Lager|Buchhaltung|Freelancer
- sort_by=name|email|created_at
- page=1
- per_page=20

Response 200:
{
  "data": [
    {
      "id": "user_uuid",
      "email": "user@company.de",
      "first_name": "Max",
      "last_name": "Mustermann",
      "status": "active",
      "roles": ["Projektleiter"],
      "last_login_at": "2026-03-20T14:00:00Z",
      "created_at": "2025-12-01T10:00:00Z",
      "created_by": "admin_name"
    }
  ],
  "pagination": { ... }
}
```

### 2. Benutzer abrufen (Admin only)
```
GET /api/v1/admin/users/{user_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 200:
{
  "id": "user_uuid",
  "email": "user@company.de",
  "first_name": "Max",
  "last_name": "Mustermann",
  "phone": "+49123456789",
  "status": "active",
  "roles": ["Projektleiter"],
  "custom_permissions": [...],
  "created_at": "2025-12-01T10:00:00Z",
  "updated_at": "2026-03-20T14:30:00Z"
}
```

### 3. Benutzer Rollen aktualisieren (Admin only)
```
PATCH /api/v1/admin/users/{user_id}/roles
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "roles": ["Projektleiter", "Buchhaltung"],
  "custom_permissions": [
    "project:read",
    "project:write",
    "invoice:read",
    "invoice:write"
  ]
}

Response 200:
{
  "id": "user_uuid",
  "roles": ["Projektleiter", "Buchhaltung"],
  "custom_permissions": [...]
}
```

### 4. Benutzer Status ändern (Admin only)
```
PATCH /api/v1/admin/users/{user_id}/status
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "status": "inactive",
  "reason": "Mitarbeiter ausgeschieden"
}

Response 200:
{
  "id": "user_uuid",
  "status": "inactive"
}
```

### 5. Benutzer löschen (Admin only)
```
DELETE /api/v1/admin/users/{user_id}
Authorization: Bearer <access_token>
Required Role: Admin

Response 204: No Content
```

### Company Settings

### 6. Firmen-Einstellungen abrufen (Admin only)
```
GET /api/v1/admin/settings/company
Authorization: Bearer <access_token>
Required Role: Admin

Response 200:
{
  "company_id": "company_uuid",
  "company_name": "Beispiel VT GmbH",
  "logo_url": "https://...",
  "address": "Hauptstraße 123",
  "postal_code": "10115",
  "city": "Berlin",
  "country": "Deutschland",
  "phone": "+4930123456",
  "email": "info@beispiel-vt.de",
  "website": "https://beispiel-vt.de",
  "tax_id": "DE123456789",
  "bank_name": "Commerzbank",
  "bank_account": "DE89370400440532013000",
  "bank_bic": "COBADEFFXXX",
  "vat_rate": 19.0,
  "standard_payment_terms_days": 30,
  "currency": "EUR",
  "language": "de",
  "timezone": "Europe/Berlin",
  "fiscal_year_start": "2026-01-01"
}
```

### 7. Firmen-Einstellungen aktualisieren (Admin only)
```
PATCH /api/v1/admin/settings/company
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "company_name": "Beispiel VT GmbH",
  "address": "Hauptstraße 124",
  "phone": "+4930123457",
  "tax_id": "DE123456789",
  "vat_rate": 19.0,
  "standard_payment_terms_days": 45
}

Response 200:
{
  "company_id": "company_uuid",
  ...
}
```

### 8. Firmen-Logo hochladen (Admin only)
```
POST /api/v1/admin/settings/company/logo
Authorization: Bearer <access_token>
Required Role: Admin
Content-Type: multipart/form-data

Request:
{
  "file": <binary image>
}

Response 200:
{
  "logo_url": "https://...",
  "updated_at": "2026-03-20T14:30:00Z"
}
```

### System Health & Maintenance

### 9. System Health Check
```
GET /api/v1/admin/health
Authorization: Bearer <access_token>
Required Role: Admin

Response 200:
{
  "status": "healthy",
  "timestamp": "2026-03-20T14:30:00Z",
  "components": {
    "database": {
      "status": "ok",
      "response_time_ms": 2
    },
    "storage": {
      "status": "ok",
      "used_gb": 45.3,
      "total_gb": 500.0
    },
    "api": {
      "status": "ok",
      "uptime_hours": 168
    },
    "cache": {
      "status": "ok",
      "hit_rate": 0.89
    }
  }
}
```

### 10. Backup Status (Admin only)
```
GET /api/v1/admin/backup/status
Authorization: Bearer <access_token>
Required Role: Admin

Response 200:
{
  "last_backup_at": "2026-03-20T03:00:00Z",
  "last_backup_size_gb": 2.5,
  "next_backup_scheduled_at": "2026-03-21T03:00:00Z",
  "backup_retention_days": 30,
  "backups": [
    {
      "id": "backup_uuid",
      "created_at": "2026-03-20T03:00:00Z",
      "size_gb": 2.5,
      "status": "completed",
      "location": "s3://backups/company-uuid/2026-03-20"
    }
  ]
}
```

### 11. Backup durchführen (Admin only)
```
POST /api/v1/admin/backup/create
Authorization: Bearer <access_token>
Required Role: Admin

Response 201:
{
  "backup_id": "backup_uuid",
  "status": "in_progress",
  "created_at": "2026-03-20T14:30:00Z"
}
```

### 12. Update Check (Admin only)
```
GET /api/v1/admin/updates/check
Authorization: Bearer <access_token>
Required Role: Admin

Response 200:
{
  "current_version": "1.2.5",
  "latest_version": "1.3.0",
  "update_available": true,
  "release_notes": "Neue Features, Bug Fixes, Performance Improvements",
  "download_url": "https://releases.example.com/v1.3.0",
  "breaking_changes": false,
  "estimated_downtime_minutes": 5
}
```

### 13. System Update durchführen (Admin only)
```
POST /api/v1/admin/updates/apply
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "version": "1.3.0",
  "backup_before_update": true,
  "maintenance_window_start": "2026-03-21T02:00:00Z"
}

Response 202:
{
  "update_id": "update_uuid",
  "status": "scheduled",
  "scheduled_start": "2026-03-21T02:00:00Z",
  "estimated_duration_minutes": 15
}
```

### 14. Audit Log (Admin only)
```
GET /api/v1/admin/audit-log
Authorization: Bearer <access_token>
Required Role: Admin

Query Parameters:
- filter[user_id]=uuid
- filter[resource_type]=project|equipment|invoice|user
- filter[action]=create|update|delete
- from_date=2026-03-01
- to_date=2026-03-31
- page=1
- per_page=50

Response 200:
{
  "data": [
    {
      "id": "audit_log_uuid",
      "timestamp": "2026-03-20T14:30:00Z",
      "user_id": "user_uuid",
      "user_name": "Max Mustermann",
      "action": "update",
      "resource_type": "invoice",
      "resource_id": "invoice_uuid",
      "changes": {
        "status": {
          "old": "Entwurf",
          "new": "Versendet"
        }
      },
      "ip_address": "192.168.1.1"
    }
  ],
  "pagination": { ... }
}
```

---

## Scanner Module

### 1. Barcode/QR-Code Lookup
```
POST /api/v1/scanner/lookup
Authorization: Bearer <access_token>
Required Role: Lager, Projektleiter, Admin

Request:
{
  "barcode": "EQ-12345",
  "context": "general|packing_list|project_return"
}

Response 200:
{
  "found": true,
  "equipment_id": "equipment_uuid",
  "equipment_name": "Beamer Sony VPL-FHZ100",
  "type_name": "Beamer",
  "serial_number": "BM-001-2024",
  "status": "aktiv",
  "storage_location": "Lager / Regal A / Fach 3",
  "is_available": true,
  "current_project": null,
  "defect_notes": null
}

Response 404:
{
  "found": false,
  "error": {
    "code": "BARCODE_NOT_FOUND",
    "message": "Barcode nicht gefunden"
  }
}
```

### 2. Scan-In für Projekt
```
POST /api/v1/scanner/scan-in
Authorization: Bearer <access_token>
Required Role: Lager, Admin

Request:
{
  "barcode": "EQ-12345",
  "project_id": "project_uuid",
  "location": "Laderampe Punkt 2",
  "scan_timestamp": "2026-03-20T14:30:00Z"
}

Response 200:
{
  "scan_event_id": "scan_event_uuid",
  "event": "checkin",
  "equipment_id": "equipment_uuid",
  "equipment_name": "Beamer Sony VPL-FHZ100",
  "project_id": "project_uuid",
  "scanned_at": "2026-03-20T14:30:00Z",
  "scanned_by": "Max Mustermann",
  "status": "success",
  "message": "Equipment erfolgreich eingescannt"
}
```

### 3. Scan-Out für Projekt
```
POST /api/v1/scanner/scan-out
Authorization: Bearer <access_token>
Required Role: Lager, Admin

Request:
{
  "barcode": "EQ-12345",
  "project_id": "project_uuid",
  "scan_timestamp": "2026-03-20T14:30:00Z",
  "notes": "Gerät in gutem Zustand zurück"
}

Response 200:
{
  "scan_event_id": "scan_event_uuid",
  "event": "checkout",
  "equipment_id": "equipment_uuid",
  "equipment_name": "Beamer Sony VPL-FHZ100",
  "project_id": "project_uuid",
  "scanned_at": "2026-03-20T14:30:00Z",
  "scanned_by": "Max Mustermann",
  "status": "success",
  "message": "Equipment erfolgreich ausgescannt"
}
```

### 4. Scan-Fehler behandeln
```
POST /api/v1/scanner/scan-error
Authorization: Bearer <access_token>
Required Role: Lager, Admin

Request:
{
  "barcode": "UNKNOWN-123",
  "error_type": "barcode_not_found|wrong_project|damaged",
  "project_id": "project_uuid",
  "notes": "Barcode nicht lesbar, Gerät beschädigt",
  "photos": ["file_uuid_1"]
}

Response 201:
{
  "error_id": "scan_error_uuid",
  "barcode": "UNKNOWN-123",
  "error_type": "barcode_not_found",
  "created_at": "2026-03-20T14:30:00Z",
  "assigned_to": "admin_user"
}
```

### 5. Scan-Historie für Equipment
```
GET /api/v1/scanner/equipment/{equipment_id}/history
Authorization: Bearer <access_token>

Query Parameters:
- from_date=2026-03-01
- to_date=2026-03-31
- filter[project_id]=uuid
- page=1
- per_page=50

Response 200:
{
  "equipment_id": "equipment_uuid",
  "equipment_name": "Beamer Sony VPL-FHZ100",
  "data": [
    {
      "scan_event_id": "scan_event_uuid",
      "event": "checkin",
      "project_id": "project_uuid",
      "project_name": "Konferenz TechWorld 2026",
      "scanned_at": "2026-03-20T08:00:00Z",
      "scanned_by": "Max Mustermann"
    },
    {
      "scan_event_id": "scan_event_uuid",
      "event": "checkout",
      "project_id": "project_uuid",
      "project_name": "Konferenz TechWorld 2026",
      "scanned_at": "2026-03-23T22:00:00Z",
      "scanned_by": "Anna Müller"
    }
  ],
  "pagination": { ... }
}
```

---

## Dashboard Module

### 1. Dashboard Overview
```
GET /api/v1/dashboard/overview
Authorization: Bearer <access_token>

Response 200:
{
  "summary": {
    "total_projects": 12,
    "active_projects": 3,
    "equipment_total": 156,
    "equipment_in_use": 45,
    "outstanding_invoices": 5,
    "outstanding_amount": 35000.00,
    "overdue_invoices": 1,
    "overdue_amount": 5000.00
  },
  "upcoming_projects": [
    {
      "id": "project_uuid",
      "name": "Festival Summer 2026",
      "start_date": "2026-07-15T08:00:00Z",
      "days_until_start": 117,
      "equipment_assigned": 34,
      "status": "Geplant"
    }
  ],
  "equipment_utilization": [
    {
      "type_name": "Beamer",
      "total_count": 8,
      "in_use": 5,
      "utilization_percent": 62.5
    }
  ],
  "cash_flow": {
    "received_this_month": 12500.00,
    "expected_this_month": 8000.00,
    "outstanding_total": 35000.00
  }
}
```

### 2. Projekt-Timeline
```
GET /api/v1/dashboard/project-timeline
Authorization: Bearer <access_token>

Query Parameters:
- from_date=2026-03-01
- to_date=2026-06-30
- filter[status]=Geplant|Aufbau|Live|Abbau|Abgeschlossen

Response 200:
{
  "timeline": [
    {
      "date": "2026-03-25",
      "events": [
        {
          "id": "project_uuid",
          "name": "Konferenz TechWorld 2026",
          "event_type": "project_start",
          "duration_days": 3,
          "equipment_count": 12,
          "status": "Geplant"
        }
      ]
    }
  ]
}
```

### 3. Equipment Bottlenecks
```
GET /api/v1/dashboard/bottlenecks
Authorization: Bearer <access_token>

Response 200:
{
  "bottlenecks": [
    {
      "type_name": "Beamer",
      "available_count": 1,
      "requested_count": 4,
      "conflict_projects": [
        {
          "project_id": "uuid1",
          "project_name": "Konferenz 1",
          "date": "2026-04-10"
        },
        {
          "project_id": "uuid2",
          "project_name": "Konferenz 2",
          "date": "2026-04-10"
        }
      ]
    }
  ],
  "actions_recommended": [
    "Sub-Rental für zusätzliche Beamer organisieren",
    "Konferenz 2 auf 2026-04-11 verschieben"
  ]
}
```

### 4. Financial Overview
```
GET /api/v1/dashboard/financials
Authorization: Bearer <access_token>
Required Role: Admin, Buchhaltung, Projektleiter

Response 200:
{
  "period": "2026-03",
  "revenue": {
    "invoiced": 45000.00,
    "paid": 38000.00,
    "outstanding": 7000.00,
    "overdue": 2000.00
  },
  "expenses": {
    "sub_rentals": 8500.00,
    "freelancers": 5200.00,
    "other": 1300.00,
    "total": 15000.00
  },
  "profit": 23000.00,
  "profit_margin": 51.11,
  "projects_this_month": 4,
  "average_project_value": 11250.00,
  "payment_compliance": {
    "on_time_percent": 85,
    "average_days_to_pay": 18
  }
}
```

---

## Zeiterfassung Module

### 1. Check-In für Freelancer
```
POST /api/v1/timesheets/check-in
Authorization: Bearer <access_token>
Required Role: Freelancer, Admin

Request:
{
  "project_id": "project_uuid",
  "location": "Berlin Convention Center",
  "notes": "Elektriker angekommen"
}

Response 201:
{
  "timesheet_entry_id": "entry_uuid",
  "freelancer_id": "user_uuid",
  "project_id": "project_uuid",
  "check_in_at": "2026-03-25T08:15:00Z",
  "check_out_at": null,
  "duration_hours": null,
  "status": "checked_in",
  "location": "Berlin Convention Center"
}
```

### 2. Check-Out für Freelancer
```
POST /api/v1/timesheets/check-out
Authorization: Bearer <access_token>
Required Role: Freelancer, Admin

Request:
{
  "project_id": "project_uuid",
  "notes": "Alle Elektrik abgebaut und getestet"
}

Response 200:
{
  "timesheet_entry_id": "entry_uuid",
  "freelancer_id": "user_uuid",
  "project_id": "project_uuid",
  "check_in_at": "2026-03-25T08:15:00Z",
  "check_out_at": "2026-03-25T18:30:00Z",
  "duration_hours": 10.25,
  "status": "checked_out",
  "earned_amount": 3587.50
}
```

### 3. Stundenübersicht für Freelancer
```
GET /api/v1/timesheets/my-hours
Authorization: Bearer <access_token>
Required Role: Freelancer

Query Parameters:
- from_date=2026-03-01
- to_date=2026-03-31
- filter[project_id]=uuid
- sort_by=date|project_name

Response 200:
{
  "freelancer_id": "user_uuid",
  "freelancer_name": "Anna Freelancer",
  "period": {
    "from": "2026-03-01",
    "to": "2026-03-31"
  },
  "entries": [
    {
      "date": "2026-03-25",
      "project_name": "Konferenz TechWorld 2026",
      "check_in_at": "2026-03-25T08:15:00Z",
      "check_out_at": "2026-03-25T18:30:00Z",
      "duration_hours": 10.25,
      "daily_rate": 350.00,
      "earned_amount": 3587.50
    }
  ],
  "summary": {
    "total_hours": 64.5,
    "total_earned": 22575.00,
    "project_count": 4
  }
}
```

### 4. Stundenübersicht für Projekt (Projektleiter/Admin)
```
GET /api/v1/projects/{project_id}/timesheets
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter

Query Parameters:
- filter[freelancer_id]=uuid
- sort_by=date|freelancer_name|hours

Response 200:
{
  "project_id": "project_uuid",
  "project_name": "Konferenz TechWorld 2026",
  "entries": [
    {
      "freelancer_id": "user_uuid",
      "freelancer_name": "Anna Freelancer",
      "date": "2026-03-25",
      "check_in_at": "2026-03-25T08:15:00Z",
      "check_out_at": "2026-03-25T18:30:00Z",
      "duration_hours": 10.25,
      "daily_rate": 350.00,
      "earned_amount": 3587.50
    }
  ],
  "summary": {
    "total_freelancers": 3,
    "total_hours": 32.75,
    "total_cost": 11462.50
  }
}
```

### 5. Stundenübersicht korrekt (Admin only)
```
PATCH /api/v1/timesheets/{entry_id}
Authorization: Bearer <access_token>
Required Role: Admin

Request:
{
  "check_in_at": "2026-03-25T08:00:00Z",
  "check_out_at": "2026-03-25T18:45:00Z",
  "notes": "Zeit korrigiert nach Projektleiter-Angaben"
}

Response 200:
{
  "entry_id": "entry_uuid",
  "duration_hours": 10.75,
  "earned_amount": 3762.50
}
```

### 6. Stundenabrechnung generieren
```
POST /api/v1/timesheets/generate-invoice
Authorization: Bearer <access_token>
Required Role: Admin, Projektleiter, Buchhaltung

Request:
{
  "project_id": "project_uuid",
  "period_from": "2026-03-01",
  "period_to": "2026-03-31",
  "invoice_date": "2026-04-01"
}

Response 201:
{
  "invoice_id": "invoice_uuid",
  "invoice_number": "REC-HOURS-2026-001",
  "project_id": "project_uuid",
  "total_hours": 32.75,
  "total_cost": 11462.50,
  "freelancer_breakdown": [
    {
      "freelancer_name": "Anna Freelancer",
      "hours": 20.5,
      "daily_rate": 350.00,
      "cost": 7175.00
    }
  ],
  "status": "Entwurf"
}
```

---

## Global Pagination Standard

Alle Endpunkte mit Liste verwenden folgendes Format:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8,
    "has_next_page": true,
    "has_previous_page": false
  }
}
```

---

## Global Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "status": 400,
    "timestamp": "2026-03-20T14:30:00Z",
    "request_id": "req_abc123xyz",
    "details": {
      "field": "value",
      "additional_info": "..."
    }
  }
}
```

---

## Rate Limiting Headers

Alle Responses beinhalten:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1711003600
```

---

## API Response Consistency

Alle Responses verwenden:
- ISO 8601 Timestamps
- Konsistente Fehlerformate
- Konsistente Pagination
- Consistent field naming (snake_case)
- UUIDs für alle IDs

---

**End of API Design Document**
