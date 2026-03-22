package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type ExpenseCategory struct {
	events.AggregateRoot
	TenantID  string
	Name      string
	SKR03Code string
	SKR04Code string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewExpenseCategory(id, tenantID, name, skr03Code, skr04Code string) *ExpenseCategory {
	now := time.Now()
	return &ExpenseCategory{
		AggregateRoot: *events.NewAggregateRoot(id, "expense_category"),
		TenantID:      tenantID,
		Name:          name,
		SKR03Code:     skr03Code,
		SKR04Code:     skr04Code,
		IsDefault:     false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (c *ExpenseCategory) SetDefault(isDefault bool) error {
	c.IsDefault = isDefault
	c.UpdatedAt = time.Now()
	return nil
}

func (c *ExpenseCategory) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("category ID cannot be empty")
	}
	if c.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if c.Name == "" {
		return fmt.Errorf("category name cannot be empty")
	}
	return nil
}
