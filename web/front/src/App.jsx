import { useEffect, useMemo, useState, useCallback } from 'react'
import Header from './components/Header'
import FilterBar from './components/FilterBar'
import OutageList from './components/OutageList'
import FavoriteOutages from './components/FavoriteOutages'
import WeatherProviderSelector from './components/WeatherProviderSelector'
import InstallButton from './components/InstallButton'
import { useOutages } from './hooks/useOutages'
import { useLocalState } from './hooks/useLocalStorage'
import { useFavorites, getLocationId } from './hooks/useFavorites'
import { getKnownCities } from './data/cityCoordinates'
import { outageStatus, durationMinutes } from './utils/dateUtils'
import { validateShape, UPDATED_AT_SCHEMA } from './utils/validateApi'

// Schema for localStorage validation
const SCHEMA = {
  city: { default: 'all' },
  status: { default: 'all', values: ['all', 'active', 'upcoming', 'past', 'active-upcoming'] },
  date: { default: 'all', values: ['all', 'today', 'tomorrow', 'week'] },
  sort: { default: 'start_asc', values: ['start_asc', 'start_desc', 'duration_desc', 'city'] },
  provider: { default: 'met-no', values: ['open-meteo', 'met-no'] },
  q: { default: '', maxLength: 200 },
}

const DEFAULTS = {
  city: 'ساری',
  status: 'all',
  q: '',
  date: 'all',
  sort: 'start_asc',
  provider: 'met-no',
}

const STORAGE_KEY = 'outage-tracker.filters.v1'

export default function App() {
  const [state, updateState] = useLocalState(STORAGE_KEY, SCHEMA)
  const [lastUpdated, setLastUpdated] = useState(null)
  const [now, setNow] = useState(() => new Date())
  const {
    favoriteOutages,
    favoritesLoading,
    addFavorite,
    removeFavorite,
    isFavorite,
    refreshFavorites,
  } = useFavorites()

  // Derive UI state from local state
  const filters = useMemo(() => ({
    city: state.city ?? DEFAULTS.city,
    status: state.status ?? DEFAULTS.status,
    q: state.q ?? DEFAULTS.q,
    date: state.date ?? DEFAULTS.date,
  }), [state])

  const { outages, error, loading, refresh } = useOutages(filters.city)

  // Toggle a location as favorite (add or remove)
  const onToggleFavorite = useCallback(
    (outage) => {
      const locId = getLocationId(outage)
      if (isFavorite(locId)) {
        removeFavorite(locId)
      } else {
        addFavorite(outage, filters.q)
      }
    },
    [isFavorite, addFavorite, removeFavorite, filters.q],
  )

  const sort = state.sort ?? DEFAULTS.sort
  const providerId = state.provider ?? DEFAULTS.provider
  const expandedId = state.open ?? null

  const setFilters = (patch) => {
    if (typeof patch === 'function') patch = patch(filters)
    updateState(patch)
  }
  const setSort = (v) => updateState({ sort: v })
  const setProviderId = (v) => updateState({ provider: v })
  const setExpandedId = (v) => updateState({ open: v })

  // Tick "now" so active/upcoming/past recompute every 30s
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 30_000)
    return () => clearInterval(id)
  }, [])

  // Update lastUpdated whenever a successful fetch lands
  useEffect(() => {
    if (loading || error) return

    const fetchUpdatedAt = async () => {
      try {
        const res = await fetch('/api/updated_at')
        if (!res.ok) return

        const data = await res.json()
        validateShape(data, UPDATED_AT_SCHEMA, '/api/updated_at response')
        if (data?.created_at) {
          setLastUpdated(new Date(data.created_at))
        }
      } catch (e) {
        // silent fail or optional logging
      }
    }

    fetchUpdatedAt()
  }, [loading, error])

  // Show all known Mazandaran cities in the dropdown by default, plus any
  // extra ones that appear in the live data.
  const cities = useMemo(() => {
    const known = new Set(getKnownCities())
    outages.forEach((o) => o.city && known.add(o.city))
    return Array.from(known).sort((a, b) => a.localeCompare(b, 'fa'))
  }, [outages])

  const counts = useMemo(() => {
    let active = 0, upcoming = 0, past = 0
    outages.forEach((o) => {
      const s = outageStatus(o.start_at, o.end_at, now)
      if (s === 'active') active++
      else if (s === 'upcoming') upcoming++
      else past++
    })
    return { active, upcoming, past, total: outages.length }
  }, [outages, now])

  const filtered = useMemo(() => {
    const q = filters.q.trim().toLowerCase()
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const tomorrow = new Date(today)
    tomorrow.setDate(tomorrow.getDate() + 1)
    const dayAfter = new Date(tomorrow)
    dayAfter.setDate(dayAfter.getDate() + 1)
    const weekAhead = new Date(today)
    weekAhead.setDate(weekAhead.getDate() + 7)

    return outages.filter((o) => {
      if (o.city !== filters.city) return false
      const status = outageStatus(o.start_at, o.end_at, now)
      if (filters.status == "active-upcoming"){
        if (status != "upcoming" && status!="active"){
          return false
        }
      }else if (filters.status !== 'all' && (status !== filters.status)) return false
      if (q && !((o.address || '').toLowerCase().includes(q) || (o.city || '').toLowerCase().includes(q))) {
        return false
      }
      const start = new Date(o.start_at)
      if (filters.date === 'today' && !(start >= today && start < tomorrow)) return false
      if (filters.date === 'tomorrow' && !(start >= tomorrow && start < dayAfter)) return false
      if (filters.date === 'week' && !(start >= today && start < weekAhead)) return false
      return true
    })
  }, [outages, filters, now])

  const sorted = useMemo(() => {
    const arr = [...filtered]
    arr.sort((a, b) => {
      switch (sort) {
        case 'start_asc':
          return new Date(a.start_at) - new Date(b.start_at)
        case 'start_desc':
          return new Date(b.start_at) - new Date(a.start_at)
        case 'duration_desc':
          return durationMinutes(b.start_at, b.end_at) - durationMinutes(a.start_at, a.end_at)
        case 'city':
          return (a.city || '').localeCompare(b.city || '', 'fa')
        default:
          return 0
      }
    })
    return arr
  }, [filtered, sort])

  return (
    <div className="min-h-screen px-4 md:px-6 py-6 md:py-8 max-w-[1400px] mx-auto">
      <div className="space-y-6">
        <Header
          total={counts.total}
          active={counts.active}
          upcoming={counts.upcoming}
          past={counts.past}
          lastUpdated={lastUpdated}
        />

        <div className="grid grid-cols-1 lg:grid-cols-[1fr_300px] gap-6">
          <div className="space-y-4 min-w-0">
            {/* Pinned Favorites section — shown above filters */}
            <FavoriteOutages
              favoriteOutages={favoriteOutages}
              favoritesLoading={favoritesLoading}
              expandedId={expandedId}
              setExpandedId={setExpandedId}
              weatherProviderId={providerId}
              isFavorite={isFavorite}
              onRemoveFavorite={removeFavorite}
              onRefresh={refreshFavorites}
            />
            <FilterBar
              cities={cities}
              filters={filters}
              setFilters={setFilters}
              sort={sort}
              setSort={setSort}
              resultCount={sorted.length}
              totalCount={counts.total}
              onRefresh={refresh}
              loading={loading}
            />
            <OutageList
              loading={loading}
              error={error}
              outages={sorted}
              expandedId={expandedId}
              setExpandedId={setExpandedId}
              weatherProviderId={providerId}
              onRetry={refresh}
              isFavorite={isFavorite}
              onToggleFavorite={onToggleFavorite}
            />
          </div>

          <aside className="space-y-4">
            <InstallButton />
            <WeatherProviderSelector value={providerId} onChange={setProviderId} />
            <section className="card p-4 md:p-5">
              <h3 className="text-sm font-semibold text-slate-100 mb-2 flex items-center gap-2">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-cyan-400">
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="16" x2="12" y2="12" />
                  <line x1="12" y1="8" x2="12.01" y2="8" />
                </svg>
                راهنما
              </h3>
              <ul className="text-xs text-slate-400 space-y-1.5 leading-relaxed">
                <li>· روی «هواشناسی» هر رویداد بزنید تا دما، رطوبت و پوشش ابر نمایش داده شود.</li>
                <li>· ارائه‌دهنده آب و هوا از ستون سمت چپ (برای موبایل پایین صفحه) قابل تغییر است.</li>
                <li>· فیلترها در حافظه مرورگر ذخیره می‌شوند و در بازدید بعدی بازمی‌گردند.</li>
                <li>· لیست شهرها شامل مراکز توزیع برق مازندران است.</li>
              </ul>
            </section>
          </aside>
        </div>

        <footer className="text-center text-[11px] text-slate-500 pt-4 pb-2">
          CopyRight{" "}
          <a
            className="text-slate-400 hover:text-cyan-300 transition"
            href="https://t.me/fmotalleb"
            target="_blank"
            rel="noopener noreferrer"
          >
            @fmotalleb
          </a>
          <br />
          داده‌ها از سرویس{" "}
          <a
            className="text-slate-400 hover:text-cyan-300 transition"
            href="https://khamooshi.maztozi.ir/"
            target="_blank"
            rel="noopener noreferrer"
          >
            https://khamooshi.maztozi.ir
          </a>{" "}
          دریافت می‌شوند.
        </footer>
      </div>
    </div>
  )
}
