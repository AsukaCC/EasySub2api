import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ModelFingerprintModal from '../ModelFingerprintModal.vue'
import ModelFingerprintResult from '../ModelFingerprintResult.vue'
import type { Account } from '@/types'
import type { ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'

const { getAvailableModels, getModelFingerprint, startModelFingerprint } = vi.hoisted(() => ({
  getAvailableModels: vi.fn(), getModelFingerprint: vi.fn(), startModelFingerprint: vi.fn()
}))
vi.mock('@/api/admin', () => ({ adminAPI: { accounts: { getAvailableModels } } }))
vi.mock('@/api/admin/modelFingerprint', async importOriginal => ({
  ...await importOriginal<typeof import('@/api/admin/modelFingerprint')>(), getModelFingerprint, startModelFingerprint
}))
vi.mock('vue-i18n', async importOriginal => ({ ...await importOriginal<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key }) }))

const running: ModelFingerprintSnapshot = {
  id: 'job-one', model: 'gpt-6-astra', status: 'running', sampling_mode: 'single_conversation',
  total: 3, completed: 0, valid: 0, started_at: '2026-09-21T00:00:00Z', expires_at: '2026-09-21T00:05:00Z'
}
const account = (id: string) => ({ id, name: `Account ${id}`, extra: {} }) as Account
const mounts: ReturnType<typeof mount>[] = []
const mountModal = () => {
  const wrapper = mount(ModelFingerprintModal, {
    props: { show: true, account: account('one') },
    global: { stubs: {
      BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
      Select: { props: ['options', 'modelValue'], template: '<select><option v-for="item in options" :key="item.id">{{ item.id }}</option></select>' },
      Icon: true
    } }
  })
  mounts.push(wrapper)
  return wrapper
}

describe('Model fingerprint workflow', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    getAvailableModels.mockResolvedValue([{ id: 'gpt-6-astra' }, { id: 'gpt-image-2' }])
    getModelFingerprint.mockResolvedValue(null)
    startModelFingerprint.mockResolvedValue(running)
  })
  afterEach(() => { mounts.splice(0).forEach(wrapper => wrapper.unmount()); vi.useRealTimers() })

  it('selects text models and starts exactly once, then polls without generating more samples', async () => {
    const wrapper = mountModal()
    await flushPromises()
    expect(wrapper.findAll('option').map(item => item.text())).toEqual(['gpt-6-astra'])
    const start = wrapper.get('.btn-primary')
    await start.trigger('click')
    await flushPromises()
    await start.trigger('click')
    expect(startModelFingerprint).toHaveBeenCalledTimes(1)
    expect(startModelFingerprint).toHaveBeenCalledWith('one', 'gpt-6-astra')
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
})
