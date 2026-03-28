package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/sdk"
)

func (r *Router) listTemplates(w http.ResponseWriter, req *http.Request) {
	templates, err := r.Templates.List()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list templates", err)
		return
	}
	items := make([]sdk.TemplateResponse, len(templates))
	for i, t := range templates {
		items[i] = templateToResponse(t)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createTemplate(w http.ResponseWriter, req *http.Request) {
	var body sdk.CreateTemplateRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	template := &phishing.EmailTemplate{
		Name:     body.Name,
		Subject:  body.Subject,
		HTMLBody: body.HTMLBody,
		TextBody: body.TextBody,
	}

	created, err := r.Templates.Create(template)
	if err != nil {
		if isValidationError(err) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to create template", err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, templateToResponse(created))
}

func (r *Router) getTemplate(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	template, err := r.Templates.Get(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "template not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to get template", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, templateToResponse(template))
}

func (r *Router) updateTemplate(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	var body sdk.UpdateTemplateRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	updated, err := r.Templates.Update(id, &phishing.EmailTemplate{
		Name:     body.Name,
		Subject:  body.Subject,
		HTMLBody: body.HTMLBody,
		TextBody: body.TextBody,
	})
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "template not found", err)
		case isValidationError(err):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to update template", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, templateToResponse(updated))
}

func (r *Router) deleteTemplate(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Templates.Delete(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "template not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to delete template", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func templateToResponse(t *phishing.EmailTemplate) sdk.TemplateResponse {
	return sdk.TemplateResponse{
		ID:        t.ID,
		Name:      t.Name,
		Subject:   t.Subject,
		HTMLBody:  t.HTMLBody,
		TextBody:  t.TextBody,
		CreatedAt: t.CreatedAt,
	}
}

// isValidationError returns true for domain validation errors that should
// be returned as 422 to the client.
func isValidationError(err error) bool {
	return errors.Is(err, phishing.ErrNameRequired) ||
		errors.Is(err, phishing.ErrEmailRequired) ||
		errors.Is(err, phishing.ErrHostRequired) ||
		errors.Is(err, phishing.ErrFromAddrRequired) ||
		errors.Is(err, phishing.ErrSubjectRequired) ||
		errors.Is(err, phishing.ErrBodyRequired)
}
