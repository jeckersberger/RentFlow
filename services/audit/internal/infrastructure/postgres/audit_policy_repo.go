package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/audit/internal/domain"
)

const auditPolicyColumns = `
	id, tenant_id, entity_type, retention_days, log_reads,
	log_writes, is_active, created_at, updated_at`

type AuditPolicyRepo struct {
	pool *pgxpool.Pool
}

func NewAuditPolicyRepo(pool *pgxpool.Pool) *AuditPolicyRepo {
	return &AuditPolicyRepo{pool: pool}
}

func scanAuditPolicy(row pgx.Row) (*domain.AuditPolicy, error) {
	p := &domain.AuditPolicy{}
	err := row.Scan(
		&p.ID, &p.TenantID, &p.EntityType, &p.RetentionDays,
		&p.LogReads, &p.LogWrites, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *AuditPolicyRepo) Create(ctx context.Context, policy *domain.AuditPolicy) error {
	query := `
		INSERT INTO audit_policies (
			id, tenant_id, entity_type, retention_days,
			log_reads, log_writes, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		policy.ID, policy.TenantID, policy.EntityType,
		policy.RetentionDays, policy.LogReads, policy.LogWrites, policy.IsActive,
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)
	if err != nil {
		return fmt.Errorf("audit_policy_repo: create: %w", err)
	}
	return nil
}

func (r *AuditPolicyRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.AuditPolicy, error) {
	query := fmt.Sprintf(`SELECT %s FROM audit_policies WHERE id = $1 AND tenant_id = $2`, auditPolicyColumns)
	policy, err := scanAuditPolicy(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("audit_policy_repo: get_by_id: %w", err)
	}
	return policy, nil
}

func (r *AuditPolicyRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.AuditPolicyFilter) ([]*domain.AuditPolicy, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_policies WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_policy_repo: list count: %w", err)
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
		`SELECT %s FROM audit_policies WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		auditPolicyColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("audit_policy_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.AuditPolicy
	for rows.Next() {
		policy, scanErr := scanAuditPolicy(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("audit_policy_repo: list scan: %w", scanErr)
		}
		items = append(items, policy)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("audit_policy_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *AuditPolicyRepo) Update(ctx context.Context, policy *domain.AuditPolicy) error {
	query := `
		UPDATE audit_policies SET
			entity_type = $3, retention_days = $4,
			log_reads = $5, log_writes = $6, is_active = $7,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		policy.ID, policy.TenantID,
		policy.EntityType, policy.RetentionDays,
		policy.LogReads, policy.LogWrites, policy.IsActive,
	).Scan(&policy.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("audit_policy_repo: update: %w", err)
	}
	return nil
}
