package phishing

import "github.com/travisbale/barb/sdk"

// CampaignEvent is a domain event describing a state change to a campaign
// or one of its results.
type CampaignEvent struct {
	Type       sdk.CampaignEventType `json:"type"`
	CampaignID string                `json:"campaign_id"`
	Result     *CampaignResult       `json:"result,omitempty"`
	Status     string                `json:"status,omitempty"`
}

// newStatusEvent constructs a CampaignEvent describing the campaign's
// current status.
func newStatusEvent(campaign *Campaign) CampaignEvent {
	return CampaignEvent{
		Type:       sdk.EventCampaignStatus,
		CampaignID: campaign.ID,
		Status:     string(campaign.Status),
	}
}

// newResultEvent constructs a CampaignEvent describing a change to a
// campaign result.
func newResultEvent(result *CampaignResult) CampaignEvent {
	return CampaignEvent{
		Type:       sdk.EventResultUpdated,
		CampaignID: result.CampaignID,
		Result:     result,
	}
}

// eventBus is the publish/subscribe port for campaign events. Concrete
// implementations live in internal/events and are injected into services
// that need to emit or observe campaign events.
//
// Publish must never block: if a subscriber's channel is full the event is
// dropped for that subscriber. The returned unsubscribe func is safe to
// call multiple times.
type eventBus interface {
	Publish(event CampaignEvent)
	Subscribe(campaignID string) (events <-chan CampaignEvent, unsubscribe func())
}
