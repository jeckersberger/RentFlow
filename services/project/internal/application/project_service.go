package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateProjectRequest struct {
	Name          string     `json:"name"`
	ProjectNumber string     `json:"project_number,omitempty"`
	Description   string     `json:"description,omitempty"`
	CustomerID    *uuid.UUID `json:"customer_id,omitempty"`
	ContactName   string     `json:"contact_name,omitempty"`
	ContactEmail  string     `json:"contact_email,omitempty"`
	ContactPhone  string     `json:"contact_phone,omitempty"`
	VenueName     string     `json:"venue_name,omitempty"`
	VenueAddress  string     `json:"venue_address,omitempty"`
	VenueLat      *float64   `json:"venue_lat,omitempty"`
	VenueLng      *float64   `json:"venue_lng,omitempty"`
	StartDate     *string    `json:"start_date,omitempty"`
	EndDate       *string    `json:"end_date,omitempty"`
	SetupDate     *string    `json:"setup_date,omitempty"`
	TeardownDate  *string    `json:"teardown_date,omitempty"`
	Color         string     `json:"color,omitempty"`
	Budget        int64      `json:"budget"`
	Currency      string     `json:"currency,omitempty"`
	ManagerID     *uuid.UUID `json:"manager_id,omitempty"`
	Notes         string     `json:"notes,omitempty"`
}

type UpdateProjectRequest struct {
	Name          *string    `json:"name,omitempty"`
	ProjectNumber *string    `json:"project_number,omitempty"`
	Description   *string    `json:"description,omitempty"`
	CustomerID    *uuid.UUID `json:"customer_id,omitempty"`
	ContactName   *string    `json:"contact_name,omitempty"`
	ContactEmail  *string    `json:"contact_email,omitempty"`
	ContactPhone  *string    `json:"contact_phone,omitempty"`
	VenueName     *string    `json:"venue_name,omitempty"`
	VenueAddress  *string    `json:"venue_address,omitempty"`
	VenueLat      *float64   `json:"venue_lat,omitempty"`
	VenueLng      *float64   `json:"venue_lng,omitempty"`
	StartDate     *string    `json:"start_date,omitempty"`
	EndDate       *string    `json:"end_date,omitempty"`
	SetupDate     *string    `json:"setup_date,omitempty"`
	TeardownDate  *string    `json:"teardown_date,omitempty"`
	Color         *string    `json:"color,omitempty"`
	Budget        *int64     `json:"budget,omitempty"`
	Currency      *string    `json:"currency,omitempty"`
	ManagerID     *uuid.UUID `json:"manager_id,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
}

type UpdateProjectStatusRequest struct {
	Status string `json:"status"`
}

type AddProjectEquipmentRequest struct {
	EquipmentID    uuid.UUID `json:"equipment_id"`
	Quantity       int       `json:"quantity"`
	AllocatedFrom  *string   `json:"allocated_from,omitempty"`
	AllocatedUntil *string   `json:"allocated_until,omitempty"`
	Notes          string    `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ProjectService struct {
	projectRepo   domain.ProjectRepository
	equipmentRepo domain.ProjectEquipmentRepository
	logger        zerolog.Logger
}

func NewProjectService(
	projectRepo domain.ProjectRepository,
	equipmentRepo domain.ProjectEquipmentRepository,
	logger zerolog.Logger,
) *ProjectService {
	return &ProjectService{
		projectRepo:   projectRepo,
		equipmentRepo: equipmentRepo,
		logger:        logger.With().Str("service", "project").Logger(),
	}
}

func (s *ProjectService) Create(ctx context.Context, tenantID uuid.UUID, req CreateProjectRequest) (*domain.Project, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	now := time.Now()
	project := &domain.Project{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Name:          req.Name,
		ProjectNumber: req.ProjectNumber,
		Description:   req.Description,
		Status:        domain.ProjectStatusDraft,
		CustomerID:    req.CustomerID,
		ContactName:   req.ContactName,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
		VenueName:     req.VenueName,
		VenueAddress:  req.VenueAddress,
		VenueLat:      req.VenueLat,
		VenueLng:      req.VenueLng,
		Color:         req.Color,
		Budget:        req.Budget,
		Currency:      req.Currency,
		ManagerID:     req.ManagerID,
		Notes:         req.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if project.Color == "" {
		project.Color = "#3b82f6"
	}
	if project.Currency == "" {
		project.Currency = "EUR"
	}

	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format (YYYY-MM-DD): %w", err)
		}
		project.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format (YYYY-MM-DD): %w", err)
		}
		project.EndDate = &t
	}
	if req.SetupDate != nil {
		t, err := time.Parse("2006-01-02", *req.SetupDate)
		if err != nil {
			return nil, fmt.Errorf("invalid setup_date format (YYYY-MM-DD): %w", err)
		}
		project.SetupDate = &t
	}
	if req.TeardownDate != nil {
		t, err := time.Parse("2006-01-02", *req.TeardownDate)
		if err != nil {
			return nil, fmt.Errorf("invalid teardown_date format (YYYY-MM-DD): %w", err)
		}
		project.TeardownDate = &t
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		s.logger.Error().Err(err).Str("tenant_id", tenantID.String()).Msg("failed to create project")
		return nil, fmt.Errorf("create project: %w", err)
	}

	s.logger.Info().Str("project_id", project.ID.String()).Str("tenant_id", tenantID.String()).Msg("project created")
	return project, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return project, nil
}

func (s *ProjectService) List(ctx context.Context, tenantID uuid.UUID, filter domain.ProjectFilter) ([]*domain.Project, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.projectRepo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	return items, total, nil
}

func (s *ProjectService) Update(ctx context.Context, id, tenantID uuid.UUID, req UpdateProjectRequest) (*domain.Project, error) {
	existing, err := s.projectRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update project – fetch: %w", err)
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.ProjectNumber != nil {
		existing.ProjectNumber = *req.ProjectNumber
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.CustomerID != nil {
		existing.CustomerID = req.CustomerID
	}
	if req.ContactName != nil {
		existing.ContactName = *req.ContactName
	}
	if req.ContactEmail != nil {
		existing.ContactEmail = *req.ContactEmail
	}
	if req.ContactPhone != nil {
		existing.ContactPhone = *req.ContactPhone
	}
	if req.VenueName != nil {
		existing.VenueName = *req.VenueName
	}
	if req.VenueAddress != nil {
		existing.VenueAddress = *req.VenueAddress
	}
	if req.VenueLat != nil {
		existing.VenueLat = req.VenueLat
	}
	if req.VenueLng != nil {
		existing.VenueLng = req.VenueLng
	}
	if req.Color != nil {
		existing.Color = *req.Color
	}
	if req.Budget != nil {
		existing.Budget = *req.Budget
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.ManagerID != nil {
		existing.ManagerID = req.ManagerID
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		existing.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		existing.EndDate = &t
	}
	if req.SetupDate != nil {
		t, err := time.Parse("2006-01-02", *req.SetupDate)
		if err != nil {
			return nil, fmt.Errorf("invalid setup_date format: %w", err)
		}
		existing.SetupDate = &t
	}
	if req.TeardownDate != nil {
		t, err := time.Parse("2006-01-02", *req.TeardownDate)
		if err != nil {
			return nil, fmt.Errorf("invalid teardown_date format: %w", err)
		}
		existing.TeardownDate = &t
	}

	existing.UpdatedAt = time.Now()

	if err := s.projectRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}

	s.logger.Info().Str("project_id", id.String()).Str("tenant_id", tenantID.String()).Msg("project updated")
	return existing, nil
}

func (s *ProjectService) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, req UpdateProjectStatusRequest) error {
	p := &domain.Project{}
	if !p.ValidateStatus(req.Status) {
		return domain.ErrInvalidProjectStatus
	}

	// Fetch current project to validate state transition.
	existing, err := s.projectRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return fmt.Errorf("update project status – fetch: %w", err)
	}

	if !existing.ValidateTransition(req.Status) {
		s.logger.Warn().
			Str("project_id", id.String()).
			Str("current_status", existing.Status).
			Str("requested_status", req.Status).
			Msg("invalid status transition")
		return domain.ErrInvalidStatusTransition
	}

	if err := s.projectRepo.UpdateStatus(ctx, id, tenantID, req.Status); err != nil {
		return fmt.Errorf("update project status: %w", err)
	}

	s.logger.Info().Str("project_id", id.String()).Str("status", req.Status).Msg("project status updated")
	return nil
}

func (s *ProjectService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	if err := s.projectRepo.Delete(ctx, id, tenantID); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	s.logger.Info().Str("project_id", id.String()).Msg("project deleted")
	return nil
}

func (s *ProjectService) Search(ctx context.Context, tenantID uuid.UUID, query string, page, perPage int) ([]*domain.Project, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	items, total, err := s.projectRepo.Search(ctx, tenantID, query, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("search projects: %w", err)
	}
	return items, total, nil
}

func (s *ProjectService) GetByDateRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*domain.Project, error) {
	items, err := s.projectRepo.GetByDateRange(ctx, tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get projects by date range: %w", err)
	}
	return items, nil
}

// --- Project Equipment ---

func (s *ProjectService) AddEquipment(ctx context.Context, tenantID, projectID uuid.UUID, req AddProjectEquipmentRequest) (*domain.ProjectEquipment, error) {
	// Verify project exists and belongs to tenant
	if _, err := s.projectRepo.GetByID(ctx, projectID, tenantID); err != nil {
		return nil, fmt.Errorf("add equipment – verify project: %w", err)
	}

	if req.Quantity < 1 {
		req.Quantity = 1
	}

	pe := &domain.ProjectEquipment{
		ID:          uuid.New(),
		ProjectID:   projectID,
		EquipmentID: req.EquipmentID,
		Quantity:    req.Quantity,
		Status:      domain.PEStatusPlanned,
		Notes:       req.Notes,
		CreatedAt:   time.Now(),
	}

	if req.AllocatedFrom != nil {
		t, err := time.Parse("2006-01-02", *req.AllocatedFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid allocated_from format: %w", err)
		}
		pe.AllocatedFrom = &t
	}
	if req.AllocatedUntil != nil {
		t, err := time.Parse("2006-01-02", *req.AllocatedUntil)
		if err != nil {
			return nil, fmt.Errorf("invalid allocated_until format: %w", err)
		}
		pe.AllocatedUntil = &t
	}

	if err := s.equipmentRepo.Add(ctx, pe); err != nil {
		return nil, fmt.Errorf("add project equipment: %w", err)
	}

	return pe, nil
}

func (s *ProjectService) RemoveEquipment(ctx context.Context, tenantID, id uuid.UUID) error {
	// Verify the equipment assignment belongs to a project owned by the tenant
	pe, err := s.equipmentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("remove project equipment – fetch: %w", err)
	}
	if _, err := s.projectRepo.GetByID(ctx, pe.ProjectID, tenantID); err != nil {
		return fmt.Errorf("remove project equipment – verify tenant: %w", err)
	}

	if err := s.equipmentRepo.Remove(ctx, id); err != nil {
		return fmt.Errorf("remove project equipment: %w", err)
	}
	return nil
}

func (s *ProjectService) ListEquipment(ctx context.Context, tenantID, projectID uuid.UUID) ([]*domain.ProjectEquipment, error) {
	// Verify project belongs to tenant
	if _, err := s.projectRepo.GetByID(ctx, projectID, tenantID); err != nil {
		return nil, fmt.Errorf("list project equipment – verify project: %w", err)
	}

	items, err := s.equipmentRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project equipment: %w", err)
	}
	return items, nil
}

// ---------------------------------------------------------------------------
// Calendar
// ---------------------------------------------------------------------------

// CalendarEvent is a lightweight projection of a project for calendar display.
type CalendarEvent struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Status    string    `json:"status"`
	Color     string    `json:"color"`
	Customer  string    `json:"customer"`
	VenueName string    `json:"venue_name"`
}

// GetCalendar returns projects within a date range as calendar events.
func (s *ProjectService) GetCalendar(ctx context.Context, tenantID uuid.UUID, from, to string) ([]CalendarEvent, error) {
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, fmt.Errorf("invalid from date (YYYY-MM-DD): %w", err)
	}
	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, fmt.Errorf("invalid to date (YYYY-MM-DD): %w", err)
	}

	projects, err := s.projectRepo.GetByDateRange(ctx, tenantID, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("get calendar: %w", err)
	}

	events := make([]CalendarEvent, 0, len(projects))
	for _, p := range projects {
		color := p.Color
		if color == "" {
			color = domain.StatusColor(p.Status)
		}

		ev := CalendarEvent{
			ID:        p.ID,
			Title:     p.Name,
			Status:    p.Status,
			Color:     color,
			Customer:  p.ContactName,
			VenueName: p.VenueName,
		}
		if p.StartDate != nil {
			ev.StartDate = p.StartDate.Format("2006-01-02")
		}
		if p.EndDate != nil {
			ev.EndDate = p.EndDate.Format("2006-01-02")
		}
		events = append(events, ev)
	}

	return events, nil
}

// icsStatus maps project statuses to iCalendar STATUS values.
func icsStatus(status string) string {
	switch status {
	case domain.ProjectStatusConfirmed, domain.ProjectStatusActive,
		domain.ProjectStatusInPreparation, domain.ProjectStatusCompleted,
		domain.ProjectStatusInvoiced:
		return "CONFIRMED"
	case domain.ProjectStatusInquiry, domain.ProjectStatusOfferSent:
		return "TENTATIVE"
	case domain.ProjectStatusCancelled:
		return "CANCELLED"
	default:
		return "CONFIRMED"
	}
}

// ExportICS returns an iCalendar (RFC 5545) representation of projects in a date range.
func (s *ProjectService) ExportICS(ctx context.Context, tenantID uuid.UUID, from, to string) (string, error) {
	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		return "", fmt.Errorf("invalid from date (YYYY-MM-DD): %w", err)
	}
	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		return "", fmt.Errorf("invalid to date (YYYY-MM-DD): %w", err)
	}

	projects, err := s.projectRepo.GetByDateRange(ctx, tenantID, fromTime, toTime)
	if err != nil {
		return "", fmt.Errorf("export ics: %w", err)
	}

	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//CrateDesk//Events//DE\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")

	for _, p := range projects {
		b.WriteString("BEGIN:VEVENT\r\n")
		b.WriteString(fmt.Sprintf("UID:%s@cratedesk\r\n", p.ID.String()))
		if p.StartDate != nil {
			b.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", p.StartDate.Format("20060102")))
		}
		if p.EndDate != nil {
			// iCal DATE end is exclusive, so add one day.
			endExclusive := p.EndDate.AddDate(0, 0, 1)
			b.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", endExclusive.Format("20060102")))
		}
		b.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeICS(p.Name)))
		if p.VenueName != "" {
			b.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeICS(p.VenueName)))
		}
		if p.Description != "" {
			b.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeICS(p.Description)))
		}
		b.WriteString(fmt.Sprintf("STATUS:%s\r\n", icsStatus(p.Status)))
		b.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", p.CreatedAt.UTC().Format("20060102T150405Z")))
		b.WriteString("END:VEVENT\r\n")
	}

	b.WriteString("END:VCALENDAR\r\n")
	return b.String(), nil
}

// escapeICS escapes special characters in iCalendar text values.
func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
