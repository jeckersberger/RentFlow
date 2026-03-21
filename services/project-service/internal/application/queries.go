package application

import "time"

// Query objects for filtering and search

type ListProjectsQuery struct {
	TenantID string
	Status   *string
	Limit    int
	Offset   int
}

type SearchProjectsQuery struct {
	TenantID   string
	SearchTerm string
	Limit      int
	Offset     int
}

type ListPacklistsQuery struct {
	TenantID  string
	ProjectID string
	Status    *string
	Limit     int
	Offset    int
}

type ListReservationsQuery struct {
	TenantID  string
	ProjectID string
	Status    *string
}

type CheckReservationConflictQuery struct {
	TenantID    string
	EquipmentID string
	StartDate   time.Time
	EndDate     time.Time
}
