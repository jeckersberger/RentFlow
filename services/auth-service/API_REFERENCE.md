# Auth-Service API Reference

## Base URL
```
http://localhost:8001
```

## Authentication
Most endpoints require JWT authentication via the `Authorization` header:
```
Authorization: Bearer {accessToken}
```

---

## Health & Readiness

### GET /health
Always available health check.

**Response (200):**
```json
{
  "status": "healthy",
  "service": "auth-service",
  "timestamp": "2026-03-21T10:00:00Z"
}
```

### GET /ready
Database readiness check.

**Response (200):**
```json
{
  "status": "ready",
  "service": "auth-service"
}
```

**Response (503):**
```json
{
  "status": "not ready",
  "service": "auth-service",
  "reason": "database connection failed"
}
```

---

## Authentication Endpoints

### POST /api/v1/auth/register
Register a new user.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response (201):**
```json
{
  "data": {
    "id": "user-uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "roles": ["readonly"],
    "tenant_id": "tenant-uuid",
    "status": "active",
    "created_at": "2026-03-21T10:00:00Z",
    "updated_at": "2026-03-21T10:00:00Z"
  },
  "message": "User registered successfully"
}
```

**Error (409) - Email exists:**
```json
{
  "code": "EMAIL_EXISTS",
  "message": "Email already exists"
}
```

**Error (400) - Weak password:**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "password must contain at least one uppercase letter"
}
```

---

### POST /api/v1/auth/login
Authenticate user and receive tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200):**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "token_type": "Bearer"
  },
  "message": "Login successful"
}
```

**Cookies Set:**
- `refresh_token` (HttpOnly, Secure, SameSite=Lax, MaxAge=604800)

**Error (401) - Invalid credentials:**
```json
{
  "code": "INVALID_CREDENTIALS",
  "message": "Invalid email or password"
}
```

**Error (403) - User locked:**
```json
{
  "code": "USER_LOCKED",
  "message": "User account is locked"
}
```

---

### POST /api/v1/auth/refresh
Refresh access token using refresh token.

**Request Body (option 1 - via body):**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Request (option 2 - via cookie):**
Cookie: `refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

**Response (200):**
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900,
    "token_type": "Bearer"
  },
  "message": "Token refreshed successfully"
}
```

**Error (401) - Invalid refresh token:**
```json
{
  "code": "INVALID_REFRESH_TOKEN",
  "message": "Invalid or expired refresh token"
}
```

---

### POST /api/v1/auth/logout
Logout and invalidate tokens. **Requires authentication.**

**Response (204):**
No content.

---

## User Profile Endpoints

### GET /api/v1/auth/me
Get current authenticated user. **Requires authentication.**

**Response (200):**
```json
{
  "data": {
    "id": "user-uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "roles": ["admin", "manager"],
    "tenant_id": "tenant-uuid",
    "status": "active",
    "created_at": "2026-03-21T10:00:00Z",
    "updated_at": "2026-03-21T10:00:00Z"
  },
  "message": "User retrieved successfully"
}
```

---

### PUT /api/v1/auth/password
Change password. **Requires authentication.**

**Request Body:**
```json
{
  "old_password": "OldPass123!",
  "new_password": "NewPass456!"
}
```

**Response (200):**
```json
{
  "message": "Password changed successfully"
}
```

**Error (401) - Invalid current password:**
```json
{
  "code": "INVALID_CREDENTIALS",
  "message": "Current password is incorrect"
}
```

---

### PUT /api/v1/auth/profile
Update user profile. **Requires authentication.**

**Request Body:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith"
}
```

**Response (200):**
```json
{
  "data": {
    "id": "user-uuid",
    "email": "user@example.com",
    "first_name": "Jane",
    "last_name": "Smith",
    "roles": ["admin"],
    "tenant_id": "tenant-uuid",
    "status": "active",
    "created_at": "2026-03-21T10:00:00Z",
    "updated_at": "2026-03-21T10:00:00Z"
  },
  "message": "Profile updated successfully"
}
```

---

## User Management Endpoints

### GET /api/v1/users
List users in tenant. **Requires authentication + admin role.**

**Query Parameters:**
- `page` (optional, default: 1)
- `per_page` (optional, default: 20, max: 100)

**Response (200):**
```json
{
  "data": {
    "data": [
      {
        "id": "user-uuid-1",
        "email": "user1@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "roles": ["admin"],
        "tenant_id": "tenant-uuid",
        "status": "active",
        "created_at": "2026-03-21T10:00:00Z",
        "updated_at": "2026-03-21T10:00:00Z"
      }
    ],
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  },
  "message": "Users retrieved successfully"
}
```

**Error (403) - Insufficient permissions:**
```json
{
  "code": "FORBIDDEN",
  "message": "Admin role required"
}
```

---

### GET /api/v1/users/{id}
Get specific user. **Requires authentication.**

**Response (200):**
```json
{
  "data": {
    "id": "user-uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "roles": ["manager"],
    "tenant_id": "tenant-uuid",
    "status": "active",
    "created_at": "2026-03-21T10:00:00Z",
    "updated_at": "2026-03-21T10:00:00Z"
  },
  "message": "User retrieved successfully"
}
```

---

### PUT /api/v1/users/{id}/roles
Assign role to user. **Requires authentication + admin role.**

**Request Body:**
```json
{
  "role": "manager"
}
```

**Response (200):**
```json
{
  "message": "Role assigned successfully"
}
```

---

### DELETE /api/v1/users/{id}
Deactivate user. **Requires authentication + admin role.**

**Response (204):**
No content.

---

## Tenant Management Endpoints

### POST /api/v1/tenants
Create new tenant.

**Request Body:**
```json
{
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "default_language": "de",
  "currency": "EUR",
  "tax_rate": 19.0,
  "invoice_prefix": "ACM"
}
```

**Response (201):**
```json
{
  "data": {
    "id": "tenant-uuid",
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "status": "active",
    "default_language": "de",
    "currency": "EUR",
    "tax_rate": 19.0,
    "invoice_prefix": "ACM",
    "created_at": "2026-03-21T10:00:00Z",
    "updated_at": "2026-03-21T10:00:00Z"
  },
  "message": "Tenant created successfully"
}
```

---

### GET /api/v1/tenants/{id}
Get tenant details.

**Response (200):**
```json
{
  "data": {
    "id": "tenant-uuid",
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "status": "active",
    "default_language": "de",
    "currency": "EUR",
    "tax_rate": 19.0,
    "invoice_prefix": "ACM",
    "created_at": "2026-03-21T10:00:00Z",
    "updated_at": "2026-03-21T10:00:00Z"
  },
  "message": "Tenant retrieved successfully"
}
```

---

### PUT /api/v1/tenants/{id}
Update tenant.

**Request Body:**
```json
{
  "name": "Acme Corporation Updated",
  "slug": "acme-corp",
  "default_language": "en",
  "currency": "USD",
  "tax_rate": 8.0,
  "invoice_prefix": "ACM"
}
```

**Response (200):**
```json
{
  "data": {
    "id": "tenant-uuid",
    "name": "Acme Corporation Updated",
    "slug": "acme-corp",
    "status": "active",
    "default_language": "en",
    "currency": "USD",
    "tax_rate": 8.0,
    "invoice_prefix": "ACM",
    "created_at": "2026-03-21T10:00:00Z",
    "updated_at": "2026-03-21T10:00:00Z"
  },
  "message": "Tenant updated successfully"
}
```

---

## Error Response Format

All error responses follow this format:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message"
}
```

### Common Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 500 | Internal Server Error |

---

## Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| INVALID_JSON | 400 | Request body is not valid JSON |
| INVALID_REQUEST | 400 | Missing required parameters |
| EMAIL_EXISTS | 409 | Email already registered in tenant |
| USER_NOT_FOUND | 404 | User ID not found |
| TENANT_NOT_FOUND | 404 | Tenant ID not found |
| INVALID_CREDENTIALS | 401 | Email or password incorrect |
| USER_LOCKED | 403 | Account locked due to failed attempts |
| UNAUTHORIZED | 401 | Missing or invalid authentication |
| FORBIDDEN | 403 | Insufficient permissions (e.g., admin required) |
| INTERNAL_ERROR | 500 | Server error |

---

## Rate Limiting

Rate limiting is applied to prevent abuse:

- **Login endpoint**: Max 10 attempts per IP per minute
- **Register endpoint**: Max 5 registrations per IP per hour

Headers will include:
- `X-RateLimit-Limit`: Total allowed requests
- `X-RateLimit-Remaining`: Requests remaining
- `X-RateLimit-Reset`: Unix timestamp of reset time

---

## Pagination

List endpoints support pagination with:
- `page` - Page number (1-indexed)
- `per_page` - Items per page (default 20, max 100)

Response includes:
```json
{
  "data": [...],
  "page": 1,
  "per_page": 20,
  "total": 100,
  "total_pages": 5
}
```

---

## Examples

### Complete Login Flow
```bash
# 1. Register
curl -X POST http://localhost:8001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000"
  }'

# 2. Login
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'

# 3. Get current user
curl -X GET http://localhost:8001/api/v1/auth/me \
  -H "Authorization: Bearer {accessToken}"

# 4. Refresh token
curl -X POST http://localhost:8001/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "{refreshToken}"}'

# 5. Logout
curl -X POST http://localhost:8001/api/v1/auth/logout \
  -H "Authorization: Bearer {accessToken}"
```

### Admin Operations
```bash
# List all users
curl -X GET "http://localhost:8001/api/v1/users?page=1&per_page=20" \
  -H "Authorization: Bearer {adminToken}"

# Assign role to user
curl -X PUT http://localhost:8001/api/v1/users/{userId}/roles \
  -H "Authorization: Bearer {adminToken}" \
  -H "Content-Type: application/json" \
  -d '{"role": "manager"}'

# Delete user
curl -X DELETE http://localhost:8001/api/v1/users/{userId} \
  -H "Authorization: Bearer {adminToken}"
```
