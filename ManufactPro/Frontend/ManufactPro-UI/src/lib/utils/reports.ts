// Pure helpers for the Reports view — formatting + margin bar visuals.
// UI-free so they can be unit-tested in isolation.

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

// Format an ISO "YYYY-MM-DD" date as "D Mon YYYY". Empty input → empty string.
export function formatDate(d: string): string {
  if (!d) return ''
  const parts = d.split('-')
  return `${parseInt(parts[2])} ${MONTHS[parseInt(parts[1]) - 1]} ${parts[0]}`
}

// Tailwind background class for a margin-percentage bar.
export function colorForBar(marginPct: number): string {
  if (marginPct >= 45) return 'bg-status-success'
  if (marginPct >= 35) return 'bg-primary'
  if (marginPct >= 20) return 'bg-status-warning'
  return 'bg-status-critical'
}

// Bar width (%) for a margin percentage, clamped to [15, 100].
export function barWidth(marginPct: number): number {
  return Math.min(100, Math.max(15, marginPct * 5))
}
