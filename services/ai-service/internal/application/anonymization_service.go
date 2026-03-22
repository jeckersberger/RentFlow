package application

import (
	"regexp"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// AnonymizationService handles PII anonymization
type AnonymizationService struct {
	logger logger.Logger
}

// NewAnonymizationService creates a new anonymization service
func NewAnonymizationService(log logger.Logger) *AnonymizationService {
	return &AnonymizationService{
		logger: log,
	}
}

// Anonymize anonymizes text containing PII
func (s *AnonymizationService) Anonymize(text string) string {
	result := text

	// Name pattern: common German names (simplified)
	namePattern := regexp.MustCompile(`\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`)
	result = namePattern.ReplaceAllString(result, "[NAME]")

	// Email pattern
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	result = emailPattern.ReplaceAllString(result, "[EMAIL]")

	// IBAN pattern: German IBAN
	ibanPattern := regexp.MustCompile(`DE\d{20}`)
	result = ibanPattern.ReplaceAllString(result, "[IBAN]")

	// Phone pattern: German phone formats
	phonePattern := regexp.MustCompile(`(?:\+49|0)[1-9]\d{1,14}`)
	result = phonePattern.ReplaceAllString(result, "[PHONE]")

	// Address pattern: street + number
	addressPattern := regexp.MustCompile(`\b[A-Z][a-z]+(?:str(?:aße)?|weg|platz|allee)\s+\d+[a-z]?\b`)
	result = addressPattern.ReplaceAllString(result, "[ADDRESS]")

	// Tax ID pattern: German Steuernummer
	taxIdPattern := regexp.MustCompile(`(?:DE\d{9}|\d{2,3}/\d{3}/\d{4,5})`)
	result = taxIdPattern.ReplaceAllString(result, "[TAX_ID]")

	// Customer ID pattern
	customerIdPattern := regexp.MustCompile(`\bK-\d{4,8}\b`)
	result = customerIdPattern.ReplaceAllString(result, "[CUSTOMER_ID]")

	return result
}

// GetPatterns returns all anonymization patterns
func (s *AnonymizationService) GetPatterns() []domain.AnonymizationPattern {
	return []domain.AnonymizationPattern{
		{
			PatternType: domain.PatternTypeName,
			Pattern:     `\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`,
			Replacement: "[NAME]",
		},
		{
			PatternType: domain.PatternTypeEmail,
			Pattern:     `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			Replacement: "[EMAIL]",
		},
		{
			PatternType: domain.PatternTypeIBAN,
			Pattern:     `DE\d{20}`,
			Replacement: "[IBAN]",
		},
		{
			PatternType: domain.PatternTypePhone,
			Pattern:     `(?:\+49|0)[1-9]\d{1,14}`,
			Replacement: "[PHONE]",
		},
		{
			PatternType: domain.PatternTypeAddress,
			Pattern:     `\b[A-Z][a-z]+(?:str(?:aße)?|weg|platz|allee)\s+\d+[a-z]?\b`,
			Replacement: "[ADDRESS]",
		},
		{
			PatternType: domain.PatternTypeTaxID,
			Pattern:     `(?:DE\d{9}|\d{2,3}/\d{3}/\d{4,5})`,
			Replacement: "[TAX_ID]",
		},
		{
			PatternType: domain.PatternTypeCustomerID,
			Pattern:     `\bK-\d{4,8}\b`,
			Replacement: "[CUSTOMER_ID]",
		},
	}
}

// AnonymizeWithReport anonymizes text and returns a report
func (s *AnonymizationService) AnonymizeWithReport(text string) (string, map[string]int) {
	result := text
	report := make(map[string]int)

	patterns := s.GetPatterns()
	for _, p := range patterns {
		re := regexp.MustCompile(p.Pattern)
		matches := re.FindAllString(result, -1)
		if len(matches) > 0 {
			report[string(p.PatternType)] = len(matches)
			result = re.ReplaceAllString(result, p.Replacement)
		}
	}

	return result, report
}
