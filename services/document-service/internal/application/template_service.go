package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/document-service/internal/ports"
)

type TemplateService struct {
	tplRepo ports.TemplateRepository
	logger  logger.Logger
}

func NewTemplateService(
	tplRepo ports.TemplateRepository,
	logger logger.Logger,
) *TemplateService {
	return &TemplateService{
		tplRepo: tplRepo,
		logger:  logger,
	}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, cmd CreateTemplateCommand) (*TemplateDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.Name == "" {
		return nil, domain.NewDomainError("NAME_REQUIRED", "template name is required", nil)
	}

	tplID := fmt.Sprintf("tpl_%d", hashString(cmd.TenantID+cmd.Name))
	tplType := domain.TemplateType(cmd.Type)

	tpl := domain.NewTemplate(tplID, cmd.TenantID, cmd.Name, tplType, cmd.Content, cmd.Variables, cmd.CreatedBy)

	if err := tpl.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.tplRepo.Create(ctx, tpl); err != nil {
		return nil, domain.NewDomainError("CREATE_FAILED", "failed to create template", err)
	}

	return TemplateToDTO(tpl), nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, tenantID, templateID string) (*TemplateDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	tpl, err := s.tplRepo.GetByID(ctx, tenantID, templateID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "template not found", err)
	}

	return TemplateToDTO(tpl), nil
}

func (s *TemplateService) ListTemplates(ctx context.Context, tenantID string, limit, offset int) (*TemplateListResult, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	tpls, total, err := s.tplRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, domain.NewDomainError("LIST_FAILED", "failed to list templates", err)
	}

	items := make([]*TemplateDTO, len(tpls))
	for i, tpl := range tpls {
		items[i] = TemplateToDTO(tpl)
	}

	return &TemplateListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, cmd UpdateTemplateCommand) (*TemplateDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	tpl, err := s.tplRepo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "template not found", err)
	}

	if cmd.Name != "" {
		tpl.Name = cmd.Name
	}
	if cmd.Type != "" {
		tpl.Type = domain.TemplateType(cmd.Type)
	}
	if cmd.Content != "" {
		if err := tpl.UpdateContent(cmd.Content); err != nil {
			return nil, domain.NewDomainError("UPDATE_FAILED", err.Error(), nil)
		}
	}
	if cmd.Variables != nil {
		tpl.Variables = cmd.Variables
	}

	if err := s.tplRepo.Update(ctx, tpl); err != nil {
		return nil, domain.NewDomainError("UPDATE_FAILED", "failed to update template", err)
	}

	return TemplateToDTO(tpl), nil
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, tenantID, templateID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	if err := s.tplRepo.Delete(ctx, tenantID, templateID); err != nil {
		return domain.NewDomainError("DELETE_FAILED", "failed to delete template", err)
	}

	return nil
}

func (s *TemplateService) GetPreview(ctx context.Context, tenantID, templateID string, variables map[string]string) (string, error) {
	if tenantID == "" {
		return "", domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	tpl, err := s.tplRepo.GetByID(ctx, tenantID, templateID)
	if err != nil {
		return "", domain.NewDomainError("NOT_FOUND", "template not found", err)
	}

	// Simple variable substitution (in real world, use a proper template engine)
	preview := tpl.Content
	for key, value := range variables {
		// Replace {{key}} with value
		placeholder := fmt.Sprintf("{{%s}}", key)
		preview = strings.ReplaceAll(preview, placeholder, value)
	}

	return preview, nil
}
