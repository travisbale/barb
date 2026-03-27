package server

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/travisbale/mirador/internal/store/sqlite"
)

// Config holds all dependencies for the server.
type Config struct {
	Addr     string
	DB       *sqlite.DB
	Frontend fs.FS
	Logger   *slog.Logger
	Version  string
}

// New creates an HTTP server with all routes registered.
func New(cfg Config) *http.Server {
	mux := http.NewServeMux()

	// API routes.
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version":"` + cfg.Version + `"}`))
	})

	// Frontend — serve the Vue SPA. All non-API routes fall through to index.html.
	frontendHandler := http.FileServerFS(cfg.Frontend)
	mux.Handle("/", spaHandler(frontendHandler, cfg.Frontend))

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}
}

// spaHandler serves static files from the frontend FS. If the requested file
// doesn't exist, it serves index.html so the Vue router can handle the route.
func spaHandler(fileServer http.Handler, frontend fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else {
			path = path[1:] // strip leading slash
		}

		// Check if the file exists in the embedded FS.
		if _, err := fs.Stat(frontend, path); err != nil {
			// File doesn't exist — serve index.html for SPA routing.
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
