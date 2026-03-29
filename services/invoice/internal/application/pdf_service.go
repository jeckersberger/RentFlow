package application

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf/v2"

	"github.com/jeckersberger/EquipFlow/services/invoice/internal/domain"
)

// InvoicePDFData holds all data needed for PDF generation.
type InvoicePDFData struct {
	Invoice  *domain.Invoice
	Items    []*domain.InvoiceItem
	Company  CompanyInfo
}

// CompanyInfo holds the company details for the PDF header.
type CompanyInfo struct {
	Name          string
	Street        string
	City          string
	Phone         string
	Email         string
	Website       string
	TaxNumber     string
	IBAN          string
	BIC           string
	BankName      string
}

func centsToEur(cents int64) string {
	eur := float64(cents) / 100.0
	s := fmt.Sprintf("%.2f", eur)
	s = strings.ReplaceAll(s, ".", ",")
	return s + " EUR"
}

func formatDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return dateStr
		}
	}
	return t.Format("02.01.2006")
}

// GenerateInvoicePDF creates a professional invoice PDF and writes it to w.
func GenerateInvoicePDF(w io.Writer, data InvoicePDFData) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()

	// --- Company Header ---
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(0, 90, 170)
	pdf.CellFormat(0, 10, encode(data.Company.Name), "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(100, 100, 100)
	companyLine := fmt.Sprintf("%s | %s | %s | %s",
		data.Company.Street, data.Company.City, data.Company.Phone, data.Company.Email)
	pdf.CellFormat(0, 4, encode(companyLine), "", 1, "L", false, 0, "")
	pdf.Ln(8)

	// --- Customer Address ---
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 5, encode(data.Invoice.CustomerName), "", 1, "L", false, 0, "")
	for _, line := range strings.Split(data.Invoice.CustomerAddress, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			pdf.CellFormat(0, 5, encode(line), "", 1, "L", false, 0, "")
		}
	}
	// Also try comma-separated address
	if !strings.Contains(data.Invoice.CustomerAddress, "\n") && strings.Contains(data.Invoice.CustomerAddress, ",") {
		// Already printed above as single line, that's fine
	}
	pdf.Ln(10)

	// --- Invoice Title ---
	title := "Rechnung"
	switch data.Invoice.InvoiceType {
	case "credit_note":
		title = "Gutschrift"
	case "proforma":
		title = "Proforma-Rechnung"
	case "partial":
		title = "Teilrechnung"
	case "advance":
		title = "Abschlagsrechnung"
	}
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 10, encode(title), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// --- Invoice Meta ---
	pdf.SetFont("Helvetica", "", 10)
	metaLeft := 40.0

	pdf.CellFormat(metaLeft, 6, "Rechnungsnr.:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(60, 6, data.Invoice.InvoiceNumber, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)

	pdf.CellFormat(30, 6, "Datum:", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, formatDate(data.Invoice.InvoiceDate), "", 1, "L", false, 0, "")

	if data.Invoice.DueDate != "" {
		pdf.CellFormat(metaLeft, 6, encode("Faellig bis:"), "", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, formatDate(data.Invoice.DueDate), "", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, "Status:", "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, encode(statusLabel(data.Invoice.Status)), "", 1, "L", false, 0, "")
	}
	pdf.Ln(8)

	// --- Items Table ---
	pdf.SetFillColor(0, 90, 170)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 9)

	colW := []float64{10, 80, 20, 20, 25, 35}
	headers := []string{"Pos", "Beschreibung", "Menge", "Einheit", "Einzelpreis", "Gesamt"}
	for i, h := range headers {
		pdf.CellFormat(colW[i], 8, encode(h), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 9)
	fill := false

	for _, item := range data.Items {
		if fill {
			pdf.SetFillColor(245, 245, 250)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		lineTotal := item.Quantity * item.UnitPrice
		pdf.CellFormat(colW[0], 7, fmt.Sprintf("%d", item.Position), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colW[1], 7, encode(item.Description), "1", 0, "L", fill, 0, "")
		pdf.CellFormat(colW[2], 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colW[3], 7, encode(item.Unit), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colW[4], 7, centsToEur(item.UnitPrice), "1", 0, "R", fill, 0, "")
		pdf.CellFormat(colW[5], 7, centsToEur(lineTotal), "1", 0, "R", fill, 0, "")
		pdf.Ln(-1)
		fill = !fill
	}

	pdf.Ln(4)

	// --- Totals ---
	totalsX := 130.0
	totalsW := 35.0
	valW := 35.0

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(totalsX)
	pdf.CellFormat(totalsW, 6, "Netto:", "", 0, "R", false, 0, "")
	pdf.CellFormat(valW, 6, centsToEur(data.Invoice.TotalNet), "", 1, "R", false, 0, "")

	if data.Invoice.Kleinunternehmer {
		pdf.SetX(totalsX)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.CellFormat(totalsW+valW, 5, encode("Gem. §19 UStG wird keine USt. berechnet"), "", 1, "R", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
	} else {
		vatPct := float64(data.Invoice.VatRate) / 100.0
		pdf.SetX(totalsX)
		pdf.CellFormat(totalsW, 6, fmt.Sprintf("MwSt. (%.0f%%):", vatPct), "", 0, "R", false, 0, "")
		pdf.CellFormat(valW, 6, centsToEur(data.Invoice.TotalVat), "", 1, "R", false, 0, "")
	}

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetX(totalsX)
	pdf.CellFormat(totalsW, 8, "Gesamt:", "", 0, "R", false, 0, "")
	pdf.CellFormat(valW, 8, centsToEur(data.Invoice.TotalGross), "", 1, "R", false, 0, "")

	if data.Invoice.AmountPaid > 0 {
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetX(totalsX)
		pdf.CellFormat(totalsW, 6, "Bezahlt:", "", 0, "R", false, 0, "")
		pdf.CellFormat(valW, 6, centsToEur(data.Invoice.AmountPaid), "", 1, "R", false, 0, "")

		remaining := data.Invoice.TotalGross - data.Invoice.AmountPaid
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetX(totalsX)
		pdf.CellFormat(totalsW, 6, "Restbetrag:", "", 0, "R", false, 0, "")
		pdf.CellFormat(valW, 6, centsToEur(remaining), "", 1, "R", false, 0, "")
	}

	// --- Notes ---
	if data.Invoice.Notes != "" {
		pdf.Ln(8)
		pdf.SetFont("Helvetica", "I", 9)
		pdf.SetTextColor(80, 80, 80)
		pdf.MultiCell(0, 5, encode(data.Invoice.Notes), "", "L", false)
	}

	// --- Footer: Bank Details ---
	pdf.Ln(10)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(100, 100, 100)

	if data.Company.IBAN != "" {
		bankLine := fmt.Sprintf("Bankverbindung: %s | IBAN: %s", data.Company.BankName, data.Company.IBAN)
		if data.Company.BIC != "" {
			bankLine += " | BIC: " + data.Company.BIC
		}
		pdf.CellFormat(0, 4, encode(bankLine), "", 1, "C", false, 0, "")
	}
	if data.Company.TaxNumber != "" {
		pdf.CellFormat(0, 4, encode("Steuernummer: "+data.Company.TaxNumber), "", 1, "C", false, 0, "")
	}

	return pdf.Output(w)
}

func statusLabel(s string) string {
	switch s {
	case "draft":
		return "Entwurf"
	case "finalized":
		return "Finalisiert"
	case "sent":
		return "Versendet"
	case "paid":
		return "Bezahlt"
	case "partial_paid":
		return "Teilweise bezahlt"
	case "overdue":
		return "Ueberfaellig"
	case "cancelled":
		return "Storniert"
	default:
		return s
	}
}

// encode replaces special chars for PDF (gofpdf uses ISO-8859-1 by default).
func encode(s string) string {
	r := strings.NewReplacer(
		"\u00e4", "ae", "\u00f6", "oe", "\u00fc", "ue",
		"\u00c4", "Ae", "\u00d6", "Oe", "\u00dc", "Ue",
		"\u00df", "ss",
		"\u20ac", "EUR",
	)
	return r.Replace(s)
}
