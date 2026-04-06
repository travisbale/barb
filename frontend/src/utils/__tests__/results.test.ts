import { describe, it, expect } from 'vitest'
import { countResults, completionRate, resultsToCSV } from '../results'

describe('countResults', () => {
  it('returns zeros for empty list', () => {
    expect(countResults([])).toEqual({ sent: 0, clicked: 0, captured: 0, completed: 0 })
  })

  it('does not count pending results as sent', () => {
    const results = [{ status: 'pending' }, { status: 'pending' }]
    expect(countResults(results)).toEqual({ sent: 0, clicked: 0, captured: 0, completed: 0 })
  })

  it('counts sent results (not pending, not clicked)', () => {
    const results = [{ status: 'sent' }, { status: 'failed' }]
    expect(countResults(results)).toEqual({ sent: 2, clicked: 0, captured: 0, completed: 0 })
  })

  it('counts clicked results as both sent and clicked', () => {
    const results = [{ status: 'clicked' }]
    expect(countResults(results)).toEqual({ sent: 1, clicked: 1, captured: 0, completed: 0 })
  })

  it('counts captured results as sent, clicked, and captured', () => {
    const results = [{ status: 'captured' }]
    expect(countResults(results)).toEqual({ sent: 1, clicked: 1, captured: 1, completed: 0 })
  })

  it('counts completed results as all categories', () => {
    const results = [{ status: 'completed' }]
    expect(countResults(results)).toEqual({ sent: 1, clicked: 1, captured: 1, completed: 1 })
  })

  it('handles a realistic mix of statuses', () => {
    const results = [
      { status: 'pending' },
      { status: 'sent' },
      { status: 'sent' },
      { status: 'failed' },
      { status: 'clicked' },
      { status: 'clicked' },
      { status: 'captured' },
      { status: 'completed' },
      { status: 'completed' },
    ]
    expect(countResults(results)).toEqual({ sent: 8, clicked: 5, captured: 3, completed: 2 })
  })
})

describe('completionRate', () => {
  it('returns 0 for empty results', () => {
    expect(completionRate([])).toBe(0)
  })

  it('returns 0 when no results are sent', () => {
    expect(completionRate([{ status: 'pending' }])).toBe(0)
  })

  it('returns 0 when sent but none completed', () => {
    expect(completionRate([{ status: 'sent' }, { status: 'clicked' }])).toBe(0)
  })

  it('returns correct rate', () => {
    const results = [
      { status: 'sent' },
      { status: 'clicked' },
      { status: 'completed' },
      { status: 'completed' },
    ]
    expect(completionRate(results)).toBe(0.5)
  })

  it('returns 1.0 when all results completed', () => {
    const results = [{ status: 'completed' }, { status: 'completed' }]
    expect(completionRate(results)).toBe(1)
  })
})

describe('resultsToCSV', () => {
  it('produces header row for empty results', () => {
    expect(resultsToCSV([])).toBe('email,status,sent_at,clicked_at,captured_at,session_id')
  })

  it('formats a single result row', () => {
    const results = [{
      email: 'alice@example.com',
      status: 'sent',
      sent_at: '2025-01-01T00:00:00Z',
      session_id: 'sess-1',
    }]
    const csv = resultsToCSV(results)
    const lines = csv.split('\n')
    expect(lines).toHaveLength(2)
    expect(lines[1]).toBe('"alice@example.com","sent","2025-01-01T00:00:00Z","","","sess-1"')
  })

  it('escapes double quotes in fields', () => {
    const results = [{
      email: 'bob"quotes@example.com',
      status: 'sent',
      session_id: '',
    }]
    const csv = resultsToCSV(results)
    expect(csv).toContain('"bob""quotes@example.com"')
  })

  it('handles missing optional date fields', () => {
    const results = [{
      email: 'carol@example.com',
      status: 'pending',
      session_id: '',
    }]
    const csv = resultsToCSV(results)
    const lines = csv.split('\n')
    expect(lines[1]).toBe('"carol@example.com","pending","","","",""')
  })
})
