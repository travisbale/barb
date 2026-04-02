package test_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

func TestTargetLists_CRUD(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	// Create a target list.
	created, err := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Acme Corp"})
	if err != nil {
		t.Fatalf("CreateTargetList: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Name != "Acme Corp" {
		t.Errorf("Name = %q, want %q", created.Name, "Acme Corp")
	}

	// Get by ID.
	got, err := h.Client.GetTargetList(created.ID)
	if err != nil {
		t.Fatalf("GetTargetList: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}

	// List.
	lists, err := h.Client.ListTargetLists()
	if err != nil {
		t.Fatalf("ListTargetLists: %v", err)
	}
	if len(lists) != 1 {
		t.Fatalf("expected 1 list, got %d", len(lists))
	}

	// Delete.
	if err := h.Client.DeleteTargetList(created.ID); err != nil {
		t.Fatalf("DeleteTargetList: %v", err)
	}

	lists, err = h.Client.ListTargetLists()
	if err != nil {
		t.Fatalf("ListTargetLists after delete: %v", err)
	}
	if len(lists) != 0 {
		t.Errorf("expected 0 lists after delete, got %d", len(lists))
	}
}

func TestTargetLists_DeleteNotFound(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	err := h.Client.DeleteTargetList("nonexistent")
	wantError(t, err, http.StatusNotFound, "not found")
}

func TestTargets_CRUD(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	// Create a list first.
	list, err := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Test"})
	if err != nil {
		t.Fatalf("CreateTargetList: %v", err)
	}

	// Add a target.
	target, err := h.Client.AddTarget(list.ID, sdk.AddTargetRequest{
		Email:     "alice@acme.com",
		FirstName: "Alice",
		LastName:  "Smith",
	})
	if err != nil {
		t.Fatalf("AddTarget: %v", err)
	}
	if target.Email != "alice@acme.com" {
		t.Errorf("Email = %q, want %q", target.Email, "alice@acme.com")
	}
	if target.ListID != list.ID {
		t.Errorf("ListID = %q, want %q", target.ListID, list.ID)
	}

	// List targets.
	targets, err := h.Client.ListTargets(list.ID)
	if err != nil {
		t.Fatalf("ListTargets: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}

	// Delete target.
	if err := h.Client.DeleteTarget(target.ID); err != nil {
		t.Fatalf("DeleteTarget: %v", err)
	}

	targets, err = h.Client.ListTargets(list.ID)
	if err != nil {
		t.Fatalf("ListTargets after delete: %v", err)
	}
	if len(targets) != 0 {
		t.Errorf("expected 0 targets after delete, got %d", len(targets))
	}
}

func TestTargets_CascadeDeleteWithList(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Test"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "alice@acme.com"})
	h.Client.AddTarget(list.ID, sdk.AddTargetRequest{Email: "bob@acme.com"})

	// Deleting the list should cascade-delete the targets.
	if err := h.Client.DeleteTargetList(list.ID); err != nil {
		t.Fatalf("DeleteTargetList: %v", err)
	}

	targets, err := h.Client.ListTargets(list.ID)
	if err != nil {
		t.Fatalf("ListTargets: %v", err)
	}
	if len(targets) != 0 {
		t.Errorf("expected 0 targets after list delete, got %d", len(targets))
	}
}

func TestTargets_ImportCSV(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Import Test"})

	csv := `email,first_name,last_name,department,position
alice@acme.com,Alice,Smith,Engineering,Developer
bob@acme.com,Bob,Jones,Finance,Analyst
carol@acme.com,Carol,Lee,,
`
	result, err := h.Client.ImportTargetsCSV(list.ID, strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ImportTargetsCSV: %v", err)
	}
	if result.Imported != 3 {
		t.Errorf("Imported = %d, want 3", result.Imported)
	}

	targets, err := h.Client.ListTargets(list.ID)
	if err != nil {
		t.Fatalf("ListTargets: %v", err)
	}
	if len(targets) != 3 {
		t.Errorf("expected 3 targets, got %d", len(targets))
	}
}

func TestTargets_ImportCSV_RequiresEmailColumn(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Bad CSV"})

	csv := `name,department
Alice,Engineering
`
	_, err := h.Client.ImportTargetsCSV(list.ID, strings.NewReader(csv))
	wantError(t, err, http.StatusUnprocessableEntity, "email")
}

func TestTargets_ImportCSV_SkipsEmptyEmail(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Skip Test"})

	csv := `email,first_name
alice@acme.com,Alice
,Bob
carol@acme.com,Carol
`
	result, err := h.Client.ImportTargetsCSV(list.ID, strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ImportTargetsCSV: %v", err)
	}
	if result.Imported != 2 {
		t.Errorf("Imported = %d, want 2 (should skip empty email)", result.Imported)
	}
}

func TestTargets_ImportCSV_FlexibleHeaders(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Flex Headers"})

	csv := `Email,First Name,Last Name,Title
alice@acme.com,Alice,Smith,Developer
`
	result, err := h.Client.ImportTargetsCSV(list.ID, strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ImportTargetsCSV: %v", err)
	}
	if result.Imported != 1 {
		t.Errorf("Imported = %d, want 1", result.Imported)
	}

	targets, _ := h.Client.ListTargets(list.ID)
	if targets[0].FirstName != "Alice" {
		t.Errorf("FirstName = %q, want %q", targets[0].FirstName, "Alice")
	}
	if targets[0].Position != "Developer" {
		t.Errorf("Position = %q, want %q (mapped from Title)", targets[0].Position, "Developer")
	}
}

func TestStatus(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	status, err := h.Client.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Version != "test" {
		t.Errorf("Version = %q, want %q", status.Version, "test")
	}
}
