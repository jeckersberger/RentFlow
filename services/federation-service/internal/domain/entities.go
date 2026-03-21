package domain

import "time"

type PeerStatus string

const (
	PeerStatusPending PeerStatus = "pending"
	PeerStatusActive  PeerStatus = "active"
	PeerStatusBlocked PeerStatus = "blocked"
)

type ShareRequestStatus string

const (
	ShareRequestStatusPending  ShareRequestStatus = "pending"
	ShareRequestStatusApproved ShareRequestStatus = "approved"
	ShareRequestStatusRejected ShareRequestStatus = "rejected"
)

type Peer struct {
	ID        string     `json:"id"`
	TenantID  string     `json:"tenant_id"`
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	PublicKey string     `json:"public_key"`
	Status    PeerStatus `json:"status"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type SharedEquipment struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	EquipmentID  string    `json:"equipment_id"`
	PeerID       string    `json:"peer_id"`
	Availability string    `json:"availability"`
	PricePerDay  float64   `json:"price_per_day"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ShareRequest struct {
	ID          string             `json:"id"`
	TenantID    string             `json:"tenant_id"`
	FromPeerID  string             `json:"from_peer_id"`
	ToPeerID    string             `json:"to_peer_id"`
	EquipmentID string             `json:"equipment_id"`
	StartDate   time.Time          `json:"start_date"`
	EndDate     time.Time          `json:"end_date"`
	Status      ShareRequestStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}
