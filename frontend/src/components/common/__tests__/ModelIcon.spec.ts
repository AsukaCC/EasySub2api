import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ModelIcon from '../ModelIcon.vue'
import PlatformIcon from '../PlatformIcon.vue'

describe('model icons', () => {
  it.each(['gemini-2.5-pro', 'google/gemini-3-pro', 'Gemini'])('renders the Gemini brand for %s', model => {
    const wrapper = mount(ModelIcon, { props: { model, size: '24px' } })
    expect(wrapper.getComponent(PlatformIcon).props('platform')).toBe('gemini')
    expect(wrapper.get('svg').attributes('style')).toContain('width: 24px')
  })
  it('recognizes provider-prefixed OpenAI models', () => {
    const native = mount(ModelIcon, { props: { model: 'gpt-5' } })
    const prefixed = mount(ModelIcon, { props: { model: 'openai/gpt-5' } })
    expect(prefixed.get('path').attributes('d')).toBe(native.get('path').attributes('d'))
  })
  it('keeps a visible fallback for unknown or empty model names', () => {
    expect(mount(ModelIcon, { props: { model: 'unknown' } }).text()).toBe('U')
    expect(mount(ModelIcon, { props: { model: '  ' } }).text()).toBe('?')
  })
})
