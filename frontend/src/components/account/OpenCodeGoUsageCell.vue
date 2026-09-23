<template>
  <div class="opencode-usage" data-testid="opencode-go-usage-cell">
    <UsageProgressBar v-for="window in windows" :key="window.label" :label="window.label" :utilization="window.value.percent" :resets-at="window.value.resets_at" :color="window.color" />
    <span v-if="!windows.length">{{ t('upstreamUpdate.unknown') }}</span>
    <span v-if="state?.snapshot?.status !== 'ok' && state?.snapshot" role="status">{{ t('upstreamUpdate.failed') }}</span>
    <time v-if="state?.snapshot?.fetched_at" :title="t('upstreamUpdate.updated')">{{ new Date(state.snapshot.fetched_at).toLocaleString() }}</time>
    <div class="opencode-usage__controls">
      <button type="button" class="btn btn-secondary btn-xs" :title="t('upstreamUpdate.refresh')" :aria-label="t('upstreamUpdate.refresh')" :disabled="busy" @click="refresh"><Icon name="refresh" size="sm" /></button>
      <label><input type="checkbox" :checked="state?.auto_refresh_enabled" :disabled="busy" @change="toggle(($event.target as HTMLInputElement).checked)" />{{ t('upstreamUpdate.autoRefresh') }}</label>
    </div>
    <span v-if="error" role="alert">{{ error }}</span>
  </div>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import UsageProgressBar from './UsageProgressBar.vue'
import { refreshOpenCodeUsage, setOpenCodeAutoRefresh, type OpenCodeGoUsageState } from '@/api/admin/upstreamCapabilities'
const props = defineProps<{ account: { id: string; opencode_go_usage?: OpenCodeGoUsageState } }>()
const emit = defineEmits<{ updated: [value: OpenCodeGoUsageState] }>()
const { t } = useI18n()
const state = ref(props.account.opencode_go_usage)
const busy = ref(false), error = ref('')
let generation = 0
watch(() => [props.account.id, props.account.opencode_go_usage] as const, () => { generation++; state.value = props.account.opencode_go_usage; error.value = ''; busy.value = false })
const windows = computed(() => {
  const data = state.value?.snapshot?.data
  return [{ label: '5h', value: data?.rolling, color: 'indigo' as const }, { label: '7d', value: data?.weekly, color: 'emerald' as const }, { label: '1m', value: data?.monthly, color: 'amber' as const }].filter(w => w.value != null).map(w => ({ ...w, value: w.value! }))
})
async function run(action: () => Promise<OpenCodeGoUsageState>) {
  const current = ++generation
  busy.value = true; error.value = ''
  try { const result = await action(); if (current === generation) { state.value = result; emit('updated', result) } }
  catch { if (current === generation) error.value = t('upstreamUpdate.queryFailed') }
  finally { if (current === generation) busy.value = false }
}
function refresh() { return run(() => refreshOpenCodeUsage(props.account.id)) }
function toggle(enabled: boolean) { return run(() => setOpenCodeAutoRefresh(props.account.id, enabled)) }
</script>
<style scoped>
.opencode-usage { min-width: 0; display: grid; gap: 4px; font-size: 11px; }
.opencode-usage time { font-size: 10px; overflow-wrap: anywhere; opacity: .7; }
.opencode-usage__controls, .opencode-usage label { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
</style>
