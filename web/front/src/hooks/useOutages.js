import { useEffect, useState, useCallback } from 'react'
import { validateList, EVENT_SCHEMA } from '../utils/validateApi'

const ENDPOINT = '/api/events'

export function useOutages(city = "ساری") {
  const [data, setData] = useState(null)
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async (signal) => {
    setLoading(true)
    setError(null)
    try {
      const r = await fetch(ENDPOINT + `?city=${city}`, { signal })
      if (!r.ok) throw new Error(`HTTP ${r.status}`)
      const j = await r.json()
      const list = Array.isArray(j) ? j : Array.isArray(j?.data) ? j.data : []
      validateList(list, EVENT_SCHEMA, '/api/events response')
      list.sort((a, b) => new Date(a.start_at) - new Date(b.start_at))
      setData(list)
    } catch (e) {
      if (e.name !== 'AbortError') {
        setError(e.message || 'خطا در دریافت اطلاعات')
        setData([])
      }
    } finally {
      setLoading(false)
    }
  }, [city])

  useEffect(() => {
    const ctrl = new AbortController()
    load(ctrl.signal)
    return () => ctrl.abort()
  }, [load]) // load changes when city changes, so this re-runs automatically

  return { outages: data ?? [], error, loading, refresh: () => load() }
}