import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createI18n } from 'vue-i18n'
import UserDashboardDynamicRateOffer from '../UserDashboardDynamicRateOffer.vue'
import { getDynamicRateOffers } from '@/api/userLevel'
import type { DynamicRateOffer } from '@/types'
import zh from '@/i18n/locales/zh/dashboard'
import en from '@/i18n/locales/en/dashboard'

vi.mock('@/api/userLevel', () => ({ getDynamicRateOffers: vi.fn() }))
vi.mock('@/utils/format', () => ({ formatDateTimeToMinute: (value: string) => `fmt:${value}` }))

const mockedGetOffers = vi.mocked(getDynamicRateOffers)
function sampleOffer(overrides: Partial<DynamicRateOffer> = {}): DynamicRateOffer {
  return {
    group_id: 'group-1', group_name: 'VIP', rule_id: 'rule-1', rule_name: 'weekend',
    discount_coefficient: 0.5,
    start_at: '2026-09-15T00:00:00Z', end_at: '2026-09-16T00:00:00Z',
    status: 'participating', activation_spend: 100, usage_7d: 150,
    personal_quota_amount: 0, personal_used_amount: 0,
    ...overrides,
  }
}
function mountOffer(locale = 'zh') {
  const i18n = createI18n({
    legacy: false, locale,
    messages: { zh, en },
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
    expect(wrapper.text()).toContain('参与中')
    expect(wrapper.findAll('.dashboard-dynamic-rate-offer__participation').map(item => item.text())).toEqual(['参与中', '参与中'])
    expect(wrapper.text()).not.toContain('特惠优先扣充值积分')
    expect(wrapper.text()).not.toContain('赠送积分按分组倍率 × 用户等级倍率扣除')
    wrapper.unmount()
  })
  it('renders percent saved in English', async () => {
    mockedGetOffers.mockResolvedValue([sampleOffer({ discount_coefficient: 0.85 })])
    const wrapper = mountOffer('en')
    await flushPromises()
    expect(wrapper.text()).toContain('15% off')
    expect(wrapper.text()).toContain('Expires')
    expect(wrapper.text()).toContain('Participating')
    expect(wrapper.get('.dashboard-dynamic-rate-offer__participation').text()).toBe('Participating')
    expect(wrapper.text()).not.toContain('Offers use recharge points first')
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
  it.each([
    ['zh', '未达门槛', '近 7 天消费 25 / 100 积分'],
    ['en', 'Requirements not met', '7-day spend: 25 / 100 credits'],
  ])('shows a compact spending threshold in %s', async (locale, status, progress) => {
    mockedGetOffers.mockResolvedValue([sampleOffer({ status: 'below_threshold', usage_7d: 25 })])
    const wrapper = mountOffer(locale)
    await flushPromises()
    expect(wrapper.findAll('.dashboard-dynamic-rate-offer__participation > span').map(item => item.text())).toEqual([status, progress])
    wrapper.unmount()
  })
  it.each([
    ['quota_exhausted', '优惠额度已用尽'],
    ['group_unavailable', '需分组权限'],
    ['subscription_required', '需有效订阅'],
    ['subscription_limited', '订阅额度不足'],
    ['level_required', '需配置消费等级'],
  ] as const)('shows the participation restriction for %s', async (status, message) => {
    mockedGetOffers.mockResolvedValue([sampleOffer({ status, personal_quota_amount: 10, personal_used_amount: 10 })])
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.text()).toContain(message)
    if (status === 'quota_exhausted') expect(wrapper.text()).toContain('优惠额度 10 / 10 U')
    else expect(wrapper.text()).not.toContain('10 / 10 U')
    expect(wrapper.text()).not.toContain('近 7 天消费')
    expect(wrapper.text()).not.toContain('参与中')
    wrapper.unmount()
  })
  it('refreshes eligibility when the user reaches the threshold', async () => {
    mockedGetOffers.mockResolvedValueOnce([sampleOffer({ status: 'below_threshold', usage_7d: 99 })])
      .mockResolvedValueOnce([sampleOffer({ usage_7d: 100 })])
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.text()).toContain('未达门槛')
    await vi.advanceTimersByTimeAsync(60_000)
    await flushPromises()
    expect(wrapper.text()).toContain('参与中')
    expect(wrapper.text()).not.toContain('未达门槛')
    expect(wrapper.get('.dashboard-dynamic-rate-offer__participation').text()).toBe('参与中')
    wrapper.unmount()
  })
  it('omits redundant conditions while participating with no spending minimum', async () => {
    mockedGetOffers.mockResolvedValue([sampleOffer({ activation_spend: 0, personal_quota_amount: 10, personal_used_amount: 2 })])
    const wrapper = mountOffer()
    await flushPromises()
    expect(wrapper.text()).not.toContain('无消费门槛')
    expect(wrapper.get('.dashboard-dynamic-rate-offer__participation').text()).toBe('参与中')
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
