package test_test

import (
	_ "embed"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

//go:embed testdata/test_phishlet.yaml
var validPhishletYAML string

//go:embed testdata/miraged.yaml
var miragedConfig string

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// wantError asserts that err is non-nil, contains the given substring, and
// (if status > 0) is an APIError with the expected HTTP status code. Pass 0
// for status to skip the status check (e.g., for client-side validation errors).
func wantError(t *testing.T, err error, status int, substr string) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error containing %q, got nil", substr)
		return
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("error = %q, want substring %q", err.Error(), substr)
	}
	if status > 0 {
		var apiErr *sdk.APIError
		if !errors.As(err, &apiErr) {
			t.Errorf("expected APIError with status %d, got %T: %v", status, err, err)
			return
		}
		if apiErr.StatusCode != status {
			t.Errorf("status = %d, want %d (message: %s)", apiErr.StatusCode, status, apiErr.Message)
		}
	}
}

// createTestSMTP creates an SMTP profile with sensible defaults.
// Pass functional options to override specific fields.
func createTestSMTP(t *testing.T, h *test.Harness, opts ...func(*sdk.CreateSMTPProfileRequest)) *sdk.SMTPProfileResponse {
	t.Helper()
	req := sdk.CreateSMTPProfileRequest{
		Name:     "Test SMTP",
		Host:     "smtp.example.com",
		FromAddr: "test@example.com",
	}
	for _, opt := range opts {
		opt(&req)
	}
	smtp, err := h.Client.CreateSMTPProfile(req)
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}
	return smtp
}

// createTestTemplate creates a template with sensible defaults.
func createTestTemplate(t *testing.T, h *test.Harness, opts ...func(*sdk.CreateTemplateRequest)) *sdk.TemplateResponse {
	t.Helper()
	req := sdk.CreateTemplateRequest{
		Name:     "Test Template",
		Subject:  "Action Required",
		HTMLBody: "<p>Click {{.URL}}</p>",
	}
	for _, opt := range opts {
		opt(&req)
	}
	tmpl, err := h.Client.CreateTemplate(req)
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	return tmpl
}

// createTestTargetList creates a target list and adds the given targets.
// If no targets are provided, adds two default targets (alice and bob).
func createTestTargetList(t *testing.T, h *test.Harness, targets ...sdk.AddTargetRequest) *sdk.TargetListResponse {
	t.Helper()
	list, err := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Test Targets"})
	if err != nil {
		t.Fatalf("CreateTargetList: %v", err)
	}
	if len(targets) == 0 {
		targets = []sdk.AddTargetRequest{
			{Email: "alice@acme.com", FirstName: "Alice"},
			{Email: "bob@acme.com", FirstName: "Bob"},
		}
	}
	for _, target := range targets {
		if _, err := h.Client.AddTarget(list.ID, target); err != nil {
			t.Fatalf("AddTarget: %v", err)
		}
	}
	return list
}

// createPrerequisites sets up a target list, template, and SMTP profile
// and returns their IDs for use in campaign creation.
func createPrerequisites(t *testing.T, h *test.Harness) (targetListID, templateID, smtpProfileID string) {
	t.Helper()
	list := createTestTargetList(t, h)
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)
	return list.ID, tmpl.ID, smtp.ID
}

// waitForCampaignStatus polls until the campaign reaches the expected status.
func waitForCampaignStatus(t *testing.T, h *test.Harness, id, status string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got, err := h.Client.GetCampaign(id)
		if err != nil {
			t.Fatalf("GetCampaign: %v", err)
		}
		if got.Status == status {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("campaign %s did not reach status %q within timeout", id, status)
}

// waitForEmails polls the mock mailer until at least n emails have been sent.
func waitForEmails(t *testing.T, h *test.Harness, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if h.Mailer.Count() >= n {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("expected at least %d emails, got %d", n, h.Mailer.Count())
}

func validCampaignRequest(listID, tmplID, smtpID string) sdk.CreateCampaignRequest {
	return sdk.CreateCampaignRequest{
		Name:          "Test Campaign",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
		RedirectURL:   "https://example.com",
		SendRate:      10,
	}
}
