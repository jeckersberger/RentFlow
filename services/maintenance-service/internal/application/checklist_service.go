package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/ports"
)

type ChecklistService struct {
	checklistRepo ports.ChecklistRepository
	logger        logger.Logger
}

func NewChecklistService(repo ports.ChecklistRepository, log logger.Logger) *ChecklistService {
	return &ChecklistService{
		checklistRepo: repo,
		logger:        log,
	}
}

func (s *ChecklistService) CreateChecklist(ctx context.Context, cmd CreateChecklistCommand) (*ChecklistDTO, error) {
	if cmd.TenantID == "" || cmd.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	items := make([]domain.ChecklistItem, len(cmd.Items))
	for i, itemData := range cmd.Items {
		items[i] = domain.ChecklistItem{
			ID:          uuid.New().String(),
			Name:        itemData["name"].(string),
			Description: itemData["description"].(string),
			CheckType:   itemData["check_type"].(string),
			Required:    itemData["required"].(bool),
		}
		if unit, ok := itemData["unit"].(string); ok {
			items[i].Unit = &unit
		}
		if minVal, ok := itemData["min_value"].(float64); ok {
			items[i].MinValue = &minVal
		}
		if maxVal, ok := itemData["max_value"].(float64); ok {
			items[i].MaxValue = &maxVal
		}
	}

	checklist := &domain.Checklist{
		ID:          uuid.New().String(),
		TenantID:    cmd.TenantID,
		Name:        cmd.Name,
		Description: cmd.Description,
		Items:       items,
		Version:     1,
		CreatedAt:   time.Now(),
	}

	if err := s.checklistRepo.Create(ctx, checklist); err != nil {
		s.logger.Error("Failed to create checklist", err)
		return nil, err
	}

	return ChecklistToDTO(checklist), nil
}

func (s *ChecklistService) GetChecklist(ctx context.Context, tenantID, checklistID string) (*ChecklistDTO, error) {
	if tenantID == "" || checklistID == "" {
		return nil, domain.ErrInvalidInput
	}

	checklist, err := s.checklistRepo.GetByID(ctx, tenantID, checklistID)
	if err != nil {
		return nil, err
	}
	if checklist == nil {
		return nil, domain.ErrChecklistNotFound
	}

	return ChecklistToDTO(checklist), nil
}

func (s *ChecklistService) ListChecklists(ctx context.Context, tenantID string) ([]*ChecklistDTO, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	checklists, err := s.checklistRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to list checklists", err)
		return nil, err
	}

	dtos := make([]*ChecklistDTO, len(checklists))
	for i, c := range checklists {
		dtos[i] = ChecklistToDTO(c)
	}
	return dtos, nil
}

func (s *ChecklistService) UpdateChecklist(ctx context.Context, tenantID, checklistID string, cmd CreateChecklistCommand) (*ChecklistDTO, error) {
	if tenantID == "" || checklistID == "" {
		return nil, domain.ErrInvalidInput
	}

	checklist, err := s.checklistRepo.GetByID(ctx, tenantID, checklistID)
	if err != nil {
		return nil, err
	}
	if checklist == nil {
		return nil, domain.ErrChecklistNotFound
	}

	checklist.Name = cmd.Name
	checklist.Description = cmd.Description
	checklist.Version++

	items := make([]domain.ChecklistItem, len(cmd.Items))
	for i, itemData := range cmd.Items {
		items[i] = domain.ChecklistItem{
			ID:          uuid.New().String(),
			Name:        itemData["name"].(string),
			Description: itemData["description"].(string),
			CheckType:   itemData["check_type"].(string),
			Required:    itemData["required"].(bool),
		}
		if unit, ok := itemData["unit"].(string); ok {
			items[i].Unit = &unit
		}
		if minVal, ok := itemData["min_value"].(float64); ok {
			items[i].MinValue = &minVal
		}
		if maxVal, ok := itemData["max_value"].(float64); ok {
			items[i].MaxValue = &maxVal
		}
	}
	checklist.Items = items

	if err := s.checklistRepo.Update(ctx, checklist); err != nil {
		s.logger.Error("Failed to update checklist", err)
		return nil, err
	}

	return ChecklistToDTO(checklist), nil
}

func (s *ChecklistService) DeleteChecklist(ctx context.Context, tenantID, checklistID string) error {
	if tenantID == "" || checklistID == "" {
		return domain.ErrInvalidInput
	}

	return s.checklistRepo.Delete(ctx, tenantID, checklistID)
}
