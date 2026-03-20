package application

import (
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
	commonerrors "github.com/jeckersberger/rentflow/pkg/common/errors"
)

// RegisterUserUseCase handles user registration
type RegisterUserUseCase struct {
	userRepo domain.UserRepository
}

// RegisterUserRequest is the input for registering a user
type RegisterUserRequest struct {
	Email    string
	Username string
	Password string
}

// RegisterUserResponse is the output of user registration
type RegisterUserResponse struct {
	UserID string
	Email  string
}

// NewRegisterUserUseCase creates a new registration use case
func NewRegisterUserUseCase(userRepo domain.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo: userRepo,
	}
}

// Execute registers a new user
func (uc *RegisterUserUseCase) Execute(req *RegisterUserRequest) (*RegisterUserResponse, error) {
	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return nil, commonerrors.New(commonerrors.CodeValidation, "Email, username, and password are required")
	}

	// Check if user already exists
	existing, err := uc.userRepo.GetByEmail(req.Email)
	if err == nil && existing != nil {
		return nil, commonerrors.New(commonerrors.CodeConflict, "Email already registered")
	}

	// Create user (password should be hashed in real implementation)
	user := &domain.User{
		ID:       generateID(), // TODO: Use UUID library
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password, // TODO: Hash password
		IsActive: true,
	}

	if err := uc.userRepo.Save(user); err != nil {
		return nil, commonerrors.New(commonerrors.CodeDatabaseError, "Failed to save user").WithError(err)
	}

	return &RegisterUserResponse{
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

func generateID() string {
	// TODO: Use proper UUID generation
	return "user_" + randomString(12)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
