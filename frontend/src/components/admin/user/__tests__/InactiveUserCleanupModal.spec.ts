import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import InactiveUserCleanupModal from '../InactiveUserCleanupModal.vue'

const { previewInactiveUsers, permanentlyDeleteInactiveUsers, showError, showSuccess } = vi.hoisted(() => ({
  previewInactiveUsers: vi.fn(),
  permanentlyDeleteInactiveUsers: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      previewInactiveUsers,
      permanentlyDeleteInactiveUsers
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const BaseDialogStub = {
  props: ['show', 'title'],
  emits: ['close'],
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
}

describe('InactiveUserCleanupModal', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-09-03T12:00:00+08:00'))
    previewInactiveUsers.mockReset()
    permanentlyDeleteInactiveUsers.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('previews filters and permanently deletes only after exact confirmation', async () => {
    previewInactiveUsers.mockResolvedValue({
      total: 2,
      total_balance: 0,
      total_usage_7d: 0,
      generated_at: '2026-09-03T04:00:00Z',
      snapshot_token: 'snapshot-2',
      items: [
        {
          id: '01990f00-0000-7000-8000-000000000001',
          email: 'inactive@example.com',
          balance: 0,
          last_used_at: null,
          usage_7d: 0,
          created_at: '2026-07-01T00:00:00Z'
        }
      ]
    })
    permanentlyDeleteInactiveUsers.mockResolvedValue({ deleted: 2 })

    const wrapper = mount(InactiveUserCleanupModal, {
      props: { show: true },
      global: { stubs: { BaseDialog: BaseDialogStub } }
    })

    await wrapper.get('[data-test="inactive-preview"]').trigger('click')
    await flushPromises()

    expect(previewInactiveUsers).toHaveBeenCalledWith({
      max_balance: 0,
      max_usage_7d: 0,
      last_used_before: new Date('2026-08-04T12:00:00+08:00').toISOString()
    })
    expect(wrapper.get('[data-test="inactive-delete"]').attributes('disabled')).toBeDefined()

    await wrapper.get('[data-test="inactive-confirmation"]').setValue('DELETE 2 USERS')
    await wrapper.get('[data-test="inactive-delete"]').trigger('click')
    await flushPromises()

    expect(permanentlyDeleteInactiveUsers).toHaveBeenCalledWith({
      max_balance: 0,
      max_usage_7d: 0,
      last_used_before: new Date('2026-08-04T12:00:00+08:00').toISOString(),
      expected_count: 2,
      snapshot_token: 'snapshot-2',
      confirmation: 'DELETE 2 USERS'
    })
    expect(wrapper.emitted('success')).toEqual([[2]])
    expect(wrapper.emitted('close')).toHaveLength(1)
  })
})
