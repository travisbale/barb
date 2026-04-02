package phishing

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	miragesdk "github.com/travisbale/mirage/sdk"
)

// MiragedConnection represents a configured miraged instance.
type MiragedConnection struct {
	ID             string
	Name           string
	Address        string
	SecretHostname string
	CertPEM        []byte
	KeyPEM         []byte
	CACertPEM      []byte
	CreatedAt      time.Time
}

func (c *MiragedConnection) Validate() error {
	if c.Name == "" {
		return ErrNameRequired
	}
	if c.Address == "" {
		return ErrAddressRequired
	}
	if c.SecretHostname == "" {
		return ErrSecretHostnameRequired
	}
	if len(c.CertPEM) == 0 || len(c.KeyPEM) == 0 || len(c.CACertPEM) == 0 {
		return ErrCertsRequired
	}
	return nil
}

type miragedStore interface {
	CreateConnection(c *MiragedConnection) error
	GetConnection(id string) (*MiragedConnection, error)
	DeleteConnection(id string) error
	ListConnections() ([]*MiragedConnection, error)
}

// MiragedService manages miraged connections and provides SDK clients.
type MiragedService struct {
	Store miragedStore
}

// Enroll connects to a miraged instance using an invite token, generates
// a keypair, enrolls via the API, and stores the resulting credentials.
func (s *MiragedService) Enroll(name, address, secretHostname, token string) (*MiragedConnection, error) {
	// Generate ECDSA P-256 keypair.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}

	// Create CSR.
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{CommonName: "barb"},
	}, key)
	if err != nil {
		return nil, fmt.Errorf("creating CSR: %w", err)
	}
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrDER})

	// Call the enrollment endpoint (no mTLS — we don't have certs yet).
	enrollResp, err := enrollHTTP(address, secretHostname, token, string(csrPEM))
	if err != nil {
		return nil, err
	}

	// Marshal the private key.
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshaling key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	conn := &MiragedConnection{
		ID:             uuid.New().String(),
		Name:           name,
		Address:        address,
		SecretHostname: secretHostname,
		CertPEM:        []byte(enrollResp.CertPEM),
		KeyPEM:         keyPEM,
		CACertPEM:      []byte(enrollResp.CACertPEM),
		CreatedAt:      time.Now(),
	}

	if err := s.Store.CreateConnection(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

// enrollHTTP sends a CSR to the miraged enrollment endpoint and returns the
// signed certificate and CA cert. TLS verification is skipped because we
// don't have the CA cert yet — the invite token authenticates the exchange.
func enrollHTTP(address, secretHostname, token, csrPEM string) (*miragesdk.EnrollResponse, error) {
	reqBody := miragesdk.EnrollRequest{Token: token, CSRPEM: csrPEM}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:         secretHostname,
				InsecureSkipVerify: true, //nolint:gosec // enrollment bootstrap
			},
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, address)
			},
		},
	}

	enrollURL := fmt.Sprintf("https://%s%s", secretHostname, miragesdk.RouteEnroll)
	httpReq, err := http.NewRequest(http.MethodPost, enrollURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%w: %s (HTTP %d)", ErrEnrollmentRejected, body, resp.StatusCode)
	}

	var enrollResp miragesdk.EnrollResponse
	if err := json.NewDecoder(resp.Body).Decode(&enrollResp); err != nil {
		// A non-JSON response typically means the secret hostname is wrong
		// and miraged returned an error page instead of the enrollment response.
		return nil, fmt.Errorf("%w: unexpected response format", ErrEnrollmentRejected)
	}
	return &enrollResp, nil
}

func (s *MiragedService) Get(id string) (*MiragedConnection, error) {
	return s.Store.GetConnection(id)
}

func (s *MiragedService) Delete(id string) error {
	return s.Store.DeleteConnection(id)
}

func (s *MiragedService) List() ([]*MiragedConnection, error) {
	return s.Store.ListConnections()
}

// Client constructs a Mirage SDK client for the given connection ID.
func (s *MiragedService) Client(id string) (*miragesdk.Client, error) {
	conn, err := s.Store.GetConnection(id)
	if err != nil {
		return nil, err
	}
	return miragesdk.NewClient(conn.Address, conn.SecretHostname, conn.CertPEM, conn.KeyPEM, conn.CACertPEM)
}
