package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

type AnonymizationRulePostgres struct {
	db *database.PostgresPool
}

func NewAnonymizationRulePostgres(db *database.PostgresPool) *AnonymizationRulePostgres {
	return &AnonymizationRulePostgres{db: db}
}

func (r *AnonymizationRulePostgres) Create(ctx context.Context, rule *domain.AnonymizationRule) error {
	query := `
		INSERT INTO anonymization_rules
		(id, tenant_id, pattern, replacement, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		rule.ID, rule.TenantID, rule.Pattern, rule.Replacement, rule.Type, rule.CreatedAt,
	)
	return err
}

func (r *AnonymizationRulePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.AnonymizationRule, error) {
	query := `
		SELECT id, tenant_id, pattern, replacement, type, created_at
		FROM anonymization_rules
		WHERE id = $1 AND tenant_id = $2
	`
	var rule domain.AnonymizationRule
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&rule.ID, &rule.TenantID, &rule.Pattern, &rule.Replacement, &rule.Type, &rule.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *AnonymizationRulePostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.AnonymizationRule, error) {
	query := `
		SELECT id, tenant_id, pattern, replacement, type, created_at
		FROM anonymization_rules
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*domain.AnonymizationRule
	for rows.Next() {
		var rule domain.AnonymizationRule
		err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Pattern, &rule.Replacement, &rule.Type, &rule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, rows.Err()
}

func (r *AnonymizationRulePostgres) ListByType(ctx context.Context, tenantID string, ruleType domain.AnonymizationRuleType) ([]*domain.AnonymizationRule, error) {
	query := `
		SELECT id, tenant_id, pattern, replacement, type, created_at
		FROM anonymization_rules
		WHERE tenant_id = $1 AND type = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, ruleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*domain.AnonymizationRule
	for rows.Next() {
		var rule domain.AnonymizationRule
		err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Pattern, &rule.Replacement, &rule.Type, &rule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, rows.Err()
}

func (r *AnonymizationRulePostgres) Update(ctx context.Context, rule *domain.AnonymizationRule) error {
	query := `
		UPDATE anonymization_rules
		SET pattern = $1, replacement = $2, type = $3
		WHERE id = $4 AND tenant_id = $5
	`
	_, err := r.db.Exec(ctx, query,
		rule.Pattern, rule.Replacement, rule.Type, rule.ID, rule.TenantID,
	)
	return err
}

func (r *AnonymizationRulePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM anonymization_rules WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

type AnonymizationMappingPostgres struct {
	db *database.PostgresPool
}

func NewAnonymizationMappingPostgres(db *database.PostgresPool) *AnonymizationMappingPostgres {
	return &AnonymizationMappingPostgres{db: db}
}

func (r *AnonymizationMappingPostgres) Create(ctx context.Context, mapping *domain.AnonymizationMapping) error {
	query := `
		INSERT INTO anonymization_mappings
		(id, tenant_id, original, anonymized, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		mapping.ID, mapping.TenantID, mapping.Original, mapping.Anonymized, mapping.CreatedAt,
	)
	return err
}

func (r *AnonymizationMappingPostgres) GetByID(ctx context.Context, id string) (*domain.AnonymizationMapping, error) {
	query := `
		SELECT id, tenant_id, original, anonymized, created_at
		FROM anonymization_mappings
		WHERE id = $1
	`
	var mapping domain.AnonymizationMapping
	err := r.db.QueryRow(ctx, query, id).Scan(
		&mapping.ID, &mapping.TenantID, &mapping.Original, &mapping.Anonymized, &mapping.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *AnonymizationMappingPostgres) GetByOriginal(ctx context.Context, tenantID, original string) (*domain.AnonymizationMapping, error) {
	query := `
		SELECT id, tenant_id, original, anonymized, created_at
		FROM anonymization_mappings
		WHERE tenant_id = $1 AND original = $2
	`
	var mapping domain.AnonymizationMapping
	err := r.db.QueryRow(ctx, query, tenantID, original).Scan(
		&mapping.ID, &mapping.TenantID, &mapping.Original, &mapping.Anonymized, &mapping.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *AnonymizationMappingPostgres) Delete(ctx context.Context, tenantID, mappingID string) error {
	query := `DELETE FROM anonymization_mappings WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, mappingID, tenantID)
	return err
}
