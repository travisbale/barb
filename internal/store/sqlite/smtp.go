package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/travisbale/barb/internal/crypto/aes"
	"github.com/travisbale/barb/internal/phishing"
)

type SMTP struct {
	db     *DB
	cipher *aes.Cipher
}

func NewSMTPStore(db *DB, cipher *aes.Cipher) *SMTP {
	return &SMTP{db: db, cipher: cipher}
}

func (s *SMTP) CreateProfile(p *phishing.SMTPProfile) error {
	encrypted, err := s.cipher.Encrypt(p.Password)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}
	headersJSON, err := marshalHeaders(p.CustomHeaders)
	if err != nil {
		return err
	}
	_, err = s.db.db.Exec(
		`INSERT INTO smtp_profiles (id, name, host, port, username, password, from_addr, from_name, custom_headers, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Host, p.Port, p.Username, encrypted, p.FromAddr, p.FromName, headersJSON, p.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *SMTP) GetProfile(id string) (*phishing.SMTPProfile, error) {
	row := s.db.db.QueryRow(
		`SELECT id, name, host, port, username, password, from_addr, from_name, custom_headers, created_at
		 FROM smtp_profiles WHERE id = ?`, id,
	)
	return s.scanSMTPProfile(row)
}

func (s *SMTP) DeleteProfile(id string) error {
	res, err := s.db.db.Exec(`DELETE FROM smtp_profiles WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *SMTP) UpdateProfile(p *phishing.SMTPProfile) error {
	encrypted, err := s.cipher.Encrypt(p.Password)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}
	headersJSON, err := marshalHeaders(p.CustomHeaders)
	if err != nil {
		return err
	}
	res, err := s.db.db.Exec(
		`UPDATE smtp_profiles SET name = ?, host = ?, port = ?, username = ?, password = ?, from_addr = ?, from_name = ?, custom_headers = ?
		 WHERE id = ?`,
		p.Name, p.Host, p.Port, p.Username, encrypted, p.FromAddr, p.FromName, headersJSON, p.ID,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *SMTP) ListProfiles() ([]*phishing.SMTPProfile, error) {
	rows, err := s.db.db.Query(
		`SELECT id, name, host, port, username, password, from_addr, from_name, custom_headers, created_at
		 FROM smtp_profiles ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.SMTPProfile
	for rows.Next() {
		p, err := s.scanSMTPProfile(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SMTP) scanSMTPProfile(row scanner) (*phishing.SMTPProfile, error) {
	var (
		p            phishing.SMTPProfile
		encryptedPwd string
		headersJSON  string
		createdAt    int64
	)
	err := row.Scan(&p.ID, &p.Name, &p.Host, &p.Port, &p.Username, &encryptedPwd, &p.FromAddr, &p.FromName, &headersJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Password, err = s.cipher.Decrypt(encryptedPwd)
	if err != nil {
		return nil, fmt.Errorf("decrypting password for profile %s: %w", p.ID, err)
	}
	p.CustomHeaders = make(map[string]string)
	if headersJSON != "" && headersJSON != "{}" {
		if err := json.Unmarshal([]byte(headersJSON), &p.CustomHeaders); err != nil {
			return nil, fmt.Errorf("parsing custom headers for profile %s: %w", p.ID, err)
		}
	}
	p.CreatedAt = time.Unix(createdAt, 0)
	return &p, nil
}

func marshalHeaders(headers map[string]string) (string, error) {
	if len(headers) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(headers)
	if err != nil {
		return "", fmt.Errorf("marshaling custom headers: %w", err)
	}
	return string(data), nil
}
