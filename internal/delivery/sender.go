package delivery

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	texttemplate "text/template"

	mail "github.com/wneessen/go-mail"

	"github.com/travisbale/mirador/internal/phishing"
)

// TemplateData holds the variables available in email templates.
type TemplateData struct {
	FirstName string
	LastName  string
	Email     string
	URL       string
}

// Sender implements phishing.Mailer using go-mail for MIME construction
// and SMTP delivery.
type Sender struct {
	Logger *slog.Logger
}

// Send renders an email template for a target and sends it via SMTP.
func (s *Sender) Send(profile *phishing.SMTPProfile, tmpl *phishing.EmailTemplate, target *phishing.Target, lureURL string) error {
	data := TemplateData{
		FirstName: target.FirstName,
		LastName:  target.LastName,
		Email:     target.Email,
		URL:       lureURL,
	}

	subject, err := renderSubject(tmpl.Subject, data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if profile.FromName != "" {
		if err := msg.FromFormat(profile.FromName, profile.FromAddr); err != nil {
			return fmt.Errorf("setting from: %w", err)
		}
	} else {
		if err := msg.From(profile.FromAddr); err != nil {
			return fmt.Errorf("setting from: %w", err)
		}
	}
	if err := msg.To(target.Email); err != nil {
		return fmt.Errorf("setting to: %w", err)
	}
	msg.Subject(subject)

	if tmpl.HTMLBody != "" {
		htmlBody, err := renderHTML(tmpl.HTMLBody, data)
		if err != nil {
			return err
		}
		msg.SetBodyString(mail.TypeTextHTML, htmlBody)

		// Add plain text alternative if available.
		if tmpl.TextBody != "" {
			textBody, err := renderText(tmpl.TextBody, data)
			if err != nil {
				return err
			}
			msg.AddAlternativeString(mail.TypeTextPlain, textBody)
		}
	} else if tmpl.TextBody != "" {
		textBody, err := renderText(tmpl.TextBody, data)
		if err != nil {
			return err
		}
		msg.SetBodyString(mail.TypeTextPlain, textBody)
	}

	client, err := s.dialClient(profile)
	if err != nil {
		return fmt.Errorf("connecting to SMTP server: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			s.Logger.Debug("SMTP close error", "error", err)
		}
	}()

	return client.DialAndSend(msg)
}

func (s *Sender) dialClient(profile *phishing.SMTPProfile) (*mail.Client, error) {
	opts := []mail.Option{
		mail.WithPort(profile.Port),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
	}

	if profile.Username != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(profile.Username),
			mail.WithPassword(profile.Password),
		)
	}

	return mail.NewClient(profile.Host, opts...)
}

func renderHTML(body string, data TemplateData) (string, error) {
	tmpl, err := template.New("html").Parse(body)
	if err != nil {
		return "", fmt.Errorf("parsing HTML template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering HTML template: %w", err)
	}
	return buf.String(), nil
}

func renderText(body string, data TemplateData) (string, error) {
	tmpl, err := texttemplate.New("text").Parse(body)
	if err != nil {
		return "", fmt.Errorf("parsing text template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering text template: %w", err)
	}
	return buf.String(), nil
}

func renderSubject(subject string, data TemplateData) (string, error) {
	return renderText(subject, data)
}
