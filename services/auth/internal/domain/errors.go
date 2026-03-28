package domain

import (
	"net/http"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// Domain-specific errors for the auth service.
var (
	ErrInvalidCredentials = apperrors.New("INVALID_CREDENTIALS", "Email oder Passwort falsch", http.StatusUnauthorized)
	ErrAccountLocked      = apperrors.New("ACCOUNT_LOCKED", "Konto gesperrt wegen zu vieler Fehlversuche", http.StatusForbidden)
	ErrAccountDisabled    = apperrors.New("ACCOUNT_DISABLED", "Konto ist deaktiviert", http.StatusForbidden)
	ErrInvalidToken       = apperrors.New("INVALID_TOKEN", "Token ungueltig oder abgelaufen", http.StatusUnauthorized)
	ErrEmailTaken         = apperrors.New("EMAIL_TAKEN", "Email-Adresse bereits vergeben", http.StatusConflict)
	ErrSlugTaken          = apperrors.New("SLUG_TAKEN", "Slug bereits vergeben", http.StatusConflict)
	ErrSetupComplete      = apperrors.New("SETUP_COMPLETE", "Setup wurde bereits abgeschlossen", http.StatusConflict)
	ErrInvalidSetupToken  = apperrors.New("INVALID_SETUP_TOKEN", "Setup-Token ungueltig", http.StatusForbidden)
	ErrWeakPassword       = apperrors.New("WEAK_PASSWORD", "Passwort muss mind. 8 Zeichen lang sein", http.StatusBadRequest)
	ErrInvitationExpired  = apperrors.New("INVITATION_EXPIRED", "Einladung abgelaufen", http.StatusGone)
	ErrSessionExpired     = apperrors.New("SESSION_EXPIRED", "Sitzung abgelaufen", http.StatusUnauthorized)
	ErrInsufficientRole   = apperrors.New("INSUFFICIENT_ROLE", "Keine Berechtigung fuer diese Aktion", http.StatusForbidden)
	ErrUserNotFound       = apperrors.New("USER_NOT_FOUND", "Benutzer nicht gefunden", http.StatusNotFound)
	ErrTenantNotFound     = apperrors.New("TENANT_NOT_FOUND", "Mandant nicht gefunden", http.StatusNotFound)
	ErrInvalidEmail       = apperrors.New("INVALID_EMAIL", "Ungueltige E-Mail-Adresse", http.StatusBadRequest)
	ErrSetupNotFound      = apperrors.New("SETUP_NOT_FOUND", "Setup-Status nicht gefunden", http.StatusNotFound)
)
