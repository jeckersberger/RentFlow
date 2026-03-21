package logger

import (
	"testing"
)

func TestNew_DefaultLevel(t *testing.T) {
	logger := New("", "")

	if logger == nil {
		t.Error("expected logger to be created, got nil")
	}

	simpleLogger, ok := logger.(*SimpleLogger)
	if !ok {
		t.Error("expected SimpleLogger type")
	}

	if simpleLogger.level != "info" {
		t.Errorf("expected default level 'info', got '%s'", simpleLogger.level)
	}

	if simpleLogger.prefix != "RentFlow" {
		t.Errorf("expected default prefix 'RentFlow', got '%s'", simpleLogger.prefix)
	}
}

func TestNew_CustomValues(t *testing.T) {
	logger := New("debug", "TestApp")

	simpleLogger := logger.(*SimpleLogger)

	if simpleLogger.level != "debug" {
		t.Errorf("expected level 'debug', got '%s'", simpleLogger.level)
	}

	if simpleLogger.prefix != "TestApp" {
		t.Errorf("expected prefix 'TestApp', got '%s'", simpleLogger.prefix)
	}
}

func TestSimpleLogger_Info(t *testing.T) {
	logger := New("info", "TEST")
	// Simply call the method to ensure it doesn't panic
	logger.Info("test message", "arg1", "arg2")
}

func TestSimpleLogger_Debug_NotShown(t *testing.T) {
	logger := New("info", "TEST")
	// Debug should not be logged when level is info
	logger.Debug("debug message")
}

func TestSimpleLogger_Debug_Shown(t *testing.T) {
	logger := New("debug", "TEST")
	// Debug should be logged when level is debug
	logger.Debug("debug message")
}

func TestSimpleLogger_Warn(t *testing.T) {
	logger := New("info", "TEST")
	logger.Warn("warning message")
}

func TestSimpleLogger_Error(t *testing.T) {
	logger := New("info", "TEST")
	logger.Error("error message")
}

func TestSimpleLogger_Fatal(t *testing.T) {
	// Skip fatal test in normal test run to avoid os.Exit(1)
	t.Skip("Skipping Fatal test to avoid os.Exit")
}

func TestSimpleLogger_WithPrefix(t *testing.T) {
	originalLogger := New("info", "Original").(*SimpleLogger)
	newLogger := originalLogger.WithPrefix("NewPrefix")

	simpleLogger := newLogger.(*SimpleLogger)

	if simpleLogger.prefix != "NewPrefix" {
		t.Errorf("expected prefix 'NewPrefix', got '%s'", simpleLogger.prefix)
	}

	if simpleLogger.level != "info" {
		t.Errorf("expected level preserved as 'info', got '%s'", simpleLogger.level)
	}

	// Original logger should be unchanged
	if originalLogger.prefix != "Original" {
		t.Errorf("expected original prefix 'Original' to be unchanged, got '%s'", originalLogger.prefix)
	}
}

func TestSimpleLogger_WithPrefix_PreservesLevel(t *testing.T) {
	originalLogger := New("debug", "Service1").(*SimpleLogger)
	newLogger := originalLogger.WithPrefix("Service2")

	newSimple := newLogger.(*SimpleLogger)

	if newSimple.level != "debug" {
		t.Errorf("expected level to be preserved as 'debug', got '%s'", newSimple.level)
	}

	// Just call Debug to verify it works
	newLogger.Debug("test debug")
}

func TestSimpleLogger_Implements_Logger_Interface(t *testing.T) {
	var _ Logger = New("info", "TEST")
}

func TestSimpleLogger_Info_WithMultipleArgs(t *testing.T) {
	logger := New("info", "TEST")
	logger.Info("message with args", 42, "string", 3.14)
}

func TestSimpleLogger_Warn_WithMultipleArgs(t *testing.T) {
	logger := New("info", "TEST")
	logger.Warn("warning", "code", 500)
}

func TestSimpleLogger_Error_WithMultipleArgs(t *testing.T) {
	logger := New("info", "TEST")
	logger.Error("error occurred", "status", 500, "resource", "user")
}

func TestSimpleLogger_MultipleLoggers_DifferentPrefixes(t *testing.T) {
	logger1 := New("info", "Service1")
	logger2 := New("info", "Service2")

	logger1.Info("from service 1")
	logger2.Info("from service 2")
}

// Tests for new features (Correlation-ID, Request-ID, WithField)

func TestSimpleLogger_WithCorrelationID(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithCorrID := logger.WithCorrelationID("corr-123-abc")

	simpleLogger := loggerWithCorrID.(*SimpleLogger)

	if simpleLogger.correlationID != "corr-123-abc" {
		t.Errorf("expected correlationID 'corr-123-abc', got '%s'", simpleLogger.correlationID)
	}

	if simpleLogger.prefix != "TEST" {
		t.Errorf("expected prefix 'TEST' to be preserved, got '%s'", simpleLogger.prefix)
	}

	if simpleLogger.level != "info" {
		t.Errorf("expected level 'info' to be preserved, got '%s'", simpleLogger.level)
	}
}

func TestSimpleLogger_WithRequestID(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithReqID := logger.WithRequestID("req-456-def")

	simpleLogger := loggerWithReqID.(*SimpleLogger)

	if simpleLogger.requestID != "req-456-def" {
		t.Errorf("expected requestID 'req-456-def', got '%s'", simpleLogger.requestID)
	}

	if simpleLogger.prefix != "TEST" {
		t.Errorf("expected prefix 'TEST' to be preserved, got '%s'", simpleLogger.prefix)
	}

	if simpleLogger.level != "info" {
		t.Errorf("expected level 'info' to be preserved, got '%s'", simpleLogger.level)
	}
}

func TestSimpleLogger_WithField(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithField := logger.WithField("userId", "user-789")

	simpleLogger := loggerWithField.(*SimpleLogger)

	if val, ok := simpleLogger.fields["userId"]; !ok || val != "user-789" {
		t.Errorf("expected field 'userId' with value 'user-789', got %v", simpleLogger.fields)
	}

	if simpleLogger.prefix != "TEST" {
		t.Errorf("expected prefix 'TEST' to be preserved, got '%s'", simpleLogger.prefix)
	}
}

func TestSimpleLogger_ChainedWithCorrelationIDAndRequestID(t *testing.T) {
	logger := New("info", "TEST")
	chainedLogger := logger.WithCorrelationID("corr-123").WithRequestID("req-456")

	simpleLogger := chainedLogger.(*SimpleLogger)

	if simpleLogger.correlationID != "corr-123" {
		t.Errorf("expected correlationID 'corr-123', got '%s'", simpleLogger.correlationID)
	}

	if simpleLogger.requestID != "req-456" {
		t.Errorf("expected requestID 'req-456', got '%s'", simpleLogger.requestID)
	}
}

func TestSimpleLogger_WithField_MultipleFields(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithFields := logger.
		WithField("userId", "user-123").
		WithField("action", "login").
		WithField("timestamp", 1234567890)

	simpleLogger := loggerWithFields.(*SimpleLogger)

	if val, ok := simpleLogger.fields["userId"]; !ok || val != "user-123" {
		t.Errorf("expected field 'userId', got %v", simpleLogger.fields)
	}

	if val, ok := simpleLogger.fields["action"]; !ok || val != "login" {
		t.Errorf("expected field 'action', got %v", simpleLogger.fields)
	}

	if val, ok := simpleLogger.fields["timestamp"]; !ok || val != 1234567890 {
		t.Errorf("expected field 'timestamp', got %v", simpleLogger.fields)
	}
}

func TestSimpleLogger_WithCorrelationID_PreservesExistingFields(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithField := logger.WithField("userId", "user-123")
	loggerWithCorrID := loggerWithField.WithCorrelationID("corr-123")

	simpleLogger := loggerWithCorrID.(*SimpleLogger)

	if val, ok := simpleLogger.fields["userId"]; !ok || val != "user-123" {
		t.Errorf("expected field 'userId' to be preserved, got %v", simpleLogger.fields)
	}

	if simpleLogger.correlationID != "corr-123" {
		t.Errorf("expected correlationID 'corr-123', got '%s'", simpleLogger.correlationID)
	}
}

func TestSimpleLogger_WithPrefix_PreservesCorrelationIDAndRequestID(t *testing.T) {
	logger := New("info", "Original").
		WithCorrelationID("corr-123").
		WithRequestID("req-456")

	newLogger := logger.WithPrefix("NewPrefix")
	simpleLogger := newLogger.(*SimpleLogger)

	if simpleLogger.prefix != "NewPrefix" {
		t.Errorf("expected prefix 'NewPrefix', got '%s'", simpleLogger.prefix)
	}

	if simpleLogger.correlationID != "corr-123" {
		t.Errorf("expected correlationID to be preserved as 'corr-123', got '%s'", simpleLogger.correlationID)
	}

	if simpleLogger.requestID != "req-456" {
		t.Errorf("expected requestID to be preserved as 'req-456', got '%s'", simpleLogger.requestID)
	}
}

func TestSimpleLogger_WithField_LogsData(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithFields := logger.
		WithField("userId", "user-123").
		WithField("action", "login")

	// Should not panic
	loggerWithFields.Info("user action logged")
}

func TestSimpleLogger_WithCorrelationID_LogsData(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithCorrID := logger.WithCorrelationID("corr-123")

	// Should not panic
	loggerWithCorrID.Info("correlated message")
}

func TestSimpleLogger_WithRequestID_LogsData(t *testing.T) {
	logger := New("info", "TEST")
	loggerWithReqID := logger.WithRequestID("req-456")

	// Should not panic
	loggerWithReqID.Info("request tracked message")
}
