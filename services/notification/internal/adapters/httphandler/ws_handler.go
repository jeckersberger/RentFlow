package httphandler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/services/notification/internal/domain"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, restrict origins via CORS configuration.
		return true
	},
}

// wsClient represents a single WebSocket connection.
type wsClient struct {
	hub      *WSHub
	conn     *websocket.Conn
	send     chan []byte
	userID   uuid.UUID
	tenantID uuid.UUID
}

// WSHub maintains the set of active WebSocket clients and broadcasts
// notifications to the appropriate user clients.
type WSHub struct {
	mu         sync.RWMutex
	clients    map[*wsClient]bool
	register   chan *wsClient
	unregister chan *wsClient
	logger     zerolog.Logger
}

// NewWSHub creates a new WebSocket hub.
func NewWSHub(logger zerolog.Logger) *WSHub {
	return &WSHub{
		clients:    make(map[*wsClient]bool),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
		logger:     logger.With().Str("component", "ws_hub").Logger(),
	}
}

// Run starts the hub's event loop. Should be run as a goroutine.
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Debug().
				Str("user_id", client.userID.String()).
				Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Debug().
				Str("user_id", client.userID.String()).
				Msg("WebSocket client disconnected")
		}
	}
}

// BroadcastToUser sends a notification payload to all WebSocket connections
// belonging to the specified tenant and user.
func (h *WSHub) BroadcastToUser(tenantID, userID uuid.UUID, notification *domain.Notification) {
	data, err := json.Marshal(notification)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to marshal notification for WebSocket")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.tenantID == tenantID && client.userID == userID {
			select {
			case client.send <- data:
			default:
				// Client send buffer full, skip.
				h.logger.Warn().
					Str("user_id", userID.String()).
					Msg("WebSocket client send buffer full, dropping message")
			}
		}
	}
}

// WSHandler handles WebSocket upgrade requests.
type WSHandler struct {
	hub    *WSHub
	logger zerolog.Logger
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *WSHub, logger zerolog.Logger) *WSHandler {
	return &WSHandler{
		hub:    hub,
		logger: logger.With().Str("handler", "websocket").Logger(),
	}
}

// ServeWS upgrades the HTTP connection to a WebSocket connection.
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	client := &wsClient{
		hub:      h.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   claims.UserID,
		tenantID: claims.TenantID,
	}

	h.hub.register <- client

	// Start read and write pumps in separate goroutines.
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub.
// It reads pong messages to keep the connection alive.
func (c *wsClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.hub.logger.Warn().Err(err).Msg("WebSocket unexpected close")
			}
			break
		}
		// We don't process incoming messages from the client;
		// this is a server-push-only WebSocket.
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
