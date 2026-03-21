package domain

import (
	"fmt"
	"time"
)

type Category struct {
	ID            string
	TenantID      string
	Name          string
	ParentID      *string
	Icon          string
	Color         string
	SortOrder     int
	CreatedAt     time.Time
	CreatedByUserID string
}

func NewCategory(id, tenantID, name string) *Category {
	return &Category{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		SortOrder: 0,
		CreatedAt: time.Now(),
	}
}

func (c *Category) SetParent(parentID string) error {
	if parentID == "" {
		c.ParentID = nil
		return nil
	}
	c.ParentID = &parentID
	return nil
}

func (c *Category) SetIcon(icon string) {
	c.Icon = icon
}

func (c *Category) SetColor(color string) {
	c.Color = color
}

func (c *Category) SetSortOrder(order int) {
	c.SortOrder = order
}

func (c *Category) Validate() error {
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
