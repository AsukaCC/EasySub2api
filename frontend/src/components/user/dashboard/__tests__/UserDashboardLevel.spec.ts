import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import UserDashboardLevel from '../UserDashboardLevel.vue'
import type { UserLevelDashboard } from '@/api/userLevel'
import zh from '@/i18n/locales/zh/dashboard'

vi.mock('@/utils/format', () => ({ formatPoints: (value: number) => `${value} points` }))

function profile(): UserLevelDashboard {
  return {
    user_id: 'user', level: 1, configured: true, usage_7d: 20, window_hours: 168,
    window_from: '2026-09-13T00:00:00Z', calculated_at: '2026-09-20T00:00:00Z',
    user_level_multiplier: .75, group_rule_multiplier: .2, effective_multiplier: .075,
    current_tier_ids: ['base'], rules: [{
      rule_id: 'rule', rule_name: 'Default', window_days: 7, enabled: true,
      spend: 20, window_from: '2026-09-13T00:00:00Z', calculated_at: '2026-09-20T00:00:00Z',
      current_tier_id: 'base', current_tier_name: 'Base', current_tier_order: 0,
      min_spend: 0, default_multiplier: .75,
      tiers: [
        { id: 'base', rule_id: 'rule', name: 'Base', sort_order: 0, min_spend: 0, default_multiplier: .75 },
        { id: 'gold', rule_id: 'rule', name: 'Gold', sort_order: 1, min_spend: 100, default_multiplier: .5 },
      ],
    }],
  }
}
const wrappers: ReturnType<typeof mount>[] = []
function render(value: UserLevelDashboard | null = profile(), loading = false) {
  const wrapper = mount(UserDashboardLevel, {
    attachTo: document.body,
    props: { profile: value, loading },
    global: {
      plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh } })],
      stubs: { Icon: true, LoadingState: true, BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' } },
    },
  })
  wrappers.push(wrapper)
  return wrapper
}
afterEach(() => { for (const wrapper of wrappers.splice(0)) wrapper.unmount() })

describe('UserDashboardLevel current multiplier', () => {
  it('shows the current user level rate, not the next tier or combined group rate', () => {
    const wrapper = render()
    const rate = wrapper.get('.dashboard-level__current-rate')
    expect(rate.text()).toContain('当前用户等级消费倍率')
    expect(rate.text()).toContain('×0.75')
    expect(rate.text()).not.toContain('×0.50')
    expect(rate.text()).not.toContain('×0.075')
  })
  it('shows the requested formula on hover and keyboard focus', async () => {
    const wrapper = render()
    const rate = wrapper.get('.dashboard-level__current-rate')
    const tooltip = document.body.querySelector<HTMLElement>('[role="tooltip"]')!
    expect(tooltip.style.display).toBe('none')
    await rate.trigger('mouseenter')
    expect(tooltip.style.display).not.toBe('none')
    expect(tooltip.textContent).toContain('实际消费倍率 = 分组倍率 * 用户等级倍率')
    await rate.trigger('mouseleave')
    expect(tooltip.style.display).toBe('none')
    await rate.trigger('focusin')
    expect(tooltip.style.display).not.toBe('none')
  })
  it('keeps the rate visible at the highest tier and preserves meaningful precision', () => {
    const value = profile()
    value.user_level_multiplier = .875
    value.rules[0].tiers = value.rules[0].tiers!.slice(0, 1)
    expect(render(value).get('.dashboard-level__current-rate-value').text()).toBe('×0.875')
  })
  it('uses the current tier then the neutral default when the aggregate is absent', async () => {
    const value = profile()
    value.user_level_multiplier = null
    const wrapper = render(value)
    expect(wrapper.get('.dashboard-level__current-rate-value').text()).toBe('×0.75')
    await wrapper.setProps({ profile: { ...value, rules: [{ ...value.rules[0], default_multiplier: null }] } })
    expect(wrapper.get('.dashboard-level__current-rate-value').text()).toBe('×1.00')
  })
  it('does not invent a current multiplier while loading or without a profile', () => {
    expect(render(null).find('.dashboard-level__current-rate').exists()).toBe(false)
    expect(render(profile(), true).find('.dashboard-level__current-rate').exists()).toBe(false)
  })
})
