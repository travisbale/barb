package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) listCampaigns(w http.ResponseWriter, req *http.Request) {
	campaigns, err := r.Campaigns.List()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list campaigns", err)
		return
	}
	items := make([]sdk.CampaignResponse, len(campaigns))
	for i, c := range campaigns {
		items[i] = campaignToResponse(c)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createCampaign(w http.ResponseWriter, req *http.Request) {
	var body sdk.CreateCampaignRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	campaign := &phishing.Campaign{
		Name:          body.Name,
		TemplateID:    body.TemplateID,
		SMTPProfileID: body.SMTPProfileID,
		TargetListID:  body.TargetListID,
		MiragedID:     body.MiragedID,
		Phishlet:      body.Phishlet,
		RedirectURL:   body.RedirectURL,
		LureURL:       body.LureURL,
		SendRate:      body.SendRate,
	}

	created, err := r.Campaigns.Create(campaign)
	if err != nil {
		if isValidationError(err) || isReferenceError(err) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to create campaign", err)
		}
		return
	}
	writeJSON(w, http.StatusCreated, campaignToResponse(created))
}

func (r *Router) getCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	campaign, err := r.Campaigns.Get(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "campaign not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to get campaign", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, campaignToResponse(campaign))
}

func (r *Router) updateCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	var body sdk.UpdateCampaignRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	updated, err := r.Campaigns.Update(id, &phishing.CampaignUpdate{
		Name:          body.Name,
		TemplateID:    body.TemplateID,
		SMTPProfileID: body.SMTPProfileID,
		TargetListID:  body.TargetListID,
		MiragedID:     body.MiragedID,
		Phishlet:      body.Phishlet,
		RedirectURL:   body.RedirectURL,
		SendRate:      body.SendRate,
	})
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "campaign not found", err)
		case errors.Is(err, phishing.ErrCampaignNotDraft):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		case isValidationError(err) || isReferenceError(err):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to update campaign", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, campaignToResponse(updated))
}

func (r *Router) deleteCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Campaigns.Delete(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "campaign not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to delete campaign", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) listCampaignResults(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	results, err := r.Campaigns.Results(id)
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list results", err)
		return
	}
	items := make([]sdk.CampaignResultResponse, len(results))
	for i, res := range results {
		items[i] = resultToResponse(res)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) startCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	err := r.Campaigns.Start(id)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "campaign not found", err)
		case errors.Is(err, phishing.ErrCampaignNotDraft):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		case isReferenceError(err):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to start campaign", err)
		}
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "starting"})
}

func (r *Router) cancelCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	err := r.Campaigns.Cancel(id)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrCampaignNotRunning):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "campaign not found", err)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to cancel campaign", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func campaignToResponse(c *phishing.Campaign) sdk.CampaignResponse {
	return sdk.CampaignResponse{
		ID:            c.ID,
		Name:          c.Name,
		Status:        string(c.Status),
		TemplateID:    c.TemplateID,
		SMTPProfileID: c.SMTPProfileID,
		TargetListID:  c.TargetListID,
		MiragedID:     c.MiragedID,
		Phishlet:      c.Phishlet,
		RedirectURL:   c.RedirectURL,
		LureURL:       c.LureURL,
		SendRate:      c.SendRate,
		CreatedAt:     c.CreatedAt,
		StartedAt:     c.StartedAt,
		CompletedAt:   c.CompletedAt,
	}
}

func resultToResponse(r *phishing.CampaignResult) sdk.CampaignResultResponse {
	return sdk.CampaignResultResponse{
		ID:         r.ID,
		CampaignID: r.CampaignID,
		TargetID:   r.TargetID,
		Email:      r.Email,
		Status:     r.Status,
		SentAt:     r.SentAt,
		ClickedAt:  r.ClickedAt,
		CapturedAt: r.CapturedAt,
		SessionID:  r.SessionID,
	}
}

func (r *Router) sendTestEmail(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	var body sdk.SendTestEmailRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	err := r.Campaigns.SendTestEmail(id, body.Email)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "campaign not found", err)
		case errors.Is(err, phishing.ErrEmailRequired):
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "failed to send test email", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func isReferenceError(err error) bool {
	return errors.Is(err, phishing.ErrTemplateNotFound) ||
		errors.Is(err, phishing.ErrSMTPProfileNotFound) ||
		errors.Is(err, phishing.ErrTargetListNotFound)
}
