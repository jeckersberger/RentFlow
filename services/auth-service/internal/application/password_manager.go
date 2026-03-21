package application

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordManager handles password hashing and validation
type PasswordManager struct {
	minLength int
	// Argon2id parameters
	argon2Time    uint32
	argon2Memory  uint32
	argon2Threads uint8
	argon2KeyLen  uint32
	argon2SaltLen uint32
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		minLength:     12,
		argon2Time:    1,
		argon2Memory:  64 * 1024, // 64 MB
		argon2Threads: 4,
		argon2KeyLen:  32,
		argon2SaltLen: 16,
	}
}

// HashPassword hashes a password using Argon2id
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	if len(password) < pm.minLength {
		return "", fmt.Errorf("password must be at least %d characters", pm.minLength)
	}

	if err := pm.ValidatePassword(password); err != nil {
		return "", err
	}

	// Generate a random salt
	salt := make([]byte, pm.argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password with Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		pm.argon2Time,
		pm.argon2Memory,
		pm.argon2Threads,
		pm.argon2KeyLen,
	)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt_base64>$<hash_base64>
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		pm.argon2Memory,
		pm.argon2Time,
		pm.argon2Threads,
		saltB64,
		hashB64,
	), nil
}

// VerifyPassword verifies a password against a hash (supports both Argon2id and legacy SHA256)
func (pm *PasswordManager) VerifyPassword(password, hash string) bool {
	// Check if it's an Argon2id hash
	if strings.HasPrefix(hash, "$argon2id$") {
		return pm.verifyArgon2id(password, hash)
	}

	// Legacy SHA256 support
	return pm.verifySHA256(password, hash)
}

// verifyArgon2id verifies an Argon2id hash
func (pm *PasswordManager) verifyArgon2id(password, hash string) bool {
	// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt_base64>$<hash_base64>
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false
	}

	// Extract salt and hash
	saltB64 := parts[4]
	expectedHashB64 := parts[5]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false
	}

	// Hash the provided password
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		pm.argon2Time,
		pm.argon2Memory,
		pm.argon2Threads,
		pm.argon2KeyLen,
	)

	computedHashB64 := base64.RawStdEncoding.EncodeToString(computedHash)
	return computedHashB64 == expectedHashB64
}

// verifySHA256 verifies a legacy SHA256 hash (backward compatibility)
func (pm *PasswordManager) verifySHA256(password, hash string) bool {
	parts := strings.Split(hash, ":")
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
