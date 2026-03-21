package repositories

import (
	"context"
	"encoding/json"
	"time"
	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/audit-service/internal/domain"
)

type AuditEntryPostgres struct {
	db *database.PostgresPool
}

func NewAuditEntryPostgres(db *database.PostgresPool) *AuditEntryPostgres {
	return &AuditEntryPostgres{db: db}
}

func (r *AuditEntryPostgres) Create(ctx context.Context, entry *domain.AuditEntry) error {
	prevJSON, _ := json.Marshal(entry.PreviousState)
	newJSON, _ := json.Marshal(entry.NewState)
	query := `INSERT INTO audit_entries (id, tenant_id, timestamp, user_id, action, entity_type, entity_id, previous_state, new_state, ip_address, user_agent, hash, previous_hash) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.Exec(ctx, query, entry.ID, entry.TenantID, entry.Timestamp, entry.UserID, entry.Action, entry.EntityType, entry.EntityID, prevJSON, newJSON, entry.IPAddress, entry.UserAgent, entry.Hash, entry.PreviousHash)
	return err
}

func (r *AuditEntryPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, timestamp, user_id, action, entity_type, entity_id, previous_state, new_state, ip_address, user_agent, hash, previous_hash FROM audit_entries WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	entry := &domain.AuditEntry{}
	var prevJSON, newJSON []byte
	err := row.Scan(&entry.ID, &entry.TenantID, &entry.Timestamp, &entry.UserID, &entry.Action, &entry.EntityType, &entry.EntityID, &prevJSON, &newJSON, &entry.IPAddress, &entry.UserAgent, &entry.Hash, &entry.PreviousHash)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(prevJSON, &entry.PreviousState)
	json.Unmarshal(newJSON, &entry.NewState)
	return entry, nil
}

func (r *AuditEntryPostgres) ListByTenant(ctx context.Context, tenantID string, fromTime, toTime time.Time) ([]*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, timestamp, user_id, action, entity_type, entity_id, previous_state, new_state, ip_address, user_agent, hash, previous_hash FROM audit_entries WHERE tenant_id = $1`
	args := []interface{}{tenantID}

	if !fromTime.IsZero() {
		query += ` AND timestamp >= $2`
		args = append(args, fromTime)
	}
	if !toTime.IsZero() {
		if len(args) > 1 {
			query += ` AND timestamp <= $3`
		} else {
			query += ` AND timestamp <= $2`
		}
		args = append(args, toTime)
	}

	query += ` ORDER BY timestamp DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry := &domain.AuditEntry{}
		var prevJSON, newJSON []byte
		if err := rows.Scan(&entry.ID, &entry.TenantID, &entry.Timestamp, &entry.UserID, &entry.Action, &entry.EntityType, &entry.EntityID, &prevJSON, &newJSON, &entry.IPAddress, &entry.UserAgent, &entry.Hash, &entry.PreviousHash); err != nil {
			return nil, err
		}
		json.Unmarshal(prevJSON, &entry.PreviousState)
		json.Unmarshal(newJSON, &entry.NewState)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *AuditEntryPostgres) ListByEntity(ctx context.Context, tenantID, entityType, entityID string) ([]*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, timestamp, user_id, action, entity_type, entity_id, previous_state, new_state, ip_address, user_agent, hash, previous_hash FROM audit_entries WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3 ORDER BY timestamp DESC`
	rows, err := r.db.Query(ctx, query, tenantID, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry := &domain.AuditEntry{}
		var prevJSON, newJSON []byte
		if err := rows.Scan(&entry.ID, &entry.TenantID, &entry.Timestamp, &entry.UserID, &entry.Action, &entry.EntityType, &entry.EntityID, &prevJSON, &newJSON, &entry.IPAddress, &entry.UserAgent, &entry.Hash, &entry.PreviousHash); err != nil {
			return nil, err
		}
		json.Unmarshal(prevJSON, &entry.PreviousState)
		json.Unmarshal(newJSON, &entry.NewState)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *AuditEntryPostgres) ListByUser(ctx context.Context, tenantID, userID string) ([]*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, timestamp, user_id, action, entity_type, entity_id, previous_state, new_state, ip_address, user_agent, hash, previous_hash FROM audit_entries WHERE tenant_id = $1 AND user_id = $2 ORDER BY timestamp DESC`
	rows, err := r.db.Query(ctx, query, tenantID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		entry := &domain.AuditEntry{}
		var prevJSON, newJSON []byte
		if err := rows.Scan(&entry.ID, &entry.TenantID, &entry.Timestamp, &entry.UserID, &entry.Action, &entry.EntityType, &entry.EntityID, &prevJSON, &newJSON, &entry.IPAddress, &entry.UserAgent, &entry.Hash, &entry.PreviousHash); err != nil {
			return nil, err
		}
		json.Unmarshal(prevJSON, &entry.PreviousState)
		json.Unmarshal(newJSON, &entry.NewState)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *AuditEntryPostgres) GetLastEntry(ctx context.Context, tenantID string) (*domain.AuditEntry, error) {
	query := `SELECT id, tenant_id, timestamp, user_id, action, entity_type, entity_id, previous_state, new_state, ip_address, user_agent, hash, previous_hash FROM audit_entries WHERE tenant_id = $1 ORDER BY timestamp DESC LIMIT 1`
	row := r.db.QueryRow(ctx, query, tenantID)

	entry := &domain.AuditEntry{}
	var prevJSON, newJSON []byte
	err := row.Scan(&entry.ID, &entry.TenantID, &entry.Timestamp, &entry.UserID, &entry.Action, &entry.EntityType, &entry.EntityID, &prevJSON, &newJSON, &entry.IPAddress, &entry.UserAgent, &entry.Hash, &entry.PreviousHash)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(prevJSON, &entry.PreviousState)
	json.Unmarshal(newJSON, &entry.NewState)
	return entry, nil
}

type IntegrityCheckPostgres struct {
	db *database.PostgresPool
}

func NewIntegrityCheckPostgres(db *database.PostgresPool) *IntegrityCheckPostgres {
	return &IntegrityCheckPostgres{db: db}
}

func (r *IntegrityCheckPostgres) Create(ctx context.Context, check *domain.IntegrityCheck) error {
	query := `INSERT INTO integrity_checks (last_verified, status, entries_checked, errors_found) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, check.LastVerified, check.Status, check.EntriesChecked, check.ErrorsFound)
	return err
}

func (r *IntegrityCheckPostgres) GetLatest(ctx context.Context, tenantID string) (*domain.IntegrityCheck, error) {
	query := `SELECT last_verified, status, entries_checked, errors_found FROM integrity_checks ORDER BY last_verified DESC LIMIT 1`
	row := r.db.QueryRow(ctx, query)

	check := &domain.IntegrityCheck{}
	err := row.Scan(&check.LastVerified, &check.Status, &check.EntriesChecked, &check.ErrorsFound)
	if err != nil {
		return nil, err
	}
	return check, nil
}
