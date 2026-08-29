<template>
  <div class="components-user-profile-totp-disable-dialog__panel" @click.self="$emit('close')">
    <div class="components-user-profile-totp-disable-dialog__panel-2">
      <div class="components-user-profile-totp-disable-dialog__panel-3" @click="$emit('close')"></div>

      <div class="components-user-profile-totp-disable-dialog__panel-4">
        <!-- Header -->
        <div class="components-user-profile-totp-disable-dialog__panel-5">
          <div class="components-user-profile-totp-disable-dialog__panel-6">
            <svg class="components-user-profile-totp-disable-dialog__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
            </svg>
          </div>
          <h3 class="components-user-profile-totp-disable-dialog__heading">
            {{ t('profile.totp.disableTitle') }}
          </h3>
          <p class="components-user-profile-totp-disable-dialog__description">
            {{ t('profile.totp.disableWarning') }}
          </p>
        </div>

        <!-- Loading verification method -->
        <div v-if="methodLoading" class="components-user-profile-totp-disable-dialog__panel-7">
          <div class="components-user-profile-totp-disable-dialog__panel-8"></div>
        </div>

        <form v-else @submit.prevent="handleDisable" class="components-user-profile-totp-disable-dialog__form">
          <!-- Email verification -->
          <div v-if="verificationMethod === 'email'">
            <label class="input-label">{{ t('profile.totp.emailCode') }}</label>
            <div class="components-user-profile-totp-disable-dialog__panel-9">
              <input
                v-model="form.emailCode"
                type="text"
                maxlength="6"
                inputmode="numeric"
                class="components-user-profile-totp-disable-dialog__field input"
                :placeholder="t('profile.totp.enterEmailCode')"
              />
              <button
                type="button"
                class="components-user-profile-totp-disable-dialog__action btn btn-secondary"
                :disabled="sendingCode || codeCooldown > 0"
                @click="handleSendCode"
              >
                {{ codeCooldown > 0 ? `${codeCooldown}s` : (sendingCode ? t('common.sending') : t('profile.totp.sendCode')) }}
              </button>
            </div>
          </div>

          <!-- Password verification -->
          <div v-else>
            <label for="password" class="input-label">
              {{ t('profile.currentPassword') }}
            </label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              class="input"
              :placeholder="t('profile.totp.enterPassword')"
            />
          </div>

          <!-- Actions -->
          <div class="components-user-profile-totp-disable-dialog__panel-10">
            <button type="button" class="btn btn-secondary" @click="$emit('close')">
              {{ t('common.cancel') }}
            </button>
            <button
              type="submit"
              class="btn btn-danger"
              :disabled="loading || !canSubmit"
            >
              {{ loading ? t('common.processing') : t('profile.totp.confirmDisable') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { totpAPI } from '@/api'

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const methodLoading = ref(true)
const verificationMethod = ref<'email' | 'password'>('password')
const loading = ref(false)
const sendingCode = ref(false)
const codeCooldown = ref(0)
const cooldownTimer = ref<ReturnType<typeof setInterval> | null>(null)
const form = ref({
  emailCode: '',
  password: ''
})

const canSubmit = computed(() => {
  if (verificationMethod.value === 'email') {
    return form.value.emailCode.length === 6
  }
  return form.value.password.length > 0
})

const loadVerificationMethod = async () => {
  methodLoading.value = true
  try {
    const method = await totpAPI.getVerificationMethod()
    verificationMethod.value = method.method
  } catch (err: any) {
    appStore.showError(err.response?.data?.message || t('common.error'))
    emit('close')
  } finally {
    methodLoading.value = false
  }
}

const handleSendCode = async () => {
  sendingCode.value = true
  try {
    await totpAPI.sendVerifyCode()
    appStore.showSuccess(t('profile.totp.codeSent'))
    // Start cooldown
    codeCooldown.value = 60
    if (cooldownTimer.value) {
      clearInterval(cooldownTimer.value)
      cooldownTimer.value = null
    }
    cooldownTimer.value = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) {
        if (cooldownTimer.value) {
          clearInterval(cooldownTimer.value)
          cooldownTimer.value = null
        }
      }
    }, 1000)
  } catch (err: any) {
    appStore.showError(err.response?.data?.message || t('profile.totp.sendCodeFailed'))
  } finally {
    sendingCode.value = false
  }
}

const handleDisable = async () => {
  if (!canSubmit.value) return

  loading.value = true

  try {
    const request = verificationMethod.value === 'email'
      ? { email_code: form.value.emailCode }
      : { password: form.value.password }

    await totpAPI.disable(request)
    appStore.showSuccess(t('profile.totp.disableSuccess'))
    emit('success')
  } catch (err: any) {
    appStore.showError(err.response?.data?.message || t('profile.totp.disableFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadVerificationMethod()
})

onUnmounted(() => {
  if (cooldownTimer.value) {
    clearInterval(cooldownTimer.value)
    cooldownTimer.value = null
  }
})
</script>
