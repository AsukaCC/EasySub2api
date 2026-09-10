import { apiClient } from './client'

export interface UsageGuideResponse {
  content_md: string
  updated_at: string
}

export async function getUsageGuide(options?: { signal?: AbortSignal }): Promise<UsageGuideResponse> {
  const { data } = await apiClient.get<UsageGuideResponse>('/usage-guide', {
    signal: options?.signal,
  })
  return data
}

export const usageGuideAPI = { get: getUsageGuide }

