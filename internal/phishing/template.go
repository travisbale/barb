package phishing

import (
	"time"

	"github.com/google/uuid"
)

// EmailTemplate is a reusable phishing email template with subject and body.
type EmailTemplate struct {
	ID        string
	Name      string
	Subject   string
	HTMLBody  string
	TextBody  string
	CreatedAt time.Time
}

func (t *EmailTemplate) Validate() error {
	if t.Name == "" {
		return ErrNameRequired
	}
	if t.Subject == "" {
		return ErrSubjectRequired
	}
	if t.HTMLBody == "" && t.TextBody == "" {
		return ErrBodyRequired
	}
	return nil
}

type templateStore interface {
	CreateTemplate(t *EmailTemplate) error
	GetTemplate(id string) (*EmailTemplate, error)
	UpdateTemplate(t *EmailTemplate) error
	DeleteTemplate(id string) error
	ListTemplates() ([]*EmailTemplate, error)
}

// TemplateService manages email templates.
type TemplateService struct {
	Store templateStore
}

func (s *TemplateService) Create(template *EmailTemplate) (*EmailTemplate, error) {
	if err := template.Validate(); err != nil {
		return nil, err
	}

	template.ID = uuid.New().String()
	template.CreatedAt = time.Now()

	if err := s.Store.CreateTemplate(template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *TemplateService) Get(id string) (*EmailTemplate, error) {
	return s.Store.GetTemplate(id)
}

func (s *TemplateService) Update(id string, template *EmailTemplate) (*EmailTemplate, error) {
	existing, err := s.Store.GetTemplate(id)
	if err != nil {
		return nil, err
	}

	existing.Name = template.Name
	existing.Subject = template.Subject
	existing.HTMLBody = template.HTMLBody
	existing.TextBody = template.TextBody

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := s.Store.UpdateTemplate(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *TemplateService) Delete(id string) error {
	return s.Store.DeleteTemplate(id)
}

func (s *TemplateService) List() ([]*EmailTemplate, error) {
	return s.Store.ListTemplates()
}
