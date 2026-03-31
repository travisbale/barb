//go:build !unit

package test_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/travisbale/barb/internal/delivery"
	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

// mailpitMessage represents a message summary from Mailpit's list API.
type mailpitMessage struct {
	ID   string `json:"ID"`
	From struct {
		Address string `json:"Address"`
		Name    string `json:"Name"`
	} `json:"From"`
	To []struct {
		Address string `json:"Address"`
	} `json:"To"`
	Subject    string `json:"Subject"`
	Snippet    string `json:"Snippet"`
	ReturnPath string `json:"ReturnPath"`
}

// mailpitMessageDetail represents a full message from Mailpit's detail API.
type mailpitMessageDetail struct {
	ID   string `json:"ID"`
	From struct {
		Address string `json:"Address"`
		Name    string `json:"Name"`
	} `json:"From"`
	To []struct {
		Address string `json:"Address"`
	} `json:"To"`
	Subject    string              `json:"Subject"`
	HTML       string              `json:"HTML"`
	Text       string              `json:"Text"`
	ReturnPath string              `json:"ReturnPath"`
	Headers    map[string][]string `json:"-"` // custom parsing needed
}

type mailpitList struct {
	Messages []mailpitMessage `json:"messages"`
	Total    int              `json:"messages_count"`
}

func startMailpit(t *testing.T) (smtpHost string, smtpPort int, apiURL string) {
	t.Helper()
	ctx := context.Background()

	// Disable Ryuk reaper — we clean up via t.Cleanup.
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	req := testcontainers.ContainerRequest{
		Image:        "axllent/mailpit:latest",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForHTTP("/api/v1/messages").WithPort("8025/tcp").WithStartupTimeout(30 * time.Second),
		Env: map[string]string{
			"MP_SMTP_AUTH_ACCEPT_ANY":     "1",
			"MP_SMTP_AUTH_ALLOW_INSECURE": "1",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("skipping: could not start mailpit container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("getting container host: %v", err)
	}

	smtpMapped, err := container.MappedPort(ctx, "1025/tcp")
	if err != nil {
		t.Fatalf("getting SMTP port: %v", err)
	}

	apiMapped, err := container.MappedPort(ctx, "8025/tcp")
	if err != nil {
		t.Fatalf("getting API port: %v", err)
	}

	return host, smtpMapped.Int(), fmt.Sprintf("http://%s:%d", host, apiMapped.Int())
}

func getMailpitMessages(t *testing.T, apiURL string) mailpitList {
	t.Helper()
	resp, err := http.Get(apiURL + "/api/v1/messages")
	if err != nil {
		t.Fatalf("querying mailpit: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result mailpitList
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("parsing mailpit response: %v (body: %s)", err, body)
	}
	return result
}

func getMailpitMessage(t *testing.T, apiURL, id string) mailpitMessageDetail {
	t.Helper()
	resp, err := http.Get(apiURL + "/api/v1/message/" + id)
	if err != nil {
		t.Fatalf("querying mailpit message: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result mailpitMessageDetail
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("parsing mailpit message: %v (body: %s)", err, body)
	}
	return result
}

func getMailpitMessageHeaders(t *testing.T, apiURL, id string) map[string][]string {
	t.Helper()
	resp, err := http.Get(apiURL + "/api/v1/message/" + id + "/headers")
	if err != nil {
		t.Fatalf("querying mailpit headers: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string][]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("parsing mailpit headers: %v (body: %s)", err, body)
	}
	return result
}

func TestIntegration_SendTestEmail(t *testing.T) {
	smtpHost, smtpPort, mailpitAPI := startMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	// Create SMTP profile pointing to Mailpit.
	smtp, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Mailpit",
		Host:     smtpHost,
		Port:     smtpPort,
		FromAddr: "phisher@example.com",
		FromName: "IT Support",
	})
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}

	// Create a template.
	tmpl, err := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Test Email Template",
		Subject:  "Hello {{.FirstName}}",
		HTMLBody: "<p>Dear {{.FirstName}}, click <a href=\"{{.URL}}\">here</a>.</p>",
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}

	// Create a target list.
	list, err := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Test Targets"})
	if err != nil {
		t.Fatalf("CreateTargetList: %v", err)
	}
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "victim@example.com", FirstName: "Alice"})

	// Create a draft campaign.
	campaign, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name:          "Integration Test",
		TemplateID:    tmpl.ID,
		SMTPProfileID: smtp.ID,
		TargetListID:  list.ID,
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	// Send a test email.
	err = h.Client.SendTestEmail(campaign.ID, sdk.SendTestEmailRequest{Email: "operator@example.com"})
	if err != nil {
		t.Fatalf("SendTestEmail: %v", err)
	}

	// Verify Mailpit received the email.
	time.Sleep(500 * time.Millisecond)
	messages := getMailpitMessages(t, mailpitAPI)
	if messages.Total != 1 {
		t.Fatalf("expected 1 message in mailpit, got %d", messages.Total)
	}

	msg := messages.Messages[0]
	if msg.Subject != "Hello Test" {
		t.Errorf("Subject = %q, want %q", msg.Subject, "Hello Test")
	}
	if msg.From.Address != "phisher@example.com" {
		t.Errorf("From = %q, want %q", msg.From.Address, "phisher@example.com")
	}
	if msg.From.Name != "IT Support" {
		t.Errorf("FromName = %q, want %q", msg.From.Name, "IT Support")
	}
	if len(msg.To) == 0 || msg.To[0].Address != "operator@example.com" {
		t.Errorf("To = %v, want operator@example.com", msg.To)
	}
}

func TestIntegration_CampaignSendsEmails(t *testing.T) {
	smtpHost, smtpPort, mailpitAPI := startMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Mailpit", Host: smtpHost, Port: smtpPort, FromAddr: "noreply@example.com",
	})
	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name: "Campaign Template", Subject: "Important: {{.FirstName}}", HTMLBody: "<p>Click {{.URL}}</p>",
	})
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Campaign Targets"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "alice@example.com", FirstName: "Alice"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "bob@example.com", FirstName: "Bob"})

	campaign, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Send Test", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID, SendRate: 600,
	})

	if err := h.Client.StartCampaign(campaign.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	// Wait for sending to complete.
	time.Sleep(2 * time.Second)

	messages := getMailpitMessages(t, mailpitAPI)
	if messages.Total != 2 {
		t.Fatalf("expected 2 messages in mailpit, got %d", messages.Total)
	}

	// Verify subjects were rendered per-target.
	subjects := map[string]bool{}
	for _, msg := range messages.Messages {
		subjects[msg.Subject] = true
	}
	if !subjects["Important: Alice"] {
		t.Error("missing email for Alice")
	}
	if !subjects["Important: Bob"] {
		t.Error("missing email for Bob")
	}
}

func TestIntegration_CustomHeaders(t *testing.T) {
	smtpHost, smtpPort, mailpitAPI := startMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Mailpit Headers", Host: smtpHost, Port: smtpPort, FromAddr: "test@example.com",
		CustomHeaders: map[string]string{
			"X-Mailer":     "Outlook 16.0",
			"X-Custom-Tag": "phishing-test",
		},
	})
	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name: "Headers Template", Subject: "Test Headers", HTMLBody: "<p>test</p>",
	})
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Header Targets"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "target@example.com"})

	campaign, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Headers Test", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID, SendRate: 600,
	})

	h.Client.StartCampaign(campaign.ID)
	time.Sleep(2 * time.Second)

	messages := getMailpitMessages(t, mailpitAPI)
	if messages.Total != 1 {
		t.Fatalf("expected 1 message, got %d", messages.Total)
	}

	// Verify custom headers via Mailpit's headers API.
	headers := getMailpitMessageHeaders(t, mailpitAPI, messages.Messages[0].ID)
	if v := headers["X-Mailer"]; len(v) == 0 || v[0] != "Outlook 16.0" {
		t.Errorf("X-Mailer = %v, want [Outlook 16.0]", v)
	}
	if v := headers["X-Custom-Tag"]; len(v) == 0 || v[0] != "phishing-test" {
		t.Errorf("X-Custom-Tag = %v, want [phishing-test]", v)
	}
}

func TestIntegration_EnvelopeSender(t *testing.T) {
	smtpHost, smtpPort, mailpitAPI := startMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Mailpit", Host: smtpHost, Port: smtpPort, FromAddr: "visible@example.com",
	})
	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:           "Envelope Test",
		Subject:        "Envelope Sender Test",
		HTMLBody:       "<p>test</p>",
		EnvelopeSender: "bounce@attacker.com",
	})
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Envelope Targets"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "target@example.com"})

	campaign, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Envelope Test", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID, SendRate: 600,
	})
	h.Client.StartCampaign(campaign.ID)
	time.Sleep(2 * time.Second)

	messages := getMailpitMessages(t, mailpitAPI)
	if messages.Total != 1 {
		t.Fatalf("expected 1 message, got %d", messages.Total)
	}

	// From header should be the SMTP profile's address.
	if messages.Messages[0].From.Address != "visible@example.com" {
		t.Errorf("From = %q, want visible@example.com", messages.Messages[0].From.Address)
	}

	// Return-Path should be the envelope sender.
	detail := getMailpitMessage(t, mailpitAPI, messages.Messages[0].ID)
	if detail.ReturnPath != "bounce@attacker.com" {
		t.Errorf("ReturnPath = %q, want bounce@attacker.com", detail.ReturnPath)
	}
}

func TestIntegration_CampaignCancelStopsSending(t *testing.T) {
	smtpHost, smtpPort, mailpitAPI := startMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Mailpit", Host: smtpHost, Port: smtpPort, FromAddr: "test@example.com",
	})
	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name: "Cancel Template", Subject: "Cancel Test", HTMLBody: "<p>test</p>",
	})
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Cancel Targets"})
	for i := 0; i < 20; i++ {
		h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: fmt.Sprintf("user%d@example.com", i)})
	}

	campaign, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Cancel Test", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID,
		SendRate: 1, // 1 per minute — very slow
	})

	h.Client.StartCampaign(campaign.ID)
	time.Sleep(500 * time.Millisecond)
	h.Client.CancelCampaign(campaign.ID)
	time.Sleep(500 * time.Millisecond)

	messages := getMailpitMessages(t, mailpitAPI)
	if messages.Total >= 20 {
		t.Errorf("expected fewer than 20 emails (campaign cancelled), got %d", messages.Total)
	}
	if messages.Total == 0 {
		t.Error("expected at least 1 email to be sent before cancel")
	}
}

func TestIntegration_TemplateVariablesRendered(t *testing.T) {
	smtpHost, smtpPort, mailpitAPI := startMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Mailpit", Host: smtpHost, Port: smtpPort, FromAddr: "test@example.com",
	})
	tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
		Name:     "Variable Template",
		Subject:  "Hello {{.FirstName}} {{.LastName}}",
		HTMLBody: "<p>Dear {{.FirstName}}, please visit <a href=\"{{.URL}}\">this link</a>.</p>",
		TextBody: "Dear {{.FirstName}} {{.LastName}}, visit {{.URL}}",
	})
	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Variable Targets"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{
		Email: "alice@example.com", FirstName: "Alice", LastName: "Smith",
	})

	campaign, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
		Name: "Variables Test", TemplateID: tmpl.ID, SMTPProfileID: smtp.ID, TargetListID: list.ID, SendRate: 600,
	})
	h.Client.StartCampaign(campaign.ID)
	time.Sleep(2 * time.Second)

	messages := getMailpitMessages(t, mailpitAPI)
	if messages.Total != 1 {
		t.Fatalf("expected 1 message, got %d", messages.Total)
	}

	msg := messages.Messages[0]
	if msg.Subject != "Hello Alice Smith" {
		t.Errorf("Subject = %q, want %q", msg.Subject, "Hello Alice Smith")
	}

	detail := getMailpitMessage(t, mailpitAPI, msg.ID)
	if !strings.Contains(detail.HTML, "Dear Alice") {
		t.Errorf("HTML body missing rendered FirstName: %s", detail.HTML)
	}
	if !strings.Contains(detail.Text, "Dear Alice Smith") {
		t.Errorf("Text body missing rendered name: %s", detail.Text)
	}
}
