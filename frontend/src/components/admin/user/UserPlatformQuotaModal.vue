<template>
  <BaseDialog
    :show="show"
    :title="t('admin.users.platformQuota.title')"
    width="wide"
    @close="$emit('close')"
  >
    <div v-if="user" class="components-admin-user-user-platform-quota-modal__panel">
      <div
        v-if="hasActiveSubscription"
        class="components-admin-user-user-platform-quota-modal__panel-2"
      >
        {{ t('admin.users.platformQuota.subscriptionWarning') }}
      </div>
      <p class="components-admin-user-user-platform-quota-modal__description">
        {{ t('admin.users.platformQuota.subtitle', { email: user.email }) }}
      </p>
      <div v-if="loading" class="components-admin-user-user-platform-quota-modal__panel-3">{{ t('common.loading') }}</div>
      <div v-else class="components-admin-user-user-platform-quota-modal__panel-4">
        <table class="components-admin-user-user-platform-quota-modal__table">
          <thead>
            <tr class="components-admin-user-user-platform-quota-modal__row">
              <th class="components-admin-user-user-platform-quota-modal__heading">{{ t('admin.users.platformQuota.columns.platform') }}</th>
              <th class="components-admin-user-user-platform-quota-modal__heading">{{ t('admin.users.platformQuota.columns.daily') }}</th>
              <th class="components-admin-user-user-platform-quota-modal__heading">{{ t('admin.users.platformQuota.columns.weekly') }}</th>
              <th class="components-admin-user-user-platform-quota-modal__heading">{{ t('admin.users.platformQuota.columns.monthly') }}</th>
              <th class="components-admin-user-user-platform-quota-modal__heading">{{ t('admin.users.platformQuota.columns.usage') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in quotas" :key="row.platform" class="components-admin-user-user-platform-quota-modal__row-2">
              <td class="components-admin-user-user-platform-quota-modal__cell">{{ row.platform }}</td>
              <td class="components-admin-user-user-platform-quota-modal__cell-2">
                <div class="components-admin-user-user-platform-quota-modal__panel-5">
                  <input
                    v-model.number="row.daily_limit_points"
                    type="number"
                    min="0"
                    step="0.01"
                    class="components-admin-user-user-platform-quota-modal__field input"
                    :placeholder="t('admin.users.platformQuota.placeholder')"
                  />
                  <button
                    type="button"
                    class="components-admin-user-user-platform-quota-modal__action"
                    :disabled="!!resetting[`${row.platform}.daily`]"
                    :title="t('admin.users.platformQuota.reset.button')"
                    @click="onReset(row.platform, 'daily')"
                  >↻</button>
                </div>
              </td>
              <td class="components-admin-user-user-platform-quota-modal__cell-2">
                <div class="components-admin-user-user-platform-quota-modal__panel-5">
                  <input
                    v-model.number="row.weekly_limit_points"
                    type="number"
                    min="0"
                    step="0.01"
                    class="components-admin-user-user-platform-quota-modal__field input"
                    :placeholder="t('admin.users.platformQuota.placeholder')"
                  />
                  <button
                    type="button"
                    class="components-admin-user-user-platform-quota-modal__action"
                    :disabled="!!resetting[`${row.platform}.weekly`]"
                    :title="t('admin.users.platformQuota.reset.button')"
                    @click="onReset(row.platform, 'weekly')"
                  >↻</button>
                </div>
              </td>
              <td class="components-admin-user-user-platform-quota-modal__cell-2">
                <div class="components-admin-user-user-platform-quota-modal__panel-5">
                  <input
                    v-model.number="row.monthly_limit_points"
                    type="number"
                    min="0"
                    step="0.01"
                    class="components-admin-user-user-platform-quota-modal__field input"
                    :placeholder="t('admin.users.platformQuota.placeholder')"
                  />
                  <button
                    type="button"
                    class="components-admin-user-user-platform-quota-modal__action"
                    :disabled="!!resetting[`${row.platform}.monthly`]"
                    :title="t('admin.users.platformQuota.reset.button')"
                    @click="onReset(row.platform, 'monthly')"
                  >↻</button>
                </div>
              </td>
              <td class="components-admin-user-user-platform-quota-modal__cell-3">
                {{ formatUsage(row.daily_usage_points) }} / {{ formatUsage(row.weekly_usage_points) }} / {{ formatUsage(row.monthly_usage_points) }}
              </td>
            </tr>
          </tbody>
        </table>
        <p class="components-admin-user-user-platform-quota-modal__description-2">{{ t('admin.users.platformQuota.hint') }}</p>
        <div class="components-admin-user-user-platform-quota-modal__panel-6">
          <button type="button" class="components-admin-user-user-platform-quota-modal__action-2 btn btn-secondary" @click="onClearAll">
            {{ t('admin.users.platformQuota.clearAll') }}
          </button>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="components-admin-user-user-platform-quota-modal__panel-7">
        <button type="button" class="btn btn-secondary" @click="$emit('close')">
          {{ t('admin.users.platformQuota.cancel') }}
        </button>
        <button type="button" class="btn btn-primary" :disabled="submitting || loading" @click="onSave">
          {{ submitting ? t('admin.users.platformQuota.saving') : t('admin.users.platformQuota.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser, PlatformQuotaItem, PlatformQuotaPlatform, PlatformQuotaWindow } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { formatPoints } from '@/utils/format'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close', 'success'])

const { t } = useI18n()
const appStore = useAppStore()

const PLATFORMS: PlatformQuotaPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']

interface QuotaRow {
  platform: PlatformQuotaPlatform
  daily_limit_points: number | null
  weekly_limit_points: number | null
  monthly_limit_points: number | null
  daily_usage_points: number
  weekly_usage_points: number
  monthly_usage_points: number
}

const hasActiveSubscription = computed(() =>
  props.user?.subscriptions?.some((s) => s.status === 'active') ?? false
)

const loading = ref(false)
const submitting = ref(false)
const resetting = reactive<Record<string, boolean>>({})
const quotas = ref<QuotaRow[]>([])

function emptyRow(p: PlatformQuotaPlatform): QuotaRow {
  return {
    platform: p,
    daily_limit_points: null,
    weekly_limit_points: null,
    monthly_limit_points: null,
    daily_usage_points: 0,
    weekly_usage_points: 0,
    monthly_usage_points: 0,
  }
}

function normalize(items: PlatformQuotaItem[]): QuotaRow[] {
  const byPlatform = new Map<PlatformQuotaPlatform, PlatformQuotaItem>()
  for (const it of items) byPlatform.set(it.platform, it)
  return PLATFORMS.map((p) => {
    const it = byPlatform.get(p)
    if (!it) return emptyRow(p)
    return {
      platform: p,
      daily_limit_points: it.daily_limit_points ?? it.daily_limit_usd ?? null,
      weekly_limit_points: it.weekly_limit_points ?? it.weekly_limit_usd ?? null,
      monthly_limit_points: it.monthly_limit_points ?? it.monthly_limit_usd ?? null,
      daily_usage_points: it.daily_usage_points ?? it.daily_usage_usd ?? 0,
      weekly_usage_points: it.weekly_usage_points ?? it.weekly_usage_usd ?? 0,
      monthly_usage_points: it.monthly_usage_points ?? it.monthly_usage_usd ?? 0,
    }
  })
}

function formatUsage(n: number): string {
  if (n == null || Number.isNaN(n)) return '-'
  return formatPoints(n)
}

async function load() {
  if (!props.user) return
  loading.value = true
  try {
    const data = await adminAPI.users.getPlatformQuotas(props.user.id)
    quotas.value = normalize(data.platform_quotas || [])
  } catch {
    appStore.showError(t('admin.users.platformQuota.loadFailed'))
    quotas.value = PLATFORMS.map(emptyRow)
  } finally {
    loading.value = false
  }
}

watch(
  () => props.show,
  (s) => { if (s && props.user) load() },
)

function onClearAll() {
  // 二次确认：一键清空全部平台的 daily/weekly/monthly 限额属于高风险批量操作，
  // 误点后所有平台变为"无限额"，且本地无 undo 机制（需要逐个手动重填或取消保存）。
  const confirmed = window.confirm(t('admin.users.platformQuota.clearAllConfirm'))
  if (!confirmed) return
  for (const row of quotas.value) {
    row.daily_limit_points = null
    row.weekly_limit_points = null
    row.monthly_limit_points = null
  }
}

async function onSave() {
  if (!props.user) return
  // 校验所有 input：v-model.number 在用户输入"0."等中间状态时会写回 NaN，
  // 之前的 normalizeLimit(NaN) 静默返回 null（"无限制"），把"有限额"配置悄悄改成"无限制"。
  // 这里在 save 前显式检测 NaN，提示用户修正后再提交。
  const invalid: string[] = []
  for (const row of quotas.value) {
    for (const win of ['daily', 'weekly', 'monthly'] as const) {
      const v = row[`${win}_limit_points` as const]
      if (typeof v === 'number' && Number.isNaN(v)) {
        invalid.push(`${row.platform}.${win}`)
      }
    }
  }
  if (invalid.length > 0) {
    appStore.showError(t('admin.users.platformQuota.invalidNumber', { fields: invalid.join(', ') }))
    return
  }

  submitting.value = true
  try {
    const payload = quotas.value.map((r) => ({
      platform: r.platform,
      daily_limit_points: normalizeLimit(r.daily_limit_points),
      weekly_limit_points: normalizeLimit(r.weekly_limit_points),
      monthly_limit_points: normalizeLimit(r.monthly_limit_points),
    }))
    await adminAPI.users.updatePlatformQuotas(props.user.id, payload)
    appStore.showSuccess(t('admin.users.platformQuota.updateSuccess'))
    emit('success')
    emit('close')
  } catch (e: any) {
    appStore.showError(e?.response?.data?.message || t('admin.users.platformQuota.updateFailed'))
  } finally {
    submitting.value = false
  }
}

// 仅在合法输入下返回数字：null/undefined/NaN/±Inf/负数 → null（视为"无限额"）。
// 调用方负责在 NaN 路径上做单独的用户提示（见 onSave）。
function normalizeLimit(v: number | null | undefined): number | null {
  if (v === null || v === undefined) return null
  if (typeof v === 'number' && Number.isFinite(v) && v >= 0) return v
  return null
}

async function onReset(platform: PlatformQuotaPlatform, quotaWindow: PlatformQuotaWindow) {
  if (!props.user) return
  const windowLabel = t(`admin.users.platformQuota.window${quotaWindow.charAt(0).toUpperCase() + quotaWindow.slice(1)}`)
  const confirmed = window.confirm(
    t('admin.users.platformQuota.reset.confirm', { platform, window: windowLabel })
  )
  if (!confirmed) return
  const key = `${platform}.${quotaWindow}`
  resetting[key] = true
  try {
    const data = await adminAPI.users.resetPlatformQuotaWindow(props.user.id, platform, quotaWindow)
    quotas.value = normalize(data.platform_quotas || [])
    appStore.showSuccess(t('admin.users.platformQuota.reset.success', { platform, window: windowLabel }))
  } catch (e: any) {
    appStore.showError(e?.response?.data?.message || t('admin.users.platformQuota.reset.failed'))
  } finally {
    resetting[key] = false
  }
}
</script>
