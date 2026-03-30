package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) listMiraged(w http.ResponseWriter, req *http.Request) {
	connections, err := r.Miraged.List()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list connections", err)
		return
	}
	items := make([]sdk.MiragedResponse, len(connections))
	for i, c := range connections {
		items[i] = miragedToResponse(c)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) enrollMiraged(w http.ResponseWriter, req *http.Request) {
	var body sdk.EnrollMiragedRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	conn, err := r.Miraged.Enroll(body.Name, body.Address, body.SecretHostname, body.Token)
	if err != nil {
		if isValidationError(err) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "enrollment failed", err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, miragedToResponse(conn))
}

func (r *Router) deleteMiraged(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Miraged.Delete(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "connection not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to delete connection", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) testMiraged(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	status, err := r.Miraged.TestConnection(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "connection not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to test connection", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, sdk.MiragedStatusResponse{
		Connected: status.Connected,
		Version:   status.Version,
		Error:     status.Error,
	})
}

func (r *Router) listMiragedPhishlets(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	phishlets, err := r.Miraged.ListPhishlets(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "connection not found", err)
		} else {
			r.writeError(w, http.StatusBadGateway, "failed to list phishlets from miraged", err)
		}
		return
	}

	items := make([]sdk.MiragedPhishletResponse, len(phishlets))
	for i, p := range phishlets {
		items[i] = sdk.MiragedPhishletResponse{
			Name:     p.Name,
			Hostname: p.Hostname,
			Enabled:  p.Enabled,
		}
	}
	writeJSON(w, http.StatusOK, items)
}

func miragedToResponse(c *phishing.MiragedConnection) sdk.MiragedResponse {
	return sdk.MiragedResponse{
		ID:             c.ID,
		Name:           c.Name,
		Address:        c.Address,
		SecretHostname: c.SecretHostname,
		CreatedAt:      c.CreatedAt,
	}
}
