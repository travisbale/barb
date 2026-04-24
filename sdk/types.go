package sdk

import "time"

// --- Target Lists ---

type CreateTargetListRequest struct {
	Name string `json:"name"`
}

type UpdateTargetListRequest struct {
	Name *string `json:"name,omitempty"`
}

type TargetListResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type AddTargetRequest struct {
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

type TargetResponse struct {
	ID         string `json:"id"`
	ListID     string `json:"list_id"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

// --- SMTP Profiles ---

type CreateSMTPProfileRequest struct {
	Name          string            `json:"name"`
	Host          string            `json:"host"`
	Port          int               `json:"port"`
	Username      string            `json:"username"`
	Password      string            `json:"password"`
	FromAddr      string            `json:"from_addr"`
	FromName      string            `json:"from_name"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// UpdateSMTPProfileRequest supports partial updates. Only non-nil fields
// are applied; nil fields leave the existing value unchanged.
type UpdateSMTPProfileRequest struct {
	Name          *string            `json:"name,omitempty"`
	Host          *string            `json:"host,omitempty"`
	Port          *int               `json:"port,omitempty"`
	Username      *string            `json:"username,omitempty"`
	Password      *string            `json:"password,omitempty"`
	FromAddr      *string            `json:"from_addr,omitempty"`
	FromName      *string            `json:"from_name,omitempty"`
	CustomHeaders *map[string]string `json:"custom_headers,omitempty"`
}

type SMTPProfileResponse struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Host          string            `json:"host"`
	Port          int               `json:"port"`
	Username      string            `json:"username"`
	FromAddr      string            `json:"from_addr"`
	FromName      string            `json:"from_name"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

// --- Email Templates ---

type CreateTemplateRequest struct {
	Name           string `json:"name"`
	Subject        string `json:"subject"`
	HTMLBody       string `json:"html_body"`
	TextBody       string `json:"text_body"`
	EnvelopeSender string `json:"envelope_sender,omitempty"`
}

// UpdateTemplateRequest supports partial updates. Only non-nil fields
// are applied; nil fields leave the existing value unchanged.
type UpdateTemplateRequest struct {
	Name           *string `json:"name,omitempty"`
	Subject        *string `json:"subject,omitempty"`
	HTMLBody       *string `json:"html_body,omitempty"`
	TextBody       *string `json:"text_body,omitempty"`
	EnvelopeSender *string `json:"envelope_sender,omitempty"`
}

type TemplateResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Subject        string    `json:"subject"`
	HTMLBody       string    `json:"html_body"`
	TextBody       string    `json:"text_body"`
	EnvelopeSender string    `json:"envelope_sender,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PreviewTemplateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	URL       string `json:"url"`
}

type PreviewTemplateResponse struct {
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

type RenderHTMLRequest struct {
	HTMLBody string `json:"html_body"`
	PreviewTemplateRequest
}

type RenderHTMLResponse struct {
	HTMLBody string `json:"html_body"`
}

type SendTestEmailRequest struct {
	Email string `json:"email"`
}

// --- Phishlets ---

type CreatePhishletRequest struct {
	YAML string `json:"yaml"`
}

type UpdatePhishletRequest struct {
	YAML string `json:"yaml"`
}

type PhishletResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	YAML      string    `json:"yaml"`
	CreatedAt time.Time `json:"created_at"`
}

type ImportTargetsResponse struct {
	Imported int `json:"imported"`
}

// --- Campaigns ---

type CreateCampaignRequest struct {
	Name          string `json:"name"`
	TemplateID    string `json:"template_id"`
	SMTPProfileID string `json:"smtp_profile_id"`
	TargetListID  string `json:"target_list_id"`
	MiragedID     string `json:"miraged_id"`
	Phishlet      string `json:"phishlet"`
	RedirectURL   string `json:"redirect_url"`
	LureURL       string `json:"lure_url"`
	SendRate      int    `json:"send_rate"`
}

type UpdateCampaignRequest struct {
	Name          *string `json:"name,omitempty"`
	TemplateID    *string `json:"template_id,omitempty"`
	SMTPProfileID *string `json:"smtp_profile_id,omitempty"`
	TargetListID  *string `json:"target_list_id,omitempty"`
	MiragedID     *string `json:"miraged_id,omitempty"`
	Phishlet      *string `json:"phishlet,omitempty"`
	RedirectURL   *string `json:"redirect_url,omitempty"`
	SendRate      *int    `json:"send_rate,omitempty"`
}

type CampaignResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	TemplateID    string     `json:"template_id"`
	SMTPProfileID string     `json:"smtp_profile_id"`
	TargetListID  string     `json:"target_list_id"`
	MiragedID     string     `json:"miraged_id"`
	Phishlet      string     `json:"phishlet"`
	RedirectURL   string     `json:"redirect_url"`
	LureURL       string     `json:"lure_url"`
	SendRate      int        `json:"send_rate"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type CampaignResultResponse struct {
	ID         string     `json:"id"`
	CampaignID string     `json:"campaign_id"`
	TargetID   string     `json:"target_id"`
	Email      string     `json:"email"`
	Status     string     `json:"status"`
	SentAt     *time.Time `json:"sent_at"`
	ClickedAt  *time.Time `json:"clicked_at"`
	CapturedAt *time.Time `json:"captured_at"`
	SessionID  string     `json:"session_id,omitempty"`
}

// --- Campaign Events (SSE) ---

// CampaignEventType enumerates the campaign event types
type CampaignEventType string

const (
	EventResultUpdated  CampaignEventType = "result.updated"
	EventCampaignStatus CampaignEventType = "campaign.status"
)

// CampaignEvent is delivered by StreamCampaign for result updates and status changes.
type CampaignEvent struct {
	Type       CampaignEventType       `json:"type"`
	CampaignID string                  `json:"campaign_id"`
	Result     *CampaignResultResponse `json:"result,omitempty"`
	Status     string                  `json:"status,omitempty"`
}

// --- Miraged Connections ---

type EnrollMiragedRequest struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	SecretHostname string `json:"secret_hostname"`
	Token          string `json:"token"`
}

type UpdateMiragedRequest struct {
	Name *string `json:"name,omitempty"`
}

type MiragedResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	SecretHostname string    `json:"secret_hostname"`
	CreatedAt      time.Time `json:"created_at"`
}

type MiragedStatusResponse struct {
	Connected bool   `json:"connected"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
}

type MiragedPhishletResponse struct {
	Name        string `json:"name"`
	Hostname    string `json:"hostname"`
	BaseDomain  string `json:"base_domain"`
	DNSProvider string `json:"dns_provider"`
	SpoofURL    string `json:"spoof_url"`
	Enabled     bool   `json:"enabled"`
}

type PushMiragedPhishletRequest struct {
	YAML string `json:"yaml"`
}

type EnableMiragedPhishletRequest struct {
	Hostname    string `json:"hostname"`
	DNSProvider string `json:"dns_provider"`
}

type MiragedNotificationChannelResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	URL       string    `json:"url"`
	Filter    []string  `json:"filter"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateMiragedNotificationChannelRequest struct {
	Type       string   `json:"type"`
	URL        string   `json:"url"`
	AuthHeader string   `json:"auth_header,omitempty"`
	Filter     []string `json:"filter,omitempty"`
}

// --- Auth ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MeResponse struct {
	Username               string `json:"username"`
	PasswordChangeRequired bool   `json:"password_change_required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// --- Sessions ---

type MiragedSessionResponse struct {
	ID           string                       `json:"id"`
	Phishlet     string                       `json:"phishlet"`
	RemoteAddr   string                       `json:"remote_addr"`
	UserAgent    string                       `json:"user_agent"`
	Username     string                       `json:"username"`
	Password     string                       `json:"password"`
	Custom       map[string]string            `json:"custom,omitempty"`
	CookieTokens map[string]map[string]string `json:"cookie_tokens,omitempty"`
	BodyTokens   map[string]string            `json:"body_tokens,omitempty"`
	HTTPTokens   map[string]string            `json:"http_tokens,omitempty"`
	StartedAt    string                       `json:"started_at"`
	CompletedAt  string                       `json:"completed_at,omitempty"`
}

// --- Dashboard ---

type DashboardResponse struct {
	Campaigns        CampaignCounts       `json:"campaigns"`
	TotalCompletions int                  `json:"total_completions"`
	TotalClicks      int                  `json:"total_clicks"`
	TotalEmailsSent  int                  `json:"total_emails_sent"`
	MiragedCount     int                  `json:"miraged_count"`
	ActiveCampaigns  []ActiveCampaignInfo `json:"active_campaigns"`
	RecentCaptures   []RecentCapture      `json:"recent_captures"`
}

type CampaignCounts struct {
	Draft     int `json:"draft"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Cancelled int `json:"cancelled"`
	Total     int `json:"total"`
}

type RecentCapture struct {
	Email        string `json:"email"`
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	CapturedAt   string `json:"captured_at"`
	SessionID    string `json:"session_id"`
}

type ActiveCampaignInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Sent      int    `json:"sent"`
	Failed    int    `json:"failed"`
	Captured  int    `json:"captured"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
}

// --- System ---

type StatusResponse struct {
	Version string `json:"version"`
}

// --- Errors ---

type ErrorResponse struct {
	Error string `json:"error"`
}
