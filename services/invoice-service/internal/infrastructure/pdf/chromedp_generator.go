package pdf

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

// InvoiceData enthält die Daten für die PDF-Generierung.
// Wird aus dem HTML-Template extrahiert oder direkt übergeben.
type InvoiceData struct {
	InvoiceNumber string
	ClientName    string
	ClientStreet  string
	ClientCity    string
	ClientCountry string
	ClientEmail   string
	IssueDate     string
	DueDate       string
	Status        string
	Currency      string
	Items         []InvoiceItemData
	SubTotal      float64
	TaxRate       float64
	TaxAmount     float64
	Total         float64
	Notes         string
	CompanyName   string
	CompanyAddr   string
}

// InvoiceItemData enthält die Daten einer Rechnungsposition.
type InvoiceItemData struct {
	Description string
	Quantity    float64
	Unit        string
	UnitPrice   float64
	TotalPrice  float64
}

// FPDFGenerator generiert PDFs mittels go-pdf/fpdf (reine Go-Bibliothek).
// Benötigt kein Chrome/Chromium – läuft überall wo Go läuft.
type FPDFGenerator struct{}

// NewFPDFGenerator erstellt einen neuen PDF-Generator.
func NewFPDFGenerator() *FPDFGenerator {
	return &FPDFGenerator{}
}

// GeneratePDF konvertiert einen HTML-String in ein PDF-Byte-Array.
// Da fpdf kein HTML parst, wird das HTML ignoriert und stattdessen
// ein professionelles PDF direkt aus den extrahierten Daten erzeugt.
// Für Phase 1 ist dies ein Fallback – das eigentliche PDF wird über
// GenerateInvoicePDFFromData erzeugt.
func (g *FPDFGenerator) GeneratePDF(_ context.Context, html string) ([]byte, error) {
	// Einfaches Fallback: HTML-Inhalt als Text-PDF rendern
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Titel
	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 12, "EquipFlow - Dokument")
	pdf.Ln(20)

	// HTML-Tags entfernen und Text rendern
	text := stripHTMLTags(html)
	pdf.SetFont("Helvetica", "", 10)
	pdf.MultiCell(0, 5, text, "", "", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("PDF-Generierung fehlgeschlagen: %w", err)
	}
	return buf.Bytes(), nil
}

// GenerateInvoicePDFFromData erzeugt ein professionelles Rechnungs-PDF
// direkt aus strukturierten Daten.
func (g *FPDFGenerator) GenerateInvoicePDFFromData(data *InvoiceData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 25)
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// ===== Header: Firmenname + Rechnungstitel =====
	pdf.SetFont("Helvetica", "B", 24)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(0, 12, "RECHNUNG")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(102, 102, 102)
	pdf.Cell(0, 5, fmt.Sprintf("Rechnungsnummer: %s", data.InvoiceNumber))
	pdf.Ln(12)

	// Trennlinie
	pdf.SetDrawColor(51, 51, 51)
	pdf.SetLineWidth(0.5)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(10)

	// ===== Empfänger- und Rechnungsdetails =====
	startY := pdf.GetY()

	// Linke Spalte: Empfänger
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(102, 102, 102)
	pdf.Cell(90, 4, "RECHNUNGSEMPFAENGER:")
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(90, 5, data.ClientName)
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(80, 80, 80)
	if data.ClientStreet != "" {
		pdf.Cell(90, 5, data.ClientStreet)
		pdf.Ln(5)
	}
	if data.ClientCity != "" {
		pdf.Cell(90, 5, data.ClientCity)
		pdf.Ln(5)
	}
	if data.ClientCountry != "" {
		pdf.Cell(90, 5, data.ClientCountry)
		pdf.Ln(5)
	}
	if data.ClientEmail != "" {
		pdf.Cell(90, 5, data.ClientEmail)
		pdf.Ln(5)
	}

	endLeftY := pdf.GetY()

	// Rechte Spalte: Rechnungsdetails
	pdf.SetY(startY)
	pdf.SetX(110)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(102, 102, 102)
	pdf.CellFormat(80, 4, "RECHNUNGSDETAILS:", "", 0, "", false, 0, "")
	pdf.Ln(6)

	pdf.SetX(110)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(30, 5, "Datum:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(50, 5, data.IssueDate, "", 0, "", false, 0, "")
	pdf.Ln(6)

	pdf.SetX(110)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(30, 5, "Faellig bis:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(50, 5, data.DueDate, "", 0, "", false, 0, "")
	pdf.Ln(6)

	pdf.SetX(110)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(30, 5, "Status:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(50, 5, data.Status, "", 0, "", false, 0, "")
	pdf.Ln(6)

	// Zum tiefsten Punkt beider Spalten springen
	if endLeftY > pdf.GetY() {
		pdf.SetY(endLeftY)
	}
	pdf.Ln(10)

	// ===== Positionstabelle =====
	// Tabellenkopf
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(80, 8, "Beschreibung", "B", 0, "", true, 0, "")
	pdf.CellFormat(25, 8, "Menge", "B", 0, "R", true, 0, "")
	pdf.CellFormat(30, 8, "Einzelpreis", "B", 0, "R", true, 0, "")
	pdf.CellFormat(35, 8, "Gesamt", "B", 0, "R", true, 0, "")
	pdf.Ln(-1)

	// Tabellenzeilen
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(80, 80, 80)
	for _, item := range data.Items {
		pdf.CellFormat(80, 7, item.Description, "B", 0, "", false, 0, "")
		qtyStr := fmt.Sprintf("%.2f %s", item.Quantity, item.Unit)
		pdf.CellFormat(25, 7, qtyStr, "B", 0, "R", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%s %.2f", data.Currency, item.UnitPrice), "B", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, fmt.Sprintf("%s %.2f", data.Currency, item.TotalPrice), "B", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(8)

	// ===== Summenbereich (rechts) =====
	pdf.SetX(110)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(45, 7, "Zwischensumme:", "", 0, "R", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(35, 7, fmt.Sprintf("%s %.2f", data.Currency, data.SubTotal), "", 0, "R", false, 0, "")
	pdf.Ln(7)

	pdf.SetX(110)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(45, 7, fmt.Sprintf("MwSt. (%.0f%%):", data.TaxRate), "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("%s %.2f", data.Currency, data.TaxAmount), "", 0, "R", false, 0, "")
	pdf.Ln(8)

	// Trennlinie über Gesamtsumme
	pdf.SetX(110)
	pdf.Line(110, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(3)

	pdf.SetX(110)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(51, 51, 51)
	pdf.CellFormat(45, 8, "GESAMT:", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("%s %.2f", data.Currency, data.Total), "", 0, "R", false, 0, "")
	pdf.Ln(20)

	// ===== Footer: Zahlungsbedingungen =====
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(102, 102, 102)
	pdf.Cell(0, 5, fmt.Sprintf("Zahlungsbedingungen: Faellig bis %s", data.DueDate))
	pdf.Ln(6)

	if data.Notes != "" {
		pdf.SetFont("Helvetica", "", 9)
		pdf.Cell(0, 5, fmt.Sprintf("Hinweise: %s", data.Notes))
		pdf.Ln(6)
	}

	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(153, 153, 153)
	pdf.Cell(0, 4, "Diese Rechnung wurde automatisch von EquipFlow erstellt.")
	pdf.Ln(4)
	pdf.Cell(0, 4, "Bei Fragen wenden Sie sich bitte an unsere Buchhaltung.")

	// PDF in Buffer schreiben
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("PDF-Generierung fehlgeschlagen: %w", err)
	}

	return buf.Bytes(), nil
}

// stripHTMLTags entfernt HTML-Tags aus einem String (einfacher Fallback)
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			result.WriteRune(' ')
		case !inTag:
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}
