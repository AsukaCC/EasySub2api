type CacheEntry<T> = {
  value: T
  expiresAt: number
}

/**
 * In-flight-aware TTL cache. Failed fetches are not stored.
 */
export function createTtlCache<TArgs extends unknown[], TResult>(
  loader: (...args: TArgs) => Promise<TResult>,
  ttlMs = 60_000
) {
  const cache = new Map<string, CacheEntry<TResult>>()
  const inFlight = new Map<string, Promise<TResult>>()

  function keyOf(args: TArgs): string {
    return JSON.stringify(args)
  }

  async function fetch(...args: TArgs): Promise<TResult> {
    const key = keyOf(args)
    const now = Date.now()
    const hit = cache.get(key)
    if (hit && now < hit.expiresAt) {
      return hit.value
    }

    const pending = inFlight.get(key)
    if (pending) {
      return pending
    }

    const promise = loader(...args)
      .then((value) => {
        cache.set(key, { value, expiresAt: Date.now() + ttlMs })
        inFlight.delete(key)
        return value
      })
      .catch((error) => {
        inFlight.delete(key)
        throw error
      })

    inFlight.set(key, promise)
    return promise
  }

  function clear(): void {
    cache.clear()
    inFlight.clear()
  }

  return { fetch, clear }
}
