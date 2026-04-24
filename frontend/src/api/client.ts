const BASE = '/api'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
  }
  if (body) {
    opts.body = JSON.stringify(body)
  }

  const resp = await fetch(BASE + path, opts)

  if (resp.status === 401 && !path.startsWith('/auth/')) {
    window.location.href = '/login'
    throw new Error('session expired')
  }

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: `HTTP ${resp.status}` }))
    throw new Error(err.error || `HTTP ${resp.status}`)
  }

  if (resp.status === 204) return undefined as T
  return resp.json()
}

// --- Auth ---

export interface AuthUser {
  username: string
  password_change_required: boolean
}

export function login(username: string, password: string): Promise<void> {
  return request('POST', '/auth/login', { username, password })
}

export function logout(): Promise<void> {
  return request('POST', '/auth/logout')
}

export function me(): Promise<AuthUser> {
  return request('GET', '/auth/me')
}

export function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  return request('POST', '/auth/password', { current_password: currentPassword, new_password: newPassword })
}

// --- Target Lists ---

export interface TargetList {
  id: string
  name: string
  created_at: string
}

export interface Target {
  id: string
  list_id: string
  email: string
  first_name: string
  last_name: string
  department: string
  position: string
}

export function listTargetLists(): Promise<TargetList[]> {
  return request('GET', '/target-lists')
}

export function createTargetList(name: string): Promise<TargetList> {
  return request('POST', '/target-lists', { name })
}

export function getTargetList(id: string): Promise<TargetList> {
  return request('GET', `/target-lists/${id}`)
}

export function updateTargetList(id: string, data: { name: string }): Promise<TargetList> {
  return request('PATCH', `/target-lists/${id}`, data)
}

export function deleteTargetList(id: string): Promise<void> {
  return request('DELETE', `/target-lists/${id}`)
}

export function listTargets(listId: string): Promise<Target[]> {
  return request('GET', `/target-lists/${listId}/targets`)
}

export function addTarget(listId: string, target: Omit<Target, 'id' | 'list_id'>): Promise<Target> {
  return request('POST', `/target-lists/${listId}/targets`, target)
}

export function deleteTarget(id: string): Promise<void> {
  return request('DELETE', `/targets/${id}`)
}

export interface ImportResult {
  imported: number
}

export async function importTargetsCSV(listId: string, file: File): Promise<ImportResult> {
  const form = new FormData()
  form.append('file', file)

  const resp = await fetch(`${BASE}/target-lists/${listId}/import`, {
    method: 'POST',
    body: form,
    credentials: 'include',
  })

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: `HTTP ${resp.status}` }))
    throw new Error(err.error || `HTTP ${resp.status}`)
  }

  return resp.json()
}

// --- Email Templates ---

export interface EmailTemplate {
  id: string
  name: string
  subject: string
  html_body: string
  text_body: string
  envelope_sender: string
  created_at: string
}

export function listTemplates(): Promise<EmailTemplate[]> {
  return request('GET', '/templates')
}

export function createTemplate(template: Omit<EmailTemplate, 'id' | 'created_at'>): Promise<EmailTemplate> {
  return request('POST', '/templates', template)
}

export function getTemplate(id: string): Promise<EmailTemplate> {
  return request('GET', `/templates/${id}`)
}

export function updateTemplate(id: string, template: Partial<Omit<EmailTemplate, 'id' | 'created_at'>>): Promise<EmailTemplate> {
  return request('PATCH', `/templates/${id}`, template)
}

export interface PreviewResult {
  subject: string
  html_body: string
  text_body: string
}

export function previewTemplate(id: string, data: { first_name: string; last_name: string; email: string; url: string }): Promise<PreviewResult> {
  return request('POST', `/templates/${id}/preview`, data)
}

export function renderTemplateHTML(html_body: string): Promise<{ html_body: string }> {
  return request('POST', '/templates/render', {
    html_body,
    first_name: 'Jane', last_name: 'Doe', email: 'jane.doe@example.com', url: 'https://phish.example.com/lure123',
  })
}

export function deleteTemplate(id: string): Promise<void> {
  return request('DELETE', `/templates/${id}`)
}

// --- Phishlets ---

export interface Phishlet {
  id: string
  name: string
  yaml: string
  created_at: string
}

export function listPhishlets(): Promise<Phishlet[]> {
  return request('GET', '/phishlets')
}

export function createPhishlet(yaml: string): Promise<Phishlet> {
  return request('POST', '/phishlets', { yaml })
}

export function getPhishlet(id: string): Promise<Phishlet> {
  return request('GET', `/phishlets/${id}`)
}

export function updatePhishlet(id: string, yaml: string): Promise<Phishlet> {
  return request('PATCH', `/phishlets/${id}`, { yaml })
}

export function deletePhishlet(id: string): Promise<void> {
  return request('DELETE', `/phishlets/${id}`)
}

// --- SMTP Profiles ---

export interface SMTPProfile {
  id: string
  name: string
  host: string
  port: number
  username: string
  from_addr: string
  from_name: string
  custom_headers?: Record<string, string>
  created_at: string
}

export function listSMTPProfiles(): Promise<SMTPProfile[]> {
  return request('GET', '/smtp-profiles')
}

export function createSMTPProfile(profile: Omit<SMTPProfile, 'id' | 'created_at'> & { password?: string }): Promise<SMTPProfile> {
  return request('POST', '/smtp-profiles', profile)
}

export function getSMTPProfile(id: string): Promise<SMTPProfile> {
  return request('GET', `/smtp-profiles/${id}`)
}

export function updateSMTPProfile(id: string, profile: Partial<Omit<SMTPProfile, 'id' | 'created_at'>> & { password?: string }): Promise<SMTPProfile> {
  return request('PATCH', `/smtp-profiles/${id}`, profile)
}

export function deleteSMTPProfile(id: string): Promise<void> {
  return request('DELETE', `/smtp-profiles/${id}`)
}

// --- Campaigns ---

export interface Campaign {
  id: string
  name: string
  status: string
  template_id: string
  smtp_profile_id: string
  target_list_id: string
  miraged_id: string
  phishlet: string
  redirect_url: string
  lure_url: string
  send_rate: number
  created_at: string
  started_at: string | null
  completed_at: string | null
}

export interface CampaignResult {
  id: string
  campaign_id: string
  target_id: string
  email: string
  status: string
  sent_at: string | null
  clicked_at: string | null
  captured_at: string | null
  session_id: string
}

export function listCampaigns(): Promise<Campaign[]> {
  return request('GET', '/campaigns')
}

export function createCampaign(campaign: {
  name: string
  template_id: string
  smtp_profile_id: string
  target_list_id: string
  miraged_id?: string
  phishlet?: string
  redirect_url?: string
  send_rate?: number
}): Promise<Campaign> {
  return request('POST', '/campaigns', campaign)
}

export function getCampaign(id: string): Promise<Campaign> {
  return request('GET', `/campaigns/${id}`)
}

export function startCampaign(id: string): Promise<void> {
  return request('POST', `/campaigns/${id}/start`)
}

export function completeCampaign(id: string): Promise<void> {
  return request('POST', `/campaigns/${id}/complete`)
}

export function cancelCampaign(id: string): Promise<void> {
  return request('POST', `/campaigns/${id}/cancel`)
}

export function sendTestEmail(campaignId: string, email: string): Promise<void> {
  return request('POST', `/campaigns/${campaignId}/test-email`, { email })
}

export function updateCampaign(id: string, updates: {
  name?: string
  template_id?: string
  smtp_profile_id?: string
  target_list_id?: string
  miraged_id?: string
  phishlet?: string
  redirect_url?: string
  send_rate?: number
}): Promise<Campaign> {
  return request('PATCH', `/campaigns/${id}`, updates)
}

export function deleteCampaign(id: string): Promise<void> {
  return request('DELETE', `/campaigns/${id}`)
}

export function listCampaignResults(id: string): Promise<CampaignResult[]> {
  return request('GET', `/campaigns/${id}/results`)
}

// --- Miraged Connections ---

export interface MiragedConnection {
  id: string
  name: string
  address: string
  secret_hostname: string
  created_at: string
}

export interface MiragedStatus {
  connected: boolean
  version?: string
  error?: string
}

export interface MiragedPhishlet {
  name: string
  hostname: string
  base_domain: string
  dns_provider: string
  spoof_url: string
  enabled: boolean
}

export function listMiraged(): Promise<MiragedConnection[]> {
  return request('GET', '/miraged')
}

export function enrollMiraged(conn: {
  name: string
  address: string
  secret_hostname: string
  token: string
}): Promise<MiragedConnection> {
  return request('POST', '/miraged', conn)
}

export function updateMiraged(id: string, data: { name: string }): Promise<MiragedConnection> {
  return request('PATCH', `/miraged/${id}`, data)
}

export function deleteMiraged(id: string): Promise<void> {
  return request('DELETE', `/miraged/${id}`)
}

export function testMiraged(id: string): Promise<MiragedStatus> {
  return request('GET', `/miraged/${id}/status`)
}

export function listMiragedDNSProviders(id: string): Promise<string[]> {
  return request('GET', `/miraged/${id}/dns/providers`)
}

export function pushMiragedPhishlet(id: string, yaml: string): Promise<void> {
  return request('POST', `/miraged/${id}/phishlets`, { yaml })
}

export function enableMiragedPhishlet(id: string, name: string, hostname: string, dnsProvider: string): Promise<MiragedPhishlet> {
  return request('POST', `/miraged/${id}/phishlets/${name}/enable`, { hostname, dns_provider: dnsProvider })
}

export function disableMiragedPhishlet(id: string, name: string): Promise<MiragedPhishlet> {
  return request('POST', `/miraged/${id}/phishlets/${name}/disable`)
}

// --- Miraged Notification Channels ---

export interface MiragedNotificationChannel {
  id: string
  type: 'webhook' | 'slack'
  url: string
  filter: string[]
  enabled: boolean
  created_at: string
}

export function listMiragedNotifications(id: string): Promise<MiragedNotificationChannel[]> {
  return request('GET', `/miraged/${id}/notifications/channels`)
}

export function createMiragedNotification(id: string, channel: {
  type: 'webhook' | 'slack'
  url: string
  auth_header?: string
  filter?: string[]
}): Promise<MiragedNotificationChannel> {
  return request('POST', `/miraged/${id}/notifications/channels`, channel)
}

export function deleteMiragedNotification(id: string, channelId: string): Promise<void> {
  return request('DELETE', `/miraged/${id}/notifications/channels/${channelId}`)
}

export function testMiragedNotification(id: string, channelId: string): Promise<void> {
  return request('POST', `/miraged/${id}/notifications/channels/${channelId}/test`)
}

export function listMiragedNotificationEventTypes(id: string): Promise<string[]> {
  return request('GET', `/miraged/${id}/notifications/event-types`)
}

// --- Sessions ---

export interface MiragedSession {
  id: string
  phishlet: string
  remote_addr: string
  user_agent: string
  username: string
  password: string
  custom?: Record<string, string>
  cookie_tokens?: Record<string, Record<string, string>>
  body_tokens?: Record<string, string>
  http_tokens?: Record<string, string>
  started_at: string
  completed_at?: string
}

export function getMiragedSession(connectionId: string, sessionId: string): Promise<MiragedSession> {
  return request('GET', `/miraged/${connectionId}/sessions/${sessionId}`)
}

export function exportMiragedSessionCookies(connectionId: string, sessionId: string): string {
  return `/api/miraged/${encodeURIComponent(connectionId)}/sessions/${encodeURIComponent(sessionId)}/export`
}

// --- Dashboard ---

export interface DashboardStats {
  campaigns: { draft: number; active: number; completed: number; cancelled: number; total: number }
  total_completions: number
  total_clicks: number
  total_emails_sent: number
  miraged_count: number
  active_campaigns: { id: string; name: string; sent: number; failed: number; captured: number; completed: number; total: number }[]
  recent_captures: { email: string; campaign_id: string; campaign_name: string; captured_at: string; session_id: string }[]
}

export function getDashboard(): Promise<DashboardStats> {
  return request('GET', '/dashboard')
}

// --- System ---

export interface Status {
  version: string
}

export function getStatus(): Promise<Status> {
  return request('GET', '/status')
}
