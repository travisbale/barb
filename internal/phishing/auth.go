package phishing

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents an operator account.
type User struct {
	ID                     string
	Username               string
	PasswordHash           string
	PasswordChangeRequired bool
	CreatedAt              time.Time
}

// Session represents an authenticated session.
type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

type authStore interface {
	CreateUser(user *User) error
	GetUserByID(id string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
	CreateSession(session *Session) error
	GetSession(token string) (*Session, error)
	DeleteSession(token string) error
	DeleteExpiredSessions() error
}

// AuthService manages authentication and sessions.
type AuthService struct {
	Store         authStore
	Logger        *slog.Logger
	SessionMaxAge time.Duration
}

const defaultSessionMaxAge = 7 * 24 * time.Hour

func (s *AuthService) sessionMaxAge() time.Duration {
	if s.SessionMaxAge > 0 {
		return s.SessionMaxAge
	}
	return defaultSessionMaxAge
}

// EnsureAdmin guarantees an admin user exists with temporary credentials
// logged to the console. If the admin hasn't completed their initial login
// yet, the account is recreated with a fresh password on each startup.
func (s *AuthService) EnsureAdmin() error {
	admin, err := s.Store.GetUserByUsername("admin")
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("looking up admin user: %w", err)
	}

	if admin != nil {
		// Admin has completed their initial login.
		if !admin.PasswordChangeRequired {
			return nil
		}

		// Delete the stale admin so we can recreate with a fresh password.
		if err := s.Store.DeleteUser(admin.ID); err != nil {
			return fmt.Errorf("deleting stale admin user: %w", err)
		}
	}

	password, err := generateRandomPassword(16)
	if err != nil {
		return fmt.Errorf("generating admin password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}

	user := &User{
		ID:                     uuid.New().String(),
		Username:               "admin",
		PasswordHash:           string(hash),
		PasswordChangeRequired: true,
		CreatedAt:              time.Now(),
	}
	if err := s.Store.CreateUser(user); err != nil {
		return fmt.Errorf("creating admin user: %w", err)
	}

	s.Logger.Info("admin account created", "username", "admin", "password", password)
	return nil
}

// Login validates credentials and creates a session. Returns the session token.
func (s *AuthService) Login(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.Store.GetUserByUsername(username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := generateSessionToken()
	if err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}

	session := &Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.sessionMaxAge()),
	}

	if err := s.Store.CreateSession(session); err != nil {
		return "", err
	}

	if err := s.Store.DeleteExpiredSessions(); err != nil {
		s.Logger.Warn("failed to clean up expired sessions", "error", err)
	}

	return token, nil
}

// Logout deletes the session.
func (s *AuthService) Logout(token string) error {
	return s.Store.DeleteSession(token)
}

// CurrentUser returns the user for a session token, or ErrNotFound if the
// session is invalid or expired.
func (s *AuthService) CurrentUser(token string) (*User, error) {
	if token == "" {
		return nil, ErrNotFound
	}

	session, err := s.Store.GetSession(token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.Store.DeleteSession(token)
		return nil, ErrNotFound
	}

	// Look up the user by iterating — we only have GetUserByUsername, not GetUserByID.
	// For a single-user system this is fine. If we add multi-user, add GetUserByID.
	return s.Store.GetUserByID(session.UserID)
}

// ChangePassword validates the current password and sets a new one.
func (s *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	if newPassword == "" {
		return ErrPasswordRequired
	}
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	user, err := s.Store.GetUserByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	user.PasswordHash = string(hash)
	user.PasswordChangeRequired = false
	return s.Store.UpdateUser(user)
}

func generateRandomPassword(length int) (string, error) {
	numBytes := (length + 1) / 2
	randomBytes := make([]byte, numBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes)[:length], nil
}

func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
