package sdk

import "time"

// --- Target Lists ---

type CreateTargetListRequest struct {
	Name string `json:"name"`
}

type TargetListResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type AddTargetRequest struct {
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

type TargetResponse struct {
	ID         string `json:"id"`
	ListID     string `json:"list_id"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

type ImportTargetsResponse struct {
	Imported int `json:"imported"`
}

// --- System ---

type StatusResponse struct {
	Version string `json:"version"`
}

// --- Errors ---

type ErrorResponse struct {
	Error string `json:"error"`
}
