package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// ReceiptVisionService handles AI-powered receipt/invoice analysis via Claude Vision
type ReceiptVisionService struct {
	providerSvc *ProviderService
	logger      logger.Logger
}

// NewReceiptVisionService creates a new receipt vision service
func NewReceiptVisionService(providerSvc *ProviderService, log logger.Logger) *ReceiptVisionService {
	return &ReceiptVisionService{
		providerSvc: providerSvc,
		logger:      log,
	}
}

// ReceiptAnalysisResult contains the structured data extracted from a receipt/invoice
type ReceiptAnalysisResult struct {
	// Ob es eine Rechnung/Beleg ist
	IsInvoice   bool    `json:"is_invoice"`
	DocumentType string `json:"document_type"` // "invoice", "receipt", "entertainment", "quote", "credit_note", "other"
	Confidence  float64 `json:"confidence"`     // 0-1

	// Lieferant
	Vendor        string `json:"vendor"`
	VendorAddress string `json:"vendor_address"`
	VendorVATID   string `json:"vendor_vat_id"`
	VendorIBAN    string `json:"vendor_iban"`

	// Betraege
	GrossAmount float64 `json:"gross_amount"`
	NetAmount   float64 `json:"net_amount"`
	TaxRate     float64 `json:"tax_rate"`
	TaxAmount   float64 `json:"tax_amount"`
	Currency    string  `json:"currency"`

	// Rechnungsdetails
	InvoiceNumber string `json:"invoice_number"`
	InvoiceDate   string `json:"invoice_date"`   // YYYY-MM-DD
	DueDate       string `json:"due_date"`        // YYYY-MM-DD
	ServicePeriodFrom string `json:"service_period_from"` // YYYY-MM-DD
	ServicePeriodTo   string `json:"service_period_to"`   // YYYY-MM-DD

	// Zahlungsbedingungen
	DiscountPercent float64 `json:"discount_percent"`
	DiscountDays    int     `json:"discount_days"`
	PaymentMethod   string  `json:"payment_method"` // "bank_transfer", "cash", "card"

	// Positionen
	LineItems []LineItem `json:"line_items"`
	Description string  `json:"description"` // Zusammenfassung

	// Kategorisierung
	SuggestedCategory string `json:"suggested_category"`
	SuggestedSKR03    string `json:"suggested_skr03"`
	SuggestedSKR04    string `json:"suggested_skr04"`

	// Bewirtungsbeleg-spezifisch
	IsEntertainment       bool    `json:"is_entertainment"`
	EntertainmentLocation string  `json:"entertainment_location"`
	Tip                   float64 `json:"tip"`
}

// LineItem represents a single item on an invoice
type LineItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	TaxRate     float64 `json:"tax_rate"`
}

const receiptAnalysisSystemPrompt = `Du bist ein Experte fuer deutsche Buchhaltung und Belegauswertung. Analysiere das Bild/PDF und extrahiere alle relevanten Daten.

WICHTIG:
- Erkenne ob es eine Rechnung, Bewirtungsbeleg, Quittung, Angebot oder Gutschrift ist
- Extrahiere USt-IdNr / Steuernummer des Lieferanten (wichtig fuer Vorsteuerabzug!)
- Erkenne deutsche MwSt-Saetze: 0%, 7%, 19%
- Wenn Zahlungsziel/Skonto auf dem Beleg steht, extrahiere es
- Bei Bewirtungsbelegen: erkenne Restaurant/Ort
- Schlage eine SKR03/SKR04-Kategorie vor basierend auf dem Inhalt
- Alle Betraege in EUR, Datumsformat YYYY-MM-DD
- Wenn ein Feld nicht erkennbar ist, setze es auf leer/null

Kategorien-Vorschlaege (SKR03):
- 4650: Bewirtungskosten
- 4654: Nicht abzugsfaehige Bewirtung
- 4930: Buero- und Geschaeftsbedarf
- 4940: Zeitschriften/Buecher
- 4946: Fortbildungskosten
- 4960: Miete/Leasing Bueroausstattung
- 4970: Nebenkosten des Geldverkehrs
- 4980: Porto
- 4210: Miete/Pacht
- 4240: Heizung
- 4250: Strom
- 4260: Reinigung
- 4510: Fahrzeugkosten
- 4520: Kfz-Versicherung
- 4530: Kfz-Steuer
- 4580: Sonstige Kfz-Kosten
- 4600: Werbekosten
- 4610: Reisekosten AN
- 4630: Geschenke abzugsfaehig
- 4806: Reparatur Betriebsausstattung
- 4830: Abschreibungen
- 4900: Sonstige betriebliche Aufwendungen

Antworte NUR mit validem JSON, kein anderer Text.`

const receiptAnalysisUserPrompt = `Analysiere diesen Beleg/diese Rechnung und extrahiere alle Daten im folgenden JSON-Format:

{
  "is_invoice": true/false,
  "document_type": "invoice|receipt|entertainment|quote|credit_note|other",
  "confidence": 0.0-1.0,
  "vendor": "Firmenname",
  "vendor_address": "Strasse, PLZ Ort",
  "vendor_vat_id": "DE123456789",
  "vendor_iban": "DE89...",
  "gross_amount": 0.00,
  "net_amount": 0.00,
  "tax_rate": 19.0,
  "tax_amount": 0.00,
  "currency": "EUR",
  "invoice_number": "RE-2026-001",
  "invoice_date": "YYYY-MM-DD",
  "due_date": "YYYY-MM-DD",
  "service_period_from": "YYYY-MM-DD",
  "service_period_to": "YYYY-MM-DD",
  "discount_percent": 0.0,
  "discount_days": 0,
  "payment_method": "bank_transfer|cash|card",
  "line_items": [{"description": "", "quantity": 1, "unit_price": 0.00, "total_price": 0.00, "tax_rate": 19.0}],
  "description": "Kurze Zusammenfassung",
  "suggested_category": "Kategoriename",
  "suggested_skr03": "4930",
  "suggested_skr04": "6300",
  "is_entertainment": false,
  "entertainment_location": "",
  "tip": 0.00
}`

// AnalyzeReceiptImage analyzes a receipt/invoice image using Claude Vision
func (s *ReceiptVisionService) AnalyzeReceiptImage(ctx context.Context, imageData []byte, mimeType string) (*ReceiptAnalysisResult, error) {
	provider, err := s.providerSvc.GetProvider("claude")
	if err != nil {
		return nil, fmt.Errorf("claude provider not available: %w", err)
	}

	claudeProvider, ok := provider.(*ClaudeProvider)
	if !ok {
		return nil, fmt.Errorf("provider is not a Claude provider")
	}

	return s.callClaudeVision(ctx, claudeProvider, imageData, mimeType)
}

// AnalyzeReceiptPDF analyzes a PDF receipt/invoice using Claude Vision
func (s *ReceiptVisionService) AnalyzeReceiptPDF(ctx context.Context, pdfData []byte) (*ReceiptAnalysisResult, error) {
	return s.AnalyzeReceiptImage(ctx, pdfData, "application/pdf")
}

// callClaudeVision sends an image to Claude's Vision API
func (s *ReceiptVisionService) callClaudeVision(ctx context.Context, provider *ClaudeProvider, data []byte, mimeType string) (*ReceiptAnalysisResult, error) {
	const apiURL = "https://api.anthropic.com/v1/messages"

	// Determine source type
	var contentBlock map[string]interface{}
	if mimeType == "application/pdf" {
		contentBlock = map[string]interface{}{
			"type": "document",
			"source": map[string]interface{}{
				"type":       "base64",
				"media_type": mimeType,
				"data":       base64.StdEncoding.EncodeToString(data),
			},
		}
	} else {
		contentBlock = map[string]interface{}{
			"type": "image",
			"source": map[string]interface{}{
				"type":       "base64",
				"media_type": mimeType,
				"data":       base64.StdEncoding.EncodeToString(data),
			},
		}
	}

	requestBody := map[string]interface{}{
		"model":      provider.model,
		"max_tokens": 4096,
		"system":     receiptAnalysisSystemPrompt,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []interface{}{
					contentBlock,
					map[string]interface{}{
						"type": "text",
						"text": receiptAnalysisUserPrompt,
					},
				},
			},
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vision request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create vision request: %w", err)
	}

	httpReq.Header.Set("x-api-key", provider.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude Vision API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read vision response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return nil, fmt.Errorf("Claude Vision API error (status %d): %v", resp.StatusCode, errResp)
	}

	// Parse Claude response
	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Claude Vision response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("no content in Claude Vision response")
	}

	// Extract JSON from response (Claude might wrap it in markdown code blocks)
	responseText := apiResp.Content[0].Text
	responseText = strings.TrimSpace(responseText)
	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
		responseText = strings.TrimSpace(responseText)
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
		responseText = strings.TrimSpace(responseText)
	}

	var result ReceiptAnalysisResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		s.logger.Error("failed to parse receipt analysis JSON", err, "raw_response", responseText)
		return nil, fmt.Errorf("failed to parse receipt analysis result: %w", err)
	}

	return &result, nil
}
