export function formatMoney(v: number | null | undefined): string {
  const n = Number(v ?? 0)
  const rounded = Math.round(n * 100) / 100
  const s = rounded.toLocaleString('en-IN', {
    minimumFractionDigits: Math.abs(rounded % 1) > 0.0001 ? 2 : 0,
    maximumFractionDigits: 2,
  })
  return `₹${s}`
}

export function formatTime(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleTimeString('en-IN', { hour: '2-digit', minute: '2-digit' })
}

export function formatDate(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })
}

export function formatDateTime(v: string | null | undefined): string {
  if (!v) return '—'
  return `${formatDate(v)}, ${formatTime(v)}`
}

/** "Today · 11:38 am", "Tomorrow · 9:30 am", else "16 Sep · 11:38 am" */
export function formatSchedule(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return '—'
  const now = new Date()
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
  const startOfDay = new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
  const dayDiff = Math.round((startOfDay - startOfToday) / 86400000)

  let dayLabel: string
  if (dayDiff === 0) dayLabel = 'Today'
  else if (dayDiff === 1) dayLabel = 'Tomorrow'
  else if (dayDiff === -1) dayLabel = 'Yesterday'
  else if (dayDiff > 1 && dayDiff < 7)
    dayLabel = d.toLocaleDateString('en-IN', { weekday: 'short' })
  else dayLabel = formatDate(v)

  return `${dayLabel} · ${formatTime(v)}`
}

/** Human-friendly elapsed time, e.g. "just now", "12m ago", "2h ago", "3d ago" */
export function timeAgo(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return '—'
  const diff = Date.now() - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 7) return `${days}d ago`
  return formatDate(v)
}

export function initials(name: string | null | undefined): string {
  if (!name) return '?'
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => w.charAt(0).toUpperCase())
    .join('')
}

export function greeting(): string {
  const h = new Date().getHours()
  if (h < 12) return 'Good morning'
  if (h < 17) return 'Good afternoon'
  return 'Good evening'
}

export function todayLabel(): string {
  return new Date().toLocaleDateString('en-IN', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  })
}