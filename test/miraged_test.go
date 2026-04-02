package test_test

import (
	"net/http"
	"testing"

	"github.com/travisbale/barb/test"
)

func TestMiraged_DeleteNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeleteMiraged("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}
