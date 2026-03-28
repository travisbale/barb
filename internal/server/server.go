package server

import (
	"io/fs"
	"net/http"

	"github.com/travisbale/mirador/internal/api"
)

// Config holds all dependencies for the server.
type Config struct {
	Addr     string
	API      *api.Router
	Frontend fs.FS
}

// New creates an HTTP server that serves the API and the Vue SPA.
func New(cfg Config) *http.Server {
	mux := http.NewServeMux()

	// API routes are handled by the router.
	mux.Handle("/api/", cfg.API)

	// Frontend — serve the Vue SPA.
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
			path = path[1:]
		}

		if _, err := fs.Stat(frontend, path); err != nil {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
