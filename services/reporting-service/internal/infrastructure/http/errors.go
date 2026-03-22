package http

import "errors"

var (
	ErrMissingTenantID = errors.New("X-Tenant-ID header is required")
	ErrMissingUserID   = errors.New("X-User-ID header is required")
)
