import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import OpenAIReferralCell from '../OpenAIReferralCell.vue'
import type { Account } from '@/types'
import { refreshCodexCredits } from '@/api/admin/upstreamCapabilities'

vi.mock('@/api/admin/upstreamCapabilities', () => ({
  refreshCodexCredits: vi.fn(), refreshCodexReferrals: vi.fn(), sendCodexReferral: vi.fn()
}))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

const balance = (value: string) => ({ has_credits: true, unlimited: false, balance: value })
const account = (id: string, value?: string, fetchedAt = 100): Account => ({
  id, platform: 'openai', type: 'oauth',
  extra: value === undefined ? {} : { codex_credits_snapshot: { credits: balance(value), fetched_at: fetchedAt } }
} as Account)
const wrappers: ReturnType<typeof mount>[] = []
function mountCell(value: Account) {
  const wrapper = mount(OpenAIReferralCell, {
    props: { account: value },
    global: { stubs: { Icon: true, BaseDialog: true } }
  })
  wrappers.push(wrapper)
  return wrapper
}
const refresh = (wrapper: ReturnType<typeof mount>) => wrapper.get('button[aria-label="upstreamUpdate.refresh"]')

beforeEach(() => vi.clearAllMocks())
afterEach(() => wrappers.splice(0).forEach(wrapper => wrapper.unmount()))

describe('Codex credits account cache', () => {
  it('restores precise cached balances on mount without an upstream request', () => {
    const wrapper = mountCell(account('one', '123.4567890123456789'))
    expect(wrapper.text()).toContain('123.4567890123456789')
    expect(wrapper.get('.codex-capabilities__row span').attributes('title')).toContain('upstreamUpdate.updated')
    expect(refreshCodexCredits).not.toHaveBeenCalled()
  })

  it('keeps zero and unlimited balances and isolates reused account rows', async () => {
    const wrapper = mountCell(account('one', '0'))
    expect(wrapper.text()).toContain(': 0')
    await wrapper.setProps({ account: account('two', '22') })
    expect(wrapper.text()).toContain(': 22')
    const unlimited = account('shadow', '0')
    unlimited.parent_account_id = 'two'
    unlimited.extra!.codex_credits_snapshot!.credits!.unlimited = true
    await wrapper.setProps({ account: unlimited })
    expect(wrapper.text()).toContain('upstreamUpdate.unlimited')
    await wrapper.setProps({ account: account('unknown') })
    expect(wrapper.text()).toContain('upstreamUpdate.unknown')
    expect(wrapper.text()).not.toContain('upstreamUpdate.unlimited')
    expect(refreshCodexCredits).not.toHaveBeenCalled()
  })

  it('refreshes and emits a snapshot that can restore the next mounted row', async () => {
    vi.mocked(refreshCodexCredits).mockResolvedValue({ credits: balance('42.50'), fetched_at: 200, cache_persisted: true, credits_cache_persisted: true })
    const original = account('one', '10')
    const wrapper = mountCell(original)
    await refresh(wrapper).trigger('click')
    await flushPromises()
    expect(refreshCodexCredits).toHaveBeenCalledTimes(1)
    expect(refreshCodexCredits).toHaveBeenCalledWith('one')
    expect(wrapper.text()).toContain('42.50')
    const snapshot = wrapper.emitted('credits-updated')![0]![0] as NonNullable<NonNullable<Account['extra']>['codex_credits_snapshot']>
    expect(snapshot).toEqual({ credits: balance('42.50'), fetched_at: 200 })
    const remounted = mountCell({ ...original, extra: { codex_credits_snapshot: snapshot } })
    expect(remounted.text()).toContain('42.50')
    expect(refreshCodexCredits).toHaveBeenCalledTimes(1)
    await wrapper.setProps({ account: account('one', '5', 150) })
    expect(wrapper.text()).toContain('42.50')
    await wrapper.setProps({ account: account('one', '43', 300) })
    expect(wrapper.text()).toContain(': 43')
  })

  it('preserves the last balance when refresh fails', async () => {
    vi.mocked(refreshCodexCredits).mockRejectedValue(new Error('offline'))
    const wrapper = mountCell(account('one', '10'))
    await refresh(wrapper).trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain(': 10')
    expect(wrapper.get('[role="alert"]').text()).toBe('upstreamUpdate.queryFailed')
    expect(wrapper.emitted('credits-updated')).toBeUndefined()
  })

  it('reports a failed persistent cache write while displaying fresh credits', async () => {
    vi.mocked(refreshCodexCredits).mockResolvedValue({ credits: balance('20'), fetched_at: 200, cache_persisted: false, credits_cache_persisted: false })
    const wrapper = mountCell(account('one', '10'))
    await refresh(wrapper).trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain(': 20')
    expect(wrapper.get('[role="alert"]').text()).toBe('upstreamUpdate.creditsCacheFailed')
  })

  it('ignores the previous account response when switching during refresh', async () => {
    let resolve!: (value: Awaited<ReturnType<typeof refreshCodexCredits>>) => void
    vi.mocked(refreshCodexCredits).mockReturnValue(new Promise(done => { resolve = done }))
    const wrapper = mountCell(account('one', '10'))
    await refresh(wrapper).trigger('click')
    await wrapper.setProps({ account: account('two', '22') })
    resolve({ credits: balance('999'), fetched_at: 200, cache_persisted: true })
    await flushPromises()
    expect(wrapper.text()).toContain(': 22')
    expect(wrapper.text()).not.toContain('999')
    expect(wrapper.emitted('credits-updated')).toBeUndefined()
  })
})
