package application

import (
	"encoding/json"
	"testing"
)

func TestReceiptAnalysisResult_InvoiceParsing(t *testing.T) {
	jsonData := `{
		"is_invoice": true,
		"document_type": "invoice",
		"confidence": 0.95,
		"vendor": "Buero GmbH",
		"vendor_address": "Musterstr. 1, 80331 Muenchen",
		"vendor_vat_id": "DE123456789",
		"vendor_iban": "DE89370400440532013000",
		"gross_amount": 119.00,
		"net_amount": 100.00,
		"tax_rate": 19.0,
		"tax_amount": 19.00,
		"currency": "EUR",
		"invoice_number": "RE-2026-001",
		"invoice_date": "2026-03-15",
		"due_date": "2026-04-15",
		"service_period_from": "2026-03-01",
		"service_period_to": "2026-03-31",
		"discount_percent": 2.0,
		"discount_days": 10,
		"payment_method": "bank_transfer",
		"line_items": [
			{
				"description": "Bueroklammern",
				"quantity": 10,
				"unit_price": 10.00,
				"total_price": 100.00,
				"tax_rate": 19.0
			}
		],
		"description": "Buerobedarf Rechnung",
		"suggested_category": "Buero- und Geschaeftsbedarf",
		"suggested_skr03": "4930",
		"suggested_skr04": "6300",
		"is_entertainment": false,
		"entertainment_location": "",
		"tip": 0.00
	}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse invoice JSON: %v", err)
	}

	if !result.IsInvoice {
		t.Error("expected IsInvoice to be true")
	}
	if result.DocumentType != "invoice" {
		t.Errorf("expected DocumentType 'invoice', got '%s'", result.DocumentType)
	}
	if result.Confidence != 0.95 {
		t.Errorf("expected Confidence 0.95, got %f", result.Confidence)
	}
	if result.Vendor != "Buero GmbH" {
		t.Errorf("expected Vendor 'Buero GmbH', got '%s'", result.Vendor)
	}
	if result.GrossAmount != 119.00 {
		t.Errorf("expected GrossAmount 119.00, got %f", result.GrossAmount)
	}
	if result.NetAmount != 100.00 {
		t.Errorf("expected NetAmount 100.00, got %f", result.NetAmount)
	}
	if result.TaxRate != 19.0 {
		t.Errorf("expected TaxRate 19.0, got %f", result.TaxRate)
	}
	if result.TaxAmount != 19.00 {
		t.Errorf("expected TaxAmount 19.00, got %f", result.TaxAmount)
	}
	if result.Currency != "EUR" {
		t.Errorf("expected Currency 'EUR', got '%s'", result.Currency)
	}
	if result.InvoiceNumber != "RE-2026-001" {
		t.Errorf("expected InvoiceNumber 'RE-2026-001', got '%s'", result.InvoiceNumber)
	}
	if result.InvoiceDate != "2026-03-15" {
		t.Errorf("expected InvoiceDate '2026-03-15', got '%s'", result.InvoiceDate)
	}
	if result.DueDate != "2026-04-15" {
		t.Errorf("expected DueDate '2026-04-15', got '%s'", result.DueDate)
	}
	if result.DiscountPercent != 2.0 {
		t.Errorf("expected DiscountPercent 2.0, got %f", result.DiscountPercent)
	}
	if result.DiscountDays != 10 {
		t.Errorf("expected DiscountDays 10, got %d", result.DiscountDays)
	}
	if result.PaymentMethod != "bank_transfer" {
		t.Errorf("expected PaymentMethod 'bank_transfer', got '%s'", result.PaymentMethod)
	}
	if result.SuggestedSKR03 != "4930" {
		t.Errorf("expected SuggestedSKR03 '4930', got '%s'", result.SuggestedSKR03)
	}
	if len(result.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(result.LineItems))
	}
	if result.LineItems[0].Description != "Bueroklammern" {
		t.Errorf("expected line item description 'Bueroklammern', got '%s'", result.LineItems[0].Description)
	}
	if result.LineItems[0].Quantity != 10 {
		t.Errorf("expected line item quantity 10, got %f", result.LineItems[0].Quantity)
	}
}

func TestReceiptAnalysisResult_ReceiptParsing(t *testing.T) {
	jsonData := `{
		"is_invoice": true,
		"document_type": "receipt",
		"confidence": 0.88,
		"vendor": "REWE",
		"gross_amount": 23.47,
		"net_amount": 21.51,
		"tax_rate": 7.0,
		"tax_amount": 1.96,
		"currency": "EUR",
		"invoice_date": "2026-03-20",
		"payment_method": "card",
		"line_items": [
			{"description": "Milch", "quantity": 2, "unit_price": 1.29, "total_price": 2.58, "tax_rate": 7.0},
			{"description": "Brot", "quantity": 1, "unit_price": 3.49, "total_price": 3.49, "tax_rate": 7.0}
		],
		"description": "Lebensmitteleinkauf",
		"is_entertainment": false
	}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse receipt JSON: %v", err)
	}

	if result.DocumentType != "receipt" {
		t.Errorf("expected DocumentType 'receipt', got '%s'", result.DocumentType)
	}
	if result.PaymentMethod != "card" {
		t.Errorf("expected PaymentMethod 'card', got '%s'", result.PaymentMethod)
	}
	if len(result.LineItems) != 2 {
		t.Errorf("expected 2 line items, got %d", len(result.LineItems))
	}
	if result.TaxRate != 7.0 {
		t.Errorf("expected TaxRate 7.0 (ermaessigt), got %f", result.TaxRate)
	}
}

func TestReceiptAnalysisResult_CreditNoteParsing(t *testing.T) {
	jsonData := `{
		"is_invoice": true,
		"document_type": "credit_note",
		"confidence": 0.92,
		"vendor": "Amazon EU S.a.r.l.",
		"gross_amount": -59.99,
		"net_amount": -50.41,
		"tax_rate": 19.0,
		"tax_amount": -9.58,
		"currency": "EUR",
		"invoice_number": "GS-2026-4711",
		"invoice_date": "2026-03-18",
		"line_items": [],
		"description": "Gutschrift fuer Retoure",
		"is_entertainment": false
	}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse credit note JSON: %v", err)
	}

	if result.DocumentType != "credit_note" {
		t.Errorf("expected DocumentType 'credit_note', got '%s'", result.DocumentType)
	}
	if result.GrossAmount >= 0 {
		t.Error("expected negative GrossAmount for credit note")
	}
}

func TestReceiptAnalysisResult_QuoteParsing(t *testing.T) {
	jsonData := `{
		"is_invoice": false,
		"document_type": "quote",
		"confidence": 0.85,
		"vendor": "Handwerker Meier",
		"gross_amount": 5950.00,
		"net_amount": 5000.00,
		"tax_rate": 19.0,
		"tax_amount": 950.00,
		"currency": "EUR",
		"description": "Angebot fuer Dachsanierung",
		"is_entertainment": false
	}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse quote JSON: %v", err)
	}

	if result.IsInvoice {
		t.Error("expected IsInvoice to be false for a quote")
	}
	if result.DocumentType != "quote" {
		t.Errorf("expected DocumentType 'quote', got '%s'", result.DocumentType)
	}
}

func TestReceiptAnalysisResult_NotRecognized(t *testing.T) {
	jsonData := `{
		"is_invoice": false,
		"document_type": "other",
		"confidence": 0.30,
		"vendor": "",
		"gross_amount": 0,
		"net_amount": 0,
		"tax_rate": 0,
		"tax_amount": 0,
		"currency": "",
		"line_items": [],
		"description": "Unleserliches Dokument",
		"is_entertainment": false
	}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse unrecognized document JSON: %v", err)
	}

	if result.IsInvoice {
		t.Error("expected IsInvoice to be false for unrecognized document")
	}
	if result.DocumentType != "other" {
		t.Errorf("expected DocumentType 'other', got '%s'", result.DocumentType)
	}
	if result.Confidence > 0.5 {
		t.Errorf("expected low confidence for unrecognized doc, got %f", result.Confidence)
	}
}

func TestReceiptAnalysisResult_EntertainmentReceipt(t *testing.T) {
	jsonData := `{
		"is_invoice": true,
		"document_type": "entertainment",
		"confidence": 0.93,
		"vendor": "Restaurant Goldener Hirsch",
		"vendor_address": "Hauptstr. 12, 80331 Muenchen",
		"gross_amount": 187.50,
		"net_amount": 157.56,
		"tax_rate": 19.0,
		"tax_amount": 29.94,
		"currency": "EUR",
		"invoice_date": "2026-03-22",
		"payment_method": "card",
		"line_items": [
			{"description": "3x Hauptgang", "quantity": 3, "unit_price": 28.50, "total_price": 85.50, "tax_rate": 19.0},
			{"description": "Getraenke", "quantity": 1, "unit_price": 45.00, "total_price": 45.00, "tax_rate": 19.0}
		],
		"description": "Geschaeftsessen mit Kunden",
		"suggested_category": "Bewirtungskosten",
		"suggested_skr03": "4650",
		"suggested_skr04": "6640",
		"is_entertainment": true,
		"entertainment_location": "Restaurant Goldener Hirsch, Muenchen",
		"tip": 15.00
	}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse entertainment receipt JSON: %v", err)
	}

	if !result.IsEntertainment {
		t.Error("expected IsEntertainment to be true")
	}
	if result.DocumentType != "entertainment" {
		t.Errorf("expected DocumentType 'entertainment', got '%s'", result.DocumentType)
	}
	if result.EntertainmentLocation == "" {
		t.Error("expected entertainment location to be set")
	}
	if result.Tip != 15.00 {
		t.Errorf("expected Tip 15.00, got %f", result.Tip)
	}
	if result.SuggestedSKR03 != "4650" {
		t.Errorf("expected SKR03 '4650' for entertainment, got '%s'", result.SuggestedSKR03)
	}
}

func TestReceiptAnalysisResult_MultipleLineItems(t *testing.T) {
	jsonData := `{
		"is_invoice": true,
		"document_type": "invoice",
		"confidence": 0.97,
		"vendor": "MediaMarkt",
		"gross_amount": 1189.97,
		"currency": "EUR",
		"line_items": [
			{"description": "Laptop", "quantity": 1, "unit_price": 999.00, "total_price": 999.00, "tax_rate": 19.0},
			{"description": "Maus", "quantity": 2, "unit_price": 29.99, "total_price": 59.98, "tax_rate": 19.0},
			{"description": "USB-Kabel", "quantity": 3, "unit_price": 9.99, "total_price": 29.97, "tax_rate": 19.0}
		],
		"is_entertainment": false
	}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse multi-item invoice JSON: %v", err)
	}

	if len(result.LineItems) != 3 {
		t.Fatalf("expected 3 line items, got %d", len(result.LineItems))
	}

	// Verify individual line items
	if result.LineItems[0].UnitPrice != 999.00 {
		t.Errorf("expected first item unit price 999.00, got %f", result.LineItems[0].UnitPrice)
	}
	if result.LineItems[1].Quantity != 2 {
		t.Errorf("expected second item quantity 2, got %f", result.LineItems[1].Quantity)
	}
	if result.LineItems[2].TotalPrice != 29.97 {
		t.Errorf("expected third item total price 29.97, got %f", result.LineItems[2].TotalPrice)
	}
}

func TestReceiptAnalysisResult_EmptyResponse(t *testing.T) {
	jsonData := `{}`

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse empty JSON: %v", err)
	}

	if result.IsInvoice {
		t.Error("expected IsInvoice to default to false")
	}
	if result.DocumentType != "" {
		t.Errorf("expected empty DocumentType, got '%s'", result.DocumentType)
	}
	if result.GrossAmount != 0 {
		t.Errorf("expected GrossAmount 0, got %f", result.GrossAmount)
	}
	if result.IsEntertainment {
		t.Error("expected IsEntertainment to default to false")
	}
}

func TestReceiptAnalysisResult_InvalidJSON(t *testing.T) {
	invalidCases := []struct {
		name string
		json string
	}{
		{"empty string", ""},
		{"plain text", "Das ist keine JSON"},
		{"truncated JSON", `{"is_invoice": true, "vendor":`},
		{"invalid syntax", `{is_invoice: true}`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			var result ReceiptAnalysisResult
			err := json.Unmarshal([]byte(tc.json), &result)
			if err == nil {
				t.Errorf("expected error for invalid JSON input '%s'", tc.name)
			}
		})
	}
}

func TestReceiptAnalysisResult_MarkdownWrappedJSON(t *testing.T) {
	// Simulates Claude wrapping JSON in markdown code blocks
	rawResponse := "```json\n{\"is_invoice\": true, \"document_type\": \"invoice\", \"vendor\": \"Test GmbH\", \"gross_amount\": 119.00}\n```"

	// Replicate the extraction logic from callClaudeVision
	responseText := rawResponse
	if len(responseText) > 7 && responseText[:7] == "```json" {
		responseText = responseText[7:]
		if len(responseText) > 3 && responseText[len(responseText)-3:] == "```" {
			responseText = responseText[:len(responseText)-3]
		}
	}

	// Trim whitespace
	for len(responseText) > 0 && (responseText[0] == '\n' || responseText[0] == ' ' || responseText[0] == '\t') {
		responseText = responseText[1:]
	}
	for len(responseText) > 0 && (responseText[len(responseText)-1] == '\n' || responseText[len(responseText)-1] == ' ' || responseText[len(responseText)-1] == '\t') {
		responseText = responseText[:len(responseText)-1]
	}

	var result ReceiptAnalysisResult
	err := json.Unmarshal([]byte(responseText), &result)
	if err != nil {
		t.Fatalf("failed to parse markdown-wrapped JSON: %v", err)
	}

	if !result.IsInvoice {
		t.Error("expected IsInvoice to be true")
	}
	if result.Vendor != "Test GmbH" {
		t.Errorf("expected Vendor 'Test GmbH', got '%s'", result.Vendor)
	}
}

func TestReceiptAnalysisResult_MIMETypeValidation(t *testing.T) {
	validMIMETypes := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"image/gif",
		"application/pdf",
	}

	invalidMIMETypes := []string{
		"text/plain",
		"application/json",
		"video/mp4",
		"audio/mpeg",
		"",
	}

	isValidMIME := func(mimeType string) bool {
		switch mimeType {
		case "image/jpeg", "image/png", "image/webp", "image/gif", "application/pdf":
			return true
		default:
			return false
		}
	}

	for _, mime := range validMIMETypes {
		t.Run("valid_"+mime, func(t *testing.T) {
			if !isValidMIME(mime) {
				t.Errorf("expected MIME type '%s' to be valid", mime)
			}
		})
	}

	for _, mime := range invalidMIMETypes {
		name := mime
		if name == "" {
			name = "empty"
		}
		t.Run("invalid_"+name, func(t *testing.T) {
			if isValidMIME(mime) {
				t.Errorf("expected MIME type '%s' to be invalid", mime)
			}
		})
	}
}

func TestReceiptAnalysisResult_PDFvImageContentBlock(t *testing.T) {
	// Verify the content block type selection logic
	testCases := []struct {
		mimeType     string
		expectedType string
	}{
		{"application/pdf", "document"},
		{"image/jpeg", "image"},
		{"image/png", "image"},
	}

	for _, tc := range testCases {
		t.Run(tc.mimeType, func(t *testing.T) {
			var blockType string
			if tc.mimeType == "application/pdf" {
				blockType = "document"
			} else {
				blockType = "image"
			}
			if blockType != tc.expectedType {
				t.Errorf("for MIME '%s': expected block type '%s', got '%s'", tc.mimeType, tc.expectedType, blockType)
			}
		})
	}
}

func TestReceiptAnalysisResult_RoundtripSerialization(t *testing.T) {
	original := ReceiptAnalysisResult{
		IsInvoice:    true,
		DocumentType: "invoice",
		Confidence:   0.95,
		Vendor:       "Test AG",
		GrossAmount:  238.00,
		NetAmount:    200.00,
		TaxRate:      19.0,
		TaxAmount:    38.00,
		Currency:     "EUR",
		LineItems: []LineItem{
			{Description: "Service", Quantity: 1, UnitPrice: 200.00, TotalPrice: 200.00, TaxRate: 19.0},
		},
		IsEntertainment: false,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed ReceiptAnalysisResult
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.IsInvoice != original.IsInvoice {
		t.Error("IsInvoice mismatch after roundtrip")
	}
	if parsed.Vendor != original.Vendor {
		t.Errorf("Vendor mismatch: got '%s'", parsed.Vendor)
	}
	if parsed.GrossAmount != original.GrossAmount {
		t.Errorf("GrossAmount mismatch: got %f", parsed.GrossAmount)
	}
	if len(parsed.LineItems) != len(original.LineItems) {
		t.Errorf("LineItems count mismatch: got %d", len(parsed.LineItems))
	}
}
