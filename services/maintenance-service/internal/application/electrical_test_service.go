package application

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/ports"
)

type ElectricalTestService struct {
	testRepo ports.ElectricalTestRepository
	logger   logger.Logger
}

func NewElectricalTestService(repo ports.ElectricalTestRepository, log logger.Logger) *ElectricalTestService {
	return &ElectricalTestService{
		testRepo: repo,
		logger:   log,
	}
}

func (s *ElectricalTestService) RecordTest(ctx context.Context, cmd RecordElectricalTestCommand) (*ElectricalTestDTO, error) {
	if cmd.TenantID == "" || cmd.EquipmentID == "" || cmd.TesterID == "" {
		return nil, domain.ErrInvalidInput
	}

	nextTestDate := cmd.TestDate.AddDate(1, 0, 0)

	test := &domain.ElectricalTest{
		ID:                          uuid.New().String(),
		TenantID:                    cmd.TenantID,
		TaskID:                      cmd.TaskID,
		EquipmentID:                 cmd.EquipmentID,
		TesterID:                    cmd.TesterID,
		TestType:                    domain.TestType(cmd.TestType),
		TestDate:                    cmd.TestDate,
		NextTestDate:                nextTestDate,
		Result:                      domain.TestResult(cmd.Result),
		InsulationResistanceMohm:    cmd.InsulationResistanceMohm,
		ProtectiveConductorResistanceOhm: cmd.ProtectiveConductorResistanceOhm,
		LeakageCurrentMA:            cmd.LeakageCurrentMA,
		VisualInspectionOK:          cmd.VisualInspectionOK,
		FunctionalTestOK:            cmd.FunctionalTestOK,
		TestDeviceID:                cmd.TestDeviceID,
		TestDeviceName:              cmd.TestDeviceName,
		CertificateNumber:           cmd.CertificateNumber,
		Notes:                       cmd.Notes,
		CreatedAt:                   time.Now(),
	}

	if err := s.testRepo.Create(ctx, test); err != nil {
		s.logger.Error("Failed to record electrical test", err)
		return nil, err
	}

	return TestToDTO(test), nil
}

func (s *ElectricalTestService) GetTest(ctx context.Context, tenantID, testID string) (*ElectricalTestDTO, error) {
	if tenantID == "" || testID == "" {
		return nil, domain.ErrInvalidInput
	}

	test, err := s.testRepo.GetByID(ctx, tenantID, testID)
	if err != nil {
		return nil, err
	}
	if test == nil {
		return nil, domain.ErrTestNotFound
	}

	return TestToDTO(test), nil
}

func (s *ElectricalTestService) ListTestsByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*ElectricalTestDTO, error) {
	if tenantID == "" || equipmentID == "" {
		return nil, domain.ErrInvalidInput
	}

	tests, err := s.testRepo.ListByEquipment(ctx, tenantID, equipmentID)
	if err != nil {
		s.logger.Error("Failed to list tests by equipment", err)
		return nil, err
	}

	dtos := make([]*ElectricalTestDTO, len(tests))
	for i, t := range tests {
		dtos[i] = TestToDTO(t)
	}
	return dtos, nil
}

func (s *ElectricalTestService) ListTestsByTenant(ctx context.Context, tenantID string) ([]*ElectricalTestDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	tests, err := s.testRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list tests by tenant", err)
		return nil, err
	}

	dtos := make([]*ElectricalTestDTO, len(tests))
	for i, t := range tests {
		dtos[i] = TestToDTO(t)
	}
	return dtos, nil
}

func (s *ElectricalTestService) ListTestsByTask(ctx context.Context, tenantID, taskID string) ([]*ElectricalTestDTO, error) {
	if tenantID == "" || taskID == "" {
		return nil, domain.ErrInvalidInput
	}

	tests, err := s.testRepo.ListByTask(ctx, tenantID, taskID)
	if err != nil {
		s.logger.Error("Failed to list tests by task", err)
		return nil, err
	}

	dtos := make([]*ElectricalTestDTO, len(tests))
	for i, t := range tests {
		dtos[i] = TestToDTO(t)
	}
	return dtos, nil
}

// IZYTRON XML import
type IzytronTestResult struct {
	XMLName       xml.Name `xml:"TestResult"`
	EquipmentID   string   `xml:"EquipmentId"`
	CertNumber    string   `xml:"CertificateNumber"`
	TestDate      string   `xml:"TestDate"`
	TestType      string   `xml:"TestType"`
	TestResult    string   `xml:"Result"`
	TesterID      string   `xml:"TesterId"`
	DeviceID      string   `xml:"DeviceId"`
	DeviceName    string   `xml:"DeviceName"`
	InsulationRes *float64 `xml:"InsulationResistance"`
	LeakageCurrent *float64 `xml:"LeakageCurrent"`
}

func (s *ElectricalTestService) ImportIzytronXML(ctx context.Context, tenantID, xmlData string) ([]*ElectricalTestDTO, error) {
	if tenantID == "" || xmlData == "" {
		return nil, domain.ErrInvalidInput
	}

	var results []IzytronTestResult
	if err := xml.Unmarshal([]byte(xmlData), &results); err != nil {
		s.logger.Error("Failed to parse IZYTRON XML", err)
		return nil, domain.ErrInvalidInput
	}

	var tests []*domain.ElectricalTest
	importID := uuid.New().String()

	for _, r := range results {
		testDate, err := time.Parse("2006-01-02", r.TestDate)
		if err != nil {
			testDate = time.Now()
		}

		test := &domain.ElectricalTest{
			ID:                 uuid.New().String(),
			TenantID:           tenantID,
			EquipmentID:        r.EquipmentID,
			TesterID:           r.TesterID,
			TestType:           domain.TestType(r.TestType),
			TestDate:           testDate,
			NextTestDate:       testDate.AddDate(1, 0, 0),
			Result:             domain.TestResult(r.TestResult),
			TestDeviceID:       r.DeviceID,
			TestDeviceName:     r.DeviceName,
			CertificateNumber:  r.CertNumber,
			VisualInspectionOK: true,
			FunctionalTestOK:   true,
			IzytronImportID:    &importID,
			RawXML:             &xmlData,
			CreatedAt:          time.Now(),
		}

		if r.InsulationRes != nil {
			test.InsulationResistanceMohm = r.InsulationRes
		}
		if r.LeakageCurrent != nil {
			test.LeakageCurrentMA = r.LeakageCurrent
		}

		if err := s.testRepo.Create(ctx, test); err != nil {
			s.logger.Error("Failed to create test from IZYTRON import", err)
			continue
		}

		tests = append(tests, test)
	}

	dtos := make([]*ElectricalTestDTO, len(tests))
	for i, t := range tests {
		dtos[i] = TestToDTO(t)
	}
	return dtos, nil
}
