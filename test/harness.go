package test

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/travisbale/mirador/internal/app"
	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/internal/store/sqlite"
	"github.com/travisbale/mirador/sdk"
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
	Client *sdk.Client
	Mailer *MockMailer
	Addr   string
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	application := app.New(app.Config{
		DB:       db,
		Frontend: emptyFS{},
		Mailer:   mockMailer,
		Version:  "test",
		Logger:   logger,
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	addr := listener.Addr().String()

	srv := &http.Server{Handler: application.Handler()}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	t.Cleanup(func() {
		application.Shutdown()
		_ = srv.Close()
	})

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

type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}
