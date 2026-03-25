package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/ports"
)

// ===========================================================================
// Mock implementations
// ===========================================================================

// --- mockLogger (implements logger.Logger) ---------------------------------

type mockLoggerImpl struct{}

func (m *mockLoggerImpl) Debug(msg string, args ...interface{})                    {}
func (m *mockLoggerImpl) Info(msg string, args ...interface{})                     {}
func (m *mockLoggerImpl) Warn(msg string, args ...interface{})                     {}
func (m *mockLoggerImpl) Error(msg string, args ...interface{})                    {}
func (m *mockLoggerImpl) Fatal(msg string, args ...interface{})                    {}
func (m *mockLoggerImpl) WithPrefix(prefix string) logger.Logger                   { return m }
func (m *mockLoggerImpl) WithCorrelationID(id string) logger.Logger                { return m }
func (m *mockLoggerImpl) WithRequestID(id string) logger.Logger                    { return m }
func (m *mockLoggerImpl) WithField(key string, value interface{}) logger.Logger    { return m }

// Compile-time check that mockLoggerImpl satisfies logger.Logger
var _ logger.Logger = (*mockLoggerImpl)(nil)

// --- mockUserRepository ----------------------------------------------------

type mockUserRepository struct {
	users map[string]*domain.User // keyed by ID
}

func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{users: make(map[string]*domain.User)}
}

func (r *mockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *mockUserRepository) FindByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			if tenantID == "" || u.TenantID == tenantID {
				return u, nil
			}
		}
	}
	return nil, nil
}

func (r *mockUserRepository) List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.User, int, error) {
	var result []*domain.User
	for _, u := range r.users {
		if u.TenantID == tenantID {
			result = append(result, u)
		}
	}
	return result, len(result), nil
}

func (r *mockUserRepository) Save(ctx context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepository) Delete(ctx context.Context, id string) error {
	delete(r.users, id)
	return nil
}

// --- mockTenantRepository --------------------------------------------------

type mockTenantRepository struct {
	tenants map[string]*domain.Tenant
}

func newMockTenantRepo() *mockTenantRepository {
	return &mockTenantRepository{tenants: make(map[string]*domain.Tenant)}
}

func (r *mockTenantRepository) addTenant(id, name, slug string) {
	r.tenants[id] = domain.NewTenant(id, name, slug)
}

func (r *mockTenantRepository) FindByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t, ok := r.tenants[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (r *mockTenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	for _, t := range r.tenants {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, nil
}

func (r *mockTenantRepository) List(ctx context.Context, page, perPage int) ([]*domain.Tenant, int, error) {
	var result []*domain.Tenant
	for _, t := range r.tenants {
		result = append(result, t)
	}
	return result, len(result), nil
}

func (r *mockTenantRepository) Save(ctx context.Context, tenant *domain.Tenant) error {
	r.tenants[tenant.ID] = tenant
	return nil
}

func (r *mockTenantRepository) Delete(ctx context.Context, id string) error {
	delete(r.tenants, id)
	return nil
}

// --- mockInvitationRepository (minimal, not used in most tests) ------------

type mockInvitationRepository struct {
	invitations map[string]*ports.Invitation
}

func newMockInvitationRepo() *mockInvitationRepository {
	return &mockInvitationRepository{invitations: make(map[string]*ports.Invitation)}
}

func (r *mockInvitationRepository) Save(ctx context.Context, inv *ports.Invitation) error {
	r.invitations[inv.ID] = inv
	return nil
}

func (r *mockInvitationRepository) FindByToken(ctx context.Context, token string) (*ports.Invitation, error) {
	for _, inv := range r.invitations {
		if inv.Token == token {
			return inv, nil
		}
	}
	return nil, nil
}

func (r *mockInvitationRepository) FindByEmail(ctx context.Context, tenantID, email string) (*ports.Invitation, error) {
	for _, inv := range r.invitations {
		if inv.Email == email && inv.TenantID == tenantID {
			return inv, nil
		}
	}
	return nil, nil
}

func (r *mockInvitationRepository) ListByTenant(ctx context.Context, tenantID string) ([]*ports.Invitation, error) {
	var result []*ports.Invitation
	for _, inv := range r.invitations {
		if inv.TenantID == tenantID {
			result = append(result, inv)
		}
	}
	return result, nil
}

// ===========================================================================
// Test helpers
// ===========================================================================

const (
	testTenantID = "tenant-test-1"
	testPassword = "SecurePass123!"
)

// newTestUserService creates a UserService with mocks and a fresh RSA key pair.
func newTestUserService(t *testing.T) (*UserService, *mockUserRepository, *mockTenantRepository) {
	t.Helper()

	userRepo := newMockUserRepo()
	tenantRepo := newMockTenantRepo()
	tenantRepo.addTenant(testTenantID, "Test Tenant", "test-tenant")

	tm := newTestTokenManager(t)
	log := &mockLoggerImpl{}

	svc := NewUserService(userRepo, tenantRepo, tm, log)
	return svc, userRepo, tenantRepo
}

// registerTestUser is a helper that registers a user through the service.
func registerTestUser(t *testing.T, svc *UserService) *UserDTO {
	t.Helper()
	ctx := context.Background()
	dto, err := svc.Register(ctx, RegisterUserCommand{
		Email:     "user@example.com",
		Password:  testPassword,
		FirstName: "Test",
		LastName:  "User",
		TenantID:  testTenantID,
	})
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}
	return dto
}

// ===========================================================================
// Registration Tests
// ===========================================================================

func TestRegister_Success(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	dto := registerTestUser(t, svc)

	if dto.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got '%s'", dto.Email)
	}
	if dto.FirstName != "Test" {
		t.Errorf("expected first name 'Test', got '%s'", dto.FirstName)
	}
	if dto.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", dto.Status)
	}
	if dto.TenantID != testTenantID {
		t.Errorf("expected tenant ID '%s', got '%s'", testTenantID, dto.TenantID)
	}
	// Default role should be "readonly"
	if len(dto.Roles) == 0 {
		t.Fatal("expected at least one role")
	}
	hasReadonly := false
	for _, r := range dto.Roles {
		if r == "readonly" {
			hasReadonly = true
		}
	}
	if !hasReadonly {
		t.Errorf("expected 'readonly' role, got %v", dto.Roles)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	registerTestUser(t, svc)

	// Try to register again with the same email
	_, err := svc.Register(ctx, RegisterUserCommand{
		Email:     "user@example.com",
		Password:  testPassword,
		FirstName: "Another",
		LastName:  "User",
		TenantID:  testTenantID,
	})
	if err != domain.ErrEmailExists {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

func TestRegister_InvalidTenant(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterUserCommand{
		Email:     "user@example.com",
		Password:  testPassword,
		FirstName: "Test",
		LastName:  "User",
		TenantID:  "nonexistent-tenant",
	})
	if err != domain.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, RegisterUserCommand{
		Email:     "user@example.com",
		Password:  "weak",
		FirstName: "Test",
		LastName:  "User",
		TenantID:  testTenantID,
	})
	if err == nil {
		t.Fatal("expected error for weak password")
	}
}

// ===========================================================================
// Login Tests
// ===========================================================================

func TestLogin_Success(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	registerTestUser(t, svc)

	tokens, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokens == nil {
		t.Fatal("expected tokens, got nil")
	}
	if tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if tokens.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got '%s'", tokens.TokenType)
	}
	if tokens.ExpiresIn <= 0 {
		t.Errorf("expected positive expires_in, got %d", tokens.ExpiresIn)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	registerTestUser(t, svc)

	_, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: "WrongPassword123!",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.Login(ctx, LoginCommand{
		Email:    "nobody@example.com",
		Password: testPassword,
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	// Deactivate the user directly
	user := userRepo.users[dto.ID]
	user.Deactivate()

	_, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for inactive user, got %v", err)
	}
}

func TestLogin_DeletedUser(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	// Soft-delete the user
	user := userRepo.users[dto.ID]
	user.Status = domain.UserStatusDeleted

	_, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for deleted user, got %v", err)
	}
}

// ===========================================================================
// Brute-Force / Account Lockout Tests
// ===========================================================================

func TestLogin_AccountLockoutAfter5Failures(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	// Fail 5 times
	for i := 0; i < 5; i++ {
		_, err := svc.Login(ctx, LoginCommand{
			Email:    "user@example.com",
			Password: "WrongPassword123!",
		})
		if err != domain.ErrInvalidCredentials {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", i+1, err)
		}
	}

	// User should now be locked
	user := userRepo.users[dto.ID]
	if user.Status != domain.UserStatusLocked {
		t.Errorf("expected status 'locked', got '%s'", user.Status)
	}
	if user.FailedLogins != 5 {
		t.Errorf("expected 5 failed logins, got %d", user.FailedLogins)
	}

	// Even correct password should fail with ErrUserLocked
	_, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != domain.ErrUserLocked {
		t.Errorf("expected ErrUserLocked, got %v", err)
	}
}

func TestLogin_AutoUnlockAfter15Minutes(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	// Lock the user
	user := userRepo.users[dto.ID]
	user.Status = domain.UserStatusLocked
	user.FailedLogins = 5
	lockedAt := time.Now().Add(-16 * time.Minute) // 16 minutes ago
	user.LockedAt = &lockedAt

	// Login should succeed because auto-unlock kicks in
	tokens, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("expected auto-unlock login to succeed, got %v", err)
	}
	if tokens == nil {
		t.Fatal("expected tokens after auto-unlock")
	}

	// User should now be active
	user = userRepo.users[dto.ID]
	if user.Status != domain.UserStatusActive {
		t.Errorf("expected status 'active' after auto-unlock, got '%s'", user.Status)
	}
}

func TestLogin_StillLockedBefore15Minutes(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	// Lock the user recently (5 minutes ago)
	user := userRepo.users[dto.ID]
	user.Status = domain.UserStatusLocked
	user.FailedLogins = 5
	lockedAt := time.Now().Add(-5 * time.Minute)
	user.LockedAt = &lockedAt

	_, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != domain.ErrUserLocked {
		t.Errorf("expected ErrUserLocked (still within 15 min), got %v", err)
	}
}

func TestLogin_SuccessfulLoginResetsFailedCounter(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	// Fail 3 times
	for i := 0; i < 3; i++ {
		svc.Login(ctx, LoginCommand{
			Email:    "user@example.com",
			Password: "WrongPassword123!",
		})
	}

	user := userRepo.users[dto.ID]
	if user.FailedLogins != 3 {
		t.Fatalf("expected 3 failed logins, got %d", user.FailedLogins)
	}

	// Successful login should reset counter
	_, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("expected successful login, got %v", err)
	}

	user = userRepo.users[dto.ID]
	if user.FailedLogins != 0 {
		t.Errorf("expected 0 failed logins after success, got %d", user.FailedLogins)
	}
}

// ===========================================================================
// User Lifecycle Tests (Update, Deactivate, SoftDelete, Unlock)
// ===========================================================================

func TestUpdateProfile(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	err := svc.UpdateProfile(ctx, UpdateProfileCommand{
		UserID:    dto.ID,
		FirstName: "Updated",
		LastName:  "Name",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user := userRepo.users[dto.ID]
	if user.FirstName != "Updated" {
		t.Errorf("expected first name 'Updated', got '%s'", user.FirstName)
	}
	if user.LastName != "Name" {
		t.Errorf("expected last name 'Name', got '%s'", user.LastName)
	}
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	err := svc.UpdateProfile(ctx, UpdateProfileCommand{
		UserID:    "nonexistent",
		FirstName: "Test",
		LastName:  "User",
	})
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestDeactivateUser(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	err := svc.DeactivateUser(ctx, DeactivateUserCommand{UserID: dto.ID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user := userRepo.users[dto.ID]
	if user.Status != domain.UserStatusInactive {
		t.Errorf("expected status 'inactive', got '%s'", user.Status)
	}
}

func TestDeactivateUser_NotFound(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	err := svc.DeactivateUser(ctx, DeactivateUserCommand{UserID: "nonexistent"})
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestActivateUser(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)
	svc.DeactivateUser(ctx, DeactivateUserCommand{UserID: dto.ID})

	err := svc.ActivateUser(ctx, DeactivateUserCommand{UserID: dto.ID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user := userRepo.users[dto.ID]
	if user.Status != domain.UserStatusActive {
		t.Errorf("expected status 'active', got '%s'", user.Status)
	}
}

func TestSoftDeleteUser(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	err := svc.SoftDeleteUser(ctx, DeactivateUserCommand{UserID: dto.ID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user := userRepo.users[dto.ID]
	if user.Status != domain.UserStatusDeleted {
		t.Errorf("expected status 'deleted', got '%s'", user.Status)
	}
}

func TestSoftDeleteUser_NotFound(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	err := svc.SoftDeleteUser(ctx, DeactivateUserCommand{UserID: "nonexistent"})
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUnlockUser(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	// Lock the user
	user := userRepo.users[dto.ID]
	user.Lock()

	// Unlock via service
	err := svc.UnlockUser(ctx, UnlockUserCommand{UserID: dto.ID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user = userRepo.users[dto.ID]
	if user.Status != domain.UserStatusActive {
		t.Errorf("expected status 'active', got '%s'", user.Status)
	}
	if user.FailedLogins != 0 {
		t.Errorf("expected 0 failed logins, got %d", user.FailedLogins)
	}
}

func TestUnlockUser_NotFound(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	err := svc.UnlockUser(ctx, UnlockUserCommand{UserID: "nonexistent"})
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// ===========================================================================
// Role Tests
// ===========================================================================

func TestAssignRole(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	err := svc.AssignRole(ctx, AssignRoleCommand{
		UserID:   dto.ID,
		Role:     "admin",
		TenantID: testTenantID,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user := userRepo.users[dto.ID]
	hasAdmin := false
	for _, r := range user.Roles {
		if r == "admin" {
			hasAdmin = true
		}
	}
	if !hasAdmin {
		t.Errorf("expected 'admin' role, got %v", user.Roles)
	}
}

func TestRemoveRole(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)
	svc.AssignRole(ctx, AssignRoleCommand{UserID: dto.ID, Role: "admin", TenantID: testTenantID})

	err := svc.RemoveRole(ctx, dto.ID, "admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user := userRepo.users[dto.ID]
	for _, r := range user.Roles {
		if r == "admin" {
			t.Error("expected 'admin' role to be removed")
		}
	}
}

// ===========================================================================
// Password Change Tests
// ===========================================================================

func TestChangePassword_Success(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	newPassword := "NewSecurePass456!"
	err := svc.ChangePassword(ctx, ChangePasswordCommand{
		UserID:      dto.ID,
		OldPassword: testPassword,
		NewPassword: newPassword,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Login with new password should work
	tokens, err := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: newPassword,
	})
	if err != nil {
		t.Fatalf("expected login with new password to succeed, got %v", err)
	}
	if tokens == nil {
		t.Fatal("expected tokens")
	}

	// Login with old password should fail
	_, err = svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials with old password, got %v", err)
	}
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	err := svc.ChangePassword(ctx, ChangePasswordCommand{
		UserID:      dto.ID,
		OldPassword: "WrongOldPass123!",
		NewPassword: "NewSecurePass456!",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePassword_WeakNewPassword(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	err := svc.ChangePassword(ctx, ChangePasswordCommand{
		UserID:      dto.ID,
		OldPassword: testPassword,
		NewPassword: "weak",
	})
	if err == nil {
		t.Fatal("expected error for weak new password")
	}
}

func TestChangePassword_UserNotFound(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	err := svc.ChangePassword(ctx, ChangePasswordCommand{
		UserID:      "nonexistent",
		OldPassword: testPassword,
		NewPassword: "NewSecurePass456!",
	})
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// ===========================================================================
// GetUser / GetUserByEmail Tests
// ===========================================================================

func TestGetUser_Success(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)

	fetched, err := svc.GetUser(ctx, dto.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got '%s'", fetched.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.GetUser(ctx, "nonexistent")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetUserByEmail_Success(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	registerTestUser(t, svc)

	fetched, err := svc.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got '%s'", fetched.Email)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.GetUserByEmail(ctx, "nobody@example.com")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// ===========================================================================
// ListUsers Tests
// ===========================================================================

func TestListUsers(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	// Register 3 users
	for i := 0; i < 3; i++ {
		svc.Register(ctx, RegisterUserCommand{
			Email:     fmt.Sprintf("user%d@example.com", i),
			Password:  testPassword,
			FirstName: fmt.Sprintf("User%d", i),
			LastName:  "Test",
			TenantID:  testTenantID,
		})
	}

	result, err := svc.ListUsers(ctx, ListUsersQuery{
		TenantID: testTenantID,
		Page:     1,
		PerPage:  10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected 3 users, got %d", result.Total)
	}
}

// ===========================================================================
// RefreshToken Tests
// ===========================================================================

func TestRefreshToken_Success(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	registerTestUser(t, svc)

	tokens, _ := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})

	newTokens, err := svc.RefreshToken(ctx, tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("expected non-empty new access token")
	}
	if newTokens.RefreshToken == "" {
		t.Error("expected non-empty new refresh token")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.RefreshToken(ctx, "invalid-refresh-token")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
}

func TestRefreshToken_InactiveUser(t *testing.T) {
	svc, userRepo, _ := newTestUserService(t)
	ctx := context.Background()

	dto := registerTestUser(t, svc)
	tokens, _ := svc.Login(ctx, LoginCommand{
		Email:    "user@example.com",
		Password: testPassword,
	})

	// Deactivate the user after getting tokens
	user := userRepo.users[dto.ID]
	user.Deactivate()

	_, err := svc.RefreshToken(ctx, tokens.RefreshToken)
	if err == nil {
		t.Fatal("expected error when refreshing token for inactive user")
	}
}

// ===========================================================================
// ForgotPassword Tests (limited - no DB mock)
// ===========================================================================

func TestForgotPassword_NoDB(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	registerTestUser(t, svc)

	_, err := svc.ForgotPassword(ctx, ForgotPasswordCommand{Email: "user@example.com"})
	if err == nil {
		t.Fatal("expected error when DB is not configured")
	}
}

func TestForgotPassword_UnknownEmail_NoDB(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	// Unknown email should return no error and empty token (don't reveal user existence)
	token, err := svc.ForgotPassword(ctx, ForgotPasswordCommand{Email: "unknown@example.com"})
	if err != nil {
		t.Fatalf("expected no error for unknown email, got %v", err)
	}
	if token != "" {
		t.Error("expected empty token for unknown email")
	}
}

// ===========================================================================
// Invitation Tests (basic)
// ===========================================================================

func TestInviteUser_Success(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	invRepo := newMockInvitationRepo()
	svc.SetInvitationRepo(invRepo)
	ctx := context.Background()

	invDTO, err := svc.InviteUser(ctx, InviteUserCommand{
		TenantID:  testTenantID,
		Email:     "newuser@example.com",
		Role:      "manager",
		InvitedBy: "admin-user-id",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if invDTO.Email != "newuser@example.com" {
		t.Errorf("expected email 'newuser@example.com', got '%s'", invDTO.Email)
	}
	if invDTO.Role != "manager" {
		t.Errorf("expected role 'manager', got '%s'", invDTO.Role)
	}
	if invDTO.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", invDTO.Status)
	}
}

func TestInviteUser_NoInvitationRepo(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.InviteUser(ctx, InviteUserCommand{
		TenantID:  testTenantID,
		Email:     "newuser@example.com",
		Role:      "manager",
		InvitedBy: "admin-user-id",
	})
	if err == nil {
		t.Fatal("expected error when invitation repo is not configured")
	}
}

func TestInviteUser_DuplicateEmail(t *testing.T) {
	svc, _, _ := newTestUserService(t)
	invRepo := newMockInvitationRepo()
	svc.SetInvitationRepo(invRepo)
	ctx := context.Background()

	// Register a user first
	registerTestUser(t, svc)

	// Try to invite the same email
	_, err := svc.InviteUser(ctx, InviteUserCommand{
		TenantID:  testTenantID,
		Email:     "user@example.com",
		Role:      "manager",
		InvitedBy: "admin-user-id",
	})
	if err == nil {
		t.Fatal("expected error when inviting existing user")
	}
}
