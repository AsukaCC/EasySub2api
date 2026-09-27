import { apiClient } from './client'
import { createSettingsMemoryCache } from './settingsMemoryCache'
import type {
  SupportTicketCategory,
  SupportTicketDetail,
  SupportTicketPage,
  SupportTicketStatus,
  SupportTicketSummary,
} from '@/types/supportTicket'

export interface SupportTicketFilters {
  page?: number
  page_size?: number
  category?: SupportTicketCategory | ''
  status?: SupportTicketStatus | ''
  search?: string
  unread?: boolean
}

export interface CreateSupportTicketRequest {
  category: SupportTicketCategory
  title?: string
  message: string
  api_key_id?: string
  order_id?: string
}

const userBase = '/tickets'
const adminBase = '/admin/tickets'

// The top nav and admin sidebar both show the badge and remount on every page.
const userSummaryCache = createSettingsMemoryCache<SupportTicketSummary>()
const adminSummaryCache = createSettingsMemoryCache<SupportTicketSummary>()

export function clearSupportTicketSummaryCache(): void {
  userSummaryCache.clear()
  adminSummaryCache.clear()
}

if (typeof window !== 'undefined') {
  // Registered at module load so it runs before component listeners refetch.
  window.addEventListener('support-tickets:updated', clearSupportTicketSummaryCache)
}

async function loadSummary(
  cache: ReturnType<typeof createSettingsMemoryCache<SupportTicketSummary>>,
  url: string,
): Promise<{ data: SupportTicketSummary }> {
  const data = await cache.load(false, async () => (await apiClient.get<SupportTicketSummary>(url)).data)
  return { data }
}

async function mutate<T>(request: Promise<T>): Promise<T> {
  try {
    return await request
  } finally {
    clearSupportTicketSummaryCache()
  }
}

export const supportTicketsAPI = {
  list(params?: SupportTicketFilters) {
    return apiClient.get<SupportTicketPage>(userBase, { params })
  },
  summary() {
    return loadSummary(userSummaryCache, `${userBase}/summary`)
  },
  create(data: CreateSupportTicketRequest) {
    return mutate(apiClient.post<SupportTicketDetail>(userBase, data))
  },
  detail(id: string) {
    return apiClient.get<SupportTicketDetail>(`${userBase}/${id}`)
  },
  reply(id: string, message: string) {
    return mutate(apiClient.post<SupportTicketDetail>(`${userBase}/${id}/messages`, { message }))
  },
  markRead(id: string) {
    return mutate(apiClient.post(`${userBase}/${id}/read`))
  },
  action(id: string, action: 'cancel' | 'close' | 'reopen') {
    return mutate(apiClient.post<SupportTicketDetail>(`${userBase}/${id}/${action}`))
  },
}

export const adminSupportTicketsAPI = {
  list(params?: SupportTicketFilters) {
    return apiClient.get<SupportTicketPage>(adminBase, { params })
  },
  summary() {
    return loadSummary(adminSummaryCache, `${adminBase}/summary`)
  },
  createRefund(data: { order_id: string; approved_principal_amount: number; message: string }) {
    return mutate(apiClient.post(`${adminBase}`, data))
  },
  detail(id: string) {
    return apiClient.get<SupportTicketDetail>(`${adminBase}/${id}`)
  },
  reply(id: string, message: string) {
    return mutate(apiClient.post<SupportTicketDetail>(`${adminBase}/${id}/messages`, { message }))
  },
  markRead(id: string) {
    return mutate(apiClient.post(`${adminBase}/${id}/read`))
  },
  setStatus(id: string, status: SupportTicketStatus, message?: string) {
    return mutate(apiClient.post<SupportTicketDetail>(`${adminBase}/${id}/status`, { status, message }))
  },
  reviewRefund(id: string, data: { decision: 'APPROVE' | 'REJECT' | 'RETRY'; approved_principal_amount?: number; message?: string }) {
    return mutate(apiClient.post(`${adminBase}/${id}/refund/review`, data))
  },
}
