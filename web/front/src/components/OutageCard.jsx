import {
  outageStatus,
  formatDateTime,
  formatDuration,
  durationMinutes,
  relativeFromNow,
} from '../utils/dateUtils'
import WeatherModal from './WeatherModal'

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
    <>
      <article className="rounded-xl border border-slate-700/40 bg-slate-900/40 p-4">
        {/* HEADER ROW: CITY + STATUS */}
        <div className="mt-2 flex items-center justify-between gap-3">
          <div className="text-sm text-slate-300 truncate">
            {outage.city} · {formatDuration(dur)}
          </div>
          
          <div className={`chip border flex items-center gap-2 ${meta.cls}`}>
            <span className={`w-1.5 h-1.5 rounded-full ${meta.dot}`} />
            {meta.label}
          </div>
        </div>

        {/* FULL WIDTH ADDRESS */}
        <div className="mt-1 text-sm text-slate-400 w-full">
          {outage.address}
        </div>

        {/* 50/50 PANELS ALWAYS */}
        <div className="mt-4 grid grid-cols-2 gap-3">
          <div className="rounded-xl bg-emerald-500/[0.06] border border-emerald-400/15 px-3 py-2.5">
            <div className="text-[11px] text-emerald-300/80 mb-0.5">شروع قطعی</div>
            <div className="text-sm font-medium text-emerald-100">
              {formatDateTime(outage.start)}
            </div>
            <div className="text-[11px] text-slate-400 mt-0.5">{refStart}</div>
          </div>

          <div className="rounded-xl bg-rose-500/[0.06] border border-rose-400/15 px-3 py-2.5">
            <div className="text-[11px] text-rose-300/80 mb-0.5">پایان قطعی</div>
            <div className="text-sm font-medium text-rose-100">
              {formatDateTime(outage.end)}
            </div>
            <div className="text-[11px] text-slate-400 mt-0.5">{refEnd}</div>
          </div>
        </div>
      </article>

      <WeatherModal
        open={expanded}
        onClose={onToggle}
        outage={outage}
        providerId={weatherProviderId}
      />
    </>
  )
}