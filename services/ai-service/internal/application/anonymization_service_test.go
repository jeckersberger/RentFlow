package application

import (
	"strings"
	"testing"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

func newTestAnonymizationService() *AnonymizationService {
	log := logger.New("error", "test")
	return NewAnonymizationService(log)
}

// --- Email anonymization ---

func TestAnonymize_Email(t *testing.T) {
	svc := newTestAnonymizationService()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple email",
			input: "Kontakt: max.mustermann@example.com bitte melden",
			want:  "Kontakt: [EMAIL] bitte melden",
		},
		{
			name:  "email with plus",
			input: "Mail: user+tag@gmail.com",
			want:  "Mail: [EMAIL]",
		},
		{
			name:  "multiple emails",
			input: "Von: a@b.de An: c@d.com",
			want:  "Von: [EMAIL] An: [EMAIL]",
		},
		{
			name:  "email with subdomain",
			input: "test@mail.company.co.uk",
			want:  "[EMAIL]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Anonymize(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// --- IBAN anonymization ---

func TestAnonymize_IBAN(t *testing.T) {
	svc := newTestAnonymizationService()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "German IBAN",
			input: "IBAN: DE89370400440532013000",
			want:  "IBAN: [IBAN]",
		},
		{
			name:  "IBAN in sentence",
			input: "Bitte ueberweisen an DE12345678901234567890 danke",
			want:  "Bitte ueberweisen an [IBAN] danke",
		},
		{
			name:  "multiple IBANs",
			input: "Von DE11111111111111111111 nach DE22222222222222222222",
			want:  "Von [IBAN] nach [IBAN]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Anonymize(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// --- Phone number anonymization ---

func TestAnonymize_Phone(t *testing.T) {
	svc := newTestAnonymizationService()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "German mobile +49",
			input:    "Tel: +491711234567",
			contains: "[PHONE]",
		},
		{
			name:     "German landline 0",
			input:    "Festnetz: 08912345678",
			contains: "[PHONE]",
		},
		{
			name:     "short mobile",
			input:    "Ruf an: 01761234567",
			contains: "[PHONE]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Anonymize(tc.input)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("expected result to contain %q, got %q", tc.contains, got)
			}
		})
	}
}

// --- Name anonymization ---

func TestAnonymize_Names(t *testing.T) {
	svc := newTestAnonymizationService()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "German full name",
			input:    "Mieter: Max Mustermann hat unterschrieben",
			contains: "[NAME]",
		},
		{
			name:     "female name",
			input:    "Ansprechpartnerin: Maria Schmidt",
			contains: "[NAME]",
		},
		{
			name:     "multiple names",
			input:    "Hans Mueller und Petra Wagner waren anwesend",
			contains: "[NAME]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Anonymize(tc.input)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("expected result to contain %q, got %q", tc.contains, got)
			}
		})
	}
}

func TestAnonymize_NameNotLowercase(t *testing.T) {
	svc := newTestAnonymizationService()

	// Lowercase words should NOT be anonymized as names
	input := "das haus ist gross"
	got := svc.Anonymize(input)
	if strings.Contains(got, "[NAME]") {
		t.Errorf("lowercase words should not be matched as names, got %q", got)
	}
}

// --- Combined anonymization ---

func TestAnonymize_Combined(t *testing.T) {
	svc := newTestAnonymizationService()

	input := "Max Mustermann, max.mustermann@example.com, IBAN: DE89370400440532013000, Tel: +491711234567"
	got := svc.Anonymize(input)

	if !strings.Contains(got, "[NAME]") {
		t.Error("expected name to be anonymized")
	}
	if !strings.Contains(got, "[EMAIL]") {
		t.Error("expected email to be anonymized")
	}
	if !strings.Contains(got, "[IBAN]") {
		t.Error("expected IBAN to be anonymized")
	}
	if !strings.Contains(got, "[PHONE]") {
		t.Error("expected phone to be anonymized")
	}

	// Original PII should be gone
	if strings.Contains(got, "mustermann@example.com") {
		t.Error("email should have been replaced")
	}
	if strings.Contains(got, "DE89370400440532013000") {
		t.Error("IBAN should have been replaced")
	}
}

// --- No PII ---

func TestAnonymize_NoPII(t *testing.T) {
	svc := newTestAnonymizationService()

	input := "Die Miete betraegt 850 EUR pro Monat."
	got := svc.Anonymize(input)

	// Text without PII patterns should remain (mostly) unchanged
	if strings.Contains(got, "[EMAIL]") || strings.Contains(got, "[IBAN]") || strings.Contains(got, "[PHONE]") {
		t.Errorf("no PII should be detected in plain text, got %q", got)
	}
}

func TestAnonymize_EmptyString(t *testing.T) {
	svc := newTestAnonymizationService()

	got := svc.Anonymize("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// --- AnonymizeWithReport ---

func TestAnonymizeWithReport_CountsCorrectly(t *testing.T) {
	svc := newTestAnonymizationService()

	input := "Max Mustermann (max@test.de) und Anna Schmidt (anna@test.de), IBAN: DE89370400440532013000"
	result, report := svc.AnonymizeWithReport(input)

	if !strings.Contains(result, "[EMAIL]") {
		t.Error("expected emails to be anonymized")
	}

	if report["email"] != 2 {
		t.Errorf("expected 2 email matches, got %d", report["email"])
	}
	if report["iban"] != 1 {
		t.Errorf("expected 1 IBAN match, got %d", report["iban"])
	}
}

func TestAnonymizeWithReport_EmptyInput(t *testing.T) {
	svc := newTestAnonymizationService()

	result, report := svc.AnonymizeWithReport("")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
	if len(report) != 0 {
		t.Errorf("expected empty report, got %v", report)
	}
}

// --- GetPatterns ---

func TestGetPatterns_ReturnsAllPatterns(t *testing.T) {
	svc := newTestAnonymizationService()
	patterns := svc.GetPatterns()

	expectedTypes := []string{"name", "email", "iban", "phone", "address", "tax_id", "customer_id"}

	if len(patterns) != len(expectedTypes) {
		t.Fatalf("expected %d patterns, got %d", len(expectedTypes), len(patterns))
	}

	typeSet := make(map[string]bool)
	for _, p := range patterns {
		typeSet[string(p.PatternType)] = true
		if p.Pattern == "" {
			t.Errorf("pattern for type '%s' should not be empty", p.PatternType)
		}
		if p.Replacement == "" {
			t.Errorf("replacement for type '%s' should not be empty", p.PatternType)
		}
	}

	for _, expected := range expectedTypes {
		if !typeSet[expected] {
			t.Errorf("missing pattern type: %s", expected)
		}
	}
}

// --- Address anonymization ---

func TestAnonymize_Address(t *testing.T) {
	svc := newTestAnonymizationService()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "Strasse",
			input:    "Wohnung in Musterstr 5",
			contains: "[ADDRESS]",
		},
		{
			name:     "Strasse mit ae",
			input:    "Adresse: Hauptstraße 12",
			contains: "[ADDRESS]",
		},
		{
			name:     "Weg",
			input:    "Lieferung an Birkenweg 3a",
			contains: "[ADDRESS]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Anonymize(tc.input)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("expected result to contain %q, got %q", tc.contains, got)
			}
		})
	}
}

// --- Tax ID anonymization ---

func TestAnonymize_TaxID(t *testing.T) {
	svc := newTestAnonymizationService()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "USt-IdNr",
			input:    "USt-IdNr: DE123456789",
			contains: "[TAX_ID]",
		},
		{
			name:     "Steuernummer format",
			input:    "Steuernummer: 123/456/78901",
			contains: "[TAX_ID]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Anonymize(tc.input)
			if !strings.Contains(got, tc.contains) {
				t.Errorf("expected result to contain %q, got %q", tc.contains, got)
			}
		})
	}
}

// --- Customer ID anonymization ---

func TestAnonymize_CustomerID(t *testing.T) {
	svc := newTestAnonymizationService()

	input := "Kundennummer: K-12345678"
	got := svc.Anonymize(input)
	if !strings.Contains(got, "[CUSTOMER_ID]") {
		t.Errorf("expected customer ID to be anonymized, got %q", got)
	}
}
