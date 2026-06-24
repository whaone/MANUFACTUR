// Locale-aware number/currency formatting for the Indonesian locale.
// Centralizes the repeated `toLocaleString('id-ID')` calls across views.

// Thousands-grouped number, e.g. 1234567 → "1.234.567".
export function formatNumber(n: number): string {
  return n.toLocaleString('id-ID')
}

// Rupiah amount with the "Rp " prefix, e.g. 65000 → "Rp 65.000".
export function formatRupiah(n: number): string {
  return `Rp ${formatNumber(n)}`
}
