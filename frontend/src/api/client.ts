const BASE = '/api'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
  }
  if (body) {
    opts.body = JSON.stringify(body)
  }

  const resp = await fetch(BASE + path, opts)

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: `HTTP ${resp.status}` }))
    throw new Error(err.error || `HTTP ${resp.status}`)
  }

  if (resp.status === 204) return undefined as T
  return resp.json()
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

export function updateTemplate(id: string, template: Omit<EmailTemplate, 'id' | 'created_at'>): Promise<EmailTemplate> {
  return request('PATCH', `/templates/${id}`, template)
}

export function deleteTemplate(id: string): Promise<void> {
  return request('DELETE', `/templates/${id}`)
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
  lure_url?: string
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

export function cancelCampaign(id: string): Promise<void> {
  return request('POST', `/campaigns/${id}/cancel`)
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
  enabled: boolean
}

export function listMiraged(): Promise<MiragedConnection[]> {
  return request('GET', '/miraged')
}

export function createMiraged(conn: {
  name: string
  address: string
  secret_hostname: string
  cert_pem: string
  key_pem: string
  ca_cert_pem: string
}): Promise<MiragedConnection> {
  return request('POST', '/miraged', conn)
}

export function deleteMiraged(id: string): Promise<void> {
  return request('DELETE', `/miraged/${id}`)
}

export function testMiraged(id: string): Promise<MiragedStatus> {
  return request('GET', `/miraged/${id}/status`)
}

export function listMiragedPhishlets(id: string): Promise<MiragedPhishlet[]> {
  return request('GET', `/miraged/${id}/phishlets`)
}

// --- System ---

export interface Status {
  version: string
}

export function getStatus(): Promise<Status> {
  return request('GET', '/status')
}
