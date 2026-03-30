package application

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
)

// ---------------------------------------------------------------------------
// Generate — empty description
// ---------------------------------------------------------------------------

func TestGenerate_EmptyDescription(t *testing.T) {
	provider := NewClaudeProvider("", "", zerolog.Nop())
	svc := NewSmartAssetService(provider, zerolog.Nop())

	tests := []struct {
		name  string
		input SmartAssetRequest
	}{
		{name: "empty string", input: SmartAssetRequest{Description: ""}},
		{name: "whitespace only", input: SmartAssetRequest{Description: "   "}},
		{name: "tabs and newlines", input: SmartAssetRequest{Description: "\t\n  \n"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Generate(context.Background(), tt.input)
			if err == nil {
				t.Error("Generate() should return error for empty description")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// stripJSONFences
// ---------------------------------------------------------------------------

func TestStripJSONFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "json fenced",
			input: "```json\n{\"name\": \"SM58\"}\n```",
			want:  "{\"name\": \"SM58\"}",
		},
		{
			name:  "plain fenced",
			input: "```\n{\"key\": \"value\"}\n```",
			want:  "{\"key\": \"value\"}",
		},
		{
			name:  "no fences",
			input: "{\"key\": \"value\"}",
			want:  "{\"key\": \"value\"}",
		},
		{
			name:  "with whitespace around fences",
			input: "  ```json\n{\"a\":1}\n```  ",
			want:  "{\"a\":1}",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only fences",
			input: "```json\n```",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripJSONFences(tt.input)
			if got != tt.want {
				t.Errorf("stripJSONFences(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewSmartAssetService — construction
// ---------------------------------------------------------------------------

func TestNewSmartAssetService(t *testing.T) {
	provider := NewClaudeProvider("key", "model", zerolog.Nop())
	svc := NewSmartAssetService(provider, zerolog.Nop())

	if svc == nil {
		t.Fatal("NewSmartAssetService() returned nil")
	}
	if svc.provider != provider {
		t.Error("provider was not set correctly")
	}
}
