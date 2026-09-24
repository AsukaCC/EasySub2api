import { describe, expect, it, vi } from 'vitest'
import { createTtlCache } from '../ttlCache'

describe('createTtlCache', () => {
  it('reuses in-flight and cached values until TTL expires', async () => {
    const loader = vi.fn().mockResolvedValue('ok')
    const cache = createTtlCache(loader, 60_000)

    const first = cache.fetch('a')
    const second = cache.fetch('a')
    await expect(Promise.all([first, second])).resolves.toEqual(['ok', 'ok'])
    expect(loader).toHaveBeenCalledTimes(1)

    await expect(cache.fetch('a')).resolves.toBe('ok')
    expect(loader).toHaveBeenCalledTimes(1)
  })

  it('does not cache failed fetches', async () => {
    const loader = vi.fn()
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce('ok')
    const cache = createTtlCache(loader, 60_000)

    await expect(cache.fetch()).rejects.toThrow('offline')
    await expect(cache.fetch()).resolves.toBe('ok')
    expect(loader).toHaveBeenCalledTimes(2)
  })
})
