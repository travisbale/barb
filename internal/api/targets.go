package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

func (r *Router) listTargetLists(w http.ResponseWriter, req *http.Request) {
	lists, err := r.Targets.ListLists()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list target lists", err)
		return
	}

	items := make([]sdk.TargetListResponse, len(lists))
	for i, l := range lists {
		items[i] = targetListToResponse(l)
	}

	writeJSON(w, http.StatusOK, items)
}

func (r *Router) createTargetList(w http.ResponseWriter, req *http.Request) {
	var body sdk.CreateTargetListRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	list, err := r.Targets.CreateList(body.Name)
	if err != nil {
		if errors.Is(err, phishing.ErrNameRequired) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to create target list", err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, targetListToResponse(list))
}

func (r *Router) getTargetList(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	list, err := r.Targets.GetList(id)
	if err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "target list not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to get target list", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, targetListToResponse(list))
}

func (r *Router) deleteTargetList(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Targets.DeleteList(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "target list not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to delete target list", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (r *Router) listTargets(w http.ResponseWriter, req *http.Request) {
	listID := req.PathValue("id")
	targets, err := r.Targets.ListTargets(listID)
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to list targets", err)
		return
	}

	items := make([]sdk.TargetResponse, len(targets))
	for i, t := range targets {
		items[i] = targetToResponse(t)
	}

	writeJSON(w, http.StatusOK, items)
}

func (r *Router) addTarget(w http.ResponseWriter, req *http.Request) {
	listID := req.PathValue("id")

	var body sdk.AddTargetRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		r.writeError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	target := &phishing.Target{
		Email:      body.Email,
		FirstName:  body.FirstName,
		LastName:   body.LastName,
		Department: body.Department,
		Position:   body.Position,
	}

	if err := r.Targets.AddTarget(listID, target); err != nil {
		if errors.Is(err, phishing.ErrEmailRequired) {
			r.writeError(w, http.StatusUnprocessableEntity, err.Error(), nil)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to add target", err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, targetToResponse(target))
}

func (r *Router) importTargets(w http.ResponseWriter, req *http.Request) {
	listID := req.PathValue("id")

	file, _, err := req.FormFile("file")
	if err != nil {
		r.writeError(w, http.StatusBadRequest, "missing CSV file", nil)
		return
	}
	defer file.Close()

	count, err := r.Targets.ImportCSV(listID, file)
	if err != nil {
		r.writeError(w, http.StatusUnprocessableEntity, err.Error(), err)
		return
	}

	writeJSON(w, http.StatusOK, sdk.ImportTargetsResponse{Imported: count})
}

func (r *Router) deleteTarget(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	if err := r.Targets.DeleteTarget(id); err != nil {
		if errors.Is(err, phishing.ErrNotFound) {
			r.writeError(w, http.StatusNotFound, "target not found", err)
		} else {
			r.writeError(w, http.StatusInternalServerError, "failed to delete target", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func targetListToResponse(l *phishing.TargetList) sdk.TargetListResponse {
	return sdk.TargetListResponse{
		ID:        l.ID,
		Name:      l.Name,
		CreatedAt: l.CreatedAt,
	}
}

func targetToResponse(t *phishing.Target) sdk.TargetResponse {
	return sdk.TargetResponse{
		ID:         t.ID,
		ListID:     t.ListID,
		Email:      t.Email,
		FirstName:  t.FirstName,
		LastName:   t.LastName,
		Department: t.Department,
		Position:   t.Position,
	}
}
