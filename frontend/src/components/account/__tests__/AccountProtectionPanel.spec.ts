import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AccountProtectionPanel from '../AccountProtectionPanel.vue'
import Select from '@/components/common/Select.vue'
import type { Account } from '@/types'

const api = vi.hoisted(() => ({ strategies: vi.fn(), preview: vi.fn(), apply: vi.fn(), set: vi.fn(), integrity: vi.fn() }))
vi.mock('@/api/admin/accountProtection', () => ({ accountProtection: api }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const account = { id: 'a', platform: 'openai', type: 'oauth', anti_degradation: true, protection_mode: 'legacy', extra: {} } as Account
const options = {
  global: { stubs: {
    Icon: true,
    ConfirmDialog: { props: ['show'], emits: ['confirm', 'cancel'], template: '<div v-if="show" data-testid="confirm"><button type="button" @click="$emit(\'confirm\')">confirm</button><slot /></div>' }
  } }
}
beforeEach(() => {
  vi.clearAllMocks()
  api.strategies.mockResolvedValue([{ id: 'legacy', apply_supported: true, diagnostic_only: false, requires_openai: false }])
  api.preview.mockResolvedValue({ enabled: true, eligible: true, changes: [], runtime: { integrity_mode: 'observe', configured_tls: 'nodejs24', effective_tls: 'standard', concurrency: 16, observed: false } })
  api.set.mockResolvedValue({ ...account, anti_degradation: false })
  api.apply.mockResolvedValue(account)
  api.integrity.mockResolvedValue({ ...account, extra: { request_integrity_mode: 'enforce' } })
})

describe('AccountProtectionPanel', () => {
  it('requires confirmation before disabling protection', async () => {
    const wrapper = mount(AccountProtectionPanel, { ...options, props: { account, compact: true } })
    await wrapper.get('input[type=checkbox]').setValue(false)
    expect(api.set).not.toHaveBeenCalled()
    await wrapper.get('[data-testid=confirm] button').trigger('click')
    await flushPromises()
    expect(api.set).toHaveBeenCalledWith('a', false)
    expect(wrapper.emitted('updated')?.[0]?.[0]).toMatchObject({ anti_degradation: false })
  })

  it('shows configured and effective TLS separately and previews before apply', async () => {
    const wrapper = mount(AccountProtectionPanel, { ...options, props: { account } })
    await flushPromises()
    expect(wrapper.text()).toContain('nodejs24')
    expect(wrapper.text()).toContain('standard')
    await wrapper.get('button.btn-secondary').trigger('click')
    await flushPromises()
    expect(api.apply).not.toHaveBeenCalled()
    await wrapper.get('[data-testid=confirm] button').trigger('click')
    await flushPromises()
    expect(api.apply).toHaveBeenCalledWith('a', 'legacy')
  })

  it('saves integrity independently without submitting an account form', async () => {
    const wrapper = mount(AccountProtectionPanel, { ...options, props: { account } })
    await flushPromises()
    await wrapper.findAllComponents(Select)[1]!.setValue('enforce')
    await flushPromises()
    expect(api.integrity).toHaveBeenCalledWith('a', 'enforce')
    expect(wrapper.emitted('updated')).toHaveLength(1)
  })

  it('keeps an error visible and allows retry after a failed change', async () => {
    api.set.mockRejectedValueOnce(new Error('conflict'))
    const wrapper = mount(AccountProtectionPanel, { ...options, props: { account, compact: true } })
    await wrapper.get('input[type=checkbox]').setValue(false)
    await wrapper.get('[data-testid=confirm] button').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role=alert]').text()).toBe('conflict')
    expect(wrapper.get('fieldset').attributes('disabled')).toBeUndefined()
    expect(wrapper.emitted('updated')).toBeUndefined()
  })
})
