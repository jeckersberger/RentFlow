package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// SetupInitRequest carries the setup token for initialization.
type SetupInitRequest struct {
	Token string `json:"token"`
}

// SetupCompleteRequest carries all data to complete the first-time setup.
type SetupCompleteRequest struct {
	Token    string       `json:"token"`
	Company  SetupCompany `json:"company"`
	Admin    SetupAdmin   `json:"admin"`
	Industry string       `json:"industry"`
}

// SetupCompany holds company information for setup.
type SetupCompany struct {
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// SetupAdmin holds admin credentials for setup.
type SetupAdmin struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// SetupService handles the initial system setup wizard.
type SetupService struct {
	setupRepo  domain.SetupRepository
	tenantRepo domain.TenantRepository
	userRepo   domain.UserRepository
	logger     zerolog.Logger
}

// NewSetupService creates a new SetupService.
func NewSetupService(
	setupRepo domain.SetupRepository,
	tenantRepo domain.TenantRepository,
	userRepo domain.UserRepository,
	logger zerolog.Logger,
) *SetupService {
	return &SetupService{
		setupRepo:  setupRepo,
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

// Init handles the setup initialization handshake.
// Returns true if setup may proceed, false if already complete.
func (s *SetupService) Init(ctx context.Context, req SetupInitRequest) (bool, error) {
	if req.Token == "" {
		return false, domain.ErrInvalidSetupToken
	}

	state, err := s.setupRepo.GetState(ctx)
	if err != nil && err != domain.ErrSetupNotFound {
		s.logger.Error().Err(err).Msg("failed to get setup state")
		return false, fmt.Errorf("setup state lookup failed: %w", err)
	}

	// Already completed.
	if state != nil && state.IsComplete {
		return false, domain.ErrSetupComplete
	}

	// First access — create the state with a hashed token.
	if state == nil {
		tokenHash := hashSetupToken(req.Token)
		newState := &domain.SetupState{
			ID:         uuid.New(),
			TokenHash:  tokenHash,
			IsComplete: false,
			CreatedAt:  time.Now(),
		}
		if err := s.setupRepo.CreateState(ctx, newState); err != nil {
			s.logger.Error().Err(err).Msg("failed to create setup state")
			return false, fmt.Errorf("setup state creation failed: %w", err)
		}
		s.logger.Info().Msg("setup state initialized")
		return true, nil
	}

	// State exists but not complete — verify token.
	if hashSetupToken(req.Token) != state.TokenHash {
		return false, domain.ErrInvalidSetupToken
	}

	return true, nil
}

// Complete finishes the setup wizard: creates the tenant and admin user.
func (s *SetupService) Complete(ctx context.Context, req SetupCompleteRequest) (*domain.Tenant, *domain.User, error) {
	state, err := s.setupRepo.GetState(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get setup state for completion")
		return nil, nil, fmt.Errorf("setup state lookup failed: %w", err)
	}
	if state != nil && state.IsComplete {
		return nil, nil, domain.ErrSetupComplete
	}

	// Verify token.
	if state == nil || hashSetupToken(req.Token) != state.TokenHash {
		return nil, nil, domain.ErrInvalidSetupToken
	}

	// Validate inputs.
	if req.Company.Name == "" {
		return nil, nil, fmt.Errorf("company name is required")
	}
	if req.Company.Slug == "" {
		return nil, nil, fmt.Errorf("company slug is required")
	}
	if req.Admin.Email == "" {
		return nil, nil, domain.ErrInvalidEmail
	}
	if err := validatePassword(req.Admin.Password); err != nil {
		return nil, nil, err
	}

	// Create tenant.
	now := time.Now()
	tenant := &domain.Tenant{
		ID:        uuid.New(),
		Name:      req.Company.Name,
		Slug:      req.Company.Slug,
		Email:     req.Company.Email,
		Phone:     req.Company.Phone,
		Industry:  req.Industry,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		s.logger.Error().Err(err).Str("company", req.Company.Name).Msg("failed to create tenant")
		return nil, nil, fmt.Errorf("tenant creation failed: %w", err)
	}

	// Hash admin password.
	passwordHash, err := HashPassword(req.Admin.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash admin password")
		return nil, nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// Create admin user.
	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		Email:        req.Admin.Email,
		PasswordHash: passwordHash,
		FirstName:    req.Admin.FirstName,
		LastName:     req.Admin.LastName,
		Role:         "admin",
		IsActive:     true,
		FailedLogins: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("email", req.Admin.Email).Msg("failed to create admin user")
		return nil, nil, fmt.Errorf("admin user creation failed: %w", err)
	}

	// Mark setup complete.
	completedAt := time.Now()
	state.IsComplete = true
	state.CompletedAt = &completedAt
	state.CompletedBy = &user.ID

	if err := s.setupRepo.MarkComplete(ctx, state); err != nil {
		s.logger.Error().Err(err).Msg("failed to mark setup as complete")
		return nil, nil, fmt.Errorf("setup completion failed: %w", err)
	}

	s.logger.Info().
		Str("tenant_id", tenant.ID.String()).
		Str("admin_email", user.Email).
		Msg("setup wizard completed")

	return tenant, user, nil
}

func hashSetupToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
