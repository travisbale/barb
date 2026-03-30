package test_test

import (
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestMiraged_EnrollRequiresName(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.EnrollMiraged(sdk.EnrollMiragedRequest{
		Address:        "127.0.0.1:443",
		SecretHostname: "mgmt.phish.local",
		Token:          "fake-token",
	})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestMiraged_EnrollRequiresAddress(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.EnrollMiraged(sdk.EnrollMiragedRequest{
		Name:           "Test",
		SecretHostname: "mgmt.phish.local",
		Token:          "fake-token",
	})
	if err == nil {
		t.Error("expected error for missing address")
	}
}

func TestMiraged_EnrollRequiresSecretHostname(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.EnrollMiraged(sdk.EnrollMiragedRequest{
		Name:    "Test",
		Address: "127.0.0.1:443",
		Token:   "fake-token",
	})
	if err == nil {
		t.Error("expected error for missing secret hostname")
	}
}

func TestMiraged_EnrollRequiresToken(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.EnrollMiraged(sdk.EnrollMiragedRequest{
		Name:           "Test",
		Address:        "127.0.0.1:443",
		SecretHostname: "mgmt.phish.local",
	})
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestMiraged_DeleteNotFound(t *testing.T) {
	h := test.NewHarness(t)

	err := h.Client.DeleteMiraged("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}
