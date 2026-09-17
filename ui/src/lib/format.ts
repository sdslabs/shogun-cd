const relativeFormatter = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" })

export function formatRelativeTime(value?: string | null) {
  if (!value) return "Never"
  const date = new Date(value)
  const seconds = Math.round((date.getTime() - Date.now()) / 1000)
  const absolute = Math.abs(seconds)
  if (absolute < 60) return relativeFormatter.format(seconds, "second")
  if (absolute < 3_600) return relativeFormatter.format(Math.round(seconds / 60), "minute")
  if (absolute < 86_400) return relativeFormatter.format(Math.round(seconds / 3_600), "hour")
  if (absolute < 2_592_000) return relativeFormatter.format(Math.round(seconds / 86_400), "day")
  return new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", year: "numeric" }).format(date)
}

export function formatDateTime(value?: string | null) {
  if (!value) return "—"
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(value))
}

export function formatDuration(start?: string | null, finish?: string | null) {
  if (!start) return "—"
  const end = finish ? new Date(finish).getTime() : Date.now()
  const elapsed = Math.max(0, end - new Date(start).getTime())
  if (elapsed < 1_000) return `${elapsed}ms`
  const seconds = Math.floor(elapsed / 1_000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  if (minutes < 60) return `${minutes}m ${remainingSeconds}s`
  const hours = Math.floor(minutes / 60)
  return `${hours}h ${minutes % 60}m`
}

export function shorten(value: string, length = 24) {
  return value.length <= length ? value : `${value.slice(0, length - 1)}…`
}
