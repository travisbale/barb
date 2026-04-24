package test_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/internal/store/sqlite"
	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestMiraged_Rename(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	// Insert a connection directly — enrollment requires a real miraged.
	store := sqlite.NewMiragedStore(h.DB, h.Cipher)
	conn := &phishing.MiragedConnection{
		ID: "rename-test", Name: "Old Name", Address: "127.0.0.1:443",
		SecretHostname: "mgmt.local", CertPEM: []byte("cert"), KeyPEM: []byte("key"),
		CACertPEM: []byte("ca"), CreatedAt: time.Now(),
	}
	if err := store.CreateConnection(conn); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	updated, err := h.Client.UpdateMiraged("rename-test", sdk.UpdateMiragedRequest{
		Name: strPtr("New Name"),
	})
	if err != nil {
		t.Fatalf("UpdateMiraged: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("Name = %q, want %q", updated.Name, "New Name")
	}

	// Verify persisted.
	got, _ := store.GetConnection("rename-test")
	if got.Name != "New Name" {
		t.Errorf("stored Name = %q, want %q", got.Name, "New Name")
	}
}

func TestMiraged_RenameNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	_, err := h.Client.UpdateMiraged("nonexistent", sdk.UpdateMiragedRequest{
		Name: strPtr("Nope"),
	})
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestMiraged_DeleteNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeleteMiraged("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}

// An empty patch (nil Name) is a no-op and returns the current state;
// this also guards against a regression of the nil-deref panic that
// occurred when the handler unconditionally dereferenced *body.Name.
func TestMiraged_UpdateEmptyPatch(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	store := sqlite.NewMiragedStore(h.DB, h.Cipher)
	conn := &phishing.MiragedConnection{
		ID: "empty-patch-test", Name: "Unchanged", Address: "127.0.0.1:443",
		SecretHostname: "mgmt.local", CertPEM: []byte("cert"), KeyPEM: []byte("key"),
		CACertPEM: []byte("ca"), CreatedAt: time.Now(),
	}
	if err := store.CreateConnection(conn); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	got, err := h.Client.UpdateMiraged("empty-patch-test", sdk.UpdateMiragedRequest{})
	if err != nil {
		t.Fatalf("UpdateMiraged: %v", err)
	}
	if got.Name != "Unchanged" {
		t.Errorf("Name = %q, want %q", got.Name, "Unchanged")
	}
}
