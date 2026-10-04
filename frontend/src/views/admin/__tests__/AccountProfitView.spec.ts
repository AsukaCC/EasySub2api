import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AccountProfitView from '../AccountProfitView.vue'
import { formatPoints } from '@/utils/format'

const { listProfit, listSubscriptionTiers } = vi.hoisted(() => ({ listProfit: vi.fn(), listSubscriptionTiers: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { accounts: { listProfit, listSubscriptionTiers } } }))
vi.mock('vue-i18n', async importOriginal => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key })
}))

const period = { revenue_points: 123, cost_usd: 30, profit_points: 98, requests: 5, tokens: 400, subscription_cost_points: 25 }
const fixture = {
  id: 'account-one', name: 'No expiry account', platform: 'openai', status: 'active', subscription_tier: '',
  created_at: '2026-01-01T00:00:00Z', subscription_cost_points: 25, quota_7d: { known: false },
  period_7d: { ...period, revenue_points: 10 }, period_30d: period, lifetime: { ...period, revenue_points: 500 }
}
const mounts: ReturnType<typeof mount>[] = []
function render() {
  const wrapper = mount(AccountProfitView, { global: { stubs: {
    AppLayout: { template: '<div><slot /></div>' },
    TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>' },
    AccountProfitModal: true, BaseDialog: true, Pagination: true, Icon: true, Select: true, SearchInput: true
  } } })
  mounts.push(wrapper)
  return wrapper
}

describe('Account profit rolling 30 days', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listSubscriptionTiers.mockResolvedValue([])
    listProfit.mockResolvedValue({ items: [fixture], total: 1, page: 1, page_size: 20 })
  })
  afterEach(() => { mounts.splice(0).forEach(wrapper => wrapper.unmount()) })

  it('shows revenue and subscription cost even when the account has no expiry', async () => {
    const wrapper = render()
    await flushPromises()
    expect(wrapper.text()).toContain('admin.accounts.profit.period30d')
    expect(wrapper.text()).not.toContain('admin.accounts.profit.expiry30d')
    const cells = wrapper.findAll('tbody .is-group-recent')
    expect(cells[0].text()).toBe(formatPoints(123))
    expect(cells[2].text()).toBe(formatPoints(25))
  })

  it('sorts using the rolling-period API field in both directions', async () => {
    const wrapper = render()
    await flushPromises()
    await wrapper.get('thead .is-group-recent button').trigger('click')
    await flushPromises()
    expect(listProfit).toHaveBeenLastCalledWith(1, 20, expect.objectContaining({ sort_by: 'period_30d_revenue', sort_order: 'desc' }))
    await wrapper.get('thead .is-group-recent button').trigger('click')
    await flushPromises()
    expect(listProfit).toHaveBeenLastCalledWith(1, 20, expect.objectContaining({ sort_by: 'period_30d_revenue', sort_order: 'asc' }))
  })

  it('renders zero usage as zero rather than an unavailable period', async () => {
    listProfit.mockResolvedValue({ items: [{ ...fixture, period_30d: { revenue_points: 0, cost_usd: 0, profit_points: 0, requests: 0, tokens: 0 } }], total: 1 })
    const wrapper = render()
    await flushPromises()
    expect(wrapper.findAll('tbody .is-group-recent')[0].text()).toBe(formatPoints(0))
  })
})
