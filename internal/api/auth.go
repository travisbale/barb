package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) loginHandler(w http.ResponseWriter, req *http.Request) {
	var body sdk.LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	token, err := r.Auth.Login(body.Username, body.Password)
	if err != nil {
		r.writeError(w, http.StatusUnauthorized, "invalid username or password", nil)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) logoutHandler(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie("session")
	if err == nil {
		if logoutErr := r.Auth.Logout(cookie.Value); logoutErr != nil {
			r.Logger.Error("failed to delete session on logout", "error", logoutErr)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) meHandler(w http.ResponseWriter, req *http.Request) {
	user := userFromContext(req.Context())
	if user == nil {
		r.writeError(w, http.StatusUnauthorized, "not authenticated", nil)
		return
	}

	writeJSON(w, http.StatusOK, sdk.MeResponse{
		Username:               user.Username,
		PasswordChangeRequired: user.PasswordChangeRequired,
	})
}

func (r *Router) changePasswordHandler(w http.ResponseWriter, req *http.Request) {
	user := userFromContext(req.Context())
	if user == nil {
		r.writeError(w, http.StatusUnauthorized, "not authenticated", nil)
		return
	}

	var body sdk.ChangePasswordRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	err := r.Auth.ChangePassword(user.ID, body.CurrentPassword, body.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrInvalidCredentials):
			r.writeError(w, http.StatusUnauthorized, "current password is incorrect", nil)
		case errors.Is(err, phishing.ErrPasswordRequired), errors.Is(err, phishing.ErrPasswordTooShort):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to change password", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
