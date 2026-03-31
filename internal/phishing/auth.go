package phishing

import (
	"crypto/rand"
	"encoding/hex"
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
	CreateSession(session *Session) error
	GetSession(token string) (*Session, error)
	DeleteSession(token string) error
	DeleteExpiredSessions() error
	CountUsers() (int, error)
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

// EnsureAdmin creates an admin user if none exist and logs the temporary
// credentials to the console.
func (s *AuthService) EnsureAdmin() {
	count, err := s.Store.CountUsers()
	if err != nil {
		s.Logger.Error("failed to check for existing users", "error", err)
		return
	}
	if count > 0 {
		return
	}

	password, err := generateRandomPassword(16)
	if err != nil {
		s.Logger.Error("failed to generate admin password", "error", err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.Logger.Error("failed to hash admin password", "error", err)
		return
	}

	user := &User{
		ID:                     uuid.New().String(),
		Username:               "admin",
		PasswordHash:           string(hash),
		PasswordChangeRequired: true,
		CreatedAt:              time.Now(),
	}

	if err := s.Store.CreateUser(user); err != nil {
		s.Logger.Error("failed to create admin user", "error", err)
		return
	}

	s.Logger.Info("admin account created", "username", "admin", "password", password)
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
