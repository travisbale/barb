package test_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	miragesdk "github.com/travisbale/mirage/sdk"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/internal/store/sqlite"
	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

// testCerts generates a self-signed CA, server cert, and client cert for
// mTLS testing. Returns PEM-encoded bytes.
func testCerts(t *testing.T, serverAddr string) (caCertPEM, serverCertPEM, serverKeyPEM, clientCertPEM, clientKeyPEM []byte) {
	t.Helper()

	// Generate CA key and certificate.
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating CA key: %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("creating CA cert: %v", err)
	}
	caCert, _ := x509.ParseCertificate(caCertDER)
	caCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

	// Generate server cert signed by CA.
	host, _, _ := net.SplitHostPort(serverAddr)
	serverKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "mock-miraged"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"mock-miraged"},
		IPAddresses:  []net.IP{net.ParseIP(host)},
	}
	serverCertDER, _ := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	serverCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	serverKeyDER, _ := x509.MarshalECPrivateKey(serverKey)
	serverKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyDER})

	// Generate client cert signed by CA.
	clientKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "barb-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCertDER, _ := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	clientCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})
	clientKeyDER, _ := x509.MarshalECPrivateKey(clientKey)
	clientKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER})

	return
}

// mockMiraged starts an mTLS HTTPS server that serves SSE events. Send
// SessionEvents to the returned channel to push them to connected clients.
func mockMiraged(t *testing.T) (address, secretHostname string, clientCertPEM, clientKeyPEM, caCertPEM []byte, events chan<- miragesdk.SessionEvent) {
	t.Helper()

	// Listen on a random port first to get the address for cert generation.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	addr := listener.Addr().String()

	ca, serverCert, serverKey, clientCert, clientKey := testCerts(t, addr)

	ch := make(chan miragesdk.SessionEvent, 10)

	mux := http.NewServeMux()

	// SSE stream for session events.
	mux.HandleFunc("/api/sessions/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		flusher.Flush()

		for {
			select {
			case evt := <-ch:
				data, _ := json.Marshal(evt.Session)
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	// Stub: push phishlet.
	mux.HandleFunc("POST /api/phishlets", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"name": "example", "enabled": false})
	})

	// Stub: create lure.
	mux.HandleFunc("POST /api/lures", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"id": "lure-1", "phishlet": "example", "url": "https://mock-miraged/lure-1",
		})
	})

	// Stub: generate lure URL with tracking params + delete lure.
	mux.HandleFunc("/api/lures/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			json.NewEncoder(w).Encode(map[string]any{"url": "https://mock-miraged/lure-1?p=encrypted"})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	// Stub: disable phishlet (cleanup).
	mux.HandleFunc("POST /api/phishlets/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"name": "example", "enabled": false})
	})

	// Parse server TLS cert.
	serverTLSCert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		t.Fatalf("parsing server cert: %v", err)
	}

	// Build CA pool for client verification.
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(ca)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}

	tlsListener := tls.NewListener(listener, tlsConfig)
	srv := &http.Server{Handler: mux}
	go srv.Serve(tlsListener)
	t.Cleanup(func() { srv.Close() })

	return addr, "mock-miraged", clientCert, clientKey, ca, ch
}

func TestCampaigns_SessionCorrelation(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	// Start a mock miraged server.
	address, secretHostname, clientCertPEM, clientKeyPEM, caCertPEM, events := mockMiraged(t)

	// Insert a miraged connection directly into the store, bypassing the
	// enrollment handshake which requires a real miraged instance.
	miragedStore := sqlite.NewMiragedStore(h.DB, h.Cipher)
	conn := &phishing.MiragedConnection{
		ID:             "test-miraged",
		Name:           "Mock Miraged",
		Address:        address,
		SecretHostname: secretHostname,
		CertPEM:        clientCertPEM,
		KeyPEM:         clientKeyPEM,
		CACertPEM:      caCertPEM,
		CreatedAt:      time.Now(),
	}
	if err := miragedStore.CreateConnection(conn); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	// Create a phishlet so ensureLure can push it.
	if _, err := h.Client.CreatePhishlet(sdk.CreatePhishletRequest{YAML: validPhishletYAML}); err != nil {
		t.Fatalf("CreatePhishlet: %v", err)
	}

	// Create campaign prerequisites.
	list := createTestTargetList(t, h, sdk.AddTargetRequest{
		Email: "alice@acme.com", FirstName: "Alice",
	})
	tmpl := createTestTemplate(t, h)
	smtp := createTestSMTP(t, h)

	// Create a campaign linked to the mock miraged.
	req := validCampaignRequest(list.ID, tmpl.ID, smtp.ID)
	req.MiragedID = conn.ID
	req.Phishlet = "example"
	req.SendRate = 600
	campaign, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	if err := h.Client.StartCampaign(campaign.ID); err != nil {
		t.Fatalf("StartCampaign: %v", err)
	}
	waitForEmails(t, h, 1)

	// Get the result ID to use as a tracking token.
	results, _ := h.Client.ListCampaignResults(campaign.ID)
	var resultID string
	for _, r := range results {
		if r.Email == "alice@acme.com" {
			resultID = r.ID
			break
		}
	}
	if resultID == "" {
		t.Fatal("could not find result for alice@acme.com")
	}

	sessionEvent := miragesdk.SessionResponse{
		ID:         "session-123",
		Phishlet:   "example",
		Username:   "alice@acme.com",
		Password:   "hunter2",
		LureParams: map[string]string{"t": resultID},
		StartedAt:  time.Now().Add(-30 * time.Second),
	}

	// Stage 1: Push session created with tracking param — result should become "clicked".
	events <- miragesdk.SessionEvent{Type: miragesdk.EventSessionCreated, Session: sessionEvent}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		results, _ := h.Client.ListCampaignResults(campaign.ID)
		for _, r := range results {
			if r.Email == "alice@acme.com" && r.Status == "clicked" {
				if r.SessionID != "session-123" {
					t.Errorf("SessionID = %q, want %q", r.SessionID, "session-123")
				}
				if r.ClickedAt == nil {
					t.Error("ClickedAt is nil after click")
				}
				goto stage2
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("timed out waiting for clicked status")

stage2:
	// Stage 2: Push credential capture — result should become "captured".
	events <- miragesdk.SessionEvent{Type: miragesdk.EventCredsCaptured, Session: sessionEvent}

	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		results, _ := h.Client.ListCampaignResults(campaign.ID)
		for _, r := range results {
			if r.Email == "alice@acme.com" && r.Status == "captured" {
				if r.CapturedAt == nil {
					t.Error("CapturedAt is nil after capture")
				}
				goto stage3
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("timed out waiting for captured status")

stage3:
	// Stage 3: Push session completed — result should become "completed".
	events <- miragesdk.SessionEvent{Type: miragesdk.EventSessionCompleted, Session: sessionEvent}

	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		results, _ := h.Client.ListCampaignResults(campaign.ID)
		for _, r := range results {
			if r.Email == "alice@acme.com" && r.Status == "completed" {
				return // success — full lifecycle: sent → clicked → captured → completed
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("timed out waiting for completed status")
}
