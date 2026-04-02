package api

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/travisbale/barb/internal/phishing"
	"github.com/travisbale/barb/sdk"
)

// Router is the HTTP handler for the Barb API.
type Router struct {
	Miraged   *phishing.MiragedService
	Campaigns *phishing.CampaignService
	Targets   *phishing.TargetService
	Templates *phishing.TemplateService
	Phishlets *phishing.PhishletService
	SMTP      *phishing.SMTPService
	Dashboard *phishing.DashboardService
	Auth      *phishing.AuthService
	Version   string
	Logger    *slog.Logger

	once    sync.Once
	handler http.Handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.once.Do(func() {
		mux := http.NewServeMux()
		r.registerRoutes(mux)
		r.handler = mux
	})

	r.handler.ServeHTTP(w, req)
}

func (r *Router) registerRoutes(mux *http.ServeMux) {
	public := func(method, route string, handler http.HandlerFunc) {
		mux.HandleFunc(method+" "+route, handler)
	}
	auth := func(method, route string, handler http.HandlerFunc) {
		mux.HandleFunc(method+" "+route, r.requireAuth(handler))
	}

	// Public
	public("POST", sdk.RouteLogin, r.loginHandler)
	public("GET", sdk.RouteStatus, r.getStatus)

	// Auth (protected)
	auth("POST", sdk.RouteLogout, r.logoutHandler)
	auth("GET", sdk.RouteMe, r.meHandler)
	auth("POST", sdk.RouteChangePassword, r.changePasswordHandler)

	// Dashboard
	auth("GET", sdk.RouteDashboard, r.getDashboard)

	// Target lists
	auth("GET", sdk.RouteTargetLists, r.listTargetLists)
	auth("POST", sdk.RouteTargetLists, r.createTargetList)
	auth("GET", sdk.RouteTargetList, r.getTargetList)
	auth("DELETE", sdk.RouteTargetList, r.deleteTargetList)
	auth("GET", sdk.RouteTargets, r.listTargets)
	auth("POST", sdk.RouteTargets, r.addTarget)
	auth("POST", sdk.RouteTargetsImport, r.importTargets)
	auth("DELETE", sdk.RouteTarget, r.deleteTarget)

	// Email Templates
	auth("GET", sdk.RouteTemplates, r.listTemplates)
	auth("POST", sdk.RouteTemplates, r.createTemplate)
	auth("POST", sdk.RouteTemplatePreview, r.previewTemplate)
	auth("GET", sdk.RouteTemplate, r.getTemplate)
	auth("PATCH", sdk.RouteTemplate, r.updateTemplate)
	auth("DELETE", sdk.RouteTemplate, r.deleteTemplate)

	// Phishlets
	auth("GET", sdk.RoutePhishlets, r.listPhishlets)
	auth("POST", sdk.RoutePhishlets, r.createPhishlet)
	auth("GET", sdk.RoutePhishlet, r.getPhishlet)
	auth("PATCH", sdk.RoutePhishlet, r.updatePhishlet)
	auth("DELETE", sdk.RoutePhishlet, r.deletePhishlet)

	// SMTP Profiles
	auth("GET", sdk.RouteSMTPProfiles, r.listSMTPProfiles)
	auth("POST", sdk.RouteSMTPProfiles, r.createSMTPProfile)
	auth("GET", sdk.RouteSMTPProfile, r.getSMTPProfile)
	auth("PATCH", sdk.RouteSMTPProfile, r.updateSMTPProfile)
	auth("DELETE", sdk.RouteSMTPProfile, r.deleteSMTPProfile)

	// Campaigns
	auth("GET", sdk.RouteCampaigns, r.listCampaigns)
	auth("POST", sdk.RouteCampaigns, r.createCampaign)
	auth("GET", sdk.RouteCampaign, r.getCampaign)
	auth("PATCH", sdk.RouteCampaign, r.updateCampaign)
	auth("DELETE", sdk.RouteCampaign, r.deleteCampaign)
	auth("POST", sdk.RouteCampaignStart, r.startCampaign)
	auth("POST", sdk.RouteCampaignCancel, r.cancelCampaign)
	auth("POST", sdk.RouteCampaignTestEmail, r.sendTestEmail)
	auth("GET", sdk.RouteCampaignResults, r.listCampaignResults)

	// Miraged connections
	auth("GET", sdk.RouteMiraged, r.listMiraged)
	auth("POST", sdk.RouteMiraged, r.enrollMiraged)
	auth("DELETE", sdk.RouteMiragedInstance, r.deleteMiraged)
	auth("GET", sdk.RouteMiragedStatus, r.testMiraged)
	auth("GET", sdk.RouteMiragedDNSProviders, r.listMiragedDNSProviders)
	auth("POST", sdk.RouteMiragedPhishlets, r.pushMiragedPhishlet)
	auth("POST", sdk.RouteMiragedPhishletEnable, r.enableMiragedPhishlet)
	auth("POST", sdk.RouteMiragedPhishletDisable, r.disableMiragedPhishlet)
	auth("GET", sdk.RouteMiragedSession, r.getMiragedSession)
	auth("GET", sdk.RouteMiragedSessionExport, r.exportMiragedSessionCookies)
}
