import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { bulkUpdate } from '../keys'

const { put } = vi.hoisted(() => ({ put: vi.fn() }))

vi.mock('../client', () => ({ apiClient: { put } }))

describe('API key bulk updates', () => {
  beforeEach(() => {
    put.mockReset()
  })

  it('updates each selected string ID once and preserves partial failures', async () => {
    const failure = { response: { status: 403, data: { detail: 'Forbidden' } } }
    put.mockImplementation((url: string) => url === '/keys/key-b'
      ? Promise.reject(failure)
      : Promise.resolve({ data: { id: url.split('/').pop() } }))

    const updates = { group_id: 'group-12', quota: 0, expires_at: '', ip_whitelist: [] }
    const result = await bulkUpdate(['key-a', 'key-b', 'key-a', 'key-c'], updates)

    expect(put.mock.calls).toEqual([
      ['/keys/key-a', updates],
      ['/keys/key-b', updates],
      ['/keys/key-c', updates]
    ])
    expect(result).toEqual({
      succeededIds: ['key-a', 'key-c'],
      failures: [{ id: 'key-b', error: failure }]
    })
  })

  it('limits in-flight requests and continues after a failed batch member', async () => {
    const settle: Array<() => void> = []
    let inFlight = 0
    let peak = 0
    put.mockImplementation(() => new Promise((resolve, reject) => {
      inFlight++
      peak = Math.max(peak, inFlight)
      const index = settle.length
      settle.push(() => {
        inFlight--
        if (index === 0) reject(new Error('network error'))
        else resolve({ data: {} })
      })
    }))

    const ids = ['key-1', 'key-2', 'key-3', 'key-4', 'key-5', 'key-6', 'key-7']
    const pending = bulkUpdate(ids, { status: 'inactive' })
    expect(put).toHaveBeenCalledTimes(5)

    settle.slice(0, 4).forEach((finish) => finish())
    await flushPromises()
    expect(put).toHaveBeenCalledTimes(5)

    settle[4]()
    await flushPromises()
    expect(put).toHaveBeenCalledTimes(7)

    settle.slice(5).forEach((finish) => finish())
    const result = await pending

    expect(peak).toBe(5)
    expect(result.succeededIds).toEqual(ids.slice(1))
    expect(result.failures.map(({ id }) => id)).toEqual(['key-1'])
  })

  it('does not send requests when no keys are selected', async () => {
    expect(await bulkUpdate([], { quota: 0 })).toEqual({ succeededIds: [], failures: [] })
    expect(put).not.toHaveBeenCalled()
  })
})
