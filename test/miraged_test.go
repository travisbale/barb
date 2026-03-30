package test_test

import (
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestMiraged_CRUD(t *testing.T) {
	h := test.NewHarness(t)

	created, err := h.Client.CreateMiraged(sdk.CreateMiragedRequest{
		Name:           "Local Dev",
		Address:        "127.0.0.1:443",
		SecretHostname: "mgmt.phish.local",
		CertPEM:        "-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----",
		KeyPEM:         "-----BEGIN EC PRIVATE KEY-----\nfake\n-----END EC PRIVATE KEY-----",
		CACertPEM:      "-----BEGIN CERTIFICATE-----\nfake-ca\n-----END CERTIFICATE-----",
	})
	if err != nil {
		t.Fatalf("CreateMiraged: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Name != "Local Dev" {
		t.Errorf("Name = %q, want %q", created.Name, "Local Dev")
	}

	// List.
	connections, err := h.Client.ListMiraged()
	if err != nil {
		t.Fatalf("ListMiraged: %v", err)
	}
	if len(connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(connections))
	}

	// Certs should not be in the response.
	// (MiragedResponse doesn't have cert fields — they're omitted by design)

	// Delete.
	if err := h.Client.DeleteMiraged(created.ID); err != nil {
		t.Fatalf("DeleteMiraged: %v", err)
	}

	connections, _ = h.Client.ListMiraged()
	if len(connections) != 0 {
		t.Errorf("expected 0 connections after delete, got %d", len(connections))
	}
}

func TestMiraged_RequiresName(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateMiraged(sdk.CreateMiragedRequest{
		Address:        "127.0.0.1:443",
		SecretHostname: "mgmt.phish.local",
		CertPEM:        "cert",
		KeyPEM:         "key",
		CACertPEM:      "ca",
	})
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestMiraged_RequiresAddress(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateMiraged(sdk.CreateMiragedRequest{
		Name:           "Test",
		SecretHostname: "mgmt.phish.local",
		CertPEM:        "cert",
		KeyPEM:         "key",
		CACertPEM:      "ca",
	})
	if err == nil {
		t.Error("expected error for missing address")
	}
}

func TestMiraged_RequiresCerts(t *testing.T) {
	h := test.NewHarness(t)

	_, err := h.Client.CreateMiraged(sdk.CreateMiragedRequest{
		Name:           "Test",
		Address:        "127.0.0.1:443",
		SecretHostname: "mgmt.phish.local",
	})
	if err == nil {
		t.Error("expected error for missing certs")
	}
}

func TestMiraged_DeleteNotFound(t *testing.T) {
	h := test.NewHarness(t)

	err := h.Client.DeleteMiraged("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}

func TestMiraged_TestConnectionFails(t *testing.T) {
	h := test.NewHarness(t)

	// Create with fake certs — test connection should fail.
	created, _ := h.Client.CreateMiraged(sdk.CreateMiragedRequest{
		Name:           "Unreachable",
		Address:        "127.0.0.1:19999",
		SecretHostname: "mgmt.phish.local",
		CertPEM:        "-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----",
		KeyPEM:         "-----BEGIN EC PRIVATE KEY-----\nfake\n-----END EC PRIVATE KEY-----",
		CACertPEM:      "-----BEGIN CERTIFICATE-----\nfake-ca\n-----END CERTIFICATE-----",
	})

	status, err := h.Client.TestMiraged(created.ID)
	if err != nil {
		t.Fatalf("TestMiraged: %v", err)
	}
	if status.Connected {
		t.Error("expected Connected = false for fake certs")
	}
	if status.Error == "" {
		t.Error("expected non-empty error message")
	}
}
