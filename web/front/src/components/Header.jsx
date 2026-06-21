import { formatDateTime } from '../utils/dateUtils'

export default function Header({ total, active, upcoming, past, lastUpdated }) {
  const stats = [
    { label: 'کل رویدادها', value: total, gradient: 'from-slate-500 to-slate-700' },
    { label: 'در جریان', value: active, gradient: 'from-emerald-500 to-cyan-500' },
    { label: 'پیش‌رو', value: upcoming, gradient: 'from-amber-500 to-orange-500' },
    { label: 'پایان‌یافته', value: past, gradient: 'from-slate-600 to-slate-800' },
  ]
  return (
    <header className="relative">
      <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="relative">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-400 to-violet-500 flex items-center justify-center shadow-lg shadow-cyan-500/30">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" className="text-white">
                  <path d="M13 2 L4 14 H11 L9 22 L20 10 H13 Z" />
                </svg>
              </div>
              <div className="absolute -inset-1 rounded-2xl bg-gradient-to-br from-cyan-400/40 to-violet-500/40 blur-md -z-10" />
            </div>
            <h1 className="text-2xl md:text-3xl font-extrabold tracking-tight bg-gradient-to-l from-white via-cyan-100 to-violet-200 bg-clip-text text-transparent">
              قطعی‌های برنامه‌ریزی‌شده برق
            </h1>
          </div>
          <p className="text-slate-400 text-sm md:text-base">
            پایش زنده رویدادهای قطعی برق استان مازندران
            {lastUpdated && (
              <span className="mr-2 text-slate-500">
                · آخرین بروزرسانی: {formatDateTime(lastUpdated.toISOString())}
              </span>
            )}
          </p>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-2 md:gap-3">
          {stats.map((s) => (
            <div key={s.label} className="card px-3.5 py-2.5 min-w-[110px]">
              <div className="text-[11px] uppercase tracking-wider text-slate-400">{s.label}</div>
              <div className={`mt-0.5 text-2xl font-bold bg-gradient-to-l ${s.gradient} bg-clip-text text-transparent`}>
                {s.value}
              </div>
            </div>
          ))}
        </div>
      </div>
    </header>
  )
}
