package api

import (
	"errors"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) listPhishlets(w http.ResponseWriter, req *http.Request) {
	phishlets, err := r.Phishlets.List()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "Failed to list phishlets.", err)
		return
	}
	items := make([]sdk.PhishletResponse, len(phishlets))
	for i, p := range phishlets {
		items[i] = phishletToResponse(p)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createPhishlet(w http.ResponseWriter, req *http.Request) {
	body, ok := decodeAndValidate[sdk.CreatePhishletRequest](w, req)
	if !ok {
		return
	}

	created, err := r.Phishlets.Create(body.YAML)
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "Failed to create phishlet.", err)
		return
	}
	writeJSON(w, http.StatusCreated, phishletToResponse(created))
}

func (r *Router) getPhishlet(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	p, err := r.Phishlets.Get(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Phishlet not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to get phishlet.", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, phishletToResponse(p))
}

func (r *Router) updatePhishlet(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	body, ok := decodeAndValidate[sdk.UpdatePhishletRequest](w, req)
	if !ok {
		return
	}

	updated, err := r.Phishlets.Update(id, body.YAML)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Phishlet not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to update phishlet.", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, phishletToResponse(updated))
}

func (r *Router) deletePhishlet(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Phishlets.Delete(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Phishlet not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to delete phishlet.", err)
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
