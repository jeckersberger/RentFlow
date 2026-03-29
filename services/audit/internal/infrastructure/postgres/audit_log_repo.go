package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/audit/internal/domain"
)

const auditLogColumns = `
	id, tenant_id, user_id, action, entity_type, entity_id,
	old_data, new_data, ip_address, user_agent, created_at`

type AuditLogRepo struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepo(pool *pgxpool.Pool) *AuditLogRepo {
	return &AuditLogRepo{pool: pool}
}

func scanAuditLog(row pgx.Row) (*domain.AuditLog, error) {
	entry := &domain.AuditLog{}
	var (
		userID    *uuid.UUID
		entityID  *uuid.UUID
		oldData   []byte
		newData   []byte
		ipAddress *string
		userAgent *string
	)

	err := row.Scan(
		&entry.ID, &entry.TenantID, &userID, &entry.Action,
		&entry.EntityType, &entityID,
		&oldData, &newData, &ipAddress, &userAgent, &entry.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	entry.UserID = userID
	entry.EntityID = entityID
	if oldData != nil {
		entry.OldData = json.RawMessage(oldData)
	}
	if newData != nil {
		entry.NewData = json.RawMessage(newData)
	}
	if ipAddress != nil {
		entry.IPAddress = *ipAddress
	}
	if userAgent != nil {
		entry.UserAgent = *userAgent
	}

	return entry, nil
}

func (r *AuditLogRepo) Create(ctx context.Context, entry *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, tenant_id, user_id, action, entity_type, entity_id,
			old_data, new_data, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.TenantID, entry.UserID, entry.Action,
		entry.EntityType, entry.EntityID,
		nullableJSON(entry.OldData), nullableJSON(entry.NewData),
		nilIfEmpty(entry.IPAddress), nilIfEmpty(entry.UserAgent),
	).Scan(&entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("audit_log_repo: create: %w", err)
	}
	return nil
}

func (r *AuditLogRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.AuditLog, error) {
	query := fmt.Sprintf(`SELECT %s FROM audit_logs WHERE id = $1 AND tenant_id = $2`, auditLogColumns)
	entry, err := scanAuditLog(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("audit_log_repo: get_by_id: %w", err)
	}
	return entry, nil
}

func (r *AuditLogRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.AuditLogFilter) ([]*domain.AuditLog, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", argIdx))
		args = append(args, filter.EntityType)
		argIdx++
	}
	if filter.EntityID != nil {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argIdx))
		args = append(args, *filter.EntityID)
		argIdx++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_log_repo: list count: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		auditLogColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_log_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.AuditLog
	for rows.Next() {
		entry, scanErr := scanAuditLog(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("audit_log_repo: list scan: %w", scanErr)
		}
		items = append(items, entry)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit_log_repo: list rows: %w", err)
	}
	return items, total, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableJSON(raw json.RawMessage) interface{} {
	if raw == nil || len(raw) == 0 {
		return nil
	}
	return []byte(raw)
}

// CreateWithHash inserts a hash-chained audit log entry including all new columns.
func (r *AuditLogRepo) CreateWithHash(ctx context.Context, entry *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, tenant_id, user_id, action, entity_type, entity_id,
			old_data, new_data, ip_address, user_agent,
			checksum, prev_checksum, aggregate_type, aggregate_id,
			event_type, event_payload, service_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING created_at, sequence_number`

	err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.TenantID, entry.UserID, entry.Action,
		entry.EntityType, entry.EntityID,
		nullableJSON(entry.OldData), nullableJSON(entry.NewData),
		nilIfEmpty(entry.IPAddress), nilIfEmpty(entry.UserAgent),
		nilIfEmpty(entry.Checksum), nilIfEmpty(entry.PrevChecksum),
		nilIfEmpty(entry.AggregateType), entry.AggregateID,
		nilIfEmpty(entry.EventType), nullableJSON(entry.EventPayload),
		nilIfEmpty(entry.ServiceName),
	).Scan(&entry.CreatedAt, &entry.SequenceNumber)
	if err != nil {
		return fmt.Errorf("audit_log_repo: create_with_hash: %w", err)
	}
	return nil
}

// GetLastChecksum returns the checksum of the most recent audit log entry
// for the given tenant. Returns empty string if no entries exist.
func (r *AuditLogRepo) GetLastChecksum(ctx context.Context, tenantID uuid.UUID) (string, error) {
	query := `SELECT COALESCE(checksum, '') FROM audit_logs WHERE tenant_id = $1 ORDER BY sequence_number DESC LIMIT 1`
	var checksum string
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(&checksum)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("audit_log_repo: get_last_checksum: %w", err)
	}
	return checksum, nil
}

// ListBySequenceRange returns all audit log entries for a tenant within the
// given sequence number range (inclusive), ordered by sequence_number ASC.
func (r *AuditLogRepo) ListBySequenceRange(ctx context.Context, tenantID uuid.UUID, fromSeq, toSeq int64) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, action, entity_type, entity_id,
			old_data, new_data, ip_address, user_agent, created_at,
			sequence_number, checksum, prev_checksum,
			aggregate_type, aggregate_id, event_type, event_payload, service_name
		FROM audit_logs
		WHERE tenant_id = $1 AND sequence_number >= $2 AND sequence_number <= $3
		ORDER BY sequence_number ASC`

	rows, err := r.pool.Query(ctx, query, tenantID, fromSeq, toSeq)
	if err != nil {
		return nil, fmt.Errorf("audit_log_repo: list_by_sequence_range: %w", err)
	}
	defer rows.Close()

	return scanExtendedRows(rows)
}

// ListByDateRange returns all audit log entries for a tenant within the given
// date range (inclusive), ordered by sequence_number ASC.
func (r *AuditLogRepo) ListByDateRange(ctx context.Context, tenantID uuid.UUID, fromDate, toDate time.Time) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, tenant_id, user_id, action, entity_type, entity_id,
			old_data, new_data, ip_address, user_agent, created_at,
			sequence_number, checksum, prev_checksum,
			aggregate_type, aggregate_id, event_type, event_payload, service_name
		FROM audit_logs
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY sequence_number ASC`

	rows, err := r.pool.Query(ctx, query, tenantID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("audit_log_repo: list_by_date_range: %w", err)
	}
	defer rows.Close()

	return scanExtendedRows(rows)
}

// scanExtendedRows scans rows that include all hash chain columns.
func scanExtendedRows(rows pgx.Rows) ([]*domain.AuditLog, error) {
	var items []*domain.AuditLog
	for rows.Next() {
		entry := &domain.AuditLog{}
		var (
			userID        *uuid.UUID
			entityID      *uuid.UUID
			oldData       []byte
			newData       []byte
			ipAddress     *string
			userAgent     *string
			checksum      *string
			prevChecksum  *string
			aggregateType *string
			aggregateID   *uuid.UUID
			eventType     *string
			eventPayload  []byte
			serviceName   *string
		)

		err := rows.Scan(
			&entry.ID, &entry.TenantID, &userID, &entry.Action,
			&entry.EntityType, &entityID,
			&oldData, &newData, &ipAddress, &userAgent, &entry.CreatedAt,
			&entry.SequenceNumber, &checksum, &prevChecksum,
			&aggregateType, &aggregateID, &eventType, &eventPayload, &serviceName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan extended row: %w", err)
		}

		entry.UserID = userID
		entry.EntityID = entityID
		if oldData != nil {
			entry.OldData = json.RawMessage(oldData)
		}
		if newData != nil {
			entry.NewData = json.RawMessage(newData)
		}
		if ipAddress != nil {
			entry.IPAddress = *ipAddress
		}
		if userAgent != nil {
			entry.UserAgent = *userAgent
		}
		if checksum != nil {
			entry.Checksum = *checksum
		}
		if prevChecksum != nil {
			entry.PrevChecksum = *prevChecksum
		}
		if aggregateType != nil {
			entry.AggregateType = *aggregateType
		}
		entry.AggregateID = aggregateID
		if eventType != nil {
			entry.EventType = *eventType
		}
		if eventPayload != nil {
			entry.EventPayload = json.RawMessage(eventPayload)
		}
		if serviceName != nil {
			entry.ServiceName = *serviceName
		}

		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan extended rows: %w", err)
	}
	return items, nil
}
