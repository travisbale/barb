package phishing

import "errors"

var (
	ErrNotFound               = errors.New("not found")
	ErrConflict               = errors.New("conflict")
	ErrNameRequired           = errors.New("name is required")
	ErrEmailRequired          = errors.New("email is required")
	ErrHostRequired           = errors.New("host is required")
	ErrFromAddrRequired       = errors.New("from address is required")
	ErrSubjectRequired        = errors.New("subject is required")
	ErrBodyRequired           = errors.New("HTML or text body is required")
	ErrTemplateRequired       = errors.New("template is required")
	ErrSMTPProfileRequired    = errors.New("SMTP profile is required")
	ErrTargetListRequired     = errors.New("target list is required")
	ErrTemplateNotFound       = errors.New("template not found")
	ErrSMTPProfileNotFound    = errors.New("SMTP profile not found")
	ErrTargetListNotFound     = errors.New("target list not found")
	ErrCampaignNotDraft       = errors.New("campaign can only be started from draft status")
	ErrCampaignNotRunning     = errors.New("campaign is not running")
	ErrAddressRequired        = errors.New("address is required")
	ErrSecretHostnameRequired = errors.New("secret hostname is required")
	ErrCertsRequired          = errors.New("certificate, key, and CA cert are required")
	ErrYAMLRequired           = errors.New("YAML content is required")
	ErrPhishletNotFound       = errors.New("phishlet not found")
)
