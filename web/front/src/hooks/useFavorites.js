import { useState, useCallback, useEffect, useRef } from 'react'

const STORAGE_KEY = 'outage-tracker.favorites.v2'

/**
 * Each favorite stores:
 *   id:        unique_hash of the outage
 *   city:      city name
 *   q:         search term (address query) used when favorited
 *   address:   outage address
 *   start_at:  outage start time
 *   end_at:    outage end time
 *   favoritedAt: ISO timestamp when it was favorited
 */
function loadFavorites() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch (_) {
    return []
  }
}

function saveFavorites(list) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(list))
  } catch (_) { /* quota or private mode */ }
}

export function useFavorites() {
  const [favorites, setFavorites] = useState(loadFavorites)
  const [favoriteOutages, setFavoriteOutages] = useState([])
  const [favoritesLoading, setFavoritesLoading] = useState(false)
  const isInitialMount = useRef(true)

  // Persist on every change (skip first render)
  useEffect(() => {
    if (isInitialMount.current) {
      isInitialMount.current = false
      return
    }
    saveFavorites(favorites)
  }, [favorites])

  // Cross-tab sync
  useEffect(() => {
    function onStorage(e) {
      if (e.key !== STORAGE_KEY || !e.newValue) return
      try {
        const parsed = JSON.parse(e.newValue)
        if (Array.isArray(parsed)) setFavorites(parsed)
      } catch (_) { /* ignore */ }
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])
  // Fetch live data for all favorited cities
  const refreshFavorites = useCallback(async () => {
    const favs = loadFavorites()
    if (favs.length === 0) {
      setFavoriteOutages([])
      return
    }

    setFavoritesLoading(true)

    // Collect unique cities
    const cities = [...new Set(favs.map((f) => f.city))]
    const favIdSet = new Set(favs.map((f) => f.id))

    try {
      const results = await Promise.allSettled(
        cities.map((city) =>
          fetch(`/api/events?city=${city}`).then((r) => {
            if (!r.ok) throw new Error(`HTTP ${r.status}`)
            return r.json()
          }),
        ),
      )

      const matched = []
      for (const result of results) {
        if (result.status !== 'fulfilled') continue
        const j = result.value
        const list = Array.isArray(j) ? j : Array.isArray(j?.data) ? j.data : []
        for (const item of list) {
          if (favIdSet.has(item.unique_hash)) {
            matched.push(item)
          }
        }
      }

      // Sort by start time ascending
      matched.sort((a, b) => new Date(a.start_at) - new Date(b.start_at))
      setFavoriteOutages(matched)
    } catch (_) {
      // Silently fail — just show whatever we already have
    } finally {
      setFavoritesLoading(false)
    }
  }, [])

  // Auto-refresh on mount
  useEffect(() => {
    refreshFavorites()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const addFavorite = useCallback(
    (outage, searchTerm) => {
      const id = outage.unique_hash
      setFavorites((prev) => {
        if (prev.some((f) => f.id === id)) return prev
        return [
          ...prev,
          {
            id,
            city: outage.city,
            q: searchTerm || '',
            address: outage.address || '',
            start_at: outage.start_at,
            end_at: outage.end_at,
            favoritedAt: new Date().toISOString(),
          },
        ]
      })
      // Immediately add the outage to the live list so it appears right away
      setFavoriteOutages((prev) => {
        if (prev.some((o) => o.unique_hash === id)) return prev
        return [...prev, outage].sort(
          (a, b) => new Date(a.start_at) - new Date(b.start_at),
        )
      })
      // Also schedule a full re-fetch to pick up any other changes
      setTimeout(() => refreshFavorites(), 2000)
    },
    [refreshFavorites],
  )

  const removeFavorite = useCallback(
    (id) => {
      setFavorites((prev) => prev.filter((f) => f.id !== id))
      setFavoriteOutages((prev) => prev.filter((o) => o.unique_hash !== id))
    },
    [],
  )

  const isFavorite = useCallback(
    (id) => favorites.some((f) => f.id === id),
    [favorites],
  )

  const clearAllFavorites = useCallback(() => {
    setFavorites([])
    setFavoriteOutages([])
  }, [])

  return {
    favorites,
    favoriteOutages,
    favoritesLoading,
    addFavorite,
    removeFavorite,
    isFavorite,
    refreshFavorites,
    clearAllFavorites,
  }
}
