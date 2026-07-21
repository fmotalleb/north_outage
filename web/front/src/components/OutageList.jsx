import OutageCard from './OutageCard'
import { getLocationId } from '../hooks/useFavorites'

function SkeletonCard() {
  return (
    <div className="card p-5">
      <div className="flex gap-2 mb-3">
        <div className="h-5 w-20 rounded-full shimmer" />
        <div className="h-5 w-24 rounded-full shimmer" />
        <div className="h-5 w-16 rounded-full shimmer" />
      </div>
      <div className="h-5 w-3/4 rounded shimmer" />
      <div className="mt-4 grid grid-cols-2 gap-3">
        <div className="h-16 rounded-xl shimmer" />
        <div className="h-16 rounded-xl shimmer" />
      </div>
    </div>
  )
}

export default function OutageList({
  loading,
  error,
  outages,
  expandedId,
  setExpandedId,
  weatherProviderId,
  onRetry,
  isFavorite,
  onToggleFavorite,
}) {
  if (loading) {
    return (
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="card p-8 text-center">
        <div className="mx-auto w-12 h-12 rounded-full bg-rose-500/15 border border-rose-500/30 flex items-center justify-center mb-3">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-rose-300">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
        </div>
        <h3 className="text-lg font-semibold text-slate-100 mb-1">خطا در دریافت اطلاعات</h3>
        <p className="text-sm text-slate-400 mb-4">{error}</p>
        <button className="btn-primary" onClick={onRetry}>
          تلاش دوباره
        </button>
      </div>
    )
  }

  if (!outages.length) {
    return (
      <div className="card p-10 text-center">
        <div className="mx-auto w-14 h-14 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mb-3">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-slate-400">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </div>
        <h3 className="text-lg font-semibold text-slate-100 mb-1">نتیجه‌ای یافت نشد</h3>
        <p className="text-sm text-slate-400">با فیلترهای فعلی هیچ رویدادی پیدا نشد. فیلترها را تغییر دهید.</p>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
      {outages.map((o) => (
        <OutageCard
          key={o.unique_hash || o.id}
          outage={o}
          weatherProviderId={weatherProviderId}
          expanded={expandedId === o.unique_hash}
          onToggle={() =>
            setExpandedId(expandedId === o.unique_hash ? null : o.unique_hash)
          }
          showFavorite={true}
          isFavorited={isFavorite ? isFavorite(getLocationId(o)) : false}
          onToggleFavorite={() => onToggleFavorite && onToggleFavorite(o)}
        />
      ))}
    </div>
  )
}
