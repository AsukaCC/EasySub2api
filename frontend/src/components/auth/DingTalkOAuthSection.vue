<template>
  <div class="components-auth-ding-talk-oauth-section__panel">
    <button type="button" :disabled="disabled" class="components-auth-ding-talk-oauth-section__action btn btn-secondary" @click="startLogin">
      <svg
        class="components-auth-ding-talk-oauth-section__icon icon"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        aria-hidden="true"
        style="flex-shrink: 0"
      >
        <circle cx="12" cy="12" r="12" fill="#1677FF" />
        <text
          x="12"
          y="17"
          font-family="sans-serif"
          font-size="13"
          font-weight="bold"
          fill="white"
          text-anchor="middle"
        >D</text>
      </svg>
      {{ t('auth.dingtalk.signIn') }}
    </button>

    <div v-if="showDivider" class="components-auth-ding-talk-oauth-section__panel-2">
      <div class="components-auth-ding-talk-oauth-section__panel-3"></div>
      <span class="components-auth-ding-talk-oauth-section__text">
        {{ t('auth.oauthOrContinue') }}
      </span>
      <div class="components-auth-ding-talk-oauth-section__panel-3"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import type { OAuthLoginStart } from '@/api/auth'
import { resolveAffiliateReferralCode, storeOAuthAffiliateCode } from '@/utils/oauthAffiliate'

const props = withDefaults(defineProps<{
  disabled?: boolean
  affCode?: string
  showDivider?: boolean
}>(), {
  showDivider: true
})
const emit = defineEmits<{
  start: [request: OAuthLoginStart]
}>()

const route = useRoute()
const { t } = useI18n()

function startLogin(): void {
  const redirectTo = (route.query.redirect as string) || '/dashboard'
  storeOAuthAffiliateCode(resolveAffiliateReferralCode(props.affCode, route.query.aff, route.query.aff_code))
  emit('start', { provider: 'dingtalk', params: { redirect: redirectTo } })
}
</script>
