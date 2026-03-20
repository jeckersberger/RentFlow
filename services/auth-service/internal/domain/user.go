package domain

// User represents a user in the auth domain
type User struct {
	ID       string
	Email    string
	Username string
	Password string // Should be hashed
	IsActive bool
}

// UserRepository defines user data access
type UserRepository interface {
	Save(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id string) error
}
