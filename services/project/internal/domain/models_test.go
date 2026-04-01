package domain

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ValidateTransition — Project status transitions
// ---------------------------------------------------------------------------

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		next      string
		wantValid bool
	}{
		// --- Valid forward transitions ---
		{name: "inquiry -> offer_sent", current: ProjectStatusInquiry, next: ProjectStatusOfferSent, wantValid: true},
		{name: "offer_sent -> confirmed", current: ProjectStatusOfferSent, next: ProjectStatusConfirmed, wantValid: true},
		{name: "confirmed -> in_preparation", current: ProjectStatusConfirmed, next: ProjectStatusInPreparation, wantValid: true},
		{name: "in_preparation -> active", current: ProjectStatusInPreparation, next: ProjectStatusActive, wantValid: true},
		{name: "active -> completed", current: ProjectStatusActive, next: ProjectStatusCompleted, wantValid: true},
		{name: "completed -> invoiced", current: ProjectStatusCompleted, next: ProjectStatusInvoiced, wantValid: true},
		{name: "invoiced -> archived", current: ProjectStatusInvoiced, next: ProjectStatusArchived, wantValid: true},

		// --- Cancellation (any except archived) ---
		{name: "inquiry -> cancelled", current: ProjectStatusInquiry, next: ProjectStatusCancelled, wantValid: true},
		{name: "offer_sent -> cancelled", current: ProjectStatusOfferSent, next: ProjectStatusCancelled, wantValid: true},
		{name: "confirmed -> cancelled", current: ProjectStatusConfirmed, next: ProjectStatusCancelled, wantValid: true},
		{name: "in_preparation -> cancelled", current: ProjectStatusInPreparation, next: ProjectStatusCancelled, wantValid: true},
		{name: "active -> cancelled", current: ProjectStatusActive, next: ProjectStatusCancelled, wantValid: true},
		{name: "completed -> cancelled", current: ProjectStatusCompleted, next: ProjectStatusCancelled, wantValid: true},
		{name: "invoiced -> cancelled", current: ProjectStatusInvoiced, next: ProjectStatusCancelled, wantValid: true},
		{name: "cancelled -> cancelled (idempotent)", current: ProjectStatusCancelled, next: ProjectStatusCancelled, wantValid: true},

		// --- Invalid: archived cannot be cancelled ---
		{name: "archived -> cancelled INVALID", current: ProjectStatusArchived, next: ProjectStatusCancelled, wantValid: false},

		// --- Invalid: skip steps ---
		{name: "inquiry -> active SKIP", current: ProjectStatusInquiry, next: ProjectStatusActive, wantValid: false},
		{name: "inquiry -> confirmed SKIP", current: ProjectStatusInquiry, next: ProjectStatusConfirmed, wantValid: false},
		{name: "offer_sent -> active SKIP", current: ProjectStatusOfferSent, next: ProjectStatusActive, wantValid: false},
		{name: "confirmed -> completed SKIP", current: ProjectStatusConfirmed, next: ProjectStatusCompleted, wantValid: false},

		// --- Invalid: backward transitions ---
		{name: "completed -> inquiry BACKWARD", current: ProjectStatusCompleted, next: ProjectStatusInquiry, wantValid: false},
		{name: "active -> confirmed BACKWARD", current: ProjectStatusActive, next: ProjectStatusConfirmed, wantValid: false},
		{name: "invoiced -> active BACKWARD", current: ProjectStatusInvoiced, next: ProjectStatusActive, wantValid: false},
		{name: "archived -> invoiced BACKWARD", current: ProjectStatusArchived, next: ProjectStatusInvoiced, wantValid: false},

		// --- Invalid: transition from terminal states ---
		{name: "archived -> active INVALID", current: ProjectStatusArchived, next: ProjectStatusActive, wantValid: false},
		{name: "cancelled -> active INVALID", current: ProjectStatusCancelled, next: ProjectStatusActive, wantValid: false},

		// --- Invalid: unknown status ---
		{name: "inquiry -> unknown_status INVALID", current: ProjectStatusInquiry, next: "unknown_status", wantValid: false},
		{name: "inquiry -> empty string INVALID", current: ProjectStatusInquiry, next: "", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{Status: tt.current}
			got := p.ValidateTransition(tt.next)
			if got != tt.wantValid {
				t.Errorf("Project{Status: %q}.ValidateTransition(%q) = %v, want %v",
					tt.current, tt.next, got, tt.wantValid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidateStatus — Project status validation
// ---------------------------------------------------------------------------

func TestValidateStatus(t *testing.T) {
	p := &Project{}

	validStatuses := []string{
		ProjectStatusInquiry,
		ProjectStatusOfferSent,
		ProjectStatusConfirmed,
		ProjectStatusInPreparation,
		ProjectStatusActive,
		ProjectStatusCompleted,
		ProjectStatusInvoiced,
		ProjectStatusArchived,
		ProjectStatusCancelled,
	}

	for _, s := range validStatuses {
		t.Run("valid_"+s, func(t *testing.T) {
			if !p.ValidateStatus(s) {
				t.Errorf("ValidateStatus(%q) = false, want true", s)
			}
		})
	}

	invalidStatuses := []string{"", "unknown", "deleted", "ACTIVE", "Active"}
	for _, s := range invalidStatuses {
		t.Run("invalid_"+s, func(t *testing.T) {
			if p.ValidateStatus(s) {
				t.Errorf("ValidateStatus(%q) = true, want false", s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ProjectStatusDraft alias
// ---------------------------------------------------------------------------

func TestProjectStatusDraftAlias(t *testing.T) {
	if ProjectStatusDraft != ProjectStatusInquiry {
		t.Errorf("ProjectStatusDraft = %q, want %q (alias for inquiry)",
			ProjectStatusDraft, ProjectStatusInquiry)
	}
}

// ---------------------------------------------------------------------------
// ValidateItemTransition — Packlist item status transitions
// ---------------------------------------------------------------------------

func TestValidateItemTransition(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		next      string
		wantValid bool
	}{
		// --- Valid forward transitions ---
		{name: "planned -> packed", current: PacklistItemStatusPlanned, next: PacklistItemStatusPacked, wantValid: true},
		{name: "packed -> loaded", current: PacklistItemStatusPacked, next: PacklistItemStatusLoaded, wantValid: true},
		{name: "loaded -> on_site", current: PacklistItemStatusLoaded, next: PacklistItemStatusOnSite, wantValid: true},
		{name: "on_site -> returned", current: PacklistItemStatusOnSite, next: PacklistItemStatusReturned, wantValid: true},

		// --- Damaged from any status ---
		{name: "planned -> damaged", current: PacklistItemStatusPlanned, next: PacklistItemStatusDamaged, wantValid: true},
		{name: "packed -> damaged", current: PacklistItemStatusPacked, next: PacklistItemStatusDamaged, wantValid: true},
		{name: "loaded -> damaged", current: PacklistItemStatusLoaded, next: PacklistItemStatusDamaged, wantValid: true},
		{name: "on_site -> damaged", current: PacklistItemStatusOnSite, next: PacklistItemStatusDamaged, wantValid: true},
		{name: "returned -> damaged", current: PacklistItemStatusReturned, next: PacklistItemStatusDamaged, wantValid: true},
		{name: "damaged -> damaged (idempotent)", current: PacklistItemStatusDamaged, next: PacklistItemStatusDamaged, wantValid: true},

		// --- Invalid: skip steps ---
		{name: "planned -> loaded SKIP", current: PacklistItemStatusPlanned, next: PacklistItemStatusLoaded, wantValid: false},
		{name: "planned -> on_site SKIP", current: PacklistItemStatusPlanned, next: PacklistItemStatusOnSite, wantValid: false},
		{name: "planned -> returned SKIP", current: PacklistItemStatusPlanned, next: PacklistItemStatusReturned, wantValid: false},
		{name: "packed -> on_site SKIP", current: PacklistItemStatusPacked, next: PacklistItemStatusOnSite, wantValid: false},

		// --- Invalid: backward transitions ---
		{name: "returned -> packed BACKWARD", current: PacklistItemStatusReturned, next: PacklistItemStatusPacked, wantValid: false},
		{name: "on_site -> loaded BACKWARD", current: PacklistItemStatusOnSite, next: PacklistItemStatusLoaded, wantValid: false},
		{name: "loaded -> planned BACKWARD", current: PacklistItemStatusLoaded, next: PacklistItemStatusPlanned, wantValid: false},
		{name: "packed -> planned BACKWARD", current: PacklistItemStatusPacked, next: PacklistItemStatusPlanned, wantValid: false},

		// --- Invalid: from terminal states (except damaged) ---
		{name: "returned -> on_site INVALID", current: PacklistItemStatusReturned, next: PacklistItemStatusOnSite, wantValid: false},
		{name: "returned -> planned INVALID", current: PacklistItemStatusReturned, next: PacklistItemStatusPlanned, wantValid: false},

		// --- Invalid: unknown status ---
		{name: "planned -> unknown INVALID", current: PacklistItemStatusPlanned, next: "unknown", wantValid: false},
		{name: "planned -> empty INVALID", current: PacklistItemStatusPlanned, next: "", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateItemTransition(tt.current, tt.next)
			if got != tt.wantValid {
				t.Errorf("ValidateItemTransition(%q, %q) = %v, want %v",
					tt.current, tt.next, got, tt.wantValid)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidateItemStatus — Packlist item status validation
// ---------------------------------------------------------------------------

func TestValidateItemStatus(t *testing.T) {
	validStatuses := []string{
		PacklistItemStatusPlanned,
		PacklistItemStatusPacked,
		PacklistItemStatusLoaded,
		PacklistItemStatusOnSite,
		PacklistItemStatusReturned,
		PacklistItemStatusDamaged,
	}

	for _, s := range validStatuses {
		t.Run("valid_"+s, func(t *testing.T) {
			if !ValidateItemStatus(s) {
				t.Errorf("ValidateItemStatus(%q) = false, want true", s)
			}
		})
	}

	invalidStatuses := []string{"", "unknown", "PACKED", "broken"}
	for _, s := range invalidStatuses {
		t.Run("invalid_"+s, func(t *testing.T) {
			if ValidateItemStatus(s) {
				t.Errorf("ValidateItemStatus(%q) = true, want false", s)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StatusColor — Calendar rendering colors
// ---------------------------------------------------------------------------

func TestStatusColor(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{ProjectStatusInquiry, "#64748b"},
		{ProjectStatusOfferSent, "#f59e0b"},
		{ProjectStatusConfirmed, "#3b82f6"},
		{ProjectStatusInPreparation, "#6366f1"},
		{ProjectStatusActive, "#22c55e"},
		{ProjectStatusCompleted, "#8b5cf6"},
		{ProjectStatusInvoiced, "#14b8a6"},
		{ProjectStatusCancelled, "#ef4444"},
		{ProjectStatusArchived, "#94a3b8"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StatusColor(tt.status)
			if got != tt.want {
				t.Errorf("StatusColor(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}

	// Default case for unknown status
	t.Run("unknown_returns_default", func(t *testing.T) {
		got := StatusColor("nonexistent")
		if got != "#3b82f6" {
			t.Errorf("StatusColor(\"nonexistent\") = %q, want #3b82f6 (default)", got)
		}
	})
}

// ---------------------------------------------------------------------------
// Status constants — Verify string values are correct
// ---------------------------------------------------------------------------

func TestProjectStatusConstants(t *testing.T) {
	expected := map[string]string{
		"ProjectStatusInquiry":       "inquiry",
		"ProjectStatusOfferSent":     "offer_sent",
		"ProjectStatusConfirmed":     "confirmed",
		"ProjectStatusInPreparation": "in_preparation",
		"ProjectStatusActive":        "active",
		"ProjectStatusCompleted":     "completed",
		"ProjectStatusInvoiced":      "invoiced",
		"ProjectStatusArchived":      "archived",
		"ProjectStatusCancelled":     "cancelled",
	}

	actual := map[string]string{
		"ProjectStatusInquiry":       ProjectStatusInquiry,
		"ProjectStatusOfferSent":     ProjectStatusOfferSent,
		"ProjectStatusConfirmed":     ProjectStatusConfirmed,
		"ProjectStatusInPreparation": ProjectStatusInPreparation,
		"ProjectStatusActive":        ProjectStatusActive,
		"ProjectStatusCompleted":     ProjectStatusCompleted,
		"ProjectStatusInvoiced":      ProjectStatusInvoiced,
		"ProjectStatusArchived":      ProjectStatusArchived,
		"ProjectStatusCancelled":     ProjectStatusCancelled,
	}

	for name, want := range expected {
		t.Run(name, func(t *testing.T) {
			got := actual[name]
			if got != want {
				t.Errorf("%s = %q, want %q", name, got, want)
			}
		})
	}
}

func TestPacklistItemStatusConstants(t *testing.T) {
	expected := map[string]string{
		"planned":  PacklistItemStatusPlanned,
		"packed":   PacklistItemStatusPacked,
		"loaded":   PacklistItemStatusLoaded,
		"on_site":  PacklistItemStatusOnSite,
		"returned": PacklistItemStatusReturned,
		"damaged":  PacklistItemStatusDamaged,
	}

	for want, got := range expected {
		t.Run(want, func(t *testing.T) {
			if got != want {
				t.Errorf("PacklistItemStatus constant = %q, want %q", got, want)
			}
		})
	}
}
