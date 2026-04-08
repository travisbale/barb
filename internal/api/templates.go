package api

import (
	"errors"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) listTemplates(w http.ResponseWriter, req *http.Request) {
	templates, err := r.Templates.List()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "Failed to list templates.", err)
		return
	}
	items := make([]sdk.TemplateResponse, len(templates))
	for i, t := range templates {
		items[i] = templateToResponse(t)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createTemplate(w http.ResponseWriter, req *http.Request) {
	body, ok := decodeAndValidate[sdk.CreateTemplateRequest](w, req)
	if !ok {
		return
	}

	created, err := r.Templates.Create(&phishing.EmailTemplate{
		Name:           body.Name,
		Subject:        body.Subject,
		HTMLBody:       body.HTMLBody,
		TextBody:       body.TextBody,
		EnvelopeSender: body.EnvelopeSender,
	})
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "Failed to create template.", err)
		return
	}
	writeJSON(w, http.StatusCreated, templateToResponse(created))
}

func (r *Router) getTemplate(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	template, err := r.Templates.Get(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Template not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to get template.", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, templateToResponse(template))
}

func (r *Router) updateTemplate(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	body, ok := decodeAndValidate[sdk.UpdateTemplateRequest](w, req)
	if !ok {
		return
	}

	updated, err := r.Templates.Update(id, &phishing.TemplateUpdate{
		Name:           body.Name,
		Subject:        body.Subject,
		HTMLBody:       body.HTMLBody,
		TextBody:       body.TextBody,
		EnvelopeSender: body.EnvelopeSender,
	})
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Template not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to update template.", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, templateToResponse(updated))
}

func (r *Router) previewTemplate(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	body, ok := decodeAndValidate[sdk.PreviewTemplateRequest](w, req)
	if !ok {
		return
	}

	rendered, err := r.Templates.Preview(id, phishing.PreviewData{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		URL:       body.URL,
	})
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Template not found.", err)
		} else {
			r.writeError(w, http.StatusUnprocessableEntity, "Failed to render template preview.", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, sdk.PreviewTemplateResponse{
		Subject:  rendered.Subject,
		HTMLBody: rendered.HTMLBody,
		TextBody: rendered.TextBody,
	})
}

func (r *Router) renderTemplateHTML(w http.ResponseWriter, req *http.Request) {
	body, ok := decodeAndValidate[sdk.RenderHTMLRequest](w, req)
	if !ok {
		return
	}

	rendered, err := r.Templates.RenderHTML(body.HTMLBody, phishing.PreviewData{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		URL:       body.URL,
	})
	if err != nil {
		r.writeError(w, http.StatusUnprocessableEntity, "Failed to render template.", err)
		return
	}

	writeJSON(w, http.StatusOK, sdk.RenderHTMLResponse{
		HTMLBody: rendered,
	})
}

func (r *Router) deleteTemplate(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Templates.Delete(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Template not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to delete template.", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func templateToResponse(t *phishing.EmailTemplate) sdk.TemplateResponse {
	return sdk.TemplateResponse{
		ID:             t.ID,
		Name:           t.Name,
		Subject:        t.Subject,
		HTMLBody:       t.HTMLBody,
		TextBody:       t.TextBody,
		EnvelopeSender: t.EnvelopeSender,
		CreatedAt:      t.CreatedAt,
	}
}
