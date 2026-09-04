<template>
  <BaseDialog :show="show" :title="t('admin.groups.rateMultipliersTitle')" width="wide" @close="handleClose">
    <div v-if="group" class="components-admin-group-group-rate-multipliers-modal__panel">
      <!-- 分组信息 -->
      <div class="components-admin-group-group-rate-multipliers-modal__panel-2">
        <span class="components-admin-group-group-rate-multipliers-modal__text" :class="platformColorClass">
          <PlatformIcon :platform="group.platform" size="sm" />
          {{ t('admin.groups.platforms.' + group.platform) }}
        </span>
        <span class="components-admin-group-group-rate-multipliers-modal__text-2">|</span>
        <span class="components-admin-group-group-rate-multipliers-modal__text-3">{{ group.name }}</span>
        <span class="components-admin-group-group-rate-multipliers-modal__text-2">|</span>
        <span class="components-admin-group-group-rate-multipliers-modal__text-4">
          {{ t('admin.groups.columns.rateMultiplier') }}: {{ group.rate_multiplier }}x
        </span>
      </div>

      <!-- 操作区 -->
      <div class="components-admin-group-group-rate-multipliers-modal__panel-3">
        <!-- 添加用户 -->
        <h4 class="components-admin-group-group-rate-multipliers-modal__heading">
          {{ t('admin.groups.addUserRate') }}
        </h4>
        <div class="components-admin-group-group-rate-multipliers-modal__panel-4">
          <div class="components-admin-group-group-rate-multipliers-modal__panel-5">
            <input
              v-model="searchQuery"
              type="text"
              autocomplete="off"
              class="components-admin-group-group-rate-multipliers-modal__field input"
              :placeholder="t('admin.groups.searchUserPlaceholder')"
              @input="handleSearchUsers"
              @focus="showDropdown = true"
            />
            <div
              v-if="showDropdown && searchResults.length > 0"
              class="components-admin-group-group-rate-multipliers-modal__panel-6"
            >
              <button
                v-for="user in searchResults"
                :key="user.id"
                type="button"
                class="components-admin-group-group-rate-multipliers-modal__action"
                @click="selectUser(user)"
              >
                <span class="components-admin-group-group-rate-multipliers-modal__text-2">#{{ user.id }}</span>
                <span class="components-admin-group-group-rate-multipliers-modal__text-5">{{ user.username || user.email }}</span>
                <span v-if="user.username" class="components-admin-group-group-rate-multipliers-modal__text-6">{{ user.email }}</span>
              </button>
            </div>
          </div>
          <div class="components-admin-group-group-rate-multipliers-modal__panel-7">
            <input
              v-model.number="newRate"
              type="number"
              step="0.001"
              min="0"
              autocomplete="off"
              class="components-admin-group-group-rate-multipliers-modal__field hide-spinner input"
              placeholder="1.0"
            />
          </div>
          <button
            type="button"
            class="components-admin-group-group-rate-multipliers-modal__action-2 btn btn-primary"
            :disabled="!selectedUser || !newRate"
            @click="handleAddLocal"
          >
            {{ t('common.add') }}
          </button>
        </div>

        <!-- 批量调整 + 全部清空 -->
        <div v-if="localEntries.length > 0" class="components-admin-group-group-rate-multipliers-modal__panel-8">
          <span class="components-admin-group-group-rate-multipliers-modal__text-7">{{ t('admin.groups.batchAdjust') }}</span>
          <div class="components-admin-group-group-rate-multipliers-modal__panel-9">
            <span class="components-admin-group-group-rate-multipliers-modal__text-6">×</span>
            <input
              v-model.number="batchFactor"
              type="number"
              step="0.1"
              min="0"
              autocomplete="off"
              class="components-admin-group-group-rate-multipliers-modal__field-2 hide-spinner"
              placeholder="0.5"
            />
            <button
              type="button"
              class="components-admin-group-group-rate-multipliers-modal__action-3 btn btn-primary btn-sm"
              :disabled="!batchFactor || batchFactor <= 0"
              @click="applyBatchFactor"
            >
              {{ t('admin.groups.applyMultiplier') }}
            </button>
          </div>
          <div class="components-admin-group-group-rate-multipliers-modal__panel-10">
            <button
              type="button"
              class="components-admin-group-group-rate-multipliers-modal__action-4"
              @click="clearAllLocal"
            >
              {{ t('admin.groups.clearAll') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 加载状态 -->
      <LoadingState v-if="loading" variant="section" size="sm" class="components-admin-group-group-rate-multipliers-modal__panel-11" />

      <!-- 已设置的用户列表 -->
      <div v-else>
        <h4 class="components-admin-group-group-rate-multipliers-modal__heading">
          {{ t('admin.groups.rateMultipliers') }} ({{ localEntries.length }})
        </h4>

        <div v-if="localEntries.length === 0" class="components-admin-group-group-rate-multipliers-modal__panel-12">
          {{ t('admin.groups.noRateMultipliers') }}
        </div>

        <div v-else>
          <!-- 表格 -->
          <div class="components-admin-group-group-rate-multipliers-modal__panel-13">
            <div class="components-admin-group-group-rate-multipliers-modal__panel-14">
              <table class="components-admin-group-group-rate-multipliers-modal__table">
                <thead class="components-admin-group-group-rate-multipliers-modal__header">
                  <tr class="components-admin-group-group-rate-multipliers-modal__row">
                    <th class="components-admin-group-group-rate-multipliers-modal__heading-2">{{ t('admin.groups.columns.userEmail') }}</th>
                    <th class="components-admin-group-group-rate-multipliers-modal__heading-2">ID</th>
                    <th class="components-admin-group-group-rate-multipliers-modal__heading-2">{{ t('admin.groups.columns.userName') }}</th>
                    <th class="components-admin-group-group-rate-multipliers-modal__heading-2">{{ t('admin.groups.columns.userNotes') }}</th>
                    <th class="components-admin-group-group-rate-multipliers-modal__heading-2">{{ t('admin.groups.columns.userStatus') }}</th>
                    <th class="components-admin-group-group-rate-multipliers-modal__heading-2">{{ t('admin.groups.columns.rateMultiplier') }}</th>
                    <th v-if="showFinalRate" class="components-admin-group-group-rate-multipliers-modal__heading-3">{{ t('admin.groups.finalRate') }}</th>
                    <th class="components-admin-group-group-rate-multipliers-modal__heading-4"></th>
                  </tr>
                </thead>
                <tbody class="components-admin-group-group-rate-multipliers-modal__body">
                  <tr
                    v-for="entry in paginatedLocalEntries"
                    :key="entry.user_id"
                    class="components-admin-group-group-rate-multipliers-modal__row-2"
                  >
                    <td class="components-admin-group-group-rate-multipliers-modal__cell">{{ entry.user_email }}</td>
                    <td class="components-admin-group-group-rate-multipliers-modal__cell-2">{{ entry.user_id }}</td>
                    <td class="components-admin-group-group-rate-multipliers-modal__cell-3">{{ entry.user_name || '-' }}</td>
                    <td class="components-admin-group-group-rate-multipliers-modal__cell-4" :title="entry.user_notes">{{ entry.user_notes || '-' }}</td>
                    <td class="components-admin-group-group-rate-multipliers-modal__cell-5">
                      <span
                        :class="[
                          'components-admin-group-group-rate-multipliers-modal__text-9',
                          entry.user_status === 'active'
                            ? 'components-admin-group-group-rate-multipliers-modal__text-11'
                            : 'components-admin-group-group-rate-multipliers-modal__text-12'
                        ]"
                      >
                        {{ entry.user_status }}
                      </span>
                    </td>
                    <td class="components-admin-group-group-rate-multipliers-modal__cell-5">
                      <input
                        type="number"
                        step="0.001"
                        min="0.001"
                        autocomplete="off"
                        :value="entry.rate_multiplier ?? ''"
                        :placeholder="String(props.group?.rate_multiplier ?? 1)"
                        class="components-admin-group-group-rate-multipliers-modal__field-3 hide-spinner"
                        @change="updateLocalRate(entry.user_id, ($event.target as HTMLInputElement).value)"
                      />
                    </td>
                    <td v-if="showFinalRate" class="components-admin-group-group-rate-multipliers-modal__cell-6">
                      {{ computeFinalRate(entry.rate_multiplier) }}
                    </td>
                    <td class="components-admin-group-group-rate-multipliers-modal__cell-7">
                      <button
                        type="button"
                        class="components-admin-group-group-rate-multipliers-modal__action-5"
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

          <!-- 分页 -->
          <Pagination
            :total="localEntries.length"
            :page="currentPage"
            :page-size="pageSize"
            @update:page="currentPage = $event"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </div>

      <!-- 底部操作栏 -->
      <div class="components-admin-group-group-rate-multipliers-modal__panel-15">
        <!-- 左侧：未保存提示 + 撤销 -->
        <template v-if="isDirty">
          <span class="components-admin-group-group-rate-multipliers-modal__text-8">{{ t('admin.groups.unsavedChanges') }}</span>
          <button
            type="button"
            class="components-admin-group-group-rate-multipliers-modal__action-6"
            @click="handleCancel"
          >
            {{ t('admin.groups.revertChanges') }}
          </button>
        </template>
        <!-- 右侧：关闭 / 保存 -->
        <div class="components-admin-group-group-rate-multipliers-modal__panel-16">
          <button type="button" class="components-admin-group-group-rate-multipliers-modal__action-7 btn btn-sm" @click="handleClose">
            {{ t('common.close') }}
          </button>
          <button
            v-if="isDirty"
            type="button"
            class="components-admin-group-group-rate-multipliers-modal__action-7 btn btn-primary btn-sm"
            :disabled="saving"
            @click="handleSave"
          >
            <Icon v-if="saving" name="refresh" size="sm" class="components-admin-group-group-rate-multipliers-modal__icon-2" />
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
import type { GroupRateMultiplierEntry } from '@/api/admin/groups'
import type { AdminGroup, AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'

interface LocalEntry extends GroupRateMultiplierEntry {}

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
const serverEntries = ref<GroupRateMultiplierEntry[]>([])
const localEntries = ref<LocalEntry[]>([])
const searchQuery = ref('')
const searchResults = ref<AdminUser[]>([])
const showDropdown = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const newRate = ref<number | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const batchFactor = ref<number | null>(null)

let searchTimeout: ReturnType<typeof setTimeout>

const platformColorClass = computed(() => {
  switch (props.group?.platform) {
    case 'anthropic': return 'components-admin-group-group-rate-multipliers-modal__state'
    case 'openai': return 'components-admin-group-group-rate-multipliers-modal__state-2'
    default: return 'components-admin-group-group-rate-multipliers-modal__state-4'
  }
})

// 是否显示"最终倍率"预览列
const showFinalRate = computed(() => {
  return batchFactor.value != null && batchFactor.value > 0 && batchFactor.value !== 1
})

// 计算最终倍率预览
const computeFinalRate = (rate: number | null | undefined) => {
  const base = rate ?? props.group?.rate_multiplier ?? 1
  if (!batchFactor.value) return base
  return parseFloat((base * batchFactor.value).toFixed(6))
}

// 检测是否有未保存的修改
const isDirty = computed(() => {
  if (localEntries.value.length !== serverEntries.value.length) return true
  const serverMap = new Map(serverEntries.value.map(e => [e.user_id, e.rate_multiplier ?? null]))
  return localEntries.value.some(e => serverMap.get(e.user_id) !== (e.rate_multiplier ?? null))
})

const paginatedLocalEntries = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return localEntries.value.slice(start, start + pageSize.value)
})

const cloneEntries = (entries: GroupRateMultiplierEntry[]): LocalEntry[] => {
  return entries.map(e => ({ ...e }))
}

const loadEntries = async () => {
  if (!props.group) return
  loading.value = true
  try {
    const raw = await adminAPI.groups.getGroupRateMultipliers(props.group.id)
    // 仅显示已设置 rate_multiplier 的条目；rpm_override 在另一个弹窗管理，保留不动
    serverEntries.value = raw.filter(e => e.rate_multiplier != null)
    localEntries.value = cloneEntries(serverEntries.value)
    adjustPage()
  } catch (error) {
    appStore.showError(t('admin.groups.failedToLoad'))
    console.error('Error loading group rate multipliers:', error)
  } finally {
    loading.value = false
  }
}

const adjustPage = () => {
  const totalPages = Math.max(1, Math.ceil(localEntries.value.length / pageSize.value))
  if (currentPage.value > totalPages) {
    currentPage.value = totalPages
  }
}

watch(() => props.show, (val) => {
  if (val && props.group) {
    currentPage.value = 1
    batchFactor.value = null
    searchQuery.value = ''
    searchResults.value = []
    selectedUser.value = null
    newRate.value = null
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

// 本地添加（或覆盖已有用户）
const handleAddLocal = () => {
  if (!selectedUser.value || !newRate.value) return
  const user = selectedUser.value
  const idx = localEntries.value.findIndex(e => e.user_id === user.id)
  const entry: LocalEntry = {
    user_id: user.id,
    user_name: user.username || '',
    user_email: user.email,
    user_notes: user.notes || '',
    user_status: user.status || 'active',
    rate_multiplier: newRate.value,
    rpm_override: null
  }
  if (idx >= 0) {
    localEntries.value[idx] = entry
  } else {
    localEntries.value.push(entry)
  }
  searchQuery.value = ''
  selectedUser.value = null
  newRate.value = null
  adjustPage()
}

// 本地修改倍率
const updateLocalRate = (userId: string, value: string) => {
  const entry = localEntries.value.find(e => e.user_id === userId)
  if (!entry) return
  if (value.trim() === '') {
    entry.rate_multiplier = null
    return
  }
  const num = parseFloat(value)
  if (isNaN(num)) return
  entry.rate_multiplier = num
}

// 本地删除
const removeLocal = (userId: string) => {
  localEntries.value = localEntries.value.filter(e => e.user_id !== userId)
  adjustPage()
}

// 批量乘数应用到本地
const applyBatchFactor = () => {
  if (!batchFactor.value || batchFactor.value <= 0) return
  for (const entry of localEntries.value) {
    if (entry.rate_multiplier != null) {
      entry.rate_multiplier = parseFloat((entry.rate_multiplier * batchFactor.value).toFixed(6))
    }
  }
  batchFactor.value = null
}

// 本地清空
const clearAllLocal = () => {
  localEntries.value = []
}

// 取消：恢复到服务器数据
const handleCancel = () => {
  localEntries.value = cloneEntries(serverEntries.value)
  batchFactor.value = null
  adjustPage()
}

// 保存：一次性提交所有数据（只提交 rate_multiplier；rpm_override 由独立弹窗管理）
const handleSave = async () => {
  if (!props.group) return
  saving.value = true
  try {
    const entries = localEntries.value
      .filter(e => e.rate_multiplier != null)
      .map(e => ({
        user_id: e.user_id,
        rate_multiplier: e.rate_multiplier as number
      }))
    await adminAPI.groups.batchSetGroupRateMultipliers(props.group.id, entries)
    appStore.showSuccess(t('admin.groups.rateSaved'))
    emit('success')
    emit('close')
  } catch (error) {
    appStore.showError(t('admin.groups.failedToSave'))
    console.error('Error saving rate multipliers:', error)
  } finally {
    saving.value = false
  }
}

// 关闭时如果有未保存修改，先恢复
const handleClose = () => {
  if (isDirty.value) {
    localEntries.value = cloneEntries(serverEntries.value)
  }
  emit('close')
}

// 点击外部关闭下拉
const handleClickOutside = () => {
  showDropdown.value = false
}

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
