import apiClient from '../client'

export interface ModelFingerprintSnapshot {
  id: string
  model: string
  status: 'running' | 'completed' | 'failed'
  sampling_mode: 'single_conversation' | 'independent'
  protocol?: FingerprintProtocol
  resolved_protocol?: FingerprintProtocol
  reasoning_effort?: string
  source?: 'manual' | 'scheduled'
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

export type FingerprintProtocol = 'auto' | 'chat' | 'anthropic'
export interface FingerprintOptions { model_id: string; protocol: FingerprintProtocol; reasoning_effort: string }
export interface FingerprintModel { id: string; display_name: string; reasoning_levels: string[] | null }
export interface FingerprintSchedule { enabled: boolean; options: FingerprintOptions; next_run_at?: string }
export interface FingerprintHistory { items: ModelFingerprintSnapshot[]; total: number; page: number; page_size: number }

export async function startModelFingerprint(id: string, model: string, options?: Omit<FingerprintOptions, 'model_id'>) {
  const { data } = await apiClient.post<ModelFingerprintSnapshot>(`/admin/accounts/${id}/model-fingerprint`, { model_id: model, ...options })
  return data
}

export async function getFingerprintModels(id: string, signal?: AbortSignal) {
  const { data } = await apiClient.get<FingerprintModel[]>(`/admin/accounts/${id}/model-fingerprint/models`, { signal })
  return data
}
export async function getFingerprintHistory(id: string, page = 1) {
  const { data } = await apiClient.get<FingerprintHistory>(`/admin/accounts/${id}/model-fingerprint/history`, { params: { page, page_size: 10 } })
  return data
}
export async function getFingerprintSchedule(id: string) {
  const { data } = await apiClient.get<FingerprintSchedule>(`/admin/accounts/${id}/model-fingerprint/schedule`)
  return data
}
export async function setFingerprintSchedule(id: string, schedule: FingerprintSchedule) {
  const { data } = await apiClient.put<FingerprintSchedule>(`/admin/accounts/${id}/model-fingerprint/schedule`, schedule)
  return data
}

export async function getModelFingerprint(id: string, signal?: AbortSignal) {
  const { data } = await apiClient.get<ModelFingerprintSnapshot | null>(`/admin/accounts/${id}/model-fingerprint`, { signal })
  return data
}

export function isFingerprintTextModel(model: string) {
  return model.trim() !== '' && !/image|imagine|video|audio|voice|tts|stt|whisper|realtime|embedding|rerank|moderation|dall-e|sora|veo|imagen/i.test(model)
}
