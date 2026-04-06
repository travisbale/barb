package test_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestCampaigns_CRUD(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.Name = "Q1 Phishing"
	created, err := h.Client.CreateCampaign(req)
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
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))
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

func TestCampaigns_RejectsInvalidReferences(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	// Invalid template.
	req := validCampaignRequest(listID, tmplID, smtpID)
	req.TemplateID = "nonexistent"
	_, err := h.Client.CreateCampaign(req)
	wantError(t, err, http.StatusUnprocessableEntity, "template not found")

	// Invalid SMTP profile.
	req = validCampaignRequest(listID, tmplID, smtpID)
	req.SMTPProfileID = "nonexistent"
	_, err = h.Client.CreateCampaign(req)
	wantError(t, err, http.StatusUnprocessableEntity, "SMTP profile not found")

	// Invalid target list.
	req = validCampaignRequest(listID, tmplID, smtpID)
	req.TargetListID = "nonexistent"
	_, err = h.Client.CreateCampaign(req)
	wantError(t, err, http.StatusUnprocessableEntity, "target list not found")
}

func TestCampaigns_Start(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.SendRate = 600 // fast — 10 per second
	created, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := h.Client.StartCampaign(created.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	// Wait for all emails to be sent.
	waitForEmails(t, h, 2)

	// Campaign stays active until explicitly completed.
	got, err := h.Client.GetCampaign(created.ID)
	if err != nil {
		t.Fatalf("GetCampaign: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}

	// Complete the campaign.
	if err := h.Client.CompleteCampaign(created.ID); err != nil {
		t.Fatalf("CompleteCampaign: %v", err)
	}
	got, _ = h.Client.GetCampaign(created.ID)
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
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.SendRate = 600
	created, _ := h.Client.CreateCampaign(req)

	// Start it once.
	h.Client.StartCampaign(created.ID)
	waitForCampaignStatus(t, h, created.ID, "active")

	// Starting again should fail — it's now active, not draft.
	err := h.Client.StartCampaign(created.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "campaign can only be started from draft status")
}

func TestCampaigns_ResultStatusesAfterCompletion(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.SendRate = 600
	created, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	h.Client.StartCampaign(created.ID)
	waitForEmails(t, h, 2)
	h.Client.CompleteCampaign(created.ID)

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
	t.Parallel()
	h, bm := test.NewHarnessWithBlockingMailer(t)

	targets := make([]sdk.AddTargetRequest, 5)
	for i := range targets {
		targets[i] = sdk.AddTargetRequest{Email: fmt.Sprintf("user%d@acme.com", i)}
	}
	list := createTestTargetList(t, h, targets...)
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	created, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := h.Client.StartCampaign(created.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	// Let exactly 2 emails through, then cancel while the rest are blocked.
	bm.Release(2)
	waitForEmails(t, h, 2)

	if err := h.Client.CancelCampaign(created.ID); err != nil {
		t.Fatalf("CancelCampaign: %v", err)
	}

	got, err := h.Client.GetCampaign(created.ID)
	if err != nil {
		t.Fatalf("GetCampaign: %v", err)
	}
	if got.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", got.Status, "cancelled")
	}

	// Exactly 2 emails should have been sent — deterministic, not timing-dependent.
	if h.Mailer.Count() != 2 {
		t.Errorf("expected 2 emails sent, got %d", h.Mailer.Count())
	}
}

func TestCampaigns_CancelNotRunning(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))

	// Cancel a draft campaign — should fail.
	err := h.Client.CancelCampaign(created.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "campaign is not running")
}

func TestCampaigns_CancelAlreadyCompleted(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.SendRate = 600
	created, _ := h.Client.CreateCampaign(req)

	h.Client.StartCampaign(created.ID)
	waitForCampaignStatus(t, h, created.ID, "active")

	// Complete it explicitly.
	h.Client.CompleteCampaign(created.ID)

	// Cancelling a completed campaign should fail.
	err := h.Client.CancelCampaign(created.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "campaign is not running")
}

func TestCampaigns_Update(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.Name = "Original Name"
	req.SendRate = 5
	created, err := h.Client.CreateCampaign(req)
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
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))
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
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))

	_, err := h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		TemplateID: strPtr("nonexistent"),
	})
	wantError(t, err, http.StatusUnprocessableEntity, "template not found")

	_, err = h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		SMTPProfileID: strPtr("nonexistent"),
	})
	wantError(t, err, http.StatusUnprocessableEntity, "SMTP profile not found")

	_, err = h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		TargetListID: strPtr("nonexistent"),
	})
	wantError(t, err, http.StatusUnprocessableEntity, "target list not found")
}

func TestCampaigns_UpdateRejectsNonDraft(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.SendRate = 600
	created, _ := h.Client.CreateCampaign(req)

	h.Client.StartCampaign(created.ID)
	waitForCampaignStatus(t, h, created.ID, "active")
	h.Client.CompleteCampaign(created.ID)

	// Campaign is now completed — update should fail.
	_, err := h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		Name: strPtr("Should Fail"),
	})
	wantError(t, err, http.StatusUnprocessableEntity, "campaign can only be started from draft status")
}

func TestCampaigns_UpdateNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.UpdateCampaign("nonexistent", sdk.UpdateCampaignRequest{
		Name: strPtr("Nope"),
	})
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestCampaigns_SendTestEmail(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))
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
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))

	err := h.Client.SendTestEmail(created.ID, sdk.SendTestEmailRequest{Email: ""})
	wantError(t, err, http.StatusUnprocessableEntity, "email is required")
}

func TestCampaigns_DeleteRejectsActive(t *testing.T) {
	t.Parallel()
	h, _ := test.NewHarnessWithBlockingMailer(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	h.Client.StartCampaign(created.ID)

	err = h.Client.DeleteCampaign(created.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "Active campaigns cannot be deleted")
}

func TestCampaigns_DeleteNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeleteCampaign("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestCampaigns_InvalidTemplateSyntax(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	// Templates with broken syntax are accepted at creation time — the syntax
	// is only validated when rendering. The preview endpoint catches this.
	tmpl := createTestTemplate(t, h, func(r *sdk.CreateTemplateRequest) {
		r.HTMLBody = "<p>{{.Unclosed</p>"
	})

	_, err := h.Client.PreviewTemplate(tmpl.ID, sdk.PreviewTemplateRequest{
		FirstName: "Alice",
	})
	wantError(t, err, http.StatusUnprocessableEntity, "rendering HTML body")
}

func TestCampaigns_EmptyTargetList(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	// Create a target list with zero targets.
	list, err := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Empty List"})
	if err != nil {
		t.Fatalf("CreateTargetList: %v", err)
	}
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	created, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Verify no results were pre-populated.
	results, _ := h.Client.ListCampaignResults(created.ID)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty target list, got %d", len(results))
	}

	// Starting the campaign should succeed — sending loop simply has nothing to send.
	if err := h.Client.StartCampaign(created.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	waitForCampaignStatus(t, h, created.ID, "active")

	// No emails should have been sent.
	if h.Mailer.Count() != 0 {
		t.Errorf("expected 0 emails sent, got %d", h.Mailer.Count())
	}
}

func TestCampaigns_ConcurrentStartAndCancel(t *testing.T) {
	t.Parallel()
	h, bm := test.NewHarnessWithBlockingMailer(t)

	list := createTestTargetList(t, h)
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	created, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Start the campaign — sends will block until released.
	if err := h.Client.StartCampaign(created.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}
	waitForCampaignStatus(t, h, created.ID, "active")

	// Cancel immediately while sends are blocked.
	if err := h.Client.CancelCampaign(created.ID); err != nil {
		t.Fatalf("CancelCampaign: %v", err)
	}

	got, _ := h.Client.GetCampaign(created.ID)
	if got.Status != "cancelled" {
		t.Errorf("Status = %q, want %q", got.Status, "cancelled")
	}

	// Release sends so the goroutine can clean up.
	bm.Release(10)
}

func TestCampaigns_ConcurrentCompleteAndCancel(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.SendRate = 600
	created, _ := h.Client.CreateCampaign(req)

	h.Client.StartCampaign(created.ID)
	waitForEmails(t, h, 2)

	// Complete immediately.
	h.Client.CompleteCampaign(created.ID)

	// Cancel after complete — should fail, but must not panic.
	err := h.Client.CancelCampaign(created.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "campaign is not running")

	got, _ := h.Client.GetCampaign(created.ID)
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}
}

func TestCampaigns_UpdateTargetListDoesNotRepopulateResults(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, err := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Verify original results (alice and bob from createPrerequisites).
	results, _ := h.Client.ListCampaignResults(created.ID)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Create a different target list.
	newList := createTestTargetList(t, h,
		sdk.AddTargetRequest{Email: "carol@acme.com"},
		sdk.AddTargetRequest{Email: "dave@acme.com"},
		sdk.AddTargetRequest{Email: "eve@acme.com"},
	)

	// Update the campaign's target list.
	_, err = h.Client.UpdateCampaign(created.ID, sdk.UpdateCampaignRequest{
		TargetListID: strPtr(newList.ID),
	})
	if err != nil {
		t.Fatalf("UpdateCampaign: %v", err)
	}

	// Results should still reflect the original target list — the update
	// does not re-populate results (this documents current behavior).
	results, _ = h.Client.ListCampaignResults(created.ID)
	if len(results) != 2 {
		t.Errorf("expected 2 results (unchanged), got %d", len(results))
	}
	emails := map[string]bool{}
	for _, r := range results {
		emails[r.Email] = true
	}
	if !emails["alice@acme.com"] || !emails["bob@acme.com"] {
		t.Errorf("expected original results (alice, bob), got %v", emails)
	}
}

func TestCampaigns_SendTestEmailOnActiveCampaign(t *testing.T) {
	t.Parallel()
	h, bm := test.NewHarnessWithBlockingMailer(t)

	list := createTestTargetList(t, h)
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	created, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Start the campaign — sends will block.
	h.Client.StartCampaign(created.ID)
	waitForCampaignStatus(t, h, created.ID, "active")

	// Release enough sends for both the campaign emails (2 targets) and the
	// test email (1). The campaign goroutine and SendTestEmail each open their
	// own connection, so all sends share the same gate.
	bm.Release(5)

	err = h.Client.SendTestEmail(created.ID, sdk.SendTestEmailRequest{Email: "tester@example.com"})
	if err != nil {
		t.Errorf("SendTestEmail on active campaign: %v", err)
	}

	h.Client.CancelCampaign(created.ID)
}

func TestCampaigns_CompleteRequiresRunning(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	created, _ := h.Client.CreateCampaign(validCampaignRequest(listID, tmplID, smtpID))

	// Completing a draft campaign should fail.
	err := h.Client.CompleteCampaign(created.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "campaign is not running")
}

func TestCampaigns_CompleteAlreadyCompleted(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)
	listID, tmplID, smtpID := createPrerequisites(t, h)

	req := validCampaignRequest(listID, tmplID, smtpID)
	req.SendRate = 600
	created, _ := h.Client.CreateCampaign(req)

	h.Client.StartCampaign(created.ID)
	waitForEmails(t, h, 2)
	h.Client.CompleteCampaign(created.ID)

	// Completing again should fail.
	err := h.Client.CompleteCampaign(created.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "campaign is not running")
}

func TestCampaigns_SendTestEmailNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.SendTestEmail("nonexistent", sdk.SendTestEmailRequest{Email: "test@example.com"})
	wantError(t, err, http.StatusNotFound, "not found")
}
