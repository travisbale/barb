package test_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestStream_ResultUpdates(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list := createTestTargetList(t, h)
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	campaign, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Connect to the SSE stream before starting the campaign.
	ch, cancel, err := h.Client.StreamCampaign(campaign.ID)
	if err != nil {
		t.Fatalf("StreamCampaign: %v", err)
	}
	defer cancel()

	// Start the campaign — triggers email sends and result.updated events.
	if err := h.Client.StartCampaign(campaign.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	// Collect result.updated events for both targets.
	sentEmails := map[string]bool{}
	timeout := time.After(10 * time.Second)
	for len(sentEmails) < 2 {
		select {
		case evt, ok := <-ch:
			if !ok {
				t.Fatal("stream closed unexpectedly")
			}
			if evt.Type == sdk.EventResultUpdated && evt.Result != nil && evt.Result.Status == "sent" {
				sentEmails[evt.Result.Email] = true
			}
		case <-timeout:
			t.Fatalf("timed out waiting for result events; got %d", len(sentEmails))
		}
	}

	if !sentEmails["alice@acme.com"] || !sentEmails["bob@acme.com"] {
		t.Errorf("expected events for alice and bob, got %v", sentEmails)
	}
}

func TestStream_CampaignStatusChange(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list := createTestTargetList(t, h)
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	campaign, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	ch, cancel, err := h.Client.StreamCampaign(campaign.ID)
	if err != nil {
		t.Fatalf("StreamCampaign: %v", err)
	}
	defer cancel()

	// Start and wait for emails to send.
	h.Client.StartCampaign(campaign.ID)
	waitForEmails(t, h, 2)

	// Complete the campaign.
	h.Client.CompleteCampaign(campaign.ID)

	// Read events until we see the completed status.
	timeout := time.After(10 * time.Second)
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				t.Fatal("stream closed unexpectedly")
			}
			if evt.Type == sdk.EventCampaignStatus && evt.Status == "completed" {
				return // success
			}
		case <-timeout:
			t.Fatal("timed out waiting for campaign.status event")
		}
	}
}

func TestStream_CatchesUpOnLateConnect(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list := createTestTargetList(t, h)
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	campaign, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Start the campaign and wait for all emails to send BEFORE connecting.
	h.Client.StartCampaign(campaign.ID)
	waitForEmails(t, h, 2)

	// Now connect to the stream — should receive the initial snapshot
	// with both results already in "sent" status.
	ch, cancel, err := h.Client.StreamCampaign(campaign.ID)
	if err != nil {
		t.Fatalf("StreamCampaign: %v", err)
	}
	defer cancel()

	sentEmails := map[string]bool{}
	gotActiveStatus := false
	timeout := time.After(5 * time.Second)
	for !gotActiveStatus || len(sentEmails) < 2 {
		select {
		case evt, ok := <-ch:
			if !ok {
				t.Fatal("stream closed unexpectedly")
			}
			if evt.Type == sdk.EventResultUpdated && evt.Result != nil && evt.Result.Status == "sent" {
				sentEmails[evt.Result.Email] = true
			}
			if evt.Type == sdk.EventCampaignStatus && evt.Status == "active" {
				gotActiveStatus = true
			}
		case <-timeout:
			t.Fatalf("timed out waiting for catch-up events; got %d results, activeStatus=%v", len(sentEmails), gotActiveStatus)
		}
	}

	if !sentEmails["alice@acme.com"] || !sentEmails["bob@acme.com"] {
		t.Errorf("expected catch-up events for alice and bob, got %v", sentEmails)
	}
}

func TestStream_NotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, _, err := h.Client.StreamCampaign("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}
