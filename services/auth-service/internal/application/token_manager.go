package application

import (
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

	// Parse private key
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

	// Parse public key
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

	tm := &TokenManager{
		privateKey:         privKey,
		publicKey:          pubKey,
		accessTokenExpiry:  1 * time.Hour,
		refreshTokenExpiry: 7 * 24 * time.Hour,
		issuer:             "rentflow-auth-service",
		audience:           []string{"rentflow-api"},
	}

	return tm, nil
}

// GenerateKeyPair generates a new RSA 2048-bit key pair
func GenerateKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// MarshalPrivateKeyPEM serializes an RSA private key to PEM string
func MarshalPrivateKeyPEM(key *rsa.PrivateKey) string {
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	return string(privPEM)
}

// MarshalPublicKeyPEM serializes an RSA public key to PEM string
func MarshalPublicKeyPEM(key *rsa.PublicKey) string {
	pubBytes, _ := x509.MarshalPKIXPublicKey(key)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	return string(pubPEM)
}

// PrivateKeyPEM returns the private key as PEM string
func (tm *TokenManager) PrivateKeyPEM() string {
	if tm.privateKey == nil {
		return ""
	}
	privBytes := x509.MarshalPKCS1PrivateKey(tm.privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	return string(privPEM)
}

// PublicKeyPEM returns the public key as PEM string
func (tm *TokenManager) PublicKeyPEM() string {
	if tm.publicKey == nil {
		return ""
	}
	pubBytes, _ := x509.MarshalPKIXPublicKey(tm.publicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	return string(pubPEM)
}

// AccessTokenClaims represents claims in an access token
type AccessTokenClaims struct {
	Subject       string   `json:"sub"` // user_id
	Issuer        string   `json:"iss"` // issuer
	Audience      []string `json:"aud"` // audience
	ExpiresAt     int64    `json:"exp"` // expiration time
	IssuedAt      int64    `json:"iat"` // issued at
	NotBefore     int64    `json:"nbf"` // not before
	JWTID         string   `json:"jti"` // JWT ID
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
	Subject   string   `json:"sub"` // user_id
	Issuer    string   `json:"iss"` // issuer
	Audience  []string `json:"aud"` // audience
	ExpiresAt int64    `json:"exp"` // expiration time
	IssuedAt  int64    `json:"iat"` // issued at
	JWTID     string   `json:"jti"` // JWT ID
	TenantID  string   `json:"tenant_id"`
	Email     string   `json:"email"`
}

// CreateAccessToken creates a new access token
func (tm *TokenManager) CreateAccessToken(userID, tenantID, email, name string, roles []string, jti, ipAddress, userAgentHash, sessionID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(tm.accessTokenExpiry)

	claims := AccessTokenClaims{
		Subject:       userID,
		Issuer:        tm.issuer,
		Audience:      tm.audience,
		ExpiresAt:     expiresAt.Unix(),
		IssuedAt:      now.Unix(),
		NotBefore:     now.Unix(),
		JWTID:         jti,
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
	expiresAt := now.Add(tm.refreshTokenExpiry)

	claims := RefreshTokenClaims{
		Subject:   userID,
		Issuer:    tm.issuer,
		Audience:  tm.audience,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  now.Unix(),
		JWTID:     jti,
		TenantID:  tenantID,
		Email:     email,
	}

	return tm.createToken(claims)
}

// createToken creates a JWT token with RS256 signature
func (tm *TokenManager) createToken(claims interface{}) (string, error) {
	if tm.privateKey == nil {
		return "", fmt.Errorf("private key not available")
	}

	// Create header
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Create payload
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, tm.privateKey, sha256.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)
	return message + "." + signatureB64, nil
}

// VerifyAccessToken verifies and returns the claims from an access token
func (tm *TokenManager) VerifyAccessToken(token string) (*AccessTokenClaims, error) {
	if tm.publicKey == nil {
		return nil, fmt.Errorf("public key not available")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(message))

	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	err = rsa.VerifyPKCS1v15(tm.publicKey, sha256.SHA256, hash[:], signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims AccessTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	// Check issuer
	if claims.Issuer != tm.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	return &claims, nil
}

// VerifyRefreshToken verifies and returns the claims from a refresh token
func (tm *TokenManager) VerifyRefreshToken(token string) (*RefreshTokenClaims, error) {
	if tm.publicKey == nil {
		return nil, fmt.Errorf("public key not available")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(message))

	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	err = rsa.VerifyPKCS1v15(tm.publicKey, sha256.SHA256, hash[:], signature)
	if err != nil {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims RefreshTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
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
