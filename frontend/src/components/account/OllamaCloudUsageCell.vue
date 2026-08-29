<template>
  <div
    v-if="state?.eligible"
    class="ollama-usage-cell"
    data-testid="ollama-cloud-usage-cell"
  >
    <UsageProgressBar
      v-if="snapshot?.data?.five_hour"
      label="5h"
      :utilization="snapshot.data.five_hour.used_percent"
      :resets-at="snapshot.data.five_hour.reset_at"
      color="indigo"
      data-testid="ollama-cloud-five-hour"
    />
    <UsageProgressBar
      v-if="snapshot?.data?.seven_day"
      label="7d"
      :utilization="snapshot.data.seven_day.used_percent"
      :resets-at="snapshot.data.seven_day.reset_at"
      color="emerald"
      data-testid="ollama-cloud-seven-day"
    />
    <div v-if="state.configured" class="ollama-usage-cell__actions">
      <button
        type="button"
        class="ollama-usage-cell__refresh"
        :disabled="refreshing"
        data-testid="ollama-cloud-usage-query"
        @click="refreshUsage"
      >
        <svg
          class="ollama-usage-cell__refresh-icon"
          :class="{ 'ollama-usage-cell__refresh-icon--spinning': refreshing }"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
        {{ t('admin.accounts.usageWindow.activeQuery') }}
      </button>
    </div>
  </div>
  <span v-else class="ollama-usage-cell__empty">-</span>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { Account, OllamaCloudUsageState } from '@/types'
import UsageProgressBar from './UsageProgressBar.vue'

const props = defineProps<{ account: Account }>()
const emit = defineEmits<{ updated: [state: OllamaCloudUsageState] }>()
const { t } = useI18n()
const state = ref(props.account.ollama_cloud_usage)
const refreshing = ref(false)
const snapshot = computed(() => state.value?.snapshot)

watch(() => props.account.ollama_cloud_usage, (next) => {
  state.value = next
})

const refreshUsage = async () => {
  if (refreshing.value) return
  refreshing.value = true
  try {
    const next = await adminAPI.accounts.refreshOllamaCloudUsage(props.account.id)
    state.value = next
    emit('updated', next)
  } catch (error) {
    console.error('Failed to refresh Ollama Cloud usage:', error)
  } finally {
    refreshing.value = false
  }
}
</script>

<style scoped>
.ollama-usage-cell {
  min-width: 0;
  max-width: 100%;
}

.ollama-usage-cell > * + * { margin-top: 0.25rem; }

.ollama-usage-cell__actions {
  display: flex;
  align-items: center;
  padding-top: 0.125rem;
}

.ollama-usage-cell__refresh {
  display: inline-flex;
  align-items: center;
  gap: 0.125rem;
  padding: 0.125rem 0.375rem;
  border-radius: var(--radius-sm);
  color: var(--color-text-brand);
  font-size: var(--type-micro-size);
  font-weight: var(--font-weight-medium);
  transition: box-shadow 150ms ease, backdrop-filter 150ms ease;
}

.ollama-usage-cell__refresh:hover:not(:disabled) {
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.ollama-usage-cell__refresh:disabled { cursor: not-allowed; opacity: 0.5; }
.ollama-usage-cell__refresh-icon { width: 0.625rem; height: 0.625rem; }
.ollama-usage-cell__refresh-icon--spinning { animation: ollama-usage-spin 1s linear infinite; }
.ollama-usage-cell__empty { color: var(--color-text-tertiary); font-size: var(--type-control-size); }

@keyframes ollama-usage-spin { to { transform: rotate(360deg); } }
</style>
