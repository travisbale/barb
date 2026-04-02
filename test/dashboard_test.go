package test_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestDashboard_Empty(t *testing.T) {
	h := test.NewHarness(t)

	dashboard, err := h.Client.Dashboard()
	if err != nil {
		t.Fatalf("Dashboard: %v", err)
	}
	if dashboard.Campaigns.Total != 0 {
		t.Errorf("Campaigns.Total = %d, want 0", dashboard.Campaigns.Total)
	}
	if dashboard.TotalCaptures != 0 {
		t.Errorf("TotalCaptures = %d, want 0", dashboard.TotalCaptures)
	}
	if dashboard.MiragedCount != 0 {
		t.Errorf("MiragedCount = %d, want 0", dashboard.MiragedCount)
	}
	if len(dashboard.ActiveCampaigns) != 0 {
		t.Errorf("expected no active campaigns, got %d", len(dashboard.ActiveCampaigns))
	}
	if len(dashboard.RecentCaptures) != 0 {
		t.Errorf("expected no recent captures, got %d", len(dashboard.RecentCaptures))
	}
}

func TestDashboard_CampaignCounts(t *testing.T) {
	h := test.NewHarness(t)

	// Create prerequisites.
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Dashboard Targets"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "alice@acme.com"})
	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name: "Dashboard Template", Subject: "Test", HTMLBody: "<p>test</p>",
	})
	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Dashboard SMTP", Host: "smtp.example.com", FromAddr: "from@example.com",
	})

	// Create a draft campaign.
	h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Draft One", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID, RedirectURL: "https://example.com",
	})

	// Create, start, and complete a campaign.
	started, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Started One", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID, RedirectURL: "https://example.com", SendRate: 600,
	})
	h.Client.StartCampaign(started.ID)
	time.Sleep(1 * time.Second)
	h.Client.CompleteCampaign(started.ID)

	dashboard, err := h.Client.Dashboard()
	if err != nil {
		t.Fatalf("Dashboard: %v", err)
	}
	if dashboard.Campaigns.Total != 2 {
		t.Errorf("Campaigns.Total = %d, want 2", dashboard.Campaigns.Total)
	}
	if dashboard.Campaigns.Draft != 1 {
		t.Errorf("Campaigns.Draft = %d, want 1", dashboard.Campaigns.Draft)
	}
	if dashboard.Campaigns.Completed != 1 {
		t.Errorf("Campaigns.Completed = %d, want 1", dashboard.Campaigns.Completed)
	}
}

func TestDashboard_ActiveCampaignProgress(t *testing.T) {
	h := test.NewHarness(t)

	// Create prerequisites with many targets.
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Dashboard Active Targets"})
	for i := 0; i < 20; i++ {
		h.Client.AddTarget(list.ID, sdk.AddTargetRequest{
			Email: fmt.Sprintf("user%d@acme.com", i),
		})
	}
	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name: "Dashboard Active Template", Subject: "Test", HTMLBody: "<p>test</p>",
	})
	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Dashboard Active SMTP", Host: "smtp.example.com", FromAddr: "from@example.com",
	})

	// Start a campaign with slow send rate so it's still active.
	campaign, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Active Dashboard", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID, RedirectURL: "https://example.com", SendRate: 1,
	})
	h.Client.StartCampaign(campaign.ID)
	time.Sleep(200 * time.Millisecond)

	dashboard, err := h.Client.Dashboard()
	if err != nil {
		t.Fatalf("Dashboard: %v", err)
	}
	if len(dashboard.ActiveCampaigns) != 1 {
		t.Fatalf("expected 1 active campaign, got %d", len(dashboard.ActiveCampaigns))
	}
	active := dashboard.ActiveCampaigns[0]
	if active.Name != "Active Dashboard" {
		t.Errorf("Name = %q", active.Name)
	}
	if active.Total != 20 {
		t.Errorf("Total = %d, want 20", active.Total)
	}

	// Clean up — cancel so the goroutine stops.
	h.Client.CancelCampaign(campaign.ID)
}

func TestDashboard_TemplatePreview(t *testing.T) {
	h := test.NewHarness(t)

	tmpl, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Preview Test",
		Subject:  "Hello {{.FirstName}}",
		HTMLBody: "<p>Dear {{.FirstName}} {{.LastName}}, click <a href=\"{{.URL}}\">here</a>.</p>",
		TextBody: "Dear {{.FirstName}}, visit {{.URL}}",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	preview, err := h.Client.PreviewTemplate(tmpl.ID, sdk.PreviewTemplateRequest{
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@acme.com",
		URL:       "https://phish.example.com/abc",
	})
	if err != nil {
		t.Fatalf("PreviewTemplate: %v", err)
	}
	if preview.Subject != "Hello Alice" {
		t.Errorf("Subject = %q, want %q", preview.Subject, "Hello Alice")
	}
	if preview.HTMLBody != `<p>Dear Alice Smith, click <a href="https://phish.example.com/abc">here</a>.</p>` {
		t.Errorf("HTMLBody = %q", preview.HTMLBody)
	}
	if preview.TextBody != "Dear Alice, visit https://phish.example.com/abc" {
		t.Errorf("TextBody = %q", preview.TextBody)
	}
}

func TestDashboard_TemplatePreviewNotFound(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.PreviewTemplate("nonexistent", sdk.PreviewTemplateRequest{
		FirstName: "Alice",
	})
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}
