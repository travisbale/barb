package test_test

import (
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestAuth_UnauthenticatedRequestReturns401(t *testing.T) {
	h := test.NewHarness(t)

	// Create a second client without a session.
	unauthenticated := sdk.NewClient("http://" + h.Addr)

	_, err := unauthenticated.ListCampaigns()
	if err == nil {
		t.Error("expected error for unauthenticated request")
	}
}

func TestAuth_LoginWithBadPassword(t *testing.T) {
	h := test.NewHarness(t)

	unauthenticated := sdk.NewClient("http://" + h.Addr)
	err := unauthenticated.Login(sdk.LoginRequest{Username: "admin", Password: "wrong"})
	if err == nil {
		t.Error("expected error for bad password")
	}
}

func TestAuth_LoginWithBadUsername(t *testing.T) {
	h := test.NewHarness(t)

	unauthenticated := sdk.NewClient("http://" + h.Addr)
	err := unauthenticated.Login(sdk.LoginRequest{Username: "nobody", Password: "whatever"})
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestAuth_MeReturnsCurrentUser(t *testing.T) {
	h := test.NewHarness(t)

	user, err := h.Client.Me()
	if err != nil {
		t.Fatalf("Me: %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("Username = %q, want %q", user.Username, "admin")
	}
	if user.PasswordChangeRequired {
		t.Error("PasswordChangeRequired should be false after password change")
	}
}

func TestAuth_ChangePasswordValidation(t *testing.T) {
	h := test.NewHarness(t)

	// Wrong current password.
	err := h.Client.ChangePassword(sdk.ChangePasswordRequest{
		CurrentPassword: "wrong",
		NewPassword:     "new-password-123",
	})
	if err == nil {
		t.Error("expected error for wrong current password")
	}

	// Too short new password.
	err = h.Client.ChangePassword(sdk.ChangePasswordRequest{
		CurrentPassword: "test-password-12345",
		NewPassword:     "short",
	})
	if err == nil {
		t.Error("expected error for password too short")
	}
}

func TestAuth_StatusEndpointIsPublic(t *testing.T) {
	h := test.NewHarness(t)

	unauthenticated := sdk.NewClient("http://" + h.Addr)
	status, err := unauthenticated.Status()
	if err != nil {
		t.Fatalf("Status should be public, got: %v", err)
	}
	if status.Version != "test" {
		t.Errorf("Version = %q, want %q", status.Version, "test")
	}
}
