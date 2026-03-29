package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// InvoiceEmailRequest holds data needed to send an invoice email.
type InvoiceEmailRequest struct {
	To          string
	Subject     string
	HTMLBody    string
	PDFData     []byte
	PDFFilename string
	InvoiceID   uuid.UUID
	TenantID    uuid.UUID
}

// SendInvoiceEmail sends an invoice email via the notification service.
func (s *InvoiceService) SendInvoiceEmail(ctx context.Context, req InvoiceEmailRequest) error {
	// Call notification-service email endpoint
	payload := map[string]interface{}{
		"to":        req.To,
		"subject":   req.Subject,
		"html_body": req.HTMLBody,
		"attachments": []map[string]string{
			{
				"filename":  req.PDFFilename,
				"data":      base64.StdEncoding.EncodeToString(req.PDFData),
				"mime_type": "application/pdf",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal email payload: %w", err)
	}

	// Direct HTTP call to notification-service (same Docker network)
	notifURL := "http://cratedesk-notification:8015/api/v1/email/send"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", notifURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create email request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Forward auth token from context if available
	if token := ctx.Value("auth_token"); token != nil {
		httpReq.Header.Set("Authorization", "Bearer "+token.(string))
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("E-Mail-Versand fehlgeschlagen — Notification-Service nicht erreichbar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("E-Mail-Versand fehlgeschlagen (Status %d)", resp.StatusCode)
	}

	return nil
}

// UpdateStatus updates an invoice's status.
func (s *InvoiceService) UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	return s.invoiceRepo.UpdateStatus(ctx, id, tenantID, status)
}

// RenderInvoiceEmailHTML generates the HTML body for an invoice email.
func RenderInvoiceEmailHTML(invoice *domain.Invoice, companyName string) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; color: #333;">
  <div style="border-bottom: 3px solid #005ab6; padding-bottom: 16px; margin-bottom: 24px;">
    <h2 style="margin: 0; color: #005ab6;">{{.CompanyName}}</h2>
  </div>
  <p>Sehr geehrte/r {{.CustomerName}},</p>
  <p>anbei erhalten Sie unsere {{.DocType}} <strong>{{.InvoiceNumber}}</strong> vom {{.InvoiceDate}}.</p>
  <div style="background: #f8f9fa; border-radius: 8px; padding: 16px; margin: 20px 0;">
    <table style="width: 100%; border-collapse: collapse;">
      <tr><td style="padding: 4px 0; color: #666;">Nummer:</td><td style="padding: 4px 0; font-weight: bold; text-align: right;">{{.InvoiceNumber}}</td></tr>
      <tr><td style="padding: 4px 0; color: #666;">Datum:</td><td style="padding: 4px 0; text-align: right;">{{.InvoiceDate}}</td></tr>
      {{if .DueDate}}<tr><td style="padding: 4px 0; color: #666;">Faellig bis:</td><td style="padding: 4px 0; text-align: right;">{{.DueDate}}</td></tr>{{end}}
      <tr style="border-top: 1px solid #ddd;"><td style="padding: 8px 0 4px; font-weight: bold;">Gesamtbetrag:</td><td style="padding: 8px 0 4px; font-weight: bold; text-align: right; font-size: 1.2em; color: #005ab6;">{{.TotalGross}}</td></tr>
    </table>
  </div>
  <p>Das Dokument finden Sie als PDF im Anhang.</p>
  <p>Bei Fragen stehen wir Ihnen gerne zur Verfuegung.</p>
  <p>Mit freundlichen Gruessen<br><strong>{{.CompanyName}}</strong></p>
  <div style="border-top: 1px solid #eee; padding-top: 12px; margin-top: 24px; font-size: 12px; color: #999;">
    Diese E-Mail wurde automatisch von CrateDesk generiert.
  </div>
</body>
</html>`

	docType := "Rechnung"
	if invoice.InvoiceType == "credit_note" {
		docType = "Gutschrift"
	}

	data := struct {
		CompanyName   string
		CustomerName  string
		DocType       string
		InvoiceNumber string
		InvoiceDate   string
		DueDate       string
		TotalGross    string
	}{
		CompanyName:   companyName,
		CustomerName:  invoice.CustomerName,
		DocType:       docType,
		InvoiceNumber: invoice.InvoiceNumber,
		InvoiceDate:   formatDateDE(invoice.InvoiceDate),
		DueDate:       formatDateDE(invoice.DueDate),
		TotalGross:    centsToEur(invoice.TotalGross),
	}

	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatDateDE(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return s
		}
	}
	return t.Format("02.01.2006")
}

func centsToEurEmail(cents int64) string {
	eur := float64(cents) / 100.0
	s := fmt.Sprintf("%.2f", eur)
	s = strings.ReplaceAll(s, ".", ",")
	return s + " EUR"
}
