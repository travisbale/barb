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

// --- System ---

export interface Status {
  version: string
}

export function getStatus(): Promise<Status> {
  return request('GET', '/status')
}
