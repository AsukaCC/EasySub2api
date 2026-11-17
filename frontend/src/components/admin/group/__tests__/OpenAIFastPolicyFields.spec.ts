import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import OpenAIFastPolicyFields from '../OpenAIFastPolicyFields.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

describe('OpenAIFastPolicyFields', () => {
  it('emits independent force, ultrafast and billing choices', async () => {
    const wrapper = mount(OpenAIFastPolicyFields, {
      props: { idPrefix: 'group', force: false, free: false },
      global: { stubs: { Toggle: { props: ['modelValue'], emits: ['update:modelValue'], template: '<button @click="$emit(\'update:modelValue\', !modelValue)" />' } } }
    })
    await wrapper.get('#group-ultrafast').trigger('click')
    expect(wrapper.emitted('update:ultrafast')).toEqual([[true]])
    expect(wrapper.emitted('update:force')).toBeUndefined()
    expect(wrapper.emitted('update:free')).toBeUndefined()
    wrapper.unmount()
  })
})
