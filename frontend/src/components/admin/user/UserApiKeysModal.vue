<template>
  <BaseDialog :show="show" :title="t('admin.users.userApiKeys')" width="wide" @close="handleClose">
    <div v-if="user" class="components-admin-user-user-api-keys-modal__panel">
      <div class="components-admin-user-user-api-keys-modal__panel-2">
        <div class="components-admin-user-user-api-keys-modal__panel-3">
          <span class="components-admin-user-user-api-keys-modal__text">{{ user.email.charAt(0).toUpperCase() }}</span>
        </div>
        <div><p class="components-admin-user-user-api-keys-modal__description">{{ user.email }}</p><p class="components-admin-user-user-api-keys-modal__description-2">{{ user.username }}</p></div>
      </div>
      <LoadingState v-if="loading" variant="section" size="sm" class="components-admin-user-user-api-keys-modal__panel-4" />
      <div v-else-if="apiKeys.length === 0" class="components-admin-user-user-api-keys-modal__panel-5"><p class="components-admin-user-user-api-keys-modal__description-3">{{ t('admin.users.noApiKeys') }}</p></div>
      <div v-else ref="scrollContainerRef" class="components-admin-user-user-api-keys-modal__panel-6" @scroll="closeGroupSelector">
        <div v-for="key in apiKeys" :key="key.id" class="components-admin-user-user-api-keys-modal__panel-7">
          <div class="components-admin-user-user-api-keys-modal__panel-8">
            <div class="components-admin-user-user-api-keys-modal__panel-9">
              <div class="components-admin-user-user-api-keys-modal__panel-10"><span class="components-admin-user-user-api-keys-modal__description">{{ key.name }}</span><span :class="['components-admin-user-user-api-keys-modal__text-5 badge', key.status === 'active' ? 'badge-success' : 'badge-danger']">{{ key.status }}</span></div>
              <p class="components-admin-user-user-api-keys-modal__description-4">{{ key.key.substring(0, 20) }}...{{ key.key.substring(key.key.length - 8) }}</p>
            </div>
            <button :aria-busy="updatingKeyIds.has(key.id)"
              type="button"
              class="components-admin-user-user-api-keys-modal__action"
              :disabled="updatingKeyIds.has(key.id)"
              :title="t('admin.users.manageGroups')"
              @click="openGroupSelector(key)"
            >
<LoadingSpinner v-if="updatingKeyIds.has(key.id)" size="sm" color="inherit" decorative />
              <Icon v-else name="grid" size="xs" />
              <span>{{ t('admin.users.manageGroups') }}</span>
            </button>
          </div>
          <div class="components-admin-user-user-api-keys-modal__panel-11">
            <div class="components-admin-user-user-api-keys-modal__panel-12">
              <span>{{ t('admin.users.group') }}:</span>
              <div class="components-admin-user-user-api-keys-modal__panel-13">
                <template v-if="keyGroupOptions(key).length > 0">
                  <GroupBadge
                    v-for="group in keyGroupOptions(key).slice(0, 2)"
                    :key="group.id"
                    :name="group.name"
                    :platform="group.platform"
                    :subscription-type="group.subscription_type"
                    :rate-multiplier="group.rate_multiplier"
                    :peak-rate-enabled="group.peak_rate_enabled"
                    :peak-start="group.peak_start"
                    :peak-end="group.peak_end"
                    :peak-rate-multiplier="group.peak_rate_multiplier"
                  />
                  <span v-if="keyGroupOptions(key).length > 2" class="components-admin-user-user-api-keys-modal__text-2">+{{ keyGroupOptions(key).length - 2 }}</span>
                </template>
                <span v-else class="components-admin-user-user-api-keys-modal__text-3">{{ t('admin.users.none') }}</span>
              </div>
            </div>
            <div class="components-admin-user-user-api-keys-modal__panel-14"><span>{{ t('admin.users.columns.created') }}: {{ formatDateTime(key.created_at) }}</span></div>
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>

  <BaseDialog
    :show="showGroupManager"
    :title="t('admin.users.manageGroups')"
    width="normal"
    :z-index="60"
    @close="closeGroupSelector"
  >
    <div class="components-admin-user-user-api-keys-modal__panel-15">
      <p class="components-admin-user-user-api-keys-modal__description-5">
        {{ selectedKeyForGroup?.name }}
      </p>
      <GroupTransferPicker
        v-model="selectedGroupIds"
        :groups="allGroups"
        :available-label="t('admin.users.availableGroups')"
        :selected-label="t('admin.users.selectedGroups')"
        :search-placeholder="t('admin.users.searchGroups')"
        :empty-label="t('admin.users.noGroups')"
      />
      <div class="components-admin-user-user-api-keys-modal__panel-16">
        <span class="components-admin-user-user-api-keys-modal__text-4">{{ selectedGroupIds.length }} {{ t('admin.users.group') }}</span>
        <button class="components-admin-user-user-api-keys-modal__action-2" :disabled="updatingKeyIds.has(selectedKeyForGroup?.id || '')" @click="applyGroups">
          {{ t('common.save') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import LoadingState from '@/components/common/LoadingState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { AdminUser, AdminGroup, ApiKey } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupTransferPicker from '@/components/common/GroupTransferPicker.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close'])
const { t } = useI18n()
const appStore = useAppStore()

const apiKeys = ref<ApiKey[]>([])
const allGroups = ref<AdminGroup[]>([])
const loading = ref(false)
const updatingKeyIds = ref(new Set<string>())
const groupSelectorKeyId = ref<string | null>(null)
const showGroupManager = ref(false)
const scrollContainerRef = ref<HTMLElement | null>(null)

const selectedKeyForGroup = computed(() => {
  if (groupSelectorKeyId.value === null) return null
  return apiKeys.value.find((k) => k.id === groupSelectorKeyId.value) || null
})
const selectedGroupIds = ref<string[]>([])

const keyGroupOptions = (key: ApiKey): AdminGroup[] => {
  const ids = key.group_ids?.length ? key.group_ids : (key.group_id ? [key.group_id] : [])
  return ids.map((id) => allGroups.value.find((group) => group.id === id) || (key.group_id === id && key.group ? key.group : null)).filter((group): group is AdminGroup => Boolean(group))
}

watch(() => props.show, (v) => {
  if (v && props.user) {
    load()
    loadGroups()
  } else {
    closeGroupSelector()
  }
})

const load = async () => {
  if (!props.user) return
  loading.value = true
  try {
    const res = await adminAPI.users.getUserApiKeys(props.user.id)
    apiKeys.value = res.items || []
  } catch (error) {
    console.error('Failed to load API keys:', error)
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  try {
    const groups = await adminAPI.groups.getAll()
    allGroups.value = groups
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

const openGroupSelector = (key: ApiKey) => {
  if (groupSelectorKeyId.value === key.id) {
    closeGroupSelector()
  } else {
    selectedGroupIds.value = key.group_ids?.length ? [...key.group_ids] : (key.group_id ? [key.group_id] : [])
    groupSelectorKeyId.value = key.id
    showGroupManager.value = true
  }
}

const closeGroupSelector = () => {
  groupSelectorKeyId.value = null
  showGroupManager.value = false
}

const applyGroups = async () => {
  const key = selectedKeyForGroup.value
  if (!key) return
  closeGroupSelector()
  const current = key.group_ids?.length ? key.group_ids : (key.group_id ? [key.group_id] : [])
  if (JSON.stringify(current) === JSON.stringify(selectedGroupIds.value)) return

  updatingKeyIds.value.add(key.id)
  try {
    const result = await adminAPI.apiKeys.updateApiKeyGroups(key.id, selectedGroupIds.value)
    // Update local data
    const idx = apiKeys.value.findIndex((k) => k.id === key.id)
    if (idx !== -1) {
      apiKeys.value[idx] = result.api_key
    }
    if (result.auto_granted_group_access && result.granted_group_name) {
      appStore.showSuccess(t('admin.users.groupChangedWithGrant', { group: result.granted_group_name }))
    } else {
      appStore.showSuccess(t('admin.users.groupChangedSuccess'))
    }
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.users.groupChangeFailed'))
  } finally {
    updatingKeyIds.value.delete(key.id)
  }
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && groupSelectorKeyId.value !== null) {
    event.stopPropagation()
    closeGroupSelector()
  }
}

const handleClose = () => {
  closeGroupSelector()
  emit('close')
}

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown, true)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown, true)
})
</script>
