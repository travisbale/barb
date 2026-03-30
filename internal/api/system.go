package api

import (
	"net/http"

	"github.com/travisbale/barb/sdk"
)

func (r *Router) getStatus(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, http.StatusOK, sdk.StatusResponse{Version: r.Version})
}
