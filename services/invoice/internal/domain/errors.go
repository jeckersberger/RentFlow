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
)
