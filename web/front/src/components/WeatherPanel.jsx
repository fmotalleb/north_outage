import { useWeather } from '../hooks/useWeather'
import { getCoords } from '../data/cityCoordinates'
import { formatTimeOnly, formatDateOnly } from '../utils/dateUtils'

function WeatherIcon({ code }) {
  // code: 'temp' | 'humidity'
  if (code === 'temp') {
    return (
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M14 14.76V3.5a2.5 2.5 0 0 0-5 0v11.26a4.5 4.5 0 1 0 5 0z" />
      </svg>
    )
  }
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z" />
    </svg>
  )
}

export default function WeatherPanel({ providerId, outage }) {
  const fallbackCoords = getCoords(outage.city)
  const { loading, data, error } = useWeather(providerId, {
    city: outage.city,
    lat: fallbackCoords?.latitude,
    lon: fallbackCoords?.longitude,
    startISO: outage.start,
    endISO: outage.end,
    enabled: true,
  })

  if (loading) {
    return (
      <div className="grid grid-cols-2 gap-2 mt-3">
        <div className="h-12 rounded-lg shimmer" />
        <div className="h-12 rounded-lg shimmer" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="mt-3 text-xs text-rose-300/90 bg-rose-500/10 border border-rose-500/20 rounded-lg px-3 py-2">
        خطا در دریافت آب و هوا: {error}
      </div>
    )
  }

  if (!data || !data.available) {
    return (
      <div className="mt-3 text-xs text-slate-400 bg-white/5 border border-white/10 rounded-lg px-3 py-2">
        {data?.reason === 'no-coords'
          ? `مختصاتی برای «${outage.city}» یافت نشد`
          : 'داده‌ای برای این بازه زمانی موجود نیست'}
      </div>
    )
  }

  const fmt = (v, digits = 1) => (Number.isFinite(v) ? v.toFixed(digits) : '—')
  const temps = data.sampled.map((s) => s.temp).filter((v) => Number.isFinite(v))
  const hums = data.sampled.map((s) => s.humidity).filter((v) => Number.isFinite(v))
  const peakTemp = temps.length ? Math.max(...temps) : null
  const peakHum = hums.length ? Math.max(...hums) : null
  const minTemp = temps.length ? Math.min(...temps) : null

  return (
    <div className="mt-3">
      <div className="grid grid-cols-2 gap-2">
        <div className="rounded-lg bg-gradient-to-l from-orange-500/10 to-rose-500/10 border border-orange-400/20 px-3 py-2">
          <div className="flex items-center gap-1.5 text-orange-300">
            <WeatherIcon code="temp" />
            <span className="text-[11px]">دما (میانگین)</span>
          </div>
          <div className="mt-0.5 text-lg font-bold text-orange-100">
            {fmt(data.avgTemp)}°C
          </div>
          {Number.isFinite(minTemp) && Number.isFinite(peakTemp) && (
            <div className="text-[10px] text-slate-400 mt-0.5">
              بازه {fmt(minTemp)}° – {fmt(peakTemp)}°
            </div>
          )}
        </div>
        <div className="rounded-lg bg-gradient-to-l from-cyan-500/10 to-blue-500/10 border border-cyan-400/20 px-3 py-2">
          <div className="flex items-center gap-1.5 text-cyan-300">
            <WeatherIcon code="humidity" />
            <span className="text-[11px]">رطوبت (میانگین)</span>
          </div>
          <div className="mt-0.5 text-lg font-bold text-cyan-100">
            {fmt(data.avgHumidity, 0)}%
          </div>
          {Number.isFinite(peakHum) && (
            <div className="text-[10px] text-slate-400 mt-0.5">
              اوج {fmt(peakHum, 0)}%
            </div>
          )}
        </div>
      </div>

      {data.sampled.length > 1 && (
        <div className="mt-2 flex items-center gap-1 overflow-x-auto pb-1">
          {data.sampled.slice(0, 8).map((s, i) => (
            <div
              key={i}
              className="shrink-0 text-[10px] text-slate-300 bg-white/5 border border-white/10 rounded-md px-2 py-1"
              title={s.time}
            >
              <div className="font-medium">{formatTimeOnly(s.time)}</div>
              <div className="text-orange-300">
                {Number.isFinite(s.temp) ? `${s.temp.toFixed(1)}°` : '—'}
              </div>
              <div className="text-cyan-300">
                {Number.isFinite(s.humidity) ? `${s.humidity.toFixed(0)}%` : '—'}
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="mt-2 flex items-center justify-between text-[10px] text-slate-500">
        <span>منبع: {data.source}</span>
        {data.coords && (
          <span>
            {data.coords.latitude.toFixed(3)}, {data.coords.longitude.toFixed(3)}
          </span>
        )}
      </div>
      {data.note && (
        <div className="mt-1 text-[10px] text-amber-300/80">{data.note}</div>
      )}
    </div>
  )
}
