package test_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
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

func TestCampaigns_Start(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Start Test",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
		SendRate:      600, // fast — 10 per second
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := h.Client.StartCampaign(created.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	// Wait for the background goroutine to finish sending.
	time.Sleep(1 * time.Second)

	// Verify the campaign status changed.
	got, err := h.Client.GetCampaign(created.ID)
	if err != nil {
		t.Fatalf("GetCampaign: %v", err)
	}
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}

	// Verify results were updated.
	results, err := h.Client.ListCampaignResults(created.ID)
	if err != nil {
		t.Fatalf("ListCampaignResults: %v", err)
	}
	for _, r := range results {
		if r.Status != "sent" {
			t.Errorf("result %s status = %q, want %q", r.Email, r.Status, "sent")
		}
	}

	// Verify the mock mailer was called.
	if h.Mailer.Count() != 2 {
		t.Errorf("expected 2 emails sent, got %d", h.Mailer.Count())
	}
}

func TestCampaigns_StartRequiresDraft(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Draft Test",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
		SendRate:      600,
	})

	// Start it once.
	h.Client.StartCampaign(created.ID)
	time.Sleep(500 * time.Millisecond)

	// Starting again should fail — it's now completed, not draft.
	err := h.Client.StartCampaign(created.ID)
	if err == nil {
		t.Error("expected error starting a non-draft campaign")
	}
}

func TestCampaigns_ResultStatusesAfterCompletion(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Results Test", TemplateID: tmplID, SMTPProfileID: smtpID, TargetListID: listID, SendRate: 600,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	h.Client.StartCampaign(created.ID)
	time.Sleep(1 * time.Second)

	results, err := h.Client.ListCampaignResults(created.ID)
	if err != nil {
		t.Fatalf("ListCampaignResults: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, result := range results {
		if result.Status != "sent" {
			t.Errorf("result %s status = %q, want %q", result.Email, result.Status, "sent")
		}
		if result.SentAt == nil {
			t.Errorf("result %s SentAt is nil, expected a timestamp", result.Email)
		}
	}
}

func TestCampaigns_Cancel(t *testing.T) {
	h := test.NewHarness(t)

	// Create a campaign with many targets and slow send rate so it's
	// still running when we cancel.
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Cancel Targets"})
	for i := 0; i < 20; i++ {
		h.Client.AddTarget(list.ID, sdk.AddTargetRequest{
			Email: fmt.Sprintf("user%d@acme.com", i),
		})
	}

	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Cancel Template",
		Subject:  "Test",
		HTMLBody: "<p>{{.URL}}</p>",
	})
	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Cancel SMTP",
		Host:     "smtp.example.com",
		FromAddr: "it@example.com",
	})

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Cancel Test",
		TemplateID:    tmpl.ID,
		SMTPProfileID: smtp.ID,
		TargetListID:  list.ID,
		SendRate:      1, // 1 per minute — very slow, so it's still running when we cancel
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := h.Client.StartCampaign(created.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	// Give the goroutine a moment to start sending.
	// Give the goroutine a moment to start sending.
	time.Sleep(200 * time.Millisecond)

	// Cancel it.
	if err := h.Client.CancelCampaign(created.ID); err != nil {
		t.Fatalf("CancelCampaign: %v", err)
	}

	// Verify the status is cancelled.
	got, err := h.Client.GetCampaign(created.ID)
	if err != nil {
		t.Fatalf("GetCampaign: %v", err)
	}
	if got.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", got.Status, "cancelled")
	}

	// Not all emails should have been sent (campaign was interrupted).
	if h.Mailer.Count() >= 20 {
		t.Errorf("expected fewer than 20 emails sent (campaign was cancelled), got %d", h.Mailer.Count())
	}
}

func TestCampaigns_CancelNotRunning(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Not Running",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})

	// Cancel a draft campaign — should fail.
	err := h.Client.CancelCampaign(created.ID)
	if err == nil {
		t.Error("expected error cancelling a draft campaign")
	}
}

func TestCampaigns_CancelAlreadyCompleted(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Already Done",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
		SendRate:      600,
	})

	h.Client.StartCampaign(created.ID)
	time.Sleep(1 * time.Second)

	// Campaign should be completed by now.
	err := h.Client.CancelCampaign(created.ID)
	if err == nil {
		t.Error("expected error cancelling a completed campaign")
	}
}

func TestCampaigns_Update(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Original Name",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
		SendRate:      5,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Update name and send rate.
	updated, err := h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		Name:     strPtr("Updated Name"),
		SendRate: intPtr(20),
	})
	if err != nil {
		t.Fatalf("UpdateCampaign: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Name = %q, want %q", updated.Name, "Updated Name")
	}
	if updated.SendRate != 20 {
		t.Errorf("SendRate = %d, want 20", updated.SendRate)
	}

	// Unchanged fields should be preserved.
	if updated.TemplateID != tmplID {
		t.Errorf("TemplateID changed unexpectedly")
	}
	if updated.SMTPProfileID != smtpID {
		t.Errorf("SMTPProfileID changed unexpectedly")
	}

	// Verify via Get.
	got, err := h.Client.GetCampaign(created.ID)
	if err != nil {
		t.Fatalf("GetCampaign: %v", err)
	}
	if got.Name != "Updated Name" || got.SendRate != 20 {
		t.Errorf("Get returned stale data: name=%q sendRate=%d", got.Name, got.SendRate)
	}
}

func TestCampaigns_UpdateReferences(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Ref Update",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Create a second template and switch to it.
	newTmpl, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "New Template",
		Subject:  "Updated Subject",
		HTMLBody: "<p>New {{.URL}}</p>",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	updated, err := h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		TemplateID: strPtr(newTmpl.ID),
	})
	if err != nil {
		t.Fatalf("UpdateCampaign: %v", err)
	}
	if updated.TemplateID != newTmpl.ID {
		t.Errorf("TemplateID = %q, want %q", updated.TemplateID, newTmpl.ID)
	}

	// Verify persisted via Get.
	got, _ := h.Client.GetCampaign(created.ID)
	if got.TemplateID != newTmpl.ID {
		t.Errorf("Get returned stale TemplateID: %q", got.TemplateID)
	}
}

func TestCampaigns_UpdateRejectsInvalidReferences(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Bad Ref Update",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})

	_, err := h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		TemplateID: strPtr("nonexistent"),
	})
	if err == nil {
		t.Error("expected error for invalid template reference")
	}

	_, err = h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		SMTPProfileID: strPtr("nonexistent"),
	})
	if err == nil {
		t.Error("expected error for invalid SMTP profile reference")
	}

	_, err = h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		TargetListID: strPtr("nonexistent"),
	})
	if err == nil {
		t.Error("expected error for invalid target list reference")
	}
}

func TestCampaigns_UpdateRejectsNonDraft(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Started Campaign",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
		SendRate:      600,
	})

	h.Client.StartCampaign(created.ID)
	time.Sleep(1 * time.Second)

	// Campaign is now completed — update should fail.
	_, err := h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		Name: strPtr("Should Fail"),
	})
	if err == nil {
		t.Error("expected error updating a non-draft campaign")
	}
}

func TestCampaigns_UpdateNotFound(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.UpdateCampaign("nonexistent", sdk.UpdateCampaignRequest{
		Name: strPtr("Nope"),
	})
	if err == nil {
		t.Error("expected error for nonexistent campaign")
	}
}

func TestCampaigns_SendTestEmail(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Test Email Campaign",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Send a test email.
	if err := h.Client.SendTestEmail(created.ID, sdk.SendTestEmailRequest{Email: "tester@example.com"}); err != nil {
		t.Fatalf("SendTestEmail: %v", err)
	}

	if h.Mailer.Count() != 1 {
		t.Errorf("expected 1 email sent, got %d", h.Mailer.Count())
	}

	// Sending a second test should also work (no lure conflicts).
	if err := h.Client.SendTestEmail(created.ID, sdk.SendTestEmailRequest{Email: "tester2@example.com"}); err != nil {
		t.Fatalf("second SendTestEmail: %v", err)
	}
	if h.Mailer.Count() != 2 {
		t.Errorf("expected 2 emails sent, got %d", h.Mailer.Count())
	}
}

func TestCampaigns_SendTestEmailRequiresAddress(t *testing.T) {
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "No Email",
		TemplateID:    tmplID,
		SMTPProfileID: smtpID,
		TargetListID:  listID,
	})

	err := h.Client.SendTestEmail(created.ID, sdk.SendTestEmailRequest{Email: ""})
	if err == nil {
		t.Error("expected error for empty email address")
	}
}

func TestCampaigns_DeleteNotFound(t *testing.T) {
	h := test.NewHarness(t)

	err := h.Client.DeleteCampaign("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent campaign")
	}
}
