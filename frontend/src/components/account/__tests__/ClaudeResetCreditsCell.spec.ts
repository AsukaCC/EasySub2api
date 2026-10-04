import { describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import type { Account } from '@/types'
import ClaudeResetCreditsCell from '../ClaudeResetCreditsCell.vue'

const api = vi.hoisted(() => ({ query: vi.fn(), redeem: vi.fn() }))
vi.mock('@/api/admin/accounts', () => ({ queryClaudeResetCredits: api.query, redeemClaudeResetCredit: api.redeem }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

describe('Claude native reset confirmation', () => {
  it('loads on demand and reuses the idempotency key after an ambiguous transport failure', async () => {
    api.query.mockResolvedValue({ eligible: true, available_count: 1, credits: [{ redeemable: true }] })
    api.redeem.mockRejectedValueOnce(new Error('network lost')).mockResolvedValueOnce({ outcome: 'unknown', replayed: true })
    const wrapper = mount(ClaudeResetCreditsCell, {
      props: { account: { id: 'account-1', platform: 'anthropic', type: 'oauth' } as Account },
      global: { stubs: { ConfirmDialog: { props: ['show'], emits: ['confirm', 'cancel'], template: '<button v-if="show" data-test="confirm" @click="$emit(\'confirm\')">confirm</button>' } } }
    })
    expect(api.query).not.toHaveBeenCalled()
    await wrapper.findAll('button')[0].trigger('click')
    await flushPromises()
    const confirm = async () => {
      await wrapper.findAll('button').find(b => b.text() === 'admin.accounts.claudeReset.reset')!.trigger('click')
      await wrapper.get('[data-test="confirm"]').trigger('click')
      await flushPromises()
    }
    await confirm()
    await confirm()
    expect(api.redeem).toHaveBeenCalledTimes(2)
    expect(api.redeem.mock.calls[0][1]).toBe(api.redeem.mock.calls[1][1])
    expect(wrapper.text()).toContain('admin.accounts.claudeReset.outcomes.unknown')
    wrapper.unmount()
  })
})
