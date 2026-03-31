package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) listSMTPProfiles(w http.ResponseWriter, req *http.Request) {
	profiles, err := r.SMTP.ListProfiles()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list SMTP profiles", err)
		return
	}

	items := make([]sdk.SMTPProfileResponse, len(profiles))
	for i, p := range profiles {
		items[i] = smtpProfileToResponse(p)
	}

	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createSMTPProfile(w http.ResponseWriter, req *http.Request) {
	var body sdk.CreateSMTPProfileRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	profile := &phishing.SMTPProfile{
		Name:          body.Name,
		Host:          body.Host,
		Port:          body.Port,
		Username:      body.Username,
		Password:      body.Password,
		FromAddr:      body.FromAddr,
		FromName:      body.FromName,
		CustomHeaders: body.CustomHeaders,
	}

	created, err := r.SMTP.CreateProfile(profile)
	if err != nil {
		if isValidationError(err) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to create SMTP profile", err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, smtpProfileToResponse(created))
}

func (r *Router) getSMTPProfile(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	profile, err := r.SMTP.GetProfile(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "SMTP profile not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to get SMTP profile", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, smtpProfileToResponse(profile))
}

func (r *Router) updateSMTPProfile(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	var body sdk.UpdateSMTPProfileRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	updated, err := r.SMTP.UpdateProfile(id, &phishing.SMTPProfileUpdate{
		Name:          body.Name,
		Host:          body.Host,
		Port:          body.Port,
		Username:      body.Username,
		Password:      body.Password,
		FromAddr:      body.FromAddr,
		FromName:      body.FromName,
		CustomHeaders: body.CustomHeaders,
	})
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "SMTP profile not found", err)
		case isValidationError(err):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to update SMTP profile", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, smtpProfileToResponse(updated))
}

func (r *Router) deleteSMTPProfile(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.SMTP.DeleteProfile(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "SMTP profile not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to delete SMTP profile", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func smtpProfileToResponse(p *phishing.SMTPProfile) sdk.SMTPProfileResponse {
	return sdk.SMTPProfileResponse{
		ID:            p.ID,
		Name:          p.Name,
		Host:          p.Host,
		Port:          p.Port,
		Username:      p.Username,
		FromAddr:      p.FromAddr,
		FromName:      p.FromName,
		CustomHeaders: p.CustomHeaders,
		CreatedAt:     p.CreatedAt,
	}
}
