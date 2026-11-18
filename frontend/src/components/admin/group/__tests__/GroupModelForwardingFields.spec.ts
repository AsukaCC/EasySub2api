import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import GroupModelForwardingFields from '../GroupModelForwardingFields.vue'
import { isValidModelForwarding } from '@/utils/modelForwarding'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

describe('group model forwarding', () => {
  it('offers group candidates and accepts an upstream model missing from discovery', async () => {
    const wrapper = mount(GroupModelForwardingFields, {
      props: { modelValue: [{ from: 'old-*', to: 'previous' }], candidates: ['available'] },
      global: { stubs: { Icon: true } }
    })
    expect(wrapper.get('datalist option').attributes('value')).toBe('available')
    await wrapper.get('[data-testid="forwarding-target"]').setValue('custom')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([[{ from: 'old-*', to: 'custom' }]])
    await wrapper.get('button[aria-label="common.delete"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[1]).toEqual([[]])
  })

  it('rejects duplicate sources and wildcard targets', () => {
    const rule = { from: 'old-*', to: 'target' }
    expect(isValidModelForwarding([rule])).toBe(true)
    expect(isValidModelForwarding([rule, rule])).toBe(false)
    expect(isValidModelForwarding([{ from: '*', to: 'new-*' }])).toBe(false)
  })
})
