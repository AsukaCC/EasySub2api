import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { createI18n } from 'vue-i18n'
import DateTimePicker from '../DateTimePicker.vue'
import zh from '@/i18n/locales/zh/dashboard'

const wrappers: ReturnType<typeof mount>[] = []
function setup(props: Record<string, unknown> = {}) {
  const wrapper = mount(DateTimePicker, { props, attachTo: document.body,
    global: { plugins: [createI18n({ legacy: false, locale: 'zh', messages: { zh } })] } })
  wrappers.push(wrapper)
  return wrapper
}
afterEach(() => { wrappers.splice(0).forEach(wrapper => wrapper.unmount()) })

describe('DateTimePicker', () => {
  it('preserves local datetime values and seconds', async () => {
    const wrapper = setup({ modelValue: '2026-09-20T10:15:32', step: 1 })
    expect(wrapper.get('input').element.value).toBe('2026-09-20 10:15:32')
    await wrapper.get('input').setValue('2026-09-21 11:16:33')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['2026-09-21T11:16:33'])
    expect(wrapper.emitted('change')?.at(-1)).toEqual(['2026-09-21T11:16:33'])
  })
  it('rejects invalid dates and out-of-range values instead of retaining an old selection', async () => {
    const wrapper = setup({ type: 'date', modelValue: '2026-09-20', min: '2026-09-01', max: '2026-09-30' })
    await wrapper.get('input').setValue('2026-02-31')
    expect(wrapper.get('input').attributes('aria-invalid')).toBe('true')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([''])
    await wrapper.setProps({ modelValue: '' })
    expect(wrapper.get('input').element.value).toBe('2026-02-31')
    await wrapper.get('input').setValue('2026-10-01')
    expect(wrapper.get('input').element.checkValidity()).toBe(false)
    await wrapper.get('input').setValue('2026-09-21')
    expect(wrapper.get('input').element.checkValidity()).toBe(true)
  })
  it('selects a calendar date, updates the parent, and restores focus', async () => {
    const wrapper = setup({ type: 'date', modelValue: '2026-09-20' })
    await wrapper.get('button[aria-haspopup]').trigger('click')
    const button = document.querySelector<HTMLButtonElement>('[data-date="2026-09-21"]')!
    button.click()
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('change')?.at(-1)).toEqual(['2026-09-21'])
    expect(document.querySelector('[data-date-picker-panel]')).toBeNull()
    expect(document.activeElement).toBe(wrapper.get('button[aria-haspopup]').element)
  })
  it('supports arrow navigation and Escape without changing the value', async () => {
    const wrapper = setup({ type: 'date', modelValue: '2026-09-20' })
    await wrapper.get('button[aria-haspopup]').trigger('click')
    const button = document.querySelector<HTMLButtonElement>('[data-date="2026-09-20"]')!
    button.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(document.activeElement?.getAttribute('data-date')).toBe('2026-09-21')
    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(document.querySelector('[data-date-picker-panel]')).toBeNull()
    expect(wrapper.emitted('change')).toBeUndefined()
  })
  it('respects disabled state and allows clearing optional fields', async () => {
    const wrapper = setup({ type: 'date', modelValue: '2026-09-20', disabled: true })
    await wrapper.get('button').trigger('click')
    expect(document.querySelector('[data-date-picker-panel]')).toBeNull()
    await wrapper.setProps({ disabled: false })
    await wrapper.get('button[aria-label="清空"]').trigger('click')
    expect(wrapper.emitted('change')?.at(-1)).toEqual([''])
  })
  it('normalizes time-only midnight and enforces step precision', async () => {
    const wrapper = setup({ type: 'time', step: 1 })
    await wrapper.get('input').setValue('24：00：00')
    expect(wrapper.emitted('change')?.at(-1)).toEqual(['00:00:00'])
    await wrapper.setProps({ step: 900 })
    await wrapper.get('input').setValue('10:16')
    expect(wrapper.get('input').element.checkValidity()).toBe(false)
    await wrapper.get('input').setValue('10:15')
    expect(wrapper.get('input').element.checkValidity()).toBe(true)
  })
})
