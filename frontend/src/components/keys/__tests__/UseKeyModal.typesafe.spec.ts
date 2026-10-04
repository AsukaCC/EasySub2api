import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UseKeyModal from '../UseKeyModal.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))

describe('TypeSafe native configuration', () => {
  it('offers native System One examples for each shell without Codex configuration', async () => {
    const wrapper = mount(UseKeyModal, {
      props: { show: true, platform: 'typesafe', apiKey: 'synthetic-key', baseUrl: 'https://example.test/v1/' },
      global: { stubs: { BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' } } }
    })
    const code = () => wrapper.findAll('pre code').map(block => block.text()).join('\n')
    expect(code()).toContain('https://example.test/v1/systemone')
    expect(code()).toContain('"model": "jev-latest"')
    expect(code()).not.toContain('config.toml')
    expect(wrapper.text()).not.toContain('keys.useKeyModal.cliTabs.codexCli')
    await wrapper.findAll('button').find(button => button.text() === 'Windows CMD')!.trigger('click')
    expect(code()).toContain(String.raw`\"model\":\"jev-latest\"`)
    await wrapper.findAll('button').find(button => button.text() === 'PowerShell')!.trigger('click')
    expect(code()).toContain('Invoke-RestMethod')
    expect(code()).toContain('"questions"')
    wrapper.unmount()
  })
})
