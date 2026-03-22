package domain

import "errors"

var (
	ErrAuditEntryNotFound = errors.New("audit entry not found")
	ErrExportNotFound     = errors.New("export not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrDatabaseError      = errors.New("database error")
	ErrChecksumMismatch   = errors.New("checksum mismatch detected")
	ErrInvalidExportType  = errors.New("invalid export type")
)
