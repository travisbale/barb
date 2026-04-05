package phishing

import (
	"sort"
	"time"
)

// DashboardStats holds aggregate metrics for the dashboard.
type DashboardStats struct {
	Campaigns       CampaignCounts
	TotalCaptures   int
	TotalClicks     int
	TotalEmailsSent int
	MiragedCount    int
	ActiveCampaigns []ActiveCampaign
	RecentCaptures  []RecentCapture
}

type CampaignCounts struct {
	Draft     int
	Active    int
	Completed int
	Cancelled int
	Total     int
}

type ActiveCampaign struct {
	ID        string
	Name      string
	Sent      int
	Failed    int
	Captured  int
	Completed int
	Total     int
}

type RecentCapture struct {
	Email        string
	CampaignID   string
	CampaignName string
	CapturedAt   string
	SessionID    string
}

// DashboardService aggregates stats from existing stores.
type DashboardService struct {
	Campaigns campaignStore
	Miraged   miragedStore
}

func (s *DashboardService) Stats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	campaigns, err := s.Campaigns.ListCampaigns()
	if err != nil {
		return nil, err
	}

	for _, campaign := range campaigns {
		stats.Campaigns.Total++
		switch campaign.Status {
		case CampaignDraft:
			stats.Campaigns.Draft++
		case CampaignActive:
			stats.Campaigns.Active++
		case CampaignCompleted:
			stats.Campaigns.Completed++
		case CampaignCancelled:
			stats.Campaigns.Cancelled++
		}
	}

	// Build active campaign cards and collect all captures.
	var allCaptures []RecentCapture

	for _, campaign := range campaigns {
		results, err := s.Campaigns.ListResults(campaign.ID)
		if err != nil {
			return nil, err
		}

		var sent, failed, clicked, captured, completed int
		for _, result := range results {
			switch result.Status {
			case ResultSent:
				sent++
			case ResultFailed:
				failed++
			case ResultClicked:
				clicked++
			case ResultCaptured:
				captured++
				if result.CapturedAt != nil {
					allCaptures = append(allCaptures, RecentCapture{
						Email:        result.Email,
						CampaignID:   campaign.ID,
						CampaignName: campaign.Name,
						CapturedAt:   result.CapturedAt.Format(time.RFC3339),
						SessionID:    result.SessionID,
					})
				}
			case ResultCompleted:
				completed++
				if result.CapturedAt != nil {
					allCaptures = append(allCaptures, RecentCapture{
						Email:        result.Email,
						CampaignID:   campaign.ID,
						CampaignName: campaign.Name,
						CapturedAt:   result.CapturedAt.Format(time.RFC3339),
						SessionID:    result.SessionID,
					})
				}
			}
		}

		stats.TotalCaptures += captured + completed
		stats.TotalClicks += clicked + captured + completed
		stats.TotalEmailsSent += sent + clicked + captured + completed

		if campaign.Status == CampaignActive {
			stats.ActiveCampaigns = append(stats.ActiveCampaigns, ActiveCampaign{
				ID:        campaign.ID,
				Name:      campaign.Name,
				Sent:      sent + clicked + captured + completed,
				Failed:    failed,
				Captured:  captured,
				Completed: completed,
				Total:     len(results),
			})
		}
	}

	// Sort captures by time descending and take the most recent 10.
	sort.Slice(allCaptures, func(i, j int) bool {
		return allCaptures[i].CapturedAt > allCaptures[j].CapturedAt
	})
	if len(allCaptures) > 10 {
		allCaptures = allCaptures[:10]
	}
	stats.RecentCaptures = allCaptures

	connections, err := s.Miraged.ListConnections()
	if err != nil {
		return nil, err
	}
	stats.MiragedCount = len(connections)

	return stats, nil
}
