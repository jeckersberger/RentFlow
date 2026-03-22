package ports

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
	"github.com/lib/pq"
)

type PartnerRepository interface {
	Create(ctx context.Context, partner *domain.FederationPartner) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.FederationPartner, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.FederationPartner, error)
	Update(ctx context.Context, partner *domain.FederationPartner) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SubRentalRequestRepository interface {
	Create(ctx context.Context, request *domain.SubRentalRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SubRentalRequest, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.SubRentalRequest, error)
	ListByPartner(ctx context.Context, partnerID uuid.UUID) ([]*domain.SubRentalRequest, error)
	Update(ctx context.Context, request *domain.SubRentalRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type EquipmentCacheRepository interface {
	Create(ctx context.Context, cache *domain.PartnerEquipmentCache) error
	ListByPartner(ctx context.Context, partnerID uuid.UUID) ([]*domain.PartnerEquipmentCache, error)
	DeleteByPartner(ctx context.Context, partnerID uuid.UUID) error
	Update(ctx context.Context, cache *domain.PartnerEquipmentCache) error
}

type CertificateRepository interface {
	Create(ctx context.Context, cert *domain.FederationCertificate) error
	GetByFingerprint(ctx context.Context, fingerprint string) (*domain.FederationCertificate, error)
	ListByPartner(ctx context.Context, partnerID uuid.UUID) ([]*domain.FederationCertificate, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.FederationCertificate, error)
	Update(ctx context.Context, cert *domain.FederationCertificate) error
}

// PostgreSQL Repositories

type postgresPartnerRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresPartnerRepository(db *sql.DB, log logger.Logger) PartnerRepository {
	return &postgresPartnerRepository{db: db, log: log}
}

func (r *postgresPartnerRepository) Create(ctx context.Context, partner *domain.FederationPartner) error {
	query := `
		INSERT INTO federation_partners 
		(id, tenant_id, partner_name, partner_endpoint, status, trust_level, cert_fingerprint, cert_expires_at, shared_categories, data_policy)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		partner.ID, partner.TenantID, partner.PartnerName, partner.PartnerEndpoint,
		partner.Status, partner.TrustLevel, partner.CertFingerprint, partner.CertExpiresAt,
		pq.Array(partner.SharedCategories), partner.DataPolicy)
	return err
}

func (r *postgresPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.FederationPartner, error) {
	query := `
		SELECT id, tenant_id, partner_name, partner_endpoint, status, trust_level, cert_fingerprint, cert_expires_at, shared_categories, data_policy, created_at, updated_at
		FROM federation_partners WHERE id = $1
	`
	p := &domain.FederationPartner{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.TenantID, &p.PartnerName, &p.PartnerEndpoint, &p.Status, &p.TrustLevel,
		&p.CertFingerprint, &p.CertExpiresAt, pq.Array(&p.SharedCategories), &p.DataPolicy, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPartnerNotFound
	}
	return p, err
}

func (r *postgresPartnerRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.FederationPartner, error) {
	query := `
		SELECT id, tenant_id, partner_name, partner_endpoint, status, trust_level, cert_fingerprint, cert_expires_at, shared_categories, data_policy, created_at, updated_at
		FROM federation_partners WHERE tenant_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var partners []*domain.FederationPartner
	for rows.Next() {
		p := &domain.FederationPartner{}
		err := rows.Scan(&p.ID, &p.TenantID, &p.PartnerName, &p.PartnerEndpoint, &p.Status, &p.TrustLevel,
			&p.CertFingerprint, &p.CertExpiresAt, pq.Array(&p.SharedCategories), &p.DataPolicy, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		partners = append(partners, p)
	}
	return partners, rows.Err()
}

func (r *postgresPartnerRepository) Update(ctx context.Context, partner *domain.FederationPartner) error {
	query := `
		UPDATE federation_partners 
		SET partner_name = $1, partner_endpoint = $2, status = $3, trust_level = $4, 
		    cert_fingerprint = $5, cert_expires_at = $6, shared_categories = $7, data_policy = $8, updated_at = NOW()
		WHERE id = $9
	`
	_, err := r.db.ExecContext(ctx, query,
		partner.PartnerName, partner.PartnerEndpoint, partner.Status, partner.TrustLevel,
		partner.CertFingerprint, partner.CertExpiresAt, pq.Array(partner.SharedCategories), partner.DataPolicy, partner.ID)
	return err
}

func (r *postgresPartnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM federation_partners WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

type postgresSubRentalRequestRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresSubRentalRequestRepository(db *sql.DB, log logger.Logger) SubRentalRequestRepository {
	return &postgresSubRentalRequestRepository{db: db, log: log}
}

func (r *postgresSubRentalRequestRepository) Create(ctx context.Context, req *domain.SubRentalRequest) error {
	query := `
		INSERT INTO sub_rental_requests 
		(id, tenant_id, partner_id, direction, status, equipment_category, equipment_description, quantity, start_date, end_date, daily_rate, total_amount, handover_document_id, invoice_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.ExecContext(ctx, query,
		req.ID, req.TenantID, req.PartnerID, req.Direction, req.Status, req.EquipmentCategory,
		req.EquipmentDescription, req.Quantity, req.StartDate, req.EndDate, req.DailyRate,
		req.TotalAmount, req.HandoverDocumentID, req.InvoiceID, req.Notes)
	return err
}

func (r *postgresSubRentalRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SubRentalRequest, error) {
	query := `
		SELECT id, tenant_id, partner_id, direction, status, equipment_category, equipment_description, quantity, start_date, end_date, daily_rate, total_amount, handover_document_id, invoice_id, notes, created_at, updated_at
		FROM sub_rental_requests WHERE id = $1
	`
	req := &domain.SubRentalRequest{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.TenantID, &req.PartnerID, &req.Direction, &req.Status, &req.EquipmentCategory,
		&req.EquipmentDescription, &req.Quantity, &req.StartDate, &req.EndDate, &req.DailyRate,
		&req.TotalAmount, &req.HandoverDocumentID, &req.InvoiceID, &req.Notes, &req.CreatedAt, &req.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRequestNotFound
	}
	return req, err
}

func (r *postgresSubRentalRequestRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.SubRentalRequest, error) {
	query := `
		SELECT id, tenant_id, partner_id, direction, status, equipment_category, equipment_description, quantity, start_date, end_date, daily_rate, total_amount, handover_document_id, invoice_id, notes, created_at, updated_at
		FROM sub_rental_requests WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT 100
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*domain.SubRentalRequest
	for rows.Next() {
		req := &domain.SubRentalRequest{}
		err := rows.Scan(&req.ID, &req.TenantID, &req.PartnerID, &req.Direction, &req.Status, &req.EquipmentCategory,
			&req.EquipmentDescription, &req.Quantity, &req.StartDate, &req.EndDate, &req.DailyRate,
			&req.TotalAmount, &req.HandoverDocumentID, &req.InvoiceID, &req.Notes, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}
	return reqs, rows.Err()
}

func (r *postgresSubRentalRequestRepository) ListByPartner(ctx context.Context, partnerID uuid.UUID) ([]*domain.SubRentalRequest, error) {
	query := `
		SELECT id, tenant_id, partner_id, direction, status, equipment_category, equipment_description, quantity, start_date, end_date, daily_rate, total_amount, handover_document_id, invoice_id, notes, created_at, updated_at
		FROM sub_rental_requests WHERE partner_id = $1 ORDER BY created_at DESC LIMIT 100
	`
	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*domain.SubRentalRequest
	for rows.Next() {
		req := &domain.SubRentalRequest{}
		err := rows.Scan(&req.ID, &req.TenantID, &req.PartnerID, &req.Direction, &req.Status, &req.EquipmentCategory,
			&req.EquipmentDescription, &req.Quantity, &req.StartDate, &req.EndDate, &req.DailyRate,
			&req.TotalAmount, &req.HandoverDocumentID, &req.InvoiceID, &req.Notes, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}
	return reqs, rows.Err()
}

func (r *postgresSubRentalRequestRepository) Update(ctx context.Context, req *domain.SubRentalRequest) error {
	query := `
		UPDATE sub_rental_requests 
		SET status = $1, handover_document_id = $2, invoice_id = $3, notes = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, req.Status, req.HandoverDocumentID, req.InvoiceID, req.Notes, req.ID)
	return err
}

func (r *postgresSubRentalRequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM sub_rental_requests WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

type postgresEquipmentCacheRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresEquipmentCacheRepository(db *sql.DB, log logger.Logger) EquipmentCacheRepository {
	return &postgresEquipmentCacheRepository{db: db, log: log}
}

func (r *postgresEquipmentCacheRepository) Create(ctx context.Context, cache *domain.PartnerEquipmentCache) error {
	query := `
		INSERT INTO partner_equipment_cache (id, partner_id, category, item_name, quantity_available, daily_rate)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, cache.ID, cache.PartnerID, cache.Category, cache.ItemName, cache.QuantityAvailable, cache.DailyRate)
	return err
}

func (r *postgresEquipmentCacheRepository) ListByPartner(ctx context.Context, partnerID uuid.UUID) ([]*domain.PartnerEquipmentCache, error) {
	query := `
		SELECT id, partner_id, category, item_name, quantity_available, daily_rate, last_synced_at
		FROM partner_equipment_cache WHERE partner_id = $1 ORDER BY category, item_name
	`
	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []*domain.PartnerEquipmentCache
	for rows.Next() {
		c := &domain.PartnerEquipmentCache{}
		err := rows.Scan(&c.ID, &c.PartnerID, &c.Category, &c.ItemName, &c.QuantityAvailable, &c.DailyRate, &c.LastSyncedAt)
		if err != nil {
			return nil, err
		}
		caches = append(caches, c)
	}
	return caches, rows.Err()
}

func (r *postgresEquipmentCacheRepository) DeleteByPartner(ctx context.Context, partnerID uuid.UUID) error {
	query := "DELETE FROM partner_equipment_cache WHERE partner_id = $1"
	_, err := r.db.ExecContext(ctx, query, partnerID)
	return err
}

func (r *postgresEquipmentCacheRepository) Update(ctx context.Context, cache *domain.PartnerEquipmentCache) error {
	query := `
		UPDATE partner_equipment_cache 
		SET quantity_available = $1, daily_rate = $2, last_synced_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, cache.QuantityAvailable, cache.DailyRate, cache.ID)
	return err
}

type postgresCertificateRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresCertificateRepository(db *sql.DB, log logger.Logger) CertificateRepository {
	return &postgresCertificateRepository{db: db, log: log}
}

func (r *postgresCertificateRepository) Create(ctx context.Context, cert *domain.FederationCertificate) error {
	query := `
		INSERT INTO federation_certificates (id, tenant_id, partner_id, cert_type, cert_pem, key_pem_encrypted, fingerprint, issued_at, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		cert.ID, cert.TenantID, cert.PartnerID, cert.CertType, cert.CertPem, cert.KeyPemEncr,
		cert.Fingerprint, cert.IssuedAt, cert.ExpiresAt, cert.IsActive)
	return err
}

func (r *postgresCertificateRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*domain.FederationCertificate, error) {
	query := `
		SELECT id, tenant_id, partner_id, cert_type, cert_pem, key_pem_encrypted, fingerprint, issued_at, expires_at, is_active, created_at
		FROM federation_certificates WHERE fingerprint = $1
	`
	cert := &domain.FederationCertificate{}
	err := r.db.QueryRowContext(ctx, query, fingerprint).Scan(
		&cert.ID, &cert.TenantID, &cert.PartnerID, &cert.CertType, &cert.CertPem, &cert.KeyPemEncr,
		&cert.Fingerprint, &cert.IssuedAt, &cert.ExpiresAt, &cert.IsActive, &cert.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrCertificateNotFound
	}
	return cert, err
}

func (r *postgresCertificateRepository) ListByPartner(ctx context.Context, partnerID uuid.UUID) ([]*domain.FederationCertificate, error) {
	query := `
		SELECT id, tenant_id, partner_id, cert_type, cert_pem, key_pem_encrypted, fingerprint, issued_at, expires_at, is_active, created_at
		FROM federation_certificates WHERE partner_id = $1 AND is_active = true ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []*domain.FederationCertificate
	for rows.Next() {
		cert := &domain.FederationCertificate{}
		err := rows.Scan(&cert.ID, &cert.TenantID, &cert.PartnerID, &cert.CertType, &cert.CertPem, &cert.KeyPemEncr,
			&cert.Fingerprint, &cert.IssuedAt, &cert.ExpiresAt, &cert.IsActive, &cert.CreatedAt)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	return certs, rows.Err()
}

func (r *postgresCertificateRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.FederationCertificate, error) {
	query := `
		SELECT id, tenant_id, partner_id, cert_type, cert_pem, key_pem_encrypted, fingerprint, issued_at, expires_at, is_active, created_at
		FROM federation_certificates WHERE tenant_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []*domain.FederationCertificate
	for rows.Next() {
		cert := &domain.FederationCertificate{}
		err := rows.Scan(&cert.ID, &cert.TenantID, &cert.PartnerID, &cert.CertType, &cert.CertPem, &cert.KeyPemEncr,
			&cert.Fingerprint, &cert.IssuedAt, &cert.ExpiresAt, &cert.IsActive, &cert.CreatedAt)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	return certs, rows.Err()
}

func (r *postgresCertificateRepository) Update(ctx context.Context, cert *domain.FederationCertificate) error {
	query := `
		UPDATE federation_certificates 
		SET is_active = $1 WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, cert.IsActive, cert.ID)
	return err
}
