package application

// CreateCrewMemberCommand represents a command to create a crew member
type CreateCrewMemberCommand struct {
	TenantID           string   `json:"tenant_id"`
	FirstName          string   `json:"first_name"`
	LastName           string   `json:"last_name"`
	Email              string   `json:"email"`
	Phone              string   `json:"phone"`
	Role               string   `json:"role"`
	Status             string   `json:"status"`
	HourlyRate         *float64 `json:"hourly_rate,omitempty"`
	DailyRate          *float64 `json:"daily_rate,omitempty"`
	PreferredVehicleID *string  `json:"preferred_vehicle_id,omitempty"`
	EmergencyContact   string   `json:"emergency_contact"`
	Notes              string   `json:"notes"`
}

// UpdateCrewMemberCommand represents a command to update a crew member
type UpdateCrewMemberCommand struct {
	ID                 string   `json:"id"`
	TenantID           string   `json:"tenant_id"`
	FirstName          string   `json:"first_name"`
	LastName           string   `json:"last_name"`
	Email              string   `json:"email"`
	Phone              string   `json:"phone"`
	Role               string   `json:"role"`
	Status             string   `json:"status"`
	HourlyRate         *float64 `json:"hourly_rate,omitempty"`
	DailyRate          *float64 `json:"daily_rate,omitempty"`
	PreferredVehicleID *string  `json:"preferred_vehicle_id,omitempty"`
	EmergencyContact   string   `json:"emergency_contact"`
	Notes              string   `json:"notes"`
}

// CreateQualificationCommand represents a command to create a qualification
type CreateQualificationCommand struct {
	TenantID          string `json:"tenant_id"`
	CrewMemberID      string `json:"crew_member_id"`
	QualificationType string `json:"qualification_type"`
	IssuedAt          string `json:"issued_at"`
	ExpiresAt         *string `json:"expires_at,omitempty"`
	CertificateNumber string `json:"certificate_number"`
	IssuingAuthority  string `json:"issuing_authority"`
}

// CreateCrewAssignmentCommand represents a command to create an assignment
type CreateCrewAssignmentCommand struct {
	TenantID     string `json:"tenant_id"`
	CrewMemberID string `json:"crew_member_id"`
	ProjectID    *string `json:"project_id,omitempty"`
	TourID       *string `json:"tour_id,omitempty"`
	Role         string `json:"role"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
}

// UpdateCrewAssignmentCommand represents a command to update an assignment
type UpdateCrewAssignmentCommand struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	CrewMemberID string `json:"crew_member_id"`
	ProjectID    *string `json:"project_id,omitempty"`
	TourID       *string `json:"tour_id,omitempty"`
	Role         string `json:"role"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Status       string `json:"status"`
	Notes        string `json:"notes"`
}

// StartTimeRecordCommand represents a command to start a time record
type StartTimeRecordCommand struct {
	TenantID     string `json:"tenant_id"`
	CrewMemberID string `json:"crew_member_id"`
	AssignmentID *string `json:"assignment_id,omitempty"`
	Date         string `json:"date"`
	Notes        string `json:"notes"`
}

// StopTimeRecordCommand represents a command to stop a time record
type StopTimeRecordCommand struct {
	ID            string `json:"id"`
	BreakMinutes  int `json:"break_minutes"`
	OvertimeMinutes int `json:"overtime_minutes"`
	Notes         string `json:"notes"`
}

// ApproveTimeRecordCommand represents a command to approve a time record
type ApproveTimeRecordCommand struct {
	ID string `json:"id"`
}
