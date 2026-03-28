package test_test

import (
	"testing"

	"github.com/travisbale/mirador/sdk"
	"github.com/travisbale/mirador/test"
)

// createPrerequisites sets up a target list, template, and SMTP profile
// and returns their IDs for use in campaign creation.
func createPrerequisites(t *testing.T, h *test.Harness) (targetListID, templateID, smtpProfileID string) {
	t.Helper()

	list, err := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Campaign Targets"})
	if err != nil {
		t.Fatalf("CreateTargetList: %v", err)
	}
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "alice@acme.com", FirstName: "Alice"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "bob@acme.com", FirstName: "Bob"})

	tmpl, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Test Template",
		Subject:  "Action Required",
		HTMLBody: "<p>Click {{.URL}}</p>",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	smtp, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Test SMTP",
		Host:     "smtp.example.com",
		FromAddr: "it@example.com",
	})
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}

	return list.ID, tmpl.ID, smtp.ID
}

func TestCampaigns_CRUD(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Q1 Phishing",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Status != "draft" {
		t.Errorf("Status = %q, want %q", created.Status, "draft")
	}
	if created.SendRate != 10 {
		t.Errorf("SendRate = %d, want 10 (default)", created.SendRate)
	}

	// Get.
	got, err := h.Client.GetCampaign(created.ID)
	if err != nil {
		t.Fatalf("GetCampaign: %v", err)
	}
	if got.Name != "Q1 Phishing" {
		t.Errorf("Name = %q", got.Name)
	}

	// List.
	campaigns, err := h.Client.ListCampaigns()
	if err != nil {
		t.Fatalf("ListCampaigns: %v", err)
	}
	if len(campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(campaigns))
	}

	// Delete.
	if err := h.Client.DeleteCampaign(created.ID); err != nil {
		t.Fatalf("DeleteCampaign: %v", err)
	}

	campaigns, _ = h.Client.ListCampaigns()
	if len(campaigns) != 0 {
		t.Errorf("expected 0 campaigns after delete, got %d", len(campaigns))
	}
}

func TestCampaigns_PrePopulatesResults(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Results Test",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	results, err := h.Client.ListCampaignResults(created.ID)
	if err != nil {
		t.Fatalf("ListCampaignResults: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (one per target), got %d", len(results))
	}

	emails := map[string]bool{}
	for _, r := range results {
		emails[r.Email] = true
		if r.Status != "pending" {
			t.Errorf("result status = %q, want %q", r.Status, "pending")
		}
	}
	if !emails["alice@acme.com"] || !emails["bob@acme.com"] {
		t.Errorf("expected results for alice and bob, got %v", emails)
	}
}

func TestCampaigns_RequiresName(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	_, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCampaigns_RequiresTemplate(t *testing.T) {
	h := test.NewHarness(t)
	listID, _, smtpID := createPrerequisites(t, h)

	_, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Test",
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})
	if err == nil {
		t.Error("expected error for missing template")
	}
}

func TestCampaigns_RejectsInvalidReferences(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	// Invalid template.
	_, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Bad Template",
		TemplateID:    "nonexistent",
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})
	if err == nil {
		t.Error("expected error for invalid template reference")
	}

	// Invalid SMTP profile.
	_, err = h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Bad SMTP",
		TemplateID:    tmplID,
		SMTPProfileID: "nonexistent",
		TargetListID:  listID,
	})
	if err == nil {
		t.Error("expected error for invalid SMTP profile reference")
	}

	// Invalid target list.
	_, err = h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Bad List",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  "nonexistent",
	})
	if err == nil {
		t.Error("expected error for invalid target list reference")
	}
}

func TestCampaigns_DeleteNotFound(t *testing.T) {
	h := test.NewHarness(t)

	err := h.Client.DeleteCampaign("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent campaign")
	}
}
