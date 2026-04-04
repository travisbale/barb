package api

import (
	"net/http"

	"github.com/travisbale/barb/sdk"
)

func (r *Router) getDashboard(w http.ResponseWriter, req *http.Request) {
	stats, err := r.Dashboard.Stats()
	if err != nil {
		r.writeError(w, http.StatusInternalServerError, "failed to load dashboard", err)
		return
	}

	active := make([]sdk.ActiveCampaignInfo, len(stats.ActiveCampaigns))
	for i, a := range stats.ActiveCampaigns {
		active[i] = sdk.ActiveCampaignInfo{
			ID:        a.ID,
			Name:      a.Name,
			Sent:      a.Sent,
			Failed:    a.Failed,
			Captured:  a.Captured,
			Completed: a.Completed,
			Total:     a.Total,
		}
	}

	captures := make([]sdk.RecentCapture, len(stats.RecentCaptures))
	for i, c := range stats.RecentCaptures {
		captures[i] = sdk.RecentCapture{
			Email:        c.Email,
			CampaignName: c.CampaignName,
			CapturedAt:   c.CapturedAt,
			SessionID:    c.SessionID,
		}
	}

	writeJSON(w, http.StatusOK, sdk.DashboardResponse{
		Campaigns: sdk.CampaignCounts{
			Draft:     stats.Campaigns.Draft,
			Active:    stats.Campaigns.Active,
			Completed: stats.Campaigns.Completed,
			Cancelled: stats.Campaigns.Cancelled,
			Total:     stats.Campaigns.Total,
		},
		TotalCaptures:   stats.TotalCaptures,
		TotalClicks:     stats.TotalClicks,
		TotalEmailsSent: stats.TotalEmailsSent,
		MiragedCount:    stats.MiragedCount,
		ActiveCampaigns: active,
		RecentCaptures:  captures,
	})
}
