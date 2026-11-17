import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CNProviderQuotaCell from '../CNProviderQuotaCell.vue'
import UsageProgressBar from '../UsageProgressBar.vue'
import type { Account } from '@/types'

const { queryQuota } = vi.hoisted(() => ({ queryQuota: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { cnProviders: { queryQuota } } }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

function account(id = 'account-a', percent = 30): Account {
  return {
    id, platform: 'kimi', credentials: { account_mode: 'coding' },
    extra: { kimi_monthly_used_percent: percent, kimi_usage_updated_at: new Date().toISOString() }
  } as Account
}

describe('CNProviderQuotaCell', () => {
  beforeEach(() => { queryQuota.mockReset() })

  it('updates monthly quota from refreshed account snapshots', async () => {
    const wrapper = mount(CNProviderQuotaCell, { props: { account: account() } })
    expect(wrapper.text()).toContain('30%')
    expect(queryQuota).not.toHaveBeenCalled()
    await wrapper.setProps({ account: account('account-a', 70) })
    expect(wrapper.text()).toContain('70%')
    wrapper.unmount()
  })

  it('keeps a successful probe during unrelated list refreshes', async () => {
    const initial = account()
    queryQuota.mockResolvedValue({ success: true, tiers: [{ window: 'monthly', used_percent: 82 }] })
    const wrapper = mount(CNProviderQuotaCell, { props: { account: initial } })
    await wrapper.get('button').trigger('click')
    await flushPromises()
    await wrapper.setProps({ account: { ...initial, name: 'renamed', extra: { ...initial.extra } } })
    expect(wrapper.text()).toContain('82%')
    await wrapper.setProps({ account: account('account-a', 90) })
    expect(wrapper.text()).toContain('90%')
    wrapper.unmount()
  })

  it('discards an in-flight probe after switching accounts', async () => {
    let resolve!: (value: unknown) => void
    queryQuota.mockImplementation(() => new Promise(r => { resolve = r }))
    const wrapper = mount(CNProviderQuotaCell, { props: { account: account() } })
    await wrapper.get('button').trigger('click')
    await wrapper.setProps({ account: account('account-b', 45) })
    resolve({ success: true, tiers: [{ window: 'monthly', used_percent: 99 }] })
    await flushPromises()
    expect(wrapper.text()).toContain('45%')
    expect(wrapper.text()).not.toContain('99%')
    wrapper.unmount()
  })

  it('does not display unknown utilization as zero percent', () => {
    const wrapper = mount(UsageProgressBar, { props: { label: '5h', utilization: 0, utilizationUnknown: true, color: 'indigo' } })
    expect(wrapper.text()).not.toContain('0%')
    wrapper.unmount()
  })
})
