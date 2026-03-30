package phishing

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Phishlet is a stored phishlet YAML config managed by Mirador.
// The Name is extracted automatically from the YAML content.
type Phishlet struct {
	ID        string
	Name      string
	YAML      string
	CreatedAt time.Time
}

// extractName parses just the top-level name field from the YAML.
func extractName(yamlContent string) (string, error) {
	var header struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal([]byte(yamlContent), &header); err != nil {
		return "", fmt.Errorf("invalid YAML: %w", err)
	}
	if header.Name == "" {
		return "", ErrNameRequired
	}
	return header.Name, nil
}

type phishletStore interface {
	CreatePhishlet(p *Phishlet) error
	GetPhishlet(id string) (*Phishlet, error)
	GetPhishletByName(name string) (*Phishlet, error)
	UpdatePhishlet(p *Phishlet) error
	DeletePhishlet(id string) error
	ListPhishlets() ([]*Phishlet, error)
}

// PhishletService manages phishlet YAML configs stored in Mirador.
type PhishletService struct {
	Store phishletStore
}

func (s *PhishletService) Create(yamlContent string) (*Phishlet, error) {
	if yamlContent == "" {
		return nil, ErrYAMLRequired
	}

	name, err := extractName(yamlContent)
	if err != nil {
		return nil, err
	}

	phishlet := &Phishlet{
		ID:        uuid.New().String(),
		Name:      name,
		YAML:      yamlContent,
		CreatedAt: time.Now(),
	}

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

func (s *PhishletService) Update(id string, yamlContent string) (*Phishlet, error) {
	existing, err := s.Store.GetPhishlet(id)
	if err != nil {
		return nil, err
	}

	if yamlContent == "" {
		return nil, ErrYAMLRequired
	}

	name, err := extractName(yamlContent)
	if err != nil {
		return nil, err
	}

	existing.Name = name
	existing.YAML = yamlContent

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
