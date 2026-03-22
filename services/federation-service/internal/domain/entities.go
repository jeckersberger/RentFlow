package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PartnerStatus represents partner status
type PartnerStatus string

const (
	PartnerStatusPending   PartnerStatus = "pending"
	PartnerStatusActive    PartnerStatus = "active"
	PartnerStatusSuspended PartnerStatus = "suspended"
	PartnerStatusRevoked   PartnerStatus = "revoked"
)

// TrustLevel represents trust level in federation
type TrustLevel string

const (
	TrustLevelBasic    TrustLevel = "basic"
	TrustLevelVerified TrustLevel = "verified"
	TrustLevelTrusted  TrustLevel = "trusted"
)

// RequestStatus represents sub-rental request status
type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "pending"
	RequestStatusAccepted  RequestStatus = "accepted"
	RequestStatusRejected  RequestStatus = "rejected"
	RequestStatusCompleted RequestStatus = "completed"
	RequestStatusCancelled RequestStatus = "cancelled"
)

// RequestDirection represents if request is inbound or outbound
type RequestDirection string

const (
	RequestDirectionInbound  RequestDirection = "inbound"
	RequestDirectionOutbound RequestDirection = "outbound"
)

// FederationPartner represents a partner in the federation
type FederationPartner struct {
	ID              uuid.UUID      `json:"id"`
	TenantID        uuid.UUID      `json:"tenant_id"`
	PartnerName     string         `json:"partner_name"`
	PartnerEndpoint string         `json:"partner_endpoint"`
	Status          PartnerStatus  `json:"status"`
	TrustLevel      TrustLevel     `json:"trust_level"`
	CertFingerprint string         `json:"cert_fingerprint"`
	CertExpiresAt   *time.Time     `json:"cert_expires_at"`
	SharedCategories []string      `json:"shared_categories"`
	DataPolicy      json.RawMessage `json:"data_policy"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// SubRentalRequest represents a sub-rental request between partners
type SubRentalRequest struct {
	ID                   uuid.UUID       `json:"id"`
	TenantID             uuid.UUID       `json:"tenant_id"`
	PartnerID            uuid.UUID       `json:"partner_id"`
	Direction            RequestDirection `json:"direction"`
	Status               RequestStatus   `json:"status"`
	EquipmentCategory    string          `json:"equipment_category"`
	EquipmentDescription string          `json:"equipment_description"`
	Quantity             int             `json:"quantity"`
	StartDate            time.Time       `json:"start_date"`
	EndDate              time.Time       `json:"end_date"`
	DailyRate            float64         `json:"daily_rate"`
	TotalAmount          float64         `json:"total_amount"`
	HandoverDocumentID   *uuid.UUID      `json:"handover_document_id"`
	InvoiceID            *uuid.UUID      `json:"invoice_id"`
	Notes                string          `json:"notes"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// PartnerEquipmentCache caches partner equipment availability
type PartnerEquipmentCache struct {
	ID              uuid.UUID `json:"id"`
	PartnerID       uuid.UUID `json:"partner_id"`
	Category        string    `json:"category"`
	ItemName        string    `json:"item_name"`
	QuantityAvailable int     `json:"quantity_available"`
	DailyRate       float64   `json:"daily_rate"`
	LastSyncedAt    time.Time `json:"last_synced_at"`
}

// FederationCertificate represents mTLS certificates
type FederationCertificate struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	PartnerID   *uuid.UUID `json:"partner_id"`
	CertType    string     `json:"cert_type"` // client, server, ca
	CertPem     string     `json:"cert_pem"`
	KeyPemEncr  string     `json:"key_pem_encrypted"`
	Fingerprint string     `json:"fingerprint"`
	IssuedAt    time.Time  `json:"issued_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Implement driver.Valuer for custom types
func (p PartnerStatus) Value() (driver.Value, error) {
	return string(p), nil
}

func (t TrustLevel) Value() (driver.Value, error) {
	return string(t), nil
}

func (r RequestStatus) Value() (driver.Value, error) {
	return string(r), nil
}

func (r RequestDirection) Value() (driver.Value, error) {
	return string(r), nil
}
