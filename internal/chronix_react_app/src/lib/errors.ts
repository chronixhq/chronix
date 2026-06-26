// Centralized helpers for formatting API errors from @dsherwin/react-api-interface
// Usage: in catch blocks of apiGet/apiPost/etc., call formatAPIError(err, 'Fallback message')
// This inspects the custom APIError shape exposed by the package:
// - err.APIErrorData: extended server error payload with fields like { code, message, details? }
// - err.status / err.statusText: HTTP information when available
// - err.cause or err.message for network or unexpected errors

export type APIErrorDataLike = {
  code?: string | number
  message?: string
  details?: unknown
  [k: string]: any
} | undefined

export type APIErrorLike = Error & {
  APIErrorData?: APIErrorDataLike
  status?: number
  statusText?: string
}

export function isAPIError(e: unknown): e is APIErrorLike {
  return !!(e && typeof e === 'object' && 'APIErrorData' in (e as any))
}

export function extractAPIErrorData(e: unknown): APIErrorDataLike {
  if (isAPIError(e)) return e.APIErrorData
  return undefined
}

export function formatAPIError(e: unknown, fallback = 'Request failed'): string {
  try {
    // Prefer explicit APIErrorData message when present
    const data = extractAPIErrorData(e)
    if (data && typeof data === 'object') {
      const msg = String((data as any).message || '').trim()
      if (msg) return msg
      // Sometimes servers send {error:"..."}
      const alt = String((data as any).error || '').trim()
      if (alt) return alt
      // Or nested details
      if ((data as any).details) {
        const det = (data as any).details
        if (typeof det === 'string' && det.trim()) return det.trim()
        if (Array.isArray(det) && det.length) {
          const first = det.find((d) => typeof d === 'string') as string | undefined
          if (first && first.trim()) return first.trim()
        }
      }
    }

    // Then try to use HTTP status text when present on error
    if (e && typeof e === 'object') {
      const anyErr = e as any
      const st = anyErr.statusText as string | undefined
      const code = anyErr.status as number | undefined
      if (st && String(st).trim()) return st
      if (code && Number.isFinite(code)) return `HTTP ${code}`
    }

    // Finally, fall back to error.message or provided fallback
    if (e instanceof Error && e.message) return e.message
    return fallback
  } catch {
    return fallback
  }
}

// Optional convenience for building a toast-friendly message with a consistent prefix
export function toastAPIError(e: unknown, base: string): string {
  const detail = formatAPIError(e, base)
  if (!detail || detail === base) return base
  // Avoid duplicating when detail already contains base
  if (detail.toLowerCase().startsWith(base.toLowerCase())) return detail
  return `${base}: ${detail}`
}
