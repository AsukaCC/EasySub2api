<template>
  <AppLayout>
    <main class="system-updates">
      <header class="system-updates__header">
        <div>
          <h1>{{ t('admin.settings.systemUpdates.title') }}</h1>
          <p>{{ t('admin.settings.systemUpdates.description') }}</p>
        </div>
        <button class="btn btn-secondary" type="button" :disabled="loading" @click="checkForUpdates(true)">
          <Icon name="refresh" size="sm" :class="{ 'system-updates__spin': loading }" />
          {{ loading ? t('admin.settings.systemUpdates.checking') : t('admin.settings.systemUpdates.check') }}
        </button>
      </header>

      <section v-if="loading && !versionInfo" class="system-updates__loading" aria-live="polite">
        <Icon name="refresh" size="lg" class="system-updates__spin" />
        <span>{{ t('common.loading') }}</span>
      </section>

      <template v-else-if="versionInfo">
        <section class="system-updates__summary" :class="{ 'is-update': versionInfo.has_update }">
          <div class="system-updates__summary-icon">
            <Icon :name="versionInfo.has_update ? 'download' : 'check'" size="lg" />
          </div>
          <div class="system-updates__summary-copy">
            <div class="system-updates__status-line">
              <h2>{{ versionInfo.has_update ? t('version.updateAvailable') : t('version.upToDate') }}</h2>
              <span class="system-updates__build-tag">
                {{ versionInfo.build_type === 'release' ? t('admin.settings.systemUpdates.releaseBuild') : t('admin.settings.systemUpdates.sourceBuild') }}
              </span>
            </div>
            <div class="system-updates__versions">
              <span>{{ t('version.currentVersion') }} <strong>v{{ versionInfo.current_version || '--' }}</strong></span>
              <span>{{ t('version.latestVersion') }} <strong>v{{ versionInfo.latest_version || versionInfo.current_version || '--' }}</strong></span>
            </div>
            <p v-if="versionInfo.build_type !== 'release'" class="system-updates__hint">
              {{ t('version.sourceModeHint') }}
            </p>
            <p v-if="versionInfo.warning" class="system-updates__hint is-warning">{{ versionInfo.warning }}</p>
          </div>
          <button
            v-if="versionInfo.has_update && versionInfo.build_type === 'release'"
            class="btn btn-primary system-updates__update-button"
            type="button"
            :disabled="updating"
            @click="performUpdateNow"
          >
            <Icon v-if="updating" name="refresh" size="sm" class="system-updates__spin" />
            <Icon v-else name="download" size="sm" />
            {{ updating ? t('version.updating') : t('version.updateNow') }}
          </button>
        </section>

        <section v-if="versionInfo.release_info" class="system-updates__release">
          <div class="system-updates__section-heading">
            <div>
              <h2>{{ t('version.releaseNotes') }}</h2>
              <p>{{ versionInfo.release_info.name || ('v' + versionInfo.latest_version) }}</p>
            </div>
            <a
              v-if="versionInfo.release_info.html_url && versionInfo.release_info.html_url !== '#'"
              class="btn btn-secondary btn-sm"
              :href="versionInfo.release_info.html_url"
              target="_blank"
              rel="noopener noreferrer"
            >
              <Icon name="externalLink" size="sm" />
              {{ t('version.viewChangelog') }}
            </a>
          </div>
          <p v-if="versionInfo.release_info.published_at" class="system-updates__published">
            {{ formatPublishedAt(versionInfo.release_info.published_at) }}
          </p>
          <pre v-if="versionInfo.release_info.body" class="system-updates__release-body">{{ versionInfo.release_info.body }}</pre>
          <p v-else class="system-updates__empty">{{ t('version.noReleaseNotes') }}</p>
        </section>

        <section class="system-updates__rollback">
          <div class="system-updates__section-heading">
            <div>
              <h2>{{ t('version.rollback') }}</h2>
              <p>{{ t('admin.settings.systemUpdates.rollbackDescription') }}</p>
            </div>
            <button
              v-if="versionInfo.build_type === 'release'"
              class="btn btn-secondary btn-sm"
              type="button"
              :disabled="rollbackLoading"
              @click="loadRollbackVersions"
            >
              <Icon name="clock" size="sm" />
              {{ rollbackLoading ? t('common.loading') : t('admin.settings.systemUpdates.loadRollback') }}
            </button>
          </div>

          <p v-if="versionInfo.build_type !== 'release'" class="system-updates__hint">
            {{ t('version.rollbackSourceHint') }}
          </p>
          <template v-else>
            <p v-if="rollbackError" class="system-updates__error">{{ rollbackError }}</p>
            <p v-else-if="rollbackLoaded && rollbackVersions.length === 0" class="system-updates__empty">
              {{ t('version.noRollbackVersions') }}
            </p>
            <div v-else-if="rollbackLoaded" class="system-updates__versions-list">
              <button
                v-for="item in rollbackVersions"
                :key="item.version"
                class="system-updates__version-option"
                :class="{ 'is-selected': selectedRollbackVersion === item.version }"
                type="button"
                :disabled="rollingBack"
                @click="selectedRollbackVersion = selectedRollbackVersion === item.version ? '' : item.version"
              >
                <span>v{{ item.version }}</span>
                <small>{{ formatPublishedAt(item.published_at) }}</small>
              </button>
              <button
                v-if="selectedRollbackVersion"
                class="btn btn-danger system-updates__rollback-button"
                type="button"
                :disabled="rollingBack"
                @click="rollbackNow"
              >
                <Icon v-if="rollingBack" name="refresh" size="sm" class="system-updates__spin" />
                <Icon v-else name="clock" size="sm" />
                {{ rollingBack ? t('version.rollingBack') : t('version.rollbackConfirm', { version: 'v' + selectedRollbackVersion }) }}
              </button>
            </div>
            <p v-else class="system-updates__empty">{{ t('admin.settings.systemUpdates.rollbackPrompt') }}</p>
          </template>
        </section>

        <section v-if="updateSuccess && needRestart" class="system-updates__restart">
          <div>
            <h2>{{ t('version.updateComplete') }}</h2>
            <p>{{ t('version.restartRequired') }}</p>
          </div>
          <button class="btn btn-primary" type="button" :disabled="restarting" @click="restartNow">
            <Icon v-if="restarting" name="refresh" size="sm" class="system-updates__spin" />
            <Icon v-else name="refresh" size="sm" />
            {{ restarting ? t('version.restarting') : t('version.restartNow') }}
          </button>
        </section>

        <p v-if="pageError" class="system-updates__error" role="alert">{{ pageError }}</p>
      </template>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api'
import type { RollbackVersionInfo, VersionInfo } from '@/api/admin/system'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const versionInfo = ref<VersionInfo | null>(null)
const pageError = ref('')
const updating = ref(false)
const updateSuccess = ref(false)
const needRestart = ref(false)
const restarting = ref(false)
const rollbackLoading = ref(false)
const rollbackLoaded = ref(false)
const rollbackVersions = ref<RollbackVersionInfo[]>([])
const rollbackError = ref('')
const selectedRollbackVersion = ref('')
const rollingBack = ref(false)

async function checkForUpdates(force = true) {
  loading.value = true
  pageError.value = ''
  try {
    const data = await adminAPI.system.checkUpdates(force)
    versionInfo.value = data
    await appStore.fetchVersion(force)
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    pageError.value = err.response?.data?.message || err.message || t('admin.settings.systemUpdates.checkFailed')
  } finally {
    loading.value = false
  }
}

async function performUpdateNow() {
  if (updating.value) return
  updating.value = true
  pageError.value = ''
  updateSuccess.value = false
  try {
    const result = await adminAPI.system.performUpdate()
    updateSuccess.value = true
    needRestart.value = result.need_restart
    appStore.clearVersionCache()
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    pageError.value = err.response?.data?.message || err.message || t('version.updateFailed')
  } finally {
    updating.value = false
  }
}

async function loadRollbackVersions() {
  if (rollbackLoading.value) return
  rollbackLoading.value = true
  rollbackError.value = ''
  try {
    const data = await adminAPI.system.getRollbackVersions()
    rollbackVersions.value = data.versions || []
    rollbackLoaded.value = true
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    rollbackError.value = err.response?.data?.message || err.message || t('version.loadVersionsFailed')
  } finally {
    rollbackLoading.value = false
  }
}

async function rollbackNow() {
  if (!selectedRollbackVersion.value || rollingBack.value) return
  rollingBack.value = true
  pageError.value = ''
  try {
    const result = await adminAPI.system.rollback(selectedRollbackVersion.value)
    updateSuccess.value = true
    needRestart.value = result.need_restart
    selectedRollbackVersion.value = ''
    appStore.clearVersionCache()
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    pageError.value = err.response?.data?.message || err.message || t('version.rollbackFailed')
  } finally {
    rollingBack.value = false
  }
}

async function restartNow() {
  if (restarting.value) return
  restarting.value = true
  try {
    await adminAPI.system.restartService()
  } catch {
    // The connection is expected to close while the service restarts.
  }
  window.setTimeout(() => window.location.reload(), 8000)
}

function formatPublishedAt(value: string): string {
  if (!value) return ''
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleDateString()
}

onMounted(() => checkForUpdates(false))
</script>

<style scoped>
.system-updates { width: min(1080px, 100%); margin: 0 auto; padding: 2rem; }
.system-updates__header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; margin-bottom: 1.5rem; }
.system-updates__header h1 { margin: 0; font-size: var(--font-size-2xl); }
.system-updates__header p { margin: .45rem 0 0; color: var(--color-text-secondary); }
.system-updates__header .btn, .system-updates__section-heading .btn, .system-updates__restart .btn { display: inline-flex; align-items: center; gap: .45rem; }
.system-updates__summary, .system-updates__release, .system-updates__rollback, .system-updates__restart { border: 1px solid var(--glass-border); border-radius: 8px; background: var(--glass-bg); padding: 1.25rem; margin-bottom: 1rem; }
.system-updates__summary { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; gap: 1rem; align-items: center; }
.system-updates__summary.is-update { border-color: color-mix(in srgb, var(--color-primary) 45%, var(--glass-border)); }
.system-updates__summary-icon { display: grid; place-items: center; width: 3rem; height: 3rem; border-radius: 8px; color: var(--color-success); background: color-mix(in srgb, var(--color-success) 12%, transparent); }
.system-updates__summary.is-update .system-updates__summary-icon { color: var(--color-primary); background: color-mix(in srgb, var(--color-primary) 12%, transparent); }
.system-updates__status-line { display: flex; flex-wrap: wrap; align-items: center; gap: .6rem; }
.system-updates__status-line h2, .system-updates__section-heading h2, .system-updates__restart h2 { margin: 0; font-size: var(--font-size-base); }
.system-updates__build-tag { border-radius: 4px; padding: .2rem .45rem; color: var(--color-text-secondary); background: var(--color-surface-muted); font-size: var(--font-size-xs); }
.system-updates__versions { display: flex; flex-wrap: wrap; gap: 1.25rem; margin-top: .55rem; color: var(--color-text-secondary); font-size: var(--font-size-sm); }
.system-updates__versions strong { color: var(--color-text-primary); font-weight: 600; }
.system-updates__hint, .system-updates__section-heading p, .system-updates__restart p { margin: .55rem 0 0; color: var(--color-text-secondary); font-size: var(--font-size-sm); line-height: 1.5; }
.system-updates__hint.is-warning, .system-updates__error { color: var(--color-danger); }
.system-updates__section-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; margin-bottom: 1rem; }
.system-updates__published { margin: .25rem 0 .75rem; color: var(--color-text-secondary); font-size: var(--font-size-xs); }
.system-updates__release-body { max-height: 360px; overflow: auto; margin: 0; padding: .9rem; border-radius: 6px; background: var(--color-surface-muted); color: var(--color-text-secondary); font: inherit; font-size: var(--font-size-sm); line-height: 1.6; white-space: pre-wrap; }
.system-updates__empty { margin: 0; color: var(--color-text-secondary); font-size: var(--font-size-sm); }
.system-updates__versions-list { display: flex; flex-wrap: wrap; align-items: center; gap: .6rem; }
.system-updates__version-option { display: inline-flex; flex-direction: column; align-items: flex-start; gap: .2rem; min-width: 130px; padding: .65rem .75rem; border: 1px solid var(--glass-border); border-radius: 6px; color: var(--color-text-primary); background: transparent; cursor: pointer; text-align: left; }
.system-updates__version-option:hover, .system-updates__version-option.is-selected { border-color: var(--color-primary); background: color-mix(in srgb, var(--color-primary) 10%, transparent); }
.system-updates__version-option small { color: var(--color-text-secondary); }
.system-updates__rollback-button { display: inline-flex; align-items: center; gap: .45rem; }
.system-updates__restart { display: flex; align-items: center; justify-content: space-between; gap: 1rem; border-color: color-mix(in srgb, var(--color-success) 45%, var(--glass-border)); }
.system-updates__loading { min-height: 18rem; display: grid; place-content: center; justify-items: center; gap: .75rem; color: var(--color-text-secondary); }
.system-updates__spin { animation: system-updates-spin 1s linear infinite; }
@keyframes system-updates-spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) { .system-updates { padding: 1rem; } .system-updates__header, .system-updates__section-heading, .system-updates__restart { flex-direction: column; align-items: stretch; } .system-updates__header .btn, .system-updates__section-heading .btn, .system-updates__restart .btn { justify-content: center; } .system-updates__summary { grid-template-columns: auto minmax(0, 1fr); } .system-updates__update-button { grid-column: 1 / -1; justify-content: center; } }
</style>
