// Free & open weather providers (no API key required).
// Each provider exposes a `fetch({ city, lat, lon, startISO, endISO, signal })`
// method and normalizes the result into:
//   {
//     source, available, coords,
//     sampled:  [{ time, temp, humidity, cloud }],
//     avgTemp, avgHumidity, avgCloud,
//     peakTemp, peakHumidity, peakCloud,
//   }

const API = 'https://geocoding-api.open-meteo.com/v1/search'
const FORECAST_API = 'https://api.open-meteo.com/v1/forecast'
const ARCHIVE_API = 'https://archive-api.open-meteo.com/v1/archive'
const MET_API = 'https://api.met.no/weatherapi/locationforecast/2.0/compact'

// ---------- Helpers ----------

function toDateStr(iso) {
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

function avgOf(arr) {
  const v = arr.filter((x) => Number.isFinite(x))
  return v.length ? v.reduce((a, b) => a + b, 0) / v.length : null
}

// ---------- Open-Meteo ----------

async function geocodeOpenMeteo(cityName, signal) {
  try {
    const url = `${API}?name=${encodeURIComponent(cityName)}&count=1&language=fa&format=json`
    const r = await fetch(url, { signal })
    if (r.ok) {
      const j = await r.json()
      if (j.results?.length) return { latitude: j.results[0].latitude, longitude: j.results[0].longitude }
    }
  } catch (_) { /* fall through */ }
  try {
    const url = `${API}?name=${encodeURIComponent(cityName)}&count=1&language=en&format=json`
    const r = await fetch(url, { signal })
    if (r.ok) {
      const j = await r.json()
      if (j.results?.length) return { latitude: j.results[0].latitude, longitude: j.results[0].longitude }
    }
  } catch (_) { /* fall through */ }
  return null
}

async function fetchOpenMeteo({ city, lat, lon, startISO, endISO, signal }) {
  const start = toDateStr(startISO)
  const end = toDateStr(endISO)

  let coords = lat && lon ? { latitude: lat, longitude: lon } : null
  if (!coords) coords = await geocodeOpenMeteo(city, signal)
  if (!coords) return { source: 'Open-Meteo', available: false, reason: 'no-coords', sampled: [] }

  // Decide forecast vs archive based on whether end is in the past.
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
    hourly: 'temperature_2m,relative_humidity_2m,cloud_cover',
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
  const clouds = j.hourly?.cloud_cover || []

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
        cloud: clouds[i],
      })
    }
  }
  // Ensure the start sample is present
  const startIdx = pickHourlyIndex(times, startISO)
  if (startIdx >= 0 && !sampled.some((s) => s.time === times[startIdx])) {
    sampled.unshift({
      time: times[startIdx],
      temp: temps[startIdx],
      humidity: hums[startIdx],
      cloud: clouds[startIdx],
    })
  }

  const validTemps = sampled.map((s) => s.temp).filter((v) => Number.isFinite(v))
  const validHums = sampled.map((s) => s.humidity).filter((v) => Number.isFinite(v))
  const validClouds = sampled.map((s) => s.cloud).filter((v) => Number.isFinite(v))

  return {
    source: 'Open-Meteo',
    provider: 'open-meteo',
    available: true,
    coords,
    sampled,
    avgTemp: avgOf(validTemps),
    avgHumidity: avgOf(validHums),
    avgCloud: avgOf(validClouds),
    peakTemp: validTemps.length ? Math.max(...validTemps) : null,
    peakHumidity: validHums.length ? Math.max(...validHums) : null,
    peakCloud: validClouds.length ? Math.max(...validClouds) : null,
    minCloud: validClouds.length ? Math.min(...validClouds) : null,
  }
}

// ---------- Met.no (Norwegian Meteorological Institute) ----------

async function fetchMetNo({ lat, lon, startISO, endISO, signal }) {
  if (!lat || !lon) {
    return { source: 'Met.no', available: false, reason: 'no-coords', sampled: [] }
  }
  const url = `${MET_API}?lat=${lat}&lon=${lon}`
  const r = await fetch(url, {
    signal,
    headers: { 'User-Agent': 'outage-tracker-demo/1.0 (educational)' },
  })
  if (!r.ok) throw new Error(`Met.no HTTP ${r.status}`)
  const j = await r.json()
  const series = j.properties?.timeseries || []
  if (!series.length) return { source: 'Met.no', available: false, reason: 'empty', sampled: [] }

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
        // Met.no's cloud_area_fraction is already in percent (0–100).
        cloud: Number.isFinite(inst.cloud_area_fraction)
          ? Math.round(inst.cloud_area_fraction)
          : null,
      })
    }
  }
  const validTemps = sampled.map((s) => s.temp).filter((v) => Number.isFinite(v))
  const validHums = sampled.map((s) => s.humidity).filter((v) => Number.isFinite(v))
  const validClouds = sampled.map((s) => s.cloud).filter((v) => Number.isFinite(v))

  return {
    source: 'Met.no',
    provider: 'met-no',
    available: true,
    coords: { latitude: lat, longitude: lon },
    sampled,
    avgTemp: avgOf(validTemps),
    avgHumidity: avgOf(validHums),
    avgCloud: avgOf(validClouds),
    peakTemp: validTemps.length ? Math.max(...validTemps) : null,
    peakHumidity: validHums.length ? Math.max(...validHums) : null,
    peakCloud: validClouds.length ? Math.max(...validClouds) : null,
    minCloud: validClouds.length ? Math.min(...validClouds) : null,
    note: 'Met.no فقط حدود ۳ روز آینده را پیش‌بینی می‌کند. برای قطعی‌های گذشته ممکن است داده‌ای موجود نباشد.',
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
    description: 'Free, no key. Forecast window ~3 days.',
    homepage: 'https://api.met.no',
    fetch: fetchMetNo,
  },
]

export function getProvider(id) {
  return WEATHER_PROVIDERS.find((p) => p.id === id) || WEATHER_PROVIDERS[0]
}
