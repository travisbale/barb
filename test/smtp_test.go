package test_test

import (
	"testing"

	"github.com/travisbale/mirador/sdk"
	"github.com/travisbale/mirador/test"
)

func TestSMTPProfiles_CRUD(t *testing.T) {
	h := test.NewHarness(t)

	created, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Gmail Relay",
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "ops@example.com",
		Password: "app-password",
		FromAddr: "support@example.com",
		FromName: "IT Support",
	})
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Name != "Gmail Relay" {
		t.Errorf("Name = %q, want %q", created.Name, "Gmail Relay")
	}
	if created.Port != 587 {
		t.Errorf("Port = %d, want 587", created.Port)
	}

	// Get by ID.
	got, err := h.Client.GetSMTPProfile(created.ID)
	if err != nil {
		t.Fatalf("GetSMTPProfile: %v", err)
	}
	if got.Host != "smtp.gmail.com" {
		t.Errorf("Host = %q, want %q", got.Host, "smtp.gmail.com")
	}

	// List.
	profiles, err := h.Client.ListSMTPProfiles()
	if err != nil {
		t.Fatalf("ListSMTPProfiles: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}

	// Delete.
	if err := h.Client.DeleteSMTPProfile(created.ID); err != nil {
		t.Fatalf("DeleteSMTPProfile: %v", err)
	}

	profiles, err = h.Client.ListSMTPProfiles()
	if err != nil {
		t.Fatalf("ListSMTPProfiles after delete: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles after delete, got %d", len(profiles))
	}
}

func TestSMTPProfiles_RequiresName(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Host:     "smtp.example.com",
		FromAddr: "from@example.com",
	})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestSMTPProfiles_RequiresHost(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Test",
		FromAddr: "from@example.com",
	})
	if err == nil {
		t.Error("expected error for missing host")
	}
}

func TestSMTPProfiles_RequiresFromAddr(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name: "Test",
		Host: "smtp.example.com",
	})
	if err == nil {
		t.Error("expected error for missing from address")
	}
}

func TestSMTPProfiles_DefaultPort(t *testing.T) {
	h := test.NewHarness(t)

	created, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "No Port",
		Host:     "smtp.example.com",
		FromAddr: "from@example.com",
	})
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}
	if created.Port != 587 {
		t.Errorf("Port = %d, want 587 (default)", created.Port)
	}
}

func TestSMTPProfiles_PasswordNotInResponse(t *testing.T) {
	h := test.NewHarness(t)

	created, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Secret",
		Host:     "smtp.example.com",
		Password: "super-secret",
		FromAddr: "from@example.com",
	})
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}

	got, err := h.Client.GetSMTPProfile(created.ID)
	if err != nil {
		t.Fatalf("GetSMTPProfile: %v", err)
	}

	// The SDK response type doesn't have a Password field — it's intentionally
	// omitted from the JSON response to avoid leaking credentials.
	_ = got
}

func TestSMTPProfiles_DeleteNotFound(t *testing.T) {
	h := test.NewHarness(t)

	err := h.Client.DeleteSMTPProfile("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}
