// Convert an HTML <input type="date"> value ("YYYY-MM-DD") to RFC3339 ISO,
// which Go's time.Time JSON decoder accepts. Empty/invalid → null.
export function toISO(d?: string | null): string | null {
  if (!d) return null
  const dt = new Date(d)
  return isNaN(dt.getTime()) ? null : dt.toISOString()
}
