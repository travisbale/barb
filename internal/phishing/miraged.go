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

// Client constructs a Mirage SDK client for the given connection ID.
func (s *MiragedService) Client(id string) (*miragesdk.Client, error) {
	conn, err := s.Store.GetConnection(id)
	if err != nil {
		return nil, err
	}
	return miragesdk.NewClient(conn.Address, conn.SecretHostname, conn.CertPEM, conn.KeyPEM, conn.CACertPEM)
}

// TestConnection verifies that a connection to the miraged instance works.
func (s *MiragedService) TestConnection(id string) error {
	client, err := s.Client(id)
	if err != nil {
		return err
	}
	_, err = client.Status()
	return err
}
