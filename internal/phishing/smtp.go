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
	UpdateProfile(p *SMTPProfile) error
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

	if profile.Port == 0 {
		profile.Port = 587
	}
	profile.ID = uuid.New().String()
	profile.CreatedAt = time.Now()

	if err := s.Store.CreateProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *SMTPService) UpdateProfile(id string, profile *SMTPProfile) (*SMTPProfile, error) {
	existing, err := s.Store.GetProfile(id)
	if err != nil {
		return nil, err
	}

	existing.Name = profile.Name
	existing.Host = profile.Host
	existing.Port = profile.Port
	existing.Username = profile.Username
	existing.FromAddr = profile.FromAddr
	existing.FromName = profile.FromName

	// Only update password if a new one was provided.
	if profile.Password != "" {
		existing.Password = profile.Password
	}

	if existing.Port == 0 {
		existing.Port = 587
	}

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := s.Store.UpdateProfile(existing); err != nil {
		return nil, err
	}
	return existing, nil
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
