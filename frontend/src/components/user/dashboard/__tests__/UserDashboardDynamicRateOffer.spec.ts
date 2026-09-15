import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UserDashboardDynamicRateOffer from '../UserDashboardDynamicRateOffer.vue'
import { getDynamicRateOffers } from '@/api/userLevel'
import type { DynamicRateOffer } from '@/types'

vi.mock('@/api/userLevel', () => ({
  getDynamicRateOffers: vi.fn(),
}))

vi.mock('@/utils/format', () => ({
  formatDateTimeToMinute: (value: string) => `fmt:${value}`,
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const mockedGetOffers = vi.mocked(getDynamicRateOffers)

function sampleOffer(overrides: Partial<DynamicRateOffer> = {}): DynamicRateOffer {
  return {
    group_id: 'group-1',
    group_name: 'VIP',
    rule_id: 'rule-1',
    rule_name: 'weekend',
    start_at: '2026-09-15T00:00:00Z',
    end_at: '2026-09-16T00:00:00Z',
    ...overrides,
  }
}

describe('UserDashboardDynamicRateOffer', () => {
  it('hides the card when there is no active offer', async () => {
    mockedGetOffers.mockResolvedValueOnce([])

    const wrapper = mount(UserDashboardDynamicRateOffer, {
      global: { stubs: { Icon: true } },
    })
    await flushPromises()

    expect(wrapper.find('section').exists()).toBe(false)
    expect(wrapper.text()).toBe('')
    wrapper.unmount()
  })

  it('renders group name and formatted start/end times', async () => {
    mockedGetOffers.mockResolvedValueOnce([sampleOffer()])

    const wrapper = mount(UserDashboardDynamicRateOffer, {
      global: { stubs: { Icon: true } },
    })
    await flushPromises()

    expect(wrapper.find('section').exists()).toBe(true)
    expect(wrapper.text()).toContain('VIP')
    expect(wrapper.text()).toContain('dashboard.dynamicRateOffer.startsAt')
    expect(wrapper.text()).toContain('dashboard.dynamicRateOffer.endsAt')
    expect(wrapper.text()).toContain('fmt:2026-09-15T00:00:00Z')
    expect(wrapper.text()).toContain('fmt:2026-09-16T00:00:00Z')
    wrapper.unmount()
  })
})
