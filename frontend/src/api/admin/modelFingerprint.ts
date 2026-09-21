import apiClient from '../client'

export interface ModelFingerprintSnapshot {
  id: string
  model: string
  status: 'running' | 'completed' | 'failed'
  sampling_mode: 'single_conversation'
  completed: number
  valid: number
  total: number
  started_at: string
  expires_at: string
  finished_at?: string
  error?: string
  result?: {
    models: { model: string; family: string; probability: number }[]
    families: { family: string; display_name: string; probability: number }[]
    revision: string
  }
}

export async function startModelFingerprint(id: string, model: string) {
  const { data } = await apiClient.post<ModelFingerprintSnapshot>(`/admin/accounts/${id}/model-fingerprint`, { model_id: model })
  return data
}

export async function getModelFingerprint(id: string, signal?: AbortSignal) {
  const { data } = await apiClient.get<ModelFingerprintSnapshot | null>(`/admin/accounts/${id}/model-fingerprint`, { signal })
  return data
}

export function isFingerprintTextModel(model: string) {
  return model.trim() !== '' && !/image|imagine|video|audio|voice|tts|stt|whisper|realtime|embedding|rerank|moderation|dall-e|sora|veo|imagen/i.test(model)
}
