package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	code := "TEST_ERROR"
	message := "This is a test error"

	err := New(code, message)

	if err.Code != code {
		t.Errorf("expected Code '%s', got '%s'", code, err.Code)
	}

	if err.Message != message {
		t.Errorf("expected Message '%s', got '%s'", message, err.Message)
	}

	if err.Err != nil {
		t.Errorf("expected Err to be nil, got %v", err.Err)
	}

	if len(err.Details) != 0 {
		t.Errorf("expected empty Details, got %v", err.Details)
	}
}

func TestWithError(t *testing.T) {
	underlyingErr := fmt.Errorf("database connection failed")
	appErr := New(CodeDatabaseError, "Failed to connect to database").WithError(underlyingErr)

	if appErr.Err != underlyingErr {
		t.Errorf("expected Err to be the underlying error, got %v", appErr.Err)
	}

	if appErr.Code != CodeDatabaseError {
		t.Errorf("expected Code '%s', got '%s'", CodeDatabaseError, appErr.Code)
	}
}

func TestWithDetail(t *testing.T) {
	appErr := New(CodeValidation, "Validation failed")
	appErr.WithDetail("field", "email").WithDetail("reason", "invalid format")

	if appErr.Details["field"] != "email" {
		t.Errorf("expected Details['field'] to be 'email', got %v", appErr.Details["field"])
	}

	if appErr.Details["reason"] != "invalid format" {
		t.Errorf("expected Details['reason'] to be 'invalid format', got %v", appErr.Details["reason"])
	}

	if len(appErr.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(appErr.Details))
	}
}

func TestWithDetail_Chaining(t *testing.T) {
	appErr := New(CodeBadRequest, "Invalid input")
	result := appErr.WithDetail("param1", "value1").WithDetail("param2", "value2")

	if result != appErr {
		t.Error("expected method chaining to return the same error instance")
	}

	if len(appErr.Details) != 2 {
		t.Errorf("expected 2 details after chaining, got %d", len(appErr.Details))
	}
}

func TestError_WithoutUnderlyingError(t *testing.T) {
	appErr := New(CodeNotFound, "User not found")

	expected := "[NOT_FOUND] User not found"
	actual := appErr.Error()

	if actual != expected {
		t.Errorf("expected Error() '%s', got '%s'", expected, actual)
	}
}

func TestError_WithUnderlyingError(t *testing.T) {
	underlyingErr := errors.New("connection timeout")
	appErr := New(CodeInternalServer, "Database error").WithError(underlyingErr)

	expected := "[INTERNAL_SERVER_ERROR] Database error: connection timeout"
	actual := appErr.Error()

	if actual != expected {
		t.Errorf("expected Error() '%s', got '%s'", expected, actual)
	}
}

func TestError_ImplementsErrorInterface(t *testing.T) {
	appErr := New(CodeConflict, "Resource already exists")

	var err error = appErr

	if err == nil {
		t.Error("expected AppError to be assignable to error interface")
	}

	if err.Error() != appErr.Error() {
		t.Error("expected Error() method to work through interface")
	}
}

func TestIsCode_WithAppError(t *testing.T) {
	appErr := New(CodeUnauthorized, "Invalid credentials")

	if !IsCode(appErr, CodeUnauthorized) {
		t.Errorf("expected IsCode to return true for matching code")
	}

	if IsCode(appErr, CodeForbidden) {
		t.Errorf("expected IsCode to return false for non-matching code")
	}
}

func TestIsCode_WithNonAppError(t *testing.T) {
	regularErr := errors.New("some error")

	if IsCode(regularErr, CodeValidation) {
		t.Errorf("expected IsCode to return false for non-AppError")
	}
}

func TestIsCode_WithNil(t *testing.T) {
	if IsCode(nil, CodeNotFound) {
		t.Errorf("expected IsCode to return false for nil error")
	}
}

func TestCommonErrorCodes(t *testing.T) {
	tests := []struct {
		code  string
		value string
	}{
		{CodeValidation, "VALIDATION_ERROR"},
		{CodeNotFound, "NOT_FOUND"},
		{CodeConflict, "CONFLICT"},
		{CodeInternalServer, "INTERNAL_SERVER_ERROR"},
		{CodeUnauthorized, "UNAUTHORIZED"},
		{CodeForbidden, "FORBIDDEN"},
		{CodeBadRequest, "BAD_REQUEST"},
		{CodeDatabaseError, "DATABASE_ERROR"},
		{CodeEventSourcingErr, "EVENT_SOURCING_ERROR"},
	}

	for _, tt := range tests {
		if tt.code != tt.value {
			t.Errorf("expected code constant '%s' to match '%s'", tt.code, tt.value)
		}
	}
}

func TestWithError_Chaining(t *testing.T) {
	err1 := errors.New("error 1")
	appErr := New(CodeDatabaseError, "DB failed")
	result := appErr.WithError(err1)

	if result != appErr {
		t.Error("expected WithError to return the same error instance")
	}
}

func TestAppError_MultipleDetails(t *testing.T) {
	appErr := New(CodeValidation, "Multiple validation errors")
	appErr.WithDetail("email", "invalid format").
		WithDetail("password", "too short").
		WithDetail("username", "already taken")

	if len(appErr.Details) != 3 {
		t.Errorf("expected 3 details, got %d", len(appErr.Details))
	}

	if appErr.Details["email"] != "invalid format" {
		t.Errorf("expected email detail to be 'invalid format', got %v", appErr.Details["email"])
	}

	if appErr.Details["password"] != "too short" {
		t.Errorf("expected password detail to be 'too short', got %v", appErr.Details["password"])
	}

	if appErr.Details["username"] != "already taken" {
		t.Errorf("expected username detail to be 'already taken', got %v", appErr.Details["username"])
	}
}

func TestAppError_OverwriteDetail(t *testing.T) {
	appErr := New(CodeValidation, "Error")
	appErr.WithDetail("field", "old value")
	appErr.WithDetail("field", "new value")

	if appErr.Details["field"] != "new value" {
		t.Errorf("expected detail to be overwritten to 'new value', got %v", appErr.Details["field"])
	}

	if len(appErr.Details) != 1 {
		t.Errorf("expected only 1 detail after overwrite, got %d", len(appErr.Details))
	}
}

func TestAppError_DetailTypes(t *testing.T) {
	appErr := New(CodeBadRequest, "Mixed types")
	appErr.WithDetail("count", 42).
		WithDetail("active", true).
		WithDetail("rate", 3.14).
		WithDetail("nested", map[string]interface{}{"key": "value"})

	if appErr.Details["count"] != 42 {
		t.Errorf("expected integer detail, got %T", appErr.Details["count"])
	}

	if appErr.Details["active"] != true {
		t.Errorf("expected boolean detail, got %T", appErr.Details["active"])
	}

	if appErr.Details["rate"] != 3.14 {
		t.Errorf("expected float detail, got %T", appErr.Details["rate"])
	}

	nested := appErr.Details["nested"].(map[string]interface{})
	if nested["key"] != "value" {
		t.Errorf("expected nested detail to work, got %v", nested)
	}
}
