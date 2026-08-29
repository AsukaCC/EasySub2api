import { apiClient } from '../client'
import type { AccountPlatform } from '@/types'

export interface AccountModelRule {
  id: string
  name: string
  description: string | null
  platform: AccountPlatform
  whitelist: string[]
  mapping: Record<string, string>
  created_at: string
  updated_at: string
}

export interface CreateAccountModelRuleRequest {
  name: string
  description?: string | null
  platform: AccountPlatform
  whitelist: string[]
  mapping: Record<string, string>
}

export type UpdateAccountModelRuleRequest = Partial<CreateAccountModelRuleRequest>

export async function list(platform?: AccountPlatform | ''): Promise<AccountModelRule[]> {
  const { data } = await apiClient.get<AccountModelRule[]>('/admin/account-model-rules', {
    params: platform ? { platform } : undefined
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

const accountModelRulesAPI = {
  list,
  getById,
  create,
  update,
  delete: deleteRule
}

export default accountModelRulesAPI
