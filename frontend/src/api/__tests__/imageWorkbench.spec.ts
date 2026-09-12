import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  isAsyncImageTaskUnavailable,
  submitImageTask,
  type ImageGenerationParams,
  type ImageTaskRequestError,
} from '@/api/imageWorkbench'

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
})
