import { validateList, EVENT_SCHEMA } from './validateApi'

const ENDPOINT = '/api/events'

/**
 * Fetch the outage list for a city and normalize it into a sorted array.
 *
 * Handles the two response shapes the server has historically returned
 * (bare array, or `{ data: [...] }`), validates it in dev mode, and sorts
 * events by start time so every caller gets consistent ordering.
 */
export async function fetchEventList(city, { signal } = {}) {
  const url = `${ENDPOINT}?city=${encodeURIComponent(city)}`
  const r = await fetch(url, { signal })
  if (!r.ok) throw new Error(`HTTP ${r.status}`)
  const j = await r.json()
  const list = Array.isArray(j) ? j : Array.isArray(j?.data) ? j.data : []
  validateList(list, EVENT_SCHEMA, '/api/events response')
  list.sort((a, b) => new Date(a.start_at) - new Date(b.start_at))
  return list
}