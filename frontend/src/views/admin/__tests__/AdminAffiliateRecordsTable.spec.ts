import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AffiliateRebateRecord } from '@/api/admin/affiliates'
import AdminAffiliateRecordsTable from '../affiliates/AdminAffiliateRecordsTable.vue'

const apiMocks = vi.hoisted(() => ({
  listInviteRecords: vi.fn(),
  listRebateRecords: vi.fn(),
  listTransferRecords: vi.fn(),
  getUserOverview: vi.fn(),
}))

vi.mock('@/api/admin/affiliates', () => ({
  affiliatesAPI: apiMocks,
  default: apiMocks,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn() }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const labels: Record<string, string> = {
    'admin.affiliates.records.rebateAmount': 'Rebate Amount',
    'admin.affiliates.records.reversedPoints': 'Reversed Points',
    'admin.affiliates.records.netRebatePoints': 'Net Rebate Points',
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) => {
        if (key === 'admin.affiliates.records.reservedReversalPoints') {
          return `${params?.amount ?? ''} pending reversal`
        }
        return labels[key] ?? key
      },
    }),
  }
})

const rebate: AffiliateRebateRecord = {
  order_id: 'order-1',
  out_trade_no: 'trade-1',
  inviter_id: 'user-1',
  inviter_email: 'inviter@example.com',
  inviter_username: 'inviter',
  invitee_id: 'user-2',
  invitee_email: 'invitee@example.com',
  invitee_username: 'invitee',
  order_amount: 100,
  pay_amount: 103,
  rebate_amount: 10,
  reserved_reversal_points: 1,
  reversed_points: 4,
  net_rebate_points: 6,
  payment_type: 'alipay',
  order_status: 'PARTIALLY_REFUNDED',
  created_at: '2026-08-28T00:00:00Z',
}

const layoutStub = {
  template: '<div><slot /><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
}

const dataTableStub = {
  props: ['columns', 'data'],
  template: `
    <div>
      <span v-for="column in columns" :key="column.key" class="column-label">{{ column.label }}</span>
      <div v-for="row in data" :key="row.order_id">
        <div data-testid="rebate-points"><slot name="cell-rebate_amount" :row="row" /></div>
        <div data-testid="reversed-points"><slot name="cell-reversed_points" :row="row" /></div>
        <div data-testid="net-rebate-points"><slot name="cell-net_rebate_points" :row="row" /></div>
      </div>
    </div>
  `,
}

describe('AdminAffiliateRecordsTable rebate reversals', () => {
  beforeEach(() => {
    apiMocks.listRebateRecords.mockReset().mockResolvedValue({
      items: [rebate],
      total: 1,
      page: 1,
      page_size: 20,
    })
  })

  it('shows original, reversed, pending, and net rebate point amounts', async () => {
    const wrapper = mount(AdminAffiliateRecordsTable, {
      props: { type: 'rebates' },
      global: {
        stubs: {
          AppLayout: layoutStub,
          TablePageLayout: layoutStub,
          DataTable: dataTableStub,
          Pagination: true,
          BaseDialog: true,
          Icon: true,
          OrderStatusBadge: true,
        },
      },
    })
    await flushPromises()

    const labels = wrapper.findAll('.column-label').map((item) => item.text())
    expect(labels).toContain('Rebate Amount')
    expect(labels).toContain('Reversed Points')
    expect(labels).toContain('Net Rebate Points')
    expect(wrapper.get('[data-testid="rebate-points"]').text()).toContain('10.00')
    expect(wrapper.get('[data-testid="reversed-points"]').text()).toContain('4.00')
    expect(wrapper.get('[data-testid="reversed-points"]').text()).toContain('1.00')
    expect(wrapper.get('[data-testid="reversed-points"]').text()).toContain('pending reversal')
    expect(wrapper.get('[data-testid="net-rebate-points"]').text()).toContain('6.00')
  })
})
