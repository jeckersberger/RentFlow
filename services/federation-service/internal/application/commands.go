package application

import "time"

type CreatePeerCommand struct {
	TenantID  string `json:"tenant_id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	PublicKey string `json:"public_key"`
}

type UpdatePeerCommand struct {
	TenantID string `json:"tenant_id"`
	PeerID   string `json:"peer_id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Status   string `json:"status"`
}

type CreateShareRequestCommand struct {
	TenantID    string    `json:"tenant_id"`
	FromPeerID  string    `json:"from_peer_id"`
	ToPeerID    string    `json:"to_peer_id"`
	EquipmentID string    `json:"equipment_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

type ApproveShareRequestCommand struct {
	TenantID  string `json:"tenant_id"`
	RequestID string `json:"request_id"`
}

type RejectShareRequestCommand struct {
	TenantID  string `json:"tenant_id"`
	RequestID string `json:"request_id"`
}

type ShareEquipmentCommand struct {
	TenantID    string  `json:"tenant_id"`
	EquipmentID string  `json:"equipment_id"`
	PeerID      string  `json:"peer_id"`
	Availability string  `json:"availability"`
	PricePerDay float64 `json:"price_per_day"`
}
