import { apiClient } from '../client'
import type { OpenAIReferralEligibility } from '@/types/openaiReferrals'

export interface OpenCodeGoUsageSettings { enabled: boolean; interval_minutes: number; debounce_minutes: number }
export interface OpenCodeGoUsageWindow { status: string; percent: number; resets_at: string }
export interface OpenCodeGoUsageState {
  account_id: string
  eligible: boolean
  auto_refresh_enabled: boolean
  snapshot?: {
    status: 'ok' | 'failed' | 'unauthorized'
    data?: { rolling?: OpenCodeGoUsageWindow; weekly?: OpenCodeGoUsageWindow; monthly?: OpenCodeGoUsageWindow }
    fetched_at?: string
    last_attempt_at: string
    next_refresh_at: string
    last_error?: string
  }
}
export interface CodexCredits { has_credits: boolean; unlimited: boolean; balance: string | null }
const usageBase = '/admin/accounts'
export async function getOpenCodeSettings() { return (await apiClient.get<OpenCodeGoUsageSettings>(`${usageBase}/opencode-go-usage/settings`)).data }
export async function saveOpenCodeSettings(value: OpenCodeGoUsageSettings) { return (await apiClient.put<OpenCodeGoUsageSettings>(`${usageBase}/opencode-go-usage/settings`, value)).data }
export async function refreshOpenCodeUsage(id: string) { return (await apiClient.post<OpenCodeGoUsageState>(`${usageBase}/${id}/opencode-go-usage/refresh`)).data }
export async function setOpenCodeAutoRefresh(id: string, enabled: boolean) { return (await apiClient.put<OpenCodeGoUsageState>(`${usageBase}/${id}/opencode-go-usage/auto-refresh`, { enabled })).data }
export async function refreshCodexCredits(id: string) { return (await apiClient.post<{ credits?: CodexCredits }>(`/admin/openai/accounts/${id}/quota/refresh`)).data }
export async function refreshCodexReferrals(id: string) { return (await apiClient.post<{ eligibility: OpenAIReferralEligibility; cache_persisted: boolean }>(`/admin/openai/accounts/${id}/referrals/refresh`)).data }
export async function sendCodexReferral(id: string, email: string, program_id: string) { return (await apiClient.post<{ sent: boolean; refresh_failed: boolean; eligibility?: OpenAIReferralEligibility }>(`/admin/openai/accounts/${id}/referrals/invite`, { email, program_id, confirmed: true })).data }
