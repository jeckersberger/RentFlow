package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/document-service/internal/ports"
)

type ChecksumService struct {
	docRepo ports.DocumentRepository
	logger  logger.Logger
}

func NewChecksumService(docRepo ports.DocumentRepository, log logger.Logger) *ChecksumService {
	return &ChecksumService{
		docRepo: docRepo,
		logger:  log,
	}
}

// GenerateChecksum creates a SHA-256 hash for a document including the previous checksum (GoBD chain)
func (s *ChecksumService) GenerateChecksum(doc *domain.Document) string {
	hash := sha256.New()

	// Include document fields in the hash
	hash.Write([]byte(doc.ID))
	hash.Write([]byte(doc.TenantID))
	hash.Write([]byte(doc.DocumentType))
	hash.Write([]byte(doc.ReferenceID))
	hash.Write([]byte(doc.DocumentNumber))
	hash.Write([]byte(doc.Title))
	hash.Write([]byte(doc.Status))
	hash.Write([]byte(doc.CreatedBy))
	hash.Write([]byte(doc.CreatedAt.String()))

	// Include metadata deterministically
	if doc.Metadata != nil {
		keys := make([]string, 0, len(doc.Metadata))
		for k := range doc.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			hash.Write([]byte(fmt.Sprintf("%v", doc.Metadata[k])))
		}
	}

	// Include previous checksum to create the chain
	if doc.PreviousChecksum != "" {
		hash.Write([]byte(doc.PreviousChecksum))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// VerifyChecksumChain walks through all documents for a tenant and verifies their checksum chain integrity
func (s *ChecksumService) VerifyChecksumChain(ctx context.Context, tenantID string) (*ChecksumChainVerificationResponse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantIDRequired
	}

	docs, err := s.docRepo.ListAllForChecksumChain(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to fetch documents for checksum verification", err)
		return nil, err
	}

	// Sort by creation time to verify chain
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].CreatedAt.Before(docs[j].CreatedAt)
	})

	var errors []string
	validCount := 0

	for i, doc := range docs {
		// Recalculate the checksum
		calculatedChecksum := s.GenerateChecksum(doc)

		if calculatedChecksum != doc.ChecksumSHA256 {
			errorMsg := fmt.Sprintf("Document %s (v%d) checksum mismatch: expected %s, got %s",
				doc.ID, doc.CurrentVersion, doc.ChecksumSHA256, calculatedChecksum)
			errors = append(errors, errorMsg)
			s.logger.Error(errorMsg)
		} else {
			validCount++
		}

		// Verify chain continuity for subsequent documents
		if i > 0 {
			prevDoc := docs[i-1]
			if doc.PreviousChecksum != prevDoc.ChecksumSHA256 {
				errorMsg := fmt.Sprintf("Document %s chain integrity broken: previous checksum mismatch", doc.ID)
				errors = append(errors, errorMsg)
				s.logger.Error(errorMsg)
			}
		}
	}

	integrity := len(errors) == 0
	status := "valid"
	if !integrity {
		status = "invalid"
	}

	return &ChecksumChainVerificationResponse{
		Status:         status,
		DocumentCount:  len(docs),
		IntegrityValid: integrity,
		Errors:         errors,
	}, nil
}
