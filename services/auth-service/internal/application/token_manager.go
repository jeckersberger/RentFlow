package application

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

const (
	accessTokenUse  = "access"
	refreshTokenUse = "refresh"
	jwtClockSkew    = 30 * time.Second
)

// TokenManager handles JWT token creation and validation using RS256
type TokenManager struct {
	privateKey         *rsa.PrivateKey
	publicKey          *rsa.PublicKey
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
	audience           []string
}

// NewTokenManager creates a new token manager with RSA key pair
func NewTokenManager(privateKeyPEM, publicKeyPEM string) (*TokenManager, error) {
	var privKey *rsa.PrivateKey
	var pubKey *rsa.PublicKey
	var err error

	if privateKeyPEM != "" {
		block, _ := pem.Decode([]byte(privateKeyPEM))
		if block == nil {
			return nil, fmt.Errorf("failed to parse private key PEM")
		}
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	if publicKeyPEM != "" {
		block, _ := pem.Decode([]byte(publicKeyPEM))
		if block == nil {
			return nil, fmt.Errorf("failed to parse public key PEM")
		}
		pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		pubKey, _ = pubKeyInterface.(*rsa.PublicKey)
		if pubKey == nil {
			return nil, fmt.Errorf("public key is not RSA")
		}
	}

	return &TokenManager{
		privateKey:         privKey,
		publicKey:          pubKey,
		accessTokenExpiry:  time.Hour,
		refreshTokenExpiry: 7 * 24 * time.Hour,
		issuer:             "rentflow-auth-service",
		audience:           []string{"rentflow-api"},
	}, nil
}

// GenerateKeyPair generates a new RSA 2048-bit key pair
func GenerateKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// MarshalPrivateKeyPEM serializes an RSA private key to PEM string
func MarshalPrivateKeyPEM(key *rsa.PrivateKey) string {
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}))
}

// MarshalPublicKeyPEM serializes an RSA public key to PEM string
func MarshalPublicKeyPEM(key *rsa.PublicKey) string {
	pubBytes, _ := x509.MarshalPKIXPublicKey(key)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}))
}

// PrivateKeyPEM returns the private key as PEM string
func (tm *TokenManager) PrivateKeyPEM() string {
	if tm.privateKey == nil {
		return ""
	}
	return MarshalPrivateKeyPEM(tm.privateKey)
}

// PublicKeyPEM returns the public key as PEM string
func (tm *TokenManager) PublicKeyPEM() string {
	if tm.publicKey == nil {
		return ""
	}
	return MarshalPublicKeyPEM(tm.publicKey)
}

// AccessTokenClaims represents claims in an access token
type AccessTokenClaims struct {
	Subject       string   `json:"sub"`
	Issuer        string   `json:"iss"`
	Audience      []string `json:"aud"`
	ExpiresAt     int64    `json:"exp"`
	IssuedAt      int64    `json:"iat"`
	NotBefore     int64    `json:"nbf"`
	JWTID         string   `json:"jti"`
	TokenUse      string   `json:"token_use"`
	TenantID      string   `json:"tenant_id"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles"`
	IPAddress     string   `json:"ip_address"`
	UserAgentHash string   `json:"user_agent_hash"`
	SessionID     string   `json:"session_id"`
}

// RefreshTokenClaims represents claims in a refresh token
type RefreshTokenClaims struct {
	Subject   string   `json:"sub"`
	Issuer    string   `json:"iss"`
	Audience  []string `json:"aud"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	JWTID     string   `json:"jti"`
	TokenUse  string   `json:"token_use"`
	TenantID  string   `json:"tenant_id"`
	Email     string   `json:"email"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid,omitempty"`
}

// CreateAccessToken creates a new access token
func (tm *TokenManager) CreateAccessToken(userID, tenantID, email, name string, roles []string, jti, ipAddress, userAgentHash, sessionID string) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		Subject:       userID,
		Issuer:        tm.issuer,
		Audience:      tm.audience,
		ExpiresAt:     now.Add(tm.accessTokenExpiry).Unix(),
		IssuedAt:      now.Unix(),
		NotBefore:     now.Unix(),
		JWTID:         jti,
		TokenUse:      accessTokenUse,
		TenantID:      tenantID,
		Email:         email,
		EmailVerified: true,
		Name:          name,
		Roles:         roles,
		IPAddress:     ipAddress,
		UserAgentHash: userAgentHash,
		SessionID:     sessionID,
	}
	return tm.createToken(claims)
}

// CreateRefreshToken creates a new refresh token
func (tm *TokenManager) CreateRefreshToken(userID, tenantID, email string, jti string) (string, error) {
	now := time.Now()
	claims := RefreshTokenClaims{
		Subject:   userID,
		Issuer:    tm.issuer,
		Audience:  tm.audience,
		ExpiresAt: now.Add(tm.refreshTokenExpiry).Unix(),
		IssuedAt:  now.Unix(),
		JWTID:     jti,
		TokenUse:  refreshTokenUse,
		TenantID:  tenantID,
		Email:     email,
	}
	return tm.createToken(claims)
}

// createToken creates a JWT token with RS256 signature.
func (tm *TokenManager) createToken(claims interface{}) (string, error) {
	if tm.privateKey == nil {
		return "", fmt.Errorf("private key not available")
	}

	header := jwtHeader{Algorithm: "RS256", Type: "JWT"}
	if jwk, err := tm.JWK(); err == nil {
		header.KeyID = jwk.KeyID
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, tm.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return message + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

// VerifyAccessToken verifies and returns the claims from an access token.
func (tm *TokenManager) VerifyAccessToken(token string) (*AccessTokenClaims, error) {
	payloadJSON, err := tm.verifySignedToken(token)
	if err != nil {
		return nil, err
	}

	var claims AccessTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}
	// Reject the wrong token purpose before evaluating claims that only exist on
	// access tokens (such as nbf/session_id). This prevents token-type confusion
	// and yields deterministic fail-closed behavior.
	if claims.TokenUse != accessTokenUse {
		return nil, fmt.Errorf("invalid token use")
	}
	if err := tm.validateRegisteredClaims(claims.Issuer, claims.Audience, claims.ExpiresAt, claims.IssuedAt); err != nil {
		return nil, err
	}
	if claims.NotBefore == 0 || time.Now().Add(jwtClockSkew).Unix() < claims.NotBefore {
		return nil, fmt.Errorf("token not yet valid")
	}
	if claims.Subject == "" || claims.JWTID == "" || claims.TenantID == "" || claims.SessionID == "" {
		return nil, fmt.Errorf("missing required access token claims")
	}
	return &claims, nil
}

// VerifyRefreshToken verifies and returns the claims from a refresh token.
func (tm *TokenManager) VerifyRefreshToken(token string) (*RefreshTokenClaims, error) {
	payloadJSON, err := tm.verifySignedToken(token)
	if err != nil {
		return nil, err
	}

	var claims RefreshTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}
	if claims.TokenUse != refreshTokenUse {
		return nil, fmt.Errorf("invalid token use")
	}
	if err := tm.validateRegisteredClaims(claims.Issuer, claims.Audience, claims.ExpiresAt, claims.IssuedAt); err != nil {
		return nil, err
	}
	if claims.Subject == "" || claims.JWTID == "" || claims.TenantID == "" {
		return nil, fmt.Errorf("missing required refresh token claims")
	}
	return &claims, nil
}

func (tm *TokenManager) verifySignedToken(token string) ([]byte, error) {
	if tm.publicKey == nil {
		return nil, fmt.Errorf("public key not available")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, fmt.Errorf("invalid token format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}
	if header.Algorithm != "RS256" || header.Type != "JWT" {
		return nil, fmt.Errorf("invalid token header")
	}
	if expected, err := tm.JWK(); err == nil && header.KeyID != "" && header.KeyID != expected.KeyID {
		return nil, fmt.Errorf("invalid key id")
	}

	message := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(message))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}
	if err := rsa.VerifyPKCS1v15(tm.publicKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, fmt.Errorf("invalid signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}
	return payloadJSON, nil
}

func (tm *TokenManager) validateRegisteredClaims(issuer string, audience []string, expiresAt, issuedAt int64) error {
	now := time.Now()
	if expiresAt == 0 || now.Unix() >= expiresAt {
		return fmt.Errorf("token expired")
	}
	if issuedAt == 0 || issuedAt > now.Add(jwtClockSkew).Unix() {
		return fmt.Errorf("invalid issued-at time")
	}
	if issuer != tm.issuer {
		return fmt.Errorf("invalid issuer")
	}
	if !audienceContains(audience, tm.audience) {
		return fmt.Errorf("invalid audience")
	}
	return nil
}

func audienceContains(actual, expected []string) bool {
	for _, want := range expected {
		found := false
		for _, got := range actual {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return len(expected) > 0
}

// ExtractUserIDFromToken extracts the user ID from a token
func (tm *TokenManager) ExtractUserIDFromToken(token string) (string, error) {
	claims, err := tm.VerifyAccessToken(token)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// GetAccessTokenExpiry returns the access token expiry duration
func (tm *TokenManager) GetAccessTokenExpiry() time.Duration {
	return tm.accessTokenExpiry
}
