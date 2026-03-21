package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
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
	// Capture log output
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("info", "TEST")
	logger.Info("test message", "arg1", "arg2")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Errorf("expected [INFO] in output, got: %s", output)
	}

	if !strings.Contains(output, "TEST") {
		t.Errorf("expected prefix 'TEST' in output, got: %s", output)
	}

	if !strings.Contains(output, "test message") {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestSimpleLogger_Debug_NotShown(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("info", "TEST")
	logger.Debug("debug message")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Errorf("expected debug message not to appear when level is 'info', got: %s", output)
	}
}

func TestSimpleLogger_Debug_Shown(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("debug", "TEST")
	logger.Debug("debug message")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("expected [DEBUG] in output when level is 'debug', got: %s", output)
	}

	if !strings.Contains(output, "debug message") {
		t.Errorf("expected debug message in output, got: %s", output)
	}
}

func TestSimpleLogger_Warn(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("info", "TEST")
	logger.Warn("warning message")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[WARN]") {
		t.Errorf("expected [WARN] in output, got: %s", output)
	}

	if !strings.Contains(output, "warning message") {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestSimpleLogger_Error(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("info", "TEST")
	logger.Error("error message")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("expected [ERROR] in output, got: %s", output)
	}

	if !strings.Contains(output, "error message") {
		t.Errorf("expected message in output, got: %s", output)
	}
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

	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	newLogger.Debug("test debug")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("expected [DEBUG] in output when level is 'debug', got: %s", output)
	}
}

func TestSimpleLogger_Implements_Logger_Interface(t *testing.T) {
	var _ Logger = New("info", "TEST")
}

func TestSimpleLogger_Info_WithMultipleArgs(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("info", "TEST")
	logger.Info("message with args", 42, "string", 3.14)

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Errorf("expected [INFO] in output, got: %s", output)
	}

	if !strings.Contains(output, "message with args") {
		t.Errorf("expected message in output, got: %s", output)
	}
}

func TestSimpleLogger_Warn_WithMultipleArgs(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("info", "TEST")
	logger.Warn("warning", "code", 500)

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[WARN]") {
		t.Errorf("expected [WARN] in output, got: %s", output)
	}
}

func TestSimpleLogger_Error_WithMultipleArgs(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger := New("info", "TEST")
	logger.Error("error occurred", "status", 500, "resource", "user")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("expected [ERROR] in output, got: %s", output)
	}
}

func TestSimpleLogger_MultipleLoggers_DifferentPrefixes(t *testing.T) {
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(os.Stderr)

	logger1 := New("info", "Service1")
	logger2 := New("info", "Service2")

	logger1.Info("from service 1")
	logger2.Info("from service 2")

	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Service1") {
		t.Errorf("expected Service1 prefix in output, got: %s", output)
	}

	if !strings.Contains(output, "Service2") {
		t.Errorf("expected Service2 prefix in output, got: %s", output)
	}
}
