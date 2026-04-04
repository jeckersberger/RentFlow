package application

import (
	"testing"
)

// ---------------------------------------------------------------------------
// HashPassword + VerifyPassword round-trip
// ---------------------------------------------------------------------------

func TestHashAndVerifyPassword(t *testing.T) {
	password := "SuperSecret123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned empty string")
	}

	// PHC format check: $argon2id$v=19$...
	if len(hash) < 20 {
		t.Fatalf("hash too short: %q", hash)
	}

	if !VerifyPassword(hash, password) {
		t.Error("VerifyPassword() should return true for correct password")
	}

	if VerifyPassword(hash, "WrongPassword!") {
		t.Error("VerifyPassword() should return false for wrong password")
	}
}

// ---------------------------------------------------------------------------
// HashPassword produces unique salts
// ---------------------------------------------------------------------------

func TestHashPasswordUniqueSalts(t *testing.T) {
	h1, err := HashPassword("test1234")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	h2, err := HashPassword("test1234")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if h1 == h2 {
		t.Error("two hashes of the same password should differ (unique salt)")
	}
}

// ---------------------------------------------------------------------------
// VerifyPassword edge cases
// ---------------------------------------------------------------------------

func TestVerifyPasswordEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{name: "empty hash", hash: "", password: "test", want: false},
		{name: "garbage hash", hash: "not-a-hash", password: "test", want: false},
		{name: "wrong format", hash: "$bcrypt$something", password: "test", want: false},
		{name: "partial argon2id", hash: "$argon2id$v=19$", password: "test", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyPassword(tt.hash, tt.password); got != tt.want {
				t.Errorf("VerifyPassword(%q, %q) = %v, want %v", tt.hash, tt.password, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// validatePassword
// ---------------------------------------------------------------------------

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "valid 8 chars", password: "12345678", wantErr: false},
		{name: "valid long", password: "SuperSecretPassword123!", wantErr: false},
		{name: "too short 7", password: "1234567", wantErr: true},
		{name: "empty", password: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePassword(%q) error = %v, wantErr = %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// hashToken determinism
// ---------------------------------------------------------------------------

func TestHashTokenDeterministic(t *testing.T) {
	token := "myRefreshToken123"
	h1 := hashToken(token)
	h2 := hashToken(token)

	if h1 != h2 {
		t.Errorf("hashToken should be deterministic: %q != %q", h1, h2)
	}

	if h1 == "" {
		t.Error("hashToken should not return empty string")
	}

	// Different input should produce different hash
	h3 := hashToken("differentToken")
	if h1 == h3 {
		t.Error("different tokens should produce different hashes")
	}
}

// ---------------------------------------------------------------------------
// generateRefreshToken uniqueness
// ---------------------------------------------------------------------------

func TestGenerateRefreshTokenUnique(t *testing.T) {
	t1, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken() error = %v", err)
	}

	t2, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("generateRefreshToken() error = %v", err)
	}

	if t1 == "" || t2 == "" {
		t.Error("refresh tokens should not be empty")
	}

	if t1 == t2 {
		t.Error("two refresh tokens should be different")
	}

	// Should be base64url encoded (no padding)
	if len(t1) < 40 {
		t.Errorf("refresh token too short: len=%d", len(t1))
	}
}
