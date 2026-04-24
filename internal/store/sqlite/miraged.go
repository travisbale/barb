package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/travisbale/barb/internal/crypto/aes"
	"github.com/travisbale/barb/internal/phishing"
)

type Miraged struct {
	db     *DB
	cipher *aes.Cipher
}

func NewMiragedStore(db *DB, cipher *aes.Cipher) *Miraged {
	return &Miraged{db: db, cipher: cipher}
}

func (s *Miraged) CreateConnection(c *phishing.MiragedConnection) error {
	encryptedKey, err := s.cipher.Encrypt(c.KeyPEM)
	if err != nil {
		return fmt.Errorf("encrypting client key: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO miraged_connections (id, name, address, secret_hostname, cert_pem, key_pem, ca_cert_pem, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Address, c.SecretHostname, c.CertPEM, encryptedKey, c.CACertPEM, c.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Miraged) GetConnection(id string) (*phishing.MiragedConnection, error) {
	row := s.db.QueryRow(
		`SELECT id, name, address, secret_hostname, cert_pem, key_pem, ca_cert_pem, created_at
		 FROM miraged_connections WHERE id = ?`, id,
	)
	return s.scanMiragedConnection(row)
}

func (s *Miraged) UpdateConnectionName(id, name string) (*phishing.MiragedConnection, error) {
	res, err := s.db.Exec(`UPDATE miraged_connections SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return nil, err
	}
	if err := requireOneRow(res); err != nil {
		return nil, err
	}
	return s.GetConnection(id)
}

func (s *Miraged) DeleteConnection(id string) error {
	res, err := s.db.Exec(`DELETE FROM miraged_connections WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Miraged) ListConnections() ([]*phishing.MiragedConnection, error) {
	rows, err := s.db.Query(
		`SELECT id, name, address, secret_hostname, cert_pem, key_pem, ca_cert_pem, created_at
		 FROM miraged_connections ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.MiragedConnection
	for rows.Next() {
		c, err := s.scanMiragedConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Miraged) scanMiragedConnection(row scanner) (*phishing.MiragedConnection, error) {
	var (
		c            phishing.MiragedConnection
		encryptedKey []byte
		createdAt    int64
	)
	err := row.Scan(&c.ID, &c.Name, &c.Address, &c.SecretHostname, &c.CertPEM, &encryptedKey, &c.CACertPEM, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.KeyPEM, err = s.cipher.Decrypt(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting client key for connection %s: %w", c.ID, err)
	}
	c.CreatedAt = time.Unix(createdAt, 0)
	return &c, nil
}
