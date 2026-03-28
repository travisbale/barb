package phishing

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// CampaignStatus represents the lifecycle state of a campaign.
type CampaignStatus string

const (
	CampaignDraft     CampaignStatus = "draft"
	CampaignActive    CampaignStatus = "active"
	CampaignPaused    CampaignStatus = "paused"
	CampaignCompleted CampaignStatus = "completed"
)

// Result status constants.
const (
	ResultPending = "pending"
	ResultSent    = "sent"
	ResultFailed  = "failed"
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
}

// Mailer sends a single rendered email. Implementations may reuse
// connections across calls.
type Mailer interface {
	Send(profile *SMTPProfile, tmpl *EmailTemplate, target *Target, lureURL string) error
}

// CampaignService manages campaign lifecycle.
type CampaignService struct {
	Store     campaignStore
	Targets   targetStore
	Templates templateStore
	SMTP      smtpStore
	Mailer    Mailer
	Logger    *slog.Logger
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

	go s.run(context.Background(), campaign)
	return nil
}

// run loads campaign dependencies and sends emails at the configured rate.
func (s *CampaignService) run(ctx context.Context, campaign *Campaign) {
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

	targets := make(map[string]*Target, len(targetList))
	for _, t := range targetList {
		targets[t.ID] = t
	}

	now := time.Now()
	campaign.Status = CampaignActive
	campaign.StartedAt = &now
	if err := s.Store.UpdateCampaign(campaign); err != nil {
		s.Logger.Error("failed to update campaign status", "error", err)
		return
	}

	interval := time.Minute / time.Duration(max(campaign.SendRate, 1))

	for i, result := range results {
		if result.Status != ResultPending {
			continue
		}

		// Throttle between sends (not before the first).
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

		if err := s.Mailer.Send(profile, tmpl, target, campaign.LureURL); err != nil {
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

func (s *CampaignService) Delete(id string) error {
	return s.Store.DeleteCampaign(id)
}

func (s *CampaignService) List() ([]*Campaign, error) {
	return s.Store.ListCampaigns()
}

func (s *CampaignService) Results(campaignID string) ([]*CampaignResult, error) {
	return s.Store.ListResults(campaignID)
}
