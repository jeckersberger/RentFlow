package application

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PDFGenerator handles PDF generation for documents
type PDFGenerator struct{}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

// GenerateBasicPDF creates a minimal valid PDF file with the given title and metadata.
// Returns the generated PDF bytes.
func (pg *PDFGenerator) GenerateBasicPDF(title, documentNumber, author string) ([]byte, error) {
	// Minimal valid PDF structure (~200 bytes)
	// This is a valid PDF 1.4 with a single text object
	pdf := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 200 >>
stream
BT
/F1 12 Tf
50 750 Td
(%s) Tj
0 -20 Td
(Document Number: %s) Tj
0 -20 Td
(Generated: %s) Tj
0 -20 Td
(Author: %s) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
xref
0 6
0000000000 65535 f
0000000010 00000 n
0000000074 00000 n
0000000133 00000 n
0000000281 00000 n
0000000538 00000 n
trailer
<< /Size 6 /Root 1 0 R >>
startxref
632
%%%%EOF
`, title, documentNumber, time.Now().Format("2006-01-02"), author)

	return []byte(pdf), nil
}

// GenerateAndWritePDF generates a PDF and writes it to the specified file path
func (pg *PDFGenerator) GenerateAndWritePDF(filePath, title, documentNumber, author string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Generate PDF bytes
	pdfBytes, err := pg.GenerateBasicPDF(title, documentNumber, author)
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	return nil
}
