const resultStatusColors: Record<string, string> = {
  pending: 'text-dim',
  sent: 'text-muted',
  failed: 'text-danger',
  clicked: 'text-amber',
  captured: 'text-teal',
  completed: 'text-green',
}

export function resultStatusColor(status: string): string {
  return resultStatusColors[status] ?? 'text-dim'
}

// Counts campaign results by status in a single pass.
// Each status implies all prior statuses: completed > captured > clicked > sent.
export function countResults(results: { status: string }[]) {
  let sent = 0, clicked = 0, captured = 0, completed = 0
  for (const r of results) {
    if (r.status === 'completed') { sent++; clicked++; captured++; completed++ }
    else if (r.status === 'captured') { sent++; clicked++; captured++ }
    else if (r.status === 'clicked') { sent++; clicked++ }
    else if (r.status !== 'pending') { sent++ }
  }
  return { sent, clicked, captured, completed }
}

export function completionRate(results: { status: string }[]) {
  const { sent, completed } = countResults(results)
  return sent > 0 ? completed / sent : 0
}

// Formats campaign results as a CSV string.
export function resultsToCSV(results: { email: string; status: string; sent_at?: string | null; clicked_at?: string | null; captured_at?: string | null; session_id: string }[]) {
  const headers = ['email', 'status', 'sent_at', 'clicked_at', 'captured_at', 'session_id']
  const rows = results.map(result => [
    result.email,
    result.status,
    result.sent_at ?? '',
    result.clicked_at ?? '',
    result.captured_at ?? '',
    result.session_id,
  ].map(field => `"${String(field).replace(/"/g, '""')}"`).join(','))
  return [headers.join(','), ...rows].join('\n')
}
