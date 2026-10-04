import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ModelFingerprintModal from '../ModelFingerprintModal.vue'
import ModelFingerprintResult from '../ModelFingerprintResult.vue'
import ModelFingerprintCell from '../ModelFingerprintCell.vue'
import type { Account } from '@/types'
import type { ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'

const { getFingerprintModels, getModelFingerprint, startModelFingerprint, getFingerprintSchedule, setFingerprintSchedule, getFingerprintHistory, list } = vi.hoisted(() => ({
  getFingerprintModels: vi.fn(), getModelFingerprint: vi.fn(), startModelFingerprint: vi.fn(),
  getFingerprintSchedule: vi.fn(), setFingerprintSchedule: vi.fn(), getFingerprintHistory: vi.fn(), list: vi.fn()
}))
vi.mock('@/api/admin', () => ({ adminAPI: { accounts: { list } } }))
vi.mock('@/api/admin/modelFingerprint', async importOriginal => ({
  ...await importOriginal<typeof import('@/api/admin/modelFingerprint')>(), getModelFingerprint, startModelFingerprint,
  getFingerprintModels, getFingerprintSchedule, setFingerprintSchedule, getFingerprintHistory
}))
vi.mock('vue-i18n', async importOriginal => ({ ...await importOriginal<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key }) }))

const running: ModelFingerprintSnapshot = {
  id: 'job-one', model: 'gpt-6-astra', status: 'running', sampling_mode: 'single_conversation',
  total: 3, completed: 0, valid: 0, started_at: '2026-09-21T00:00:00Z', expires_at: '2026-09-21T00:05:00Z'
}
const account = (id: string) => ({ id, name: `Account ${id}`, type: 'apikey', platform: 'openai', extra: {} }) as Account
const completed: ModelFingerprintSnapshot = {
  ...running, status: 'completed', finished_at: '2026-09-21T00:00:00Z',
  result: {
    revision: 'reference',
    models: [{ model: 'gpt-5.4', family: 'gpt', probability: 0.3 }, { model: 'claude-sonnet-4-6', family: 'claude', probability: 0.7 }],
    families: [{ family: 'gpt', display_name: 'GPT', probability: 0.3 }, { family: 'claude', display_name: 'Claude', probability: 0.7 }]
  }
}
const mounts: ReturnType<typeof mount>[] = []
const mountModal = () => {
  const wrapper = mount(ModelFingerprintModal, {
    props: { show: true, account: account('one') },
    global: { stubs: {
      BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
      Select: { props: ['options', 'modelValue', 'valueKey'], emits: ['update:modelValue'], template: '<select :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value)"><option v-for="item in options" :key="item.id || item.value" :value="item[valueKey || \'value\']">{{ item.id || item.value }}</option></select>' },
      RouterLink: { template: '<a><slot /></a>' }, DataTable: true, Pagination: true,
      Icon: true
    } }
  })
  mounts.push(wrapper)
  return wrapper
}

describe('Model fingerprint workflow', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(running.started_at))
    vi.clearAllMocks()
    getFingerprintModels.mockResolvedValue([{ id: 'gpt-6-astra', reasoning_levels: ['low', 'high'] }])
    list.mockResolvedValue({ items: [account('one'), account('two')] })
    getFingerprintSchedule.mockResolvedValue({ enabled: false, options: {} })
    setFingerprintSchedule.mockResolvedValue({ enabled: true, next_run_at: '2026-09-21T00:30:00Z' })
    getFingerprintHistory.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 10 })
    getModelFingerprint.mockResolvedValue(null)
    startModelFingerprint.mockResolvedValue(running)
  })
  afterEach(() => { mounts.splice(0).forEach(wrapper => wrapper.unmount()); vi.useRealTimers() })

  it('selects text models and starts exactly once, then polls without generating more samples', async () => {
    const wrapper = mountModal()
    await flushPromises()
    expect(getFingerprintModels).toHaveBeenCalledWith('one')
    expect(wrapper.findAll('select')[2].findAll('option').map(item => item.text())).toEqual(['gpt-6-astra'])
    const start = wrapper.get('.btn-primary')
    await start.trigger('click')
    await flushPromises()
    await start.trigger('click')
    expect(startModelFingerprint).toHaveBeenCalledTimes(1)
    expect(startModelFingerprint).toHaveBeenCalledWith('one', 'gpt-6-astra', { protocol: 'auto', reasoning_effort: '' })
    expect(start.attributes('disabled')).toBeDefined()
    getModelFingerprint.mockResolvedValue({ ...running, status: 'failed', completed: 3, error: 'insufficient_samples' })
    await vi.advanceTimersByTimeAsync(1600)
    await flushPromises()
    expect(wrapper.text()).toContain('insufficient_samples')
    expect(startModelFingerprint).toHaveBeenCalledTimes(1)
    const calls = getModelFingerprint.mock.calls.length
    await vi.advanceTimersByTimeAsync(6000)
    expect(getModelFingerprint).toHaveBeenCalledTimes(calls)
  })

  it('closing stops polling and does not cancel the server job', async () => {
    getModelFingerprint.mockResolvedValue(running)
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.setProps({ show: false })
    const calls = getModelFingerprint.mock.calls.length
    await vi.advanceTimersByTimeAsync(10000)
    expect(getModelFingerprint).toHaveBeenCalledTimes(calls)
    expect(startModelFingerprint).not.toHaveBeenCalled()
  })

  it('does not apply an old account response after switching accounts', async () => {
    let resolveOld!: (value: ModelFingerprintSnapshot) => void
    getModelFingerprint.mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve }))
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.setProps({ account: account('two') })
    await flushPromises()
    resolveOld(running)
    await flushPromises()
    expect(wrapper.get('.btn-primary').attributes('disabled')).toBeUndefined()
    expect(wrapper.emitted('update')?.some(args => args[0] === 'two' && (args[1] as ModelFingerprintSnapshot)?.id === running.id)).toBeFalsy()
  })

  it('renders model shares and family proportions without implying requested-model identity', () => {
    const wrapper = mount(ModelFingerprintResult, {
      props: { snapshot: { ...running, status: 'completed', result: {
        revision: 'reference',
        models: [{ model: 'claude-sonnet-4-6', family: 'claude', probability: 0.7 }, { model: 'gpt-5.4', family: 'gpt', probability: 0.3 }],
        families: [{ family: 'gpt', display_name: 'GPT', probability: 0.3 }, { family: 'claude', display_name: 'Claude', probability: 0.7 }]
      } } }, global: { stubs: { Icon: true } }
    })
    mounts.push(wrapper)
    expect(wrapper.text()).toContain('claude-sonnet-4-6')
    expect(wrapper.text()).toContain('70.0%')
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe('GPT 30.0%, Claude 70.0%')
  })

  it('shows only the highest share and expires saved list results exactly after 24 hours', async () => {
    vi.setSystemTime(new Date('2026-09-21T23:59:59Z'))
    const wrapper = mount(ModelFingerprintCell, {
      props: { account: { ...account('one'), extra: { model_fingerprint: completed } } },
      global: { stubs: { Icon: true } }
    })
    mounts.push(wrapper)
    expect(wrapper.findAll('.fingerprint-row > span').map(item => item.text())).toEqual(['claude-sonnet-4-6', '70.0%'])
    expect(wrapper.find('.fingerprint-model').exists()).toBe(false)
    expect(wrapper.find('.fingerprint-bar').exists()).toBe(false)
    expect(wrapper.get('.fingerprint-candidate').attributes('title')).toBe('claude-sonnet-4-6')
    await vi.advanceTimersByTimeAsync(999)
    expect(wrapper.text()).toContain('70.0%')
    await vi.advanceTimersByTimeAsync(1)
    expect(wrapper.find('.fingerprint-result').exists()).toBe(false)
    expect(wrapper.emitted('update')).toEqual([['one', null]])
    expect(getModelFingerprint).not.toHaveBeenCalled()
    expect(startModelFingerprint).not.toHaveBeenCalled()
  })

  it('replaces the expiry timer when a newer result arrives', async () => {
    vi.setSystemTime(new Date('2026-09-21T23:59:59Z'))
    const wrapper = mount(ModelFingerprintCell, {
      props: { account: { ...account('one'), extra: { model_fingerprint: completed } } },
      global: { stubs: { Icon: true } }
    })
    mounts.push(wrapper)
    const newer = { ...completed, id: 'job-two', finished_at: new Date().toISOString() }
    await wrapper.setProps({ account: { ...account('one'), extra: { model_fingerprint: newer } } })
    await vi.advanceTimersByTimeAsync(1000)
    expect(wrapper.text()).toContain('70.0%')
    expect(wrapper.emitted('update')).toBeUndefined()
    await vi.advanceTimersByTimeAsync(24 * 60 * 60 * 1000 - 1000)
    expect(wrapper.emitted('update')).toEqual([['one', null]])
  })

  it('expires dialog results without polling or rerunning model tests', async () => {
    vi.setSystemTime(new Date('2026-09-21T23:59:59Z'))
    getModelFingerprint.mockResolvedValue(completed)
    const wrapper = mountModal()
    await flushPromises()
    expect(wrapper.find('.fingerprint-output').exists()).toBe(true)
    await vi.advanceTimersByTimeAsync(1000)
    expect(wrapper.find('.fingerprint-output').exists()).toBe(false)
    expect(wrapper.emitted('update')?.at(-1)).toEqual(['one', null])
    expect(getModelFingerprint).toHaveBeenCalledTimes(1)
    expect(startModelFingerprint).not.toHaveBeenCalled()
  })

  it('hides already expired API results and rechecks saved results after tab suspension', async () => {
    vi.setSystemTime(new Date('2026-09-22T00:00:00Z'))
    getModelFingerprint.mockResolvedValue(completed)
    const modal = mountModal()
    await flushPromises()
    expect(modal.find('.fingerprint-output').exists()).toBe(false)
    expect(modal.emitted('update')).toEqual([['one', null]])

    vi.setSystemTime(new Date('2026-09-21T01:00:00Z'))
    const cell = mount(ModelFingerprintCell, {
      props: { account: { ...account('two'), extra: { model_fingerprint: completed } } },
      global: { stubs: { Icon: true } }
    })
    mounts.push(cell)
    vi.setSystemTime(new Date('2026-09-22T01:00:00Z'))
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(cell.find('.fingerprint-result').exists()).toBe(false)
    expect(cell.emitted('update')).toEqual([['two', null]])
  })

  it('uses the selected account, protocol and effort for manual and scheduled tests', async () => {
    const wrapper = mountModal()
    await flushPromises()
    await wrapper.findAll('select')[0].setValue('two')
    await flushPromises()
    expect(getFingerprintModels).toHaveBeenLastCalledWith('two')
    await wrapper.findAll('select')[1].setValue('anthropic')
    await wrapper.findAll('select')[3].setValue('high')
    await wrapper.get('input[type="checkbox"]').setValue(true)
    await wrapper.get('.fingerprint-schedule button').trigger('click')
    await flushPromises()
    expect(setFingerprintSchedule).toHaveBeenCalledWith('two', { enabled: true, options: { model_id: 'gpt-6-astra', protocol: 'anthropic', reasoning_effort: 'high' } })
    await wrapper.get('.btn-primary').trigger('click')
    await flushPromises()
    expect(startModelFingerprint).toHaveBeenCalledWith('two', 'gpt-6-astra', { protocol: 'anthropic', reasoning_effort: 'high' })
  })

  it('restores saved parameters and cannot start when upstream models fail to load', async () => {
    getFingerprintSchedule.mockResolvedValue({ enabled: true, options: { model_id: 'gpt-6-astra', protocol: 'chat', reasoning_effort: 'high' } })
    const wrapper = mountModal()
    await flushPromises()
    expect((wrapper.findAll('select')[3].element as HTMLSelectElement).value).toBe('high')
    getFingerprintModels.mockRejectedValueOnce(new Error('offline'))
    await wrapper.findAll('select')[0].setValue('two')
    await flushPromises()
    expect(wrapper.get('.btn-primary').attributes('disabled')).toBeDefined()
    expect(startModelFingerprint).not.toHaveBeenCalled()
  })
})
