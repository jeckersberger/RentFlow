package response

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information in the response
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// JSON writes a successful JSON response
// status should be one of the HTTP status codes (e.g., http.StatusOK)
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := APIResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	_ = json.NewEncoder(w).Encode(response)
}

// Error writes an error JSON response
func Error(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

// ErrorWithDetails writes an error response with detailed information
func ErrorWithDetails(w http.ResponseWriter, status int, code string, message string, details map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	_ = json.NewEncoder(w).Encode(response)
}

// ValidationError writes a validation error response
// errors is a map of field names to error messages
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	ErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", errors)
}

// Paginated writes a paginated response
func Paginated(w http.ResponseWriter, data interface{}, page, perPage, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}

	response := PaginatedResponse{
		Success:    true,
		Data:       data,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}

	_ = json.NewEncoder(w).Encode(response)
}

// Created writes a 201 Created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// Accepted writes a 202 Accepted response
func Accepted(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusAccepted, data)
}

// NoContent writes a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// OK writes a 200 OK response
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// BadRequest writes a 400 Bad Request error
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized writes a 401 Unauthorized error
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden writes a 403 Forbidden error
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound writes a 404 Not Found error
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict writes a 409 Conflict error
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// UnprocessableEntity writes a 422 Unprocessable Entity error
func UnprocessableEntity(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", message)
}

// InternalError writes a 500 Internal Server Error
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// ServiceUnavailable writes a 503 Service Unavailable error
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// TooManyRequests writes a 429 Too Many Requests error
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message)
}

// MethodNotAllowed writes a 405 Method Not Allowed error
func MethodNotAllowed(w http.ResponseWriter, message string) {
	Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", message)
}

// Custom writes a custom response with the given status and data
func Custom(w http.ResponseWriter, status int, data interface{}) {
	JSON(w, status, data)
}

// RawJSON writes raw JSON data without wrapping
func RawJSON(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

// ListResponse represents a list response
type ListResponse struct {
	Items []interface{} `json:"items"`
	Count int           `json:"count"`
}

// List writes a list response
func List(w http.ResponseWriter, items []interface{}) {
	response := ListResponse{
		Items: items,
		Count: len(items),
	}
	JSON(w, http.StatusOK, response)
}
