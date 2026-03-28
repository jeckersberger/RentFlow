package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// loadRSAKeys loads RSA keys from PEM files. If the files do not exist,
// it generates a dev key pair, saves them to /tmp, and updates the env vars
// so that subsequent config.Load() picks up the correct paths.
func loadRSAKeys(privatePath, publicPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privData, privErr := os.ReadFile(privatePath)
	pubData, pubErr := os.ReadFile(publicPath)

	if privErr == nil && pubErr == nil {
		privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privData)
		if err != nil {
			return nil, nil, err
		}
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubData)
		if err != nil {
			return nil, nil, err
		}
		return privKey, pubKey, nil
	}

	// Keys not found — generate for development.
	return generateDevKeys()
}

// generateDevKeys creates a 2048-bit RSA key pair and saves it to /tmp.
func generateDevKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Encode private key.
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})

	// Encode public key.
	pubDER, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

	// Save to /tmp for the JWTAuth middleware to read.
	privPath := "/tmp/cratedesk_private.pem"
	pubPath := "/tmp/cratedesk_public.pem"

	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		return nil, nil, err
	}
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return nil, nil, err
	}

	// Update env so config.Load and middleware.JWTAuth use the generated keys.
	os.Setenv("JWT_PRIVATE_KEY_PATH", privPath)
	os.Setenv("JWT_PUBLIC_KEY_PATH", pubPath)

	os.Stderr.WriteString("[WARN] RSA keys not found — generated dev keys at " + privPath + " and " + pubPath + "\n")

	return privKey, &privKey.PublicKey, nil
}
