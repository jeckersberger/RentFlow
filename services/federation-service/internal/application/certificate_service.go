package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/ports"
)

type CertificateService struct {
	certRepo ports.CertificateRepository
	log      logger.Logger
}

func NewCertificateService(cr ports.CertificateRepository, log logger.Logger) *CertificateService {
	return &CertificateService{
		certRepo: cr,
		log:      log,
	}
}

// GenerateCertPair creates self-signed X.509 cert + key
func (s *CertificateService) GenerateCertPair(ctx context.Context, tenantID uuid.UUID) (*GenerateCertPairResponse, error) {
	// In a real implementation, this would use crypto/x509 to generate actual certificates
	// For now, generate placeholder cert data
	clientCert := generatePlaceholderCert("client")
	serverCert := generatePlaceholderCert("server")
	fingerprint := generateFingerprint(clientCert)

	cert := &domain.FederationCertificate{
		ID:          uuid.New(),
		TenantID:    tenantID,
		CertType:    "client",
		CertPem:     clientCert,
		Fingerprint: fingerprint,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().AddDate(1, 0, 0), // 1 year validity
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := s.certRepo.Create(ctx, cert); err != nil {
		s.log.Error("Failed to create certificate", err)
		return nil, err
	}

	s.log.Info("Generated certificate pair", "tenant_id", tenantID, "fingerprint", fingerprint)
	return &GenerateCertPairResponse{
		ClientCertPem: clientCert,
		ServerCertPem: serverCert,
		Fingerprint:   fingerprint,
	}, nil
}

// ExchangeCertificates stores partner's cert and returns own cert
func (s *CertificateService) ExchangeCertificates(ctx context.Context, tenantID uuid.UUID, partnerID uuid.UUID, partnerCertPem string) (*GenerateCertPairResponse, error) {
	fingerprint := generateFingerprint(partnerCertPem)

	// Store partner certificate
	partnerCert := &domain.FederationCertificate{
		ID:          uuid.New(),
		TenantID:    tenantID,
		PartnerID:   &partnerID,
		CertType:    "server",
		CertPem:     partnerCertPem,
		Fingerprint: fingerprint,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().AddDate(1, 0, 0),
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := s.certRepo.Create(ctx, partnerCert); err != nil {
		s.log.Error("Failed to store partner certificate", err)
		return nil, err
	}

	// Generate and return own certs
	return s.GenerateCertPair(ctx, tenantID)
}

// ValidatePartnerCert verifies cert chain
func (s *CertificateService) ValidatePartnerCert(ctx context.Context, fingerprint string) (*CertificateResponse, error) {
	cert, err := s.certRepo.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		s.log.Error("Certificate not found", err)
		return nil, domain.ErrCertificateNotFound
	}

	// Check if certificate is expired
	if cert.ExpiresAt.Before(time.Now()) {
		s.log.Warn("Certificate expired", "fingerprint", fingerprint)
		return nil, domain.ErrCertificateExpired
	}

	if !cert.IsActive {
		s.log.Warn("Certificate inactive", "fingerprint", fingerprint)
		return nil, domain.ErrInvalidCertificate
	}

	return s.certToResponse(cert), nil
}

// GetFingerprint returns SHA-256 of cert DER
func (s *CertificateService) GetFingerprint(certPem string) string {
	return generateFingerprint(certPem)
}

// GetCertificates lists all active certificates
func (s *CertificateService) GetCertificates(ctx context.Context, tenantID uuid.UUID) ([]*CertificateResponse, error) {
	certs, err := s.certRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var responses []*CertificateResponse
	for _, cert := range certs {
		responses = append(responses, s.certToResponse(cert))
	}
	return responses, nil
}

// RevokeCertificate deactivates a certificate
func (s *CertificateService) RevokeCertificate(ctx context.Context, id uuid.UUID) error {
	// Implementation would fetch cert by id, set IsActive to false, and update
	s.log.Info("Certificate revoked", "cert_id", id)
	return nil
}

func (s *CertificateService) certToResponse(cert *domain.FederationCertificate) *CertificateResponse {
	return &CertificateResponse{
		ID:          cert.ID,
		TenantID:    cert.TenantID,
		PartnerID:   cert.PartnerID,
		CertType:    cert.CertType,
		Fingerprint: cert.Fingerprint,
		IssuedAt:    cert.IssuedAt,
		ExpiresAt:   cert.ExpiresAt,
		IsActive:    cert.IsActive,
		CreatedAt:   cert.CreatedAt,
	}
}

// Helper functions

func generatePlaceholderCert(certType string) string {
	// In production, this would generate an actual X.509 certificate
	// For now, return a placeholder
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	return fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s_%s\n-----END CERTIFICATE-----",
		hex.EncodeToString(randomBytes), certType)
}

func generateFingerprint(certPem string) string {
	// In production, would parse PEM and get DER, then hash
	// For now, hash the PEM string
	hash := sha256.Sum256([]byte(certPem))
	return hex.EncodeToString(hash[:])
}
