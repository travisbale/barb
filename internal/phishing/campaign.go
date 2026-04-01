package phishing

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
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
	ResultPending  ResultStatus = "pending"
	ResultSent     ResultStatus = "sent"
	ResultFailed   ResultStatus = "failed"
	ResultClicked  ResultStatus = "clicked"
	ResultCaptured ResultStatus = "captured"
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
	LureURL       string
	SendRate      int
	CreatedAt     time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

func (c *Campaign) Validate() error {
	if c.Name == "" {
		return ErrNameRequired
	}
	if c.TemplateID == "" {
		return ErrTemplateRequired
	}
	if c.SMTPProfileID == "" {
		return ErrSMTPProfileRequired
	}
	if c.TargetListID == "" {
		return ErrTargetListRequired
	}
	return nil
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
	ListResults(campaignID string) ([]*CampaignResult, error)
	ListActiveCampaignsByMiraged(miragedID string) ([]*Campaign, error)
	GetResultByEmail(campaignID, email string) (*CampaignResult, error)
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
	Logger    *slog.Logger

	running   map[string]context.CancelFunc
	runningMu sync.Mutex
}

func (s *CampaignService) Create(campaign *Campaign) (*Campaign, error) {
	if err := campaign.Validate(); err != nil {
		return nil, err
	}

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
// emails in a background goroutine.
func (s *CampaignService) Start(id string) error {
	campaign, err := s.Store.GetCampaign(id)
	if err != nil {
		return err
	}
	if campaign.Status != CampaignDraft {
		return ErrCampaignNotDraft
	}

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

// Cancel stops a running campaign and sets its status to cancelled.
func (s *CampaignService) Cancel(id string) error {
	s.runningMu.Lock()
	cancel, ok := s.running[id]
	s.runningMu.Unlock()

	if !ok {
		return ErrCampaignNotRunning
	}

	cancel()

	campaign, err := s.Store.GetCampaign(id)
	if err != nil {
		return err
	}
	campaign.Status = CampaignCancelled
	return s.Store.UpdateCampaign(campaign)
}

// Shutdown cancels all running campaigns.
func (s *CampaignService) Shutdown() {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	for id, cancel := range s.running {
		cancel()
		s.Logger.Info("cancelled running campaign", "campaign_id", id)
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

// run orchestrates the campaign: creates the lure, sends emails, and marks complete.
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

	// Only mark complete if the campaign wasn't cancelled.
	if ctx.Err() == nil {
		s.complete(campaign)
	}
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
	if s.Phishlets != nil {
		phishlet, err := s.Phishlets.GetPhishletByName(campaign.Phishlet)
		if err == nil {
			if err := s.Miraged.PushPhishlet(campaign.MiragedID, phishlet.YAML); err != nil {
				s.Logger.Error("failed to push phishlet to miraged", "phishlet", campaign.Phishlet, "error", err)
				return err
			}
			s.Logger.Info("phishlet pushed to miraged", "phishlet", campaign.Phishlet)
		}
	}

	client, err := s.Miraged.client(campaign.MiragedID)
	if err != nil {
		s.Logger.Error("failed to connect to miraged", "error", err)
		return err
	}
	lure, err := client.CreateLure(miragesdk.CreateLureRequest{
		Phishlet:    campaign.Phishlet,
		RedirectURL: campaign.RedirectURL,
	})
	if err != nil {
		s.Logger.Error("failed to create lure", "error", err)
		return err
	}
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

		sentAt := time.Now()
		result.SentAt = &sentAt

		if err := conn.Send(profile, tmpl, target, campaign.LureURL); err != nil {
			result.Status = ResultFailed
			s.Logger.Error("failed to send email", "campaign_id", campaign.ID, "email", target.Email, "error", err)
		} else {
			result.Status = ResultSent
			s.Logger.Info("email sent", "campaign_id", campaign.ID, "email", target.Email)
		}

		if err := s.Store.UpdateResult(result); err != nil {
			s.Logger.Error("failed to update result", "error", err)
		}
	}
}

func (s *CampaignService) complete(campaign *Campaign) {
	completedAt := time.Now()
	campaign.Status = CampaignCompleted
	campaign.CompletedAt = &completedAt
	if err := s.Store.UpdateCampaign(campaign); err != nil {
		s.Logger.Error("failed to mark campaign completed", "error", err)
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

	if err := campaign.Validate(); err != nil {
		return nil, err
	}

	if err := s.Store.UpdateCampaign(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

func (s *CampaignService) Delete(id string) error {
	return s.Store.DeleteCampaign(id)
}

func (s *CampaignService) List() ([]*Campaign, error) {
	return s.Store.ListCampaigns()
}

func (s *CampaignService) Results(campaignID string) ([]*CampaignResult, error) {
	return s.Store.ListResults(campaignID)
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
