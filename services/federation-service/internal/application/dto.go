package application

import (
	"github.com/jeckersberger/rentflow/services/federation-service/internal/domain"
	"time"
)

type PeerDTO struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	PublicKey string     `json:"public_key"`
	Status    string     `json:"status"`
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type SharedEquipmentDTO struct {
	ID           string  `json:"id"`
	EquipmentID  string  `json:"equipment_id"`
	PeerID       string  `json:"peer_id"`
	Availability string  `json:"availability"`
	PricePerDay  float64 `json:"price_per_day"`
}

type ShareRequestDTO struct {
	ID          string    `json:"id"`
	FromPeerID  string    `json:"from_peer_id"`
	ToPeerID    string    `json:"to_peer_id"`
	EquipmentID string    `json:"equipment_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func PeerToDTO(p *domain.Peer) *PeerDTO {
	return &PeerDTO{
		ID:        p.ID,
		Name:      p.Name,
		URL:       p.URL,
		PublicKey: p.PublicKey,
		Status:    string(p.Status),
		LastSeen:  p.LastSeen,
		CreatedAt: p.CreatedAt,
	}
}

func SharedEquipmentToDTO(se *domain.SharedEquipment) *SharedEquipmentDTO {
	return &SharedEquipmentDTO{
		ID:           se.ID,
		EquipmentID:  se.EquipmentID,
		PeerID:       se.PeerID,
		Availability: se.Availability,
		PricePerDay:  se.PricePerDay,
	}
}

func ShareRequestToDTO(sr *domain.ShareRequest) *ShareRequestDTO {
	return &ShareRequestDTO{
		ID:          sr.ID,
		FromPeerID:  sr.FromPeerID,
		ToPeerID:    sr.ToPeerID,
		EquipmentID: sr.EquipmentID,
		StartDate:   sr.StartDate,
		EndDate:     sr.EndDate,
		Status:      string(sr.Status),
		CreatedAt:   sr.CreatedAt,
	}
}
