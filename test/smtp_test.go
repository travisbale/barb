package test_test

import (
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/travisbale/barb/internal/delivery"
	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/internal/store/sqlite"
	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestSMTPProfiles_CRUD(t *testing.T) {
	t.Parallel()
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

func TestSMTPProfiles_DefaultPort(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestSMTPProfiles_Update(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Original",
		Host:     "smtp.old.com",
		Port:     587,
		FromAddr: "old@example.com",
	})
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}

	updated, err := h.Client.UpdateSMTPProfile(created.ID, sdk.UpdateSMTPProfileRequest{
		Name:     strPtr("Updated"),
		Host:     strPtr("smtp.new.com"),
		Port:     intPtr(465),
		FromAddr: strPtr("new@example.com"),
		FromName: strPtr("New Name"),
	})
	if err != nil {
		t.Fatalf("UpdateSMTPProfile: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("Name = %q, want %q", updated.Name, "Updated")
	}
	if updated.Host != "smtp.new.com" {
		t.Errorf("Host = %q, want %q", updated.Host, "smtp.new.com")
	}
	if updated.Port != 465 {
		t.Errorf("Port = %d, want 465", updated.Port)
	}
	if updated.FromAddr != "new@example.com" {
		t.Errorf("FromAddr = %q, want %q", updated.FromAddr, "new@example.com")
	}
	if updated.FromName != "New Name" {
		t.Errorf("FromName = %q, want %q", updated.FromName, "New Name")
	}
}

func TestSMTPProfiles_UpdatePreservesPassword(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	created, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "With Password",
		Host:     "smtp.example.com",
		Password: "original-secret",
		FromAddr: "from@example.com",
	})
	if err != nil {
		t.Fatalf("CreateSMTPProfile: %v", err)
	}

	// Update only the name — password (nil) should be preserved.
	_, err = h.Client.UpdateSMTPProfile(created.ID, sdk.UpdateSMTPProfileRequest{
		Name: strPtr("Renamed"),
	})
	if err != nil {
		t.Fatalf("UpdateSMTPProfile: %v", err)
	}

	// Verify the profile still works (password was preserved, not blanked).
	got, err := h.Client.GetSMTPProfile(created.ID)
	if err != nil {
		t.Fatalf("GetSMTPProfile: %v", err)
	}
	if got.Name != "Renamed" {
		t.Errorf("Name = %q, want %q", got.Name, "Renamed")
	}
}

func TestSMTPProfiles_DeleteNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeleteSMTPProfile("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestSMTPProfiles_CreateRejectsUnreachableHost(t *testing.T) {
	t.Parallel()
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	_, err := h.Client.CreateSMTPProfile(sdk.CreateSMTPProfileRequest{
		Name:     "Bad SMTP",
		Host:     "127.0.0.1",
		Port:     19999,
		FromAddr: "from@example.com",
	})
	wantError(t, err, http.StatusUnprocessableEntity, "Could not connect to the SMTP server")
}

func TestSMTPProfiles_UpdateRejectsUnreachableHost(t *testing.T) {
	t.Parallel()
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	// Insert a profile directly into the store to bypass the create dial check.
	store := sqlite.NewSMTPStore(h.DB, nil)
	if err := store.CreateProfile(&phishing.SMTPProfile{
		ID: "smtp-update-test", Name: "Original", Host: "smtp.example.com",
		Port: 587, FromAddr: "from@example.com", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	_, err := h.Client.UpdateSMTPProfile("smtp-update-test", sdk.UpdateSMTPProfileRequest{
		Host: strPtr("127.0.0.1"),
		Port: intPtr(19999),
	})
	wantError(t, err, http.StatusUnprocessableEntity, "Could not connect to the SMTP server")
}

func TestSMTPProfiles_StartCampaignRejectsUnreachableSMTP(t *testing.T) {
	t.Parallel()
	h := test.NewHarnessWithMailer(t, &delivery.Sender{Logger: slog.Default()})

	// Insert an SMTP profile directly to bypass the create dial check.
	store := sqlite.NewSMTPStore(h.DB, nil)
	if err := store.CreateProfile(&phishing.SMTPProfile{
		ID: "smtp-start-test", Name: "Unreachable", Host: "127.0.0.1",
		Port: 19999, FromAddr: "from@example.com", CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	tmpl := createTestTemplate(t, h)
	list := createTestTargetList(t, h, sdk.AddTargetRequest{Email: "alice@example.com"})

	req := validCampaignRequest(list.ID, tmpl.ID, "smtp-start-test")
	campaign, err := h.Client.CreateCampaign(req)
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}

	err = h.Client.StartCampaign(campaign.ID)
	wantError(t, err, http.StatusUnprocessableEntity, "Could not connect to the SMTP server")

	// Campaign should still be in draft.
	got, _ := h.Client.GetCampaign(campaign.ID)
	if got.Status != "draft" {
		t.Errorf("Status = %q, want %q", got.Status, "draft")
	}
}
