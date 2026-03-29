package domain

import "errors"

var (
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrInvoiceNotDraft     = errors.New("invoice is not in draft status")
	ErrInvoiceAlreadyPaid  = errors.New("invoice is already fully paid")
	ErrInvalidInvoiceType  = errors.New("invalid invoice type")
	ErrPaymentExceedsTotal = errors.New("payment exceeds remaining balance")
	ErrItemNotFound        = errors.New("invoice item not found")

	ErrQuoteNotFound         = errors.New("quote not found")
	ErrQuoteNotDraft         = errors.New("quote is not in draft status")
	ErrQuoteAlreadyConverted = errors.New("quote has already been converted to an invoice")

	ErrDunningConfigNotFound = errors.New("dunning config not found")
	ErrInvoiceNotOverdue     = errors.New("invoice is not overdue")

	ErrBankTxNotFound       = errors.New("bank transaction not found")
	ErrBankTxAlreadyMatched = errors.New("bank transaction is already matched")
	ErrCSVParseError        = errors.New("CSV-Datei konnte nicht gelesen werden")
	ErrInvalidPercentage    = errors.New("Prozentsatz muss zwischen 1 und 100 liegen")
)
