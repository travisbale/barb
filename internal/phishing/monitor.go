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
	Bus       eventBus
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
	session := event.Session

	resultID, ok := session.LureParams["t"]
	if !ok {
		return
	}

	m.correlateByToken(resultID, event.Type, session)
}

// correlateByToken uses the tracking token (result ID) from the lure URL
// to directly look up and update the campaign result.
func (m *SessionMonitor) correlateByToken(resultID string, eventType miragesdk.EventType, session miragesdk.SessionResponse) {
	result, err := m.Campaigns.GetResult(resultID)
	if err != nil {
		m.Logger.Debug("no result for tracking token", "token", resultID)
		return
	}

	if !m.updateResult(result, eventType, session) {
		return
	}

	if err := m.Campaigns.UpdateResult(result); err != nil {
		m.Logger.Error("failed to update result", "result_id", resultID, "error", err)
		return
	}

	m.Bus.Publish(newResultEvent(result))
	m.Logger.Info("session correlated",
		"campaign_id", result.CampaignID,
		"event", string(eventType),
		"email", result.Email,
		"session_id", session.ID,
	)
}

// updateResult applies the event to the result. Returns false if no update is needed.
func (m *SessionMonitor) updateResult(result *CampaignResult, eventType miragesdk.EventType, session miragesdk.SessionResponse) bool {
	switch eventType {
	case miragesdk.EventSessionCreated:
		if result.ClickedAt != nil {
			return false
		}
		clickedAt := session.StartedAt
		result.ClickedAt = &clickedAt
		result.Status = ResultClicked
		result.SessionID = session.ID

	case miragesdk.EventCredsCaptured:
		if result.Status == ResultCaptured || result.Status == ResultCompleted {
			return false
		}
		if result.ClickedAt == nil {
			clickedAt := session.StartedAt
			result.ClickedAt = &clickedAt
		}
		now := time.Now()
		result.Status = ResultCaptured
		result.CapturedAt = &now
		result.SessionID = session.ID

	case miragesdk.EventSessionCompleted:
		if result.Status == ResultCompleted {
			return false
		}
		result.Status = ResultCompleted

	default:
		return false
	}
	return true
}
