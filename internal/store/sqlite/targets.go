package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/travisbale/barb/internal/phishing"
)

type Targets struct{ db *DB }

func NewTargetStore(db *DB) *Targets { return &Targets{db: db} }

func (s *Targets) CreateList(list *phishing.TargetList) error {
	_, err := s.db.Exec(
		`INSERT INTO target_lists (id, name, created_at) VALUES (?, ?, ?)`,
		list.ID, list.Name, list.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Targets) GetList(id string) (*phishing.TargetList, error) {
	row := s.db.QueryRow(`SELECT id, name, created_at FROM target_lists WHERE id = ?`, id)
	return scanTargetList(row)
}

func (s *Targets) DeleteList(id string) error {
	res, err := s.db.Exec(`DELETE FROM target_lists WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Targets) ListLists() ([]*phishing.TargetList, error) {
	rows, err := s.db.Query(`SELECT id, name, created_at FROM target_lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.TargetList
	for rows.Next() {
		list, err := scanTargetList(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, list)
	}
	return out, rows.Err()
}

func (s *Targets) CreateTarget(target *phishing.Target) error {
	_, err := s.db.Exec(
		`INSERT INTO targets (id, list_id, email, first_name, last_name, department, position)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		target.ID, target.ListID, target.Email, target.FirstName, target.LastName,
		target.Department, target.Position,
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Targets) CreateTargets(targets []*phishing.Target) error {
	return s.db.WithTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(
			`INSERT INTO targets (id, list_id, email, first_name, last_name, department, position)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, t := range targets {
			if _, err := stmt.Exec(t.ID, t.ListID, t.Email, t.FirstName, t.LastName, t.Department, t.Position); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Targets) ListTargets(listID string) ([]*phishing.Target, error) {
	rows, err := s.db.Query(
		`SELECT id, list_id, email, first_name, last_name, department, position
		 FROM targets WHERE list_id = ? ORDER BY email ASC`, listID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*phishing.Target
	for rows.Next() {
		var t phishing.Target
		if err := rows.Scan(&t.ID, &t.ListID, &t.Email, &t.FirstName, &t.LastName, &t.Department, &t.Position); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

func (s *Targets) DeleteTarget(id string) error {
	res, err := s.db.Exec(`DELETE FROM targets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func scanTargetList(row scanner) (*phishing.TargetList, error) {
	var (
		list      phishing.TargetList
		createdAt int64
	)
	err := row.Scan(&list.ID, &list.Name, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	list.CreatedAt = time.Unix(createdAt, 0)
	return &list, nil
}
