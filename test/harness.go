package test

import (
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/travisbale/mirador/internal/api"
	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/internal/server"
	"github.com/travisbale/mirador/internal/store/sqlite"
	"github.com/travisbale/mirador/sdk"

	"log/slog"
	"os"
)

// MockMailer records Send calls instead of sending real emails.
type MockMailer struct {
	mu   sync.Mutex
	Sent []MockEmail
}

type MockEmail struct {
	To      string
	Subject string
}

func (m *MockMailer) Send(profile *phishing.SMTPProfile, tmpl *phishing.EmailTemplate, target *phishing.Target, lureURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Sent = append(m.Sent, MockEmail{To: target.Email, Subject: tmpl.Subject})
	return nil
}

func (m *MockMailer) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Sent)
}

// Harness is a fully-wired test environment. Obtain one via NewHarness.
type Harness struct {
	// Client is the SDK client pointed at the test server.
	Client *sdk.Client

	// Mailer is the mock mailer used by the campaign service.
	Mailer *MockMailer

	// Addr is the listen address of the test server (e.g. "127.0.0.1:PORT").
	Addr string
}

// NewHarness starts a server in-process with an in-memory database and returns
// an SDK client. All resources are cleaned up via t.Cleanup.
func NewHarness(t *testing.T) *Harness {
	t.Helper()

	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("sqlite.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mockMailer := &MockMailer{}

	targetStore := sqlite.NewTargetStore(db)
	templateStore := sqlite.NewTemplateStore(db)
	smtpStore := sqlite.NewSMTPStore(db)

	targetSvc := &phishing.TargetService{Store: targetStore}
	templateSvc := &phishing.TemplateService{Store: templateStore}
	smtpSvc := &phishing.SMTPService{Store: smtpStore}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	miragedSvc := &phishing.MiragedService{Store: sqlite.NewMiragedStore(db)}

	campaignSvc := &phishing.CampaignService{
		Store:     sqlite.NewCampaignStore(db),
		Targets:   targetStore,
		Templates: templateStore,
		SMTP:      smtpStore,
		Miraged:   miragedSvc,
		Mailer:    mockMailer,
		Logger:    logger,
	}

	apiRouter := &api.Router{
		Miraged:   miragedSvc,
		Campaigns: campaignSvc,
		Targets:   targetSvc,
		Templates: templateSvc,
		SMTP:      smtpSvc,
		Version:   "test",
		Logger:    logger,
	}

	// Use an empty fs.FS for the frontend — tests don't need the SPA.
	srv := server.New(server.Config{
		Addr:     "127.0.0.1:0",
		API:      apiRouter,
		Frontend: emptyFS{},
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	addr := listener.Addr().String()

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	t.Cleanup(func() { _ = srv.Close() })

	// Wait for server to be ready.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	client := sdk.NewClient(fmt.Sprintf("http://%s", addr))

	return &Harness{
		Client: client,
		Mailer: mockMailer,
		Addr:   addr,
	}
}

// emptyFS is a minimal fs.FS that contains no files.
type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}
