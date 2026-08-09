import { useEffect, useState, useCallback } from 'react'
import { fetchEventList } from '../utils/api'
import { DEFAULT_CITY } from '../data/cityCoordinates'

export function useOutages(city = DEFAULT_CITY) {
  const [data, setData] = useState(null)
  const [error, setError] = useState(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async (signal) => {
    setLoading(true)
    setError(null)
    try {
      const list = await fetchEventList(city, { signal })
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