package sdk

import (
	"fmt"
	"net"
	"strings"
)

// ValidationError represents a user-facing validation message.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func invalid(msg string) error {
	return &ValidationError{Message: msg}
}

// --- Target Lists ---

func (r CreateTargetListRequest) Validate() error {
	if r.Name == "" {
		return invalid("Name is required.")
	}
	return nil
}

func (r UpdateTargetListRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return invalid("Name cannot be empty.")
	}
	return nil
}

func (r AddTargetRequest) Validate() error {
	if r.Email == "" {
		return invalid("Email is required.")
	}
	return nil
}

// --- SMTP Profiles ---

var reservedHeaders = map[string]bool{
	"from": true, "to": true, "cc": true, "bcc": true,
	"subject": true, "date": true, "message-id": true,
	"content-type": true, "content-transfer-encoding": true,
	"mime-version": true, "reply-to": true, "sender": true,
	"return-path": true,
}

func (r CreateSMTPProfileRequest) Validate() error {
	if r.Name == "" {
		return invalid("Name is required.")
	}
	if r.Host == "" {
		return invalid("Host is required.")
	}
	if r.FromAddr == "" {
		return invalid("From address is required.")
	}
	return validateHeaders(r.CustomHeaders)
}

func (r UpdateSMTPProfileRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return invalid("Name cannot be empty.")
	}
	if r.Host != nil && *r.Host == "" {
		return invalid("Host cannot be empty.")
	}
	if r.FromAddr != nil && *r.FromAddr == "" {
		return invalid("From address cannot be empty.")
	}
	if r.CustomHeaders != nil {
		return validateHeaders(*r.CustomHeaders)
	}
	return nil
}

// --- Email Templates ---

func (r CreateTemplateRequest) Validate() error {
	if r.Name == "" {
		return invalid("Name is required.")
	}
	if r.Subject == "" {
		return invalid("Subject is required.")
	}
	if r.HTMLBody == "" && r.TextBody == "" {
		return invalid("HTML or text body is required.")
	}
	return nil
}

func (r UpdateTemplateRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return invalid("Name cannot be empty.")
	}
	if r.Subject != nil && *r.Subject == "" {
		return invalid("Subject cannot be empty.")
	}
	return nil
}

func (r RenderHTMLRequest) Validate() error {
	if r.HTMLBody == "" {
		return invalid("HTML body is required.")
	}
	return nil
}

// --- Phishlets ---

func (r CreatePhishletRequest) Validate() error {
	if r.YAML == "" {
		return invalid("YAML is required.")
	}
	return nil
}

func (r UpdatePhishletRequest) Validate() error {
	if r.YAML == "" {
		return invalid("YAML is required.")
	}
	return nil
}

// --- Campaigns ---

func (r CreateCampaignRequest) Validate() error {
	if r.Name == "" {
		return invalid("Name is required.")
	}
	if r.TemplateID == "" {
		return invalid("Template is required.")
	}
	if r.SMTPProfileID == "" {
		return invalid("SMTP profile is required.")
	}
	if r.TargetListID == "" {
		return invalid("Target list is required.")
	}
	if r.RedirectURL == "" {
		return invalid("Redirect URL is required.")
	}
	return nil
}

func (r UpdateCampaignRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return invalid("Name cannot be empty.")
	}
	return nil
}

// --- Miraged Connections ---

func (r EnrollMiragedRequest) Validate() error {
	if r.Name == "" {
		return invalid("Name is required.")
	}
	if r.Address == "" {
		return invalid("Address is required.")
	}
	if _, _, err := net.SplitHostPort(r.Address); err != nil {
		return invalid("Address must be in host:port format.")
	}
	if r.SecretHostname == "" {
		return invalid("Secret hostname is required.")
	}
	if r.Token == "" {
		return invalid("Token is required.")
	}
	return nil
}

func (r UpdateMiragedRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return invalid("Name cannot be empty.")
	}
	return nil
}

func (r PushMiragedPhishletRequest) Validate() error {
	if r.YAML == "" {
		return invalid("YAML is required.")
	}
	return nil
}

func (r CreateMiragedNotificationChannelRequest) Validate() error {
	if r.Type != "webhook" && r.Type != "slack" {
		return invalid("Type must be \"webhook\" or \"slack\".")
	}
	if r.URL == "" {
		return invalid("URL is required.")
	}
	return nil
}

// --- Preview / Enable (no required fields) ---

func (r PreviewTemplateRequest) Validate() error       { return nil }
func (r EnableMiragedPhishletRequest) Validate() error { return nil }

func validateHeaders(headers map[string]string) error {
	for key := range headers {
		if reservedHeaders[strings.ToLower(key)] {
			return invalid(fmt.Sprintf("%q conflicts with a standard email header.", key))
		}
	}
	return nil
}
