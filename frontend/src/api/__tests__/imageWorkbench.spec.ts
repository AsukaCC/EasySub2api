import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  eligibleImageKeys,
  imagePlatformAdapters,
  listImageModels,
  loadWorkbenchCredentials,
  isAsyncImageTaskUnavailable,
  submitImageTask,
  cancelImageTask,
  retryImageTask,
  deleteImageTask,
  type ImageGenerationParams,
  type ImageTaskRequestError,
} from '@/api/imageWorkbench'
import { keysAPI } from '@/api/keys'
import { userGroupsAPI } from '@/api/groups'
import type { ApiKey, Group } from '@/types'

const adapter = imagePlatformAdapters[0]
const imageGroup = { id: 'images', platform: 'openai', allow_image_generation: true } as Group
const textGroup = { id: 'text', platform: 'openai', allow_image_generation: false } as Group

function imageKey(overrides: Partial<ApiKey> = {}): ApiKey {
  return { id: 'key-1', key: 'sk-test', status: 'active', group_id: imageGroup.id, group_ids: [imageGroup.id], group: imageGroup, ...overrides } as ApiKey
}

describe('image workbench model discovery', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('accepts an image-enabled secondary group on a multi-group key', () => {
    const key = imageKey({ group_id: textGroup.id, group_ids: [textGroup.id, imageGroup.id], group: textGroup })
    expect(eligibleImageKeys([key], adapter, [textGroup, imageGroup])).toEqual([key])
  })

  it('accepts image-enabled composite groups', () => {
    const key = imageKey({ group: { ...imageGroup, platform: 'composite' } })
    expect(eligibleImageKeys([key], adapter)).toEqual([key])
  })

  it('resolves group_ids without a legacy group or group_id', () => {
    const key = imageKey({ group: undefined, group_id: null })
    expect(eligibleImageKeys([key], adapter, [imageGroup])).toEqual([key])
  })

  it('keeps legacy single-group keys working', () => {
    const key = imageKey({ group_ids: [] })
    expect(eligibleImageKeys([key], adapter)).toEqual([key])
  })

  it('rejects inactive keys, disabled image groups, other platforms and unbound groups', () => {
    const keys = [
      imageKey({ status: 'inactive' }),
      imageKey({ group: { ...imageGroup, allow_image_generation: false } }),
      imageKey({ group: { ...imageGroup, platform: 'grok' } }),
      imageKey({ group_ids: ['missing'] }),
    ]
    expect(eligibleImageKeys(keys, adapter)).toEqual([])
    expect(eligibleImageKeys([imageKey({ group_ids: ['missing'], group: undefined })], adapter, [imageGroup])).toEqual([])
  })

  it('uses current group permissions from the available-group response', () => {
    expect(eligibleImageKeys([imageKey()], adapter, [{ ...imageGroup, allow_image_generation: false }])).toEqual([])
  })

  it('loads all active key pages using the server page size', async () => {
    const first = imageKey({ id: 'first', group: undefined })
    const second = imageKey({ id: 'second' })
    const list = vi.spyOn(keysAPI, 'list')
      .mockResolvedValueOnce({ items: [first], page: 1, page_size: 1, total: 2, pages: 2 })
      .mockResolvedValueOnce({ items: [second], page: 2, page_size: 1, total: 2, pages: 2 })
    vi.spyOn(userGroupsAPI, 'getAvailable').mockResolvedValue([imageGroup])

    const result = await loadWorkbenchCredentials()
    expect(result.keys.map((key) => key.id)).toEqual(['first', 'second'])
    expect(result.keys[0].group).toEqual(imageGroup)
    expect(list).toHaveBeenNthCalledWith(2, 2, 1, { status: 'active' })
  })

  it.each(['list', 'array'])('filters and deduplicates the gateway %s response', async (shape) => {
    const data = [null, {}, { id: 123 }, { id: 'gpt-6-astra' }, { id: 'gpt-image-2' }, { id: 'gpt-image-2' }, { id: 'gpt-image-2.5-flare' }]
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(shape === 'array' ? data : { object: 'list', data })))
    vi.stubGlobal('fetch', fetchMock)
    expect(await listImageModels('sk-test', adapter)).toEqual([{ id: 'gpt-image-2' }, { id: 'gpt-image-2.5-flare' }])
    expect(fetchMock).toHaveBeenCalledWith('http://localhost:3000/v1/models', { headers: { Authorization: 'Bearer sk-test' } })
  })

  it('distinguishes an empty catalog from an invalid response', async () => {
    vi.stubGlobal('fetch', vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ data: [] })))
      .mockResolvedValueOnce(new Response('<html>proxy error</html>')))
    await expect(listImageModels('sk-test', adapter)).resolves.toEqual([])
    await expect(listImageModels('sk-test', adapter)).rejects.toThrow('Invalid models response')
  })

  it('preserves gateway errors instead of reporting an empty catalog', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { message: 'API key expired' } }), { status: 401 })))
    await expect(listImageModels('sk-test', adapter)).rejects.toThrow('API key expired')
  })
})

const params: ImageGenerationParams = {
  prompt: 'a lighthouse',
  model: 'gpt-image-1',
  size: '1024x1024',
  quality: 'auto',
  output_format: 'png',
  output_compression: 100,
  background: 'opaque',
  moderation: 'auto',
  n: 1,
}

function requestError(status: number, message: string, details?: Record<string, unknown>): ImageTaskRequestError {
  const error = new Error(message) as ImageTaskRequestError
  error.status = status
  error.details = details
  return error
}

describe('image workbench async fallback signal', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it.each([
    requestError(404, 'async image tasks are not enabled'),
    requestError(503, 'failed to store image task request', { code: 'IMAGE_TASK_UNAVAILABLE' }),
    requestError(503, 'async image task storage is unavailable', { reason: 'IMAGE_TASK_UNAVAILABLE' }),
    requestError(404, 'Image task request failed (404)'),
    requestError(503, 'Image task request failed (503)'),
    requestError(503, 'upstream connect error or disconnect/reset before headers'),
  ])('recognizes unavailable async infrastructure', (error) => {
    expect(isAsyncImageTaskUnavailable(error)).toBe(true)
  })

  it.each([
    requestError(503, 'Our servers are currently overloaded', { type: 'server_error' }),
    requestError(404, 'Images API is not supported for this platform', { type: 'not_found_error' }),
    requestError(429, 'rate limit exceeded', { type: 'rate_limit_error' }),
  ])('keeps upstream and business errors terminal', (error) => {
    expect(isAsyncImageTaskUnavailable(error)).toBe(false)
  })

  it('preserves the task infrastructure status and error code', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ error: { code: 'IMAGE_TASK_UNAVAILABLE', message: 'failed to store image task request' } }),
      { status: 503, headers: { 'Content-Type': 'application/json' } },
    )))

    await expect(submitImageTask('sk-test', params)).rejects.toMatchObject({
      status: 503,
      details: { code: 'IMAGE_TASK_UNAVAILABLE' },
    })
  })

  it('retries the existing task id with its original generation parameters', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ id: 'imgtask_existing', task_id: 'imgtask_existing', status: 'queued' }),
      { status: 202, headers: { 'Content-Type': 'application/json' } },
    ))
    vi.stubGlobal('fetch', fetchMock)

    const result = await retryImageTask('sk-test', 'imgtask_existing', params)

    expect(result.task_id).toBe('imgtask_existing')
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:3000/v1/images/tasks/imgtask_existing/retry',
      expect.objectContaining({ method: 'POST', body: JSON.stringify(params) }),
    )
  })

  it('can retry without replacing the server-retained request body', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ id: 'imgtask_existing', task_id: 'imgtask_existing', status: 'queued' }),
      { status: 202, headers: { 'Content-Type': 'application/json' } },
    ))
    vi.stubGlobal('fetch', fetchMock)

    await retryImageTask('sk-test', 'imgtask_existing')

    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:3000/v1/images/tasks/imgtask_existing/retry',
      expect.objectContaining({ method: 'POST', body: undefined }),
    )
  })

  it('deletes a terminal task without requiring a response body', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(deleteImageTask('sk-test', 'imgtask_existing')).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:3000/v1/images/tasks/imgtask_existing',
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('accepts an empty response when a cancel races with terminal deletion', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 204 })))

    await expect(cancelImageTask('sk-test', 'imgtask_existing')).resolves.toBeUndefined()
  })
})
