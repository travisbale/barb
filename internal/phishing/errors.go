package phishing

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("conflict")
	ErrNameRequired  = errors.New("name is required")
	ErrEmailRequired = errors.New("email is required")
)
