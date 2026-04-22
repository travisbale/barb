package phishing

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/travisbale/barb/sdk"
	miragesdk "github.com/travisbale/mirage/sdk"
)

// CampaignStatus represents the lifecycle state of a campaign.
type CampaignStatus string

const (
	CampaignDraft     CampaignStatus = "draft"
	CampaignActive    CampaignStatus = "active"
	CampaignPaused    CampaignStatus = "paused"
	CampaignCompleted CampaignStatus = "completed"
	CampaignCancelled CampaignStatus = "cancelled"
)

// ResultStatus represents the delivery state of a campaign result.
type ResultStatus = string

// Result status constants.
const (
	ResultPending   ResultStatus = "pending"
	ResultSent      ResultStatus = "sent"
	ResultFailed    ResultStatus = "failed"
	ResultClicked   ResultStatus = "clicked"
	ResultCaptured  ResultStatus = "captured"
	ResultCompleted ResultStatus = "completed"
)

// Campaign ties together a target list, email template, and SMTP profile
// into a single phishing operation.
type Campaign struct {
	ID            string
	Name          string
	Status        CampaignStatus
	TemplateID    string
	SMTPProfileID string
	TargetListID  string
	MiragedID     string
	Phishlet      string
	RedirectURL   string
	LureID        string
	LureURL       string
	SendRate      int
	CreatedAt     time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

// CampaignResult tracks the status of a single target within a campaign.
type CampaignResult struct {
	ID         string
	CampaignID string
	TargetID   string
	Email      string
	Status     string
	SentAt     *time.Time
	ClickedAt  *time.Time
	CapturedAt *time.Time
	SessionID  string
}

type campaignStore interface {
	CreateCampaign(c *Campaign) error
	GetCampaign(id string) (*Campaign, error)
	UpdateCampaign(c *Campaign) error
	DeleteCampaign(id string) error
	ListCampaigns() ([]*Campaign, error)
	CreateResults(results []*CampaignResult) error
	UpdateResult(result *CampaignResult) error
	GetResult(id string) (*CampaignResult, error)
	ListResults(campaignID string) ([]*CampaignResult, error)
}

// MailConn represents an open SMTP connection that can send multiple
// messages before being closed.
type MailConn interface {
	Send(profile *SMTPProfile, tmpl *EmailTemplate, target *Target, lureURL string) error
	Close() error
}

// Mailer opens SMTP connections for sending email.
type Mailer interface {
	Dial(ctx context.Context, profile *SMTPProfile) (MailConn, error)
}

// CampaignService manages campaign lifecycle.
type CampaignService struct {
	Store     campaignStore
	Targets   targetStore
	Templates templateStore
	SMTP      smtpStore
	Phishlets phishletStore
	Miraged   *MiragedService
	Monitor   *SessionMonitor
	Mailer    Mailer
	Bus       eventBus
	Logger    *slog.Logger

	running   map[string]context.CancelFunc
	runningMu sync.Mutex
}

func (s *CampaignService) Create(campaign *Campaign) (*Campaign, error) {
	// Verify references exist.
	if _, err := s.Templates.GetTemplate(campaign.TemplateID); err != nil {
		return nil, ErrTemplateNotFound
	}
	if _, err := s.SMTP.GetProfile(campaign.SMTPProfileID); err != nil {
		return nil, ErrSMTPProfileNotFound
	}
	if _, err := s.Targets.GetList(campaign.TargetListID); err != nil {
		return nil, ErrTargetListNotFound
	}

	campaign.ID = uuid.New().String()
	campaign.Status = CampaignDraft
	campaign.CreatedAt = time.Now()
	if campaign.SendRate == 0 {
		campaign.SendRate = 10
	}

	if err := s.Store.CreateCampaign(campaign); err != nil {
		return nil, err
	}

	// Pre-populate results from target list.
	targets, err := s.Targets.ListTargets(campaign.TargetListID)
	if err != nil {
		return nil, err
	}

	results := make([]*CampaignResult, len(targets))
	for i, t := range targets {
		results[i] = &CampaignResult{
			ID:         uuid.New().String(),
			CampaignID: campaign.ID,
			TargetID:   t.ID,
			Email:      t.Email,
			Status:     ResultPending,
		}
	}
	if len(results) > 0 {
		if err := s.Store.CreateResults(results); err != nil {
			return nil, err
		}
	}

	return campaign, nil
}

// Start validates the campaign is in draft status and begins sending
// emails in a background goroutine. It verifies SMTP connectivity before
// launching to surface configuration errors immediately.
func (s *CampaignService) Start(ctx context.Context, id string) error {
	campaign, err := s.Store.GetCampaign(id)
	if err != nil {
		return err
	}
	if campaign.Status != CampaignDraft {
		return ErrCampaignNotDraft
	}

	// Verify the SMTP server is reachable before starting.
	profile, err := s.SMTP.GetProfile(campaign.SMTPProfileID)
	if err != nil {
		return fmt.Errorf("loading SMTP profile: %w", err)
	}
	conn, err := s.Mailer.Dial(ctx, profile)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSMTPConnectionFailed, err)
	}
	_ = conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	s.trackRunning(campaign.ID, cancel)

	go func() {
		defer s.untrackRunning(campaign.ID)
		defer func() {
			if r := recover(); r != nil {
				s.Logger.Error("campaign goroutine panicked", "campaign_id", campaign.ID, "panic", r)
			}
		}()
		s.run(ctx, campaign)
	}()
	return nil
}

// Complete ends a running campaign and sets its status to completed.
func (s *CampaignService) Complete(id string) error {
	return s.endCampaign(id, CampaignCompleted)
}

// Cancel stops a running campaign and sets its status to cancelled.
func (s *CampaignService) Cancel(id string) error {
	return s.endCampaign(id, CampaignCancelled)
}

func (s *CampaignService) endCampaign(id string, status CampaignStatus) error {
	campaign, err := s.Store.GetCampaign(id)
	if err != nil {
		return err
	}
	if campaign.Status != CampaignActive {
		return ErrCampaignNotRunning
	}

	// Stop the goroutine if one is running (may not be after a restart).
	s.runningMu.Lock()
	if cancel, ok := s.running[id]; ok {
		cancel()
	}
	s.runningMu.Unlock()

	now := time.Now()
	campaign.Status = status
	campaign.CompletedAt = &now
	if err := s.Store.UpdateCampaign(campaign); err != nil {
		return err
	}
	s.Bus.Publish(CampaignEvent{
		Type:       sdk.EventCampaignStatus,
		CampaignID: campaign.ID,
		Status:     string(campaign.Status),
	})

	// Clean up miraged resources only on explicit complete/cancel.
	s.cleanup(campaign)

	return nil
}

// Resume restarts session monitors for any campaigns that are still active
// in the database. Called on startup to reconnect after a restart.
func (s *CampaignService) Resume() {
	campaigns, err := s.Store.ListCampaigns()
	if err != nil {
		s.Logger.Error("failed to list campaigns for resume", "error", err)
		return
	}

	for _, campaign := range campaigns {
		if campaign.Status != CampaignActive {
			continue
		}
		if campaign.MiragedID == "" || s.Monitor == nil {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		s.trackRunning(campaign.ID, cancel)

		go func(c *Campaign) {
			defer s.untrackRunning(c.ID)
			s.Logger.Info("resuming session monitor", "campaign_id", c.ID)
			s.Monitor.Watch(ctx, c.MiragedID)
		}(campaign)
	}
}

// Shutdown stops all campaign goroutines without changing campaign status
// or cleaning up miraged resources. Campaigns remain active in the database
// and their lures stay live — the session monitor will reconnect on restart.
func (s *CampaignService) Shutdown() {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	for id, cancel := range s.running {
		cancel()
		s.Logger.Info("stopping campaign goroutine", "campaign_id", id)
	}
	s.running = nil
}

func (s *CampaignService) trackRunning(id string, cancel context.CancelFunc) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	if s.running == nil {
		s.running = make(map[string]context.CancelFunc)
	}
	s.running[id] = cancel
}

func (s *CampaignService) untrackRunning(id string) {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()
	delete(s.running, id)
}

// run orchestrates the campaign: creates the lure, sends emails, and waits
// for the context to be cancelled (by Complete, Cancel, or Shutdown).
func (s *CampaignService) run(ctx context.Context, campaign *Campaign) {
	if err := s.createLure(campaign); err != nil {
		return
	}
	if err := s.activate(campaign); err != nil {
		return
	}

	if campaign.MiragedID != "" && s.Monitor != nil {
		go s.Monitor.Watch(ctx, campaign.MiragedID)
	}

	s.sendEmails(ctx, campaign)

	// Wait for the operator to complete, cancel, or barb to shut down.
	<-ctx.Done()
}

func (s *CampaignService) createLure(campaign *Campaign) error {
	if campaign.MiragedID == "" || campaign.Phishlet == "" || s.Miraged == nil {
		return nil
	}

	// If a lure was already created (e.g. by a test email), reuse it.
	if campaign.LureURL != "" {
		s.Logger.Info("reusing existing lure", "campaign_id", campaign.ID, "lure_url", campaign.LureURL)
		return nil
	}

	return s.ensureLure(campaign)
}

// ensureLure pushes the phishlet and creates a lure on miraged, storing the
// resulting URL on the campaign. The caller is responsible for persisting
// the campaign afterwards.
func (s *CampaignService) ensureLure(campaign *Campaign) error {
	client, err := s.Miraged.Client(campaign.MiragedID)
	if err != nil {
		s.Logger.Error("failed to connect to miraged", "error", err)
		return err
	}

	if s.Phishlets != nil {
		phishlet, err := s.Phishlets.GetPhishletByName(campaign.Phishlet)
		if err == nil {
			if _, err := client.PushPhishlet(miragesdk.PushPhishletRequest{YAML: phishlet.YAML}); err != nil {
				s.Logger.Error("failed to push phishlet to miraged", "phishlet", campaign.Phishlet, "error", err)
				return err
			}
			s.Logger.Info("phishlet pushed to miraged", "phishlet", campaign.Phishlet)
		}
	}
	lure, err := client.CreateLure(miragesdk.CreateLureRequest{
		Phishlet:    campaign.Phishlet,
		RedirectURL: campaign.RedirectURL,
	})
	if err != nil {
		s.Logger.Error("failed to create lure", "error", err)
		return err
	}
	campaign.LureID = lure.ID
	campaign.LureURL = lure.URL
	s.Logger.Info("lure created", "campaign_id", campaign.ID, "lure_url", lure.URL)
	return nil
}

func (s *CampaignService) activate(campaign *Campaign) error {
	now := time.Now()
	campaign.Status = CampaignActive
	campaign.StartedAt = &now
	if err := s.Store.UpdateCampaign(campaign); err != nil {
		s.Logger.Error("failed to activate campaign", "error", err)
		return err
	}
	s.Bus.Publish(CampaignEvent{
		Type:       sdk.EventCampaignStatus,
		CampaignID: campaign.ID,
		Status:     string(campaign.Status),
	})
	return nil
}

func (s *CampaignService) sendEmails(ctx context.Context, campaign *Campaign) {
	profile, err := s.SMTP.GetProfile(campaign.SMTPProfileID)
	if err != nil {
		s.Logger.Error("failed to load SMTP profile", "error", err)
		return
	}
	tmpl, err := s.Templates.GetTemplate(campaign.TemplateID)
	if err != nil {
		s.Logger.Error("failed to load template", "error", err)
		return
	}
	targetList, err := s.Targets.ListTargets(campaign.TargetListID)
	if err != nil {
		s.Logger.Error("failed to load targets", "error", err)
		return
	}
	results, err := s.Store.ListResults(campaign.ID)
	if err != nil {
		s.Logger.Error("failed to load results", "error", err)
		return
	}

	conn, err := s.Mailer.Dial(ctx, profile)
	if err != nil {
		s.Logger.Error("failed to connect to SMTP server", "campaign_id", campaign.ID, "error", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.Logger.Debug("SMTP connection close error", "error", err)
		}
	}()

	targets := make(map[string]*Target, len(targetList))
	for _, t := range targetList {
		targets[t.ID] = t
	}

	// Build a miraged client for per-target URL generation if available.
	var miragedClient *miragesdk.Client
	if campaign.MiragedID != "" && campaign.LureID != "" && s.Miraged != nil {
		mc, err := s.Miraged.Client(campaign.MiragedID)
		if err != nil {
			s.Logger.Error("failed to connect to miraged for URL generation", "error", err)
		} else {
			miragedClient = mc
		}
	}

	interval := time.Minute / time.Duration(max(campaign.SendRate, 1))

	for i, result := range results {
		if result.Status != ResultPending {
			continue
		}

		if i > 0 {
			select {
			case <-ctx.Done():
				s.Logger.Info("campaign sending cancelled", "campaign_id", campaign.ID)
				return
			case <-time.After(interval):
			}
		}

		target := targets[result.TargetID]
		if target == nil {
			continue
		}

		// Generate a per-target URL with an encrypted tracking param.
		lureURL := campaign.LureURL
		if miragedClient != nil {
			resp, err := miragedClient.GenerateLureURL(campaign.LureID, miragesdk.GenerateURLRequest{
				Params: map[string]string{"t": result.ID},
			})
			if err != nil {
				s.Logger.Error("failed to generate tracked URL", "email", target.Email, "error", err)
			} else {
				lureURL = resp.URL
			}
		}

		sentAt := time.Now()
		result.SentAt = &sentAt

		if err := conn.Send(profile, tmpl, target, lureURL); err != nil {
			result.Status = ResultFailed
			s.Logger.Error("failed to send email", "campaign_id", campaign.ID, "email", target.Email, "error", err)
		} else {
			result.Status = ResultSent
			s.Logger.Info("email sent", "campaign_id", campaign.ID, "email", target.Email)
		}

		if err := s.Store.UpdateResult(result); err != nil {
			s.Logger.Error("failed to update result", "error", err)
		}
		s.Bus.Publish(CampaignEvent{
			Type:       sdk.EventResultUpdated,
			CampaignID: campaign.ID,
			Result:     result,
		})
	}
}

// cleanup deletes the lure and disables the phishlet on miraged.
func (s *CampaignService) cleanup(campaign *Campaign) {
	if s.Miraged == nil {
		return
	}

	client, err := s.Miraged.Client(campaign.MiragedID)
	if err != nil {
		s.Logger.Error("failed to connect to miraged for cleanup", "error", err)
		return
	}

	if err := client.DeleteLure(campaign.LureID); err != nil {
		s.Logger.Error("failed to delete lure", "campaign_id", campaign.ID, "error", err)
	} else {
		s.Logger.Info("lure deleted", "campaign_id", campaign.ID)
	}

	if _, err := client.DisablePhishlet(campaign.Phishlet); err != nil {
		s.Logger.Error("failed to disable phishlet", "campaign_id", campaign.ID, "error", err)
	} else {
		s.Logger.Info("phishlet disabled", "campaign_id", campaign.ID, "phishlet", campaign.Phishlet)
	}
}

func (s *CampaignService) Get(id string) (*Campaign, error) {
	return s.Store.GetCampaign(id)
}

// CampaignUpdate holds optional fields for updating a draft campaign.
type CampaignUpdate struct {
	Name          *string
	TemplateID    *string
	SMTPProfileID *string
	TargetListID  *string
	MiragedID     *string
	Phishlet      *string
	RedirectURL   *string
	SendRate      *int
}

func (s *CampaignService) Update(id string, update *CampaignUpdate) (*Campaign, error) {
	campaign, err := s.Store.GetCampaign(id)
	if err != nil {
		return nil, err
	}
	if campaign.Status != CampaignDraft {
		return nil, ErrCampaignNotDraft
	}

	if update.Name != nil {
		campaign.Name = *update.Name
	}
	if update.TemplateID != nil {
		if _, err := s.Templates.GetTemplate(*update.TemplateID); err != nil {
			return nil, ErrTemplateNotFound
		}
		campaign.TemplateID = *update.TemplateID
	}
	if update.SMTPProfileID != nil {
		if _, err := s.SMTP.GetProfile(*update.SMTPProfileID); err != nil {
			return nil, ErrSMTPProfileNotFound
		}
		campaign.SMTPProfileID = *update.SMTPProfileID
	}
	if update.TargetListID != nil {
		if _, err := s.Targets.GetList(*update.TargetListID); err != nil {
			return nil, ErrTargetListNotFound
		}
		campaign.TargetListID = *update.TargetListID
	}
	if update.MiragedID != nil {
		campaign.MiragedID = *update.MiragedID
	}
	if update.Phishlet != nil {
		campaign.Phishlet = *update.Phishlet
	}
	if update.RedirectURL != nil {
		campaign.RedirectURL = *update.RedirectURL
	}
	if update.SendRate != nil {
		campaign.SendRate = *update.SendRate
	}

	if err := s.Store.UpdateCampaign(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

func (s *CampaignService) Delete(id string) error {
	campaign, err := s.Store.GetCampaign(id)
	if err != nil {
		return err
	}
	if campaign.Status == CampaignActive {
		return ErrCampaignActive
	}
	return s.Store.DeleteCampaign(id)
}

func (s *CampaignService) List() ([]*Campaign, error) {
	return s.Store.ListCampaigns()
}

func (s *CampaignService) Results(campaignID string) ([]*CampaignResult, error) {
	return s.Store.ListResults(campaignID)
}

// Stream returns a channel that delivers the campaign's current state
// (status + results that have moved past pending) followed by every
// subsequent CampaignEvent published for the campaign. The channel is
// closed and the bus subscription released when ctx is cancelled.
//
// Returns ErrNotFound if the campaign does not exist, without starting
// the stream.
func (s *CampaignService) Stream(ctx context.Context, campaignID string) (<-chan CampaignEvent, error) {
	campaign, err := s.Store.GetCampaign(campaignID)
	if err != nil {
		return nil, err
	}

	chEvents := make(chan CampaignEvent, 64)
	go func() {
		defer close(chEvents)

		// Subscribe before taking the snapshot so events that fire mid-snapshot
		// queue in the bus channel rather than being lost.
		ch := s.Bus.Subscribe(campaignID)
		defer s.Bus.Unsubscribe(campaignID, ch)

		s.emitSnapshot(ctx, campaign, chEvents)

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-ch:
				if !sendOrCancel(ctx, chEvents, event) {
					return
				}
			}
		}
	}()
	return chEvents, nil
}

// emitSnapshot sends the campaign's current state to out: the current status
// and every result whose delivery has progressed past pending. Pending
// results are skipped because they have no state worth broadcasting yet —
// the EventResultUpdated fired on their first transition will deliver them
// to subscribers via the live stream.
func (s *CampaignService) emitSnapshot(ctx context.Context, campaign *Campaign, out chan<- CampaignEvent) {
	if !sendOrCancel(ctx, out, CampaignEvent{
		Type:       sdk.EventCampaignStatus,
		CampaignID: campaign.ID,
		Status:     string(campaign.Status),
	}) {
		return
	}

	results, err := s.Store.ListResults(campaign.ID)
	if err != nil {
		s.Logger.Warn("stream: list results failed", "campaign_id", campaign.ID, "error", err)
		return
	}
	for _, result := range results {
		if result.Status == ResultPending {
			continue
		}
		if !sendOrCancel(ctx, out, CampaignEvent{
			Type:       sdk.EventResultUpdated,
			CampaignID: campaign.ID,
			Result:     result,
		}) {
			return
		}
	}
}

// sendOrCancel attempts to send event on out, returning false if ctx is
// cancelled before the send succeeds.
func sendOrCancel(ctx context.Context, out chan<- CampaignEvent, event CampaignEvent) bool {
	select {
	case out <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// SendTestEmail sends a single test email using a campaign's configuration.
// If miraged is configured and no lure exists yet, one is created and
// persisted on the campaign so it can be reused by subsequent tests and
// the actual campaign run.
func (s *CampaignService) SendTestEmail(campaignID, email string) error {
	if email == "" {
		return ErrEmailRequired
	}

	campaign, err := s.Store.GetCampaign(campaignID)
	if err != nil {
		return err
	}

	profile, err := s.SMTP.GetProfile(campaign.SMTPProfileID)
	if err != nil {
		return fmt.Errorf("loading SMTP profile: %w", err)
	}
	tmpl, err := s.Templates.GetTemplate(campaign.TemplateID)
	if err != nil {
		return fmt.Errorf("loading template: %w", err)
	}

	// Create a persistent lure if miraged is configured and one doesn't exist yet.
	lureURL := campaign.LureURL
	if lureURL == "" && campaign.MiragedID != "" && campaign.Phishlet != "" && s.Miraged != nil {
		if err := s.ensureLure(campaign); err != nil {
			return fmt.Errorf("creating lure: %w", err)
		}
		lureURL = campaign.LureURL
		if err := s.Store.UpdateCampaign(campaign); err != nil {
			return fmt.Errorf("saving lure URL: %w", err)
		}
	}
	if lureURL == "" {
		lureURL = "https://example.com/test-lure"
	}

	conn, err := s.Mailer.Dial(context.Background(), profile)
	if err != nil {
		return fmt.Errorf("connecting to SMTP: %w", err)
	}
	defer conn.Close()

	target := &Target{Email: email, FirstName: "Test", LastName: "User"}
	return conn.Send(profile, tmpl, target, lureURL)
}
