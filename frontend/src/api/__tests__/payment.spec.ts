import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, put, del } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  del: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
    put,
    delete: del,
  },
}))

import { paymentAPI } from '@/api/payment'
import { adminPaymentAPI } from '@/api/admin/payment'

describe('payment api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    get.mockResolvedValue({ data: {} })
    post.mockResolvedValue({ data: {} })
    put.mockReset()
    del.mockReset()
  })

  it('keeps legacy public out_trade_no verification for upgrade compatibility', async () => {
    await paymentAPI.verifyOrderPublic('legacy-order-no')

    expect(post).toHaveBeenCalledWith('/payment/public/orders/verify', {
      out_trade_no: 'legacy-order-no',
    })
  })

  it('keeps signed public resume-token resolve endpoint', async () => {
    await paymentAPI.resolveOrderPublicByResumeToken('resume-token-123')

    expect(post).toHaveBeenCalledWith('/payment/public/orders/resolve', {
      resume_token: 'resume-token-123',
    })
  })

  it('quotes and creates idempotent self-service refunds using principal RMB', async () => {
    await paymentAPI.getRefundQuote('order-7', 88.5)
    await paymentAPI.createRefund(
      'order-7',
      { principal_amount: 88.5, reason: 'duplicate recharge' },
      'refund-order-7-stable-key',
    )

    expect(get).toHaveBeenCalledWith('/payment/orders/order-7/refund-quote', {
      params: { principal_amount: 88.5 },
    })
    expect(post).toHaveBeenCalledWith(
      '/payment/orders/order-7/refunds',
      { principal_amount: 88.5, reason: 'duplicate recharge' },
      { headers: { 'Idempotency-Key': 'refund-order-7-stable-key' } },
    )
  })

  it('uses the refund ticket collection and pending-ticket cancel routes', async () => {
    await paymentAPI.createRefundTicket('order-9', { comment: 'manual review' })
    await paymentAPI.getMyRefundTickets({ page: 2, page_size: 10 })
    await paymentAPI.cancelRefundTicket('ticket-3')

    expect(post).toHaveBeenCalledWith('/payment/orders/order-9/refund-tickets', { comment: 'manual review' })
    expect(get).toHaveBeenCalledWith('/payment/refund-tickets', { params: { page: 2, page_size: 10 } })
    expect(post).toHaveBeenCalledWith('/payment/refund-tickets/ticket-3/cancel')
  })

  it('uses manual affiliate handling for admin refund-ticket reviews', async () => {
    await adminPaymentAPI.getRefundTickets({ status: 'PENDING', page: 1, page_size: 20 })
    await adminPaymentAPI.reviewRefundTicket('ticket-8', {
      decision: 'APPROVE',
      approved_principal_amount: 25,
      review_note: 'verified',
      affiliate_action: 'MANUAL',
    })

    expect(get).toHaveBeenCalledWith('/admin/payment/refund-tickets', {
      params: { status: 'PENDING', page: 1, page_size: 20 },
    })
    expect(post).toHaveBeenCalledWith('/admin/payment/refund-tickets/ticket-8/review', {
      decision: 'APPROVE',
      approved_principal_amount: 25,
      review_note: 'verified',
      affiliate_action: 'MANUAL',
    })
  })

  it('creates admin refunds with principal RMB and a stable idempotency header', async () => {
    await adminPaymentAPI.refundOrder(
      'order-12',
      { principal_amount: 45.25, reason: 'admin approved' },
      'admin-refund-order-12-stable-key',
    )

    expect(post).toHaveBeenCalledWith(
      '/admin/payment/orders/order-12/refund',
      { principal_amount: 45.25, reason: 'admin approved' },
      { headers: { 'Idempotency-Key': 'admin-refund-order-12-stable-key' } },
    )
  })
})
