import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import GroupTranslationsFields from '../GroupTranslationsFields.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/i18n', () => ({ availableLocales: [{ code: 'en', name: 'English' }, { code: 'zh', name: '中文' }] }))

describe('group translations', () => {
  it('edits only the selected translation without mutating the source object', async () => {
    const original = { en: { name: 'Original', description: 'Description' }, zh: { name: '中文' } }
    const wrapper = mount(GroupTranslationsFields, {
      props: { modelValue: original },
      global: { stubs: { Select: { props: ['modelValue', 'options'], emits: ['update:modelValue'], template: '<select @change="$emit(\'update:modelValue\', $event.target.value)"><option v-for="o in options" :value="o.value">{{ o.label }}</option></select>' } } }
    })
    await wrapper.get('select').setValue('en')
    await wrapper.get('input').setValue('Translated')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([{ en: { name: 'Translated', description: 'Description' }, zh: { name: '中文' } }])
    expect(original.en.name).toBe('Original')
    wrapper.unmount()
  })
})
