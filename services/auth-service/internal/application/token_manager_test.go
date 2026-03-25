package application

import (
	"crypto/rsa"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helper: generate a TokenManager with fresh RSA keys for tests
// ---------------------------------------------------------------------------

func newTestTokenManager(t *testing.T) *TokenManager {
	t.Helper()
	privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	privPEM := MarshalPrivateKeyPEM(privKey)
	pubPEM := MarshalPublicKeyPEM(&privKey.PublicKey)

	tm, err := NewTokenManager(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}
	return tm
}

// newExpiredTokenManager creates a TokenManager whose tokens expire immediately.
func newExpiredTokenManager(t *testing.T) *TokenManager {
	t.Helper()
	tm := newTestTokenManager(t)
	// Set expiry to -1 second so tokens are already expired when created
	tm.accessTokenExpiry = -1 * time.Second
	tm.refreshTokenExpiry = -1 * time.Second
	return tm
}

// ---------------------------------------------------------------------------
// Access Token Tests
// ---------------------------------------------------------------------------

func TestCreateAndVerifyAccessToken(t *testing.T) {
	tm := newTestTokenManager(t)

	token, err := tm.CreateAccessToken(
		"user-123", "tenant-456", "test@example.com", "Test User",
		[]string{"admin"}, "jti-789", "127.0.0.1", "ua-hash", "session-abc",
	)
	if err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	// Token should have 3 parts (header.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}

	// Verify the token
	claims, err := tm.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("failed to verify access token: %v", err)
	}

	if claims.Subject != "user-123" {
		t.Errorf("expected subject 'user-123', got '%s'", claims.Subject)
	}
	if claims.TenantID != "tenant-456" {
		t.Errorf("expected tenant_id 'tenant-456', got '%s'", claims.TenantID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", claims.Name)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Errorf("expected roles [admin], got %v", claims.Roles)
	}
	if claims.JWTID != "jti-789" {
		t.Errorf("expected jti 'jti-789', got '%s'", claims.JWTID)
	}
	if claims.SessionID != "session-abc" {
		t.Errorf("expected session_id 'session-abc', got '%s'", claims.SessionID)
	}
	if claims.Issuer != "rentflow-auth-service" {
		t.Errorf("expected issuer 'rentflow-auth-service', got '%s'", claims.Issuer)
	}
}

func TestVerifyAccessToken_Expired(t *testing.T) {
	tm := newExpiredTokenManager(t)

	token, err := tm.CreateAccessToken(
		"user-123", "tenant-456", "test@example.com", "Test User",
		[]string{"admin"}, "jti-789", "", "", "session-abc",
	)
	if err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	_, err = tm.VerifyAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected 'expired' in error message, got '%s'", err.Error())
	}
}

func TestVerifyAccessToken_InvalidSignature(t *testing.T) {
	tm := newTestTokenManager(t)

	token, err := tm.CreateAccessToken(
		"user-123", "tenant-456", "test@example.com", "Test User",
		[]string{"admin"}, "jti-789", "", "", "session-abc",
	)
	if err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	// Tamper with the token by modifying the payload
	parts := strings.Split(token, ".")
	parts[1] = parts[1] + "tampered"
	tamperedToken := strings.Join(parts, ".")

	_, err = tm.VerifyAccessToken(tamperedToken)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestVerifyAccessToken_InvalidFormat(t *testing.T) {
	tm := newTestTokenManager(t)
	_, err := tm.VerifyAccessToken("not.a.valid.token.format")
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestVerifyAccessToken_WrongKey(t *testing.T) {
	tm1 := newTestTokenManager(t)
	tm2 := newTestTokenManager(t) // different key pair

	token, _ := tm1.CreateAccessToken(
		"user-123", "tenant-456", "test@example.com", "Test User",
		[]string{"admin"}, "jti-789", "", "", "session-abc",
	)

	_, err := tm2.VerifyAccessToken(token)
	if err == nil {
		t.Fatal("expected error when verifying with wrong public key")
	}
}

// ---------------------------------------------------------------------------
// Refresh Token Tests
// ---------------------------------------------------------------------------

func TestCreateAndVerifyRefreshToken(t *testing.T) {
	tm := newTestTokenManager(t)

	token, err := tm.CreateRefreshToken("user-123", "tenant-456", "test@example.com", "jti-789")
	if err != nil {
		t.Fatalf("failed to create refresh token: %v", err)
	}

	claims, err := tm.VerifyRefreshToken(token)
	if err != nil {
		t.Fatalf("failed to verify refresh token: %v", err)
	}

	if claims.Subject != "user-123" {
		t.Errorf("expected subject 'user-123', got '%s'", claims.Subject)
	}
	if claims.TenantID != "tenant-456" {
		t.Errorf("expected tenant_id 'tenant-456', got '%s'", claims.TenantID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.JWTID != "jti-789" {
		t.Errorf("expected jti 'jti-789', got '%s'", claims.JWTID)
	}
}

func TestVerifyRefreshToken_Expired(t *testing.T) {
	tm := newExpiredTokenManager(t)

	token, err := tm.CreateRefreshToken("user-123", "tenant-456", "test@example.com", "jti-789")
	if err != nil {
		t.Fatalf("failed to create refresh token: %v", err)
	}

	_, err = tm.VerifyRefreshToken(token)
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
}

func TestVerifyRefreshToken_WrongKey(t *testing.T) {
	tm1 := newTestTokenManager(t)
	tm2 := newTestTokenManager(t)

	token, _ := tm1.CreateRefreshToken("user-123", "tenant-456", "test@example.com", "jti-789")

	_, err := tm2.VerifyRefreshToken(token)
	if err == nil {
		t.Fatal("expected error when verifying with wrong public key")
	}
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestCreateToken_NoPrivateKey(t *testing.T) {
	// Create a TokenManager with only a public key (no private key)
	privKey, _ := GenerateKeyPair()
	pubPEM := MarshalPublicKeyPEM(&privKey.PublicKey)

	tm, err := NewTokenManager("", pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = tm.CreateAccessToken("user-123", "tenant-456", "test@example.com", "Test User",
		[]string{"admin"}, "jti-789", "", "", "session-abc")
	if err == nil {
		t.Fatal("expected error when creating token without private key")
	}
}

func TestVerifyToken_NoPublicKey(t *testing.T) {
	privKey, _ := GenerateKeyPair()
	privPEM := MarshalPrivateKeyPEM(privKey)

	tm, err := NewTokenManager(privPEM, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, _ := tm.CreateAccessToken("user-123", "tenant-456", "test@example.com", "Test User",
		[]string{"admin"}, "jti-789", "", "", "session-abc")

	_, err = tm.VerifyAccessToken(token)
	if err == nil {
		t.Fatal("expected error when verifying without public key")
	}
}

func TestExtractUserIDFromToken(t *testing.T) {
	tm := newTestTokenManager(t)

	token, _ := tm.CreateAccessToken("user-abc", "tenant-456", "test@example.com", "Test",
		[]string{"admin"}, "jti", "", "", "session")

	userID, err := tm.ExtractUserIDFromToken(token)
	if err != nil {
		t.Fatalf("failed to extract user ID: %v", err)
	}
	if userID != "user-abc" {
		t.Errorf("expected 'user-abc', got '%s'", userID)
	}
}

func TestGetAccessTokenExpiry(t *testing.T) {
	tm := newTestTokenManager(t)
	expiry := tm.GetAccessTokenExpiry()
	if expiry != 1*time.Hour {
		t.Errorf("expected 1h expiry, got %v", expiry)
	}
}

func TestKeyPairMarshalRoundTrip(t *testing.T) {
	privKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	privPEM := MarshalPrivateKeyPEM(privKey)
	pubPEM := MarshalPublicKeyPEM(&privKey.PublicKey)

	// Create a new TokenManager from the PEM strings
	tm, err := NewTokenManager(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("failed to create token manager from marshaled keys: %v", err)
	}

	// Verify keys work end-to-end
	token, err := tm.CreateAccessToken("u1", "t1", "e@e.com", "N", []string{"r"}, "j", "", "", "s")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	claims, err := tm.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}
	if claims.Subject != "u1" {
		t.Errorf("expected subject 'u1', got '%s'", claims.Subject)
	}
}

func TestNewTokenManager_InvalidPrivateKeyPEM(t *testing.T) {
	_, err := NewTokenManager("not-a-pem", "")
	if err == nil {
		t.Fatal("expected error for invalid private key PEM")
	}
}

func TestNewTokenManager_InvalidPublicKeyPEM(t *testing.T) {
	_, err := NewTokenManager("", "not-a-pem")
	if err == nil {
		t.Fatal("expected error for invalid public key PEM")
	}
}

// Ensure the public key returned by PublicKeyPEM/PrivateKeyPEM works
func TestTokenManager_PEMAccessors(t *testing.T) {
	tm := newTestTokenManager(t)

	privPEM := tm.PrivateKeyPEM()
	pubPEM := tm.PublicKeyPEM()

	if !strings.Contains(privPEM, "RSA PRIVATE KEY") {
		t.Error("PrivateKeyPEM should contain RSA PRIVATE KEY")
	}
	if !strings.Contains(pubPEM, "PUBLIC KEY") {
		t.Error("PublicKeyPEM should contain PUBLIC KEY")
	}

	// No-key cases
	tmEmpty := &TokenManager{}
	if tmEmpty.PrivateKeyPEM() != "" {
		t.Error("expected empty string for nil private key")
	}
	if tmEmpty.PublicKeyPEM() != "" {
		t.Error("expected empty string for nil public key")
	}
}

func TestNewTokenManager_WrongPublicKeyType(t *testing.T) {
	// Provide a valid private key PEM as public key PEM - should fail
	privKey, _ := GenerateKeyPair()
	privPEM := MarshalPrivateKeyPEM(privKey)

	_, err := NewTokenManager("", privPEM)
	if err == nil {
		t.Fatal("expected error when providing private key as public key")
	}
}

// Verify that GenerateKeyPair produces valid keys
func TestGenerateKeyPair(t *testing.T) {
	key, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
	// Check it's a 2048-bit key
	if key.N.BitLen() != 2048 {
		t.Errorf("expected 2048-bit key, got %d", key.N.BitLen())
	}
	// Validate the key
	if err := key.Validate(); err != nil {
		t.Fatalf("generated key is invalid: %v", err)
	}
	// Public key should be extractable
	var _ *rsa.PublicKey = &key.PublicKey
}
