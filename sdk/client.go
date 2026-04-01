package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
)

// Client communicates with the Barb API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client that talks to the Barb API at baseURL.
func NewClient(baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Jar: jar},
	}
}

// --- Auth ---

func (c *Client) Login(req LoginRequest) error {
	resp, err := c.do(http.MethodPost, RouteLogin, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

func (c *Client) Me() (*MeResponse, error) {
	return get[MeResponse](c, RouteMe)
}

func (c *Client) ChangePassword(req ChangePasswordRequest) error {
	resp, err := c.do(http.MethodPost, RouteChangePassword, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

// --- Target Lists ---

func (c *Client) CreateTargetList(req CreateTargetListRequest) (*TargetListResponse, error) {
	return send[TargetListResponse](c, http.MethodPost, RouteTargetLists, req)
}

func (c *Client) ListTargetLists() ([]TargetListResponse, error) {
	resp, err := get[[]TargetListResponse](c, RouteTargetLists)
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) GetTargetList(id string) (*TargetListResponse, error) {
	return get[TargetListResponse](c, ResolveRoute(RouteTargetList, "id", id))
}

func (c *Client) DeleteTargetList(id string) error {
	return discard(c, http.MethodDelete, ResolveRoute(RouteTargetList, "id", id))
}

func (c *Client) AddTarget(listID string, req AddTargetRequest) (*TargetResponse, error) {
	return send[TargetResponse](c, http.MethodPost, ResolveRoute(RouteTargets, "id", listID), req)
}

func (c *Client) ListTargets(listID string) ([]TargetResponse, error) {
	resp, err := get[[]TargetResponse](c, ResolveRoute(RouteTargets, "id", listID))
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) ImportTargetsCSV(listID string, csvData io.Reader) (*ImportTargetsResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "targets.csv")
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, csvData); err != nil {
		return nil, fmt.Errorf("copying CSV data: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+ResolveRoute(RouteTargetsImport, "id", listID), &buf)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var result ImportTargetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unable to decode response, expected JSON from API")
	}
	return &result, nil
}

func (c *Client) DeleteTarget(id string) error {
	return discard(c, http.MethodDelete, ResolveRoute(RouteTarget, "id", id))
}

// --- Campaigns ---

func (c *Client) CreateCampaign(req CreateCampaignRequest) (*CampaignResponse, error) {
	return send[CampaignResponse](c, http.MethodPost, RouteCampaigns, req)
}

func (c *Client) ListCampaigns() ([]CampaignResponse, error) {
	resp, err := get[[]CampaignResponse](c, RouteCampaigns)
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) GetCampaign(id string) (*CampaignResponse, error) {
	return get[CampaignResponse](c, ResolveRoute(RouteCampaign, "id", id))
}

func (c *Client) UpdateCampaign(id string, req UpdateCampaignRequest) (*CampaignResponse, error) {
	return send[CampaignResponse](c, http.MethodPatch, ResolveRoute(RouteCampaign, "id", id), req)
}

func (c *Client) StartCampaign(id string) error {
	return discard(c, http.MethodPost, ResolveRoute(RouteCampaignStart, "id", id))
}

func (c *Client) CancelCampaign(id string) error {
	return discard(c, http.MethodPost, ResolveRoute(RouteCampaignCancel, "id", id))
}

func (c *Client) SendTestEmail(campaignID string, req SendTestEmailRequest) error {
	resp, err := c.do(http.MethodPost, ResolveRoute(RouteCampaignTestEmail, "id", campaignID), req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

func (c *Client) DeleteCampaign(id string) error {
	return discard(c, http.MethodDelete, ResolveRoute(RouteCampaign, "id", id))
}

func (c *Client) ListCampaignResults(id string) ([]CampaignResultResponse, error) {
	resp, err := get[[]CampaignResultResponse](c, ResolveRoute(RouteCampaignResults, "id", id))
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

// --- Email Templates ---

func (c *Client) CreateTemplate(req CreateTemplateRequest) (*TemplateResponse, error) {
	return send[TemplateResponse](c, http.MethodPost, RouteTemplates, req)
}

func (c *Client) ListTemplates() ([]TemplateResponse, error) {
	resp, err := get[[]TemplateResponse](c, RouteTemplates)
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) GetTemplate(id string) (*TemplateResponse, error) {
	return get[TemplateResponse](c, ResolveRoute(RouteTemplate, "id", id))
}

func (c *Client) UpdateTemplate(id string, req UpdateTemplateRequest) (*TemplateResponse, error) {
	return send[TemplateResponse](c, http.MethodPatch, ResolveRoute(RouteTemplate, "id", id), req)
}

func (c *Client) PreviewTemplate(id string, req PreviewTemplateRequest) (*PreviewTemplateResponse, error) {
	return send[PreviewTemplateResponse](c, http.MethodPost, ResolveRoute(RouteTemplatePreview, "id", id), req)
}

func (c *Client) DeleteTemplate(id string) error {
	return discard(c, http.MethodDelete, ResolveRoute(RouteTemplate, "id", id))
}

// --- Phishlets ---

func (c *Client) CreatePhishlet(req CreatePhishletRequest) (*PhishletResponse, error) {
	return send[PhishletResponse](c, http.MethodPost, RoutePhishlets, req)
}

func (c *Client) ListPhishlets() ([]PhishletResponse, error) {
	resp, err := get[[]PhishletResponse](c, RoutePhishlets)
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) GetPhishlet(id string) (*PhishletResponse, error) {
	return get[PhishletResponse](c, ResolveRoute(RoutePhishlet, "id", id))
}

func (c *Client) UpdatePhishlet(id string, req UpdatePhishletRequest) (*PhishletResponse, error) {
	return send[PhishletResponse](c, http.MethodPatch, ResolveRoute(RoutePhishlet, "id", id), req)
}

func (c *Client) DeletePhishlet(id string) error {
	return discard(c, http.MethodDelete, ResolveRoute(RoutePhishlet, "id", id))
}

// --- SMTP Profiles ---

func (c *Client) CreateSMTPProfile(req CreateSMTPProfileRequest) (*SMTPProfileResponse, error) {
	return send[SMTPProfileResponse](c, http.MethodPost, RouteSMTPProfiles, req)
}

func (c *Client) ListSMTPProfiles() ([]SMTPProfileResponse, error) {
	resp, err := get[[]SMTPProfileResponse](c, RouteSMTPProfiles)
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) GetSMTPProfile(id string) (*SMTPProfileResponse, error) {
	return get[SMTPProfileResponse](c, ResolveRoute(RouteSMTPProfile, "id", id))
}

func (c *Client) UpdateSMTPProfile(id string, req UpdateSMTPProfileRequest) (*SMTPProfileResponse, error) {
	return send[SMTPProfileResponse](c, http.MethodPatch, ResolveRoute(RouteSMTPProfile, "id", id), req)
}

func (c *Client) DeleteSMTPProfile(id string) error {
	return discard(c, http.MethodDelete, ResolveRoute(RouteSMTPProfile, "id", id))
}

// --- Miraged Connections ---

func (c *Client) EnrollMiraged(req EnrollMiragedRequest) (*MiragedResponse, error) {
	return send[MiragedResponse](c, http.MethodPost, RouteMiraged, req)
}

func (c *Client) ListMiraged() ([]MiragedResponse, error) {
	resp, err := get[[]MiragedResponse](c, RouteMiraged)
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) DeleteMiraged(id string) error {
	return discard(c, http.MethodDelete, ResolveRoute(RouteMiragedInstance, "id", id))
}

func (c *Client) TestMiraged(id string) (*MiragedStatusResponse, error) {
	return get[MiragedStatusResponse](c, ResolveRoute(RouteMiragedStatus, "id", id))
}

func (c *Client) ListMiragedPhishlets(id string) ([]MiragedPhishletResponse, error) {
	resp, err := get[[]MiragedPhishletResponse](c, ResolveRoute(RouteMiragedPhishlets, "id", id))
	if err != nil {
		return nil, err
	}
	return *resp, nil
}

func (c *Client) EnableMiragedPhishlet(connectionID, name string, req EnableMiragedPhishletRequest) (*MiragedPhishletResponse, error) {
	route := ResolveRoute(RouteMiragedPhishletEnable, "id", connectionID, "name", name)
	return send[MiragedPhishletResponse](c, http.MethodPost, route, req)
}

func (c *Client) DisableMiragedPhishlet(connectionID, name string) (*MiragedPhishletResponse, error) {
	route := ResolveRoute(RouteMiragedPhishletDisable, "id", connectionID, "name", name)
	return send[MiragedPhishletResponse](c, http.MethodPost, route, nil)
}

func (c *Client) PushMiragedPhishlet(connectionID string, yaml string) error {
	resp, err := c.do(http.MethodPost, ResolveRoute(RouteMiragedPhishletPush, "id", connectionID), PushMiragedPhishletRequest{YAML: yaml})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

// --- Dashboard ---

func (c *Client) Dashboard() (*DashboardResponse, error) {
	return get[DashboardResponse](c, RouteDashboard)
}

// --- System ---

func (c *Client) Status() (*StatusResponse, error) {
	return get[StatusResponse](c, RouteStatus)
}

// --- Internals ---

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		req, err = http.NewRequest(method, c.baseURL+path, bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, c.baseURL+path, nil)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
	}

	return c.httpClient.Do(req)
}

func get[T any](c *Client, path string) (*T, error) {
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unable to decode response, expected JSON from API")
	}
	return &result, nil
}

func send[T any](c *Client, method, path string, body any) (*T, error) {
	resp, err := c.do(method, path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unable to decode response, expected JSON from API")
	}
	return &result, nil
}

func discard(c *Client, method, path string) error {
	resp, err := c.do(method, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkStatus(resp)
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}
	var e ErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&e)
	if e.Error != "" {
		return fmt.Errorf("api: %s", e.Error)
	}
	return fmt.Errorf("api: HTTP %d", resp.StatusCode)
}
