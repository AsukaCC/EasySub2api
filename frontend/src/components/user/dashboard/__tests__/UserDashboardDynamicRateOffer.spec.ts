import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import UserDashboardDynamicRateOffer from '../UserDashboardDynamicRateOffer.vue'
import { getDynamicRateOffers } from '@/api/userLevel'
import type { DynamicRateOffer } from '@/types'

vi.mock('@/api/userLevel', () => ({ getDynamicRateOffers: vi.fn() }))
vi.mock('@/utils/format', () => ({ formatDateTimeToMinute: (value: string) => `fmt:${value}` }))

const mockedGetOffers = vi.mocked(getDynamicRateOffers)
function sampleOffer(overrides: Partial<DynamicRateOffer> = {}): DynamicRateOffer {
  return {
    group_id: 'group-1', group_name: 'VIP', rule_id: 'rule-1', rule_name: 'weekend',
    discount_coefficient: 0.5,
    start_at: '2026-09-15T00:00:00Z', end_at: '2026-09-16T00:00:00Z',
    ...overrides,
  }
}
function mountOffer(locale = 'zh') {
  const i18n = createI18n({
    legacy: false, locale,
    messages: {
      zh: { dashboard: { dynamicRateOffer: { title: '分时优惠', groups: '{count} 个分组', discount: '{percent}% off', discountDetail: '优惠系数 ×{coefficient}，减免 {percent}%', endsAt: '到期' } } },
      en: { dashboard: { dynamicRateOffer: { title: 'Timed Discount', groups: '{count} groups', discount: '{percent}% off', discountDetail: 'Multiplier ×{coefficient}, save {percent}%', endsAt: 'Expires' } } },
    },
  })
  return mount(UserDashboardDynamicRateOffer, { global: { plugins: [i18n], stubs: { Icon: true } } })
}
beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date', 'setInterval', 'clearInterval'] })
  vi.setSystemTime(new Date('2026-09-15T12:00:00Z'))
  mockedGetOffers.mockReset().mockResolvedValue([])
})
afterEach(() => { vi.useRealTimers() })

describe('UserDashboardDynamicRateOffer', () => {
  it('hides when there is no active offer', async () => {
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.find('section').exists()).toBe(false)
    wrapper.unmount()
  })
  it('renders every discounted group with discount and expiry, never the start time', async () => {
    mockedGetOffers.mockResolvedValue([
      sampleOffer(),
      sampleOffer({ group_id: 'group-2', group_name: 'Standard', discount_coefficient: 0.85 }),
    ])
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.findAll('li')).toHaveLength(2)
    expect(wrapper.text()).toContain('VIP')
    expect(wrapper.text()).toContain('Standard')
    expect(wrapper.text()).toContain('50% off')
    expect(wrapper.text()).toContain('15% off')
    expect(wrapper.text()).toContain('到期')
    expect(wrapper.text()).toContain('fmt:2026-09-16T00:00:00Z')
    expect(wrapper.text()).not.toContain('2026-09-15T00:00:00Z')
    expect(wrapper.findAll('time')).toHaveLength(2)
    expect(wrapper.text()).toContain('2 个分组')
    wrapper.unmount()
  })
  it('renders percent saved in English', async () => {
    mockedGetOffers.mockResolvedValue([sampleOffer({ discount_coefficient: 0.85 })])
    const wrapper = mountOffer('en')
    await flushPromises()
    expect(wrapper.text()).toContain('15% off')
    expect(wrapper.text()).toContain('Expires')
    wrapper.unmount()
  })
  it('excludes future, expired, invalid and non-discounted offers', async () => {
    mockedGetOffers.mockResolvedValue([
      sampleOffer({ start_at: '2026-09-15T13:00:00Z' }),
      sampleOffer({ end_at: '2026-09-15T12:00:00Z' }),
      sampleOffer({ discount_coefficient: 1 }),
      sampleOffer({ discount_coefficient: 0 }),
      sampleOffer({ discount_coefficient: Number.NaN }),
      sampleOffer({ end_at: 'invalid' }),
    ])
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.find('section').exists()).toBe(false)
    wrapper.unmount()
  })
  it('removes an offer at its expiry without waiting for the next request', async () => {
    mockedGetOffers.mockResolvedValue([sampleOffer({ end_at: '2026-09-15T12:00:01Z' })])
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.find('section').exists()).toBe(true)
    await vi.advanceTimersByTimeAsync(1000)
    expect(wrapper.find('section').exists()).toBe(false)
    wrapper.unmount()
  })
  it('refreshes active discounts and releases timers on unmount', async () => {
    mockedGetOffers.mockResolvedValueOnce([]).mockResolvedValueOnce([sampleOffer()])
    const wrapper = mountOffer()
    await flushPromises()
    await vi.advanceTimersByTimeAsync(60_000)
    await flushPromises()
    expect(wrapper.text()).toContain('50% off')
    expect(mockedGetOffers).toHaveBeenCalledTimes(2)
    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })
  it('hides unavailable offers on a failed response', async () => {
    mockedGetOffers.mockRejectedValueOnce(new Error('unavailable'))
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.find('section').exists()).toBe(false)
    wrapper.unmount()
  })
})
