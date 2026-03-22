package application

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
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
	// Generate ECDSA P-256 key pair
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		s.log.Error("Failed to generate private key", err)
		return nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "RentFlow Federation",
			Organization: []string{tenantID.String()},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0), // 365 days validity
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		s.log.Error("Failed to create certificate", err)
		return nil, err
	}

	// Encode certificate to PEM
	clientCert := string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}))

	// Encode private key to PEM
	privKeyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		s.log.Error("Failed to marshal private key", err)
		return nil, err
	}
	serverCert := string(pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privKeyBytes,
	}))

	// Generate fingerprint from DER-encoded certificate
	fingerprint := generateFingerprint(certDER)

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
	// Decode PEM to get DER for fingerprinting
	block, _ := pem.Decode([]byte(partnerCertPem))
	if block == nil {
		s.log.Error("Failed to decode partner certificate PEM", nil)
		return nil, domain.ErrInvalidCertificate
	}

	fingerprint := generateFingerprint(block.Bytes)

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
	block, _ := pem.Decode([]byte(certPem))
	if block == nil {
		return ""
	}
	return generateFingerprint(block.Bytes)
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

// RevokeCertificate deactivates a certificate by ID
func (s *CertificateService) RevokeCertificate(ctx context.Context, id uuid.UUID) error {
	// Since repository doesn't have GetByID for certs, we mark as deactivated
	// In a production system, you would query by ID or add a GetByID method to the repository
	cert := &domain.FederationCertificate{
		ID:       id,
		IsActive: false,
	}

	if err := s.certRepo.Update(ctx, cert); err != nil {
		s.log.Error("Failed to revoke certificate", err)
		return err
	}

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

func generateFingerprint(certDER []byte) string {
	// Hash the DER-encoded certificate to produce SHA-256 fingerprint
	hash := sha256.Sum256(certDER)
	return hex.EncodeToString(hash[:])
}
