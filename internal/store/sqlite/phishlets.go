package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/travisbale/mirador/internal/phishing"
)

type Phishlets struct{ db *DB }

func NewPhishletStore(db *DB) *Phishlets { return &Phishlets{db: db} }

func (s *Phishlets) CreatePhishlet(p *phishing.Phishlet) error {
	_, err := s.db.db.Exec(
		`INSERT INTO phishlets (id, name, yaml, created_at) VALUES (?, ?, ?, ?)`,
		p.ID, p.Name, p.YAML, p.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Phishlets) GetPhishlet(id string) (*phishing.Phishlet, error) {
	row := s.db.db.QueryRow(`SELECT id, name, yaml, created_at FROM phishlets WHERE id = ?`, id)
	return scanPhishlet(row)
}

func (s *Phishlets) GetPhishletByName(name string) (*phishing.Phishlet, error) {
	row := s.db.db.QueryRow(`SELECT id, name, yaml, created_at FROM phishlets WHERE name = ?`, name)
	return scanPhishlet(row)
}

func (s *Phishlets) UpdatePhishlet(p *phishing.Phishlet) error {
	res, err := s.db.db.Exec(
		`UPDATE phishlets SET name = ?, yaml = ? WHERE id = ?`,
		p.Name, p.YAML, p.ID,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Phishlets) DeletePhishlet(id string) error {
	res, err := s.db.db.Exec(`DELETE FROM phishlets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Phishlets) ListPhishlets() ([]*phishing.Phishlet, error) {
	rows, err := s.db.db.Query(`SELECT id, name, yaml, created_at FROM phishlets ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.Phishlet
	for rows.Next() {
		p, err := scanPhishlet(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanPhishlet(row scanner) (*phishing.Phishlet, error) {
	var (
		p         phishing.Phishlet
		createdAt int64
	)
	err := row.Scan(&p.ID, &p.Name, &p.YAML, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt = time.Unix(createdAt, 0)
	return &p, nil
}
