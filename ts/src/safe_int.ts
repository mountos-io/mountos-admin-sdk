// Server-side int64 fields (ids, byte counts, ...) can exceed
// Number.MAX_SAFE_INTEGER; JSON.parse silently rounds such values to the
// nearest representable double. This walks a parsed JSON value and throws
// as soon as it finds an integer outside the safe range, so a precision
// loss surfaces as a loud error instead of a silently wrong id or count.
export class UnsafeIntegerError extends Error {
  constructor(path: string, value: number) {
    super(`value at ${path} (${value}) exceeds Number.MAX_SAFE_INTEGER and cannot be represented exactly as a JS number`)
    this.name = 'UnsafeIntegerError'
  }
}

export function assertSafeIntegers(value: unknown, path = '$'): void {
  if (typeof value === 'number') {
    if (Number.isInteger(value) && !Number.isSafeInteger(value)) {
      throw new UnsafeIntegerError(path, value)
    }
    return
  }
  if (Array.isArray(value)) {
    for (let i = 0; i < value.length; i++) assertSafeIntegers(value[i], `${path}[${i}]`)
    return
  }
  if (value !== null && typeof value === 'object') {
    for (const [k, v] of Object.entries(value)) assertSafeIntegers(v, `${path}.${k}`)
  }
}
