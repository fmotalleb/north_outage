// Free & open weather providers (no API key required for any of these).
// Each provider exposes a `fetchWeather({ lat, lon, startISO, endISO, signal })` method
// and normalizes the result into { source, hourly: [{ time, temp, humidity }] }.

const API = 'https://geocoding-api.open-meteo.com/v1/search'
const FORECAST_API = 'https://api.open-meteo.com/v1/forecast'
const ARCHIVE_API = 'https://archive-api.open-meteo.com/v1/archive'
const MET_API = 'https://api.met.no/weatherapi/locationforecast/2.0/compact'
const SEVEN_TIMER_API = 'https://www.7timer.info/bin/api.pl'

// ---------- Helpers ----------

function toDateStr(iso) {
  // YYYY-MM-DD in the ISO's local timezone
  const d = new Date(iso)
  const y = d.getUTCFullYear()
  const m = String(d.getUTCMonth() + 1).padStart(2, '0')
  const day = String(d.getUTCDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function pickHourlyIndex(times, targetISO) {
  if (!times || !times.length) return -1
  const t = new Date(targetISO).getTime()
  let best = 0
  let bestDiff = Infinity
  for (let i = 0; i < times.length; i++) {
    const diff = Math.abs(new Date(times[i]).getTime() - t)
    if (diff < bestDiff) {
      bestDiff = diff
      best = i
    }
  }
  return best
}

// ---------- Open-Meteo ----------

async function geocodeOpenMeteo(cityName, signal) {
  // Try Persian first, then English fallback, then hard-coded coordinates.
  try {
    const url = `${API}?name=${encodeURIComponent(cityName)}&count=1&language=fa&format=json`
    const r = await fetch(url, { signal })
    if (r.ok) {
      const j = await r.json()
      if (j.results && j.results.length) {
        return { latitude: j.results[0].latitude, longitude: j.results[0].longitude }
      }
    }
  } catch (_) { /* fall through */ }

  try {
    const url = `${API}?name=${encodeURIComponent(cityName)}&count=1&language=en&format=json`
    const r = await fetch(url, { signal })
    if (r.ok) {
      const j = await r.json()
      if (j.results && j.results.length) {
        return { latitude: j.results[0].latitude, longitude: j.results[0].longitude }
      }
    }
  } catch (_) { /* fall through */ }

  return null
}

async function fetchOpenMeteo({ city, lat, lon, startISO, endISO, signal }) {
  const start = toDateStr(startISO)
  const end = toDateStr(endISO)

  let coords = lat && lon ? { latitude: lat, longitude: lon } : null
  if (!coords) coords = await geocodeOpenMeteo(city, signal)

  if (!coords) {
    return {
      source: 'Open-Meteo',
      available: false,
      reason: 'no-coords',
      hourly: [],
    }
  }

  // Decide forecast vs archive based on whether end is in the past.
  // Archive is available up to a few days behind today.
  const endDate = new Date(endISO)
  const now = new Date()
  const hoursAhead = (endDate.getTime() - now.getTime()) / 36e5
  const isFuture = hoursAhead > 24
  const hoursBehind = (now.getTime() - endDate.getTime()) / 36e5
  const isPastTooFar = hoursBehind > 6 * 24

  const base = isFuture ? FORECAST_API : isPastTooFar ? ARCHIVE_API : FORECAST_API
  const params = new URLSearchParams({
    latitude: String(coords.latitude),
    longitude: String(coords.longitude),
    hourly: 'temperature_2m,relative_humidity_2m',
    start_date: start,
    end_date: end,
    timezone: 'auto',
  })
  const url = `${base}?${params.toString()}`
  const r = await fetch(url, { signal })
  if (!r.ok) throw new Error(`Open-Meteo HTTP ${r.status}`)
  const j = await r.json()

  const times = j.hourly?.time || []
  const temps = j.hourly?.temperature_2m || []
  const hums = j.hourly?.relative_humidity_2m || []

  // Sample a few representative hours across the outage window.
  const sampled = []
  const startMs = new Date(startISO).getTime()
  const endMs = new Date(endISO).getTime()
  const step = Math.max(1, Math.floor(times.length / 6))
  for (let i = 0; i < times.length; i += step) {
    const t = new Date(times[i]).getTime()
    if (t >= startMs - 6e5 && t <= endMs + 6e5) {
      sampled.push({
        time: times[i],
        temp: temps[i],
        humidity: hums[i],
      })
    }
  }
  // Ensure at least the start sample is present
  const startIdx = pickHourlyIndex(times, startISO)
  if (startIdx >= 0 && !sampled.some((s) => s.time === times[startIdx])) {
    sampled.unshift({
      time: times[startIdx],
      temp: temps[startIdx],
      humidity: hums[startIdx],
    })
  }

  // Averages
  const validTemps = sampled.filter((s) => Number.isFinite(s.temp)).map((s) => s.temp)
  const validHums = sampled.filter((s) => Number.isFinite(s.humidity)).map((s) => s.humidity)
  const avg = (arr) => (arr.length ? arr.reduce((a, b) => a + b, 0) / arr.length : null)

  return {
    source: 'Open-Meteo',
    provider: 'open-meteo',
    available: true,
    coords,
    sampled,
    avgTemp: avg(validTemps),
    avgHumidity: avg(validHums),
    peakTemp: validTemps.length ? Math.max(...validTemps) : null,
    peakHumidity: validHums.length ? Math.max(...validHums) : null,
  }
}

// ---------- Met.no (Norwegian Meteorological Institute) ----------

async function fetchMetNo({ lat, lon, startISO, endISO, signal }) {
  if (!lat || !lon) {
    return { source: 'Met.no', available: false, reason: 'no-coords', hourly: [] }
  }
  const url = `${MET_API}?lat=${lat}&lon=${lon}`
  const r = await fetch(url, {
    signal,
    headers: { 'User-Agent': 'outage-tracker-demo/1.0 (educational)' },
  })
  if (!r.ok) throw new Error(`Met.no HTTP ${r.status}`)
  const j = await r.json()
  const series = j.properties?.timeseries || []
  if (!series.length) {
    return { source: 'Met.no', available: false, reason: 'empty', hourly: [] }
  }

  const startMs = new Date(startISO).getTime()
  const endMs = new Date(endISO).getTime()
  const sampled = []
  for (const entry of series) {
    const t = new Date(entry.time).getTime()
    if (t >= startMs - 6e5 && t <= endMs + 6e5) {
      const inst = entry.data?.instant?.details || {}
      sampled.push({
        time: entry.time,
        temp: inst.air_temperature ?? null,
        humidity: inst.relative_humidity ?? null,
      })
    }
  }
  const validTemps = sampled.map((s) => s.temp).filter((v) => Number.isFinite(v))
  const validHums = sampled.map((s) => s.humidity).filter((v) => Number.isFinite(v))
  const avg = (arr) => (arr.length ? arr.reduce((a, b) => a + b, 0) / arr.length : null)

  return {
    source: 'Met.no',
    provider: 'met-no',
    available: true,
    coords: { latitude: lat, longitude: lon },
    sampled,
    avgTemp: avg(validTemps),
    avgHumidity: avg(validHums),
    peakTemp: validTemps.length ? Math.max(...validTemps) : null,
    peakHumidity: validHums.length ? Math.max(...validHums) : null,
    note: 'Met.no returns a ~3 day forecast window. Past outages may not have data.',
  }
}

// ---------- 7Timer! ----------

async function fetchSevenTimer({ city, lat, lon, signal }) {
  // 7Timer! requires lat/lon. If not provided, skip.
  if (!lat || !lon) {
    return { source: '7Timer!', available: false, reason: 'no-coords', hourly: [] }
  }
  // The api.pl endpoint redirects to meteo.php for the "meteo" product which
  // has hourly samples (3-hour step) and includes humidity (rh2m) + temperature (temp2m).
  const url = `${SEVEN_TIMER_API}?lon=${lon}&lat=${lat}&product=meteo&output=json`
  const r = await fetch(url, { signal, redirect: 'follow' })
  if (!r.ok) throw new Error(`7Timer! HTTP ${r.status}`)
  const j = await r.json()
  const series = j.dataseries || []
  if (!series.length) {
    return { source: '7Timer!', available: false, reason: 'empty', hourly: [] }
  }
  // Each entry has `timepoint` (3 = +3h, 6 = +6h, …) from the init run time.
  const init = String(j.init || '')
  const iy = init.slice(0, 4)
  const imo = init.slice(4, 6)
  const id = init.slice(6, 8)
  const ihh = init.slice(8, 10) || '00'
  const initISO = `${iy}-${imo}-${id}T${ihh}:00:00Z`
  const sampled = series.map((entry) => {
    const tp = Number(entry.timepoint) || 0
    const t = new Date(initISO)
    t.setUTCHours(t.getUTCHours() + tp)
    return {
      time: t.toISOString(),
      temp: Number.isFinite(entry.temp2m) ? entry.temp2m : null,
      humidity: Number.isFinite(entry.rh2m) ? entry.rh2m : null,
    }
  })
  const validTemps = sampled.map((s) => s.temp).filter((v) => Number.isFinite(v))
  const validHums = sampled.map((s) => s.humidity).filter((v) => Number.isFinite(v))
  const avg = (arr) => (arr.length ? arr.reduce((a, b) => a + b, 0) / arr.length : null)
  return {
    source: '7Timer!',
    provider: '7timer',
    available: true,
    coords: { latitude: lat, longitude: lon },
    sampled,
    avgTemp: avg(validTemps),
    avgHumidity: avg(validHums),
    note: '7Timer! provides 3-hour step forecasts (~7 days). The rh2m field is a humidity index, not standard % RH.',
  }
}

// ---------- Public registry ----------

export const WEATHER_PROVIDERS = [
  {
    id: 'open-meteo',
    name: 'Open-Meteo',
    description: 'Free, no API key. Best historical + forecast coverage worldwide.',
    homepage: 'https://open-meteo.com',
    fetch: fetchOpenMeteo,
  },
  {
    id: 'met-no',
    name: 'Met.no',
    description: 'Norwegian Meteorological Institute. Free, no key. Forecast window ~3 days.',
    homepage: 'https://api.met.no',
    fetch: fetchMetNo,
  },
  {
    id: '7timer',
    name: '7Timer!',
    description: 'Free, no key. Daily aggregates only.',
    homepage: 'https://www.7timer.info',
    fetch: fetchSevenTimer,
  },
]

export function getProvider(id) {
  return WEATHER_PROVIDERS.find((p) => p.id === id) || WEATHER_PROVIDERS[0]
}
