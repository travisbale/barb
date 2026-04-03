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
	m.Logger.Debug("opening session stream", "miraged_id", miragedID)
	client, err := m.Miraged.Client(miragedID)
	if err != nil {
		return err
	}

	m.Logger.Debug("connecting to session stream", "miraged_id", miragedID)
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
			m.Logger.Debug("session event received",
				"miraged_id", miragedID,
				"type", event.Type,
				"session_id", event.Session.ID,
				"username", event.Session.Username,
				"phishlet", event.Session.Phishlet,
			)
			m.handleEvent(miragedID, event)
		}
	}
}

func (m *SessionMonitor) handleEvent(miragedID string, event miragesdk.SessionEvent) {
	switch event.Type {
	case miragesdk.EventCredsCaptured, miragesdk.EventSessionCompleted:
	default:
		m.Logger.Debug("ignoring non-capture event", "type", event.Type)
		return
	}

	session := event.Session
	if session.Username == "" {
		m.Logger.Debug("ignoring event with empty username", "session_id", session.ID)
		return
	}

	campaigns, err := m.Campaigns.ListActiveCampaignsByMiraged(miragedID)
	if err != nil {
		m.Logger.Error("failed to list campaigns for correlation", "error", err)
		return
	}
	m.Logger.Debug("correlating session against campaigns",
		"session_id", session.ID, "username", session.Username,
		"phishlet", session.Phishlet, "campaign_count", len(campaigns),
	)

	for _, campaign := range campaigns {
		if campaign.Phishlet != "" && campaign.Phishlet != session.Phishlet {
			m.Logger.Debug("skipping campaign — phishlet mismatch",
				"campaign_id", campaign.ID, "campaign_phishlet", campaign.Phishlet,
				"session_phishlet", session.Phishlet,
			)
			continue
		}
		m.correlate(campaign, session)
	}
}

func (m *SessionMonitor) correlate(campaign *Campaign, session miragesdk.SessionResponse) {
	result, err := m.Campaigns.GetResultByEmail(campaign.ID, session.Username)
	if err != nil {
		m.Logger.Debug("no matching result for session",
			"campaign_id", campaign.ID, "username", session.Username, "error", err,
		)
		return
	}
	if result.SessionID != "" {
		m.Logger.Debug("session already correlated",
			"campaign_id", campaign.ID, "result_id", result.ID, "session_id", result.SessionID,
		)
		return
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
