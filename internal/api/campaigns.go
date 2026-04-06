package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) listCampaigns(w http.ResponseWriter, req *http.Request) {
	campaigns, err := r.Campaigns.List()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "Failed to list campaigns.", err)
		return
	}
	items := make([]sdk.CampaignResponse, len(campaigns))
	for i, c := range campaigns {
		items[i] = campaignToResponse(c)
	}
	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createCampaign(w http.ResponseWriter, req *http.Request) {
	body, ok := decodeAndValidate[sdk.CreateCampaignRequest](w, req)
	if !ok {
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
		if isReferenceError(err) {
			r.writeError(w, http.StatusUnprocessableEntity, referenceErrorMessage(err), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to create campaign.", err)
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
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to get campaign.", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, campaignToResponse(campaign))
}

func (r *Router) updateCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	body, ok := decodeAndValidate[sdk.UpdateCampaignRequest](w, req)
	if !ok {
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
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		case errors.Is(err, phishing.ErrCampaignNotDraft):
			r.writeError(w, http.StatusUnprocessableEntity, "Campaign can only be edited while in draft status.", nil)
		case isReferenceError(err):
			r.writeError(w, http.StatusUnprocessableEntity, referenceErrorMessage(err), nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "Failed to update campaign.", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, campaignToResponse(updated))
}

func (r *Router) deleteCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Campaigns.Delete(id); err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		case errors.Is(err, phishing.ErrCampaignActive):
			r.writeError(w, http.StatusUnprocessableEntity, "Active campaigns cannot be deleted. Complete or cancel the campaign first.", err)
		default:
			r.writeError(w, http.StatusInternalServerError, "Failed to delete campaign.", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) listCampaignResults(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	results, err := r.Campaigns.Results(id)
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "Failed to list results.", err)
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

	err := r.Campaigns.Start(req.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		case errors.Is(err, phishing.ErrCampaignNotDraft):
			r.writeError(w, http.StatusUnprocessableEntity, "Campaign can only be started from draft status.", nil)
		case isReferenceError(err):
			r.writeError(w, http.StatusUnprocessableEntity, referenceErrorMessage(err), nil)
		case errors.Is(err, phishing.ErrSMTPConnectionFailed):
			r.writeError(w, http.StatusUnprocessableEntity, "Could not connect to the SMTP server. Please check the host, port, and credentials.", err)
		default:
			r.writeError(w, http.StatusInternalServerError, "Failed to start campaign.", err)
		}
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "starting"})
}

func (r *Router) completeCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	err := r.Campaigns.Complete(id)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrCampaignNotRunning):
			r.writeError(w, http.StatusUnprocessableEntity, "Campaign is not currently running.", nil)
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		default:
			r.writeError(w, http.StatusInternalServerError, "Failed to complete campaign.", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (r *Router) cancelCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	err := r.Campaigns.Cancel(id)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrCampaignNotRunning):
			r.writeError(w, http.StatusUnprocessableEntity, "Campaign is not currently running.", nil)
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		default:
			r.writeError(w, http.StatusInternalServerError, "Failed to cancel campaign.", err)
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
		r.writeError(w, http.StatusBadRequest, "Invalid request body.", nil)
		return
	}

	err := r.Campaigns.SendTestEmail(id, body.Email)
	if err != nil {
		switch {
		case errors.Is(err, phishing.ErrNotFound):
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		case errors.Is(err, phishing.ErrEmailRequired):
			r.writeError(w, http.StatusUnprocessableEntity, "Email address is required.", nil)
		default:
			r.writeError(w, http.StatusInternalServerError, "Failed to send test email.", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (r *Router) streamCampaign(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")

	// Verify the campaign exists.
	if _, err := r.Campaigns.Get(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "Campaign not found.", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "Failed to get campaign.", err)
		}
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		r.writeError(w, http.StatusInternalServerError, "Streaming not supported.", nil)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := r.Campaigns.Bus.Subscribe(id)
	defer r.Campaigns.Bus.Unsubscribe(id, ch)

	// Send current state so the client doesn't miss events that fired
	// before the subscription was established.
	campaign, _ := r.Campaigns.Get(id)
	if campaign != nil {
		statusData, _ := json.Marshal(sdk.CampaignEvent{
			Type:       sdk.EventCampaignStatus,
			CampaignID: id,
			Status:     string(campaign.Status),
		})
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", sdk.EventCampaignStatus, statusData)
		flusher.Flush()
	}
	results, _ := r.Campaigns.Results(id)
	for _, result := range results {
		if result.Status == "pending" {
			continue
		}
		resp := resultToResponse(result)
		resultData, _ := json.Marshal(sdk.CampaignEvent{
			Type:       sdk.EventResultUpdated,
			CampaignID: id,
			Result:     &resp,
		})
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", sdk.EventResultUpdated, resultData)
	}
	flusher.Flush()

	for {
		select {
		case <-req.Context().Done():
			return
		case event := <-ch:
			sseEvent := sdk.CampaignEvent{
				Type:       event.Type,
				CampaignID: event.CampaignID,
				Status:     event.Status,
			}
			if event.Result != nil {
				r := resultToResponse(event.Result)
				sseEvent.Result = &r
			}
			data, err := json.Marshal(sseEvent)
			if err != nil {
				r.Logger.Error("failed to marshal SSE event", "error", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}

func isReferenceError(err error) bool {
	return errors.Is(err, phishing.ErrTemplateNotFound) ||
		errors.Is(err, phishing.ErrSMTPProfileNotFound) ||
		errors.Is(err, phishing.ErrTargetListNotFound)
}

func referenceErrorMessage(err error) string {
	switch {
	case errors.Is(err, phishing.ErrTemplateNotFound):
		return "The selected template no longer exists."
	case errors.Is(err, phishing.ErrSMTPProfileNotFound):
		return "The selected SMTP profile no longer exists."
	case errors.Is(err, phishing.ErrTargetListNotFound):
		return "The selected target list no longer exists."
	default:
		return "A referenced resource no longer exists."
	}
}
