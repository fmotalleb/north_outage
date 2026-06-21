import { outageStatus, formatDateTime, formatDuration, durationMinutes, relativeFromNow } from '../utils/dateUtils'
import WeatherPanel from './WeatherPanel'

const STATUS_META = {
  active: {
    label: 'در جریان',
    cls: 'bg-emerald-500/15 text-emerald-300 border-emerald-400/30',
    dot: 'bg-emerald-400 shadow-[0_0_10px_rgba(16,185,129,0.8)] animate-pulse_soft',
  },
  upcoming: {
    label: 'پیش‌رو',
    cls: 'bg-amber-500/15 text-amber-300 border-amber-400/30',
    dot: 'bg-amber-400',
  },
  past: {
    label: 'پایان‌یافته',
    cls: 'bg-slate-500/15 text-slate-400 border-slate-400/20',
    dot: 'bg-slate-500',
  },
}

export default function OutageCard({ outage, weatherProviderId, expanded, onToggle }) {
  const status = outageStatus(outage.start, outage.end)
  const meta = STATUS_META[status]
  const dur = durationMinutes(outage.start, outage.end)
  const refStart = relativeFromNow(outage.start)
  const refEnd = relativeFromNow(outage.end)

  return (
    <article className="card overflow-hidden animate-slide_up">
      <div className="p-4 md:p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 flex-wrap mb-1.5">
              <span className={`chip border ${meta.cls}`}>
                <span className={`w-1.5 h-1.5 rounded-full ${meta.dot}`} />
                {meta.label}
              </span>
              <span className="chip bg-white/5 border border-white/10 text-slate-300">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                  <circle cx="12" cy="10" r="3" />
                </svg>
                {outage.city}
              </span>
              <span className="chip bg-white/5 border border-white/10 text-slate-300">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <circle cx="12" cy="12" r="10" />
                  <polyline points="12 6 12 12 16 14" />
                </svg>
                {formatDuration(dur)}
              </span>
            </div>
            <h3 className="text-base md:text-lg font-semibold text-slate-100 leading-snug">
              {outage.address}
            </h3>
          </div>
          <button
            onClick={onToggle}
            className="btn-ghost !py-1.5 !px-2.5 text-xs shrink-0"
            aria-expanded={expanded}
          >
            <svg
              width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor"
              strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
              className={`transition-transform ${expanded ? 'rotate-180' : ''}`}
            >
              <polyline points="6 9 12 15 18 9" />
            </svg>
            {expanded ? 'بستن' : 'جزئیات'}
          </button>
        </div>

        <div className="mt-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div className="rounded-xl bg-emerald-500/[0.06] border border-emerald-400/15 px-3 py-2.5">
            <div className="text-[11px] text-emerald-300/80 mb-0.5">شروع قطعی</div>
            <div className="text-sm font-medium text-emerald-100">{formatDateTime(outage.start)}</div>
            <div className="text-[11px] text-slate-400 mt-0.5">{refStart}</div>
          </div>
          <div className="rounded-xl bg-rose-500/[0.06] border border-rose-400/15 px-3 py-2.5">
            <div className="text-[11px] text-rose-300/80 mb-0.5">پایان قطعی</div>
            <div className="text-sm font-medium text-rose-100">{formatDateTime(outage.end)}</div>
            <div className="text-[11px] text-slate-400 mt-0.5">{refEnd}</div>
          </div>
        </div>

        {expanded && (
          <div className="mt-4 pt-4 border-t border-white/10 animate-slide_up">
            <div className="flex items-center justify-between mb-2">
              <h4 className="text-sm font-semibold text-slate-200 flex items-center gap-2">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M3 15a4 4 0 0 0 4 4h10a4 4 0 0 0 1-7.87" />
                  <path d="M19.07 7.87A4 4 0 0 0 17 4H7a4 4 0 0 0-3.9 3" />
                </svg>
                شرایط آب و هوا در بازه قطعی
              </h4>
            </div>
            <WeatherPanel providerId={weatherProviderId} outage={outage} />
          </div>
        )}
      </div>
    </article>
  )
}
