import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import OpenAIModelForwardingFields from '../OpenAIModelForwardingFields.vue'
import Select from '@/components/common/Select.vue'

const mocks = vi.hoisted(() => ({ sync: vi.fn(), preview: vi.fn() }))
vi.mock('@/api/admin/accounts', () => ({ accountsAPI: {
  syncUpstreamModels: mocks.sync,
  syncUpstreamModelsPreview: mocks.preview
} }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

function mountFields(props: Record<string, unknown> = {}) {
  return mount(OpenAIModelForwardingFields, {
    props: { modelValue: [{ from: 'client-*', to: '' }], accountId: 'account-1', ...props },
    global: { stubs: { Teleport: true, Icon: true } }
  })
}

describe('OpenAIModelForwardingFields', () => {
  beforeEach(() => {
    mocks.sync.mockReset().mockResolvedValue({ models: ['upstream-only-model', 'gpt-6.1-sol', 'upstream-only-model'] })
    mocks.preview.mockReset().mockResolvedValue({ models: ['preview-model'] })
  })

  it('offers discovered upstream models and saves a selected target without changing the source', async () => {
    const wrapper = mountFields()
    await flushPromises()
    expect(mocks.sync).toHaveBeenCalledWith('account-1')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    const picker = wrapper.getComponent(Select)
    expect(picker.props('options')).toEqual([
      { value: 'upstream-only-model', label: 'upstream-only-model', disabled: false },
      { value: 'gpt-6.1-sol', label: 'gpt-6.1-sol', disabled: false }
    ])
    await picker.get('button.select-trigger').trigger('click')
    await picker.get('input').setValue('upstream-only')
    const option = picker.get('[role="option"]')
    expect(option.text()).toBe('upstream-only-model')
    await option.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[[{ from: 'client-*', to: 'upstream-only-model' }]]])
    wrapper.unmount()
  })

  it('previews unsaved API key credentials without saving or selecting whitelist models', async () => {
    const syncCredentials = { platform: 'openai', type: 'apikey', api_key: 'test-key', base_url: 'https://example.com' }
    const wrapper = mountFields({ accountId: undefined, syncCredentials })
    await flushPromises()
    expect(mocks.preview).toHaveBeenCalledWith(syncCredentials)
    expect(mocks.sync).not.toHaveBeenCalled()
    expect(wrapper.getComponent(Select).props('options')).toEqual([{ value: 'preview-model', label: 'preview-model', disabled: false }])
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })

  it('preserves a saved target when discovery fails and allows retrying', async () => {
    mocks.sync.mockRejectedValueOnce({ message: 'Upstream unavailable' })
    const wrapper = mountFields({ modelValue: [{ from: 'alias', to: 'saved-model' }] })
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Upstream unavailable')
    expect(wrapper.getComponent(Select).props('options')).toEqual([{ value: 'saved-model', label: 'saved-model', disabled: true }])
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await wrapper.get('[data-testid="load-forwarding-models"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('ignores a previous account response after switching accounts', async () => {
    let resolveOld!: (result: { models: string[] }) => void
    mocks.sync.mockReturnValueOnce(new Promise(resolve => { resolveOld = resolve }))
    const wrapper = mountFields()
    await wrapper.setProps({ accountId: 'account-2' })
    await flushPromises()
    resolveOld({ models: ['old-account-model'] })
    await flushPromises()
    expect(wrapper.getComponent(Select).props('options')).not.toEqual(expect.arrayContaining([expect.objectContaining({ value: 'old-account-model' })]))
    expect(mocks.sync).toHaveBeenLastCalledWith('account-2')
    wrapper.unmount()
  })
})
