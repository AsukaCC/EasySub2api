import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const mocks = vi.hoisted(() => ({
  copyToClipboard: vi.fn().mockResolvedValue(true),
  syncUpstreamModels: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  showInfo: vi.fn(),
  showWarning: vi.fn()
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'common.copy') return '复制'
        if (key === 'admin.accounts.syncUpstreamModelsError') return `同步失败：${params?.message ?? ''}`
        return key
      }
    })
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: mocks.showError,
    showSuccess: mocks.showSuccess,
    showInfo: mocks.showInfo,
    showWarning: mocks.showWarning
  })
}))

vi.mock('@/api/admin/accounts', () => ({
  accountsAPI: {
    syncUpstreamModels: mocks.syncUpstreamModels,
    syncUpstreamModelsPreview: vi.fn()
  }
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: mocks.copyToClipboard
  })
}))

import ModelWhitelistSelector from '../ModelWhitelistSelector.vue'

function mountSelector(props: Record<string, unknown> = {}) {
  return mount(ModelWhitelistSelector, {
    props: {
      modelValue: [],
      platform: 'openai',
      ...props
    },
    global: {
      stubs: {
        ModelIcon: true,
        Teleport: true
      }
    }
  })
}

function findModelRow(wrapper: ReturnType<typeof mountSelector>, modelId: string) {
  const row = wrapper
    .findAll('[data-testid="model-option"]')
    .find(candidate => candidate.text().includes(modelId))

  if (!row) {
    throw new Error(`Model row not found: ${modelId}`)
  }

  return row
}

describe('ModelWhitelistSelector', () => {
  beforeEach(() => {
    mocks.copyToClipboard.mockClear()
    mocks.syncUpstreamModels.mockReset()
    mocks.showError.mockClear()
    mocks.showSuccess.mockClear()
    mocks.showInfo.mockClear()
    mocks.showWarning.mockClear()
  })

  it('copies a model ID without selecting the model', async () => {
    const wrapper = mountSelector()
    await wrapper.get('.components-account-model-whitelist-selector__panel-2').trigger('click')

    const row = findModelRow(wrapper, 'gpt-5.6-sol')

    const copyButton = row.get('[data-testid="copy-model-id"]')
    expect(copyButton.attributes('aria-label')).toBe('复制 gpt-5.6-sol')

    await copyButton.trigger('click')
    await flushPromises()

    expect(mocks.copyToClipboard).toHaveBeenCalledWith('gpt-5.6-sol')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('keeps the existing model selection behavior', async () => {
    const wrapper = mountSelector()
    await wrapper.get('.components-account-model-whitelist-selector__panel-2').trigger('click')

    const row = findModelRow(wrapper, 'gpt-5.6-sol')
    await row.get('[data-testid="select-model"]').trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[['gpt-5.6-sol']]])
    expect(mocks.copyToClipboard).not.toHaveBeenCalled()
  })

  it('shows upstream sync only for supported saved account types', () => {
    const openAIOAuth = mountSelector({ accountId: 'account-1', accountType: 'oauth' })
    expect(openAIOAuth.find('[data-testid="sync-upstream-models"]').exists()).toBe(true)

    const antigravityOAuth = mountSelector({
      accountId: 'account-2',
      platform: 'antigravity',
      accountType: 'oauth'
    })
    expect(antigravityOAuth.find('[data-testid="sync-upstream-models"]').exists()).toBe(false)
  })

  it('shows the backend message from a plain API error object', async () => {
    mocks.syncUpstreamModels.mockRejectedValueOnce({
      status: 400,
      message: 'No OpenAI access token is available'
    })
    const wrapper = mountSelector({ accountId: 'account-1', accountType: 'oauth' })

    await wrapper.get('[data-testid="sync-upstream-models"]').trigger('click')
    await flushPromises()

    expect(mocks.showError).toHaveBeenCalledWith('同步失败：No OpenAI access token is available')
  })
})
