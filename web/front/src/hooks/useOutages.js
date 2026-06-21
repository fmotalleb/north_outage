import { useEffect, useState, useCallback } from 'react'

const ENDPOINT = '/api/events'

export function useOutages() {
  const [data, setData] = useState(null)
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async (signal) => {
    setLoading(true)
    setError(null)
    try {
      const r = await fetch(ENDPOINT, { signal })
      if (!r.ok) throw new Error(`HTTP ${r.status}`)
      const j = await r.json()
      // API may return either an array directly or { data: [...] }
      const list = Array.isArray(j) ? j : Array.isArray(j?.data) ? j.data : []
      // Normalize + sort by start ascending
      list.sort((a, b) => new Date(a.start) - new Date(b.start))
      setData(list)
    } catch (e) {
      if (e.name !== 'AbortError') {
        setError(e.message || 'خطا در دریافت اطلاعات')
        setData([])
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const ctrl = new AbortController()
    load(ctrl.signal)
    return () => ctrl.abort()
  }, [load])

  return { outages: data ?? [], error, loading, refresh: () => load() }
}
