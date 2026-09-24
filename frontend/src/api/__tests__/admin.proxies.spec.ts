import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get } = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('@/api/client', () => ({ apiClient: { get } }))

import { getAll, getAllWithCount, getStats, invalidateProxyListCache, list } from '@/api/admin/proxies'

describe.each([
  { name: 'paginated list', load: () => list(), wrap: (items: unknown[]) => ({ items, total: items.length, pages: 1 }) },
  { name: 'account selector', load: getAll, wrap: (items: unknown[]) => items },
  { name: 'account selector with counts', load: getAllWithCount, wrap: (items: unknown[]) => items },
])('$name', ({ load, wrap }) => {
  beforeEach(() => {
    get.mockReset()
    invalidateProxyListCache()
  })

  it.each(['', undefined, null, {}, { items: null }, { items: {} }])(
    'rejects an empty or malformed successful response: %j',
    async (data) => {
      get.mockResolvedValue({ data })
      await expect(load()).rejects.toThrow('Invalid proxy list response')
    },
  )

  it.each([{ items: [] }, { items: [{ id: 1, name: 'valid proxy' }] }])('preserves valid rows: %j', async ({ items }) => {
    const data = wrap(items)
    get.mockResolvedValue({ data })
    await expect(load()).resolves.toBe(data)
  })
})

describe('proxy statistics', () => {
  beforeEach(() => { get.mockReset() })

  it('preserves unavailable historical metrics instead of reporting a healthy proxy', async () => {
    const stats = {
      total_accounts: 3,
      active_accounts: 2,
      total_requests: null,
      success_rate: null,
      average_latency: null,
    }
    get.mockResolvedValue({ data: stats })
    await expect(getStats('proxy-id')).resolves.toEqual(stats)
    expect(get).toHaveBeenCalledWith('/admin/proxies/proxy-id/stats')
  })

  it('propagates statistics errors', async () => {
    get.mockRejectedValue(new Error('Statistics unavailable'))
    await expect(getStats('proxy-id')).rejects.toThrow('Statistics unavailable')
  })
})
