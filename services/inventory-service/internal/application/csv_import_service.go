package application

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
)

type CSVImportService struct {
	equipRepo ports.EquipmentRepository
	catRepo   ports.CategoryRepository
	logger    logger.Logger
}

func NewCSVImportService(
	equipRepo ports.EquipmentRepository,
	catRepo ports.CategoryRepository,
	logger logger.Logger,
) *CSVImportService {
	return &CSVImportService{
		equipRepo: equipRepo,
		catRepo:   catRepo,
		logger:    logger,
	}
}

// ImportFromCSV imports equipment from a CSV file.
// Expected columns: name, description, sku, serial_number, barcode, category_name, daily_rate, weekly_rate
func (s *CSVImportService) ImportFromCSV(
	ctx context.Context,
	tenantID string,
	file io.Reader,
	userID string,
) (*ImportResult, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Map column names to indices
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Validate required columns
	requiredColumns := []string{"name", "barcode", "category_name"}
	for _, col := range requiredColumns {
		if _, ok := columnMap[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	result := &ImportResult{
		Errors: make([]ImportErrorItem, 0),
	}

	rowNum := 1
	for {
		rowNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, ImportErrorItem{
				Row:   rowNum,
				Error: fmt.Sprintf("failed to read row: %v", err),
			})
			continue
		}

		// Parse row
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue // Skip empty rows
		}

		name := getColumnValue(record, columnMap, "name")
		if name == "" {
			result.Errors = append(result.Errors, ImportErrorItem{
				Row:   rowNum,
				Error: "name is required",
			})
			result.Skipped++
			continue
		}

		barcode := getColumnValue(record, columnMap, "barcode")
		if barcode == "" {
			result.Errors = append(result.Errors, ImportErrorItem{
				Row:   rowNum,
				Error: "barcode is required",
			})
			result.Skipped++
			continue
		}

		categoryName := getColumnValue(record, columnMap, "category_name")
		description := getColumnValue(record, columnMap, "description")
		sku := getColumnValue(record, columnMap, "sku")
		serialNumber := getColumnValue(record, columnMap, "serial_number")

		dailyRateStr := getColumnValue(record, columnMap, "daily_rate")
		dailyRate := 0.0
		if dailyRateStr != "" {
			if rate, err := strconv.ParseFloat(dailyRateStr, 64); err == nil {
				dailyRate = rate
			}
		}

		weeklyRateStr := getColumnValue(record, columnMap, "weekly_rate")
		weeklyRate := 0.0
		if weeklyRateStr != "" {
			if rate, err := strconv.ParseFloat(weeklyRateStr, 64); err == nil {
				weeklyRate = rate
			}
		}

		// Find category by name
		categories, err := s.catRepo.List(ctx, tenantID)
		if err != nil {
			result.Errors = append(result.Errors, ImportErrorItem{
				Row:     rowNum,
				Barcode: barcode,
				Error:   fmt.Sprintf("failed to list categories: %v", err),
			})
			result.Skipped++
			continue
		}

		var categoryID string
		for _, cat := range categories {
			if cat.Name == categoryName {
				categoryID = cat.ID
				break
			}
		}

		if categoryID == "" {
			result.Errors = append(result.Errors, ImportErrorItem{
				Row:     rowNum,
				Barcode: barcode,
				Error:   fmt.Sprintf("category not found: %s", categoryName),
			})
			result.Skipped++
			continue
		}

		// Create equipment
		cmd := CreateEquipmentCommand{
			TenantID:        tenantID,
			Name:            name,
			Description:     description,
			CategoryID:      categoryID,
			SKU:             sku,
			SerialNumber:    serialNumber,
			Barcode:         barcode,
			RentalPriceDay:  dailyRate,
			RentalPriceWeek: weeklyRate,
			CreatedByUserID: userID,
		}

		// Create equipment service to handle creation
		// For now, we'll use the repository directly
		equipmentID := fmt.Sprintf("equip_%d", hashString(tenantID+barcode))
		eq := domain.NewEquipment(equipmentID, tenantID, name, categoryID, sku, barcode, userID)
		eq.Description = description
		eq.SerialNumber = serialNumber
		eq.RentalPriceDay = dailyRate
		eq.RentalPriceWeek = weeklyRate

		if err := eq.Validate(); err != nil {
			result.Errors = append(result.Errors, ImportErrorItem{
				Row:     rowNum,
				Barcode: barcode,
				Error:   fmt.Sprintf("validation error: %v", err),
			})
			result.Skipped++
			continue
		}

		if err := s.equipRepo.Create(ctx, eq); err != nil {
			// Check if it's a duplicate barcode error
			if strings.Contains(err.Error(), "unique constraint") {
				result.Errors = append(result.Errors, ImportErrorItem{
					Row:     rowNum,
					Barcode: barcode,
					Error:   "barcode already exists",
				})
			} else {
				result.Errors = append(result.Errors, ImportErrorItem{
					Row:     rowNum,
					Barcode: barcode,
					Error:   fmt.Sprintf("failed to create equipment: %v", err),
				})
			}
			result.Skipped++
			continue
		}

		_ = cmd // Suppress unused warning
		result.Created++
		s.logger.Info("Equipment imported from CSV", "barcode", barcode, "name", name)
	}

	return result, nil
}

func getColumnValue(record []string, columnMap map[string]int, columnName string) string {
	if idx, ok := columnMap[columnName]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}
