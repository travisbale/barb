package phishing

import (
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

func (s *MiragedService) Create(conn *MiragedConnection) (*MiragedConnection, error) {
	if err := conn.Validate(); err != nil {
		return nil, err
	}

	conn.ID = uuid.New().String()
	conn.CreatedAt = time.Now()

	if err := s.Store.CreateConnection(conn); err != nil {
		return nil, err
	}
	return conn, nil
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

// client constructs a Mirage SDK client for the given connection ID.
func (s *MiragedService) client(id string) (*miragesdk.Client, error) {
	conn, err := s.Store.GetConnection(id)
	if err != nil {
		return nil, err
	}
	return miragesdk.NewClient(conn.Address, conn.SecretHostname, conn.CertPEM, conn.KeyPEM, conn.CACertPEM)
}

// MiragedStatus holds the result of a connectivity test to a miraged instance.
type MiragedStatus struct {
	Connected bool
	Version   string
	Error     string
}

// TestConnection verifies connectivity to the miraged instance and returns
// its status. A connection or protocol error is reported in the returned
// status, not as a Go error — the only errors returned are lookup failures.
func (s *MiragedService) TestConnection(id string) (*MiragedStatus, error) {
	client, err := s.client(id)
	if err != nil {
		return &MiragedStatus{Error: err.Error()}, nil
	}
	status, err := client.Status()
	if err != nil {
		return &MiragedStatus{Error: err.Error()}, nil
	}
	return &MiragedStatus{Connected: true, Version: status.Version}, nil
}

// MiragedPhishlet is a phishlet reported by a miraged instance.
type MiragedPhishlet struct {
	Name     string
	Hostname string
	Enabled  bool
}

// ListPhishlets retrieves the phishlet list from the miraged instance.
func (s *MiragedService) ListPhishlets(id string) ([]MiragedPhishlet, error) {
	client, err := s.client(id)
	if err != nil {
		return nil, err
	}
	resp, err := client.ListPhishlets()
	if err != nil {
		return nil, err
	}
	phishlets := make([]MiragedPhishlet, len(resp.Items))
	for i, p := range resp.Items {
		phishlets[i] = MiragedPhishlet{
			Name:     p.Name,
			Hostname: p.Hostname,
			Enabled:  p.Enabled,
		}
	}
	return phishlets, nil
}

// PushPhishlet deploys a phishlet YAML config to the miraged instance.
func (s *MiragedService) PushPhishlet(connectionID, yamlContent string) error {
	mirageClient, err := s.client(connectionID)
	if err != nil {
		return err
	}
	_, err = mirageClient.PushPhishlet(miragesdk.PushPhishletRequest{YAML: yamlContent})
	return err
}
