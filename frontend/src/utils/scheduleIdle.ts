/**
 * Schedule non-critical work after the browser is idle, with a timeout fallback.
 */
export function scheduleIdle(callback: () => void, timeout = 2000): number | ReturnType<typeof setTimeout> {
  if (typeof window !== 'undefined' && typeof window.requestIdleCallback === 'function') {
    return window.requestIdleCallback(() => callback(), { timeout })
  }
  return setTimeout(callback, Math.min(timeout, 200))
}
