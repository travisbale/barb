package test_test

import (
	"net/http"
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

// These tests exercise the barb notification proxy handlers' not-found
// path, which returns before any call to miraged. Request validation is
// covered in sdk/validate_test.go.

func TestMiragedNotifications_CreateConnectionNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.CreateMiragedNotification("nonexistent", sdk.CreateMiragedNotificationChannelRequest{
		Type: "slack",
		URL:  "https://hooks.slack.com/fake",
	})
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestMiragedNotifications_ListConnectionNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.ListMiragedNotifications("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestMiragedNotifications_DeleteConnectionNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeleteMiragedNotification("nonexistent", "some-channel-id")
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestMiragedNotifications_TestConnectionNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.TestMiragedNotification("nonexistent", "some-channel-id")
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestMiragedNotifications_ListEventTypesConnectionNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.ListMiragedNotificationEventTypes("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}
