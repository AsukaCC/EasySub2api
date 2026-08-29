/**
 * User Payment API endpoints
 * Handles payment operations for regular users
 */

import { apiClient } from './client'
import type {
  PaymentConfig,
  SubscriptionPlan,
  MethodLimitsResponse,
  CheckoutInfoResponse,
  CreateOrderRequest,
  CreateOrderResult,
  PaymentOrder,
  PaymentRefund,
  RefundQuote,
  RefundTicket
} from '@/types/payment'
import type { BasePaginationResponse } from '@/types'

export interface PublicOrderVerifyResult {
  out_trade_no: string
  status: string
  paid: boolean
  created_at: string
  expires_at: string
}

export const paymentAPI = {
  /** Get payment configuration (enabled types, limits, etc.) */
  getConfig() {
    return apiClient.get<PaymentConfig>('/payment/config')
  },

  /** Get available subscription plans */
  getPlans() {
    return apiClient.get<SubscriptionPlan[]>('/payment/plans')
  },

  /** Get all checkout page data in a single call */
  getCheckoutInfo() {
    return apiClient.get<CheckoutInfoResponse>('/payment/checkout-info')
  },

  /** Get payment method limits and fee rates */
  getLimits() {
    return apiClient.get<MethodLimitsResponse>('/payment/limits')
  },

  /** Create a new payment order */
  createOrder(data: CreateOrderRequest) {
    return apiClient.post<CreateOrderResult>('/payment/orders', data)
  },

  /** Get current user's orders */
  getMyOrders(params?: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get<BasePaginationResponse<PaymentOrder>>('/payment/orders/my', { params })
  },

  /** Get a specific order by ID */
  getOrder(id: string) {
    return apiClient.get<PaymentOrder>(`/payment/orders/${id}`)
  },

  /** Cancel a pending order */
  cancelOrder(id: string) {
    return apiClient.post(`/payment/orders/${id}/cancel`)
  },

  /** Verify order payment status with upstream provider */
  verifyOrder(outTradeNo: string) {
    return apiClient.post<PaymentOrder>('/payment/orders/verify', { out_trade_no: outTradeNo })
  },

  /** Legacy-compatible public order lookup by out_trade_no */
  verifyOrderPublic(outTradeNo: string) {
    return apiClient.post<PublicOrderVerifyResult>('/payment/public/orders/verify', { out_trade_no: outTradeNo })
  },

  /** Resolve an order from a signed resume token without auth */
  resolveOrderPublicByResumeToken(resumeToken: string) {
    return apiClient.post<PublicOrderVerifyResult>('/payment/public/orders/resolve', { resume_token: resumeToken })
  },

  /** Quote the refundable RMB principal and point recovery for a recharge order. */
  getRefundQuote(id: string, principalAmount?: number) {
    return apiClient.get<RefundQuote>(`/payment/orders/${id}/refund-quote`, {
      params: principalAmount == null ? undefined : { principal_amount: principalAmount }
    })
  },

  /** Start an idempotent self-service refund. */
  createRefund(
    id: string,
    data: { principal_amount?: number; reason: string },
    idempotencyKey: string
  ) {
    return apiClient.post<PaymentRefund>(`/payment/orders/${id}/refunds`, data, {
      headers: { 'Idempotency-Key': idempotencyKey }
    })
  },

  /** Submit an out-of-window refund ticket for manual review. */
  createRefundTicket(id: string, data: { comment: string }) {
    return apiClient.post<RefundTicket>(`/payment/orders/${id}/refund-tickets`, data)
  },

  /** Get the authenticated user's refund tickets. */
  getMyRefundTickets(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<RefundTicket>>('/payment/refund-tickets', { params })
  },

  /** Cancel a pending refund ticket. */
  cancelRefundTicket(id: string) {
    return apiClient.post<RefundTicket>(`/payment/refund-tickets/${id}/cancel`)
  },

  /** Get provider instances that support direct refunds or reviewed tickets. */
  getRefundEligibleProviders() {
    return apiClient.get<{
      provider_instance_ids: string[]
      refund_enabled_provider_instance_ids?: string[]
    }>('/payment/orders/refund-eligible-providers')
  }
}
