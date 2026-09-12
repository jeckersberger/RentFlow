package application

import (
	"encoding/hex"
	"testing"
)

func TestGenerateSetupTokenReturnsRandom256BitToken(t *testing.T) {
	first, err := generateSetupToken()
	if err != nil {
		t.Fatalf("generateSetupToken: %v", err)
	}
	second, err := generateSetupToken()
	if err != nil {
		t.Fatalf("generateSetupToken second: %v", err)
	}
	if len(first) != 64 {
		t.Fatalf("token length = %d, want 64 hex characters", len(first))
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("token is not hex: %v", err)
	}
	if first == second {
		t.Fatal("two independently generated setup tokens are equal")
	}
}

func TestHashSetupTokenIsDeterministicAndDoesNotRevealToken(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	first := hashSetupToken(token)
	second := hashSetupToken(token)
	if first != second {
		t.Fatal("setup token hash is not deterministic")
	}
	if first == token {
		t.Fatal("setup token hash equals plaintext token")
	}
	if len(first) != 64 {
		t.Fatalf("hash length = %d, want 64", len(first))
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("hash is not hex: %v", err)
	}
	if first == hashSetupToken(token+"x") {
		t.Fatal("different setup tokens produced the same hash")
	}
}
