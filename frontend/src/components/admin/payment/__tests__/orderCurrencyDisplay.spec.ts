import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import type { PaymentOrder } from '@/types/payment'
import AdminOrderDetail from '../AdminOrderDetail.vue'
import AdminOrderTable from '../AdminOrderTable.vue'
import AdminRefundDialog from '../AdminRefundDialog.vue'
import OrderTable from '@/components/payment/OrderTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'en' },
    }),
  }
})

const BaseDialogStub = {
  props: ['show'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
}

const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.id">
        <slot name="cell-pay_amount" :value="row.pay_amount" :row="row" />
      </div>
    </div>
  `,
}

function orderFactory(overrides: Partial<PaymentOrder> = {}): PaymentOrder {
  return {
    id: '1',
    user_id: '10',
    amount: 120,
    pay_amount: 108,
    principal_amount: 100,
    fee_amount: 8,
    base_points: 100,
    bonus_points: 20,
    credited_points: 120,
    affiliate_rebate_points: 0,
    wallet_amount: 0,
    wallet_bonus_amount: 0,
    wallet_recharge_amount: 0,
    gateway_base_amount: 100,
    wallet_only: false,
    currency: 'CNY',
    fee_rate: 8,
    payment_type: 'stripe',
    out_trade_no: 'sub2_202606250001',
    status: 'COMPLETED',
    order_type: 'balance',
    created_at: '2026-06-25T10:00:00Z',
    expires_at: '2026-06-25T10:30:00Z',
    refund_amount: 25,
    ...overrides,
  }
}

describe('payment order unit display', () => {
  it('separates recharge CNY amounts, credited points, and refund snapshots', () => {
    const wrapper = mount(AdminOrderDetail, {
      props: {
        show: true,
        order: orderFactory({
          refunded_gateway_amount: 25,
          reversed_base_points: 20,
          reversed_bonus_points: 5,
        }),
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('¥100.00')
    expect(text).toContain('¥8.00')
    expect(text).toContain('¥108.00')
    expect(text).toContain('100.00 points')
    expect(text).toContain('20.00 points')
    expect(text).toContain('120.00 points')
    expect(text).toContain('¥25.00')
    expect(text).toContain('25.00 points')
  })

  it('uses CNY for refund principal and points for account reversal', () => {
    const wrapper = mount(AdminRefundDialog, {
      props: {
        show: true,
        order: orderFactory({
          status: 'PARTIALLY_REFUNDED',
          refund_amount: 20,
          refunded_principal_amount: 20,
        }),
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('¥108.00')
    expect(text).toContain('120.00 points')
    expect(text).toContain('¥20.00')
    expect(text).toContain('¥80.00')
    expect(text).toContain('96.00 points')
  })

  it('does not reinterpret a legacy CNY refund amount as reversed points', () => {
    const wrapper = mount(AdminOrderDetail, {
      props: {
        show: true,
        order: orderFactory({
          refunded_gateway_amount: 25,
          reversed_base_points: 0,
          reversed_bonus_points: 0,
        }),
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    expect(wrapper.text()).toContain('¥25.00')
    expect(wrapper.text()).not.toContain('25.00 points')
  })

  it('renders recharge money and subscription charges with their respective units', () => {
    const wrapper = mount(OrderTable, {
      props: {
        orders: [
          orderFactory({ id: '1' }),
          orderFactory({
            id: '2',
            order_type: 'subscription',
            amount: 40,
            wallet_amount: 40,
            pay_amount: 0,
          }),
        ],
        loading: false,
        showUser: true,
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          OrderStatusBadge: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('¥108.00')
    expect(text).toContain('120.00 points')
    expect(text).toContain('40.00 points')
    expect(text).not.toContain('$')
  })

  it('uses the same split in the admin order table', () => {
    const wrapper = mount(AdminOrderTable, {
      props: {
        orders: [
          orderFactory({ id: '1' }),
          orderFactory({
            id: '2',
            order_type: 'subscription',
            amount: 40,
            wallet_amount: 40,
            pay_amount: 0,
          }),
        ],
        loading: false,
        page: 1,
        pageSize: 20,
        total: 2,
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          Icon: true,
          Pagination: true,
          Select: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('¥108.00')
    expect(text).toContain('120.00 points')
    expect(text).toContain('40.00 points')
    expect(text).not.toContain('$')
  })
})
