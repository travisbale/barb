package phishing

import (
	"time"

	"github.com/google/uuid"
)

// SMTPProfile is a configured mail relay for sending phishing emails.
type SMTPProfile struct {
	ID        string
	Name      string
	Host      string
	Port      int
	Username  string
	Password  string
	FromAddr  string
	FromName  string
	CreatedAt time.Time
}

type smtpStore interface {
	CreateProfile(p *SMTPProfile) error
	GetProfile(id string) (*SMTPProfile, error)
	DeleteProfile(id string) error
	ListProfiles() ([]*SMTPProfile, error)
}

func (p *SMTPProfile) Validate() error {
	if p.Name == "" {
		return ErrNameRequired
	}
	if p.Host == "" {
		return ErrHostRequired
	}
	if p.FromAddr == "" {
		return ErrFromAddrRequired
	}
	if p.Port == 0 {
		p.Port = 587
	}
	return nil
}

// SMTPService manages SMTP relay profiles.
type SMTPService struct {
	Store smtpStore
}

func (s *SMTPService) CreateProfile(profile *SMTPProfile) (*SMTPProfile, error) {
	if err := profile.Validate(); err != nil {
		return nil, err
	}

	profile.ID = uuid.New().String()
	profile.CreatedAt = time.Now()

	if err := s.Store.CreateProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *SMTPService) GetProfile(id string) (*SMTPProfile, error) {
	return s.Store.GetProfile(id)
}

func (s *SMTPService) DeleteProfile(id string) error {
	return s.Store.DeleteProfile(id)
}

func (s *SMTPService) ListProfiles() ([]*SMTPProfile, error) {
	return s.Store.ListProfiles()
}
