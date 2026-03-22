package ocr

import (
	"context"
	"os"
	"os/exec"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/document-service/internal/ports"
)

// TesseractProcessor uses tesseract to extract text from images
type TesseractProcessor struct {
	logger logger.Logger
}

// NoopOCRProcessor is a fallback that returns empty results
type NoopOCRProcessor struct {
	logger logger.Logger
}

// NewTesseractProcessor creates a new tesseract processor if tesseract is available
func NewTesseractProcessor(log logger.Logger) ports.OCRProcessor {
	// Check if tesseract binary is available
	_, err := exec.LookPath("tesseract")
	if err != nil {
		log.Warn("Tesseract binary not found, using no-op OCR processor")
		return &NoopOCRProcessor{logger: log}
	}

	return &TesseractProcessor{logger: log}
}

// ExtractText extracts text from an image file using tesseract
func (t *TesseractProcessor) ExtractText(ctx context.Context, filePath string) (*ports.OCRResult, error) {
	// Check file exists
	if _, err := os.Stat(filePath); err != nil {
		t.logger.Error("File not found for OCR", err, "path", filePath)
		return nil, err
	}

	// Create output file path (without extension)
	outputFile := filePath + ".txt"

	// Run tesseract command
	cmd := exec.CommandContext(ctx, "tesseract", filePath, outputFile)
	if err := cmd.Run(); err != nil {
		t.logger.Error("Tesseract command failed", err, "path", filePath)
		return nil, err
	}

	// Read the output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.logger.Error("Failed to read OCR output", err, "output_file", outputFile)
		return nil, err
	}

	// Clean up the output file
	_ = os.Remove(outputFile)

	result := &ports.OCRResult{
		Text:       string(content),
		Confidence: 0.95, // Default confidence for tesseract
		Language:   "eng",
	}

	t.logger.Info("OCR extraction completed", "path", filePath, "text_length", len(result.Text))
	return result, nil
}

// ExtractText returns an empty result with a log message
func (n *NoopOCRProcessor) ExtractText(ctx context.Context, filePath string) (*ports.OCRResult, error) {
	n.logger.Info("No-op OCR processor: skipping text extraction", "path", filePath)
	return &ports.OCRResult{
		Text:       "",
		Confidence: 0,
		Language:   "",
	}, nil
}
