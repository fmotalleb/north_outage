import { useEffect, useState, useCallback } from 'react'

const KEYS = ['city', 'status', 'q', 'date', 'sort', 'provider', 'open']

/**
 * Bi-directional sync between component state and URL search params.
 * - Reads from URL on mount (so shared links restore filters).
 * - Writes to URL on every change via history.replaceState (no scroll,
 *   no extra history entries — keeps the back button clean).
 *
 * Pass a key schema to validate URL values:
 *   schema = {
 *     city: { default: 'all' },                  // any string allowed
 *     status: { default: 'all', values: [...] }, // whitelist
 *     q: { default: '', maxLength: 200 },        // length cap
 *   }
 */
export function useUrlState(schema = {}) {
  // 1. Read initial values from URL
  const readInitial = useCallback(() => {
    if (typeof window === 'undefined') return {}
    const params = new URLSearchParams(window.location.search)
    const out = {}
    for (const key of KEYS) {
      if (!params.has(key)) continue
      const raw = params.get(key)
      if (!schema[key]) {
        out[key] = raw
        continue
      }
      // Whitelist: drop value if it's not in the allowed list
      if (schema[key].values && !schema[key].values.includes(raw)) continue
      // Length cap (string inputs)
      if (typeof schema[key].maxLength === 'number' && raw.length > schema[key].maxLength) continue
      out[key] = raw
    }
    return out
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const [urlState, setUrlState] = useState(readInitial)

  // 2. On mount, also hydrate from localStorage for cross-session memory.
  //    URL params always win over stored values.
  useEffect(() => {
    try {
      const stored = localStorage.getItem('outage-tracker.filters.v1')
      if (!stored) return
      const parsed = JSON.parse(stored)
      setUrlState((current) => ({ ...parsed, ...current }))
    } catch (_) { /* ignore */ }
  }, [])

  // 3. Write to URL + localStorage whenever state changes
  useEffect(() => {
    if (typeof window === 'undefined') return
    const params = new URLSearchParams(window.location.search)
    for (const key of KEYS) {
      const v = urlState[key]
      if (v == null || v === '' || v === schemaDefault(key, schema)) {
        params.delete(key)
      } else {
        params.set(key, v)
      }
    }
    const qs = params.toString()
    const newUrl = qs
      ? `${window.location.pathname}?${qs}${window.location.hash}`
      : `${window.location.pathname}${window.location.hash}`
    window.history.replaceState(null, '', newUrl)

    // Persist to localStorage (excluding transient `open`)
    try {
      const toStore = { ...urlState }
      delete toStore.open
      localStorage.setItem('outage-tracker.filters.v1', JSON.stringify(toStore))
    } catch (_) { /* ignore */ }
  }, [urlState]) // eslint-disable-line react-hooks/exhaustive-deps

  const update = useCallback((patch) => {
    setUrlState((current) => ({ ...current, ...patch }))
  }, [])

  const reset = useCallback(() => {
    setUrlState({})
  }, [])

  return [urlState, update, reset, setUrlState]
}

function schemaDefault(key, schema) {
  if (!schema[key] || !schema[key].default) return ''
  return schema[key].default
}
