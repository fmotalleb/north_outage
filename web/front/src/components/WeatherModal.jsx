import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useWeather } from '../hooks/useWeather'
import { getCoords } from '../data/cityCoordinates'
import { formatTimeOnly } from '../utils/dateUtils'

// ── Sun + cloud icon ──────────────────────────────────────────────────────
// Single composite SVG. The sun stays put; the cloud slides + scales
// based on `cloudPct` (0–100) so it visually represents coverage.
//   0–10   → sun alone (no cloud)
//   10–40  → small cloud, sun mostly visible
//   40–70  → cloud covers half the sun
//   70–100 → cloud mostly covers the sun
function SunCloudIcon({ pct = 0, size = 96 }) {
  const showCloud = pct >= 10
  const cover = Math.min(1, Math.max(0, pct / 100)) // 0..1

  const cloudX = 18 + cover * 14     // 18 → 32
  const cloudY = 28 + cover * 8      // 28 → 36
  const cloudScale = 0.55 + cover * 0.45 // 0.55 → 1.0
  const sunOpacity = 1 - cover * 0.6

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 96 96"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      role="img"
      aria-label={`پوشش ابر ${Math.round(pct)} درصد`}
    >
      <defs>
        <radialGradient id="sunGrad" cx="0.5" cy="0.5" r="0.5">
          <stop offset="0" stopColor="#fde68a" />
          <stop offset="0.6" stopColor="#fbbf24" />
          <stop offset="1" stopColor="#f59e0b" />
        </radialGradient>
        <linearGradient id="cloudGrad" x1="0" x2="1" y1="0" y2="1">
          <stop offset="0" stopColor="#cbd5e1" />
          <stop offset="1" stopColor="#64748b" />
        </linearGradient>
      </defs>

      {/* Sun (behind cloud) */}
      <g style={{ opacity: sunOpacity }}>
        <circle cx="40" cy="42" r="20" fill="url(#sunGrad)" />
        {[0, 45, 90, 135, 180, 225, 270, 315].map((deg) => (
          <line
            key={deg}
            x1="40"
            y1="42"
            x2="40"
            y2="14"
            stroke="#fbbf24"
            strokeWidth="3"
            strokeLinecap="round"
            transform={`rotate(${deg} 40 42)`}
            opacity={0.7}
          />
        ))}
      </g>

      {/* Cloud (in front, scaled/positioned by cover) */}
      {showCloud && (
        <g
          transform={`translate(${cloudX - cloudX * cloudScale} ${cloudY - cloudY * cloudScale}) scale(${cloudScale})`}
        >
          <path
            d="M22 60 a16 16 0 0 1 8-31 18 18 0 0 1 34 4 14 14 0 0 1 6 27 H22 z"
            fill="url(#cloudGrad)"
            stroke="rgba(15,23,42,0.4)"
            strokeWidth="0.5"
          />
          <ellipse cx="44" cy="40" rx="14" ry="6" fill="rgba(255,255,255,0.25)" />
        </g>
      )}
    </svg>
  )
}

function cloudDescription(pct) {
  if (pct == null) return '—'
  if (pct < 10) return 'صاف'
  if (pct < 30) return 'کمی ابری'
  if (pct < 60) return 'نیمه ابری'
  if (pct < 85) return 'ابری'
  return 'تمام ابری'
}

// ── Modal ──────────────────────────────────────────────────────────────────
export default function WeatherModal({ open, onClose, outage, providerId }) {
  const fallbackCoords = outage ? getCoords(outage.city) : null

  // Fetch weather ourselves (the modal owns its own data lifecycle).
  const { loading, data, error } = useWeather(providerId, {
    city: outage?.city,
    lat: fallbackCoords?.latitude,
    lon: fallbackCoords?.longitude,
    startISO: outage?.start,
    endISO: outage?.end,
    // Only fetch while the modal is open so we don't burn API quota
    enabled: !!open && !!outage,
  })

  // Reset internal state when closed so reopening shows a fresh shimmer
  const [version, setVersion] = useState(0)
  useEffect(() => {
    if (open) setVersion((v) => v + 1)
  }, [open, outage?.unique_hash])

  // Close on Esc
  useEffect(() => {
    if (!open) return
    function onKey(e) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onClose])

  // Lock body scroll while open
  useEffect(() => {
    if (!open) return
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => { document.body.style.overflow = prev }
  }, [open])

  if (!open || !outage) return null

  return createPortal(
    <div
      className="fixed inset-0 z-[10000] flex items-center justify-center p-4 animate-slide_up"
      role="dialog"
      aria-modal="true"
      aria-labelledby="weather-modal-title"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/70 backdrop-blur-md"
        aria-hidden="true"
      />

      {/* Dialog */}
      <div
        className="relative w-full max-w-md rounded-2xl border border-white/10 shadow-2xl shadow-black/60 overflow-hidden"
        style={{
          background:
            'linear-gradient(180deg, rgba(27,36,56,0.97), rgba(12,18,32,0.97))',
        }}
      >
        {/* Header */}
        <div className="flex items-start justify-between p-5 border-b border-white/10">
          <div className="min-w-0">
            <h2 id="weather-modal-title" className="text-lg font-bold text-slate-100 flex items-center gap-2">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-cyan-400 shrink-0">
                <circle cx="12" cy="12" r="5" />
                <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" />
              </svg>
              هواشناسی بازه قطعی
            </h2>
            <div className="mt-1 text-xs text-slate-400 truncate">
              {outage.city} · {outage.address?.slice(0, 60)}{outage.address?.length > 60 ? '…' : ''}
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="btn-ghost !p-1.5 shrink-0"
            aria-label="بستن"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        {/* Body */}
        <div className="p-5" key={version}>
          {loading && (
            <div className="space-y-3">
              <div className="h-32 rounded-xl shimmer" />
              <div className="grid grid-cols-3 gap-3">
                <div className="h-20 rounded-lg shimmer" />
                <div className="h-20 rounded-lg shimmer" />
                <div className="h-20 rounded-lg shimmer" />
              </div>
            </div>
          )}

          {error && !loading && (
            <div className="text-sm text-rose-300/90 bg-rose-500/10 border border-rose-500/20 rounded-lg px-3 py-2">
              خطا در دریافت آب و هوا: {error}
            </div>
          )}

          {data && !data.available && !loading && (
            <div className="text-sm text-slate-300 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-center">
              {data.reason === 'no-coords'
                ? `مختصاتی برای «${outage.city}» یافت نشد`
                : 'داده‌ای برای این بازه زمانی موجود نیست'}
            </div>
          )}

          {data && data.available && !loading && (
            <>
              {/* Sun + cloud icon */}
              <div className="flex flex-col items-center justify-center py-4">
                <SunCloudIcon pct={data.avgCloud ?? 0} size={120} />
                <div className="mt-3 text-center">
                  <div className="text-2xl font-bold text-slate-100">
                    {cloudDescription(data.avgCloud)}
                  </div>
                  <div className="text-xs text-slate-400 mt-0.5">
                    پوشش ابر {Number.isFinite(data.avgCloud) ? Math.round(data.avgCloud) : '—'}٪
                  </div>
                </div>
              </div>

              {/* Stats grid */}
              <div className="grid grid-cols-3 gap-2">
                <div className="rounded-lg bg-gradient-to-l from-orange-500/15 to-rose-500/15 border border-orange-400/25 px-3 py-3 text-center">
                  <div className="text-[11px] text-orange-300 mb-1">دما</div>
                  <div className="text-xl font-bold text-orange-100">
                    {Number.isFinite(data.avgTemp) ? data.avgTemp.toFixed(1) + '°' : '—'}
                  </div>
                  {Number.isFinite(data.peakTemp) && (
                    <div className="text-[10px] text-slate-400 mt-1">اوج {data.peakTemp.toFixed(1)}°</div>
                  )}
                </div>
                <div className="rounded-lg bg-gradient-to-l from-cyan-500/15 to-blue-500/15 border border-cyan-400/25 px-3 py-3 text-center">
                  <div className="text-[11px] text-cyan-300 mb-1">رطوبت</div>
                  <div className="text-xl font-bold text-cyan-100">
                    {Number.isFinite(data.avgHumidity) ? Math.round(data.avgHumidity) + '٪' : '—'}
                  </div>
                  {Number.isFinite(data.peakHumidity) && (
                    <div className="text-[10px] text-slate-400 mt-1">اوج {Math.round(data.peakHumidity)}٪</div>
                  )}
                </div>
                <div className="rounded-lg bg-gradient-to-l from-slate-400/15 to-slate-600/15 border border-slate-400/25 px-3 py-3 text-center">
                  <div className="text-[11px] text-slate-300 mb-1">ابر</div>
                  <div className="text-xl font-bold text-slate-100">
                    {Number.isFinite(data.avgCloud) ? Math.round(data.avgCloud) + '٪' : '—'}
                  </div>
                  {Number.isFinite(data.peakCloud) && (
                    <div className="text-[10px] text-slate-400 mt-1">اوج {Math.round(data.peakCloud)}٪</div>
                  )}
                </div>
              </div>

              {/* Cloud cover bar */}
              {Number.isFinite(data.avgCloud) && (
                <div className="mt-3">
                  <div className="h-2 w-full rounded-full bg-white/5 overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-l from-cyan-400 via-slate-300 to-slate-500"
                      style={{ width: `${Math.max(2, Math.min(100, data.avgCloud))}%` }}
                    />
                  </div>
                </div>
              )}

              {/* Timeline */}
              {data.sampled.length > 1 && (
                <div className="mt-4">
                  <div className="text-[11px] text-slate-400 mb-1.5">نمونه‌های ساعتی</div>
                  <div className="flex items-stretch gap-1.5 overflow-x-auto pb-1">
                    {data.sampled.slice(0, 8).map((s, i) => (
                      <div
                        key={i}
                        className="shrink-0 text-[10px] text-slate-300 bg-white/5 border border-white/10 rounded-md px-2 py-1.5 min-w-[72px]"
                      >
                        <div className="font-medium text-slate-200">{formatTimeOnly(s.time)}</div>
                        <div className="text-orange-300 mt-0.5">
                          {Number.isFinite(s.temp) ? s.temp.toFixed(1) + '°' : '—'}
                        </div>
                        <div className="text-cyan-300">
                          {Number.isFinite(s.humidity) ? Math.round(s.humidity) + '٪' : '—'}
                        </div>
                        <div className="text-slate-400">
                          {Number.isFinite(s.cloud) ? Math.round(s.cloud) + '٪' : '—'}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Footer */}
              <div className="mt-4 flex items-center justify-between text-[10px] text-slate-500 pt-3 border-t border-white/10">
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
            </>
          )}
        </div>
      </div>
    </div>,
    document.body,
  )
}
