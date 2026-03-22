package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type ElectricalTestPostgres struct {
	db *database.PostgresPool
}

func NewElectricalTestPostgres(db *database.PostgresPool) *ElectricalTestPostgres {
	return &ElectricalTestPostgres{db: db}
}

func (r *ElectricalTestPostgres) Create(ctx context.Context, test *domain.ElectricalTest) error {
	query := `
		INSERT INTO electrical_tests (id, tenant_id, task_id, equipment_id, tester_id, test_type, test_date, next_test_date, result, insulation_resistance_mohm, protective_conductor_resistance_ohm, leakage_current_ma, visual_inspection_ok, functional_test_ok, test_device_id, test_device_name, certificate_number, notes, izytron_import_id, raw_xml, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`
	_, err := r.db.Exec(ctx, query,
		test.ID, test.TenantID, test.TaskID, test.EquipmentID, test.TesterID, test.TestType, test.TestDate, test.NextTestDate,
		test.Result, test.InsulationResistanceMohm, test.ProtectiveConductorResistanceOhm, test.LeakageCurrentMA,
		test.VisualInspectionOK, test.FunctionalTestOK, test.TestDeviceID, test.TestDeviceName, test.CertificateNumber,
		test.Notes, test.IzytronImportID, test.RawXML, test.CreatedAt,
	)
	return err
}

func (r *ElectricalTestPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.ElectricalTest, error) {
	query := `
		SELECT id, tenant_id, task_id, equipment_id, tester_id, test_type, test_date, next_test_date, result, insulation_resistance_mohm, protective_conductor_resistance_ohm, leakage_current_ma, visual_inspection_ok, functional_test_ok, test_device_id, test_device_name, certificate_number, notes, izytron_import_id, raw_xml, created_at
		FROM electrical_tests WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	test := &domain.ElectricalTest{}
	err := row.Scan(&test.ID, &test.TenantID, &test.TaskID, &test.EquipmentID, &test.TesterID, &test.TestType, &test.TestDate, &test.NextTestDate,
		&test.Result, &test.InsulationResistanceMohm, &test.ProtectiveConductorResistanceOhm, &test.LeakageCurrentMA,
		&test.VisualInspectionOK, &test.FunctionalTestOK, &test.TestDeviceID, &test.TestDeviceName, &test.CertificateNumber,
		&test.Notes, &test.IzytronImportID, &test.RawXML, &test.CreatedAt)
	if err != nil {
		return nil, err
	}

	return test, nil
}

func (r *ElectricalTestPostgres) ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.ElectricalTest, error) {
	query := `
		SELECT id, tenant_id, task_id, equipment_id, tester_id, test_type, test_date, next_test_date, result, insulation_resistance_mohm, protective_conductor_resistance_ohm, leakage_current_ma, visual_inspection_ok, functional_test_ok, test_device_id, test_device_name, certificate_number, notes, izytron_import_id, raw_xml, created_at
		FROM electrical_tests WHERE tenant_id = $1 AND equipment_id = $2 ORDER BY test_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []*domain.ElectricalTest
	for rows.Next() {
		test := &domain.ElectricalTest{}
		err := rows.Scan(&test.ID, &test.TenantID, &test.TaskID, &test.EquipmentID, &test.TesterID, &test.TestType, &test.TestDate, &test.NextTestDate,
			&test.Result, &test.InsulationResistanceMohm, &test.ProtectiveConductorResistanceOhm, &test.LeakageCurrentMA,
			&test.VisualInspectionOK, &test.FunctionalTestOK, &test.TestDeviceID, &test.TestDeviceName, &test.CertificateNumber,
			&test.Notes, &test.IzytronImportID, &test.RawXML, &test.CreatedAt)
		if err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}

	return tests, rows.Err()
}

func (r *ElectricalTestPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.ElectricalTest, error) {
	query := `
		SELECT id, tenant_id, task_id, equipment_id, tester_id, test_type, test_date, next_test_date, result, insulation_resistance_mohm, protective_conductor_resistance_ohm, leakage_current_ma, visual_inspection_ok, functional_test_ok, test_device_id, test_device_name, certificate_number, notes, izytron_import_id, raw_xml, created_at
		FROM electrical_tests WHERE tenant_id = $1 ORDER BY test_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []*domain.ElectricalTest
	for rows.Next() {
		test := &domain.ElectricalTest{}
		err := rows.Scan(&test.ID, &test.TenantID, &test.TaskID, &test.EquipmentID, &test.TesterID, &test.TestType, &test.TestDate, &test.NextTestDate,
			&test.Result, &test.InsulationResistanceMohm, &test.ProtectiveConductorResistanceOhm, &test.LeakageCurrentMA,
			&test.VisualInspectionOK, &test.FunctionalTestOK, &test.TestDeviceID, &test.TestDeviceName, &test.CertificateNumber,
			&test.Notes, &test.IzytronImportID, &test.RawXML, &test.CreatedAt)
		if err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}

	return tests, rows.Err()
}

func (r *ElectricalTestPostgres) ListByTask(ctx context.Context, tenantID, taskID string) ([]*domain.ElectricalTest, error) {
	query := `
		SELECT id, tenant_id, task_id, equipment_id, tester_id, test_type, test_date, next_test_date, result, insulation_resistance_mohm, protective_conductor_resistance_ohm, leakage_current_ma, visual_inspection_ok, functional_test_ok, test_device_id, test_device_name, certificate_number, notes, izytron_import_id, raw_xml, created_at
		FROM electrical_tests WHERE tenant_id = $1 AND task_id = $2 ORDER BY test_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []*domain.ElectricalTest
	for rows.Next() {
		test := &domain.ElectricalTest{}
		err := rows.Scan(&test.ID, &test.TenantID, &test.TaskID, &test.EquipmentID, &test.TesterID, &test.TestType, &test.TestDate, &test.NextTestDate,
			&test.Result, &test.InsulationResistanceMohm, &test.ProtectiveConductorResistanceOhm, &test.LeakageCurrentMA,
			&test.VisualInspectionOK, &test.FunctionalTestOK, &test.TestDeviceID, &test.TestDeviceName, &test.CertificateNumber,
			&test.Notes, &test.IzytronImportID, &test.RawXML, &test.CreatedAt)
		if err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}

	return tests, rows.Err()
}

func (r *ElectricalTestPostgres) Update(ctx context.Context, test *domain.ElectricalTest) error {
	query := `
		UPDATE electrical_tests
		SET task_id = $1, result = $2, insulation_resistance_mohm = $3, protective_conductor_resistance_ohm = $4, leakage_current_ma = $5, visual_inspection_ok = $6, functional_test_ok = $7, notes = $8
		WHERE id = $9 AND tenant_id = $10
	`
	_, err := r.db.Exec(ctx, query,
		test.TaskID, test.Result, test.InsulationResistanceMohm, test.ProtectiveConductorResistanceOhm, test.LeakageCurrentMA,
		test.VisualInspectionOK, test.FunctionalTestOK, test.Notes, test.ID, test.TenantID,
	)
	return err
}

func (r *ElectricalTestPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM electrical_tests WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}
