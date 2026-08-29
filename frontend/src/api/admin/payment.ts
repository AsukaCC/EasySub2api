/**
 * Admin Payment API endpoints
 * Handles payment management operations for administrators
 */

import { apiClient } from '../client'
import type {
  DashboardStats,
  PaymentOrder,
  PaymentChannel,
  SubscriptionPlan,
  ProviderInstance,
  PaymentRefund,
  RechargeBonusTier,
  RefundTicket
} from '@/types/payment'
import type { BasePaginationResponse } from '@/types'

/** Admin-facing payment config returned by GET /admin/payment/config */
export interface AdminPaymentConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  order_timeout_minutes: number
  max_pending_orders: number
  enabled_payment_types: string[]
  balance_disabled: boolean
  /** Deprecated compatibility field; recharge points are permanently 1 RMB = 1 point. */
  balance_recharge_multiplier: number
  recharge_bonus_tiers: RechargeBonusTier[]
  /** Deprecated compatibility field; subscription prices are point-denominated. */
  subscription_usd_to_cny_rate: number
  recharge_fee_rate: number
  load_balance_strategy: string
  product_name_prefix: string
  product_name_suffix: string
  help_image_url: string
  help_text: string
}

/** Fields accepted by PUT /admin/payment/config (all optional via pointer semantics) */
export interface UpdatePaymentConfigRequest {
  enabled?: boolean
  min_amount?: number
  max_amount?: number
  daily_limit?: number
  order_timeout_minutes?: number
  max_pending_orders?: number
  enabled_payment_types?: string[]
  balance_disabled?: boolean
  recharge_bonus_tiers?: RechargeBonusTier[]
  recharge_fee_rate?: number
  load_balance_strategy?: string
  product_name_prefix?: string
  product_name_suffix?: string
  help_image_url?: string
  help_text?: string
}

export interface RefundTicketReviewResult {
  ticket: RefundTicket
  refund?: PaymentRefund
}

export interface AdminPaymentOrderDetail {
  order: PaymentOrder
  auditLogs?: Array<Record<string, unknown>>
}

export const adminPaymentAPI = {
  // ==================== Config ====================

  /** Get payment configuration (admin view) */
  getConfig() {
    return apiClient.get<AdminPaymentConfig>('/admin/payment/config')
  },

  /** Update payment configuration */
  updateConfig(data: UpdatePaymentConfigRequest) {
    return apiClient.put('/admin/payment/config', data)
  },

  // ==================== Dashboard ====================

  /** Get payment dashboard statistics */
  getDashboard(days?: number) {
    return apiClient.get<DashboardStats>('/admin/payment/dashboard', {
      params: days ? { days } : undefined
    })
  },

  // ==================== Orders ====================

  /** Get all orders (paginated, with filters) */
  getOrders(params?: {
    page?: number
    page_size?: number
    status?: string
    payment_type?: string
    user_id?: string
    keyword?: string
    start_date?: string
    end_date?: string
    order_type?: string
  }) {
    return apiClient.get<BasePaginationResponse<PaymentOrder>>('/admin/payment/orders', { params })
  },

  /** Get a specific order by ID */
  getOrder(id: string) {
    return apiClient.get<AdminPaymentOrderDetail>(`/admin/payment/orders/${id}`)
  },

  /** Cancel an order (admin) */
  cancelOrder(id: string) {
    return apiClient.post(`/admin/payment/orders/${id}/cancel`)
  },

  /** Retry recharge for a failed order */
  retryRecharge(id: string) {
    return apiClient.post(`/admin/payment/orders/${id}/retry`)
  },

  /** Process a refund */
  refundOrder(
    id: string,
    data: { principal_amount?: number; reason: string },
    idempotencyKey: string
  ) {
    return apiClient.post<PaymentRefund>(`/admin/payment/orders/${id}/refund`, data, {
      headers: { 'Idempotency-Key': idempotencyKey }
    })
  },

  /** Query and finalize a pending refund */
  queryRefund(id: string) {
    return apiClient.post<PaymentRefund>(`/admin/payment/orders/${id}/refund/query`)
  },

  /** Get refund tickets requiring or awaiting manual review. */
  getRefundTickets(params?: { status?: string; page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<RefundTicket>>('/admin/payment/refund-tickets', { params })
  },

  /** Approve or reject a refund ticket. Affiliate handling is manual by contract. */
  reviewRefundTicket(
    id: string,
    data: {
      decision: 'APPROVE' | 'REJECT'
      approved_principal_amount?: number
      review_note: string
      affiliate_action: 'MANUAL'
    }
  ) {
    return apiClient.post<RefundTicketReviewResult>(`/admin/payment/refund-tickets/${id}/review`, data)
  },

  // ==================== Channels ====================

  /** Get all payment channels */
  getChannels() {
    return apiClient.get<PaymentChannel[]>('/admin/payment/channels')
  },

  /** Create a payment channel */
  createChannel(data: Partial<PaymentChannel>) {
    return apiClient.post<PaymentChannel>('/admin/payment/channels', data)
  },

  /** Update a payment channel */
  updateChannel(id: string, data: Partial<PaymentChannel>) {
    return apiClient.put<PaymentChannel>(`/admin/payment/channels/${id}`, data)
  },

  /** Delete a payment channel */
  deleteChannel(id: string) {
    return apiClient.delete(`/admin/payment/channels/${id}`)
  },

  // ==================== Subscription Plans ====================

  /** Get all subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/admin/payment/plans')
  },

  /** Create a subscription plan */
  createPlan(data: Record<string, unknown>) {
    return apiClient.post<SubscriptionPlan>('/admin/payment/plans', data)
  },

  /** Update a subscription plan */
  updatePlan(id: string, data: Record<string, unknown>) {
    return apiClient.put<SubscriptionPlan>(`/admin/payment/plans/${id}`, data)
  },

  /** Delete a subscription plan */
  deletePlan(id: string) {
    return apiClient.delete(`/admin/payment/plans/${id}`)
  },

  // ==================== Provider Instances ====================

  /** Get all provider instances */
  getProviders() {
    return apiClient.get<ProviderInstance[]>('/admin/payment/providers')
  },

  /** Create a provider instance */
  createProvider(data: Partial<ProviderInstance>) {
    return apiClient.post<ProviderInstance>('/admin/payment/providers', data)
  },

  /** Update a provider instance */
  updateProvider(id: string, data: Partial<ProviderInstance>) {
    return apiClient.put<ProviderInstance>(`/admin/payment/providers/${id}`, data)
  },

  /** Delete a provider instance */
  deleteProvider(id: string) {
    return apiClient.delete(`/admin/payment/providers/${id}`)
  }
}

export default adminPaymentAPI
