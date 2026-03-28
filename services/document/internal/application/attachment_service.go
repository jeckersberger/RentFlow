package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/document/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateAttachmentRequest holds the data needed to create an attachment.
type CreateAttachmentRequest struct {
	ReferenceID   uuid.UUID `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	FileName      string    `json:"file_name"`
	FilePath      string    `json:"file_path"`
	FileSize      int       `json:"file_size,omitempty"`
	MimeType      string    `json:"mime_type,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// AttachmentService implements the application-level use cases for attachments.
type AttachmentService struct {
	attachmentRepo domain.AttachmentRepository
	logger         zerolog.Logger
}

// NewAttachmentService constructs a new AttachmentService.
func NewAttachmentService(
	attachmentRepo domain.AttachmentRepository,
	logger zerolog.Logger,
) *AttachmentService {
	return &AttachmentService{
		attachmentRepo: attachmentRepo,
		logger:         logger.With().Str("service", "attachment").Logger(),
	}
}

// Create creates a new attachment.
func (s *AttachmentService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	uploadedBy *uuid.UUID,
	req CreateAttachmentRequest,
) (*domain.Attachment, error) {
	if req.ReferenceID == uuid.Nil || req.ReferenceType == "" {
		return nil, domain.ErrReferenceRequired
	}
	if req.FileName == "" {
		return nil, domain.ErrFileNameRequired
	}
	if req.FilePath == "" {
		return nil, domain.ErrFilePathRequired
	}

	att := &domain.Attachment{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
		FileName:      req.FileName,
		FilePath:      req.FilePath,
		FileSize:      req.FileSize,
		MimeType:      req.MimeType,
		UploadedBy:    uploadedBy,
	}

	if err := s.attachmentRepo.Create(ctx, att); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create attachment")
		return nil, fmt.Errorf("create attachment: %w", err)
	}

	s.logger.Info().
		Str("attachment_id", att.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("attachment created")

	return att, nil
}

// GetByID retrieves an attachment by ID.
func (s *AttachmentService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.Attachment, error) {
	att, err := s.attachmentRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("attachment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get attachment")
		return nil, fmt.Errorf("get attachment: %w", err)
	}
	return att, nil
}

// ListByReference returns all attachments for a given reference.
func (s *AttachmentService) ListByReference(
	ctx context.Context,
	tenantID uuid.UUID,
	referenceID uuid.UUID,
	referenceType string,
) ([]*domain.Attachment, error) {
	attachments, err := s.attachmentRepo.ListByReference(ctx, tenantID, referenceID, referenceType)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Str("reference_id", referenceID.String()).
			Msg("failed to list attachments")
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	return attachments, nil
}

// Delete deletes an attachment by ID.
func (s *AttachmentService) Delete(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) error {
	if err := s.attachmentRepo.Delete(ctx, id, tenantID); err != nil {
		s.logger.Error().Err(err).
			Str("attachment_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to delete attachment")
		return fmt.Errorf("delete attachment: %w", err)
	}

	s.logger.Info().
		Str("attachment_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("attachment deleted")

	return nil
}
