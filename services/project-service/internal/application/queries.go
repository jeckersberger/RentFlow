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

type ListCustomersQuery struct {
	TenantID string
	Limit    int
	Offset   int
}

type SearchCustomersQuery struct {
	TenantID   string
	SearchTerm string
	Limit      int
	Offset     int
}

type GetCalendarQuery struct {
	TenantID  string
	StartDate string
	EndDate   string
}

type CalendarEventDTO struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ClientName string    `json:"client_name"`
	Status     string    `json:"status"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Color      string    `json:"color"`
}
