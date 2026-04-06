package test_test

import (
	"net/http"
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestPhishlets_CRUD(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreatePhishlet(sdk.CreatePhishletRequest{
		YAML: validPhishletYAML,
	})
	if err != nil {
		t.Fatalf("CreatePhishlet: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Name != "example" {
		t.Errorf("Name = %q, want %q (extracted from YAML)", created.Name, "example")
	}
	if created.YAML != validPhishletYAML {
		t.Error("YAML content not preserved")
	}

	// Get.
	got, err := h.Client.GetPhishlet(created.ID)
	if err != nil {
		t.Fatalf("GetPhishlet: %v", err)
	}
	if got.Name != "example" {
		t.Errorf("Name = %q", got.Name)
	}

	// List.
	phishlets, err := h.Client.ListPhishlets()
	if err != nil {
		t.Fatalf("ListPhishlets: %v", err)
	}
	if len(phishlets) != 1 {
		t.Fatalf("expected 1 phishlet, got %d", len(phishlets))
	}

	// Update.
	updatedYAML := `name: updated
author: test
version: "2.0"
proxy_hosts:
  - phish_sub: login
    orig_sub: login
    domain: example.com
    is_landing: true
auth_tokens:
  - domain: example.com
    keys:
      - name: session
login:
  domain: login.example.com
  path: /login
`
	updated, err := h.Client.UpdatePhishlet(created.ID, sdk.UpdatePhishletRequest{
		YAML: updatedYAML,
	})
	if err != nil {
		t.Fatalf("UpdatePhishlet: %v", err)
	}
	if updated.Name != "updated" {
		t.Errorf("Name after update = %q, want %q", updated.Name, "updated")
	}

	// Delete.
	if err := h.Client.DeletePhishlet(created.ID); err != nil {
		t.Fatalf("DeletePhishlet: %v", err)
	}
	phishlets, _ = h.Client.ListPhishlets()
	if len(phishlets) != 0 {
		t.Errorf("expected 0 phishlets after delete, got %d", len(phishlets))
	}
}

func TestPhishlets_ExtractsNameFromYAML(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreatePhishlet(sdk.CreatePhishletRequest{
		YAML: validPhishletYAML,
	})
	if err != nil {
		t.Fatalf("CreatePhishlet: %v", err)
	}
	if created.Name != "example" {
		t.Errorf("Name = %q, want %q", created.Name, "example")
	}
}

func TestPhishlets_RequiresNameInYAML(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.CreatePhishlet(sdk.CreatePhishletRequest{
		YAML: "author: test\nversion: '1.0'\n",
	})
	wantError(t, err, http.StatusInternalServerError, "Failed to create phishlet")
}

func TestPhishlets_InvalidYAML(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.CreatePhishlet(sdk.CreatePhishletRequest{
		YAML: "{{invalid yaml",
	})
	wantError(t, err, http.StatusInternalServerError, "Failed to create phishlet")
}

func TestPhishlets_UpdateInvalidYAML(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreatePhishlet(sdk.CreatePhishletRequest{YAML: validPhishletYAML})
	if err != nil {
		t.Fatalf("CreatePhishlet: %v", err)
	}

	_, err = h.Client.UpdatePhishlet(created.ID, sdk.UpdatePhishletRequest{YAML: "{{invalid"})
	wantError(t, err, http.StatusInternalServerError, "Failed to update phishlet")

	// Verify the original is unchanged.
	got, err := h.Client.GetPhishlet(created.ID)
	if err != nil {
		t.Fatalf("GetPhishlet: %v", err)
	}
	if got.Name != "example" {
		t.Errorf("Name = %q, want %q (should be unchanged)", got.Name, "example")
	}
}

func TestPhishlets_DeleteNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeletePhishlet("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}
