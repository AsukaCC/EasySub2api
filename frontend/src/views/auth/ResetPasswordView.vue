<template>
  <AuthLayout>
    <div class="views-auth-reset-password-view__panel">
      <!-- Title -->
      <div class="views-auth-reset-password-view__panel-2">
        <h2 class="views-auth-reset-password-view__heading">
          {{ t('auth.resetPasswordTitle') }}
        </h2>
        <p class="views-auth-reset-password-view__description">
          {{ t('auth.resetPasswordHint') }}
        </p>
      </div>

      <!-- Invalid Link State -->
      <div v-if="isInvalidLink" class="views-auth-reset-password-view__panel">
        <div class="views-auth-reset-password-view__panel-3">
          <div class="views-auth-reset-password-view__panel-4">
            <div class="views-auth-reset-password-view__panel-5">
              <Icon name="exclamationCircle" size="lg" class="views-auth-reset-password-view__icon" />
            </div>
            <div>
              <h3 class="views-auth-reset-password-view__heading-2">
                {{ t('auth.invalidResetLink') }}
              </h3>
              <p class="views-auth-reset-password-view__description-2">
                {{ t('auth.invalidResetLinkHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="views-auth-reset-password-view__panel-2">
          <router-link
            to="/forgot-password"
            class="views-auth-reset-password-view__router-link"
          >
            {{ t('auth.requestNewResetLink') }}
          </router-link>
        </div>
      </div>

      <!-- Success State -->
      <div v-else-if="isSuccess" class="views-auth-reset-password-view__panel">
        <div class="views-auth-reset-password-view__panel-6">
          <div class="views-auth-reset-password-view__panel-4">
            <div class="views-auth-reset-password-view__panel-7">
              <Icon name="checkCircle" size="lg" class="views-auth-reset-password-view__icon-2" />
            </div>
            <div>
              <h3 class="views-auth-reset-password-view__heading-3">
                {{ t('auth.passwordResetSuccess') }}
              </h3>
              <p class="views-auth-reset-password-view__description-3">
                {{ t('auth.passwordResetSuccessHint') }}
              </p>
            </div>
          </div>
        </div>

        <div class="views-auth-reset-password-view__panel-2">
          <router-link
            to="/login"
            class="views-auth-reset-password-view__router-link-2 btn btn-primary"
          >
            <Icon name="login" size="md" />
            {{ t('auth.signIn') }}
          </router-link>
        </div>
      </div>

      <!-- Form State -->
      <form v-else @submit.prevent="handleSubmit" class="views-auth-reset-password-view__form">
        <!-- Email (readonly) -->
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="views-auth-reset-password-view__panel-8">
            <div class="views-auth-reset-password-view__panel-9">
              <Icon name="mail" size="md" class="views-auth-reset-password-view__icon-3" />
            </div>
            <input
              id="email"
              :value="email"
              type="email"
              readonly
              disabled
              class="views-auth-reset-password-view__field input"
            />
          </div>
        </div>

        <!-- New Password Input -->
        <div>
          <label for="password" class="input-label">
            {{ t('auth.newPassword') }}
          </label>
          <div class="views-auth-reset-password-view__panel-8">
            <div class="views-auth-reset-password-view__panel-9">
              <Icon name="lock" size="md" class="views-auth-reset-password-view__icon-3" />
            </div>
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="isLoading"
              class="views-auth-reset-password-view__field-2 input"
              :class="{ 'input-error': errors.password }"
              :placeholder="t('auth.newPasswordPlaceholder')"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="views-auth-reset-password-view__action"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
        </div>

        <!-- Confirm Password Input -->
        <div>
          <label for="confirmPassword" class="input-label">
            {{ t('auth.confirmPassword') }}
          </label>
          <div class="views-auth-reset-password-view__panel-8">
            <div class="views-auth-reset-password-view__panel-9">
              <Icon name="lock" size="md" class="views-auth-reset-password-view__icon-3" />
            </div>
            <input
              id="confirmPassword"
              v-model="formData.confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              required
              autocomplete="new-password"
              :disabled="isLoading"
              class="views-auth-reset-password-view__field-2 input"
              :class="{ 'input-error': errors.confirmPassword }"
              :placeholder="t('auth.confirmPasswordPlaceholder')"
            />
            <button
              type="button"
              @click="showConfirmPassword = !showConfirmPassword"
              class="views-auth-reset-password-view__action"
            >
              <Icon v-if="showConfirmPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="isLoading"
          class="views-auth-reset-password-view__action-2 btn btn-primary"
        >
          <svg
            v-if="isLoading"
            class="views-auth-reset-password-view__icon-4"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="views-auth-reset-password-view__circle"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="views-auth-reset-password-view__path"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <Icon v-else name="checkCircle" size="md" class="views-auth-reset-password-view__icon-5" />
          {{ isLoading ? t('auth.resettingPassword') : t('auth.resetPassword') }}
        </button>
      </form>
    </div>

    <!-- Footer -->
    <template #footer>
      <p class="views-auth-reset-password-view__description-4">
        {{ t('auth.rememberedPassword') }}
        <router-link
          to="/login"
          class="views-auth-reset-password-view__router-link-3"
        >
          {{ t('auth.signIn') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { resetPassword } from '@/api/auth'

const { t } = useI18n()

// ==================== Router & Stores ====================

const route = useRoute()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const isSuccess = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const showConfirmPassword = ref<boolean>(false)

// URL parameters
const email = ref<string>('')
const token = ref<string>('')

const formData = reactive({
  password: '',
  confirmPassword: ''
})

const errors = reactive({
  password: '',
  confirmPassword: ''
})

const validationToastMessage = computed(
  () => errors.password || errors.confirmPassword || ''
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// Check if the reset link is valid (has email and token)
const isInvalidLink = computed(() => !email.value || !token.value)

// ==================== Lifecycle ====================

onMounted(() => {
  // Get email and token from URL query parameters
  email.value = (route.query.email as string) || ''
  token.value = (route.query.token as string) || ''

  if (!email.value || !token.value) {
    appStore.showError(t('auth.invalidResetLink'))
  }
})

// ==================== Validation ====================

function validateForm(): boolean {
  errors.password = ''
  errors.confirmPassword = ''

  let isValid = true

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Confirm password validation
  if (!formData.confirmPassword) {
    errors.confirmPassword = t('auth.confirmPasswordRequired')
    isValid = false
  } else if (formData.password !== formData.confirmPassword) {
    errors.confirmPassword = t('auth.passwordsDoNotMatch')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    await resetPassword({
      email: email.value,
      token: token.value,
      new_password: formData.password
    })

    isSuccess.value = true
    appStore.showSuccess(t('auth.passwordResetSuccess'))
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { detail?: string; code?: string } } }

    // Check for invalid/expired token error
    if (err.response?.data?.code === 'INVALID_RESET_TOKEN') {
      errorMessage.value = t('auth.invalidOrExpiredToken')
    } else if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.resetPasswordFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
