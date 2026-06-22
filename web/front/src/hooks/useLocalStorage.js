import { useEffect, useRef, useState, useCallback } from 'react'

/**
 * Persist React state in localStorage.
 *
 * - Reads the initial value from localStorage on mount (or falls back to `defaultValue`).
 * - Writes the value to localStorage on every change.
 * - Listens for the `storage` event so updates in another tab sync here too.
 *
 * Optionally takes a `schema` (same shape as the old useUrlState):
 *   schema = {
 *     city:    { default: 'all' },
 *     status:  { default: 'all', values: ['all', 'active', 'upcoming', 'past'] },
 *     sort:    { default: 'start_asc', values: [...] },
 *     q:       { default: '', maxLength: 200 },
 *   }
 */
export function useLocalState(storageKey, schema = {}) {
  const readInitial = useCallback(() => {
    if (typeof window === 'undefined') return {}
    try {
      const raw = window.localStorage.getItem(storageKey)
      if (!raw) return {}
      const parsed = JSON.parse(raw)
      if (!parsed || typeof parsed !== 'object') return {}
      const out = {}
      for (const [key, value] of Object.entries(parsed)) {
        if (!schema[key]) { out[key] = value; continue }
        if (schema[key].values && !schema[key].values.includes(value)) continue
        if (typeof schema[key].maxLength === 'number' && String(value).length > schema[key].maxLength) continue
        out[key] = value
      }
      return out
    } catch (_) { return {} }
  }, [storageKey]) // eslint-disable-line react-hooks/exhaustive-deps

  const [state, setState] = useState(readInitial)
  const isFirstRender = useRef(true)

  // Persist on every change (skip the very first render to avoid clobbering
  // what we just read).
  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false
      return
    }
    try {
      window.localStorage.setItem(storageKey, JSON.stringify(state))
    } catch (_) { /* quota or private mode — silently ignore */ }
  }, [storageKey, state])

  // Cross-tab sync via the `storage` event
  useEffect(() => {
    if (typeof window === 'undefined') return
    function onStorage(e) {
      if (e.key !== storageKey || !e.newValue) return
      try {
        const parsed = JSON.parse(e.newValue)
        if (parsed && typeof parsed === 'object') setState(parsed)
      } catch (_) { /* ignore */ }
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [storageKey])

  const update = useCallback((patch) => {
    setState((current) => ({ ...current, ...patch }))
  }, [])

  const reset = useCallback(() => setState({}), [])

  return [state, update, reset]
}
