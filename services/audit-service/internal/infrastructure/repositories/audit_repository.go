package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
)

type PostgresAuditRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewPostgresAuditRepository(db *sql.DB, log logger.Logger) *PostgresAuditRepository {
	return &PostgresAuditRepository{db: db, log: log}
}

func (r *PostgresAuditRepository) Create(ctx context.Context, entry *domain.AuditEntry) (*domain.AuditEntry, error) {
	oldValuesBytes, _ := json.Marshal(entry.OldValues)
	newValuesBytes, _ := json.Marshal(entry.NewValues)

	query := `INSERT INTO audit_log (id, tenant_id, timestamp, service_name, operation, entity_type, entity_id, user_id, user_name, old_values, new_values, ip_address, user_agent, checksum, previous_checksum, is_pseudonymized, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW())
	RETURNING id, sequence_number, created_at`

	err := r.db.QueryRowContext(ctx, query,
		entry.ID, entry.TenantID, entry.Timestamp, entry.ServiceName, entry.Operation, entry.EntityType,
		entry.EntityID, entry.UserID, entry.UserName, oldValuesBytes, newValuesBytes,
		entry.IPAddress, entry.UserAgent, entry.Checksum, entry.PreviousChecksum, entry.IsPseudonymized).
		Scan(&entry.ID, &entry.SequenceNumber, &entry.CreatedAt)

	if err != nil {
		r.log.Error("Failed to create audit entry", err)
		return nil, domain.ErrDatabaseError
	}

	return entry, nil
}

func (r *PostgresAuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, sequence_number, timestamp, service_name, operation, entity_type, entity_id, user_id, user_name, old_values, new_values, ip_address, user_agent, checksum, previous_checksum, is_pseudonymized, created_at
	FROM audit_log WHERE id = $1`

	entry := &domain.AuditEntry{}
	var oldValuesBytes, newValuesBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID, &entry.TenantID, &entry.SequenceNumber, &entry.Timestamp, &entry.ServiceName, &entry.Operation, &entry.EntityType,
		&entry.EntityID, &entry.UserID, &entry.UserName, &oldValuesBytes, &newValuesBytes, &entry.IPAddress, &entry.UserAgent,
		&entry.Checksum, &entry.PreviousChecksum, &entry.IsPseudonymized, &entry.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, domain.ErrAuditEntryNotFound
	}
	if err != nil {
		return nil, domain.ErrDatabaseError
	}

	json.Unmarshal(oldValuesBytes, &entry.OldValues)
	json.Unmarshal(newValuesBytes, &entry.NewValues)
	return entry, nil
}

func (r *PostgresAuditRepository) GetLastChecksum(ctx context.Context, tenantID uuid.UUID) (*string, error) {
	query := `SELECT checksum FROM audit_log WHERE tenant_id = $1 ORDER BY sequence_number DESC LIMIT 1`
	var checksum *string
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&checksum)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	return checksum, nil
}

func (r *PostgresAuditRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, sequence_number, timestamp, service_name, operation, entity_type, entity_id, user_id, user_name, old_values, new_values, ip_address, user_agent, checksum, previous_checksum, is_pseudonymized, created_at
	FROM audit_log WHERE tenant_id = $1 ORDER BY sequence_number ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var entries []*domain.AuditEntry

	for rows.Next() {
		entry := &domain.AuditEntry{}
		var oldValuesBytes, newValuesBytes []byte

		err := rows.Scan(&entry.ID, &entry.TenantID, &entry.SequenceNumber, &entry.Timestamp, &entry.ServiceName, &entry.Operation, &entry.EntityType,
			&entry.EntityID, &entry.UserID, &entry.UserName, &oldValuesBytes, &newValuesBytes, &entry.IPAddress, &entry.UserAgent,
			&entry.Checksum, &entry.PreviousChecksum, &entry.IsPseudonymized, &entry.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(oldValuesBytes, &entry.OldValues)
		json.Unmarshal(newValuesBytes, &entry.NewValues)
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresAuditRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, sequence_number, timestamp, service_name, operation, entity_type, entity_id, user_id, user_name, old_values, new_values, ip_address, user_agent, checksum, previous_checksum, is_pseudonymized, created_at
	FROM audit_log WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3 ORDER BY sequence_number DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, entityType, entityID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var entries []*domain.AuditEntry

	for rows.Next() {
		entry := &domain.AuditEntry{}
		var oldValuesBytes, newValuesBytes []byte

		err := rows.Scan(&entry.ID, &entry.TenantID, &entry.SequenceNumber, &entry.Timestamp, &entry.ServiceName, &entry.Operation, &entry.EntityType,
			&entry.EntityID, &entry.UserID, &entry.UserName, &oldValuesBytes, &newValuesBytes, &entry.IPAddress, &entry.UserAgent,
			&entry.Checksum, &entry.PreviousChecksum, &entry.IsPseudonymized, &entry.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(oldValuesBytes, &entry.OldValues)
		json.Unmarshal(newValuesBytes, &entry.NewValues)
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresAuditRepository) ListByUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, sequence_number, timestamp, service_name, operation, entity_type, entity_id, user_id, user_name, old_values, new_values, ip_address, user_agent, checksum, previous_checksum, is_pseudonymized, created_at
	FROM audit_log WHERE tenant_id = $1 AND user_id = $2 ORDER BY sequence_number DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var entries []*domain.AuditEntry

	for rows.Next() {
		entry := &domain.AuditEntry{}
		var oldValuesBytes, newValuesBytes []byte

		err := rows.Scan(&entry.ID, &entry.TenantID, &entry.SequenceNumber, &entry.Timestamp, &entry.ServiceName, &entry.Operation, &entry.EntityType,
			&entry.EntityID, &entry.UserID, &entry.UserName, &oldValuesBytes, &newValuesBytes, &entry.IPAddress, &entry.UserAgent,
			&entry.Checksum, &entry.PreviousChecksum, &entry.IsPseudonymized, &entry.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(oldValuesBytes, &entry.OldValues)
		json.Unmarshal(newValuesBytes, &entry.NewValues)
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresAuditRepository) ListByDateRange(ctx context.Context, tenantID uuid.UUID, from, to interface{}) ([]*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, sequence_number, timestamp, service_name, operation, entity_type, entity_id, user_id, user_name, old_values, new_values, ip_address, user_agent, checksum, previous_checksum, is_pseudonymized, created_at
	FROM audit_log WHERE tenant_id = $1 AND timestamp >= $2 AND timestamp <= $3 ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, from, to)
	if err != nil {
		return nil, domain.ErrDatabaseError
	}
	defer rows.Close()

	var entries []*domain.AuditEntry

	for rows.Next() {
		entry := &domain.AuditEntry{}
		var oldValuesBytes, newValuesBytes []byte

		err := rows.Scan(&entry.ID, &entry.TenantID, &entry.SequenceNumber, &entry.Timestamp, &entry.ServiceName, &entry.Operation, &entry.EntityType,
			&entry.EntityID, &entry.UserID, &entry.UserName, &oldValuesBytes, &newValuesBytes, &entry.IPAddress, &entry.UserAgent,
			&entry.Checksum, &entry.PreviousChecksum, &entry.IsPseudonymized, &entry.CreatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(oldValuesBytes, &entry.OldValues)
		json.Unmarshal(newValuesBytes, &entry.NewValues)
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresAuditRepository) Pseudonymize(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE audit_log SET is_pseudonymized = true, user_name = CONCAT('PSEUDONYM_', SUBSTRING(MD5(user_id::text), 1, 8))
	WHERE tenant_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, tenantID, userID)
	if err != nil {
		r.log.Error("Failed to pseudonymize user entries", err)
		return domain.ErrDatabaseError
	}

	return nil
}

func (r *PostgresAuditRepository) GetTodayCount(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM audit_log WHERE tenant_id = $1 AND created_at::date = CURRENT_DATE`
	var count int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, domain.ErrDatabaseError
	}
	return count, nil
}

func (r *PostgresAuditRepository) GetAll(ctx context.Context, tenantID uuid.UUID) ([]*domain.AuditEntry, error) {
	return r.ListByTenant(ctx, tenantID)
}
