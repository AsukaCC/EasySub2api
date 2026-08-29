<template>
  <div
    v-if="uniqueIps.length > 0"
    class="ip-geo-batch-toolbar"
  >
    <span v-if="pendingCount > 0" class="ip-geo-batch-toolbar__status">
      {{ t('usage.ipGeo.pending', { count: pendingCount }) }}
    </span>
    <button
      type="button"
      class="ip-geo-batch-toolbar__fetch"
      :disabled="loading || pendingCount === 0"
      @click="run"
    >
      {{ loading ? t('usage.ipGeo.batchFetching') : t('usage.ipGeo.batchFetch') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchBatch, getEntry } from '@/utils/ipGeoLookup'

// 当前页 IP 批量地理查询工具条:传入原始 IP 列表(可含空值),内部去重;
// 无 IP 时自身不渲染。批量失败 emit failed,由使用方弹提示。
const props = defineProps<{
  ips: Array<string | null | undefined>
}>()

const emit = defineEmits<{
  (e: 'failed'): void
}>()

const { t } = useI18n()

const uniqueIps = computed(() =>
  Array.from(new Set(props.ips.filter((ip): ip is string => Boolean(ip))))
)

const pendingCount = computed(() =>
  uniqueIps.value.filter((ip) => {
    const status = getEntry(ip).status
    return status === 'idle' || status === 'error'
  }).length
)

const loading = ref(false)

const run = async () => {
  loading.value = true
  try {
    const ok = await fetchBatch(uniqueIps.value)
    if (!ok) emit('failed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.ip-geo-batch-toolbar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.ip-geo-batch-toolbar__status,
.ip-geo-batch-toolbar__fetch {
  font-size: var(--type-caption-size);
  line-height: var(--type-caption-line-height);
}

.ip-geo-batch-toolbar__status {
  color: var(--color-text-secondary);
}

.ip-geo-batch-toolbar__fetch {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.5rem;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: var(--color-text-brand);
  font-weight: var(--font-weight-medium);
  transition: border-color 150ms ease, box-shadow 150ms ease, backdrop-filter 150ms ease;
}

.ip-geo-batch-toolbar__fetch:hover:not(:disabled) {
  border-color: var(--color-primary-border);
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.ip-geo-batch-toolbar__fetch:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
