package phishing

import (
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

// Campaign ties together a target list, email template, and SMTP profile
// into a single phishing operation.
type Campaign struct {
	ID            string
	Name          string
	Status        CampaignStatus
	TemplateID    string
	SMTPProfileID string
	TargetListID  string
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
	ListResults(campaignID string) ([]*CampaignResult, error)
}

// CampaignService manages campaign lifecycle.
type CampaignService struct {
	Store     campaignStore
	Targets   targetStore
	Templates templateStore
	SMTP      smtpStore
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
			Status:     "pending",
		}
	}
	if len(results) > 0 {
		if err := s.Store.CreateResults(results); err != nil {
			return nil, err
		}
	}

	return campaign, nil
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
