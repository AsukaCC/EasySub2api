import { buildGatewayUrl } from './url'
import { keysAPI } from './keys'
import { userGroupsAPI } from './groups'
import type { ApiKey, Group } from '@/types'

export type ImagePlatform = 'openai' | 'grok'

export interface ImagePlatformCapabilities {
  generation: boolean
  editing: boolean
  size: boolean
  quality: boolean
  outputFormat: boolean
  transparency: boolean
}

export interface ImagePlatformAdapter {
  id: ImagePlatform
  label: string
  enabled: boolean
  groupPlatform: string
  defaultModel: string
  isImageModel: (model: string) => boolean
  capabilities: ImagePlatformCapabilities
}

export const imagePlatformAdapters: ImagePlatformAdapter[] = [
  {
    id: 'openai',
    label: 'GPT',
    enabled: true,
    groupPlatform: 'openai',
    defaultModel: 'gpt-image-1',
    isImageModel: (model) => /^gpt-image(?:-|$)/i.test(model),
    capabilities: {
      generation: true,
      editing: true,
      size: true,
      quality: true,
      outputFormat: true,
      transparency: true,
    },
  },
  {
    id: 'grok',
    label: 'Grok',
    // Keep the adapter ready for a future rollout without exposing it in the first release.
    enabled: false,
    groupPlatform: 'grok',
    defaultModel: 'grok-imagine-image-quality',
    isImageModel: (model) => /^grok-imagine-image/i.test(model),
    capabilities: {
      generation: true,
      editing: true,
      size: true,
      quality: false,
      outputFormat: false,
      transparency: false,
    },
  },
]

export interface ImageModel {
  id: string
  name?: string
  owned_by?: string
}

export interface ImageGenerationParams {
  prompt: string
  model: string
  size: string
  quality: string
  output_format: string
  output_compression: number
  background: string
  moderation: string
  n: number
}

export interface ImageResult {
  url?: string
  b64_json?: string
  revised_prompt?: string
}

export interface ImageTaskError {
  message?: string
  type?: string
  code?: string
  request_id?: string
  [key: string]: unknown
}

export interface ImageTaskRequestError extends Error {
  status?: number
  details?: ImageTaskError
}

export interface ImageTask {
  id: string
  task_id: string
  object?: string
  status: 'queued' | 'processing' | 'completed' | 'failed' | 'canceled' | 'cancelled'
  result?: { data?: ImageResult[]; created?: number; [key: string]: unknown }
  error?: ImageTaskError
  http_status?: number
  created_at?: number
  completed_at?: number
  expires_at?: number
  poll_url?: string
}

function parseTaskError(payload: any, status: number): ImageTaskRequestError {
  const error = payload?.error || payload
  const message = error?.message || payload?.message || `Image task request failed (${status})`
  const result = new Error(message) as ImageTaskRequestError
  result.status = status
  result.details = error
  return result
}

// The async task endpoint can be unavailable when its Redis/object-storage
// infrastructure is disabled or unreachable. The workbench can still use the
// synchronous image endpoint in that deployment, while upstream and business
// errors must remain terminal errors.
export function isAsyncImageTaskUnavailable(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const candidate = error as ImageTaskRequestError
  if (candidate.status !== 404 && candidate.status !== 503) return false
  const message = String(candidate.message || '').trim().toLowerCase()
  const code = String(
    candidate.details?.code ||
    candidate.details?.reason ||
    candidate.details?.type ||
    '',
  ).trim().toLowerCase()
  if (code === 'image_task_unavailable') return true
  if (message.includes('async image tasks are not enabled') ||
    message.includes('async image object storage is unavailable') ||
    message.includes('failed to store image task request') ||
    message.includes('image task storage is unavailable')) return true

  // A reverse proxy can replace the gateway JSON body with a plain 503 page.
  // There is no upstream image request at this point, so an otherwise-untyped
  // submit failure is still an async infrastructure compatibility signal.
  if (candidate.status === 503 && !code) return true

  // Older gateways may not know the async route at all. Only treat generic
  // route-level 404s as a compatibility signal; business 404s such as an
  // unsupported platform must stay terminal and must not be retried as sync.
  return candidate.status === 404 && (
    message === 'image task request failed (404)' ||
    message === 'not found' ||
    message === '404 not found' ||
    message === '404 page not found'
  )
}

export async function submitImageTask(key: string, params: ImageGenerationParams, reference?: File): Promise<ImageTask> {
  const endpoint = reference ? '/v1/images/edits/async' : '/v1/images/generations/async'
  let response: Response
  if (reference) {
    const body = new FormData()
    for (const [name, value] of Object.entries(params)) body.append(name, String(value))
    body.append('image', reference)
    response = await fetch(buildGatewayUrl(endpoint), { method: 'POST', headers: { Authorization: `Bearer ${key}` }, body })
  } else {
    response = await fetch(buildGatewayUrl(endpoint), { method: 'POST', headers: { Authorization: `Bearer ${key}`, 'Content-Type': 'application/json' }, body: JSON.stringify(params) })
  }
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) throw parseTaskError(payload, response.status)
  return payload as ImageTask
}

export async function getImageTask(key: string, taskId: string): Promise<ImageTask> {
  const response = await fetch(buildGatewayUrl(`/v1/images/tasks/${encodeURIComponent(taskId)}`), { headers: { Authorization: `Bearer ${key}` } })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) throw parseTaskError(payload, response.status)
  return payload as ImageTask
}

export async function cancelImageTask(key: string, taskId: string): Promise<ImageTask | undefined> {
  const response = await fetch(buildGatewayUrl(`/v1/images/tasks/${encodeURIComponent(taskId)}`), { method: 'DELETE', headers: { Authorization: `Bearer ${key}` } })
  if (response.status === 204) return undefined
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) throw parseTaskError(payload, response.status)
  return payload as ImageTask
}

export async function retryImageTask(key: string, taskId: string, params?: ImageGenerationParams, reference?: File): Promise<ImageTask> {
  let body: BodyInit | undefined
  const headers: Record<string, string> = { Authorization: `Bearer ${key}` }
  if (params && reference) {
    const form = new FormData()
    for (const [name, value] of Object.entries(params)) form.append(name, String(value))
    form.append('image', reference)
    body = form
  } else if (params) {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(params)
  }
  const response = await fetch(buildGatewayUrl(`/v1/images/tasks/${encodeURIComponent(taskId)}/retry`), {
    method: 'POST',
    headers,
    body,
  })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) throw parseTaskError(payload, response.status)
  return payload as ImageTask
}

export async function deleteImageTask(key: string, taskId: string): Promise<void> {
  const response = await fetch(buildGatewayUrl(`/v1/images/tasks/${encodeURIComponent(taskId)}`), {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${key}` },
  })
  if (!response.ok) {
    const payload = await response.json().catch(() => ({}))
    throw parseTaskError(payload, response.status)
  }
}

export async function loadWorkbenchCredentials(): Promise<{ keys: ApiKey[]; groups: Group[] }> {
  const [keyPage, groups] = await Promise.all([
    keysAPI.list(1, 200, { status: 'active' }),
    userGroupsAPI.getAvailable(),
  ])
  const groupMap = new Map(groups.map((group) => [group.id, group]))
  const keys = keyPage.items.map((key) => ({
    ...key,
    group: key.group || (key.group_id ? groupMap.get(key.group_id) : undefined),
  }))
  return { keys, groups }
}

export function eligibleImageKeys(keys: ApiKey[], adapter: ImagePlatformAdapter): ApiKey[] {
  return keys.filter((key) => {
    const group = key.group
    return key.status === 'active' && group?.platform === adapter.groupPlatform && group.allow_image_generation === true
  })
}

export async function listImageModels(key: string, adapter: ImagePlatformAdapter): Promise<ImageModel[]> {
  const response = await fetch(buildGatewayUrl('/v1/models'), {
    headers: { Authorization: `Bearer ${key}` },
  })
  if (!response.ok) throw new Error(`Models request failed (${response.status})`)
  const payload = await response.json() as { data?: ImageModel[] } | ImageModel[]
  const models = Array.isArray(payload) ? payload : payload.data || []
  return models.filter((model) => adapter.isImageModel(model.id))
}

export async function generateImage(key: string, params: ImageGenerationParams, reference?: File): Promise<ImageResult[]> {
  const endpoint = reference ? '/v1/images/edits' : '/v1/images/generations'
  let response: Response
  if (reference) {
    const body = new FormData()
    body.append('model', params.model)
    body.append('prompt', params.prompt)
    body.append('size', params.size)
    body.append('quality', params.quality)
    body.append('output_format', params.output_format)
    body.append('output_compression', String(params.output_compression))
    body.append('background', params.background)
    body.append('moderation', params.moderation)
    body.append('n', String(params.n))
    body.append('image', reference)
    response = await fetch(buildGatewayUrl(endpoint), { method: 'POST', headers: { Authorization: `Bearer ${key}` }, body })
  } else {
    response = await fetch(buildGatewayUrl(endpoint), {
      method: 'POST',
      headers: { Authorization: `Bearer ${key}`, 'Content-Type': 'application/json' },
      body: JSON.stringify(params),
    })
  }
  const payload = await response.json().catch(() => ({})) as { data?: ImageResult[]; error?: { message?: string }; message?: string; request_id?: string; task_id?: string }
  if (!response.ok) throw new Error(payload.error?.message || payload.message || `Image request failed (${response.status})`)
  const taskId = payload.task_id || payload.request_id
  if ((!payload.data || payload.data.length === 0) && taskId) {
    for (let attempt = 0; attempt < 30; attempt += 1) {
      await new Promise((resolve) => window.setTimeout(resolve, 1000))
      const taskResponse = await fetch(buildGatewayUrl(`/v1/images/tasks/${encodeURIComponent(taskId)}`), { headers: { Authorization: `Bearer ${key}` } })
      const taskPayload = await taskResponse.json().catch(() => ({})) as { data?: ImageResult[]; status?: string; error?: { message?: string }; message?: string }
      if (!taskResponse.ok) throw new Error(taskPayload.error?.message || taskPayload.message || `Image task failed (${taskResponse.status})`)
      if (taskPayload.data?.length) return taskPayload.data
      if (['failed', 'cancelled', 'canceled'].includes(String(taskPayload.status).toLowerCase())) throw new Error(taskPayload.message || 'Image task failed')
    }
    throw new Error('Image generation timed out')
  }
  return payload.data || []
}
