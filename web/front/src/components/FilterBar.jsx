export default function FilterBar({
  cities,
  filters,
  setFilters,
  sort,
  setSort,
  resultCount,
  onRefresh,
  loading,
}) {
  const update = (k) => (e) => setFilters((f) => ({ ...f, [k]: e.target.value }))

  const reset = () =>
    setFilters({ city: 'all', status: 'all', q: '', date: 'all' })

  return (
    <section className="card p-4 md:p-5">
      <div className="grid grid-cols-1 md:grid-cols-12 gap-3 md:gap-4">
        <div className="md:col-span-4">
          <label className="label">جستجو در آدرس</label>
          <div className="relative">
            <input
              type="text"
              className="input pr-9"
              placeholder="مثلاً: شهرک صنعتی، روستا، خیابان…"
              value={filters.q}
              onChange={update('q')}
            />
            <svg className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          </div>
        </div>

        <div className="md:col-span-3">
          <label className="label">شهر</label>
          <select className="select" value={filters.city} onChange={update('city')}>
            <option value="all">همه شهرها</option>
            {cities.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
        </div>

        <div className="md:col-span-3">
          <label className="label">وضعیت</label>
          <select className="select" value={filters.status} onChange={update('status')}>
            <option value="all">همه وضعیت‌ها</option>
            <option value="active">در جریان</option>
            <option value="upcoming">پیش‌رو</option>
            <option value="past">پایان‌یافته</option>
          </select>
        </div>

        <div className="md:col-span-2">
          <label className="label">بازه تاریخ</label>
          <select className="select" value={filters.date} onChange={update('date')}>
            <option value="all">همه</option>
            <option value="today">امروز</option>
            <option value="tomorrow">فردا</option>
            <option value="week">هفت روز آینده</option>
          </select>
        </div>
      </div>

      <div className="mt-4 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-sm text-slate-400">
          <span className="chip bg-white/5 border border-white/10 text-slate-300">
            {resultCount.toLocaleString('fa-IR')} نتیجه
          </span>
          <select
            className="select !py-1.5 !text-xs w-auto"
            value={sort}
            onChange={(e) => setSort(e.target.value)}
          >
            <option value="start_asc">شروع: نزدیک‌ترین</option>
            <option value="start_desc">شروع: دورترین</option>
            <option value="duration_desc">مدت: طولانی‌ترین</option>
            <option value="city">شهر (الفبایی)</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <button className="btn-ghost" onClick={reset} type="button">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="1 4 1 10 7 10" />
              <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10" />
            </svg>
            پاک کردن فیلترها
          </button>
          <button className="btn-primary" onClick={onRefresh} disabled={loading} type="button">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={loading ? 'animate-spin' : ''}>
              <polyline points="23 4 23 10 17 10" />
              <polyline points="1 20 1 14 7 14" />
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10" />
              <path d="M20.49 15a9 9 0 0 1-14.85 3.36L1 14" />
            </svg>
            بروزرسانی
          </button>
        </div>
      </div>
    </section>
  )
}
