package domain

import "errors"

var (
	ErrCustomerNotFound = errors.New("customer not found")
	ErrContactNotFound  = errors.New("contact not found")
	ErrNoteNotFound     = errors.New("contact note not found")
)
