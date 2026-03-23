/**
 * CSV Export Utility
 * Generates CSV with UTF-8 BOM for Excel compatibility and German column names.
 */

// UTF-8 BOM for Excel
const BOM = '\uFEFF'

export function generateCSV(
  headers: { key: string; label: string }[],
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  rows: Record<string, any>[],
): string {
  const headerLine = headers.map((h) => escapeCSV(h.label)).join(';')
  const dataLines = rows.map((row) =>
    headers.map((h) => escapeCSV(String(row[h.key] ?? ''))).join(';'),
  )
  return BOM + [headerLine, ...dataLines].join('\r\n')
}

function escapeCSV(value: string): string {
  if (value.includes(';') || value.includes('"') || value.includes('\n') || value.includes('\r')) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return value
}

export function downloadCSV(csv: string, filename: string): void {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.style.display = 'none'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

export function formatDateForExport(): string {
  return new Date().toISOString().split('T')[0]
}
