package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/cache"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// Session represents a user session stored in Redis
type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TenantID   string    `json:"tenant_id"`
	Roles      []string  `json:"roles"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Rotations  int       `json:"rotations"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// UserSessionIndex stores all session IDs for a user
type UserSessionIndex struct {
	SessionIDs []string `json:"session_ids"`
}

// SessionManager handles session lifecycle with Redis
type SessionManager struct {
	cache  cache.Cache
	logger logger.Logger
}

const (
	sessionKeyPrefix      = "session:"
	userSessionsKeyPrefix = "user_sessions:"
	bruteforceKeyPrefix   = "bruteforce:"
	sessionTTL            = 7 * 24 * time.Hour  // 7 Tage
	maxRotations          = 50
	maxFailedLogins       = 10
)

// NewSessionManager erstellt einen neuen SessionManager
func NewSessionManager(cache cache.Cache, logger logger.Logger) *SessionManager {
	return &SessionManager{
		cache:  cache,
		logger: logger,
	}
}

// CreateSession erstellt eine neue Session und speichert sie in Redis
func (sm *SessionManager) CreateSession(ctx context.Context, userID, tenantID string, roles []string, ipAddress, userAgent string) (*Session, error) {
	if userID == "" || tenantID == "" {
		return nil, fmt.Errorf("userID und tenantID sind erforderlich")
	}

	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:         sessionID,
		UserID:     userID,
		TenantID:   tenantID,
		Roles:      roles,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Rotations:  0,
		CreatedAt:  now,
		LastUsedAt: now,
	}

	// Session als JSON serialisieren und in Redis speichern
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("session serialisierung fehlgeschlagen: %w", err)
	}

	sessionKey := sessionKeyPrefix + sessionID
	if err := sm.cache.Set(ctx, sessionKey, string(sessionJSON), sessionTTL); err != nil {
		return nil, fmt.Errorf("session speichern fehlgeschlagen: %w", err)
	}

	// Session-ID zum User-Sessions-Index hinzufuegen
	sm.addToUserSessions(ctx, userID, sessionID)

	sm.logger.Info("Session erstellt", "sessionID", sessionID, "userID", userID)
	return session, nil
}

// GetSession ruft eine Session aus Redis ab
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID ist erforderlich")
	}

	sessionKey := sessionKeyPrefix + sessionID
	sessionJSON, err := sm.cache.Get(ctx, sessionKey)
	if err != nil {
		return nil, fmt.Errorf("session nicht gefunden: %w", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("session deserialisierung fehlgeschlagen: %w", err)
	}

	return &session, nil
}

// RefreshSession verlaengert die TTL und zaehlt Rotationen hoch
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session refresh fehlgeschlagen: %w", err)
	}

	if session.Rotations >= maxRotations {
		return fmt.Errorf("maximale Rotationen erreicht (%d)", maxRotations)
	}

	session.Rotations++
	session.LastUsedAt = time.Now()

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("session serialisierung fehlgeschlagen: %w", err)
	}

	sessionKey := sessionKeyPrefix + sessionID
	if err := sm.cache.Set(ctx, sessionKey, string(sessionJSON), sessionTTL); err != nil {
		return fmt.Errorf("session update fehlgeschlagen: %w", err)
	}

	return nil
}

// InvalidateSession entfernt eine Session aus Redis
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID ist erforderlich")
	}

	sessionKey := sessionKeyPrefix + sessionID
	if err := sm.cache.Delete(ctx, sessionKey); err != nil {
		sm.logger.Error("session loeschen fehlgeschlagen", err)
	}

	return nil
}

// InvalidateAllUserSessions entfernt alle Sessions eines Users
func (sm *SessionManager) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("userID ist erforderlich")
	}

	userSessionsKey := userSessionsKeyPrefix + userID
	indexJSON, err := sm.cache.Get(ctx, userSessionsKey)
	if err != nil {
		// Keine Sessions vorhanden
		return nil
	}

	var index UserSessionIndex
	if err := json.Unmarshal([]byte(indexJSON), &index); err != nil {
		return nil
	}

	// Jede Session loeschen
	for _, sid := range index.SessionIDs {
		sessionKey := sessionKeyPrefix + sid
		_ = sm.cache.Delete(ctx, sessionKey)
	}

	// Index loeschen
	_ = sm.cache.Delete(ctx, userSessionsKey)

	sm.logger.Info("alle Sessions invalidiert", "userID", userID, "count", len(index.SessionIDs))
	return nil
}

// bruteforceTTLForCount berechnet den verlaengernden Timeout basierend auf Fehlversuchen
// 1-9 Versuche: 1 Minute, 10: 5 Min, 15: 15 Min, 20: 30 Min, 25+: 1 Stunde
func bruteforceTTLForCount(count int) time.Duration {
	switch {
	case count < maxFailedLogins:
		return 1 * time.Minute
	case count < 15:
		return 5 * time.Minute
	case count < 20:
		return 15 * time.Minute
	case count < 25:
		return 30 * time.Minute
	default:
		return 1 * time.Hour
	}
}

// RecordFailedLogin zaehlt fehlgeschlagene Logins pro IP mit verlaengerndem Timeout
func (sm *SessionManager) RecordFailedLogin(ctx context.Context, ipAddress string) (int, error) {
	if ipAddress == "" {
		return 0, fmt.Errorf("ipAddress ist erforderlich")
	}

	bruteforceKey := bruteforceKeyPrefix + ipAddress
	countStr, err := sm.cache.Get(ctx, bruteforceKey)
	var count int
	if err == nil {
		count, _ = strconv.Atoi(countStr)
	}

	count++

	// Counter mit verlaengerndem TTL speichern
	ttl := bruteforceTTLForCount(count)
	if err := sm.cache.Set(ctx, bruteforceKey, strconv.Itoa(count), ttl); err != nil {
		return 0, fmt.Errorf("brute-force counter speichern fehlgeschlagen: %w", err)
	}

	sm.logger.Warn("fehlgeschlagener Login", "ipAddress", ipAddress, "count", count, "lockout", ttl.String())
	return count, nil
}

// GetFailedLoginCount gibt die Anzahl fehlgeschlagener Logins zurueck
func (sm *SessionManager) GetFailedLoginCount(ctx context.Context, ipAddress string) (int, error) {
	bruteforceKey := bruteforceKeyPrefix + ipAddress
	countStr, err := sm.cache.Get(ctx, bruteforceKey)
	if err != nil {
		return 0, nil // Kein Eintrag = 0 Fehlversuche
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, nil
	}

	return count, nil
}

// IsIPBlocked prueft ob eine IP gesperrt ist (>= 5 Fehlversuche)
func (sm *SessionManager) IsIPBlocked(ctx context.Context, ipAddress string) (bool, error) {
	count, err := sm.GetFailedLoginCount(ctx, ipAddress)
	if err != nil {
		return false, err
	}
	return count >= maxFailedLogins, nil
}

// ResetFailedLogins setzt den Counter zurueck (nach erfolgreichem Login)
func (sm *SessionManager) ResetFailedLogins(ctx context.Context, ipAddress string) error {
	bruteforceKey := bruteforceKeyPrefix + ipAddress
	return sm.cache.Delete(ctx, bruteforceKey)
}

// addToUserSessions fuegt eine Session-ID zum User-Index hinzu
func (sm *SessionManager) addToUserSessions(ctx context.Context, userID, sessionID string) {
	userSessionsKey := userSessionsKeyPrefix + userID

	var index UserSessionIndex
	indexJSON, err := sm.cache.Get(ctx, userSessionsKey)
	if err == nil {
		_ = json.Unmarshal([]byte(indexJSON), &index)
	}

	index.SessionIDs = append(index.SessionIDs, sessionID)

	newIndexJSON, err := json.Marshal(index)
	if err != nil {
		sm.logger.Error("user sessions index serialisierung fehlgeschlagen", err)
		return
	}

	_ = sm.cache.Set(ctx, userSessionsKey, string(newIndexJSON), sessionTTL)
}
