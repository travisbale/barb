package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/travisbale/barb/internal/phishing"
)

type Auth struct{ db *DB }

func NewAuthStore(db *DB) *Auth { return &Auth{db: db} }

func (s *Auth) CreateUser(user *phishing.User) error {
	_, err := s.db.Exec(
		`INSERT INTO users (id, username, password_hash, password_change_required, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Username, user.PasswordHash, user.PasswordChangeRequired, user.CreatedAt.Unix(),
	)
	if isConflict(err) {
		return phishing.ErrConflict
	}
	return err
}

func (s *Auth) GetUserByID(id string) (*phishing.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, password_change_required, created_at
		 FROM users WHERE id = ?`, id,
	)
	return scanUser(row)
}

func (s *Auth) GetUserByUsername(username string) (*phishing.User, error) {
	row := s.db.QueryRow(
		`SELECT id, username, password_hash, password_change_required, created_at
		 FROM users WHERE username = ?`, username,
	)
	return scanUser(row)
}

func scanUser(row scanner) (*phishing.User, error) {
	var (
		user      phishing.User
		pwdChange int
		createdAt int64
	)
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &pwdChange, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	user.PasswordChangeRequired = pwdChange != 0
	user.CreatedAt = time.Unix(createdAt, 0)
	return &user, nil
}

func (s *Auth) UpdateUser(user *phishing.User) error {
	res, err := s.db.Exec(
		`UPDATE users SET password_hash = ?, password_change_required = ? WHERE id = ?`,
		user.PasswordHash, user.PasswordChangeRequired, user.ID,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (s *Auth) CreateSession(session *phishing.Session) error {
	_, err := s.db.Exec(
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		session.Token, session.UserID, session.ExpiresAt.Unix(),
	)
	return err
}

func (s *Auth) GetSession(token string) (*phishing.Session, error) {
	row := s.db.QueryRow(
		`SELECT token, user_id, expires_at FROM sessions WHERE token = ?`, token,
	)
	var (
		session   phishing.Session
		expiresAt int64
	)
	err := row.Scan(&session.Token, &session.UserID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, phishing.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	session.ExpiresAt = time.Unix(expiresAt, 0)
	return &session, nil
}

func (s *Auth) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (s *Auth) DeleteExpiredSessions() error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().Unix())
	return err
}

func (s *Auth) DeleteUser(id string) error {
	res, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}
