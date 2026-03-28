package sdk

import "strings"

const (
	// Target lists
	RouteTargetLists   = "/api/target-lists"
	RouteTargetList    = "/api/target-lists/{id}"
	RouteTargets       = "/api/target-lists/{id}/targets"
	RouteTargetsImport = "/api/target-lists/{id}/import"
	RouteTarget        = "/api/targets/{id}"

	// Campaigns
	RouteCampaigns       = "/api/campaigns"
	RouteCampaign        = "/api/campaigns/{id}"
	RouteCampaignStart   = "/api/campaigns/{id}/start"
	RouteCampaignCancel  = "/api/campaigns/{id}/cancel"
	RouteCampaignResults = "/api/campaigns/{id}/results"

	// Email Templates
	RouteTemplates = "/api/templates"
	RouteTemplate  = "/api/templates/{id}"

	// SMTP Profiles
	RouteSMTPProfiles = "/api/smtp-profiles"
	RouteSMTPProfile  = "/api/smtp-profiles/{id}"

	// Miraged connections
	RouteMiraged          = "/api/miraged"
	RouteMiragedInstance  = "/api/miraged/{id}"
	RouteMiragedStatus    = "/api/miraged/{id}/status"
	RouteMiragedPhishlets = "/api/miraged/{id}/phishlets"

	// System
	RouteStatus = "/api/status"
)

// ResolveRoute replaces {param} placeholders in a route pattern with concrete values.
// Parameters are passed as alternating name/value pairs:
//
//	ResolveRoute("/api/target-lists/{id}", "id", "abc123")
func ResolveRoute(pattern string, pairs ...string) string {
	for i := 0; i+1 < len(pairs); i += 2 {
		pattern = strings.ReplaceAll(pattern, "{"+pairs[i]+"}", pairs[i+1])
	}
	return pattern
}
