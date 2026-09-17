import { apiClient } from '../client'
import type { Account } from '@/types'

export type IntegrityMode = 'off' | 'observe' | 'enforce'
export interface ProtectionStrategy {
  id: string
  name: string
  identity_mode: string
  tls_profile: string
  diagnostic_only: boolean
  apply_supported: boolean
  requires_openai: boolean
}
export interface ProtectionPreview {
  enabled: boolean
  eligible: boolean
  active_mode: string
  reason?: string
  issues?: string[]
  changes: { key: string; from?: unknown; to: unknown }[]
  runtime?: {
    integrity_mode: IntegrityMode
    configured_tls: string
    effective_tls: string
    tls_reason?: string
    observed: boolean
    concurrency: number
  }
}
const path = (id: string) => `/admin/accounts/${encodeURIComponent(id)}/anti-degrade`
export const accountProtection = {
  async strategies() {
    return (await apiClient.get<{ strategies: ProtectionStrategy[] }>('/admin/accounts/anti-degrade/strategies')).data.strategies
  },
  async preview(id: string, mode = 'legacy') {
    return (await apiClient.get<ProtectionPreview>(path(id), { params: { mode } })).data
  },
  async apply(id: string, mode: string) {
    return (await apiClient.post<Account>(path(id) + '/apply', undefined, { params: { mode } })).data
  },
  async set(id: string, enabled: boolean) {
    return (await apiClient.put<Account>(path(id), { enabled, confirm_disable: !enabled })).data
  },
  async integrity(id: string, mode: IntegrityMode) {
    return (await apiClient.put<Account>(path(id) + '/integrity', { mode })).data
  },
  async batch(ids: string[]) {
    return (await apiClient.post<{ success_ids: string[]; failures: Record<string, string> }>(
      '/admin/accounts/anti-degrade/enable-batch', { account_ids: ids }
    )).data
  }
}
