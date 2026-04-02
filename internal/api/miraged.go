package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
	miragesdk "github.com/travisbale/mirage/sdk"
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
	body, ok := decodeAndValidate[sdk.EnrollMiragedRequest](w, req)
	if !ok {
		return
	}

	conn, err := r.Miraged.Enroll(body.Name, body.Address, body.SecretHostname, body.Token)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrConnectionFailed):
			r.writeError(w, http.StatusBadGateway, "could not reach miraged server", err)
		case errors.Is(err, phishing.ErrEnrollmentRejected):
			r.writeError(w, http.StatusBadGateway, "enrollment rejected — check the secret hostname and invite token", err)
		default:
			r.writeError(w, http.StatusInternalServerError, "enrollment failed", err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, miragedToResponse(conn))
}

func (r *Router) updateMiraged(w http.ResponseWriter, req *http.Request) {
	body, ok := decodeAndValidate[sdk.UpdateMiragedRequest](w, req)
	if !ok {
		return
	}

	id := req.PathValue("id")
	updated, err := r.Miraged.Rename(id, *body.Name)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "connection not found", err)
		case errors.Is(err, phishing.ErrNameRequired):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to update connection", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, miragedToResponse(updated))
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
	client, err := r.miragedClient(w, req)
	if err != nil {
		return
	}
	status, err := client.Status()
	if err != nil {
		writeJSON(w, http.StatusOK, sdk.MiragedStatusResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sdk.MiragedStatusResponse{
		Connected: true,
		Version:   status.Version,
	})
}

func (r *Router) listMiragedDNSProviders(w http.ResponseWriter, req *http.Request) {
	client, err := r.miragedClient(w, req)
	if err != nil {
		return
	}
	providers, err := client.ListDNSProviders()
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to list DNS providers from miraged", err)
		return
	}
	writeJSON(w, http.StatusOK, providers)
}

func (r *Router) pushMiragedPhishlet(w http.ResponseWriter, req *http.Request) {
	client, err := r.miragedClient(w, req)
	if err != nil {
		return
	}

	var body sdk.PushMiragedPhishletRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if _, err := client.PushPhishlet(miragesdk.PushPhishletRequest{YAML: body.YAML}); err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to push phishlet", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) enableMiragedPhishlet(w http.ResponseWriter, req *http.Request) {
	client, err := r.miragedClient(w, req)
	if err != nil {
		return
	}
	name := req.PathValue("name")

	var body sdk.EnableMiragedPhishletRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	resp, err := client.EnablePhishlet(name, miragesdk.EnablePhishletRequest{
		Hostname:    body.Hostname,
		DNSProvider: body.DNSProvider,
	})
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to enable phishlet", err)
		return
	}
	writeJSON(w, http.StatusOK, miragedPhishletToResponse(*resp))
}

func (r *Router) disableMiragedPhishlet(w http.ResponseWriter, req *http.Request) {
	client, err := r.miragedClient(w, req)
	if err != nil {
		return
	}
	name := req.PathValue("name")

	resp, err := client.DisablePhishlet(name)
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to disable phishlet", err)
		return
	}
	writeJSON(w, http.StatusOK, miragedPhishletToResponse(*resp))
}

func (r *Router) getMiragedSession(w http.ResponseWriter, req *http.Request) {
	client, err := r.miragedClient(w, req)
	if err != nil {
		return
	}
	sessionID := req.PathValue("sessionId")

	session, err := client.GetSession(sessionID)
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to get session", err)
		return
	}

	resp := sdk.MiragedSessionResponse{
		ID:           session.ID,
		Phishlet:     session.Phishlet,
		RemoteAddr:   session.RemoteAddr,
		UserAgent:    session.UserAgent,
		Username:     session.Username,
		Password:     session.Password,
		Custom:       session.Custom,
		CookieTokens: session.CookieTokens,
		BodyTokens:   session.BodyTokens,
		HTTPTokens:   session.HTTPTokens,
		StartedAt:    session.StartedAt.Format(time.RFC3339),
	}
	if session.CompletedAt != nil {
		resp.CompletedAt = session.CompletedAt.Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (r *Router) exportMiragedSessionCookies(w http.ResponseWriter, req *http.Request) {
	client, err := r.miragedClient(w, req)
	if err != nil {
		return
	}
	sessionID := req.PathValue("sessionId")

	data, err := client.ExportSessionCookies(sessionID)
	if err != nil {
		r.writeError(w, http.StatusBadGateway, "failed to export cookies", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=cookies.json")
	if _, err := w.Write(data); err != nil {
		r.Logger.Error("failed to write cookie export response", "error", err)
	}
}

// miragedClient extracts the connection ID from the request path and returns
// a mirage SDK client. It writes an error response and returns an error if
// the connection is not found or the client cannot be constructed.
func (r *Router) miragedClient(w http.ResponseWriter, req *http.Request) (*miragesdk.Client, error) {
	id := req.PathValue("id")
	client, err := r.Miraged.Client(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "connection not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to connect to miraged", err)
		}
		return nil, err
	}
	return client, nil
}

func miragedPhishletToResponse(p miragesdk.PhishletResponse) sdk.MiragedPhishletResponse {
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
