package repositories

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/auth-service/internal/domain"
)

// PostgresUserRepository implements the UserRepository interface using PostgreSQL
type PostgresUserRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sql.DB, log logger.Logger) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves a user by ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name,
		       roles, status, failed_logins, last_login_at, locked_at, created_at, updated_at
		FROM auth.users
		WHERE id = $1
	`

	return r.scanUser(ctx, query, id)
}

// FindByEmail retrieves a user by email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	// If tenantID is empty, search across all tenants
	var query string
	var args []interface{}

	if tenantID != "" {
		query = `
			SELECT id, tenant_id, email, password_hash, first_name, last_name,
			       roles, status, failed_logins, last_login_at, locked_at, created_at, updated_at
			FROM auth.users
			WHERE tenant_id = $1 AND email = $2
		`
		args = []interface{}{tenantID, email}
	} else {
		query = `
			SELECT id, tenant_id, email, password_hash, first_name, last_name,
			       roles, status, failed_logins, last_login_at, locked_at, created_at, updated_at
			FROM auth.users
			WHERE email = $1
		`
		args = []interface{}{email}
	}

	row := r.db.QueryRowContext(ctx, query, args...)
	return r.scanUserRow(row)
}

// List retrieves users in a tenant with pagination
func (r *PostgresUserRepository) List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.User, int, error) {
	// Calculate offset
	offset := (page - 1) * perPage

	// Count total
	countQuery := `SELECT COUNT(*) FROM auth.users WHERE tenant_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		r.logger.Error("failed to count users", err, "tenantID", tenantID)
		return nil, 0, err
	}

	// Query users
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name,
		       roles, status, failed_logins, last_login_at, locked_at, created_at, updated_at
		FROM auth.users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, perPage, offset)
	if err != nil {
		r.logger.Error("failed to query users", err, "tenantID", tenantID)
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user, err := r.scanUserFromRow(rows)
		if err != nil {
			r.logger.Error("failed to scan user", err)
			continue
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating users", err)
		return nil, 0, err
	}

	return users, total, nil
}

// Save persists a user (creates or updates)
func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	// Check if user exists
	exists := false
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM auth.users WHERE id = $1)", user.ID).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check user existence", err, "id", user.ID)
		return err
	}

	now := time.Now()
	rolesJSON := strings.Join(user.Roles, ",")

	if exists {
		// Update
		query := `
			UPDATE auth.users
			SET email = $2, password_hash = $3, first_name = $4, last_name = $5,
			    roles = $6, status = $7, failed_logins = $8, last_login_at = $9, locked_at = $10, updated_at = $11
			WHERE id = $1
		`
		_, err := r.db.ExecContext(ctx, query,
			user.ID,
			user.Email,
			user.PasswordHash,
			user.FirstName,
			user.LastName,
			rolesJSON,
			string(user.Status),
			user.FailedLogins,
			user.LastLoginAt,
			user.LockedAt,
			now,
		)
		if err != nil {
			r.logger.Error("failed to update user", err, "id", user.ID)
			return err
		}
	} else {
		// Insert
		query := `
			INSERT INTO auth.users
			(id, tenant_id, email, password_hash, first_name, last_name, roles, status, failed_logins, locked_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err := r.db.ExecContext(ctx, query,
			user.ID,
			user.TenantID,
			user.Email,
			user.PasswordHash,
			user.FirstName,
			user.LastName,
			rolesJSON,
			string(user.Status),
			user.FailedLogins,
			user.LockedAt,
			now,
			now,
		)
		if err != nil {
			r.logger.Error("failed to insert user", err, "id", user.ID, "email", user.Email)
			return err
		}
	}

	return nil
}

// Delete deletes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM auth.users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete user", err, "id", id)
		return err
	}
	return nil
}

// Helper functions

func (r *PostgresUserRepository) scanUser(ctx context.Context, query string, args ...interface{}) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	return r.scanUserRow(row)
}

func (r *PostgresUserRepository) scanUserRow(row *sql.Row) (*domain.User, error) {
	var id, tenantID, email, passwordHash, firstName, lastName, status string
	var rolesStr sql.NullString
	var failedLogins int
	var lastLoginAt, lockedAt sql.NullTime
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &tenantID, &email, &passwordHash, &firstName, &lastName,
		&rolesStr, &status, &failedLogins, &lastLoginAt, &lockedAt, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse roles
	var roles []string
	if rolesStr.Valid && rolesStr.String != "" {
		roles = strings.Split(rolesStr.String, ",")
	}

	user := &domain.User{
		AggregateRoot: domain.AggregateRoot{
			ID:      id,
			Type:    "User",
			Changes: []interface{}{},
		},
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		TenantID:     tenantID,
		Roles:        roles,
		Status:       domain.UserStatus(status),
		FailedLogins: failedLogins,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	if lockedAt.Valid {
		user.LockedAt = &lockedAt.Time
	}

	return user, nil
}

func (r *PostgresUserRepository) scanUserFromRow(rows *sql.Rows) (*domain.User, error) {
	var id, tenantID, email, passwordHash, firstName, lastName, status string
	var rolesStr sql.NullString
	var failedLogins int
	var lastLoginAt, lockedAt sql.NullTime
	var createdAt, updatedAt time.Time

	err := rows.Scan(&id, &tenantID, &email, &passwordHash, &firstName, &lastName,
		&rolesStr, &status, &failedLogins, &lastLoginAt, &lockedAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	// Parse roles
	var roles []string
	if rolesStr.Valid && rolesStr.String != "" {
		roles = strings.Split(rolesStr.String, ",")
	}

	user := &domain.User{
		AggregateRoot: domain.AggregateRoot{
			ID:      id,
			Type:    "User",
			Changes: []interface{}{},
		},
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		TenantID:     tenantID,
		Roles:        roles,
		Status:       domain.UserStatus(status),
		FailedLogins: failedLogins,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	if lockedAt.Valid {
		user.LockedAt = &lockedAt.Time
	}

	return user, nil
}
