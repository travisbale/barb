package phishing

import (
	"sync"

	"github.com/travisbale/barb/sdk"
)

// CampaignEvent represents a real-time update to a campaign or its results.
type CampaignEvent struct {
	Type       string          `json:"type"`
	CampaignID string          `json:"campaign_id"`
	Result     *CampaignResult `json:"result,omitempty"`
	Status     string          `json:"status,omitempty"`
}

// CampaignBus is a pub/sub for campaign events. Subscribers register by
// campaign ID and receive events on a channel. Safe for concurrent use.
type CampaignBus struct {
	mu   sync.Mutex
	subs map[string][]chan CampaignEvent
}

func NewCampaignBus() *CampaignBus {
	return &CampaignBus{subs: make(map[string][]chan CampaignEvent)}
}

// Subscribe returns a channel that receives events for the given campaign.
func (b *CampaignBus) Subscribe(campaignID string) chan CampaignEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan CampaignEvent, 16)
	b.subs[campaignID] = append(b.subs[campaignID], ch)
	return ch
}

// Unsubscribe removes a channel from the campaign's subscriber list.
func (b *CampaignBus) Unsubscribe(campaignID string, ch chan CampaignEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[campaignID]
	for i, s := range subs {
		if s == ch {
			b.subs[campaignID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(b.subs[campaignID]) == 0 {
		delete(b.subs, campaignID)
	}
}

// Publish sends an event to all subscribers of the campaign. Non-blocking —
// if a subscriber's channel is full, the event is dropped.
func (b *CampaignBus) Publish(event CampaignEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subs[event.CampaignID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// PublishResultUpdate sends a result.updated event.
func (b *CampaignBus) PublishResultUpdate(campaignID string, result *CampaignResult) {
	b.Publish(CampaignEvent{
		Type:       sdk.EventResultUpdated,
		CampaignID: campaignID,
		Result:     result,
	})
}

// PublishStatusChange sends a campaign.status event.
func (b *CampaignBus) PublishStatusChange(campaignID string, status CampaignStatus) {
	b.Publish(CampaignEvent{
		Type:       sdk.EventCampaignStatus,
		CampaignID: campaignID,
		Status:     string(status),
	})
}
