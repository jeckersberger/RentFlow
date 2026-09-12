package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
)

// PostgresSigningKeyRepository stores the single active auth signing key pair.
type PostgresSigningKeyRepository struct {
	db *sql.DB
}

func NewPostgresSigningKeyRepository(db *sql.DB) *PostgresSigningKeyRepository {
	return &PostgresSigningKeyRepository{db: db}
}

func (r *PostgresSigningKeyRepository) Load(ctx context.Context) (application.SigningKeyPair, error) {
	var pair application.SigningKeyPair
	err := r.db.QueryRowContext(ctx, `
		SELECT private_key_pem, public_key_pem
		FROM auth.signing_keys
		WHERE id = 'active'
	`).Scan(&pair.PrivateKeyPEM, &pair.PublicKeyPEM)
	if errors.Is(err, sql.ErrNoRows) {
		return application.SigningKeyPair{}, application.ErrSigningKeyNotFound
	}
	if err != nil {
		return application.SigningKeyPair{}, fmt.Errorf("query signing key: %w", err)
	}
	return pair, nil
}

func (r *PostgresSigningKeyRepository) CreateIfAbsent(ctx context.Context, pair application.SigningKeyPair) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth.signing_keys (id, private_key_pem, public_key_pem)
		VALUES ('active', $1, $2)
		ON CONFLICT (id) DO NOTHING
	`, pair.PrivateKeyPEM, pair.PublicKeyPEM)
	if err != nil {
		return fmt.Errorf("insert signing key: %w", err)
	}
	return nil
}
