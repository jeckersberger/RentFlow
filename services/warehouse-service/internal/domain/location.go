package domain

import (
	"fmt"
	"strings"
	"time"
)

type LocationType string

const (
	LocationSite     LocationType = "site"
	LocationBuilding LocationType = "building"
	LocationRoom     LocationType = "room"
	LocationRack     LocationType = "rack"
	LocationShelf    LocationType = "shelf"
	LocationBin      LocationType = "bin"
)

type Location struct {
	ID           string
	TenantID     string
	Name         string
	Type         LocationType
	ParentID     *string
	Path         string
	Capacity     int
	CurrentCount int
	SortOrder    int
	Barcode      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewLocation(id, tenantID, name string, locType LocationType) *Location {
	return &Location{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Type:      locType,
		Capacity:  100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (l *Location) Validate() error {
	if l.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if l.Name == "" {
		return fmt.Errorf("name is required")
	}
	if l.Capacity <= 0 {
		return fmt.Errorf("capacity must be greater than 0")
	}
	return nil
}

func (l *Location) UpdatePath(parentPath string) {
	if parentPath == "" {
		l.Path = l.ID
	} else {
		l.Path = fmt.Sprintf("%s/%s", parentPath, l.ID)
	}
}

func (l *Location) GetHierarchyLevel() int {
	if l.Path == "" {
		return 0
	}
	return strings.Count(l.Path, "/") + 1
}

func (l *Location) CanAddCapacity(count int) bool {
	return l.CurrentCount+count <= l.Capacity
}

func (l *Location) AddToCapacity(count int) error {
	if !l.CanAddCapacity(count) {
		return fmt.Errorf("insufficient capacity: %d available, %d requested", l.Capacity-l.CurrentCount, count)
	}
	l.CurrentCount += count
	l.UpdatedAt = time.Now()
	return nil
}

func (l *Location) RemoveFromCapacity(count int) error {
	if l.CurrentCount < count {
		return fmt.Errorf("cannot remove more items than available")
	}
	l.CurrentCount -= count
	l.UpdatedAt = time.Now()
	return nil
}
