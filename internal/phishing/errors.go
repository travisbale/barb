package phishing

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrNameRequired     = errors.New("name is required")
	ErrEmailRequired    = errors.New("email is required")
	ErrHostRequired     = errors.New("host is required")
	ErrFromAddrRequired = errors.New("from address is required")
	ErrSubjectRequired  = errors.New("subject is required")
	ErrBodyRequired     = errors.New("HTML or text body is required")
)
