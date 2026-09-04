<template>
  <BaseDialog :show="show" :title="t('admin.groups.rpmOverridesTitle')" width="wide" @close="handleClose">
    <div v-if="group" class="components-admin-group-group-rpmoverrides-modal__panel">
      <!-- 分组信息 -->
      <div class="components-admin-group-group-rpmoverrides-modal__panel-2">
        <span class="components-admin-group-group-rpmoverrides-modal__text" :class="platformColorClass">
          <PlatformIcon :platform="group.platform" size="sm" />
          {{ t('admin.groups.platforms.' + group.platform) }}
        </span>
        <span class="components-admin-group-group-rpmoverrides-modal__text-2">|</span>
        <span class="components-admin-group-group-rpmoverrides-modal__text-3">{{ group.name }}</span>
        <span class="components-admin-group-group-rpmoverrides-modal__text-2">|</span>
        <span class="components-admin-group-group-rpmoverrides-modal__text-4">
          {{ t('admin.groups.groupRpmDefault') }}: {{ group.rpm_limit || 0 }}
        </span>
      </div>

      <!-- 操作区：添加用户 -->
      <div class="components-admin-group-group-rpmoverrides-modal__panel-3">
        <h4 class="components-admin-group-group-rpmoverrides-modal__heading">
          {{ t('admin.groups.addUserRpm') }}
        </h4>
        <div class="components-admin-group-group-rpmoverrides-modal__panel-4">
          <div class="components-admin-group-group-rpmoverrides-modal__panel-5">
            <input
              v-model="searchQuery"
              type="text"
              autocomplete="off"
              class="components-admin-group-group-rpmoverrides-modal__field input"
              :placeholder="t('admin.groups.searchUserPlaceholder')"
              @input="handleSearchUsers"
              @focus="showDropdown = true"
            />
            <div
              v-if="showDropdown && searchResults.length > 0"
              class="components-admin-group-group-rpmoverrides-modal__panel-6"
            >
              <button
                v-for="user in searchResults"
                :key="user.id"
                type="button"
                class="components-admin-group-group-rpmoverrides-modal__action"
                @click="selectUser(user)"
              >
                <span class="components-admin-group-group-rpmoverrides-modal__text-2">#{{ user.id }}</span>
                <span class="components-admin-group-group-rpmoverrides-modal__text-5">{{ user.username || user.email }}</span>
                <span v-if="user.username" class="components-admin-group-group-rpmoverrides-modal__text-6">{{ user.email }}</span>
              </button>
            </div>
          </div>
          <div class="components-admin-group-group-rpmoverrides-modal__panel-7">
            <input
              v-model.number="newRpm"
              type="number"
              step="1"
              min="0"
              autocomplete="off"
              class="components-admin-group-group-rpmoverrides-modal__field hide-spinner input"
              placeholder="100"
            />
          </div>
          <button
            type="button"
            class="components-admin-group-group-rpmoverrides-modal__action-2 btn btn-primary"
            :disabled="!selectedUser || newRpm == null || newRpm < 0"
            @click="handleAddLocal"
          >
            {{ t('common.add') }}
          </button>
        </div>

        <div v-if="localEntries.length > 0" class="components-admin-group-group-rpmoverrides-modal__panel-8">
          <button
            type="button"
            :disabled="clearing"
            class="components-admin-group-group-rpmoverrides-modal__action-3"
            @click="clearAllLocal"
          >
            <Icon v-if="clearing" name="refresh" size="sm" class="components-admin-group-group-rpmoverrides-modal__icon" />
            {{ t('admin.groups.clearAll') }}
          </button>
        </div>
      </div>

      <!-- 加载状态 -->
      <LoadingState v-if="loading" variant="section" size="sm" class="components-admin-group-group-rpmoverrides-modal__panel-9" />

      <!-- 列表 -->
      <div v-else>
        <h4 class="components-admin-group-group-rpmoverrides-modal__heading">
          {{ t('admin.groups.rpmOverrides') }} ({{ localEntries.length }})
        </h4>

        <div v-if="localEntries.length === 0" class="components-admin-group-group-rpmoverrides-modal__panel-10">
          {{ t('admin.groups.noRpmOverrides') }}
        </div>

        <div v-else>
          <div class="components-admin-group-group-rpmoverrides-modal__panel-11">
            <div class="components-admin-group-group-rpmoverrides-modal__panel-12">
              <table class="components-admin-group-group-rpmoverrides-modal__table">
                <thead class="components-admin-group-group-rpmoverrides-modal__header">
                  <tr class="components-admin-group-group-rpmoverrides-modal__row">
                    <th class="components-admin-group-group-rpmoverrides-modal__heading-2">{{ t('admin.groups.columns.userEmail') }}</th>
                    <th class="components-admin-group-group-rpmoverrides-modal__heading-2">ID</th>
                    <th class="components-admin-group-group-rpmoverrides-modal__heading-2">{{ t('admin.groups.columns.userName') }}</th>
                    <th class="components-admin-group-group-rpmoverrides-modal__heading-2">{{ t('admin.groups.columns.userNotes') }}</th>
                    <th class="components-admin-group-group-rpmoverrides-modal__heading-2">{{ t('admin.groups.columns.userStatus') }}</th>
                    <th class="components-admin-group-group-rpmoverrides-modal__heading-2" :title="t('admin.groups.columns.rpmOverrideHint')">{{ t('admin.groups.columns.rpmOverride') }}</th>
                    <th class="components-admin-group-group-rpmoverrides-modal__heading-3"></th>
                  </tr>
                </thead>
                <tbody class="components-admin-group-group-rpmoverrides-modal__body">
                  <tr
                    v-for="entry in paginatedLocalEntries"
                    :key="entry.user_id"
                    class="components-admin-group-group-rpmoverrides-modal__row-2"
                  >
                    <td class="components-admin-group-group-rpmoverrides-modal__cell">{{ entry.user_email }}</td>
                    <td class="components-admin-group-group-rpmoverrides-modal__cell-2">{{ entry.user_id }}</td>
                    <td class="components-admin-group-group-rpmoverrides-modal__cell-3">{{ entry.user_name || '-' }}</td>
                    <td class="components-admin-group-group-rpmoverrides-modal__cell-4" :title="entry.user_notes">{{ entry.user_notes || '-' }}</td>
                    <td class="components-admin-group-group-rpmoverrides-modal__cell-5">
                      <span
                        :class="[
                          'components-admin-group-group-rpmoverrides-modal__text-8',
                          entry.user_status === 'active'
                            ? 'components-admin-group-group-rpmoverrides-modal__text-10'
                            : 'components-admin-group-group-rpmoverrides-modal__text-11'
                        ]"
                      >
                        {{ entry.user_status }}
                      </span>
                    </td>
                    <td class="components-admin-group-group-rpmoverrides-modal__cell-5">
                      <input
                        type="number"
                        step="1"
                        min="0"
                        autocomplete="off"
                        :value="entry.rpm_override"
                        class="components-admin-group-group-rpmoverrides-modal__field-2 hide-spinner"
                        @change="updateLocalRpm(entry.user_id, ($event.target as HTMLInputElement).value)"
                      />
                    </td>
                    <td class="components-admin-group-group-rpmoverrides-modal__cell-6">
                      <button
                        type="button"
                        class="components-admin-group-group-rpmoverrides-modal__action-4"
                        @click="removeLocal(entry.user_id)"
                      >
                        <Icon name="trash" size="sm" />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <Pagination
            :total="localEntries.length"
            :page="currentPage"
            :page-size="pageSize"
            @update:page="currentPage = $event"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </div>

      <!-- 底部 -->
      <div class="components-admin-group-group-rpmoverrides-modal__panel-13">
        <template v-if="isDirty">
          <span class="components-admin-group-group-rpmoverrides-modal__text-7">{{ t('admin.groups.unsavedChanges') }}</span>
          <button
            type="button"
            class="components-admin-group-group-rpmoverrides-modal__action-5"
            @click="handleCancel"
          >
            {{ t('admin.groups.revertChanges') }}
          </button>
        </template>
        <div class="components-admin-group-group-rpmoverrides-modal__panel-14">
          <button type="button" class="components-admin-group-group-rpmoverrides-modal__action-6 btn btn-sm" @click="handleClose">
            {{ t('common.close') }}
          </button>
          <button
            v-if="isDirty"
            type="button"
            class="components-admin-group-group-rpmoverrides-modal__action-6 btn btn-primary btn-sm"
            :disabled="saving"
            @click="handleSave"
          >
            <Icon v-if="saving" name="refresh" size="sm" class="components-admin-group-group-rpmoverrides-modal__icon-3" />
            {{ t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'

import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { GroupRPMOverrideEntry } from '@/api/admin/groups'
import type { AdminGroup, AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'

interface LocalEntry extends GroupRPMOverrideEntry {}

const props = defineProps<{
  show: boolean
  group: AdminGroup | null
}>()

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
const serverEntries = ref<GroupRPMOverrideEntry[]>([])
const localEntries = ref<LocalEntry[]>([])
const searchQuery = ref('')
const searchResults = ref<AdminUser[]>([])
const showDropdown = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const newRpm = ref<number | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)

let searchTimeout: ReturnType<typeof setTimeout>

const platformColorClass = computed(() => {
  switch (props.group?.platform) {
    case 'anthropic': return 'components-admin-group-group-rpmoverrides-modal__state'
    case 'openai': return 'components-admin-group-group-rpmoverrides-modal__state-2'
    default: return 'components-admin-group-group-rpmoverrides-modal__state-4'
  }
})

const isDirty = computed(() => {
  if (localEntries.value.length !== serverEntries.value.length) return true
  const serverMap = new Map(serverEntries.value.map(e => [e.user_id, e.rpm_override]))
  return localEntries.value.some(e => serverMap.get(e.user_id) !== e.rpm_override)
})

const paginatedLocalEntries = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return localEntries.value.slice(start, start + pageSize.value)
})

const cloneEntries = (entries: GroupRPMOverrideEntry[]): LocalEntry[] => {
  return entries.map(e => ({ ...e }))
}

const loadEntries = async () => {
  if (!props.group) return
  loading.value = true
  try {
    serverEntries.value = await adminAPI.groups.getGroupRPMOverrides(props.group.id)
    localEntries.value = cloneEntries(serverEntries.value)
    adjustPage()
  } catch (error) {
    appStore.showError(t('admin.groups.failedToLoad'))
    console.error('Error loading RPM overrides:', error)
  } finally {
    loading.value = false
  }
}

const adjustPage = () => {
  const totalPages = Math.max(1, Math.ceil(localEntries.value.length / pageSize.value))
  if (currentPage.value > totalPages) currentPage.value = totalPages
}

watch(() => props.show, (val) => {
  if (val && props.group) {
    currentPage.value = 1
    searchQuery.value = ''
    searchResults.value = []
    selectedUser.value = null
    newRpm.value = null
    loadEntries()
  }
})

const handlePageSizeChange = (newSize: number) => {
  pageSize.value = newSize
  currentPage.value = 1
}

const handleSearchUsers = () => {
  clearTimeout(searchTimeout)
  selectedUser.value = null
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    showDropdown.value = false
    return
  }
  searchTimeout = setTimeout(async () => {
    try {
      const res = await adminAPI.users.list(1, 10, { search: searchQuery.value.trim() })
      searchResults.value = res.items
      showDropdown.value = true
    } catch {
      searchResults.value = []
    }
  }, 300)
}

const selectUser = (user: AdminUser) => {
  selectedUser.value = user
  searchQuery.value = user.email
  showDropdown.value = false
  searchResults.value = []
}

const handleAddLocal = () => {
  if (!selectedUser.value || newRpm.value == null || newRpm.value < 0) return
  const user = selectedUser.value
  const idx = localEntries.value.findIndex(e => e.user_id === user.id)
  const entry: LocalEntry = {
    user_id: user.id,
    user_name: user.username || '',
    user_email: user.email,
    user_notes: user.notes || '',
    user_status: user.status || 'active',
    rpm_override: newRpm.value
  }
  if (idx >= 0) {
    localEntries.value[idx] = entry
  } else {
    localEntries.value.push(entry)
  }
  searchQuery.value = ''
  selectedUser.value = null
  newRpm.value = null
  adjustPage()
}

const updateLocalRpm = (userId: string, value: string) => {
  const num = parseInt(value, 10)
  if (isNaN(num) || num < 0) return
  const entry = localEntries.value.find(e => e.user_id === userId)
  if (entry) entry.rpm_override = num
}

const removeLocal = (userId: string) => {
  localEntries.value = localEntries.value.filter(e => e.user_id !== userId)
  adjustPage()
}

const clearing = ref(false)
const clearAllLocal = async () => {
  if (!props.group || clearing.value) return
  clearing.value = true
  try {
    await adminAPI.groups.clearGroupRPMOverrides(props.group.id)
    localEntries.value = []
    serverEntries.value = []
    appStore.showSuccess(t('admin.groups.rpmSaved'))
  } catch (error) {
    appStore.showError(t('admin.groups.failedToSave'))
    console.error('Error clearing RPM overrides:', error)
  } finally {
    clearing.value = false
  }
}

const handleCancel = () => {
  localEntries.value = cloneEntries(serverEntries.value)
  adjustPage()
}

const handleSave = async () => {
  if (!props.group) return
  saving.value = true
  try {
    const entries = localEntries.value.map(e => ({
      user_id: e.user_id,
      rpm_override: e.rpm_override
    }))
    await adminAPI.groups.batchSetGroupRPMOverrides(props.group.id, entries)
    appStore.showSuccess(t('admin.groups.rpmSaved'))
    emit('success')
    emit('close')
  } catch (error) {
    appStore.showError(t('admin.groups.failedToSave'))
    console.error('Error saving RPM overrides:', error)
  } finally {
    saving.value = false
  }
}

const handleClose = () => {
  if (isDirty.value) {
    localEntries.value = cloneEntries(serverEntries.value)
  }
  emit('close')
}

const handleClickOutside = () => { showDropdown.value = false }
if (typeof document !== 'undefined') {
  document.addEventListener('click', handleClickOutside)
}
</script>

<style scoped>
.hide-spinner::-webkit-outer-spin-button,
.hide-spinner::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.hide-spinner {
  -moz-appearance: textfield;
}
</style>
