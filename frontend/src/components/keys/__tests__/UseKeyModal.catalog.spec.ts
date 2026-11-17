import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import UseKeyModal from '../UseKeyModal.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
afterEach(() => vi.unstubAllGlobals())

describe('Codex catalog configuration', () => {
  it('uses the group WebSocket default when opening a previously hidden modal', async () => {
    const wrapper = mount(UseKeyModal, {
      props: { show: false, apiKey: '', baseUrl: '', platform: null },
      global: { stubs: { BaseDialog: { template: '<div><slot /></div>' }, Icon: true } }
    })
    await wrapper.setProps({ show: true, apiKey: 'synthetic-key', baseUrl: 'https://example.com', platform: 'openai', codexWebsocketEnabled: true })
    await flushPromises()
    expect(wrapper.findAll('pre code').map(block => block.text()).join('\n')).toContain('supports_websockets = true')
    wrapper.unmount()
  })

  it('exports edited models, follows the main model by default, and toggles WebSocket', async () => {
    const wrapper = mount(UseKeyModal, {
      props: { show: true, apiKey: 'synthetic-key', baseUrl: 'https://example.com', platform: 'openai' },
      global: { stubs: { BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }, Icon: true } }
    })
    const code = () => wrapper.findAll('pre code').map(block => block.text()).join('\n')
    expect(code()).not.toContain('default_subagent_model')
    await wrapper.get('[data-testid="codex-main-model"]').setValue('gpt-6-astra')
    expect(code()).toContain('model = "gpt-6-astra"')
    expect(code()).not.toContain('default_subagent_model')
    await wrapper.get('[data-testid="codex-subagent-follow"]').setValue(false)
    await wrapper.get('[data-testid="codex-subagent-model"]').setValue('gpt-6.1-terra')
    expect(code()).toContain('default_subagent_model = "gpt-6.1-terra"')
    await wrapper.get('[data-testid="codex-subagent-model"]').setValue('')
    expect(wrapper.find('[data-testid="codex-subagent-model"]').exists()).toBe(true)
    await wrapper.get('[data-testid="codex-subagent-follow"]').setValue(true)
    expect(code()).not.toContain('default_subagent_model')
    await wrapper.get('[data-testid="codex-websocket"]').setValue(true)
    expect(code()).toContain('supports_websockets = true')
    await wrapper.get('[data-testid="codex-websocket"]').setValue(false)
    expect(code()).toContain('supports_websockets = false')
    wrapper.unmount()
  })

  it('supports remote and fetched local catalogs for HTTP and WebSocket clients', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ models: [{ slug: 'gpt-6.1-sol' }] })))
    vi.stubGlobal('fetch', fetchMock)
    const wrapper = mount(UseKeyModal, {
      props: { show: true, apiKey: 'synthetic-key', baseUrl: 'https://example.com/v1', platform: 'openai' },
      global: { stubs: { BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }, Icon: true } }
    })
    const code = () => wrapper.findAll('pre code').map(block => block.text()).join('\n')
    expect(code()).toContain('model_catalog_url = "https://example.com/v1/models"')
    expect(code()).toContain('api_key_model_discovery = true')
    expect(fetchMock).not.toHaveBeenCalled()
    await wrapper.get('[data-testid="codex-model-catalog-local"]').trigger('click')
    await flushPromises()
    expect(code()).toContain('model_catalog_json = "~/.codex/models.json"')
    expect(code()).toContain('"slug": "gpt-6.1-sol"')
    expect(code()).not.toContain('model_catalog_url')
    expect(code()).not.toContain('api_key_model_discovery')
    fetchMock.mockResolvedValue(new Response(JSON.stringify({ models: [{ slug: 'gpt-6.1-sol' }] })))
    await wrapper.findAll('button').find(b => b.text() === 'keys.useKeyModal.cliTabs.codexCliWs')!.trigger('click')
    await flushPromises()
    expect(code()).toContain('supports_websockets = true')
    expect(code()).toContain('model_catalog_json')
    await wrapper.get('[data-testid="codex-model-catalog-remote"]').trigger('click')
    expect(code()).toContain('api_key_model_discovery = true')
    expect(code()).not.toContain('"slug"')
    wrapper.unmount()
  })
})
