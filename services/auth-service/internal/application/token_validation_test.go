package application

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAccessTokenCannotBeUsedAsRefreshToken(t *testing.T) {
	tm := newTestTokenManager(t)
	access, err := tm.CreateAccessToken("user", "tenant", "u@example.com", "User", []string{"admin"}, "access-jti", "", "", "session")
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	if _, err := tm.VerifyRefreshToken(access); err == nil || !strings.Contains(err.Error(), "token use") {
		t.Fatalf("VerifyRefreshToken(access) error = %v, want invalid token use", err)
	}
}

func TestRefreshTokenCannotBeUsedAsAccessToken(t *testing.T) {
	tm := newTestTokenManager(t)
	refresh, err := tm.CreateRefreshToken("user", "tenant", "u@example.com", "refresh-jti")
	if err != nil {
		t.Fatalf("CreateRefreshToken: %v", err)
	}

	if _, err := tm.VerifyAccessToken(refresh); err == nil || !strings.Contains(err.Error(), "token use") {
		t.Fatalf("VerifyAccessToken(refresh) error = %v, want invalid token use", err)
	}
}

func TestVerifyAccessTokenRejectsWrongAudience(t *testing.T) {
	tm := newTestTokenManager(t)
	now := time.Now()
	token, err := tm.createToken(AccessTokenClaims{
		Subject:   "user",
		Issuer:    tm.issuer,
		Audience:  []string{"other-api"},
		ExpiresAt: now.Add(time.Hour).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		JWTID:     "jti",
		TokenUse:  accessTokenUse,
		TenantID:  "tenant",
		SessionID: "session",
	})
	if err != nil {
		t.Fatalf("createToken: %v", err)
	}

	if _, err := tm.VerifyAccessToken(token); err == nil || !strings.Contains(err.Error(), "audience") {
		t.Fatalf("VerifyAccessToken error = %v, want invalid audience", err)
	}
}

func TestVerifyRefreshTokenRejectsWrongIssuer(t *testing.T) {
	tm := newTestTokenManager(t)
	now := time.Now()
	token, err := tm.createToken(RefreshTokenClaims{
		Subject:   "user",
		Issuer:    "other-issuer",
		Audience:  tm.audience,
		ExpiresAt: now.Add(time.Hour).Unix(),
		IssuedAt:  now.Unix(),
		JWTID:     "jti",
		TokenUse:  refreshTokenUse,
		TenantID:  "tenant",
	})
	if err != nil {
		t.Fatalf("createToken: %v", err)
	}

	if _, err := tm.VerifyRefreshToken(token); err == nil || !strings.Contains(err.Error(), "issuer") {
		t.Fatalf("VerifyRefreshToken error = %v, want invalid issuer", err)
	}
}

func TestVerifyAccessTokenRejectsFutureNotBefore(t *testing.T) {
	tm := newTestTokenManager(t)
	now := time.Now()
	token, err := tm.createToken(AccessTokenClaims{
		Subject:   "user",
		Issuer:    tm.issuer,
		Audience:  tm.audience,
		ExpiresAt: now.Add(time.Hour).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Add(5 * time.Minute).Unix(),
		JWTID:     "jti",
		TokenUse:  accessTokenUse,
		TenantID:  "tenant",
		SessionID: "session",
	})
	if err != nil {
		t.Fatalf("createToken: %v", err)
	}

	if _, err := tm.VerifyAccessToken(token); err == nil || !strings.Contains(err.Error(), "not yet valid") {
		t.Fatalf("VerifyAccessToken error = %v, want not yet valid", err)
	}
}

func TestVerifyTokenRejectsSignedUnexpectedAlgorithmHeader(t *testing.T) {
	tm := newTestTokenManager(t)
	now := time.Now()
	claims := AccessTokenClaims{
		Subject:   "user",
		Issuer:    tm.issuer,
		Audience:  tm.audience,
		ExpiresAt: now.Add(time.Hour).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		JWTID:     "jti",
		TokenUse:  accessTokenUse,
		TenantID:  "tenant",
		SessionID: "session",
	}
	token := signTestJWT(t, tm.privateKey, map[string]string{"alg": "RS512", "typ": "JWT"}, claims)

	if _, err := tm.VerifyAccessToken(token); err == nil || !strings.Contains(err.Error(), "header") {
		t.Fatalf("VerifyAccessToken error = %v, want invalid token header", err)
	}
}

func TestCreatedTokenIncludesJWKKeyID(t *testing.T) {
	tm := newTestTokenManager(t)
	token, err := tm.CreateAccessToken("user", "tenant", "u@example.com", "User", nil, "jti", "", "", "session")
	if err != nil {
		t.Fatalf("CreateAccessToken: %v", err)
	}

	parts := strings.Split(token, ".")
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}
	jwk, err := tm.JWK()
	if err != nil {
		t.Fatalf("JWK: %v", err)
	}
	if header.KeyID == "" || header.KeyID != jwk.KeyID {
		t.Fatalf("kid = %q, want %q", header.KeyID, jwk.KeyID)
	}
}

func signTestJWT(t *testing.T, key *rsa.PrivateKey, header, claims interface{}) string {
	t.Helper()
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	message := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return message + "." + base64.RawURLEncoding.EncodeToString(signature)
}
