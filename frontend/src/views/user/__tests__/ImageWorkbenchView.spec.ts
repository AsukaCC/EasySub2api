import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import ImageWorkbenchView from '../ImageWorkbenchView.vue'
import { listImageModels, loadWorkbenchCredentials, type ImageModel } from '@/api/imageWorkbench'
import type { ApiKey, Group } from '@/types'

vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key }),
}))
vi.mock('@/stores', () => ({ useAppStore: () => ({ publicSettingsLoaded: true }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
vi.mock('@/api/imageWorkbench', async (importOriginal) => ({
  ...await importOriginal<typeof import('@/api/imageWorkbench')>(),
  listImageModels: vi.fn(),
  loadWorkbenchCredentials: vi.fn(),
}))
vi.mock('@/features/image-workbench/storage', () => ({
  listHistory: vi.fn().mockResolvedValue([]),
  listTasks: vi.fn().mockResolvedValue([]),
}))

function deferred() {
  let resolve!: (models: ImageModel[]) => void
  let reject!: (error: Error) => void
  const promise = new Promise<ImageModel[]>((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}

let wrapper: ReturnType<typeof shallowMount> | undefined

async function mountWorkbench() {
  wrapper = shallowMount(ImageWorkbenchView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: { template: '<div><slot /></div>' },
      },
    },
  })
  await flushPromises()
  const keySelect = wrapper.findComponent('[data-testid="workbench-key-select"]')
  const modelSelect = wrapper.findComponent('[data-testid="workbench-model-select"]')
  return { keySelect, modelSelect }
}

describe('image workbench model selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    const group = { id: 'images', platform: 'openai', allow_image_generation: true } as Group
    vi.mocked(loadWorkbenchCredentials).mockResolvedValue({
      keys: ['first', 'second'].map((id) => ({ id, key: `sk-${id}`, status: 'active', group_id: group.id, group_ids: [group.id], group }) as ApiKey),
      groups: [group],
    })
  })

  afterEach(() => { wrapper?.unmount() })

  it.each(['success', 'failure'])('ignores stale request %s after switching keys', async (outcome) => {
    const old = deferred()
    vi.mocked(listImageModels).mockReturnValueOnce(old.promise).mockResolvedValueOnce([{ id: 'gpt-image-2' }])
    const { keySelect, modelSelect } = await mountWorkbench()
    expect(modelSelect.props('disabled')).toBe(true)
    await keySelect.setValue('second')
    await flushPromises()
    if (outcome === 'success') old.resolve([{ id: 'gpt-image-1' }])
    else old.reject(new Error('old failure'))
    await flushPromises()
    expect(modelSelect.props('modelValue')).toBe('gpt-image-2')
    expect(JSON.stringify(modelSelect.props('options'))).not.toContain('gpt-image-1')
    expect(wrapper!.find('[role="alert"]').exists()).toBe(false)
    expect(listImageModels).toHaveBeenCalledTimes(2)
  })

  it('clears stale models on failure and clears the error after recovery', async () => {
    vi.mocked(listImageModels)
      .mockResolvedValueOnce([{ id: 'gpt-image-1' }])
      .mockRejectedValueOnce(new Error('API key expired'))
      .mockResolvedValueOnce([{ id: 'gpt-image-2' }])
    const { keySelect, modelSelect } = await mountWorkbench()
    await keySelect.setValue('second')
    await flushPromises()
    expect(modelSelect.props('modelValue')).toBe('')
    expect(modelSelect.props('disabled')).toBe(true)
    expect(wrapper!.get('[role="alert"]').text()).toBe('API key expired')
    await keySelect.setValue('first')
    await flushPromises()
    expect(modelSelect.props('modelValue')).toBe('gpt-image-2')
    expect(wrapper!.find('[role="alert"]').exists()).toBe(false)
  })

  it('preserves a saved model when it is still available', async () => {
    localStorage.setItem('image-workbench-preferences', JSON.stringify({
      selectedByPlatform: { openai: { keyId: 'first', model: 'gpt-image-2' } },
    }))
    vi.mocked(listImageModels).mockResolvedValue([{ id: 'gpt-image-1' }, { id: 'gpt-image-2' }])
    const { modelSelect } = await mountWorkbench()
    expect(modelSelect.props('modelValue')).toBe('gpt-image-2')
    expect(listImageModels).toHaveBeenCalledTimes(1)
  })
})
