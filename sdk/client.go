package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// Client communicates with the Mirador API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client that talks to the Mirador API at baseURL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
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
	w.Close()

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

// --- System ---

func (c *Client) Status() (*StatusResponse, error) {
	return get[StatusResponse](c, RouteStatus)
}

// --- Internals ---

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequest(method, c.baseURL+path, bodyReader)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, c.baseURL+path, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
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
