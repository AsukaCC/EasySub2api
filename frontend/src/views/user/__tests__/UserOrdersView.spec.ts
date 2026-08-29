import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import UserOrdersView from '../UserOrdersView.vue'
import { formatCNY, formatPoints } from '@/utils/format'
import type { PaymentOrder, RefundQuote } from '@/types/payment'

const {
  getMyOrders,
  cancelOrder,
  getRefundEligibleProviders,
  getRefundQuote,
  createRefund,
  createRefundTicket,
  getMyRefundTickets,
  cancelRefundTicket,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getMyOrders: vi.fn(),
  cancelOrder: vi.fn(),
  getRefundEligibleProviders: vi.fn(),
  getRefundQuote: vi.fn(),
  createRefund: vi.fn(),
  createRefundTicket: vi.fn(),
  getMyRefundTickets: vi.fn(),
  cancelRefundTicket: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getMyOrders,
    cancelOrder,
    getRefundEligibleProviders,
    getRefundQuote,
    createRefund,
    createRefundTicket,
    getMyRefundTickets,
    cancelRefundTicket,
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: () => 'refund error',
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

const OrderTableStub = defineComponent({
  props: { orders: { type: Array, default: () => [] } },
  template: '<div><div v-for="row in orders" :key="row.id"><slot name="actions" :row="row" /></div></div>',
})

const order: PaymentOrder = {
  id: 'order-7',
  user_id: 'user-2',
  amount: 100,
  principal_amount: 100,
  pay_amount: 102,
  base_points: 100,
  bonus_points: 20,
  credited_points: 120,
  wallet_amount: 0,
  wallet_bonus_amount: 0,
  wallet_recharge_amount: 0,
  gateway_base_amount: 100,
  wallet_only: false,
  fee_rate: 2,
  payment_type: 'wxpay',
  out_trade_no: 'trade-7',
  status: 'COMPLETED',
  order_type: 'balance',
  provider_instance_id: 'provider-1',
  created_at: '2026-08-20T00:00:00Z',
  expires_at: '2026-08-20T00:10:00Z',
  refund_amount: 0,
}

const quote: RefundQuote = {
  order_id: 'order-7',
  currency: 'CNY',
  requested_principal_amount: 50,
  principal_amount: 50,
  fee_amount: 1,
  gateway_amount: 51,
  base_points: 50,
  bonus_points: 10,
  points_to_hold: 58,
  bonus_expired_offset: 2,
  affiliate_rebate_points: 5,
  remaining_principal_amount: 100,
  max_refundable_principal_amount: 50,
  refund_deadline: '2026-08-27T00:00:00Z',
  self_service_eligible: true,
  requires_ticket: false,
}

function mountView() {
  return mount(UserOrdersView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: BaseDialogStub,
        OrderTable: OrderTableStub,
        Pagination: true,
        Select: true,
        Icon: true,
      },
    },
  })
}

describe('UserOrdersView refunds', () => {
  beforeEach(() => {
    getMyOrders.mockReset().mockResolvedValue({ data: { items: [order], total: 1 } })
    getRefundEligibleProviders.mockReset().mockResolvedValue({
      data: { provider_instance_ids: ['provider-1'] },
    })
    getRefundQuote.mockReset().mockResolvedValue({ data: quote })
    getMyRefundTickets.mockReset().mockResolvedValue({ data: { items: [], total: 0 } })
    createRefund.mockReset()
    createRefundTicket.mockReset()
    cancelOrder.mockReset()
    cancelRefundTicket.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('shows the RMB/point quote and reuses the refund key when a submission is retried', async () => {
    vi.spyOn(globalThis.crypto, 'randomUUID').mockReturnValue('refund-intent-7')
    createRefund
      .mockRejectedValueOnce(new Error('temporary failure'))
      .mockResolvedValueOnce({ data: { status: 'SUCCEEDED' } })
    const wrapper = mountView()
    await flushPromises()

    const refundButton = wrapper.findAll('button')
      .find((button) => button.text().includes('payment.orders.requestRefund'))
    await refundButton?.trigger('click')
    await flushPromises()

    expect(getRefundQuote).toHaveBeenCalledWith('order-7', undefined)
    expect(wrapper.text()).toContain(formatCNY(51, 'en'))
    expect(wrapper.text()).toContain(formatPoints(50, 'en'))
    expect(wrapper.text()).toContain(formatPoints(10, 'en'))
    expect(wrapper.text()).toContain(formatPoints(2, 'en'))
    expect(wrapper.text()).toContain(formatPoints(58, 'en'))
    expect(wrapper.text()).toContain(formatPoints(5, 'en'))
    expect((wrapper.get('#refund-principal').element as HTMLInputElement).value).toBe('100')
    expect(wrapper.text()).toContain('payment.refundQuote.clamped')

    await wrapper.find('textarea').setValue('duplicate recharge')
    const confirmButton = wrapper.findAll('button')
      .find((button) => button.text().includes('payment.refundQuote.confirm'))
    await confirmButton?.trigger('click')
    await flushPromises()
    await confirmButton?.trigger('click')
    await flushPromises()

    expect(createRefund).toHaveBeenCalledTimes(2)
    expect(createRefund).toHaveBeenNthCalledWith(
      1,
      'order-7',
      { principal_amount: 100, reason: 'duplicate recharge' },
      'self-refund-order-7-refund-intent-7',
    )
    expect(createRefund.mock.calls[1]?.[2]).toBe(createRefund.mock.calls[0]?.[2])
    expect(showError).toHaveBeenCalledWith('refund error')
    expect(showSuccess).toHaveBeenCalledWith('payment.refundQuote.submitted')
  })
})
