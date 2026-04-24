package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	"github.com/travisbale/barb/internal/phishing"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// DB wraps a SQLite database connection.
type DB struct {
	*sql.DB
}

// Open creates or opens a SQLite database and applies the schema.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("applying schema: %w", err)
	}

	return &DB{DB: db}, nil
}

// WithTx executes fn within a transaction. The transaction is committed if fn
// returns nil; otherwise it is rolled back.
func (d *DB) WithTx(fn func(tx *sql.Tx) error) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// scanner is satisfied by *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func isConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// requireOneRow returns ErrNotFound if no rows were affected.
func requireOneRow(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return phishing.ErrNotFound
	}
	return nil
}
