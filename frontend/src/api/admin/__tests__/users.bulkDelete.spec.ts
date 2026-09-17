import { flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { bulkDelete } from '../users'

const { deleteRequest } = vi.hoisted(() => ({ deleteRequest: vi.fn() }))

vi.mock('../../client', () => ({
  apiClient: {
    delete: deleteRequest
  }
}))

describe('admin user bulk deletion', () => {
  beforeEach(() => {
    deleteRequest.mockReset()
  })

  it('deduplicates string IDs and preserves per-user failures', async () => {
    const failure = { response: { status: 400, data: { detail: 'cannot delete admin user' } } }
    deleteRequest.mockImplementation((url: string) => url.endsWith('/user-admin')
      ? Promise.reject(failure)
      : Promise.resolve({ data: { message: 'deleted', mode: 'permanently_deleted' } }))

    const result = await bulkDelete(['user-a', 'user-admin', 'user-a', 'user-b'])

    expect(deleteRequest.mock.calls).toEqual([
      ['/admin/users/user-a'],
      ['/admin/users/user-admin'],
      ['/admin/users/user-b']
    ])
    expect(result).toEqual({
      succeededIds: ['user-a', 'user-b'],
      failures: [{ id: 'user-admin', error: failure }]
    })
  })

  it('runs no more than three cascading deletions concurrently', async () => {
    const settle: Array<() => void> = []
    let inFlight = 0
    let peak = 0
    deleteRequest.mockImplementation(() => new Promise((resolve) => {
      inFlight++
      peak = Math.max(peak, inFlight)
      settle.push(() => {
        inFlight--
        resolve({ data: { message: 'deleted', mode: 'permanently_deleted' } })
      })
    }))

    const ids = ['user-1', 'user-2', 'user-3', 'user-4', 'user-5']
    const pending = bulkDelete(ids)
    expect(deleteRequest).toHaveBeenCalledTimes(3)

    settle.slice(0, 2).forEach((finish) => finish())
    await flushPromises()
    expect(deleteRequest).toHaveBeenCalledTimes(3)

    settle[2]()
    await flushPromises()
    expect(deleteRequest).toHaveBeenCalledTimes(5)

    settle.slice(3).forEach((finish) => finish())
    expect(await pending).toEqual({ succeededIds: ids, failures: [] })
    expect(peak).toBe(3)
  })

  it('does not send requests for an empty selection', async () => {
    expect(await bulkDelete([])).toEqual({ succeededIds: [], failures: [] })
    expect(deleteRequest).not.toHaveBeenCalled()
  })
})
