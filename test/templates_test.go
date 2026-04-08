package test_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestTemplates_CRUD(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Password Reset",
		Subject:  "Your password is expiring",
		HTMLBody: "<p>Click <a href='{{.URL}}'>here</a> to reset.</p>",
		TextBody: "Click here to reset: {{.URL}}",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Name != "Password Reset" {
		t.Errorf("Name = %q, want %q", created.Name, "Password Reset")
	}

	// Get.
	got, err := h.Client.GetTemplate(created.ID)
	if err != nil {
		t.Fatalf("GetTemplate: %v", err)
	}
	if got.Subject != "Your password is expiring" {
		t.Errorf("Subject = %q", got.Subject)
	}

	// Update.
	updated, err := h.Client.UpdateTemplate(created.ID, sdk.UpdateTemplateRequest{
		Name:    strPtr("Password Expiry"),
		Subject: strPtr("Action required: password expiring"),
	})
	if err != nil {
		t.Fatalf("UpdateTemplate: %v", err)
	}
	if updated.Name != "Password Expiry" {
		t.Errorf("Name after update = %q", updated.Name)
	}
	if updated.Subject != "Action required: password expiring" {
		t.Errorf("Subject after update = %q", updated.Subject)
	}

	// List.
	templates, err := h.Client.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates: %v", err)
	}
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}

	// Delete.
	if err := h.Client.DeleteTemplate(created.ID); err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}

	templates, err = h.Client.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates after delete: %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("expected 0 templates after delete, got %d", len(templates))
	}
}

func TestTemplates_HTMLOnlyIsValid(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "HTML Only",
		Subject:  "Test",
		HTMLBody: "<p>html only</p>",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if created.TextBody != "" {
		t.Errorf("TextBody = %q, want empty", created.TextBody)
	}
}

func TestTemplates_TextOnlyIsValid(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Text Only",
		Subject:  "Test",
		TextBody: "plain text only",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if created.HTMLBody != "" {
		t.Errorf("HTMLBody = %q, want empty", created.HTMLBody)
	}
}

func TestTemplates_DeleteNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeleteTemplate("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestTemplates_UpdateNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.UpdateTemplate("nonexistent", sdk.UpdateTemplateRequest{
		Name:    strPtr("Test"),
		Subject: strPtr("Test"),
	})
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestTemplates_RenderHTML(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	result, err := h.Client.RenderTemplateHTML(sdk.RenderHTMLRequest{
		HTMLBody: "<p>Hello {{.FirstName}} {{.LastName}}, click <a href=\"{{.URL}}\">here</a>.</p>",
		PreviewTemplateRequest: sdk.PreviewTemplateRequest{
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "alice@acme.com",
			URL:       "https://phish.example.com/abc",
		},
	})
	if err != nil {
		t.Fatalf("RenderTemplateHTML: %v", err)
	}
	want := `<p>Hello Alice Smith, click <a href="https://phish.example.com/abc">here</a>.</p>`
	if result.HTMLBody != want {
		t.Errorf("HTMLBody = %q, want %q", result.HTMLBody, want)
	}
}

func TestTemplates_RenderHTML_EmptyBody(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.RenderTemplateHTML(sdk.RenderHTMLRequest{})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	var ve *sdk.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	if ve.Message != "HTML body is required." {
		t.Errorf("Message = %q, want %q", ve.Message, "HTML body is required.")
	}
}

func TestTemplates_RenderHTML_InvalidTemplate(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.RenderTemplateHTML(sdk.RenderHTMLRequest{
		HTMLBody: "<p>{{.Unclosed</p>",
	})
	wantError(t, err, http.StatusUnprocessableEntity, "Failed to render template.")
}

func TestTemplates_PartialUpdate(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Original",
		Subject:  "Original Subject",
		HTMLBody: "<p>Original body</p>",
		TextBody: "Original text",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Update only the subject — other fields should be preserved.
	updated, err := h.Client.UpdateTemplate(created.ID, sdk.UpdateTemplateRequest{
		Subject: strPtr("New Subject"),
	})
	if err != nil {
		t.Fatalf("UpdateTemplate: %v", err)
	}
	if updated.Subject != "New Subject" {
		t.Errorf("Subject = %q, want %q", updated.Subject, "New Subject")
	}
	if updated.Name != "Original" {
		t.Errorf("Name = %q, want %q (should be preserved)", updated.Name, "Original")
	}
	if updated.HTMLBody != "<p>Original body</p>" {
		t.Errorf("HTMLBody = %q (should be preserved)", updated.HTMLBody)
	}
	if updated.TextBody != "Original text" {
		t.Errorf("TextBody = %q (should be preserved)", updated.TextBody)
	}
}
