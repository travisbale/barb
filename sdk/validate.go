package sdk

import (
	"fmt"
	"net"
	"strings"
)

// --- Target Lists ---

func (r CreateTargetListRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name: required")
	}
	return nil
}

func (r AddTargetRequest) Validate() error {
	if r.Email == "" {
		return fmt.Errorf("email: required")
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
		return fmt.Errorf("name: required")
	}
	if r.Host == "" {
		return fmt.Errorf("host: required")
	}
	if r.FromAddr == "" {
		return fmt.Errorf("from_addr: required")
	}
	return validateHeaders(r.CustomHeaders)
}

func (r UpdateSMTPProfileRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return fmt.Errorf("name: cannot be empty")
	}
	if r.Host != nil && *r.Host == "" {
		return fmt.Errorf("host: cannot be empty")
	}
	if r.FromAddr != nil && *r.FromAddr == "" {
		return fmt.Errorf("from_addr: cannot be empty")
	}
	if r.CustomHeaders != nil {
		return validateHeaders(*r.CustomHeaders)
	}
	return nil
}

// --- Email Templates ---

func (r CreateTemplateRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name: required")
	}
	if r.Subject == "" {
		return fmt.Errorf("subject: required")
	}
	if r.HTMLBody == "" && r.TextBody == "" {
		return fmt.Errorf("body: HTML or text body is required")
	}
	return nil
}

func (r UpdateTemplateRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return fmt.Errorf("name: cannot be empty")
	}
	if r.Subject != nil && *r.Subject == "" {
		return fmt.Errorf("subject: cannot be empty")
	}
	return nil
}

// --- Phishlets ---

func (r CreatePhishletRequest) Validate() error {
	if r.YAML == "" {
		return fmt.Errorf("yaml: required")
	}
	return nil
}

func (r UpdatePhishletRequest) Validate() error {
	if r.YAML == "" {
		return fmt.Errorf("yaml: required")
	}
	return nil
}

// --- Campaigns ---

func (r CreateCampaignRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name: required")
	}
	if r.TemplateID == "" {
		return fmt.Errorf("template_id: required")
	}
	if r.SMTPProfileID == "" {
		return fmt.Errorf("smtp_profile_id: required")
	}
	if r.TargetListID == "" {
		return fmt.Errorf("target_list_id: required")
	}
	if r.RedirectURL == "" {
		return fmt.Errorf("redirect_url: required")
	}
	return nil
}

func (r UpdateCampaignRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return fmt.Errorf("name: cannot be empty")
	}
	return nil
}

// --- Miraged Connections ---

func (r EnrollMiragedRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name: required")
	}
	if r.Address == "" {
		return fmt.Errorf("address: required")
	}
	if _, _, err := net.SplitHostPort(r.Address); err != nil {
		return fmt.Errorf("address: must be in host:port format")
	}
	if r.SecretHostname == "" {
		return fmt.Errorf("secret_hostname: required")
	}
	if r.Token == "" {
		return fmt.Errorf("token: required")
	}
	return nil
}

func (r PushMiragedPhishletRequest) Validate() error {
	if r.YAML == "" {
		return fmt.Errorf("yaml: required")
	}
	return nil
}

func validateHeaders(headers map[string]string) error {
	for key := range headers {
		if reservedHeaders[strings.ToLower(key)] {
			return fmt.Errorf("custom_headers: %q conflicts with a standard email header", key)
		}
	}
	return nil
}
