package application

import (
	"encoding/base64"
	"math/big"
	"testing"
)

func TestTokenManagerJWK(t *testing.T) {
	key, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	manager, err := NewTokenManager(MarshalPrivateKeyPEM(key), MarshalPublicKeyPEM(&key.PublicKey))
	if err != nil {
		t.Fatalf("NewTokenManager: %v", err)
	}

	jwk, err := manager.JWK()
	if err != nil {
		t.Fatalf("JWK: %v", err)
	}
	if jwk.KeyType != "RSA" || jwk.Use != "sig" || jwk.Algorithm != "RS256" {
		t.Fatalf("unexpected JWK metadata: %+v", jwk)
	}
	if jwk.KeyID == "" {
		t.Fatal("kid is empty")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.Modulus)
	if err != nil {
		t.Fatalf("decode n: %v", err)
	}
	if new(big.Int).SetBytes(nBytes).Cmp(key.PublicKey.N) != 0 {
		t.Fatal("JWK modulus does not match RSA public key")
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.Exponent)
	if err != nil {
		t.Fatalf("decode e: %v", err)
	}
	if int(new(big.Int).SetBytes(eBytes).Int64()) != key.PublicKey.E {
		t.Fatal("JWK exponent does not match RSA public key")
	}

	second, err := manager.JWK()
	if err != nil {
		t.Fatalf("second JWK: %v", err)
	}
	if second.KeyID != jwk.KeyID {
		t.Fatal("kid is not stable for the same public key")
	}
}

func TestTokenManagerJWKRequiresPublicKey(t *testing.T) {
	manager, err := NewTokenManager("", "")
	if err != nil {
		t.Fatalf("NewTokenManager: %v", err)
	}
	if _, err := manager.JWK(); err == nil {
		t.Fatal("expected JWK to fail without public key")
	}
}
