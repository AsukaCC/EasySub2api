import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import AdminRefundDialog from '../AdminRefundDialog.vue'
import type { PaymentOrder } from '@/types/payment'

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

const BaseDialogStub = defineComponent({
  props: { show: Boolean },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const order: PaymentOrder = {
  id: 'order-12',
  user_id: 'user-3',
  amount: 100,
  principal_amount: 100,
  pay_amount: 102,
  base_points: 100,
  bonus_points: 20,
  credited_points: 120,
  wallet_amount: 0,
  wallet_bonus_amount: 0,
  wallet_recharge_amount: 0,
  gateway_base_amount: 100,
  wallet_only: false,
  fee_rate: 2,
  payment_type: 'wxpay',
  out_trade_no: 'trade-12',
  status: 'COMPLETED',
  order_type: 'balance',
  created_at: '2026-08-20T00:00:00Z',
  expires_at: '2026-08-20T00:10:00Z',
  refund_amount: 0,
}

afterEach(() => {
  vi.restoreAllMocks()
})

describe('AdminRefundDialog', () => {
  it('emits principal RMB and reuses one idempotency key for retries in the same dialog', async () => {
    const randomUUID = vi.spyOn(globalThis.crypto, 'randomUUID')
      .mockReturnValueOnce('refund-intent-1')
      .mockReturnValueOnce('refund-intent-2')
    const wrapper = mount(AdminRefundDialog, {
      props: { show: true, order },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
        },
      },
    })

    await wrapper.find('textarea').setValue('approved by support')
    await wrapper.find('form').trigger('submit.prevent')
    await wrapper.find('form').trigger('submit.prevent')

    const first = wrapper.emitted('confirm')?.[0]?.[0] as Record<string, unknown>
    const retry = wrapper.emitted('confirm')?.[1]?.[0] as Record<string, unknown>
    expect(first).toEqual({
      principal_amount: 100,
      reason: 'approved by support',
      idempotency_key: 'admin-refund-order-12-refund-intent-1',
    })
    expect(retry.idempotency_key).toBe(first.idempotency_key)
    expect(first).not.toHaveProperty('amount')
    expect(first).not.toHaveProperty('deduct_balance')
    expect(first).not.toHaveProperty('force')

    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await flushPromises()
    await wrapper.find('form').trigger('submit.prevent')

    const nextIntent = wrapper.emitted('confirm')?.[2]?.[0] as Record<string, unknown>
    expect(nextIntent.idempotency_key).toBe('admin-refund-order-12-refund-intent-2')
    expect(randomUUID).toHaveBeenCalledTimes(2)
  })
})
