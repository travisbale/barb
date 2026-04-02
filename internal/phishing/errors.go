package phishing

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrNameRequired        = errors.New("name is required")
	ErrEmailRequired       = errors.New("email is required")
	ErrTemplateNotFound    = errors.New("template not found")
	ErrSMTPProfileNotFound = errors.New("SMTP profile not found")
	ErrTargetListNotFound  = errors.New("target list not found")
	ErrCampaignNotDraft    = errors.New("campaign can only be started from draft status")
	ErrCampaignNotRunning  = errors.New("campaign is not running")
	ErrPhishletNotFound    = errors.New("phishlet not found")
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrPasswordRequired    = errors.New("password is required")
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters")
	ErrConnectionFailed    = errors.New("could not reach server")
	ErrEnrollmentRejected  = errors.New("enrollment rejected by server")
)
