package sdk

import "time"

// --- Target Lists ---

type CreateTargetListRequest struct {
	Name string `json:"name"`
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
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	FromAddr string `json:"from_addr"`
	FromName string `json:"from_name"`
}

type SMTPProfileResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	FromAddr  string    `json:"from_addr"`
	FromName  string    `json:"from_name"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Email Templates ---

type CreateTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

type UpdateTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

type TemplateResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	HTMLBody  string    `json:"html_body"`
	TextBody  string    `json:"text_body"`
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
	LureURL       string `json:"lure_url"`
	SendRate      int    `json:"send_rate"`
}

type CampaignResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	TemplateID    string     `json:"template_id"`
	SMTPProfileID string     `json:"smtp_profile_id"`
	TargetListID  string     `json:"target_list_id"`
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

// --- System ---

type StatusResponse struct {
	Version string `json:"version"`
}

// --- Errors ---

type ErrorResponse struct {
	Error string `json:"error"`
}
