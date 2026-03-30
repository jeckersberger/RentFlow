// Package templates embeds the default HTML templates for PDF generation.
package templates

import "embed"

// DefaultFS contains the embedded default HTML template files.
//
//go:embed invoice.html quote.html delivery_note.html reminder.html
var DefaultFS embed.FS
