import Dropdown from './Dropdown'

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
  const update = (k) => (v) => setFilters((f) => ({ ...f, [k]: v }))

  const cityOptions = [
    { value: 'all', label: 'همه شهرها' },
    ...cities.map((c) => ({ value: c, label: c })),
  ]
  const statusOptions = [
    { value: 'all', label: 'همه وضعیت‌ها' },
    { value: 'active', label: 'در جریان' },
    { value: 'upcoming', label: 'پیش‌رو' },
    { value: 'past', label: 'پایان‌یافته' },
  ]
  const dateOptions = [
    { value: 'all', label: 'همه تاریخ‌ها' },
    { value: 'today', label: 'امروز' },
    { value: 'tomorrow', label: 'فردا' },
    { value: 'week', label: 'هفت روز آینده' },
  ]
  const sortOptions = [
    { value: 'start_asc', label: 'شروع: نزدیک‌ترین' },
    { value: 'start_desc', label: 'شروع: دورترین' },
    { value: 'duration_desc', label: 'مدت: طولانی‌ترین' },
    { value: 'city', label: 'شهر (الفبایی)' },
  ]

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
              onChange={(e) => update('q')(e.target.value)}
            />
            <svg className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          </div>
        </div>

        <div className="md:col-span-3">
          <Dropdown
            label="شهر"
            value={filters.city}
            onChange={update('city')}
            options={cityOptions}
          />
        </div>

        <div className="md:col-span-3">
          <Dropdown
            label="وضعیت"
            value={filters.status}
            onChange={update('status')}
            options={statusOptions}
          />
        </div>

        <div className="md:col-span-2">
          <Dropdown
            label="بازه تاریخ"
            value={filters.date}
            onChange={update('date')}
            options={dateOptions}
          />
        </div>
      </div>

      <div className="mt-4 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-sm text-slate-400">
          <Dropdown
            value={sort}
            onChange={setSort}
            options={sortOptions}
            className="w-auto min-w-[180px]"
          />
        </div>

        <div className="flex items-center gap-2">
          <span className="chip bg-white/5 border border-white/10 text-slate-300">
            {resultCount.toLocaleString('fa-IR')} نتیجه
          </span>
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
