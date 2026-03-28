package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/travisbale/mirador/internal/phishing"
)

type Miraged struct{ db *DB }

func NewMiragedStore(db *DB) *Miraged { return &Miraged{db: db} }

func (s *Miraged) CreateConnection(c *phishing.MiragedConnection) error {
	_, err := s.db.db.Exec(
		`INSERT INTO miraged_connections (id, name, address, secret_hostname, cert_pem, key_pem, ca_cert_pem, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Address, c.SecretHostname, c.CertPEM, c.KeyPEM, c.CACertPEM, c.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Miraged) GetConnection(id string) (*phishing.MiragedConnection, error) {
	row := s.db.db.QueryRow(
		`SELECT id, name, address, secret_hostname, cert_pem, key_pem, ca_cert_pem, created_at
		 FROM miraged_connections WHERE id = ?`, id,
	)
	return scanMiragedConnection(row)
}

func (s *Miraged) DeleteConnection(id string) error {
	res, err := s.db.db.Exec(`DELETE FROM miraged_connections WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Miraged) ListConnections() ([]*phishing.MiragedConnection, error) {
	rows, err := s.db.db.Query(
		`SELECT id, name, address, secret_hostname, cert_pem, key_pem, ca_cert_pem, created_at
		 FROM miraged_connections ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.MiragedConnection
	for rows.Next() {
		c, err := scanMiragedConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scanMiragedConnection(row scanner) (*phishing.MiragedConnection, error) {
	var (
		c         phishing.MiragedConnection
		createdAt int64
	)
	err := row.Scan(&c.ID, &c.Name, &c.Address, &c.SecretHostname, &c.CertPEM, &c.KeyPEM, &c.CACertPEM, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.CreatedAt = time.Unix(createdAt, 0)
	return &c, nil
}
