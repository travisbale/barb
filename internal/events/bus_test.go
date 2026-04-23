package events_test

import (
	"sync"
	"testing"
	"time"

	"github.com/travisbale/barb/internal/events"
	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

// TestPublishUnsubscribeRace stress-tests for the send-on-closed-channel
// panic that occurs if Publish iterates the subscriber list while
// Unsubscribe closes one of those channels. The race detector does not
// catch this — channel close/send use channel-internal sync, not the bus
// mutex — so we rely on the panic itself as the failure signal.
//
// Without the lock-held-across-publish fix, this test reliably panics
// within milliseconds. With the fix it always passes.
func TestPublishUnsubscribeRace(t *testing.T) {
	bus := events.NewBus()

	const campaignID = "stress-campaign"
	const publishers = 8
	const churners = 8
	const duration = 200 * time.Millisecond

	deadline := time.Now().Add(duration)
	panics := make(chan any, publishers+churners)

	work := func(wg *sync.WaitGroup, fn func()) {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				panics <- r
			}
		}()
		fn()
	}

	var wg sync.WaitGroup

	// Publishers spam events as fast as possible.
	for range publishers {
		wg.Add(1)
		go work(&wg, func() {
			for time.Now().Before(deadline) {
				bus.Publish(phishing.CampaignEvent{
					Type:       sdk.EventCampaignStatus,
					CampaignID: campaignID,
					Status:     "active",
				})
			}
		})
	}

	// Churners subscribe and immediately unsubscribe to maximize the chance
	// that a publisher iterates to a channel just as it's being closed.
	for range churners {
		wg.Add(1)
		go work(&wg, func() {
			for time.Now().Before(deadline) {
				_, unsubscribe := bus.Subscribe(campaignID)
				unsubscribe()
			}
		})
	}

	wg.Wait()
	close(panics)

	for r := range panics {
		t.Errorf("goroutine panicked: %v", r)
	}
}
