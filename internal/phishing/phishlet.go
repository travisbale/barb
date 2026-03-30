package phishing

import (
	"time"

	"github.com/google/uuid"
)

// Phishlet is a stored phishlet YAML config managed by Mirador.
type Phishlet struct {
	ID        string
	Name      string
	YAML      string
	CreatedAt time.Time
}

func (p *Phishlet) Validate() error {
	if p.Name == "" {
		return ErrNameRequired
	}
	if p.YAML == "" {
		return ErrYAMLRequired
	}
	return nil
}

type phishletStore interface {
	CreatePhishlet(p *Phishlet) error
	GetPhishlet(id string) (*Phishlet, error)
	GetPhishletByName(name string) (*Phishlet, error)
	UpdatePhishlet(p *Phishlet) error
	DeletePhishlet(id string) error
	ListPhishlets() ([]*Phishlet, error)
}

// PhishletUpdate holds optional fields for a partial phishlet update.
type PhishletUpdate struct {
	Name *string
	YAML *string
}

// PhishletService manages phishlet YAML configs stored in Mirador.
type PhishletService struct {
	Store phishletStore
}

func (s *PhishletService) Create(phishlet *Phishlet) (*Phishlet, error) {
	if err := phishlet.Validate(); err != nil {
		return nil, err
	}

	phishlet.ID = uuid.New().String()
	phishlet.CreatedAt = time.Now()

	if err := s.Store.CreatePhishlet(phishlet); err != nil {
		return nil, err
	}
	return phishlet, nil
}

func (s *PhishletService) Get(id string) (*Phishlet, error) {
	return s.Store.GetPhishlet(id)
}

func (s *PhishletService) GetByName(name string) (*Phishlet, error) {
	return s.Store.GetPhishletByName(name)
}

func (s *PhishletService) Update(id string, update *PhishletUpdate) (*Phishlet, error) {
	existing, err := s.Store.GetPhishlet(id)
	if err != nil {
		return nil, err
	}

	if update.Name != nil {
		existing.Name = *update.Name
	}
	if update.YAML != nil {
		existing.YAML = *update.YAML
	}

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := s.Store.UpdatePhishlet(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *PhishletService) Delete(id string) error {
	return s.Store.DeletePhishlet(id)
}

func (s *PhishletService) List() ([]*Phishlet, error) {
	return s.Store.ListPhishlets()
}
