package phishing

// CampaignEvent is a domain event describing a state change to a campaign
// or one of its results.
type CampaignEvent struct {
	Type       string          `json:"type"`
	CampaignID string          `json:"campaign_id"`
	Result     *CampaignResult `json:"result,omitempty"`
	Status     string          `json:"status,omitempty"`
}

// eventBus is the publish/subscribe port for campaign events. Concrete
// implementations live in internal/events and are injected into services
// that need to emit or observe campaign events.
//
// Publish must never block: if a subscriber's channel is full the event is
// dropped for that subscriber.
type eventBus interface {
	Publish(event CampaignEvent)
	Subscribe(campaignID string) <-chan CampaignEvent
	Unsubscribe(campaignID string, ch <-chan CampaignEvent)
}
