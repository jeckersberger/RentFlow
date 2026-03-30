package application

import (
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// ---------------------------------------------------------------------------
// IsConfigured
// ---------------------------------------------------------------------------

func TestIsConfigured_NoAPIKey(t *testing.T) {
	provider := NewClaudeProvider("", "", zerolog.Nop())
	if provider.IsConfigured() {
		t.Error("IsConfigured() = true, want false when API key is empty")
	}
}

func TestIsConfigured_WithAPIKey(t *testing.T) {
	provider := NewClaudeProvider("sk-test-key", "", zerolog.Nop())
	if !provider.IsConfigured() {
		t.Error("IsConfigured() = false, want true when API key is set")
	}
}

// ---------------------------------------------------------------------------
// ModelName
// ---------------------------------------------------------------------------

func TestModelName_Default(t *testing.T) {
	provider := NewClaudeProvider("sk-test", "", zerolog.Nop())
	got := provider.ModelName()
	want := defaultClaudeModel
	if got != want {
		t.Errorf("ModelName() = %q, want %q (default)", got, want)
	}
}

func TestModelName_Custom(t *testing.T) {
	provider := NewClaudeProvider("sk-test", "claude-opus-4-20250514", zerolog.Nop())
	got := provider.ModelName()
	want := "claude-opus-4-20250514"
	if got != want {
		t.Errorf("ModelName() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Complete — error when not configured
// ---------------------------------------------------------------------------

func TestComplete_NotConfigured(t *testing.T) {
	provider := NewClaudeProvider("", "", zerolog.Nop())

	_, err := provider.Complete(context.Background(), AIRequest{
		UserMessage: "Hello",
	})
	if err == nil {
		t.Fatal("Complete() should return error when provider is not configured")
	}
	if !strings.Contains(err.Error(), "nicht konfiguriert") {
		t.Errorf("error should mention not configured, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// truncate helper
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{name: "short string", input: "hello", maxLen: 10, want: "hello"},
		{name: "exact length", input: "hello", maxLen: 5, want: "hello"},
		{name: "truncated", input: "hello world", maxLen: 5, want: "hello..."},
		{name: "empty", input: "", maxLen: 5, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewClaudeProvider — verify initialization
// ---------------------------------------------------------------------------

func TestNewClaudeProvider_SetsFields(t *testing.T) {
	provider := NewClaudeProvider("my-key", "my-model", zerolog.Nop())

	if provider.apiKey != "my-key" {
		t.Errorf("apiKey = %q, want %q", provider.apiKey, "my-key")
	}
	if provider.model != "my-model" {
		t.Errorf("model = %q, want %q", provider.model, "my-model")
	}
	if provider.baseURL != defaultClaudeBaseURL {
		t.Errorf("baseURL = %q, want %q", provider.baseURL, defaultClaudeBaseURL)
	}
	if provider.client == nil {
		t.Error("client should not be nil")
	}
}
