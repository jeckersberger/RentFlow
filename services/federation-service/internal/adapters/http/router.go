package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/federation-service/internal/application"
)

func NewRouter(
	peerSvc *application.PeerService,
	sharingSvc *application.SharingService,
	log logger.Logger,
) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(peerSvc, sharingSvc, log)

	// Health
	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// Peers
	router.HandleFunc("POST /api/v1/federation/peers", handler.CreatePeer)
	router.HandleFunc("GET /api/v1/federation/peers", handler.ListPeers)
	router.HandleFunc("GET /api/v1/federation/peers/{id}", handler.GetPeer)
	router.HandleFunc("PUT /api/v1/federation/peers/{id}", handler.UpdatePeer)
	router.HandleFunc("DELETE /api/v1/federation/peers/{id}", handler.DeletePeer)

	// Equipment sharing
	router.HandleFunc("POST /api/v1/federation/share", handler.ShareEquipment)
	router.HandleFunc("GET /api/v1/federation/catalog", handler.ListCatalog)

	// Share requests
	router.HandleFunc("GET /api/v1/federation/requests", handler.ListRequests)
	router.HandleFunc("POST /api/v1/federation/requests", handler.CreateShareRequest)
	router.HandleFunc("PUT /api/v1/federation/requests/{id}", handler.UpdateRequest)

	// Handshake
	router.HandleFunc("POST /api/v1/federation/handshake", handler.Handshake)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"federation-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"federation-service"}`))
}
