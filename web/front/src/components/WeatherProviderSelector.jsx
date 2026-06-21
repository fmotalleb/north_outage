import { WEATHER_PROVIDERS } from '../utils/weatherProviders'

export default function WeatherProviderSelector({ value, onChange }) {
  return (
    <section className="card p-4 md:p-5">
      <div className="flex items-start justify-between gap-3 mb-3">
        <div>
          <h2 className="text-sm font-semibold text-slate-100 flex items-center gap-2">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-cyan-400">
              <path d="M3 15a4 4 0 0 0 4 4h10a4 4 0 0 0 1-7.87" />
              <path d="M19.07 7.87A4 4 0 0 0 17 4H7a4 4 0 0 0-3.9 3" />
            </svg>
            ارائه‌دهنده آب و هوا
          </h2>
          <p className="text-xs text-slate-400 mt-1">
            یکی از سرویس‌های رایگان و بدون کلید API را انتخاب کنید.
          </p>
        </div>
      </div>
      <div className="space-y-2">
        {WEATHER_PROVIDERS.map((p) => {
          const active = value === p.id
          return (
            <button
              key={p.id}
              onClick={() => onChange(p.id)}
              className={`w-full text-right rounded-xl px-3 py-2.5 transition border ${
                active
                  ? 'bg-gradient-to-l from-cyan-500/15 to-violet-500/15 border-cyan-400/40 shadow-[0_0_0_1px_rgba(34,211,238,0.25)]'
                  : 'bg-white/5 border-white/10 hover:bg-white/10'
              }`}
              type="button"
            >
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <div className={`text-sm font-semibold ${active ? 'text-cyan-100' : 'text-slate-100'}`}>
                    {p.name}
                  </div>
                  <div className="text-[11px] text-slate-400 mt-0.5 line-clamp-2">{p.description}</div>
                </div>
                <div
                  className={`shrink-0 w-4 h-4 rounded-full border-2 flex items-center justify-center ${
                    active ? 'border-cyan-400' : 'border-slate-500'
                  }`}
                >
                  {active && <div className="w-2 h-2 rounded-full bg-cyan-400" />}
                </div>
              </div>
            </button>
          )
        })}
      </div>
      <p className="mt-3 text-[10px] text-slate-500 leading-relaxed">
        داده‌ها مستقیماً از API ارائه‌دهنده دریافت می‌شوند. هیچ کلیدی در فرانت‌اند ذخیره نمی‌شود.
      </p>
    </section>
  )
}
