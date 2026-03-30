package application

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/document/internal/domain"
	"github.com/jeckersberger/EquipFlow/services/document/templates"
)

// validTemplateTypes maps template type keys to their embedded file names.
var validTemplateTypes = map[string]string{
	"invoice":       "invoice.html",
	"quote":         "quote.html",
	"delivery_note": "delivery_note.html",
	"reminder":      "reminder.html",
}

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// RenderRequest holds the data for rendering a document template.
type RenderRequest struct {
	TemplateType string                 `json:"template_type"`
	Data         map[string]interface{} `json:"data"`
	CustomCSS    string                 `json:"custom_css,omitempty"`
}

// RenderResult contains the rendered HTML output.
type RenderResult struct {
	HTML         string `json:"html"`
	TemplateType string `json:"template_type"`
	IsCustom     bool   `json:"is_custom"`
}

// TemplateTypeInfo describes an available default template type.
type TemplateTypeInfo struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// TemplateRenderer renders document templates with support for tenant-specific
// custom templates that fall back to embedded defaults.
type TemplateRenderer struct {
	templateRepo domain.DocumentTemplateRepository
	logger       zerolog.Logger
}

// NewTemplateRenderer constructs a new TemplateRenderer.
func NewTemplateRenderer(
	templateRepo domain.DocumentTemplateRepository,
	logger zerolog.Logger,
) *TemplateRenderer {
	return &TemplateRenderer{
		templateRepo: templateRepo,
		logger:       logger.With().Str("service", "template_renderer").Logger(),
	}
}

// Render loads a custom template for the tenant if one exists, otherwise falls
// back to the embedded default, then executes the template with the given data.
func (r *TemplateRenderer) Render(
	ctx context.Context,
	tenantID uuid.UUID,
	req RenderRequest,
) (*RenderResult, error) {
	if req.TemplateType == "" {
		return nil, fmt.Errorf("template_type is required")
	}

	if _, ok := validTemplateTypes[req.TemplateType]; !ok {
		return nil, fmt.Errorf("unsupported template type: %s", req.TemplateType)
	}

	// Try to load a tenant-specific custom template.
	isCustom := false
	htmlContent, err := r.loadCustomTemplate(ctx, tenantID, req.TemplateType)
	if err != nil {
		r.logger.Debug().Err(err).
			Str("template_type", req.TemplateType).
			Msg("no custom template found, using default")
	}
	if htmlContent != "" {
		isCustom = true
	} else {
		// Fall back to the embedded default.
		htmlContent, err = GetDefaultTemplate(req.TemplateType)
		if err != nil {
			return nil, fmt.Errorf("load default template: %w", err)
		}
	}

	// If custom CSS is provided, inject it before </head>.
	if req.CustomCSS != "" {
		cssTag := "<style>\n" + req.CustomCSS + "\n</style>\n</head>"
		htmlContent = strings.Replace(htmlContent, "</head>", cssTag, 1)
	}

	// Parse and execute the template.
	tmpl, err := template.New(req.TemplateType).Parse(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.Data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return &RenderResult{
		HTML:         buf.String(),
		TemplateType: req.TemplateType,
		IsCustom:     isCustom,
	}, nil
}

// GetDefaultTemplate reads an embedded default template by type name.
func GetDefaultTemplate(templateType string) (string, error) {
	fileName, ok := validTemplateTypes[templateType]
	if !ok {
		return "", fmt.Errorf("unsupported template type: %s", templateType)
	}

	data, err := templates.DefaultFS.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("read embedded template %s: %w", fileName, err)
	}

	return string(data), nil
}

// ListTemplateTypes returns all available default template types.
func ListTemplateTypes() []TemplateTypeInfo {
	return []TemplateTypeInfo{
		{
			Type:        "invoice",
			Name:        "Rechnung",
			Description: "Professionelle Rechnungsvorlage mit Positionen, MwSt. und Bankdaten",
		},
		{
			Type:        "quote",
			Name:        "Angebot",
			Description: "Angebotsvorlage mit Gueltigkeitsdauer und Einleitungs-/Schlusstext",
		},
		{
			Type:        "delivery_note",
			Name:        "Lieferschein",
			Description: "Lieferschein mit Equipment-Liste, Barcodes und Unterschriftsfeld",
		},
		{
			Type:        "reminder",
			Name:        "Mahnung",
			Description: "Mahnungsvorlage mit Mahnstufen (0-2), Gebuehren und Fristen",
		},
	}
}

// GetSampleData returns sample data for previewing a template type.
func GetSampleData(templateType string) map[string]interface{} {
	company := map[string]interface{}{
		"Name":    "Muster Veranstaltungstechnik GmbH",
		"Street":  "Musterstrasse 42",
		"Zip":     "80331",
		"City":    "Muenchen",
		"Phone":   "+49 89 1234567",
		"Email":   "info@muster-technik.de",
		"Website": "www.muster-technik.de",
		"TaxID":   "143/123/12345",
		"VatID":   "DE123456789",
		"LogoURL": "",
	}

	customer := map[string]interface{}{
		"Company": "Beispiel Events GmbH",
		"Name":    "Max Mustermann",
		"Street":  "Beispielweg 7",
		"Zip":     "10115",
		"City":    "Berlin",
		"Country": "Deutschland",
	}

	bank := map[string]interface{}{
		"AccountHolder": "Muster Veranstaltungstechnik GmbH",
		"IBAN":          "DE89 3704 0044 0532 0130 00",
		"BIC":           "COBADEFFXXX",
		"BankName":      "Commerzbank",
	}

	switch templateType {
	case "invoice":
		return map[string]interface{}{
			"Company":  company,
			"Customer": customer,
			"Bank":     bank,
			"Invoice": map[string]interface{}{
				"Number":         "RE-2026-0042",
				"Date":           "30.03.2026",
				"DueDate":        "13.04.2026",
				"CustomerNumber": "KD-1001",
				"ProjectName":    "Firmengala 2026",
				"IntroText":      "Vielen Dank fuer Ihren Auftrag. Wir berechnen Ihnen folgende Leistungen:",
				"OutroText":      "Bitte ueberweisen Sie den Betrag unter Angabe der Rechnungsnummer bis zum Faelligkeitsdatum.",
			},
			"Items": []map[string]interface{}{
				{"Position": 1, "Description": "PA-Anlage JBL VTX V25-II", "SubText": "inkl. Zubehoer und Verkabelung", "Quantity": "2", "Unit": "Stk", "UnitPrice": "450,00 EUR", "Total": "900,00 EUR"},
				{"Position": 2, "Description": "Lichtanlage MA Lighting grandMA3", "SubText": "", "Quantity": "1", "Unit": "Stk", "UnitPrice": "320,00 EUR", "Total": "320,00 EUR"},
				{"Position": 3, "Description": "Techniker vor Ort", "SubText": "Auf- und Abbau, Betreuung", "Quantity": "8", "Unit": "Std", "UnitPrice": "65,00 EUR", "Total": "520,00 EUR"},
			},
			"Totals": map[string]interface{}{
				"Net":             "1.740,00 EUR",
				"VatRate":         "19%",
				"Vat":             "330,60 EUR",
				"Gross":           "2.070,60 EUR",
				"IsSmallBusiness": false,
				"Paid":            "",
				"Remaining":       "",
			},
		}

	case "quote":
		return map[string]interface{}{
			"Company":  company,
			"Customer": customer,
			"Quote": map[string]interface{}{
				"Number":         "AN-2026-0015",
				"Date":           "30.03.2026",
				"ValidUntil":     "30.04.2026",
				"CustomerNumber": "KD-1001",
				"ProjectName":    "Sommerfest 2026",
				"IntroText":      "Vielen Dank fuer Ihre Anfrage. Gerne unterbreiten wir Ihnen folgendes Angebot:",
				"OutroText":      "Wir freuen uns auf Ihre Rueckmeldung und stehen bei Fragen gerne zur Verfuegung.",
			},
			"Items": []map[string]interface{}{
				{"Position": 1, "Description": "Buehne 8x6m", "SubText": "inkl. Aufbau und Abbau", "Quantity": "1", "Unit": "Stk", "UnitPrice": "1.200,00 EUR", "Total": "1.200,00 EUR"},
				{"Position": 2, "Description": "LED-Wand 4x3m", "SubText": "Indoor P3.9", "Quantity": "1", "Unit": "Stk", "UnitPrice": "800,00 EUR", "Total": "800,00 EUR"},
				{"Position": 3, "Description": "Beschallungsanlage", "SubText": "fuer bis zu 500 Personen", "Quantity": "1", "Unit": "Psch", "UnitPrice": "650,00 EUR", "Total": "650,00 EUR"},
			},
			"Totals": map[string]interface{}{
				"Net":             "2.650,00 EUR",
				"VatRate":         "19%",
				"Vat":             "503,50 EUR",
				"Gross":           "3.153,50 EUR",
				"IsSmallBusiness": false,
			},
		}

	case "delivery_note":
		return map[string]interface{}{
			"Company":  company,
			"Customer": customer,
			"DeliveryNote": map[string]interface{}{
				"Number":       "LS-2026-0028",
				"Date":         "30.03.2026",
				"Time":         "09:00 Uhr",
				"OrderNumber":  "RE-2026-0042",
				"ProjectName":  "Firmengala 2026",
				"ProjectStart": "01.04.2026",
				"ProjectEnd":   "02.04.2026",
				"Venue":        "Messehalle 3, Muenchen",
				"TotalItems":   4,
				"Notes":        "Bitte Equipment nach Veranstaltungsende am Buehnenhintereingang bereitstellen.",
			},
			"Items": []map[string]interface{}{
				{"Position": 1, "Name": "JBL VTX V25-II Lautsprecher", "Quantity": "2x", "Barcode": "EQ-2026-00142"},
				{"Position": 2, "Name": "JBL VTX S28 Subwoofer", "Quantity": "2x", "Barcode": "EQ-2026-00143"},
				{"Position": 3, "Name": "grandMA3 Light Pult", "Quantity": "1x", "Barcode": "EQ-2026-00087"},
				{"Position": 4, "Name": "Moving Head Robe BMFL Spot", "Quantity": "8x", "Barcode": "EQ-2026-00201"},
			},
		}

	case "reminder":
		return map[string]interface{}{
			"Company":  company,
			"Customer": customer,
			"Bank":     bank,
			"Reminder": map[string]interface{}{
				"Title":                "1. Mahnung",
				"Level":               1,
				"Date":                "30.03.2026",
				"NewDueDate":          "13.04.2026",
				"CustomerNumber":      "KD-1001",
				"InvoiceNumber":       "RE-2026-0038",
				"InvoiceDate":         "01.03.2026",
				"OriginalDueDate":     "15.03.2026",
				"InvoiceAmount":       "2.070,60 EUR",
				"OutstandingAmount":   "2.070,60 EUR",
				"Fee":                 "5,00 EUR",
				"Interest":            "",
				"TotalDue":            "2.075,60 EUR",
				"PreviousReminderDate": "20.03.2026",
			},
		}

	default:
		return map[string]interface{}{}
	}
}

// loadCustomTemplate attempts to find an active custom template for the given
// tenant and template type. It returns the template content or an error.
func (r *TemplateRenderer) loadCustomTemplate(
	ctx context.Context,
	tenantID uuid.UUID,
	templateType string,
) (string, error) {
	tpl, err := r.templateRepo.GetByType(ctx, tenantID, templateType)
	if err != nil {
		return "", err
	}
	if tpl == nil || !tpl.IsActive || tpl.Content == "" {
		return "", fmt.Errorf("no active custom template for type %s", templateType)
	}
	return tpl.Content, nil
}
