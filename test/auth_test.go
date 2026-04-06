package test_test

import (
	"net/http"
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestAuth_UnauthenticatedRequestReturns401(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	// Create a second client without a session.
	unauthenticated := sdk.NewClient("http://" + h.Addr)

	_, err := unauthenticated.ListCampaigns()
	wantError(t, err, http.StatusUnauthorized, "Authentication required")
}

func TestAuth_LoginWithBadPassword(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	unauthenticated := sdk.NewClient("http://" + h.Addr)
	err := unauthenticated.Login(sdk.LoginRequest{Username: "admin", Password: "wrong"})
	wantError(t, err, http.StatusUnauthorized, "Invalid username or password")
}

func TestAuth_LoginWithBadUsername(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	unauthenticated := sdk.NewClient("http://" + h.Addr)
	err := unauthenticated.Login(sdk.LoginRequest{Username: "nobody", Password: "whatever"})
	wantError(t, err, http.StatusUnauthorized, "Invalid username or password")
}

func TestAuth_MeReturnsCurrentUser(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	h := test.NewHarness(t)

	// Wrong current password.
	err := h.Client.ChangePassword(sdk.ChangePasswordRequest{
		CurrentPassword: "wrong",
		NewPassword:     "new-password-123",
	})
	wantError(t, err, http.StatusUnauthorized, "Current password is incorrect")

	// Too short new password.
	err = h.Client.ChangePassword(sdk.ChangePasswordRequest{
		CurrentPassword: "test-password-12345",
		NewPassword:     "short",
	})
	wantError(t, err, http.StatusUnprocessableEntity, "Password must be at least 8 characters")
}

func TestAuth_StatusEndpointIsPublic(t *testing.T) {
	t.Parallel()
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
