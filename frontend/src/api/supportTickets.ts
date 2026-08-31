import { apiClient } from './client'
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

export const supportTicketsAPI = {
  list(params?: SupportTicketFilters) {
    return apiClient.get<SupportTicketPage>(userBase, { params })
  },
  summary() {
    return apiClient.get<SupportTicketSummary>(`${userBase}/summary`)
  },
  create(data: CreateSupportTicketRequest) {
    return apiClient.post<SupportTicketDetail>(userBase, data)
  },
  detail(id: string) {
    return apiClient.get<SupportTicketDetail>(`${userBase}/${id}`)
  },
  reply(id: string, message: string) {
    return apiClient.post<SupportTicketDetail>(`${userBase}/${id}/messages`, { message })
  },
  markRead(id: string) {
    return apiClient.post(`${userBase}/${id}/read`)
  },
  action(id: string, action: 'cancel' | 'close' | 'reopen') {
    return apiClient.post<SupportTicketDetail>(`${userBase}/${id}/${action}`)
  },
}

export const adminSupportTicketsAPI = {
  list(params?: SupportTicketFilters) {
    return apiClient.get<SupportTicketPage>(adminBase, { params })
  },
  summary() {
    return apiClient.get<SupportTicketSummary>(`${adminBase}/summary`)
  },
  createRefund(data: { order_id: string; approved_principal_amount: number; message: string }) {
    return apiClient.post(`${adminBase}`, data)
  },
  detail(id: string) {
    return apiClient.get<SupportTicketDetail>(`${adminBase}/${id}`)
  },
  reply(id: string, message: string) {
    return apiClient.post<SupportTicketDetail>(`${adminBase}/${id}/messages`, { message })
  },
  markRead(id: string) {
    return apiClient.post(`${adminBase}/${id}/read`)
  },
  setStatus(id: string, status: SupportTicketStatus, message?: string) {
    return apiClient.post<SupportTicketDetail>(`${adminBase}/${id}/status`, { status, message })
  },
  reviewRefund(id: string, data: { decision: 'APPROVE' | 'REJECT' | 'RETRY'; approved_principal_amount?: number; message?: string }) {
    return apiClient.post(`${adminBase}/${id}/refund/review`, data)
  },
}
