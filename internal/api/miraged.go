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
		items[i] = miragedPhishletToResponse(p)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) pushMiragedPhishlet(w http.ResponseWriter, req *http.Request) {
	connectionID := req.PathValue("id")

	var body sdk.PushMiragedPhishletRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if err := r.Miraged.PushPhishlet(connectionID, body.YAML); err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to push phishlet", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) enableMiragedPhishlet(w http.ResponseWriter, req *http.Request) {
	connectionID := req.PathValue("id")
	name := req.PathValue("name")

	var body sdk.EnableMiragedPhishletRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	phishlet, err := r.Miraged.EnablePhishlet(connectionID, name, body.Hostname, body.DNSProvider)
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to enable phishlet", err)
		return
	}
	writeJSON(w, http.StatusOK, miragedPhishletToResponse(*phishlet))
}

func (r *Router) disableMiragedPhishlet(w http.ResponseWriter, req *http.Request) {
	connectionID := req.PathValue("id")
	name := req.PathValue("name")

	phishlet, err := r.Miraged.DisablePhishlet(connectionID, name)
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to disable phishlet", err)
		return
	}
	writeJSON(w, http.StatusOK, miragedPhishletToResponse(*phishlet))
}

func miragedPhishletToResponse(p phishing.MiragedPhishlet) sdk.MiragedPhishletResponse {
	return sdk.MiragedPhishletResponse{
		Name:        p.Name,
		Hostname:    p.Hostname,
		BaseDomain:  p.BaseDomain,
		DNSProvider: p.DNSProvider,
		SpoofURL:    p.SpoofURL,
		Enabled:     p.Enabled,
	}
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
