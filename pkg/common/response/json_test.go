package response

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_SuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "success"}

	JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}

	body := parseBody(w.Body)

	if !body["success"].(bool) {
		t.Errorf("expected success to be true")
	}

	if body["data"].(map[string]interface{})["message"] != "success" {
		t.Errorf("expected data.message to be 'success'")
	}
}

func TestJSON_ErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusBadRequest, nil)

	body := parseBody(w.Body)

	if body["success"].(bool) {
		t.Errorf("expected success to be false for status 400")
	}
}

func TestError_WritesErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["success"].(bool) {
		t.Errorf("expected success to be false")
	}

	if body["error"] == nil {
		t.Errorf("expected error object in response")
	}

	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "NOT_FOUND" {
		t.Errorf("expected error code 'NOT_FOUND', got '%s'", errObj["code"])
	}

	if errObj["message"] != "Resource not found" {
		t.Errorf("expected error message 'Resource not found', got '%s'", errObj["message"])
	}
}

func TestErrorWithDetails_IncludesDetails(t *testing.T) {
	w := httptest.NewRecorder()
	details := map[string]string{
		"field": "email",
		"issue": "invalid format",
	}

	ErrorWithDetails(w, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)

	body := parseBody(w.Body)

	if body["success"].(bool) {
		t.Errorf("expected success to be false")
	}

	errObj := body["error"].(map[string]interface{})
	detailsObj := errObj["details"].(map[string]interface{})

	if detailsObj["field"] != "email" {
		t.Errorf("expected details.field to be 'email', got '%v'", detailsObj["field"])
	}

	if detailsObj["issue"] != "invalid format" {
		t.Errorf("expected details.issue to be 'invalid format', got '%v'", detailsObj["issue"])
	}
}

func TestValidationError_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	errors := map[string]string{
		"email":    "invalid email",
		"password": "too short",
	}

	ValidationError(w, errors)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["success"].(bool) {
		t.Errorf("expected success to be false")
	}

	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", errObj["code"])
	}
}

func TestPaginated_Response(t *testing.T) {
	w := httptest.NewRecorder()
	items := []interface{}{"item1", "item2", "item3"}

	Paginated(w, items, 1, 10, 25)

	body := parseBody(w.Body)

	if !body["success"].(bool) {
		t.Errorf("expected success to be true")
	}

	if int(body["page"].(float64)) != 1 {
		t.Errorf("expected page 1, got %v", body["page"])
	}

	if int(body["per_page"].(float64)) != 10 {
		t.Errorf("expected per_page 10, got %v", body["per_page"])
	}

	if int(body["total"].(float64)) != 25 {
		t.Errorf("expected total 25, got %v", body["total"])
	}

	expectedPages := (25 + 10 - 1) / 10 // ceiling division
	if int(body["total_pages"].(float64)) != expectedPages {
		t.Errorf("expected total_pages %d, got %v", expectedPages, body["total_pages"])
	}
}

func TestPaginated_CalculatesTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		perPage  int
		expected int
	}{
		{"Exact division", 20, 10, 2},
		{"With remainder", 25, 10, 3},
		{"Single page", 5, 10, 1},
		{"Large dataset", 100, 15, 7},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		Paginated(w, []interface{}{}, 1, tt.perPage, tt.total)

		body := parseBody(w.Body)
		totalPages := int(body["total_pages"].(float64))

		if totalPages != tt.expected {
			t.Errorf("Test %s: expected %d pages, got %d", tt.name, tt.expected, totalPages)
		}
	}
}

func TestCreated_Returns201(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]interface{}{"id": "123"}

	Created(w, data)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if !body["success"].(bool) {
		t.Errorf("expected success to be true")
	}
}

func TestAccepted_Returns202(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"job_id": "abc123"}

	Accepted(w, data)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}
}

func TestNoContent_Returns204(t *testing.T) {
	w := httptest.NewRecorder()

	NoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	if len(body) > 0 {
		t.Errorf("expected empty body for NoContent, got %d bytes", len(body))
	}
}

func TestOK_Returns200(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	OK(w, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if !body["success"].(bool) {
		t.Errorf("expected success to be true")
	}
}

func TestBadRequest_Returns400(t *testing.T) {
	w := httptest.NewRecorder()

	BadRequest(w, "Invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["success"].(bool) {
		t.Errorf("expected success to be false")
	}

	if body["error"].(map[string]interface{})["code"] != "BAD_REQUEST" {
		t.Errorf("expected code 'BAD_REQUEST'")
	}
}

func TestUnauthorized_Returns401(t *testing.T) {
	w := httptest.NewRecorder()

	Unauthorized(w, "Invalid credentials")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code 'UNAUTHORIZED'")
	}
}

func TestForbidden_Returns403(t *testing.T) {
	w := httptest.NewRecorder()

	Forbidden(w, "Access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "FORBIDDEN" {
		t.Errorf("expected code 'FORBIDDEN'")
	}
}

func TestNotFound_Returns404(t *testing.T) {
	w := httptest.NewRecorder()

	NotFound(w, "User not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND'")
	}
}

func TestConflict_Returns409(t *testing.T) {
	w := httptest.NewRecorder()

	Conflict(w, "Resource already exists")

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "CONFLICT" {
		t.Errorf("expected code 'CONFLICT'")
	}
}

func TestUnprocessableEntity_Returns422(t *testing.T) {
	w := httptest.NewRecorder()

	UnprocessableEntity(w, "Cannot process entity")

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "UNPROCESSABLE_ENTITY" {
		t.Errorf("expected code 'UNPROCESSABLE_ENTITY'")
	}
}

func TestInternalError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()

	InternalError(w, "Internal server error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "INTERNAL_SERVER_ERROR" {
		t.Errorf("expected code 'INTERNAL_SERVER_ERROR'")
	}
}

func TestServiceUnavailable_Returns503(t *testing.T) {
	w := httptest.NewRecorder()

	ServiceUnavailable(w, "Service temporarily unavailable")

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "SERVICE_UNAVAILABLE" {
		t.Errorf("expected code 'SERVICE_UNAVAILABLE'")
	}
}

func TestTooManyRequests_Returns429(t *testing.T) {
	w := httptest.NewRecorder()

	TooManyRequests(w, "Rate limit exceeded")

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "RATE_LIMIT_EXCEEDED" {
		t.Errorf("expected code 'RATE_LIMIT_EXCEEDED'")
	}
}

func TestMethodNotAllowed_Returns405(t *testing.T) {
	w := httptest.NewRecorder()

	MethodNotAllowed(w, "Method not allowed")

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["error"].(map[string]interface{})["code"] != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected code 'METHOD_NOT_ALLOWED'")
	}
}

func TestCustom_CustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"custom": "data"}

	Custom(w, http.StatusTeapot, data)

	if w.Code != http.StatusTeapot {
		t.Errorf("expected status 418, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if body["data"] == nil {
		t.Errorf("expected data in response")
	}
}

func TestRawJSON_WritesRawBytes(t *testing.T) {
	w := httptest.NewRecorder()
	rawJSON := []byte(`{"raw":"json"}`)

	RawJSON(w, http.StatusOK, rawJSON)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json'")
	}

	body, _ := io.ReadAll(w.Body)
	if string(body) != string(rawJSON) {
		t.Errorf("expected raw JSON to be preserved, got '%s'", string(body))
	}
}

func TestList_Response(t *testing.T) {
	w := httptest.NewRecorder()
	items := []interface{}{"item1", "item2", "item3"}

	List(w, items)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := parseBody(w.Body)

	if !body["success"].(bool) {
		t.Errorf("expected success to be true")
	}

	// Note: the response structure wraps items in data
	if body["data"] == nil {
		t.Errorf("expected data in response")
	}
}

func TestAPIResponse_Structure(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	JSON(w, http.StatusOK, data)

	body := parseBody(w.Body)

	if _, ok := body["success"]; !ok {
		t.Errorf("expected 'success' field in response")
	}

	if _, ok := body["data"]; !ok {
		t.Errorf("expected 'data' field in response")
	}
}

func TestErrorInfo_Structure(t *testing.T) {
	w := httptest.NewRecorder()

	Error(w, http.StatusBadRequest, "TEST_CODE", "Test message")

	body := parseBody(w.Body)

	if body["error"] == nil {
		t.Errorf("expected 'error' field in response")
	}

	errObj := body["error"].(map[string]interface{})

	if _, ok := errObj["code"]; !ok {
		t.Errorf("expected 'code' field in error")
	}

	if _, ok := errObj["message"]; !ok {
		t.Errorf("expected 'message' field in error")
	}
}

// Helper function to parse JSON body
func parseBody(body io.Reader) map[string]interface{} {
	var result map[string]interface{}
	json.NewDecoder(body).Decode(&result)
	return result
}
