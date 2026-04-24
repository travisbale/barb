package sdk

import (
	"errors"
	"testing"
)

func TestCreateTargetListRequest_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		req     CreateTargetListRequest
		wantErr string
	}{
		{"valid", CreateTargetListRequest{Name: "Targets"}, ""},
		{"missing name", CreateTargetListRequest{}, "Name is required."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestAddTargetRequest_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		req     AddTargetRequest
		wantErr string
	}{
		{"valid", AddTargetRequest{Email: "alice@example.com"}, ""},
		{"missing email", AddTargetRequest{FirstName: "Alice"}, "Email is required."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestCreateSMTPProfileRequest_Validate(t *testing.T) {
	t.Parallel()
	valid := CreateSMTPProfileRequest{Name: "SMTP", Host: "smtp.example.com", FromAddr: "from@example.com"}
	tests := []struct {
		name    string
		req     CreateSMTPProfileRequest
		wantErr string
	}{
		{"valid", valid, ""},
		{"missing name", CreateSMTPProfileRequest{Host: "h", FromAddr: "f"}, "Name is required."},
		{"missing host", CreateSMTPProfileRequest{Name: "n", FromAddr: "f"}, "Host is required."},
		{"missing from_addr", CreateSMTPProfileRequest{Name: "n", Host: "h"}, "From address is required."},
		{"reserved header", CreateSMTPProfileRequest{Name: "n", Host: "h", FromAddr: "f", CustomHeaders: map[string]string{"From": "x"}}, "conflicts with a standard email header"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestUpdateSMTPProfileRequest_Validate(t *testing.T) {
	t.Parallel()
	empty := ""
	valid := "valid"
	tests := []struct {
		name    string
		req     UpdateSMTPProfileRequest
		wantErr string
	}{
		{"valid", UpdateSMTPProfileRequest{Name: &valid}, ""},
		{"empty name", UpdateSMTPProfileRequest{Name: &empty}, "Name cannot be empty."},
		{"empty host", UpdateSMTPProfileRequest{Host: &empty}, "Host cannot be empty."},
		{"empty from_addr", UpdateSMTPProfileRequest{FromAddr: &empty}, "From address cannot be empty."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestCreateTemplateRequest_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		req     CreateTemplateRequest
		wantErr string
	}{
		{"valid html", CreateTemplateRequest{Name: "T", Subject: "S", HTMLBody: "<p>hi</p>"}, ""},
		{"valid text", CreateTemplateRequest{Name: "T", Subject: "S", TextBody: "hi"}, ""},
		{"missing name", CreateTemplateRequest{Subject: "S", HTMLBody: "b"}, "Name is required."},
		{"missing subject", CreateTemplateRequest{Name: "T", HTMLBody: "b"}, "Subject is required."},
		{"missing body", CreateTemplateRequest{Name: "T", Subject: "S"}, "HTML or text body is required."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestUpdateTemplateRequest_Validate(t *testing.T) {
	t.Parallel()
	empty := ""
	valid := "valid"
	tests := []struct {
		name    string
		req     UpdateTemplateRequest
		wantErr string
	}{
		{"valid", UpdateTemplateRequest{Name: &valid}, ""},
		{"empty name", UpdateTemplateRequest{Name: &empty}, "Name cannot be empty."},
		{"empty subject", UpdateTemplateRequest{Subject: &empty}, "Subject cannot be empty."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestCreatePhishletRequest_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		req     CreatePhishletRequest
		wantErr string
	}{
		{"valid", CreatePhishletRequest{YAML: "name: test"}, ""},
		{"missing yaml", CreatePhishletRequest{}, "YAML is required."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestCreateCampaignRequest_Validate(t *testing.T) {
	t.Parallel()
	valid := CreateCampaignRequest{
		Name: "C", TemplateID: "t", SMTPProfileID: "s", TargetListID: "l", RedirectURL: "https://example.com",
	}
	tests := []struct {
		name    string
		modify  func(*CreateCampaignRequest)
		wantErr string
	}{
		{"valid", nil, ""},
		{"missing name", func(r *CreateCampaignRequest) { r.Name = "" }, "Name is required."},
		{"missing template_id", func(r *CreateCampaignRequest) { r.TemplateID = "" }, "Template is required."},
		{"missing smtp_profile_id", func(r *CreateCampaignRequest) { r.SMTPProfileID = "" }, "SMTP profile is required."},
		{"missing target_list_id", func(r *CreateCampaignRequest) { r.TargetListID = "" }, "Target list is required."},
		{"missing redirect_url", func(r *CreateCampaignRequest) { r.RedirectURL = "" }, "Redirect URL is required."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			if tt.modify != nil {
				tt.modify(&req)
			}
			err := req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestEnrollMiragedRequest_Validate(t *testing.T) {
	t.Parallel()
	valid := EnrollMiragedRequest{
		Name: "M", Address: "127.0.0.1:443", SecretHostname: "mgmt.local", Token: "tok",
	}
	tests := []struct {
		name    string
		modify  func(*EnrollMiragedRequest)
		wantErr string
	}{
		{"valid", nil, ""},
		{"missing name", func(r *EnrollMiragedRequest) { r.Name = "" }, "Name is required."},
		{"missing address", func(r *EnrollMiragedRequest) { r.Address = "" }, "Address is required."},
		{"bad address format", func(r *EnrollMiragedRequest) { r.Address = "no-port" }, "host:port format"},
		{"missing secret_hostname", func(r *EnrollMiragedRequest) { r.SecretHostname = "" }, "Secret hostname is required."},
		{"missing token", func(r *EnrollMiragedRequest) { r.Token = "" }, "Token is required."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			if tt.modify != nil {
				tt.modify(&req)
			}
			err := req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestUpdateTargetListRequest_Validate(t *testing.T) {
	t.Parallel()
	empty := ""
	valid := "valid"
	tests := []struct {
		name    string
		req     UpdateTargetListRequest
		wantErr string
	}{
		{"valid", UpdateTargetListRequest{Name: &valid}, ""},
		{"nil name", UpdateTargetListRequest{}, ""},
		{"empty name", UpdateTargetListRequest{Name: &empty}, "Name cannot be empty."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestCreateMiragedNotificationChannelRequest_Validate(t *testing.T) {
	t.Parallel()
	valid := CreateMiragedNotificationChannelRequest{Type: "slack", URL: "https://hooks.slack.com/x"}
	tests := []struct {
		name    string
		modify  func(*CreateMiragedNotificationChannelRequest)
		wantErr string
	}{
		{"valid slack", nil, ""},
		{"valid webhook", func(r *CreateMiragedNotificationChannelRequest) { r.Type = "webhook" }, ""},
		{"missing type", func(r *CreateMiragedNotificationChannelRequest) { r.Type = "" }, "webhook"},
		{"invalid type", func(r *CreateMiragedNotificationChannelRequest) { r.Type = "email" }, "webhook"},
		{"missing url", func(r *CreateMiragedNotificationChannelRequest) { r.URL = "" }, "URL is required."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			if tt.modify != nil {
				tt.modify(&req)
			}
			err := req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

func TestUpdateMiragedRequest_Validate(t *testing.T) {
	t.Parallel()
	empty := ""
	valid := "valid"
	tests := []struct {
		name    string
		req     UpdateMiragedRequest
		wantErr string
	}{
		{"valid", UpdateMiragedRequest{Name: &valid}, ""},
		{"nil name", UpdateMiragedRequest{}, ""},
		{"empty name", UpdateMiragedRequest{Name: &empty}, "Name cannot be empty."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			checkValidation(t, err, tt.wantErr)
		})
	}
}

// checkValidation asserts that an error matches expectations.
// If wantErr is empty, the error should be nil.
func checkValidation(t *testing.T, err error, wantErr string) {
	t.Helper()
	if wantErr == "" {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Errorf("expected error containing %q, got nil", wantErr)
		return
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected *ValidationError, got %T", err)
		return
	}
	if got := err.Error(); got != wantErr && !contains(got, wantErr) {
		t.Errorf("error = %q, want substring %q", got, wantErr)
	}
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
