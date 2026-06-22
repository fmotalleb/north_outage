// Date utilities. The outage timestamps use Iran time (IRST, +03:30),
// but we render in the user's local timezone for clarity.

export function parseISO(iso) {
  return new Date(iso)
}

export function formatDateTime(iso,showDate = true, opts = {}) {
  if (!iso) return '—'
  const d = new Date(iso)
  const dateOpts = {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    ...opts.date,
  }
  const timeOpts = {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
    ...opts.time,
  }
  // Intl doesn't support separate date + time cleanly in one call in all browsers;
  // build them manually.
  const time = new Intl.DateTimeFormat('fa-IR', timeOpts).format(d)
  if (!showDate){
    return time
  }
  const date = new Intl.DateTimeFormat('fa-IR', dateOpts).format(d)
  return `${date} · ${time}`
}

export function formatTimeOnly(iso) {
  if (!iso) return '—'
  const d = new Date(iso)
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(d)
}

export function formatDateOnly(iso) {
  if (!iso) return '—'
  const d = new Date(iso)
  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  }).format(d)
}

export function durationMinutes(startISO, endISO) {
  const s = new Date(startISO).getTime()
  const e = new Date(endISO).getTime()
  return Math.max(0, Math.round((e - s) / 60000))
}

export function formatDuration(minutes) {
  if (!Number.isFinite(minutes) || minutes <= 0) return '—'
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  if (h === 0) return `${m} دقیقه`
  if (m === 0) return `${h} ساعت`
  return `${h} ساعت و ${m} دقیقه`
}

export function outageStatus(startISO, endISO, now = new Date()) {
  const s = new Date(startISO).getTime()
  const e = new Date(endISO).getTime()
  const n = now.getTime()
  if (n < s) return 'upcoming'
  if (n >= s && n < e) return 'active'
  return 'past'
}

export function relativeFromNow(iso, now = new Date()) {
  const diff = new Date(iso).getTime() - now.getTime()
  const abs = Math.abs(diff)
  const mins = Math.round(abs / 60000)
  const hours = Math.floor(mins / 60)
  const remMins = mins % 60
  let str
  if (mins < 60) str = `${mins} دقیقه`
  else if (hours < 24) str = `${hours} ساعت${remMins ? ' و ' + remMins + ' دقیقه' : ''}`
  else str = `${Math.floor(hours / 24)} روز`
  return diff >= 0 ? `${str} دیگر` : `${str} پیش`
}
