<template>
    <div class="views-admin-backup-view__panel settings-backup-content">
      <!-- S3 Storage Config -->
      <div class="views-admin-backup-view__panel-2 card-body card">
        <div class="views-admin-backup-view__panel-3">
          <div>
            <h3 class="views-admin-backup-view__heading">
              {{ t('admin.backup.s3.title') }}
            </h3>
            <p class="views-admin-backup-view__description">
              {{ t('admin.backup.s3.descriptionPrefix') }}
              <button type="button" class="views-admin-backup-view__action" @click="showR2Guide = true">Cloudflare R2</button>
              {{ t('admin.backup.s3.descriptionSuffix') }}
            </p>
          </div>
        </div>
        <div class="views-admin-backup-view__panel-4">
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.endpoint') }}</label>
            <input v-model="s3Form.endpoint" class="views-admin-backup-view__field input" placeholder="https://<account_id>.r2.cloudflarestorage.com" />
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.region') }}</label>
            <input v-model="s3Form.region" class="views-admin-backup-view__field input" placeholder="auto" />
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.bucket') }}</label>
            <input v-model="s3Form.bucket" class="views-admin-backup-view__field input" />
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.prefix') }}</label>
            <input v-model="s3Form.prefix" class="views-admin-backup-view__field input" placeholder="backups/" />
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.accessKeyId') }}</label>
            <input v-model="s3Form.access_key_id" class="views-admin-backup-view__field input" />
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.secretAccessKey') }}</label>
            <input v-model="s3Form.secret_access_key" type="password" class="views-admin-backup-view__field input" :placeholder="s3SecretConfigured ? t('admin.backup.s3.secretConfigured') : ''" />
          </div>
          <label class="views-admin-backup-view__label-2">
            <input v-model="s3Form.force_path_style" type="checkbox" />
            <span>{{ t('admin.backup.s3.forcePathStyle') }}</span>
          </label>
        </div>
        <div class="views-admin-backup-view__panel-5">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="testingS3" @click="testS3">
            {{ testingS3 ? t('common.loading') : t('admin.backup.s3.testConnection') }}
          </button>
          <button type="button" class="btn btn-primary btn-sm" :disabled="savingS3" @click="saveS3Config">
            {{ savingS3 ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>

      <!-- Async image object storage -->
      <div class="views-admin-backup-view__panel-2 card-body card">
        <div class="views-admin-backup-view__panel-3">
          <div>
            <h3 class="views-admin-backup-view__heading">
              {{ t('admin.backup.imageStorage.title') }}
            </h3>
            <p class="views-admin-backup-view__description">
              {{ t('admin.backup.imageStorage.description') }}
            </p>
          </div>
          <label class="views-admin-backup-view__label-3">
            <input v-model="imageStorageForm.enabled" type="checkbox" />
            <span>{{ t('admin.backup.imageStorage.enabled') }}</span>
          </label>
        </div>

        <label class="views-admin-backup-view__label-3">
          <input v-model="imageStorageForm.reuse_backup_s3" type="checkbox" />
          <span>{{ t('admin.backup.imageStorage.reuseBackupS3') }}</span>
        </label>

        <div class="views-admin-backup-view__panel-6">
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.imageStorage.bucket') }}</label>
            <input v-model="imageStorageForm.bucket" class="views-admin-backup-view__field input" :placeholder="imageStorageForm.reuse_backup_s3 ? t('admin.backup.imageStorage.bucketInherited') : ''" />
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.imageStorage.prefix') }}</label>
            <input v-model="imageStorageForm.prefix" class="views-admin-backup-view__field input" placeholder="images/" />
          </div>

          <template v-if="!imageStorageForm.reuse_backup_s3">
            <div>
              <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.endpoint') }}</label>
              <input v-model="imageStorageForm.endpoint" class="views-admin-backup-view__field input" placeholder="https://<account_id>.r2.cloudflarestorage.com" />
            </div>
            <div>
              <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.region') }}</label>
              <input v-model="imageStorageForm.region" class="views-admin-backup-view__field input" placeholder="auto" />
            </div>
            <div>
              <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.accessKeyId') }}</label>
              <input v-model="imageStorageForm.access_key_id" class="views-admin-backup-view__field input" />
            </div>
            <div>
              <label class="views-admin-backup-view__label">{{ t('admin.backup.s3.secretAccessKey') }}</label>
              <input v-model="imageStorageForm.secret_access_key" type="password" class="views-admin-backup-view__field input" :placeholder="imageStorageSecretConfigured ? t('admin.backup.s3.secretConfigured') : ''" />
            </div>
            <label class="views-admin-backup-view__label-2">
              <input v-model="imageStorageForm.force_path_style" type="checkbox" />
              <span>{{ t('admin.backup.s3.forcePathStyle') }}</span>
            </label>
          </template>

          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.imageStorage.publicBaseUrl') }}</label>
            <input v-model="imageStorageForm.public_base_url" class="views-admin-backup-view__field input" :placeholder="t('admin.backup.imageStorage.publicBaseUrlPlaceholder')" />
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.imageStorage.presignExpiryHours') }}</label>
            <input v-model.number="imageStorageForm.presign_expiry_hours" type="number" min="1" class="views-admin-backup-view__field input" />
          </div>
        </div>

        <div class="views-admin-backup-view__panel-5">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="testingImageStorage" @click="testImageStorage">
            {{ testingImageStorage ? t('common.loading') : t('admin.backup.s3.testConnection') }}
          </button>
          <button type="button" class="btn btn-primary btn-sm" :disabled="savingImageStorage" @click="saveImageStorageConfig">
            {{ savingImageStorage ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>

      <!-- Schedule Config -->
      <div class="views-admin-backup-view__panel-2 card-body card">
        <div class="views-admin-backup-view__panel-7">
          <h3 class="views-admin-backup-view__heading">
            {{ t('admin.backup.schedule.title') }}
          </h3>
          <p class="views-admin-backup-view__description">
            {{ t('admin.backup.schedule.description') }}
          </p>
        </div>
        <div class="views-admin-backup-view__panel-4">
          <label class="views-admin-backup-view__label-2">
            <input v-model="scheduleForm.enabled" type="checkbox" />
            <span>{{ t('admin.backup.schedule.enabled') }}</span>
          </label>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.schedule.cronExpr') }}</label>
            <input v-model="scheduleForm.cron_expr" class="views-admin-backup-view__field input" placeholder="0 2 * * *" />
            <p class="views-admin-backup-view__description-2">{{ t('admin.backup.schedule.cronHint') }}</p>
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.schedule.retainDays') }}</label>
            <input v-model.number="scheduleForm.retain_days" type="number" min="0" class="views-admin-backup-view__field input" />
            <p class="views-admin-backup-view__description-2">{{ t('admin.backup.schedule.retainDaysHint') }}</p>
          </div>
          <div>
            <label class="views-admin-backup-view__label">{{ t('admin.backup.schedule.retainCount') }}</label>
            <input v-model.number="scheduleForm.retain_count" type="number" min="0" class="views-admin-backup-view__field input" />
            <p class="views-admin-backup-view__description-2">{{ t('admin.backup.schedule.retainCountHint') }}</p>
          </div>
        </div>
        <div class="views-admin-backup-view__panel-8">
          <button type="button" class="btn btn-primary btn-sm" :disabled="savingSchedule" @click="saveSchedule">
            {{ savingSchedule ? t('common.loading') : t('common.save') }}
          </button>
        </div>
      </div>

      <!-- Backup Operations -->
      <div class="views-admin-backup-view__panel-2 card-body card">
        <div class="views-admin-backup-view__panel-3">
          <div>
            <h3 class="views-admin-backup-view__heading">
              {{ t('admin.backup.operations.title') }}
            </h3>
            <p class="views-admin-backup-view__description">
              {{ t('admin.backup.operations.description') }}
            </p>
          </div>
          <div class="views-admin-backup-view__panel-9">
            <div class="views-admin-backup-view__panel-10">
              <label class="views-admin-backup-view__label-4">{{ t('admin.backup.operations.expireDays') }}</label>
              <input v-model.number="manualExpireDays" type="number" min="0" class="views-admin-backup-view__field-2 input" />
            </div>
            <button type="button" class="btn btn-primary btn-sm" :disabled="creatingBackup" @click="createBackup">
              {{ creatingBackup ? t('admin.backup.operations.backing') : t('admin.backup.operations.createBackup') }}
            </button>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="loadingBackups" @click="loadBackups">
              {{ loadingBackups ? t('common.loading') : t('common.refresh') }}
            </button>
          </div>
        </div>

        <div class="views-admin-backup-view__panel-11">
          <table class="views-admin-backup-view__table">
            <thead>
              <tr class="views-admin-backup-view__row">
                <th class="views-admin-backup-view__heading-2">ID</th>
                <th class="views-admin-backup-view__heading-2">{{ t('admin.backup.columns.status') }}</th>
                <th class="views-admin-backup-view__heading-2">{{ t('admin.backup.columns.fileName') }}</th>
                <th class="views-admin-backup-view__heading-2">{{ t('admin.backup.columns.size') }}</th>
                <th class="views-admin-backup-view__heading-2">{{ t('admin.backup.columns.parts') }}</th>
                <th class="views-admin-backup-view__heading-2">{{ t('admin.backup.columns.expiresAt') }}</th>
                <th class="views-admin-backup-view__heading-2">{{ t('admin.backup.columns.triggeredBy') }}</th>
                <th class="views-admin-backup-view__heading-2">{{ t('admin.backup.columns.startedAt') }}</th>
                <th class="views-admin-backup-view__heading-3">{{ t('admin.backup.columns.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="record in backups" :key="record.id" class="views-admin-backup-view__row-2">
                <td class="views-admin-backup-view__cell">{{ record.id }}</td>
                <td class="views-admin-backup-view__cell-2">
                  <span
                    class="views-admin-backup-view__text"
                    :class="statusClass(record.status)"
                  >
                    {{ record.status === 'running' && record.progress
                      ? t(`admin.backup.progress.${record.progress}`)
                      : t(`admin.backup.status.${record.status}`) }}
                  </span>
                </td>
                <td class="views-admin-backup-view__cell-3">{{ record.file_name }}</td>
                <td class="views-admin-backup-view__cell-3">{{ formatSize(record.size_bytes) }}</td>
                <td class="views-admin-backup-view__cell-3">{{ record.parts?.length || (record.status === 'running' ? '-' : 1) }}</td>
                <td class="views-admin-backup-view__cell-3">
                  {{ record.expires_at ? formatDate(record.expires_at) : t('admin.backup.neverExpire') }}
                </td>
                <td class="views-admin-backup-view__cell-3">
                  {{ record.triggered_by === 'scheduled' ? t('admin.backup.trigger.scheduled') : t('admin.backup.trigger.manual') }}
                </td>
                <td class="views-admin-backup-view__cell-3">{{ formatDate(record.started_at) }}</td>
                <td class="views-admin-backup-view__cell-4">
                  <div class="views-admin-backup-view__panel-12">
                    <button
                      v-if="record.status === 'completed'"
                      type="button"
                      class="btn btn-secondary btn-xs"
                      @click="downloadBackup(record.id)"
                    >
                      {{ t('admin.backup.actions.download') }}
                    </button>
                    <button
                      v-if="record.status === 'completed'"
                      type="button"
                      class="btn btn-secondary btn-xs"
                      :disabled="restoringId === record.id"
                      @click="restoreBackup(record.id)"
                    >
                      {{ restoringId === record.id ? t('common.loading') : t('admin.backup.actions.restore') }}
                    </button>
                    <button
                      v-if="record.status !== 'running'"
                      type="button"
                      class="btn btn-danger btn-xs"
                      @click="removeBackup(record.id)"
                    >
                      {{ t('common.delete') }}
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="backups.length === 0">
                <td colspan="9" class="views-admin-backup-view__cell-5">
                  {{ t('admin.backup.empty') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Cloudflare R2 Setup Guide Modal -->
    <BaseDialog
      :show="showR2Guide"
      :title="t('admin.backup.r2Guide.title')"
      width="wide"
      close-on-click-outside
      @close="showR2Guide = false"
    >
      <div class="r2-guide-dialog">
        <p class="r2-guide-dialog__intro">{{ t('admin.backup.r2Guide.intro') }}</p>

        <!-- Step 1 -->
        <div class="r2-guide-step">
          <div class="r2-guide-step__header">
            <span class="r2-guide-step__badge">1</span>
            <h4 class="r2-guide-step__title">{{ t('admin.backup.r2Guide.step1.title') }}</h4>
          </div>
          <ol class="r2-guide-step__list">
            <li>{{ t('admin.backup.r2Guide.step1.line1') }}</li>
            <li>{{ t('admin.backup.r2Guide.step1.line2') }}</li>
            <li>{{ t('admin.backup.r2Guide.step1.line3') }}</li>
          </ol>
        </div>

        <!-- Step 2 -->
        <div class="r2-guide-step">
          <div class="r2-guide-step__header">
            <span class="r2-guide-step__badge">2</span>
            <h4 class="r2-guide-step__title">{{ t('admin.backup.r2Guide.step2.title') }}</h4>
          </div>
          <ol class="r2-guide-step__list">
            <li>{{ t('admin.backup.r2Guide.step2.line1') }}</li>
            <li>{{ t('admin.backup.r2Guide.step2.line2') }}</li>
            <li>{{ t('admin.backup.r2Guide.step2.line3') }}</li>
            <li>{{ t('admin.backup.r2Guide.step2.line4') }}</li>
          </ol>
          <div class="r2-guide-alert r2-guide-alert--warning">
            <Icon name="exclamationTriangle" size="xs" />
            <span>{{ t('admin.backup.r2Guide.step2.warning') }}</span>
          </div>
        </div>

        <!-- Step 3 -->
        <div class="r2-guide-step">
          <div class="r2-guide-step__header">
            <span class="r2-guide-step__badge">3</span>
            <h4 class="r2-guide-step__title">{{ t('admin.backup.r2Guide.step3.title') }}</h4>
          </div>
          <p class="r2-guide-step__desc">{{ t('admin.backup.r2Guide.step3.desc') }}</p>
          <div class="r2-guide-code-box">
            <code>https://&lt;{{ t('admin.backup.r2Guide.step3.accountId') }}&gt;.r2.cloudflarestorage.com</code>
          </div>
        </div>

        <!-- Step 4: Fill form -->
        <div class="r2-guide-step">
          <div class="r2-guide-step__header">
            <span class="r2-guide-step__badge">4</span>
            <h4 class="r2-guide-step__title">{{ t('admin.backup.r2Guide.step4.title') }}</h4>
          </div>
          <div class="r2-guide-table-wrap">
            <table class="table r2-guide-table">
              <tbody>
                <tr v-for="(row, i) in r2ConfigRows" :key="i">
                  <td class="r2-guide-table__key">{{ row.field }}</td>
                  <td class="r2-guide-table__val"><code>{{ row.value }}</code></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Free tier note -->
        <div class="r2-guide-alert r2-guide-alert--info">
          <Icon name="infoCircle" size="xs" />
          <span>{{ t('admin.backup.r2Guide.freeTier') }}</span>
        </div>
      </div>

      <template #footer>
        <button type="button" class="btn btn-secondary btn-sm" @click="showR2Guide = false">{{ t('common.close') }}</button>
      </template>
    </BaseDialog>

    <!-- 分卷下载链接 -->
    <BaseDialog
      :show="downloadPartsModalOpen"
      :title="t('admin.backup.actions.downloadParts')"
      width="normal"
      close-on-click-outside
      @close="closeDownloadParts"
    >
      <div class="parts-dialog">
        <p class="parts-dialog__hint">{{ t('admin.backup.actions.downloadPartsHint') }}</p>
        <div class="parts-dialog__list">
          <div
            v-for="part in downloadParts"
            :key="part.index"
            class="parts-dialog__item"
          >
            <div class="parts-dialog__meta">
              <span class="parts-dialog__name">{{ t('admin.backup.actions.partLabel', { index: part.index }) }}</span>
              <span class="parts-dialog__size">{{ formatSize(part.size_bytes) }}</span>
            </div>
            <a :href="part.url" class="btn btn-secondary btn-xs" rel="noopener">
              {{ t('admin.backup.actions.download') }}
            </a>
          </div>
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary btn-sm" @click="closeDownloadParts">{{ t('common.close') }}</button>
      </template>
    </BaseDialog>
    <TotpStepUpDialog :controller="backupStepUp" />
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import { useAppStore } from '@/stores'
import type {
  BackupS3Config,
  BackupScheduleConfig,
  BackupRecord,
  BackupDownloadPart,
  ImageStorageConfig,
} from '@/api/admin/backup'
import { useStepUp, isStepUpBlocked, isStepUpCancelled, stepUpBlockReason } from '@/composables/useStepUp'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()
const backupStepUp = useStepUp()

// 敏感操作被 2FA 门控拦截时的统一提示。
function reportStepUpBlocked(error: unknown): boolean {
  if (!isStepUpBlocked(error)) return false
  appStore.showError(
    stepUpBlockReason(error) === 'STEP_UP_ADMIN_API_KEY_FORBIDDEN'
      ? t('stepUp.adminApiKeyForbidden')
      : t('stepUp.notEnabled')
  )
  return true
}

// S3 config
const s3Form = ref<BackupS3Config>({
  endpoint: '',
  region: 'auto',
  bucket: '',
  access_key_id: '',
  secret_access_key: '',
  prefix: 'backups/',
  force_path_style: false,
})
const s3SecretConfigured = ref(false)
const savingS3 = ref(false)
const testingS3 = ref(false)

// Async image object storage. Shares the S3 client with backups, so the default is
// to reuse the credentials configured above and only differ by prefix.
const imageStorageForm = ref<ImageStorageConfig>({
  enabled: false,
  reuse_backup_s3: true,
  bucket: '',
  prefix: 'images/',
  public_base_url: '',
  presign_expiry_hours: 24,
  max_download_bytes: 33554432,
  endpoint: '',
  region: 'auto',
  access_key_id: '',
  secret_access_key: '',
  force_path_style: false,
})
const imageStorageSecretConfigured = ref(false)
const savingImageStorage = ref(false)
const testingImageStorage = ref(false)

// Schedule config
const scheduleForm = ref<BackupScheduleConfig>({
  enabled: false,
  cron_expr: '0 2 * * *',
  retain_days: 14,
  retain_count: 10,
})
const savingSchedule = ref(false)

// Backups
const backups = ref<BackupRecord[]>([])
const loadingBackups = ref(false)
const creatingBackup = ref(false)
const restoringId = ref('')
const manualExpireDays = ref(14)
const downloadParts = ref<BackupDownloadPart[]>([])
const downloadPartsModalOpen = ref(false)

// Polling
const pollingTimer = ref<ReturnType<typeof setInterval> | null>(null)
const restoringPollingTimer = ref<ReturnType<typeof setInterval> | null>(null)
const MAX_POLL_COUNT = 900

function updateRecordInList(updated: BackupRecord) {
  const idx = backups.value.findIndex(r => r.id === updated.id)
  if (idx >= 0) {
    backups.value[idx] = updated
  }
}

function startPolling(backupId: string) {
  stopPolling()
  let count = 0
  pollingTimer.value = setInterval(async () => {
    if (count++ >= MAX_POLL_COUNT) {
      stopPolling()
      creatingBackup.value = false
      appStore.showWarning(t('admin.backup.operations.backupRunning'))
      return
    }
    try {
      const record = await adminAPI.backup.getBackup(backupId)
      updateRecordInList(record)
      if (record.status === 'completed' || record.status === 'failed') {
        stopPolling()
        creatingBackup.value = false
        if (record.status === 'completed') {
          appStore.showSuccess(t('admin.backup.operations.backupCreated'))
        } else {
          appStore.showError(record.error_message || t('admin.backup.operations.backupFailed'))
        }
        await loadBackups()
      }
    } catch {
      // 轮询失败时不中断
    }
  }, 2000)
}

function stopPolling() {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value)
    pollingTimer.value = null
  }
}

function startRestorePolling(backupId: string) {
  stopRestorePolling()
  let count = 0
  restoringPollingTimer.value = setInterval(async () => {
    if (count++ >= MAX_POLL_COUNT) {
      stopRestorePolling()
      restoringId.value = ''
      appStore.showWarning(t('admin.backup.operations.restoreRunning'))
      return
    }
    try {
      const record = await adminAPI.backup.getBackup(backupId)
      updateRecordInList(record)
      if (record.restore_status === 'completed' || record.restore_status === 'failed') {
        stopRestorePolling()
        restoringId.value = ''
        if (record.restore_status === 'completed') {
          appStore.showSuccess(t('admin.backup.actions.restoreSuccess'))
        } else {
          appStore.showError(record.restore_error || t('admin.backup.operations.restoreFailed'))
        }
        await loadBackups()
      }
    } catch {
      // 轮询失败时不中断
    }
  }, 2000)
}

function stopRestorePolling() {
  if (restoringPollingTimer.value) {
    clearInterval(restoringPollingTimer.value)
    restoringPollingTimer.value = null
  }
}

function handleVisibilityChange() {
  if (document.hidden) {
    stopPolling()
    stopRestorePolling()
  } else {
    // 标签页恢复时刷新列表，检查是否仍有活跃操作
    loadBackups().then(() => {
      const running = backups.value.find(r => r.status === 'running')
      if (running) {
        creatingBackup.value = true
        startPolling(running.id)
      }
      const restoring = backups.value.find(r => r.restore_status === 'running')
      if (restoring) {
        restoringId.value = restoring.id
        startRestorePolling(restoring.id)
      }
    })
  }
}

// R2 guide
const showR2Guide = ref(false)
const r2ConfigRows = computed(() => [
  { field: t('admin.backup.s3.endpoint'), value: 'https://<account_id>.r2.cloudflarestorage.com' },
  { field: t('admin.backup.s3.region'), value: 'auto' },
  { field: t('admin.backup.s3.bucket'), value: t('admin.backup.r2Guide.step4.bucketValue') },
  { field: t('admin.backup.s3.prefix'), value: 'backups/' },
  { field: 'Access Key ID', value: t('admin.backup.r2Guide.step4.fromStep2') },
  { field: 'Secret Access Key', value: t('admin.backup.r2Guide.step4.fromStep2') },
  { field: t('admin.backup.s3.forcePathStyle'), value: t('admin.backup.r2Guide.step4.unchecked') },
])

async function loadS3Config() {
  try {
    const cfg = await adminAPI.backup.getS3Config()
    s3Form.value = {
      endpoint: cfg.endpoint || '',
      region: cfg.region || 'auto',
      bucket: cfg.bucket || '',
      access_key_id: cfg.access_key_id || '',
      secret_access_key: '',
      prefix: cfg.prefix || 'backups/',
      force_path_style: cfg.force_path_style,
    }
    s3SecretConfigured.value = Boolean(cfg.access_key_id)
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  }
}

async function saveS3Config() {
  savingS3.value = true
  try {
    await backupStepUp.run(() => adminAPI.backup.updateS3Config(s3Form.value))
    appStore.showSuccess(t('admin.backup.s3.saved'))
    await loadS3Config()
  } catch (error) {
    if (isStepUpCancelled(error)) {
      savingS3.value = false
      return
    }
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  } finally {
    savingS3.value = false
  }
}

async function loadImageStorageConfig() {
  try {
    const { config, secret_configured } = await adminAPI.backup.getImageStorageConfig()
    imageStorageForm.value = {
      ...config,
      prefix: config.prefix || 'images/',
      region: config.region || 'auto',
      secret_access_key: '',
    }
    imageStorageSecretConfigured.value = secret_configured
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  }
}

async function saveImageStorageConfig() {
  savingImageStorage.value = true
  try {
    await backupStepUp.run(() => adminAPI.backup.updateImageStorageConfig(imageStorageForm.value))
    appStore.showSuccess(t('admin.backup.imageStorage.saved'))
    await loadImageStorageConfig()
  } catch (error) {
    if (isStepUpCancelled(error)) {
      savingImageStorage.value = false
      return
    }
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  } finally {
    savingImageStorage.value = false
  }
}

async function testImageStorage() {
  testingImageStorage.value = true
  try {
    const result = await adminAPI.backup.testImageStorageConnection(imageStorageForm.value)
    if (result.ok) {
      appStore.showSuccess(result.message || t('admin.backup.s3.testSuccess'))
    } else {
      appStore.showError(result.message || t('admin.backup.s3.testFailed'))
    }
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  } finally {
    testingImageStorage.value = false
  }
}

async function testS3() {
  testingS3.value = true
  try {
    const result = await adminAPI.backup.testS3Connection(s3Form.value)
    if (result.ok) {
      appStore.showSuccess(result.message || t('admin.backup.s3.testSuccess'))
    } else {
      appStore.showError(result.message || t('admin.backup.s3.testFailed'))
    }
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  } finally {
    testingS3.value = false
  }
}

async function loadSchedule() {
  try {
    const cfg = await adminAPI.backup.getSchedule()
    scheduleForm.value = {
      enabled: cfg.enabled,
      cron_expr: cfg.cron_expr || '0 2 * * *',
      retain_days: cfg.retain_days || 14,
      retain_count: cfg.retain_count || 10,
    }
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  }
}

async function saveSchedule() {
  savingSchedule.value = true
  try {
    await adminAPI.backup.updateSchedule(scheduleForm.value)
    appStore.showSuccess(t('admin.backup.schedule.saved'))
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  } finally {
    savingSchedule.value = false
  }
}

async function loadBackups() {
  loadingBackups.value = true
  try {
    const result = await adminAPI.backup.listBackups()
    backups.value = result.items || []
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  } finally {
    loadingBackups.value = false
  }
}

async function createBackup() {
  creatingBackup.value = true
  try {
    const record = await backupStepUp.run(() => adminAPI.backup.createBackup({ expire_days: manualExpireDays.value }))
    // 插入到列表顶部
    backups.value.unshift(record)
    startPolling(record.id)
  } catch (error: any) {
    if (isStepUpCancelled(error)) {
      creatingBackup.value = false
      return
    }
    if (reportStepUpBlocked(error)) {
      creatingBackup.value = false
      return
    }
    if (error?.response?.status === 409) {
      appStore.showWarning(t('admin.backup.operations.alreadyInProgress'))
    } else {
      appStore.showError(error?.message || t('errors.networkError'))
    }
    creatingBackup.value = false
  }
}

async function downloadBackup(id: string) {
  try {
    const result = await backupStepUp.run(() => adminAPI.backup.getDownloadURL(id))
    if (result.parts && result.parts.length > 0) {
      downloadParts.value = result.parts
      downloadPartsModalOpen.value = true
      return
    }
    if (!result.url) {
      throw new Error(t('admin.backup.actions.downloadFailed'))
    }
    // 预签名 URL 带 attachment disposition，同页 anchor 导航直接触发下载；
    // 不用 window.open：step-up 弹窗 await 会耗尽瞬态用户激活，新标签页会被浏览器拦截。
    const link = document.createElement('a')
    link.href = result.url
    link.rel = 'noopener'
    link.click()
  } catch (error) {
    if (isStepUpCancelled(error)) return
    if (reportStepUpBlocked(error)) return
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  }
}

function closeDownloadParts() {
  downloadPartsModalOpen.value = false
  downloadParts.value = []
}

async function restoreBackup(id: string) {
  if (!window.confirm(t('admin.backup.actions.restoreConfirm'))) return
  const password = window.prompt(t('admin.backup.actions.restorePasswordPrompt'))
  if (!password) return
  restoringId.value = id
  try {
    const record = await backupStepUp.run(() => adminAPI.backup.restoreBackup(id, password))
    updateRecordInList(record)
    startRestorePolling(id)
  } catch (error: any) {
    restoringId.value = ''
    if (isStepUpCancelled(error)) return
    if (reportStepUpBlocked(error)) return
    // apiClient 拦截器把 HTTP 错误归一化为顶层 { status } 平面对象（无 response 字段）
    if (error?.status === 409 || error?.response?.status === 409) {
      appStore.showWarning(t('admin.backup.operations.restoreRunning'))
    } else {
      appStore.showError(error?.message || t('errors.networkError'))
    }
  }
}

async function removeBackup(id: string) {
  if (!window.confirm(t('admin.backup.actions.deleteConfirm'))) return
  try {
    await adminAPI.backup.deleteBackup(id)
    appStore.showSuccess(t('admin.backup.actions.deleted'))
    await loadBackups()
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || t('errors.networkError'))
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'completed':
      return 'views-admin-backup-view__state'
    case 'running':
      return 'views-admin-backup-view__state-2'
    case 'failed':
      return 'views-admin-backup-view__state-3'
    default:
      return 'views-admin-backup-view__state-4'
  }
}

function formatSize(bytes: number): string {
  if (!bytes || bytes <= 0) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

onMounted(async () => {
  document.addEventListener('visibilitychange', handleVisibilityChange)
  await Promise.all([loadS3Config(), loadImageStorageConfig(), loadSchedule(), loadBackups()])

  // 如果有正在 running 的备份，恢复轮询
  const runningBackup = backups.value.find(r => r.status === 'running')
  if (runningBackup) {
    creatingBackup.value = true
    startPolling(runningBackup.id)
  }
  const restoringBackup = backups.value.find(r => r.restore_status === 'running')
  if (restoringBackup) {
    restoringId.value = restoringBackup.id
    startRestorePolling(restoringBackup.id)
  }
})

onBeforeUnmount(() => {
  stopPolling()
  stopRestorePolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.settings-backup-content {
  padding-bottom: calc(5.5rem + env(safe-area-inset-bottom));
}

@media (max-width: 640px) {
  .settings-backup-content {
    padding-bottom: calc(5rem + env(safe-area-inset-bottom));
  }
}

.r2-guide-dialog {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.r2-guide-dialog__intro {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.5;
}

.r2-guide-step {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.r2-guide-step__header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.r2-guide-step__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--theme-accent) 16%, transparent);
  color: var(--theme-accent);
  font-size: var(--font-size-xs);
  font-weight: 700;
  flex-shrink: 0;
}

.r2-guide-step__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--font-size-sm);
  font-weight: 600;
}

.r2-guide-step__list {
  margin: 0;
  padding-left: 2.125rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  line-height: 1.6;
}

.r2-guide-step__list li + li {
  margin-top: 0.25rem;
}

.r2-guide-step__desc {
  margin: 0;
  padding-left: 2.125rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
}

.r2-guide-code-box {
  margin-left: 2.125rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  background: var(--color-surface-muted);
  overflow-x: auto;
}

.r2-guide-code-box code {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-primary);
  word-break: break-all;
}

.r2-guide-alert {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-left: 2.125rem;
  padding: 0.625rem 0.85rem;
  border-radius: var(--radius-md);
  font-size: var(--font-size-xs);
  line-height: 1.5;
}

.r2-guide-alert--warning {
  border: 1px solid color-mix(in srgb, #f59e0b 35%, transparent);
  background: color-mix(in srgb, #f59e0b 10%, transparent);
  color: #d97706;
}

:is(.dark) .r2-guide-alert--warning {
  color: #fbbf24;
}

.r2-guide-alert--info {
  margin-left: 0;
  border: 1px solid color-mix(in srgb, var(--theme-accent) 30%, transparent);
  background: color-mix(in srgb, var(--theme-accent) 10%, transparent);
  color: var(--color-text-secondary);
}

.r2-guide-alert :deep(.app-icon) {
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.r2-guide-table-wrap {
  margin-left: 2.125rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.r2-guide-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-xs);
}

.r2-guide-table td {
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.r2-guide-table tr:last-child td {
  border-bottom: none;
}

.r2-guide-table__key {
  color: var(--color-text-secondary);
  font-weight: 500;
  width: 38%;
  background: color-mix(in srgb, var(--color-surface-muted) 50%, transparent);
}

.r2-guide-table__val {
  color: var(--color-text-primary);
}

.r2-guide-table__val code {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  background: var(--color-surface-muted);
  padding: 0.15rem 0.35rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border-subtle);
}

.parts-dialog {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.parts-dialog__hint {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.parts-dialog__list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.parts-dialog__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-surface-muted);
}

.parts-dialog__meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.parts-dialog__name {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-primary);
}

.parts-dialog__size {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}
</style>
