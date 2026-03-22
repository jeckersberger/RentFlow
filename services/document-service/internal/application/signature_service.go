package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/document-service/internal/ports"
)

type SignatureService struct {
	docRepo  ports.DocumentRepository
	sigRepo  ports.SignatureRepository
	logger   logger.Logger
}

func NewSignatureService(
	docRepo ports.DocumentRepository,
	sigRepo ports.SignatureRepository,
	log logger.Logger,
) *SignatureService {
	return &SignatureService{
		docRepo:  docRepo,
		sigRepo:  sigRepo,
		logger:   log,
	}
}

// RequestSignature creates a new signature request for a document
func (s *SignatureService) RequestSignature(
	ctx context.Context,
	tenantID, docID string,
	req RequestSignatureRequest,
) (*SignatureResponse, error) {
	if tenantID == "" || docID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.SignerName == "" || req.SignerEmail == "" {
		return nil, domain.ErrInvalidInput
	}

	// Verify document exists
	doc, err := s.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}

	// Check if signature already exists for this email
	existing, err := s.sigRepo.GetByDocumentAndEmail(ctx, docID, req.SignerEmail)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.SignedAt != nil {
		return nil, domain.ErrDocumentAlreadySigned
	}

	// Create new signature request
	sig := &domain.Signature{
		ID:          uuid.New().String(),
		DocumentID:  docID,
		SignerName:  req.SignerName,
		SignerEmail: req.SignerEmail,
		SignerRole:  req.SignerRole,
		Verified:    false,
		CreatedAt:   time.Now(),
	}

	if err := s.sigRepo.Create(ctx, sig); err != nil {
		s.logger.Error("Failed to create signature request", err)
		return nil, err
	}

	return SignatureToResponse(sig), nil
}

// SubmitSignature submits a canvas signature for a document
func (s *SignatureService) SubmitSignature(
	ctx context.Context,
	tenantID, docID, sigID string,
	req SubmitSignatureRequest,
) (*SignatureResponse, error) {
	if tenantID == "" || docID == "" || sigID == "" {
		return nil, domain.ErrTenantIDRequired
	}
	if req.SignatureData == "" {
		return nil, domain.ErrInvalidInput
	}

	// Verify document exists
	doc, err := s.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}

	// Get the signature request
	sig, err := s.sigRepo.GetByID(ctx, sigID)
	if err != nil {
		return nil, err
	}
	if sig == nil {
		return nil, domain.ErrSignatureNotFound
	}

	if sig.DocumentID != docID {
		return nil, domain.ErrInvalidInput
	}

	// Update signature with submitted data
	now := time.Now()
	sig.SignatureData = req.SignatureData
	sig.SignedAt = &now
	sig.IPAddress = req.IPAddress
	sig.UserAgent = req.UserAgent
	sig.Verified = true

	if err := s.sigRepo.Update(ctx, sig); err != nil {
		s.logger.Error("Failed to update signature", err)
		return nil, err
	}

	return SignatureToResponse(sig), nil
}

// GetSignatures retrieves all signatures for a document
func (s *SignatureService) GetSignatures(
	ctx context.Context,
	tenantID, docID string,
) ([]*SignatureResponse, error) {
	if tenantID == "" || docID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	// Verify document exists
	doc, err := s.docRepo.GetByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, domain.ErrDocumentNotFound
	}

	sigs, err := s.sigRepo.ListByDocument(ctx, docID)
	if err != nil {
		s.logger.Error("Failed to fetch signatures", err)
		return nil, err
	}

	responses := make([]*SignatureResponse, len(sigs))
	for i, sig := range sigs {
		responses[i] = SignatureToResponse(sig)
	}

	return responses, nil
}
