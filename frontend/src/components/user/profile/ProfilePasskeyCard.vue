<template>
  <div class="card">
    <div class="components-user-profile-profile-passkey-card__panel">
      <div>
        <h2 class="components-user-profile-profile-passkey-card__heading">
          {{ t('profile.passkey.title') }}
        </h2>
        <p class="components-user-profile-profile-passkey-card__description">
          {{ t('profile.passkey.description') }}
        </p>
      </div>
      <button
        v-if="enabled && supported && !showAddForm"
        type="button"
        class="btn btn-primary"
        :disabled="busy"
        @click="showAddForm = true"
      >
        {{ t('profile.passkey.add') }}
      </button>
    </div>

    <div class="components-user-profile-profile-passkey-card__panel-2">
      <div v-if="!enabled" class="components-user-profile-profile-passkey-card__panel-3">
        {{ t('profile.passkey.featureDisabled') }}
      </div>
      <div v-if="enabled && !supported" class="components-user-profile-profile-passkey-card__panel-4">
        {{ t('profile.passkey.unsupported') }}
      </div>
      <div>
        <form
          v-if="enabled && supported && showAddForm"
          class="components-user-profile-profile-passkey-card__form"
          @submit.prevent="addPasskey"
        >
          <div class="components-user-profile-profile-passkey-card__panel-5">
            <div>
              <label for="passkey-name" class="input-label">{{ t('profile.passkey.name') }}</label>
              <input
                id="passkey-name"
                v-model="newName"
                class="input"
                maxlength="100"
                :placeholder="t('profile.passkey.namePlaceholder')"
                autofocus
              />
            </div>
            <div>
              <label for="passkey-add-password" class="input-label">{{
                t('profile.currentPassword')
              }}</label>
              <input
                id="passkey-add-password"
                v-model="newPassword"
                type="password"
                autocomplete="current-password"
                class="input"
                :placeholder="t('profile.passkey.passwordPlaceholder')"
              />
            </div>
          </div>
          <div class="components-user-profile-profile-passkey-card__panel-6">
            <button type="button" class="btn btn-secondary" :disabled="busy" @click="cancelAdd">
              {{ t('common.cancel') }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="busy || newPassword.length === 0">
              {{ busy ? t('common.processing') : t('profile.passkey.continue') }}
            </button>
          </div>
        </form>

        <div v-if="loading" class="components-user-profile-profile-passkey-card__panel-7">
          <div class="components-user-profile-profile-passkey-card__panel-8"></div>
        </div>

        <div
          v-else-if="credentials.length === 0"
          class="components-user-profile-profile-passkey-card__panel-9"
        >
          {{ t('profile.passkey.empty') }}
        </div>

        <div v-else class="components-user-profile-profile-passkey-card__panel-10">
          <div
            v-for="credential in credentials"
            :key="credential.id"
            class="components-user-profile-profile-passkey-card__panel-11"
          >
            <div class="components-user-profile-profile-passkey-card__panel-12">
              <div class="components-user-profile-profile-passkey-card__panel-13">
                <Icon name="key" size="md" class="components-user-profile-profile-passkey-card__icon" />
                <p class="components-user-profile-profile-passkey-card__description-2">
                  {{ credential.name }}
                </p>
                <span
                  v-if="credential.backup"
                  class="components-user-profile-profile-passkey-card__text"
                >
                  {{ t('profile.passkey.synced') }}
                </span>
              </div>
              <p class="components-user-profile-profile-passkey-card__description-3">
                {{ t('profile.passkey.createdAt', { date: formatDate(credential.created_at) }) }}
                <template v-if="credential.last_used_at">
                  · {{ t('profile.passkey.lastUsed', { date: formatDate(credential.last_used_at) }) }}
                </template>
              </p>
            </div>
            <div class="components-user-profile-profile-passkey-card__panel-14">
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="busy"
                @click="renamePasskey(credential)"
              >
                {{ t('common.edit') }}
              </button>
              <button
                type="button"
                class="components-user-profile-profile-passkey-card__action btn btn-ghost btn-sm"
                :disabled="busy"
                @click="deletePasskey(credential)"
              >
                {{ t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 删除确认：吊销凭据需验证当前密码，防止被窃会话静默移除 Passkey -->
    <div v-if="deleteTarget" class="components-user-profile-profile-passkey-card__panel-15">
      <div class="components-user-profile-profile-passkey-card__panel-16">
        <div class="components-user-profile-profile-passkey-card__panel-17" @click="closeDeleteDialog"></div>
        <div
          class="components-user-profile-profile-passkey-card__panel-18"
        >
          <h3 class="components-user-profile-profile-passkey-card__heading-2">
            {{ t('profile.passkey.deleteTitle') }}
          </h3>
          <p class="components-user-profile-profile-passkey-card__description-4">
            {{ t('profile.passkey.deleteConfirm', { name: deleteTarget.name }) }}
          </p>
          <form class="components-user-profile-profile-passkey-card__form-2" @submit.prevent="confirmDelete">
            <div>
              <label for="passkey-delete-password" class="input-label">{{
                t('profile.currentPassword')
              }}</label>
              <input
                id="passkey-delete-password"
                v-model="deletePassword"
                type="password"
                autocomplete="current-password"
                class="input"
                :placeholder="t('profile.passkey.passwordPlaceholder')"
                autofocus
              />
            </div>
            <div class="components-user-profile-profile-passkey-card__panel-19">
              <button type="button" class="btn btn-secondary" :disabled="busy" @click="closeDeleteDialog">
                {{ t('common.cancel') }}
              </button>
              <button
                type="submit"
                class="btn btn-danger"
                :disabled="busy || deletePassword.length === 0"
              >
                {{ busy ? t('common.processing') : t('common.delete') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { passkeyAPI, type PasskeyCredentialSummary } from '@/api'
import { Icon } from '@/components/icons'
import { useAppStore } from '@/stores/app'

const props = defineProps<{ enabled: boolean }>()

const { t } = useI18n()
const appStore = useAppStore()
const supported = passkeyAPI.isSupported()
const loading = ref(false)
const busy = ref(false)
const showAddForm = ref(false)
const newName = ref('')
const newPassword = ref('')
const deleteTarget = ref<PasskeyCredentialSummary | null>(null)
const deletePassword = ref('')
const credentials = ref<PasskeyCredentialSummary[]>([])

// apiClient 拦截器把错误规范化为 { code, reason, message }；
// 透出后端消息（如密码错误），否则回退到通用文案。
function extractErrorMessage(error: unknown, fallback: string): string {
  const message = (error as { message?: string }).message
  return typeof message === 'string' && message.length > 0 ? message : fallback
}

async function loadCredentials(): Promise<void> {
  if (!props.enabled) {
    credentials.value = []
    return
  }
  loading.value = true
  try {
    credentials.value = await passkeyAPI.list()
  } catch (error) {
    // 字符串错误码在 reason 字段（code 是数字状态码）；
    // 设置变更竞态下后端仍可能返回 PASSKEY_DISABLED，静默处理
    const reason = (error as { reason?: string }).reason
    if (reason !== 'PASSKEY_DISABLED') {
      appStore.showError(t('profile.passkey.loadFailed'))
    }
  } finally {
    loading.value = false
  }
}

async function addPasskey(): Promise<void> {
  if (newPassword.value.length === 0) return
  busy.value = true
  try {
    await passkeyAPI.register(newName.value.trim(), newPassword.value)
    appStore.showSuccess(t('profile.passkey.added'))
    cancelAdd()
    await loadCredentials()
  } catch (error) {
    if (!(error instanceof DOMException && error.name === 'NotAllowedError')) {
      appStore.showError(extractErrorMessage(error, t('profile.passkey.addFailed')))
    }
  } finally {
    busy.value = false
  }
}

function cancelAdd(): void {
  showAddForm.value = false
  newName.value = ''
  newPassword.value = ''
}

async function renamePasskey(credential: PasskeyCredentialSummary): Promise<void> {
  const name = window.prompt(t('profile.passkey.renamePrompt'), credential.name)?.trim()
  if (!name || name === credential.name) return
  busy.value = true
  try {
    await passkeyAPI.rename(credential.id, name)
    credential.name = name
    appStore.showSuccess(t('profile.passkey.renamed'))
  } catch {
    appStore.showError(t('profile.passkey.renameFailed'))
  } finally {
    busy.value = false
  }
}

function deletePasskey(credential: PasskeyCredentialSummary): void {
  deleteTarget.value = credential
  deletePassword.value = ''
}

function closeDeleteDialog(): void {
  deleteTarget.value = null
  deletePassword.value = ''
}

async function confirmDelete(): Promise<void> {
  const credential = deleteTarget.value
  if (!credential || deletePassword.value.length === 0) return
  busy.value = true
  try {
    await passkeyAPI.remove(credential.id, deletePassword.value)
    credentials.value = credentials.value.filter((item) => item.id !== credential.id)
    appStore.showSuccess(t('profile.passkey.deleted'))
    closeDeleteDialog()
  } catch (error) {
    // 密码错误等失败保持对话框打开，允许重试
    appStore.showError(extractErrorMessage(error, t('profile.passkey.deleteFailed')))
  } finally {
    busy.value = false
  }
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  }).format(new Date(value))
}

watch(
  () => props.enabled,
  () => {
    void loadCredentials()
  },
  { immediate: true }
)
</script>
