package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TokenManager handles JWT token creation and validation
type TokenManager struct {
	secret              string
	accessTokenExpiry   time.Duration
	refreshTokenExpiry  time.Duration
	issuer              string
	audience            []string
}

// NewTokenManager creates a new token manager
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secret:             secret,
		accessTokenExpiry:  15 * time.Minute,
		refreshTokenExpiry: 7 * 24 * time.Hour,
		issuer:             "rentflow-auth-service",
		audience:           []string{"rentflow-api"},
	}
}

// AccessTokenClaims represents claims in an access token
type AccessTokenClaims struct {
	Subject       string   `json:"sub"`     // user_id
	Issuer        string   `json:"iss"`     // issuer
	Audience      []string `json:"aud"`     // audience
	ExpiresAt     int64    `json:"exp"`     // expiration time
	IssuedAt      int64    `json:"iat"`     // issued at
	NotBefore     int64    `json:"nbf"`     // not before
	JWTID         string   `json:"jti"`     // JWT ID
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
	Subject   string `json:"sub"`     // user_id
	Issuer    string `json:"iss"`     // issuer
	Audience  []string `json:"aud"`   // audience
	ExpiresAt int64  `json:"exp"`     // expiration time
	IssuedAt  int64  `json:"iat"`     // issued at
	JWTID     string `json:"jti"`     // JWT ID
	TenantID  string `json:"tenant_id"`
	Email     string `json:"email"`
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

// createToken creates a JWT token with the given claims
func (tm *TokenManager) createToken(claims interface{}) (string, error) {
	// Create header
	header := map[string]interface{}{
		"alg": "HS256",
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
	h := hmac.New(sha256.New, []byte(tm.secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature, nil
}

// VerifyAccessToken verifies and returns the claims from an access token
func (tm *TokenManager) VerifyAccessToken(token string) (*AccessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + payloadB64
	h := hmac.New(sha256.New, []byte(tm.secret))
	h.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if signatureB64 != expectedSignature {
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
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + payloadB64
	h := hmac.New(sha256.New, []byte(tm.secret))
	h.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if signatureB64 != expectedSignature {
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
