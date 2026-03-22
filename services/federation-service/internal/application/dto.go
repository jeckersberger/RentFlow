package application

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreatePartnerRequest struct {
	PartnerName     string   `json:"partner_name"`
	PartnerEndpoint string   `json:"partner_endpoint"`
	SharedCategories []string `json:"shared_categories"`
	DataPolicy      json.RawMessage `json:"data_policy"`
}

type PartnerResponse struct {
	ID              uuid.UUID   `json:"id"`
	TenantID        uuid.UUID   `json:"tenant_id"`
	PartnerName     string      `json:"partner_name"`
	PartnerEndpoint string      `json:"partner_endpoint"`
	Status          string      `json:"status"`
	TrustLevel      string      `json:"trust_level"`
	CertFingerprint string      `json:"cert_fingerprint"`
	CertExpiresAt   *time.Time  `json:"cert_expires_at"`
	SharedCategories []string   `json:"shared_categories"`
	CreatedAt       time.Time   `json:"created_at"`
}

type CreateSubRentalRequestRequest struct {
	PartnerID            uuid.UUID `json:"partner_id"`
	Direction            string    `json:"direction"`
	EquipmentCategory    string    `json:"equipment_category"`
	EquipmentDescription string    `json:"equipment_description"`
	Quantity             int       `json:"quantity"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	DailyRate            float64   `json:"daily_rate"`
	Notes                string    `json:"notes"`
}

type SubRentalRequestResponse struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	PartnerID            uuid.UUID  `json:"partner_id"`
	Direction            string     `json:"direction"`
	Status               string     `json:"status"`
	EquipmentCategory    string     `json:"equipment_category"`
	EquipmentDescription string     `json:"equipment_description"`
	Quantity             int        `json:"quantity"`
	StartDate            time.Time  `json:"start_date"`
	EndDate              time.Time  `json:"end_date"`
	DailyRate            float64    `json:"daily_rate"`
	TotalAmount          float64    `json:"total_amount"`
	HandoverDocumentID   *uuid.UUID `json:"handover_document_id"`
	InvoiceID            *uuid.UUID `json:"invoice_id"`
	CreatedAt            time.Time  `json:"created_at"`
}

type EquipmentCacheResponse struct {
	ID               uuid.UUID `json:"id"`
	PartnerID        uuid.UUID `json:"partner_id"`
	Category         string    `json:"category"`
	ItemName         string    `json:"item_name"`
	QuantityAvailable int      `json:"quantity_available"`
	DailyRate        float64   `json:"daily_rate"`
	LastSyncedAt     time.Time `json:"last_synced_at"`
}

type CertificateResponse struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	PartnerID   *uuid.UUID `json:"partner_id"`
	CertType    string     `json:"cert_type"`
	Fingerprint string     `json:"fingerprint"`
	IssuedAt    time.Time  `json:"issued_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type GenerateCertPairResponse struct {
	ClientCertPem string `json:"client_cert_pem"`
	ServerCertPem string `json:"server_cert_pem"`
	Fingerprint   string `json:"fingerprint"`
}

type DashboardResponse struct {
	PartnerCount    int     `json:"partner_count"`
	ActiveRentals   int     `json:"active_rentals"`
	TotalRevenue    float64 `json:"total_revenue"`
}
