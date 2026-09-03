import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, del } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  del: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
    delete: del,
  },
}))

import {
  batchUpdateLimits,
  bindUserAuthIdentity,
  deleteUser,
  listArchived,
  permanentlyDeleteInactiveUsers,
  previewInactiveUsers,
  restoreArchivedUser,
  type AdminBindAuthIdentityRequest,
  type AdminBoundAuthIdentity,
  type BatchUpdateUserLimitsRequest,
  type BatchUpdateUserLimitsResponse,
} from '@/api/admin/users'

type Assert<T extends true> = T
type IsExact<T, U> = (
  (<G>() => G extends T ? 1 : 2) extends (<G>() => G extends U ? 1 : 2)
    ? ((<G>() => G extends U ? 1 : 2) extends (<G>() => G extends T ? 1 : 2) ? true : false)
    : false
)

type ExpectedAdminBindAuthIdentityRequest = {
  provider_type: string
  provider_key: string
  provider_subject: string
  issuer?: string
  metadata?: Record<string, unknown>
  channel?: {
    channel: string
    channel_app_id: string
    channel_subject: string
    metadata?: Record<string, unknown>
  }
}

type ExpectedAdminBoundAuthIdentity = {
  user_id: string
  provider_type: string
  provider_key: string
  provider_subject: string
  verified_at?: string | null
  issuer?: string | null
  metadata: Record<string, unknown> | null
  created_at: string
  updated_at: string
  channel?: {
    channel: string
    channel_app_id: string
    channel_subject: string
    metadata: Record<string, unknown> | null
    created_at: string
    updated_at: string
  } | null
}

const requestContractExact: Assert<
  IsExact<AdminBindAuthIdentityRequest, ExpectedAdminBindAuthIdentityRequest>
> = true
const responseContractExact: Assert<
  IsExact<AdminBoundAuthIdentity, ExpectedAdminBoundAuthIdentity>
> = true
const batchRequestContractExact: Assert<
  IsExact<
    BatchUpdateUserLimitsRequest,
    {
      user_ids: string[]
      all?: boolean
      concurrency?: number
      rpm_limit?: number
    }
  >
> = true
const batchResponseContractExact: Assert<
  IsExact<BatchUpdateUserLimitsResponse, { affected: number }>
> = true

describe('admin users api auth identity binding', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    del.mockReset()
  })

  it('posts the backend-compatible auth identity bind payload and returns the backend response shape', async () => {
    const payload: AdminBindAuthIdentityRequest = {
      provider_type: 'wechat',
      provider_key: 'wechat-main',
      provider_subject: 'union-123',
      metadata: { source: 'admin-repair' },
      channel: {
        channel: 'open',
        channel_app_id: 'wx-open',
        channel_subject: 'openid-123',
        metadata: { scene: 'migration' },
      },
    }

    const response: AdminBoundAuthIdentity = {
      user_id: 9,
      provider_type: 'wechat',
      provider_key: 'wechat-main',
      provider_subject: 'union-123',
      verified_at: '2026-04-22T00:00:00Z',
      issuer: null,
      metadata: { source: 'admin-repair' },
      created_at: '2026-04-22T00:00:00Z',
      updated_at: '2026-04-22T00:00:00Z',
      channel: {
        channel: 'open',
        channel_app_id: 'wx-open',
        channel_subject: 'openid-123',
        metadata: { scene: 'migration' },
        created_at: '2026-04-22T00:00:00Z',
        updated_at: '2026-04-22T00:00:00Z',
      },
    }
    post.mockResolvedValue({ data: response })

    const result = await bindUserAuthIdentity(9, payload)

    expect(post).toHaveBeenCalledWith('/admin/users/9/auth-identities', payload)
    expect(result).toEqual(response)
  })

  it('keeps bind auth identity request and response types aligned with the backend contract', () => {
    expect(requestContractExact).toBe(true)
    expect(responseContractExact).toBe(true)
  })

  it('posts batch limit updates once with only the supplied limit fields', async () => {
    const request: BatchUpdateUserLimitsRequest = {
      user_ids: [4, 7],
      all: false,
      rpm_limit: 0,
    }
    post.mockResolvedValue({ data: { affected: 2 } satisfies BatchUpdateUserLimitsResponse })

    const result = await batchUpdateLimits(request)

    expect(post).toHaveBeenCalledWith('/admin/users/batch-limits', request)
    expect(result).toEqual({ affected: 2 })
    expect(batchRequestContractExact).toBe(true)
    expect(batchResponseContractExact).toBe(true)
  })

  it('posts inactive cleanup preview and permanent delete payloads', async () => {
    const filters = {
      max_balance: 0,
      last_used_before: '2026-08-01T00:00:00Z',
      max_usage_7d: 0
    }
    post
      .mockResolvedValueOnce({
        data: {
          total: 2,
          total_balance: 0,
          total_usage_7d: 0,
          generated_at: '2026-09-03T00:00:00Z',
          snapshot_token: 'snapshot-2',
          items: []
        }
      })
      .mockResolvedValueOnce({ data: { deleted: 2 } })

    await previewInactiveUsers(filters)
    await permanentlyDeleteInactiveUsers({
      ...filters,
      expected_count: 2,
      snapshot_token: 'snapshot-2',
      confirmation: 'DELETE 2 USERS'
    })

    expect(post).toHaveBeenNthCalledWith(1, '/admin/users/inactive/preview', filters)
    expect(post).toHaveBeenNthCalledWith(2, '/admin/users/inactive/permanent-delete', {
      ...filters,
      expected_count: 2,
      snapshot_token: 'snapshot-2',
      confirmation: 'DELETE 2 USERS'
    })
  })

  it('lists, restores, and deletes users with archive disposition contracts', async () => {
    get.mockResolvedValue({ data: { items: [], total: 0, page: 1, page_size: 20, pages: 1 } })
    post.mockResolvedValue({ data: { id: '0199-user', email: 'restored@example.com' } })
    del.mockResolvedValue({ data: { message: 'archived', mode: 'archived' } })

    await listArchived(1, 20, 'paid@example.com')
    await restoreArchivedUser('0199-user')
    const result = await deleteUser('0199-user')

    expect(get).toHaveBeenCalledWith('/admin/users/archived', {
      params: { page: 1, page_size: 20, search: 'paid@example.com' }
    })
    expect(post).toHaveBeenCalledWith('/admin/users/0199-user/restore')
    expect(del).toHaveBeenCalledWith('/admin/users/0199-user')
    expect(result.mode).toBe('archived')
  })
})
