import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AccountModelRulesManager from '../AccountModelRulesManager.vue'

const { listRules, updateRule, listSubscriptionTiers, listAccounts, syncUpstreamModels } = vi.hoisted(() => ({
  listRules: vi.fn(),
  updateRule: vi.fn(),
  listSubscriptionTiers: vi.fn(),
  listAccounts: vi.fn(),
  syncUpstreamModels: vi.fn()
}))

vi.mock('vue-i18n', async importOriginal => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accountModelRules: {
      list: listRules,
      create: vi.fn(),
      update: updateRule,
      delete: vi.fn()
    },
    accounts: {
      listSubscriptionTiers,
      list: listAccounts,
      syncUpstreamModels
    }
  }
}))

const BaseDialogStub = {
  props: ['show'],
  template: '<div v-if="show" data-testid="dialog"><slot /><slot name="footer" /></div>'
}

const SelectStub = {
  props: ['modelValue', 'options'],
  inheritAttrs: false,
  emits: ['update:modelValue'],
  methods: {
    handleChange(event: Event) {
      this.$emit('update:modelValue', (event.target as HTMLSelectElement).value)
    }
  },
  template: `
    <select v-bind="$attrs" :value="modelValue" @change="handleChange">
      <option v-for="option in options" :key="String(option.value)" :value="option.value">
        {{ option.label }}
      </option>
    </select>
  `
}

function mountManager() {
  return mount(AccountModelRulesManager, {
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        ConfirmDialog: true,
        Select: SelectStub,
        Icon: true,
        PlatformIcon: true,
        LoadingState: true,
        LoadingButtonContent: { template: '<span><slot /></span>' }
      }
    }
  })
}

describe('AccountModelRulesManager', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    listRules.mockReset()
    updateRule.mockReset()
    listSubscriptionTiers.mockReset()
    listAccounts.mockReset()
    syncUpstreamModels.mockReset()
    updateRule.mockResolvedValue(undefined)
    listSubscriptionTiers.mockResolvedValue([{ value: 'pro', label: 'Pro', account_count: 1 }])
    listAccounts.mockResolvedValue({
      items: [{ id: 'account-1', name: 'Upstream account', platform: 'openai', type: 'oauth', subscription_tier: 'pro' }],
      total: 1
    })
    syncUpstreamModels.mockResolvedValue({ models: ['gpt-5.6', 'gpt-5.6-sol'] })
    listRules.mockResolvedValue([{
      id: 'rule-1',
      name: 'OpenAI routing',
      description: 'Routing description',
      platform: 'openai',
      subscription_tier: 'pro',
      model_routes: [{ request_model: 'gpt-5.6', upstream_model: 'gpt-5.6-sol', reasoning_effort: 'high' }],
      bound_account_count: 0,
      created_at: '2026-09-01T00:00:00Z',
      updated_at: '2026-09-01T00:00:00Z'
    }])
  })

  it('does not render a duplicate page title', async () => {
    const wrapper = mountManager()
    await flushPromises()

    expect(wrapper.find('h1').exists()).toBe(false)
    expect(wrapper.text()).toContain('admin.accounts.modelRules.description')
  })

  it('round-trips unified model routes and reasoning effort', async () => {
    const wrapper = mountManager()
    await flushPromises()

    await wrapper.get('button[title="common.edit"]').trigger('click')
    await flushPromises()

    expect((wrapper.get('input[placeholder="admin.accounts.modelRules.fromModel"]').element as HTMLInputElement).value).toBe('gpt-5.6')
    expect((wrapper.get('input[placeholder="admin.accounts.modelRules.toModel"]').element as HTMLInputElement).value).toBe('gpt-5.6-sol')
    expect(wrapper.get('[data-testid="rule-reasoning-effort"]').element.value).toBe('high')

    await wrapper.get('[data-testid="rule-reasoning-effort"]').setValue('xhigh')
    const primaryButtons = wrapper.findAll('[data-testid="dialog"] button.btn-primary')
    await primaryButtons[primaryButtons.length - 1].trigger('click')
    await flushPromises()

    expect(updateRule).toHaveBeenCalledWith('rule-1', expect.objectContaining({
      platform: 'openai',
      subscription_tier: 'pro',
      model_routes: [{ request_model: 'gpt-5.6', upstream_model: 'gpt-5.6-sol', reasoning_effort: 'xhigh' }]
    }))
  })

  it('imports models from the selected real upstream account as identity routes', async () => {
    const wrapper = mountManager()
    await flushPromises()

    await wrapper.get('button[title="common.edit"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="rule-source-account"]').setValue('account-1')
    await wrapper.get('button.btn-secondary').trigger('click')
    await flushPromises()

    expect(syncUpstreamModels).toHaveBeenCalledWith('account-1')
    const importButton = wrapper.findAll('button.btn-secondary')[1]
    await importButton.trigger('click')

    const requestModels = wrapper.findAll('input[placeholder="admin.accounts.modelRules.fromModel"]')
      .map(input => (input.element as HTMLInputElement).value)
    expect(requestModels).toContain('gpt-5.6')
    expect(requestModels).toContain('gpt-5.6-sol')
  })
})
