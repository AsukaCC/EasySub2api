<template>
  <div class="components-auth-wechat-oauth-section__panel">
    <button type="button" :disabled="buttonDisabled" class="components-auth-wechat-oauth-section__action btn btn-secondary" @click="startLogin">
      <span
        class="components-auth-wechat-oauth-section__text"
      >
        W
      </span>
      {{ t('auth.oidc.signIn', { providerName }) }}
    </button>

    <p
      v-if="disabledHint"
      data-testid="wechat-oauth-hint"
      class="components-auth-wechat-oauth-section__description"
    >
      {{ disabledHint }}
    </p>

    <div v-if="showDivider" class="components-auth-wechat-oauth-section__panel-2">
      <div class="components-auth-wechat-oauth-section__panel-3"></div>
      <span class="components-auth-wechat-oauth-section__text-2">
        {{ t('auth.oauthOrContinue') }}
      </span>
      <div class="components-auth-wechat-oauth-section__panel-3"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { resolveWeChatOAuthStart, type OAuthLoginStart } from '@/api/auth'
import { useAppStore } from '@/stores'
import { resolveAffiliateReferralCode, storeOAuthAffiliateCode } from '@/utils/oauthAffiliate'

const props = withDefaults(defineProps<{
  disabled?: boolean
  affCode?: string
  promoCode?: string
  showDivider?: boolean
}>(), {
  showDivider: true,
})
const emit = defineEmits<{
  start: [request: OAuthLoginStart]
}>()

const appStore = useAppStore()
const route = useRoute()
const { t, locale } = useI18n()
const providerName = computed(() => t('auth.wechatProviderName'))

function localizeWeChatHint(zh: string, en: string): string {
  return locale.value.startsWith('zh') ? zh : en
}

const resolvedStart = computed(() => resolveWeChatOAuthStart(appStore.cachedPublicSettings))
const buttonDisabled = computed(() => props.disabled || resolvedStart.value.mode === null)
const disabledHint = computed(() => {
  if (props.disabled) {
    return ''
  }
  switch (resolvedStart.value.unavailableReason) {
    case 'external_browser_required':
      return t('auth.oauthFlow.wechatSystemBrowserOnly')
    case 'wechat_browser_required':
      return t('auth.oauthFlow.wechatBrowserOnly')
    case 'native_app_required':
      return localizeWeChatHint(
        '当前仅配置微信移动应用登录，需要在原生 App 中通过微信 SDK 发起授权。',
        'This site only has WeChat mobile app login configured. Continue from the native app through the WeChat SDK.',
      )
    case 'not_configured':
      return t('auth.oauthFlow.wechatNotConfigured')
    default:
      return ''
  }
})

onMounted(() => {
  if (!appStore.cachedPublicSettings && !appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})

function startLogin(): void {
  if (buttonDisabled.value || !resolvedStart.value.mode) {
    return
  }
  const redirectTo = (route.query.redirect as string) || '/dashboard'
  storeOAuthAffiliateCode(resolveAffiliateReferralCode(props.affCode, route.query.aff, route.query.aff_code))
  const mode = resolvedStart.value.mode
  const params: Record<string, string> = { mode, redirect: redirectTo }
  const promoCode = props.promoCode?.trim()
  if (promoCode) {
    params.promo_code = promoCode
  }
  emit('start', {
    provider: 'wechat',
    params
  })
}
</script>
