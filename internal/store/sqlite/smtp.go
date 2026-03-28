package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/travisbale/mirador/internal/phishing"
)

type SMTP struct{ db *DB }

func NewSMTPStore(db *DB) *SMTP { return &SMTP{db: db} }

func (s *SMTP) CreateProfile(p *phishing.SMTPProfile) error {
	_, err := s.db.db.Exec(
		`INSERT INTO smtp_profiles (id, name, host, port, username, password, from_addr, from_name, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Host, p.Port, p.Username, p.Password, p.FromAddr, p.FromName, p.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *SMTP) GetProfile(id string) (*phishing.SMTPProfile, error) {
	row := s.db.db.QueryRow(
		`SELECT id, name, host, port, username, password, from_addr, from_name, created_at
		 FROM smtp_profiles WHERE id = ?`, id,
	)
	return scanSMTPProfile(row)
}

func (s *SMTP) DeleteProfile(id string) error {
	res, err := s.db.db.Exec(`DELETE FROM smtp_profiles WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *SMTP) ListProfiles() ([]*phishing.SMTPProfile, error) {
	rows, err := s.db.db.Query(
		`SELECT id, name, host, port, username, password, from_addr, from_name, created_at
		 FROM smtp_profiles ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.SMTPProfile
	for rows.Next() {
		p, err := scanSMTPProfile(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanSMTPProfile(row scanner) (*phishing.SMTPProfile, error) {
	var (
		p         phishing.SMTPProfile
		createdAt int64
	)
	err := row.Scan(&p.ID, &p.Name, &p.Host, &p.Port, &p.Username, &p.Password, &p.FromAddr, &p.FromName, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt = time.Unix(createdAt, 0)
	return &p, nil
}
