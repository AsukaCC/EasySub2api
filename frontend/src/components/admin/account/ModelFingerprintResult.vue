<template>
  <div v-if="snapshot" class="fingerprint-result" :class="{ 'fingerprint-result--compact': compact }">
    <div class="fingerprint-model" :title="snapshot.model">{{ snapshot.model }}</div>
    <div v-if="snapshot.status === 'running'" class="fingerprint-status" role="status">
      <Icon name="refresh" size="sm" class="fingerprint-spin" />
      {{ t('admin.accounts.fingerprint.progress', { count: snapshot.completed, total: snapshot.total }) }}
    </div>
    <div v-else-if="snapshot.status === 'failed'" class="fingerprint-error" role="status">
      {{ t(`admin.accounts.fingerprint.errors.${snapshot.error || 'upstream_failed'}`) }}
    </div>
    <template v-else-if="snapshot.result">
      <div class="fingerprint-bar" role="img" :aria-label="familyLabel" :title="familyLabel">
        <span v-for="family in snapshot.result.families" :key="family.family"
          :class="family.family === 'gpt' ? 'fingerprint-gpt' : 'fingerprint-claude'"
          :style="{ width: `${family.probability * 100}%` }" />
      </div>
      <div v-if="!compact" class="fingerprint-families">
        <span v-for="family in snapshot.result.families" :key="family.family">
          <i :class="family.family === 'gpt' ? 'fingerprint-gpt' : 'fingerprint-claude'" />
          {{ family.display_name }} {{ percent(family.probability) }}
        </span>
      </div>
      <div v-for="item in visibleModels" :key="item.model" class="fingerprint-row">
        <span class="fingerprint-candidate" :title="item.model">{{ item.model }}</span>
        <span>{{ percent(item.probability) }}</span>
        <div v-if="!compact" class="fingerprint-model-bar"><span :style="{ width: `${item.probability * 100}%` }" /></div>
      </div>
      <details v-if="!compact && snapshot.result.models.length > 3" class="fingerprint-more">
        <summary>{{ t('admin.accounts.fingerprint.allCandidates') }}</summary>
        <div v-for="item in snapshot.result.models.slice(3)" :key="item.model" class="fingerprint-row">
          <span class="fingerprint-candidate">{{ item.model }}</span><span>{{ percent(item.probability) }}</span>
        </div>
      </details>
    </template>
    <time v-if="!compact && snapshot.finished_at" class="fingerprint-time">{{ new Date(snapshot.finished_at).toLocaleString() }}</time>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import type { ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'

const props = defineProps<{ snapshot: ModelFingerprintSnapshot | null; compact?: boolean }>()
const { t } = useI18n()
const percent = (value: number) => `${(value * 100).toFixed(1)}%`
const visibleModels = computed(() => props.snapshot?.result?.models.slice(0, props.compact ? 1 : 3) || [])
const familyLabel = computed(() => props.snapshot?.result?.families.map(f => `${f.display_name} ${percent(f.probability)}`).join(', ') || '')
</script>

<style scoped>
.fingerprint-result { display: grid; gap: 10px; min-width: 0; font-size: var(--font-size-sm); }
.fingerprint-result--compact { width: 180px; gap: 5px; font-size: var(--font-size-xs); text-align: left; }
.fingerprint-model { color: var(--color-text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.fingerprint-bar { display: flex; height: 6px; overflow: hidden; border-radius: 3px; background: var(--color-surface-muted); }
.fingerprint-gpt { background: #138a72; }
.fingerprint-claude { background: #c46a38; }
.fingerprint-families { display: flex; flex-wrap: wrap; gap: 16px; }
.fingerprint-families span { display: inline-flex; align-items: center; gap: 5px; }
.fingerprint-families i { width: 8px; height: 8px; border-radius: 2px; }
.fingerprint-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 4px 12px; font-variant-numeric: tabular-nums; }
.fingerprint-candidate { overflow-wrap: anywhere; }
.fingerprint-model-bar { grid-column: 1 / -1; height: 3px; background: var(--color-surface-muted); }
.fingerprint-model-bar span { display: block; height: 100%; background: var(--color-text-secondary); }
.fingerprint-status { display: flex; align-items: center; gap: 6px; }
.fingerprint-error { color: var(--color-text-danger); white-space: normal; }
.fingerprint-time, .fingerprint-more { font-size: var(--font-size-xs); color: var(--color-text-secondary); }
.fingerprint-more summary { cursor: pointer; margin-bottom: 8px; }
.fingerprint-more .fingerprint-row + .fingerprint-row { margin-top: 6px; }
.fingerprint-spin { animation: fingerprint-rotate 1.5s linear infinite; }
@keyframes fingerprint-rotate { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) { .fingerprint-spin { animation: none; } }
</style>
