package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/scanner/internal/domain"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// CreateSessionRequest holds the data for creating a scan session.
type CreateSessionRequest struct {
	SessionType string     `json:"session_type"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
}

// SignatureRequest holds the base64-encoded signature data.
type SignatureRequest struct {
	SignatureData string `json:"signature_data"`
}

// ResolveResult is the response from resolving an equipment identifier.
type ResolveResult struct {
	Found     bool        `json:"found"`
	Equipment interface{} `json:"equipment,omitempty"`
	Source    string      `json:"source,omitempty"`
}

// SessionProtocol holds a session with all its events.
type SessionProtocol struct {
	Session *domain.ScanSession `json:"session"`
	Events  []*domain.ScanEvent `json:"events"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// SessionService implements the application-level use cases for scan sessions.
type SessionService struct {
	sessionRepo      domain.ScanSessionRepository
	logger           zerolog.Logger
	inventoryBaseURL string
}

// NewSessionService constructs a new SessionService.
func NewSessionService(
	sessionRepo domain.ScanSessionRepository,
	logger zerolog.Logger,
	inventoryBaseURL string,
) *SessionService {
	return &SessionService{
		sessionRepo:      sessionRepo,
		logger:           logger.With().Str("service", "session").Logger(),
		inventoryBaseURL: inventoryBaseURL,
	}
}

// CreateSession creates a new scan session.
func (s *SessionService) CreateSession(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateSessionRequest,
) (*domain.ScanSession, error) {
	if req.SessionType == "" {
		req.SessionType = domain.SessionTypeCheckout
	}
	if !domain.ValidSessionType(req.SessionType) {
		return nil, domain.ErrInvalidSessionType
	}

	now := time.Now()
	session := &domain.ScanSession{
		ID:          uuid.New(),
		TenantID:    tenantID,
		SessionType: req.SessionType,
		ProjectID:   req.ProjectID,
		StartedBy:   userID,
		StartedAt:   now,
		ItemsCount:  0,
		Status:      domain.SessionStatusActive,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create scan session")
		return nil, fmt.Errorf("create session: %w", err)
	}

	s.logger.Info().
		Str("session_id", session.ID.String()).
		Str("tenant_id", tenantID.String()).
		Str("type", req.SessionType).
		Msg("scan session created")

	return session, nil
}

// EndSession marks a session as completed.
func (s *SessionService) EndSession(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.ScanSession, error) {
	if err := s.sessionRepo.End(ctx, id, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("session_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to end scan session")
		return nil, fmt.Errorf("end session: %w", err)
	}

	session, err := s.sessionRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("end session – fetch: %w", err)
	}

	s.logger.Info().
		Str("session_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Int("items_count", session.ItemsCount).
		Msg("scan session ended")

	return session, nil
}

// GetProtocol returns a session with all its scan events.
func (s *SessionService) GetProtocol(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*SessionProtocol, error) {
	session, err := s.sessionRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("session_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get session for protocol")
		return nil, fmt.Errorf("get protocol – session: %w", err)
	}

	events, err := s.sessionRepo.GetSessionEvents(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("session_id", id.String()).
			Msg("failed to get session events")
		return nil, fmt.Errorf("get protocol – events: %w", err)
	}

	return &SessionProtocol{
		Session: session,
		Events:  events,
	}, nil
}

// SaveSignature saves a base64 signature to a session.
func (s *SessionService) SaveSignature(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req SignatureRequest,
) error {
	if req.SignatureData == "" {
		return fmt.Errorf("signature_data is required")
	}

	if err := s.sessionRepo.UpdateSignature(ctx, id, tenantID, req.SignatureData); err != nil {
		s.logger.Error().Err(err).
			Str("session_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to save signature")
		return fmt.Errorf("save signature: %w", err)
	}

	s.logger.Info().
		Str("session_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("signature saved")

	return nil
}

// Resolve looks up equipment by a barcode, QR code, serial number, or RFID tag
// by calling the inventory service internally.
func (s *SessionService) Resolve(
	ctx context.Context,
	tenantID uuid.UUID,
	identifier string,
	authHeader string,
) (*ResolveResult, error) {
	if identifier == "" {
		return nil, fmt.Errorf("identifier is required")
	}

	// Try the inventory service's scan resolve endpoint.
	resolveURL := fmt.Sprintf(
		"%s/api/v1/scan/resolve/%s",
		s.inventoryBaseURL,
		url.PathEscape(identifier),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolveURL, nil)
	if err != nil {
		return nil, fmt.Errorf("resolve – create request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("X-Tenant-ID", tenantID.String())

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Warn().Err(err).
			Str("identifier", identifier).
			Str("url", resolveURL).
			Msg("failed to call inventory service for resolve")
		return &ResolveResult{Found: false}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &ResolveResult{Found: false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Warn().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("inventory service returned non-OK for resolve")
		return &ResolveResult{Found: false}, nil
	}

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		s.logger.Error().Err(err).Msg("failed to decode inventory resolve response")
		return &ResolveResult{Found: false}, nil
	}

	var equipment interface{}
	if err := json.Unmarshal(envelope.Data, &equipment); err != nil {
		return &ResolveResult{Found: false}, nil
	}

	return &ResolveResult{
		Found:     true,
		Equipment: equipment,
		Source:    "inventory",
	}, nil
}
