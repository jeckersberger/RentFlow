package application

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
)

// PasswordManager handles password hashing and validation
type PasswordManager struct {
	saltLength int
	minLength  int
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		saltLength: 32,
		minLength:  12,
	}
}

// HashPassword hashes a password using SHA256 with a salt
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	if len(password) < pm.minLength {
		return "", fmt.Errorf("password must be at least %d characters", pm.minLength)
	}

	if err := pm.ValidatePassword(password); err != nil {
		return "", err
	}

	// Generate a random salt
	salt := make([]byte, pm.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password with the salt
	hash := sha256.Sum256(append(salt, []byte(password)...))

	// Return salt + hash as hex
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash[:]), nil
}

// VerifyPassword verifies a password against a hash
func (pm *PasswordManager) VerifyPassword(password, hash string) bool {
	parts := parseHash(hash)
	if len(parts) != 2 {
		return false
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}

	expectedHash := sha256.Sum256(append(salt, []byte(password)...))
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	return expectedHashStr == parts[1]
}

// ValidatePassword validates password requirements
func (pm *PasswordManager) ValidatePassword(password string) error {
	if len(password) < pm.minLength {
		return fmt.Errorf("password must be at least %d characters", pm.minLength)
	}

	// At least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// At least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// At least one digit
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	// At least one special character
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// parseHash parses a salted hash
func parseHash(hash string) []string {
	return regexp.MustCompile(`:+`).Split(hash, -1)
}
