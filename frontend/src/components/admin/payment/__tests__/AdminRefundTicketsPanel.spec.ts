import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminRefundTicketsPanel from '../AdminRefundTicketsPanel.vue'
import { formatCNY, formatPoints } from '@/utils/format'
import type { PaymentOrder, RefundTicket } from '@/types/payment'

const { getRefundTickets, getOrder, reviewRefundTicket, showError, showSuccess } = vi.hoisted(() => ({
  getRefundTickets: vi.fn(),
  getOrder: vi.fn(),
  reviewRefundTicket: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: { getRefundTickets, getOrder, reviewRefundTicket },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'en' },
    }),
  }
})

const BaseDialogStub = defineComponent({
  props: { show: Boolean },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const ticket: RefundTicket = {
  id: 'ticket-8',
  order_id: 'order-12',
  user_id: 'user-3',
  status: 'PENDING',
  comment: 'Needs manual review',
  created_at: '2026-08-20T00:00:00Z',
  updated_at: '2026-08-20T00:00:00Z',
}

const order: PaymentOrder = {
  id: 'order-12',
  user_id: 'user-3',
  amount: 100,
  principal_amount: 100,
  pay_amount: 102,
  base_points: 100,
  bonus_points: 20,
  credited_points: 120,
  affiliate_rebate_points: 12.5,
  wallet_amount: 0,
  wallet_bonus_amount: 0,
  wallet_recharge_amount: 0,
  gateway_base_amount: 100,
  wallet_only: false,
  fee_rate: 2,
  payment_type: 'wxpay',
  out_trade_no: 'trade-12',
  status: 'PARTIALLY_REFUNDED',
  order_type: 'balance',
  created_at: '2026-08-20T00:00:00Z',
  expires_at: '2026-08-20T00:10:00Z',
  refund_amount: 25,
  refunded_principal_amount: 25,
}

describe('AdminRefundTicketsPanel', () => {
  beforeEach(() => {
    getRefundTickets.mockReset().mockResolvedValue({
      data: { items: [ticket], total: 1, page: 1, page_size: 20 },
    })
    getOrder.mockReset().mockResolvedValue({ data: { order } })
    reviewRefundTicket.mockReset().mockResolvedValue({ data: {} })
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads the related order snapshot and submits a bounded manual-affiliate approval', async () => {
    const wrapper = mount(AdminRefundTicketsPanel, {
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true,
          Pagination: true,
          Select: true,
        },
      },
    })
    await flushPromises()

    const reviewButton = wrapper.findAll('button')
      .find((button) => button.text().includes('payment.admin.refundTickets.review'))
    await reviewButton?.trigger('click')
    await flushPromises()

    expect(getOrder).toHaveBeenCalledWith('order-12')
    expect(wrapper.text()).toContain(formatCNY(75, 'en'))
    expect(wrapper.text()).toContain(formatPoints(100, 'en'))
    expect(wrapper.text()).toContain(formatPoints(20, 'en'))
    expect(wrapper.text()).toContain(formatPoints(12.5, 'en'))

    const approvedPrincipal = wrapper.get('#approved-principal')
    expect((approvedPrincipal.element as HTMLInputElement).value).toBe('75')
    await approvedPrincipal.setValue('')
    await wrapper.get('#refund-review-note').setValue('Verified against the recharge order')

    const submitButton = wrapper.findAll('button')
      .find((button) => button.text().includes('payment.admin.refundTickets.submitReview'))
    expect(submitButton?.attributes('disabled')).toBeDefined()

    await approvedPrincipal.setValue('50')
    expect(submitButton?.attributes('disabled')).toBeUndefined()
    await submitButton?.trigger('click')
    await flushPromises()

    expect(reviewRefundTicket).toHaveBeenCalledWith('ticket-8', {
      decision: 'APPROVE',
      approved_principal_amount: 50,
      review_note: 'Verified against the recharge order',
      affiliate_action: 'MANUAL',
    })
  })
})
