package api

import (
	"encoding/json"
	"net/http"

	"github.com/travisbale/barb/sdk"
)

type validatable interface {
	Validate() error
}

// decodeAndValidate decodes a JSON request body into T and calls Validate on it.
// On failure it writes the appropriate error response and returns false.
func decodeAndValidate[T validatable](w http.ResponseWriter, req *http.Request) (T, bool) {
	var body T
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, sdk.ErrorResponse{Error: "Invalid request body."})
		return body, false
	}

	if err := body.Validate(); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, sdk.ErrorResponse{Error: err.Error()})
		return body, false
	}

	return body, true
}
