<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.archive.title')"
    width="wide"
    @close="emit('close')"
  >
    <div class="archived-users">
      <p class="archived-users__hint">{{ t('admin.users.archive.hint') }}</p>

      <div class="archived-users__toolbar">
        <input
          v-model="search"
          class="input archived-users__search"
          type="text"
          :placeholder="t('admin.users.archive.search')"
          @keyup.enter="applySearch"
        />
        <button class="btn btn-secondary" type="button" :disabled="loading" @click="applySearch">
          {{ t('common.search') }}
        </button>
        <button class="btn btn-secondary" type="button" :disabled="loading" @click="loadUsers">
          {{ t('common.refresh') }}
        </button>
      </div>

      <div v-if="loading" class="archived-users__state">{{ t('common.loading') }}</div>
      <div v-else-if="users.length === 0" class="archived-users__state">
        {{ t('admin.users.archive.empty') }}
      </div>
      <div v-else class="archived-users__table-wrap">
        <table class="archived-users__table">
          <thead>
            <tr>
              <th>{{ t('admin.users.email') }}</th>
              <th>{{ t('admin.users.columns.balance') }}</th>
              <th>{{ t('admin.users.columns.lastUsed') }}</th>
              <th>{{ t('admin.users.archive.archivedAt') }}</th>
              <th>{{ t('admin.users.columns.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>
                <strong>{{ user.email }}</strong>
                <small v-if="user.username">{{ user.username }}</small>
              </td>
              <td>{{ formatPoints(user.balance) }}</td>
              <td>{{ user.last_used_at ? formatDateTime(user.last_used_at) : '-' }}</td>
              <td>{{ user.deleted_at ? formatDateTime(user.deleted_at) : '-' }}</td>
              <td>
                <button
                  class="btn btn-primary btn-sm"
                  type="button"
                  :disabled="restoringId === user.id"
                  @click="restoreUser(user)"
                >
                  {{ restoringId === user.id ? t('admin.users.archive.restoring') : t('admin.users.archive.restore') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.pageSize"
        :show-page-size-selector="false"
        @update:page="changePage"
      />
    </div>

    <template #footer>
      <button class="btn btn-secondary" type="button" @click="emit('close')">
        {{ t('common.close') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import { useAppStore } from '@/stores/app'
import { formatDateTime, formatPoints } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: []; restored: [user: AdminUser] }>()
const { t } = useI18n()
const appStore = useAppStore()

const users = ref<AdminUser[]>([])
const search = ref('')
const appliedSearch = ref('')
const loading = ref(false)
const restoringId = ref<string | null>(null)
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const loadUsers = async () => {
  loading.value = true
  try {
    const response = await adminAPI.users.listArchived(
      pagination.page,
      pagination.pageSize,
      appliedSearch.value
    )
    users.value = response.items
    pagination.total = response.total
    if (users.value.length === 0 && pagination.page > 1) {
      pagination.page--
      await loadUsers()
    }
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.users.archive.loadFailed'))
  } finally {
    loading.value = false
  }
}

const applySearch = () => {
  appliedSearch.value = search.value.trim()
  pagination.page = 1
  void loadUsers()
}

const changePage = (page: number) => {
  pagination.page = page
  void loadUsers()
}

const restoreUser = async (user: AdminUser) => {
  restoringId.value = user.id
  try {
    const restored = await adminAPI.users.restoreArchivedUser(user.id)
    emit('restored', restored)
    await loadUsers()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.users.archive.restoreFailed'))
  } finally {
    restoringId.value = null
  }
}

watch(
  () => props.show,
  (show) => {
    if (!show) return
    pagination.page = 1
    void loadUsers()
  }
)
</script>

<style scoped>
.archived-users { display: flex; flex-direction: column; gap: 1rem; }
.archived-users__hint { margin: 0; color: var(--color-text-secondary); }
.archived-users__toolbar { display: flex; gap: .75rem; align-items: center; }
.archived-users__search { flex: 1; }
.archived-users__state { padding: 3rem 1rem; text-align: center; color: var(--color-text-secondary); }
.archived-users__table-wrap { overflow-x: auto; border: 1px solid var(--color-border); border-radius: .75rem; }
.archived-users__table { width: 100%; border-collapse: collapse; }
.archived-users__table th, .archived-users__table td { padding: .85rem 1rem; text-align: left; border-bottom: 1px solid var(--color-border); white-space: nowrap; }
.archived-users__table th { color: var(--color-text-secondary); font-size: var(--type-caption-size); }
.archived-users__table tbody tr:last-child td { border-bottom: 0; }
.archived-users__table td:first-child { white-space: normal; min-width: 14rem; }
.archived-users__table small { display: block; margin-top: .2rem; color: var(--color-text-secondary); }
@media (max-width: 640px) { .archived-users__toolbar { align-items: stretch; flex-direction: column; } }
</style>
