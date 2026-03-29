package api

import (
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/travisbale/mirador/internal/phishing"
	"github.com/travisbale/mirador/sdk"
	miragesdk "github.com/travisbale/mirage/sdk"
)

type miragedManager interface {
	Create(conn *phishing.MiragedConnection) (*phishing.MiragedConnection, error)
	Get(id string) (*phishing.MiragedConnection, error)
	Delete(id string) error
	List() ([]*phishing.MiragedConnection, error)
	Client(id string) (*miragesdk.Client, error)
	TestConnection(id string) error
}

type campaignManager interface {
	Create(campaign *phishing.Campaign) (*phishing.Campaign, error)
	Get(id string) (*phishing.Campaign, error)
	Delete(id string) error
	List() ([]*phishing.Campaign, error)
	Start(id string) error
	Cancel(id string) error
	Results(campaignID string) ([]*phishing.CampaignResult, error)
}

type templateManager interface {
	Create(template *phishing.EmailTemplate) (*phishing.EmailTemplate, error)
	Get(id string) (*phishing.EmailTemplate, error)
	Update(id string, update *phishing.TemplateUpdate) (*phishing.EmailTemplate, error)
	Delete(id string) error
	List() ([]*phishing.EmailTemplate, error)
}

type smtpManager interface {
	CreateProfile(profile *phishing.SMTPProfile) (*phishing.SMTPProfile, error)
	GetProfile(id string) (*phishing.SMTPProfile, error)
	UpdateProfile(id string, update *phishing.SMTPProfileUpdate) (*phishing.SMTPProfile, error)
	DeleteProfile(id string) error
	ListProfiles() ([]*phishing.SMTPProfile, error)
}

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
	Miraged   miragedManager
	Campaigns campaignManager
	Targets   targetManager
	Templates templateManager
	SMTP      smtpManager
	Version   string
	Logger    *slog.Logger

	once sync.Once
	mux  *http.ServeMux
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.once.Do(func() {
		r.mux = http.NewServeMux()
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

	// Email Templates
	h("GET", sdk.RouteTemplates, r.listTemplates)
	h("POST", sdk.RouteTemplates, r.createTemplate)
	h("GET", sdk.RouteTemplate, r.getTemplate)
	h("PATCH", sdk.RouteTemplate, r.updateTemplate)
	h("DELETE", sdk.RouteTemplate, r.deleteTemplate)

	// SMTP Profiles
	h("GET", sdk.RouteSMTPProfiles, r.listSMTPProfiles)
	h("POST", sdk.RouteSMTPProfiles, r.createSMTPProfile)
	h("GET", sdk.RouteSMTPProfile, r.getSMTPProfile)
	h("PATCH", sdk.RouteSMTPProfile, r.updateSMTPProfile)
	h("DELETE", sdk.RouteSMTPProfile, r.deleteSMTPProfile)

	// Campaigns
	h("GET", sdk.RouteCampaigns, r.listCampaigns)
	h("POST", sdk.RouteCampaigns, r.createCampaign)
	h("GET", sdk.RouteCampaign, r.getCampaign)
	h("DELETE", sdk.RouteCampaign, r.deleteCampaign)
	h("POST", sdk.RouteCampaignStart, r.startCampaign)
	h("POST", sdk.RouteCampaignCancel, r.cancelCampaign)
	h("GET", sdk.RouteCampaignResults, r.listCampaignResults)

	// Miraged connections
	h("GET", sdk.RouteMiraged, r.listMiraged)
	h("POST", sdk.RouteMiraged, r.createMiraged)
	h("DELETE", sdk.RouteMiragedInstance, r.deleteMiraged)
	h("GET", sdk.RouteMiragedStatus, r.testMiraged)
	h("GET", sdk.RouteMiragedPhishlets, r.listMiragedPhishlets)
}
