package api

import (
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
	body, ok := decodeAndValidate[sdk.CreateSMTPProfileRequest](w, req)
	if !ok {
		return
	}

	created, err := r.SMTP.CreateProfile(req.Context(), &phishing.SMTPProfile{
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
		if errors.Is(err, phishing.ErrSMTPConnectionFailed) {
			r.writeError(w, http.StatusUnprocessableEntity, "Could not connect to the SMTP server. Please check the host, port, and credentials.", err)
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

	body, ok := decodeAndValidate[sdk.UpdateSMTPProfileRequest](w, req)
	if !ok {
		return
	}

	updated, err := r.SMTP.UpdateProfile(req.Context(), id, &phishing.SMTPProfileUpdate{
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
		case errors.Is(err, phishing.ErrSMTPConnectionFailed):
			r.writeError(w, http.StatusUnprocessableEntity, "Could not connect to the SMTP server. Please check the host, port, and credentials.", err)
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
