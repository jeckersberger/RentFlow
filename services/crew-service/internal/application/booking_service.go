package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/crew-service/internal/ports"
)

// BookingService handles booking request business logic
type BookingService struct {
	db              *sql.DB
	assignmentRepo  ports.CrewAssignmentRepository
	crewRepo        ports.CrewMemberRepository
	logger          logger.Logger
	notificationURL string
}

// NewBookingService creates a new booking service
func NewBookingService(
	db *sql.DB,
	assignmentRepo ports.CrewAssignmentRepository,
	crewRepo ports.CrewMemberRepository,
	log logger.Logger,
) *BookingService {
	return &BookingService{
		db:             db,
		assignmentRepo: assignmentRepo,
		crewRepo:       crewRepo,
		logger:         log,
	}
}

// SetNotificationURL sets the notification-service base URL for sending emails
func (s *BookingService) SetNotificationURL(url string) {
	s.notificationURL = url
}

// generateToken creates a cryptographically secure 32-byte hex token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateBookingRequest creates a new booking request for an assignment
func (s *BookingService) CreateBookingRequest(ctx context.Context, cmd CreateBookingRequestCommand) (*BookingRequestDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	// Get assignment
	assignment, err := s.assignmentRepo.FindByID(ctx, cmd.AssignmentID)
	if err != nil {
		s.logger.Error("failed to find assignment", err)
		return nil, err
	}
	if assignment == nil {
		return nil, domain.ErrAssignmentNotFound
	}

	// Get crew member
	member, err := s.crewRepo.FindByID(ctx, assignment.CrewMemberID)
	if err != nil {
		s.logger.Error("failed to find crew member", err)
		return nil, err
	}
	if member == nil {
		return nil, domain.ErrCrewMemberNotFound
	}

	// Generate token
	token, err := generateToken()
	if err != nil {
		s.logger.Error("failed to generate token", err)
		return nil, err
	}

	// Determine project ID
	projectID := ""
	if assignment.ProjectID != nil {
		projectID = *assignment.ProjectID
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	query := `
		INSERT INTO booking_requests
		(id, tenant_id, assignment_id, crew_member_id, project_id, token, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
	`
	_, err = s.db.ExecContext(ctx, query, id, cmd.TenantID, cmd.AssignmentID, assignment.CrewMemberID, projectID, token, now)
	if err != nil {
		s.logger.Error("failed to create booking request", err)
		return nil, err
	}

	dto := &BookingRequestDTO{
		ID:             id,
		AssignmentID:   cmd.AssignmentID,
		CrewMemberID:   assignment.CrewMemberID,
		CrewMemberName: member.FirstName + " " + member.LastName,
		ProjectID:      projectID,
		Token:          token,
		Status:         "pending",
		CreatedAt:      now.Format(time.RFC3339),
	}

	// Send booking request email asynchronously (fire-and-forget)
	if s.notificationURL != "" && member.Email != "" {
		projectDates := ""
		if !assignment.StartDate.IsZero() && !assignment.EndDate.IsZero() {
			projectDates = assignment.StartDate.Format("02.01.2006") + " - " + assignment.EndDate.Format("02.01.2006")
		}
		sendEmailAsync(s.notificationURL, "/api/v1/notifications/send-email/booking-request", map[string]string{
			"to":             member.Email,
			"recipient_name": member.FirstName + " " + member.LastName,
			"project_name":   "Projekt " + projectID,
			"dates":          projectDates,
			"link":           "http://localhost:3000/booking/" + token,
		}, s.logger)
	}

	return dto, nil
}

// GetBookingDetails retrieves booking details by token (public, no auth)
func (s *BookingService) GetBookingDetails(ctx context.Context, token string) (*BookingDetailsDTO, error) {
	query := `
		SELECT br.crew_member_id, br.project_id, br.status,
		       ca.role, ca.start_date, ca.end_date, ca.notes
		FROM booking_requests br
		JOIN crew_assignments ca ON br.assignment_id = ca.id
		WHERE br.token = $1
	`

	var crewMemberID, projectID, status, role, notes string
	var startDate, endDate time.Time

	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&crewMemberID, &projectID, &status, &role, &startDate, &endDate, &notes,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrBookingNotFound
	}
	if err != nil {
		s.logger.Error("failed to get booking details", err)
		return nil, err
	}

	// Get crew member name
	member, err := s.crewRepo.FindByID(ctx, crewMemberID)
	if err != nil {
		s.logger.Error("failed to find crew member", err)
		return nil, err
	}

	freelancerName := ""
	if member != nil {
		freelancerName = member.FirstName + " " + member.LastName
	}

	projectDates := startDate.Format("02.01.2006") + " - " + endDate.Format("02.01.2006")

	return &BookingDetailsDTO{
		ProjectName:    "Projekt " + projectID,
		ProjectDates:   projectDates,
		Role:           role,
		Location:       "",
		Message:        notes,
		FreelancerName: freelancerName,
		Status:         status,
	}, nil
}

// RespondToBooking processes a freelancer's response to a booking request
func (s *BookingService) RespondToBooking(ctx context.Context, cmd BookingResponseCommand) error {
	// Get the booking request
	query := `
		SELECT id, assignment_id, status
		FROM booking_requests
		WHERE token = $1
	`

	var bookingID, assignmentID, currentStatus string
	err := s.db.QueryRowContext(ctx, query, cmd.Token).Scan(&bookingID, &assignmentID, &currentStatus)
	if err == sql.ErrNoRows {
		return domain.ErrBookingNotFound
	}
	if err != nil {
		s.logger.Error("failed to get booking request", err)
		return err
	}

	if currentStatus != "pending" {
		return domain.ErrBookingAlreadyResponded
	}

	now := time.Now().UTC()

	// Update booking request
	updateQuery := `
		UPDATE booking_requests
		SET status = $1, response_message = $2, responded_at = $3
		WHERE token = $4
	`
	_, err = s.db.ExecContext(ctx, updateQuery, cmd.Status, cmd.Message, now, cmd.Token)
	if err != nil {
		s.logger.Error("failed to update booking request", err)
		return err
	}

	// Update assignment status based on response
	var assignmentStatus string
	switch cmd.Status {
	case "accepted":
		assignmentStatus = "confirmed"
	case "declined":
		assignmentStatus = "cancelled"
	case "alternative":
		// Keep as planned, but the message will contain the alternative proposal
		assignmentStatus = "planned"
	default:
		return domain.ErrInvalidStatus
	}

	assignmentUpdateQuery := `
		UPDATE crew_assignments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err = s.db.ExecContext(ctx, assignmentUpdateQuery, assignmentStatus, now, assignmentID)
	if err != nil {
		s.logger.Error("failed to update assignment status", err)
		return err
	}

	return nil
}

// ListBookingRequests lists all booking requests for a tenant
func (s *BookingService) ListBookingRequests(ctx context.Context, tenantID string) ([]*BookingRequestDTO, error) {
	query := `
		SELECT br.id, br.assignment_id, br.crew_member_id, br.project_id,
		       br.token, br.status, br.response_message, br.responded_at, br.created_at,
		       cm.first_name, cm.last_name
		FROM booking_requests br
		JOIN crew_members cm ON br.crew_member_id = cm.id
		WHERE br.tenant_id = $1
		ORDER BY br.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		s.logger.Error("failed to list booking requests", err)
		return nil, err
	}
	defer rows.Close()

	var results []*BookingRequestDTO
	for rows.Next() {
		var dto BookingRequestDTO
		var firstName, lastName string
		var responseMessage sql.NullString
		var respondedAt sql.NullTime
		var createdAt time.Time

		if err := rows.Scan(
			&dto.ID, &dto.AssignmentID, &dto.CrewMemberID, &dto.ProjectID,
			&dto.Token, &dto.Status, &responseMessage, &respondedAt, &createdAt,
			&firstName, &lastName,
		); err != nil {
			s.logger.Error("failed to scan booking request", err)
			return nil, err
		}

		dto.CrewMemberName = firstName + " " + lastName
		dto.CreatedAt = createdAt.Format(time.RFC3339)

		if responseMessage.Valid {
			dto.ResponseMessage = &responseMessage.String
		}
		if respondedAt.Valid {
			t := respondedAt.Time.Format(time.RFC3339)
			dto.RespondedAt = &t
		}

		results = append(results, &dto)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("error iterating booking requests", err)
		return nil, err
	}

	if results == nil {
		results = []*BookingRequestDTO{}
	}

	return results, nil
}
