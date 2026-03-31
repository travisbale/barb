package phishing

import (
	"context"
	"log/slog"
	"time"

	miragesdk "github.com/travisbale/mirage/sdk"
)

// SessionMonitor watches miraged instances for captured sessions and
// correlates them to active campaign results by matching the session's
// captured username against the result's email.
type SessionMonitor struct {
	Campaigns campaignStore
	Miraged   *MiragedService
	Logger    *slog.Logger
}

// Watch connects to the miraged instance's SSE stream and processes
// session events until the context is cancelled or the stream ends.
func (m *SessionMonitor) Watch(ctx context.Context, miragedID string) {
	client, err := m.Miraged.client(miragedID)
	if err != nil {
		m.Logger.Error("failed to connect to miraged for monitoring", "miraged_id", miragedID, "error", err)
		return
	}

	ch, cancel, err := client.StreamSessions()
	if err != nil {
		m.Logger.Error("failed to start session stream", "miraged_id", miragedID, "error", err)
		return
	}
	defer cancel()

	m.Logger.Info("session monitor started", "miraged_id", miragedID)

	for {
		select {
		case <-ctx.Done():
			m.Logger.Info("session monitor stopped", "miraged_id", miragedID)
			return
		case event, ok := <-ch:
			if !ok {
				m.Logger.Warn("session stream closed", "miraged_id", miragedID)
				return
			}
			m.handleEvent(miragedID, event)
		}
	}
}

func (m *SessionMonitor) handleEvent(miragedID string, event miragesdk.SessionEvent) {
	switch event.Type {
	case miragesdk.EventCredsCaptured, miragesdk.EventSessionCompleted:
	default:
		return
	}

	session := event.Session
	if session.Username == "" {
		return
	}

	campaigns, err := m.Campaigns.ListActiveCampaignsByMiraged(miragedID)
	if err != nil {
		m.Logger.Error("failed to list campaigns for correlation", "error", err)
		return
	}

	for _, campaign := range campaigns {
		if campaign.Phishlet != "" && campaign.Phishlet != session.Phishlet {
			continue
		}
		m.correlate(campaign, session)
	}
}

func (m *SessionMonitor) correlate(campaign *Campaign, session miragesdk.SessionResponse) {
	result, err := m.Campaigns.GetResultByEmail(campaign.ID, session.Username)
	if err != nil {
		return // not found or DB error — nothing to correlate
	}
	if result.SessionID != "" {
		return // already correlated
	}

	// Record the click time from when the session started (target first visited the lure).
	clickedAt := session.StartedAt
	result.ClickedAt = &clickedAt

	now := time.Now()
	result.Status = ResultCaptured
	result.CapturedAt = &now
	result.SessionID = session.ID

	if err := m.Campaigns.UpdateResult(result); err != nil {
		m.Logger.Error("failed to update result", "result_id", result.ID, "error", err)
	} else {
		m.Logger.Info("session correlated",
			"campaign_id", campaign.ID,
			"email", result.Email,
			"session_id", session.ID,
		)
	}
}
