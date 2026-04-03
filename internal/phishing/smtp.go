package phishing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SMTPProfile is a configured mail relay for sending phishing emails.
type SMTPProfile struct {
	ID            string
	Name          string
	Host          string
	Port          int
	Username      string
	Password      string
	FromAddr      string
	FromName      string
	CustomHeaders map[string]string
	CreatedAt     time.Time
}

type smtpStore interface {
	CreateProfile(p *SMTPProfile) error
	GetProfile(id string) (*SMTPProfile, error)
	UpdateProfile(p *SMTPProfile) error
	DeleteProfile(id string) error
	ListProfiles() ([]*SMTPProfile, error)
}

// SMTPService manages SMTP relay profiles.
type SMTPService struct {
	Store  smtpStore
	Mailer Mailer
}

// verifyConnection dials the SMTP server to confirm it's reachable.
func (s *SMTPService) verifyConnection(ctx context.Context, profile *SMTPProfile) error {
	if s.Mailer == nil {
		return nil
	}
	conn, err := s.Mailer.Dial(ctx, profile)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSMTPConnectionFailed, err)
	}
	_ = conn.Close()
	return nil
}

func (s *SMTPService) CreateProfile(ctx context.Context, profile *SMTPProfile) (*SMTPProfile, error) {
	if profile.Port == 0 {
		profile.Port = 587
	}
	profile.ID = uuid.New().String()
	profile.CreatedAt = time.Now()

	if err := s.verifyConnection(ctx, profile); err != nil {
		return nil, err
	}
	if err := s.Store.CreateProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// SMTPProfileUpdate holds optional fields for a partial SMTP profile update.
// Nil fields are left unchanged.
type SMTPProfileUpdate struct {
	Name          *string
	Host          *string
	Port          *int
	Username      *string
	Password      *string
	FromAddr      *string
	FromName      *string
	CustomHeaders *map[string]string
}

func (s *SMTPService) UpdateProfile(ctx context.Context, id string, update *SMTPProfileUpdate) (*SMTPProfile, error) {
	existing, err := s.Store.GetProfile(id)
	if err != nil {
		return nil, err
	}

	if update.Name != nil {
		existing.Name = *update.Name
	}
	if update.Host != nil {
		existing.Host = *update.Host
	}
	if update.Port != nil {
		existing.Port = *update.Port
	}
	if update.Username != nil {
		existing.Username = *update.Username
	}
	if update.Password != nil {
		existing.Password = *update.Password
	}
	if update.FromAddr != nil {
		existing.FromAddr = *update.FromAddr
	}
	if update.FromName != nil {
		existing.FromName = *update.FromName
	}
	if update.CustomHeaders != nil {
		existing.CustomHeaders = *update.CustomHeaders
	}

	if existing.Port == 0 {
		existing.Port = 587
	}

	// Verify connectivity when connection details change.
	if update.Host != nil || update.Port != nil || update.Username != nil || update.Password != nil {
		if err := s.verifyConnection(ctx, existing); err != nil {
			return nil, err
		}
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
