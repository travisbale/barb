package test_test

import (
	"testing"

	"github.com/travisbale/mirador/sdk"
	"github.com/travisbale/mirador/test"
)

func TestTemplates_CRUD(t *testing.T) {
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
		Name:     "Password Expiry",
		Subject:  "Action required: password expiring",
		HTMLBody: got.HTMLBody,
		TextBody: got.TextBody,
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

func TestTemplates_RequiresName(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Subject:  "Test",
		HTMLBody: "<p>test</p>",
	})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestTemplates_RequiresSubject(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Test",
		HTMLBody: "<p>test</p>",
	})
	if err == nil {
		t.Error("expected error for missing subject")
	}
}

func TestTemplates_RequiresBody(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:    "Test",
		Subject: "Test",
	})
	if err == nil {
		t.Error("expected error for missing body")
	}
}

func TestTemplates_HTMLOnlyIsValid(t *testing.T) {
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
	h := test.NewHarness(t)

	err := h.Client.DeleteTemplate("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestTemplates_UpdateNotFound(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.UpdateTemplate("nonexistent", sdk.UpdateTemplateRequest{
		Name:     "Test",
		Subject:  "Test",
		HTMLBody: "<p>test</p>",
	})
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}
