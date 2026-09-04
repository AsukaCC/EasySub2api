import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AccountModelRulesManager from '../AccountModelRulesManager.vue'

const { listRules, updateRule } = vi.hoisted(() => ({
  listRules: vi.fn(),
  updateRule: vi.fn()
}))

vi.mock('vue-i18n', async importOriginal => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accountModelRules: {
      list: listRules,
      create: vi.fn(),
      update: updateRule,
      delete: vi.fn()
    },
    channels: {
      syncPricingModels: vi.fn()
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
        ModelWhitelistSelector: true
      }
    }
  })
}

describe('AccountModelRulesManager', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    listRules.mockReset()
    updateRule.mockReset()
    updateRule.mockResolvedValue(undefined)
    listRules.mockResolvedValue([
      {
        id: 'rule-1',
        name: 'OpenAI mapping',
        description: 'Mapping description',
        platform: 'openai',
        whitelist: [],
        mapping: { 'gpt-4o': 'gpt-4.1' },
        reasoning_efforts: { 'gpt-4o': 'high' },
        created_at: '2026-09-01T00:00:00Z',
        updated_at: '2026-09-01T00:00:00Z'
      }
    ])
  })

  it('keeps the existing mapping when editing a non-default platform rule', async () => {
    const wrapper = mountManager()
    await flushPromises()

    await wrapper.get('button[title="common.edit"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[data-testid="dialog"] select').element.value).toBe('openai')
    expect(wrapper.get('button[role="tab"][aria-selected="true"]').text()).toContain(
      'admin.accounts.modelRules.mapping'
    )
    expect(
      (wrapper.get('input[placeholder="admin.accounts.modelRules.fromModel"]').element as HTMLInputElement).value
    ).toBe('gpt-4o')
    expect(
      (wrapper.get('input[placeholder="admin.accounts.modelRules.toModel"]').element as HTMLInputElement).value
    ).toBe('gpt-4.1')
    expect(wrapper.get('select[aria-label="admin.accounts.modelRules.reasoningEffort"]').element.value).toBe('high')
  })

  it('saves the selected reasoning effort for OpenAI mappings', async () => {
    const wrapper = mountManager()
    await flushPromises()

    await wrapper.get('button[title="common.edit"]').trigger('click')
    await wrapper.get('select[aria-label="admin.accounts.modelRules.reasoningEffort"]').setValue('xhigh')
    const primaryButtons = wrapper.findAll('[data-testid="dialog"] button.btn-primary')
    await primaryButtons[primaryButtons.length - 1].trigger('click')
    await flushPromises()

    expect(updateRule).toHaveBeenCalledWith(
      'rule-1',
      expect.objectContaining({ reasoning_efforts: { 'gpt-4o': 'xhigh' } })
    )
  })

  it('still clears restrictions when the administrator changes platform', async () => {
    const wrapper = mountManager()
    await flushPromises()

    await wrapper.get('button[title="common.edit"]').trigger('click')
    await wrapper.get('[data-testid="dialog"] select').setValue('gemini')
    await wrapper
      .get('button[role="tab"][aria-selected="false"]')
      .trigger('click')

    expect(
      (wrapper.get('input[placeholder="admin.accounts.modelRules.fromModel"]').element as HTMLInputElement).value
    ).toBe('')
    expect(
      (wrapper.get('input[placeholder="admin.accounts.modelRules.toModel"]').element as HTMLInputElement).value
    ).toBe('')
  })
})
