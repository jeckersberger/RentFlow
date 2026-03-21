package domain

import (
	"fmt"
	"time"
)

type Packlist struct {
	ID        string
	ProjectID string
	TenantID  string
	Name      string
	Items     []PacklistItem
	Status    PacklistStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PacklistStatus string

const (
	PacklistDraft     PacklistStatus = "draft"
	PacklistConfirmed PacklistStatus = "confirmed"
	PacklistPacked    PacklistStatus = "packed"
	PacklistLoaded    PacklistStatus = "loaded"
	PacklistReturned  PacklistStatus = "returned"
)

type PacklistItem struct {
	ID               string
	PacklistID       string
	EquipmentID      string
	EquipmentName    string
	Quantity         int
	QuantityPacked   int
	QuantityReturned int
	Notes            string
	Status           PacklistItemStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PacklistItemStatus string

const (
	PacklistItemPending   PacklistItemStatus = "pending"
	PacklistItemPacked    PacklistItemStatus = "packed"
	PacklistItemLoaded    PacklistItemStatus = "loaded"
	PacklistItemReturned  PacklistItemStatus = "returned"
	PacklistItemMissing   PacklistItemStatus = "missing"
	PacklistItemDamaged   PacklistItemStatus = "damaged"
)

func NewPacklist(id, projectID, tenantID, name string) *Packlist {
	now := time.Now()
	return &Packlist{
		ID:        id,
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Items:     []PacklistItem{},
		Status:    PacklistDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (p *Packlist) AddItem(itemID, equipmentID, equipmentName string, quantity int) error {
	if equipmentID == "" {
		return fmt.Errorf("equipment ID cannot be empty")
	}
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// Check if equipment already in list
	for _, item := range p.Items {
		if item.EquipmentID == equipmentID {
			return fmt.Errorf("equipment already in packlist")
		}
	}

	item := PacklistItem{
		ID:            itemID,
		PacklistID:    p.ID,
		EquipmentID:   equipmentID,
		EquipmentName: equipmentName,
		Quantity:      quantity,
		Status:        PacklistItemPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	p.Items = append(p.Items, item)
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Packlist) RemoveItem(equipmentID string) error {
	for i, item := range p.Items {
		if item.EquipmentID == equipmentID {
			p.Items = append(p.Items[:i], p.Items[i+1:]...)
			p.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("equipment not found in packlist")
}

func (p *Packlist) MarkItemPacked(equipmentID string, quantityPacked int) error {
	if quantityPacked <= 0 {
		return fmt.Errorf("quantity packed must be greater than 0")
	}

	for i, item := range p.Items {
		if item.EquipmentID == equipmentID {
			if quantityPacked > item.Quantity {
				return fmt.Errorf("quantity packed cannot exceed required quantity")
			}
			p.Items[i].QuantityPacked = quantityPacked
			p.Items[i].Status = PacklistItemPacked
			p.Items[i].UpdatedAt = time.Now()
			p.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("equipment not found in packlist")
}

func (p *Packlist) MarkItemReturned(equipmentID string, quantityReturned int) error {
	if quantityReturned <= 0 {
		return fmt.Errorf("quantity returned must be greater than 0")
	}

	for i, item := range p.Items {
		if item.EquipmentID == equipmentID {
			if quantityReturned > item.Quantity {
				return fmt.Errorf("quantity returned cannot exceed required quantity")
			}
			p.Items[i].QuantityReturned = quantityReturned
			if quantityReturned == item.Quantity {
				p.Items[i].Status = PacklistItemReturned
			}
			p.Items[i].UpdatedAt = time.Now()
			p.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("equipment not found in packlist")
}

func (p *Packlist) ChangeStatus(newStatus PacklistStatus) error {
	if newStatus == p.Status {
		return fmt.Errorf("packlist already in status %s", newStatus)
	}

	switch p.Status {
	case PacklistDraft:
		if newStatus != PacklistConfirmed && newStatus != PacklistPacked {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case PacklistConfirmed:
		if newStatus != PacklistPacked {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case PacklistPacked:
		if newStatus != PacklistLoaded {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case PacklistLoaded:
		if newStatus != PacklistReturned {
			return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
		}
	case PacklistReturned:
		return fmt.Errorf("cannot transition from returned status")
	}

	p.Status = newStatus
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Packlist) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("packlist ID cannot be empty")
	}
	if p.ProjectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if p.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if p.Name == "" {
		return fmt.Errorf("packlist name cannot be empty")
	}
	return nil
}
