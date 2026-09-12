package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
)

// SetupService handles the one-time setup wizard
type SetupService struct {
	db        *sql.DB
	userSvc   *UserService
	tenantSvc *TenantService
	logger    logger.Logger
}

// SetupStatus represents the current setup status
type SetupStatus struct {
	IsCompleted bool       `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// SetupRequest represents the setup wizard request
type SetupRequest struct {
	// Step 1: Company
	CompanyName    string  `json:"company_name"`
	CompanyAddress string  `json:"company_address"`
	CompanySlug    string  `json:"company_slug"`
	Currency       string  `json:"currency"`
	TaxRate        float64 `json:"tax_rate"`
	InvoicePrefix  string  `json:"invoice_prefix"`

	// Step 2: Admin User
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
	AdminName     string `json:"admin_name"`

	// Step 3: Settings
	Language string `json:"language"`

	// Security
	SetupToken string `json:"setup_token"`
}

// NewSetupService creates a new setup service
func NewSetupService(
	db *sql.DB,
	userSvc *UserService,
	tenantSvc *TenantService,
	log logger.Logger,
) *SetupService {
	return &SetupService{
		db:        db,
		userSvc:   userSvc,
		tenantSvc: tenantSvc,
		logger:    log,
	}
}

// GetStatus returns whether setup is completed
func (s *SetupService) GetStatus(ctx context.Context) (*SetupStatus, error) {
	var isCompleted bool
	var completedAt *time.Time

	err := s.db.QueryRowContext(
		ctx,
		"SELECT is_completed, completed_at FROM auth.setup_state WHERE id = 1",
	).Scan(&isCompleted, &completedAt)

	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("failed to get setup status", err)
		return nil, err
	}

	return &SetupStatus{
		IsCompleted: isCompleted,
		CompletedAt: completedAt,
	}, nil
}

// InitializeSetupState initializes setup state on server startup.
// It returns the setup token to the caller for secure local delivery; the
// service itself never logs the token value.
func (s *SetupService) InitializeSetupState(ctx context.Context) (string, error) {
	var tableExists bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema='auth' AND table_name='setup_state')",
	).Scan(&tableExists)

	if err != nil {
		s.logger.Error("failed to check if setup_state table exists", err)
		return "", err
	}

	if !tableExists {
		s.logger.Warn("setup_state table does not exist; ensure migrations are run")
		return "", errors.New("setup_state table not found")
	}

	var count int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth.setup_state").Scan(&count)
	if err != nil {
		s.logger.Error("failed to check setup_state", err)
		return "", err
	}

	if count == 0 {
		token := generateSetupToken()
		_, err := s.db.ExecContext(
			ctx,
			"INSERT INTO auth.setup_state (id, is_completed, setup_token) VALUES (1, FALSE, $1)",
			token,
		)
		if err != nil {
			s.logger.Error("failed to initialize setup_state", err)
			return "", err
		}
		s.logger.Info("setup_state initialized with new token")
		return token, nil
	}

	var token string
	var isCompleted bool
	err = s.db.QueryRowContext(
		ctx,
		"SELECT setup_token, is_completed FROM auth.setup_state WHERE id = 1",
	).Scan(&token, &isCompleted)

	if err != nil {
		s.logger.Error("failed to retrieve setup token", err)
		return "", err
	}

	if isCompleted {
		s.logger.Info("setup is already completed")
		return "", nil
	}

	return token, nil
}

// CompleteSetup completes the setup wizard
// Validates the setup token, creates the tenant and admin user, and marks setup as completed
// Returns 403 Forbidden equivalent if already completed
func (s *SetupService) CompleteSetup(ctx context.Context, req SetupRequest) error {
	status, err := s.GetStatus(ctx)
	if err != nil {
		return err
	}

	if status.IsCompleted {
		return ErrSetupAlreadyCompleted
	}

	var storedToken string
	err = s.db.QueryRowContext(
		ctx,
		"SELECT setup_token FROM auth.setup_state WHERE id = 1",
	).Scan(&storedToken)

	if err != nil {
		s.logger.Error("failed to retrieve setup token for verification", err)
		return err
	}

	if req.SetupToken != storedToken {
		// Never log caller-supplied secret material, even partially. The previous
		// prefix logging also panicked when a malformed token was shorter than 8 bytes.
		s.logger.Warn("invalid setup token provided")
		return ErrInvalidSetupToken
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("failed to begin transaction", err)
		return err
	}
	defer tx.Rollback()

	tenantID := generateID()
	createTenantCmd := CreateTenantCommand{
		Name:            req.CompanyName,
		Slug:            req.CompanySlug,
		DefaultLanguage: req.Language,
		Currency:        req.Currency,
		TaxRate:         req.TaxRate,
		InvoicePrefix:   req.InvoicePrefix,
	}

	tenant := domain.NewTenant(tenantID, createTenantCmd.Name, createTenantCmd.Slug)
	tenant.Settings = domain.TenantSettings{
		DefaultLanguage: createTenantCmd.DefaultLanguage,
		Currency:        createTenantCmd.Currency,
		TaxRate:         createTenantCmd.TaxRate,
		InvoicePrefix:   createTenantCmd.InvoicePrefix,
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO auth.tenants (id, name, slug, default_language, currency, tax_rate, invoice_prefix)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tenant.ID, tenant.Name, tenant.Slug,
		tenant.Settings.DefaultLanguage, tenant.Settings.Currency,
		tenant.Settings.TaxRate, tenant.Settings.InvoicePrefix,
	)
	if err != nil {
		s.logger.Error("failed to create tenant", err)
		return err
	}

	passwordMgr := NewPasswordManager()
	if err := passwordMgr.ValidatePassword(req.AdminPassword); err != nil {
		s.logger.Warn("weak admin password provided", "error", err.Error())
		return err
	}

	passwordHash, err := passwordMgr.HashPassword(req.AdminPassword)
	if err != nil {
		s.logger.Error("failed to hash admin password", err)
		return err
	}

	userID := generateID()
	user := domain.NewUser(userID, req.AdminEmail, passwordHash, req.AdminName, "Administrator", tenantID)
	user.AssignRole("admin")

	rolesStr := domain.RolesToString(user.Roles)
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO auth.users (id, tenant_id, email, password_hash, first_name, last_name, roles, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.TenantID, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, rolesStr, "active",
	)
	if err != nil {
		s.logger.Error("failed to create admin user", err)
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE auth.setup_state SET is_completed = TRUE, completed_at = NOW(), completed_by = $1
		 WHERE id = 1`,
		userID,
	)
	if err != nil {
		s.logger.Error("failed to mark setup as completed", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transaction", err)
		return err
	}

	s.logger.Info("setup wizard completed successfully",
		"tenant_id", tenantID,
		"admin_email", req.AdminEmail,
		"company_name", req.CompanyName,
	)

	return nil
}

// IsSetupRequired returns true if setup is not yet completed
func (s *SetupService) IsSetupRequired(ctx context.Context) (bool, error) {
	status, err := s.GetStatus(ctx)
	if err != nil {
		return false, err
	}
	return !status.IsCompleted, nil
}

func generateSetupToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

var (
	ErrSetupAlreadyCompleted = errors.New("setup is already completed")
	ErrInvalidSetupToken     = errors.New("invalid setup token")
)
