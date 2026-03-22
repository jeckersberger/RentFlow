package http

import (
	nethttp "net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/application"
)

func SetupRoutes(
	router *nethttp.ServeMux,
	partnerService *application.PartnerService,
	sharingService *application.SharingService,
	certificateService *application.CertificateService,
	log logger.Logger,
) {
	h := NewHandlers(partnerService, sharingService, certificateService, log)

	// Federation Partners
	router.HandleFunc("POST /api/v1/federation/partners", h.CreatePartner)
	router.HandleFunc("GET /api/v1/federation/partners", h.ListPartners)
	router.HandleFunc("GET /api/v1/federation/partners/{id}", h.GetPartner)
	router.HandleFunc("POST /api/v1/federation/partners/{id}/activate", h.ActivatePartner)
	router.HandleFunc("POST /api/v1/federation/partners/{id}/suspend", h.SuspendPartner)

	// Partner Equipment
	router.HandleFunc("GET /api/v1/federation/partners/{id}/equipment", h.GetPartnerEquipment)
	router.HandleFunc("POST /api/v1/federation/partners/{id}/sync", h.SyncPartnerEquipment)

	// Sub-Rental Requests
	router.HandleFunc("POST /api/v1/federation/requests", h.CreateRequest)
	router.HandleFunc("GET /api/v1/federation/requests", h.ListRequests)
	router.HandleFunc("GET /api/v1/federation/requests/{id}", h.GetRequest)
	router.HandleFunc("POST /api/v1/federation/requests/{id}/accept", h.AcceptRequest)
	router.HandleFunc("POST /api/v1/federation/requests/{id}/reject", h.RejectRequest)
	router.HandleFunc("POST /api/v1/federation/requests/{id}/complete", h.CompleteRequest)

	// Certificates
	router.HandleFunc("GET /api/v1/federation/certificates", h.ListCertificates)
	router.HandleFunc("POST /api/v1/federation/certificates/generate", h.GenerateCertPair)
}
