package delivery

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	texttemplate "text/template"

	mail "github.com/wneessen/go-mail"

	"github.com/travisbale/barb/internal/phishing"
)

// TemplateData holds the variables available in email templates.
type TemplateData struct {
	FirstName string
	LastName  string
	Email     string
	URL       string
}

// Sender implements phishing.Mailer by opening persistent SMTP connections.
type Sender struct {
	Logger *slog.Logger
}

// Dial connects to the SMTP server described by profile and returns a
// connection that can send multiple messages before being closed.
func (s *Sender) Dial(ctx context.Context, profile *phishing.SMTPProfile) (phishing.MailConn, error) {
	client, err := dialClient(profile)
	if err != nil {
		return nil, fmt.Errorf("creating SMTP client: %w", err)
	}
	if err := client.DialWithContext(ctx); err != nil {
		return nil, fmt.Errorf("connecting to SMTP server: %w", err)
	}
	return &conn{client: client, logger: s.Logger}, nil
}

// conn wraps a connected go-mail client and implements phishing.MailConn.
type conn struct {
	client *mail.Client
	logger *slog.Logger
}

// Send renders and sends a single email on the open connection.
func (c *conn) Send(profile *phishing.SMTPProfile, tmpl *phishing.EmailTemplate, target *phishing.Target, lureURL string) error {
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
	if tmpl.EnvelopeSender != "" {
		if err := msg.EnvelopeFrom(tmpl.EnvelopeSender); err != nil {
			return fmt.Errorf("setting envelope sender: %w", err)
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

	// Apply custom headers from the SMTP profile.
	for key, value := range profile.CustomHeaders {
		msg.SetGenHeader(mail.Header(key), value)
	}

	return c.client.Send(msg)
}

// Close terminates the SMTP connection.
func (c *conn) Close() error {
	return c.client.Close()
}

func dialClient(profile *phishing.SMTPProfile) (*mail.Client, error) {
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
