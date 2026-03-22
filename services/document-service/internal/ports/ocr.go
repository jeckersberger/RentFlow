package ports

import "context"

// OCRResult contains the result of OCR processing
type OCRResult struct {
	Text       string
	Confidence float64
	Language   string
}

// OCRProcessor defines the interface for OCR operations
type OCRProcessor interface {
	ExtractText(ctx context.Context, filePath string) (*OCRResult, error)
}
