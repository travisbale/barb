package sdk

import (
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
		{"missing name", CreateTargetListRequest{}, "name: required"},
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
		{"missing email", AddTargetRequest{FirstName: "Alice"}, "email: required"},
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
		{"missing name", CreateSMTPProfileRequest{Host: "h", FromAddr: "f"}, "name: required"},
		{"missing host", CreateSMTPProfileRequest{Name: "n", FromAddr: "f"}, "host: required"},
		{"missing from_addr", CreateSMTPProfileRequest{Name: "n", Host: "h"}, "from_addr: required"},
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
		{"empty name", UpdateSMTPProfileRequest{Name: &empty}, "name: cannot be empty"},
		{"empty host", UpdateSMTPProfileRequest{Host: &empty}, "host: cannot be empty"},
		{"empty from_addr", UpdateSMTPProfileRequest{FromAddr: &empty}, "from_addr: cannot be empty"},
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
		{"missing name", CreateTemplateRequest{Subject: "S", HTMLBody: "b"}, "name: required"},
		{"missing subject", CreateTemplateRequest{Name: "T", HTMLBody: "b"}, "subject: required"},
		{"missing body", CreateTemplateRequest{Name: "T", Subject: "S"}, "body: HTML or text body is required"},
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
		{"empty name", UpdateTemplateRequest{Name: &empty}, "name: cannot be empty"},
		{"empty subject", UpdateTemplateRequest{Subject: &empty}, "subject: cannot be empty"},
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
		{"missing yaml", CreatePhishletRequest{}, "yaml: required"},
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
		{"missing name", func(r *CreateCampaignRequest) { r.Name = "" }, "name: required"},
		{"missing template_id", func(r *CreateCampaignRequest) { r.TemplateID = "" }, "template_id: required"},
		{"missing smtp_profile_id", func(r *CreateCampaignRequest) { r.SMTPProfileID = "" }, "smtp_profile_id: required"},
		{"missing target_list_id", func(r *CreateCampaignRequest) { r.TargetListID = "" }, "target_list_id: required"},
		{"missing redirect_url", func(r *CreateCampaignRequest) { r.RedirectURL = "" }, "redirect_url: required"},
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
		{"missing name", func(r *EnrollMiragedRequest) { r.Name = "" }, "name: required"},
		{"missing address", func(r *EnrollMiragedRequest) { r.Address = "" }, "address: required"},
		{"bad address format", func(r *EnrollMiragedRequest) { r.Address = "no-port" }, "host:port format"},
		{"missing secret_hostname", func(r *EnrollMiragedRequest) { r.SecretHostname = "" }, "secret_hostname: required"},
		{"missing token", func(r *EnrollMiragedRequest) { r.Token = "" }, "token: required"},
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
