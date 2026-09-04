import { apiClient } from '../client'
import type { AccountPlatform } from '@/types'

export interface AccountModelRoute {
  request_model: string
  upstream_model: string
  reasoning_effort?: string
}

export interface AccountModelRule {
  id: string
  name: string
  description: string | null
  platform: AccountPlatform
  subscription_tier: string | null
  model_routes: AccountModelRoute[]
  bound_account_count?: number
  created_at: string
  updated_at: string
}

export interface CreateAccountModelRuleRequest {
  name: string
  description?: string | null
  platform: AccountPlatform
  subscription_tier?: string | null
  model_routes: AccountModelRoute[]
}

export type UpdateAccountModelRuleRequest = Partial<CreateAccountModelRuleRequest>

export async function list(platform?: AccountPlatform | '', subscriptionTier?: string): Promise<AccountModelRule[]> {
  const params: Record<string, string> = {}
  if (platform) params.platform = platform
  if (subscriptionTier) params.subscription_tier = subscriptionTier
  const { data } = await apiClient.get<AccountModelRule[]>('/admin/account-model-rules', {
    params: Object.keys(params).length ? params : undefined
  })
  return data
}

export async function getById(id: string): Promise<AccountModelRule> {
  const { data } = await apiClient.get<AccountModelRule>(`/admin/account-model-rules/${id}`)
  return data
}

export async function create(payload: CreateAccountModelRuleRequest): Promise<AccountModelRule> {
  const { data } = await apiClient.post<AccountModelRule>('/admin/account-model-rules', payload)
  return data
}

export async function update(id: string, payload: UpdateAccountModelRuleRequest): Promise<AccountModelRule> {
  const { data } = await apiClient.put<AccountModelRule>(`/admin/account-model-rules/${id}`, payload)
  return data
}

export async function deleteRule(id: string): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/account-model-rules/${id}`)
  return data
}

export default { list, getById, create, update, delete: deleteRule }
