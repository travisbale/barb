package test

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/travisbale/barb/internal/app"
	"github.com/travisbale/barb/internal/crypto/aes"
	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/internal/store/sqlite"
	"github.com/travisbale/barb/sdk"
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

func (m *MockMailer) Dial(_ context.Context, _ *phishing.SMTPProfile) (phishing.MailConn, error) {
	return &mockConn{mailer: m}, nil
}

func (m *MockMailer) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Sent)
}

type mockConn struct {
	mailer *MockMailer
}

func (c *mockConn) Send(_ *phishing.SMTPProfile, tmpl *phishing.EmailTemplate, target *phishing.Target, _ string) error {
	c.mailer.mu.Lock()
	defer c.mailer.mu.Unlock()
	c.mailer.Sent = append(c.mailer.Sent, MockEmail{To: target.Email, Subject: tmpl.Subject})
	return nil
}

func (c *mockConn) Close() error { return nil }

// BlockingMailer wraps MockMailer but blocks each Send until explicitly released.
// Use this for deterministic cancel/concurrency tests.
type BlockingMailer struct {
	MockMailer
	// gate is buffered — send a value to release one blocked Send call.
	gate chan struct{}
	// done is closed during test cleanup to unblock any pending sends.
	done chan struct{}
}

func NewBlockingMailer() *BlockingMailer {
	return &BlockingMailer{
		gate: make(chan struct{}, 100),
		done: make(chan struct{}),
	}
}

// Release allows n pending or future Send calls to proceed.
func (b *BlockingMailer) Release(n int) {
	for i := 0; i < n; i++ {
		b.gate <- struct{}{}
	}
}

// Close unblocks all pending and future Send calls. Called during test cleanup.
func (b *BlockingMailer) CloseGate() {
	select {
	case <-b.done:
	default:
		close(b.done)
	}
}

func (b *BlockingMailer) Dial(ctx context.Context, profile *phishing.SMTPProfile) (phishing.MailConn, error) {
	return &blockingConn{mailer: b}, nil
}

type blockingConn struct {
	mailer *BlockingMailer
}

func (c *blockingConn) Send(profile *phishing.SMTPProfile, tmpl *phishing.EmailTemplate, target *phishing.Target, lureURL string) error {
	// Block until released or test cleanup.
	select {
	case <-c.mailer.gate:
	case <-c.mailer.done:
		return fmt.Errorf("mailer closed")
	}
	c.mailer.mu.Lock()
	defer c.mailer.mu.Unlock()
	c.mailer.Sent = append(c.mailer.Sent, MockEmail{To: target.Email, Subject: tmpl.Subject})
	return nil
}

func (c *blockingConn) Close() error { return nil }

// Harness is a fully-wired test environment. Obtain one via NewHarness.
type Harness struct {
	Client *sdk.Client
	Mailer *MockMailer
	Addr   string
	DB     *sqlite.DB
	Cipher *aes.Cipher
}

// NewHarness starts a server in-process with an in-memory database and a mock
// mailer. All resources are cleaned up via t.Cleanup.
func NewHarness(t *testing.T) *Harness {
	return newHarness(t, &MockMailer{})
}

// NewHarnessWithBlockingMailer starts a server with a BlockingMailer that
// blocks each Send until explicitly released via Release().
func NewHarnessWithBlockingMailer(t *testing.T) (*Harness, *BlockingMailer) {
	bm := NewBlockingMailer()
	h := newHarness(t, bm)
	h.Mailer = &bm.MockMailer
	// Unblock any pending sends when the test finishes.
	t.Cleanup(func() { bm.CloseGate() })
	return h, bm
}

// NewHarnessWithMailer starts a server with a real mailer for integration tests.
func NewHarnessWithMailer(t *testing.T, mailer phishing.Mailer) *Harness {
	return newHarness(t, mailer)
}

func newHarness(t *testing.T, mailer phishing.Mailer) *Harness {
	t.Helper()

	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("sqlite.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	var mockMailer *MockMailer
	switch m := mailer.(type) {
	case *MockMailer:
		mockMailer = m
	case *BlockingMailer:
		mockMailer = &m.MockMailer
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Use a fixed test key for deterministic encryption.
	testKey := make([]byte, 32)
	for i := range testKey {
		testKey[i] = byte(i)
	}

	// Create a test admin with a known password before app.New
	// so EnsureAdmin inside New() is a no-op.
	const testPassword = "test-password-12345"
	testHash, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	authStore := sqlite.NewAuthStore(db)
	_ = authStore.CreateUser(&phishing.User{
		ID:           "test-admin",
		Username:     "admin",
		PasswordHash: string(testHash),
		CreatedAt:    time.Now(),
	})

	cipher := aes.NewCipher(testKey)
	application, err := app.New(app.Config{
		DB:       db,
		Cipher:   cipher,
		Frontend: emptyFS{},
		Mailer:   mailer,
		Version:  "test",
		Logger:   logger,
	})
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}

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

	// Authenticate with the test admin account.
	if err := client.Login(sdk.LoginRequest{Username: "admin", Password: testPassword}); err != nil {
		t.Fatalf("test login: %v", err)
	}

	return &Harness{
		Client: client,
		Mailer: mockMailer,
		Addr:   addr,
		DB:     db,
		Cipher: cipher,
	}
}

type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}
