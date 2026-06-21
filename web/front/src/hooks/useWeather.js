import { useEffect, useRef, useState } from 'react'
import { getProvider } from '../utils/weatherProviders'

// Simple in-memory cache so we don't hammer the API when filtering.
const cache = new Map()

function cacheKey(providerId, city, startISO) {
  return `${providerId}|${city}|${startISO}`
}

export function useWeather(providerId, { city, lat, lon, startISO, endISO, enabled = true }) {
  const [state, setState] = useState({
    loading: false,
    data: null,
    error: null,
  })
  const ctrlRef = useRef(null)

  useEffect(() => {
    if (!enabled || !providerId || !startISO || !endISO || !city) return

    const provider = getProvider(providerId)
    const key = cacheKey(providerId, city, startISO)
    if (cache.has(key)) {
      setState({ loading: false, data: cache.get(key), error: null })
      return
    }

    ctrlRef.current?.abort?.()
    const ctrl = new AbortController()
    ctrlRef.current = ctrl
    setState({ loading: true, data: null, error: null })

    provider
      .fetch({ city, lat, lon, startISO, endISO, signal: ctrl.signal })
      .then((res) => {
        cache.set(key, res)
        if (!ctrl.signal.aborted) setState({ loading: false, data: res, error: null })
      })
      .catch((e) => {
        if (ctrl.signal.aborted) return
        setState({
          loading: false,
          data: null,
          error: e.name === 'AbortError' ? null : e.message,
        })
      })

    return () => ctrl.abort()
  }, [providerId, city, lat, lon, startISO, endISO, enabled])

  return state
}
