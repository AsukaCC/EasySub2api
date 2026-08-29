import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'

import PlanEditDialog from '../PlanEditDialog.vue'
import type { AdminGroup } from '@/types'

const { createPlan, updatePlan } = vi.hoisted(() => ({
  createPlan: vi.fn(),
  updatePlan: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
      locale: { value: 'en' },
    }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    createPlan,
    updatePlan,
  },
}))

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: Boolean,
    title: String,
    width: String,
  },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: [String, Number],
    options: {
      type: Array,
      default: () => [],
    },
    placeholder: String,
  },
  emits: ['update:modelValue'],
  setup(_props, { emit }) {
    const onChange = (event: Event) => {
      const value = (event.target as HTMLSelectElement).value
      emit('update:modelValue', value === '' ? null : Number(value))
    }
    return { onChange }
  },
  template: `
    <select
      :value="modelValue ?? ''"
      @change="onChange"
    >
      <option value="">{{ placeholder }}</option>
      <option
        v-for="option in options"
        :key="option.value"
        :value="option.value"
        :data-platform="option.platform"
      >
        {{ option.label }}
      </option>
    </select>
  `,
})

const groupFixture = (overrides: Partial<AdminGroup>): AdminGroup => ({
  id: 1,
  name: 'OpenAI',
  description: null,
  platform: 'openai',
  rate_multiplier: 1,
  rpm_limit: 0,
  is_exclusive: false,
  status: 'active',
  subscription_type: 'subscription',
  daily_limit_usd: null,
  weekly_limit_usd: null,
  monthly_limit_usd: null,
  allow_image_generation: false,
  image_rate_independent: false,
  image_rate_multiplier: 1,
  image_price_1k: null,
  image_price_2k: null,
  image_price_4k: null,
  peak_rate_enabled: false,
  peak_start: '',
  peak_end: '',
  peak_rate_multiplier: 1,
  claude_code_only: false,
  fallback_group_id: null,
  fallback_group_id_on_invalid_request: null,
  allow_messages_dispatch: false,
  require_oauth_only: false,
  require_privacy_set: false,
  created_at: '2026-07-01T00:00:00Z',
  updated_at: '2026-07-01T00:00:00Z',
  model_routing: null,
  model_routing_enabled: false,
  mcp_xml_inject: false,
  sort_order: 0,
  ...overrides,
})

function mountDialog({ groups = [] }: { groups?: AdminGroup[] } = {}) {
  return mount(PlanEditDialog, {
    props: {
      show: true,
      plan: null,
      groups,
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Select: SelectStub,
        Icon: true,
        GroupBadge: true,
      },
    },
  })
}

describe('PlanEditDialog', () => {
  beforeEach(() => {
    createPlan.mockReset().mockResolvedValue({})
    updatePlan.mockReset().mockResolvedValue({})
  })

  it('allows composite subscription groups for payment plans', () => {
    const wrapper = mountDialog({
      groups: [
        groupFixture({
          id: 10,
          name: 'OpenAI + Claude + Grok',
          platform: 'composite',
          rate_multiplier: 1.2,
          subscription_type: 'subscription',
        }),
        groupFixture({
          id: 11,
          name: 'Standard OpenAI',
          platform: 'openai',
          subscription_type: 'standard',
        }),
      ],
    })

    const options = wrapper.findAll('option').map(option => option.text())

    expect(options).toContain('OpenAI + Claude + Grok — composite (1.2x)')
    expect(options).not.toContain('Standard OpenAI — openai (1x)')
  })

  it('shows group quotas and plan prices as points without a currency selector', async () => {
    const wrapper = mountDialog({
      groups: [
        groupFixture({
          id: 10,
          name: 'Point plan group',
          subscription_type: 'subscription',
          daily_limit_points: 50,
          weekly_limit_points: 250,
          monthly_limit_usd: 900,
        }),
      ],
    })

    await wrapper.find('select').setValue('10')

    expect(wrapper.text()).toContain('50.00 points')
    expect(wrapper.text()).toContain('250.00 points')
    expect(wrapper.text()).toContain('900.00 points')
    expect(wrapper.text()).toContain('payment.admin.planPointsHint')
    expect(wrapper.text()).not.toContain('payment.admin.currency')
    expect(wrapper.text()).not.toContain('¥')
    expect(wrapper.text()).not.toContain('$')
  })

  it('creates a point-denominated plan without sending a currency field', async () => {
    const wrapper = mountDialog({
      groups: [groupFixture({ id: 10, subscription_type: 'subscription' })],
    })

    await wrapper.find('input[type="text"]').setValue('Starter points')
    await wrapper.find('select').setValue('10')
    await wrapper.find('textarea').setValue('Point-only subscription')
    await wrapper.findAll('input[type="number"]')[0]?.setValue('19.5')
    await wrapper.find('form').trigger('submit.prevent')

    expect(createPlan).toHaveBeenCalledWith(expect.objectContaining({
      name: 'Starter points',
      group_id: 10,
      price: 19.5,
    }))
    expect(createPlan.mock.calls[0]?.[0]).not.toHaveProperty('currency')
  })
})
