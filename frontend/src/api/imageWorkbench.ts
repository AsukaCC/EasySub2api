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
    enabled: true,
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
