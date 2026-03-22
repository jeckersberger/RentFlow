package domain

import (
	"time"
)

// CrewMember represents a crew member entity
type CrewMember struct {
	ID                 string
	TenantID           string
	FirstName          string
	LastName           string
	Email              string
	Phone              string
	Role               CrewRole
	Status             CrewStatus
	HourlyRate         *float64
	DailyRate          *float64
	PreferredVehicleID *string
	EmergencyContact   string
	Notes              string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// CrewRole represents crew member roles
type CrewRole string

const (
	CrewRoleTechnician    CrewRole = "technician"
	CrewRoleRigger        CrewRole = "rigger"
	CrewRoleDriver        CrewRole = "driver"
	CrewRoleSoundEngineer CrewRole = "sound_engineer"
	CrewRoleLightingTech  CrewRole = "lighting_tech"
	CrewRoleFreelancer    CrewRole = "freelancer"
)

// CrewStatus represents crew member status
type CrewStatus string

const (
	CrewStatusActive   CrewStatus = "active"
	CrewStatusInactive CrewStatus = "inactive"
	CrewStatusOnLeave  CrewStatus = "on_leave"
)

// Qualification represents crew member qualifications
type Qualification struct {
	ID                string
	TenantID          string
	CrewMemberID      string
	QualificationType QualificationType
	IssuedAt          time.Time
	ExpiresAt         *time.Time
	CertificateNumber string
	IssuingAuthority  string
	Status            QualificationStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// QualificationType represents types of qualifications
type QualificationType string

const (
	QualTypeIPAF          QualificationType = "IPAF"
	QualTypeRiggerCert    QualificationType = "rigger_cert"
	QualTypeSoundEngineer QualificationType = "sound_engineer"
	QualTypeLightingTech  QualificationType = "lighting_tech"
	QualTypeElectricalCert QualificationType = "electrical_cert"
	QualTypeDriverLicenseC QualificationType = "driver_license_c"
	QualTypeDriverLicenseCE QualificationType = "driver_license_ce"
	QualTypeFirstAid      QualificationType = "first_aid"
	QualTypeForklift      QualificationType = "forklift"
)

// QualificationStatus represents qualification validity status
type QualificationStatus string

const (
	QualStatusValid         QualificationStatus = "valid"
	QualStatusExpired       QualificationStatus = "expired"
	QualStatusPendingRenewal QualificationStatus = "pending_renewal"
)

// CrewAssignment represents crew assignments to projects/tours
type CrewAssignment struct {
	ID           string
	TenantID     string
	CrewMemberID string
	ProjectID    *string
	TourID       *string
	Role         string
	StartDate    time.Time
	EndDate      time.Time
	Status       AssignmentStatus
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AssignmentStatus represents assignment status
type AssignmentStatus string

const (
	AssignmentStatusPlanned    AssignmentStatus = "planned"
	AssignmentStatusConfirmed  AssignmentStatus = "confirmed"
	AssignmentStatusActive     AssignmentStatus = "active"
	AssignmentStatusCompleted  AssignmentStatus = "completed"
	AssignmentStatusCancelled  AssignmentStatus = "cancelled"
)

// TimeRecord represents time tracking for crew members
type TimeRecord struct {
	ID              string
	TenantID        string
	CrewMemberID    string
	AssignmentID    *string
	Date            time.Time
	StartTime       time.Time
	EndTime         *time.Time
	BreakMinutes    int
	OvertimeMinutes int
	Status          TimeRecordStatus
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TimeRecordStatus represents time record status
type TimeRecordStatus string

const (
	TimeRecordStatusRunning   TimeRecordStatus = "running"
	TimeRecordStatusCompleted TimeRecordStatus = "completed"
	TimeRecordStatusApproved  TimeRecordStatus = "approved"
)

// IsAvailable checks if crew member is available for a date range
func (c *CrewMember) IsAvailable(startDate, endDate time.Time, assignments []*CrewAssignment) bool {
	if c.Status != CrewStatusActive {
		return false
	}

	for _, assignment := range assignments {
		if assignment.CrewMemberID != c.ID {
			continue
		}
		if assignment.Status == AssignmentStatusCancelled {
			continue
		}

		// Check for overlap
		if !(endDate.Before(assignment.StartDate) || startDate.After(assignment.EndDate)) {
			return false
		}
	}

	return true
}

// HasQualification checks if crew member has a specific qualification that is valid
func (c *CrewMember) HasQualification(qualType QualificationType, qualifications []*Qualification) bool {
	for _, q := range qualifications {
		if q.CrewMemberID != c.ID {
			continue
		}
		if q.QualificationType != qualType {
			continue
		}
		if q.Status != QualStatusValid {
			continue
		}
		// Check if not expired
		if q.ExpiresAt != nil && q.ExpiresAt.Before(time.Now()) {
			continue
		}
		return true
	}
	return false
}

// CanDrive checks if crew member can drive
func (c *CrewMember) CanDrive(qualifications []*Qualification) bool {
	return c.HasQualification(QualTypeDriverLicenseC, qualifications) ||
		c.HasQualification(QualTypeDriverLicenseCE, qualifications)
}

// CalculateHours calculates total hours worked in a time record
func (t *TimeRecord) CalculateHours() float64 {
	if t.EndTime == nil {
		return 0
	}

	duration := t.EndTime.Sub(t.StartTime)
	totalMinutes := int(duration.Minutes())
	workMinutes := totalMinutes - t.BreakMinutes
	return float64(workMinutes) / 60.0
}

// HasConflict checks if assignment conflicts with other assignments
func (a *CrewAssignment) HasConflict(otherAssignments []*CrewAssignment) bool {
	for _, other := range otherAssignments {
		if other.ID == a.ID {
			continue
		}
		if other.CrewMemberID != a.CrewMemberID {
			continue
		}
		if other.Status == AssignmentStatusCancelled {
			continue
		}

		// Check for overlap
		if !(a.EndDate.Before(other.StartDate) || a.StartDate.After(other.EndDate)) {
			return true
		}
	}

	return false
}
