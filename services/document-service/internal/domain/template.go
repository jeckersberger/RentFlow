package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type TemplateType string

const (
	TemplateTypeInvoice      TemplateType = "invoice"
	TemplateTypeQuote        TemplateType = "quote"
	TemplateTypeDeliveryNote TemplateType = "delivery_note"
	TemplateTypeContract     TemplateType = "contract"
)

type Template struct {
	events.AggregateRoot
	TenantID    string
	Name        string
	Type        TemplateType
	Content     string
	Variables   []string
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
}

func NewTemplate(id, tenantID, name string, templateType TemplateType, content string, variables []string, createdBy string) *Template {
	now := time.Now()
	return &Template{
		AggregateRoot: *events.NewAggregateRoot(id, "template"),
		TenantID:      tenantID,
		Name:          name,
		Type:          templateType,
		Content:       content,
		Variables:     variables,
		IsDefault:     false,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     createdBy,
	}
}

func (t *Template) UpdateContent(content string) error {
	if content == "" {
		return fmt.Errorf("template content cannot be empty")
	}
	t.Content = content
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Template) SetDefault(isDefault bool) error {
	t.IsDefault = isDefault
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Template) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}
	if t.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if t.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	if t.Content == "" {
		return fmt.Errorf("template content cannot be empty")
	}
	return nil
}
