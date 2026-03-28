package domain

import "errors"

// Sentinel errors for the document domain.
var (
	ErrTemplateNotFound   = errors.New("document template not found")
	ErrDocumentNotFound   = errors.New("document not found")
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrNameRequired       = errors.New("name is required")
	ErrTypeRequired       = errors.New("type is required")
	ErrTitleRequired      = errors.New("title is required")
	ErrFileNameRequired   = errors.New("file_name is required")
	ErrFilePathRequired   = errors.New("file_path is required")
	ErrReferenceRequired  = errors.New("reference_id and reference_type are required")
)
