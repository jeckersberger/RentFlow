package ports

// UserRepository defines the interface for user data access
// This is a port that adapters will implement
type UserRepository interface {
	// Save persists a new user
	Save(id, email, username, password string) error

	// GetByEmail retrieves a user by email
	GetByEmail(email string) (interface{}, error)

	// Verify checks user credentials
	Verify(email, password string) (bool, error)
}
