<template>
  <section aria-labelledby="prompt-runtime-title" class="features-prompt-audit-components-runtime-overview__section">
    <div class="features-prompt-audit-components-runtime-overview__panel">
      <div>
        <h2 id="prompt-runtime-title" class="features-prompt-audit-components-runtime-overview__heading">
          {{ t('admin.promptAudit.runtime.title') }}
        </h2>
        <p class="features-prompt-audit-components-runtime-overview__description">
          {{ t('admin.promptAudit.runtime.description') }}
        </p>
      </div>
      <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="$emit('refresh')">
        {{ t('admin.promptAudit.actions.refresh') }}
      </button>
    </div>

    <div v-if="error" role="alert" class="features-prompt-audit-components-runtime-overview__panel-2">
      {{ error }}
    </div>
    <div v-else-if="loading && !runtime" class="features-prompt-audit-components-runtime-overview__panel-3" aria-busy="true">
      <div v-for="index in 6" :key="index" class="features-prompt-audit-components-runtime-overview__panel-4" />
    </div>
    <template v-else-if="runtime">
      <dl class="features-prompt-audit-components-runtime-overview__panel-3">
        <div v-for="item in statusItems" :key="item.label" class="features-prompt-audit-components-runtime-overview__panel-5">
          <dt class="features-prompt-audit-components-runtime-overview__dt">{{ item.label }}</dt>
          <dd class="features-prompt-audit-components-runtime-overview__dd">
            <span v-if="item.dot" class="features-prompt-audit-components-runtime-overview__text" :class="item.dot" />
            <span class="features-prompt-audit-components-runtime-overview__text-2">{{ item.value }}</span>
          </dd>
        </div>
      </dl>

      <div class="features-prompt-audit-components-runtime-overview__panel-6">
        <div class="features-prompt-audit-components-runtime-overview__panel-7">
          <h3 class="features-prompt-audit-components-runtime-overview__heading-2">{{ t('admin.promptAudit.runtime.guardMetrics') }}</h3>
          <div class="features-prompt-audit-components-runtime-overview__panel-8">
            <div v-for="metric in guardMetricItems" :key="metric.label" class="features-prompt-audit-components-runtime-overview__panel-9">
              <p class="features-prompt-audit-components-runtime-overview__description-2">{{ metric.label }}</p>
              <p class="features-prompt-audit-components-runtime-overview__description-3">{{ metric.value }}</p>
            </div>
          </div>
          <p class="features-prompt-audit-components-runtime-overview__description-4">
            {{ t('admin.promptAudit.runtime.queueBreakdown', {
              queued: runtime.queue.queued,
              processing: runtime.queue.processing,
              retry: runtime.queue.retry,
              done: runtime.queue.done,
              failed: runtime.queue.failed,
            }) }}
            <span class="features-prompt-audit-components-runtime-overview__text-3">·</span>
            {{ t('admin.promptAudit.runtime.deliveryTotals', { enqueued: runtime.enqueued_total, dropped: runtime.dropped_total, processed: runtime.processed_total, failed: runtime.failed_total }) }}
          </p>
        </div>
        <div class="features-prompt-audit-components-runtime-overview__panel-7">
          <h3 class="features-prompt-audit-components-runtime-overview__heading-2">{{ t('admin.promptAudit.runtime.latest') }}</h3>
          <p class="features-prompt-audit-components-runtime-overview__description-5">
            {{ runtime.last_processed_at ? formatDate(runtime.last_processed_at) : t('admin.promptAudit.common.never') }}
          </p>
          <p v-if="runtime.last_error_code" class="features-prompt-audit-components-runtime-overview__description-6">
            {{ runtime.last_error_code }}<span v-if="runtime.last_error_message"> · {{ runtime.last_error_message }}</span>
          </p>
          <div v-if="Object.keys(runtime.endpoints).length" class="features-prompt-audit-components-runtime-overview__panel-10">
            <span v-for="(probe, id) in runtime.endpoints" :key="id" class="features-prompt-audit-components-runtime-overview__text-4" :class="probe.ok ? 'features-prompt-audit-components-runtime-overview__text-5' : 'features-prompt-audit-components-runtime-overview__text-6'">
              {{ id }} · {{ probe.status }} · {{ probe.latency_ms }} ms
            </span>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PromptAuditRuntime } from '../types'

const props = defineProps<{ runtime: PromptAuditRuntime | null; loading: boolean; error: string }>()
defineEmits<{ (event: 'refresh'): void }>()
const { t, locale } = useI18n()

const statusItems = computed(() => {
  const runtime = props.runtime
  if (!runtime) return []
  return [
    { label: t('admin.promptAudit.runtime.process'), value: t(`admin.promptAudit.status.${runtime.process_status}`), dot: statusDot(runtime.process_status) },
    { label: t('admin.promptAudit.runtime.mode'), value: t(`admin.promptAudit.mode.${runtime.effective_mode}`) },
    { label: t('admin.promptAudit.runtime.version'), value: `${runtime.active_config_version} / ${runtime.expected_config_version}` },
    { label: t('admin.promptAudit.runtime.workers'), value: `${runtime.worker_active} / ${runtime.worker_total}` },
    { label: t('admin.promptAudit.runtime.queue'), value: `${runtime.queue.active} / ${runtime.queue_capacity}` },
    { label: t('admin.promptAudit.runtime.dependencies'), value: `DB ${runtime.database_status} · Redis ${runtime.redis_status}` },
  ]
})

const guardMetricItems = computed(() => {
  const metrics = props.runtime?.guard_metrics
  if (!metrics) return []
  return [
    { label: t('admin.promptAudit.metrics.total'), value: metrics.total },
    { label: t('admin.promptAudit.metrics.allowed'), value: metrics.allowed },
    { label: t('admin.promptAudit.metrics.flagged'), value: metrics.flagged },
    { label: t('admin.promptAudit.metrics.blocked'), value: metrics.blocked },
    { label: t('admin.promptAudit.metrics.unavailable'), value: metrics.unavailable },
    { label: t('admin.promptAudit.metrics.timeouts'), value: metrics.timeouts },
    { label: t('admin.promptAudit.metrics.failovers'), value: metrics.failovers },
    { label: 'P95', value: metrics.latency_p95_ms != null ? `${metrics.latency_p95_ms} ms` : '—' },
  ]
})

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(locale.value, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value))
}

function statusDot(status: string): string {
  if (status === 'running') return 'status-fill--success'
  if (status === 'disabled') return 'status-fill--neutral'
  if (status === 'degraded') return 'status-fill--warning'
  return 'status-fill--danger'
}
</script>
