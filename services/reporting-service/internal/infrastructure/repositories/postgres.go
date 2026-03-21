package repositories

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type ReportPostgres struct {
	db *pgxpool.Pool
}

func NewReportPostgres(db *pgxpool.Pool) *ReportPostgres {
	return &ReportPostgres{db: db}
}

func (r *ReportPostgres) Create(ctx context.Context, report *domain.Report) error {
	paramsJSON, _ := json.Marshal(report.Parameters)
	query := `INSERT INTO reports (id, tenant_id, type, parameters, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, report.ID, report.TenantID, report.Type, paramsJSON, report.Status, report.CreatedAt, report.UpdatedAt)
	return err
}

func (r *ReportPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Report, error) {
	query := `SELECT id, tenant_id, type, parameters, status, generated_at, file_ref, created_at, updated_at FROM reports WHERE id = $1 AND tenant_id = $2`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	report := &domain.Report{}
	var paramsJSON []byte
	err := row.Scan(&report.ID, &report.TenantID, &report.Type, &paramsJSON, &report.Status, &report.GeneratedAt, &report.FileRef, &report.CreatedAt, &report.UpdatedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(paramsJSON, &report.Parameters)
	return report, nil
}

func (r *ReportPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Report, error) {
	query := `SELECT id, tenant_id, type, parameters, status, generated_at, file_ref, created_at, updated_at FROM reports WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*domain.Report
	for rows.Next() {
		report := &domain.Report{}
		var paramsJSON []byte
		if err := rows.Scan(&report.ID, &report.TenantID, &report.Type, &paramsJSON, &report.Status, &report.GeneratedAt, &report.FileRef, &report.CreatedAt, &report.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(paramsJSON, &report.Parameters)
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

func (r *ReportPostgres) Update(ctx context.Context, report *domain.Report) error {
	paramsJSON, _ := json.Marshal(report.Parameters)
	query := `UPDATE reports SET status = $1, generated_at = $2, file_ref = $3, updated_at = $4 WHERE id = $5 AND tenant_id = $6`
	_, err := r.db.Exec(ctx, query, report.Status, report.GeneratedAt, report.FileRef, report.UpdatedAt, report.ID, report.TenantID)
	return err
}

func (r *ReportPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM reports WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}
