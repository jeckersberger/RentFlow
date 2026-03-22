package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
)

type SignaturePostgres struct {
	db *database.PostgresPool
}

func NewSignaturePostgres(db *database.PostgresPool) *SignaturePostgres {
	return &SignaturePostgres{db: db}
}

func (r *SignaturePostgres) Create(ctx context.Context, sig *domain.Signature) error {
	query := `
		INSERT INTO signatures
		(id, document_id, signer_name, signer_email, signer_role, signature_data,
		 signed_at, ip_address, user_agent, verified, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		sig.ID, sig.DocumentID, sig.SignerName, sig.SignerEmail, sig.SignerRole,
		sig.SignatureData, sig.SignedAt, sig.IPAddress, sig.UserAgent, sig.Verified,
		sig.CreatedAt,
	)
	return err
}

func (r *SignaturePostgres) GetByID(ctx context.Context, sigID string) (*domain.Signature, error) {
	query := `
		SELECT id, document_id, signer_name, signer_email, signer_role, signature_data,
		       signed_at, ip_address, user_agent, verified, created_at
		FROM signatures
		WHERE id = $1
	`
	var sig domain.Signature
	err := r.db.QueryRow(ctx, query, sigID).Scan(
		&sig.ID, &sig.DocumentID, &sig.SignerName, &sig.SignerEmail, &sig.SignerRole,
		&sig.SignatureData, &sig.SignedAt, &sig.IPAddress, &sig.UserAgent, &sig.Verified,
		&sig.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sig, nil
}

func (r *SignaturePostgres) ListByDocument(ctx context.Context, docID string) ([]*domain.Signature, error) {
	query := `
		SELECT id, document_id, signer_name, signer_email, signer_role, signature_data,
		       signed_at, ip_address, user_agent, verified, created_at
		FROM signatures
		WHERE document_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signatures []*domain.Signature
	for rows.Next() {
		var sig domain.Signature
		if err := rows.Scan(
			&sig.ID, &sig.DocumentID, &sig.SignerName, &sig.SignerEmail, &sig.SignerRole,
			&sig.SignatureData, &sig.SignedAt, &sig.IPAddress, &sig.UserAgent, &sig.Verified,
			&sig.CreatedAt,
		); err != nil {
			return nil, err
		}
		signatures = append(signatures, &sig)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return signatures, nil
}

func (r *SignaturePostgres) Update(ctx context.Context, sig *domain.Signature) error {
	query := `
		UPDATE signatures
		SET signature_data = $1, signed_at = $2, ip_address = $3,
		    user_agent = $4, verified = $5
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query,
		sig.SignatureData, sig.SignedAt, sig.IPAddress, sig.UserAgent,
		sig.Verified, sig.ID,
	)
	return err
}

func (r *SignaturePostgres) GetByDocumentAndEmail(ctx context.Context, docID, email string) (*domain.Signature, error) {
	query := `
		SELECT id, document_id, signer_name, signer_email, signer_role, signature_data,
		       signed_at, ip_address, user_agent, verified, created_at
		FROM signatures
		WHERE document_id = $1 AND signer_email = $2
	`
	var sig domain.Signature
	err := r.db.QueryRow(ctx, query, docID, email).Scan(
		&sig.ID, &sig.DocumentID, &sig.SignerName, &sig.SignerEmail, &sig.SignerRole,
		&sig.SignatureData, &sig.SignedAt, &sig.IPAddress, &sig.UserAgent, &sig.Verified,
		&sig.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sig, nil
}
