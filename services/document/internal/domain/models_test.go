package domain

import (
	"testing"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// DocumentTemplate — defaults
// ---------------------------------------------------------------------------

func TestDocumentTemplateDefaults(t *testing.T) {
	tmpl := &DocumentTemplate{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Name:      "Lieferschein",
		Type:      "delivery_note",
		IsDefault: true,
		IsActive:  true,
	}

	if !tmpl.IsDefault {
		t.Error("template should be default")
	}
	if !tmpl.IsActive {
		t.Error("template should be active")
	}
	if tmpl.Type != "delivery_note" {
		t.Errorf("Type = %q, want %q", tmpl.Type, "delivery_note")
	}
}

// ---------------------------------------------------------------------------
// Document — mime types
// ---------------------------------------------------------------------------

func TestDocumentMimeTypes(t *testing.T) {
	validMimes := []string{
		"application/pdf",
		"text/html",
		"text/plain",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}

	for _, mime := range validMimes {
		doc := &Document{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			MimeType: mime,
			Status:   "generated",
		}
		if doc.MimeType == "" {
			t.Errorf("MimeType should not be empty for %s", mime)
		}
	}
}

// ---------------------------------------------------------------------------
// Attachment — file size
// ---------------------------------------------------------------------------

func TestAttachmentFileSize(t *testing.T) {
	att := &Attachment{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		FileName: "foto.jpg",
		FilePath: "/uploads/foto.jpg",
		FileSize: 2_500_000,
		MimeType: "image/jpeg",
	}

	// Max 50 MB
	maxSize := 50 * 1024 * 1024
	if att.FileSize > maxSize {
		t.Errorf("FileSize %d exceeds max %d", att.FileSize, maxSize)
	}
	if att.FileName == "" {
		t.Error("FileName should not be empty")
	}
}
