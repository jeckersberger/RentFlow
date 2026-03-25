package application

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// PasswordManager Tests
// ---------------------------------------------------------------------------

func TestHashPassword_ValidPassword(t *testing.T) {
	pm := NewPasswordManager()
	hash, err := pm.HashPassword("SecurePass123!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected argon2id hash prefix, got %s", hash)
	}
}

func TestHashPassword_TooShort(t *testing.T) {
	pm := NewPasswordManager()
	_, err := pm.HashPassword("Short1!")
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
}

func TestHashPassword_MissingUppercase(t *testing.T) {
	pm := NewPasswordManager()
	_, err := pm.HashPassword("alllowercase123!")
	if err == nil {
		t.Fatal("expected error for missing uppercase")
	}
}

func TestHashPassword_MissingLowercase(t *testing.T) {
	pm := NewPasswordManager()
	_, err := pm.HashPassword("ALLUPPERCASE123!")
	if err == nil {
		t.Fatal("expected error for missing lowercase")
	}
}

func TestHashPassword_MissingDigit(t *testing.T) {
	pm := NewPasswordManager()
	_, err := pm.HashPassword("NoDigitsHere!!!")
	if err == nil {
		t.Fatal("expected error for missing digit")
	}
}

func TestHashPassword_MissingSpecialChar(t *testing.T) {
	pm := NewPasswordManager()
	_, err := pm.HashPassword("NoSpecialChar123")
	if err == nil {
		t.Fatal("expected error for missing special character")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	pm := NewPasswordManager()
	_, err := pm.HashPassword("")
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "SecurePass123!"
	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	if !pm.VerifyPassword(password, hash) {
		t.Fatal("expected password to verify correctly")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	pm := NewPasswordManager()
	hash, err := pm.HashPassword("SecurePass123!")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	if pm.VerifyPassword("WrongPassword123!", hash) {
		t.Fatal("expected wrong password to be rejected")
	}
}

func TestVerifyPassword_InvalidHashFormat(t *testing.T) {
	pm := NewPasswordManager()
	if pm.VerifyPassword("anything", "not-a-valid-hash") {
		t.Fatal("expected invalid hash to return false")
	}
}

func TestVerifyPassword_DifferentHashesSamePassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "SecurePass123!"

	hash1, _ := pm.HashPassword(password)
	hash2, _ := pm.HashPassword(password)

	// Two hashes of the same password should be different (different salt)
	if hash1 == hash2 {
		t.Fatal("expected different hashes due to random salt")
	}

	// But both should verify
	if !pm.VerifyPassword(password, hash1) {
		t.Fatal("hash1 should verify")
	}
	if !pm.VerifyPassword(password, hash2) {
		t.Fatal("hash2 should verify")
	}
}

func TestValidatePassword_AllRequirementsMet(t *testing.T) {
	pm := NewPasswordManager()
	err := pm.ValidatePassword("GoodPassword1!")
	if err != nil {
		t.Fatalf("expected valid password, got error: %v", err)
	}
}

func TestValidatePassword_ExactMinLength(t *testing.T) {
	pm := NewPasswordManager()
	// Exactly 12 characters, meets all requirements
	err := pm.ValidatePassword("Abcdefghij1!")
	if err != nil {
		t.Fatalf("expected valid password at min length, got error: %v", err)
	}
}

func TestValidatePassword_OneBelowMinLength(t *testing.T) {
	pm := NewPasswordManager()
	// 11 characters
	err := pm.ValidatePassword("Abcdefghi1!")
	if err == nil {
		t.Fatal("expected error for 11-char password")
	}
}
