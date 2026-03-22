# Inventory Service API Reference

## Base URL
```
http://localhost:8002/api/v1
```

## Required Headers
All requests require:
```
X-Tenant-ID: <tenant-uuid>
Content-Type: application/json
```

## Equipment Endpoints

### Create Equipment
```
POST /equipment
Content-Type: application/json

{
  "name": "Shure SM58 Microphone",
  "description": "Dynamic microphone for vocals",
  "category_id": "cat_123",
  "sku": "SM58",
  "serial_number": "SN12345",
  "barcode": "BARCODE123",
  "purchase_date": "2024-01-15",
  "purchase_price": 99.99,
  "rental_price_day": 25.00,
  "rental_price_week": 150.00,
  "weight": 0.3,
  "dimensions": {
    "length": 10,
    "width": 5,
    "height": 5,
    "unit": "cm"
  },
  "location_id": "loc_456",
  "tags": ["audio", "microphone"],
  "custom_fields": {
    "brand": "Shure",
    "model": "SM58"
  }
}

Response: 201 Created
{
  "id": "equip_789",
  "tenant_id": "tenant_uuid",
  "name": "Shure SM58 Microphone",
  ...
}
```

### Get Equipment
```
GET /equipment/{id}

Response: 200 OK
{
  "id": "equip_789",
  "name": "Shure SM58 Microphone",
  ...
}
```

### List Equipment
```
GET /equipment?status=available&category_id=cat_123&location_id=loc_456&limit=20&offset=0

Query Parameters:
- status: available, reserved, checked_out, in_maintenance, damaged, retired
- category_id: filter by category
- location_id: filter by location
- limit: results per page (default: 20)
- offset: pagination offset (default: 0)

Response: 200 OK
{
  "data": [
    {
      "id": "equip_789",
      "name": "Shure SM58 Microphone",
      ...
    }
  ],
  "total": 150,
  "limit": 20,
  "offset": 0
}
```

### Search Equipment
```
GET /equipment/search?q=microphone&limit=20&offset=0

Query Parameters:
- q: search term (searches name, description, sku, barcode, serial_number)
- limit: results per page
- offset: pagination offset

Response: 200 OK
{
  "data": [
    {...},
    {...}
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

### Get Equipment by Barcode
```
GET /equipment/barcode/BARCODE123

Response: 200 OK
{
  "id": "equip_789",
  "barcode": "BARCODE123",
  ...
}
```

### Update Equipment
```
PUT /equipment/{id}
Content-Type: application/json

{
  "name": "Shure SM58 Microphone",
  "description": "Dynamic microphone for vocals (updated)",
  "category_id": "cat_123",
  "purchase_date": "2024-01-15",
  "purchase_price": 99.99,
  "rental_price_day": 25.00,
  "rental_price_week": 150.00,
  "weight": 0.3,
  "dimensions": {...},
  "tags": ["audio", "microphone", "updated"],
  "custom_fields": {...}
}

Response: 200 OK
{
  "id": "equip_789",
  "name": "Shure SM58 Microphone",
  ...
}
```

### Change Equipment Status
```
PATCH /equipment/{id}/status
Content-Type: application/json

{
  "status": "checked_out",
  "reason": "rented_to_customer"
}

Valid statuses:
- available
- reserved
- checked_out
- in_maintenance
- damaged
- retired

Response: 200 OK
{
  "message": "status updated"
}
```

### Update Equipment Condition
```
PATCH /equipment/{id}/condition
Content-Type: application/json

{
  "condition": "good"
}

Valid conditions:
- new
- good
- fair
- poor
- defective

Response: 200 OK
{
  "message": "condition updated"
}
```

### Add Equipment Image
```
POST /equipment/{id}/images
Content-Type: multipart/form-data

Form data:
- image: <file>

Response: 200 OK
{
  "id": "equip_789",
  "name": "Shure SM58 Microphone",
  "image_refs": ["image_ref_1", "image_ref_2"],
  ...
}
```

### Delete Equipment
```
DELETE /equipment/{id}

Response: 204 No Content
```

## Category Endpoints

### Create Category
```
POST /categories
Content-Type: application/json

{
  "name": "Audio Equipment",
  "parent_id": null,
  "icon": "speaker",
  "color": "#FF5733",
  "sort_order": 1
}

Response: 201 Created
{
  "id": "cat_123",
  "tenant_id": "tenant_uuid",
  "name": "Audio Equipment",
  ...
}
```

### List Categories
```
GET /categories

Response: 200 OK
[
  {
    "id": "cat_123",
    "name": "Audio Equipment",
    "parent_id": null,
    ...
  },
  {
    "id": "cat_456",
    "name": "Microphones",
    "parent_id": "cat_123",
    ...
  }
]
```

### Get Category
```
GET /categories/{id}

Response: 200 OK
{
  "id": "cat_123",
  "name": "Audio Equipment",
  ...
}
```

### Update Category
```
PUT /categories/{id}
Content-Type: application/json

{
  "name": "Audio Equipment (Updated)",
  "icon": "speaker",
  "color": "#FF5733",
  "sort_order": 1
}

Response: 200 OK
{
  "id": "cat_123",
  "name": "Audio Equipment (Updated)",
  ...
}
```

### Delete Category
```
DELETE /categories/{id}

Response: 204 No Content
```

## Flightcase Endpoints

### Create Flightcase
```
POST /flightcases
Content-Type: application/json

{
  "name": "Moving Heads Case",
  "description": "4x Claypaky Sharpy + Galgen",
  "barcode": "FLIGHTCASE001",
  "weight": 45.5
}

Response: 201 Created
{
  "id": "fc_789",
  "tenant_id": "tenant_uuid",
  "name": "Moving Heads Case",
  "contents": [],
  ...
}
```

### List Flightcases
```
GET /flightcases?limit=20&offset=0

Query Parameters:
- limit: results per page (default: 20)
- offset: pagination offset (default: 0)

Response: 200 OK
{
  "data": [
    {
      "id": "fc_789",
      "name": "Moving Heads Case",
      "contents": [
        {
          "equipment_id": "equip_123",
          "quantity": 4,
          "added_at": "2024-01-15T10:30:00Z"
        }
      ],
      ...
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

### Get Flightcase
```
GET /flightcases/{id}

Response: 200 OK
{
  "id": "fc_789",
  "name": "Moving Heads Case",
  "contents": [...],
  ...
}
```

### Update Flightcase
```
PUT /flightcases/{id}
Content-Type: application/json

{
  "name": "Moving Heads Case (Updated)",
  "description": "4x Claypaky Sharpy + Galgen (Updated)",
  "weight": 46.0
}

Response: 200 OK
{
  "id": "fc_789",
  "name": "Moving Heads Case (Updated)",
  ...
}
```

### Add Item to Flightcase
```
POST /flightcases/{id}/items
Content-Type: application/json

{
  "equipment_id": "equip_123",
  "quantity": 4
}

Response: 200 OK
{
  "id": "fc_789",
  "name": "Moving Heads Case",
  "contents": [
    {
      "equipment_id": "equip_123",
      "quantity": 4,
      "added_at": "2024-01-15T10:30:00Z"
    }
  ],
  ...
}
```

### Remove Item from Flightcase
```
DELETE /flightcases/{id}/items
Content-Type: application/json

{
  "equipment_id": "equip_123",
  "quantity": 2
}

Response: 200 OK
{
  "id": "fc_789",
  "name": "Moving Heads Case",
  "contents": [
    {
      "equipment_id": "equip_123",
      "quantity": 2,
      "added_at": "2024-01-15T10:30:00Z"
    }
  ],
  ...
}
```

### Delete Flightcase
```
DELETE /flightcases/{id}

Response: 204 No Content
```

## Health & Readiness

### Health Check
```
GET /health

Response: 200 OK
{
  "status": "healthy",
  "service": "inventory-service"
}
```

### Readiness Check
```
GET /ready

Response: 200 OK
{
  "status": "ready",
  "service": "inventory-service"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message describing the issue"
}
```

### Common HTTP Status Codes

| Status | Meaning |
|--------|---------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (success, no response body) |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (missing/invalid tenant ID) |
| 404 | Not Found |
| 409 | Conflict (e.g., duplicate barcode) |
| 500 | Internal Server Error |

### Example Error Response

```json
{
  "error": "barcode already exists"
}
```

## Pagination

List endpoints support pagination:
- `limit`: Number of results per page (default: 20)
- `offset`: Number of results to skip (default: 0)

Response includes:
- `data`: Array of results
- `total`: Total number of records
- `limit`: Requested limit
- `offset`: Requested offset

Example:
```
GET /equipment?limit=10&offset=20

Returns items 20-29 out of total count
```
