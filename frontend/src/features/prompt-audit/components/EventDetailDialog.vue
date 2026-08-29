<template>
  <BaseDialog :show="show" :title="t('admin.promptAudit.events.detailTitle')" width="extra-wide" @close="$emit('close')">
    <div v-if="loading" class="features-prompt-audit-components-event-detail-dialog__panel" aria-busy="true">{{ t('common.loading') }}</div>
    <div v-else-if="event" class="features-prompt-audit-components-event-detail-dialog__panel-2">
      <div class="features-prompt-audit-components-event-detail-dialog__panel-3" role="tablist">
        <button v-for="tab in tabs" :key="tab" type="button" role="tab" :aria-selected="activeTab === tab" class="features-prompt-audit-components-event-detail-dialog__action" :class="activeTab === tab ? 'features-prompt-audit-components-event-detail-dialog__action-2' : 'features-prompt-audit-components-event-detail-dialog__action-3'" @click="activeTab = tab">
          {{ t(`admin.promptAudit.events.tabs.${tab}`) }}
        </button>
      </div>

      <!-- Fixed panel height so switching tabs does not resize the dialog -->
      <div class="features-prompt-audit-components-event-detail-dialog__panel-4" data-test="event-detail-tab-panel">
        <div v-show="activeTab === 'summary'" class="features-prompt-audit-components-event-detail-dialog__panel-5" role="tabpanel">
          <div>
            <h4 class="features-prompt-audit-components-event-detail-dialog__heading">{{ t('admin.promptAudit.events.promptFull') }}</h4>
            <pre class="features-prompt-audit-components-event-detail-dialog__pre" data-test="summary-prompt-full">{{ displayPrompt(event) }}</pre>
          </div>
          <dl class="features-prompt-audit-components-event-detail-dialog__dl">
            <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.decision') }}</dt><dd class="features-prompt-audit-components-event-detail-dialog__dd">{{ formatDecisionAction(event.decision, event.action) }}</dd>
            <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.user') }}</dt><dd>{{ event.snapshot.username || '—' }}</dd>
            <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.email') }}</dt><dd>{{ event.snapshot.user_email || '—' }}</dd>
            <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.apiKey') }}</dt><dd>{{ event.snapshot.api_key_name || '—' }}</dd>
            <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.group') }}</dt><dd>{{ event.snapshot.group_name || '—' }}</dd>
            <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.model') }}</dt><dd>{{ event.snapshot.model || '—' }}</dd>
            <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.categories') }}</dt><dd>{{ formatCategories(event.categories) }}</dd>
          </dl>
        </div>

        <div v-show="activeTab === 'risks'" class="features-prompt-audit-components-event-detail-dialog__panel-6" role="tabpanel">
          <div class="features-prompt-audit-components-event-detail-dialog__panel-7">
            <section data-test="risk-prompt-preview">
              <h4 class="features-prompt-audit-components-event-detail-dialog__heading">{{ t('admin.promptAudit.events.promptFull') }}</h4>
              <p class="features-prompt-audit-components-event-detail-dialog__description">{{ t('admin.promptAudit.events.promptFullHint') }}</p>
              <pre class="features-prompt-audit-components-event-detail-dialog__pre-2" data-test="risk-prompt-full">{{ displayPrompt(event) }}</pre>
            </section>
            <section data-test="risk-guard-return">
              <h4 class="features-prompt-audit-components-event-detail-dialog__heading">{{ t('admin.promptAudit.events.guardReturn') }}</h4>
              <p class="features-prompt-audit-components-event-detail-dialog__description">{{ t('admin.promptAudit.events.guardReturnHint') }}</p>
              <pre class="features-prompt-audit-components-event-detail-dialog__pre-3">{{ formatGuardReturn(event) }}</pre>
            </section>
          </div>

          <div class="features-prompt-audit-components-event-detail-dialog__panel-8">
            <h4 class="features-prompt-audit-components-event-detail-dialog__heading">{{ t('admin.promptAudit.events.riskSummaries') }}</h4>
            <article v-for="issue in event.issue_summaries" :key="`${issue.scanner_id}-${issue.code}`" class="features-prompt-audit-components-event-detail-dialog__article" data-test="risk-issue">
              <div class="features-prompt-audit-components-event-detail-dialog__panel-9">
                <h5 class="features-prompt-audit-components-event-detail-dialog__dd">{{ issueTitle(issue) }}</h5>
                <span class="features-prompt-audit-components-event-detail-dialog__text">{{ issueSeverity(issue) }} · {{ issueAction(issue) }}</span>
              </div>
              <p class="features-prompt-audit-components-event-detail-dialog__description-2">{{ issueDescription(issue) }}</p>
              <dl class="features-prompt-audit-components-event-detail-dialog__dl-2">
                <div><dt class="features-prompt-audit-components-event-detail-dialog__dt-2">{{ t('admin.promptAudit.events.categories') }} · </dt><dd class="features-prompt-audit-components-event-detail-dialog__dd-2">{{ translateCategory(issue.category || issue.scanner_id) }}</dd></div>
                <div><dt class="features-prompt-audit-components-event-detail-dialog__dt-2">{{ t('admin.promptAudit.events.score') }} · </dt><dd class="features-prompt-audit-components-event-detail-dialog__dd-2">{{ issue.score }}</dd></div>
                <div class="features-prompt-audit-components-event-detail-dialog__panel-10"><dt class="features-prompt-audit-components-event-detail-dialog__dt-2">{{ t('admin.promptAudit.events.evidence') }} · </dt><dd class="features-prompt-audit-components-event-detail-dialog__dd-3">{{ issue.evidence ? translateEvidence(issue.evidence) : '—' }}</dd></div>
              </dl>
            </article>
            <p v-if="event.issue_summaries.length === 0" class="features-prompt-audit-components-event-detail-dialog__description-3">{{ t('admin.promptAudit.events.noRisks') }}</p>
          </div>
        </div>

        <dl v-show="activeTab === 'technical'" class="features-prompt-audit-components-event-detail-dialog__dl-3" role="tabpanel">
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.requestId') }}</dt><dd class="features-prompt-audit-components-event-detail-dialog__dd-4">{{ event.snapshot.request_id || '—' }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.promptHash') }}</dt><dd class="features-prompt-audit-components-event-detail-dialog__dd-4">{{ event.snapshot.prompt_hash }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.technical.scanner') }}</dt><dd>{{ event.scanner_backend }} · {{ event.scanner_version }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.technical.policy') }}</dt><dd>{{ event.policy_id }} · v{{ event.policy_version }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.technical.guardEndpoint') }}</dt><dd>{{ event.guard_endpoint_id }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.technical.config') }}</dt><dd>v{{ event.config_version }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.technical.chunks') }}</dt><dd>{{ event.chunk_total }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.technical.latency') }}</dt><dd>{{ event.latency_ms }} ms</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.stage') }}</dt><dd>{{ event.snapshot.stage || 'http' }}</dd>
          <dt class="features-prompt-audit-components-event-detail-dialog__dt">{{ t('admin.promptAudit.events.technical.protocol') }}</dt><dd>{{ event.snapshot.protocol }} · {{ event.snapshot.endpoint }}</dd>
        </dl>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { PromptAuditEvent, PromptIssueSummary } from '../types'
import { SCANNER_CATALOG } from '../viewModel'

const props = defineProps<{ show: boolean; event: PromptAuditEvent | null; loading: boolean }>()
defineEmits<{ (event: 'close'): void }>()
const { t } = useI18n()
const tabs = ['summary', 'risks', 'technical'] as const
const activeTab = ref<(typeof tabs)[number]>('summary')
watch(() => props.event?.id, () => { activeTab.value = 'summary' })

const DECISIONS = new Set(['pass', 'flag', 'critical'])
const ACTIONS = new Set(['Allow', 'Warn', 'Block'])
const RISK_LEVELS = new Set(['low', 'medium', 'high', 'critical'])

function displayPrompt(event: PromptAuditEvent): string {
  return event.snapshot.full_prompt || event.snapshot.redacted_preview || '—'
}

function formatDecisionAction(decision: string, action: string): string {
  const decisionLabel = DECISIONS.has(decision) ? t(`admin.promptAudit.decisions.${decision}`) : decision
  const actionLabel = ACTIONS.has(action) ? t(`admin.promptAudit.actions.${action}`) : action
  return `${decisionLabel} · ${actionLabel}`
}
function translateCategory(category: string): string {
  return SCANNER_CATALOG.some((scanner) => scanner.id === category)
    ? t(`admin.promptAudit.scanners.${category}`)
    : category
}
function formatCategories(categories: string[]): string {
  if (!categories.length) return '—'
  return categories.map(translateCategory).join(', ')
}
function translateEvidence(value: string): string {
  const byId = SCANNER_CATALOG.find((scanner) => scanner.id === value)
  if (byId) return t(`admin.promptAudit.scanners.${byId.id}`)
  const byLabel = SCANNER_CATALOG.find((scanner) => scanner.label === value)
  if (byLabel) return t(`admin.promptAudit.scanners.${byLabel.id}`)
  return value
}
function formatGuardReturn(event: PromptAuditEvent): string {
  const evidence: Record<string, string> = {}
  for (const [key, value] of Object.entries(event.scanner_evidence || {})) {
    evidence[key] = translateEvidence(value)
  }
  return JSON.stringify({
    decision: DECISIONS.has(event.decision) ? t(`admin.promptAudit.decisions.${event.decision}`) : event.decision,
    risk_level: RISK_LEVELS.has(event.risk_level) ? t(`admin.promptAudit.riskLevels.${event.risk_level}`) : event.risk_level,
    action: ACTIONS.has(event.action) ? t(`admin.promptAudit.actions.${event.action}`) : event.action,
    categories: event.categories.map(translateCategory),
    matched_scanners: event.matched_scanners.map(translateCategory),
    scanner_scores: event.scanner_scores,
    scanner_evidence: evidence,
    scanner_backend: event.scanner_backend,
    scanner_version: event.scanner_version,
    guard_endpoint_id: event.guard_endpoint_id,
    chunk_total: event.chunk_total,
    latency_ms: event.latency_ms,
  }, null, 2)
}
function issueTitle(issue: PromptIssueSummary): string {
  return translateCategory(issue.category || issue.scanner_id) || issue.title
}
function issueDescription(issue: PromptIssueSummary): string {
  const category = issue.category || issue.scanner_id
  const key = `admin.promptAudit.scannerDescriptions.${category}`
  const label = t(key)
  return label === key ? issue.description : label
}
function issueSeverity(issue: PromptIssueSummary): string {
  return RISK_LEVELS.has(issue.severity) ? t(`admin.promptAudit.riskLevels.${issue.severity}`) : issue.severity_label || issue.severity
}
function issueAction(issue: PromptIssueSummary): string {
  return ACTIONS.has(issue.action) ? t(`admin.promptAudit.actions.${issue.action}`) : issue.action_label || issue.action
}
</script>
