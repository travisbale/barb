package phishing

import (
	"context"
	"fmt"
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
// session events until the context is cancelled. Automatically reconnects
// with exponential backoff if the stream fails or disconnects.
func (m *SessionMonitor) Watch(ctx context.Context, miragedID string) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		start := time.Now()
		err := m.stream(ctx, miragedID)
		if ctx.Err() != nil {
			m.Logger.Info("session monitor stopped", "miraged_id", miragedID)
			return
		}

		// Reset backoff if the stream was connected for a reasonable duration.
		if time.Since(start) > maxBackoff {
			backoff = time.Second
		}

		m.Logger.Warn("session stream disconnected, reconnecting",
			"miraged_id", miragedID, "error", err, "backoff", backoff)

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		backoff = min(backoff*2, maxBackoff)
	}
}

// stream opens a single SSE connection and processes events until it
// closes or the context is cancelled. Returns the reason for disconnection.
func (m *SessionMonitor) stream(ctx context.Context, miragedID string) error {
	client, err := m.Miraged.Client(miragedID)
	if err != nil {
		return err
	}

	ch, cancel, err := client.StreamSessions()
	if err != nil {
		return err
	}
	defer cancel()

	m.Logger.Info("session monitor connected", "miraged_id", miragedID)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-ch:
			if !ok {
				return fmt.Errorf("stream closed by server")
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
