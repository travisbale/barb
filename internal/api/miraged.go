package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/sdk"
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

func (r *Router) createMiraged(w http.ResponseWriter, req *http.Request) {
	var body sdk.CreateMiragedRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	conn := &phishing.MiragedConnection{
		Name:           body.Name,
		Address:        body.Address,
		SecretHostname: body.SecretHostname,
		CertPEM:        []byte(body.CertPEM),
		KeyPEM:         []byte(body.KeyPEM),
		CACertPEM:      []byte(body.CACertPEM),
	}

	created, err := r.Miraged.Create(conn)
	if err != nil {
		if isValidationError(err) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to create connection", err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, miragedToResponse(created))
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
	err := r.Miraged.TestConnection(id)
	if errors.Is(err, phishing.ErrNotFound) {
		r.writeError(w, http.StatusNotFound, "connection not found", err)
		return
	}

	resp := sdk.MiragedStatusResponse{Connected: err == nil}
	if err != nil {
		resp.Error = err.Error()
	} else {
		client, _ := r.Miraged.Client(id)
		if status, err := client.Status(); err == nil {
			resp.Version = status.Version
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) listMiragedPhishlets(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	client, err := r.Miraged.Client(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "connection not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to connect", err)
		}
		return
	}

	phishlets, err := client.ListPhishlets()
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to list phishlets from miraged", err)
		return
	}

	items := make([]sdk.MiragedPhishletResponse, len(phishlets.Items))
	for i, p := range phishlets.Items {
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
