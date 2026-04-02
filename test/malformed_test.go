package test_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/travisbale/barb/sdk"
	"github.com/travisbale/barb/test"
)

// rawRequest sends a raw HTTP request using the SDK client's session cookie
// and returns the status code and decoded error message.
func rawRequest(t *testing.T, h *test.Harness, method, path, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, h.Client.BaseURL()+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.HTTPClient().Do(req)
	if err != nil {
		t.Fatalf("sending request: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var errResp sdk.ErrorResponse
	json.Unmarshal(respBody, &errResp)
	return resp.StatusCode, errResp.Error
}

// wantRawError asserts the status code and error message from a rawRequest.
func wantRawError(t *testing.T, gotStatus int, gotMsg string, wantStatus int, wantMsg string) {
	t.Helper()
	if gotStatus != wantStatus {
		t.Errorf("status = %d, want %d", gotStatus, wantStatus)
	}
	if !strings.Contains(gotMsg, wantMsg) {
		t.Errorf("error = %q, want substring %q", gotMsg, wantMsg)
	}
}

// --- Malformed JSON (400) ---

func TestMalformedJSON(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"create campaign", "POST", "/api/campaigns"},
		{"update campaign", "PATCH", "/api/campaigns/some-id"},
		{"create template", "POST", "/api/templates"},
		{"update template", "PATCH", "/api/templates/some-id"},
		{"create smtp profile", "POST", "/api/smtp-profiles"},
		{"update smtp profile", "PATCH", "/api/smtp-profiles/some-id"},
		{"create phishlet", "POST", "/api/phishlets"},
		{"update phishlet", "PATCH", "/api/phishlets/some-id"},
		{"create target list", "POST", "/api/target-lists"},
		{"add target", "POST", "/api/target-lists/some-id/targets"},
		{"enroll miraged", "POST", "/api/miraged"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := rawRequest(t, h, tt.method, tt.path, "{invalid json")
			wantRawError(t, status, msg, http.StatusBadRequest, "invalid request body")
		})
	}
}

// --- Server-side validation (422) ---
// These bypass the SDK's client-side Validate() to test that the server
// validates independently.

func TestServerValidation_CreateCampaign(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"missing name", `{"template_id":"t","smtp_profile_id":"s","target_list_id":"l","redirect_url":"https://x.com"}`, "name: required"},
		{"missing template_id", `{"name":"C","smtp_profile_id":"s","target_list_id":"l","redirect_url":"https://x.com"}`, "template_id: required"},
		{"missing smtp_profile_id", `{"name":"C","template_id":"t","target_list_id":"l","redirect_url":"https://x.com"}`, "smtp_profile_id: required"},
		{"missing target_list_id", `{"name":"C","template_id":"t","smtp_profile_id":"s","redirect_url":"https://x.com"}`, "target_list_id: required"},
		{"missing redirect_url", `{"name":"C","template_id":"t","smtp_profile_id":"s","target_list_id":"l"}`, "redirect_url: required"},
		{"empty body", `{}`, "name: required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := rawRequest(t, h, "POST", "/api/campaigns", tt.body)
			wantRawError(t, status, msg, http.StatusUnprocessableEntity, tt.wantMsg)
		})
	}
}

func TestServerValidation_CreateSMTPProfile(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"missing name", `{"host":"h","from_addr":"f"}`, "name: required"},
		{"missing host", `{"name":"n","from_addr":"f"}`, "host: required"},
		{"missing from_addr", `{"name":"n","host":"h"}`, "from_addr: required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := rawRequest(t, h, "POST", "/api/smtp-profiles", tt.body)
			wantRawError(t, status, msg, http.StatusUnprocessableEntity, tt.wantMsg)
		})
	}
}

func TestServerValidation_CreateTemplate(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"missing name", `{"subject":"S","html_body":"b"}`, "name: required"},
		{"missing subject", `{"name":"T","html_body":"b"}`, "subject: required"},
		{"missing body", `{"name":"T","subject":"S"}`, "body: HTML or text body is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := rawRequest(t, h, "POST", "/api/templates", tt.body)
			wantRawError(t, status, msg, http.StatusUnprocessableEntity, tt.wantMsg)
		})
	}
}

func TestServerValidation_EnrollMiraged(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	tests := []struct {
		name    string
		body    string
		wantMsg string
	}{
		{"missing name", `{"address":"127.0.0.1:443","secret_hostname":"h","token":"t"}`, "name: required"},
		{"missing address", `{"name":"M","secret_hostname":"h","token":"t"}`, "address: required"},
		{"bad address format", `{"name":"M","address":"no-port","secret_hostname":"h","token":"t"}`, "host:port format"},
		{"missing secret_hostname", `{"name":"M","address":"127.0.0.1:443","token":"t"}`, "secret_hostname: required"},
		{"missing token", `{"name":"M","address":"127.0.0.1:443","secret_hostname":"h"}`, "token: required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg := rawRequest(t, h, "POST", "/api/miraged", tt.body)
			wantRawError(t, status, msg, http.StatusUnprocessableEntity, tt.wantMsg)
		})
	}
}

func TestServerValidation_UpdateMiraged(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	status, msg := rawRequest(t, h, "PATCH", "/api/miraged/some-id", `{"name":""}`)
	wantRawError(t, status, msg, http.StatusUnprocessableEntity, "name: cannot be empty")
}

func TestServerValidation_AddTarget(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	list, _ := h.Client.CreateTargetList(sdk.CreateTargetListRequest{Name: "Test"})

	status, msg := rawRequest(t, h, "POST", "/api/target-lists/"+list.ID+"/targets", `{"first_name":"Bob"}`)
	wantRawError(t, status, msg, http.StatusUnprocessableEntity, "email: required")
}

func TestServerValidation_CreateTargetList(t *testing.T) {
	t.Parallel()
	h := test.NewHarness(t)

	status, msg := rawRequest(t, h, "POST", "/api/target-lists", `{}`)
	wantRawError(t, status, msg, http.StatusUnprocessableEntity, "name: required")
}
