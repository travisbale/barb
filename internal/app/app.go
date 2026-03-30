package app

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/travisbale/barb/internal/api"
	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/internal/store/sqlite"
)

// Config holds the parameters needed to construct an App.
type Config struct {
	DB       *sqlite.DB
	Frontend fs.FS
	Mailer   phishing.Mailer
	Version  string
	Logger   *slog.Logger
}

// App owns the service graph and HTTP handler.
type App struct {
	Campaigns *phishing.CampaignService
	handler   http.Handler
}

// New constructs all services and wires them into the API router.
func New(cfg Config) *App {
	targetStore := sqlite.NewTargetStore(cfg.DB)
	templateStore := sqlite.NewTemplateStore(cfg.DB)
	smtpStore := sqlite.NewSMTPStore(cfg.DB)
	campaignStore := sqlite.NewCampaignStore(cfg.DB)
	miragedStore := sqlite.NewMiragedStore(cfg.DB)
	phishletStore := sqlite.NewPhishletStore(cfg.DB)

	targetSvc := &phishing.TargetService{Store: targetStore}
	templateSvc := &phishing.TemplateService{Store: templateStore}
	smtpSvc := &phishing.SMTPService{Store: smtpStore}
	miragedSvc := &phishing.MiragedService{Store: miragedStore}
	phishletSvc := &phishing.PhishletService{Store: phishletStore}

	monitor := &phishing.SessionMonitor{
		Campaigns: campaignStore,
		Miraged:   miragedSvc,
		Logger:    cfg.Logger,
	}

	campaignSvc := &phishing.CampaignService{
		Store:     campaignStore,
		Targets:   targetStore,
		Templates: templateStore,
		SMTP:      smtpStore,
		Phishlets: phishletStore,
		Miraged:   miragedSvc,
		Monitor:   monitor,
		Mailer:    cfg.Mailer,
		Logger:    cfg.Logger,
	}

	router := &api.Router{
		Miraged:   miragedSvc,
		Campaigns: campaignSvc,
		Targets:   targetSvc,
		Templates: templateSvc,
		Phishlets: phishletSvc,
		SMTP:      smtpSvc,
		Version:   cfg.Version,
		Logger:    cfg.Logger,
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", router)
	if cfg.Frontend != nil {
		frontendHandler := http.FileServerFS(cfg.Frontend)
		mux.Handle("/", spaHandler(frontendHandler, cfg.Frontend))
	}

	return &App{
		Campaigns: campaignSvc,
		handler:   mux,
	}
}

// Handler returns the HTTP handler for the application.
func (a *App) Handler() http.Handler {
	return a.handler
}

// Shutdown cancels all running campaigns.
func (a *App) Shutdown() {
	a.Campaigns.Shutdown()
}

// spaHandler serves static files, falling back to index.html for SPA routing.
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
