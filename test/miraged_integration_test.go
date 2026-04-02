//go:build !unit

package test_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/travisbale/barb/internal/delivery"
	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func startMiraged(t *testing.T) (address, secretHostname, token string) {
	t.Helper()
	ctx := context.Background()
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../mirage",
			Dockerfile: "Dockerfile",
		},
		ExposedPorts: []string{"443/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            strings.NewReader(miragedConfig),
				ContainerFilePath: "/etc/mirage/miraged.yaml",
				FileMode:          0644,
			},
		},
		Tmpfs:      map[string]string{"/var/lib/mirage": "uid=65532,gid=65532"},
		WaitingFor: wait.ForLog("enroll with:").WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("skipping: could not start miraged container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("getting container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "443/tcp")
	if err != nil {
		t.Fatalf("getting mapped port: %v", err)
	}

	token = extractInviteToken(t, ctx, container)
	return fmt.Sprintf("%s:%d", host, port.Int()), "mgmt.phish.local", token
}

func extractInviteToken(t *testing.T, ctx context.Context, container testcontainers.Container) string {
	t.Helper()
	logs, err := container.Logs(ctx)
	if err != nil {
		t.Fatalf("reading container logs: %v", err)
	}
	defer logs.Close()
	logBytes, _ := io.ReadAll(logs)

	tokenRe := regexp.MustCompile(`--token\s+([0-9a-f-]+)`)
	matches := tokenRe.FindStringSubmatch(string(logBytes))
	if len(matches) < 2 {
		t.Fatalf("could not find invite token in miraged logs: %s", logBytes)
	}
	return matches[1]
}

// TestIntegration_Miraged starts a single miraged container and runs all
// miraged-related subtests against it.
func TestIntegration_Miraged(t *testing.T) {
	address, secretHostname, token := startMiraged(t)
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	// Enroll once — the token is single-use.
	conn, err := h.Client.EnrollMiraged(sdk.EnrollMiragedRequest{
		Name:           "Test Miraged",
		Address:        address,
		SecretHostname: secretHostname,
		Token:          token,
	})
	if err != nil {
		t.Fatalf("EnrollMiraged: %v", err)
	}

	t.Run("ConnectionTest", func(t *testing.T) {
		status, err := h.Client.TestMiraged(conn.ID)
		if err != nil {
			t.Fatalf("TestMiraged: %v", err)
		}
		if !status.Connected {
			t.Errorf("expected connected, got error: %s", status.Error)
		}
	})

	t.Run("PushAndListPhishlets", func(t *testing.T) {
		_, err := h.Client.CreatePhishlet(sdk.CreatePhishletRequest{
			YAML: validPhishletYAML,
		})
		if err != nil {
			t.Fatalf("CreatePhishlet: %v", err)
		}

		if err := h.Client.PushMiragedPhishlet(conn.ID, validPhishletYAML); err != nil {
			t.Fatalf("PushMiragedPhishlet: %v", err)
		}
	})

	t.Run("EnableAndDisablePhishlet", func(t *testing.T) {
		enabled, err := h.Client.EnableMiragedPhishlet(conn.ID, "example", sdk.EnableMiragedPhishletRequest{
			Hostname: "login.phish.local",
		})
		if err != nil {
			t.Fatalf("EnableMiragedPhishlet: %v", err)
		}
		if !enabled.Enabled {
			t.Error("expected phishlet to be enabled")
		}
		if enabled.Hostname != "login.phish.local" {
			t.Errorf("Hostname = %q, want %q", enabled.Hostname, "login.phish.local")
		}

		disabled, err := h.Client.DisableMiragedPhishlet(conn.ID, "example")
		if err != nil {
			t.Fatalf("DisableMiragedPhishlet: %v", err)
		}
		if disabled.Enabled {
			t.Error("expected phishlet to be disabled")
		}
	})

	t.Run("CreateLure", func(t *testing.T) {
		// Re-enable the phishlet so we can create a lure.
		_, err := h.Client.EnableMiragedPhishlet(conn.ID, "example", sdk.EnableMiragedPhishletRequest{
			Hostname: "login.phish.local",
		})
		if err != nil {
			t.Fatalf("EnableMiragedPhishlet: %v", err)
		}

		// Create a campaign that references this miraged + phishlet.
		smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
			Name: "Lure SMTP", Host: "localhost", Port: 1025, FromAddr: "test@example.com",
		})
		tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
			Name: "Lure Template", Subject: "Test", HTMLBody: "<p>{{.URL}}</p>",
		})
		list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Lure Targets"})
		h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "target@example.com"})

		campaign, err := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
			Name:          "Lure Test",
			TemplateID:    tmpl.ID,
			SMTPProfileID: smtp.ID,
			TargetListID:  list.ID,
			MiragedID:     conn.ID,
			Phishlet:      "example",
			RedirectURL:   "https://example.com",
			SendRate:      600,
		})
		if err != nil {
			t.Fatalf("CreateCampaign: %v", err)
		}

		// Start the campaign — this creates the lure on miraged.
		if err := h.Client.StartCampaign(campaign.ID); err != nil {
			t.Fatalf("StartCampaign: %v", err)
		}

		time.Sleep(2 * time.Second)

		// Verify the campaign has a lure URL.
		got, err := h.Client.GetCampaign(campaign.ID)
		if err != nil {
			t.Fatalf("GetCampaign: %v", err)
		}
		if got.LureURL == "" {
			t.Error("expected campaign to have a lure URL after starting")
		}
	})

	t.Run("TestEmailWithLure", func(t *testing.T) {
		smtpHost, smtpPort, mailpitAPI := startMailpit(t)

		smtp, _ := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
			Name: "Mailpit Lure", Host: smtpHost, Port: smtpPort, FromAddr: "test@example.com",
		})
		tmpl, _ := h.Client.CreateTemplate(sdk.CreateTemplateRequest{
			Name: "Lure Email", Subject: "Click here", HTMLBody: "<p>Visit <a href=\"{{.URL}}\">this link</a></p>",
		})
		list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Lure Email Targets"})
		h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "victim@example.com"})

		campaign, _ := h.Client.CreateCampaign(sdk.CreateCampaignRequest{
			Name:          "Test Email Lure",
			TemplateID:    tmpl.ID,
			SMTPProfileID: smtp.ID,
			TargetListID:  list.ID,
			MiragedID:     conn.ID,
			Phishlet:      "example",
			RedirectURL:   "https://example.com",
		})

		// Send a test email — this creates a persistent lure and includes its URL.
		err := h.Client.SendTestEmail(campaign.ID, sdk.SendTestEmailRequest{Email: "operator@example.com"})
		if err != nil {
			t.Fatalf("SendTestEmail: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		messages := getMailpitMessages(t, mailpitAPI)
		if messages.Total != 1 {
			t.Fatalf("expected 1 message, got %d", messages.Total)
		}

		// Verify the email body contains a real lure URL (not the placeholder).
		detail := getMailpitMessage(t, mailpitAPI, messages.Messages[0].ID)
		if strings.Contains(detail.HTML, "example.com/test-lure") {
			t.Error("email contains placeholder URL instead of real lure URL")
		}
		if !strings.Contains(detail.HTML, "https://") {
			t.Error("email body missing lure URL")
		}
	})
}
