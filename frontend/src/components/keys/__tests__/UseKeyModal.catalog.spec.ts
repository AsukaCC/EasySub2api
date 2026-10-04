import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import UseKeyModal from '../UseKeyModal.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
afterEach(() => vi.unstubAllGlobals())

describe('Codex catalog configuration', () => {
  it('supports remote and fetched local catalogs for HTTP and WebSocket clients', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ models: [{ slug: 'gpt-6.1-sol' }] })))
    vi.stubGlobal('fetch', fetchMock)
    const wrapper = mount(UseKeyModal, {
      props: { show: true, apiKey: 'synthetic-key', baseUrl: 'https://example.com/v1', platform: 'openai' },
      global: { stubs: { BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }, Icon: true } }
    })
    const code = () => wrapper.findAll('pre code').map(block => block.text()).join('\n')
    expect(code()).toContain('model_catalog_url = "https://example.com/v1/models"')
    expect(fetchMock).not.toHaveBeenCalled()
    await wrapper.get('[data-testid="codex-model-catalog-local"]').trigger('click')
    await flushPromises()
    expect(code()).toContain('model_catalog_json = "~/.codex/models.json"')
    expect(code()).toContain('"slug": "gpt-6.1-sol"')
    expect(code()).not.toContain('model_catalog_url')
    fetchMock.mockResolvedValue(new Response(JSON.stringify({ models: [{ slug: 'gpt-6.1-sol' }] })))
    await wrapper.findAll('button').find(b => b.text() === 'keys.useKeyModal.cliTabs.codexCliWs')!.trigger('click')
    await flushPromises()
    expect(code()).toContain('supports_websockets = true')
    expect(code()).toContain('model_catalog_json')
    await wrapper.get('[data-testid="codex-model-catalog-remote"]').trigger('click')
    expect(code()).not.toContain('"slug"')
    wrapper.unmount()
  })
})
