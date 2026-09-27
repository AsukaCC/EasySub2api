const DEFAULT_TTL_MS = 30_000

type CacheEntry<T> = {
  identity: string
  value: T
  expiresAt: number
}

type InflightEntry<T> = {
  identity: string
  promise: Promise<T>
}

export function readAuthIdentity(): string {
  try {
    return localStorage.getItem('auth_token') || ''
  } catch {
    return ''
  }
}

function copyCachedValue<T>(value: T): T {
  if (typeof structuredClone === 'function') {
    try {
      return structuredClone(value)
    } catch {
      return value
    }
  }
  return value
}

/**
 * Short-lived in-memory cache for one settings payload.
 * Failed fetches are not stored. Callers must clear this on settings updates,
 * logout, and account switches; the TTL only covers a single page initialization.
 */
export function createSettingsMemoryCache<T>(ttlMs = DEFAULT_TTL_MS) {
  let cached: CacheEntry<T> | null = null
  let inflight: InflightEntry<T> | null = null
  let epoch = 0

  function clear(): void {
    epoch += 1
    cached = null
    inflight = null
  }

  function load(force: boolean, fetcher: () => Promise<T>): Promise<T> {
    const identity = readAuthIdentity()
    if (!force && cached && cached.identity === identity && Date.now() < cached.expiresAt) {
      return Promise.resolve(copyCachedValue(cached.value))
    }
    if (!force && inflight && inflight.identity === identity) {
      return inflight.promise
    }

    const writeEpoch = epoch
    const promise = fetcher()
      .then((value) => {
        const stored = copyCachedValue(value)
        if (writeEpoch === epoch && readAuthIdentity() === identity) {
          cached = {
            identity,
            value: stored,
            expiresAt: Date.now() + ttlMs,
          }
        }
        return copyCachedValue(stored)
      })
      .finally(() => {
        if (inflight?.promise === promise) {
          inflight = null
        }
      })

    inflight = { identity, promise }
    return promise
  }

  return { clear, load }
}

export const publicSettingsMemoryCache = createSettingsMemoryCache<unknown>()

export function clearPublicSettingsMemoryCache(): void {
  publicSettingsMemoryCache.clear()
}
