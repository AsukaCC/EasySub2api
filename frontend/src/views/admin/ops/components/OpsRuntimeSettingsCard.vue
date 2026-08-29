<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { opsAPI } from '@/api/admin/ops'
import type { OpsAlertRuntimeSettings } from '../types'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)

const alertSettings = ref<OpsAlertRuntimeSettings | null>(null)

const showAlertEditor = ref(false)
const draftAlert = ref<OpsAlertRuntimeSettings | null>(null)

type ValidationResult = { valid: boolean; errors: string[] }

function normalizeSeverities(input: Array<string | null | undefined> | null | undefined): string[] {
  if (!input || input.length === 0) return []
  const allowed = new Set(['P0', 'P1', 'P2', 'P3'])
  const out: string[] = []
  const seen = new Set<string>()
  for (const raw of input) {
    const s = String(raw || '')
      .trim()
      .toUpperCase()
    if (!s) continue
    if (!allowed.has(s)) continue
    if (seen.has(s)) continue
    seen.add(s)
    out.push(s)
  }
  return out
}

function validateRuntimeSettings(settings: OpsAlertRuntimeSettings): ValidationResult {
  const errors: string[] = []

  const evalSeconds = settings.evaluation_interval_seconds
  if (!Number.isFinite(evalSeconds) || evalSeconds < 1 || evalSeconds > 86400) {
    errors.push(t('admin.ops.runtime.validation.evalIntervalRange'))
  }

  // Thresholds validation
  const thresholds = settings.thresholds
  if (thresholds) {
    if (thresholds.sla_percent_min != null) {
      if (!Number.isFinite(thresholds.sla_percent_min) || thresholds.sla_percent_min < 0 || thresholds.sla_percent_min > 100) {
        errors.push(t('admin.ops.runtime.validation.slaMinPercentRange'))
      }
    }
    if (thresholds.ttft_p99_ms_max != null) {
      if (!Number.isFinite(thresholds.ttft_p99_ms_max) || thresholds.ttft_p99_ms_max < 0) {
        errors.push(t('admin.ops.runtime.validation.ttftP99MaxRange'))
      }
    }
    if (thresholds.request_error_rate_percent_max != null) {
      if (!Number.isFinite(thresholds.request_error_rate_percent_max) || thresholds.request_error_rate_percent_max < 0 || thresholds.request_error_rate_percent_max > 100) {
        errors.push(t('admin.ops.runtime.validation.requestErrorRateMaxRange'))
      }
    }
    if (thresholds.upstream_error_rate_percent_max != null) {
      if (!Number.isFinite(thresholds.upstream_error_rate_percent_max) || thresholds.upstream_error_rate_percent_max < 0 || thresholds.upstream_error_rate_percent_max > 100) {
        errors.push(t('admin.ops.runtime.validation.upstreamErrorRateMaxRange'))
      }
    }
  }

  const lock = settings.distributed_lock
  if (lock?.enabled) {
    if (!lock.key || lock.key.trim().length < 3) {
      errors.push(t('admin.ops.runtime.validation.lockKeyRequired'))
    } else if (!lock.key.startsWith('ops:')) {
      errors.push(t('admin.ops.runtime.validation.lockKeyPrefix', { prefix: 'ops:' }))
    }
    if (!Number.isFinite(lock.ttl_seconds) || lock.ttl_seconds < 1 || lock.ttl_seconds > 86400) {
      errors.push(t('admin.ops.runtime.validation.lockTtlRange'))
    }
  }

  // Silencing validation (alert-only)
  const silencing = settings.silencing
  if (silencing?.enabled) {
    const until = (silencing.global_until_rfc3339 || '').trim()
    if (until) {
      const parsed = Date.parse(until)
      if (!Number.isFinite(parsed)) errors.push(t('admin.ops.runtime.silencing.validation.timeFormat'))
    }

    const entries = Array.isArray(silencing.entries) ? silencing.entries : []
    for (let idx = 0; idx < entries.length; idx++) {
      const entry = entries[idx]
      const untilEntry = (entry?.until_rfc3339 || '').trim()
      if (!untilEntry) {
        errors.push(t('admin.ops.runtime.silencing.entries.validation.untilRequired'))
        break
      }
      const parsedEntry = Date.parse(untilEntry)
      if (!Number.isFinite(parsedEntry)) {
        errors.push(t('admin.ops.runtime.silencing.entries.validation.untilFormat'))
        break
      }
      const ruleId = (entry as any)?.rule_id
      if (ruleId != null && (typeof ruleId !== 'string' || !ruleId.trim())) {
        errors.push(t('admin.ops.runtime.silencing.entries.validation.ruleIdPositive'))
        break
      }
      if ((entry as any)?.severities) {
        const raw = (entry as any).severities
        const normalized = normalizeSeverities(Array.isArray(raw) ? raw : [raw])
        if (Array.isArray(raw) && raw.length > 0 && normalized.length === 0) {
          errors.push(t('admin.ops.runtime.silencing.entries.validation.severitiesFormat'))
          break
        }
      }
    }
  }

  return { valid: errors.length === 0, errors }
}

const alertValidation = computed(() => {
  if (!draftAlert.value) return { valid: true, errors: [] as string[] }
  return validateRuntimeSettings(draftAlert.value)
})

async function loadSettings() {
  loading.value = true
  try {
    alertSettings.value = await opsAPI.getAlertRuntimeSettings()
  } catch (err: any) {
    console.error('[OpsRuntimeSettingsCard] Failed to load runtime settings', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.runtime.loadFailed'))
  } finally {
    loading.value = false
  }
}

function openAlertEditor() {
  if (!alertSettings.value) return
  draftAlert.value = JSON.parse(JSON.stringify(alertSettings.value))

  // Backwards-compat: ensure nested settings exist even if API payload is older.
  if (draftAlert.value) {
    if (!draftAlert.value.distributed_lock) {
      draftAlert.value.distributed_lock = { enabled: true, key: 'ops:alert:evaluator:leader', ttl_seconds: 30 }
    }
    if (!draftAlert.value.silencing) {
      draftAlert.value.silencing = { enabled: false, global_until_rfc3339: '', global_reason: '', entries: [] }
    }
    if (!Array.isArray(draftAlert.value.silencing.entries)) {
      draftAlert.value.silencing.entries = []
    }
    if (!draftAlert.value.thresholds) {
      draftAlert.value.thresholds = {
        sla_percent_min: 99.5,
        ttft_p99_ms_max: 500,
        request_error_rate_percent_max: 5,
        upstream_error_rate_percent_max: 5
      }
    }
  }

  showAlertEditor.value = true
}

function addSilenceEntry() {
  if (!draftAlert.value) return
  if (!draftAlert.value.silencing) {
    draftAlert.value.silencing = { enabled: true, global_until_rfc3339: '', global_reason: '', entries: [] }
  }
  if (!Array.isArray(draftAlert.value.silencing.entries)) {
    draftAlert.value.silencing.entries = []
  }
  draftAlert.value.silencing.entries.push({
    rule_id: undefined,
    severities: [],
    until_rfc3339: '',
    reason: ''
  })
}

function removeSilenceEntry(index: number) {
  if (!draftAlert.value?.silencing?.entries) return
  draftAlert.value.silencing.entries.splice(index, 1)
}

function updateSilenceEntryRuleId(index: number, raw: string) {
  const entries = draftAlert.value?.silencing?.entries
  if (!entries || !entries[index]) return
  const trimmed = raw.trim()
  if (!trimmed) {
    delete (entries[index] as any).rule_id
    return
  }
  (entries[index] as any).rule_id = trimmed
}

function updateSilenceEntrySeverities(index: number, raw: string) {
  const entries = draftAlert.value?.silencing?.entries
  if (!entries || !entries[index]) return
  const parts = raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
  ;(entries[index] as any).severities = normalizeSeverities(parts)
}

async function saveAlertSettings() {
  if (!draftAlert.value) return
  if (!alertValidation.value.valid) {
    appStore.showError(alertValidation.value.errors[0] || t('admin.ops.runtime.validation.invalid'))
    return
  }

  saving.value = true
  try {
    alertSettings.value = await opsAPI.updateAlertRuntimeSettings(draftAlert.value)
    showAlertEditor.value = false
    appStore.showSuccess(t('admin.ops.runtime.saveSuccess'))
  } catch (err: any) {
    console.error('[OpsRuntimeSettingsCard] Failed to save alert runtime settings', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.runtime.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<template>
  <div class="views-admin-ops-components-ops-runtime-settings-card__panel">
    <div class="views-admin-ops-components-ops-runtime-settings-card__panel-2">
      <div>
        <h3 class="views-admin-ops-components-ops-runtime-settings-card__heading">{{ t('admin.ops.runtime.title') }}</h3>
        <p class="views-admin-ops-components-ops-runtime-settings-card__description">{{ t('admin.ops.runtime.description') }}</p>
      </div>
      <button
        class="views-admin-ops-components-ops-runtime-settings-card__action"
        :disabled="loading"
        @click="loadSettings"
      >
        <svg class="views-admin-ops-components-ops-runtime-settings-card__icon" :class="{ 'views-admin-ops-components-ops-runtime-settings-card__icon-2': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        {{ t('common.refresh') }}
      </button>
    </div>

    <div v-if="!alertSettings" class="views-admin-ops-components-ops-runtime-settings-card__panel-3">
      <span v-if="loading">{{ t('admin.ops.runtime.loading') }}</span>
      <span v-else>{{ t('admin.ops.runtime.noData') }}</span>
    </div>

    <div v-else class="views-admin-ops-components-ops-runtime-settings-card__panel-4">
      <div class="views-admin-ops-components-ops-runtime-settings-card__panel-5">
        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-6">
          <h4 class="views-admin-ops-components-ops-runtime-settings-card__heading-2">{{ t('admin.ops.runtime.alertTitle') }}</h4>
          <button class="btn btn-sm btn-secondary" @click="openAlertEditor">{{ t('common.edit') }}</button>
        </div>
        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-7">
          <div class="views-admin-ops-components-ops-runtime-settings-card__panel-8">
            {{ t('admin.ops.runtime.evalIntervalSeconds') }}:
            <span class="views-admin-ops-components-ops-runtime-settings-card__text">{{ alertSettings.evaluation_interval_seconds }}s</span>
          </div>
          <div
            v-if="alertSettings.silencing?.enabled && alertSettings.silencing.global_until_rfc3339"
            class="views-admin-ops-components-ops-runtime-settings-card__panel-9"
          >
            {{ t('admin.ops.runtime.silencing.globalUntil') }}:
            <span class="views-admin-ops-components-ops-runtime-settings-card__text-2">{{ alertSettings.silencing.global_until_rfc3339 }}</span>
          </div>

          <details class="views-admin-ops-components-ops-runtime-settings-card__details">
            <summary class="views-admin-ops-components-ops-runtime-settings-card__summary">
              {{ t('admin.ops.runtime.showAdvancedDeveloperSettings') }}
            </summary>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-10">
              <div class="views-admin-ops-components-ops-runtime-settings-card__panel-11">
                {{ t('admin.ops.runtime.lockEnabled') }}:
                <span class="views-admin-ops-components-ops-runtime-settings-card__text-3">{{ alertSettings.distributed_lock.enabled }}</span>
              </div>
              <div class="views-admin-ops-components-ops-runtime-settings-card__panel-11">
                {{ t('admin.ops.runtime.lockKey') }}:
                <span class="views-admin-ops-components-ops-runtime-settings-card__text-3">{{ alertSettings.distributed_lock.key }}</span>
              </div>
              <div class="views-admin-ops-components-ops-runtime-settings-card__panel-11">
                {{ t('admin.ops.runtime.lockTTLSeconds') }}:
                <span class="views-admin-ops-components-ops-runtime-settings-card__text-3">{{ alertSettings.distributed_lock.ttl_seconds }}s</span>
              </div>
            </div>
          </details>
        </div>
      </div>
    </div>
  </div>

  <BaseDialog :show="showAlertEditor" :title="t('admin.ops.runtime.alertTitle')" width="extra-wide" @close="showAlertEditor = false">
    <div v-if="draftAlert" class="views-admin-ops-components-ops-runtime-settings-card__panel-12">
      <div
        v-if="!alertValidation.valid"
        class="views-admin-ops-components-ops-runtime-settings-card__panel-13"
      >
        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-14">{{ t('admin.ops.runtime.validation.title') }}</div>
        <ul class="views-admin-ops-components-ops-runtime-settings-card__list">
          <li v-for="msg in alertValidation.errors" :key="msg">{{ msg }}</li>
        </ul>
      </div>

      <div>
        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.evalIntervalSeconds') }}</div>
        <input
          v-model.number="draftAlert.evaluation_interval_seconds"
          type="number"
          min="1"
          max="86400"
          class="input"
          :aria-invalid="!alertValidation.valid"
        />
        <p class="views-admin-ops-components-ops-runtime-settings-card__description-2">{{ t('admin.ops.runtime.evalIntervalHint') }}</p>
      </div>

      <div class="views-admin-ops-components-ops-runtime-settings-card__panel-5">
        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-16">{{ t('admin.ops.runtime.metricThresholds') }}</div>
        <p class="views-admin-ops-components-ops-runtime-settings-card__description-3">{{ t('admin.ops.runtime.metricThresholdsHint') }}</p>

        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-17">
          <div>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.slaMinPercent') }}</div>
            <input
              v-model.number="draftAlert.thresholds.sla_percent_min"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              placeholder="99.5"
            />
            <p class="views-admin-ops-components-ops-runtime-settings-card__description">{{ t('admin.ops.runtime.slaMinPercentHint') }}</p>
          </div>



          <div>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.ttftP99MaxMs') }}</div>
            <input
              v-model.number="draftAlert.thresholds.ttft_p99_ms_max"
              type="number"
              min="0"
              step="100"
              class="input"
              placeholder="500"
            />
            <p class="views-admin-ops-components-ops-runtime-settings-card__description">{{ t('admin.ops.runtime.ttftP99MaxMsHint') }}</p>
          </div>

          <div>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.requestErrorRateMaxPercent') }}</div>
            <input
              v-model.number="draftAlert.thresholds.request_error_rate_percent_max"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              placeholder="5"
            />
            <p class="views-admin-ops-components-ops-runtime-settings-card__description">{{ t('admin.ops.runtime.requestErrorRateMaxPercentHint') }}</p>
          </div>

          <div>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.upstreamErrorRateMaxPercent') }}</div>
            <input
              v-model.number="draftAlert.thresholds.upstream_error_rate_percent_max"
              type="number"
              min="0"
              max="100"
              step="0.1"
              class="input"
              placeholder="5"
            />
            <p class="views-admin-ops-components-ops-runtime-settings-card__description">{{ t('admin.ops.runtime.upstreamErrorRateMaxPercentHint') }}</p>
          </div>
        </div>
      </div>

      <div class="views-admin-ops-components-ops-runtime-settings-card__panel-5">
        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-16">{{ t('admin.ops.runtime.silencing.title') }}</div>

        <label class="views-admin-ops-components-ops-runtime-settings-card__label">
          <input v-model="draftAlert.silencing.enabled" type="checkbox" class="views-admin-ops-components-ops-runtime-settings-card__field" />
          <span>{{ t('admin.ops.runtime.silencing.enabled') }}</span>
        </label>

        <div v-if="draftAlert.silencing.enabled" class="views-admin-ops-components-ops-runtime-settings-card__panel-18">
          <div>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.silencing.globalUntil') }}</div>
            <input
              v-model="draftAlert.silencing.global_until_rfc3339"
              type="text"
              class="views-admin-ops-components-ops-runtime-settings-card__field-2 input"
                      placeholder="2026-01-05T00:00:00Z"
            />
            <p class="views-admin-ops-components-ops-runtime-settings-card__description">{{ t('admin.ops.runtime.silencing.untilHint') }}</p>
          </div>

          <div>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.silencing.reason') }}</div>
            <input
              v-model="draftAlert.silencing.global_reason"
              type="text"
              class="input"
              :placeholder="t('admin.ops.runtime.silencing.reasonPlaceholder')"
            />
          </div>

          <div class="views-admin-ops-components-ops-runtime-settings-card__panel-19">
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-20">
              <div>
                <div class="views-admin-ops-components-ops-runtime-settings-card__panel-21">{{ t('admin.ops.runtime.silencing.entries.title') }}</div>
                <p class="views-admin-ops-components-ops-runtime-settings-card__description-4">{{ t('admin.ops.runtime.silencing.entries.hint') }}</p>
              </div>
              <button class="btn btn-sm btn-secondary" type="button" @click="addSilenceEntry">
                {{ t('admin.ops.runtime.silencing.entries.add') }}
              </button>
            </div>

            <div v-if="!draftAlert.silencing.entries?.length" class="views-admin-ops-components-ops-runtime-settings-card__panel-22">
              {{ t('admin.ops.runtime.silencing.entries.empty') }}
            </div>

            <div v-else class="views-admin-ops-components-ops-runtime-settings-card__panel-18">
              <div
                v-for="(entry, idx) in draftAlert.silencing.entries"
                :key="idx"
                class="views-admin-ops-components-ops-runtime-settings-card__panel-23"
              >
                <div class="views-admin-ops-components-ops-runtime-settings-card__panel-6">
                  <div class="views-admin-ops-components-ops-runtime-settings-card__panel-21">
                    {{ t('admin.ops.runtime.silencing.entries.entryTitle', { n: idx + 1 }) }}
                  </div>
                  <button class="btn btn-sm btn-danger" type="button" @click="removeSilenceEntry(idx)">{{ t('common.delete') }}</button>
                </div>

                <div class="views-admin-ops-components-ops-runtime-settings-card__panel-7">
                  <div>
                    <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.silencing.entries.ruleId') }}</div>
                    <input
                      :value="typeof (entry as any).rule_id === 'string' ? (entry as any).rule_id : ''"
                      type="text"
                      class="views-admin-ops-components-ops-runtime-settings-card__field-2 input"
                      :placeholder="t('admin.ops.runtime.silencing.entries.ruleIdPlaceholder')"
                      @input="updateSilenceEntryRuleId(idx, ($event.target as HTMLInputElement).value)"
                    />
                  </div>

                  <div>
                    <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.silencing.entries.severities') }}</div>
                    <input
                      :value="Array.isArray((entry as any).severities) ? (entry as any).severities.join(', ') : ''"
                      type="text"
                      class="views-admin-ops-components-ops-runtime-settings-card__field-2 input"
                      :placeholder="t('admin.ops.runtime.silencing.entries.severitiesPlaceholder')"
                      @input="updateSilenceEntrySeverities(idx, ($event.target as HTMLInputElement).value)"
                    />
                  </div>

                  <div>
                    <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.silencing.entries.until') }}</div>
                    <input
                      v-model="(entry as any).until_rfc3339"
                      type="text"
                      class="views-admin-ops-components-ops-runtime-settings-card__field-2 input"
              placeholder="2026-01-05T00:00:00Z"
                    />
                  </div>

                  <div>
                    <div class="views-admin-ops-components-ops-runtime-settings-card__panel-15">{{ t('admin.ops.runtime.silencing.entries.reason') }}</div>
                    <input
                      v-model="(entry as any).reason"
                      type="text"
                      class="input"
                      :placeholder="t('admin.ops.runtime.silencing.reasonPlaceholder')"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <details class="views-admin-ops-components-ops-runtime-settings-card__details-2">
        <summary class="views-admin-ops-components-ops-runtime-settings-card__summary-2">{{ t('admin.ops.runtime.advancedSettingsSummary') }}</summary>
        <div class="views-admin-ops-components-ops-runtime-settings-card__panel-24">
          <div>
            <label class="views-admin-ops-components-ops-runtime-settings-card__label-2">
              <input v-model="draftAlert.distributed_lock.enabled" type="checkbox" class="views-admin-ops-components-ops-runtime-settings-card__field" />
              <span>{{ t('admin.ops.runtime.lockEnabled') }}</span>
            </label>
          </div>
          <div class="views-admin-ops-components-ops-runtime-settings-card__panel-25">
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-26">{{ t('admin.ops.runtime.lockKey') }}</div>
            <input v-model="draftAlert.distributed_lock.key" type="text" class="views-admin-ops-components-ops-runtime-settings-card__field-3 input" />
            <p v-if="draftAlert.distributed_lock.enabled" class="views-admin-ops-components-ops-runtime-settings-card__description-5">
              {{ t('admin.ops.runtime.validation.lockKeyHint', { prefix: 'ops:' }) }}
            </p>
          </div>
          <div>
            <div class="views-admin-ops-components-ops-runtime-settings-card__panel-26">{{ t('admin.ops.runtime.lockTTLSeconds') }}</div>
            <input v-model.number="draftAlert.distributed_lock.ttl_seconds" type="number" min="1" max="86400" class="views-admin-ops-components-ops-runtime-settings-card__field-3 input" />
          </div>
        </div>
      </details>
    </div>

    <template #footer>
      <div class="views-admin-ops-components-ops-runtime-settings-card__panel-27">
        <button class="btn btn-secondary" @click="showAlertEditor = false">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="saving || !alertValidation.valid" @click="saveAlertSettings">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>
