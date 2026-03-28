package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// UserRepo implements domain.UserRepository using PostgreSQL.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// scanUser scans a single user row into a domain.User, handling nullable columns.
func scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	var username, phone, avatarURL *string

	err := row.Scan(
		&u.ID, &u.TenantID, &u.Email, &username, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &phone, &avatarURL,
		&u.IsActive, &u.FailedLogins, &u.LockedUntil, &u.LastLogin,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if username != nil {
		u.Username = *username
	}
	if phone != nil {
		u.Phone = *phone
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	return u, nil
}

const userColumns = `
	id, tenant_id, email, username, password_hash,
	first_name, last_name, role, phone, avatar_url,
	is_active, failed_logins, locked_until, last_login,
	created_at, updated_at`

// Create inserts a new user and scans back the generated fields.
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, tenant_id, email, username, password_hash,
			first_name, last_name, role, phone, avatar_url,
			is_active, failed_logins, locked_until
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13
		) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		user.ID, user.TenantID, user.Email, nilIfEmpty(user.Username), user.PasswordHash,
		user.FirstName, user.LastName, user.Role, nilIfEmpty(user.Phone), nilIfEmpty(user.AvatarURL),
		user.IsActive, user.FailedLogins, user.LockedUntil,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("user_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a user by primary key.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, userColumns)
	u, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("user_repo: get_by_id: %w", err)
	}
	return u, nil
}

// GetByIDAndTenant retrieves a user by primary key scoped to a tenant.
func (r *UserRepo) GetByIDAndTenant(ctx context.Context, id, tenantID uuid.UUID) (*domain.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1 AND tenant_id = $2`, userColumns)
	u, err := scanUser(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("user_repo: get_by_id_and_tenant: %w", err)
	}
	return u, nil
}

// GetByEmail retrieves a user by tenant and email.
func (r *UserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	query := fmt.Sprintf(`SELECT %s FROM users WHERE tenant_id = $1 AND email = $2`, userColumns)
	u, err := scanUser(r.pool.QueryRow(ctx, query, tenantID, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("user_repo: get_by_email: %w", err)
	}
	return u, nil
}

// List returns a paginated list of users for a tenant plus total count.
func (r *UserRepo) List(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]*domain.User, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE tenant_id = $1`, tenantID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("user_repo: list_by_tenant count: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT %s FROM users WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userColumns,
	)
	rows, err := r.pool.Query(ctx, query, tenantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("user_repo: list_by_tenant query: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("user_repo: list_by_tenant scan: %w", scanErr)
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("user_repo: list_by_tenant rows: %w", err)
	}
	return users, total, nil
}

// Update modifies user profile fields (not password or login counters).
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET
			email = $2, username = $3, first_name = $4, last_name = $5,
			role = $6, phone = $7, avatar_url = $8, is_active = $9,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		user.ID, user.Email, nilIfEmpty(user.Username), user.FirstName, user.LastName,
		user.Role, nilIfEmpty(user.Phone), nilIfEmpty(user.AvatarURL), user.IsActive,
	).Scan(&user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailTaken
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("user_repo: update: %w", err)
	}
	return nil
}

// UpdatePassword sets a new password hash for a user.
func (r *UserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`,
		userID, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("user_repo: update_password: %w", err)
	}
	return nil
}

// IncrementFailedLogins increments the failed login counter and returns the new value.
func (r *UserRepo) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET failed_logins = failed_logins + 1 WHERE id = $1 RETURNING failed_logins`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("user_repo: increment_failed_logins: %w", err)
	}
	return count, nil
}

// ResetFailedLogins sets the failed login counter to zero.
func (r *UserRepo) ResetFailedLogins(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET failed_logins = 0 WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("user_repo: reset_failed_logins: %w", err)
	}
	return nil
}

// LockUntil sets the account lock timestamp.
func (r *UserRepo) LockUntil(ctx context.Context, userID uuid.UUID, lockedUntil time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET locked_until = $2 WHERE id = $1`,
		userID, lockedUntil,
	)
	if err != nil {
		return fmt.Errorf("user_repo: lock_until: %w", err)
	}
	return nil
}

// UpdateLastLogin sets the last_login timestamp to now.
func (r *UserRepo) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET last_login = NOW() WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("user_repo: update_last_login: %w", err)
	}
	return nil
}

// Deactivate marks a user as inactive.
func (r *UserRepo) Deactivate(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("user_repo: deactivate: %w", err)
	}
	return nil
}

// CountByTenant returns the number of active users in a tenant.
func (r *UserRepo) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND is_active = true`,
		tenantID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("user_repo: count_by_tenant: %w", err)
	}
	return count, nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
