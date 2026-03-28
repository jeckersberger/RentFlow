package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard API response wrapper.
type Envelope struct {
	Data  interface{}  `json:"data"`
	Meta  *Meta        `json:"meta,omitempty"`
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail describes an error in the response.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta holds optional metadata (e.g. pagination).
type Meta struct {
	Page    int   `json:"page,omitempty"`
	PerPage int   `json:"per_page,omitempty"`
	Total   int64 `json:"total,omitempty"`
}

// Success sends a 200 JSON response with data.
func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Envelope{
		Data:  data,
		Error: nil,
	})
}

// SuccessWithMeta sends a 200 JSON response with data and metadata.
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta Meta) {
	writeJSON(w, http.StatusOK, Envelope{
		Data:  data,
		Meta:  &meta,
		Error: nil,
	})
}

// Created sends a 201 JSON response with the created resource.
func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, Envelope{
		Data:  data,
		Error: nil,
	})
}

// NoContent sends a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error sends a JSON error response with the given status code.
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, Envelope{
		Data: nil,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
