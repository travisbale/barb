package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/sdk"
)

func (r *Router) listPhishlets(w http.ResponseWriter, req *http.Request) {
	phishlets, err := r.Phishlets.List()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list phishlets", err)
		return
	}
	items := make([]sdk.PhishletResponse, len(phishlets))
	for i, p := range phishlets {
		items[i] = phishletToResponse(p)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createPhishlet(w http.ResponseWriter, req *http.Request) {
	var body sdk.CreatePhishletRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	created, err := r.Phishlets.Create(body.YAML)
	if err != nil {
		if isValidationError(err) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to create phishlet", err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, phishletToResponse(created))
}

func (r *Router) getPhishlet(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	p, err := r.Phishlets.Get(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "phishlet not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to get phishlet", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, phishletToResponse(p))
}

func (r *Router) updatePhishlet(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	var body sdk.UpdatePhishletRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	updated, err := r.Phishlets.Update(id, body.YAML)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "phishlet not found", err)
		case isValidationError(err):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to update phishlet", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, phishletToResponse(updated))
}

func (r *Router) deletePhishlet(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Phishlets.Delete(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "phishlet not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to delete phishlet", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func phishletToResponse(p *phishing.Phishlet) sdk.PhishletResponse {
	return sdk.PhishletResponse{
		ID:        p.ID,
		Name:      p.Name,
		YAML:      p.YAML,
		CreatedAt: p.CreatedAt,
	}
}
