package api

import (
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/sdk"
)

type targetManager interface {
	CreateList(name string) (*phishing.TargetList, error)
	GetList(id string) (*phishing.TargetList, error)
	DeleteList(id string) error
	ListLists() ([]*phishing.TargetList, error)
	AddTarget(listID string, target *phishing.Target) error
	ListTargets(listID string) ([]*phishing.Target, error)
	ImportCSV(listID string, r io.Reader) (int, error)
	DeleteTarget(id string) error
}

// Router is the HTTP handler for the Mirador API.
type Router struct {
	Targets targetManager
	Version string
	Logger  *slog.Logger

	once      sync.Once
	mux       *http.ServeMux
	startedAt time.Time
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.once.Do(func() {
		r.mux = http.NewServeMux()
		r.startedAt = time.Now()
		r.registerRoutes()
	})

	r.mux.ServeHTTP(w, req)
}

func (r *Router) registerRoutes() {
	h := func(method, route string, handler http.HandlerFunc) {
		r.mux.HandleFunc(method+" "+route, handler)
	}

	// System
	h("GET", sdk.RouteStatus, r.getStatus)

	// Target lists
	h("GET", sdk.RouteTargetLists, r.listTargetLists)
	h("POST", sdk.RouteTargetLists, r.createTargetList)
	h("GET", sdk.RouteTargetList, r.getTargetList)
	h("DELETE", sdk.RouteTargetList, r.deleteTargetList)
	h("GET", sdk.RouteTargets, r.listTargets)
	h("POST", sdk.RouteTargets, r.addTarget)
	h("POST", sdk.RouteTargetsImport, r.importTargets)
	h("DELETE", sdk.RouteTarget, r.deleteTarget)
}
