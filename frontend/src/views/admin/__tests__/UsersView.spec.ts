import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import type { AdminUser } from '@/types'
import UsersView from '../UsersView.vue'

const {
  listUsers,
  getAllGroups,
  getBatchUsersUsage,
  listEnabledDefinitions,
  getBatchUserAttributes,
  bulkDeleteUsers,
  showError,
  showSuccess
} = vi.hoisted(() => ({
  listUsers: vi.fn(),
  getAllGroups: vi.fn(),
  getBatchUsersUsage: vi.fn(),
  listEnabledDefinitions: vi.fn(),
  getBatchUserAttributes: vi.fn(),
  bulkDeleteUsers: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      list: listUsers,
      toggleStatus: vi.fn(),
      delete: vi.fn(),
      bulkDelete: bulkDeleteUsers
    },
    groups: {
      getAll: getAllGroups
    },
    dashboard: {
      getBatchUsersUsage
    },
    userAttributes: {
      listEnabledDefinitions,
      getBatchUserAttributes
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
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

const createAdminUser = (overrides: Partial<AdminUser> = {}): AdminUser => ({
  id: 'user-42',
  username: 'scoped-user',
  email: 'scoped@example.com',
  role: 'user',
  available_balance: 0,
  recharge_balance: 0,
  bonus_balance: 0,
  concurrency: 1,
  status: 'active',
  allowed_groups: [],
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: '2026-04-17T00:00:00Z',
  updated_at: '2026-04-17T00:00:00Z',
  notes: '',
  last_active_at: '2026-04-16T02:00:00Z',
  last_used_at: '2026-04-17T02:00:00Z',
  current_concurrency: 0,
  ...overrides
})

const DataTableStub = {
  props: ['columns', 'data', 'selectedKeys'],
  emits: ['sort', 'update:selectedKeys'],
  template: `
    <div>
      <div data-test="columns">{{ columns.map(col => col.key).join(',') }}</div>
      <div data-test="row-order">{{ data.map(row => row.email).join(',') }}</div>
      <div data-test="selected-keys">{{ (selectedKeys || []).join(',') }}</div>
      <button data-test="sort-last-used" @click="$emit('sort', 'last_used_at', 'desc')">sort</button>
      <button
        v-for="row in data"
        :key="'select-' + row.id"
        :data-test="'select-' + row.id"
        @click="$emit('update:selectedKeys', Array.from(new Set([...(selectedKeys || []), row.id])))"
      >
        select
      </button>
      <template v-for="col in columns" :key="col.key">
        <slot :name="'header-' + col.key" :column="col" />
      </template>
      <div v-for="row in data" :key="row.id">
        <slot name="cell-last_used_at" :value="row.last_used_at" :row="row" />
      </div>
    </div>
  `
}

const PaginationStub = {
  emits: ['update:page'],
  template: '<button data-test="next-page" @click="$emit(\'update:page\', 2)">next</button>'
}

const BulkEditUserModalStub = {
  props: ['show', 'selectedIds'],
  emits: ['close', 'success'],
  template: `
    <div v-if="show" data-test="bulk-modal">
      <span data-test="bulk-modal-ids">{{ selectedIds.join(',') }}</span>
      <button data-test="bulk-success" @click="$emit('success', selectedIds.length)">success</button>
    </div>
  `
}

const ConfirmDialogStub = {
  name: 'ConfirmDialog',
  props: ['show', 'title', 'message', 'confirmText', 'danger'],
  emits: ['confirm', 'cancel'],
  template: `
    <div v-if="show" data-test="confirm-dialog">
      <span data-test="confirm-dialog-title">{{ title }}</span>
      <span data-test="confirm-dialog-message">{{ message }}</span>
      <button data-test="confirm-dialog-confirm" @click="$emit('confirm')">confirm</button>
      <button data-test="confirm-dialog-cancel" @click="$emit('cancel')">cancel</button>
    </div>
  `
}

describe('admin UsersView', () => {
  beforeEach(() => {
    vi.useRealTimers()
    localStorage.clear()

    listUsers.mockReset()
    getAllGroups.mockReset()
    getBatchUsersUsage.mockReset()
    listEnabledDefinitions.mockReset()
    getBatchUserAttributes.mockReset()
    bulkDeleteUsers.mockReset()
    showError.mockReset()
    showSuccess.mockReset()

    listUsers.mockResolvedValue({
      items: [createAdminUser()],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
    getAllGroups.mockResolvedValue([])
    getBatchUsersUsage.mockResolvedValue({ stats: {} })
    listEnabledDefinitions.mockResolvedValue([])
    getBatchUserAttributes.mockResolvedValue({ values: {} })
    bulkDeleteUsers.mockResolvedValue({ succeededIds: [], failures: [] })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows active, used, and created activity columns in order and requests last_used_at sort', async () => {
    const wrapper = mount(UsersView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          EmptyState: true,
          GroupBadge: true,
          Select: true,
          UserAttributesConfigModal: true,
          UserConcurrencyCell: true,
          UserCreateModal: true,
          UserEditModal: true,
          BulkEditUserModal: BulkEditUserModalStub,
          UserPlatformQuotaModal: true,
          UserApiKeysModal: true,
          UserAllowedGroupsModal: true,
          UserBalanceModal: true,
          UserBalanceHistoryModal: true,
          GroupReplaceModal: true,
          Icon: true,
          Teleport: true
        }
      }
    })

    await flushPromises()

    const columns = wrapper.get('[data-test="columns"]').text()
    const visibleColumns = columns.split(',')
    expect(visibleColumns.slice(-4, -1)).toEqual(['last_active_at', 'last_used_at', 'created_at'])
    expect(visibleColumns).not.toContain('last_login_at')

    await wrapper.get('[data-test="sort-last-used"]').trigger('click')
    await flushPromises()

    expect(listUsers).toHaveBeenLastCalledWith(
      1,
      20,
      expect.objectContaining({
        sort_by: 'last_used_at',
        sort_order: 'desc'
      }),
      expect.any(Object)
    )
  })

  it('switches from usage server sort to last_used_at server sort', async () => {
    vi.useFakeTimers()
    localStorage.setItem('user-column-settings-version', '3')
    localStorage.setItem(
      'user-hidden-columns',
      JSON.stringify([
        'notes',
        'groups',
        'subscriptions',
        'concurrency',
        'usage_anthropic',
        'usage_openai',
        'balance_platform_quota'
      ])
    )

    listUsers.mockResolvedValue({
      items: [
        createAdminUser({ id: 'user-1', email: 'last-used-first@example.com' }),
        createAdminUser({ id: 'user-2', email: 'usage-first@example.com' })
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })
    getBatchUsersUsage.mockResolvedValue({
      stats: {
        'user-1': { user_id: 'user-1', today_actual_cost: 1, total_actual_cost: 1, by_platform: [] },
        'user-2': { user_id: 'user-2', today_actual_cost: 9, total_actual_cost: 9, by_platform: [] }
      }
    })

    const wrapper = mount(UsersView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          EmptyState: true,
          GroupBadge: true,
          Select: true,
          UserAttributesConfigModal: true,
          UserConcurrencyCell: true,
          UserCreateModal: true,
          UserEditModal: true,
          BulkEditUserModal: BulkEditUserModalStub,
          UserPlatformQuotaModal: true,
          UserApiKeysModal: true,
          UserAllowedGroupsModal: true,
          UserBalanceModal: true,
          UserBalanceHistoryModal: true,
          GroupReplaceModal: true,
          Icon: true,
          Teleport: true
        }
      }
    })

    await flushPromises()
    await vi.advanceTimersByTimeAsync(50)
    await flushPromises()

    expect(wrapper.get('[data-test="row-order"]').text()).toBe('last-used-first@example.com,usage-first@example.com')

    await wrapper.get('[data-test="usage-sort-trigger-usage"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-test="usage-sort-usage-today"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[data-test="row-order"]').text()).toBe('last-used-first@example.com,usage-first@example.com')
    expect(localStorage.getItem('admin-users-usage-sort')).toContain('"key":"usage"')
    expect(listUsers).toHaveBeenLastCalledWith(
      1,
      20,
      expect.objectContaining({
        sort_by: 'usage_today',
        sort_order: 'desc'
      }),
      expect.any(Object)
    )

    await wrapper.get('[data-test="sort-last-used"]').trigger('click')
    await flushPromises()

    expect(localStorage.getItem('admin-users-usage-sort')).toBeNull()
    expect(wrapper.get('[data-test="row-order"]').text()).toBe('last-used-first@example.com,usage-first@example.com')
    expect(listUsers).toHaveBeenLastCalledWith(
      1,
      20,
      expect.objectContaining({
        sort_by: 'last_used_at',
        sort_order: 'desc'
      }),
      expect.any(Object)
    )
  })

  it('keeps selected user IDs across pages and clears them after a successful bulk update', async () => {
    let refreshed = false
    listUsers.mockImplementation(async (page: number) => {
      const user = page === 2
        ? createAdminUser({
            id: 'user-43',
            email: refreshed ? 'refreshed-page-two@example.com' : 'page-two@example.com'
          })
        : createAdminUser({ id: 'user-42', email: 'page-one@example.com' })
      return {
        items: [user],
        total: 2,
        page,
        page_size: 20,
        pages: 2
      }
    })

    const wrapper = mount(UsersView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: PaginationStub,
          ConfirmDialog: ConfirmDialogStub,
          EmptyState: true,
          GroupBadge: true,
          Select: true,
          UserAttributesConfigModal: true,
          UserConcurrencyCell: true,
          UserCreateModal: true,
          UserEditModal: true,
          BulkEditUserModal: BulkEditUserModalStub,
          UserPlatformQuotaModal: true,
          UserApiKeysModal: true,
          UserAllowedGroupsModal: true,
          UserBalanceModal: true,
          UserBalanceHistoryModal: true,
          GroupReplaceModal: true,
          Icon: true,
          Teleport: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.find('[data-test="bulk-edit-limits"]').exists()).toBe(false)
    await wrapper.get('[data-test="select-user-42"]').trigger('click')
    expect(wrapper.get('[data-test="selected-keys"]').text()).toBe('user-42')
    expect(wrapper.find('[data-test="bulk-edit-limits"]').exists()).toBe(true)

    await wrapper.get('[data-test="next-page"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-test="selected-keys"]').text()).toBe('user-42')

    await wrapper.get('[data-test="select-user-43"]').trigger('click')
    expect(wrapper.get('[data-test="selected-keys"]').text()).toBe('user-42,user-43')

    await wrapper.get('[data-test="bulk-edit-limits"]').trigger('click')
    expect(wrapper.get('[data-test="bulk-modal-ids"]').text()).toBe('user-42,user-43')

    const callsBeforeSuccess = listUsers.mock.calls.length
    refreshed = true
    await wrapper.get('[data-test="bulk-success"]').trigger('click')
    await flushPromises()

    expect(listUsers.mock.calls.length).toBeGreaterThan(callsBeforeSuccess)
    expect(wrapper.get('[data-test="row-order"]').text()).toBe('refreshed-page-two@example.com')
    expect(wrapper.find('[data-test="bulk-edit-limits"]').exists()).toBe(false)
    expect(wrapper.get('[data-test="selected-keys"]').text()).toBe('')
  })

  it('keeps failed user deletions selected and retries only those string IDs', async () => {
    listUsers.mockResolvedValue({
      items: [
        createAdminUser({ id: 'user-delete', email: 'delete@example.com' }),
        createAdminUser({ id: 'user-admin', email: 'admin@example.com', role: 'admin' })
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1
    })
    bulkDeleteUsers.mockResolvedValueOnce({
      succeededIds: ['user-delete'],
      failures: [{ id: 'user-admin', error: new Error('cannot delete admin user') }]
    }).mockResolvedValueOnce({
      succeededIds: ['user-admin'],
      failures: []
    })

    const wrapper = mount(UsersView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: {
            template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
          },
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          EmptyState: true,
          GroupBadge: true,
          Select: true,
          UserAttributesConfigModal: true,
          UserConcurrencyCell: true,
          UserCreateModal: true,
          UserEditModal: true,
          BulkEditUserModal: BulkEditUserModalStub,
          ArchivedUsersModal: true,
          UserPlatformQuotaModal: true,
          UserApiKeysModal: true,
          UserAllowedGroupsModal: true,
          UserBalanceModal: true,
          UserBalanceHistoryModal: true,
          GroupReplaceModal: true,
          Icon: true,
          Teleport: true
        }
      }
    })
    await flushPromises()

    await wrapper.get('[data-test="select-user-delete"]').trigger('click')
    await wrapper.get('[data-test="select-user-admin"]').trigger('click')
    await wrapper.get('[data-test="bulk-delete-users"]').trigger('click')
    expect(wrapper.get('[data-test="confirm-dialog-title"]').text()).toBe('admin.users.bulkDelete.title')
    await wrapper.get('[data-test="confirm-dialog-confirm"]').trigger('click')
    await flushPromises()

    expect(bulkDeleteUsers).toHaveBeenNthCalledWith(1, ['user-delete', 'user-admin'])
    expect(wrapper.get('[data-test="selected-keys"]').text()).toBe('user-admin')
    expect(showSuccess).toHaveBeenCalledWith('admin.users.bulkDelete.success')
    expect(showError).toHaveBeenCalledWith('admin.users.bulkDelete.failed')

    await wrapper.get('[data-test="bulk-delete-users"]').trigger('click')
    await wrapper.get('[data-test="confirm-dialog-confirm"]').trigger('click')
    await flushPromises()

    expect(bulkDeleteUsers).toHaveBeenNthCalledWith(2, ['user-admin'])
    expect(wrapper.get('[data-test="selected-keys"]').text()).toBe('')
  })
})
