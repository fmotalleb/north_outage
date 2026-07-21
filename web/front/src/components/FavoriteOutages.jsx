import OutageCard from './OutageCard'

function FavoriteSkeletonCard() {
  return (
    <div className="rounded-xl border border-amber-500/20 bg-amber-500/[0.03] p-4 flex flex-col">
      <div className="flex items-center gap-2 mb-3">
        <div className="h-5 w-16 rounded-full shimmer" />
        <div className="h-5 w-20 rounded-full shimmer" />
      </div>
      <div className="h-4 w-3/4 rounded shimmer mb-1" />
      <div className="h-4 w-1/2 rounded shimmer" />
      <div className="mt-4 grid grid-cols-2 gap-3">
        <div className="h-14 rounded-xl shimmer" />
        <div className="h-14 rounded-xl shimmer" />
      </div>
    </div>
  )
}

export default function FavoriteOutages({
  favoriteOutages,
  favoritesLoading,
  expandedId,
  setExpandedId,
  weatherProviderId,
  isFavorite,
  onRemoveFavorite,
  onRefresh,
}) {
  if (favoritesLoading && !favoriteOutages.length) {
    return (
      <section className="card p-4 md:p-5 border-amber-500/20">
        <div className="flex items-center gap-2 mb-4">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" stroke="none" className="text-amber-400">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
          </svg>
          <h2 className="text-sm font-semibold text-amber-200">علاقه‌مندی‌ها</h2>
          <div className="mr-auto">
            <div className="w-4 h-4 rounded-full border-2 border-amber-400/40 border-t-transparent animate-spin" />
          </div>
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {Array.from({ length: 2 }).map((_, i) => (
            <FavoriteSkeletonCard key={i} />
          ))}
        </div>
      </section>
    )
  }

  if (!favoriteOutages.length) return null

  return (
    <section className="card p-4 md:p-5 border-amber-500/20 bg-gradient-to-br from-amber-500/[0.04] to-transparent">
      {/* HEADER */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          {/* Star icon */}
          <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" stroke="none" className="text-amber-400">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
          </svg>
          <h2 className="text-sm font-semibold text-amber-200">علاقه‌مندی‌ها</h2>
          <span className="chip border border-amber-400/20 bg-amber-500/10 text-amber-300 text-[11px]">
            {favoriteOutages.length}
          </span>
          {favoritesLoading && (
            <div className="w-3 h-3 rounded-full border-2 border-amber-400/40 border-t-transparent animate-spin" />
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            className="btn-ghost text-[11px] px-2 py-1"
            onClick={onRefresh}
            disabled={favoritesLoading}
            type="button"
            title="بروزرسانی علاقه‌مندی‌ها"
          >
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={favoritesLoading ? 'animate-spin' : ''}>
              <polyline points="23 4 23 10 17 10" />
              <polyline points="1 20 1 14 7 14" />
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10" />
              <path d="M20.49 15a9 9 0 0 1-14.85 3.36L1 14" />
            </svg>
          </button>
        </div>
      </div>

      {/* GRID */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {favoriteOutages.map((o) => (
          <OutageCard
            key={o.unique_hash || o.id}
            outage={o}
            weatherProviderId={weatherProviderId}
            expanded={expandedId === o.unique_hash}
            onToggle={() =>
              setExpandedId(expandedId === o.unique_hash ? null : o.unique_hash)
            }
            showFavorite={true}
            isFavorited={isFavorite(o.unique_hash)}
            onToggleFavorite={() => onRemoveFavorite(o.unique_hash)}
          />
        ))}
      </div>
    </section>
  )
}
