//go:build !unit

package test_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
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

// Shared Mailpit container — started once and reused across all integration tests.
// Terminated by TestMain after m.Run() returns.
var (
	mailpitOnce     sync.Once
	sharedContainer testcontainers.Container
	sharedSMTPHost  string
	sharedSMTPPort  int
	sharedAPIURL    string
	mailpitSkipMsg  string
	mailpitStartErr error
)

func TestMain(m *testing.M) {
	code := m.Run()
	if sharedContainer != nil {
		if err := sharedContainer.Terminate(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "failed to terminate mailpit container: %v\n", err)
		}
	}
	os.Exit(code)
}

func startSharedMailpit() {
	// Disable Ryuk reaper — testcontainers can't detect process exit in all environments.
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()
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
		mailpitSkipMsg = fmt.Sprintf("could not start mailpit container: %v", err)
		mailpitStartErr = err
		return
	}
	sharedContainer = container

	host, err := container.Host(ctx)
	if err != nil {
		mailpitStartErr = fmt.Errorf("getting container host: %w", err)
		return
	}
	smtpMapped, err := container.MappedPort(ctx, "1025/tcp")
	if err != nil {
		mailpitStartErr = fmt.Errorf("getting SMTP port: %w", err)
		return
	}
	apiMapped, err := container.MappedPort(ctx, "8025/tcp")
	if err != nil {
		mailpitStartErr = fmt.Errorf("getting API port: %w", err)
		return
	}

	sharedSMTPHost = host
	sharedSMTPPort = smtpMapped.Int()
	sharedAPIURL = fmt.Sprintf("http://%s:%d", host, apiMapped.Int())
}

// requireMailpit ensures the shared Mailpit container is running.
// Each test uses unique recipient addresses so no clearing is needed.
func requireMailpit(t *testing.T) (smtpHost string, smtpPort int, apiURL string) {
	t.Helper()
	mailpitOnce.Do(startSharedMailpit)
	if mailpitStartErr != nil {
		t.Skipf("skipping: %s", mailpitSkipMsg)
	}
	return sharedSMTPHost, sharedSMTPPort, sharedAPIURL
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

// searchMailpit queries Mailpit for messages matching a search query.
func searchMailpit(t *testing.T, apiURL, query string) mailpitList {
	t.Helper()
	resp, err := http.Get(apiURL + "/api/v1/search?query=" + url.QueryEscape(query))
	if err != nil {
		t.Fatalf("searching mailpit: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result mailpitList
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("parsing mailpit search: %v (body: %s)", err, body)
	}
	return result
}

// waitForMailpit polls Mailpit until at least n messages matching the query
// have been received. If query is empty, returns all messages.
func waitForMailpit(t *testing.T, apiURL string, n int, query string) mailpitList {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var result mailpitList
		if query == "" {
			result = getMailpitMessages(t, apiURL)
		} else {
			result = searchMailpit(t, apiURL, query)
		}
		if result.Total >= n {
			return result
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d messages (query=%q)", n, query)
	return mailpitList{}
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
	t.Parallel()
	smtpHost, smtpPort, mailpitAPI := requireMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp := createTestSMTP(t, h, func(r *sdk.CreateSMTPProfileRequest) {
		r.Host = smtpHost
		r.Port = smtpPort
		r.FromAddr = "phisher@sendtest.example.com"
		r.FromName = "IT Support"
	})
	tmpl := createTestTemplate(t, h, func(r *sdk.CreateTemplateRequest) {
		r.Subject = "Hello {{.FirstName}}"
		r.HTMLBody = "<p>Dear {{.FirstName}}, click <a href=\"{{.URL}}\">here</a>.</p>"
	})
	list := createTestTargetList(t, h, sdk.AddTargetRequest{Email: "victim@sendtest.example.com", FirstName: "Alice"})

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	campaign, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	err = h.Client.SendTestEmail(campaign.ID, sdk.SendTestEmailRequest{Email: "operator@sendtest.example.com"})
	if err != nil {
		t.Fatalf("SendTestEmail: %v", err)
	}

	messages := waitForMailpit(t, mailpitAPI, 1, "to:sendtest.example.com")

	msg := messages.Messages[0]
	if msg.Subject != "Hello Test" {
		t.Errorf("Subject = %q, want %q", msg.Subject, "Hello Test")
	}
	if msg.From.Address != "phisher@sendtest.example.com" {
		t.Errorf("From = %q, want %q", msg.From.Address, "phisher@sendtest.example.com")
	}
	if msg.From.Name != "IT Support" {
		t.Errorf("FromName = %q, want %q", msg.From.Name, "IT Support")
	}
	if len(msg.To) == 0 || msg.To[0].Address != "operator@sendtest.example.com" {
		t.Errorf("To = %v, want operator@sendtest.example.com", msg.To)
	}
}

func TestIntegration_CampaignSendsEmails(t *testing.T) {
	t.Parallel()
	smtpHost, smtpPort, mailpitAPI := requireMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp := createTestSMTP(t, h, func(r *sdk.CreateSMTPProfileRequest) {
		r.Host = smtpHost
		r.Port = smtpPort
	})
	tmpl := createTestTemplate(t, h, func(r *sdk.CreateTemplateRequest) {
		r.Subject = "Important: {{.FirstName}}"
	})
	list := createTestTargetList(t, h,
		sdk.AddTargetRequest{Email: "alice@sends.example.com", FirstName: "Alice"},
		sdk.AddTargetRequest{Email: "bob@sends.example.com", FirstName: "Bob"},
	)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	campaign, _ := h.Client.CreateCampaign(req)

	if err := h.Client.StartCampaign(campaign.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}

	messages := waitForMailpit(t, mailpitAPI, 2, "to:sends.example.com")

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
	t.Parallel()
	smtpHost, smtpPort, mailpitAPI := requireMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp := createTestSMTP(t, h, func(r *sdk.CreateSMTPProfileRequest) {
		r.Host = smtpHost
		r.Port = smtpPort
		r.CustomHeaders = map[string]string{
			"X-Mailer":     "Outlook 16.0",
			"X-Custom-Tag": "phishing-test",
		}
	})
	tmpl := createTestTemplate(t, h, func(r *sdk.CreateTemplateRequest) {
		r.Subject = "Test Headers"
	})
	list := createTestTargetList(t, h, sdk.AddTargetRequest{Email: "target@headers.example.com"})

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	campaign, _ := h.Client.CreateCampaign(req)

	h.Client.StartCampaign(campaign.ID)

	messages := waitForMailpit(t, mailpitAPI, 1, "to:headers.example.com")

	headers := getMailpitMessageHeaders(t, mailpitAPI, messages.Messages[0].ID)
	if v := headers["X-Mailer"]; len(v) == 0 || v[0] != "Outlook 16.0" {
		t.Errorf("X-Mailer = %v, want [Outlook 16.0]", v)
	}
	if v := headers["X-Custom-Tag"]; len(v) == 0 || v[0] != "phishing-test" {
		t.Errorf("X-Custom-Tag = %v, want [phishing-test]", v)
	}
}

func TestIntegration_EnvelopeSender(t *testing.T) {
	t.Parallel()
	smtpHost, smtpPort, mailpitAPI := requireMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp := createTestSMTP(t, h, func(r *sdk.CreateSMTPProfileRequest) {
		r.Host = smtpHost
		r.Port = smtpPort
		r.FromAddr = "visible@envelope.example.com"
	})
	tmpl := createTestTemplate(t, h, func(r *sdk.CreateTemplateRequest) {
		r.Subject = "Envelope Sender Test"
		r.EnvelopeSender = "bounce@attacker.com"
	})
	list := createTestTargetList(t, h, sdk.AddTargetRequest{Email: "target@envelope.example.com"})

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	campaign, _ := h.Client.CreateCampaign(req)
	h.Client.StartCampaign(campaign.ID)

	messages := waitForMailpit(t, mailpitAPI, 1, "to:envelope.example.com")

	if messages.Messages[0].From.Address != "visible@envelope.example.com" {
		t.Errorf("From = %q, want visible@envelope.example.com", messages.Messages[0].From.Address)
	}

	detail := getMailpitMessage(t, mailpitAPI, messages.Messages[0].ID)
	if detail.ReturnPath != "bounce@attacker.com" {
		t.Errorf("ReturnPath = %q, want bounce@attacker.com", detail.ReturnPath)
	}
}

func TestIntegration_CampaignCancelStopsSending(t *testing.T) {
	t.Parallel()
	smtpHost, smtpPort, mailpitAPI := requireMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp := createTestSMTP(t, h, func(r *sdk.CreateSMTPProfileRequest) {
		r.Host = smtpHost
		r.Port = smtpPort
	})
	tmpl := createTestTemplate(t, h)

	targets := make([]sdk.AddTargetRequest, 20)
	for i := range targets {
		targets[i] = sdk.AddTargetRequest{Email: fmt.Sprintf("user%d@cancel.example.com", i)}
	}
	list := createTestTargetList(t, h, targets...)

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 1 // 1 per minute — very slow
	campaign, _ := h.Client.CreateCampaign(req)

	h.Client.StartCampaign(campaign.ID)
	waitForMailpit(t, mailpitAPI, 1, "to:cancel.example.com")
	h.Client.CancelCampaign(campaign.ID)

	// Brief pause for cancellation to propagate.
	time.Sleep(100 * time.Millisecond)
	messages := searchMailpit(t, mailpitAPI, "to:cancel.example.com")
	if messages.Total >= 20 {
		t.Errorf("expected fewer than 20 emails (campaign cancelled), got %d", messages.Total)
	}
	if messages.Total == 0 {
		t.Error("expected at least 1 email to be sent before cancel")
	}
}

func TestIntegration_TemplateVariablesRendered(t *testing.T) {
	t.Parallel()
	smtpHost, smtpPort, mailpitAPI := requireMailpit(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	smtp := createTestSMTP(t, h, func(r *sdk.CreateSMTPProfileRequest) {
		r.Host = smtpHost
		r.Port = smtpPort
	})
	tmpl := createTestTemplate(t, h, func(r *sdk.CreateTemplateRequest) {
		r.Subject = "Hello {{.FirstName}} {{.LastName}}"
		r.HTMLBody = "<p>Dear {{.FirstName}}, please visit <a href=\"{{.URL}}\">this link</a>.</p>"
		r.TextBody = "Dear {{.FirstName}} {{.LastName}}, visit {{.URL}}"
	})
	list := createTestTargetList(t, h, sdk.AddTargetRequest{
		Email: "alice@variables.example.com", FirstName: "Alice", LastName: "Smith",
	})

	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.SendRate = 600
	campaign, _ := h.Client.CreateCampaign(req)
	h.Client.StartCampaign(campaign.ID)

	messages := waitForMailpit(t, mailpitAPI, 1, "to:variables.example.com")

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
