import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import LoadingButtonContent from '../LoadingButtonContent.vue'
import LoadingSpinner from '../LoadingSpinner.vue'
import LoadingState from '../LoadingState.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

describe('LoadingSpinner', () => {
  it('supports xs size and inherited color', () => {
    const wrapper = mount(LoadingSpinner, {
      props: { size: 'xs', color: 'inherit' }
    })

    expect(wrapper.classes()).toContain('loading-spinner--xs')
    expect(wrapper.classes()).toContain('loading-spinner--inherit')
  })

  it('announces standalone loading state', () => {
    const wrapper = mount(LoadingSpinner)

    expect(wrapper.attributes('role')).toBe('status')
    expect(wrapper.attributes('aria-label')).toBe('common.loading')
    expect(wrapper.text()).toContain('common.loading')
  })

  it('is silent when decorative', () => {
    const wrapper = mount(LoadingSpinner, { props: { decorative: true } })

    expect(wrapper.attributes('role')).toBeUndefined()
    expect(wrapper.attributes('aria-label')).toBeUndefined()
    expect(wrapper.attributes('aria-hidden')).toBe('true')
  })
})

describe('LoadingState', () => {
  it.each(['page', 'section', 'inline'] as const)('renders the %s layout', (variant) => {
    const wrapper = mount(LoadingState, {
      props: { variant, label: 'Working' }
    })

    expect(wrapper.classes()).toContain(`loading-state--${variant}`)
    expect(wrapper.attributes('role')).toBe('status')
    expect(wrapper.attributes('aria-live')).toBe('polite')
    expect(wrapper.attributes('aria-busy')).toBe('true')
    expect(wrapper.text()).toContain('Working')
    expect(wrapper.getComponent(LoadingSpinner).props('decorative')).toBe(true)
  })
})

describe('LoadingButtonContent', () => {
  it('keeps both layers mounted to preserve button width', async () => {
    const wrapper = mount(LoadingButtonContent, {
      props: { loading: false, loadingText: 'Saving' },
      slots: { default: 'Save settings' }
    })

    const layers = wrapper.findAll('.loading-button-content__layer')
    expect(layers).toHaveLength(2)
    expect(wrapper.get('.loading-button-content__normal').text()).toBe('Save settings')
    expect(wrapper.get('.loading-button-content__loading').classes()).toContain('loading-button-content__layer--hidden')

    await wrapper.setProps({ loading: true })

    expect(wrapper.get('.loading-button-content__normal').classes()).toContain('loading-button-content__layer--hidden')
    expect(wrapper.get('.loading-button-content__loading').text()).toContain('Saving')
    expect(wrapper.getComponent(LoadingSpinner).props()).toMatchObject({
      size: 'xs',
      color: 'inherit',
      decorative: true
    })
  })
})
