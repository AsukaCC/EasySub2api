<template>
  <div class="components-admin-monitor-monitor-advanced-request-config__panel">
    <!-- Headers key-value rows -->
    <div>
      <label class="input-label">{{ t('admin.channelMonitor.advanced.headers') }}</label>
      <div class="components-admin-monitor-monitor-advanced-request-config__panel-2">
        <div
          v-for="(row, i) in headerRows"
          :key="i"
          class="components-admin-monitor-monitor-advanced-request-config__panel-3"
        >
          <input
            v-model="row.name"
            type="text"
            spellcheck="false"
            :placeholder="t('admin.channelMonitor.advanced.headerNamePlaceholder')"
            class="components-admin-monitor-monitor-advanced-request-config__field input"
            @blur="commitHeaders"
          />
          <input
            v-model="row.value"
            type="text"
            spellcheck="false"
            :placeholder="t('admin.channelMonitor.advanced.headerValuePlaceholder')"
            class="components-admin-monitor-monitor-advanced-request-config__field-2 input"
            @blur="commitHeaders"
          />
          <button
            type="button"
            class="components-admin-monitor-monitor-advanced-request-config__action"
            :title="t('common.delete')"
            @click="removeRow(i)"
          >
            <svg class="components-admin-monitor-monitor-advanced-request-config__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <button
          type="button"
          class="components-admin-monitor-monitor-advanced-request-config__action-2"
          @click="addRow"
        >
          <svg class="components-admin-monitor-monitor-advanced-request-config__icon-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('admin.channelMonitor.advanced.headerAddRow') }}
        </button>
      </div>
      <p v-if="headersError" class="components-admin-monitor-monitor-advanced-request-config__description">{{ headersError }}</p>
      <p v-else class="components-admin-monitor-monitor-advanced-request-config__description-2">
        {{ t('admin.channelMonitor.advanced.headersHint') }}
      </p>
    </div>

    <!-- Body mode radio -->
    <div>
      <label class="input-label">{{ t('admin.channelMonitor.advanced.bodyMode') }}</label>
      <div class="components-admin-monitor-monitor-advanced-request-config__panel-4">
        <button
          v-for="opt in bodyModeOptions"
          :key="opt.value"
          type="button"
          class="components-admin-monitor-monitor-advanced-request-config__action-3"
          :class="bodyModeButtonClass(opt.value)"
          @click="updateBodyMode(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
      <p class="components-admin-monitor-monitor-advanced-request-config__description-2">
        {{ bodyModeHint }}
      </p>
    </div>

    <!-- Body JSON (仅当 mode != off) -->
    <div v-if="bodyOverrideMode !== 'off'">
      <div class="components-admin-monitor-monitor-advanced-request-config__panel-5">
        <label class="components-admin-monitor-monitor-advanced-request-config__label input-label">{{ t('admin.channelMonitor.advanced.bodyJson') }}</label>
        <button
          type="button"
          class="components-admin-monitor-monitor-advanced-request-config__action-4"
          :disabled="!bodyText.trim()"
          @click="formatBody"
        >
          {{ t('admin.channelMonitor.advanced.bodyJsonFormat') }}
        </button>
      </div>
      <textarea
        v-model="bodyText"
        rows="10"
        :placeholder="bodyPlaceholder"
        class="components-admin-monitor-monitor-advanced-request-config__field-3 input"
        style="white-space: pre; overflow-wrap: normal; overflow-x: auto;"
        spellcheck="false"
        @blur="commitBody"
      />
      <p v-if="bodyError" class="components-admin-monitor-monitor-advanced-request-config__description">{{ bodyError }}</p>
      <p v-else class="components-admin-monitor-monitor-advanced-request-config__description-2">
        {{ t('admin.channelMonitor.advanced.bodyJsonHint') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { APIMode, BodyOverrideMode, Provider } from '@/api/admin/channelMonitor'
import {
  API_MODE_RESPONSES,
  DEFAULT_GROK_MODEL,
  PROVIDER_GROK,
  PROVIDER_OPENAI,
} from '@/constants/channelMonitor'

const props = defineProps<{
  provider?: Provider
  apiMode?: APIMode
  extraHeaders: Record<string, string>
  bodyOverrideMode: BodyOverrideMode
  bodyOverride: Record<string, unknown> | null
}>()

const emit = defineEmits<{
  (e: 'update:extraHeaders', value: Record<string, string>): void
  (e: 'update:bodyOverrideMode', value: BodyOverrideMode): void
  (e: 'update:bodyOverride', value: Record<string, unknown> | null): void
}>()

const { t } = useI18n()

// ---- Headers key-value rows ----
interface HeaderRow {
  name: string
  value: string
}

const headerRows = ref<HeaderRow[]>(toRows(props.extraHeaders))
const headersError = ref('')

watch(
  () => props.extraHeaders,
  (v) => {
    // 外部重置时（切换平台 / 应用模板）同步行。
    // 同值不回写，避免每次 commit 都把行重排。
    if (!isSameHeaderMap(toMap(headerRows.value), v)) {
      headerRows.value = toRows(v)
    }
    headersError.value = ''
  },
)

function toRows(h: Record<string, string>): HeaderRow[] {
  const entries = Object.entries(h || {})
  if (entries.length === 0) return [{ name: '', value: '' }]
  return entries.map(([name, value]) => ({ name, value }))
}

function toMap(rows: HeaderRow[]): Record<string, string> {
  const out: Record<string, string> = {}
  for (const row of rows) {
    const name = row.name.trim()
    if (name === '') continue
    out[name] = row.value
  }
  return out
}

function isSameHeaderMap(a: Record<string, string>, b: Record<string, string>): boolean {
  const ak = Object.keys(a)
  const bk = Object.keys(b || {})
  if (ak.length !== bk.length) return false
  for (const k of ak) {
    if (a[k] !== b[k]) return false
  }
  return true
}

function commitHeaders() {
  // 空白 name + 空白 value 的行允许保留作为"占位新行"，不报错；
  // name 非空但 value 为空（或反之）都视为用户正在编辑，同样不报错。
  // 只在 name 里含冒号这种明显不合法时兜一下。
  for (const row of headerRows.value) {
    const name = row.name.trim()
    if (name === '') continue
    if (name.includes(':') || /\s/.test(name)) {
      headersError.value = t('admin.channelMonitor.advanced.headerNameInvalid', { name })
      return
    }
  }
  headersError.value = ''
  emit('update:extraHeaders', toMap(headerRows.value))
}

function addRow() {
  headerRows.value.push({ name: '', value: '' })
}

function removeRow(index: number) {
  headerRows.value.splice(index, 1)
  if (headerRows.value.length === 0) {
    headerRows.value.push({ name: '', value: '' })
  }
  commitHeaders()
}

// ---- Body mode + JSON ----
const bodyText = ref(serializeBody(props.bodyOverride))
const bodyError = ref('')

watch(
  () => props.bodyOverride,
  (v) => {
    bodyText.value = serializeBody(v)
    bodyError.value = ''
  },
)

function commitBody() {
  if (props.bodyOverrideMode === 'off') {
    return
  }
  const trimmed = bodyText.value.trim()
  if (trimmed === '') {
    emit('update:bodyOverride', null)
    bodyError.value = ''
    return
  }
  try {
    const parsed = JSON.parse(trimmed)
    if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
      bodyError.value = t('admin.channelMonitor.advanced.bodyJsonObjectError')
      return
    }
    emit('update:bodyOverride', parsed as Record<string, unknown>)
    bodyError.value = ''
  } catch (e) {
    bodyError.value =
      t('admin.channelMonitor.advanced.bodyJsonError') +
      ': ' +
      (e instanceof Error ? e.message : String(e))
  }
}

function formatBody() {
  const trimmed = bodyText.value.trim()
  if (trimmed === '') return
  try {
    const parsed = JSON.parse(trimmed)
    bodyText.value = JSON.stringify(parsed, null, 2)
    bodyError.value = ''
    // 同步把校验过的对象提交，避免格式化后焦点未移走时父组件读到旧值
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      emit('update:bodyOverride', parsed as Record<string, unknown>)
    }
  } catch (e) {
    bodyError.value =
      t('admin.channelMonitor.advanced.bodyJsonError') +
      ': ' +
      (e instanceof Error ? e.message : String(e))
  }
}

function serializeBody(body: Record<string, unknown> | null): string {
  if (!body || Object.keys(body).length === 0) return ''
  return JSON.stringify(body, null, 2)
}

function updateBodyMode(mode: BodyOverrideMode) {
  emit('update:bodyOverrideMode', mode)
  // 切换到 off 时清掉 body（提示用户）
  if (mode === 'off') {
    emit('update:bodyOverride', null)
  }
}

const bodyModeOptions = computed<{ value: BodyOverrideMode; label: string }[]>(() => [
  { value: 'off', label: t('admin.channelMonitor.advanced.bodyModeOff') },
  { value: 'merge', label: t('admin.channelMonitor.advanced.bodyModeMerge') },
  { value: 'replace', label: t('admin.channelMonitor.advanced.bodyModeReplace') },
])

function bodyModeButtonClass(mode: BodyOverrideMode): string {
  const active = props.bodyOverrideMode === mode
  if (active) {
    return 'components-admin-monitor-monitor-advanced-request-config__state'
  }
  return 'components-admin-monitor-monitor-advanced-request-config__state-2'
}

const bodyModeHint = computed(() => {
  switch (props.bodyOverrideMode) {
    case 'merge':
      return t('admin.channelMonitor.advanced.bodyModeHintMerge')
    case 'replace':
      return t('admin.channelMonitor.advanced.bodyModeHintReplace')
    default:
      return t('admin.channelMonitor.advanced.bodyModeHintOff')
  }
})

const bodyPlaceholder = computed(() => {
  if (props.provider === PROVIDER_OPENAI && props.apiMode === API_MODE_RESPONSES) {
    if (props.bodyOverrideMode === 'merge') {
      return '{\n  "max_output_tokens": 20\n}'
    }
    return '{\n  "model": "gpt-4o-mini",\n  "instructions": "You are a health check endpoint. Reply briefly.",\n  "input": "Reply with exactly: ok",\n  "max_output_tokens": 20,\n  "stream": false\n}'
  }
  if (props.provider === PROVIDER_OPENAI || props.provider === PROVIDER_GROK) {
    if (props.bodyOverrideMode === 'merge') {
      return '{\n  "max_tokens": 20\n}'
    }
    const model = props.provider === PROVIDER_GROK ? DEFAULT_GROK_MODEL : 'gpt-4o-mini'
    return `{\n  "model": "${model}",\n  "messages": [{"role":"user","content":"Reply with exactly: ok"}],\n  "max_tokens": 20,\n  "stream": false\n}`
  }
  if (props.bodyOverrideMode === 'merge') {
    return '{\n  "system": "You are Claude Code..."\n}'
  }
  return '{\n  "model": "claude-x",\n  "messages": [{"role":"user","content":"hi"}],\n  "max_tokens": 10\n}'
})
</script>
