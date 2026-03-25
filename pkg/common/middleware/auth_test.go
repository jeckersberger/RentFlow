package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewSimpleJWT(t *testing.T) {
	secret := "test-secret"
	jwt := NewSimpleJWT(secret)

	if jwt == nil {
		t.Error("expected SimpleJWT to be created, got nil")
	}

	if jwt.secret != secret {
		t.Errorf("expected secret to be '%s', got '%s'", secret, jwt.secret)
	}
}

func TestCreateToken(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID:   "user123",
		TenantID: "tenant456",
		Email:    "user@example.com",
		Roles:    []string{"admin", "user"},
	}

	token, err := jwt.CreateToken(claims)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if token == "" {
		t.Error("expected token to be generated")
	}

	// Token should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected token to have 3 parts, got %d", len(parts))
	}
}

func TestCreateToken_SetsIssuedAt(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID: "user123",
	}

	before := time.Now().Unix()
	token, _ := jwt.CreateToken(claims)
	after := time.Now().Unix()

	// Verify token
	verifyClaims, _ := jwt.VerifyToken(token)

	if verifyClaims.IssuedAt < before || verifyClaims.IssuedAt > after+1 {
		t.Errorf("expected IssuedAt between %d and %d, got %d", before, after, verifyClaims.IssuedAt)
	}
}

func TestCreateToken_SetsExpiration(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID: "user123",
	}

	token, _ := jwt.CreateToken(claims)
	verifyClaims, _ := jwt.VerifyToken(token)

	// Should expire in about 1 hour (3600 seconds)
	expiresIn := verifyClaims.ExpiresAt - verifyClaims.IssuedAt

	if expiresIn < 3599 || expiresIn > 3601 {
		t.Errorf("expected token to expire in ~3600 seconds, got %d", expiresIn)
	}
}

func TestVerifyToken_Valid(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	originalClaims := &Claims{
		UserID:   "user123",
		TenantID: "tenant456",
		Email:    "user@example.com",
		Roles:    []string{"admin"},
	}

	token, _ := jwt.CreateToken(originalClaims)
	verifyClaims, err := jwt.VerifyToken(token)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if verifyClaims.UserID != originalClaims.UserID {
		t.Errorf("expected UserID '%s', got '%s'", originalClaims.UserID, verifyClaims.UserID)
	}

	if verifyClaims.TenantID != originalClaims.TenantID {
		t.Errorf("expected TenantID '%s', got '%s'", originalClaims.TenantID, verifyClaims.TenantID)
	}

	if verifyClaims.Email != originalClaims.Email {
		t.Errorf("expected Email '%s', got '%s'", originalClaims.Email, verifyClaims.Email)
	}

	if len(verifyClaims.Roles) != 1 || verifyClaims.Roles[0] != "admin" {
		t.Errorf("expected Roles [admin], got %v", verifyClaims.Roles)
	}
}

func TestVerifyToken_InvalidFormat(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")

	_, err := jwt.VerifyToken("invalid")

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyToken_InvalidSignature(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID: "user123",
	}

	token, _ := jwt.CreateToken(claims)

	// Tamper with the token by changing the last character of the signature
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		lastChar := parts[2][len(parts[2])-1:]
		newLastChar := string(rune(lastChar[0]) + 1)
		parts[2] = parts[2][:len(parts[2])-1] + newLastChar
		tamperedToken := strings.Join(parts, ".")

		_, err := jwt.VerifyToken(tamperedToken)

		if err != ErrInvalidSignature {
			t.Errorf("expected ErrInvalidSignature, got %v", err)
		}
	}
}

func TestVerifyToken_TooFewParts(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")

	_, err := jwt.VerifyToken("header.payload")

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyToken_TooManyParts(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")

	_, err := jwt.VerifyToken("header.payload.signature.extra")

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyToken_ExpiredToken(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID:    "user123",
		IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}

	// We need to manually create a token with past expiration
	// Use a different approach: create token and manually modify expiration
	_, _ = jwt.CreateToken(claims)

	// Since we set expiration in CreateToken, we need to use a custom approach
	// For now, we'll test the expiration logic separately
	// This test demonstrates the concept but may need adjustment based on implementation
}

func TestWithClaims(t *testing.T) {
	ctx := context.Background()
	claims := &Claims{
		UserID: "user123",
	}

	newCtx := WithClaims(ctx, claims)

	if newCtx == ctx {
		t.Error("expected new context with claims added")
	}

	retrieved := ExtractClaims(newCtx)
	if retrieved.UserID != claims.UserID {
		t.Errorf("expected UserID '%s', got '%s'", claims.UserID, retrieved.UserID)
	}
}

func TestExtractClaims_NotPresent(t *testing.T) {
	ctx := context.Background()

	claims := ExtractClaims(ctx)

	if claims != nil {
		t.Error("expected nil when claims not present in context")
	}
}

func TestGetUserID(t *testing.T) {
	ctx := context.Background()
	claims := &Claims{
		UserID: "user123",
	}

	ctx = WithClaims(ctx, claims)
	userID := GetUserID(ctx)

	if userID != "user123" {
		t.Errorf("expected UserID 'user123', got '%s'", userID)
	}
}

func TestGetUserID_NoClaims(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)

	if userID != "" {
		t.Errorf("expected empty string, got '%s'", userID)
	}
}

func TestGetTenantIDFromClaims(t *testing.T) {
	ctx := context.Background()
	claims := &Claims{
		TenantID: "tenant456",
	}

	ctx = WithClaims(ctx, claims)
	tenantID := GetTenantIDFromClaims(ctx)

	if tenantID != "tenant456" {
		t.Errorf("expected TenantID 'tenant456', got '%s'", tenantID)
	}
}

func TestGetEmail(t *testing.T) {
	ctx := context.Background()
	claims := &Claims{
		Email: "user@example.com",
	}

	ctx = WithClaims(ctx, claims)
	email := GetEmail(ctx)

	if email != "user@example.com" {
		t.Errorf("expected Email 'user@example.com', got '%s'", email)
	}
}

func TestGetRoles(t *testing.T) {
	ctx := context.Background()
	claims := &Claims{
		Roles: []string{"admin", "moderator"},
	}

	ctx = WithClaims(ctx, claims)
	roles := GetRoles(ctx)

	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}

	if roles[0] != "admin" || roles[1] != "moderator" {
		t.Errorf("expected [admin, moderator], got %v", roles)
	}
}

func TestGetRoles_NoClaims(t *testing.T) {
	ctx := context.Background()
	roles := GetRoles(ctx)

	if len(roles) != 0 {
		t.Errorf("expected empty roles, got %v", roles)
	}
}

func TestHasRole(t *testing.T) {
	ctx := context.Background()
	claims := &Claims{
		Roles: []string{"admin", "user"},
	}

	ctx = WithClaims(ctx, claims)

	if !HasRole(ctx, "admin") {
		t.Error("expected HasRole to return true for 'admin'")
	}

	if !HasRole(ctx, "user") {
		t.Error("expected HasRole to return true for 'user'")
	}

	if HasRole(ctx, "superadmin") {
		t.Error("expected HasRole to return false for 'superadmin'")
	}
}

func TestHasRole_NoClaims(t *testing.T) {
	ctx := context.Background()

	if HasRole(ctx, "admin") {
		t.Error("expected HasRole to return false when no claims in context")
	}
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID: "user123",
	}

	token, _ := jwt.CreateToken(claims)

	// Create a handler to test
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := GetUserID(ctx)

		if userID != "user123" {
			t.Errorf("expected UserID 'user123', got '%s'", userID)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_MissingToken(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["code"] != "MISSING_TOKEN" {
		t.Errorf("expected code 'MISSING_TOKEN', got '%v'", response["code"])
	}
}

func TestJWTAuthMiddleware_InvalidScheme(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestRequireRole_WithValidRole(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID: "user123",
		Roles:  []string{"admin"},
	}

	token, _ := jwt.CreateToken(claims)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Chain auth middleware with role requirement
	authMiddleware := JWTAuthMiddleware("test-secret")
	roleMiddleware := RequireRole("admin")
	handler := authMiddleware(roleMiddleware(testHandler))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequireRole_WithoutValidRole(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID: "user123",
		Roles:  []string{"user"},
	}

	token, _ := jwt.CreateToken(claims)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := JWTAuthMiddleware("test-secret")
	roleMiddleware := RequireRole("admin")
	handler := authMiddleware(roleMiddleware(testHandler))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID: "user123",
		Roles:  []string{"moderator"},
	}

	token, _ := jwt.CreateToken(claims)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authMiddleware := JWTAuthMiddleware("test-secret")
	roleMiddleware := RequireRole("admin", "moderator")
	handler := authMiddleware(roleMiddleware(testHandler))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthError_Error(t *testing.T) {
	err := &AuthError{
		Code:    "TEST_CODE",
		Message: "Test message",
	}

	if err.Error() != "Test message" {
		t.Errorf("expected 'Test message', got '%s'", err.Error())
	}
}

// --- Extended JWT / Auth Tests ---

func TestVerifyToken_DifferentSecret(t *testing.T) {
	jwtCreator := NewSimpleJWT("secret-A")
	jwtVerifier := NewSimpleJWT("secret-B")

	claims := &Claims{
		UserID:   "user123",
		TenantID: "tenant456",
	}

	token, err := jwtCreator.CreateToken(claims)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = jwtVerifier.VerifyToken(token)
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature when verifying with wrong secret, got %v", err)
	}
}

func TestVerifyToken_ExpiredToken_Manual(t *testing.T) {
	// Manually construct a token with past expiration to bypass CreateToken's auto-set
	jwt := NewSimpleJWT("test-secret")

	claims := &Claims{
		UserID:    "user123",
		TenantID:  "tenant456",
		Email:     "user@example.com",
		Roles:     []string{"admin"},
		IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
		Issuer:    "rentflow",
	}

	// Manually build a token with expired claims (replicate CreateToken logic without overwriting times)
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64EncodeRawURL(headerJSON)

	payloadJSON, _ := json.Marshal(claims)
	payloadB64 := base64EncodeRawURL(payloadJSON)

	message := headerB64 + "." + payloadB64
	signature := hmacSHA256([]byte(jwt.secret), []byte(message))
	signatureB64 := base64EncodeRawURL(signature)

	expiredToken := message + "." + signatureB64

	_, err := jwt.VerifyToken(expiredToken)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestVerifyToken_TenantIDExtraction(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID:   "user-abc",
		TenantID: "f6f63dc5-d253-58a7-8e1d-ab57e309120e",
		Email:    "admin@rentflow.de",
		Roles:    []string{"admin"},
	}

	token, err := jwt.CreateToken(claims)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	verified, err := jwt.VerifyToken(token)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}

	if verified.TenantID != "f6f63dc5-d253-58a7-8e1d-ab57e309120e" {
		t.Errorf("expected TenantID 'f6f63dc5-d253-58a7-8e1d-ab57e309120e', got '%s'", verified.TenantID)
	}
}

func TestJWTAuthMiddleware_TenantIDInContext(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID:   "user-abc",
		TenantID: "tenant-xyz-123",
		Email:    "test@test.de",
		Roles:    []string{"user"},
	}

	token, _ := jwt.CreateToken(claims)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := GetTenantIDFromClaims(r.Context())
		if tenantID != "tenant-xyz-123" {
			t.Errorf("expected TenantID 'tenant-xyz-123' in context, got '%s'", tenantID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_EmptyBearerToken(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for empty bearer token, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_InvalidSignatureViaMiddleware(t *testing.T) {
	// Create a token with one secret, try to verify through middleware with another
	jwtOther := NewSimpleJWT("other-secret")
	claims := &Claims{
		UserID: "user123",
		Roles:  []string{"admin"},
	}
	token, _ := jwtOther.CreateToken(claims)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for invalid signature")
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("correct-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["code"] != "INVALID_SIGNATURE" {
		t.Errorf("expected code 'INVALID_SIGNATURE', got '%v'", response["code"])
	}
}

func TestJWTAuthMiddleware_ExpiredTokenViaMiddleware(t *testing.T) {
	secret := "test-secret"
	jwt := NewSimpleJWT(secret)

	// Build an expired token manually
	claims := &Claims{
		UserID:    "user123",
		TenantID:  "tenant456",
		IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64EncodeRawURL(headerJSON)
	payloadJSON, _ := json.Marshal(claims)
	payloadB64 := base64EncodeRawURL(payloadJSON)
	message := headerB64 + "." + payloadB64
	signature := hmacSHA256([]byte(jwt.secret), []byte(message))
	signatureB64 := base64EncodeRawURL(signature)
	expiredToken := message + "." + signatureB64

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for expired token")
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware(secret)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for expired token, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["code"] != "TOKEN_EXPIRED" {
		t.Errorf("expected code 'TOKEN_EXPIRED', got '%v'", response["code"])
	}
}

func TestJWTAuthMiddleware_AllClaimsInContext(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID:   "user-999",
		TenantID: "tenant-888",
		Email:    "all@claims.test",
		Roles:    []string{"admin", "manager"},
	}

	token, _ := jwt.CreateToken(claims)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if GetUserID(ctx) != "user-999" {
			t.Errorf("UserID mismatch: got '%s'", GetUserID(ctx))
		}
		if GetTenantIDFromClaims(ctx) != "tenant-888" {
			t.Errorf("TenantID mismatch: got '%s'", GetTenantIDFromClaims(ctx))
		}
		if GetEmail(ctx) != "all@claims.test" {
			t.Errorf("Email mismatch: got '%s'", GetEmail(ctx))
		}
		roles := GetRoles(ctx)
		if len(roles) != 2 || roles[0] != "admin" || roles[1] != "manager" {
			t.Errorf("Roles mismatch: got %v", roles)
		}
		if !HasRole(ctx, "admin") {
			t.Error("expected HasRole('admin') to be true")
		}
		if !HasRole(ctx, "manager") {
			t.Error("expected HasRole('manager') to be true")
		}
		if HasRole(ctx, "superadmin") {
			t.Error("expected HasRole('superadmin') to be false")
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestJWTAuthMiddleware_OnlyBearerScheme(t *testing.T) {
	schemes := []string{
		"Token abc123",
		"bearer abc123",
		"BEARER abc123",
		"Basic dXNlcjpwYXNz",
		"Digest username=test",
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for non-Bearer scheme")
	})

	middleware := JWTAuthMiddleware("test-secret")
	handler := middleware(testHandler)

	for _, scheme := range schemes {
		t.Run(scheme, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", scheme)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected 401 for scheme '%s', got %d", scheme, w.Code)
			}
		})
	}
}

func TestVerifyToken_EmptyString(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	_, err := jwt.VerifyToken("")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for empty token, got %v", err)
	}
}

func TestVerifyToken_PreservesAllFields(t *testing.T) {
	jwt := NewSimpleJWT("test-secret")
	claims := &Claims{
		UserID:   "uid-001",
		TenantID: "tid-002",
		Email:    "preserve@test.de",
		Roles:    []string{"role1", "role2", "role3"},
		Issuer:   "rentflow-auth",
	}

	token, _ := jwt.CreateToken(claims)
	verified, err := jwt.VerifyToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if verified.UserID != claims.UserID {
		t.Errorf("UserID: got '%s', want '%s'", verified.UserID, claims.UserID)
	}
	if verified.TenantID != claims.TenantID {
		t.Errorf("TenantID: got '%s', want '%s'", verified.TenantID, claims.TenantID)
	}
	if verified.Email != claims.Email {
		t.Errorf("Email: got '%s', want '%s'", verified.Email, claims.Email)
	}
	if verified.Issuer != claims.Issuer {
		t.Errorf("Issuer: got '%s', want '%s'", verified.Issuer, claims.Issuer)
	}
	if len(verified.Roles) != 3 {
		t.Errorf("expected 3 roles, got %d", len(verified.Roles))
	}
}

func TestWriteAuthError_WithAuthError(t *testing.T) {
	w := httptest.NewRecorder()
	writeAuthError(w, http.StatusUnauthorized, &AuthError{Code: "CUSTOM", Message: "custom msg"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["code"] != "CUSTOM" {
		t.Errorf("expected code 'CUSTOM', got '%v'", response["code"])
	}
}

func TestWriteAuthError_WithGenericError(t *testing.T) {
	w := httptest.NewRecorder()
	genericErr := fmt.Errorf("something went wrong")
	writeAuthError(w, http.StatusInternalServerError, genericErr)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["code"] != "UNKNOWN_ERROR" {
		t.Errorf("expected code 'UNKNOWN_ERROR' for generic error, got '%v'", response["code"])
	}
	if response["message"] != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%v'", response["message"])
	}
}

// Helper functions for manually constructing tokens in tests

func base64EncodeRawURL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func hmacSHA256(key, message []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return h.Sum(nil)
}
