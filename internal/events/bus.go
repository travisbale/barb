// Package events provides in-process publish/subscribe implementations for
// barb's domain event ports. Each concrete bus satisfies the corresponding
// interface declared in the domain package it serves.
package events

import (
	"log/slog"
	"sync"

	"github.com/travisbale/barb/internal/phishing"
)

// DefaultBufferSize is the channel buffer size used for new subscriptions.
const DefaultBufferSize = 16

// Bus is a goroutine-safe publish/subscribe bus for campaign events keyed
// by campaign ID. It implements the eventBus port declared in the phishing
// package.
//
// Delivery guarantee: best-effort, in-process only. Events are never
// persisted. If a subscriber's channel is full when Publish is called, the
// event is dropped for that subscriber and a warning is logged.
type Bus struct {
	bufSize int
	mu      sync.RWMutex
	subs    map[string][]*subEntry
}

type subEntry struct {
	ch chan phishing.CampaignEvent
}

func NewBus() *Bus {
	return &Bus{
		bufSize: DefaultBufferSize,
		subs:    make(map[string][]*subEntry),
	}
}

// Publish sends event to every subscriber registered for event.CampaignID.
// The publisher is never blocked: a subscriber whose buffer is full loses
// the event and a warning is logged.
//
// The read lock is held across the send loop to prevent Unsubscribe from
// closing a subscriber's channel between iteration and send (which would
// panic). Sends are non-blocking (select + default), so the critical section
// stays short: concurrent publishers still run in parallel under RLock, and
// an Unsubscribe call only waits for in-flight publishes to return.
func (b *Bus) Publish(event phishing.CampaignEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, entry := range b.subs[event.CampaignID] {
		select {
		case entry.ch <- event:
		default:
			slog.Warn("events: campaign subscriber channel full, dropping event",
				"campaign_id", event.CampaignID,
				"type", event.Type,
			)
		}
	}
}

// Subscribe registers a subscriber for events tagged with campaignID and
// returns a Subscription handle. The caller releases resources by calling
// Subscription.Unsubscribe; it is safe to invoke multiple times.
func (b *Bus) Subscribe(campaignID string) *phishing.Subscription {
	ch := make(chan phishing.CampaignEvent, b.bufSize)
	entry := &subEntry{ch: ch}

	b.mu.Lock()
	b.subs[campaignID] = append(b.subs[campaignID], entry)
	b.mu.Unlock()

	return &phishing.Subscription{
		Events:      ch,
		Unsubscribe: func() { b.unsubscribe(campaignID, entry) },
	}
}

// unsubscribe removes entry from the subscriber list for campaignID and
// closes its channel. Idempotent: after the first call the entry is no
// longer in the slice and subsequent calls are no-ops, so
// Subscription.Unsubscribe is safe to invoke multiple times.
func (b *Bus) unsubscribe(campaignID string, entry *subEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	entries := b.subs[campaignID]
	for i, e := range entries {
		if e == entry {
			close(e.ch)
			last := len(entries) - 1
			entries[i] = entries[last]
			b.subs[campaignID] = entries[:last]
			if len(b.subs[campaignID]) == 0 {
				delete(b.subs, campaignID)
			}
			return
		}
	}
}
