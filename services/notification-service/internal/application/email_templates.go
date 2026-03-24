package application

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

// --- Template Data Structs ---

// BookingRequestData holds data for the booking request email template
type BookingRequestData struct {
	RecipientName string
	ProjectName   string
	Dates         string
	Link          string
	SenderName    string
	CompanyName   string
}

// InvitationData holds data for the invitation email template
type InvitationData struct {
	RecipientEmail string
	InviterName    string
	CompanyName    string
	Link           string
}

// PasswordResetData holds data for the password reset email template
type PasswordResetData struct {
	RecipientName string
	Link          string
	CompanyName   string
	ExpiresIn     string
}

// --- HTML Templates ---
// The base layout uses %%s as a placeholder for content (fmt.Sprintf),
// while {{.Field}} is for Go template variables.

const emailBaseLayout = `<!DOCTYPE html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#f4f5f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f4f5f7;padding:40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <!-- Header with gradient -->
          <tr>
            <td style="background:linear-gradient(135deg,#06b6d4,#0891b2);padding:32px 40px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:700;letter-spacing:-0.5px;">RentFlow</h1>
            </td>
          </tr>
          <!-- Content -->
          <tr>
            <td style="padding:40px;">
              %s
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding:24px 40px;background-color:#f8fafc;border-top:1px solid #e2e8f0;text-align:center;">
              <p style="margin:0;color:#94a3b8;font-size:12px;line-height:1.5;">
                Diese E-Mail wurde automatisch von RentFlow versendet.<br>
                &copy; {{.Year}} {{.CompanyName}}
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

const bookingRequestContent = `
<h2 style="margin:0 0 16px;color:#1e293b;font-size:20px;font-weight:600;">Neue Buchungsanfrage</h2>
<p style="margin:0 0 12px;color:#475569;font-size:15px;line-height:1.6;">
  Hallo {{.RecipientName}},
</p>
<p style="margin:0 0 20px;color:#475569;font-size:15px;line-height:1.6;">
  {{.SenderName}} hat Sie fuer das folgende Projekt angefragt:
</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f0fdfa;border:1px solid #99f6e4;border-radius:8px;margin-bottom:24px;">
  <tr>
    <td style="padding:20px;">
      <p style="margin:0 0 8px;color:#0f766e;font-size:16px;font-weight:600;">{{.ProjectName}}</p>
      <p style="margin:0;color:#0d9488;font-size:14px;">Zeitraum: {{.Dates}}</p>
    </td>
  </tr>
</table>
<table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 16px;">
  <tr>
    <td style="background:linear-gradient(135deg,#06b6d4,#0891b2);border-radius:8px;">
      <a href="{{.Link}}" style="display:inline-block;padding:14px 32px;color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;">
        Anfrage beantworten
      </a>
    </td>
  </tr>
</table>
<p style="margin:0;color:#94a3b8;font-size:13px;text-align:center;">
  Oder kopieren Sie diesen Link: <a href="{{.Link}}" style="color:#0891b2;">{{.Link}}</a>
</p>`

const invitationContent = `
<h2 style="margin:0 0 16px;color:#1e293b;font-size:20px;font-weight:600;">Einladung zu RentFlow</h2>
<p style="margin:0 0 20px;color:#475569;font-size:15px;line-height:1.6;">
  {{.InviterName}} hat Sie eingeladen, {{.CompanyName}} auf RentFlow beizutreten.
</p>
<p style="margin:0 0 24px;color:#475569;font-size:15px;line-height:1.6;">
  Erstellen Sie jetzt Ihr Konto, um loszulegen:
</p>
<table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 16px;">
  <tr>
    <td style="background:linear-gradient(135deg,#06b6d4,#0891b2);border-radius:8px;">
      <a href="{{.Link}}" style="display:inline-block;padding:14px 32px;color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;">
        Konto erstellen
      </a>
    </td>
  </tr>
</table>
<p style="margin:0;color:#94a3b8;font-size:13px;text-align:center;">
  Oder kopieren Sie diesen Link: <a href="{{.Link}}" style="color:#0891b2;">{{.Link}}</a>
</p>`

const passwordResetContent = `
<h2 style="margin:0 0 16px;color:#1e293b;font-size:20px;font-weight:600;">Passwort zuruecksetzen</h2>
<p style="margin:0 0 12px;color:#475569;font-size:15px;line-height:1.6;">
  Hallo {{.RecipientName}},
</p>
<p style="margin:0 0 24px;color:#475569;font-size:15px;line-height:1.6;">
  Sie haben eine Anfrage zum Zuruecksetzen Ihres Passworts gestellt.
  Klicken Sie auf den folgenden Button, um ein neues Passwort zu vergeben:
</p>
<table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 16px;">
  <tr>
    <td style="background:linear-gradient(135deg,#06b6d4,#0891b2);border-radius:8px;">
      <a href="{{.Link}}" style="display:inline-block;padding:14px 32px;color:#ffffff;text-decoration:none;font-size:15px;font-weight:600;">
        Passwort zuruecksetzen
      </a>
    </td>
  </tr>
</table>
<p style="margin:0 0 16px;color:#94a3b8;font-size:13px;text-align:center;">
  Oder kopieren Sie diesen Link: <a href="{{.Link}}" style="color:#0891b2;">{{.Link}}</a>
</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#fef3c7;border:1px solid #fde68a;border-radius:8px;margin-top:20px;">
  <tr>
    <td style="padding:16px;">
      <p style="margin:0;color:#92400e;font-size:13px;line-height:1.5;">
        Dieser Link ist {{.ExpiresIn}} gueltig. Falls Sie diese Anfrage nicht gestellt haben, koennen Sie diese E-Mail ignorieren.
      </p>
    </td>
  </tr>
</table>`

// RenderBookingRequestEmail renders the booking request email template
func RenderBookingRequestEmail(data BookingRequestData) (subject string, bodyHTML string, err error) {
	subject = "Buchungsanfrage: " + data.ProjectName
	bodyHTML, err = renderTemplate("booking_request", bookingRequestContent, data, data.CompanyName)
	return
}

// RenderInvitationEmail renders the invitation email template
func RenderInvitationEmail(data InvitationData) (subject string, bodyHTML string, err error) {
	subject = "Einladung zu RentFlow von " + data.InviterName
	bodyHTML, err = renderTemplate("invitation", invitationContent, data, data.CompanyName)
	return
}

// RenderPasswordResetEmail renders the password reset email template
func RenderPasswordResetEmail(data PasswordResetData) (subject string, bodyHTML string, err error) {
	subject = "Passwort zuruecksetzen - RentFlow"
	bodyHTML, err = renderTemplate("password_reset", passwordResetContent, data, data.CompanyName)
	return
}

// renderTemplate renders an email template with the base layout using a two-pass approach.
// Pass 1 renders the content block with the specific data fields.
// Pass 2 renders the full layout (with rendered content inserted) to fill in footer fields.
func renderTemplate(name string, content string, data interface{}, companyName string) (string, error) {
	// Pass 1: render the content part with the specific data
	contentTmpl, err := template.New(name + "_content").Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse content template: %w", err)
	}
	var contentBuf bytes.Buffer
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute content template: %w", err)
	}

	// Pass 2: insert rendered content into the base layout, then render footer fields
	layoutWithContent := fmt.Sprintf(emailBaseLayout, contentBuf.String())
	layoutTmpl, err := template.New(name + "_layout").Parse(layoutWithContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse layout template: %w", err)
	}

	footerData := struct {
		Year        int
		CompanyName string
	}{
		Year:        time.Now().Year(),
		CompanyName: companyName,
	}

	var buf bytes.Buffer
	if err := layoutTmpl.Execute(&buf, footerData); err != nil {
		return "", fmt.Errorf("failed to execute layout template: %w", err)
	}

	return buf.String(), nil
}
