# CrateDesk — Auth, Security & Compliance (Implementierungsreife Spezifikation)

**Stand:** 21. März 2026
**Version:** 1.0
**Status:** Implementierungsreif

---

## 1. Authentication System (JWT RS256)

### 1.1 JWT Flow — Vollständiger Sequence-Diagram

```
┌─────────────┐                    ┌────────────────┐                ┌─────────────┐
│   Client    │                    │  auth-service  │                │  Redis      │
└──────┬──────┘                    └────────┬───────┘                └──────┬──────┘
       │                                    │                               │
       │  1. POST /auth/login               │                               │
       │    {email, password}               │                               │
       ├───────────────────────────────────>│                               │
       │                                    │                               │
       │                  2. Validate credentials (bcrypt)                  │
       │                     Check rate limiting                            │
       │                                    │                               │
       │                  3. Generate JWT (RS256)                           │
       │                     Generate Refresh Token (secure random)         │
       │                                    │                               │
       │                                    │  4. HSET session:{token_jti}   │
       │                                    │      {user_id, tenant_id,     │
       │                                    │       ip, ua, created_at}      │
       │                                    ├──────────────────────────────>│
       │                                    │        (TTL: 7 days)          │
       │                                    │                               │
       │  5. HTTP 200 {access_token,        │                               │
       │    refresh_token, expires_in}      │                               │
       │<───────────────────────────────────┤                               │
       │                                    │                               │
       │  6. Client speichert access_token  │                               │
       │     (Memory oder secure Storage)   │                               │
       │     refresh_token (secure Cookie)  │                               │
       │                                    │                               │
       │  7. GET /api/equipment             │                               │
       │    Header: Authorization: Bearer {access_token}                   │
       ├───────────────────────────────────>│                               │
       │                                    │                               │
       │                  8. Validate JWT (RS256 signature)                 │
       │                     Extract claims (sub, tenant_id, permissions)   │
       │                     Check not blacklisted                          │
       │                                    │                               │
       │  9. HTTP 200 [equipment...]        │                               │
       │<───────────────────────────────────┤                               │
       │                                    │                               │
       │ (nach 15 Minuten: access_token abgelaufen)                        │
       │                                    │                               │
       │  10. POST /auth/refresh            │                               │
       │      {refresh_token}               │                               │
       ├───────────────────────────────────>│                               │
       │                                    │                               │
       │                  11. Validate refresh_token                        │
       │                      Check in Redis session                        │
       │                      Check not blacklisted                         │
       │                                    │                               │
       │                  12. Rotate: Issue NEW access + refresh token     │
       │                      Invalidate OLD refresh_token                  │
       │                                    │  13. DEL session:{old_jti}    │
       │                                    ├──────────────────────────────>│
       │                                    │                               │
       │                                    │  14. HSET session:{new_jti}   │
       │                                    ├──────────────────────────────>│
       │                                    │                               │
       │  15. HTTP 200 {new_access_token,   │                               │
       │        new_refresh_token}          │                               │
       │<───────────────────────────────────┤                               │
       │                                    │                               │
       │  16. POST /auth/logout             │                               │
       │      {refresh_token}               │                               │
       ├───────────────────────────────────>│                               │
       │                                    │                               │
       │                  17. Extract JTI from access_token                 │
       │                      Invalidate both tokens                        │
       │                                    │  18. DEL session:{jti}        │
       │                                    ├──────────────────────────────>│
       │                                    │                               │
       │                                    │  19. HSET blacklist:{jti}     │
       │                                    ├──────────────────────────────>│
       │                                    │      (TTL: 15 min)            │
       │                                    │                               │
       │  20. HTTP 204 No Content           │                               │
       │<───────────────────────────────────┤                               │
       │                                    │                               │
```

### 1.2 JWT Token-Struktur (Exakt)

#### Access Token (RS256)

```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "auth-service-key-v1"
}
```

**Payload:**

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "iss": "myrms-auth-service",
  "aud": ["myrms-api"],
  "exp": 1711100400,
  "iat": 1711096800,
  "nbf": 1711096800,
  "jti": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "tenant_id": "a7d5e9c2-1b34-4f8e-9c7d-3a5f8e1c2b9a",
  "email": "marco@berger-gruppe.de",
  "email_verified": true,
  "name": "Marco Berger",
  "roles": ["admin", "project_manager"],
  "permissions": [
    "equipment.read",
    "equipment.write",
    "equipment.delete",
    "project.read",
    "project.write",
    "invoice.read",
    "invoice.write",
    "users.read",
    "users.write",
    "roles.write"
  ],
  "mfa_verified": true,
  "mfa_method": "totp",
  "ip_address": "192.168.1.100",
  "user_agent_hash": "sha256:8f14e45f...",
  "session_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479"
}
```

**Go Struct:**

```go
type AccessTokenClaims struct {
    // Standard JWT Claims (RFC 7519)
    Subject   string   `json:"sub"`                // user_id
    Issuer    string   `json:"iss"`                // "myrms-auth-service"
    Audience  []string `json:"aud"`                // ["myrms-api"]
    ExpiresAt int64    `json:"exp"`                // unix timestamp
    IssuedAt  int64    `json:"iat"`                // unix timestamp
    NotBefore int64    `json:"nbf"`                // unix timestamp
    JWTID     string    `json:"jti"`                // unique token id

    // CrateDesk Custom Claims
    TenantID      uuid.UUID `json:"tenant_id"`
    Email         string    `json:"email"`
    EmailVerified bool      `json:"email_verified"`
    Name          string    `json:"name"`
    Roles         []string  `json:"roles"`
    Permissions   []string  `json:"permissions"`
    MFAVerified   bool      `json:"mfa_verified"`
    MFAMethod     string    `json:"mfa_method"` // "totp", "email", "sms"
    IPAddress     string    `json:"ip_address"` // Anti-token-stealing
    UserAgentHash string    `json:"user_agent_hash"` // SHA256(User-Agent)
    SessionID     string    `json:"session_id"` // Redis key
}

// Validators
func (c *AccessTokenClaims) Valid() error {
    now := time.Now().Unix()
    if c.ExpiresAt < now {
        return jwt.NewValidationError("token expired", jwt.ValidationErrorExpired)
    }
    if c.NotBefore > now {
        return jwt.NewValidationError("token not yet valid", jwt.ValidationErrorNotValidYet)
    }
    if c.Issuer != "myrms-auth-service" {
        return jwt.NewValidationError("invalid issuer", jwt.ValidationErrorClaimsInvalid)
    }
    return nil
}
```

#### Refresh Token (Opaque, in Redis)

```
Structure: Secure Random (256-bit), stored as Redis Hash

Redis Key: session:{jti}
Redis Value:
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": "a7d5e9c2-1b34-4f8e-9c7d-3a5f8e1c2b9a",
  "refresh_token_hash": "sha256:...",  // sha256(refresh_token_value)
  "created_at": 1711096800,
  "ip_address": "192.168.1.100",
  "user_agent_hash": "sha256:8f14e45f...",
  "used_at": 1711096900,
  "rotation_count": 1,
  "max_rotations": 5  // Anti-replay: max 5 refreshes pro Refresh Token
}

TTL: 7 days (604800 seconds)
```

**Go Struct:**

```go
type RefreshTokenSession struct {
    UserID           uuid.UUID `json:"user_id"`
    TenantID         uuid.UUID `json:"tenant_id"`
    RefreshTokenHash string    `json:"refresh_token_hash"`  // hex-encoded
    CreatedAt        int64     `json:"created_at"`
    IPAddress        string    `json:"ip_address"`
    UserAgentHash    string    `json:"user_agent_hash"`
    UsedAt           int64     `json:"used_at"`
    RotationCount    int       `json:"rotation_count"`
    MaxRotations     int       `json:"max_rotations"`
}

// Tokens sind opaque (keine JWT), 32-byte random
type RefreshToken []byte

func GenerateRefreshToken() RefreshToken {
    b := make([]byte, 32)
    rand.Read(b)  // crypto/rand
    return RefreshToken(b)
}

func (rt RefreshToken) String() string {
    return base64.URLEncoding.EncodeToString(rt)
}

func (rt RefreshToken) Hash() string {
    h := sha256.Sum256(rt)
    return hex.EncodeToString(h[:])
}
```

### 1.3 Lifetime & Rotation Strategy

| Token | Lifetime | Rotation | Revocation | Binding |
|-------|----------|----------|-----------|---------|
| Access Token | 15 Minuten | Nicht rotierbar | Nach Logout (Redis Blacklist) | IP + User-Agent (Hash) |
| Refresh Token | 7 Tage | Nach jedem Refresh (max 5x) | Nach Logout (Redis Del) | IP + User-Agent (Hash) |
| Session (Redis) | 7 Tage | Mit Refresh Token | Sofort bei Logout | Enthält alle Audit-Daten |

**Rotation Strategy (Rolling Refresh):**

```go
// Wenn Refresh Token verwendet wird:
1. Validiere alten Refresh Token (Hash, IP, UA, RotationCount < MaxRotations)
2. Generiere NEW Access Token + NEW Refresh Token
3. Speichere neue Session in Redis
4. Lösche ALTE Refresh Token aus Redis
5. Token mit altem JTI können nicht mehr verwendet werden (verhindert Replay)
```

### 1.4 Token Storage in Redis

**Session-Key Schema:**

```
session:{access_token_jti}  (TTL: 15 Minuten)
  → user_id
  → tenant_id
  → refresh_token_jti (zum schnellen Logout)
  → ip_address
  → user_agent_hash
  → created_at
  → last_activity
  → mfa_verified

session:refresh:{refresh_token_jti}  (TTL: 7 Tage)
  → user_id
  → tenant_id
  → refresh_token_hash
  → created_at
  → ip_address
  → user_agent_hash
  → rotation_count
  → used_at

blacklist:{jti}  (TTL: 15 Minuten nach Logout)
  → blacklisted_at
  → reason ("logout", "password_reset", "suspicious_activity")
```

**Go Code (Session Persistence):**

```go
func (r *RedisSessionStore) SaveSession(ctx context.Context, claims *AccessTokenClaims, refreshTokenJTI string) error {
    sessionData := map[string]interface{}{
        "user_id":           claims.Subject,
        "tenant_id":         claims.TenantID.String(),
        "refresh_token_jti": refreshTokenJTI,
        "ip_address":        claims.IPAddress,
        "user_agent_hash":   claims.UserAgentHash,
        "created_at":        claims.IssuedAt,
        "last_activity":     time.Now().Unix(),
        "mfa_verified":      claims.MFAVerified,
    }

    key := fmt.Sprintf("session:%s", claims.JWTID)
    pipe := r.client.Pipeline()
    pipe.HSet(ctx, key, sessionData)
    pipe.Expire(ctx, key, 15*time.Minute)
    _, err := pipe.Exec(ctx)
    return err
}

func (r *RedisSessionStore) SaveRefreshSession(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, refreshToken string, claims *AccessTokenClaims) error {
    refreshTokenHash := hashToken(refreshToken)
    sessionData := map[string]interface{}{
        "user_id":           userID.String(),
        "tenant_id":         tenantID.String(),
        "refresh_token_hash": refreshTokenHash,
        "created_at":        time.Now().Unix(),
        "ip_address":        claims.IPAddress,
        "user_agent_hash":   claims.UserAgentHash,
        "rotation_count":    0,
        "used_at":           time.Now().Unix(),
    }

    key := fmt.Sprintf("session:refresh:%s", claims.JWTID)
    pipe := r.client.Pipeline()
    pipe.HSet(ctx, key, sessionData)
    pipe.Expire(ctx, key, 7*24*time.Hour)
    _, err := pipe.Exec(ctx)
    return err
}

func (r *RedisSessionStore) Blacklist(ctx context.Context, jti string, reason string) error {
    blacklistData := map[string]interface{}{
        "blacklisted_at": time.Now().Unix(),
        "reason":         reason,
    }

    key := fmt.Sprintf("blacklist:%s", jti)
    pipe := r.client.Pipeline()
    pipe.HSet(ctx, key, blacklistData)
    pipe.Expire(ctx, key, 15*time.Minute)
    _, err := pipe.Exec(ctx)
    return err
}
```

### 1.5 Password Hashing (Argon2id)

**Parameter (OWASP-Empfohlen für 2026):**

```go
type Argon2Params struct {
    Memory      uint32 // 64 MB
    Iterations  uint32 // 2
    Parallelism uint8  // 2
    SaltLength  uint32 // 16 bytes
    KeyLength   uint32 // 32 bytes
}

var DefaultArgon2Params = Argon2Params{
    Memory:      65536,  // 64 MB (für Schnelligkeit, aber Brute-Force-resistent)
    Iterations: 2,       // 2 Iterationen (genug für Login-Geschwindigkeit)
    Parallelism: 2,      // 2 Threads
    SaltLength: 16,      // 16-byte Random Salt
    KeyLength:  32,      // 32-byte Key
}
```

**Hash-Format (PHC String):**

```
$argon2id$v=19$m=65536,t=2,p=2$<salt>$<hash>

Beispiel:
$argon2id$v=19$m=65536,t=2,p=2$J7xHKj8L9mPq2vZe5nAb$x7F9dK2pL1mN4qR7sT0uV3wX6yZ8aB5cD9eFgHjKlM
```

**Go Implementation:**

```go
import "golang.org/x/crypto/argon2"

func HashPassword(password string) (string, error) {
    salt := make([]byte, DefaultArgon2Params.SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }

    hash := argon2.IDKey(
        []byte(password),
        salt,
        DefaultArgon2Params.Iterations,
        DefaultArgon2Params.Memory,
        DefaultArgon2Params.Parallelism,
        DefaultArgon2Params.KeyLength,
    )

    // PHC String-Format
    phcString := fmt.Sprintf(
        "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        DefaultArgon2Params.Memory,
        DefaultArgon2Params.Iterations,
        DefaultArgon2Params.Parallelism,
        base64.RawStdEncoding.EncodeToString(salt),
        base64.RawStdEncoding.EncodeToString(hash),
    )

    return phcString, nil
}

func VerifyPassword(hash, password string) (bool, error) {
    // Parse PHC String
    parts := strings.Split(hash, "$")
    if len(parts) != 6 || parts[1] != "argon2id" {
        return false, fmt.Errorf("invalid hash format")
    }

    salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
    storedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])

    computedHash := argon2.IDKey(
        []byte(password),
        salt,
        DefaultArgon2Params.Iterations,
        DefaultArgon2Params.Memory,
        DefaultArgon2Params.Parallelism,
        DefaultArgon2Params.KeyLength,
    )

    return subtle.ConstantTimeCompare(storedHash, computedHash) == 1, nil
}
```

### 1.6 Brute-Force Protection

**Progressive Delay Strategy:**

```go
type BruteForceProtection struct {
    redis *redis.Client
}

// Attempt-Tracking
type LoginAttempt struct {
    Email      string
    IPAddress  string
    Timestamp  int64
    Success    bool
    FailCount  int  // Pro Email + IP
}

func (b *BruteForceProtection) CheckRateLimit(ctx context.Context, email, ipAddress string) error {
    key := fmt.Sprintf("login:attempts:%s:%s", email, ipAddress)
    failCount, _ := b.redis.Incr(ctx, key).Val(), nil

    // Exponential backoff: 2^failCount Sekunden
    switch failCount {
    case 1, 2, 3:
        // Kein Delay
    case 4:
        time.Sleep(2 * time.Second)  // 2^1
    case 5:
        time.Sleep(4 * time.Second)  // 2^2
    case 6:
        time.Sleep(8 * time.Second)  // 2^3
    case 7:
        time.Sleep(16 * time.Second) // 2^4
    default:
        return fmt.Errorf("too many login attempts, try again in 1 hour")
    }

    // TTL: 1 Stunde (count wird nach 1h zurückgesetzt)
    b.redis.Expire(ctx, key, 1*time.Hour)

    return nil
}

// Account Lockout (nach 10 Fehlversuchen)
func (b *BruteForceProtection) IsAccountLocked(ctx context.Context, email string) (bool, error) {
    key := fmt.Sprintf("login:locked:%s", email)
    locked, err := b.redis.Exists(ctx, key).Result()
    return locked > 0, err
}

func (b *BruteForceProtection) LockAccount(ctx context.Context, email string, durationMinutes int) error {
    key := fmt.Sprintf("login:locked:%s", email)
    return b.redis.SetEx(ctx, key, "true", time.Duration(durationMinutes)*time.Minute).Err()
}

func (b *BruteForceProtection) RecordFailure(ctx context.Context, email, ipAddress string) {
    key := fmt.Sprintf("login:attempts:%s:%s", email, ipAddress)
    failCount, _ := b.redis.Incr(ctx, key).Val(), nil
    b.redis.Expire(ctx, key, 1*time.Hour)

    // Lock nach 10 Versuchen (30 Minuten)
    if failCount >= 10 {
        b.LockAccount(ctx, email, 30)
    }
}

func (b *BruteForceProtection) RecordSuccess(ctx context.Context, email, ipAddress string) {
    key := fmt.Sprintf("login:attempts:%s:%s", email, ipAddress)
    b.redis.Del(ctx, key)
}
```

**Account Lockout Policy:**

```
10 Fehlversuche (pro E-Mail + IP) → Account gesperrt für 30 Minuten
Nach Sperrung: Admin kann manuell freischalten oder automatisch nach 30min
```

### 1.7 Login Handler (Go Code)

```go
// POST /api/auth/login
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"` // 900 (15 Minuten)
    TokenType    string `json:"token_type"` // "Bearer"
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // Validierung
    if err := h.validator.StructCtx(ctx, req); err != nil {
        http.Error(w, "validation failed", http.StatusBadRequest)
        return
    }

    // Rate Limiting
    clientIP := getRealIP(r)
    if err := h.bruteForce.CheckRateLimit(ctx, req.Email, clientIP); err != nil {
        http.Error(w, "too many attempts", http.StatusTooManyRequests)
        return
    }

    // Account Lockout Check
    locked, _ := h.bruteForce.IsAccountLocked(ctx, req.Email)
    if locked {
        http.Error(w, "account locked, try again later", http.StatusForbidden)
        return
    }

    // User Lookup
    user, err := h.userRepo.FindByEmail(ctx, req.Email)
    if err != nil || user == nil {
        h.bruteForce.RecordFailure(ctx, req.Email, clientIP)
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }

    // Tenant Validation
    if user.TenantID == uuid.Nil || user.DeletedAt != nil {
        h.bruteForce.RecordFailure(ctx, req.Email, clientIP)
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }

    // Password Verification
    valid, err := h.passwordHasher.VerifyPassword(user.PasswordHash, req.Password)
    if !valid || err != nil {
        h.bruteForce.RecordFailure(ctx, req.Email, clientIP)
        h.logger.Warn("login failed", "email", req.Email, "ip", clientIP)
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }

    // MFA Check (wenn aktiviert)
    if user.MFAEnabled {
        // Temporären MFA-Challenge-Token erstellen
        mfaToken := h.mfaService.CreateChallenge(ctx, user.ID, user.TenantID)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusAccepted)  // 202 statt 200
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status":       "mfa_required",
            "mfa_token":    mfaToken,
            "mfa_methods":  user.MFAMethods,
            "expires_in":   300, // 5 Minuten
        })
        return
    }

    // Event: UserLoggedIn
    h.eventStore.Append(ctx, "user-"+user.ID.String(), events.UserLoggedIn{
        UserID:       user.ID,
        TenantID:     user.TenantID,
        Email:        user.Email,
        IPAddress:    clientIP,
        UserAgent:    r.Header.Get("User-Agent"),
        Timestamp:    time.Now(),
    })

    // JWT & Refresh Token generieren
    claims := &AccessTokenClaims{
        Subject:       user.ID.String(),
        Issuer:        "myrms-auth-service",
        Audience:      []string{"myrms-api"},
        ExpiresAt:     time.Now().Add(15 * time.Minute).Unix(),
        IssuedAt:      time.Now().Unix(),
        NotBefore:     time.Now().Unix(),
        JWTID:         uuid.New().String(),
        TenantID:      user.TenantID,
        Email:         user.Email,
        EmailVerified: user.EmailVerified,
        Name:          user.FullName,
        Roles:         user.Roles,
        Permissions:   h.permissionService.GetPermissionsForUser(ctx, user.ID),
        MFAVerified:   false,
        MFAMethod:     "",
        IPAddress:     clientIP,
        UserAgentHash: hashUserAgent(r.Header.Get("User-Agent")),
    }

    accessToken, err := h.jwtService.SignToken(claims)
    if err != nil {
        http.Error(w, "token generation failed", http.StatusInternalServerError)
        return
    }

    // Refresh Token generieren
    refreshToken := GenerateRefreshToken()

    // Sessions speichern
    h.sessionStore.SaveSession(ctx, claims, claims.JWTID)
    h.sessionStore.SaveRefreshSession(ctx, user.ID, user.TenantID, refreshToken.String(), claims)

    // Erfolgreicher Login
    h.bruteForce.RecordSuccess(ctx, user.Email, clientIP)
    h.logger.Info("login successful", "user_id", user.ID, "email", user.Email, "ip", clientIP)

    // Cookies setzen (secure, httpOnly, sameSite)
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken.String(),
        Path:     "/",
        HttpOnly: true,
        Secure:   true,  // HTTPS only
        SameSite: http.SameSiteLax,
        MaxAge:   7 * 24 * 3600, // 7 Tage
    })

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken.String(),
        ExpiresIn:    900,
        TokenType:    "Bearer",
    })
}
```

### 1.8 JWT Middleware

```go
// Middleware für alle geschützten Routen
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Token aus Authorization Header extrahieren
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "missing authorization header", http.StatusUnauthorized)
            return
        }

        parts := strings.Fields(authHeader)
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "invalid authorization header", http.StatusUnauthorized)
            return
        }

        tokenString := parts[1]

        // Token validieren
        claims := &AccessTokenClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            // Signature mit Public Key verifizieren
            if token.Method != jwt.SigningMethodRS256 {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return m.publicKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        // Token in Blacklist?
        blacklisted, _ := m.sessionStore.IsBlacklisted(ctx, claims.JWTID)
        if blacklisted {
            http.Error(w, "token revoked", http.StatusUnauthorized)
            return
        }

        // Session in Redis validieren
        session, err := m.sessionStore.GetSession(ctx, claims.JWTID)
        if err != nil || session == nil {
            http.Error(w, "invalid session", http.StatusUnauthorized)
            return
        }

        // IP + User-Agent Binding prüfen (Optional: Strict)
        clientIP := getRealIP(r)
        clientUA := r.Header.Get("User-Agent")
        if m.strictBinding && (claims.IPAddress != clientIP || claims.UserAgentHash != hashUserAgent(clientUA)) {
            m.logger.Warn("suspicious request", "user_id", claims.Subject, "original_ip", claims.IPAddress, "current_ip", clientIP)
            http.Error(w, "token binding validation failed", http.StatusUnauthorized)
            return
        }

        // Claims in Context speichern
        ctx = context.WithValue(ctx, "claims", claims)
        ctx = context.WithValue(ctx, "user_id", uuid.MustParse(claims.Subject))
        ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
        ctx = context.WithValue(ctx, "permissions", claims.Permissions)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Middleware für bestimmte Permissions
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("claims").(*AccessTokenClaims)

            if !contains(claims.Permissions, permission) {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 1.9 Refresh Token Handler

```go
// POST /api/auth/refresh
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    var req RefreshRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // Refresh Token validieren
    refreshSession, err := h.sessionStore.GetRefreshSession(ctx, req.RefreshToken)
    if err != nil || refreshSession == nil {
        http.Error(w, "invalid refresh token", http.StatusUnauthorized)
        return
    }

    // Rotation Count prüfen
    if refreshSession.RotationCount >= refreshSession.MaxRotations {
        // Verdacht auf Token-Reuse
        h.logger.Error("refresh token max rotations exceeded", "user_id", refreshSession.UserID)
        h.sessionStore.Blacklist(ctx, "", "max_rotations_exceeded")
        http.Error(w, "refresh token limit exceeded", http.StatusUnauthorized)
        return
    }

    // User & Permissions laden
    user, _ := h.userRepo.FindByID(ctx, refreshSession.UserID)
    permissions := h.permissionService.GetPermissionsForUser(ctx, user.ID)

    // Neue Claims generieren
    newClaims := &AccessTokenClaims{
        Subject:       user.ID.String(),
        Issuer:        "myrms-auth-service",
        Audience:      []string{"myrms-api"},
        ExpiresAt:     time.Now().Add(15 * time.Minute).Unix(),
        IssuedAt:      time.Now().Unix(),
        NotBefore:     time.Now().Unix(),
        JWTID:         uuid.New().String(),
        TenantID:      user.TenantID,
        Email:         user.Email,
        Roles:         user.Roles,
        Permissions:   permissions,
        MFAVerified:   refreshSession.MFAVerified,
        IPAddress:     getRealIP(r),
        UserAgentHash: hashUserAgent(r.Header.Get("User-Agent")),
    }

    // Neue Tokens generieren
    newAccessToken, _ := h.jwtService.SignToken(newClaims)
    newRefreshToken := GenerateRefreshToken()

    // Session-Update: alte Token invalidieren, neue speichern
    h.sessionStore.SaveSession(ctx, newClaims, newClaims.JWTID)
    h.sessionStore.SaveRefreshSession(ctx, user.ID, user.TenantID, newRefreshToken.String(), newClaims)
    h.sessionStore.Blacklist(ctx, refreshSession.JWTID, "rotated")

    // Event
    h.eventStore.Append(ctx, "user-"+user.ID.String(), events.TokenRefreshed{
        UserID:    user.ID,
        TenantID:  user.TenantID,
        OldJTI:    refreshSession.JWTID,
        NewJTI:    newClaims.JWTID,
        Timestamp: time.Now(),
    })

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(LoginResponse{
        AccessToken:  newAccessToken,
        RefreshToken: newRefreshToken.String(),
        ExpiresIn:    900,
        TokenType:    "Bearer",
    })
}
```

### 1.10 Logout Handler

```go
// POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    claims := r.Context().Value("claims").(*AccessTokenClaims)

    // Alle Tokens blacklisten
    h.sessionStore.Blacklist(ctx, claims.JWTID, "logout")

    // Session löschen
    h.sessionStore.DeleteSession(ctx, claims.JWTID)

    // Event
    h.eventStore.Append(ctx, "user-"+claims.Subject, events.UserLoggedOut{
        UserID:    uuid.MustParse(claims.Subject),
        TenantID:  claims.TenantID,
        Timestamp: time.Now(),
    })

    // Refresh Token Cookie löschen
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLax,
        MaxAge:   -1,
    })

    w.WriteHeader(http.StatusNoContent)
}
```

---

## 2. RBAC (Role-Based Access Control)

### 2.1 Rollen-Hierarchie

```
┌─────────────────────────────────────────────────────────┐
│                        ADMIN                             │
│  Vollzugriff auf alle Funktionen + System-Verwaltung     │
│  Permissions: *.read, *.write, *.delete, users.write     │
│  Kann andere Rollen zuweisen                             │
└─────────────────────────────────────────────────────────┘
           ↑                 ↑                    ↑
      ┌────┴────┐      ┌────┴────┐      ┌────────┴─────┐
      │          │      │         │      │              │
   ┌──▼──┐  ┌──▼──┐  ┌──▼──┐  ┌──▼──┐ ┌──▼──┐      ┌──▼──┐
   │PL   │  │WARE │  │BUCH │  │CREW │ │FREE │      │PART │
   │MGMT │  │MGMT │  │HALT │  │MGMT │ │LANCER      │NER  │
   └─────┘  └─────┘  └─────┘  └─────┘ └─────┘      └─────┘
```

**Rollen-Definitionen (System-Rollen):**

| Rolle | Code | Beschreibung | Berechtigungen |
|-------|------|-------------|----------------|
| **Admin** | `admin` | Vollzugriff, Nutzer-Verwaltung, System-Config | Alle |
| **Projektleiter** | `project_manager` | Projekte, Packlisten, Equipment-Reservierung | project.*, equipment.read, notification.* |
| **Lagermeister** | `warehouse_manager` | Warenbewegung, Lagerplätze, Inventur, Check-In/Out | warehouse.*, equipment.*, inventory.read, project.read |
| **Buchhalter** | `accountant` | Rechnungen, Mahnungen, DATEV-Export | invoice.*, equipment.read, project.read, crew.read (für Zeiterfassung) |
| **Crew-Leiter** | `crew_manager` | Personalplanung, Zeiterfassung, CalDAV | crew.*, project.read |
| **Freelancer** | `freelancer` | Nur eigene Zeiterfassung + verfügbare Projekte | crew.my_profile, crew.my_timecard (write), project.read |
| **Partner** | `partner` | Nur Federation: Equipment-Sharing-Anfragen | federation.read, federation.requests.write |
| **Custom** | `custom_*` | Admin-definierte Rollen | Kombinierbar |

### 2.2 Permissions-Matrix (Alle Routes × Rollen)

```markdown
┌─────────────────────────────────────────────────────────────────────────────┐
│  Route / Ressource       │ Admin │ PL │ WM │ BH │ CM │ FREE │ PART │ Custom│
├─────────────────────────────────────────────────────────────────────────────┤
│ /api/equipment           │       │    │    │    │    │      │      │       │
│   GET (list)             │  ✓   │ ✓  │ ✓  │ ✓  │    │      │      │ cfg  │
│   GET (single)           │  ✓   │ ✓  │ ✓  │ ✓  │    │      │      │ cfg  │
│   POST (create)          │  ✓   │    │ ✓  │    │    │      │      │ cfg  │
│   PUT (update)           │  ✓   │    │ ✓  │    │    │      │      │ cfg  │
│   DELETE                 │  ✓   │    │    │    │    │      │      │ cfg  │
│   POST check-out         │  ✓   │ ✓  │ ✓  │    │    │      │      │ cfg  │
│   POST check-in          │  ✓   │ ✓  │ ✓  │    │    │      │      │ cfg  │
├─────────────────────────────────────────────────────────────────────────────┤
│ /api/projects            │       │    │    │    │    │      │      │       │
│   GET (list)             │  ✓   │ ✓  │    │ ✓  │    │      │      │ cfg  │
│   POST (create)          │  ✓   │ ✓  │    │    │    │      │      │ cfg  │
│   PUT (update)           │  ✓   │ ✓  │    │    │    │      │      │ cfg  │
│   DELETE                 │  ✓   │    │    │    │    │      │      │ cfg  │
├─────────────────────────────────────────────────────────────────────────────┤
│ /api/invoices            │       │    │    │    │    │      │      │       │
│   GET                    │  ✓   │    │    │ ✓  │    │      │      │ cfg  │
│   POST (create)          │  ✓   │    │    │ ✓  │    │      │      │ cfg  │
│   PUT                    │  ✓   │    │    │ ✓  │    │      │      │ cfg  │
│   DELETE                 │  ✓   │    │    │ ✓  │    │      │      │ cfg  │
│   POST /send             │  ✓   │    │    │ ✓  │    │      │      │ cfg  │
├─────────────────────────────────────────────────────────────────────────────┤
│ /api/users               │       │    │    │    │    │      │      │       │
│   GET                    │  ✓   │    │    │    │    │      │      │ no   │
│   POST                   │  ✓   │    │    │    │    │      │      │ no   │
│   PUT                    │  ✓   │    │    │    │    │      │      │ no   │
│   DELETE                 │  ✓   │    │    │    │    │      │      │ no   │
├─────────────────────────────────────────────────────────────────────────────┤
│ /api/roles               │       │    │    │    │    │      │      │       │
│   GET                    │  ✓   │    │    │    │    │      │      │ cfg  │
│   POST                   │  ✓   │    │    │    │    │      │      │ no   │
│   PUT                    │  ✓   │    │    │    │    │      │      │ no   │
│   DELETE                 │  ✓   │    │    │    │    │      │      │ no   │
├─────────────────────────────────────────────────────────────────────────────┤
│ /api/crew                │       │    │    │    │    │      │      │       │
│   GET (all)              │  ✓   │ ✓  │    │    │ ✓  │      │      │ cfg  │
│   GET /me                │  ✓   │ ✓  │    │    │ ✓  │ ✓    │      │ auto │
│   POST timecard          │  ✓   │    │    │    │ ✓  │ ✓*   │      │ cfg  │
│   PUT timecard           │  ✓   │    │    │    │ ✓  │ ✓*   │      │ cfg  │
│     (* nur eigene)       │      │    │    │    │    │      │      │      │
├─────────────────────────────────────────────────────────────────────────────┤
│ /api/federation          │       │    │    │    │    │      │      │       │
│   GET                    │  ✓   │    │    │    │    │      │ ✓    │ cfg  │
│   POST request           │  ✓   │    │    │    │    │      │ ✓    │ cfg  │
│   PUT response           │  ✓   │    │    │    │    │      │ ✓    │ cfg  │

Legend: ✓ = Zugriff gewährt, cfg = konfigurierbar (Custom Roles), no = niemals
```

### 2.3 Resource-Level Permissions

**Freelancer sieht nur eigene Daten:**

```go
// GET /api/crew/{id}
func (h *CrewHandler) GetCrew(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    claims := ctx.Value("claims").(*AccessTokenClaims)
    crewID := r.PathValue("id")

    // Ressourcen-Level RBAC
    if claims.Subject != crewID && !contains(claims.Roles, "admin") && !contains(claims.Roles, "crew_manager") {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    crew, _ := h.crewRepo.FindByID(ctx, crewID)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(crew)
}
```

**Project Manager sieht nur Projekte seines Tenants:**

```go
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    claims := ctx.Value("claims").(*AccessTokenClaims)

    // Tenant-Isolation wird durch Claims erzwungen
    projects, _ := h.projectRepo.ListByTenant(ctx, claims.TenantID)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(projects)
}
```

### 2.4 RBAC Middleware (Go)

```go
type RBACMiddleware struct {
    permissionService *PermissionService
    logger            *zerolog.Logger
}

// RequireRole: Mindestens eine der angegebenen Rollen erforderlich
func (m *RBACMiddleware) RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("claims").(*AccessTokenClaims)

            hasRole := false
            for _, allowed := range allowedRoles {
                if contains(claims.Roles, allowed) {
                    hasRole = true
                    break
                }
            }

            if !hasRole {
                m.logger.Warn("rbac violation", "user_id", claims.Subject, "allowed", allowedRoles, "actual", claims.Roles)
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// RequirePermission: Exakte Permission erforderlich (z.B. "equipment.write")
func (m *RBACMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("claims").(*AccessTokenClaims)

            if !contains(claims.Permissions, permission) {
                m.logger.Warn("permission denied", "user_id", claims.Subject, "permission", permission)
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// ResourceOwner: Nur Eigentümer oder Admin kann Ressource ändern
func (m *RBACMiddleware) ResourceOwner(ownerIDExtractor func(*http.Request) uuid.UUID) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("claims").(*AccessTokenClaims)
            resourceOwnerID := ownerIDExtractor(r)
            userID := uuid.MustParse(claims.Subject)

            if userID != resourceOwnerID && !contains(claims.Roles, "admin") {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 2.5 Custom Roles (Admin-definiert)

**Rollen-Modell in PostgreSQL:**

```sql
CREATE TABLE auth_schema.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,  -- z.B. "custom_supervisor"
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,  -- FALSE für Custom Roles
    created_at TIMESTAMP DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, code)
);

CREATE TABLE auth_schema.role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,  -- z.B. "equipment.write"
    granted_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(role_id, permission)
);
```

**Go Code für Custom Role Erstellung:**

```go
type CreateCustomRoleRequest struct {
    Name        string   `json:"name" validate:"required,max=255"`
    Description string   `json:"description" validate:"max=1000"`
    Permissions []string `json:"permissions" validate:"required,min=1"`
}

func (h *RoleHandler) CreateCustomRole(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    claims := ctx.Value("claims").(*AccessTokenClaims)

    var req CreateCustomRoleRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Nur Admin darf Roles erstellen
    if !contains(claims.Roles, "admin") {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // Code auto-generieren
    code := "custom_" + strings.ToLower(strings.ReplaceAll(req.Name, " ", "_"))

    // Permissions validieren (nur erlaubte Permissions)
    allowedPermissions := h.permissionService.GetAllAvailablePermissions(ctx)
    for _, p := range req.Permissions {
        if !contains(allowedPermissions, p) {
            http.Error(w, fmt.Sprintf("invalid permission: %s", p), http.StatusBadRequest)
            return
        }
    }

    // Role in DB erstellen
    role := &Role{
        ID:          uuid.New(),
        TenantID:    claims.TenantID,
        Name:        req.Name,
        Code:        code,
        Description: req.Description,
        IsSystem:    false,
        CreatedBy:   uuid.MustParse(claims.Subject),
        CreatedAt:   time.Now(),
    }

    h.roleRepo.Create(ctx, role)

    // Permissions eintragen
    for _, p := range req.Permissions {
        h.roleRepo.AddPermission(ctx, role.ID, p)
    }

    // Event
    h.eventStore.Append(ctx, "role-"+role.ID.String(), events.CustomRoleCreated{
        RoleID:      role.ID,
        TenantID:    role.TenantID,
        Name:        role.Name,
        Code:        code,
        Permissions: req.Permissions,
        CreatedBy:   uuid.MustParse(claims.Subject),
        Timestamp:   time.Now(),
    })

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(role)
}
```

### 2.6 Permission Inheritance

System-Rollen vererben ihre Permissions automatisch:

```
admin
  └─ Alle Permissions (wildcard: *)

project_manager
  ├─ project.read
  ├─ project.write
  ├─ project.delete
  ├─ equipment.read
  └─ notification.*

warehouse_manager
  ├─ warehouse.read
  ├─ warehouse.write
  ├─ warehouse.delete
  ├─ equipment.read
  ├─ equipment.write
  ├─ inventory.read
  └─ project.read (nur für Verfügbarkeits-Check)

Custom Roles: Explizit konfiguriert, keine Vererbung
```

---

## 3. Multi-Tenancy

### 3.1 Tenant-Isolation-Strategie: Row-Level Security (RLS)

**Architektur:** Alle Daten in einer PostgreSQL-Instanz, RLS erzwingt Isolation per Tenant:

```sql
-- Beispiel: equipment Tabelle
CREATE TABLE inventory_schema.equipment (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255),
    quantity_total INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (tenant_id) REFERENCES public.tenants(id)
);

-- RLS aktivieren
ALTER TABLE inventory_schema.equipment ENABLE ROW LEVEL SECURITY;

-- Policy: Jeder User sieht nur Equipment seines Tenants
CREATE POLICY equipment_tenant_isolation ON inventory_schema.equipment
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

CREATE POLICY equipment_tenant_modify ON inventory_schema.equipment
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

**Jede Query setzt `app.current_tenant_id`:**

```go
func (r *EquipmentRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*Equipment, error) {
    // Set PostgreSQL GUC (Good User Configuration) variable
    rows, err := r.db.QueryContext(
        ctx,
        fmt.Sprintf(
            "SELECT * FROM inventory_schema.equipment WHERE tenant_id = $1",
        ),
        tenantID,
    )
    // Oder mit pgx Connection Pool:
    // rows, err := r.db.Query(ctx, "SET app.current_tenant_id TO $1; SELECT ...", tenantID)
    return scanEquipment(rows)
}
```

### 3.2 Tenant-Context-Propagation

**Tenant-ID Extraktion aus JWT:**

```go
type TenantContextMiddleware struct {}

func (m *TenantContextMiddleware) ExtractTenant(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        claims := ctx.Value("claims").(*AccessTokenClaims)

        // Tenant-ID aus JWT in Context speichern
        ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)

        // Auch in Request-Header für Service-zu-Service Calls
        r.Header.Set("X-Tenant-ID", claims.TenantID.String())

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Service-zu-Service Calls (gRPC) mit Tenant-Propagation:**

```go
// inventory-service ruft project-service auf
func (s *AvailabilityService) CheckAvailability(ctx context.Context, equipmentID uuid.UUID, ...) error {
    tenantID := ctx.Value("tenant_id").(uuid.UUID)

    // gRPC Call mit Tenant im Metadata
    md := metadata.Pairs("tenant-id", tenantID.String())
    grpcCtx := metadata.NewOutgoingContext(ctx, md)

    resp, err := s.projectClient.GetReservedQuantity(grpcCtx, &pb.GetReservedQuantityRequest{
        EquipmentID: equipmentID.String(),
    })
    return err
}

// project-service extractiert Tenant aus gRPC Metadata
func (s *ProjectService) GetReservedQuantity(ctx context.Context, req *pb.GetReservedQuantityRequest) (*pb.QuantityResponse, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    tenantIDStr := md.Get("tenant-id")[0]
    tenantID := uuid.MustParse(tenantIDStr)

    // Jetzt alle DB-Queries mit diesem tenantID
    reservations, _ := s.reservationRepo.ListByEquipment(ctx, tenantID, uuid.MustParse(req.EquipmentID))
    return &pb.QuantityResponse{Total: int32(len(reservations))}, nil
}
```

### 3.3 Tenant-Provisioning (Onboarding)

**Ablauf für neue Firma:**

```
1. Super-Admin registriert neue Firma
   → POST /api/admin/tenants
   → Erstelle Tenant-Record in public.tenants
   → Erstelle Tenant-spezifische PostgreSQL Schemas

2. Erste Benutzer (Geschäftsführer)
   → Email-Link mit Einladungstoken
   → Benutzer setzt Passwort
   → Erhält "admin" Rolle für seinen Tenant

3. Setup-Wizard (optional)
   → Kategorien anlegen
   → Equipment-Template importieren
   → Teams konfigurieren
```

**Go Code für Tenant Creation:**

```go
type CreateTenantRequest struct {
    CompanyName   string `json:"company_name" validate:"required"`
    Email         string `json:"email" validate:"required,email"`
    Country       string `json:"country" validate:"required,len=2"` // z.B. "DE"
    Industry      string `json:"industry"` // z.B. "event_tech"
}

func (h *AdminHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var req CreateTenantRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Tenant anlegen
    tenant := &Tenant{
        ID:           uuid.New(),
        CompanyName:  req.CompanyName,
        Country:      req.Country,
        Industry:     req.Industry,
        Status:       "active",
        CreatedAt:    time.Now(),
    }

    h.tenantRepo.Create(ctx, tenant)

    // PostgreSQL Schemas für Tenant erstellen
    h.db.ExecContext(ctx, fmt.Sprintf(`
        CREATE SCHEMA %s;
        CREATE SCHEMA %s;
        ...
    `, fmt.Sprintf("tenant_%s_inventory", tenant.ID), ...))

    // Erster User mit Admin-Rolle
    firstAdmin := &User{
        ID:       uuid.New(),
        TenantID: tenant.ID,
        Email:    req.Email,
        // Password wird nach Email-Bestätigung gesetzt
    }
    h.userRepo.Create(ctx, firstAdmin)
    h.roleAssignmentRepo.Assign(ctx, firstAdmin.ID, "admin")

    // Einladungs-Token
    token := generateInvitationToken()
    h.invitationRepo.Create(ctx, &Invitation{
        ID:        uuid.New(),
        UserID:    firstAdmin.ID,
        TenantID:  tenant.ID,
        Token:     token,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
    })

    // Event
    h.eventStore.Append(ctx, "tenant-"+tenant.ID.String(), events.TenantCreated{
        TenantID:    tenant.ID,
        CompanyName: req.CompanyName,
        CreatedAt:   time.Now(),
    })

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "tenant_id":        tenant.ID,
        "invitation_url":   fmt.Sprintf("https://myrms.local/join?token=%s", token),
        "expires_in_hours": 168,
    })
}
```

### 3.4 Daten-Isolation Garantien

**Was passiert bei fehlerhafter Tenant-ID?**

```go
// Szenario: Bug in Code setzt falsche Tenant-ID
func buggyHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    wrongTenantID := uuid.New()  // OOPS, sollte aus JWT sein!
    ctx = context.WithValue(ctx, "tenant_id", wrongTenantID)

    equipment, _ := h.equipmentRepo.List(ctx, wrongTenantID)
    // PostgreSQL RLS verhindert Zugriff:
    // ERROR: SELECT permission denied on equipment (RLS policy violiert)
    return
}
```

**Weitere Schutzmaßnahmen:**

```sql
-- Constraint: Equipment.tenant_id kann nicht NULL sein
ALTER TABLE inventory_schema.equipment
    ADD CONSTRAINT equipment_tenant_id_not_null
    CHECK (tenant_id IS NOT NULL);

-- Audit: Jede Cross-Tenant-Query wird geloggt
CREATE TRIGGER audit_tenant_violation
BEFORE SELECT ON inventory_schema.equipment
FOR EACH ROW
WHEN (NEW.tenant_id != current_setting('app.current_tenant_id')::uuid)
EXECUTE FUNCTION log_security_incident();

-- Foreign Keys: Unmöglich andere Tenants zu referenzieren
ALTER TABLE project_schema.projects
    ADD CONSTRAINT project_tenant_matches_equipment
    FOREIGN KEY (tenant_id, equipment_id)
    REFERENCES inventory_schema.equipment(tenant_id, id);
```

---

## 4. OWASP Top 10 Abdeckung (2025)

### 4.1 A01:2021 – Broken Access Control

**Angriffsvektor in CrateDesk:** Admin erstellt URL `/api/users/admin-user-id`, andere User versuchen zu ändern

**Schutzmaßnahme:**

```go
// RequireAdmin Middleware
func (m *RBACMiddleware) RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims := r.Context().Value("claims").(*AccessTokenClaims)
        if !contains(claims.Roles, "admin") {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// PUT /api/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("claims").(*AccessTokenClaims)
    targetUserID := uuid.MustParse(r.PathValue("id"))
    userID := uuid.MustParse(claims.Subject)

    // Nur Admin oder der User selbst darf sein Profil ändern
    if !contains(claims.Roles, "admin") && userID != targetUserID {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // Wenn nicht Admin: darf man Rolle nicht ändern
    var req UpdateUserRequest
    json.NewDecoder(r.Body).Decode(&req)
    if !contains(claims.Roles, "admin") && req.Roles != nil {
        http.Error(w, "cannot change roles", http.StatusForbidden)
        return
    }
    // ...
}
```

**Header zur Sicherheit setzen:**

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

**Test-Szenario:**

```bash
# 1. Normaler User loggt sich ein
curl -X POST https://myrms.local/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"freelancer@example.com", "password":"..."}' \
  # Antwort: access_token, mit Rollen=["freelancer"]

# 2. Versucht Admin-Benutzer zu ändern
curl -X PUT https://myrms.local/api/users/admin-uuid \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"roles":[""]}'
  # Erwartet: 403 Forbidden

# 3. Mit Admin-Token: erlaubt
curl -X PUT https://myrms.local/api/users/freelancer-uuid \
  -H "Authorization: Bearer <admin_token>" \
  -d '{"roles":["crew_manager"]}'
  # Erwartet: 200 OK
```

### 4.2 A02:2021 – Cryptographic Failures

**Angriffsvektor:** Tokens ohne TLS übertragen, schwache Kryptographie

**Schutzmaßnahme (TLS + RS256 + Argon2):**

```go
// Server mit HTTPS erzwingwn
func main() {
    // In Production: TLS mit Let's Encrypt Zertifikat (via Traefik)
    server := &http.Server{
        Addr:    ":8443",
        Handler: router,
        TLSConfig: &tls.Config{
            MinVersion:               tls.VersionTLS12,
            CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
            PreferServerCipherSuites: true,
            CipherSuites: []uint16{
                tls.TLS_AES_256_GCM_SHA384,
                tls.TLS_CHACHA20_POLY1305_SHA256,
                tls.TLS_AES_128_GCM_SHA256,
            },
        },
    }

    // HSTS Header
    server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        router.ServeHTTP(w, r)
    })

    log.Fatal(server.ListenAndServeTLS("cert.pem", "key.pem"))
}

// Cookies immer secure + httpOnly
http.SetCookie(w, &http.Cookie{
    Name:     "refresh_token",
    Value:    token,
    Path:     "/",
    HttpOnly: true,  // JavaScript kann nicht zugreifen
    Secure:   true,  // HTTPS only
    SameSite: http.SameSiteLax,
})
```

**Database-Verschlüsselung (PGCrypto):**

```sql
-- Sensitive Felder verschlüsseln
CREATE EXTENSION pgcrypto;

ALTER TABLE auth_schema.users
    ADD password_hash_encrypted BYTEA;

-- Beim Speichern
UPDATE auth_schema.users
SET password_hash_encrypted = encrypt(password_hash::bytea, 'secret_key'::bytea, 'aes')
WHERE id = $1;

-- Beim Lesen
SELECT decrypt(password_hash_encrypted, 'secret_key'::bytea, 'aes') FROM users;
```

**Test-Szenario:**

```bash
# 1. HTTP-Request sollte Fehler geben
curl -X POST http://myrms.local/api/auth/login \
  -H "Content-Type: application/json" \
  # Erwartet: 301 Redirect zu HTTPS oder 403 Forbidden

# 2. TLS Version < 1.2 sollte abgelehnt werden
openssl s_client -connect myrms.local:443 -tls1_0
  # Erwartet: "TLSV1 ALERT PROTOCOL VERSION"

# 3. JWT ohne RS256 sollte abgelehnt werden
jwt_with_hs256 = jwt.encode(claims, "secret", algorithm="HS256")
curl -H "Authorization: Bearer $jwt_with_hs256" https://myrms.local/api/equipment
  # Erwartet: 401 Unauthorized
```

### 4.3 A03:2021 – Injection

**SQL Injection Prävention:**

```go
// FALSCH: SQL Injection Anfällig!
query := fmt.Sprintf("SELECT * FROM equipment WHERE name = '%s'", name)
rows, err := db.Query(query)

// RICHTIG: Parametrisierte Queries (pgx)
rows, err := db.Query(ctx, "SELECT * FROM equipment WHERE name = $1 AND tenant_id = $2", name, tenantID)

for rows.Next() {
    var eq Equipment
    rows.Scan(&eq.ID, &eq.Name, ...)
}
```

**Event Injection (KurrentDB):**

```go
// FALSCH: Unkontrolliertes Event-Encoding
func (s *Service) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    var evt map[string]interface{}
    json.NewDecoder(r.Body).Decode(&evt)
    // Direkt speichern → kann beliebige Struktur enthalten!
    s.eventStore.Append(ctx, "equipment-id", evt)
}

// RICHTIG: Strict Event-Typprüfung
type EquipmentCheckedOut struct {
    EquipmentID uuid.UUID `json:"equipment_id"`
    ProjectID   uuid.UUID `json:"project_id"`
    CheckedOutBy uuid.UUID `json:"checked_out_by"`
    Timestamp   time.Time `json:"timestamp"`
}

func (s *Service) HandleCheckOut(w http.ResponseWriter, r *http.Request) {
    var evt EquipmentCheckedOut
    if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
        http.Error(w, "invalid event", http.StatusBadRequest)
        return
    }
    // Validierung
    if evt.EquipmentID == uuid.Nil || evt.ProjectID == uuid.Nil {
        http.Error(w, "missing required fields", http.StatusBadRequest)
        return
    }
    s.eventStore.Append(ctx, "equipment-"+evt.EquipmentID.String(), evt)
}
```

**Input Validation (Go Validator):**

```go
import "github.com/go-playground/validator/v10"

type CreateEquipmentRequest struct {
    Name              string  `json:"name" validate:"required,max=255"`
    InternalNumber    string  `json:"internal_number" validate:"required,alphanumeric,max=50"`
    ManufacturerName  string  `json:"manufacturer" validate:"max=255"`
    QuantityTotal     int     `json:"quantity_total" validate:"required,gt=0,lt=100000"`
    PurchasePrice     float64 `json:"purchase_price" validate:"omitempty,gt=0"`
    ReplacementValue  float64 `json:"replacement_value" validate:"omitempty,gt=0"`
}

func (h *InventoryHandler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
    var req CreateEquipmentRequest
    json.NewDecoder(r.Body).Decode(&req)

    validate := validator.New()
    if err := validate.Struct(req); err != nil {
        // Return validation errors
        errs := err.(validator.ValidationErrors)
        for _, e := range errs {
            fmt.Printf("Field: %s, Tag: %s, Param: %s\n", e.Field(), e.Tag(), e.Param())
        }
        http.Error(w, "validation failed", http.StatusBadRequest)
        return
    }
    // Safe to proceed
}
```

**Test-Szenario:**

```bash
# 1. SQL Injection
curl -X GET "https://myrms.local/api/equipment?name=' OR '1'='1" \
  # Erwartet: 400 Bad Request oder 0 Ergebnisse (nicht exploitabel)

# 2. Command Injection (falls externer Befehl)
# z.B. beim PDF-Generieren mit Chromedp
curl -X POST "https://myrms.local/api/documents/generate" \
  -d '{"template":""; rm -rf /"}' \
  # Erwartet: 400 Bad Request, nicht ausgeführt

# 3. Event-Typ-Konfusion
curl -X POST "https://myrms.local/api/equipment/123/check-out" \
  -H "Content-Type: application/json" \
  -d '{"malicious_field":"value"}' \
  # Erwartet: 400 Bad Request oder ignoriert
```

### 4.4 A04:2021 – Insecure Design

**Sicheres Design-Pattern: Positive Allowlist**

```go
// FALSCH: Blacklist (was ist alles verboten?)
func isBlacklistedPermission(perm string) bool {
    blacklist := []string{"delete_all_users", "drop_database", ...}
    for _, b := range blacklist {
        if perm == b { return true }
    }
    return false
}

// RICHTIG: Whitelist (was ist alles erlaubt?)
var AllowedPermissions = []string{
    "equipment.read", "equipment.write", "equipment.delete",
    "project.read", "project.write",
    "invoice.read", "invoice.write",
    "users.read", "users.write",
    "roles.read", "roles.write",
    "warehouse.read", "warehouse.write",
    "crew.read", "crew.write",
    "federation.read", "federation.write",
    "audit.read",
    "admin.system",
}

func isValidPermission(perm string) bool {
    for _, allowed := range AllowedPermissions {
        if perm == allowed { return true }
    }
    return false
}
```

### 4.5 A05:2021 – Broken Authentication

**Mehrstufige Authentifizierung (MFA):**

```go
// TOTP (Time-based One-Time Password)
type MFASetupRequest struct {
    Method string `json:"method"` // "totp", "email", "sms"
}

func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    claims := ctx.Value("claims").(*AccessTokenClaims)

    var req MFASetupRequest
    json.NewDecoder(r.Body).Decode(&req)

    switch req.Method {
    case "totp":
        secret, err := totp.GenerateSecret(totp.GenerateSecretInput{
            Name:     claims.Email,
            Issuer:   "CrateDesk",
            AccountName: claims.Email,
        })
        // Zurück: QR-Code + Secret
        json.NewEncoder(w).Encode(map[string]interface{}{
            "secret": secret.Secret(),
            "qr_code": secret.String(), // Data-URI mit QR-Code
        })

    case "email":
        code := generateRandomCode(6)
        h.emailService.SendMFACode(ctx, claims.Email, code)
        h.mfaCodeRepo.Store(ctx, claims.Subject, "email", code)
        json.NewEncoder(w).Encode(map[string]string{"status": "code sent"})
    }
}

// Verify MFA
type VerifyMFARequest struct {
    MFAToken string `json:"mfa_token"`
    Code     string `json:"code"`
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var req VerifyMFARequest
    json.NewDecoder(r.Body).Decode(&req)

    mfaChallenge, _ := h.mfaChallengeRepo.Get(ctx, req.MFAToken)
    if mfaChallenge == nil || time.Now().After(mfaChallenge.ExpiresAt) {
        http.Error(w, "mfa challenge expired", http.StatusUnauthorized)
        return
    }

    valid := false
    switch mfaChallenge.Method {
    case "totp":
        valid = totp.Validate(req.Code, mfaChallenge.Secret)
    case "email":
        valid = req.Code == mfaChallenge.Code && time.Now().Before(mfaChallenge.ExpiresAt)
    }

    if !valid {
        http.Error(w, "invalid mfa code", http.StatusUnauthorized)
        return
    }

    // Neue JWT mit MFAVerified=true
    claims := &AccessTokenClaims{
        // ... übliche Claims
        MFAVerified: true,
        MFAMethod:   mfaChallenge.Method,
    }
    accessToken, _ := h.jwtService.SignToken(claims)

    json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}
```

### 4.6 A06:2021 – Vulnerable and Outdated Components

**Dependency Management:**

```bash
# Go modules für Dependency Tracking
go mod tidy  # Entfernt ungenutzte

# Regelmäßige Updates (wöchentlich)
go get -u all

# Security Audit
go list -json -m all | nancy sleuth

# Dockerfile: Alpine Base Image (minimal attack surface)
FROM golang:1.22-alpine AS builder
FROM alpine:3.19
# Nur Runtime-Binary kopieren, nicht ganzer Go Toolchain
```

**SBOM (Software Bill of Materials):**

```bash
# Alle Dependencies auflisten
go mod graph

# Mit Versionen
go list -json -m all | jq '.Version'
```

### 4.7 A07:2021 – Identification and Authentication Failures

**Session Timeout + Reauthentication:**

```go
const (
    AccessTokenLifetime  = 15 * time.Minute
    RefreshTokenLifetime = 7 * 24 * time.Hour
    InactivityTimeout    = 30 * time.Minute  // Auch mit gültigen Token
)

func (m *AuthMiddleware) CheckInactivity(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        claims := ctx.Value("claims").(*AccessTokenClaims)

        session, _ := m.sessionStore.GetSession(ctx, claims.JWTID)
        lastActivity := time.Unix(session.LastActivity, 0)

        if time.Since(lastActivity) > InactivityTimeout {
            m.sessionStore.Blacklist(ctx, claims.JWTID, "inactivity")
            http.Error(w, "session expired due to inactivity", http.StatusUnauthorized)
            return
        }

        // Update LastActivity
        m.sessionStore.UpdateActivity(ctx, claims.JWTID)
        next.ServeHTTP(w, r)
    })
}
```

### 4.8 A08:2021 – Software and Data Integrity Failures

**Software Integrity (Checksum Verification):**

```go
// Docker Images signieren (Cosign)
cosign sign --key cosign.key myrms/auth-service:1.0

// Verify
cosign verify --key cosign.pub myrms/auth-service:1.0

// In Deployment: signature verification erzwingen
```

**Data Integrity (Checksums für kritische Transaktionen):**

```go
// Bei Invoice-Erstellung: Prüfsumme generieren
type Invoice struct {
    ID             uuid.UUID
    Amount         float64
    Description    string
    CalculatedHash string // SHA-256 aller Felder
}

func (inv *Invoice) CalculateHash() string {
    data := fmt.Sprintf("%s|%f|%s", inv.ID, inv.Amount, inv.Description)
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}

// Bei Abruf: Hash validieren
func (repo *InvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*Invoice, error) {
    inv, _ := repo.db.QueryRow(...).Scan(&inv)
    if inv.CalculateHash() != inv.CalculatedHash {
        return nil, fmt.Errorf("data integrity check failed")
    }
    return inv, nil
}
```

### 4.9 A09:2021 – Logging and Monitoring Failures

**Audit Logging (alle Sicherheits-Events):**

```go
type AuditLog struct {
    ID        uuid.UUID `json:"id"`
    TenantID  uuid.UUID `json:"tenant_id"`
    UserID    uuid.UUID `json:"user_id"`
    Action    string    `json:"action"` // "login", "create_user", "delete_equipment"
    Resource  string    `json:"resource"`
    Result    string    `json:"result"` // "success", "failure"
    Reason    string    `json:"reason"` // Warum fehlgeschlagen
    IPAddress string    `json:"ip_address"`
    UserAgent string    `json:"user_agent"`
    Timestamp time.Time `json:"timestamp"`
}

func (h *AuthHandler) LogSecurityEvent(ctx context.Context, action string, result string, reason string) {
    claims := ctx.Value("claims").(*AccessTokenClaims)
    audit := &AuditLog{
        ID:        uuid.New(),
        TenantID:  claims.TenantID,
        UserID:    uuid.MustParse(claims.Subject),
        Action:    action,
        Result:    result,
        Reason:    reason,
        IPAddress: getRealIP(nil),
        Timestamp: time.Now(),
    }
    h.auditRepo.Create(ctx, audit)

    // Auch in Event Store (für GoBD Compliance)
    h.eventStore.Append(ctx, "audit", events.SecurityEventLogged{
        AuditLog:  audit,
        Timestamp: time.Now(),
    })
}
```

**Alerts & Monitoring (Prometheus):**

```go
var (
    LoginAttemppts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_login_attempts_total",
            Help: "Total login attempts",
        },
        []string{"result"},  // "success", "failed"
    )

    BruteForceDetections = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_brute_force_detections_total",
        },
        []string{"user_email"},
    )
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ...
    if !valid {
        LoginAttempts.WithLabelValues("failed").Inc()
        BruteForceDetections.WithLabelValues(req.Email).Inc()
    } else {
        LoginAttempts.WithLabelValues("success").Inc()
    }
}
```

### 4.10 A10:2021 – Server-Side Request Forgery (SSRF)

**SSRF Prevention bei Federation (mTLS P2P):**

```go
// FALSCH: Dynamische URL von Client
func (h *FederationHandler) RequestEquipment(w http.ResponseWriter, r *http.Request) {
    var req map[string]string
    json.NewDecoder(r.Body).Decode(&req)
    partnerURL := req["partner_url"]  // DANGER!

    // Könnte sein: http://localhost:6379 (Redis) oder http://metadata.service/
    client := &http.Client{}
    resp, _ := client.Get(partnerURL)  // SSRF!
}

// RICHTIG: Partner aus DB, kein Fretext-URL
type Partner struct {
    ID      uuid.UUID
    Name    string
    CertURL string  // z.B. https://partner.example.com/.well-known/certs
    MutualTLSCert *tls.Certificate
}

func (h *FederationHandler) RequestEquipment(w http.ResponseWriter, r *http.Request) {
    var req RequestEquipmentRequest
    json.NewDecoder(r.Body).Decode(&req)
    partnerID := uuid.MustParse(req.PartnerID)

    partner, _ := h.partnerRepo.FindByID(ctx, partnerID)
    if partner == nil {
        http.Error(w, "partner not found", http.StatusNotFound)
        return
    }

    // Nur TLS mit Zertifikat
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{*partner.MutualTLSCert},
        ServerName:   partner.Domain,
    }
    client := &http.Client{
        Transport: &http.Transport{TLSClientConfig: tlsConfig},
    }

    // URL ist Basis-URL des Partners, nicht frei konfigurierbar
    resp, _ := client.Get(partner.BaseURL + "/api/equipment")
}

// URL Validation (wenn doch dynamisch):
func ValidateInternalURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return err
    }

    // Nur HTTP/HTTPS erlaubt
    if u.Scheme != "http" && u.Scheme != "https" {
        return fmt.Errorf("invalid scheme")
    }

    // Keine lokalen IPs
    forbiddenHosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0", "169.254.169.254"}
    for _, h := range forbiddenHosts {
        if u.Host == h || strings.HasPrefix(u.Host, h+":") {
            return fmt.Errorf("access to %s forbidden", h)
        }
    }

    // Private IP ranges blocken
    ip := net.ParseIP(u.Hostname())
    if ip != nil && (ip.IsPrivate() || ip.IsLoopback()) {
        return fmt.Errorf("private ip address not allowed")
    }

    return nil
}
```

---

## 5. DSGVO-Compliance

### 5.1 Personenbezogene Daten in CrateDesk

| Datentyp | Tabelle | Zweck | Rechtsgrundlage | Löschfrist | Speicherort |
|----------|---------|-------|-----------------|-----------|-------------|
| **User-Stammdaten** | users (email, name) | Benutzer-Authentifizierung, Rechnungsadresse | Vertrag (Art. 6.1.b DSGVO) | Nach Kontoauflösung + 90 Tage | PostgreSQL |
| **IP-Adressen** | audit_log, session | Sicherheit, Brute-Force-Detection | Berechtigtes Interesse (Art. 6.1.f) | 90 Tage | PostgreSQL + Redis |
| **User-Agent** | session | Geräte-Binding zur Sicherheit | Berechtigtes Interesse | 90 Tage | Redis |
| **Crew-Daten** | crew (Name, Telefon, IBAN) | Zeiterfassung, Bezahlung | Vertrag + Arbeitsrecht | 3 Jahre (Steuern) | PostgreSQL |
| **Kundeninfos** | project.customer_details | Rechnungsstellung | Vertrag | 10 Jahre (GoBD) | PostgreSQL |
| **Zeitstempel Login** | audit_log.timestamp | Accountability | DSGVO Art. 5 | 90 Tage | PostgreSQL |
| **Cookie/Tracking** | - | NICHT in CrateDesk! | N/A | N/A | N/A |
| **Sensitive Payment Data** | - | NICHT speichern! (PCI-DSS) | - | Keine lokale Speicherung | Stripe/PayPal |

### 5.2 Recht auf Löschung (Art. 17 DSGVO)

**3-Tier Deletion Strategy:**

```go
type DeletionRequest struct {
    UserID uuid.UUID
    Reason string  // "account_closure", "data_breach", "business_decision"
    Force  bool    // Sofortige Hard-Delete (Admin only)
}

// TIER 1: Soft Delete (standardmäßig)
func (svc *UserService) SoftDeleteUser(ctx context.Context, userID uuid.UUID) error {
    // Nur Markieren als gelöscht
    _, err := svc.db.ExecContext(ctx,
        "UPDATE auth_schema.users SET deleted_at = NOW(), email = CONCAT('deleted_', $1) WHERE id = $2",
        uuid.New(), userID,
    )
    return err
}

// TIER 2: Anonymisierung (nach 90 Tagen)
func (job *AnonymizationJob) Run(ctx context.Context) error {
    // Täglich: Alle gelöschten User älter als 90 Tage anonymisieren
    deleted90DaysAgo := time.Now().AddDate(0, 0, -90)

    rows, _ := svc.db.QueryContext(ctx,
        "SELECT id FROM auth_schema.users WHERE deleted_at < $1 AND anonymized_at IS NULL",
        deleted90DaysAgo,
    )

    for rows.Next() {
        var userID uuid.UUID
        rows.Scan(&userID)

        // Anonymisieren
        svc.db.ExecContext(ctx, `
            UPDATE auth_schema.users
            SET
                email = CONCAT('anon_', $1),
                name = 'Anonymized User',
                phone = NULL,
                address = NULL,
                anonymized_at = NOW()
            WHERE id = $2
        `, uuid.New(), userID)

        // Events anonymisieren
        svc.eventStore.Append(ctx, "user-"+userID.String(), events.UserAnonymized{
            UserID:     userID,
            Timestamp:  time.Now(),
        })
    }

    return nil
}

// TIER 3: Hard Delete (nach 3 Jahren, für GDPR Art. 17 Erfüllung)
func (job *HardDeleteJob) Run(ctx context.Context) error {
    // Jährlich: Anonymisierte Daten nach 3 Jahren endgültig löschen
    deleteAfter := time.Now().AddDate(-3, 0, 0)

    svc.db.ExecContext(ctx,
        "DELETE FROM auth_schema.users WHERE anonymized_at < $1",
        deleteAfter,
    )

    // Event für Audit-Trail
    svc.eventStore.Append(ctx, "audit", events.UserHardDeleted{
        Count:     deletedCount,
        Timestamp: time.Now(),
    })

    return nil
}

// REST API für User-getriggerte Löschung
func (h *UserHandler) RequestDeletion(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    claims := ctx.Value("claims").(*AccessTokenClaims)
    userID := uuid.MustParse(claims.Subject)

    // 1. Event: DeletionRequested
    h.eventStore.Append(ctx, "user-"+userID.String(), events.UserDeletionRequested{
        UserID:    userID,
        RequestedAt: time.Now(),
    })

    // 2. Bestätigungs-Email versenden
    h.emailService.SendDeletionConfirmation(ctx, claims.Email)

    // 3. 14-Tage Wartefrist setzen
    h.db.ExecContext(ctx,
        "UPDATE auth_schema.users SET deletion_requested_at = NOW(), deletion_confirmation_token = $1 WHERE id = $2",
        uuid.New(), userID,
    )

    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "deletion requested, check email for confirmation",
        "expires_in_days": "14",
    })
}
```

**Event-basierte Anonymisierung (Audit-Trail):**

```go
// Events in KurrentDB dürfen NICHT gelöscht werden (immutable)
// Aber sensitive Daten anonymisieren
type UserAnonymizedEvent struct {
    UserID uuid.UUID `json:"user_id"`
    // Alle anderen Felder: NULL/redacted
    AnonymizedAt time.Time `json:"anonymized_at"`
}

// Event-Projektion ersetzt sensitives Daten in Read-Models
func (proj *EventProjector) ProjectUserAnonymized(ctx context.Context, evt UserAnonymizedEvent) {
    svc.db.ExecContext(ctx,
        `UPDATE auth_schema.users
         SET name='Anonymized', email=NULL, phone=NULL
         WHERE id = $1`, evt.UserID,
    )
}
```

### 5.3 Recht auf Auskunft (Art. 15 DSGVO)

**Data Export API:**

```go
// GET /api/users/me/export
func (h *UserHandler) ExportPersonalData(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    claims := ctx.Value("claims").(*AccessTokenClaims)
    userID := uuid.MustParse(claims.Subject)

    // Sammle alle Daten dieses Users
    userData := map[string]interface{}{}

    // User-Stammdaten
    user, _ := h.userRepo.FindByID(ctx, userID)
    userData["user"] = user

    // Crew-Daten
    crew, _ := h.crewRepo.FindByUserID(ctx, userID)
    userData["crew"] = crew

    // Projekte die dieser User erstellt/bearbeitet hat
    projects, _ := h.projectRepo.FindByCreator(ctx, userID)
    userData["projects"] = projects

    // Audit-Log dieses Users
    audits, _ := h.auditRepo.FindByUserID(ctx, userID)
    userData["audit_logs"] = audits

    // Sessions
    sessions, _ := h.sessionRepo.FindByUserID(ctx, userID)
    userData["sessions"] = sessions

    // Als JSON oder ZIP exportieren
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", "attachment; filename=personal-data.json")
    json.NewEncoder(w).Encode(userData)

    // Event für Audit
    h.eventStore.Append(ctx, "user-"+userID.String(), events.PersonalDataExported{
        UserID:    userID,
        ExportedAt: time.Now(),
    })
}
```

### 5.4 Datenminimierung (Art. 5.1.c DSGVO)

**Was wird NICHT gespeichert:**

```
✗ Passwort-Wiederherstellungstexte (nur Token)
✗ Kreditkarten-Nummern (PCI-DSS-Ausnahme, externe Zahlungs-Gateway)
✗ Cookies für Tracking (nur Session-Management)
✗ Genaue Geburtsdaten (nur Altersverifikation wenn nötig)
✗ Genetische oder biometrische Daten
✗ IP-Adressen länger als 90 Tage
✗ Detaillierte Aufenthaltsorte (nur Projekt-Ort)

✓ Gespeichert: Email, Name, Telefon (geschäftsnotwendig)
✓ Gespeichert: IBAN für Freelancer-Bezahlung (Geschäftsnotwendig)
✓ Gespeichert: IP + User-Agent (90 Tage, für Sicherheit)
✓ Gespeichert: Login-Timestamps (90 Tage Audit)
```

### 5.5 Verschlüsselung At-Rest + In-Transit

**At-Rest (PostgreSQL pgcrypto):**

```sql
-- Sensitive Felder verschlüsseln
CREATE EXTENSION pgcrypto;

ALTER TABLE auth_schema.users ADD COLUMN phone_encrypted BYTEA;
ALTER TABLE crew_schema.crew ADD COLUMN iban_encrypted BYTEA;

-- Insert mit Verschlüsselung
INSERT INTO auth_schema.users (id, phone_encrypted)
VALUES ($1, encrypt('+49123456789'::bytea, 'ENCRYPTION_KEY'::bytea, 'aes'));

-- Select mit Entschlüsselung
SELECT decrypt(phone_encrypted, 'ENCRYPTION_KEY'::bytea, 'aes') FROM users;

-- Encryption Key in Environment Variable
ENCRYPTION_KEY=<random_256bit_hex>
```

**In-Transit (TLS 1.3):**

```go
// Traefik Config (docker-compose.yml)
traefik:
  command:
    - "--entrypoints.websecure.address=:443"
    - "--certificatesresolvers.letsencrypt.acme.email=${ADMIN_EMAIL}"
    - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
    - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"

  # Nur TLS 1.3 + 1.2
  labels:
    - "traefik.http.middlewares.ssl-headers.headers.sslredirect=true"
    - "traefik.http.middlewares.ssl-headers.headers.sslhost=${DOMAIN}"
    - "traefik.http.middlewares.ssl-headers.headers.sslprotos=TLSv1.3,TLSv1.2"
    - "traefik.http.middlewares.ssl-headers.headers.sslminversion=VersionTLS13"
```

### 5.6 Auftragsverarbeitung (Data Processing Agreement)

**Was Admin konfigurieren muss:**

```yaml
# config/dpa-config.yaml
data_retention:
  audit_logs: 90 days
  sessions: 7 days
  deleted_users_anonymization: 90 days
  hard_delete_after: 3 years

encryption:
  at_rest: true
  algorithm: "AES-256-GCM"
  encryption_key: "${ENCRYPTION_KEY}"

data_export:
  enabled: true
  format: "JSON"
  frequency: "on-demand"

data_deletion:
  soft_delete_on_request: true
  anonymize_after_days: 90
  hard_delete_after_years: 3

third_party_processors:
  - name: "Email Service Provider"
    service: "SendGrid"
    purpose: "Transactional emails"
    dpa_signed: true

  - name: "Cloud Storage"
    service: "Self-hosted S3"
    purpose: "Invoice archival"
    dpa_signed: false  # Internal only
```

---

## 6. GoBD-Compliance (Rechnungsmodul)

### 6.1 Unveränderbarkeit von Rechnungen

**Technische Umsetzung:**

```sql
CREATE TABLE invoice_schema.invoices (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    invoice_number VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(50),  -- "draft", "issued", "paid", "cancelled"

    -- Unveränderbare Felder (nach Issuance)
    created_at TIMESTAMP NOT NULL,
    issued_at TIMESTAMP,
    issued_by UUID,
    content_hash CHAR(64),  -- SHA-256 Checksumme

    -- Änderungshistorie
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,

    UNIQUE(tenant_id, invoice_number)
);

-- Trigger: Nur bestimmte Felder darf man ändern (z.B. status, payment_date)
CREATE TRIGGER invoice_immutability
BEFORE UPDATE ON invoice_schema.invoices
FOR EACH ROW
EXECUTE FUNCTION check_invoice_immutability();

CREATE FUNCTION check_invoice_immutability()
RETURNS TRIGGER AS $$
BEGIN
    -- Felder die nach Issuance nicht geändert darf:
    IF NEW.issued_at IS NOT NULL THEN
        IF OLD.amount != NEW.amount
           OR OLD.content_hash != NEW.content_hash
           OR OLD.created_at != NEW.created_at THEN
            RAISE EXCEPTION 'invoice is immutable after issuance';
        END IF;
    END IF;

    -- Erlaubte Felder (nur diese):
    IF NEW.status NOT IN ('draft', 'issued', 'paid', 'cancelled') THEN
        RAISE EXCEPTION 'invalid status';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### 6.2 Prüfsummenkette (SHA-256 Chain)

```go
type InvoiceImmutabilityChain struct {
    InvoiceID       uuid.UUID
    ContentHash     string  // SHA-256(Amount + Description + Items)
    PreviousHash    string  // Verkettung mit vorheriger Rechnung
    ChainHash       string  // SHA-256(ThisHash + PreviousHash)
    CreatedAt       time.Time
}

func (svc *InvoiceService) IssueInvoice(ctx context.Context, inv *Invoice) error {
    // 1. Content Hash berechnen
    contentData := fmt.Sprintf(
        "%s|%.2f|%s",
        inv.ID,
        inv.Amount,
        inv.Description,
    )
    contentHash := sha256.Sum256([]byte(contentData))
    inv.ContentHash = hex.EncodeToString(contentHash[:])

    // 2. Vorherige Rechnung laden (für Verkettung)
    prevInvoice, _ := svc.repo.GetLastIssuedInvoice(ctx, inv.TenantID)
    prevHash := ""
    if prevInvoice != nil {
        prevHash = prevInvoice.ContentHash
    }

    // 3. Kette berechnen
    chainData := fmt.Sprintf("%s|%s", inv.ContentHash, prevHash)
    chainHash := sha256.Sum256([]byte(chainData))

    // 4. In DB speichern
    svc.repo.Create(ctx, inv)
    svc.repo.UpdateChainHash(ctx, inv.ID, hex.EncodeToString(chainHash[:]), prevHash)

    // Event
    svc.eventStore.Append(ctx, "invoice-"+inv.ID.String(), events.InvoiceIssued{
        InvoiceID:    inv.ID,
        Amount:       inv.Amount,
        ContentHash:  inv.ContentHash,
        ChainHash:    hex.EncodeToString(chainHash[:]),
        IssuedAt:     time.Now(),
    })

    return nil
}

// Validierung der Prüfsummenkette
func (svc *InvoiceService) VerifyIntegrity(ctx context.Context, invoiceID uuid.UUID) (bool, error) {
    inv, _ := svc.repo.GetByID(ctx, invoiceID)

    // Content Hash neu berechnen
    contentData := fmt.Sprintf("%s|%.2f|%s", inv.ID, inv.Amount, inv.Description)
    contentHash := sha256.Sum256([]byte(contentData))

    if hex.EncodeToString(contentHash[:]) != inv.ContentHash {
        return false, fmt.Errorf("content hash mismatch (tampering detected)")
    }

    // Chain Hash neu berechnen
    prevInvoice, _ := svc.repo.GetLastIssuedBefore(ctx, inv.TenantID, inv.CreatedAt)
    prevHash := ""
    if prevInvoice != nil {
        prevHash = prevInvoice.ContentHash
    }

    chainData := fmt.Sprintf("%s|%s", inv.ContentHash, prevHash)
    chainHash := sha256.Sum256([]byte(chainData))

    if hex.EncodeToString(chainHash[:]) != inv.ChainHash {
        return false, fmt.Errorf("chain hash mismatch (record tampered)")
    }

    return true, nil
}
```

### 6.3 Aufbewahrungsfristen

```yaml
retention_policies:
  invoices:
    retention_years: 10
    deletion_trigger: "invoice_date + 10 years"
    legal_basis: "German Tax Code (AStG §147)"

  delivery_notes:
    retention_years: 6
    deletion_trigger: "delivery_date + 6 years"
    legal_basis: "German Tax Code §257"

  business_letters:
    retention_years: 6
    deletion_trigger: "letter_date + 6 years"

  time_cards:
    retention_years: 3
    deletion_trigger: "period_end + 3 years"
    legal_basis: "German Labor Law"

  audit_logs:
    retention_years: 10
    deletion_trigger: "log_date + 10 years"
    legal_basis: "GoBD Audit Trail"
```

**Implementierung der Aufbewahrungsfrist:**

```go
type RetentionJob struct {
    db *sql.DB
}

func (job *RetentionJob) Run(ctx context.Context) error {
    // Täglich: Prüfe welche Rechnungen gelöscht werden können
    deleteBefore := time.Now().AddDate(-10, 0, 0)

    rows, _ := job.db.QueryContext(ctx,
        "SELECT id FROM invoice_schema.invoices WHERE issued_at < $1 AND deleted_at IS NULL",
        deleteBefore,
    )

    for rows.Next() {
        var invoiceID uuid.UUID
        rows.Scan(&invoiceID)

        // Soft-Delete (Archivierung)
        job.db.ExecContext(ctx,
            "UPDATE invoice_schema.invoices SET deleted_at = NOW() WHERE id = $1",
            invoiceID,
        )

        // Hard-Delete (nur nach nochmal 10 Jahren, also insgesamt 20)
    }

    return nil
}
```

### 6.4 Export für Betriebsprüfung (GDPdU/GoBD Format)

```go
// GET /api/admin/export/gobd?year=2024&format=gdsb
func (h *ExportHandler) ExportGOBD(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    year := r.URL.Query().Get("year")

    // 1. Alle Rechnungen des Jahres
    invoices, _ := h.invoiceRepo.FindByYear(ctx, year)

    // 2. Generiere ASC-Format (Standard für Betriebsprüfung)
    ascData := generateASCFormat(invoices)

    // 3. Digitale Signatur
    signature := h.signer.Sign(ascData)

    // 4. Als ZIP verpacken
    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=export_%s.zip", year))

    zipWriter := zip.NewWriter(w)

    // Daten
    f1, _ := zipWriter.Create("invoices.asc")
    f1.Write(ascData)

    // Signatur
    f2, _ := zipWriter.Create("signature.txt")
    f2.Write([]byte(signature))

    // Verfahrensdokumentation
    f3, _ := zipWriter.Create("README.txt")
    f3.Write([]byte(generateProcedureDocumentation()))

    zipWriter.Close()
}

func generateASCFormat(invoices []*Invoice) []byte {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)
    writer.Comma = ';'  // Standard für deutsche Steuerbehörden

    for _, inv := range invoices {
        writer.Write([]string{
            inv.ID.String(),
            inv.InvoiceNumber,
            inv.IssuedAt.Format("2006-01-02"),
            fmt.Sprintf("%.2f", inv.Amount),
            inv.CustomerName,
            inv.Description,
        })
    }

    writer.Flush()
    return buf.Bytes()
}
```

### 6.5 Verfahrensdokumentation

**Zu dokumentieren (GoBD §4 Abs. 3):**

```
1. Geschäftsvorfälle
   - Art und Weise der Erstellung
   - Verarbeitung
   - Speicherung
   - Abruf

2. System- und Organisationsmittel
   - Hardware
   - Software (Version, Update-History)
   - Benutzerrechte

3. Sicherheitsmaßnahmen
   - Verschlüsselung
   - Zugriffskontrolle
   - Backup-Strategie

4. Änderungen am System (Change Log)
   - Datum
   - Beschreibung
   - Begründung
```

**Als Dokument in der Anwendung speichern:**

```go
type ProcedureDocumentation struct {
    Title       string
    Version     string
    CreatedAt   time.Time
    Content     string  // Markdown
    LastUpdated time.Time
}

// In Tenant-Konfiguration speichern
func (svc *ConfigService) GetProcedureDocumentation(ctx context.Context, tenantID uuid.UUID) (*ProcedureDocumentation, error) {
    var doc ProcedureDocumentation
    svc.db.QueryRowContext(ctx,
        "SELECT title, version, content FROM tenant_config WHERE tenant_id = $1 AND key = 'procedure_documentation'",
        tenantID,
    ).Scan(&doc.Title, &doc.Version, &doc.Content)
    return &doc, nil
}
```

### 6.6 Audit-Trail (immutable Log)

```go
// audit-service subscribet auf alle Events und speichert in PostgreSQL
CREATE TABLE audit_schema.events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    event_type VARCHAR(100),
    aggregate_id UUID,
    aggregate_type VARCHAR(50),
    data JSONB,
    metadata JSONB,
    timestamp TIMESTAMP DEFAULT NOW(),
    checksum_previous CHAR(64),  -- Verkettung
    checksum_this CHAR(64),      -- Dieser Event

    UNIQUE(tenant_id, timestamp, id)
);

-- Immutable: RLS mit nur SELECT, kein UPDATE
CREATE POLICY audit_immutable ON audit_schema.events
    AS RESTRICTIVE
    FOR ALL
    USING (true);

-- Events vom audit-service
func (svc *AuditService) LogEvent(ctx context.Context, evt *Event) error {
    // Checksummen berechnen
    thisChecksum := svc.hashEvent(evt)

    // Vorherigen Event laden
    prevEvent, _ := svc.repo.GetLastEvent(ctx, evt.TenantID)
    prevChecksum := ""
    if prevEvent != nil {
        prevChecksum = prevEvent.ChecksumThis
    }

    // Speichern
    svc.repo.Create(ctx, &AuditLogEntry{
        ID:              uuid.New(),
        TenantID:        evt.TenantID,
        EventType:       evt.Type,
        AggregateID:     evt.AggregateID,
        Data:            evt.Data,
        ChecksumPrevious: prevChecksum,
        ChecksumThis:     thisChecksum,
        Timestamp:       time.Now(),
    })

    return nil
}
```

---

## 7. DGUV V3 / E-Check Integration

### 7.1 Prüfprotokolle (Datenmodell)

```sql
CREATE TABLE maintenance_schema.echeck_protocols (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    equipment_id UUID NOT NULL REFERENCES inventory_schema.equipment(id),

    -- Prüfer-Info
    inspector_name VARCHAR(255) NOT NULL,
    inspector_license VARCHAR(100),
    inspection_date TIMESTAMP NOT NULL,

    -- Prüf-Details
    visual_inspection BOOLEAN,
    continuity_test BOOLEAN,
    insulation_test BOOLEAN,
    leakage_current_test BOOLEAN,

    -- Messwerte
    insulation_resistance_ohms NUMERIC,
    leakage_current_ma NUMERIC,
    test_voltage_v INTEGER,

    -- Ergebnis
    result VARCHAR(50),  -- "passed", "failed", "conditional"
    remarks TEXT,

    -- Next Inspection
    next_inspection_due DATE,

    -- GoBD Compliance
    created_at TIMESTAMP DEFAULT NOW(),
    hash_checksum CHAR(64),  -- SHA-256 des Protokolls

    UNIQUE(equipment_id, inspection_date)
);
```

### 7.2 IZYTRON.IQ Import

```go
// Datei-Upload von IZYTRON.IQ Export
type IZYTRONImportRequest struct {
    File     multipart.FileHeader `form:"file"`
    TenantID uuid.UUID           `form:"tenant_id"`
}

func (h *MaintenanceHandler) ImportIZYTRON(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    file, _, _ := r.FormFile("file")
    defer file.Close()

    // CSV Parser (IZYTRON Export Format)
    reader := csv.NewReader(file)
    records, _ := reader.ReadAll()

    importedCount := 0
    for _, record := range records {
        // CSV Columns: EquipmentID, InspectionDate, Voltage, Resistance, ...
        eq := &ECheckProtocol{
            ID:                     uuid.New(),
            TenantID:               uuid.MustParse(r.FormValue("tenant_id")),
            EquipmentID:            uuid.MustParse(record[0]),
            InspectionDate:         parseDate(record[1]),
            TestVoltage:            parseInt(record[2]),
            InsulationResistance:   parseFloat(record[3]),
            LeakageCurrent:         parseFloat(record[4]),
            Result:                 record[5],  // "passed" oder "failed"
        }

        h.maintenanceRepo.CreateECheckProtocol(ctx, eq)
        importedCount++
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]int{"imported": importedCount})
}
```

### 7.3 Prüffristen & Automatische Erinnerungen

```go
type ECheckReminder struct {
    EquipmentID uuid.UUID
    LastCheck   time.Time
    NextDue     time.Time  // Automatisch berechnet
    DaysUntilDue int
    Status      string     // "ok", "warning", "overdue"
}

// Job: Täglich prüfen welche Geräte überfällige Prüfungen haben
func (job *ECheckReminderJob) Run(ctx context.Context) error {
    // Equipment, bei dem E-Check überfällig ist
    overdue, _ := job.repo.FindOverdueEChecks(ctx)

    for _, eq := range overdue {
        // 1. E-Mail an Prüfer
        job.notificationService.SendEmail(ctx, &Notification{
            Type:     "echeck_overdue",
            Recipient: eq.ResponsibleParty,
            Subject:  fmt.Sprintf("E-Check überfällig: %s", eq.Name),
            Message:  fmt.Sprintf("Letzte Prüfung: %s, überfällig seit: %d Tagen", eq.LastCheckDate, eq.DaysOverdue),
        })

        // 2. Equipment sperren (kann nicht ausgecheckt werden)
        job.repo.UpdateStatus(ctx, eq.ID, "echeck_overdue")

        // Event
        job.eventStore.Append(ctx, "equipment-"+eq.ID.String(), events.ECheckOverdue{
            EquipmentID: eq.ID,
            LastCheckDate: eq.LastCheckDate,
            OverdueBy: eq.DaysOverdue,
        })
    }

    return nil
}

// Prüffristen nach DGUV V3
var ECheckIntervals = map[string]int{
    "office_equipment":    24,  // Monatlich
    "site_equipment":      12,  // Halbjährlich
    "fixed_installation": 24,   // Jährlich
    "high_risk":           6,   // Vierteljährlich
}
```

---

## 8. API Security

### 8.1 Rate Limiting

```go
// Redis-basiertes Distributed Rate Limiting
type RateLimiter struct {
    redis *redis.Client
}

// Per Endpoint konfigurierbar
type RateLimitConfig struct {
    RequestsPerSecond float64  // z.B. 10.0
    BurstCapacity     int      // z.B. 20
    Window            time.Duration  // z.B. 1 Minute
}

var EndpointLimits = map[string]*RateLimitConfig{
    "/api/auth/login":    {RequestsPerSecond: 5, BurstCapacity: 10, Window: 1 * time.Minute},
    "/api/equipment":     {RequestsPerSecond: 100, BurstCapacity: 200, Window: 1 * time.Minute},
    "/api/invoices":      {RequestsPerSecond: 50, BurstCapacity: 100, Window: 1 * time.Minute},
    "/api/admin/*":       {RequestsPerSecond: 10, BurstCapacity: 20, Window: 1 * time.Minute},
}

func (rl *RateLimiter) Limit(ctx context.Context, clientID string, config *RateLimitConfig) error {
    key := fmt.Sprintf("ratelimit:%s", clientID)

    // Token Bucket Algorithm
    current, _ := rl.redis.Incr(ctx, key).Val(), nil

    if current == 1 {
        // Erste Request in diesem Window
        rl.redis.Expire(ctx, key, config.Window)
    }

    allowedPerSecond := int(config.RequestsPerSecond * config.Window.Seconds())
    if current > allowedPerSecond {
        // Rate Limit exceeded
        remaining := rl.redis.TTL(ctx, key).Val()
        return fmt.Errorf("rate limit exceeded, retry after %d seconds", int(remaining.Seconds()))
    }

    return nil
}

// Middleware
func (m *RateLimitMiddleware) Limit(endpoint string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("claims").(*AccessTokenClaims)
            clientID := claims.Subject

            config := EndpointLimits[endpoint]
            if config == nil {
                config = &RateLimitConfig{RequestsPerSecond: 100, BurstCapacity: 200}
            }

            if err := m.limiter.Limit(r.Context(), clientID, config); err != nil {
                w.Header().Set("Retry-After", "60")
                http.Error(w, err.Error(), http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 8.2 Input Validation (Go Validator Tags)

```go
type CreateInvoiceRequest struct {
    CustomerName     string  `json:"customer_name" validate:"required,max=255"`
    CustomerEmail    string  `json:"customer_email" validate:"required,email"`
    InvoiceDate      string  `json:"invoice_date" validate:"required,datetime=2006-01-02"`
    DueDate          string  `json:"due_date" validate:"required,datetime=2006-01-02"`
    Items            []Item  `json:"items" validate:"required,dive,max=1000"`
    Amount           float64 `json:"amount" validate:"required,gt=0,lt=999999.99"`
    VAT              float64 `json:"vat" validate:"omitempty,gte=0,lte=100"`
    Currency         string  `json:"currency" validate:"required,len=3,uppercase"`
}

type Item struct {
    Description string  `validate:"required,max=500"`
    Quantity    int     `validate:"required,gt=0,lt=100000"`
    UnitPrice   float64 `validate:"required,gt=0"`
}

func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateInvoiceRequest
    json.NewDecoder(r.Body).Decode(&req)

    validate := validator.New()
    if err := validate.Struct(req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        for _, err := range err.(validator.ValidationErrors) {
            fmt.Fprintf(w, "Field %s failed validation: %s\n", err.Field(), err.Tag())
        }
        return
    }

    // Validated data, safe to process
}
```

### 8.3 SQL Injection Prevention

```go
// FALSCH
query := fmt.Sprintf("SELECT * FROM equipment WHERE name = '%s'", userInput)

// RICHTIG - pgx mit parametrisierten Queries
rows, err := db.Query(ctx,
    "SELECT * FROM equipment WHERE name = $1 AND tenant_id = $2",
    userInput,  // Wird automatisch escaped
    tenantID,
)
```

### 8.4 XSS Prevention (Content-Security-Policy)

```go
func (m *SecurityHeadersMiddleware) Apply(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; "+
            "script-src 'self' 'nonce-{random}' cdn.jsdelivr.net; "+
            "style-src 'self' 'unsafe-inline' fonts.googleapis.com; "+
            "img-src 'self' data: https:; "+
            "font-src 'self' fonts.gstatic.com; "+
            "connect-src 'self' https:; "+
            "frame-ancestors 'none'; "+
            "base-uri 'self'; "+
            "form-action 'self'",
        )
        next.ServeHTTP(w, r)
    })
}
```

### 8.5 CORS Configuration

```go
// GET /api/equipment
// POST /api/equipment
// PUT /api/equipment/{id}
// DELETE /api/equipment/{id}

type CORSConfig struct {
    AllowedOrigins   []string      // z.B. ["https://myrms.local", "https://app.myrms.local"]
    AllowedMethods   []string      // ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    AllowedHeaders   []string      // ["Authorization", "Content-Type"]
    ExposedHeaders   []string      // ["X-Total-Count", "X-Page"]
    AllowCredentials bool          // true (für Cookies)
    MaxAge           int           // 86400 (1 Tag)
}

func (m *CORSMiddleware) Apply(config *CORSConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // Whitelist check
            allowed := false
            for _, o := range config.AllowedOrigins {
                if origin == o {
                    allowed = true
                    break
                }
            }

            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Credentials", "true")
                w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
                w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
                w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
                w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
            }

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 8.6 Security Headers

```go
var SecurityHeaders = map[string]string{
    "Strict-Transport-Security":      "max-age=31536000; includeSubDomains; preload",
    "X-Content-Type-Options":         "nosniff",
    "X-Frame-Options":                "DENY",
    "X-XSS-Protection":               "1; mode=block",
    "Referrer-Policy":                "strict-origin-when-cross-origin",
    "Permissions-Policy":             "geolocation=(), microphone=(), camera=()",
    "Cache-Control":                  "no-store, no-cache, must-revalidate, proxy-revalidate",
    "Pragma":                         "no-cache",
    "Expires":                        "0",
}

func (m *SecurityHeadersMiddleware) Apply(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        for header, value := range SecurityHeaders {
            w.Header().Set(header, value)
        }
        next.ServeHTTP(w, r)
    })
}
```

### 8.7 API Key Management (für externe Integrationen)

```sql
CREATE TABLE api_schema.api_keys (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255),
    key_hash CHAR(64),  -- SHA-256(key)
    scope VARCHAR(100),  -- "invoices.read", "equipment.write"
    rate_limit INTEGER,  -- Requests pro Minute
    created_at TIMESTAMP DEFAULT NOW(),
    last_used TIMESTAMP,
    expires_at TIMESTAMP,
    revoked_at TIMESTAMP,

    UNIQUE(key_hash)
);
```

```go
// Generiere API Key
func (svc *APIKeyService) GenerateKey(ctx context.Context, tenantID uuid.UUID, name string, scope string) (string, error) {
    // 32-byte random key
    b := make([]byte, 32)
    rand.Read(b)
    key := base64.URLEncoding.EncodeToString(b)
    keyHash := hashAPIKey(key)

    svc.db.ExecContext(ctx,
        `INSERT INTO api_schema.api_keys (id, tenant_id, name, key_hash, scope, created_at)
         VALUES ($1, $2, $3, $4, $5, NOW())`,
        uuid.New(), tenantID, name, keyHash, scope,
    )

    return key, nil  // Nur einmal angezeigt!
}

// Validiere API Key Header
func (m *APIKeyMiddleware) Validate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            http.Error(w, "missing api key", http.StatusUnauthorized)
            return
        }

        keyHash := hashAPIKey(apiKey)
        apiKeyRecord, _ := m.repo.FindByHash(r.Context(), keyHash)
        if apiKeyRecord == nil || apiKeyRecord.RevokedAt != nil {
            http.Error(w, "invalid api key", http.StatusUnauthorized)
            return
        }

        // Rate limit check
        if m.limiter.ExceedsLimit(r.Context(), apiKeyRecord.ID, apiKeyRecord.RateLimit) {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        ctx := context.WithValue(r.Context(), "tenant_id", apiKeyRecord.TenantID)
        ctx = context.WithValue(ctx, "api_key_id", apiKeyRecord.ID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 8.8 mTLS für Federation

```go
// Alle Federation P2P Requests müssen mit mTLS signiert sein
type mTLSConfig struct {
    CertFile string  // /etc/myrms/federation.crt
    KeyFile  string  // /etc/myrms/federation.key
    CAFile   string  // /etc/myrms/federation-ca.crt
}

func (svc *FederationService) CreateSecureClient(config *mTLSConfig) (*http.Client, error) {
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    if err != nil {
        return nil, err
    }

    caCert, _ := ioutil.ReadFile(config.CAFile)
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        ClientAuth:   tls.RequireAndVerifyClientCert,
        MinVersion:   tls.VersionTLS13,
    }

    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: tlsConfig,
        },
    }

    return client, nil
}
```

---

## 9. Backup & Disaster Recovery

### 9.1 Backup-Strategie

```bash
#!/bin/bash
# backup-script.sh

BACKUP_DIR="/backups"
ENCRYPTION_KEY="$BACKUP_ENCRYPTION_KEY"  # 32-byte Key aus Environment

# 1. PostgreSQL Dump (vollständig)
pg_dump \
    --host=db \
    --username=myrms \
    --format=directory \
    --file="$BACKUP_DIR/postgres_$(date +%Y%m%d_%H%M%S)" \
    myrms

# 2. KurrentDB Backup (Event Store)
# KurrentDB v26 hat built-in Backup API
curl -X POST http://kurrentdb:2113/admin/backup \
    -d '{"path":"/var/lib/kurrentdb/backups"}' \
    > "$BACKUP_DIR/kurrentdb_$(date +%Y%m%d_%H%M%S).tar.gz"

# 3. Redis RDB Snapshot
redis-cli --rdb "$BACKUP_DIR/redis_$(date +%Y%m%d_%H%M%S).rdb"

# 4. Verschlüsselung (AES-256-GCM)
for backup in "$BACKUP_DIR"/*.{sql,tar.gz,rdb}; do
    openssl enc -aes-256-cbc -e -in "$backup" -out "$backup.enc" -K "$ENCRYPTION_KEY"
    rm "$backup"  # Unverschlüsselten Backup löschen
done

# 5. S3 Upload (oder andere Remote Storage)
aws s3 sync "$BACKUP_DIR" "s3://myrms-backups/$(date +%Y/%m/%d)/" --sse AES256

# 6. Lokal aufbewahren (7 täglich, 4 wöchentlich, 3 monatlich)
# Automatische Cleanup mittels Lifecycle Policy
```

**Backup-Konfiguration in Docker Compose:**

```yaml
services:
  backup:
    image: myrms/backup:latest
    environment:
      BACKUP_ENCRYPTION_KEY: ${BACKUP_ENCRYPTION_KEY}
      AWS_ACCESS_KEY_ID: ${AWS_ACCESS_KEY_ID}
      AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY}
      AWS_REGION: eu-central-1
      BACKUP_S3_BUCKET: ${BACKUP_S3_BUCKET}
      BACKUP_SCHEDULE: "0 2 * * *"  # Täglich 02:00 Uhr

    volumes:
      - backups:/backups

    depends_on:
      - db
      - kurrentdb
      - redis
```

### 9.2 Encryption: AES-256-GCM

```go
type BackupEncryption struct {
    keyDerivation func(password string) []byte
}

func (be *BackupEncryption) EncryptBackup(ctx context.Context, backupPath string, encryptionKey string) (string, error) {
    // 1. Backup-Datei lesen
    backupData, _ := ioutil.ReadFile(backupPath)

    // 2. IV (Initialization Vector) generieren
    iv := make([]byte, 12)  // GCM Standard: 12 Bytes
    rand.Read(iv)

    // 3. Cipher-Block erstellen
    keyBytes := []byte(encryptionKey)  // Idealerweise: Key Derivation Function (PBKDF2)
    block, _ := aes.NewCipher(keyBytes)
    aesgcm, _ := cipher.NewGCM(block)

    // 4. Verschlüsseln
    ciphertext := aesgcm.Seal(nil, iv, backupData, nil)

    // 5. IV + Ciphertext speichern
    encryptedData := append(iv, ciphertext...)
    encryptedPath := backupPath + ".enc"
    ioutil.WriteFile(encryptedPath, encryptedData, 0600)

    return encryptedPath, nil
}

func (be *BackupEncryption) DecryptBackup(encryptedPath string, encryptionKey string) ([]byte, error) {
    // Inverse: Read IV, Decrypt
    encryptedData, _ := ioutil.ReadFile(encryptedPath)

    iv := encryptedData[:12]
    ciphertext := encryptedData[12:]

    keyBytes := []byte(encryptionKey)
    block, _ := aes.NewCipher(keyBytes)
    aesgcm, _ := cipher.NewGCM(block)

    plaintext, _ := aesgcm.Open(nil, iv, ciphertext, nil)
    return plaintext, nil
}
```

### 9.3 Retention Policy

```yaml
backup_retention:
  # Development
  dev:
    daily_count: 3
    weekly_count: 0
    monthly_count: 0

  # Production
  production:
    daily_count: 7      # 7 Tage täglich
    weekly_count: 4     # 4 Wochen wöchentlich
    monthly_count: 3    # 3 Monate monatlich

  # Archive (Compliance: 10 Jahre für Rechnungen)
  archive:
    yearly_count: 10

lifecycle_policy:
  daily_to_weekly: "after 7 days"
  weekly_to_monthly: "after 4 weeks"
  monthly_to_archive: "after 3 months"
  delete_archive: "after 10 years"
```

### 9.4 Restore-Procedure

```bash
#!/bin/bash
# restore-backup.sh

BACKUP_DATE="20260320"
BACKUP_ENCRYPTION_KEY="$BACKUP_ENCRYPTION_KEY"
RESTORE_DIR="/restore"

# 1. Backup von S3 downloaden
aws s3 cp "s3://myrms-backups/2026/03/20/postgres.sql.enc" "$RESTORE_DIR/"
aws s3 cp "s3://myrms-backups/2026/03/20/kurrentdb.tar.gz.enc" "$RESTORE_DIR/"
aws s3 cp "s3://myrms-backups/2026/03/20/redis.rdb.enc" "$RESTORE_DIR/"

# 2. Entschlüsseln
openssl enc -aes-256-cbc -d -in "$RESTORE_DIR/postgres.sql.enc" \
    -out "$RESTORE_DIR/postgres.sql" -K "$BACKUP_ENCRYPTION_KEY"

# 3. Datenbank stoppen
docker-compose stop db kurrentdb redis

# 4. Daten wiederherstellen
psql -U myrms -d myrms < "$RESTORE_DIR/postgres.sql"
# KurrentDB restore procedure (vendor-specific)
# Redis RDB restore

# 5. Services neu starten
docker-compose up -d db kurrentdb redis

# 6. Verifizierung
echo "Checking database integrity..."
psql -U myrms -d myrms -c "SELECT COUNT(*) FROM inventory_schema.equipment;"

# 7. Cleanup
rm -rf "$RESTORE_DIR"
```

**Go Code für Restore-Workflow:**

```go
type RestoreService struct {
    db *sql.DB
    s3Client *s3.Client
}

func (svc *RestoreService) RestoreFromBackup(ctx context.Context, backupDate time.Time) error {
    // 1. Backup-Dateien von S3 laden
    pgBackup, _ := svc.s3Client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String("myrms-backups"),
        Key:    aws.String(fmt.Sprintf("postgres_%s.sql.enc", backupDate.Format("20060102"))),
    })
    defer pgBackup.Body.Close()

    // 2. Entschlüsseln
    decrypted, _ := svc.decrypt(pgBackup.Body)

    // 3. In leere Datenbank importieren
    // ... SQL execution

    // 4. Validierung
    var count int
    svc.db.QueryRow("SELECT COUNT(*) FROM inventory_schema.equipment").Scan(&count)
    if count == 0 {
        return fmt.Errorf("restoration verification failed: no equipment found")
    }

    return nil
}
```

### 9.5 Recovery Time Objective (RTO) < 30 Minuten

```
RTO-Ziel: 30 Minuten vom Fehler bis System online

Aufschlüsselung:
1. Fehler erkannt: 2 Minuten (Monitoring Alert)
2. Backup-Download: 5 Minuten (S3 → lokale Disk)
3. Entschlüsselung: 2 Minuten (AES-256)
4. Database Restore: 15 Minuten (PostgreSQL Import)
5. Service-Start: 2 Minuten (Container restart)
6. Smoke Tests: 4 Minuten (DB Connectivity, API Health)

Total: 30 Minuten

Pre-requisites für RTO:
- Backups mindestens täglich
- Schnelle Netzwerk-Verbindung (S3)
- Pre-provisioned Restore-Hardware
- Automatisiertes Restore-Skript
```

### 9.6 Recovery Point Objective (RPO) < 1 Stunde

```
RPO-Ziel: Maximal 1 Stunde Datenverlust

Backup-Frequenz:
- Stündlich: PostgreSQL (via pg_dump incremental)
- Stündlich: KurrentDB (Relational Sink ist Live-Projektion)
- Stündlich: Redis (RDB + AOF)

Falls Ausfall zwischen Backup und Fehler:
- Events in KurrentDB: Nicht verloren (immutable)
- Read-Models in PostgreSQL: Können rekonstruiert werden (Replay)
- Sessions in Redis: Maximal 1 Stunde Datenverlust

Nachbericht nach Ausfalls-Incident:
- Welche Events sind verloren gegangen? (keiner, siehe KurrentDB)
- Welche Transaktionen sind verloren? (maximal 1 Stunde)
- Welche Business-Daten sind aktuell? (Read-Models aus KurrentDB Replay)
```

---

## Implementierungs-Checkliste

### Phase 1: Authentication & JWT (Woche 1-2)

- [ ] JWT RS256 Key-Pair generieren (RSA 2048)
- [ ] Login-Handler mit Argon2id Password Hashing
- [ ] Access Token + Refresh Token Lifecycle
- [ ] Redis Session Store
- [ ] JWT Middleware + RBAC Middleware
- [ ] Unit Tests für Brute-Force Protection
- [ ] Integration Tests für Token Refresh

### Phase 2: Multi-Tenancy & Isolation (Woche 3)

- [ ] PostgreSQL Row-Level Security Policies
- [ ] Tenant Context Extraction (Middleware)
- [ ] Tenant Provisioning API
- [ ] Cross-Tenant Isolation Tests

### Phase 3: API Security (Woche 4-5)

- [ ] Input Validation (go-playground/validator)
- [ ] Rate Limiting (Redis Token Bucket)
- [ ] CORS Configuration
- [ ] Security Headers (CSP, HSTS, etc.)
- [ ] mTLS für Federation

### Phase 4: DSGVO Compliance (Woche 6)

- [ ] Data Export API (Art. 15)
- [ ] Deletion Request Handler (Art. 17, soft + hard delete)
- [ ] Audit Logging (90-Tage Retention)
- [ ] PII Anonymization Job

### Phase 5: GoBD Compliance (Woche 7)

- [ ] Invoice Immutability Enforcement
- [ ] Prüfsummenkette (SHA-256)
- [ ] Audit-Trail in Event Store
- [ ] GoDPdU Export

### Phase 6: Backup & DR (Woche 8)

- [ ] PostgreSQL Backup Script
- [ ] KurrentDB Backup Integration
- [ ] AES-256-GCM Encryption
- [ ] S3 Upload + Lifecycle Policy
- [ ] Restore Procedure (automated)
- [ ] RTO/RPO Tests

### Phase 7: Testing & Hardening (Woche 9-10)

- [ ] OWASP Top 10 Security Tests
- [ ] Penetration Testing (externe)
- [ ] Load Testing (Rate Limit Verification)
- [ ] Disaster Recovery Drill
- [ ] Security Audit

---

**Dokumentversion:** 1.0
**Letztes Update:** 21. März 2026
**Autor:** Security Engineering Team
**Status:** Implementation Ready

