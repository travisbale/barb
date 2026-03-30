package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/travisbale/barb/internal/phishing"
)

type Templates struct{ db *DB }

func NewTemplateStore(db *DB) *Templates { return &Templates{db: db} }

func (s *Templates) CreateTemplate(t *phishing.EmailTemplate) error {
	_, err := s.db.db.Exec(
		`INSERT INTO email_templates (id, name, subject, html_body, text_body, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.Subject, t.HTMLBody, t.TextBody, t.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Templates) GetTemplate(id string) (*phishing.EmailTemplate, error) {
	row := s.db.db.QueryRow(
		`SELECT id, name, subject, html_body, text_body, created_at
		 FROM email_templates WHERE id = ?`, id,
	)
	return scanTemplate(row)
}

func (s *Templates) UpdateTemplate(t *phishing.EmailTemplate) error {
	res, err := s.db.db.Exec(
		`UPDATE email_templates SET name = ?, subject = ?, html_body = ?, text_body = ?
		 WHERE id = ?`,
		t.Name, t.Subject, t.HTMLBody, t.TextBody, t.ID,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Templates) DeleteTemplate(id string) error {
	res, err := s.db.db.Exec(`DELETE FROM email_templates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Templates) ListTemplates() ([]*phishing.EmailTemplate, error) {
	rows, err := s.db.db.Query(
		`SELECT id, name, subject, html_body, text_body, created_at
		 FROM email_templates ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.EmailTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanTemplate(row scanner) (*phishing.EmailTemplate, error) {
	var (
		t         phishing.EmailTemplate
		createdAt int64
	)
	err := row.Scan(&t.ID, &t.Name, &t.Subject, &t.HTMLBody, &t.TextBody, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.CreatedAt = time.Unix(createdAt, 0)
	return &t, nil
}
