import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { AdminUser } from '@/types'
import UserBalanceModal from '../UserBalanceModal.vue'

const mocks = vi.hoisted(() => ({ getConfig: vi.fn(), updateBalance: vi.fn(), showError: vi.fn() }))
vi.mock('@/api/admin/payment', () => ({ adminPaymentAPI: { getConfig: mocks.getConfig } }))
vi.mock('@/api/admin', () => ({ adminAPI: { users: { updateBalance: mocks.updateBalance } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: mocks.showError, showSuccess: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/utils/format', () => ({ formatPoints: (value: number) => `${value.toFixed(2)} points` }))

const config = { data: { recharge_bonus_tiers: [
  { threshold_cny: 50, bonus_points: 3 },
  { threshold_cny: 200, bonus_points: 20 },
] } }

function mountModal(operation: 'add' | 'subtract' = 'add') {
  return mount(UserBalanceModal, {
    props: { show: true, operation, user: { id: 'user-1', email: 'user@example.com', balance: 10 } as AdminUser },
    global: { stubs: {
      BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
      Icon: true,
      Select: {
        props: ['modelValue', 'options'], emits: ['update:modelValue'],
        template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"><option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option></select>',
      },
    } },
  })
}

beforeEach(() => {
  vi.resetAllMocks()
  mocks.getConfig.mockResolvedValue(config)
  mocks.updateBalance.mockResolvedValue({})
})

describe('admin recharge bonus', () => {
  it('previews only the highest tier and submits the principal without adding the bonus twice', async () => {
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.get('#balance-amount').setValue(49.99)
    expect(wrapper.text()).toContain('0.00 points')
    expect(wrapper.text()).toContain('59.99 points')
    await wrapper.get('#balance-amount').setValue(50)
    expect(wrapper.text()).toContain('3.00 points')
    expect(wrapper.text()).toContain('63.00 points')
    await wrapper.get('#balance-amount').setValue(200)
    expect(wrapper.text()).toContain('20.00 points')
    expect(wrapper.text()).toContain('230.00 points')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.updateBalance).toHaveBeenCalledWith('user-1', 200, 'add', '', 'recharge', 90)
    wrapper.unmount()
  })

  it('does not grant an additional tier bonus for manual bonus points or withdrawals', async () => {
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.get('#balance-amount').setValue(200)
    await wrapper.get('select').setValue('bonus')
    expect(wrapper.text()).not.toContain('admin.users.rechargeTierBonus')
    expect(wrapper.text()).toContain('210.00 points')
    await wrapper.setProps({ operation: 'subtract' })
    await wrapper.get('#balance-amount').setValue(5)
    expect(wrapper.text()).not.toContain('admin.users.rechargeTierBonus')
    expect(wrapper.text()).toContain('5.00 points')
    expect(mocks.getConfig).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('blocks recharge while loading or after a failure and supports retry', async () => {
    let reject!: (error: Error) => void
    mocks.getConfig.mockReturnValueOnce(new Promise((_, fail) => { reject = fail }))
    const wrapper = mountModal()
    await wrapper.get('#balance-amount').setValue(50)
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await wrapper.get('form').trigger('submit')
    expect(mocks.updateBalance).not.toHaveBeenCalled()
    reject(new Error('offline'))
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('admin.users.rechargeBonusLoadFailed')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await wrapper.get('[role="alert"] button').trigger('click')
    await flushPromises()
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.text()).toContain('63.00 points')
    wrapper.unmount()
  })

  it('ignores an outdated config response after reopening', async () => {
    let resolve!: (value: typeof config) => void
    mocks.getConfig.mockReturnValueOnce(new Promise((done) => { resolve = done }))
    const wrapper = mountModal()
    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await flushPromises()
    resolve({ data: { recharge_bonus_tiers: [] } })
    await flushPromises()
    await wrapper.get('#balance-amount').setValue(50)
    expect(wrapper.text()).toContain('63.00 points')
    wrapper.unmount()
  })
})
