package phishing

import (
	"bytes"
	"fmt"
	"html/template"
	texttemplate "text/template"
	"time"

	"github.com/google/uuid"
)

// EmailTemplate is a reusable phishing email template with subject and body.
type EmailTemplate struct {
	ID             string
	Name           string
	Subject        string
	HTMLBody       string
	TextBody       string
	EnvelopeSender string
	CreatedAt      time.Time
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

// TemplateUpdate holds optional fields for a partial template update.
// Nil fields are left unchanged.
type TemplateUpdate struct {
	Name           *string
	Subject        *string
	HTMLBody       *string
	TextBody       *string
	EnvelopeSender *string
}

func (s *TemplateService) Update(id string, update *TemplateUpdate) (*EmailTemplate, error) {
	existing, err := s.Store.GetTemplate(id)
	if err != nil {
		return nil, err
	}

	if update.Name != nil {
		existing.Name = *update.Name
	}
	if update.Subject != nil {
		existing.Subject = *update.Subject
	}
	if update.HTMLBody != nil {
		existing.HTMLBody = *update.HTMLBody
	}
	if update.TextBody != nil {
		existing.TextBody = *update.TextBody
	}
	if update.EnvelopeSender != nil {
		existing.EnvelopeSender = *update.EnvelopeSender
	}

	if err := s.Store.UpdateTemplate(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// PreviewData holds the variables available when rendering a template preview.
type PreviewData struct {
	FirstName string
	LastName  string
	Email     string
	URL       string
}

// RenderedTemplate is the result of rendering a template with preview data.
type RenderedTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

func (s *TemplateService) Preview(id string, data PreviewData) (*RenderedTemplate, error) {
	tmpl, err := s.Store.GetTemplate(id)
	if err != nil {
		return nil, err
	}

	rendered := &RenderedTemplate{}

	if rendered.Subject, err = renderText(tmpl.Subject, data); err != nil {
		return nil, fmt.Errorf("rendering subject: %w", err)
	}
	if tmpl.HTMLBody != "" {
		if rendered.HTMLBody, err = renderHTML(tmpl.HTMLBody, data); err != nil {
			return nil, fmt.Errorf("rendering HTML body: %w", err)
		}
	}
	if tmpl.TextBody != "" {
		if rendered.TextBody, err = renderText(tmpl.TextBody, data); err != nil {
			return nil, fmt.Errorf("rendering text body: %w", err)
		}
	}

	return rendered, nil
}

func renderHTML(body string, data PreviewData) (string, error) {
	tmpl, err := template.New("html").Parse(body)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderText(body string, data PreviewData) (string, error) {
	tmpl, err := texttemplate.New("text").Parse(body)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *TemplateService) Delete(id string) error {
	return s.Store.DeleteTemplate(id)
}

func (s *TemplateService) List() ([]*EmailTemplate, error) {
	return s.Store.ListTemplates()
}
