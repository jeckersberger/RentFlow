package application

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
)

// PublicJWK is the public JSON Web Key representation used by the JWKS endpoint.
type PublicJWK struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// JWK returns the active RSA public key in RFC 7517 compatible form.
func (tm *TokenManager) JWK() (PublicJWK, error) {
	if tm.publicKey == nil {
		return PublicJWK{}, fmt.Errorf("public key not available")
	}

	n := base64.RawURLEncoding.EncodeToString(tm.publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(tm.publicKey.E)).Bytes())
	thumbprintInput := fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`, e, n)
	thumbprint := sha256.Sum256([]byte(thumbprintInput))
	kid := base64.RawURLEncoding.EncodeToString(thumbprint[:])

	return PublicJWK{
		KeyType:   "RSA",
		Use:       "sig",
		Algorithm: "RS256",
		KeyID:     kid,
		Modulus:   n,
		Exponent:  e,
	}, nil
}
