export function clampDecimal(raw: string): string {
  let s = raw.replace(/,/g, '.').replace(/[^\d.]/g, '')
  const dot = s.indexOf('.')
  if (dot === -1) {
    return s
  }
  return s.slice(0, dot + 1) + s.slice(dot + 1).replace(/\./g, '').slice(0, 2)
}

export function optionalFloat(raw: string, label: string): number | undefined {
  const t = raw.trim().replace(/,/g, '.').replace(/\.$/, '')
  if (!t) {
    return undefined
  }
  if (!/^(?:\d+(?:\.\d{1,2})?|\.\d{1,2})$/.test(t)) {
    throw new Error(`${label} must be a number with at most two digits after the decimal point`)
  }
  const n = Number(t)
  if (!Number.isFinite(n)) {
    throw new Error(`${label} must be a number`)
  }
  return n
}

export function formatDecimal(n: number): string {
  return String(Number(n.toFixed(2)))
}
