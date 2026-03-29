package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/jeckersberger/EquipFlow/pkg/common/response"
)

// AppError represents a structured application error with an HTTP status code.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New creates a new AppError with the given fields.
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Common pre-defined application errors.
var (
	ErrNotFound     = New("NOT_FOUND", "The requested resource was not found", http.StatusNotFound)
	ErrUnauthorized = New("UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
	ErrForbidden    = New("FORBIDDEN", "You do not have permission to perform this action", http.StatusForbidden)
	ErrBadRequest   = New("BAD_REQUEST", "The request is invalid", http.StatusBadRequest)
	ErrConflict     = New("CONFLICT", "The resource already exists", http.StatusConflict)
	ErrInternal     = New("INTERNAL_ERROR", "An internal server error occurred", http.StatusInternalServerError)
)

// Wrap creates a copy of a base AppError with a custom message.
func Wrap(base *AppError, message string) *AppError {
	return &AppError{
		Code:       base.Code,
		Message:    message,
		HTTPStatus: base.HTTPStatus,
	}
}

// HandleError inspects the error and writes the appropriate JSON response.
// If the error (or any wrapped cause) is an *AppError, its code/message/status
// are used. Otherwise it checks for known domain error patterns and falls back
// to a generic 500.
func HandleError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		response.Error(w, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	// Map well-known domain errors to HTTP status codes.
	if mapped, ok := mapDomainError(err); ok {
		response.Error(w, mapped.HTTPStatus, mapped.Code, mapped.Message)
		return
	}

	response.Error(w, http.StatusInternalServerError, ErrInternal.Code, ErrInternal.Message)
}

// domainErrorMappings maps error message substrings to AppErrors.
var domainErrorMappings = []struct {
	substr string
	appErr *AppError
}{
	{"not found", ErrNotFound},
	{"invalid equipment status", Wrap(ErrBadRequest, "Ungueltiger Equipment-Status")},
	{"invalid equipment condition", Wrap(ErrBadRequest, "Ungueltiger Equipment-Zustand")},
	{"has child categories", Wrap(ErrConflict, "Kategorie hat Unterkategorien")},
	{"has associated equipment", Wrap(ErrConflict, "Kategorie hat zugeordnetes Equipment")},
	{"duplicate barcode", Wrap(ErrConflict, "Barcode bereits vergeben")},
	{"duplicate rfid", Wrap(ErrConflict, "RFID-Tag bereits vergeben")},
	{"not available", Wrap(ErrConflict, "Equipment ist nicht verfuegbar")},
	{"invalid project status", Wrap(ErrBadRequest, "Ungueltiger Projekt-Status")},
	{"invalid project equipment status", Wrap(ErrBadRequest, "Ungueltiger Projekt-Equipment-Status")},
	{"reservation conflict", Wrap(ErrConflict, "Reservierungskonflikt")},
	{"project has assigned equipment", Wrap(ErrConflict, "Projekt hat zugeordnetes Equipment")},
	{"name is required", Wrap(ErrBadRequest, "Name ist erforderlich")},
	{"end_date must be after start_date", Wrap(ErrBadRequest, "Enddatum muss nach Startdatum liegen")},
	{"invalid start_date", Wrap(ErrBadRequest, "Ungueltiges Startdatum")},
	{"invalid end_date", Wrap(ErrBadRequest, "Ungueltiges Enddatum")},
	{"company name is required", Wrap(ErrBadRequest, "Firmenname ist erforderlich")},
	{"first_name and last_name are required", Wrap(ErrBadRequest, "Vor- und Nachname sind erforderlich")},
	{"note content is required", Wrap(ErrBadRequest, "Notizinhalt ist erforderlich")},
	{"invoice is not in draft status", Wrap(ErrBadRequest, "Rechnung ist nicht im Entwurfsstatus")},
	{"invoice is already fully paid", Wrap(ErrBadRequest, "Rechnung ist bereits vollstaendig bezahlt")},
	{"invalid invoice type", Wrap(ErrBadRequest, "Ungueltiger Rechnungstyp")},
	{"payment exceeds remaining balance", Wrap(ErrBadRequest, "Zahlung uebersteigt den offenen Betrag")},
	{"customer_name is required", Wrap(ErrBadRequest, "Kundenname ist erforderlich")},
	{"description is required", Wrap(ErrBadRequest, "Beschreibung ist erforderlich")},
	{"amount must be positive", Wrap(ErrBadRequest, "Betrag muss positiv sein")},
	{"invoice cannot accept payments", Wrap(ErrBadRequest, "Rechnung kann keine Zahlungen empfangen")},
	// Scanner domain errors
	{"invalid scan action", Wrap(ErrBadRequest, "Ungueltige Scan-Aktion")},
	{"scan event not found", ErrNotFound},
	{"scanner device not found", ErrNotFound},
	{"duplicate device", Wrap(ErrConflict, "Scanner-Geraet bereits registriert")},
	{"barcode or rfid_tag is required", Wrap(ErrBadRequest, "Barcode oder RFID-Tag erforderlich")},
	{"equipment_id is required for checkout", Wrap(ErrBadRequest, "Equipment-ID fuer Check-Out erforderlich")},
	{"project_id is required for checkout", Wrap(ErrBadRequest, "Projekt-ID fuer Check-Out erforderlich")},
	{"scan session not found", ErrNotFound},
	{"scan session already completed", Wrap(ErrConflict, "Scan-Session bereits abgeschlossen")},
	{"invalid session type", Wrap(ErrBadRequest, "Ungueltiger Session-Typ")},
	{"signature_data is required", Wrap(ErrBadRequest, "Signatur-Daten erforderlich")},
	{"identifier is required", Wrap(ErrBadRequest, "Identifier ist erforderlich")},
	// Warehouse domain errors
	{"warehouse not found", ErrNotFound},
	{"zone not found", ErrNotFound},
	{"rack not found", ErrNotFound},
	{"location not found", ErrNotFound},
	{"movement not found", ErrNotFound},
	{"inventory check not found", ErrNotFound},
	{"inventory check already completed", Wrap(ErrConflict, "Inventur bereits abgeschlossen")},
	{"duplicate location code", Wrap(ErrConflict, "Stellplatz-Code bereits vergeben")},
	{"warehouse name is required", Wrap(ErrBadRequest, "Lagername ist erforderlich")},
	{"zone name is required", Wrap(ErrBadRequest, "Zonenname ist erforderlich")},
	{"rack name is required", Wrap(ErrBadRequest, "Regalname ist erforderlich")},
	{"location code is required", Wrap(ErrBadRequest, "Stellplatz-Code ist erforderlich")},
	{"equipment_id is required for movement", Wrap(ErrBadRequest, "Equipment-ID fuer Bewegung erforderlich")},
	// Crew domain errors
	{"crew member not found", ErrNotFound},
	{"assignment not found", ErrNotFound},
	{"qualification not found", ErrNotFound},
	{"first_name is required", Wrap(ErrBadRequest, "Vorname ist erforderlich")},
	{"last_name is required", Wrap(ErrBadRequest, "Nachname ist erforderlich")},
	// Document domain errors
	{"template not found", ErrNotFound},
	{"document not found", ErrNotFound},
	{"attachment not found", ErrNotFound},
	{"template name is required", Wrap(ErrBadRequest, "Template-Name ist erforderlich")},
	{"document title is required", Wrap(ErrBadRequest, "Dokumenttitel ist erforderlich")},
	// Expense domain errors
	{"expense not found", ErrNotFound},
	{"expense category not found", ErrNotFound},
	{"expense already approved", Wrap(ErrConflict, "Ausgabe bereits genehmigt")},
	{"category name is required", Wrap(ErrBadRequest, "Kategoriename ist erforderlich")},
	// Reporting domain errors
	{"report definition not found", ErrNotFound},
	{"report snapshot not found", ErrNotFound},
	{"widget not found", ErrNotFound},
	{"report name is required", Wrap(ErrBadRequest, "Reportname ist erforderlich")},
	{"widget name is required", Wrap(ErrBadRequest, "Widget-Name ist erforderlich")},
	// Maintenance domain errors
	{"maintenance schedule not found", ErrNotFound},
	{"maintenance task not found", ErrNotFound},
	{"task already completed", Wrap(ErrConflict, "Aufgabe bereits abgeschlossen")},
	{"task title is required", Wrap(ErrBadRequest, "Aufgabentitel ist erforderlich")},
	// Transport domain errors
	{"vehicle not found", ErrNotFound},
	{"transport order not found", ErrNotFound},
	{"order already completed", Wrap(ErrConflict, "Transportauftrag bereits abgeschlossen")},
	{"vehicle name is required", Wrap(ErrBadRequest, "Fahrzeugname ist erforderlich")},
	// Insurance domain errors
	{"policy not found", ErrNotFound},
	{"claim not found", ErrNotFound},
	{"policy name is required", Wrap(ErrBadRequest, "Policenname ist erforderlich")},
	// Workflow domain errors
	{"workflow definition not found", ErrNotFound},
	{"workflow instance not found", ErrNotFound},
	{"workflow already completed", Wrap(ErrConflict, "Workflow bereits abgeschlossen")},
	// AI domain errors
	{"prediction not found", ErrNotFound},
	{"suggestion not found", ErrNotFound},
	{"suggestion already processed", Wrap(ErrConflict, "Vorschlag bereits verarbeitet")},
	// Notification domain errors
	{"notification not found", ErrNotFound},
	// Federation domain errors
	{"partner not found", ErrNotFound},
	{"listing not found", ErrNotFound},
	{"federation request not found", ErrNotFound},
	{"partner name is required", Wrap(ErrBadRequest, "Partnername ist erforderlich")},
	// Audit domain errors
	{"audit log not found", ErrNotFound},
	{"audit policy not found", ErrNotFound},
}

func mapDomainError(err error) (*AppError, bool) {
	msg := err.Error()
	for _, m := range domainErrorMappings {
		if containsCI(msg, m.substr) {
			return m.appErr, true
		}
	}
	return nil, false
}

func containsCI(s, substr string) bool {
	// Simple case-insensitive contains without importing strings.
	ls := toLower(s)
	lsub := toLower(substr)
	return len(lsub) <= len(ls) && indexOf(ls, lsub) >= 0
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// Is checks whether the given error matches the target AppError by code.
func Is(err error, target *AppError) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == target.Code
}
