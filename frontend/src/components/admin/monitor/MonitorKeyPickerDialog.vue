<template>
  <BaseDialog
    :show="show"
    :title="t('admin.channelMonitor.form.selectKeyTitle')"
    width="wide"
    @close="$emit('close')"
  >
    <div class="components-admin-monitor-monitor-key-picker-dialog__panel">
      <p class="components-admin-monitor-monitor-key-picker-dialog__description">
        {{ t('admin.channelMonitor.form.selectKeyHint') }}
      </p>

      <div class="components-admin-monitor-monitor-key-picker-dialog__panel-2">
        <input
          v-model="search"
          type="text"
          class="components-admin-monitor-monitor-key-picker-dialog__field input"
          :placeholder="t('keys.searchPlaceholder')"
        />
        <svg class="components-admin-monitor-monitor-key-picker-dialog__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
        </svg>
      </div>

      <div v-if="loading" class="components-admin-monitor-monitor-key-picker-dialog__panel-3">
        {{ t('common.loading') }}
      </div>
      <div v-else-if="filteredKeys.length === 0" class="components-admin-monitor-monitor-key-picker-dialog__panel-3">
        {{ t('admin.channelMonitor.form.noActiveKey') }}
      </div>
      <div v-else class="components-admin-monitor-monitor-key-picker-dialog__panel-4">
        <table class="components-admin-monitor-monitor-key-picker-dialog__table">
          <thead class="components-admin-monitor-monitor-key-picker-dialog__header">
            <tr class="components-admin-monitor-monitor-key-picker-dialog__row">
              <th class="components-admin-monitor-monitor-key-picker-dialog__heading">{{ t('common.name') }}</th>
              <th class="components-admin-monitor-monitor-key-picker-dialog__heading">{{ t('keys.apiKey') }}</th>
              <th class="components-admin-monitor-monitor-key-picker-dialog__heading">{{ t('keys.group') }}</th>
            </tr>
          </thead>
          <tbody class="components-admin-monitor-monitor-key-picker-dialog__body">
            <tr
              v-for="k in filteredKeys"
              :key="k.id"
              class="components-admin-monitor-monitor-key-picker-dialog__row-2"
              @click="$emit('pick', k)"
            >
              <td class="components-admin-monitor-monitor-key-picker-dialog__cell">{{ k.name }}</td>
              <td class="components-admin-monitor-monitor-key-picker-dialog__cell-2">{{ maskApiKey(k.key) }}</td>
              <td class="components-admin-monitor-monitor-key-picker-dialog__heading">
                <GroupBadge
                  v-if="k.group"
                  :name="k.group.name"
                  :platform="k.group.platform"
                  :subscription-type="k.group.subscription_type"
                  :rate-multiplier="k.group.rate_multiplier"
                  :user-rate-multiplier="userGroupRates[k.group.id]"
                />
                <span v-else class="components-admin-monitor-monitor-key-picker-dialog__text">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <template #footer>
      <div class="components-admin-monitor-monitor-key-picker-dialog__panel-5">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ApiKey } from '@/types'
import type { Provider } from '@/api/admin/channelMonitor'
import BaseDialog from '@/components/common/BaseDialog.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import { maskApiKey } from '@/utils/maskApiKey'

const props = withDefaults(defineProps<{
  show: boolean
  loading: boolean
  keys: ApiKey[]
  provider: Provider
  userGroupRates?: Record<string, number>
}>(), {
  userGroupRates: () => ({}),
})

defineEmits<{
  (e: 'close'): void
  (e: 'pick', key: ApiKey): void
}>()

const { t } = useI18n()

const search = ref('')

watch(() => props.show, (shown) => {
  if (!shown) search.value = ''
})

const filteredKeys = computed<ApiKey[]>(() => {
  const q = search.value.trim().toLowerCase()
  return props.keys.filter((k) => {
    if (k.group?.platform !== props.provider) return false
    if (!q) return true
    return (
      k.name.toLowerCase().includes(q) ||
      k.key.toLowerCase().includes(q) ||
      (k.group?.name || '').toLowerCase().includes(q)
    )
  })
})
</script>
