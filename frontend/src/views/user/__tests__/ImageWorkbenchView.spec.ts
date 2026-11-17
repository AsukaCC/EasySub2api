import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import ImageWorkbenchView from '../ImageWorkbenchView.vue'
import { listImageModels, loadWorkbenchCredentials, submitImageTask, generateImage, type ImageModel } from '@/api/imageWorkbench'
import { putTask, putHistory } from '@/features/image-workbench/storage'
import type { ApiKey, Group } from '@/types'

vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key }),
}))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ publicSettingsLoaded: true, showSuccess: vi.fn() }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: vi.fn() }) }))
vi.mock('@/api/imageWorkbench', async (importOriginal) => ({
  ...await importOriginal<typeof import('@/api/imageWorkbench')>(),
  listImageModels: vi.fn(),
  loadWorkbenchCredentials: vi.fn(),
  submitImageTask: vi.fn(),
  generateImage: vi.fn(),
  getImageTask: vi.fn().mockResolvedValue({ status: 'processing' }),
}))
vi.mock('@/features/image-workbench/storage', () => ({
  listHistory: vi.fn().mockResolvedValue([]),
  listTasks: vi.fn().mockResolvedValue([]),
  putTask: vi.fn().mockResolvedValue(undefined),
  putHistory: vi.fn().mockResolvedValue(undefined),
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
        Transition: { template: '<div><slot /></div>' },
        BaseDialog: { name: 'BaseDialog', template: '<div><slot /></div>' },
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

  afterEach(() => { wrapper?.unmount(); vi.unstubAllGlobals() })

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

  it.each(['close', 'save', 'connection-save'])('submits confirmed preferences after %s', async (action) => {
    vi.mocked(listImageModels).mockResolvedValue([{ id: 'gpt-image-2' }])
    vi.mocked(submitImageTask).mockResolvedValue({ id: 'task', task_id: 'task', status: 'queued' })
    await mountWorkbench()
    await wrapper!.get('.toolbar-pill--model').trigger('click')
    await wrapper!.findAll('.settings-tabs button')[1].trigger('click')
    await wrapper!.get('.param-field--button').trigger('click')
    const picker = wrapper!.findComponent({ name: 'SizePickerModal' })
    picker.vm.$emit('select', '3840x2160')
    picker.vm.$emit('close')
    await wrapper!.findComponent('.settings-form select-stub').setValue('high')
    if (action === 'close') wrapper!.findComponent({ name: 'BaseDialog' }).vm.$emit('close')
    else {
      if (action === 'connection-save') await wrapper!.findAll('.settings-tabs button')[0].trigger('click')
      await wrapper!.get('.settings-form .primary-button').trigger('click')
    }
    await wrapper!.get('textarea').setValue('test image')
    await wrapper!.get('form').trigger('submit')
    await flushPromises()
    expect(submitImageTask).toHaveBeenCalledWith('sk-first', expect.objectContaining({ size: '3840x2160', quality: 'high' }), undefined)
    expect(putTask).toHaveBeenCalledWith(expect.objectContaining({ params: expect.objectContaining({ size: '3840x2160', quality: 'high' }) }))
    expect(JSON.parse(localStorage.getItem('image-workbench-preferences')!)).toMatchObject({ size: '3840x2160', quality: 'high' })
  })

  it('retains composer parameters when opening and saving unchanged preferences', async () => {
    vi.mocked(listImageModels).mockResolvedValue([{ id: 'gpt-image-2' }])
    vi.mocked(submitImageTask).mockResolvedValue({ id: 'task', task_id: 'task', status: 'queued' })
    await mountWorkbench()
    await wrapper!.get('.toolbar-pill--button').trigger('click')
    wrapper!.findComponent({ name: 'SizePickerModal' }).vm.$emit('select', '2048x1152')
    await wrapper!.get('.toolbar-pill--model').trigger('click')
    await wrapper!.findAll('.settings-tabs button')[1].trigger('click')
    await wrapper!.get('.settings-form .primary-button').trigger('click')
    await wrapper!.get('textarea').setValue('test image')
    await wrapper!.get('form').trigger('submit')
    await flushPromises()
    expect(submitImageTask).toHaveBeenCalledWith('sk-first', expect.objectContaining({ size: '2048x1152' }), undefined)
  })

  it('uses the same parameter snapshot for synchronous fallback', async () => {
    vi.mocked(listImageModels).mockResolvedValue([{ id: 'gpt-image-2' }])
    vi.mocked(submitImageTask).mockRejectedValue(Object.assign(new Error('async image tasks are not enabled'), { status: 404 }))
    vi.mocked(generateImage).mockResolvedValue([])
    await mountWorkbench()
    await wrapper!.get('.toolbar-pill--button').trigger('click')
    wrapper!.findComponent({ name: 'SizePickerModal' }).vm.$emit('select', '2048x1152')
    await wrapper!.get('textarea').setValue('test image')
    await wrapper!.get('form').trigger('submit')
    await flushPromises()
    expect(generateImage).toHaveBeenCalledWith('sk-first', expect.objectContaining({ size: '2048x1152' }), undefined)
    expect(vi.mocked(generateImage).mock.calls.at(-1)?.[1]).toEqual(vi.mocked(submitImageTask).mock.calls.at(-1)?.[1])
  })

  it('keeps transparency and output format compatible in either selection order', async () => {
    vi.mocked(listImageModels).mockResolvedValue([{ id: 'gpt-image-2' }])
    await mountWorkbench()
    await wrapper!.get('.toolbar-pill--toggle').trigger('click')
    const format = wrapper!.findComponent('[aria-label="imageWorkbench.format"]')
    const background = wrapper!.findComponent('[aria-label="imageWorkbench.transparent"]')
    expect(wrapper!.find('.composer__advanced input').exists()).toBe(false)
    await background.setValue('transparent')
    await format.setValue('jpeg')
    expect(background.props('modelValue')).toBe('opaque')
    expect(wrapper!.find('.composer__advanced input').exists()).toBe(true)
    await background.setValue('transparent')
    expect(format.props('modelValue')).toBe('png')
  })

  it.each([
    ['jpeg', '/9j/test'],
    ['webp', 'UklGRtest'],
    ['png', 'iVBORtest'],
  ])('preserves the returned %s format without changing requested parameters', async (format, encoded) => {
    vi.mocked(listImageModels).mockResolvedValue([{ id: 'gpt-image-2' }])
    vi.mocked(submitImageTask).mockResolvedValue({ id: 'task', task_id: 'task', status: 'completed', result: { data: [{ b64_json: encoded }] } })
    vi.stubGlobal('Image', class {
      naturalWidth = 1672
      naturalHeight = 941
      onload?: () => void
      set src(_: string) { queueMicrotask(() => this.onload?.()) }
    })
    await mountWorkbench()
    await wrapper!.get('textarea').setValue('test image')
    await wrapper!.get('form').trigger('submit')
    await flushPromises()
    expect(putHistory).toHaveBeenCalledWith(expect.objectContaining({
      src: `data:image/${format};base64,${encoded}`, width: 1672, height: 941,
      params: expect.objectContaining({ size: '1024x1024', output_format: 'png' }),
    }))
  })
})
