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
    cls: 'bg-rose-500/15 text-rose-300 border-rose-400/30',
    dot: 'bg-rose-400 shadow-[0_0_10px_rgba(244,63,94,0.8)] animate-pulse_soft',
    panel: 'bg-rose-500/[0.06] border-rose-400/15 text-rose-300/80',
    panelValue: 'text-rose-100',
  },
  upcoming: {
    label: 'پیش‌رو',
    cls: 'bg-orange-500/15 text-orange-300 border-orange-400/30',
    dot: 'bg-orange-400',
    panel: 'bg-orange-500/[0.06] border-orange-400/15 text-orange-300/80',
    panelValue: 'text-orange-100',
  },
  past: {
    label: 'پایان‌یافته',
    cls: 'bg-cyan-500/15 text-cyan-300 border-cyan-400/30',
    dot: 'bg-cyan-400',
    panel: 'bg-cyan-500/[0.06] border-cyan-400/15 text-cyan-300/80',
    panelValue: 'text-cyan-100',
  },
}

export default function OutageCard({ outage, weatherProviderId, expanded, onToggle, showFavorite, isFavorited, onToggleFavorite }) {
  const status = outageStatus(outage.start_at, outage.end_at)
  const meta = STATUS_META[status]

  const dur = durationMinutes(outage.start_at, outage.end_at)
  const refStart = relativeFromNow(outage.start_at)
  // const refEnd = relativeFromNow(outage.end)

  return (
    <>
      <article className={`rounded-xl border ${isFavorited ? 'border-amber-500/30 bg-gradient-to-br from-amber-500/[0.04] to-slate-900/40' : 'border-slate-700/40 bg-slate-900/40'} p-4 flex flex-col h-full`}>
        {/* HEADER ROW: CITY + STATUS + FAVORITE */}
        <div className="mt-2 flex flex-col gap-2">
          {/* LINE 1: CITY */}
          <div className="flex items-center gap-2 overflow-hidden">
            <div className="text-[15px] sm:text-sm text-slate-300 whitespace-nowrap shrink-0">
              {outage.city} · {refStart}
            </div>
          </div>
          
          {/* LINE 2: STATUS + WEATHER + FAVORITE */}
          <div className="flex items-center h-10 sm:h-8 overflow-hidden rounded-lg bg-slate-800/40 self-start max-w-full">
          
            {/* STATUS */}
            <div className={`flex items-center gap-2 px-3 h-full ${meta.cls}`}>
              <span className={`w-1.5 h-1.5 rounded-full ${meta.dot}`} />
              <span className="text-[15px] sm:text-xs leading-none whitespace-nowrap">
                {meta.label}
              </span>
            </div>
          
            {/* WEATHER */}
            <button
              onClick={onToggle}
              className="flex items-center gap-1 px-3 h-full text-[15px] sm:text-xs sm:text-sm text-slate-200 hover:text-slate-100 transition shrink-0"
              aria-expanded={expanded}
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <circle cx="12" cy="12" r="4" />
                <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
              </svg>
              هواشناسی
            </button>

            {/* FAVORITE */}
            {showFavorite && (
              <button
                onClick={(e) => { e.stopPropagation(); onToggleFavorite && onToggleFavorite() }}
                className={`flex items-center justify-center w-8 h-full text-xs transition shrink-0 ${
                  isFavorited
                    ? 'text-amber-400 hover:text-amber-300'
                    : 'text-slate-500 hover:text-amber-400'
                }`}
                title={isFavorited ? 'حذف از علاقه‌مندی‌ها' : 'افزودن به علاقه‌مندی‌ها'}
                aria-label={isFavorited ? 'حذف از علاقه‌مندی‌ها' : 'افزودن به علاقه‌مندی‌ها'}
              >
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill={isFavorited ? 'currentColor' : 'none'}
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  className={isFavorited ? 'drop-shadow-[0_0_6px_rgba(251,191,36,0.6)]' : ''}
                >
                  <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                </svg>
              </button>
            )}
          
          </div>
        </div>

        {/* FULL WIDTH ADDRESS */}
        <div className="mt-1.5 text-[15px] sm:text-sm text-slate-400 w-full leading-relaxed">
          {outage.address}
        </div>

        {/* 50/50 PANELS ALWAYS */}
        <div className="mt-auto pt-4 grid grid-cols-1 min-[420px]:grid-cols-2 gap-3">
          <div className={`rounded-xl border ${meta.panel} px-3 py-2.5`}>
            <div className="text-[13px] sm:text-xs md:text-[11px] mb-0.5">شروع قطعی</div>
            <div className={`text-base md:text-sm font-medium ${meta.panelValue}`}>
              {formatDateTime(outage.start_at)}
            </div>
          </div>

          <div className={`rounded-xl border ${meta.panel} px-3 py-2.5`}>
            <div className="text-[13px] sm:text-xs md:text-[11px] mb-0.5">پایان قطعی</div>
            <div className={`text-base md:text-sm font-medium ${meta.panelValue}`}>
              {formatDateTime(outage.end_at)}
            </div>
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