package domain

import "errors"

var (
	ErrPartnerNotFound         = errors.New("federation partner not found")
	ErrRequestNotFound         = errors.New("sub-rental request not found")
	ErrCertificateNotFound     = errors.New("certificate not found")
	ErrPartnerAlreadyExists    = errors.New("partner already exists")
	ErrInvalidPartnerStatus    = errors.New("invalid partner status")
	ErrInvalidTrustLevel       = errors.New("invalid trust level")
	ErrInvalidRequestStatus    = errors.New("invalid request status")
	ErrPartnerSuspended        = errors.New("partner is suspended")
	ErrPartnerRevoked          = errors.New("partner has been revoked")
	ErrInvalidCertificate      = errors.New("invalid certificate")
	ErrCertificateExpired      = errors.New("certificate expired")
	ErrDataSovereigntyViolated = errors.New("data sovereignty policy violated")
	ErrUnauthorizedAccess      = errors.New("unauthorized access to partner data")
	ErrCategoryNotShared       = errors.New("equipment category not shared by partner")
)
